package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
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

// ParseJSON turns a JSON array of objects or arrays into benchmark data. Each
// numeric field becomes a chart series (field name = Stat.Type); non-numeric
// fields are ignored unless named in --group/-g, whose values are joined with
// the separators from --group-pattern/-p and routed through the grouping
// machinery (-p/-r). Nested objects are
// flattened to dotted keys; array-valued fields inside objects are skipped.
func ParseJSON(input io.Reader, cfg parser.Config) ([]shared.DataPoint, parser.Config, *shared.Meta, error) {
	points, effectiveCfg, err := parseReader(input, cfg, !cfg.QuietAutoDetect)
	return points, effectiveCfg, nil, err
}

func parseReader(input io.Reader, cfg parser.Config, logAuto bool) ([]shared.DataPoint, parser.Config, error) {
	var err error
	cfg, err = parser.FinalizeGroupConfig(cfg)
	if err != nil {
		return nil, cfg, err
	}

	dec := json.NewDecoder(input)

	// Top-level must be a JSON array; otherwise unsupported (object form is
	// consumed earlier by convertToBenchmark).
	tok, err := dec.Token()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, cfg, nil
		}
		return nil, cfg, fmt.Errorf("read JSON: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '[' {
		return nil, cfg, nil
	}

	rows, colOrder, seenCol, err := decodeTopLevelRows(dec)
	if err != nil {
		return nil, cfg, fmt.Errorf("read JSON: %w", err)
	}

	if len(rows) == 0 {
		return nil, cfg, nil
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
		if logAuto {
			cfg, err = parser.AutoDetectTabularConfig(cfg, autoHeaders, stringRows)
		} else {
			cfg, err = parser.AutoDetectTabularConfigQuiet(cfg, autoHeaders, stringRows)
		}
		if err != nil {
			return nil, cfg, err
		}
	}

	if (len(cfg.SelectViews) > 0 && !parser.IsExplicitGrouping(cfg)) || len(cfg.Axes) > 0 {
		cfg.Mode = parser.ResolveMode(cfg)
		selectAxis := len(cfg.SelectViews) > 0 && !parser.IsExplicitGrouping(cfg)
		flag := parser.AxisColumnLabel(selectAxis)
		readers := make([]parser.RowReader, len(rows))
		for i, row := range rows {
			readers[i] = jsonRowReader{row: row, seenCol: seenCol, colOrder: colOrder, flag: flag}
		}
		results, err := parser.DispatchSelectMode(readers, &cfg, jsonKindFn(rows, seenCol, colOrder, flag))
		if err != nil {
			return nil, cfg, err
		}
		return results, cfg, nil
	}

	groupKeys, groupSet, err := resolveGroupKeys(colOrder, seenCol, parser.EffectiveGroupColumns(cfg))
	if err != nil {
		return nil, cfg, err
	}

	chartCols := chartColumns(colOrder, groupSet, rows)
	var fieldLabels map[string]string
	if len(cfg.Select) > 0 {
		chartCols, fieldLabels, err = resolveExplicitChartFields(colOrder, cfg, rows)
		if err != nil {
			return nil, cfg, err
		}
	}
	if len(chartCols) == 0 {
		return nil, cfg, fmt.Errorf("no numeric fields found in JSON")
	}

	var results []shared.DataPoint

	for _, row := range rows {
		groupValues := groupFieldValues(row, groupKeys)

		var name, xAxis, yAxis, zAxis string
		if len(groupValues) > 0 {
			label := parser.TabularFilterLabel(groupValues, cfg)
			include, err := parser.ShouldIncludeBenchmark(label, cfg)
			if err != nil {
				return nil, cfg, err
			}
			if !include {
				continue
			}

			group, gerr := parser.GroupTabularRow(groupValues, cfg)
			if gerr != nil {
				return nil, cfg, fmt.Errorf("parse JSON group name: %w", gerr)
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

	return results, cfg, nil
}

func decodeTopLevelRows(dec *json.Decoder) ([]map[string]any, []string, map[string]bool, error) {
	seenCol := map[string]bool{}

	if !dec.More() {
		return nil, nil, seenCol, nil
	}

	tok, err := dec.Token()
	if err != nil {
		return nil, nil, nil, err
	}

	if d, ok := tok.(json.Delim); ok {
		switch d {
		case '{':
			return decodeObjectRows(dec, seenCol)
		case '[':
			return decodeMatrixRows(dec, seenCol)
		}
	}

	rows, colOrder, err := decodeRemainingObjectRows(dec, nil, seenCol)
	return rows, colOrder, seenCol, err
}

func decodeObjectRows(dec *json.Decoder, seenCol map[string]bool) ([]map[string]any, []string, map[string]bool, error) {
	leaves, err := decodeObjectBody(dec, "")
	if err != nil {
		return nil, nil, nil, err
	}

	rows, colOrder, err := decodeRemainingObjectRows(dec, leaves, seenCol)
	return rows, colOrder, seenCol, err
}

func decodeRemainingObjectRows(dec *json.Decoder, first []leaf, seenCol map[string]bool) ([]map[string]any, []string, error) {
	var rows []map[string]any
	var colOrder []string

	if first != nil {
		rows = append(rows, leavesToRow(first, seenCol, &colOrder))
	}

	for dec.More() {
		leaves, derr := decodeElement(dec)
		if derr != nil {
			return nil, nil, derr
		}
		rows = append(rows, leavesToRow(leaves, seenCol, &colOrder))
	}

	return rows, colOrder, nil
}

func leavesToRow(leaves []leaf, seenCol map[string]bool, colOrder *[]string) map[string]any {
	row := make(map[string]any, len(leaves))
	for _, lf := range leaves {
		row[lf.key] = lf.val // last wins on duplicate key
		if !seenCol[lf.key] {
			seenCol[lf.key] = true
			*colOrder = append(*colOrder, lf.key)
		}
	}
	return row
}

type matrixCell struct {
	val      any
	ok       bool
	isString bool
}

func decodeMatrixRows(dec *json.Decoder, seenCol map[string]bool) ([]map[string]any, []string, map[string]bool, error) {
	first, err := decodeMatrixRowBody(dec)
	if err != nil {
		return nil, nil, nil, err
	}

	if matrixHeaderRow(first) {
		headers := normalizeMatrixHeaders(first)
		rows, headers, err := decodeMatrixDataRows(dec, headers, nil, false)
		if err != nil {
			return nil, nil, nil, err
		}
		colOrder := nonEmptyMatrixHeaders(headers)
		for _, h := range colOrder {
			seenCol[h] = true
		}
		return rows, colOrder, seenCol, nil
	}

	headers := syntheticMatrixHeaders(len(first))
	rows, headers, err := decodeMatrixDataRows(dec, headers, first, true)
	if err != nil {
		return nil, nil, nil, err
	}
	colOrder := nonEmptyMatrixHeaders(headers)
	for _, h := range colOrder {
		seenCol[h] = true
	}
	return rows, colOrder, seenCol, nil
}

func decodeMatrixDataRows(dec *json.Decoder, headers []string, first []matrixCell, extendSynthetic bool) ([]map[string]any, []string, error) {
	var rows []map[string]any
	if first != nil {
		if extendSynthetic {
			headers = ensureSyntheticMatrixHeaders(headers, len(first))
		}
		rows = append(rows, matrixCellsToRow(first, headers))
	}

	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, nil, err
		}

		d, ok := tok.(json.Delim)
		if !ok || d != '[' {
			if ok && d == '{' {
				if err := skipContainerBody(dec); err != nil {
					return nil, nil, err
				}
			}
			continue
		}

		cells, err := decodeMatrixRowBody(dec)
		if err != nil {
			return nil, nil, err
		}
		if extendSynthetic {
			headers = ensureSyntheticMatrixHeaders(headers, len(cells))
		}
		rows = append(rows, matrixCellsToRow(cells, headers))
	}

	return rows, headers, nil
}

func decodeMatrixRowBody(dec *json.Decoder) ([]matrixCell, error) {
	var cells []matrixCell

	for dec.More() {
		cell, err := decodeMatrixCell(dec)
		if err != nil {
			return nil, err
		}
		cells = append(cells, cell)
	}

	if _, err := dec.Token(); err != nil { // consume ']'
		return nil, err
	}

	return cells, nil
}

func decodeMatrixCell(dec *json.Decoder) (matrixCell, error) {
	tok, err := dec.Token()
	if err != nil {
		return matrixCell{}, err
	}

	switch v := tok.(type) {
	case json.Delim:
		switch v {
		case '[', '{':
			return matrixCell{}, skipContainerBody(dec)
		}
		return matrixCell{}, nil
	case float64:
		return matrixCell{val: v, ok: true}, nil
	case string:
		return matrixCell{val: v, ok: true, isString: true}, nil
	default: // bool, nil
		return matrixCell{}, nil
	}
}

func matrixHeaderRow(cells []matrixCell) bool {
	if len(cells) == 0 {
		return false
	}
	for _, c := range cells {
		if !c.ok || !c.isString {
			return false
		}
	}
	return true
}

func normalizeMatrixHeaders(cells []matrixCell) []string {
	headers := make([]string, len(cells))
	seen := map[string]int{}

	for i, c := range cells {
		h, _ := c.val.(string)
		if i == 0 {
			h = strings.TrimPrefix(h, "\ufeff")
		}
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}

		seen[h]++
		if seen[h] > 1 {
			h = fmt.Sprintf("%s (%d)", h, seen[h])
		}
		headers[i] = h
	}

	return headers
}

func syntheticMatrixHeaders(n int) []string {
	return ensureSyntheticMatrixHeaders(nil, n)
}

func ensureSyntheticMatrixHeaders(headers []string, n int) []string {
	for len(headers) < n {
		headers = append(headers, syntheticMatrixHeader(len(headers)))
	}
	return headers
}

func syntheticMatrixHeader(i int) string {
	switch i {
	case 0:
		return "x"
	case 1:
		return "y"
	case 2:
		return "z"
	case 3:
		return "metric"
	default:
		return fmt.Sprintf("col%d", i+1)
	}
}

func nonEmptyMatrixHeaders(headers []string) []string {
	var out []string
	for _, h := range headers {
		if h != "" {
			out = append(out, h)
		}
	}
	return out
}

func matrixCellsToRow(cells []matrixCell, headers []string) map[string]any {
	row := make(map[string]any)
	for i, c := range cells {
		if i >= len(headers) || headers[i] == "" || !c.ok {
			continue
		}
		row[headers[i]] = c.val
	}
	return row
}

// jsonRowReader adapts one JSON row to the parser.RowReader interface.
type jsonRowReader struct {
	row      map[string]any
	seenCol  map[string]bool
	colOrder []string
	flag     string
}

func (r jsonRowReader) Cell(source string) (string, bool) {
	v, ok := r.row[source]
	if !ok {
		return "", false
	}
	return strings.TrimSpace(stringify(v)), true
}

func (r jsonRowReader) Numeric(source string) (float64, bool) {
	v, ok := r.row[source]
	if !ok {
		return 0, false
	}
	return leafNumber(v)
}

func (r jsonRowReader) AvailableColumns() []string { return r.colOrder }
func (r jsonRowReader) FlagLabel() string          { return r.flag }

func jsonKindFn(rows []map[string]any, seenCol map[string]bool, colOrder []string, flag string) parser.AxisColumnKind {
	return func(source, axisKey string) (string, error) {
		if !seenCol[source] {
			return "", fmt.Errorf("%s field %q not found; available: %v", flag, source, colOrder)
		}
		anyNumeric, allNumeric, sawCell := false, true, false
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
			return "", fmt.Errorf("%s field %q has no data", flag, source)
		}
		if axisKey == "x" {
			if allNumeric {
				return "value", nil
			}
			return "category", nil
		}
		if !anyNumeric {
			return "", fmt.Errorf("%s field %q is not numeric", flag, source)
		}
		return "value", nil
	}
}

// resolveGroupKeys maps each non-empty --group name to a known field (preserving
// flag order). A missing name is fatal and lists available fields.
func resolveGroupKeys(colOrder []string, seenCol map[string]bool, group []string) ([]string, map[string]bool, error) {
	var keys []string
	set := map[string]bool{}

	for _, name := range group {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		if !seenCol[name] {
			return nil, nil, fmt.Errorf("group field %q not found; available: %v", name, colOrder)
		}

		keys = append(keys, name)
		set[name] = true
	}

	return keys, set, nil
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
