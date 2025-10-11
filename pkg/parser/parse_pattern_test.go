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
				"name":     "Rivet",
				"subject":  "GPlusStatic",
				"workload": "",
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
				"name":     "",
				"subject":  "Rivet_GPlusStatic",
				"workload": "",
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
				"name":     "MyLib",
				"subject":  "ComplexFunction_TestCase",
				"workload": "",
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
		{
			name:          "Skip words from first",
			benchmarkName: "Tasks/Name/Workload/Subject",
			pattern:       "/name/workload/subject",
			expected: map[string]string{
				"name":     "Name",
				"subject":  "Subject",
				"workload": "Workload",
			},
		},
		{
			name:          "Skip words from middle",
			benchmarkName: "Tasks/Name/Workload/Subject",
			pattern:       "/name//subject",
			expected: map[string]string{
				"name":     "Name",
				"subject":  "Subject",
				"workload": "",
			},
		},
		// Not enough parts in benchmark name
		{
			name:          "Not enough parts in benchmark name",
			benchmarkName: "Rivet",
			pattern:       "name_subject_workload",
			expected: map[string]string{
				"name":     "Rivet",
				"subject":  "",
				"workload": "",
			},
			expectError: false,
		},
		// Square bracket patterns for PascalCase splitting
		{
			name:          "Simple square bracket pattern",
			benchmarkName: "SubjectWorkloadName",
			pattern:       "[s,w,n]",
			expected: map[string]string{
				"subject":  "Subject",
				"workload": "Workload",
				"name":     "Name",
			},
			expectError: false,
		},
		{
			name:          "Mixed separator and square bracket pattern",
			benchmarkName: "Concat/LargeData/StringOps",
			pattern:       "s/[w]/[n]",
			expected: map[string]string{
				"subject":  "Concat",
				"workload": "Large",
				"name":     "String",
			},
			expectError: false,
		},
		{
			name:          "Square bracket pattern with skipping",
			benchmarkName: "BenchmarkConcat/LargeData/StringOps",
			pattern:       "s/[,w]/[,n]",
			expected: map[string]string{
				"subject":  "BenchmarkConcat",
				"workload": "Data",
				"name":     "Ops",
			},
			expectError: false,
		},
		{
			name:          "Square bracket pattern with empty indices",
			benchmarkName: "FirstSecondThirdFourth",
			pattern:       "[,,s]",
			expected: map[string]string{
				"subject":  "Third",
				"workload": "",
				"name":     "",
			},
			expectError: false,
		},
		{
			name:          "Complex PascalCase benchmark name",
			benchmarkName: "HTTPServerRequestHandler",
			pattern:       "[s,w,n]",
			expected: map[string]string{
				"subject":  "HTTP",
				"workload": "Server",
				"name":     "Request",
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
		{
			name:          "Invalid pattern: missing subject",
			pattern:       "name_workload",
			expectError:   true,
			errorContains: "pattern must contain subject(s)",
		},
		{
			name:          "Invalid pattern: name only",
			pattern:       "name",
			expectError:   true,
			errorContains: "pattern must contain subject(s)",
		},
		{
			name:          "Invalid pattern: workload only",
			pattern:       "workload",
			expectError:   true,
			errorContains: "pattern must contain subject(s)",
		},
		// Square bracket pattern validation tests
		{
			name:        "Valid square bracket pattern",
			pattern:     "[s,w,n]",
			expectError: false,
		},
		{
			name:        "Valid mixed pattern with square brackets",
			pattern:     "s/[w]/[n]",
			expectError: false,
		},
		{
			name:        "Valid square bracket pattern with skipping",
			pattern:     "[,s,w]",
			expectError: false,
		},
		{
			name:          "Invalid square bracket pattern: unknown part",
			pattern:       "[s,invalid,w]",
			expectError:   true,
			errorContains: "Invalid part in square brackets: 'invalid'",
		},
		{
			name:          "Invalid square bracket pattern: missing subject",
			pattern:       "[n,w]",
			expectError:   true,
			errorContains: "pattern must contain subject(s)",
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
