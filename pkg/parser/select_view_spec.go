package parser

import (
	"fmt"
	"strings"

	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
)

// ParseSelectViewFlag parses one solo --select value (e.g. "region,latency",
// "x:region,y:latency,z:sales", or "x,y,z,value" / "x:x,y:y,z:z,metric:value")
// into 2–3 spatial axis specs plus an optional visualMap metric column.
// Reuses tokenizeSelectFlag from select_spec.go for {label}/quoting.
// Multi-stat views may end with " (Chart Tab Name)" to rename stat.type.
func ParseSelectViewFlag(raw string) (SelectView, error) {
	cols, typeLabel, err := splitTrailingParenLabel(raw)
	if err != nil {
		return SelectView{}, err
	}
	specs, metricSrc, metricLabel, err := parseSelectViewColumns(cols)
	if err != nil {
		return SelectView{}, err
	}
	return SelectView{
		Columns:      specs,
		TypeLabel:    typeLabel,
		MetricSource: metricSrc,
		MetricLabel:  metricLabel,
	}, nil
}

func parseSelectViewColumns(raw string) (specs []ColumnSpec, metricSrc, metricLabel string, err error) {
	if strings.TrimSpace(raw) == "" {
		return nil, "", "", fmt.Errorf("--select requires 2–4 columns (x,y[,z][,metric]); got 0")
	}

	tokens, err := tokenizeSelectFlag(raw)
	if err != nil {
		return nil, "", "", err
	}
	n := len(tokens)
	if n < 2 || n > 4 {
		return nil, "", "", fmt.Errorf("--select requires 2–4 columns (x,y[,z][,metric]); got %d", n)
	}

	seenCol := map[string]bool{}
	seenKey := map[string]bool{}
	explicitCount := 0
	type parsed struct {
		spec       ColumnSpec
		key        string
		isExplicit bool
	}
	items := make([]parsed, 0, n)

	for _, tok := range tokens {
		spec, key, isExplicit, perr := parseAxisToken(tok)
		if perr != nil {
			return nil, "", "", perr
		}
		if spec.Source == "" {
			return nil, "", "", fmt.Errorf("empty column name in --select")
		}
		if seenCol[spec.Source] {
			return nil, "", "", fmt.Errorf("duplicate column '%s' in --select", spec.Source)
		}
		seenCol[spec.Source] = true

		if isExplicit {
			explicitCount++
			if seenKey[key] {
				return nil, "", "", fmt.Errorf("duplicate axis key '%s' in --select", key)
			}
			seenKey[key] = true
			spec.AxisKey = key
		}
		items = append(items, parsed{spec: spec, key: key, isExplicit: isExplicit})
	}

	if explicitCount > 0 && explicitCount != n {
		return nil, "", "", fmt.Errorf("--select: use explicit x:/y:/z:/metric: syntax for every column, or omit prefixes for all")
	}

	// Explicit metric:col peels off the visualMap metric; remaining are spatial axes.
	if explicitCount > 0 {
		axisItems := make([]parsed, 0, n)
		for _, it := range items {
			if it.key == "metric" {
				if metricSrc != "" {
					return nil, "", "", fmt.Errorf("duplicate metric column in --select")
				}
				metricSrc = it.spec.Source
				metricLabel = it.spec.Label
				continue
			}
			axisItems = append(axisItems, it)
		}
		if metricSrc != "" && len(axisItems) < 2 {
			return nil, "", "", fmt.Errorf("--select metric requires at least x and y axes")
		}
		if metricSrc != "" && len(axisItems) > 3 {
			return nil, "", "", fmt.Errorf("--select metric allows at most 3 spatial axes (x,y[,z])")
		}
		specs = make([]ColumnSpec, len(axisItems))
		for i, it := range axisItems {
			specs[i] = it.spec
		}
		if err := validateExplicitSelectAxisKeys(specs); err != nil {
			return nil, "", "", err
		}
		return specs, metricSrc, metricLabel, nil
	}

	// Positional: 2–3 → x,y[,z]; 4 → x,y,z + metric (last column).
	if n == 4 {
		metricSrc = items[3].spec.Source
		metricLabel = items[3].spec.Label
		items = items[:3]
	}
	keys := []string{"x", "y", "z"}
	specs = make([]ColumnSpec, len(items))
	for i, it := range items {
		spec := it.spec
		spec.AxisKey = keys[i]
		specs[i] = spec
	}
	return specs, metricSrc, metricLabel, nil
}

// splitTrailingParenLabel strips a view-level " (Title)" suffix used for
// multi-stat chart-tab names. Quoted column names may contain parentheses.
func splitTrailingParenLabel(raw string) (cols string, typeLabel string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.HasSuffix(raw, ")") {
		return raw, "", nil
	}

	splitAt := -1
	inQuote := false
	for i := 0; i < len(raw); i++ {
		switch raw[i] {
		case '"':
			inQuote = !inQuote
		case ' ':
			if !inQuote && i+1 < len(raw) && raw[i+1] == '(' {
				splitAt = i
			}
		}
	}
	if splitAt < 0 {
		return raw, "", nil
	}

	inner := strings.TrimSpace(raw[splitAt+2 : len(raw)-1])
	if inner == "" {
		return "", "", fmt.Errorf("empty chart name () in --select")
	}
	return strings.TrimSpace(raw[:splitAt]), inner, nil
}

func parseAxisToken(tok string) (ColumnSpec, string, bool, error) {
	colon := strings.Index(tok, ":")
	if colon <= 0 {
		spec, err := parseColumnToken(tok)
		return spec, "", false, err
	}

	key := strings.TrimSpace(tok[:colon])
	if key != "x" && key != "y" && key != "z" && key != "metric" {
		spec, err := parseColumnToken(tok)
		return spec, "", false, err
	}

	spec, err := parseColumnToken(strings.TrimSpace(tok[colon+1:]))
	if err != nil {
		return spec, key, true, err
	}
	return spec, key, true, nil
}

func validateExplicitSelectAxisKeys(specs []ColumnSpec) error {
	has := map[string]bool{}
	for _, s := range specs {
		has[s.AxisKey] = true
	}
	if !has["x"] {
		return fmt.Errorf("--select explicit syntax requires x:column (e.g. x:region,y:latency)")
	}
	if !has["y"] {
		return fmt.Errorf("--select explicit syntax requires y:column")
	}
	if len(specs) == 3 && !has["z"] {
		return fmt.Errorf("--select with 3 columns requires z:column in explicit syntax")
	}
	// metric: is peeled before this runs; only spatial keys remain.
	return nil
}

// HasSelect reports whether the user supplied any --select configuration.
func HasSelect(cfg Config) bool {
	return len(cfg.Select) > 0 || len(cfg.SelectViews) > 0
}

// IsExplicitGrouping reports whether the user supplied explicit grouping flags.
func IsExplicitGrouping(cfg Config) bool {
	return len(cfg.Group) > 0 ||
		cfg.GroupRegex != "" ||
		(cfg.GroupPattern != "" && cfg.GroupPattern != "x")
}

// ValidateMultiSelectStatViews ensures each repeatable --select has exactly two
// columns (dimension, metric). Three-column coordinate views use a single --select.
func ValidateMultiSelectStatViews(views []SelectView) error {
	for i, view := range views {
		if len(view.Columns) != 2 {
			return fmt.Errorf(
				"repeatable --select requires 2 columns (dim,metric) per flag (view %d has %d); use one --select for 3-axis coordinate views",
				i+1, len(view.Columns),
			)
		}
	}
	return nil
}

// selectDimName returns the display/source name for the dimension column in a view.
func selectDimName(spec ColumnSpec) string {
	if spec.Label != "" {
		return spec.Label
	}
	return spec.Source
}

// SelectStatType names the stat series for one solo --select view in multi-stat mode.
// Default: "metric by dim". Trailing (Title) overrides; {label} on the metric column
// is a legacy alias when () is absent.
func SelectStatType(view SelectView) string {
	if view.TypeLabel != "" {
		return view.TypeLabel
	}
	if len(view.Columns) < 2 {
		return ""
	}
	if view.Columns[1].Label != "" {
		return view.Columns[1].Label
	}
	return view.Columns[1].Source + " by " + selectDimName(view.Columns[0])
}

// MultiSelectSharedDim reports whether every repeatable --select uses the same
// dimension column, so per-row stats can share one xAxis value.
func MultiSelectSharedDim(views []SelectView) bool {
	if len(views) == 0 {
		return false
	}
	src := views[0].Columns[0].Source
	for _, view := range views[1:] {
		if len(view.Columns) == 0 || view.Columns[0].Source != src {
			return false
		}
	}
	return true
}

// MultiSelectRowStat is one metric read from an input row for multi-stat mode.
type MultiSelectRowStat struct {
	DimVal string
	Value  float64
}

// AppendMultiSelectStatPoint appends one DataPoint built from all views on a single
// input row when views share a dimension column; otherwise appends one point per view.
func AppendMultiSelectStatPoint(
	results *[]shared.DataPoint,
	views []SelectView,
	numberUnit string,
	merge bool,
	read func(view SelectView) (MultiSelectRowStat, bool),
) {
	if merge {
		var dp shared.DataPoint
		for _, view := range views {
			row, ok := read(view)
			if !ok {
				continue
			}
			if dp.XAxis == "" {
				dp.XAxis = row.DimVal
			}
			dp.Stats = append(dp.Stats, shared.Stat{
				Type:  utils.CreateStatType(SelectStatType(view), numberUnit, ""),
				Value: shared.F64(utils.FormatNumber(row.Value, numberUnit)),
			})
		}
		if len(dp.Stats) > 0 {
			*results = append(*results, dp)
		}
		return
	}

	for _, view := range views {
		row, ok := read(view)
		if !ok {
			continue
		}
		*results = append(*results, shared.DataPoint{
			XAxis: row.DimVal,
			Stats: []shared.Stat{{
				Type:  utils.CreateStatType(SelectStatType(view), numberUnit, ""),
				Value: shared.F64(utils.FormatNumber(row.Value, numberUnit)),
			}},
		})
	}
}

// MultiSelectStatAxes returns dataset axes for multi-stat solo --select (category x only).
// The x-axis label comes from the dimension column's {label} (or source name) on the
// first --select flag; repeatable flags are expected to share the same dimension.
func MultiSelectStatAxes(views []SelectView) []shared.Axis {
	label := ""
	if len(views) > 0 && len(views[0].Columns) > 0 {
		dim := views[0].Columns[0]
		label = dim.Label
		if label == "" {
			label = dim.Source
		}
	}
	return []shared.Axis{{Key: "x", Label: label, Type: ""}}
}

// SelectViewData holds parsed rows for one solo --select view.
type SelectViewData struct {
	View SelectView
	Data []shared.DataPoint
	Name string
}

// SelectViewDatasetName auto-names a dataset from its column sources (e.g.
// "region × latency"). Labels take precedence over source names.
func SelectViewDatasetName(view []ColumnSpec, index int) string {
	if len(view) == 0 {
		return fmt.Sprintf("View %d", index+1)
	}
	parts := make([]string, len(view))
	for i, spec := range view {
		if spec.Label != "" {
			parts[i] = spec.Label
			continue
		}
		parts[i] = spec.Source
	}
	return strings.Join(parts, " × ")
}

// AxisColumnLabel returns the flag name for axis-column error messages.
func AxisColumnLabel(selectMode bool) string {
	if selectMode {
		return "--select"
	}
	return "--axes"
}
