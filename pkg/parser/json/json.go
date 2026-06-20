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
	if len(cfg.Group) > 0 && cfg.GroupRegex == "" {
		var err error
		cfg, err = parser.ResolveGroupConfig(cfg)
		if err != nil {
			shared.ExitWithError(err.Error(), nil)
		}
		if err := parser.ValidateTabularGroupAlignment(cfg); err != nil {
			shared.ExitWithError(err.Error(), nil)
		}
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

	groupKeys, groupSet := resolveGroupKeys(colOrder, seenCol, parser.EffectiveGroupColumns(cfg))

	var chartCols []string
	var fieldLabels map[string]string
	if len(cfg.Cols) > 0 {
		var err error
		chartCols, fieldLabels, err = resolveExplicitChartFields(colOrder, cfg, rows)
		if err != nil {
			shared.ExitWithError(err.Error(), nil)
		}
	} else {
		chartCols = chartColumns(colOrder, groupSet, rows)
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
			if fieldLabels != nil {
				if l, ok := fieldLabels[k]; ok {
					label = l
				}
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
	seen := map[string]bool{}
	for _, k := range colOrder {
		seen[k] = true
	}

	fields := make([]string, 0, len(cfg.Cols))
	labels := make(map[string]string, len(cfg.Cols))

	for _, spec := range cfg.Cols {
		if !seen[spec.Source] {
			return nil, nil, fmt.Errorf("column '%s' not found in --cols; available: %v", spec.Source, colOrder)
		}

		numeric := false
		for _, row := range rows {
			if v, ok := row[spec.Source]; ok {
				if _, isNum := leafNumber(v); isNum {
					numeric = true
					break
				}
			}
		}
		if !numeric {
			return nil, nil, fmt.Errorf("column '%s' in --cols is not numeric", spec.Source)
		}

		fields = append(fields, spec.Source)
		labels[spec.Source] = spec.DisplayLabel()
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
