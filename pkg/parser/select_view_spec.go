package parser

import (
	"fmt"
	"strings"

	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
)

// ParseSelectViewFlag parses one solo --select value (e.g. "region,latency" or
// "x:region,y:latency,z:sales") into 2–3 column specs with x/y/z axis keys.
// Reuses tokenizeSelectFlag from select_spec.go for {label}/quoting.
// Multi-stat views may end with " (Chart Tab Name)" to rename stat.type.
func ParseSelectViewFlag(raw string) (SelectView, error) {
	cols, typeLabel, err := splitTrailingParenLabel(raw)
	if err != nil {
		return SelectView{}, err
	}
	specs, err := parseSelectViewColumns(cols)
	if err != nil {
		return SelectView{}, err
	}
	return SelectView{Columns: specs, TypeLabel: typeLabel}, nil
}

func parseSelectViewColumns(raw string) ([]ColumnSpec, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("--select requires 2 or 3 columns (x,y[,z]); got 0")
	}

	tokens, err := tokenizeSelectFlag(raw)
	if err != nil {
		return nil, err
	}
	if n := len(tokens); n < 2 || n > 3 {
		return nil, fmt.Errorf("--select requires 2 or 3 columns (x,y[,z]); got %d", n)
	}

	seenCol := map[string]bool{}
	seenKey := map[string]bool{}
	specs := make([]ColumnSpec, 0, len(tokens))
	explicitCount := 0

	for i, tok := range tokens {
		spec, key, isExplicit, err := parseAxisToken(tok)
		if err != nil {
			return nil, err
		}
		if spec.Source == "" {
			return nil, fmt.Errorf("empty column name in --select")
		}
		if seenCol[spec.Source] {
			return nil, fmt.Errorf("duplicate column '%s' in --select", spec.Source)
		}
		seenCol[spec.Source] = true

		if isExplicit {
			explicitCount++
			if seenKey[key] {
				return nil, fmt.Errorf("duplicate axis key '%s' in --select", key)
			}
			seenKey[key] = true
			spec.AxisKey = key
		} else {
			keys := []string{"x", "y", "z"}
			spec.AxisKey = keys[i]
		}
		specs = append(specs, spec)
	}

	if explicitCount > 0 && explicitCount != len(tokens) {
		return nil, fmt.Errorf("--select: use explicit x:/y:/z: syntax for every column, or omit prefixes for all")
	}
	if explicitCount > 0 {
		if err := validateExplicitSelectAxisKeys(specs); err != nil {
			return nil, err
		}
	}

	return specs, nil
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
	if key != "x" && key != "y" && key != "z" {
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

// IsSelectAxisMode reports solo --select axis mode: select views without explicit grouping.
func IsSelectAxisMode(cfg Config) bool {
	return len(cfg.SelectViews) > 0 && !IsExplicitGrouping(cfg)
}

// IsMultiSelectStatMode reports repeatable solo --select: each flag is an independent
// (dim, metric) pair merged into one dataset with stats[] chart separation.
func IsMultiSelectStatMode(cfg Config) bool {
	return IsSelectAxisMode(cfg) && len(cfg.SelectViews) > 1
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
