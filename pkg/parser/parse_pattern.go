package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/goptics/vizb/shared"
)

// dimensionTokens lists known pattern tokens longest-first so "xAxis" wins over "x".
var dimensionTokens = []struct {
	raw      string
	expanded string
}{
	{"xAxis", "xAxis"},
	{"yAxis", "yAxis"},
	{"zAxis", "zAxis"},
	{"name", "name"},
	{"x", "xAxis"},
	{"y", "yAxis"},
	{"z", "zAxis"},
	{"n", "name"},
}

// ParseBenchmarkNameToGroups parses a benchmark name using the given pattern
func ParseBenchmarkNameToGroups(name, pattern string) (map[string]string, error) {
	if err := ValidateGroupPattern(pattern); err != nil {
		return nil, err
	}

	patternParts := parsePatternParts(pattern)
	nameParts := splitNameByPattern(name, pattern)
	result := mapPartsToResult(patternParts, nameParts)

	return result, nil
}

type patternMatch struct {
	start    int
	raw      string
	expanded string
}

func isPatternIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func isPatternTokenBoundary(pattern string, pos, tokLen int) bool {
	beforeOK := pos == 0 || !isPatternIdentChar(pattern[pos-1])
	afterPos := pos + tokLen
	afterOK := afterPos >= len(pattern) || !isPatternIdentChar(pattern[afterPos])
	return beforeOK && afterOK
}

func findPatternMatches(pattern string) ([]patternMatch, error) {
	var matches []patternMatch
	for pos := 0; pos < len(pattern); {
		found := false
		for _, tok := range dimensionTokens {
			if strings.HasPrefix(pattern[pos:], tok.raw) && isPatternTokenBoundary(pattern, pos, len(tok.raw)) {
				matches = append(matches, patternMatch{start: pos, raw: tok.raw, expanded: tok.expanded})
				pos += len(tok.raw)
				found = true
				break
			}
		}
		if !found {
			pos++
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("Invalid part: %q; only name(n), xAxis(x), yAxis(y), zAxis(z) allowed", pattern)
	}
	return matches, nil
}

func invalidPatternRemainder(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return fmt.Errorf("Invalid part: %q; only name(n), xAxis(x), yAxis(y), zAxis(z) allowed", s)
		}
	}
	return nil
}

// tokenizePattern splits a group pattern into expanded dimension parts and separators between them.
// Leading or skipped segments produce empty parts so name splitting stays index-aligned.
func tokenizePattern(pattern string) (parts []string, separators []string, err error) {
	if pattern == "" {
		return nil, nil, errPatternEmpty
	}

	pattern, _, err = parsePatternLabels(pattern)
	if err != nil {
		return nil, nil, err
	}

	matches, err := findPatternMatches(pattern)
	if err != nil {
		return nil, nil, err
	}

	lastEnd := matches[len(matches)-1].start + len(matches[len(matches)-1].raw)
	if err := invalidPatternRemainder(pattern[lastEnd:]); err != nil {
		return nil, nil, err
	}

	if matches[0].start > 0 {
		parts = append(parts, "")
		separators = append(separators, pattern[0:matches[0].start])
	}

	for i, m := range matches {
		parts = append(parts, m.expanded)
		if i+1 < len(matches) {
			prevEnd := m.start + len(m.raw)
			separators = append(separators, pattern[prevEnd:matches[i+1].start])
		}
	}

	return parts, separators, nil
}

var (
	errPatternEmpty    = errors.New("pattern cannot be empty")
	errPatternNeedsXY  = errors.New("pattern must contain xAxis (x) or yAxis (y)")
	errPatternZNeedsXY = errors.New("zAxis (z) requires both xAxis (x) and yAxis (y)")
)

// ValidateGroupPattern validates the group pattern and regex string
func ValidateGroupPattern(pattern string) error {
	if strings.Contains(pattern, "[") {
		_, err := ParseTabularPattern(pattern)
		return err
	}

	parts, _, err := tokenizePattern(pattern)
	if err != nil {
		return err
	}

	var (
		hasXAxis bool
		hasYAxis bool
		hasZAxis bool
	)

	seen := map[string]bool{}
	for _, part := range parts {
		if part == "" {
			continue
		}

		if seen[part] {
			return fmt.Errorf("duplicate dimension '%s' in pattern", part)
		}
		seen[part] = true

		switch part {
		case "xAxis":
			hasXAxis = true
		case "yAxis":
			hasYAxis = true
		case "zAxis":
			hasZAxis = true
		}
	}

	if !hasXAxis && !hasYAxis {
		return errPatternNeedsXY
	}

	if hasZAxis && (!hasXAxis || !hasYAxis) {
		return errPatternZNeedsXY
	}

	return nil
}

func patternPartsHasBothXY(parts []string) bool {
	var hasXAxis, hasYAxis bool
	for _, part := range parts {
		switch part {
		case "xAxis":
			hasXAxis = true
		case "yAxis":
			hasYAxis = true
		}
	}
	return hasXAxis && hasYAxis
}

// PatternHasBothXY reports whether the resolved group config declares both x and y
// dimensions. Value-mode --3d requires this; grouped 3D (z present) does not.
func PatternHasBothXY(cfg Config) bool {
	if cfg.GroupRegex != "" {
		re, err := regexp.Compile(cfg.GroupRegex)
		if err != nil {
			return false
		}
		var hasXAxis, hasYAxis bool
		for _, capName := range re.SubexpNames() {
			if capName == "" {
				continue
			}
			switch expandShorthand(capName) {
			case "xAxis":
				hasXAxis = true
			case "yAxis":
				hasYAxis = true
			}
		}
		return hasXAxis && hasYAxis
	}

	if cfg.TabularPattern != nil {
		var hasXAxis, hasYAxis bool
		for _, slot := range cfg.TabularPattern.Slots {
			var dims []string
			if slot.ValueSplit {
				dims = parsePatternParts(slot.InnerPattern)
			} else if slot.Dimension != "" {
				dims = []string{slot.Dimension}
			}
			for _, dim := range dims {
				switch dim {
				case "xAxis":
					hasXAxis = true
				case "yAxis":
					hasYAxis = true
				}
			}
		}
		return hasXAxis && hasYAxis
	}

	if cfg.GroupPattern == "" {
		return false
	}

	return patternPartsHasBothXY(parsePatternParts(cfg.GroupPattern))
}

// parsePatternParts extracts and normalizes pattern parts
func parsePatternParts(pattern string) []string {
	parts, _, err := tokenizePattern(pattern)
	if err != nil {
		return nil
	}
	return parts
}

// patternSeparators returns the separator strings declared in a group pattern.
func patternSeparators(pattern string) []string {
	_, seps, err := tokenizePattern(pattern)
	if err != nil {
		return nil
	}
	return seps
}

// splitNameByPattern splits benchmark name using separators from pattern
func splitNameByPattern(name, pattern string) []string {
	separators := patternSeparators(pattern)
	return splitBySeparators(name, separators)
}

// mapPartsToResult maps pattern parts to name parts
func mapPartsToResult(patternParts, nameParts []string) map[string]string {
	result := map[string]string{
		"name":  "",
		"xAxis": "",
		"yAxis": "",
		"zAxis": "",
	}

	for i, part := range patternParts {
		if part == "" {
			continue
		}

		if i < len(nameParts) {
			result[part] = nameParts[i]
		}
	}

	return result
}

// expandShorthand converts shorthand to full name
func expandShorthand(part string) string {
	shortcuts := map[string]string{
		"n": "name",
		"x": "xAxis",
		"y": "yAxis",
		"z": "zAxis",
	}
	if expanded, exists := shortcuts[part]; exists {
		return expanded
	}
	return part
}

// ParseBenchmarkNameWithRegex parses a benchmark name using the given regex pattern
func ParseBenchmarkNameWithRegex(name, pattern string) (map[string]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	match := re.FindStringSubmatch(name)
	if match == nil {
		return nil, fmt.Errorf("benchmark name '%s' does not match regex '%s'", name, pattern)
	}

	result := map[string]string{
		"name":  "",
		"xAxis": "",
		"yAxis": "",
		"zAxis": "",
	}

	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			expandedName := expandShorthand(name)
			if i < len(match) {
				result[expandedName] = match[i]
			}
		}
	}

	if result["xAxis"] == "" && result["yAxis"] == "" {
		return nil, fmt.Errorf("regex '%s' does not contain x (xAxis) or y (yAxis)", pattern)
	}

	if result["zAxis"] != "" && (result["xAxis"] == "" || result["yAxis"] == "") {
		return nil, fmt.Errorf("regex '%s' captures zAxis (z) but z requires both xAxis (x) and yAxis (y)", pattern)
	}

	return result, nil
}

// shortAxisKey maps the full dimension name to its canonical short key used in
// shared.Axis.Key ("name"→"name", "xAxis"→"x", "yAxis"→"y", "zAxis"→"z").
func shortAxisKey(dim string) string {
	switch dim {
	case "xAxis":
		return "x"
	case "yAxis":
		return "y"
	case "zAxis":
		return "z"
	default:
		return dim
	}
}

// GroupAxes returns the ordered list of axis descriptors for the given grouping
// configuration.
func GroupAxes(cfg Config) []shared.Axis {
	if cfg.GroupRegex == "" && cfg.GroupPattern == "" {
		return nil
	}

	if cfg.TabularPattern != nil {
		return groupAxesFromTabular(cfg)
	}

	if cfg.GroupRegex != "" {
		re, err := regexp.Compile(cfg.GroupRegex)
		if err != nil {
			return nil
		}

		canonicalOrder := []string{"name", "x", "y", "z"}
		found := map[string]bool{}
		for _, capName := range re.SubexpNames() {
			if capName == "" {
				continue
			}
			expanded := expandShorthand(capName)
			switch expanded {
			case "name":
				found["name"] = true
			case "xAxis":
				found["x"] = true
			case "yAxis":
				found["y"] = true
			case "zAxis":
				found["z"] = true
			}
		}

		var axes []shared.Axis
		for _, key := range canonicalOrder {
			if found[key] {
				axes = append(axes, shared.Axis{Key: key, Label: ""})
			}
		}
		return axes
	}

	_, patternLabels, err := parsePatternLabels(cfg.GroupPattern)
	if err != nil {
		return nil
	}

	groupLabels := EffectiveGroupColumns(cfg)
	parts := parsePatternParts(cfg.GroupPattern)
	var axes []shared.Axis
	labelIdx := 0
	for _, dim := range parts {
		if dim == "" {
			continue
		}
		label := patternLabels[dim]
		if label == "" && labelIdx < len(groupLabels) {
			label = groupLabels[labelIdx]
		}
		labelIdx++
		axes = append(axes, shared.Axis{Key: shortAxisKey(dim), Label: label})
	}
	return axes
}

func GroupBenchmarkName(name string, cfg Config) (map[string]string, error) {
	if cfg.GroupRegex != "" {
		return ParseBenchmarkNameWithRegex(name, cfg.GroupRegex)
	}

	if strings.Contains(cfg.GroupPattern, "[") {
		return nil, fmt.Errorf("bracket slots [...] in --group-pattern are only supported for CSV/JSON data")
	}

	return ParseBenchmarkNameToGroups(name, cfg.GroupPattern)
}
