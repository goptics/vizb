package csv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/testutil"
	"github.com/stretchr/testify/suite"
)

func statTypes(stats []shared.Stat) []string {
	out := make([]string, len(stats))
	for i, s := range stats {
		out[i] = s.Type
	}
	return out
}

// CSVSuite exercises ParseCSV with a per-test parser.Config (built fresh in
// SetupTest), replacing the former global shared.FlagState mutation.
type CSVSuite struct {
	suite.Suite
	cfg           parser.Config
	restoreOsExit func()
}

func (s *CSVSuite) SetupTest() {
	s.restoreOsExit, _ = testutil.TrapOsExitPanic(s.T())
	s.cfg = parser.Config{GroupPattern: "x"}
}

func (s *CSVSuite) TearDownTest() {
	s.restoreOsExit()
}

func (s *CSVSuite) writeFile(content string) string {
	path := filepath.Join(s.T().TempDir(), "data.csv")
	s.Require().NoError(os.WriteFile(path, []byte(content), 0644))
	return path
}

func (s *CSVSuite) TestExplicitColsSelectsAndOrders() {
	s.cfg.Select = []parser.ColumnSpec{{Source: "price"}, {Source: "count"}}
	csv := "name,date,count,level,price\na,2024-01,10,1,100\n"

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	s.Equal([]string{"price", "count"}, statTypes(results[0].Stats))
	s.Equal(100.0, *results[0].Stats[0].Value)
	s.Equal(10.0, *results[0].Stats[1].Value)
}

func (s *CSVSuite) TestExplicitColsRename() {
	s.cfg.Select = []parser.ColumnSpec{
		{Source: "price", Label: "Unit price"},
		{Source: "count", Label: "Total"},
	}
	csv := "name,price,count\na,100,10\n"

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	s.Equal([]string{"Unit price", "Total"}, statTypes(results[0].Stats))
}

func (s *CSVSuite) TestExplicitColsMissingColumnErrors() {
	s.cfg.Select = []parser.ColumnSpec{{Source: "missing"}}
	csv := "name,price\na,10\n"

	s.Panics(func() { ParseCSV(s.writeFile(csv), s.cfg) })
}

func (s *CSVSuite) TestExplicitColsNonNumericErrors() {
	s.cfg.Select = []parser.ColumnSpec{{Source: "name"}}
	csv := "name,price\nalpha,10\n"

	s.Panics(func() { ParseCSV(s.writeFile(csv), s.cfg) })
}

func (s *CSVSuite) TestNumericColumnsBecomeChartsNoGroup() {
	csv := "name,sells,stocks,date\na,10,5,2024-01\nb,20,7,2025-02\n"

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 2)
	s.Equal([]string{"sells", "stocks"}, statTypes(results[0].Stats))
	s.Equal(10.0, *results[0].Stats[0].Value)
	s.Equal(5.0, *results[0].Stats[1].Value)
	// no -g → empty labels
	s.Empty(results[0].Name)
	s.Empty(results[0].XAxis)
	s.Empty(results[0].YAxis)
}

func (s *CSVSuite) TestGroupSingleColumnToXAxis() {
	s.cfg.Group = []string{"name"}
	csv := "name,sells,date\nalpha,10,2024-01\nbeta,20,2025-02\n"

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 2)
	s.Equal("alpha", results[0].XAxis)
	s.Equal("beta", results[1].XAxis)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *CSVSuite) TestGroupBracketValueSplitDateCategory() {
	cfg, err := parser.ResolveGroupConfig(parser.Config{
		Group:        []string{"date", "category"},
		GroupPattern: "[x-y-n],z",
	})
	s.Require().NoError(err)

	csv := "date,category,sales\n2022-2-30,Widget,100\n"
	results, _ := ParseCSV(s.writeFile(csv), cfg)

	s.Len(results, 1)
	s.Equal("2022", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
	s.Equal("30", results[0].Name)
	s.Equal("Widget", results[0].ZAxis)
	s.Equal(100.0, *results[0].Stats[0].Value)
}

func (s *CSVSuite) TestGroupBracketValueSplitSlashBenchmark() {
	cfg, err := parser.ResolveGroupConfig(parser.Config{
		Group:        []string{"benchmark"},
		GroupPattern: "[n/x/y]",
	})
	s.Require().NoError(err)

	csv := "benchmark,latency\nSort/1024/QuickSort,12\n"
	results, _ := ParseCSV(s.writeFile(csv), cfg)

	s.Len(results, 1)
	s.Equal("Sort", results[0].Name)
	s.Equal("1024", results[0].XAxis)
	s.Equal("QuickSort", results[0].YAxis)
}

func (s *CSVSuite) TestGroupBracketValueSplitMixedWithWholeColumn() {
	cfg, err := parser.ResolveGroupConfig(parser.Config{
		Group:        []string{"date", "region"},
		GroupPattern: "[n-y-x],z",
	})
	s.Require().NoError(err)

	csv := "date,region,sales\n2022-2-30,USA,80\n"
	results, _ := ParseCSV(s.writeFile(csv), cfg)

	s.Len(results, 1)
	s.Equal("2022", results[0].Name)
	s.Equal("2", results[0].YAxis)
	s.Equal("30", results[0].XAxis)
	s.Equal("USA", results[0].ZAxis)
}

func (s *CSVSuite) TestGroupMultiColumnRoutedByPattern() {
	s.cfg.Group = []string{"name", "date"}
	s.cfg.GroupPattern = "name,x"
	csv := "name,sells,date\nalpha,10,2024-01\nbeta,20,2025-02\n"

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 2)
	s.Equal("alpha", results[0].Name)
	s.Equal("2024-01", results[0].XAxis)
}

func (s *CSVSuite) TestGroupSpaceSeparatedPattern() {
	cfg, err := parser.ResolveGroupConfig(parser.Config{
		Group:        []string{"name category region"},
		GroupPattern: "x n y",
	})
	s.Require().NoError(err)

	csv := "name,category,region,sells\nalpha,beta,gamma,10\n"
	results, _ := ParseCSV(s.writeFile(csv), cfg)

	s.Len(results, 1)
	s.Equal("alpha", results[0].XAxis)
	s.Equal("beta", results[0].Name)
	s.Equal("gamma", results[0].YAxis)
}

func (s *CSVSuite) TestGroupColumnExcludedFromCharts() {
	s.cfg.Group = []string{"id"}
	csv := "id,sells\n1,10\n2,20\n"

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 2)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
	s.Equal("1", results[0].XAxis)
}

func (s *CSVSuite) TestAnyOneParsesMakesJunkChartColumn() {
	csv := "name,mostlytext\na,hello\nb,42\n"

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	// mostlytext qualifies as a chart column (>=1 numeric cell);
	// row a has no numeric cell → dropped, row b kept.
	s.Len(results, 1)
	s.Equal([]string{"mostlytext"}, statTypes(results[0].Stats))
	s.Equal(42.0, *results[0].Stats[0].Value)
}

func (s *CSVSuite) TestNaNAndInfCellsSkipped() {
	csv := "name,v\na,NaN\nb,Inf\nc,3\n"

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	// only c has a finite value
	s.Len(results, 1)
	s.Equal(3.0, *results[0].Stats[0].Value)
}

func (s *CSVSuite) TestPureNonNumericColumnIgnored() {
	csv := "label,sells\nfoo,10\nbar,20\n"

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 2)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *CSVSuite) TestBOMStrippedFromFirstHeader() {
	s.cfg.Group = []string{"name"}
	csv := "\ufeffname,sells\nalpha,10\n"

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	s.Equal("alpha", results[0].XAxis)
}

func (s *CSVSuite) TestWhitespaceTrimmedInHeadersAndGroupValues() {
	s.cfg.Group = []string{"name"}
	csv := " name , sells \n alpha , 10 \n"

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	s.Equal("alpha", results[0].XAxis)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
	s.Equal(10.0, *results[0].Stats[0].Value)
}

func (s *CSVSuite) TestRaggedRowsTolerated() {
	csv := "name,sells,stocks\na,10\nb,20,7\n"

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 2)
	// row a missing stocks cell → only sells stat
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
	s.Equal([]string{"sells", "stocks"}, statTypes(results[1].Stats))
}

func (s *CSVSuite) TestDuplicateHeadersSuffixed() {
	csv := "sells,sells\n10,20\n"

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	s.Equal([]string{"sells", "sells (2)"}, statTypes(results[0].Stats))
}

func (s *CSVSuite) TestEmptyHeaderColumnIgnored() {
	csv := "name,,sells\na,99,10\n"
	s.cfg.Group = []string{"name"}

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	// the empty-named column (value 99) is not charted
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *CSVSuite) TestEmptyGroupEntryFilteredOut() {
	s.cfg.Group = []string{"name", ""}
	csv := "name,sells\nalpha,10\n"

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	s.Equal("alpha", results[0].XAxis)
}

func (s *CSVSuite) TestFilterRegexOnGroupLabel() {
	s.cfg.Group = []string{"name"}
	s.cfg.Filter = "keep"
	csv := "name,sells\nkeep_a,10\ndrop_b,20\nkeep_c,30\n"

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 2)
	for _, r := range results {
		s.Contains(r.XAxis, "keep")
	}
}

func (s *CSVSuite) TestNumberUnitScaling() {
	s.cfg.NumberUnit = "M"
	csv := "name,sells\na,2000000\n"

	results, _ := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	s.Equal("sells (M)", results[0].Stats[0].Type)
	s.Equal(2.0, *results[0].Stats[0].Value)
}

func (s *CSVSuite) TestLessThanTwoRowsReturnsNil() {
	pts, _ := ParseCSV(s.writeFile("name,sells\n"), s.cfg)
	s.Nil(pts)
	pts, _ = ParseCSV(s.writeFile(""), s.cfg)
	s.Nil(pts)
}

func TestCSVSuite(t *testing.T) {
	suite.Run(t, new(CSVSuite))
}

// CSVFatalSuite covers the fatal (os.Exit) paths by trapping shared.OsExit.
type CSVFatalSuite struct {
	suite.Suite
	cfg        parser.Config
	origOsExit func(int)
}

func (s *CSVFatalSuite) SetupTest() {
	s.cfg = parser.Config{GroupPattern: "x"}
	s.origOsExit = shared.OsExit
	shared.OsExit = func(int) { panic("exit") }
}

func (s *CSVFatalSuite) TearDownTest() {
	shared.OsExit = s.origOsExit
}

func (s *CSVFatalSuite) writeFile(content string) string {
	path := filepath.Join(s.T().TempDir(), "data.csv")
	s.Require().NoError(os.WriteFile(path, []byte(content), 0644))
	return path
}

func (s *CSVFatalSuite) TestMissingGroupColumnIsFatal() {
	s.cfg.Group = []string{"nope"}
	path := s.writeFile("name,sells\na,10\n")

	s.PanicsWithValue("exit", func() { ParseCSV(path, s.cfg) })
}

func (s *CSVFatalSuite) TestNoNumericColumnsIsFatal() {
	path := s.writeFile("name,label\na,foo\nb,bar\n")

	s.PanicsWithValue("exit", func() { ParseCSV(path, s.cfg) })
}

func (s *CSVFatalSuite) TestValueModeMissingAxisColumnErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "missing"}, {Source: "y"}}
	path := s.writeFile("x,y\n1,2\n")

	s.PanicsWithValue("exit", func() { ParseCSV(path, s.cfg) })
}

func (s *CSVFatalSuite) TestValueModeNonNumericAxisColumnErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "name"}, {Source: "y"}}
	path := s.writeFile("name,y\nalpha,2\n")

	s.PanicsWithValue("exit", func() { ParseCSV(path, s.cfg) })
}

func (s *CSVFatalSuite) TestValueModeMetricColumnMissingErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}
	s.cfg.MetricColumn = "m"
	path := s.writeFile("x,y\n1,2\n")

	s.PanicsWithValue("exit", func() { ParseCSV(path, s.cfg) })
}

func (s *CSVFatalSuite) TestValueModeMetricColumnNonNumericErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}
	s.cfg.MetricColumn = "label"
	path := s.writeFile("x,y,label\n1,2,foo\n")

	s.PanicsWithValue("exit", func() { ParseCSV(path, s.cfg) })
}

func (s *CSVFatalSuite) TestValueModeSkipsRowWithBadMetric() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}
	s.cfg.MetricColumn = "m"
	path := s.writeFile("x,y,m\n1,2,3\n4,5,bad\n6,7,8\n")

	results, _ := ParseCSV(path, s.cfg)
	s.Len(results, 2)
	s.Equal("3", results[0].Metric)
	s.Equal("8", results[1].Metric)
}

func TestCSVFatalSuite(t *testing.T) {
	suite.Run(t, new(CSVFatalSuite))
}

// CSVAutoGroupSuite exercises ParseCSV with cfg.AutoGroup set, simulating the
// pipeline's "no grouping configured" case.
type CSVAutoGroupSuite struct {
	suite.Suite
	cfg           parser.Config
	restoreOsExit func()
}

func (s *CSVAutoGroupSuite) SetupTest() {
	s.restoreOsExit, _ = testutil.TrapOsExitPanic(s.T())
	s.cfg = parser.Config{GroupPattern: "x", AutoGroup: true, ChartTypes: []string{"scatter"}}
}

func (s *CSVAutoGroupSuite) TearDownTest() {
	s.restoreOsExit()
}

func (s *CSVAutoGroupSuite) writeFile(content string) string {
	path := filepath.Join(s.T().TempDir(), "data.csv")
	s.Require().NoError(os.WriteFile(path, []byte(content), 0644))
	return path
}

func (s *CSVAutoGroupSuite) TestCategoricalColumnBecomesXAxis() {
	csv := "region,sells\nWest,10\nEast,20\n"
	results, _ := ParseCSV(s.writeFile(csv), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("West", results[0].XAxis)
	s.Equal("East", results[1].XAxis)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *CSVAutoGroupSuite) TestHighestCardinalityCategoricalWins() {
	// product has 3 distinct values; region has 2 → xAxis=product
	csv := "region,product,sells\nWest,A,10\nEast,B,20\nWest,C,30\n"
	results, _ := ParseCSV(s.writeFile(csv), s.cfg)
	s.Require().Len(results, 3)
	s.Equal("A", results[0].XAxis)
	s.Equal("B", results[1].XAxis)
	s.Equal("C", results[2].XAxis)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *CSVAutoGroupSuite) TestAllNumericAutoValues() {
	// all numeric → auto-value-mode kicks in: first 2 cols become x,y value axes
	csv := "id,sells\n1,10\n2,20\n3,30\n"
	results, _ := ParseCSV(s.writeFile(csv), s.cfg)
	s.Require().Len(results, 3)
	s.Equal("1", results[0].XAxis)
	s.Equal("10", results[0].YAxis)
	s.Empty(results[0].Stats)
}

func (s *CSVAutoGroupSuite) TestAutoGroupPicksSingleColumnEvenWithMultipleCategoricals() {
	csv := "region,product,sells\nWest,A,10\nEast,B,20\nNorth,C,30\nSouth,D,40\nCentral,E,50\nWest,F,60\nEast,G,70\n"
	// product 7 distinct > region 5 → xAxis=product only
	results, _ := ParseCSV(s.writeFile(csv), s.cfg)
	s.Require().NotEmpty(results)
	for _, r := range results {
		s.NotEmpty(r.XAxis)
		s.Empty(r.YAxis)
	}
}

func (s *CSVAutoGroupSuite) TestExplicitGroupDisablesAutoGroup() {
	// Even though AutoGroup is true, explicit --group wins (AutoGroupApplies
	// checks len(cfg.Group)==0).
	s.cfg.Group = []string{"region"}
	csv := "region,product,sells\nWest,A,10\nEast,B,20\n"
	results, _ := ParseCSV(s.writeFile(csv), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("West", results[0].XAxis)
	s.Empty(results[0].YAxis) // explicit single-col group, no yAxis
}

func (s *CSVAutoGroupSuite) TestSingleColumnNoOp() {
	csv := "sells\n10\n20\n"
	results, _ := ParseCSV(s.writeFile(csv), s.cfg)
	// single column: auto-group cannot pick an axis; numeric col becomes a stat
	s.Require().Len(results, 2)
	s.Empty(results[0].XAxis)
}

func TestCSVAutoGroupSuite(t *testing.T) {
	suite.Run(t, new(CSVAutoGroupSuite))
}

type CSVAutoValueSuite struct {
	suite.Suite
	cfg           parser.Config
	restoreOsExit func()
}

func (s *CSVAutoValueSuite) SetupTest() {
	s.restoreOsExit, _ = testutil.TrapOsExitPanic(s.T())
	s.cfg = parser.Config{GroupPattern: "x", AutoGroup: true, ChartTypes: []string{"scatter"}}
}

func (s *CSVAutoValueSuite) TearDownTest() {
	s.restoreOsExit()
}

func (s *CSVAutoValueSuite) writeFile(content string) string {
	path := filepath.Join(s.T().TempDir(), "data.csv")
	s.Require().NoError(os.WriteFile(path, []byte(content), 0644))
	return path
}

func (s *CSVAutoValueSuite) TestTwoNumericColsProduceValueAxes() {
	csv := "price,latency\n10,5\n20,7\n30,9\n"
	results, _ := ParseCSV(s.writeFile(csv), s.cfg)
	s.Require().Len(results, 3)
	s.Equal("10", results[0].XAxis)
	s.Equal("5", results[0].YAxis)
	s.Equal("20", results[1].XAxis)
	s.Equal("7", results[1].YAxis)
	s.Empty(results[0].Stats)
}

func (s *CSVAutoValueSuite) TestThreeNumericColsProduceValueAxes() {
	csv := "price,latency,memory\n10,5,100\n20,7,200\n30,9,300\n"
	results, _ := ParseCSV(s.writeFile(csv), s.cfg)
	s.Require().Len(results, 3)
	s.Equal("10", results[0].XAxis)
	s.Equal("5", results[0].YAxis)
	s.Equal("100", results[0].ZAxis)
	s.Empty(results[0].Stats)
}

func (s *CSVAutoValueSuite) TestFourNumericColsTakeFirstThree() {
	csv := "a,b,c,d\n1,2,3,4\n5,6,7,8\n"
	results, _ := ParseCSV(s.writeFile(csv), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("1", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
	s.Equal("3", results[0].ZAxis)
	s.Equal("4", results[0].Metric)
	s.Empty(results[0].Stats)
}

func (s *CSVAutoValueSuite) TestOneNumericColFallsBackToFlat() {
	// single numeric column → auto-group and auto-value both skip, flat series
	csv := "price\n10\n20\n"
	results, _ := ParseCSV(s.writeFile(csv), s.cfg)
	s.Require().Len(results, 2)
	s.Empty(results[0].XAxis)
	s.NotEmpty(results[0].Stats)
}

func (s *CSVAutoValueSuite) TestAutoGroupTakesPriorityOverAutoValue() {
	// categorical columns exist → auto-group fires, not auto-value
	csv := "region,price,product\nWest,10,foo\nEast,20,bar\n"
	results, _ := ParseCSV(s.writeFile(csv), s.cfg)
	s.Require().Len(results, 2)
	s.NotEmpty(results[0].XAxis) // categorical xAxis from auto-group
	s.Empty(results[0].YAxis)    // single-col group
	s.NotEmpty(results[0].Stats) // price becomes chart column
}

func (s *CSVAutoValueSuite) TestMixedTypesSkipsNonNumeric() {
	// region (non-numeric), price (numeric), product (non-numeric), latency (numeric)
	// auto-group picks the categorical with highest cardinality
	// auto-value only fires when NO categoricals exist
	csv := "region,price,product,latency\nWest,10,foo,5\nEast,20,bar,7\n"
	results, _ := ParseCSV(s.writeFile(csv), s.cfg)
	// region has 2 distinct, product has 2 distinct → auto-group picks region as xAxis
	s.Require().Len(results, 2)
	s.NotEmpty(results[0].XAxis)
	s.NotEmpty(results[0].Stats)
}

func (s *CSVAutoValueSuite) TestPieChartFallsBackToFlat() {
	// pie chart type → auto-value is NOT eligible, falls back to flat series
	s.cfg.ChartTypes = []string{"pie"}
	csv := "price,latency\n10,5\n20,7\n"
	results, _ := ParseCSV(s.writeFile(csv), s.cfg)
	s.Require().Len(results, 2)
	s.Empty(results[0].XAxis)
	s.NotEmpty(results[0].Stats)
}

func (s *CSVAutoValueSuite) TestHeatmapChartFallsBackToFlat() {
	s.cfg.ChartTypes = []string{"heatmap"}
	csv := "price,latency\n10,5\n20,7\n"
	results, _ := ParseCSV(s.writeFile(csv), s.cfg)
	s.Require().Len(results, 2)
	s.Empty(results[0].XAxis)
	s.NotEmpty(results[0].Stats)
}

func (s *CSVAutoValueSuite) TestSelectSkipsAutoDetect() {
	// Solo --select (SelectViews) disables auto-value inference and routes value mode.
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "x", AxisKey: "x"}, {Source: "y", AxisKey: "y"}}},
	}
	csv := "x,y,z,w\n1,2,3,4\n"
	results, _ := ParseCSV(s.writeFile(csv), s.cfg)
	s.Require().Len(results, 1)
	s.Equal("1", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
	s.Empty(results[0].Stats)
}

func (s *CSVFatalSuite) TestSelectMixedModeMapsCategoryXAndValueY() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "latency", AxisKey: "y"}}},
	}
	path := s.writeFile("region,latency,sales\nAsia,12,100\nEU,11,60\n")

	results, _ := ParseCSV(path, s.cfg)
	s.Require().Len(results, 2)
	s.Equal("Asia", results[0].XAxis)
	s.Equal("12", results[0].YAxis)
	s.Empty(results[0].Stats)
}

func (s *CSVFatalSuite) TestSelectColumnNotFoundExits() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "missing", AxisKey: "x"}, {Source: "latency", AxisKey: "y"}}},
	}
	path := s.writeFile("region,latency\nAsia,12\n")
	s.Panics(func() { ParseCSV(path, s.cfg) })
}

func (s *CSVFatalSuite) TestSelectNonNumericYColumnExits() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "label", AxisKey: "y"}}},
	}
	path := s.writeFile("region,label\nAsia,fast\n")
	s.Panics(func() { ParseCSV(path, s.cfg) })
}

func (s *CSVFatalSuite) TestSelectEmptyColumnExits() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "latency", AxisKey: "y"}}},
	}
	path := s.writeFile("region,latency\n,\n")
	s.Panics(func() { ParseCSV(path, s.cfg) })
}

func (s *CSVFatalSuite) TestSelectValueModeAllNumeric() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "x", AxisKey: "x"}, {Source: "y", AxisKey: "y"}}},
	}
	path := s.writeFile("x,y\n1,2\n3,4\n")

	results, _ := ParseCSV(path, s.cfg)
	s.Require().Len(results, 2)
	s.Equal("1", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
	s.Empty(results[0].Stats)
}

func (s *CSVFatalSuite) TestSelectMultiStatModeIndependentCombinations() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "latency", AxisKey: "y"}}},
		{Columns: []parser.ColumnSpec{{Source: "product", AxisKey: "x"}, {Source: "sales", AxisKey: "y"}}},
	}
	path := s.writeFile("region,latency,sales,product\nAsia,12,100,Widget\n")

	results, _ := ParseCSV(path, s.cfg)
	s.Require().Len(results, 2)
	s.Equal("Asia", results[0].XAxis)
	s.Require().Len(results[0].Stats, 1)
	s.Equal("latency by region", results[0].Stats[0].Type)
	s.Equal(12.0, *results[0].Stats[0].Value)
	s.Equal("Widget", results[1].XAxis)
	s.Equal("sales by product", results[1].Stats[0].Type)
	s.Equal(100.0, *results[1].Stats[0].Value)
}

func (s *CSVFatalSuite) TestSelectMultiStatModeParenTypeLabel() {
	s.cfg.SelectViews = []parser.SelectView{
		{
			Columns:   []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "latency", AxisKey: "y"}},
			TypeLabel: "Latency by Region",
		},
		{
			Columns:   []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "sales", AxisKey: "y"}},
			TypeLabel: "Sales by Region",
		},
	}
	path := s.writeFile("region,latency,sales\nAsia,12,100\n")

	results, _ := ParseCSV(path, s.cfg)
	s.Require().Len(results, 1)
	s.Equal("Asia", results[0].XAxis)
	s.Require().Len(results[0].Stats, 2)
	s.Equal("Latency by Region", results[0].Stats[0].Type)
	s.Equal("Sales by Region", results[0].Stats[1].Type)
}

func (s *CSVFatalSuite) TestSelectMultiStatModeCustomTypeLabel() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "latency", AxisKey: "y", Label: "Custom"}}},
		{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "sales", AxisKey: "y"}}},
	}
	path := s.writeFile("region,latency,sales\nAsia,12,100\n")

	results, _ := ParseCSV(path, s.cfg)
	s.Require().Len(results, 1)
	s.Require().Len(results[0].Stats, 2)
	s.Equal("Custom", results[0].Stats[0].Type)
	s.Equal("sales by region", results[0].Stats[1].Type)
}

func TestCSVAutoValueSuite(t *testing.T) {
	suite.Run(t, new(CSVAutoValueSuite))
}
