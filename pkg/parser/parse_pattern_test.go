package parser

import (
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// ParsePatternSuite covers the name-grouping helpers and GroupAxes.
type ParsePatternSuite struct {
	suite.Suite
}

func (s *ParsePatternSuite) TestParseNameToGroups() {
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
			expected:      map[string]string{"name": "Rivet", "yAxis": "GPlusStatic", "xAxis": "", "zAxis": ""},
		},
		{
			name:          "Pattern Match: name/yAxis/xAxis pattern",
			benchmarkName: "Rivet/GPlusStatic/100k",
			pattern:       "name/yAxis/xAxis",
			expected:      map[string]string{"name": "Rivet", "yAxis": "GPlusStatic", "xAxis": "100k", "zAxis": ""},
		},
		{
			name:          "Pattern Match: yAxis/name/xAxis pattern",
			benchmarkName: "Rivet/GPlusStatic/100k",
			pattern:       "yAxis/name/xAxis",
			expected:      map[string]string{"name": "GPlusStatic", "yAxis": "Rivet", "xAxis": "100k", "zAxis": ""},
		},
		{
			name:          "Pattern Match: xAxis/yAxis/name pattern",
			benchmarkName: "Rivet/GPlusStatic/100k",
			pattern:       "xAxis/yAxis/name",
			expected:      map[string]string{"name": "100k", "yAxis": "GPlusStatic", "xAxis": "Rivet", "zAxis": ""},
		},
		{
			name:          "Pattern Match: name_yAxis_xAxis pattern",
			benchmarkName: "MyLib_ComplexFunction_TestCase",
			pattern:       "name_yAxis_xAxis",
			expected:      map[string]string{"name": "MyLib", "yAxis": "ComplexFunction", "xAxis": "TestCase", "zAxis": ""},
		},
		{
			name:          "Default behavior: yAxis only pattern",
			benchmarkName: "Rivet_GPlusStatic",
			pattern:       "yAxis",
			expected:      map[string]string{"name": "", "yAxis": "Rivet_GPlusStatic", "xAxis": "", "zAxis": ""},
		},
		{
			name:          "yAxis and xAxis without name: y/x pattern",
			benchmarkName: "Rivet_GPlusStatic/100k",
			pattern:       "yAxis/xAxis",
			expected:      map[string]string{"name": "", "yAxis": "Rivet_GPlusStatic", "xAxis": "100k", "zAxis": ""},
		},
		{
			name:          "Complex name: name_yAxis pattern with multi-part name",
			benchmarkName: "MyLib_ComplexFunction_TestCase",
			pattern:       "name_yAxis",
			expected:      map[string]string{"name": "MyLib", "yAxis": "ComplexFunction_TestCase", "xAxis": "", "zAxis": ""},
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
			expected:      map[string]string{"name": "Rivet", "yAxis": "GPlusStatic", "xAxis": "100k_extra", "zAxis": ""},
		},
		{
			name:          "Skip words",
			benchmarkName: "Tasks/Name/Workload/Subject",
			pattern:       "/name/xAxis/yAxis",
			expected:      map[string]string{"name": "Name", "yAxis": "Subject", "xAxis": "Workload", "zAxis": ""},
		},
		{
			name:          "Not enough parts in benchmark name",
			benchmarkName: "Rivet",
			pattern:       "name_yAxis_xAxis",
			expected:      map[string]string{"name": "Rivet", "yAxis": "", "xAxis": "", "zAxis": ""},
		},
		{
			name:          "Pattern Match: name/x/y/z pattern (3D)",
			benchmarkName: "Sort/bubble/1000/threads8",
			pattern:       "name/x/y/z",
			expected:      map[string]string{"name": "Sort", "xAxis": "bubble", "yAxis": "1000", "zAxis": "threads8"},
		},
		{
			name:          "Pattern Match: space-separated x n y",
			benchmarkName: "alpha beta gamma",
			pattern:       "x n y",
			expected:      map[string]string{"name": "beta", "xAxis": "alpha", "yAxis": "gamma", "zAxis": ""},
		},
		{
			name:          "Pattern Match: comma-separated y,x",
			benchmarkName: "USA,Widget",
			pattern:       "y,x",
			expected:      map[string]string{"name": "", "xAxis": "Widget", "yAxis": "USA", "zAxis": ""},
		},
		{
			name:          "Pattern Match: curly labels stripped for splitting",
			benchmarkName: "2022-2-30",
			pattern:       "n{year}-y{months}-x{dates}",
			expected:      map[string]string{"name": "2022", "yAxis": "2", "xAxis": "30", "zAxis": ""},
		},
		{
			name:          "Pattern Match: trailing separator drops extra segment",
			benchmarkName: "2024-01-01",
			pattern:       "n{Year}-x{Month}-",
			expected:      map[string]string{"name": "2024", "xAxis": "01", "yAxis": "", "zAxis": ""},
		},
		{
			name:          "Pattern Match: consecutive separators skip middle segment",
			benchmarkName: "2024-01-01",
			pattern:       "n{Year}--x{Date}",
			expected:      map[string]string{"name": "2024", "xAxis": "01", "yAxis": "", "zAxis": ""},
		},
		{
			name:          "Pattern Match: consecutive slash separators skip middle segment",
			benchmarkName: "alpha/beta/gamma",
			pattern:       "n//y",
			expected:      map[string]string{"name": "alpha", "yAxis": "gamma", "xAxis": "", "zAxis": ""},
		},
		{
			name:          "Pattern Match: mixed comma and slash x,y/z",
			benchmarkName: "USA,Widget/EU",
			pattern:       "x,y/z",
			expected:      map[string]string{"name": "", "xAxis": "USA", "yAxis": "Widget", "zAxis": "EU"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := ParseBenchmarkNameToGroups(tt.benchmarkName, tt.pattern)

			if tt.expectError {
				s.Require().Error(err)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
				return
			}

			s.Require().NoError(err)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *ParsePatternSuite) TestValidateGroupPattern() {
	tests := []struct {
		name          string
		pattern       string
		expectError   bool
		errorContains string
	}{
		{name: "Valid pattern: name_yAxis", pattern: "name_yAxis"},
		{name: "Valid pattern with shorthand: n_y/x", pattern: "n_y/x"},
		{name: "Valid pattern: yAxis only", pattern: "yAxis"},
		{name: "Invalid pattern: unknown part", pattern: "name_invalid", expectError: true, errorContains: "Invalid part:"},
		{name: "Empty pattern", pattern: "", expectError: true, errorContains: "pattern cannot be empty"},
		{name: "Valid pattern: all parts", pattern: "name_yAxis_xAxis"},
		{name: "Valid pattern: xAxis without yAxis", pattern: "name_xAxis"},
		{name: "Invalid pattern: name only (missing xAxis and yAxis)", pattern: "name", expectError: true, errorContains: "pattern must contain xAxis (x) or yAxis (y)"},
		{name: "Valid pattern: xAxis only", pattern: "xAxis"},
		{name: "Valid pattern: name/x/y/z (3D)", pattern: "name/x/y/z"},
		{name: "Valid pattern: full zAxis token", pattern: "name/xAxis/yAxis/zAxis"},
		{name: "Invalid pattern: z with x but no y", pattern: "x/z", expectError: true, errorContains: "zAxis (z) requires both xAxis (x) and yAxis (y)"},
		{name: "Invalid pattern: z with y but no x", pattern: "n/y/z", expectError: true, errorContains: "zAxis (z) requires both xAxis (x) and yAxis (y)"},
		{name: "Invalid pattern: duplicate dimension x/y/x", pattern: "x/y/x", expectError: true, errorContains: "duplicate dimension 'xAxis' in pattern"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := ValidateGroupPattern(tt.pattern)

			if tt.expectError {
				s.Require().Error(err)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
				return
			}

			s.NoError(err)
		})
	}
}

func (s *ParsePatternSuite) TestPatternHasBothXY() {
	tests := []struct {
		name    string
		cfg     Config
		resolve bool
		want    bool
	}{
		{name: "x only", cfg: Config{GroupPattern: "x"}, want: false},
		{name: "y only", cfg: Config{GroupPattern: "y"}, want: false},
		{name: "n/x/y", cfg: Config{GroupPattern: "n/x/y"}, want: true},
		{name: "n/x/y/z", cfg: Config{GroupPattern: "n/x/y/z"}, want: true},
		{
			name: "regex x and y",
			cfg:  Config{GroupRegex: `(?<n>.*)/(?<x>\d+)/(?<y>.*)`},
			want: true,
		},
		{
			name: "regex x only",
			cfg:  Config{GroupRegex: `(?<n>.*)/(?<x>\d+)`},
			want: false,
		},
		{
			name:    "tabular y,x",
			cfg:     Config{GroupPattern: "y,x"},
			resolve: true,
			want:    true,
		},
		{
			name:    "tabular y only",
			cfg:     Config{GroupPattern: "y"},
			resolve: true,
			want:    false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			cfg := tt.cfg
			if tt.resolve {
				var err error
				cfg, err = ResolveGroupConfig(cfg)
				s.Require().NoError(err)
			}
			s.Equal(tt.want, PatternHasBothXY(cfg))
		})
	}
}

func (s *ParsePatternSuite) TestExpandShorthand() {
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
		s.Run(tt.input, func() {
			s.Equal(tt.expected, expandShorthand(tt.input))
		})
	}
}

func (s *ParsePatternSuite) TestParseNameWithRegex() {
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
			expected:      map[string]string{"name": "Hashing64", "yAxis": "MD5", "xAxis": "", "zAxis": ""},
		},
		{
			name:          "Valid Regex: All Groups",
			benchmarkName: "Matrix/1024/Parallel",
			pattern:       `(?<n>.*)/(?<x>\d+)/(?<y>.*)`,
			expected:      map[string]string{"name": "Matrix", "xAxis": "1024", "yAxis": "Parallel", "zAxis": ""},
		},
		{
			name:          "Valid Regex: Named Groups 2",
			benchmarkName: "Decode/text=digits/level=speed",
			pattern:       `(?<n>.*)/text=(?<x>.*)/level=(?<y>.*)`,
			expected:      map[string]string{"name": "Decode", "xAxis": "digits", "yAxis": "speed", "zAxis": ""},
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
			expected:      map[string]string{"name": "", "xAxis": "", "yAxis": "MD5", "zAxis": ""},
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
			expected:      map[string]string{"name": "TestFunc", "xAxis": "1024", "yAxis": "", "zAxis": ""},
		},
		{
			name:          "Regex with all four groups (3D)",
			benchmarkName: "Sort/bubble/1000/threads8",
			pattern:       `(?<n>.*)/(?<x>.*)/(?<y>.*)/(?<z>.*)`,
			expected:      map[string]string{"name": "Sort", "xAxis": "bubble", "yAxis": "1000", "zAxis": "threads8"},
		},
		{
			name:          "Regex with no named groups at all",
			benchmarkName: "TestFunc/1024",
			pattern:       `(.*)/(\d+)`,
			expected:      nil,
			expectError:   true,
			errorContains: "does not contain x (xAxis) or y (yAxis)",
		},
		{
			name:          "Regex with zAxis but missing yAxis",
			benchmarkName: "Sort/bubble/threads8",
			pattern:       `(?<n>.*)/(?<x>.*)/(?<z>.*)`,
			expected:      nil,
			expectError:   true,
			errorContains: "z requires both xAxis (x) and yAxis (y)",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := ParseBenchmarkNameWithRegex(tt.benchmarkName, tt.pattern)

			if tt.expectError {
				s.Require().Error(err)
				if tt.errorContains != "" {
					s.Contains(err.Error(), tt.errorContains)
				}
				return
			}

			s.Require().NoError(err)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *ParsePatternSuite) TestGroupAxes() {
	tests := []struct {
		name        string
		pattern     string
		regex       string
		groupLabels []string
		want        []shared.Axis
	}{
		{
			name:        "pattern with labels",
			pattern:     "n_x_y",
			groupLabels: []string{"Impl", "Size", "ns/op"},
			want: []shared.Axis{
				{Key: "name", Label: "Impl"},
				{Key: "x", Label: "Size"},
				{Key: "y", Label: "ns/op"},
			},
		},
		{
			name:    "pattern without labels",
			pattern: "x_y",
			want: []shared.Axis{
				{Key: "x", Label: ""},
				{Key: "y", Label: ""},
			},
		},
		{
			name:        "pattern with partial labels",
			pattern:     "n_x_y",
			groupLabels: []string{"Impl"},
			want: []shared.Axis{
				{Key: "name", Label: "Impl"},
				{Key: "x", Label: ""},
				{Key: "y", Label: ""},
			},
		},
		{
			name:  "regex mode - 2D",
			regex: `(?P<x>[^/]+)/(?P<y>.+)`,
			want: []shared.Axis{
				{Key: "x", Label: ""},
				{Key: "y", Label: ""},
			},
		},
		{
			name:  "regex mode - 3D with name",
			regex: `(?P<n>[^/]+)/(?P<x>[^/]+)/(?P<y>[^/]+)/(?P<z>.+)`,
			want: []shared.Axis{
				{Key: "name", Label: ""},
				{Key: "x", Label: ""},
				{Key: "y", Label: ""},
				{Key: "z", Label: ""},
			},
		},
		{
			name: "no grouping",
			want: nil,
		},
		{
			name:        "pattern y/x with labels: order follows pattern",
			pattern:     "y/x",
			groupLabels: []string{"region", "product"},
			want: []shared.Axis{
				{Key: "y", Label: "region"},
				{Key: "x", Label: "product"},
			},
		},
		{
			name:        "blank group entries skipped, align with pattern",
			pattern:     "y/x",
			groupLabels: []string{"", "region", " ", "product"},
			want: []shared.Axis{
				{Key: "y", Label: "region"},
				{Key: "x", Label: "product"},
			},
		},
		{
			name:        "more group labels than pattern parts: extras ignored",
			pattern:     "y/x",
			groupLabels: []string{"region", "product", "extra"},
			want: []shared.Axis{
				{Key: "y", Label: "region"},
				{Key: "x", Label: "product"},
			},
		},
		{
			name:        "regex mode ignores group labels",
			regex:       `(?<y>.*)/(?<x>.*)`,
			groupLabels: []string{"region", "product"},
			want: []shared.Axis{
				{Key: "x", Label: ""},
				{Key: "y", Label: ""},
			},
		},
		{
			name:        "empty part in pattern does not shift group label index",
			pattern:     "x//y",
			groupLabels: []string{"A", "B"},
			want: []shared.Axis{
				{Key: "x", Label: "A"},
				{Key: "y", Label: "B"},
			},
		},
		{
			name:        "curly labels override group labels",
			pattern:     "n{year}_y{months}_x{dates}",
			groupLabels: []string{"order_date", "order_date", "order_date"},
			want: []shared.Axis{
				{Key: "name", Label: "year"},
				{Key: "y", Label: "months"},
				{Key: "x", Label: "dates"},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			cfg := Config{
				Group:        tt.groupLabels,
				GroupPattern: tt.pattern,
				GroupRegex:   tt.regex,
			}
			s.Equal(tt.want, GroupAxes(cfg))
		})
	}
}

func (s *ParsePatternSuite) TestGroupBenchmarkNameRegex() {
	cfg := Config{GroupRegex: `(?P<name>[^/]+)/(?P<xAxis>[^/]+)`}
	got, err := GroupBenchmarkName("alpha/2024", cfg)
	s.Require().NoError(err)
	s.Equal("alpha", got["name"])
	s.Equal("2024", got["xAxis"])
}

func (s *ParsePatternSuite) TestGroupBenchmarkNamePatternFallback() {
	cfg := Config{GroupPattern: "name/xAxis"}
	got, err := GroupBenchmarkName("alpha/2024", cfg)
	s.Require().NoError(err)
	s.Equal("alpha", got["name"])
	s.Equal("2024", got["xAxis"])
}

func TestParsePatternSuite(t *testing.T) {
	suite.Run(t, new(ParsePatternSuite))
}
