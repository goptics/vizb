package csv

import (
	"encoding/csv"
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
	parser.Parsers["csv"] = ParseCSV
}

// parseFinite parses a trimmed cell as a float, rejecting NaN/Inf (which would
// later crash json.Marshal). Returns the value and whether it is a finite number.
func parseFinite(s string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, false
	}
	return v, true
}

// ParseCSV turns a generic CSV into benchmark data. Each numeric column
// becomes a chart series (column name = Stat.Type); non-numeric columns are
// ignored unless named in --group/-g, whose values are joined with the
// separators from --group-pattern/-p and routed through the grouping machinery
// (-p/-r) for name/xAxis/yAxis placement.
func ParseCSV(filename string, cfg parser.Config) []shared.DataPoint {
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

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1 // allow ragged rows

	rows, err := reader.ReadAll()
	if err != nil {
		shared.ExitWithError("Error reading CSV", err)
	}

	if len(rows) < 2 { // need header + at least one data row
		return nil
	}

	headers := normalizeHeaders(rows[0])
	dataRows := rows[1:]

	// Auto-group: when no grouping is configured, infer the category axis from
	// the data so `vizb data.csv` produces a usable chart without -g/-p/-r/-x.
	if parser.AutoGroupApplies(cfg) {
		cols, pattern, ok := parser.AutoGroupColumns(headers, dataRows, cfg.WantsBothXY)
		if ok {
			cfg.Group = cols
			cfg.GroupPattern = pattern
			var err error
			cfg, err = parser.ResolveGroupConfig(cfg)
			if err != nil {
				shared.ExitWithError(err.Error(), nil)
			}
			if err := parser.ValidateTabularGroupAlignment(cfg); err != nil {
				shared.ExitWithError(err.Error(), nil)
			}
			parser.LogAutoGroup(cols, cfg.WantsBothXY)
		} else if valCols, vok := parser.AutoValueColumns(headers, dataRows); vok {
			cfg.Axes = make([]parser.ColumnSpec, len(valCols))
			for i, name := range valCols {
				cfg.Axes[i] = parser.ColumnSpec{Source: name}
			}
			parser.LogAutoValue(valCols)
		}
	}

	if len(cfg.Axes) > 0 {
		return parseCSVValueMode(headers, dataRows, cfg)
	}

	groupIdx, groupSet := resolveGroupColumns(headers, parser.EffectiveGroupColumns(cfg))

	var chartCols []int
	var colLabels map[int]string
	if len(cfg.Select) > 0 {
		var err error
		chartCols, colLabels, err = resolveExplicitChartColumns(headers, cfg, dataRows)
		if err != nil {
			shared.ExitWithError(err.Error(), nil)
		}
	} else {
		chartCols = chartColumns(headers, groupSet, dataRows)
	}
	if len(chartCols) == 0 {
		shared.ExitWithError("no numeric columns found in CSV", nil)
	}

	var results []shared.DataPoint

	for _, row := range dataRows {
		groupValues := groupColumnValues(row, groupIdx)

		var name, xAxis, yAxis, zAxis string
		if len(groupValues) > 0 {
			label := parser.TabularFilterLabel(groupValues, cfg)
			if !parser.ShouldIncludeBenchmark(label, cfg) {
				continue
			}

			group, gerr := parser.GroupTabularRow(groupValues, cfg)
			if gerr != nil {
				shared.ExitWithError("Error parsing CSV group name", gerr)
			}

			name, xAxis, yAxis, zAxis = group["name"], group["xAxis"], group["yAxis"], group["zAxis"]
		}

		var stats []shared.Stat
		for _, c := range chartCols {
			if c >= len(row) {
				continue // ragged row: missing cell
			}

			v, ok := parseFinite(row[c])
			if !ok {
				continue // non-numeric/empty cell: gap
			}

			label := headers[c]
			if colLabels != nil {
				if l, ok := colLabels[c]; ok {
					label = l
				}
			}
			stats = append(stats, shared.Stat{
				Type:  utils.CreateStatType(label, cfg.NumberUnit, ""),
				Value: shared.F64(utils.FormatNumber(v, cfg.NumberUnit)),
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

// parseCSVValueMode implements value mode: each named numeric column
// becomes a coordinate on x, y[, z] (by --axes order); each row becomes a raw
// point with no stat series. A missing or fully non-numeric axis column is
// fatal; an individual row missing a finite cell for any axis is skipped.
func parseCSVValueMode(headers []string, dataRows [][]string, cfg parser.Config) []shared.DataPoint {
	idx := make([]int, len(cfg.Axes))
	for i, spec := range cfg.Axes {
		col := -1
		for h, name := range headers {
			if name == spec.Source {
				col = h
				break
			}
		}
		if col == -1 {
			shared.ExitWithError(fmt.Sprintf("--axes column '%s' not found; available: %v", spec.Source, nonEmpty(headers)), nil)
		}

		numeric := false
		for _, row := range dataRows {
			if col < len(row) {
				if _, ok := parseFinite(row[col]); ok {
					numeric = true
					break
				}
			}
		}
		if !numeric {
			shared.ExitWithError(fmt.Sprintf("--axes column '%s' is not numeric", spec.Source), nil)
		}
		idx[i] = col
	}

	var results []shared.DataPoint
	for _, row := range dataRows {
		dp := shared.DataPoint{Stats: []shared.Stat{}}
		dst := []*string{&dp.XAxis, &dp.YAxis, &dp.ZAxis}
		complete := true
		for i, col := range idx {
			if col >= len(row) {
				complete = false
				break
			}
			v, ok := parseFinite(row[col])
			if !ok {
				complete = false
				break
			}
			*dst[i] = strconv.FormatFloat(utils.FormatNumber(v, cfg.NumberUnit), 'g', -1, 64)
		}
		if !complete {
			continue
		}
		results = append(results, dp)
	}
	return results
}

// normalizeHeaders strips a leading UTF-8 BOM, trims whitespace, and de-duplicates
// repeated names by suffixing (sells, "sells (2)") so same-named columns don't
// collapse into one lost series. Empty header cells stay empty (unusable column).
func normalizeHeaders(raw []string) []string {
	headers := make([]string, len(raw))
	seen := map[string]int{}

	for i, h := range raw {
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

// resolveGroupColumns maps each non-empty --group name to its header index
// (preserving flag order). A missing column is fatal and lists available headers.
func resolveGroupColumns(headers []string, group []string) ([]int, map[int]bool) {
	var idx []int
	set := map[int]bool{}

	for _, name := range group {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		found := -1
		for i, h := range headers {
			if h == name {
				found = i
				break
			}
		}

		if found == -1 {
			shared.ExitWithError(fmt.Sprintf("group column '%s' not found; available: %v", name, nonEmpty(headers)), nil)
		}

		idx = append(idx, found)
		set[found] = true
	}

	return idx, set
}

func resolveExplicitChartColumns(headers []string, cfg parser.Config, dataRows [][]string) ([]int, map[int]string, error) {
	numeric := make(map[int]bool, len(headers))
	for _, c := range chartColumns(headers, map[int]bool{}, dataRows) {
		numeric[c] = true
	}

	indices := make([]int, 0, len(cfg.Select))
	labels := make(map[int]string, len(cfg.Select))

	for _, spec := range cfg.Select {
		idx := -1
		for i, h := range headers {
			if h == spec.Source {
				idx = i
				break
			}
		}
		if idx == -1 {
			return nil, nil, fmt.Errorf("column '%s' not found in --select; available: %v", spec.Source, nonEmpty(headers))
		}
		if !numeric[idx] {
			return nil, nil, fmt.Errorf("column '%s' in --select is not numeric", spec.Source)
		}

		label := spec.Source
		if spec.Label != "" {
			label = spec.Label
		}
		indices = append(indices, idx)
		labels[idx] = label
	}
	return indices, labels, nil
}

// chartColumns returns, in column order, indices of columns that have a non-empty
// header, are not group columns, and contain at least one finite numeric cell.
func chartColumns(headers []string, groupSet map[int]bool, dataRows [][]string) []int {
	var cols []int

	for i, h := range headers {
		if h == "" || groupSet[i] {
			continue
		}

		for _, row := range dataRows {
			if i >= len(row) {
				continue
			}

			if _, ok := parseFinite(row[i]); ok {
				cols = append(cols, i)
				break
			}
		}
	}

	return cols
}

func groupColumnValues(row []string, groupIdx []int) []string {
	parts := make([]string, 0, len(groupIdx))
	for _, idx := range groupIdx {
		val := ""
		if idx < len(row) {
			val = strings.TrimSpace(row[idx])
		}
		parts = append(parts, val)
	}
	return parts
}

// nonEmpty returns the non-empty header names (for error messages).
func nonEmpty(headers []string) []string {
	var out []string
	for _, h := range headers {
		if h != "" {
			out = append(out, h)
		}
	}
	return out
}
