package parser

import (
	"fmt"
	"strings"
	"unicode"
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

// parseGroupSpec resolves column names and separators from the raw --group flag values
// and the separators declared in --group-pattern.
func parseGroupSpec(group []string, patternSeps []string) (GroupSpec, error) {
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
		return GroupSpec{}, fmt.Errorf(
			"--group %q does not match --group-pattern separators: expected %d column(s), found %d",
			group[0], expected, len(columns),
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

	spec, err := parseGroupSpec(cfg.Group, tp.TopSeparators)
	if err != nil {
		return cfg, err
	}

	if len(spec.Columns) == 0 {
		return cfg, nil
	}

	if spec.Structured {
		if !separatorsMatch(spec.Separators, tp.TopSeparators) {
			return cfg, fmt.Errorf(
				"--group-pattern %q separators do not match --group structure (expected %q, got %q)",
				cfg.GroupPattern, formatSeparators(spec.Separators), formatSeparators(tp.TopSeparators),
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
			"--group defines %d column(s) but --group-pattern %q defines %d slot(s)",
			len(cols), cfg.GroupPattern, slotCount,
		)
	}

	patternSeps := cfg.TabularPattern.TopSeparators
	if len(cols) > 1 && len(cfg.Group) > 1 && !isStructuredGroup(cfg.Group) {
		expected := repeatSep(",", len(cols)-1)
		if !separatorsMatch(expected, patternSeps) {
			return fmt.Errorf(
				"comma-separated --group requires comma-separated --group-pattern (e.g. -p x,y,z), not %q",
				cfg.GroupPattern,
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
