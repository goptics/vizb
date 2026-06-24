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
	barchart "github.com/goptics/vizb/config/charts/bar"
	linechart "github.com/goptics/vizb/config/charts/line"
	scatterchart "github.com/goptics/vizb/config/charts/scatter"
	"github.com/goptics/vizb/pkg/parser"
	goparser "github.com/goptics/vizb/pkg/parser/golang"
	jsonparser "github.com/goptics/vizb/pkg/parser/json"
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
	RunLinearWithConfig(cmd, args, common, common.ParseConfig(), configs, applyOnPassthrough)
}

func RunLinearWithConfig(cmd *cobra.Command, args []string, common CommonOptions, cfg parser.Config, configs []config_charts.ChartConfig, applyOnPassthrough bool) {
	target, ok := resolveInput(cmd, args)
	if !ok {
		return
	}

	// In 'auto' mode (the default), sniff the input content and surface the
	// auto-selected parser so the choice is never silent.
	if common.Parser == "auto" {
		detected := parser.DetectParser(target)
		// --json-path only makes sense for JSON; an envelope file starts with '{'
		// which auto-detect reads as the "go" fallback, so nudge it to json.
		if cfg.JSONPath != "" && detected != "json" {
			detected = "json"
		}
		common.Parser = detected
		fmt.Println(style.Info.Render("✨ Auto-detected parser: " + detected))
	}

	// Enable auto-grouping for the csv/json parsers when the user supplied no
	// explicit grouping. The csv/json parsers infer the category axis from the
	// data so `vizb data.csv` produces a usable chart without -g/-p/-r.
	if (common.Parser == "csv" || common.Parser == "json") && parser.NoExplicitGrouping(cfg) {
		cfg.AutoGroup = true
	}
	for _, c := range configs {
		cfg.ChartTypes = append(cfg.ChartTypes, c.ChartType())
	}

	WarnThreeDIfIneligible(cfg, configs)
	outFile := ResolveOutputFileName(common.OutputFile)

	// First try to read the input as an existing vizb Dataset JSON. --json-path
	// explicitly marks the input as raw enveloped data, not a vizb Dataset, so
	// skip the passthrough (an envelope object would otherwise unmarshal into an
	// empty Dataset and silently produce no output).
	var dataSet *shared.Dataset
	if cfg.JSONPath == "" {
		dataSet = convertToDataset(target)
	}
	if dataSet == nil {
		// Not Dataset JSON: parse raw/bench input into data points.
		target = preprocessInputFile(target, common.Parser)
		if common.Parser == "json" && cfg.JSONPath != "" {
			target = applyJSONPath(target, cfg.JSONPath)
		}
		results := prepareData(target, common.Parser, cfg)
		dataSet = assembleDataset(results, common, configs, cfg)
		// Validate swap only for chart subcommands (applyOnPassthrough true).
		// The root command stores swap as-is, trusting the UI to handle it.
		if applyOnPassthrough {
			for _, cc := range configs {
				if swp := cc.SwapString(); swp != "" {
					if err := shared.ValidateSwap(swp, dataSet.Axes); err != nil {
						shared.ExitWithError(err.Error(), nil)
					}
				}
			}
		}
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
	RunSingleChartWithConfig(cmd, args, common, common.ParseConfig(), configs)
}

func RunSingleChartWithConfig(cmd *cobra.Command, args []string, common CommonOptions, cfg parser.Config, configs []config_charts.ChartConfig) {
	if len(configs) == 0 {
		return
	}
	RunLinearWithConfig(cmd, args, common, cfg, configs, true)
}

// WarnThreeDIfIneligible prints a warning when threeD is baked on a bar/line
// config but the group pattern does not declare both x and y. The flag is kept.
func WarnThreeDIfIneligible(cfg parser.Config, configs []config_charts.ChartConfig) {
	if parser.PatternHasBothXY(cfg) || !shared.SettingsHasThreeDOption(configs) {
		return
	}
	shared.PrintWarning("Warning: --3d requires both x and y in --group-pattern; value 3D will not render.")
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

	_ = cmd.Help()
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

	dataSetProgressManager := NewDataProgressManager(
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

		// ReadString returns the final chunk alongside io.EOF when the input has
		// no trailing newline, so write it before handling the error — else a
		// single-line, newline-less payload (e.g. a curl'd JSON envelope) is lost.
		if len(line) > 0 {
			if _, werr := writer.WriteString(line); werr != nil {
				shared.ExitWithError("Error writing to file", werr)
			}
			dataSetProgressManager.ProcessLine(line)
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			shared.ExitWithError("Error reading from stdin", err)
		}
	}

	if err := dataSetProgressManager.Finish(); err != nil {
		shared.ExitWithError("Error finishing progress bar", err)
	}

	if err := writer.Flush(); err != nil {
		shared.ExitWithError("Error writing to file", err)
	}
	_ = inputTempFile.Sync()
}

// preprocessInputFile handles Go bench JSON → TXT conversion when needed.
func preprocessInputFile(filePath, parserKey string) string {
	if parserKey == "go" && utils.IsBenchJSONFile(filePath) {
		return goparser.ConvertGoJsonBenchToText(filePath)
	}

	return filePath
}

// applyJSONPath extracts the array named by cfg.JSONPath from the input file and
// writes it to a temp file, which is returned for the JSON parser to consume.
func applyJSONPath(filePath, path string) string {
	bytes, err := jsonparser.SelectPath(filePath, path)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}

	out := shared.MustCreateTempFile(shared.TempBenchFilePrefix, "json")
	shared.TempFiles.Store(out)
	if err := os.WriteFile(out, bytes, 0o600); err != nil {
		shared.ExitWithError("Error writing json-path result", err)
	}
	return out
}

// prepareData parses input into data points, aggregating grouped csv/json rows.
func prepareData(filePath, parserKey string, cfg parser.Config) []shared.DataPoint {
	parseFn, err := parser.GetParser(parserKey)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}

	if cfg.JSONPath != "" && parserKey != "json" {
		fmt.Fprintln(os.Stderr, "warning: --json-path is only supported for the json parser; ignoring")
	}

	if len(cfg.Select) > 0 && parserKey != "csv" && parserKey != "json" {
		fmt.Fprintln(os.Stderr, "warning: --select is only supported for csv/json parsers; ignoring")
	}

	if len(cfg.Axes) > 0 && parserKey != "csv" && parserKey != "json" {
		shared.ExitWithError("--axes is only supported for csv/json parsers", nil)
	}

	fmt.Println(style.Info.Render("🧲 Parsing data..."))
	data := parseFn(filePath, cfg)

	// CSV/JSON emit one DataPoint per row; when grouping is active, multiple rows
	// can share the same (name, xAxis, yAxis, zAxis) key. Collapse them by summing
	// so the output isn't a row-per-record dump. Benchmark parsers are excluded.
	if (parserKey == "csv" || parserKey == "json") && len(cfg.Group) > 0 {
		before := len(data)
		fmt.Println(style.Info.Render(fmt.Sprintf("🧮 Aggregating %d rows %s...", before, formatAggregationGroup(cfg))))
		data = shared.AggregateDataPoints(data)
		fmt.Println(style.Info.Render(fmt.Sprintf("✅ Aggregated into %d grouped data points", len(data))))
	}

	if len(data) == 0 {
		shared.ExitWithError("No dataSet data found", nil)
	}

	return data
}

// formatAggregationGroup describes the --group columns and dimension keys used
// when collapsing duplicate CSV/JSON rows.
func formatAggregationGroup(cfg parser.Config) string {
	cols := parser.EffectiveGroupColumns(cfg)
	colList := strings.Join(cols, ", ")
	colPhrase := "by columns: " + colList
	if len(cols) == 1 {
		colPhrase = "by column: " + colList
	}

	axes := parser.GroupAxes(cfg)
	if len(axes) == 0 {
		return colPhrase
	}

	dims := make([]string, 0, len(axes))
	for _, ax := range axes {
		if ax.Label != "" {
			dims = append(dims, ax.Key+": "+ax.Label)
			continue
		}
		dims = append(dims, ax.Key)
	}
	return colPhrase + " (" + strings.Join(dims, ", ") + ")"
}

// deriveAxesFromData scans parsed data points to determine which axes are
// populated. Used when auto-grouping modified the config inside the parser
// (invisible to the caller since Config is passed by value).
func deriveAxesFromData(results []shared.DataPoint) []shared.Axis {
	var axes []shared.Axis
	if len(results) == 0 {
		return axes
	}
	// Value mode is signaled by empty Stats + populated coordinate axes.
	valueMode := len(results[0].Stats) == 0
	hasName, hasX, hasY, hasZ := false, false, false, false
	for _, dp := range results {
		if dp.Name != "" {
			hasName = true
		}
		if dp.XAxis != "" {
			hasX = true
		}
		if dp.YAxis != "" {
			hasY = true
		}
		if dp.ZAxis != "" {
			hasZ = true
		}
	}
	axisType := ""
	if valueMode {
		axisType = "value"
	}
	if hasName {
		axes = append(axes, shared.Axis{Key: "name", Type: axisType})
	}
	if hasX {
		axes = append(axes, shared.Axis{Key: "x", Type: axisType})
	}
	if hasY {
		axes = append(axes, shared.Axis{Key: "y", Type: axisType})
	}
	if hasZ {
		axes = append(axes, shared.Axis{Key: "z", Type: axisType})
	}
	return axes
}

func isValueXYZAxes(axes []shared.Axis) bool {
	keys := map[string]bool{}
	for _, a := range axes {
		if a.Key == "metric" {
			continue
		}
		if a.Type != "value" {
			return false
		}
		keys[a.Key] = true
	}
	return keys["x"] && keys["y"] && keys["z"]
}

func appendMetricAxis(axes []shared.Axis, cfg parser.Config, results []shared.DataPoint) []shared.Axis {
	if !valueModeHasMetric(cfg, results) {
		return axes
	}
	for _, a := range axes {
		if a.Key == "metric" {
			return axes
		}
	}
	label := cfg.MetricColumn
	if label == "" {
		label = "value"
	}
	return append(axes, shared.Axis{Key: "metric", Label: label, Type: "value"})
}

func valueModeHasMetric(cfg parser.Config, results []shared.DataPoint) bool {
	if cfg.MetricColumn != "" {
		return true
	}
	for _, dp := range results {
		if dp.Metric != "" {
			return true
		}
	}
	return false
}

func applyValueMode3DFlags(threeD **bool, visualMap **bool, v bool, enableVM bool) {
	*threeD = &v
	if enableVM {
		*visualMap = &v
	}
}

// autoEnableValueMode3D sets ThreeD (and ThreeDVisualMap when metric present) on
// bar/line/scatter configs for continuous xyz value mode.
func autoEnableValueMode3D(configs []config_charts.ChartConfig, axes []shared.Axis, visualMap bool) {
	if !isValueXYZAxes(axes) {
		return
	}
	v := true
	for i, c := range configs {
		switch bc := c.(type) {
		case barchart.Config:
			applyValueMode3DFlags(&bc.ThreeD, &bc.ThreeDVisualMap, v, visualMap)
			configs[i] = bc
		case linechart.Config:
			applyValueMode3DFlags(&bc.ThreeD, &bc.ThreeDVisualMap, v, visualMap)
			configs[i] = bc
		case scatterchart.Config:
			applyValueMode3DFlags(&bc.ThreeD, &bc.ThreeDVisualMap, v, visualMap)
			configs[i] = bc
		}
	}
}

// assembleDataset builds the output Dataset from parsed results plus the
// command's metadata and the resolved per-chart configs.
func assembleDataset(results []shared.DataPoint, common CommonOptions, configs []config_charts.ChartConfig, cfg parser.Config) *shared.Dataset {
	var axes []shared.Axis
	if cfg.AutoGroup {
		// Auto-grouping modified the config inside the parser; derive axes
		// from the actual data points since the caller's cfg is unchanged.
		axes = deriveAxesFromData(results)
		autoEnableValueMode3D(configs, axes, valueModeHasMetric(cfg, results))
	} else {
		axes = parser.GroupAxes(cfg)
		if len(cfg.Axes) > 0 {
			axes = parser.ValueAxes(cfg)
		}
	}
	axes = appendMetricAxis(axes, cfg, results)

	dataSet := &shared.Dataset{
		Name:        common.Name,
		Description: common.Description,
		Data:        results,
		Settings:    configs,
		Axes:        axes,
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

// settingsNeedCorrelation reports whether any chart setting in the slice needs
// the correlation heatmap renderer shipped (math empty = all, or contains "correlations").
func settingsNeedCorrelation(settings []config_charts.ChartConfig) bool {
	for _, cfg := range settings {
		if shared.ChartConfigNeedsCorrelation(cfg) {
			return true
		}
	}
	return false
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

		needsHeatmapChunk := settingsNeedCorrelation(dataSet.Settings)
		htmlContent := template.GenerateUI(jsonData, chartTypeNames(dataSet.Settings), shared.DatasetNeeds3D(dataSet), needsHeatmapChunk, template.VizbHTMLTemplate)
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
		if _, err := f.Write(bytes); err != nil {
			shared.ExitWithError("Failed to write output file", err)
		}
		fmt.Println(style.Success.Render("🎉 Generated JSON successfully!"))
	}
}
