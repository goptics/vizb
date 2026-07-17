// Package core contains request-scoped Vizb application operations. It has no
// Cobra, terminal, output-file, or process-exit dependencies.
package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	internalcharts "github.com/goptics/vizb/internal/charts"
	barchart "github.com/goptics/vizb/internal/charts/bar"
	linechart "github.com/goptics/vizb/internal/charts/line"
	scatterchart "github.com/goptics/vizb/internal/charts/scatter"
	"github.com/goptics/vizb/pkg/parser"
	_ "github.com/goptics/vizb/pkg/parser/csv"
	jsonparser "github.com/goptics/vizb/pkg/parser/json"
	"github.com/goptics/vizb/pkg/template"
	"github.com/goptics/vizb/shared"
)

// Metadata is the caller-supplied Dataset metadata for Convert.
type Metadata struct {
	ID          string
	Name        string
	Theme       string
	Description string
	Tag         string
	System      *shared.Meta
	Timestamp   string
}

// AssembleInput contains parsed, request-local data and its resolved options.
// It lets file-based adapters and HTTP handlers share Dataset construction.
type AssembleInput struct {
	Points   []shared.DataPoint
	Parser   string
	Config   parser.Config
	Metadata Metadata
	Charts   []internalcharts.ChartConfig
}

// ConvertInput is all of the state required to convert inline input. Chart
// configs must already be materialised by the request adapter.
type ConvertInput struct {
	Input    []byte
	Parser   string
	Config   parser.Config
	Metadata Metadata
	Charts   []internalcharts.ChartConfig
}

// ConvertResult carries the Dataset plus non-fatal chart applicability notices.
type ConvertResult struct {
	Dataset  *shared.Dataset
	Warnings []string
}

// Convert parses inline tabular data, aggregates it, applies chart rules, and
// builds a Dataset. Every dependency receives request-local values and every
// failure is returned to the caller.
func Convert(in ConvertInput) (ConvertResult, error) {
	if len(in.Input) == 0 {
		return ConvertResult{}, fmt.Errorf("input is empty")
	}
	if len(in.Charts) == 0 {
		return ConvertResult{}, fmt.Errorf("at least one chart configuration is required")
	}

	key := strings.TrimSpace(in.Parser)
	if key == "" {
		key = "auto"
	}
	if key == "auto" {
		key = detectInlineParser(in.Input)
	}
	if key != "csv" && key != "json" {
		return ConvertResult{}, fmt.Errorf("parser %q does not support inline conversion", key)
	}
	if in.Config.Filter != "" {
		if _, err := regexp.Compile(in.Config.Filter); err != nil {
			return ConvertResult{}, fmt.Errorf("invalid filter regex: %w", err)
		}
	}

	cfg := in.Config
	if cfg.GroupPattern == "" {
		cfg.GroupPattern = "x"
	}
	cfg.ChartTypes = append(slices.Clone(cfg.ChartTypes), chartTypes(in.Charts)...)
	if parser.NoExplicitGrouping(cfg) && !parser.HasSelect(cfg) {
		cfg.AutoGroup = true
	}

	data := in.Input
	if cfg.JSONPath != "" {
		if key != "json" {
			return ConvertResult{}, fmt.Errorf("json path is only supported by the json parser")
		}
		var err error
		data, err = jsonparser.SelectBytes(data, cfg.JSONPath)
		if err != nil {
			return ConvertResult{}, err
		}
	}

	parseFn, err := parser.GetReaderParser(key)
	if err != nil {
		return ConvertResult{}, err
	}
	points, effectiveCfg, err := parseFn(bytes.NewReader(data), cfg)
	if err != nil {
		return ConvertResult{}, err
	}
	if len(effectiveCfg.Group) == 0 {
		points = shared.CollapseDataPointsByKey(points)
	} else {
		points = shared.AggregateDataPoints(points)
	}
	if len(points) == 0 {
		return ConvertResult{}, fmt.Errorf("no dataset found")
	}

	dataset := Assemble(AssembleInput{Points: points, Parser: key, Config: effectiveCfg, Metadata: in.Metadata, Charts: in.Charts})
	for _, chart := range in.Charts {
		if swap := chart.SwapString(); swap != "" {
			if err := shared.ValidateSwap(swap, dataset.Axes); err != nil {
				return ConvertResult{}, err
			}
		}
	}
	ruleAxes := make([]internalcharts.AxisInfo, 0, len(dataset.Axes))
	for _, axis := range dataset.Axes {
		ruleAxes = append(ruleAxes, internalcharts.AxisInfo{Key: axis.Key, Type: axis.Type})
	}
	warnings, err := internalcharts.ApplyRules(internalcharts.RuleContext{Axes: ruleAxes}, dataset.Settings)
	if err != nil {
		return ConvertResult{}, err
	}
	return ConvertResult{Dataset: dataset, Warnings: warnings}, nil
}

// Merge combines complete request datasets atomically. It intentionally accepts
// datasets, not paths or directories.
func Merge(datasets []shared.Dataset, dimension shared.Dimension) ([]shared.Dataset, error) {
	if len(datasets) == 0 {
		return nil, fmt.Errorf("at least one dataset is required to merge")
	}
	switch dimension {
	case "", "name", shared.DimensionName:
		dimension = shared.DimensionName
	case shared.DimensionXAxis, shared.DimensionYAxis, shared.DimensionZAxis:
	default:
		return nil, fmt.Errorf("invalid tag axis %q; expected name, x, y, or z", dimension)
	}
	return shared.MergeDatasets(datasets, dimension), nil
}

// GenerateUI serializes one or more datasets into a self-contained HTML page.
// When charts is provided it is both embedded as the UI selection and used to
// prune the renderer chunks.
func GenerateUI(datasets []shared.Dataset, charts []string) (string, error) {
	if len(datasets) == 0 {
		return "", fmt.Errorf("at least one dataset is required")
	}
	copyDatasets := slices.Clone(datasets)
	if len(charts) == 0 {
		charts = unionCharts(copyDatasets)
	} else {
		for i := range copyDatasets {
			copyDatasets[i].Settings = filterSettings(copyDatasets[i].Settings, charts)
		}
	}
	jsonData, err := marshalDatasets(copyDatasets)
	if err != nil {
		return "", fmt.Errorf("marshal datasets: %w", err)
	}
	needs3D := slices.ContainsFunc(copyDatasets, func(ds shared.Dataset) bool { return shared.DatasetNeeds3D(&ds) })
	needsHeatmap := slices.ContainsFunc(copyDatasets, func(ds shared.Dataset) bool {
		return slices.ContainsFunc(ds.Settings, shared.ChartConfigNeedsCorrelation)
	})
	return template.GenerateUIE(jsonData, charts, needs3D, needsHeatmap, template.VizbHTMLTemplate)
}

func detectInlineParser(data []byte) string {
	if bytes.HasPrefix(bytes.TrimSpace(data), []byte("[")) {
		return "json"
	}
	return "csv"
}

// Assemble creates a Dataset from already parsed points. It does not inspect
// global process metadata; callers supply every optional metadata field.
func Assemble(in AssembleInput) *shared.Dataset {
	points, parserKey, cfg, meta, charts := in.Points, in.Parser, in.Config, in.Metadata, in.Charts
	var view []parser.ColumnSpec
	if cfg.Mode.IsSelectAxis() && len(cfg.SelectViews) == 1 {
		view = cfg.SelectViews[0].Columns
	}
	viewName := ""
	if len(view) > 0 && cfg.Mode.IsSelectAxis() && !cfg.Mode.IsMultiStat() {
		viewName = parser.SelectViewDatasetName(view, 0)
	}
	var axes []shared.Axis
	if cfg.Mode.IsMultiStat() {
		axes = parser.MultiSelectStatAxes(cfg.SelectViews)
	} else if len(view) > 0 {
		axes = parser.DatasetAxesForSelectView(view, points)
		autoEnableValueMode3D(charts, axes, valueModeHasMetric(cfg, points))
	} else if cfg.Mode.IsSelectAxis() {
		axes = parser.DatasetAxesForSelectView(cfg.SelectViews[0].Columns, points)
		autoEnableValueMode3D(charts, axes, valueModeHasMetric(cfg, points))
	} else {
		axes = parser.GroupAxes(cfg)
		if len(cfg.Axes) > 0 {
			if parser.IsMixedAxes(cfg) {
				axes = parser.MixedAxes(cfg)
			} else {
				axes = parser.ValueAxes(cfg)
			}
			autoEnableValueMode3D(charts, axes, valueModeHasMetric(cfg, points))
		}
	}
	if cfg.ColAxis != "" {
		axes = shared.EnsureAxis(axes, shared.Dimension(cfg.ColAxis))
	}
	axes = appendMetricAxis(axes, cfg, points)
	name := meta.Name
	if name == "" && viewName != "" {
		name = viewName
	}
	timestamp := meta.Timestamp
	if timestamp == "" {
		timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	ds := &shared.Dataset{
		ID:           strings.TrimSpace(meta.ID),
		Name:         name,
		Theme:        meta.Theme,
		Description:  meta.Description,
		Tag:          meta.Tag,
		Timestamp:    timestamp,
		Meta:         meta.System,
		Axes:         axes,
		Settings:     charts,
		Data:         points,
		PreserveRows: (parserKey == "csv" || parserKey == "json") && len(cfg.Group) == 0,
	}
	return ds
}

func valueModeHasMetric(cfg parser.Config, points []shared.DataPoint) bool {
	if cfg.MetricColumn != "" {
		return true
	}
	return slices.ContainsFunc(points, func(point shared.DataPoint) bool { return point.Metric != "" })
}

func appendMetricAxis(axes []shared.Axis, cfg parser.Config, points []shared.DataPoint) []shared.Axis {
	if !valueModeHasMetric(cfg, points) {
		return axes
	}
	for _, axis := range axes {
		if axis.Key == "metric" {
			return axes
		}
	}
	label := cfg.MetricColumn
	if label == "" {
		label = "value"
	}
	return append(axes, shared.Axis{Key: "metric", Label: label, Type: "value"})
}

func autoEnableValueMode3D(configs []internalcharts.ChartConfig, axes []shared.Axis, visualMap bool) {
	keys := map[string]bool{}
	for _, axis := range axes {
		if axis.Key != "metric" && axis.Type != "value" {
			return
		}
		if axis.Key != "metric" {
			keys[axis.Key] = true
		}
	}
	if !keys["x"] || !keys["y"] || !keys["z"] {
		return
	}
	v := true
	for _, config := range configs {
		switch chart := config.(type) {
		case *barchart.Config:
			chart.ThreeD = &v
			if visualMap {
				chart.ThreeDVisualMap = &v
			}
		case *linechart.Config:
			chart.ThreeD = &v
			if visualMap {
				chart.ThreeDVisualMap = &v
			}
		case *scatterchart.Config:
			chart.ThreeD = &v
			if visualMap {
				chart.ThreeDVisualMap = &v
			}
		}
	}
}

func chartTypes(settings []internalcharts.ChartConfig) []string {
	types := make([]string, 0, len(settings))
	for _, setting := range settings {
		types = append(types, setting.ChartType())
	}
	return types
}

func unionCharts(datasets []shared.Dataset) []string {
	var charts []string
	for _, dataset := range datasets {
		for _, setting := range dataset.Settings {
			if !slices.Contains(charts, setting.ChartType()) {
				charts = append(charts, setting.ChartType())
			}
		}
	}
	return charts
}

func filterSettings(settings []internalcharts.ChartConfig, allowed []string) []internalcharts.ChartConfig {
	filtered := make([]internalcharts.ChartConfig, 0, len(settings))
	for _, setting := range settings {
		if slices.Contains(allowed, setting.ChartType()) {
			filtered = append(filtered, setting)
		}
	}
	return filtered
}

func marshalDatasets(datasets []shared.Dataset) ([]byte, error) {
	return json.Marshal(datasets)
}
