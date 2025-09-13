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
	Label      string
	Value      *string
	ValidSet   []string
	Normalizer func(string) string
	Default    string
}

func ApplyValidationRules(rules []ValidationRule) {
	for _, rule := range rules {
		// Normalize if needed
		if rule.Normalizer != nil {
			*rule.Value = rule.Normalizer(*rule.Value)
		}

		// Validate
		if !slices.Contains(rule.ValidSet, *rule.Value) {
			fmt.Fprintf(
				os.Stderr,
				"Warning: Invalid %s '%s'. Using default '%s'\n",
				rule.Label,
				*rule.Value,
				rule.Default,
			)
			*rule.Value = rule.Default
		}
	}
}

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
