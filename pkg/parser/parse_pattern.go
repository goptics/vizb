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

	validParts := regexp.MustCompile(`^[nsw]|name|subject|workload$`)
	parts := separatorRegex.Split(pattern, -1)

	// subject is required
	var hasSubject bool
	for _, part := range parts {
		// Skip empty parts (from leading/trailing separators)
		if part == "" {
			continue
		}

		if !validParts.MatchString(part) {
			return fmt.Errorf("Invalid part: '%s'; only name(n), subject(s), workload(w) allowed", part)
		}
		if part == "s" || part == "subject" {
			hasSubject = true
		}
	}

	if !hasSubject {
		return errors.New("pattern must contain subject(s)")
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
		"s": "subject",
		"w": "workload",
	}
	if expanded, exists := shortcuts[part]; exists {
		return expanded
	}
	return part
}
