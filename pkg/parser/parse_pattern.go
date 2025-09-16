package parser

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
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

	// If not enough parts in name, only return what we have
	if len(nameParts) < len(patternParts) && hasNameInPattern(patternParts) {
		filteredResult := make(map[string]string)
		for key, value := range result {
			if value != "" {
				filteredResult[key] = value
			}
		}
		return filteredResult, nil
	}

	// Default behavior: if no name in pattern, add implicit name=""
	if !hasNameInPattern(patternParts) && len(patternParts) > 1 {
		result["name"] = ""
	}

	fmt.Println("name is not removing", result)

	return result, nil
}

// ValidatePattern validates the pattern string
func ValidatePattern(pattern string) error {
	if pattern == "" {
		return errors.New("pattern cannot be empty")
	}

	validParts := regexp.MustCompile(`^[nsw]|name|subject|workload$`)
	parts := separatorRegex.Split(pattern, -1)

	for _, part := range parts {
		if !validParts.MatchString(part) {
			return fmt.Errorf("Invalid part: '%s'; only name(n), subject(s), workload(w) allowed", part)
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
		"name":     "",
		"workload": "",
		"subject":  "",
	}

	for i, part := range patternParts {
		if i < len(nameParts) {
			result[part] = nameParts[i]
		} else {
			result[part] = ""
		}
	}

	return result
}

// hasNameInPattern checks if pattern contains name
func hasNameInPattern(parts []string) bool {
	return slices.Contains(parts, "name")
}

// expandShorthand converts shorthand to full name
func expandShorthand(part string) string {
	shortcuts := map[string]string{
		"n": "name",
		"s": "subject",
		"w": "workload",
	}
	if expanded, exists := shortcuts[part]; exists {
		return expanded
	}
	return part
}
