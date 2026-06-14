package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/goptics/vizb/shared"
)

var separatorRegex = regexp.MustCompile(`[_/]`)

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

// ValidateGroupPattern validates the group pattern and regex string
func ValidateGroupPattern(pattern string) error {
	if pattern == "" {
		return errors.New("pattern cannot be empty")
	}

	validParts := regexp.MustCompile(`^[nxyz]|name|xAxis|yAxis|zAxis$`)
	parts := separatorRegex.Split(pattern, -1)

	var (
		hasXAxis bool
		hasYAxis bool
		hasZAxis bool
	)

	for _, part := range parts {
		if part == "" {
			continue
		}

		if !validParts.MatchString(part) {
			return fmt.Errorf("Invalid part: '%s'; only name(n), xAxis(x), yAxis(y), zAxis(z) allowed", part)
		}

		switch expandShorthand(part) {
		case "xAxis":
			hasXAxis = true
		case "yAxis":
			hasYAxis = true
		case "zAxis":
			hasZAxis = true
		}
	}

	if !hasXAxis && !hasYAxis {
		return errors.New("pattern must contain xAxis (x) or yAxis (y)")
	}

	// zAxis defines the third (depth) dimension of a 3D chart, which needs an
	// x/y floor; reject z unless both x and y are present.
	if hasZAxis && (!hasXAxis || !hasYAxis) {
		return errors.New("zAxis (z) requires both xAxis (x) and yAxis (y)")
	}

	return nil
}

// parsePatternParts extracts and normalizes pattern parts
func parsePatternParts(pattern string) []string {
	parts := separatorRegex.Split(pattern, -1)

	for i, part := range parts {
		parts[i] = expandShorthand(part)
	}

	return parts
}

// splitNameByPattern splits benchmark name using separators from pattern
func splitNameByPattern(name, pattern string) []string {
	separators := separatorRegex.FindAllString(pattern, -1)
	if len(separators) == 0 {
		return []string{name}
	}

	parts := []string{name}
	for _, sep := range separators {
		var newParts []string
		for _, part := range parts {
			split := strings.SplitN(part, sep, 2)
			newParts = append(newParts, split...)
		}
		parts = newParts
	}

	// Filter empty parts
	var result []string
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
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

	// zAxis is the depth dimension of a 3D chart, which needs an x/y floor;
	// reject z unless both x and y are also captured.
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
		return dim // "name" stays "name"
	}
}

// GroupAxes returns the ordered list of axis descriptors for the current
// grouping configuration, reading from shared.FlagState.
//
// Pattern mode: dims are taken from --group-pattern in order; each dim's label
// is the positional --group[i] value (empty string when no label supplied).
//
// Regex mode: named capture groups (x/y/z/n) define dims; labels are always
// empty (captures provide values, not column names).
//
// No grouping (both empty): returns nil.
func GroupAxes() []shared.Axis {
	if shared.FlagState.GroupRegex == "" && shared.FlagState.GroupPattern == "" {
		return nil
	}

	if shared.FlagState.GroupRegex != "" {
		// Regex mode: inspect named sub-expressions for known axis names.
		re, err := regexp.Compile(shared.FlagState.GroupRegex)
		if err != nil {
			return nil
		}

		// canonicalOrder defines output sequence: name, x, y, z.
		canonicalOrder := []string{"name", "x", "y", "z"}
		found := map[string]bool{}
		for _, capName := range re.SubexpNames() {
			if capName == "" {
				continue
			}
			expanded := expandShorthand(capName) // n→name, x→xAxis, y→yAxis, z→zAxis
			// Map back to short keys used in Axis.Key.
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

	// Pattern mode: dims come from --group-pattern in declaration order.
	// Collect non-empty trimmed --group labels positionally.
	var groups []string
	for _, g := range shared.FlagState.Group {
		if g = strings.TrimSpace(g); g != "" {
			groups = append(groups, g)
		}
	}

	parts := parsePatternParts(shared.FlagState.GroupPattern)
	var axes []shared.Axis
	for i, dim := range parts {
		if dim == "" {
			continue
		}
		label := ""
		if i < len(groups) {
			label = groups[i]
		}
		axes = append(axes, shared.Axis{Key: shortAxisKey(dim), Label: label})
	}
	return axes
}

func GroupBenchmarkName(name string) (map[string]string, error) {
	if shared.FlagState.GroupRegex != "" {
		return ParseBenchmarkNameWithRegex(name, shared.FlagState.GroupRegex)
	}

	return ParseBenchmarkNameToGroups(name, shared.FlagState.GroupPattern)
}
