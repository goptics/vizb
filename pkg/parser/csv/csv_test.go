package csv

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func parseCSVFile(t testing.TB, path string, cfg parser.Config) ([]shared.DataPoint, parser.Config, error) {
	t.Helper()
	input, err := os.Open(path)
	if err != nil {
		return nil, cfg, err
	}
	defer input.Close()
	points, effectiveCfg, _, err := ParseCSV(input, cfg)
	return points, effectiveCfg, err
}

func parseCSVFileError(t testing.TB, path string, cfg parser.Config) error {
	t.Helper()
	_, _, err := parseCSVFile(t, path, cfg)
	return err
}

func mustParseCSVFile(t testing.TB, path string, cfg parser.Config) ([]shared.DataPoint, parser.Config) {
	t.Helper()
	points, effectiveCfg, err := parseCSVFile(t, path, cfg)
	if err != nil {
		t.Fatalf("ParseCSV returned an error: %v", err)
	}
	return points, effectiveCfg
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

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

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

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

	s.Equal([]string{"Unit price", "Total"}, statTypes(results[0].Stats))
}

func (s *CSVSuite) TestExplicitColsMissingColumnErrors() {
	s.cfg.Select = []parser.ColumnSpec{{Source: "missing"}}
	csv := "name,price\na,10\n"

	s.Error(parseCSVFileError(s.T(), s.writeFile(csv), s.cfg))
}

func (s *CSVSuite) TestExplicitColsNonNumericErrors() {
	s.cfg.Select = []parser.ColumnSpec{{Source: "name"}}
	csv := "name,price\nalpha,10\n"

	s.Error(parseCSVFileError(s.T(), s.writeFile(csv), s.cfg))
}

func (s *CSVSuite) TestNumericColumnsBecomeChartsNoGroup() {
	csv := "name,sells,stocks,date\na,10,5,2024-01\nb,20,7,2025-02\n"

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

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

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

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
	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), cfg)

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
	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), cfg)

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
	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), cfg)

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

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

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
	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), cfg)

	s.Len(results, 1)
	s.Equal("alpha", results[0].XAxis)
	s.Equal("beta", results[0].Name)
	s.Equal("gamma", results[0].YAxis)
}

func (s *CSVSuite) TestGroupColumnExcludedFromCharts() {
	s.cfg.Group = []string{"id"}
	csv := "id,sells\n1,10\n2,20\n"

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

	s.Len(results, 2)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
	s.Equal("1", results[0].XAxis)
}

func (s *CSVSuite) TestAnyOneParsesMakesJunkChartColumn() {
	csv := "name,mostlytext\na,hello\nb,42\n"

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

	// mostlytext qualifies as a chart column (>=1 numeric cell);
	// row a has no numeric cell → dropped, row b kept.
	s.Len(results, 1)
	s.Equal([]string{"mostlytext"}, statTypes(results[0].Stats))
	s.Equal(42.0, *results[0].Stats[0].Value)
}

func (s *CSVSuite) TestNaNAndInfCellsSkipped() {
	csv := "name,v\na,NaN\nb,Inf\nc,3\n"

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

	// only c has a finite value
	s.Len(results, 1)
	s.Equal(3.0, *results[0].Stats[0].Value)
}

func (s *CSVSuite) TestPureNonNumericColumnIgnored() {
	csv := "label,sells\nfoo,10\nbar,20\n"

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

	s.Len(results, 2)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *CSVSuite) TestBOMStrippedFromFirstHeader() {
	s.cfg.Group = []string{"name"}
	csv := "\ufeffname,sells\nalpha,10\n"

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	s.Equal("alpha", results[0].XAxis)
}

func (s *CSVSuite) TestWhitespaceTrimmedInHeadersAndGroupValues() {
	s.cfg.Group = []string{"name"}
	csv := " name , sells \n alpha , 10 \n"

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	s.Equal("alpha", results[0].XAxis)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
	s.Equal(10.0, *results[0].Stats[0].Value)
}

func (s *CSVSuite) TestRaggedRowsTolerated() {
	csv := "name,sells,stocks\na,10\nb,20,7\n"

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

	s.Len(results, 2)
	// row a missing stocks cell → only sells stat
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
	s.Equal([]string{"sells", "stocks"}, statTypes(results[1].Stats))
}

func (s *CSVSuite) TestDuplicateHeadersSuffixed() {
	csv := "sells,sells\n10,20\n"

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	s.Equal([]string{"sells", "sells (2)"}, statTypes(results[0].Stats))
}

func (s *CSVSuite) TestEmptyHeaderColumnIgnored() {
	csv := "name,,sells\na,99,10\n"
	s.cfg.Group = []string{"name"}

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	// the empty-named column (value 99) is not charted
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *CSVSuite) TestEmptyGroupEntryFilteredOut() {
	s.cfg.Group = []string{"name", ""}
	csv := "name,sells\nalpha,10\n"

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	s.Equal("alpha", results[0].XAxis)
}

func (s *CSVSuite) TestFilterRegexOnGroupLabel() {
	s.cfg.Group = []string{"name"}
	s.cfg.Filter = "keep"
	csv := "name,sells\nkeep_a,10\ndrop_b,20\nkeep_c,30\n"

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

	s.Len(results, 2)
	for _, r := range results {
		s.Contains(r.XAxis, "keep")
	}
}

func (s *CSVSuite) TestInvalidFilterReturnsError() {
	s.cfg.Group = []string{"name"}
	s.cfg.Filter = "["

	err := parseCSVFileError(s.T(), s.writeFile("name,sells\nkeep,10\n"), s.cfg)
	s.ErrorContains(err, "invalid filter regex")
}

func (s *CSVSuite) TestNumberUnitScaling() {
	s.cfg.NumberUnit = "M"
	csv := "name,sells\na,2000000\n"

	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	s.Equal("sells (M)", results[0].Stats[0].Type)
	s.Equal(2.0, *results[0].Stats[0].Value)
}

func (s *CSVSuite) TestLessThanTwoRowsReturnsNil() {
	pts, _ := mustParseCSVFile(s.T(), s.writeFile("name,sells\n"), s.cfg)
	s.Nil(pts)
	pts, _ = mustParseCSVFile(s.T(), s.writeFile(""), s.cfg)
	s.Nil(pts)
}

func (s *CSVSuite) TestParseCSVReturnsResultsAndErrors() {
	results, cfg, _, err := ParseCSV(strings.NewReader("name,sells\nalpha,10\n"), parser.Config{
		GroupPattern: "x",
		Group:        []string{"name"},
	})
	s.Require().NoError(err)
	s.Equal([]string{"name"}, cfg.Group)
	s.Require().Len(results, 1)
	s.Equal("alpha", results[0].XAxis)

	_, _, _, err = ParseCSV(strings.NewReader("name,sells\nalpha,10\n"), parser.Config{
		GroupPattern: "x",
		Group:        []string{"missing"},
	})
	s.ErrorContains(err, `group column "missing" not found`)

	_, _, _, err = ParseCSV(strings.NewReader("name,sells\nalpha,\"bad\n"), parser.Config{GroupPattern: "x"})
	s.ErrorContains(err, "read CSV")

	_, _, _, err = ParseCSV(strings.NewReader("name,sells\nalpha,10\n"), parser.Config{
		Group:        []string{"name", "sells"},
		GroupPattern: "x",
	})
	s.Error(err)
}

func (s *CSVSuite) TestParseCSVReturnsAutoDetectError() {
	want := errors.New("auto detect failed")
	_, _, err := parseReader(
		strings.NewReader("name,sells\nalpha,10\n"),
		parser.Config{GroupPattern: "x"},
		func(cfg parser.Config, _ []string, _ [][]string) (parser.Config, error) {
			return cfg, want
		},
	)
	s.ErrorIs(err, want)
}

func (s *CSVSuite) TestParseCSVReturnsGroupRowError() {
	_, _, _, err := ParseCSV(strings.NewReader("name,sells\nalpha,10\n"), parser.Config{
		Group:      []string{"name"},
		GroupRegex: "explicit-regex-bypasses-tabular-pattern",
	})
	s.ErrorContains(err, "parse CSV group name")
	s.ErrorContains(err, "tabular pattern is not configured")
}

func TestCSVSuite(t *testing.T) {
	suite.Run(t, new(CSVSuite))
}

// CSVErrorSuite covers parser failures returned to callers.
type CSVErrorSuite struct {
	suite.Suite
	cfg parser.Config
}

func (s *CSVErrorSuite) SetupTest() {
	s.cfg = parser.Config{GroupPattern: "x"}
}

func (s *CSVErrorSuite) writeFile(content string) string {
	path := filepath.Join(s.T().TempDir(), "data.csv")
	s.Require().NoError(os.WriteFile(path, []byte(content), 0644))
	return path
}

func (s *CSVErrorSuite) TestMissingGroupColumnReturnsError() {
	s.cfg.Group = []string{"nope"}
	path := s.writeFile("name,sells\na,10\n")

	s.Error(parseCSVFileError(s.T(), path, s.cfg))
}

func (s *CSVErrorSuite) TestNoNumericColumnsReturnsError() {
	path := s.writeFile("name,label\na,foo\nb,bar\n")

	s.Error(parseCSVFileError(s.T(), path, s.cfg))
}

func (s *CSVErrorSuite) TestValueModeMissingAxisColumnErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "missing"}, {Source: "y"}}
	path := s.writeFile("x,y\n1,2\n")

	s.Error(parseCSVFileError(s.T(), path, s.cfg))
}

func (s *CSVErrorSuite) TestValueModeNonNumericAxisColumnErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "name"}, {Source: "y"}}
	path := s.writeFile("name,y\nalpha,2\n")

	s.Error(parseCSVFileError(s.T(), path, s.cfg))
}

func (s *CSVErrorSuite) TestValueModeMetricColumnMissingErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}
	s.cfg.MetricColumn = "m"
	path := s.writeFile("x,y\n1,2\n")

	s.Error(parseCSVFileError(s.T(), path, s.cfg))
}

func (s *CSVErrorSuite) TestValueModeMetricColumnNonNumericErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}
	s.cfg.MetricColumn = "label"
	path := s.writeFile("x,y,label\n1,2,foo\n")

	s.Error(parseCSVFileError(s.T(), path, s.cfg))
}

func (s *CSVErrorSuite) TestValueModeSkipsRowWithBadMetric() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}
	s.cfg.MetricColumn = "m"
	path := s.writeFile("x,y,m\n1,2,3\n4,5,bad\n6,7,8\n")

	results, _ := mustParseCSVFile(s.T(), path, s.cfg)
	s.Len(results, 2)
	s.Equal("3", results[0].Metric)
	s.Equal("8", results[1].Metric)
}

func TestCSVErrorSuite(t *testing.T) {
	suite.Run(t, new(CSVErrorSuite))
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
	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("West", results[0].XAxis)
	s.Equal("East", results[1].XAxis)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *CSVAutoGroupSuite) TestHighestCardinalityCategoricalWins() {
	// product has 3 distinct values; region has 2 → xAxis=product
	csv := "region,product,sells\nWest,A,10\nEast,B,20\nWest,C,30\n"
	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)
	s.Require().Len(results, 3)
	s.Equal("A", results[0].XAxis)
	s.Equal("B", results[1].XAxis)
	s.Equal("C", results[2].XAxis)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *CSVAutoGroupSuite) TestAllNumericAutoColAxis() {
	// all numeric → auto col-axis x; multi-stat points (expand is pipeline-side)
	csv := "id,sells\n1,10\n2,20\n3,30\n"
	results, effective := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)
	s.Equal("x", effective.ColAxis)
	s.Empty(effective.Axes)
	s.Require().Len(results, 3)
	s.Empty(results[0].XAxis)
	s.ElementsMatch([]string{"id", "sells"}, statTypes(results[0].Stats))
}

func (s *CSVAutoGroupSuite) TestAutoGroupPicksSingleColumnEvenWithMultipleCategoricals() {
	csv := "region,product,sells\nWest,A,10\nEast,B,20\nNorth,C,30\nSouth,D,40\nCentral,E,50\nWest,F,60\nEast,G,70\n"
	// product 7 distinct > region 5 → xAxis=product only
	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)
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
	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("West", results[0].XAxis)
	s.Empty(results[0].YAxis) // explicit single-col group, no yAxis
}

func (s *CSVAutoGroupSuite) TestSingleColumnNoOp() {
	csv := "sells\n10\n20\n"
	results, _ := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)
	// single column: auto-group cannot pick an axis; numeric col becomes a stat
	s.Require().Len(results, 2)
	s.Empty(results[0].XAxis)
}

func TestCSVAutoGroupSuite(t *testing.T) {
	suite.Run(t, new(CSVAutoGroupSuite))
}

type CSVAutoColAxisSuite struct {
	suite.Suite
	cfg           parser.Config
	restoreOsExit func()
}

func (s *CSVAutoColAxisSuite) SetupTest() {
	s.restoreOsExit, _ = testutil.TrapOsExitPanic(s.T())
	s.cfg = parser.Config{GroupPattern: "x", AutoGroup: true, ChartTypes: []string{"scatter"}}
}

func (s *CSVAutoColAxisSuite) TearDownTest() {
	s.restoreOsExit()
}

func (s *CSVAutoColAxisSuite) writeFile(content string) string {
	path := filepath.Join(s.T().TempDir(), "data.csv")
	s.Require().NoError(os.WriteFile(path, []byte(content), 0644))
	return path
}

func (s *CSVAutoColAxisSuite) TestTwoNumericColsSetColAxisMultiStat() {
	csv := "price,latency\n10,5\n20,7\n30,9\n"
	results, effective := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)
	s.Equal("x", effective.ColAxis)
	s.Empty(effective.Axes)
	s.Require().Len(results, 3)
	s.Empty(results[0].XAxis)
	s.Empty(results[0].YAxis)
	s.ElementsMatch([]string{"price", "latency"}, statTypes(results[0].Stats))
	s.Equal(10.0, *results[0].Stats[0].Value)
}

func (s *CSVAutoColAxisSuite) TestThreeNumericColsAllAsStats() {
	csv := "price,latency,memory\n10,5,100\n20,7,200\n30,9,300\n"
	results, effective := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)
	s.Equal("x", effective.ColAxis)
	s.Empty(effective.Axes)
	s.Require().Len(results, 3)
	s.Empty(results[0].XAxis)
	s.ElementsMatch([]string{"price", "latency", "memory"}, statTypes(results[0].Stats))
}

func (s *CSVAutoColAxisSuite) TestFourNumericColsAllAsStatsNoMetric() {
	csv := "a,b,c,d\n1,2,3,4\n5,6,7,8\n"
	results, effective := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)
	s.Equal("x", effective.ColAxis)
	s.Empty(effective.Axes)
	s.Empty(effective.MetricColumn)
	s.Require().Len(results, 2)
	s.Empty(results[0].Metric)
	s.ElementsMatch([]string{"a", "b", "c", "d"}, statTypes(results[0].Stats))
}

func (s *CSVAutoColAxisSuite) TestOneNumericColSetsColAxis() {
	csv := "price\n10\n20\n"
	results, effective := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)
	s.Equal("x", effective.ColAxis)
	s.Require().Len(results, 2)
	s.Empty(results[0].XAxis)
	s.Equal([]string{"price"}, statTypes(results[0].Stats))
}

func (s *CSVAutoColAxisSuite) TestAutoGroupTakesPriorityOverAutoColAxis() {
	// categorical columns exist → auto-group fires, not auto col-axis
	csv := "region,price,product\nWest,10,foo\nEast,20,bar\n"
	results, effective := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)
	s.Empty(effective.ColAxis)
	s.Require().Len(results, 2)
	s.NotEmpty(results[0].XAxis) // categorical xAxis from auto-group
	s.Empty(results[0].YAxis)    // single-col group
	s.NotEmpty(results[0].Stats) // price becomes chart column
}

func (s *CSVAutoColAxisSuite) TestMixedTypesAutoGroupNotColAxis() {
	// region (non-numeric), price (numeric), product (non-numeric), latency (numeric)
	// auto-group picks the categorical with highest cardinality
	csv := "region,price,product,latency\nWest,10,foo,5\nEast,20,bar,7\n"
	results, effective := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)
	s.Empty(effective.ColAxis)
	s.Require().Len(results, 2)
	s.NotEmpty(results[0].XAxis)
	s.NotEmpty(results[0].Stats)
}

func (s *CSVAutoColAxisSuite) TestPieChartGetsAutoColAxis() {
	s.cfg.ChartTypes = []string{"pie"}
	csv := "price,latency\n10,5\n20,7\n"
	results, effective := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)
	s.Equal("x", effective.ColAxis)
	s.Require().Len(results, 2)
	s.Empty(results[0].XAxis)
	s.ElementsMatch([]string{"price", "latency"}, statTypes(results[0].Stats))
}

func (s *CSVAutoColAxisSuite) TestHeatmapGetsAutoColAxis() {
	s.cfg.ChartTypes = []string{"heatmap"}
	csv := "price,latency\n10,5\n20,7\n"
	results, effective := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)
	s.Equal("x", effective.ColAxis)
	s.Require().Len(results, 2)
	s.ElementsMatch([]string{"price", "latency"}, statTypes(results[0].Stats))
}

func (s *CSVAutoColAxisSuite) TestSelectSkipsAutoDetect() {
	// Solo --select (SelectViews) disables auto col-axis inference and routes value mode.
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "x", AxisKey: "x"}, {Source: "y", AxisKey: "y"}}},
	}
	csv := "x,y,z,w\n1,2,3,4\n"
	results, effective := mustParseCSVFile(s.T(), s.writeFile(csv), s.cfg)
	s.Empty(effective.ColAxis)
	s.Require().Len(results, 1)
	s.Equal("1", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
	s.Empty(results[0].Stats)
}

func (s *CSVErrorSuite) TestSelectMixedModeMapsCategoryXAndValueY() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "latency", AxisKey: "y"}}},
	}
	path := s.writeFile("region,latency,sales\nAsia,12,100\nEU,11,60\n")

	results, _ := mustParseCSVFile(s.T(), path, s.cfg)
	s.Require().Len(results, 2)
	s.Equal("Asia", results[0].XAxis)
	s.Equal("12", results[0].YAxis)
	s.Empty(results[0].Stats)
}

func (s *CSVErrorSuite) TestSelectColumnNotFoundReturnsError() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "missing", AxisKey: "x"}, {Source: "latency", AxisKey: "y"}}},
	}
	path := s.writeFile("region,latency\nAsia,12\n")
	s.Error(parseCSVFileError(s.T(), path, s.cfg))
}

func (s *CSVErrorSuite) TestSelectNonNumericYColumnReturnsError() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "label", AxisKey: "y"}}},
	}
	path := s.writeFile("region,label\nAsia,fast\n")
	s.Error(parseCSVFileError(s.T(), path, s.cfg))
}

func (s *CSVErrorSuite) TestSelectEmptyColumnReturnsError() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "latency", AxisKey: "y"}}},
	}
	path := s.writeFile("region,latency\n,\n")
	s.Error(parseCSVFileError(s.T(), path, s.cfg))
}

func (s *CSVErrorSuite) TestSelectValueModeAllNumeric() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "x", AxisKey: "x"}, {Source: "y", AxisKey: "y"}}},
	}
	path := s.writeFile("x,y\n1,2\n3,4\n")

	results, _ := mustParseCSVFile(s.T(), path, s.cfg)
	s.Require().Len(results, 2)
	s.Equal("1", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
	s.Empty(results[0].Stats)
}

func (s *CSVErrorSuite) TestSelectValueModeWithMetricColumn() {
	// noise-grid shape: x,y,z position + value visualMap metric
	view, err := parser.ParseSelectViewFlag("x,y,z,value")
	s.Require().NoError(err)
	s.cfg.SelectViews = []parser.SelectView{view}
	s.cfg.Mode = parser.ResolveMode(s.cfg)
	path := s.writeFile("x,y,z,value\n0,0,0,4\n1,2,3,5.5\n")

	results, effective := mustParseCSVFile(s.T(), path, s.cfg)
	s.Equal("value", effective.MetricColumn)
	s.Require().Len(results, 2)
	s.Equal("0", results[0].XAxis)
	s.Equal("0", results[0].YAxis)
	s.Equal("0", results[0].ZAxis)
	s.Equal("4", results[0].Metric)
	s.Equal("1", results[1].XAxis)
	s.Equal("5.5", results[1].Metric)
	s.Empty(results[0].Stats)
}

func (s *CSVErrorSuite) TestSelectMultiStatModeIndependentCombinations() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "latency", AxisKey: "y"}}},
		{Columns: []parser.ColumnSpec{{Source: "product", AxisKey: "x"}, {Source: "sales", AxisKey: "y"}}},
	}
	path := s.writeFile("region,latency,sales,product\nAsia,12,100,Widget\n")

	results, _ := mustParseCSVFile(s.T(), path, s.cfg)
	s.Require().Len(results, 2)
	s.Equal("Asia", results[0].XAxis)
	s.Require().Len(results[0].Stats, 1)
	s.Equal("latency by region", results[0].Stats[0].Type)
	s.Equal(12.0, *results[0].Stats[0].Value)
	s.Equal("Widget", results[1].XAxis)
	s.Equal("sales by product", results[1].Stats[0].Type)
	s.Equal(100.0, *results[1].Stats[0].Value)
}

func (s *CSVErrorSuite) TestSelectMultiStatModeParenTypeLabel() {
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

	results, _ := mustParseCSVFile(s.T(), path, s.cfg)
	s.Require().Len(results, 1)
	s.Equal("Asia", results[0].XAxis)
	s.Require().Len(results[0].Stats, 2)
	s.Equal("Latency by Region", results[0].Stats[0].Type)
	s.Equal("Sales by Region", results[0].Stats[1].Type)
}

func (s *CSVErrorSuite) TestSelectMultiStatModeCustomTypeLabel() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "latency", AxisKey: "y", Label: "Custom"}}},
		{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "sales", AxisKey: "y"}}},
	}
	path := s.writeFile("region,latency,sales\nAsia,12,100\n")

	results, _ := mustParseCSVFile(s.T(), path, s.cfg)
	s.Require().Len(results, 1)
	s.Require().Len(results[0].Stats, 2)
	s.Equal("Custom", results[0].Stats[0].Type)
	s.Equal("sales by region", results[0].Stats[1].Type)
}

func TestCSVAutoColAxisSuite(t *testing.T) {
	suite.Run(t, new(CSVAutoColAxisSuite))
}
