package parser

import (
	"reflect"
	"testing"
)

func TestParseNameToGroups(t *testing.T) {
	tests := []struct {
		name          string
		benchmarkName string
		pattern       string
		expected      map[string]string
		expectError   bool
		errorContains string
	}{
		{
			name:          "Pattern Match: name_subject pattern",
			benchmarkName: "Rivet_GPlusStatic",
			pattern:       "name_subject",
			expected: map[string]string{
				"name":    "Rivet",
				"subject": "GPlusStatic",
			},
			expectError: false,
		},
		{
			name:          "Pattern Match: name/subject/workload pattern",
			benchmarkName: "Rivet/GPlusStatic/100k",
			pattern:       "name/subject/workload",
			expected: map[string]string{
				"name":     "Rivet",
				"subject":  "GPlusStatic",
				"workload": "100k",
			},
			expectError: false,
		},
		{
			name:          "Pattern Match: subject/name/workload pattern",
			benchmarkName: "Rivet/GPlusStatic/100k",
			pattern:       "subject/name/workload",
			expected: map[string]string{
				"name":     "GPlusStatic",
				"subject":  "Rivet",
				"workload": "100k",
			},
			expectError: false,
		},
		{
			name:          "Pattern Match: workload/subject/name pattern",
			benchmarkName: "Rivet/GPlusStatic/100k",
			pattern:       "workload/subject/name",
			expected: map[string]string{
				"name":     "100k",
				"subject":  "GPlusStatic",
				"workload": "Rivet",
			},
			expectError: false,
		},
		{
			name:          "Pattern Match: name_subject_workload pattern",
			benchmarkName: "MyLib_ComplexFunction_TestCase",
			pattern:       "name_subject_workload",
			expected: map[string]string{
				"name":     "MyLib",
				"subject":  "ComplexFunction",
				"workload": "TestCase",
			},
			expectError: false,
		},
		{
			name:          "Default behavior: subject only pattern",
			benchmarkName: "Rivet_GPlusStatic",
			pattern:       "subject",
			expected: map[string]string{
				"subject": "Rivet_GPlusStatic",
			},
			expectError: false,
		},
		// name will be consider as subject if only name found
		{
			name:          "Default behavior: name only pattern",
			benchmarkName: "Rivet",
			pattern:       "name",
			expected: map[string]string{
				"subject": "Rivet",
			},
			expectError: false,
		},
		// Subject and workload without name
		{
			name:          "Subject and workload without name: s/w pattern",
			benchmarkName: "Rivet_GPlusStatic/100k",
			pattern:       "subject/workload",
			expected: map[string]string{
				"name":     "",
				"subject":  "Rivet_GPlusStatic",
				"workload": "100k",
			},
			expectError: false,
		},
		// Complex name with multiple underscores
		{
			name:          "Complex name: name_subject pattern with multi-part name",
			benchmarkName: "MyLib_ComplexFunction_TestCase",
			pattern:       "name_subject",
			expected: map[string]string{
				"name":    "MyLib",
				"subject": "ComplexFunction_TestCase",
			},
			expectError: false,
		},
		// Empty pattern
		{
			name:          "Empty pattern",
			benchmarkName: "Rivet",
			pattern:       "",
			expected:      nil,
			expectError:   true,
			errorContains: "pattern cannot be empty",
		},
		// Mixed separators
		{
			name:          "Mixed separators: underscore and slash",
			benchmarkName: "Rivet_GPlusStatic/100k_extra",
			pattern:       "name_subject/workload",
			expected: map[string]string{
				"name":     "Rivet",
				"subject":  "GPlusStatic",
				"workload": "100k_extra",
			},
			expectError: false,
		},
		// Not enough parts in benchmark name
		{
			name:          "Not enough parts in benchmark name",
			benchmarkName: "Rivet",
			pattern:       "name_subject_workload",
			expected: map[string]string{
				"name": "Rivet",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseBenchmarkNameToGroups(tt.benchmarkName, tt.pattern)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', but got '%s'", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}

func TestValidatePattern(t *testing.T) {
	tests := []struct {
		name          string
		pattern       string
		expectError   bool
		errorContains string
	}{
		{
			name:        "Valid pattern: name_subject",
			pattern:     "name_subject",
			expectError: false,
		},
		{
			name:        "Valid pattern with shorthand: n_s/w",
			pattern:     "n_s/w",
			expectError: false,
		},
		{
			name:        "Valid pattern: subject only",
			pattern:     "subject",
			expectError: false,
		},
		{
			name:          "Invalid pattern: unknown part",
			pattern:       "name_invalid",
			expectError:   true,
			errorContains: "Invalid part: 'invalid'",
		},
		{
			name:          "Empty pattern",
			pattern:       "",
			expectError:   true,
			errorContains: "pattern cannot be empty",
		},
		{
			name:        "Valid pattern: all parts",
			pattern:     "name_subject_workload",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePattern(tt.pattern)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', but got '%s'", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestExpandShorthand(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"n", "name"},
		{"s", "subject"},
		{"w", "workload"},
		{"name", "name"},
		{"subject", "subject"},
		{"workload", "workload"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := expandShorthand(tt.input)
			if result != tt.expected {
				t.Errorf("expandShorthand(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOfSubstring(s, substr) >= 0)))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
