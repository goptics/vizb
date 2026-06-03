package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				"zAxis": "",
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
				"zAxis": "",
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
				"zAxis": "",
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
				"zAxis": "",
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
				"zAxis": "",
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
				"zAxis": "",
			},
			expectError: false,
		},
		{
			name:          "yAxis and xAxis without name: y/x pattern",
			benchmarkName: "Rivet_GPlusStatic/100k",
			pattern:       "yAxis/xAxis",
			expected: map[string]string{
				"name":  "",
				"yAxis": "Rivet_GPlusStatic",
				"xAxis": "100k",
				"zAxis": "",
			},
			expectError: false,
		},
		{
			name:          "Complex name: name_yAxis pattern with multi-part name",
			benchmarkName: "MyLib_ComplexFunction_TestCase",
			pattern:       "name_yAxis",
			expected: map[string]string{
				"name":  "MyLib",
				"yAxis": "ComplexFunction_TestCase",
				"xAxis": "",
				"zAxis": "",
			},
			expectError: false,
		},
		{
			name:          "Empty pattern",
			benchmarkName: "Rivet",
			pattern:       "",
			expected:      nil,
			expectError:   true,
			errorContains: "pattern cannot be empty",
		},
		{
			name:          "Mixed separators: underscore and slash",
			benchmarkName: "Rivet_GPlusStatic/100k_extra",
			pattern:       "name_yAxis/xAxis",
			expected: map[string]string{
				"name":  "Rivet",
				"yAxis": "GPlusStatic",
				"xAxis": "100k_extra",
				"zAxis": "",
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
				"zAxis": "",
			},
		},
		{
			name:          "Not enough parts in benchmark name",
			benchmarkName: "Rivet",
			pattern:       "name_yAxis_xAxis",
			expected: map[string]string{
				"name":  "Rivet",
				"yAxis": "",
				"xAxis": "",
				"zAxis": "",
			},
			expectError: false,
		},
		{
			name:          "Pattern Match: name/x/y/z pattern (3D)",
			benchmarkName: "Sort/bubble/1000/threads8",
			pattern:       "name/x/y/z",
			expected: map[string]string{
				"name":  "Sort",
				"xAxis": "bubble",
				"yAxis": "1000",
				"zAxis": "threads8",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseBenchmarkNameToGroups(tt.benchmarkName, tt.pattern)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateGroupPattern(t *testing.T) {
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
			name:        "Valid pattern: xAxis without yAxis",
			pattern:     "name_xAxis",
			expectError: false,
		},
		{
			name:          "Invalid pattern: name only (missing xAxis and yAxis)",
			pattern:       "name",
			expectError:   true,
			errorContains: "pattern must contain xAxis (x) or yAxis (y)",
		},
		{
			name:        "Valid pattern: xAxis only",
			pattern:     "xAxis",
			expectError: false,
		},
		{
			name:        "Valid pattern: name/x/y/z (3D)",
			pattern:     "name/x/y/z",
			expectError: false,
		},
		{
			name:        "Valid pattern: full zAxis token",
			pattern:     "name/xAxis/yAxis/zAxis",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGroupPattern(tt.pattern)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			assert.NoError(t, err)
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
		{"z", "zAxis"},
		{"name", "name"},
		{"yAxis", "yAxis"},
		{"xAxis", "xAxis"},
		{"zAxis", "zAxis"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := expandShorthand(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseNameWithRegex(t *testing.T) {
	tests := []struct {
		name          string
		benchmarkName string
		pattern       string
		expected      map[string]string
		expectError   bool
		errorContains string
	}{
		{
			name:          "Valid Regex: Named Groups",
			benchmarkName: "Hashing64MD5",
			pattern:       `(?<n>Hashing64)(?<y>.*)`,
			expected: map[string]string{
				"name":  "Hashing64",
				"yAxis": "MD5",
				"xAxis": "",
				"zAxis": "",
			},
			expectError: false,
		},
		{
			name:          "Valid Regex: All Groups",
			benchmarkName: "Matrix/1024/Parallel",
			pattern:       `(?<n>.*)/(?<x>\d+)/(?<y>.*)`,
			expected: map[string]string{
				"name":  "Matrix",
				"xAxis": "1024",
				"yAxis": "Parallel",
				"zAxis": "",
			},
			expectError: false,
		},
		{
			name:          "Valid Regex: Named Groups 2",
			benchmarkName: "Decode/text=digits/level=speed",
			pattern:       `(?<n>.*)/text=(?<x>.*)/level=(?<y>.*)`,
			expected: map[string]string{
				"name":  "Decode",
				"xAxis": "digits",
				"yAxis": "speed",
				"zAxis": "",
			},
			expectError: false,
		},
		{
			name:          "Regex No Match",
			benchmarkName: "Hashing64MD5",
			pattern:       `(?<n>Sorting)(?<y>.*)`,
			expected:      nil,
			expectError:   true,
			errorContains: "does not match regex",
		},
		{
			name:          "Invalid Regex Syntax",
			benchmarkName: "Hashing64MD5",
			pattern:       `(?<n>Hashing64)(?<y>.*`,
			expected:      nil,
			expectError:   true,
			errorContains: "invalid regex pattern",
		},
		{
			name:          "Regex with non-capturing groups (ignored)",
			benchmarkName: "Hashing64MD5",
			pattern:       `(?:Hashing64)(?<y>.*)`,
			expected: map[string]string{
				"name":  "",
				"xAxis": "",
				"yAxis": "MD5",
				"zAxis": "",
			},
			expectError: false,
		},
		{
			name:          "Regex with only name capture (missing xAxis and yAxis)",
			benchmarkName: "Hashing64MD5",
			pattern:       `(?<n>Hashing64)(?<name>.*)`,
			expected:      nil,
			expectError:   true,
			errorContains: "does not contain x (xAxis) or y (yAxis)",
		},
		{
			name:          "Regex with only xAxis (valid)",
			benchmarkName: "TestFunc/1024",
			pattern:       `(?<n>.*)/(?<x>\d+)`,
			expected: map[string]string{
				"name":  "TestFunc",
				"xAxis": "1024",
				"yAxis": "",
				"zAxis": "",
			},
			expectError: false,
		},
		{
			name:          "Regex with all four groups (3D)",
			benchmarkName: "Sort/bubble/1000/threads8",
			pattern:       `(?<n>.*)/(?<x>.*)/(?<y>.*)/(?<z>.*)`,
			expected: map[string]string{
				"name":  "Sort",
				"xAxis": "bubble",
				"yAxis": "1000",
				"zAxis": "threads8",
			},
			expectError: false,
		},
		{
			name:          "Regex with no named groups at all",
			benchmarkName: "TestFunc/1024",
			pattern:       `(.*)/(\d+)`,
			expected:      nil,
			expectError:   true,
			errorContains: "does not contain x (xAxis) or y (yAxis)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseBenchmarkNameWithRegex(tt.benchmarkName, tt.pattern)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
