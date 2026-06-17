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

// RunLinear runs the full linear pipeline shared by the root command and every
// linear chart subcommand: resolve input (file/stdin) → optional Dataset JSON
// passthrough → parse → assemble Dataset → write HTML/JSON → handle output.
//
// applyOnPassthrough controls whether the provided configs override a
// passed-through Dataset's baked chart selection. Chart subcommands pass true
// (explicit single chart intent); the root command passes false (preserve the
// dataset as-is).
func RunLinear(cmd *cobra.Command, args []string, common CommonOptions, configs []config_charts.ChartConfig, applyOnPassthrough bool) {
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
		dataSet = assembleDataset(results, common, configs, cfg)
	} else if applyOnPassthrough {
		applySelections(dataSet, configs)
	}

	f := shared.MustCreateFile(outFile)
	defer f.Close()

	writeOutput(f, dataSet, InferFormatFromExtension(outFile))

	HandleOutputResult(f, common.OutputFile)
}

// RunSingleChart is the entry point for a single-chart subcommand. It forwards
// the per-chart Config (built by the subcommand via its chart's Materialise) to
// the shared linear pipeline. An empty configs slice is treated as a no-op so
// callers can defensively guard against misconfiguration.
func RunSingleChart(cmd *cobra.Command, args []string, common CommonOptions, configs []config_charts.ChartConfig) {
	if len(configs) == 0 {
		return
	}
	RunLinear(cmd, args, common, configs, true)
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

// assembleDataset builds the output Dataset from parsed results plus the
// command's metadata and the resolved per-chart configs.
func assembleDataset(results []shared.DataPoint, common CommonOptions, configs []config_charts.ChartConfig, cfg parser.Config) *shared.Dataset {
	dataSet := &shared.Dataset{
		Name:        common.Name,
		Description: common.Description,
		Data:        results,
		Settings:    configs,
		Axes:        parser.GroupAxes(cfg),
	}

	meta := shared.Meta{OS: shared.OS, Arch: shared.Arch, Pkg: shared.Pkg}
	if cpuName := strings.TrimSpace(shared.CPU); cpuName != "" || shared.CPUCount != 0 {
		meta.CPU = &shared.CPUInfo{Name: cpuName, Cores: shared.CPUCount}
	}
	if meta != (shared.Meta{}) {
		dataSet.Meta = &meta
	}

	dataSet.Tag = common.Tag
	dataSet.Timestamp = time.Now().UTC().Format(time.RFC3339)

	return dataSet
}

// applySelections overrides a passed-through Dataset's chart selection so e.g.
// `vizb bar data.json` re-renders with the new configs.
func applySelections(dataSet *shared.Dataset, configs []config_charts.ChartConfig) {
	dataSet.Settings = configs
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
