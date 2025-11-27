package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var separatorRegex = regexp.MustCompile(`[_/]`)

// ParseBenchmarkNameToGroups parses a benchmark name using the given pattern
func ParseBenchmarkNameToGroups(name, pattern string) (map[string]string, error) {
	if err := ValidatePattern(pattern); err != nil {
		return nil, err
	}

	patternParts := parsePatternParts(pattern)
	nameParts := splitNameByPattern(name, pattern)
	result := mapPartsToResult(patternParts, nameParts)

	return result, nil
}

// ValidatePattern validates the pattern string
func ValidatePattern(pattern string) error {
	if pattern == "" {
		return errors.New("pattern cannot be empty")
	}

	validParts := regexp.MustCompile(`^[nxy]|name|xAxis|yAxis$`)
	parts := separatorRegex.Split(pattern, -1)

	for _, part := range parts {
		if part == "" {
			continue
		}

		if !validParts.MatchString(part) {
			return fmt.Errorf("Invalid part: '%s'; only name(n), xAxis(x), yAxis(y) allowed", part)
		}
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
	}

	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			expandedName := expandShorthand(name)
			if i < len(match) {
				result[expandedName] = match[i]
			}
		}
	}

	return result, nil
}
