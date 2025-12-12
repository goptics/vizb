package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
	"github.com/stretchr/testify/assert"
)

func TestFlagValidationRules(t *testing.T) {
	// Save original flag state
	origMemUnit := shared.FlagState.MemUnit
	origTimeUnit := shared.FlagState.TimeUnit
	origAllocUnit := shared.FlagState.NumberUnit
	origFormat := shared.FlagState.Format
	origSort := shared.FlagState.Sort
	origCharts := shared.FlagState.Charts

	defer func() {
		// Restore original values
		shared.FlagState.MemUnit = origMemUnit
		shared.FlagState.TimeUnit = origTimeUnit
		shared.FlagState.NumberUnit = origAllocUnit
		shared.FlagState.Format = origFormat
		shared.FlagState.Sort = origSort
		shared.FlagState.Charts = origCharts
	}()

	t.Run("Memory unit validation", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
			hasWarn  bool
		}{
			{"B", "B", false},      // skip normalization
			{"b", "b", false},      // skip normalization
			{"KB", "KB", false},    // Valid, normalized
			{"mb", "MB", false},    // Valid, already lowercase
			{"gb", "GB", false},    // Valid
			{"invalid", "B", true}, // Invalid, uses default
			{"", "B", true},        // Empty, uses default
		}

		for _, tt := range tests {
			t.Run("input_"+tt.input, func(t *testing.T) {
				shared.FlagState.MemUnit = tt.input

				// Capture stderr
				oldStderr := os.Stderr
				r, w, _ := os.Pipe()
				os.Stderr = w

				utils.ApplyValidationRules(flagValidationRules)

				w.Close()
				os.Stderr = oldStderr

				var buf bytes.Buffer
				buf.ReadFrom(r)
				stderr := buf.String()

				assert.Equal(t, tt.expected, shared.FlagState.MemUnit)
				if tt.hasWarn {
					assert.Contains(t, stderr, "Warning")
					assert.Contains(t, stderr, "memory unit")
				} else {
					assert.Empty(t, stderr)
				}
			})
		}
	})

	t.Run("Time unit validation", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
			hasWarn  bool
		}{
			{"ns", "ns", false},
			{"us", "us", false},
			{"ms", "ms", false},
			{"s", "s", false},
			{"NS", "ns", true},      // Invalid case, uses default
			{"invalid", "ns", true}, // Invalid, uses default
		}

		for _, tt := range tests {
			t.Run("input_"+tt.input, func(t *testing.T) {
				shared.FlagState.TimeUnit = tt.input

				oldStderr := os.Stderr
				r, w, _ := os.Pipe()
				os.Stderr = w

				utils.ApplyValidationRules(flagValidationRules)

				w.Close()
				os.Stderr = oldStderr

				var buf bytes.Buffer
				buf.ReadFrom(r)
				stderr := buf.String()

				assert.Equal(t, tt.expected, shared.FlagState.TimeUnit)
				if tt.hasWarn {
					assert.Contains(t, stderr, "time unit")
				}
			})
		}
	})

	t.Run("Number unit validation", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
			hasWarn  bool
		}{
			{"K", "K", false},
			{"M", "M", false},
			{"B", "B", false},
			{"T", "T", false},
			{"k", "K", false},     // Valid, normalized to uppercase
			{"m", "M", false},     // Valid, normalized
			{"", "", false},       // Empty is allowed for alloc unit
			{"invalid", "", true}, // Invalid, uses empty default
		}

		for _, tt := range tests {
			t.Run("input_"+tt.input, func(t *testing.T) {
				shared.FlagState.NumberUnit = tt.input

				oldStderr := os.Stderr
				r, w, _ := os.Pipe()
				os.Stderr = w

				utils.ApplyValidationRules(flagValidationRules)

				w.Close()
				os.Stderr = oldStderr

				var buf bytes.Buffer
				buf.ReadFrom(r)
				stderr := buf.String()

				assert.Equal(t, tt.expected, shared.FlagState.NumberUnit)
				if tt.hasWarn {
					assert.Contains(t, stderr, "number unit")
				}
			})
		}
	})

	t.Run("Format validation", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
			hasWarn  bool
		}{
			{"html", "html", false},
			{"json", "json", false},
			{"HTML", "html", false}, // Valid, normalized
			{"JSON", "json", false}, // Valid, normalized
			{"xml", "html", true},   // Invalid, uses default
			{"csv", "html", true},   // Invalid, uses default
		}

		for _, tt := range tests {
			t.Run("input_"+tt.input, func(t *testing.T) {
				shared.FlagState.Format = tt.input

				oldStderr := os.Stderr
				r, w, _ := os.Pipe()
				os.Stderr = w

				utils.ApplyValidationRules(flagValidationRules)

				w.Close()
				os.Stderr = oldStderr

				var buf bytes.Buffer
				buf.ReadFrom(r)
				stderr := buf.String()

				assert.Equal(t, tt.expected, shared.FlagState.Format)
				if tt.hasWarn {
					assert.Contains(t, stderr, "format")
				}
			})
		}
	})

	t.Run("Multiple invalid flags", func(t *testing.T) {
		shared.FlagState.MemUnit = "invalid"
		shared.FlagState.TimeUnit = "invalid"
		shared.FlagState.Format = "invalid"
		shared.FlagState.NumberUnit = "invalid"

		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		utils.ApplyValidationRules(flagValidationRules)

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		buf.ReadFrom(r)
		stderr := buf.String()

		// Should have warnings for all invalid flags
		assert.Contains(t, stderr, "memory unit")
		assert.Contains(t, stderr, "time unit")
		assert.Contains(t, stderr, "format")
		assert.Contains(t, stderr, "number unit")

		// Should use defaults
		assert.Equal(t, "B", shared.FlagState.MemUnit)
		assert.Equal(t, "ns", shared.FlagState.TimeUnit)
		assert.Equal(t, "html", shared.FlagState.Format)
		assert.Equal(t, "", shared.FlagState.NumberUnit)
	})
}

func TestFlagValidationRulesStructure(t *testing.T) {
	t.Run("All validation rules are present", func(t *testing.T) {
		// Check that all expected rules are present
		labels := make(map[string]bool)
		for _, rule := range flagValidationRules {
			labels[rule.Label] = true
		}

		assert.True(t, labels["memory unit"], "Should have memory unit rule")
		assert.True(t, labels["time unit"], "Should have time unit rule")
		assert.True(t, labels["number unit"], "Should have number unit rule")
		assert.True(t, labels["format"], "Should have format rule")
		assert.True(t, labels["sort order"], "Should have sort order rule")
		assert.True(t, labels["charts"], "Should have charts rule")
		assert.True(t, labels["group pattern"], "Should have group pattern rule")
	})

	t.Run("Validation rules have correct properties", func(t *testing.T) {
		for _, rule := range flagValidationRules {
			t.Run("rule_"+rule.Label, func(t *testing.T) {
				assert.NotEmpty(t, rule.Label, "Rule should have a label")

				if rule.SliceValue != nil {
					assert.NotNil(t, rule.SliceValue, "Rule should have a slice value pointer")
					assert.Nil(t, rule.Value, "Rule with SliceValue should not have Value")
				} else {
					assert.NotNil(t, rule.Value, "Rule should have a value pointer")
					assert.Nil(t, rule.SliceValue, "Rule with Value should not have SliceValue")
				}

				// Check specific rule properties
				switch rule.Label {
				case "memory unit":
					assert.Contains(t, rule.ValidSet, "B", "Memory unit should accept 'B'")
					assert.Contains(t, rule.ValidSet, "KB", "Memory unit should accept 'kb'")
					assert.Equal(t, "B", rule.Default)
					assert.NotNil(t, rule.Normalizer, "Memory unit should have normalizer")

				case "time unit":
					assert.Contains(t, rule.ValidSet, "ns", "Time unit should accept 'ns'")
					assert.Contains(t, rule.ValidSet, "s", "Time unit should accept 's'")
					assert.Equal(t, "ns", rule.Default)
					assert.Nil(t, rule.Normalizer, "Time unit should not have normalizer")

				case "number unit":
					assert.Contains(t, rule.ValidSet, "K", "Number unit should accept 'K'")
					assert.Contains(t, rule.ValidSet, "M", "Number unit should accept 'M'")
					assert.Equal(t, "", rule.Default)
					assert.NotNil(t, rule.Normalizer, "Number unit should have normalizer")

				case "format":
					assert.Contains(t, rule.ValidSet, "html", "Format should accept 'html'")
					assert.Contains(t, rule.ValidSet, "json", "Format should accept 'json'")
					assert.Equal(t, "html", rule.Default)
					assert.NotNil(t, rule.Normalizer, "Format should have normalizer")

				case "sort order":
					assert.Contains(t, rule.ValidSet, "asc")
					assert.Contains(t, rule.ValidSet, "desc")
					assert.Equal(t, "", rule.Default)

				case "charts":
					assert.Contains(t, rule.ValidSet, "bar")
					assert.Contains(t, rule.ValidSet, "line")
					assert.NotNil(t, rule.SliceValue)
				}
			})
		}
	})
}

// Test normalizer functions work correctly
func TestFlagNormalizers(t *testing.T) {
	t.Run("Memory unit normalizer", func(t *testing.T) {
		memRule := flagValidationRules[0] // Memory unit rule
		assert.Equal(t, "KB", memRule.Normalizer("kb"))
		assert.Equal(t, "MB", memRule.Normalizer("mb"))
		assert.Equal(t, "GB", memRule.Normalizer("gb"))
		assert.Equal(t, "b", memRule.Normalizer("b"))
		assert.Equal(t, "B", memRule.Normalizer("B"))

	})

	t.Run("Number unit normalizer", func(t *testing.T) {
		allocRule := flagValidationRules[2] // Number unit rule
		assert.Equal(t, "K", allocRule.Normalizer("k"))
		assert.Equal(t, "M", allocRule.Normalizer("m"))
		assert.Equal(t, "B", allocRule.Normalizer("b"))
		assert.Equal(t, "T", allocRule.Normalizer("t"))
	})

	t.Run("Format normalizer", func(t *testing.T) {
		formatRule := flagValidationRules[3] // Format rule
		assert.Equal(t, "html", formatRule.Normalizer("HTML"))
		assert.Equal(t, "json", formatRule.Normalizer("JSON"))
		assert.Equal(t, "html", formatRule.Normalizer("Html"))
	})
}
