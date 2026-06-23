package parser

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GroupSpecSuite struct {
	suite.Suite
}

func (s *GroupSpecSuite) TestParseGroupSpecFlatSlice() {
	spec, err := parseGroupSpec([]string{"name", "category", "region"}, "", []string{",", ","})
	s.Require().NoError(err)
	s.Equal([]string{"name", "category", "region"}, spec.Columns)
	s.False(spec.Structured)
}

func (s *GroupSpecSuite) TestParseGroupSpecSpaceSingleValue() {
	spec, err := parseGroupSpec([]string{"name category region"}, "x n y", []string{" ", " "})
	s.Require().NoError(err)
	s.Equal([]string{"name", "category", "region"}, spec.Columns)
	s.Equal([]string{" ", " "}, spec.Separators)
}

func (s *GroupSpecSuite) TestParseGroupSpecStructured() {
	spec, err := parseGroupSpec([]string{"name", "category/region"}, "", nil)
	s.Require().NoError(err)
	s.True(spec.Structured)
	s.Equal([]string{"name", "category", "region"}, spec.Columns)
	s.Equal([]string{",", "/"}, spec.Separators)
}

func (s *GroupSpecSuite) TestResolveGroupConfigSpacePattern() {
	cfg, err := ResolveGroupConfig(Config{
		Group:        []string{"name category region"},
		GroupPattern: "x n y",
	})
	s.Require().NoError(err)
	s.Equal([]string{"name", "category", "region"}, cfg.GroupColumns)
	s.Equal([]string{" ", " "}, cfg.LabelSeparators)
}

func (s *GroupSpecSuite) TestResolveGroupConfigCommaPatternFlat() {
	cfg, err := ResolveGroupConfig(Config{
		Group:        []string{"region", "product"},
		GroupPattern: "y,x",
	})
	s.Require().NoError(err)
	s.Equal([]string{"region", "product"}, cfg.GroupColumns)
	s.Equal([]string{","}, cfg.LabelSeparators)
}

func (s *GroupSpecSuite) TestGroupPatternSeparatorMismatch() {
	cases := []struct {
		name    string
		group   []string
		pattern string
		want    string
	}{
		{
			name:    "comma_group_slash_pattern",
			group:   []string{"product", "category", "region"},
			pattern: "x/y/z",
			want:    `--group "product,category,region" and --group-pattern "x/y/z" separators do not match (expected ", ,", got "/ /")`,
		},
		{
			name:    "slash_group_mixed_pattern",
			group:   []string{"product/category/region"},
			pattern: "x,y/z",
			want:    `--group "product/category/region" and --group-pattern "x,y/z" separators do not match (expected "/ /", got ", /")`,
		},
		{
			name:    "hash_group_mixed_pattern",
			group:   []string{"product#category#region"},
			pattern: "x#y,z",
			want:    `--group "product#category#region" and --group-pattern "x#y,z" separators do not match (expected "# #", got "# ,")`,
		},
		{
			name:    "structured_multi_arg",
			group:   []string{"name", "category/region"},
			pattern: "x/y/z",
			want:    `--group "name,category/region" and --group-pattern "x/y/z" separators do not match (expected ", /", got "/ /")`,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			cfg, err := ResolveGroupConfig(Config{
				Group:        tc.group,
				GroupPattern: tc.pattern,
			})
			if err == nil {
				err = ValidateTabularGroupAlignment(cfg)
			}
			s.Require().Error(err)
			s.Equal(tc.want, err.Error())
		})
	}
}

func (s *GroupSpecSuite) TestValidateTabularGroupAcceptsCommaPattern() {
	cfg, err := ResolveGroupConfig(Config{
		Group:        []string{"region", "product", "month"},
		GroupPattern: "x,y,z",
	})
	s.Require().NoError(err)
	s.Require().NoError(ValidateTabularGroupAlignment(cfg))
}

func (s *GroupSpecSuite) TestJoinLabelPartsSpaces() {
	got := JoinLabelParts([]string{"alpha", "beta", "gamma"}, []string{" ", " "})
	s.Equal("alpha beta gamma", got)
}

func (s *GroupSpecSuite) TestParseBenchmarkNameSpacePattern() {
	got, err := ParseBenchmarkNameToGroups("alpha beta gamma", "x n y")
	s.Require().NoError(err)
	s.Equal("alpha", got["xAxis"])
	s.Equal("beta", got["name"])
	s.Equal("gamma", got["yAxis"])
}

func TestGroupSpecSuite(t *testing.T) {
	suite.Run(t, new(GroupSpecSuite))
}

type AutoGroupColumnsSuite struct {
	suite.Suite
}

func (s *AutoGroupColumnsSuite) TestNonNumericPickedOverNumeric() {
	headers := []string{"region", "sells", "stocks"}
	rows := [][]string{
		{"West", "10", "5"},
		{"East", "20", "7"},
		{"West", "30", "9"},
	}
	cols, pattern, ok := AutoGroupColumns(headers, rows, false)
	s.Require().True(ok)
	s.Equal([]string{"region"}, cols)
	s.Equal("x", pattern)
}

func (s *AutoGroupColumnsSuite) TestAllNumericNoAutoGroup() {
	// Numeric columns are completely ignored; no categorical column → ok=false
	headers := []string{"id", "sells", "stocks"}
	rows := [][]string{
		{"1", "10", "5"},
		{"2", "20", "7"},
		{"3", "30", "9"},
		{"4", "40", "11"},
	}
	cols, pattern, ok := AutoGroupColumns(headers, rows, false)
	s.False(ok)
	s.Empty(cols)
	s.Empty(pattern)
}

func (s *AutoGroupColumnsSuite) TestHighestCardinalityWins() {
	headers := []string{"region", "product"}
	rows := [][]string{
		{"West", "A"},
		{"East", "B"},
		{"North", "C"},
		{"South", "D"},
		{"Central", "E"},
		{"West", "F"},
		{"East", "G"},
		{"North", "H"},
		{"South", "I"},
		{"Central", "J"},
		{"West", "K"},
		{"East", "L"},
	}
	// region has 5 distinct, product has 12 distinct → product wins
	cols, _, ok := AutoGroupColumns(headers, rows, false)
	s.Require().True(ok)
	s.Equal([]string{"product"}, cols)
}

func (s *AutoGroupColumnsSuite) TestLeftmostTieBreak() {
	headers := []string{"region", "product"}
	rows := [][]string{
		{"West", "A"},
		{"East", "B"},
		{"North", "C"},
	}
	// both have 3 distinct, equal → leftmost (region) wins
	cols, _, ok := AutoGroupColumns(headers, rows, false)
	s.Require().True(ok)
	s.Equal([]string{"region"}, cols)
}

func (s *AutoGroupColumnsSuite) TestWantXYPicksTwo() {
	headers := []string{"region", "product", "sells", "stocks"}
	rows := [][]string{
		{"West", "A", "10", "5"},
		{"East", "B", "20", "7"},
		{"North", "C", "30", "9"},
		{"South", "D", "40", "11"},
		{"Central", "E", "50", "13"},
		{"West", "F", "60", "15"},
		{"East", "G", "70", "17"},
		{"North", "H", "80", "19"},
		{"South", "I", "90", "21"},
		{"Central", "J", "100", "23"},
		{"West", "K", "110", "25"},
		{"East", "L", "120", "27"},
	}
	cols, pattern, ok := AutoGroupColumns(headers, rows, true)
	s.Require().True(ok)
	// product 12 distinct > region 5 → xAxis=product, yAxis=region
	s.Equal([]string{"product", "region"}, cols)
	s.Equal("x,y", pattern)
}

func (s *AutoGroupColumnsSuite) TestWantXYWithOneCandidatePicksOne() {
	headers := []string{"region", "sells", "stocks"}
	rows := [][]string{
		{"West", "10", "5"},
		{"East", "20", "7"},
		{"North", "30", "9"},
	}
	cols, pattern, ok := AutoGroupColumns(headers, rows, true)
	s.Require().True(ok)
	s.Equal([]string{"region"}, cols)
	s.Equal("x", pattern)
}

func (s *AutoGroupColumnsSuite) TestSingleColumnNoOp() {
	headers := []string{"sells"}
	rows := [][]string{{"10"}, {"20"}}
	cols, pattern, ok := AutoGroupColumns(headers, rows, false)
	s.False(ok)
	s.Empty(cols)
	s.Empty(pattern)
}

func (s *AutoGroupColumnsSuite) TestNoChartColumnRemainsNoOp() {
	// Two columns: picking one as axis leaves only one "chart" candidate,
	// but that candidate is the other column itself and is consumed... no:
	// header count is 2, so chartColsAfter = 1 (ok, ONE remaining numeric).
	// Use one column to test the "<2 headers" path (true no-op).
	headers := []string{"region"}
	rows := [][]string{{"West"}, {"East"}}
	cols, _, ok := AutoGroupColumns(headers, rows, false)
	s.False(ok)
	s.Empty(cols)
}

func (s *AutoGroupColumnsSuite) TestNumericStringValuesClassifiedNumeric() {
	// A non-numeric cell ("West") makes the column categorical even if most
	// cells are numeric strings.
	headers := []string{"mix", "sells"}
	rows := [][]string{
		{"West", "10"},
		{"20", "30"},
		{"30", "40"},
	}
	cols, _, ok := AutoGroupColumns(headers, rows, false)
	s.Require().True(ok)
	s.Equal([]string{"mix"}, cols) // mix is categorical (has "West")
}

func TestAutoGroupColumnsSuite(t *testing.T) {
	suite.Run(t, new(AutoGroupColumnsSuite))
}

type AutoValueColumnsSuite struct {
	suite.Suite
}

func (s *AutoValueColumnsSuite) TestThreeNumericCols() {
	headers := []string{"price", "latency", "memory"}
	rows := [][]string{
		{"10", "5", "100"},
		{"20", "7", "200"},
		{"30", "9", "300"},
	}
	cols, ok := AutoValueColumns(headers, rows)
	s.Require().True(ok)
	s.Equal([]string{"price", "latency", "memory"}, cols)
}

func (s *AutoValueColumnsSuite) TestTwoNumericCols() {
	headers := []string{"price", "latency"}
	rows := [][]string{
		{"10", "5"},
		{"20", "7"},
	}
	cols, ok := AutoValueColumns(headers, rows)
	s.Require().True(ok)
	s.Equal([]string{"price", "latency"}, cols)
}

func (s *AutoValueColumnsSuite) TestOneNumericColReturnsFalse() {
	headers := []string{"price", "region"}
	rows := [][]string{
		{"10", "West"},
		{"20", "East"},
	}
	cols, ok := AutoValueColumns(headers, rows)
	s.False(ok)
	s.Empty(cols)
}

func (s *AutoValueColumnsSuite) TestAllNonNumericReturnsFalse() {
	headers := []string{"region", "product"}
	rows := [][]string{
		{"West", "A"},
		{"East", "B"},
	}
	cols, ok := AutoValueColumns(headers, rows)
	s.False(ok)
	s.Empty(cols)
}

func (s *AutoValueColumnsSuite) TestFourNumericColsReturnsFirstThree() {
	headers := []string{"a", "b", "c", "d"}
	rows := [][]string{
		{"1", "2", "3", "4"},
		{"5", "6", "7", "8"},
	}
	cols, ok := AutoValueColumns(headers, rows)
	s.Require().True(ok)
	s.Equal([]string{"a", "b", "c"}, cols)
}

func (s *AutoValueColumnsSuite) TestMixedTypesSkipsNonNumeric() {
	headers := []string{"region", "price", "product", "latency"}
	rows := [][]string{
		{"West", "10", "foo", "5"},
		{"East", "20", "bar", "7"},
	}
	// region (non-numeric) skipped, price (numeric) kept, product (non-numeric) skipped, latency (numeric) kept
	cols, ok := AutoValueColumns(headers, rows)
	s.Require().True(ok)
	s.Equal([]string{"price", "latency"}, cols)
}

func (s *AutoValueColumnsSuite) TestEmptyHeaderSkipped() {
	headers := []string{"price", "", "latency"}
	rows := [][]string{
		{"10", "x", "5"},
		{"20", "y", "7"},
	}
	cols, ok := AutoValueColumns(headers, rows)
	s.Require().True(ok)
	s.Equal([]string{"price", "latency"}, cols)
}

func (s *AutoValueColumnsSuite) TestNumericStringValuesClassifiedNumeric() {
	headers := []string{"price", "count"}
	rows := [][]string{
		{"10.5", "100"},
		{"20.0", "200"},
	}
	cols, ok := AutoValueColumns(headers, rows)
	s.Require().True(ok)
	s.Equal([]string{"price", "count"}, cols)
}

func (s *AutoValueColumnsSuite) TestSingleColumnOnlyReturnsFalse() {
	headers := []string{"price"}
	rows := [][]string{
		{"10"},
		{"20"},
	}
	cols, ok := AutoValueColumns(headers, rows)
	s.False(ok)
	s.Empty(cols)
}

func TestAutoValueColumnsSuite(t *testing.T) {
	suite.Run(t, new(AutoValueColumnsSuite))
}

func TestAutoValueEligible(t *testing.T) {
	tests := []struct {
		name  string
		types []string
		want  bool
	}{
		{"scatter only", []string{"scatter"}, true},
		{"bar only", []string{"bar"}, true},
		{"line only", []string{"line"}, true},
		{"pie only", []string{"pie"}, false},
		{"heatmap only", []string{"heatmap"}, false},
		{"radar only", []string{"radar"}, false},
		{"mixed with eligible", []string{"pie", "bar", "radar"}, true},
		{"mixed without eligible", []string{"pie", "heatmap", "radar"}, false},
		{"empty slice", []string{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AutoValueEligible(tt.types)
			if got != tt.want {
				t.Errorf("AutoValueEligible(%v) = %v, want %v", tt.types, got, tt.want)
			}
		})
	}
}
