package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	internal_charts "github.com/goptics/vizb/internal/charts"
	barchart "github.com/goptics/vizb/internal/charts/bar"
	linechart "github.com/goptics/vizb/internal/charts/line"
	scatterchart "github.com/goptics/vizb/internal/charts/scatter"
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
//
// meta carries the dataset metadata (name/description/tag/output) and the
// selected parser; cfg is the resolved parser.Config (built by the caller's
// FlagBag.ParseConfig).
func RunLinear(cmd *cobra.Command, args []string, meta RunMeta, cfg parser.Config, configs []internal_charts.ChartConfig, applyOnPassthrough bool) {
	target, ok := resolveInput(cmd, args)
	if !ok {
		return
	}

	// In 'auto' mode (the default), sniff the input content and surface the
	// auto-selected parser so the choice is never silent.
	if meta.Parser == "auto" {
		detected := parser.DetectParser(target)
		// --json-path only makes sense for JSON; an envelope file starts with '{'
		// which auto-detect reads as the "go" fallback, so nudge it to json.
		if cfg.JSONPath != "" && detected != "json" {
			detected = "json"
		}
		meta.Parser = detected
		fmt.Println(style.Info.Render("✨ Auto-detected parser: " + detected))
	}

	// Enable auto-grouping for the csv/json parsers when the user supplied no
	// explicit grouping. The csv/json parsers infer the category axis from the
	// data so `vizb data.csv` produces a usable chart without -g/-p/-r.
	if (meta.Parser == "csv" || meta.Parser == "json") && parser.NoExplicitGrouping(cfg) && !parser.HasSelect(cfg) {
		cfg.AutoGroup = true
	}
	for _, c := range configs {
		cfg.ChartTypes = append(cfg.ChartTypes, c.ChartType())
	}

	outFile := ResolveOutputFileName(meta.OutputFile)

	// First try to read the input as an existing vizb Dataset JSON. --json-path
	// explicitly marks the input as raw enveloped data, not a vizb Dataset, so
	// skip the passthrough (an envelope object would otherwise unmarshal into an
	// empty Dataset and silently produce no output).
	var datasets []*shared.Dataset
	if cfg.JSONPath == "" {
		if ds := convertToDataset(target); ds != nil {
			datasets = []*shared.Dataset{ds}
		}
	}
	if len(datasets) == 0 {
		// Not Dataset JSON: parse raw/bench input into data points.
		target = preprocessInputFile(target, meta.Parser)
		if meta.Parser == "json" && cfg.JSONPath != "" {
			target = applyJSONPath(target, cfg.JSONPath)
		}
		results, effectiveCfg := prepareData(target, meta.Parser, cfg)
		datasets = []*shared.Dataset{assembleDataset(results, meta, configs, effectiveCfg)}
		// Validate swap only for chart subcommands (applyOnPassthrough true).
		// The root command stores swap as-is, trusting the UI to handle it.
		if applyOnPassthrough {
			for _, dataSet := range datasets {
				for _, cc := range configs {
					if swp := cc.SwapString(); swp != "" {
						if err := shared.ValidateSwap(swp, dataSet.Axes); err != nil {
							shared.ExitWithError(err.Error(), nil)
						}
					}
				}
			}
		}
	} else if applyOnPassthrough {
		applySelections(datasets[0], configs)
	}

	// Phase B: evaluate applicability rules on materialised configs with
	// data-derived axes. Rules are nil on every descriptor yet (Phase C adds
	// them), so this is a no-op in Phase B.
	for _, dataSet := range datasets {
		var ruleAxes []internal_charts.AxisInfo
		for _, a := range dataSet.Axes {
			ruleAxes = append(ruleAxes, internal_charts.AxisInfo{Key: a.Key, Type: a.Type})
		}
		ruleCtx := internal_charts.RuleContext{Axes: ruleAxes}
		warnings, fatal := internal_charts.ApplyRules(ruleCtx, configs)
		if fatal != nil {
			shared.ExitWithError(fatal.Error(), nil)
		}
		for _, w := range warnings {
			fmt.Fprintln(os.Stderr, style.Warning.Render(w))
		}
	}

	f := shared.MustCreateFile(outFile)
	defer f.Close()

	writeOutput(f, datasets, InferFormatFromExtension(outFile))

	HandleOutputResult(f, meta.OutputFile)
}

// RunSingleChart is the entry point for a single-chart subcommand. It forwards
// the per-chart Config (built by the subcommand via its chart's Materialise) to
// the shared linear pipeline. An empty configs slice is treated as a no-op so
// callers can defensively guard against misconfiguration.
func RunSingleChart(cmd *cobra.Command, args []string, meta RunMeta, cfg parser.Config, configs []internal_charts.ChartConfig) {
	if len(configs) == 0 {
		return
	}
	RunLinear(cmd, args, meta, cfg, configs, true)
}

// RunMeta carries the dataset metadata and selected parser from a command's
// FlagBag into the linear pipeline. It is plain value passing — flag declaration
// and validation live in the FlagBag, not here.
type RunMeta struct {
	ID          string
	Name        string
	Description string
	Tag         string
	OutputFile  string
	Parser      string
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
// The returned Config is the parser's effective config (auto-group/auto-value
// mutations included) for aggregation and dataset assembly.
func prepareData(filePath, parserKey string, cfg parser.Config) ([]shared.DataPoint, parser.Config) {
	parseFn, err := parser.GetParser(parserKey)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}

	if cfg.JSONPath != "" && parserKey != "json" {
		fmt.Fprintln(os.Stderr, "warning: --json-path is only supported for the json parser; ignoring")
	}

	if parser.HasSelect(cfg) && parserKey != "csv" && parserKey != "json" {
		fmt.Fprintln(os.Stderr, "warning: --select is only supported for csv/json parsers; ignoring")
	}

	if len(cfg.Axes) > 0 && parserKey != "csv" && parserKey != "json" {
		shared.ExitWithError("--axes is only supported for csv/json parsers", nil)
	}

	fmt.Println(style.Info.Render("🧲 Parsing data..."))
	data, effectiveCfg := parseFn(filePath, cfg)

	// CSV/JSON emit one DataPoint per row; when grouping is inactive, collapse rows
	// that share the same (name, x, y, z) by appending stats (no sum/average).
	if tabularParser(parserKey) && len(effectiveCfg.Group) == 0 {
		data = shared.CollapseDataPointsByKey(data)
	}

	// CSV/JSON emit one DataPoint per row; when grouping is active, multiple rows
	// can share the same (name, xAxis, yAxis, zAxis) key. Collapse them by summing
	// so the output isn't a row-per-record dump. Benchmark parsers are excluded.
	if tabularParser(parserKey) && len(effectiveCfg.Group) > 0 {
		before := len(data)
		data = shared.AggregateDataPoints(data)
		logAggregationResult(before, len(data), effectiveCfg)
	}

	if len(data) == 0 {
		shared.ExitWithError("No dataSet data found", nil)
	}

	return data, effectiveCfg
}

// logAggregationResult prints CLI feedback after summing grouped CSV/JSON rows.
// The opening line always reports the row count and group columns; the closing
// line differs when every key was already unique (no duplicates to sum).
func logAggregationResult(before, after int, cfg parser.Config) {
	if before == 0 {
		return
	}
	groupDesc := formatAggregationGroup(cfg)
	fmt.Println(style.Info.Render(fmt.Sprintf("🧮 Aggregating %d rows %s...", before, groupDesc)))
	if after < before {
		fmt.Println(style.Info.Render(fmt.Sprintf("✅ Aggregated into %d grouped data points", after)))
		return
	}
	fmt.Println(style.Info.Render(fmt.Sprintf("✅ %d grouped rows — all unique (no duplicates to sum)", after)))
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
func autoEnableValueMode3D(configs []internal_charts.ChartConfig, axes []shared.Axis, visualMap bool) {
	if !isValueXYZAxes(axes) {
		return
	}
	v := true
	for _, c := range configs {
		switch bc := c.(type) {
		case *barchart.Config:
			applyValueMode3DFlags(&bc.ThreeD, &bc.ThreeDVisualMap, v, visualMap)
		case *linechart.Config:
			applyValueMode3DFlags(&bc.ThreeD, &bc.ThreeDVisualMap, v, visualMap)
		case *scatterchart.Config:
			applyValueMode3DFlags(&bc.ThreeD, &bc.ThreeDVisualMap, v, visualMap)
		}
	}
}

// assembleDataset builds the output Dataset from parsed results plus the
// command's metadata and the resolved per-chart configs.
func assembleDataset(results []shared.DataPoint, m RunMeta, configs []internal_charts.ChartConfig, cfg parser.Config) *shared.Dataset {
	var view []parser.ColumnSpec
	if cfg.Mode.IsSelectAxis() && len(cfg.SelectViews) == 1 {
		view = cfg.SelectViews[0].Columns
	}
	name := ""
	if len(view) > 0 && cfg.Mode.IsSelectAxis() && !cfg.Mode.IsMultiStat() {
		name = parser.SelectViewDatasetName(view, 0)
	}
	return buildDataset(results, m, configs, cfg, view, name)
}

func tabularParser(parserKey string) bool {
	return parserKey == "csv" || parserKey == "json"
}

// preserveRowsForDataset is true for csv/json tabular data where the UI must
// expand collapsed stats[] or keep duplicate (x,y,z) keys (solo/multi --select).
// Grouped output is pre-aggregated to one point per cell; preserveRows there
// makes the UI emit one series row per DataPoint and duplicates x categories.
func preserveRowsForDataset(parserKey string, cfg parser.Config) bool {
	return tabularParser(parserKey) && len(cfg.Group) == 0
}

func buildDataset(results []shared.DataPoint, m RunMeta, configs []internal_charts.ChartConfig, cfg parser.Config, view []parser.ColumnSpec, viewName string) *shared.Dataset {
	var axes []shared.Axis
	if cfg.Mode.IsMultiStat() {
		axes = parser.MultiSelectStatAxes(cfg.SelectViews)
	} else if len(view) > 0 {
		axes = parser.DatasetAxesForSelectView(view, results)
		autoEnableValueMode3D(configs, axes, valueModeHasMetric(cfg, results))
	} else if cfg.Mode.IsSelectAxis() {
		axes = parser.DatasetAxesForSelectView(cfg.SelectViews[0].Columns, results)
		autoEnableValueMode3D(configs, axes, valueModeHasMetric(cfg, results))
	} else {
		axes = parser.GroupAxes(cfg)
		if len(cfg.Axes) > 0 {
			if parser.IsMixedAxes(cfg) {
				axes = parser.MixedAxes(cfg)
			} else {
				axes = parser.ValueAxes(cfg)
			}
			autoEnableValueMode3D(configs, axes, valueModeHasMetric(cfg, results))
		}
	}
	axes = appendMetricAxis(axes, cfg, results)

	name := m.Name
	if name == "" && viewName != "" {
		name = viewName
	}

	dataSet := &shared.Dataset{
		Name:         name,
		Description:  m.Description,
		Data:         results,
		Settings:     configs,
		Axes:         axes,
		PreserveRows: preserveRowsForDataset(m.Parser, cfg),
	}
	if id := strings.TrimSpace(m.ID); id != "" {
		dataSet.ID = id
	}

	meta := shared.Meta{OS: shared.OS, Arch: shared.Arch, Pkg: shared.Pkg}
	if cpuName := strings.TrimSpace(shared.CPU); cpuName != "" || shared.CPUCount != 0 {
		meta.CPU = &shared.CPUInfo{Name: cpuName, Cores: shared.CPUCount}
	}
	if meta != (shared.Meta{}) {
		dataSet.Meta = &meta
	}

	dataSet.Tag = m.Tag
	dataSet.Timestamp = time.Now().UTC().Format(time.RFC3339)

	return dataSet
}

// applySelections overrides a passed-through Dataset's chart selection so e.g.
// `vizb bar data.json` re-renders with the new configs.
func applySelections(dataSet *shared.Dataset, configs []internal_charts.ChartConfig) {
	dataSet.Settings = configs
}

// settingsNeedCorrelation reports whether any chart setting in the slice needs
// the correlation heatmap renderer shipped (math empty = all, or contains "correlations").
func settingsNeedCorrelation(settings []internal_charts.ChartConfig) bool {
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
func chartTypeNames(settings []internal_charts.ChartConfig) []string {
	out := make([]string, 0, len(settings))
	for _, c := range settings {
		out = append(out, c.ChartType())
	}
	return out
}

// writeOutput writes one or more datasets to f as HTML or JSON. HTML embeds an
// array when N>1 (like vizb ui); JSON keeps a single object when N=1 for backward
// compatibility.
func writeOutput(f *os.File, datasets []*shared.Dataset, format string) {
	if len(datasets) == 0 {
		return
	}

	switch format {
	case "html":
		fmt.Println(style.Info.Render("🔄 Generating UI..."))

		jsonData, err := marshalDatasetsForOutput(datasets)
		if err != nil {
			shared.ExitWithError("Failed to marshal dataSet data: %v", err)
		}

		needsHeatmapChunk := datasetsNeedCorrelation(datasets)
		needs3D := datasetsNeed3D(datasets)
		htmlContent := template.GenerateUI(jsonData, chartTypeNames(datasets[0].Settings), needs3D, needsHeatmapChunk, template.VizbHTMLTemplate)
		if _, err := f.WriteString(htmlContent); err != nil {
			shared.ExitWithError("Failed to write output file: %v", err)
		}

		fmt.Println(style.Success.Render("🎉 Generated HTML UI successfully!"))

	case "json":
		fmt.Println(style.Info.Render("🔄 Generating JSON..."))
		bytes, err := marshalDatasetsForOutput(datasets)
		if err != nil {
			shared.ExitWithError("Error marshaling dataSet data", err)
		}
		if _, err := f.Write(bytes); err != nil {
			shared.ExitWithError("Failed to write output file", err)
		}
		fmt.Println(style.Success.Render("🎉 Generated JSON successfully!"))
	}
}

func marshalDatasetsForOutput(datasets []*shared.Dataset) ([]byte, error) {
	if len(datasets) == 1 {
		return json.Marshal(datasets[0])
	}
	slice := make([]shared.Dataset, len(datasets))
	for i, ds := range datasets {
		slice[i] = *ds
	}
	return json.Marshal(slice)
}

func datasetsNeed3D(datasets []*shared.Dataset) bool {
	for _, ds := range datasets {
		if shared.DatasetNeeds3D(ds) {
			return true
		}
	}
	return false
}

func datasetsNeedCorrelation(datasets []*shared.Dataset) bool {
	for _, ds := range datasets {
		if settingsNeedCorrelation(ds.Settings) {
			return true
		}
	}
	return false
}
