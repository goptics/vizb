package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var separatorRegex = regexp.MustCompile(`[_/]`)

// splitPascalCase splits a PascalCase string into individual words
func splitPascalCase(s string) []string {
	if s == "" {
		return []string{}
	}

	var words []string
	var current strings.Builder
	runes := []rune(s)

	for i, r := range runes {
		if i == 0 {
			current.WriteRune(r)
		} else if unicode.IsUpper(r) {
			// Check if we have consecutive uppercase letters (like HTTP)
			// If next character is lowercase, split before current char
			// If previous char was uppercase and current is uppercase, continue building
			prevUpper := i > 0 && unicode.IsUpper(runes[i-1])
			nextLower := i < len(runes)-1 && unicode.IsLower(runes[i+1])

			if !prevUpper || nextLower {
				if current.Len() > 0 {
					words = append(words, current.String())
					current.Reset()
				}
			}
			current.WriteRune(r)
		} else {
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}

// extractSquareBracketIndices extracts the indices from square bracket notation
// e.g., "[s,w,n]" returns ["s", "w", "n"]
// e.g., "[,w]" returns ["", "w"]
func extractSquareBracketIndices(pattern string) []string {
	if !strings.HasPrefix(pattern, "[") || !strings.HasSuffix(pattern, "]") {
		return nil
	}

	content := strings.TrimPrefix(strings.TrimSuffix(pattern, "]"), "[")
	if content == "" {
		return []string{}
	}

	return strings.Split(content, ",")
}

// ParseBenchmarkNameToGroups parses a benchmark name using the given pattern
func ParseBenchmarkNameToGroups(name, pattern string) (map[string]string, error) {
	if err := ValidatePattern(pattern); err != nil {
		return nil, err
	}

	result := map[string]string{
		"name":     "",
		"workload": "",
		"subject":  "",
	}

	// Split pattern by separators to get pattern parts
	rawPatternParts := separatorRegex.Split(pattern, -1)
	separators := separatorRegex.FindAllString(pattern, -1)

	// Split name by separators first
	nameParts := []string{name}
	for _, sep := range separators {
		var newParts []string
		for _, part := range nameParts {
			split := strings.SplitN(part, sep, 2)
			newParts = append(newParts, split...)
		}
		nameParts = newParts
	}

	// Filter empty name parts
	var filteredNameParts []string
	for _, part := range nameParts {
		if part != "" {
			filteredNameParts = append(filteredNameParts, part)
		}
	}

	// Now map each pattern part to corresponding name part
	for i, rawPatternPart := range rawPatternParts {
		if rawPatternPart == "" || i >= len(filteredNameParts) {
			continue
		}

		if strings.HasPrefix(rawPatternPart, "[") && strings.HasSuffix(rawPatternPart, "]") {
			// Square bracket pattern - split name part by PascalCase and map indices
			pascalWords := splitPascalCase(filteredNameParts[i])
			indices := extractSquareBracketIndices(rawPatternPart)

			// Count leading empty indices to determine skip offset
			skipCount := 0
			for _, index := range indices {
				if index == "" {
					skipCount++
				} else {
					break
				}
			}

			// Map non-empty indices to words, starting after the skip offset
			wordIndex := skipCount
			for _, index := range indices {
				if index != "" {
					if wordIndex < len(pascalWords) {
						expandedIndex := expandShorthand(index)
						result[expandedIndex] = pascalWords[wordIndex]
					}
					wordIndex++
				}
			}
		} else {
			// Regular pattern part
			expandedPart := expandShorthand(rawPatternPart)
			result[expandedPart] = filteredNameParts[i]
		}
	}

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

		// Handle square bracket patterns
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			indices := extractSquareBracketIndices(part)
			for _, index := range indices {
				if index != "" && !validParts.MatchString(index) {
					return fmt.Errorf("Invalid part in square brackets: '%s'; only name(n), subject(s), workload(w) allowed", index)
				}
				if index == "s" || index == "subject" {
					hasSubject = true
				}
			}
		} else if !validParts.MatchString(part) {
			return fmt.Errorf("Invalid part: '%s'; only name(n), subject(s), workload(w) allowed", part)
		} else {
			if part == "s" || part == "subject" {
				hasSubject = true
			}
		}
	}

	if !hasSubject {
		return errors.New("pattern must contain subject(s)")
	}

	return nil
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
