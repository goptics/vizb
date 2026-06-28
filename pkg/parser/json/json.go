package json

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
)

func init() {
	parser.Parsers["json"] = ParseJSON
}

// leaf is a single flattened scalar field of a row. val is float64 or string.
type leaf struct {
	key string
	val any
}

// parseFinite parses a trimmed string as a float, rejecting NaN/Inf.
func parseFinite(s string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, false
	}
	return v, true
}

// leafNumber reports whether v is a finite number — either a JSON number or a
// numeric string (mirrors the CSV any-one-parses rule).
func leafNumber(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		if math.IsNaN(t) || math.IsInf(t, 0) {
			return 0, false
		}
		return t, true
	case string:
		return parseFinite(t)
	}
	return 0, false
}

// stringify renders a leaf value for use in a group label.
func stringify(v any) string {
	switch t := v.(type) {
	case float64:
		return strconv.FormatFloat(t, 'g', -1, 64)
	case string:
		return t
	}
	return ""
}

// ParseJSON turns a JSON array of objects into benchmark data. Each
// numeric field becomes a chart series (field name = Stat.Type); non-numeric
// fields are ignored unless named in --group/-g, whose values are joined with
// the separators from --group-pattern/-p and routed through the grouping
// machinery (-p/-r). Nested objects are
// flattened to dotted keys; array-valued fields are skipped.
func ParseJSON(filename string, cfg parser.Config) []shared.DataPoint {
	var err error
	cfg, err = parser.FinalizeGroupConfig(cfg)
	if err != nil {
		shared.ExitWithError(err.Error(), nil)
	}

	f, err := os.Open(filename)
	if err != nil {
		shared.ExitWithError("Error opening file", err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)

	// Top-level must be a JSON array; otherwise unsupported (object form is
	// consumed earlier by convertToBenchmark).
	tok, err := dec.Token()
	if err != nil {
		return nil
	}
	if d, ok := tok.(json.Delim); !ok || d != '[' {
		return nil
	}

	var rows []map[string]any
	var colOrder []string
	seenCol := map[string]bool{}

	for dec.More() {
		leaves, derr := decodeElement(dec)
		if derr != nil {
			shared.ExitWithError("Error reading JSON", derr)
		}

		row := make(map[string]any, len(leaves))
		for _, lf := range leaves {
			row[lf.key] = lf.val // last wins on duplicate key
			if !seenCol[lf.key] {
				seenCol[lf.key] = true
				colOrder = append(colOrder, lf.key)
			}
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		return nil
	}

	// Auto-group: when no grouping is configured, infer the category axis from
	// the data so `vizb data.json` produces a usable chart without -g/-p/-r/-x.
	if !parser.HasSelect(cfg) && parser.AutoGroupApplies(cfg) {
		autoHeaders := parser.FilterHeadersForAutoDetect(colOrder, cfg.Select)
		stringRows := make([][]string, len(rows))
		for i, row := range rows {
			cells := make([]string, len(autoHeaders))
			for j, k := range autoHeaders {
				if v, ok := row[k]; ok {
					cells[j] = stringify(v)
				}
			}
			stringRows[i] = cells
		}
		cfg, err = parser.AutoDetectTabularConfig(cfg, autoHeaders, stringRows)
		if err != nil {
			shared.ExitWithError(err.Error(), nil)
		}
	}

	if parser.IsMultiSelectStatMode(cfg) {
		return parseJSONSelectStatMode(rows, seenCol, cfg)
	}

	if parser.IsSelectAxisMode(cfg) {
		axesCfg := parser.SelectViewAxesCfg(cfg)
		flag := parser.AxisColumnLabel(true)
		if err := parser.ResolveAxesTypes(&axesCfg, jsonAxisColumnKind(rows, seenCol, colOrder, flag)); err != nil {
			shared.ExitWithError(err.Error(), nil)
		}
		if parser.IsMixedMode(axesCfg) {
			return parseJSONMixedMode(rows, colOrder, seenCol, axesCfg, flag)
		}
		return parseJSONValueMode(rows, colOrder, seenCol, axesCfg, flag)
	}

	if len(cfg.Axes) > 0 {
		flag := parser.AxisColumnLabel(false)
		if err := parser.ResolveAxesTypes(&cfg, jsonAxisColumnKind(rows, seenCol, colOrder, flag)); err != nil {
			shared.ExitWithError(err.Error(), nil)
		}
		if parser.IsMixedMode(cfg) {
			return parseJSONMixedMode(rows, colOrder, seenCol, cfg, flag)
		}
		return parseJSONValueMode(rows, colOrder, seenCol, cfg, flag)
	}

	groupKeys, groupSet := resolveGroupKeys(colOrder, seenCol, parser.EffectiveGroupColumns(cfg))

	chartCols := chartColumns(colOrder, groupSet, rows)
	var fieldLabels map[string]string
	if len(cfg.Select) > 0 {
		chartCols, fieldLabels, err = resolveExplicitChartFields(colOrder, cfg, rows)
		if err != nil {
			shared.ExitWithError(err.Error(), nil)
		}
	}
	if len(chartCols) == 0 {
		shared.ExitWithError("no numeric fields found in JSON", nil)
	}

	var results []shared.DataPoint

	for _, row := range rows {
		groupValues := groupFieldValues(row, groupKeys)

		var name, xAxis, yAxis, zAxis string
		if len(groupValues) > 0 {
			label := parser.TabularFilterLabel(groupValues, cfg)
			if !parser.ShouldIncludeBenchmark(label, cfg) {
				continue
			}

			group, gerr := parser.GroupTabularRow(groupValues, cfg)
			if gerr != nil {
				shared.ExitWithError("Error parsing JSON group name", gerr)
			}

			name, xAxis, yAxis, zAxis = group["name"], group["xAxis"], group["yAxis"], group["zAxis"]
		}

		var stats []shared.Stat
		for _, k := range chartCols {
			v, ok := row[k]
			if !ok {
				continue
			}

			num, ok := leafNumber(v)
			if !ok {
				continue
			}

			label := k
			if l, ok := fieldLabels[k]; ok {
				label = l
			}
			stats = append(stats, shared.Stat{
				Type:  utils.CreateStatType(label, cfg.NumberUnit, ""),
				Value: shared.F64(utils.FormatNumber(num, cfg.NumberUnit)),
			})
		}

		if len(stats) == 0 {
			continue
		}

		results = append(results, shared.DataPoint{
			Name:  name,
			XAxis: xAxis,
			YAxis: yAxis,
			ZAxis: zAxis,
			Stats: stats,
		})
	}

	return results
}

// jsonAxisColumnKind classifies axis fields for mixed vs value routing.
func jsonAxisColumnKind(rows []map[string]any, seenCol map[string]bool, colOrder []string, flagLabel string) parser.AxisColumnKind {
	return func(source, axisKey string) (string, error) {
		if !seenCol[source] {
			return "", fmt.Errorf("%s field %q not found; available: %v", flagLabel, source, colOrder)
		}
		anyNumeric := false
		allNumeric := true
		sawCell := false
		for _, row := range rows {
			v, ok := row[source]
			if !ok {
				continue
			}
			if strings.TrimSpace(stringify(v)) == "" {
				continue
			}
			sawCell = true
			if _, ok := leafNumber(v); ok {
				anyNumeric = true
			} else {
				allNumeric = false
			}
		}
		if !sawCell {
			return "", fmt.Errorf("%s field %q has no data", flagLabel, source)
		}
		if axisKey == "x" {
			if allNumeric {
				return "value", nil
			}
			return "category", nil
		}
		if !anyNumeric {
			return "", fmt.Errorf("%s field %q is not numeric", flagLabel, source)
		}
		return "value", nil
	}
}

// parseJSONMixedMode maps one categorical field to x and numeric fields to y[,z].
func parseJSONMixedMode(rows []map[string]any, colOrder []string, seenCol map[string]bool, cfg parser.Config, flagLabel string) []shared.DataPoint {
	type slot struct {
		key  string
		kind string
	}
	slots := make(map[string]slot, len(cfg.Axes))
	for _, spec := range cfg.Axes {
		if !seenCol[spec.Source] {
			shared.ExitWithError(fmt.Sprintf("%s field '%s' not found; available: %v", flagLabel, spec.Source, colOrder), nil)
		}
		if spec.AxisType == "value" {
			numeric := false
			for _, row := range rows {
				v, ok := row[spec.Source]
				if !ok {
					continue
				}
				if _, ok := leafNumber(v); ok {
					numeric = true
					break
				}
			}
			if !numeric {
				shared.ExitWithError(fmt.Sprintf("%s field '%s' is not numeric", flagLabel, spec.Source), nil)
			}
		}
		slots[spec.AxisKey] = slot{key: spec.Source, kind: spec.AxisType}
	}

	var results []shared.DataPoint
	for _, row := range rows {
		var dp shared.DataPoint
		complete := true
		for axisKey, sl := range slots {
			v, ok := row[sl.key]
			if !ok {
				complete = false
				break
			}
			if sl.kind == "category" {
				s := strings.TrimSpace(stringify(v))
				if s == "" {
					complete = false
					break
				}
				switch axisKey {
				case "x":
					dp.XAxis = s
				case "y":
					dp.YAxis = s
				case "z":
					dp.ZAxis = s
				}
				continue
			}
			num, ok := leafNumber(v)
			if !ok {
				complete = false
				break
			}
			formatted := strconv.FormatFloat(utils.FormatNumber(num, cfg.NumberUnit), 'g', -1, 64)
			switch axisKey {
			case "x":
				dp.XAxis = formatted
			case "y":
				dp.YAxis = formatted
			case "z":
				dp.ZAxis = formatted
			}
		}
		if !complete {
			continue
		}
		results = append(results, dp)
	}
	return results
}

// parseJSONValueMode implements value mode for JSON: each named numeric
// field becomes a coordinate on x, y[, z] (by axis order); each row becomes a
// raw point with no stat series. A missing or fully non-numeric axis field is
// fatal; a row missing a finite value for any axis is skipped.
func parseJSONValueMode(rows []map[string]any, colOrder []string, seenCol map[string]bool, cfg parser.Config, flagLabel string) []shared.DataPoint {
	keys := make([]string, len(cfg.Axes))
	for i, spec := range cfg.Axes {
		if !seenCol[spec.Source] {
			shared.ExitWithError(fmt.Sprintf("%s field '%s' not found; available: %v", flagLabel, spec.Source, colOrder), nil)
		}

		numeric := false
		for _, row := range rows {
			v, ok := row[spec.Source]
			if !ok {
				continue
			}
			if _, ok := leafNumber(v); ok {
				numeric = true
				break
			}
		}
		if !numeric {
			shared.ExitWithError(fmt.Sprintf("%s field '%s' is not numeric", flagLabel, spec.Source), nil)
		}
		keys[i] = spec.Source
	}

	metricKey := ""
	if cfg.MetricColumn != "" {
		if !seenCol[cfg.MetricColumn] {
			shared.ExitWithError(fmt.Sprintf("metric field '%s' not found; available: %v", cfg.MetricColumn, colOrder), nil)
		}
		hasNumeric := false
		for _, row := range rows {
			v, ok := row[cfg.MetricColumn]
			if !ok {
				continue
			}
			if _, ok := leafNumber(v); ok {
				hasNumeric = true
				break
			}
		}
		if !hasNumeric {
			shared.ExitWithError(fmt.Sprintf("metric field '%s' is not numeric", cfg.MetricColumn), nil)
		}
		metricKey = cfg.MetricColumn
	}

	var results []shared.DataPoint
	for _, row := range rows {
		var dp shared.DataPoint
		dst := []*string{&dp.XAxis, &dp.YAxis, &dp.ZAxis}
		complete := true
		for i, k := range keys {
			v, ok := row[k]
			if !ok {
				complete = false
				break
			}
			num, ok := leafNumber(v)
			if !ok {
				complete = false
				break
			}
			*dst[i] = strconv.FormatFloat(utils.FormatNumber(num, cfg.NumberUnit), 'g', -1, 64)
		}
		if !complete {
			continue
		}
		if metricKey != "" {
			v, ok := row[metricKey]
			if !ok {
				continue
			}
			mv, ok := leafNumber(v)
			if !ok {
				continue
			}
			dp.Metric = strconv.FormatFloat(utils.FormatNumber(mv, cfg.NumberUnit), 'g', -1, 64)
		}
		results = append(results, dp)
	}
	return results
}

// parseJSONSelectStatMode parses repeatable solo --select into one dataset. When
// every flag shares the same dimension column, each input row becomes one point
// with multiple stats; otherwise each (row × view) stays a separate point.
func parseJSONSelectStatMode(rows []map[string]any, seenCol map[string]bool, cfg parser.Config) []shared.DataPoint {
	flag := parser.AxisColumnLabel(true)
	merge := parser.MultiSelectSharedDim(cfg.SelectViews)

	for _, view := range cfg.SelectViews {
		for _, spec := range view.Columns {
			if !seenCol[spec.Source] {
				shared.ExitWithError(fmt.Sprintf("%s field %q not found", flag, spec.Source), nil)
			}
		}
	}

	var results []shared.DataPoint
	for _, row := range rows {
		parser.AppendMultiSelectStatPoint(&results, cfg.SelectViews, cfg.NumberUnit, merge, func(view parser.SelectView) (parser.MultiSelectRowStat, bool) {
			dim, metric := view.Columns[0], view.Columns[1]
			dimVal := strings.TrimSpace(stringify(row[dim.Source]))
			if dimVal == "" {
				return parser.MultiSelectRowStat{}, false
			}
			v, ok := leafNumber(row[metric.Source])
			if !ok {
				return parser.MultiSelectRowStat{}, false
			}
			return parser.MultiSelectRowStat{DimVal: dimVal, Value: v}, true
		})
	}
	if len(results) == 0 {
		shared.ExitWithError("No dataSet data found", nil)
	}
	return results
}

func readJSONArray(filename string) (rows []map[string]any, colOrder []string, seenCol map[string]bool) {
	f, err := os.Open(filename)
	if err != nil {
		shared.ExitWithError("Error opening file", err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	tok, err := dec.Token()
	if err != nil {
		shared.ExitWithError("No dataSet data found", nil)
	}
	if d, ok := tok.(json.Delim); !ok || d != '[' {
		shared.ExitWithError("No dataSet data found", nil)
	}

	seenCol = map[string]bool{}
	for dec.More() {
		leaves, derr := decodeElement(dec)
		if derr != nil {
			shared.ExitWithError("Error reading JSON", derr)
		}

		row := make(map[string]any, len(leaves))
		for _, lf := range leaves {
			row[lf.key] = lf.val
			if !seenCol[lf.key] {
				seenCol[lf.key] = true
				colOrder = append(colOrder, lf.key)
			}
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		shared.ExitWithError("No dataSet data found", nil)
	}
	return rows, colOrder, seenCol
}

// resolveGroupKeys maps each non-empty --group name to a known field (preserving
// flag order). A missing name is fatal and lists available fields.
func resolveGroupKeys(colOrder []string, seenCol map[string]bool, group []string) ([]string, map[string]bool) {
	var keys []string
	set := map[string]bool{}

	for _, name := range group {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		if !seenCol[name] {
			shared.ExitWithError(fmt.Sprintf("group field '%s' not found; available: %v", name, colOrder), nil)
		}

		keys = append(keys, name)
		set[name] = true
	}

	return keys, set
}

func resolveExplicitChartFields(colOrder []string, cfg parser.Config, rows []map[string]any) ([]string, map[string]string, error) {
	seen := make(map[string]bool, len(colOrder))
	for _, k := range colOrder {
		seen[k] = true
	}

	numeric := make(map[string]bool, len(colOrder))
	for _, k := range chartColumns(colOrder, map[string]bool{}, rows) {
		numeric[k] = true
	}

	fields := make([]string, 0, len(cfg.Select))
	labels := make(map[string]string, len(cfg.Select))

	for _, spec := range cfg.Select {
		if !seen[spec.Source] {
			return nil, nil, fmt.Errorf("column '%s' not found in --select; available: %v", spec.Source, colOrder)
		}
		if !numeric[spec.Source] {
			return nil, nil, fmt.Errorf("column '%s' in --select is not numeric", spec.Source)
		}

		label := spec.Source
		if spec.Label != "" {
			label = spec.Label
		}
		fields = append(fields, spec.Source)
		labels[spec.Source] = label
	}
	return fields, labels, nil
}

// chartColumns returns, in first-seen order, fields that have at least one
// numeric value across the rows and are not group fields.
func chartColumns(colOrder []string, groupSet map[string]bool, rows []map[string]any) []string {
	var cols []string

	for _, k := range colOrder {
		if groupSet[k] {
			continue
		}

		for _, row := range rows {
			if v, ok := row[k]; ok {
				if _, isNum := leafNumber(v); isNum {
					cols = append(cols, k)
					break
				}
			}
		}
	}

	return cols
}

// buildLabel joins the stringified group-field values using the configured separators.
func groupFieldValues(row map[string]any, groupKeys []string) []string {
	parts := make([]string, 0, len(groupKeys))
	for _, k := range groupKeys {
		parts = append(parts, stringify(row[k]))
	}
	return parts
}

// decodeElement decodes one array element. Objects are flattened to leaves;
// non-object elements (scalars, arrays) are skipped.
func decodeElement(dec *json.Decoder) ([]leaf, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}

	if d, ok := tok.(json.Delim); ok {
		switch d {
		case '{':
			return decodeObjectBody(dec, "")
		case '[':
			return nil, skipContainerBody(dec)
		}
	}

	// scalar element: token already consumed
	return nil, nil
}

// decodeObjectBody walks an object whose opening '{' has already been read,
// emitting one leaf per scalar field (recursing into nested objects with a
// dotted prefix, skipping arrays).
func decodeObjectBody(dec *json.Decoder, prefix string) ([]leaf, error) {
	var out []leaf

	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}

		key, _ := keyTok.(string)
		full := prefix + key

		leaves, err := decodeValue(dec, full)
		if err != nil {
			return nil, err
		}

		out = append(out, leaves...)
	}

	if _, err := dec.Token(); err != nil { // consume '}'
		return nil, err
	}

	return out, nil
}

// decodeValue reads a single value for the given key.
func decodeValue(dec *json.Decoder, key string) ([]leaf, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}

	switch v := tok.(type) {
	case json.Delim:
		switch v {
		case '{':
			return decodeObjectBody(dec, key+".")
		case '[':
			return nil, skipContainerBody(dec)
		}
		return nil, nil
	case float64:
		return []leaf{{key: key, val: v}}, nil
	case string:
		return []leaf{{key: key, val: v}}, nil
	default: // bool, nil
		return nil, nil
	}
}

// skipContainerBody consumes tokens until the container whose opening delimiter
// was already read is balanced (handles arbitrary nesting).
func skipContainerBody(dec *json.Decoder) error {
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return err
		}

		if d, ok := tok.(json.Delim); ok {
			switch d {
			case '[', '{':
				depth++
			case ']', '}':
				depth--
			}
		}
	}

	return nil
}
