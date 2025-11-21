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
			name:          "Pattern Match: name_yAxis pattern",
			benchmarkName: "Rivet_GPlusStatic",
			pattern:       "name_yAxis",
			expected: map[string]string{
				"name":  "Rivet",
				"yAxis": "GPlusStatic",
				"xAxis": "",
			},
			expectError: false,
		},
		{
			name:          "Pattern Match: name/yAxis/xAxis pattern",
			benchmarkName: "Rivet/GPlusStatic/100k",
			pattern:       "name/yAxis/xAxis",
			expected: map[string]string{
				"name":  "Rivet",
				"yAxis": "GPlusStatic",
				"xAxis": "100k",
			},
			expectError: false,
		},
		{
			name:          "Pattern Match: yAxis/name/xAxis pattern",
			benchmarkName: "Rivet/GPlusStatic/100k",
			pattern:       "yAxis/name/xAxis",
			expected: map[string]string{
				"name":  "GPlusStatic",
				"yAxis": "Rivet",
				"xAxis": "100k",
			},
			expectError: false,
		},
		{
			name:          "Pattern Match: xAxis/yAxis/name pattern",
			benchmarkName: "Rivet/GPlusStatic/100k",
			pattern:       "xAxis/yAxis/name",
			expected: map[string]string{
				"name":  "100k",
				"yAxis": "GPlusStatic",
				"xAxis": "Rivet",
			},
			expectError: false,
		},
		{
			name:          "Pattern Match: name_yAxis_xAxis pattern",
			benchmarkName: "MyLib_ComplexFunction_TestCase",
			pattern:       "name_yAxis_xAxis",
			expected: map[string]string{
				"name":  "MyLib",
				"yAxis": "ComplexFunction",
				"xAxis": "TestCase",
			},
			expectError: false,
		},
		{
			name:          "Default behavior: yAxis only pattern",
			benchmarkName: "Rivet_GPlusStatic",
			pattern:       "yAxis",
			expected: map[string]string{
				"name":  "",
				"yAxis": "Rivet_GPlusStatic",
				"xAxis": "",
			},
			expectError: false,
		},
		// Subject and workload without name
		{
			name:          "yAxis and xAxis without name: y/x pattern",
			benchmarkName: "Rivet_GPlusStatic/100k",
			pattern:       "yAxis/xAxis",
			expected: map[string]string{
				"name":  "",
				"yAxis": "Rivet_GPlusStatic",
				"xAxis": "100k",
			},
			expectError: false,
		},
		// Complex name with multiple underscores
		{
			name:          "Complex name: name_yAxis pattern with multi-part name",
			benchmarkName: "MyLib_ComplexFunction_TestCase",
			pattern:       "name_yAxis",
			expected: map[string]string{
				"name":  "MyLib",
				"yAxis": "ComplexFunction_TestCase",
				"xAxis": "",
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
			pattern:       "name_yAxis/xAxis",
			expected: map[string]string{
				"name":  "Rivet",
				"yAxis": "GPlusStatic",
				"xAxis": "100k_extra",
			},
			expectError: false,
		},
		{
			name:          "Skip words",
			benchmarkName: "Tasks/Name/Workload/Subject",
			pattern:       "/name/xAxis/yAxis",
			expected: map[string]string{
				"name":  "Name",
				"yAxis": "Subject",
				"xAxis": "Workload",
			},
		},
		// Not enough parts in benchmark name
		{
			name:          "Not enough parts in benchmark name",
			benchmarkName: "Rivet",
			pattern:       "name_yAxis_xAxis",
			expected: map[string]string{
				"name":  "Rivet",
				"yAxis": "",
				"xAxis": "",
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
			name:        "Valid pattern: name_yAxis",
			pattern:     "name_yAxis",
			expectError: false,
		},
		{
			name:        "Valid pattern with shorthand: n_y/x",
			pattern:     "n_y/x",
			expectError: false,
		},
		{
			name:        "Valid pattern: yAxis only",
			pattern:     "yAxis",
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
			pattern:     "name_yAxis_xAxis",
			expectError: false,
		},
		{
			name:        "Invalid pattern: missing yAxis",
			pattern:     "name_xAxis",
			expectError: false, // Requirement removed
		},
		{
			name:        "Invalid pattern: name only",
			pattern:     "name",
			expectError: false, // Requirement removed
		},
		{
			name:        "Invalid pattern: xAxis only",
			pattern:     "xAxis",
			expectError: false, // Requirement removed
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
		{"y", "yAxis"},
		{"x", "xAxis"},
		{"name", "name"},
		{"yAxis", "yAxis"},
		{"xAxis", "xAxis"},
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
