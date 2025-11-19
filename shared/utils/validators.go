package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/goptics/vizb/shared"
)

type ValidationRule struct {
	Label        string
	Value        *string
	SliceValue   *[]string
	ValidSet     []string
	Normalizer   func(string) string
	Validator    func(string) bool
	Default      string
	SliceDefault []string
}

// ApplyValidationRules validates and normalizes flag values according to the provided rules.
// It checks each rule's validation function and normalizes the value if validation passes.
// If validation fails, the program exits with an appropriate error message.
func ApplyValidationRules(rules []ValidationRule) {
	for _, rule := range rules {
		if rule.Value != nil {
			// skip validation if default and rule value are both empty
			if rule.Default == "" && *rule.Value == "" {
				continue
			}
			// Normalize if needed
			if rule.Normalizer != nil {
				*rule.Value = rule.Normalizer(*rule.Value)
			}

			isValid := slices.Contains(rule.ValidSet, *rule.Value)

			if rule.Validator != nil {
				isValid = rule.Validator(*rule.Value)
			}

			// Validate
			if !isValid {
				fmt.Fprintf(
					os.Stderr,
					"⚠️  Warning: Invalid %s '%s'. Using default '%s'\n",
					rule.Label,
					*rule.Value,
					rule.Default,
				)
				*rule.Value = rule.Default
			}
		} else if rule.SliceValue != nil {
			isValid := true
			for i, v := range *rule.SliceValue {
				if rule.Normalizer != nil {
					(*rule.SliceValue)[i] = rule.Normalizer(v)
					v = (*rule.SliceValue)[i]
				}

				itemValid := slices.Contains(rule.ValidSet, v)
				if rule.Validator != nil {
					itemValid = rule.Validator(v)
				}

				if !itemValid {
					isValid = false
					break
				}
			}

			if !isValid {
				fmt.Fprintf(
					os.Stderr,
					"⚠️  Warning: Invalid %s '%v'. Using default '%v'\n",
					rule.Label,
					*rule.SliceValue,
					rule.SliceDefault,
				)
				*rule.SliceValue = rule.SliceDefault
			}
		}
	}
}

// IsBenchJSONFile determines if the given file path contains JSON benchmark data.
// It opens the file and attempts to parse the first line as a JSON object.
// Returns true if the file contains valid JSON benchmark events, false otherwise.
func IsBenchJSONFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var ev shared.BenchEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			return false
		}

		// very Go-specific validation rules
		if ev.Action == "" {
			return false
		}

		// if it's a benchmark, "Test" usually begins with "Benchmark"
		if ev.Test != "" && strings.HasPrefix(ev.Test, "Benchmark") {
			return true
		}

		return true
	}

	return false
}
