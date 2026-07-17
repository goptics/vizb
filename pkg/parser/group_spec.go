package parser

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/goptics/vizb/pkg/style"
)

// GroupSpec describes how --group column names are laid out and joined into labels.
type GroupSpec struct {
	Columns    []string
	Separators []string
	Structured bool
}

// splitBySeparators splits s using separators in order (same algorithm as benchmark name splitting).
func splitBySeparators(s string, separators []string) []string {
	if len(separators) == 0 {
		return []string{s}
	}

	parts := []string{s}
	for _, sep := range separators {
		var newParts []string
		for _, part := range parts {
			newParts = append(newParts, strings.SplitN(part, sep, 2)...)
		}
		parts = newParts
	}

	var result []string
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}
	return result
}

func isStructuredGroup(group []string) bool {
	for _, g := range group {
		if strings.ContainsAny(g, "/_") {
			return true
		}
	}
	return false
}

func reconstructGroupSpec(group []string) string {
	return strings.Join(group, ",")
}

// extractSpecSeparators returns separator strings between column-name tokens in a structured spec.
func extractSpecSeparators(spec string) []string {
	var seps []string
	inName := false

	for _, r := range spec {
		if isColumnNameRune(r) {
			inName = true
			continue
		}
		if inName {
			seps = append(seps, string(r))
			inName = false
		}
	}
	return seps
}

func isColumnNameRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.' || r == '-'
}

func formatGroupDisplay(group []string, structured bool) string {
	if len(group) == 1 {
		return group[0]
	}
	if structured {
		return reconstructGroupSpec(group)
	}
	return strings.Join(group, ",")
}

func errGroupPatternSeparatorMismatch(groupDisplay, pattern string, expectedSeps, actualSeps []string) error {
	return fmt.Errorf(
		"--group %q and --group-pattern %q separators do not match (expected %q, got %q)",
		groupDisplay, pattern,
		formatSeparators(expectedSeps), formatSeparators(actualSeps),
	)
}

// parseGroupSpec resolves column names and separators from the raw --group flag values
// and the separators declared in --group-pattern.
func parseGroupSpec(group []string, pattern string, patternSeps []string) (GroupSpec, error) {
	if len(group) == 0 {
		return GroupSpec{}, nil
	}

	if isStructuredGroup(group) {
		spec := reconstructGroupSpec(group)
		specSeps := extractSpecSeparators(spec)
		columns := splitBySeparators(spec, specSeps)
		if len(columns) == 0 {
			return GroupSpec{}, fmt.Errorf("invalid --group spec %q: no columns found", spec)
		}
		return GroupSpec{Columns: columns, Separators: specSeps, Structured: true}, nil
	}

	// Single -g value with internal separators matching -p (e.g. -g "name category region" -p "x n y").
	if len(group) == 1 && len(patternSeps) > 0 {
		columns := splitBySeparators(group[0], patternSeps)
		expected := len(patternSeps) + 1
		if len(columns) == expected {
			return GroupSpec{Columns: columns, Separators: patternSeps, Structured: false}, nil
		}
		if len(columns) == 1 && strings.TrimSpace(columns[0]) == strings.TrimSpace(group[0]) {
			return GroupSpec{Columns: columns, Separators: nil, Structured: false}, nil
		}
		return GroupSpec{}, errGroupPatternSeparatorMismatch(
			group[0], pattern,
			extractSpecSeparators(group[0]), patternSeps,
		)
	}

	var columns []string
	for _, g := range group {
		if g = strings.TrimSpace(g); g != "" {
			columns = append(columns, g)
		}
	}
	return GroupSpec{Columns: columns, Separators: nil, Structured: false}, nil
}

func separatorsMatch(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// isNumericCell reports whether a single cell string is a finite number.
// Mirrors the csv parser's parseFinite rule so AutoGroupColumns classifies
// columns the same way chartColumns does.
func isNumericCell(s string) bool {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return err == nil && !math.IsNaN(v) && !math.IsInf(v, 0)
}

// AutoGroupColumns infers one group column for the csv/json parsers when the
// user supplies no explicit grouping. Only non-numeric (categorical) columns
// are considered candidates; numeric columns are completely ignored. If no
// categorical column exists, auto-grouping is silently skipped (ok=false).
// Candidates are ranked by descending distinct-value count (leftmost
// tie-break); the top one is returned alongside pattern "x". ok is false when
// no inference is possible (single column, zero categorical columns, or no
// chart column would remain).
func AutoGroupColumns(headers []string, rows [][]string) (cols []string, pattern string, ok bool) {
	if len(headers) < 2 {
		return nil, "", false
	}

	type col struct {
		idx      int
		distinct int
	}

	// Gather distinct values per column; classify as categorical when any
	// non-numeric cell is seen.
	distinct := make([]map[string]struct{}, len(headers))
	anyNonNumeric := make([]bool, len(headers))
	for i := range headers {
		distinct[i] = make(map[string]struct{})
	}
	for _, row := range rows {
		for i := range headers {
			if i >= len(row) {
				continue
			}
			cell := strings.TrimSpace(row[i])
			if cell == "" {
				continue
			}
			distinct[i][cell] = struct{}{}
			if !isNumericCell(cell) {
				anyNonNumeric[i] = true
			}
		}
	}

	// Candidate pool: only non-numeric (categorical) columns.
	candidates := make([]col, 0, len(headers))
	for i, h := range headers {
		if h == "" {
			continue
		}
		if !anyNonNumeric[i] {
			continue // numeric columns are ignored
		}
		candidates = append(candidates, col{
			idx:      i,
			distinct: len(distinct[i]),
		})
	}
	if len(candidates) == 0 {
		return nil, "", false
	}

	// Rank by descending distinct-value count, leftmost tie-break.
	sort.SliceStable(candidates, func(a, b int) bool {
		if candidates[a].distinct != candidates[b].distinct {
			return candidates[a].distinct > candidates[b].distinct
		}
		return candidates[a].idx < candidates[b].idx
	})

	// The chosen axis column must leave at least one other column to act as
	// a chart (Stat) series; otherwise auto-grouping would consume the only
	// numeric column.
	if len(headers)-1 < 1 {
		return nil, "", false
	}

	return []string{headers[candidates[0].idx]}, "x", true
}

// NoExplicitGrouping reports whether the user supplied no explicit grouping
// configuration (no --group, no --group-regex, default --group-pattern "x",
// no --select, no auto-value axes). The pipeline uses it to decide whether to
// opt the csv/json parsers into auto-grouping.
func NoExplicitGrouping(cfg Config) bool {
	return !IsExplicitGrouping(cfg) &&
		!HasSelect(cfg) &&
		len(cfg.Axes) == 0
}

// FilterHeadersForAutoDetect returns headers restricted to select sources (file
// order) when --select is set; otherwise all headers.
func FilterHeadersForAutoDetect(headers []string, selectCols []ColumnSpec) []string {
	if len(selectCols) == 0 {
		return headers
	}
	selSet := make(map[string]bool, len(selectCols))
	for _, spec := range selectCols {
		selSet[spec.Source] = true
	}
	out := make([]string, 0, len(selectCols))
	for _, h := range headers {
		if selSet[h] {
			out = append(out, h)
		}
	}
	return out
}

// AutoGroupApplies reports whether auto-grouping should run inside the parser:
// the pipeline has opted this config in (cfg.AutoGroup) AND the user supplied
// no explicit grouping configuration. Any explicit grouping flag disables it.
func AutoGroupApplies(cfg Config) bool {
	return cfg.AutoGroup && NoExplicitGrouping(cfg)
}

// numericColumns returns purely numeric column names in file order. A column is
// numeric when all non-empty cells parse as finite floats.
func numericColumns(headers []string, rows [][]string) []string {
	var cols []string
	for i, h := range headers {
		if h == "" {
			continue
		}
		allNumeric := true
		for _, row := range rows {
			if i >= len(row) {
				continue
			}
			cell := strings.TrimSpace(row[i])
			if cell == "" {
				continue
			}
			if !isNumericCell(cell) {
				allNumeric = false
				break
			}
		}
		if allNumeric {
			cols = append(cols, h)
		}
	}
	return cols
}

// AutoValueColumns returns the first 2-3 purely numeric columns (in file
// order) for auto-value-mode value axes (x, y, z). A column is numeric when
// all non-empty cells parse as finite floats. Returns ok=false when <2 such
// columns exist — caller falls back to flat-series.
func AutoValueColumns(headers []string, rows [][]string) (cols []string, ok bool) {
	all := numericColumns(headers, rows)
	if len(all) < 2 {
		return nil, false
	}
	if len(all) > 3 {
		return all[:3], true
	}
	return all, true
}

// AutoValueEligible reports whether the given chart types are eligible for
// auto-value-mode. Currently only scatter, bar, and line are supported;
// pie/heatmap/radar fall back to flat series when all columns are numeric.
func AutoValueEligible(types []string) bool {
	for _, t := range types {
		if t == "scatter" || t == "bar" || t == "line" {
			return true
		}
	}
	return false
}

// LogAutoGroup prints the inferred group column. The csv/json parsers call it
// from their prelude so the inference is non-silent, mirroring the CLI's
// "Auto-detected parser" message.
func LogAutoGroup(cols []string) {
	if len(cols) == 0 {
		return
	}
	noun := "column"
	if len(cols) > 1 {
		noun = "columns"
	}
	fmt.Println(style.Info.Render(fmt.Sprintf("🧠 Auto-grouped by %s: %s", noun, strings.Join(cols, ", "))))
}

// LogAutoValue prints the inferred value axis columns (parallel to LogAutoGroup).
func LogAutoValue(cols []string, metricCol string) {
	if len(cols) == 0 {
		return
	}
	noun := "columns"
	if len(cols) == 2 {
		noun = "columns (2D pattern x-y)"
	} else if len(cols) == 3 {
		noun = "columns (3D pattern x-y-z)"
	}
	msg := fmt.Sprintf("🧠 Auto-valued by %s: %s", noun, strings.Join(cols, ", "))
	if metricCol != "" {
		msg += fmt.Sprintf(", metric: %s", metricCol)
	}
	fmt.Println(style.Info.Render(msg))
}

// FinalizeGroupConfig resolves and validates explicit --group config for tabular parsers.
func FinalizeGroupConfig(cfg Config) (Config, error) {
	if len(cfg.Group) == 0 || cfg.GroupRegex != "" {
		return cfg, nil
	}
	var err error
	cfg, err = ResolveGroupConfig(cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, ValidateTabularGroupAlignment(cfg)
}

// AutoDetectTabularConfig infers group columns or value axes when auto-group is enabled.
func AutoDetectTabularConfig(cfg Config, autoHeaders []string, rows [][]string) (Config, error) {
	return autoDetectTabularConfig(cfg, autoHeaders, rows, true)
}

// AutoDetectTabularConfigQuiet has the same inference behaviour as
// AutoDetectTabularConfig but does not emit CLI progress messages. Request
// handlers use it so concurrent responses never share terminal output.
func AutoDetectTabularConfigQuiet(cfg Config, autoHeaders []string, rows [][]string) (Config, error) {
	return autoDetectTabularConfig(cfg, autoHeaders, rows, false)
}

func autoDetectTabularConfig(cfg Config, autoHeaders []string, rows [][]string, log bool) (Config, error) {
	if !AutoGroupApplies(cfg) {
		return cfg, nil
	}

	cols, pattern, ok := AutoGroupColumns(autoHeaders, rows)
	if ok {
		cfg.Group = cols
		cfg.GroupPattern = pattern
		var err error
		cfg, err = FinalizeGroupConfig(cfg)
		if err != nil {
			return cfg, err
		}
		if log {
			LogAutoGroup(cols)
		}
		return cfg, nil
	}

	if !AutoValueEligible(cfg.ChartTypes) {
		return cfg, nil
	}
	allNumeric := numericColumns(autoHeaders, rows)
	if len(allNumeric) < 2 {
		return cfg, nil
	}
	valCols := allNumeric
	if len(valCols) > 3 {
		valCols = valCols[:3]
	}
	cfg.Axes = make([]ColumnSpec, len(valCols))
	for i, name := range valCols {
		cfg.Axes[i] = ColumnSpec{Source: name}
	}
	if len(allNumeric) >= 4 {
		cfg.MetricColumn = allNumeric[3]
	}
	if log {
		LogAutoValue(valCols, cfg.MetricColumn)
	}
	return cfg, nil
}

// ResolveGroupConfig fills GroupColumns and LabelSeparators from --group and --group-pattern.
func ResolveGroupConfig(cfg Config) (Config, error) {
	if cfg.GroupRegex != "" || len(cfg.Group) == 0 {
		return cfg, nil
	}

	tp, err := ParseTabularPattern(cfg.GroupPattern)
	if err != nil {
		return cfg, err
	}
	cfg.TabularPattern = &tp

	spec, err := parseGroupSpec(cfg.Group, cfg.GroupPattern, tp.TopSeparators)
	if err != nil {
		return cfg, err
	}

	if len(spec.Columns) == 0 {
		return cfg, nil
	}

	if spec.Structured {
		if !separatorsMatch(spec.Separators, tp.TopSeparators) {
			return cfg, errGroupPatternSeparatorMismatch(
				formatGroupDisplay(cfg.Group, true), cfg.GroupPattern,
				spec.Separators, tp.TopSeparators,
			)
		}
	}

	cfg.GroupColumns = spec.Columns
	cfg.GroupStructured = spec.Structured

	cfg.LabelSeparators = tp.TopSeparators
	if spec.Structured {
		cfg.LabelSeparators = spec.Separators
	}

	return cfg, nil
}

// ValidateTabularGroupAlignment ensures csv/json --group and --group-pattern use
// matching separators. Comma-split --group (e.g. -g region,product,month) requires
// comma-separated -p (e.g. -p x,y,z); slash patterns are not accepted there.
func ValidateTabularGroupAlignment(cfg Config) error {
	if len(cfg.Group) == 0 || cfg.GroupRegex != "" {
		return nil
	}

	if cfg.TabularPattern == nil {
		return fmt.Errorf("tabular pattern is not configured")
	}

	cols := EffectiveGroupColumns(cfg)
	slotCount := len(cfg.TabularPattern.Slots)
	if len(cols) != slotCount {
		return fmt.Errorf(
			"--group and --group-pattern dimension count do not match: %d column(s) in %q, %d slot(s) in --group-pattern %q",
			len(cols), formatGroupDisplay(cfg.Group, cfg.GroupStructured), slotCount, cfg.GroupPattern,
		)
	}

	patternSeps := cfg.TabularPattern.TopSeparators
	if len(cols) > 1 && len(cfg.Group) > 1 && !isStructuredGroup(cfg.Group) {
		expected := repeatSep(",", len(cols)-1)
		if !separatorsMatch(expected, patternSeps) {
			return errGroupPatternSeparatorMismatch(
				formatGroupDisplay(cfg.Group, false), cfg.GroupPattern,
				expected, patternSeps,
			)
		}
	}

	return nil
}

func formatSeparators(seps []string) string {
	if len(seps) == 0 {
		return "(none)"
	}
	return strings.Join(seps, " ")
}

func repeatSep(sep string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = sep
	}
	return out
}

// JoinLabelParts joins row values with the configured label separators.
func JoinLabelParts(values []string, seps []string) string {
	if len(values) == 0 {
		return ""
	}
	out := values[0]
	for i, sep := range seps {
		if i+1 < len(values) {
			out += sep + values[i+1]
		}
	}
	return out
}

// EffectiveGroupColumns returns resolved columns, falling back to trimmed --group values.
func EffectiveGroupColumns(cfg Config) []string {
	if len(cfg.GroupColumns) > 0 {
		return cfg.GroupColumns
	}
	var cols []string
	for _, g := range cfg.Group {
		if g = strings.TrimSpace(g); g != "" {
			cols = append(cols, g)
		}
	}
	return cols
}

// EffectiveLabelSeparators returns separators for label joining from resolved config.
func EffectiveLabelSeparators(cfg Config) []string {
	return cfg.LabelSeparators
}
