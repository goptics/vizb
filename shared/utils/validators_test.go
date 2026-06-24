package utils

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Test constants for validation rules
const (
	validFormat1  = "json"
	validFormat2  = "html"
	validFormat3  = "csv"
	defaultFormat = "json"
	invalidFormat = "xml"
)

var errValidation = errors.New("validation error")

type ValidatorsSuite struct {
	suite.Suite
}

func (s *ValidatorsSuite) TestApplyValidationRules() {
	s.Run("Valid value passes validation", func() {
		value := validFormat1
		rule := ValidationRule{
			Label:    "format",
			Value:    &value,
			ValidSet: []string{validFormat1, validFormat2, validFormat3},
			Default:  defaultFormat,
		}

		// Capture stderr to check no warning is printed
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		ApplyValidationRules([]ValidationRule{rule})

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		s.Equal(validFormat1, value, "Valid value should remain unchanged")
		s.Empty(buf.String(), "No warning should be printed for valid value")
	})

	s.Run("Invalid value gets replaced with default", func() {
		value := invalidFormat
		rule := ValidationRule{
			Label:    "format",
			Value:    &value,
			ValidSet: []string{validFormat1, validFormat2, validFormat3},
			Default:  defaultFormat,
		}

		// Capture stderr to check warning is printed
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		ApplyValidationRules([]ValidationRule{rule})

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		s.Equal(defaultFormat, value, "Invalid value should be replaced with default")
		s.Contains(buf.String(), "Warning: Invalid format", "Warning should be printed")
		s.Contains(buf.String(), invalidFormat, "Warning should contain invalid value")
		s.Contains(buf.String(), defaultFormat, "Warning should contain default value")
	})

	s.Run("Empty value and empty default are skipped", func() {
		value := ""
		rule := ValidationRule{
			Label:    "format",
			Value:    &value,
			ValidSet: []string{validFormat1, validFormat2},
			Default:  "",
		}

		ApplyValidationRules([]ValidationRule{rule})

		s.Equal("", value, "Empty value should remain empty when default is also empty")
	})

	s.Run("Empty value gets replaced with non-empty default", func() {
		value := ""
		rule := ValidationRule{
			Label:    "format",
			Value:    &value,
			ValidSet: []string{validFormat1, validFormat2, defaultFormat},
			Default:  defaultFormat,
		}

		ApplyValidationRules([]ValidationRule{rule})

		s.Equal(defaultFormat, value, "Empty value should be replaced with default")
	})

	s.Run("Normalizer function is applied", func() {
		value := "JSON"
		normalizer := strings.ToLower
		rule := ValidationRule{
			Label:      "format",
			Value:      &value,
			ValidSet:   []string{validFormat1, validFormat2, validFormat3},
			Normalizer: normalizer,
			Default:    defaultFormat,
		}

		ApplyValidationRules([]ValidationRule{rule})

		s.Equal("json", value, "Normalizer should convert to lowercase")
	})

	s.Run("Normalizer then validation", func() {
		value := "HTML"
		normalizer := strings.ToLower
		rule := ValidationRule{
			Label:      "format",
			Value:      &value,
			ValidSet:   []string{validFormat1, validFormat2, validFormat3},
			Normalizer: normalizer,
			Default:    defaultFormat,
		}

		ApplyValidationRules([]ValidationRule{rule})

		s.Equal(validFormat2, value, "Normalized value should pass validation")
	})

	s.Run("Normalizer with invalid result uses default", func() {
		value := "INVALID"
		normalizer := strings.ToLower
		rule := ValidationRule{
			Label:      "format",
			Value:      &value,
			ValidSet:   []string{validFormat1, validFormat2, validFormat3},
			Normalizer: normalizer,
			Default:    defaultFormat,
		}

		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		ApplyValidationRules([]ValidationRule{rule})

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		s.Equal(defaultFormat, value, "Invalid normalized value should use default")
		s.Contains(buf.String(), "invalid", "Warning should contain normalized invalid value")
	})

	s.Run("Multiple rules are processed", func() {
		format := "JSON"
		output := "invalid_output"

		formatRule := ValidationRule{
			Label:      "format",
			Value:      &format,
			ValidSet:   []string{validFormat1, validFormat2, validFormat3},
			Normalizer: strings.ToLower,
			Default:    defaultFormat,
		}

		outputRule := ValidationRule{
			Label:    "output",
			Value:    &output,
			ValidSet: []string{"file", "stdout", "stderr"},
			Default:  "stdout",
		}

		ApplyValidationRules([]ValidationRule{formatRule, outputRule})

		s.Equal("json", format, "First rule should normalize and validate")
		s.Equal("stdout", output, "Second rule should use default")
	})

	s.Run("Nil normalizer is handled safely", func() {
		value := validFormat1
		rule := ValidationRule{
			Label:      "format",
			Value:      &value,
			ValidSet:   []string{validFormat1, validFormat2, validFormat3},
			Normalizer: nil,
			Default:    defaultFormat,
		}

		ApplyValidationRules([]ValidationRule{rule})

		s.Equal(validFormat1, value, "Value should remain unchanged with nil normalizer")
	})
}

func (s *ValidatorsSuite) TestIsBenchJSONFile() {
	tempDir := s.T().TempDir()

	s.Run("Valid benchmark JSON file", func() {
		content := `{"Action":"run","Test":"BenchmarkExample"}
{"Action":"pass","Test":"BenchmarkExample","Output":"result"}
`
		filePath := filepath.Join(tempDir, "valid_bench.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		s.Require().NoError(err)

		result := IsBenchJSONFile(filePath)
		s.True(result, "Valid benchmark JSON should return true")
	})

	s.Run("Valid JSON with benchmark test name", func() {
		content := `{"Action":"run","Test":"BenchmarkMemoryUsage"}
`
		filePath := filepath.Join(tempDir, "benchmark.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		s.Require().NoError(err)

		result := IsBenchJSONFile(filePath)
		s.True(result, "JSON with Benchmark prefix should return true")
	})

	s.Run("Valid JSON without benchmark test name but with action", func() {
		content := `{"Action":"run","Test":"TestSomething"}
`
		filePath := filepath.Join(tempDir, "test.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		s.Require().NoError(err)

		result := IsBenchJSONFile(filePath)
		s.True(result, "JSON with valid action should return true")
	})

	s.Run("Empty action returns false", func() {
		content := `{"Action":"","Test":"BenchmarkExample"}
`
		filePath := filepath.Join(tempDir, "empty_action.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		s.Require().NoError(err)

		result := IsBenchJSONFile(filePath)
		s.False(result, "Empty action should return false")
	})

	s.Run("Missing action field returns false", func() {
		content := `{"Test":"BenchmarkExample"}
`
		filePath := filepath.Join(tempDir, "missing_action.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		s.Require().NoError(err)

		result := IsBenchJSONFile(filePath)
		s.False(result, "Missing action field should return false")
	})

	s.Run("Invalid JSON returns false", func() {
		content := `{"Action":"run","Test":"BenchmarkExample"
invalid json here
`
		filePath := filepath.Join(tempDir, "invalid.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		s.Require().NoError(err)

		result := IsBenchJSONFile(filePath)
		s.False(result, "Invalid JSON should return false")
	})

	s.Run("Empty file returns false", func() {
		filePath := filepath.Join(tempDir, "empty.json")
		err := os.WriteFile(filePath, []byte(""), 0644)
		s.Require().NoError(err)

		result := IsBenchJSONFile(filePath)
		s.False(result, "Empty file should return false")
	})

	s.Run("File with only whitespace returns false", func() {
		content := `


	`
		filePath := filepath.Join(tempDir, "whitespace.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		s.Require().NoError(err)

		result := IsBenchJSONFile(filePath)
		s.False(result, "File with only whitespace should return false")
	})

	s.Run("File with mixed valid and invalid lines", func() {
		content := `{"Action":"run","Test":"BenchmarkExample"}
invalid line here
{"Action":"pass","Test":"BenchmarkExample"}
`
		filePath := filepath.Join(tempDir, "mixed.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		s.Require().NoError(err)

		result := IsBenchJSONFile(filePath)
		s.True(result, "File with first valid line should return true")
	})

	s.Run("File with empty lines at start", func() {
		content := `

{"Action":"run","Test":"BenchmarkExample"}
`
		filePath := filepath.Join(tempDir, "empty_start.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		s.Require().NoError(err)

		result := IsBenchJSONFile(filePath)
		s.True(result, "File with empty lines at start should still work")
	})

	s.Run("Large file with valid first line", func() {
		var content strings.Builder
		content.WriteString(`{"Action":"run","Test":"BenchmarkLargeTest"}` + "\n")

		// Add many more lines
		for i := 0; i < 1000; i++ {
			content.WriteString(`{"Action":"pass","Test":"BenchmarkLargeTest","Output":"data"}` + "\n")
		}

		filePath := filepath.Join(tempDir, "large.json")
		err := os.WriteFile(filePath, []byte(content.String()), 0644)
		s.Require().NoError(err)

		result := IsBenchJSONFile(filePath)
		s.True(result, "Large file with valid format should return true")
	})
}

func (s *ValidatorsSuite) TestValidationEdgeCases() {
	s.Run("Nil value pointer safe", func() {
		rule := ValidationRule{
			Label:    "format",
			Value:    nil,
			ValidSet: []string{validFormat1},
			Default:  defaultFormat,
		}

		s.NotPanics(func() {
			ApplyValidationRules([]ValidationRule{rule})
		}, "Nil value pointer should not cause panic")
	})

	s.Run("Empty ValidSet with valid default", func() {
		value := "anything"
		rule := ValidationRule{
			Label:    "format",
			Value:    &value,
			ValidSet: []string{},
			Default:  defaultFormat,
		}

		ApplyValidationRules([]ValidationRule{rule})

		s.Equal(defaultFormat, value, "Empty ValidSet should always use default")
	})

	s.Run("Case sensitive validation", func() {
		value := "JSON"
		rule := ValidationRule{
			Label:    "format",
			Value:    &value,
			ValidSet: []string{"json", "html", "csv"},
			Default:  "json",
		}

		ApplyValidationRules([]ValidationRule{rule})

		s.Equal("json", value, "Case sensitive validation should use default for 'JSON' vs 'json'")
	})

	s.Run("Special characters in validation", func() {
		value := "special-format_123"
		rule := ValidationRule{
			Label:    "format",
			Value:    &value,
			ValidSet: []string{"special-format_123", "other-format"},
			Default:  "default",
		}

		ApplyValidationRules([]ValidationRule{rule})

		s.Equal("special-format_123", value, "Special characters should be handled correctly")
	})
}

func (s *ValidatorsSuite) TestNormalizerFunctions() {
	s.Run("Trim whitespace normalizer", func() {
		value := "  json  "
		rule := ValidationRule{
			Label:      "format",
			Value:      &value,
			ValidSet:   []string{"json", "html", "csv"},
			Normalizer: strings.TrimSpace,
			Default:    "json",
		}

		ApplyValidationRules([]ValidationRule{rule})

		s.Equal("json", value, "Whitespace should be trimmed")
	})

	s.Run("Replace normalizer", func() {
		value := "json-format"
		normalizer := func(s string) string {
			return strings.ReplaceAll(s, "-", "_")
		}
		rule := ValidationRule{
			Label:      "format",
			Value:      &value,
			ValidSet:   []string{"json_format", "html_format"},
			Normalizer: normalizer,
			Default:    "json_format",
		}

		ApplyValidationRules([]ValidationRule{rule})

		s.Equal("json_format", value, "Replace normalizer should work")
	})

	s.Run("Chain normalizers effect", func() {
		value := "  JSON-FORMAT  "
		normalizer := func(s string) string {
			s = strings.TrimSpace(s)
			s = strings.ToLower(s)
			s = strings.ReplaceAll(s, "-", "_")
			return s
		}
		rule := ValidationRule{
			Label:      "format",
			Value:      &value,
			ValidSet:   []string{"json_format", "html_format"},
			Normalizer: normalizer,
			Default:    "json_format",
		}

		ApplyValidationRules([]ValidationRule{rule})

		s.Equal("json_format", value, "Chained normalizer operations should work")
	})
}

func (s *ValidatorsSuite) TestApplyValidationRulesSlice() {
	s.Run("Valid slice values pass validation", func() {
		values := []string{"bar", "line"}
		rule := ValidationRule{
			Label:        "charts",
			SliceValue:   &values,
			ValidSet:     []string{"bar", "line", "pie"},
			SliceDefault: []string{"bar"},
		}

		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		ApplyValidationRules([]ValidationRule{rule})

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		s.Equal([]string{"bar", "line"}, values, "Valid slice should remain unchanged")
		s.Empty(buf.String(), "No warning should be printed for valid slice")
	})

	s.Run("Invalid slice value gets replaced with default", func() {
		values := []string{"bar", "invalid"}
		rule := ValidationRule{
			Label:        "charts",
			SliceValue:   &values,
			ValidSet:     []string{"bar", "line", "pie"},
			SliceDefault: []string{"bar", "line", "pie"},
		}

		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		ApplyValidationRules([]ValidationRule{rule})

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		s.Equal([]string{"bar", "line", "pie"}, values, "Invalid slice should be replaced with default")
		s.Contains(buf.String(), "Warning: Invalid charts", "Warning should be printed")
	})

	s.Run("Slice with normalizer", func() {
		values := []string{"BAR", "LINE"}
		rule := ValidationRule{
			Label:        "charts",
			SliceValue:   &values,
			ValidSet:     []string{"bar", "line", "pie"},
			Normalizer:   strings.ToLower,
			SliceDefault: []string{"bar"},
		}

		ApplyValidationRules([]ValidationRule{rule})

		s.Equal([]string{"bar", "line"}, values, "Slice values should be normalized")
	})

	s.Run("Empty slice remains empty", func() {
		values := []string{}
		rule := ValidationRule{
			Label:        "charts",
			SliceValue:   &values,
			ValidSet:     []string{"bar", "line", "pie"},
			SliceDefault: []string{"bar"},
		}

		ApplyValidationRules([]ValidationRule{rule})

		s.Equal([]string{}, values, "Empty slice should remain empty")
	})
}

func (s *ValidatorsSuite) TestApplyValidationRulesWithCustomValidator() {
	s.Run("Custom validator returns nil (success)", func() {
		value := "custom-value"
		rule := ValidationRule{
			Label: "custom",
			Value: &value,
			Validator: func(s string) error {
				return nil // Always valid
			},
			Default: "default",
		}

		ApplyValidationRules([]ValidationRule{rule})

		s.Equal("custom-value", value, "Value should remain unchanged when validator returns nil")
	})

	s.Run("Custom validator returns error", func() {
		value := "invalid-pattern"
		rule := ValidationRule{
			Label: "pattern",
			Value: &value,
			Validator: func(s string) error {
				return errValidation
			},
			Default: "default-pattern",
		}

		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		ApplyValidationRules([]ValidationRule{rule})

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		s.Equal("default-pattern", value, "Value should be replaced with default when validator returns error")
		s.Contains(buf.String(), "Warning: Invalid pattern", "Warning should be printed")
		s.Contains(buf.String(), "Reason:", "Warning should contain reason")
	})

	s.Run("Custom validator with specific error message", func() {
		value := "name"
		customErr := "pattern must contain xAxis (x) or yAxis (y)"
		rule := ValidationRule{
			Label: "group pattern",
			Value: &value,
			Validator: func(s string) error {
				if s == "name" {
					return errValidation
				}
				return nil
			},
			Default: "xAxis",
		}

		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		ApplyValidationRules([]ValidationRule{rule})

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		s.Equal("xAxis", value, "Invalid pattern should be replaced with default")
		s.Contains(buf.String(), "group pattern", "Warning should mention the label")
		_ = customErr // suppress unused warning
	})

	s.Run("Slice with custom validator", func() {
		values := []string{"valid", "also-valid"}
		rule := ValidationRule{
			Label:      "items",
			SliceValue: &values,
			Validator: func(s string) error {
				if s == "invalid" {
					return errValidation
				}
				return nil
			},
			SliceDefault: []string{"default"},
		}

		ApplyValidationRules([]ValidationRule{rule})

		s.Equal([]string{"valid", "also-valid"}, values, "Valid slice should remain unchanged")
	})

	s.Run("Slice with custom validator returns error", func() {
		values := []string{"valid", "invalid"}
		rule := ValidationRule{
			Label:      "items",
			SliceValue: &values,
			Validator: func(s string) error {
				if s == "invalid" {
					return errValidation
				}
				return nil
			},
			SliceDefault: []string{"default"},
		}

		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		ApplyValidationRules([]ValidationRule{rule})

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		s.Equal([]string{"default"}, values, "Slice with invalid item should be replaced with default")
		s.Contains(buf.String(), "Warning: Invalid items", "Warning should be printed")
	})
}

func (s *ValidatorsSuite) TestValidatorsConcurrency() {
	s.Run("Concurrent ApplyValidationRules", func() {
		const numGoroutines = 100

		results := make(chan string, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				value := "INVALID"
				rule := ValidationRule{
					Label:      "format",
					Value:      &value,
					ValidSet:   []string{"json", "html", "csv"},
					Normalizer: strings.ToLower,
					Default:    "json",
				}

				ApplyValidationRules([]ValidationRule{rule})
				results <- value
			}(i)
		}

		// Collect all results
		for i := 0; i < numGoroutines; i++ {
			result := <-results
			s.Equal("json", result, "All concurrent operations should produce consistent results")
		}
	})

	s.Run("Concurrent IsBenchJSONFile", func() {
		tempDir := s.T().TempDir()
		content := `{"Action":"run","Test":"BenchmarkExample"}`
		filePath := filepath.Join(tempDir, "concurrent_test.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		s.Require().NoError(err)

		const numGoroutines = 50
		results := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				result := IsBenchJSONFile(filePath)
				results <- result
			}()
		}

		// Collect all results
		for i := 0; i < numGoroutines; i++ {
			result := <-results
			s.True(result, "All concurrent file reads should succeed")
		}
	})
}

func TestValidatorsSuite(t *testing.T) {
	suite.Run(t, new(ValidatorsSuite))
}
