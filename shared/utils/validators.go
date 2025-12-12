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
	Validator    func(string) error
	Default      string
	SliceDefault []string
}

// ApplyValidationRules validates and normalizes flag values according to the provided rules.
// It checks each rule's validation function and normalizes the value if validation passes.
// If validation fails, the program exits with an appropriate error message.
func ApplyValidationRules(rules []ValidationRule) {
	for _, rule := range rules {
		rule.apply()
	}
}

func (r ValidationRule) apply() {
	switch {
	case r.Value != nil:
		r.validateValue()
	case r.SliceValue != nil:
		r.validateSlice()
	}
}

func (r ValidationRule) validateValue() {
	// skip validation if default and rule value are both empty
	if r.Default == "" && *r.Value == "" {
		return
	}

	r.normalizeValue()

	err := r.validate(*r.Value)
	if err == nil {
		return
	}

	r.printWarning(*r.Value, r.Default, err)
	*r.Value = r.Default
}

func (r ValidationRule) validateSlice() {
	r.normalizeSlice()

	for _, v := range *r.SliceValue {
		if err := r.validate(v); err != nil {
			r.printWarning(*r.SliceValue, r.SliceDefault, err)
			*r.SliceValue = r.SliceDefault
			return
		}
	}
}

func (r ValidationRule) normalizeValue() {
	if r.Normalizer != nil {
		*r.Value = r.Normalizer(*r.Value)
	}
}

func (r ValidationRule) normalizeSlice() {
	if r.Normalizer == nil {
		return
	}
	for i, v := range *r.SliceValue {
		(*r.SliceValue)[i] = r.Normalizer(v)
	}
}

func (r ValidationRule) validate(value string) error {
	if r.Validator != nil {
		return r.Validator(value)
	}
	if slices.Contains(r.ValidSet, value) {
		return nil
	}
	return fmt.Errorf("value not in valid set %v", r.ValidSet)
}

func (r ValidationRule) printWarning(value any, defaultValue any, err error) {
	errMsg := fmt.Sprintf("Warning: Invalid %s '%v'.", r.Label, value)
	if err != nil {
		errMsg += fmt.Sprintf(" Reason: %s.", err.Error())
	}
	errMsg += fmt.Sprintf(" Using default '%v'", defaultValue)
	shared.PrintWarning(errMsg)
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
