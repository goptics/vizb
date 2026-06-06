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
// ignored unless named in --group/-g, whose values are '/'-joined and routed
// through the existing grouping machinery (-p/-r) for name/xAxis/yAxis placement.
func ParseCSV(filename string) []shared.DataPoint {
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

	groupIdx, groupSet := resolveGroupColumns(headers)

	chartCols := chartColumns(headers, groupSet, dataRows)
	if len(chartCols) == 0 {
		shared.ExitWithError("no numeric columns found in CSV", nil)
	}

	var results []shared.DataPoint

	for _, row := range dataRows {
		label := buildLabel(row, groupIdx)

		var name, xAxis, yAxis string
		if label != "" {
			if !parser.ShouldIncludeBenchmark(label) {
				continue
			}

			group, gerr := parser.GroupBenchmarkName(label)
			if gerr != nil {
				shared.ExitWithError("Error parsing CSV group name", gerr)
			}

			name, xAxis, yAxis = group["name"], group["xAxis"], group["yAxis"]
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

			stats = append(stats, shared.Stat{
				Type:  utils.CreateStatType(headers[c], shared.FlagState.NumberUnit, ""),
				Value: utils.FormatNumber(v, shared.FlagState.NumberUnit),
			})
		}

		if len(stats) == 0 {
			continue
		}

		results = append(results, shared.DataPoint{
			Name:  name,
			XAxis: xAxis,
			YAxis: yAxis,
			Stats: stats,
		})
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
func resolveGroupColumns(headers []string) ([]int, map[int]bool) {
	var idx []int
	set := map[int]bool{}

	for _, name := range shared.FlagState.Group {
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

// buildLabel joins the trimmed group-column values for a row with '/'. Returns
// "" when no group columns are configured.
func buildLabel(row []string, groupIdx []int) string {
	if len(groupIdx) == 0 {
		return ""
	}

	parts := make([]string, 0, len(groupIdx))
	for _, idx := range groupIdx {
		val := ""
		if idx < len(row) {
			val = strings.TrimSpace(row[idx])
		}
		parts = append(parts, val)
	}

	return strings.Join(parts, "/")
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
