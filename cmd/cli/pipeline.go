package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	config_charts "github.com/goptics/vizb/config/charts"
	"github.com/goptics/vizb/pkg/parser"
	goparser "github.com/goptics/vizb/pkg/parser/golang"
	"github.com/goptics/vizb/pkg/style"
	"github.com/goptics/vizb/pkg/template"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// LinearDefaults holds the dataset-level defaults a command bakes into
// Settings.Sort/Scale/ShowLabels. A single chart subcommand fills these from its
// own flags; the root command fills them from its global flags.
type LinearDefaults struct {
	Sort       string
	Scale      string
	ShowLabels bool
}

// ChartSelection is one chart type plus its resolved per-chart settings. Settings
// may be the zero value (no per-chart overrides beyond the global defaults).
type ChartSelection struct {
	Type     string
	Settings shared.ChartSettings
}

// SelectionsFromCharts pairs each chart type with its per-chart settings (zero
// value when the type has none). Used by the root command to combine --charts
// with the parsed --chart specs.
func SelectionsFromCharts(charts []string, specs map[string]shared.ChartSettings) []ChartSelection {
	out := make([]ChartSelection, 0, len(charts))
	for _, c := range charts {
		out = append(out, ChartSelection{Type: c, Settings: specs[c]})
	}
	return out
}

// RunLinear runs the full linear pipeline shared by the root command and every
// linear chart subcommand: resolve input (file/stdin) → optional Dataset JSON
// passthrough → parse → assemble Dataset → write HTML/JSON → handle output.
//
// applyOnPassthrough controls whether selections override a passed-through
// Dataset's baked chart selection. Chart subcommands pass true (explicit single
// chart intent); the root command passes false (preserve the dataset as-is).
func RunLinear(cmd *cobra.Command, args []string, common CommonOptions, defaults LinearDefaults, selections []ChartSelection, applyOnPassthrough bool) {
	target, ok := resolveInput(cmd, args)
	if !ok {
		return
	}

	// In 'auto' mode (the default), sniff the input content and surface the
	// auto-selected parser so the choice is never silent.
	if common.Parser == "auto" {
		detected := parser.DetectParser(target)
		common.Parser = detected
		fmt.Println(style.Info.Render("✨ Auto-detected parser: " + detected))
	}

	cfg := common.ParseConfig()
	outFile := ResolveOutputFileName(common.OutputFile)

	// First try to read the input as an existing vizb Dataset JSON.
	dataSet := convertToDataset(target)
	if dataSet == nil {
		// Not Dataset JSON: parse raw/bench input into data points.
		target = preprocessInputFile(target, common.Parser)
		results := prepareData(target, common.Parser, cfg)
		dataSet = assembleDataset(results, common, defaults, cfg, selections)
	} else if applyOnPassthrough {
		applySelections(dataSet, defaults, selections)
	}

	f := shared.MustCreateFile(outFile)
	defer f.Close()

	writeOutput(f, dataSet, InferFormatFromExtension(outFile))

	HandleOutputResult(f, common.OutputFile)
}

// RunSingleChart is the entry point for a single-chart subcommand. It validates
// the --swap value against the active grouping, assembles one ChartSelection,
// and runs the shared linear pipeline.
func RunSingleChart(cmd *cobra.Command, args []string, common CommonOptions, defaults LinearDefaults, chartType, swap string, autoRotate *bool) {
	axes := parser.GroupAxes(common.ParseConfig())
	if err := shared.ValidateSwap(swap, axes); err != nil {
		shared.ExitWithError(err.Error(), nil)
	}

	sel := ChartSelection{Type: chartType, Settings: shared.ChartSettings{Swap: swap, AutoRotate: autoRotate}}
	RunLinear(cmd, args, common, defaults, []ChartSelection{sel}, true)
}

// ValidateScale normalises and validates a scale flag (linear/log), falling back
// to "linear" with a warning. Used by bar/line which expose --scale.
func ValidateScale(scale *string) {
	utils.ApplyValidationRules([]utils.ValidationRule{{
		Label:      "scale",
		Value:      scale,
		ValidSet:   []string{"linear", "log"},
		Normalizer: strings.ToLower,
		Default:    "linear",
	}})
}

// resolveInput returns the input file path. It accepts a file arg, else reads
// piped stdin into a temp file. With neither, it prints help and exits.
func resolveInput(cmd *cobra.Command, args []string) (string, bool) {
	stat, _ := os.Stdin.Stat()
	isStdinPiped := (stat.Mode() & os.ModeCharDevice) == 0

	if len(args) > 0 {
		target := args[0]
		checkTargetFile(target)
		return target, true
	}

	if isStdinPiped {
		target := shared.MustCreateTempFile(shared.TempBenchFilePrefix, "out")
		shared.TempFiles.Store(target)
		writeStdinPipedInputs(target)
		return target, true
	}

	cmd.Help()
	shared.OsExit(0)
	return "", false
}

func checkTargetFile(filePath string) {
	fmt.Println(style.Info.Render(fmt.Sprintf("🔎 Reading data from file: %s", filePath)))

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		shared.ExitWithError(fmt.Sprintf("Error: File '%s' does not exist", filePath), nil)
	}
}

func writeStdinPipedInputs(tempfilePath string) {
	inputTempFile := shared.MustCreateFile(tempfilePath)
	defer inputTempFile.Close()

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(inputTempFile)

	dataSetProgressManager := NewBenchmarkProgressManager(
		progressbar.NewOptions(-1,
			progressbar.OptionSetDescription(style.Info.Render("Processing data sets")),
			progressbar.OptionSetWidth(50),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionOnCompletion(func() { fmt.Println() }),
		),
	)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			shared.ExitWithError("Error reading from stdin", err)
		}

		if _, err := writer.WriteString(line); err != nil {
			shared.ExitWithError("Error writing to file", err)
		}

		dataSetProgressManager.ProcessLine(line)
	}

	dataSetProgressManager.Finish()

	writer.Flush()
	inputTempFile.Sync()
}

// preprocessInputFile handles Go bench JSON → TXT conversion when needed.
func preprocessInputFile(filePath, parserKey string) string {
	if parserKey == "go" && utils.IsBenchJSONFile(filePath) {
		return goparser.ConvertGoJsonBenchToText(filePath)
	}

	return filePath
}

// prepareData parses input into data points, aggregating grouped csv/json rows.
func prepareData(filePath, parserKey string, cfg parser.Config) []shared.DataPoint {
	parseFn, err := parser.GetParser(parserKey)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}

	fmt.Println(style.Info.Render("⚙️  Parsing data..."))
	data := parseFn(filePath, cfg)

	// CSV/JSON emit one DataPoint per row; when grouping is active, multiple rows
	// can share the same (name, xAxis, yAxis, zAxis) key. Collapse them by summing
	// so the output isn't a row-per-record dump. Benchmark parsers are excluded.
	if (parserKey == "csv" || parserKey == "json") && len(cfg.Group) > 0 {
		before := len(data)
		fmt.Println(style.Info.Render(fmt.Sprintf("🧮 Aggregating %d rows...", before)))
		data = shared.AggregateDataPoints(data)
		fmt.Println(style.Info.Render(fmt.Sprintf("✅ Aggregated into %d grouped data points", len(data))))
	}

	if len(data) == 0 {
		shared.ExitWithError("No dataSet data found", nil)
	}

	return data
}

// assembleDataset builds the output Dataset from parsed results plus the command's
// metadata, defaults, and chart selections.
func assembleDataset(results []shared.DataPoint, common CommonOptions, defaults LinearDefaults, cfg parser.Config, selections []ChartSelection) *shared.Dataset {
	dataSet := &shared.Dataset{
		Name:        common.Name,
		Description: common.Description,
		Data:        results,
	}

	meta := shared.Meta{OS: shared.OS, Arch: shared.Arch, Pkg: shared.Pkg}
	if cpuName := strings.TrimSpace(shared.CPU); cpuName != "" || shared.CPUCount != 0 {
		meta.CPU = &shared.CPUInfo{Name: cpuName, Cores: shared.CPUCount}
	}
	if meta != (shared.Meta{}) {
		dataSet.Meta = &meta
	}

	// Task 2 keeps the call sites compiling; the per-chart Config assembly
	// itself is rewritten in Task 3 (which uses each chart's Materialise
	// function). For now we only need the build to pass — the test fixture
	// failures on Settings assertions are expected and noted.
	_ = selectionTypes
	_ = applyDefaults
	_ = selectionSettings
	_ = defaults
	_ = selections
	dataSet.Axes = parser.GroupAxes(cfg)

	dataSet.Tag = common.Tag
	dataSet.Timestamp = time.Now().UTC().Format(time.RFC3339)

	return dataSet
}

// applySelections overrides a passed-through Dataset's chart selection and
// defaults, so e.g. `vizb bar data.json` re-renders as bar only. Task 2 keeps
// this compiling; the rewrite is in Task 3.
func applySelections(dataSet *shared.Dataset, defaults LinearDefaults, selections []ChartSelection) {
	_ = defaults
	_ = selections
	_ = selectionTypes
	_ = applyDefaults
	_ = selectionSettings
}

// chartTypeNames extracts the chart-type discriminators from a slice of
// per-chart Configs. Used by callers that need a []string (e.g. HTML chunk
// pruning) in place of the old dataSet.Settings.Charts list.
func chartTypeNames(settings []config_charts.ChartConfig) []string {
	out := make([]string, 0, len(settings))
	for _, c := range settings {
		out = append(out, c.ChartType())
	}
	return out
}

// applyDefaults writes the dataset-level sort/scale/labels defaults into Settings.
func applyDefaults(s *shared.DatasetSettings, defaults LinearDefaults) {
	enableSorting := defaults.Sort != ""
	s.Sort.Enabled = enableSorting
	if enableSorting {
		s.Sort.Order = defaults.Sort
	} else {
		s.Sort.Order = "asc"
	}
	s.ShowLabels = defaults.ShowLabels
	s.Scale = defaults.Scale
}

func selectionTypes(selections []ChartSelection) []string {
	types := make([]string, 0, len(selections))
	for _, s := range selections {
		types = append(types, s.Type)
	}
	return types
}

// selectionSettings collects the non-zero per-chart settings keyed by type, or
// nil when none have overrides.
func selectionSettings(selections []ChartSelection) map[string]shared.ChartSettings {
	m := make(map[string]shared.ChartSettings)
	for _, s := range selections {
		if s.Settings != (shared.ChartSettings{}) {
			m[s.Type] = s.Settings
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// writeOutput writes the dataset to f as HTML or JSON.
func writeOutput(f *os.File, dataSet *shared.Dataset, format string) {
	switch format {
	case "html":
		fmt.Println(style.Info.Render("🔄 Generating UI..."))

		jsonData, err := json.Marshal(dataSet)
		if err != nil {
			shared.ExitWithError("Failed to marshal dataSet data: %v", err)
		}

		htmlContent := template.GenerateUI(jsonData, chartTypeNames(dataSet.Settings), shared.DatasetNeeds3D(dataSet), template.VizbHTMLTemplate)
		if _, err := f.WriteString(htmlContent); err != nil {
			shared.ExitWithError("Failed to write output file: %v", err)
		}

		fmt.Println(style.Success.Render("🎉 Generated HTML UI successfully!"))

	case "json":
		fmt.Println(style.Info.Render("🔄 Generating JSON..."))
		bytes, err := json.Marshal(dataSet)
		if err != nil {
			shared.ExitWithError("Error marshaling dataSet data", err)
		}
		f.Write(bytes)
		fmt.Println(style.Success.Render("🎉 Generated JSON successfully!"))
	}
}
