package utils

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants for validation rules
const (
	validFormat1  = "json"
	validFormat2  = "html"
	validFormat3  = "csv"
	defaultFormat = "json"
	invalidFormat = "xml"
)

func TestApplyValidationRules(t *testing.T) {
	t.Run("Valid value passes validation", func(t *testing.T) {
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
		buf.ReadFrom(r)

		assert.Equal(t, validFormat1, value, "Valid value should remain unchanged")
		assert.Empty(t, buf.String(), "No warning should be printed for valid value")
	})

	t.Run("Invalid value gets replaced with default", func(t *testing.T) {
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
		buf.ReadFrom(r)

		assert.Equal(t, defaultFormat, value, "Invalid value should be replaced with default")
		assert.Contains(t, buf.String(), "Warning: Invalid format", "Warning should be printed")
		assert.Contains(t, buf.String(), invalidFormat, "Warning should contain invalid value")
		assert.Contains(t, buf.String(), defaultFormat, "Warning should contain default value")
	})

	t.Run("Empty value and empty default are skipped", func(t *testing.T) {
		value := ""
		rule := ValidationRule{
			Label:    "format",
			Value:    &value,
			ValidSet: []string{validFormat1, validFormat2},
			Default:  "",
		}

		ApplyValidationRules([]ValidationRule{rule})

		assert.Equal(t, "", value, "Empty value should remain empty when default is also empty")
	})

	t.Run("Empty value gets replaced with non-empty default", func(t *testing.T) {
		value := ""
		rule := ValidationRule{
			Label:    "format",
			Value:    &value,
			ValidSet: []string{validFormat1, validFormat2, defaultFormat},
			Default:  defaultFormat,
		}

		ApplyValidationRules([]ValidationRule{rule})

		assert.Equal(t, defaultFormat, value, "Empty value should be replaced with default")
	})

	t.Run("Normalizer function is applied", func(t *testing.T) {
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

		assert.Equal(t, "json", value, "Normalizer should convert to lowercase")
	})

	t.Run("Normalizer then validation", func(t *testing.T) {
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

		assert.Equal(t, validFormat2, value, "Normalized value should pass validation")
	})

	t.Run("Normalizer with invalid result uses default", func(t *testing.T) {
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
		buf.ReadFrom(r)

		assert.Equal(t, defaultFormat, value, "Invalid normalized value should use default")
		assert.Contains(t, buf.String(), "invalid", "Warning should contain normalized invalid value")
	})

	t.Run("Multiple rules are processed", func(t *testing.T) {
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

		assert.Equal(t, "json", format, "First rule should normalize and validate")
		assert.Equal(t, "stdout", output, "Second rule should use default")
	})

	t.Run("Nil normalizer is handled safely", func(t *testing.T) {
		value := validFormat1
		rule := ValidationRule{
			Label:      "format",
			Value:      &value,
			ValidSet:   []string{validFormat1, validFormat2, validFormat3},
			Normalizer: nil,
			Default:    defaultFormat,
		}

		ApplyValidationRules([]ValidationRule{rule})

		assert.Equal(t, validFormat1, value, "Value should remain unchanged with nil normalizer")
	})
}

func TestIsBenchJSONFile(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Valid benchmark JSON file", func(t *testing.T) {
		content := `{"Action":"run","Test":"BenchmarkExample"}
{"Action":"pass","Test":"BenchmarkExample","Output":"result"}
`
		filePath := filepath.Join(tempDir, "valid_bench.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		result := IsBenchJSONFile(filePath)
		assert.True(t, result, "Valid benchmark JSON should return true")
	})

	t.Run("Valid JSON with benchmark test name", func(t *testing.T) {
		content := `{"Action":"run","Test":"BenchmarkMemoryUsage"}
`
		filePath := filepath.Join(tempDir, "benchmark.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		result := IsBenchJSONFile(filePath)
		assert.True(t, result, "JSON with Benchmark prefix should return true")
	})

	t.Run("Valid JSON without benchmark test name but with action", func(t *testing.T) {
		content := `{"Action":"run","Test":"TestSomething"}
`
		filePath := filepath.Join(tempDir, "test.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		result := IsBenchJSONFile(filePath)
		assert.True(t, result, "JSON with valid action should return true")
	})

	t.Run("Empty action returns false", func(t *testing.T) {
		content := `{"Action":"","Test":"BenchmarkExample"}
`
		filePath := filepath.Join(tempDir, "empty_action.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		result := IsBenchJSONFile(filePath)
		assert.False(t, result, "Empty action should return false")
	})

	t.Run("Missing action field returns false", func(t *testing.T) {
		content := `{"Test":"BenchmarkExample"}
`
		filePath := filepath.Join(tempDir, "missing_action.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		result := IsBenchJSONFile(filePath)
		assert.False(t, result, "Missing action field should return false")
	})

	t.Run("Invalid JSON returns false", func(t *testing.T) {
		content := `{"Action":"run","Test":"BenchmarkExample"
invalid json here
`
		filePath := filepath.Join(tempDir, "invalid.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		result := IsBenchJSONFile(filePath)
		assert.False(t, result, "Invalid JSON should return false")
	})

	t.Run("Empty file returns false", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "empty.json")
		err := os.WriteFile(filePath, []byte(""), 0644)
		require.NoError(t, err)

		result := IsBenchJSONFile(filePath)
		assert.False(t, result, "Empty file should return false")
	})

	t.Run("File with only whitespace returns false", func(t *testing.T) {
		content := `

	`
		filePath := filepath.Join(tempDir, "whitespace.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		result := IsBenchJSONFile(filePath)
		assert.False(t, result, "File with only whitespace should return false")
	})

	t.Run("File with mixed valid and invalid lines", func(t *testing.T) {
		content := `{"Action":"run","Test":"BenchmarkExample"}
invalid line here
{"Action":"pass","Test":"BenchmarkExample"}
`
		filePath := filepath.Join(tempDir, "mixed.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		result := IsBenchJSONFile(filePath)
		assert.True(t, result, "File with first valid line should return true")
	})

	t.Run("File with empty lines at start", func(t *testing.T) {
		content := `

{"Action":"run","Test":"BenchmarkExample"}
`
		filePath := filepath.Join(tempDir, "empty_start.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		result := IsBenchJSONFile(filePath)
		assert.True(t, result, "File with empty lines at start should still work")
	})

	t.Run("Large file with valid first line", func(t *testing.T) {
		var content strings.Builder
		content.WriteString(`{"Action":"run","Test":"BenchmarkLargeTest"}` + "\n")

		// Add many more lines
		for i := 0; i < 1000; i++ {
			content.WriteString(`{"Action":"pass","Test":"BenchmarkLargeTest","Output":"data"}` + "\n")
		}

		filePath := filepath.Join(tempDir, "large.json")
		err := os.WriteFile(filePath, []byte(content.String()), 0644)
		require.NoError(t, err)

		result := IsBenchJSONFile(filePath)
		assert.True(t, result, "Large file with valid format should return true")
	})
}

// Test edge cases and error conditions
func TestValidationEdgeCases(t *testing.T) {
	t.Run("Nil value pointer panics", func(t *testing.T) {
		rule := ValidationRule{
			Label:    "format",
			Value:    nil,
			ValidSet: []string{validFormat1},
			Default:  defaultFormat,
		}

		assert.Panics(t, func() {
			ApplyValidationRules([]ValidationRule{rule})
		}, "Nil value pointer should cause panic")
	})

	t.Run("Empty ValidSet with valid default", func(t *testing.T) {
		value := "anything"
		rule := ValidationRule{
			Label:    "format",
			Value:    &value,
			ValidSet: []string{},
			Default:  defaultFormat,
		}

		ApplyValidationRules([]ValidationRule{rule})

		assert.Equal(t, defaultFormat, value, "Empty ValidSet should always use default")
	})

	t.Run("Case sensitive validation", func(t *testing.T) {
		value := "JSON"
		rule := ValidationRule{
			Label:    "format",
			Value:    &value,
			ValidSet: []string{"json", "html", "csv"},
			Default:  "json",
		}

		ApplyValidationRules([]ValidationRule{rule})

		assert.Equal(t, "json", value, "Case sensitive validation should use default for 'JSON' vs 'json'")
	})

	t.Run("Special characters in validation", func(t *testing.T) {
		value := "special-format_123"
		rule := ValidationRule{
			Label:    "format",
			Value:    &value,
			ValidSet: []string{"special-format_123", "other-format"},
			Default:  "default",
		}

		ApplyValidationRules([]ValidationRule{rule})

		assert.Equal(t, "special-format_123", value, "Special characters should be handled correctly")
	})
}

// Test complex normalizer functions
func TestNormalizerFunctions(t *testing.T) {
	t.Run("Trim whitespace normalizer", func(t *testing.T) {
		value := "  json  "
		rule := ValidationRule{
			Label:      "format",
			Value:      &value,
			ValidSet:   []string{"json", "html", "csv"},
			Normalizer: strings.TrimSpace,
			Default:    "json",
		}

		ApplyValidationRules([]ValidationRule{rule})

		assert.Equal(t, "json", value, "Whitespace should be trimmed")
	})

	t.Run("Replace normalizer", func(t *testing.T) {
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

		assert.Equal(t, "json_format", value, "Replace normalizer should work")
	})

	t.Run("Chain normalizers effect", func(t *testing.T) {
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

		assert.Equal(t, "json_format", value, "Chained normalizer operations should work")
	})
}

// Test concurrent usage
func TestValidatorsConcurrency(t *testing.T) {
	t.Run("Concurrent ApplyValidationRules", func(t *testing.T) {
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
			assert.Equal(t, "json", result, "All concurrent operations should produce consistent results")
		}
	})

	t.Run("Concurrent IsBenchJSONFile", func(t *testing.T) {
		tempDir := t.TempDir()
		content := `{"Action":"run","Test":"BenchmarkExample"}`
		filePath := filepath.Join(tempDir, "concurrent_test.json")
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

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
			assert.True(t, result, "All concurrent file reads should succeed")
		}
	})
}
