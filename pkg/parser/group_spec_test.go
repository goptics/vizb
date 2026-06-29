package parser

import (
	"strings"
	"testing"

	"github.com/goptics/vizb/testutil"
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
	cols, pattern, ok := AutoGroupColumns(headers, rows)
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
	cols, pattern, ok := AutoGroupColumns(headers, rows)
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
	cols, _, ok := AutoGroupColumns(headers, rows)
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
	cols, _, ok := AutoGroupColumns(headers, rows)
	s.Require().True(ok)
	s.Equal([]string{"region"}, cols)
}

func (s *AutoGroupColumnsSuite) TestSingleColumnNoOp() {
	headers := []string{"sells"}
	rows := [][]string{{"10"}, {"20"}}
	cols, pattern, ok := AutoGroupColumns(headers, rows)
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
	cols, _, ok := AutoGroupColumns(headers, rows)
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
	cols, _, ok := AutoGroupColumns(headers, rows)
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
	all := numericColumns(headers, rows)
	s.Require().GreaterOrEqual(len(all), 4)
	s.Equal("d", all[3])
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

func (s *GroupSpecSuite) TestParseGroupSpecEmptyGroup() {
	spec, err := parseGroupSpec(nil, "x", nil)
	s.Require().NoError(err)
	s.Empty(spec.Columns)
}

func (s *GroupSpecSuite) TestParseGroupSpecSingleValueNoSplit() {
	spec, err := parseGroupSpec([]string{"region"}, "x,y", []string{","})
	s.Require().NoError(err)
	s.Equal([]string{"region"}, spec.Columns)
	s.Nil(spec.Separators)
}

func (s *GroupSpecSuite) TestParseGroupSpecSeparatorMismatch() {
	_, err := parseGroupSpec([]string{"a b"}, "x n y z", []string{" ", " "})
	s.Require().Error(err)
	s.Contains(err.Error(), "separators do not match")
}

func TestNoExplicitGrouping(t *testing.T) {
	if !NoExplicitGrouping(Config{GroupPattern: "x"}) {
		t.Fatal("expected true for default config")
	}
	if NoExplicitGrouping(Config{Group: []string{"region"}}) {
		t.Fatal("expected false when group set")
	}
	if NoExplicitGrouping(Config{GroupRegex: ".*"}) {
		t.Fatal("expected false when group-regex set")
	}
	if NoExplicitGrouping(Config{GroupPattern: "x,y"}) {
		t.Fatal("expected false for custom group-pattern")
	}
	if NoExplicitGrouping(Config{Axes: []ColumnSpec{{Source: "x"}, {Source: "y"}}}) {
		t.Fatal("expected false when axes set")
	}
	if NoExplicitGrouping(Config{Select: []ColumnSpec{{Source: "price"}}}) {
		t.Fatal("expected false when grouped select set")
	}
	if NoExplicitGrouping(Config{SelectViews: []SelectView{{Columns: []ColumnSpec{{Source: "region"}, {Source: "latency"}}}}}) {
		t.Fatal("expected false when solo select views set")
	}
}

func TestAutoGroupApplies(t *testing.T) {
	if !AutoGroupApplies(Config{AutoGroup: true, GroupPattern: "x"}) {
		t.Fatal("expected true")
	}
	if AutoGroupApplies(Config{AutoGroup: false, GroupPattern: "x"}) {
		t.Fatal("expected false when AutoGroup off")
	}
	if AutoGroupApplies(Config{AutoGroup: true, Group: []string{"region"}}) {
		t.Fatal("expected false when explicit group")
	}
}

func TestFilterHeadersForAutoDetect(t *testing.T) {
	headers := []string{"z", "x", "y", "w"}
	got := FilterHeadersForAutoDetect(headers, nil)
	if len(got) != 4 || got[0] != "z" {
		t.Fatalf("want all headers in order, got %v", got)
	}
	got = FilterHeadersForAutoDetect(headers, []ColumnSpec{{Source: "y"}, {Source: "x"}})
	if len(got) != 2 || got[0] != "x" || got[1] != "y" {
		t.Fatalf("want selected headers in file order, got %v", got)
	}
}

func TestEffectiveGroupColumnsFallback(t *testing.T) {
	cfg := Config{Group: []string{" region ", ""}}
	got := EffectiveGroupColumns(cfg)
	if len(got) != 1 || got[0] != "region" {
		t.Fatalf("unexpected columns: %v", got)
	}
}

func TestEffectiveLabelSeparators(t *testing.T) {
	cfg := Config{LabelSeparators: []string{",", "/"}}
	if len(EffectiveLabelSeparators(cfg)) != 2 {
		t.Fatal("expected label separators from config")
	}
}

func TestLogAutoGroupEmptyIsNoOp(t *testing.T) {
	out := testutil.CaptureStdout(func() { LogAutoGroup(nil) })
	if out != "" {
		t.Fatalf("expected no output, got %q", out)
	}
}

func TestLogAutoValueBranches(t *testing.T) {
	out := testutil.CaptureStdout(func() { LogAutoValue([]string{"x", "y"}, "") })
	if !strings.Contains(out, "2D") || !strings.Contains(out, "x, y") {
		t.Fatalf("unexpected 2D log: %q", out)
	}
	out = testutil.CaptureStdout(func() { LogAutoValue([]string{"x", "y", "z"}, "m") })
	if !strings.Contains(out, "3D") || !strings.Contains(out, "metric: m") {
		t.Fatalf("unexpected 3D log: %q", out)
	}
}

type AutoDetectTabularConfigSuite struct {
	suite.Suite
}

func (s *AutoDetectTabularConfigSuite) TestAutoGroupPath() {
	cfg := Config{AutoGroup: true, GroupPattern: "x", ChartTypes: []string{"bar"}}
	headers := []string{"region", "sells"}
	rows := [][]string{{"West", "10"}, {"East", "20"}}

	out := testutil.CaptureStdout(func() {
		got, err := AutoDetectTabularConfig(cfg, headers, rows)
		s.Require().NoError(err)
		s.Equal([]string{"region"}, got.Group)
		s.Equal("x", got.GroupPattern)
	})
	s.Contains(out, "Auto-grouped")
}

func (s *AutoDetectTabularConfigSuite) TestAutoValue2D() {
	cfg := Config{AutoGroup: true, GroupPattern: "x", ChartTypes: []string{"scatter"}}
	headers := []string{"price", "latency"}
	rows := [][]string{{"10", "5"}, {"20", "7"}}

	out := testutil.CaptureStdout(func() {
		got, err := AutoDetectTabularConfig(cfg, headers, rows)
		s.Require().NoError(err)
		s.Len(got.Axes, 2)
		s.Equal("price", got.Axes[0].Source)
		s.Equal("latency", got.Axes[1].Source)
	})
	s.Contains(out, "2D")
}

func (s *AutoDetectTabularConfigSuite) TestAutoValue3DWithMetric() {
	cfg := Config{AutoGroup: true, GroupPattern: "x", ChartTypes: []string{"line"}}
	headers := []string{"a", "b", "c", "d"}
	rows := [][]string{{"1", "2", "3", "4"}}

	out := testutil.CaptureStdout(func() {
		got, err := AutoDetectTabularConfig(cfg, headers, rows)
		s.Require().NoError(err)
		s.Len(got.Axes, 3)
		s.Equal("d", got.MetricColumn)
	})
	s.Contains(out, "3D")
	s.Contains(out, "metric: d")
}

func (s *AutoDetectTabularConfigSuite) TestSkipsWhenAutoGroupOff() {
	cfg := Config{AutoGroup: false, GroupPattern: "x", ChartTypes: []string{"scatter"}}
	got, err := AutoDetectTabularConfig(cfg, []string{"a", "b"}, [][]string{{"1", "2"}})
	s.Require().NoError(err)
	s.Empty(got.Axes)
}

func (s *AutoDetectTabularConfigSuite) TestSkipsWhenPieOnly() {
	cfg := Config{AutoGroup: true, GroupPattern: "x", ChartTypes: []string{"pie"}}
	got, err := AutoDetectTabularConfig(cfg, []string{"a", "b"}, [][]string{{"1", "2"}})
	s.Require().NoError(err)
	s.Empty(got.Axes)
}

func (s *AutoDetectTabularConfigSuite) TestSkipsWhenFewerThanTwoNumeric() {
	cfg := Config{AutoGroup: true, GroupPattern: "x", ChartTypes: []string{"scatter"}}
	got, err := AutoDetectTabularConfig(cfg, []string{"price"}, [][]string{{"10"}, {"20"}})
	s.Require().NoError(err)
	s.Empty(got.Axes)
}

func (s *AutoDetectTabularConfigSuite) TestFinalizeGroupConfigSkipsRegex() {
	cfg := Config{GroupRegex: ".*", Group: []string{"a"}}
	got, err := FinalizeGroupConfig(cfg)
	s.Require().NoError(err)
	s.Equal(cfg, got)
}

func TestAutoDetectTabularConfigSuite(t *testing.T) {
	suite.Run(t, new(AutoDetectTabularConfigSuite))
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
