package parser

import (
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

type TabularPatternSuite struct {
	suite.Suite
}

func (s *TabularPatternSuite) TestParseFlatTabularPattern() {
	tp, err := ParseTabularPattern("y,x")
	s.Require().NoError(err)
	s.Len(tp.Slots, 2)
	s.False(tp.Slots[0].ValueSplit)
	s.Equal("yAxis", tp.Slots[0].Dimension)
	s.Equal("xAxis", tp.Slots[1].Dimension)
	s.Equal([]string{","}, tp.TopSeparators)
}

func (s *TabularPatternSuite) TestParseBracketDateCategory() {
	tp, err := ParseTabularPattern("[x-y-n],z")
	s.Require().NoError(err)
	s.Len(tp.Slots, 2)
	s.True(tp.Slots[0].ValueSplit)
	s.Equal("x-y-n", tp.Slots[0].InnerPattern)
	s.Equal("zAxis", tp.Slots[1].Dimension)
	s.Equal([]string{","}, tp.TopSeparators)
}

func (s *TabularPatternSuite) TestParseBracketSlashBenchmark() {
	tp, err := ParseTabularPattern("[n/x/y]")
	s.Require().NoError(err)
	s.Len(tp.Slots, 1)
	s.True(tp.Slots[0].ValueSplit)
	s.Equal("n/x/y", tp.Slots[0].InnerPattern)
}

func (s *TabularPatternSuite) TestParseBracketMixedSpace() {
	tp, err := ParseTabularPattern("[x n y],z")
	s.Require().NoError(err)
	s.Len(tp.Slots, 2)
	s.True(tp.Slots[0].ValueSplit)
	s.Equal("x n y", tp.Slots[0].InnerPattern)
	s.Equal("zAxis", tp.Slots[1].Dimension)
	s.Equal([]string{","}, tp.TopSeparators)
}

func (s *TabularPatternSuite) TestParseBracketRejectsUnclosed() {
	_, err := ParseTabularPattern("[x-y-n,z")
	s.Require().Error(err)
	s.Contains(err.Error(), "unclosed")
}

func (s *TabularPatternSuite) TestParseBracketRejectsDuplicateDimension() {
	_, err := ParseTabularPattern("[x-y],x")
	s.Require().Error(err)
	s.Contains(err.Error(), "duplicate dimension")
}

func (s *TabularPatternSuite) TestGroupTabularRowDateCategory() {
	cfg, err := ResolveGroupConfig(Config{
		Group:        []string{"date", "category"},
		GroupPattern: "[x-y-n],z",
	})
	s.Require().NoError(err)

	got, err := GroupTabularRow([]string{"2022-2-30", "Widget"}, cfg)
	s.Require().NoError(err)
	s.Equal("2022", got["xAxis"])
	s.Equal("2", got["yAxis"])
	s.Equal("30", got["name"])
	s.Equal("Widget", got["zAxis"])
}

func (s *TabularPatternSuite) TestGroupTabularRowSlashPath() {
	cfg, err := ResolveGroupConfig(Config{
		Group:        []string{"benchmark"},
		GroupPattern: "[n/x/y]",
	})
	s.Require().NoError(err)

	got, err := GroupTabularRow([]string{"Sort/1024/QuickSort"}, cfg)
	s.Require().NoError(err)
	s.Equal("Sort", got["name"])
	s.Equal("1024", got["xAxis"])
	s.Equal("QuickSort", got["yAxis"])
}

func (s *TabularPatternSuite) TestGroupTabularRowUnderscoreID() {
	cfg, err := ResolveGroupConfig(Config{
		Group:        []string{"id"},
		GroupPattern: "[n_y_x]",
	})
	s.Require().NoError(err)

	got, err := GroupTabularRow([]string{"lib_func_case"}, cfg)
	s.Require().NoError(err)
	s.Equal("lib", got["name"])
	s.Equal("func", got["yAxis"])
	s.Equal("case", got["xAxis"])
}

func (s *TabularPatternSuite) TestGroupTabularRowConsecutiveSeparatorsSkipMiddle() {
	cfg, err := ResolveGroupConfig(Config{
		Group:        []string{"date"},
		GroupPattern: "[n{Year}--x{Date}]",
	})
	s.Require().NoError(err)

	got, err := GroupTabularRow([]string{"2024-01-01"}, cfg)
	s.Require().NoError(err)
	s.Equal("2024", got["name"])
	s.Equal("01", got["xAxis"])
	s.Equal("", got["yAxis"])
}

func (s *TabularPatternSuite) TestGroupTabularRowTrailingSeparatorDropsDay() {
	cfg, err := ResolveGroupConfig(Config{
		Group:        []string{"date"},
		GroupPattern: "[n{Year}-x{Month}-]",
	})
	s.Require().NoError(err)

	got, err := GroupTabularRow([]string{"2024-01-01"}, cfg)
	s.Require().NoError(err)
	s.Equal("2024", got["name"])
	s.Equal("01", got["xAxis"])
	s.Equal("", got["yAxis"])
}

func (s *TabularPatternSuite) TestGroupTabularRowPrefixSkip() {
	cfg, err := ResolveGroupConfig(Config{
		Group:        []string{"name"},
		GroupPattern: "[/n/y]",
	})
	s.Require().NoError(err)

	got, err := GroupTabularRow([]string{"BenchmarkTest/JSON/Marshal"}, cfg)
	s.Require().NoError(err)
	s.Equal("JSON", got["name"])
	s.Equal("Marshal", got["yAxis"])
}

func (s *TabularPatternSuite) TestGroupBenchmarkNameRejectsBrackets() {
	_, err := GroupBenchmarkName("Sort/1024/QuickSort", Config{GroupPattern: "[n/x/y]"})
	s.Require().Error(err)
	s.Contains(err.Error(), "only supported for CSV/JSON")
}

func (s *TabularPatternSuite) TestValidateTabularAlignmentBracketSlots() {
	cfg, err := ResolveGroupConfig(Config{
		Group:        []string{"date", "category"},
		GroupPattern: "[x-y-n],z",
	})
	s.Require().NoError(err)
	s.Require().NoError(ValidateTabularGroupAlignment(cfg))
}

func (s *TabularPatternSuite) TestParseBracketWithCurlyAxisLabels() {
	tp, err := ParseTabularPattern("[n{year}-y{months}-x{dates}],z{RenamedCategory}")
	s.Require().NoError(err)
	s.Len(tp.Slots, 2)
	s.True(tp.Slots[0].ValueSplit)
	s.Equal("n-y-x", tp.Slots[0].InnerPattern)
	s.Equal("year", tp.Slots[0].AxisLabels["name"])
	s.Equal("months", tp.Slots[0].AxisLabels["yAxis"])
	s.Equal("dates", tp.Slots[0].AxisLabels["xAxis"])
	s.Equal("zAxis", tp.Slots[1].Dimension)
	s.Equal("RenamedCategory", tp.Slots[1].AxisLabels["zAxis"])
}

func (s *TabularPatternSuite) TestGroupTabularRowWithCurlyLabels() {
	cfg, err := ResolveGroupConfig(Config{
		Group:        []string{"order_date", "category"},
		GroupPattern: "[n{year}-y{months}-x{dates}],z{category}",
	})
	s.Require().NoError(err)

	got, err := GroupTabularRow([]string{"2022-2-30", "Widget"}, cfg)
	s.Require().NoError(err)
	s.Equal("2022", got["name"])
	s.Equal("2", got["yAxis"])
	s.Equal("30", got["xAxis"])
	s.Equal("Widget", got["zAxis"])
}

func (s *TabularPatternSuite) TestGroupAxesCurlyLabels() {
	cfg, err := ResolveGroupConfig(Config{
		Group:        []string{"order_date", "category"},
		GroupPattern: "[n{year}-y{months}-x{dates}],z{RenamedCategory}",
	})
	s.Require().NoError(err)

	axes := GroupAxes(cfg)
	s.Equal([]shared.Axis{
		{Key: "name", Label: "year"},
		{Key: "y", Label: "months"},
		{Key: "x", Label: "dates"},
		{Key: "z", Label: "RenamedCategory"},
	}, axes)
}

func (s *TabularPatternSuite) TestParseFlatPatternWithCurlyLabels() {
	tp, err := ParseTabularPattern("y{region},x{product}")
	s.Require().NoError(err)
	s.Len(tp.Slots, 2)
	s.Equal("region", tp.Slots[0].AxisLabels["yAxis"])
	s.Equal("product", tp.Slots[1].AxisLabels["xAxis"])
}

func (s *TabularPatternSuite) TestValidateTabularAlignmentRejectsSlotMismatch() {
	cfg, err := ResolveGroupConfig(Config{
		Group:        []string{"date"},
		GroupPattern: "[x-y-n],z",
	})
	s.Require().NoError(err)
	err = ValidateTabularGroupAlignment(cfg)
	s.Require().Error(err)
	s.Contains(err.Error(), "slot")
}

func TestTabularPatternSuite(t *testing.T) {
	suite.Run(t, new(TabularPatternSuite))
}
