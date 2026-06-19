package csv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
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
	cfg parser.Config
}

func (s *CSVSuite) SetupTest() {
	s.cfg = parser.Config{GroupPattern: "x"}
}

func (s *CSVSuite) writeFile(content string) string {
	path := filepath.Join(s.T().TempDir(), "data.csv")
	s.Require().NoError(os.WriteFile(path, []byte(content), 0644))
	return path
}

func (s *CSVSuite) TestNumericColumnsBecomeChartsNoGroup() {
	csv := "name,sells,stocks,date\na,10,5,2024-01\nb,20,7,2025-02\n"

	results := ParseCSV(s.writeFile(csv), s.cfg)

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

	results := ParseCSV(s.writeFile(csv), s.cfg)

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
	results := ParseCSV(s.writeFile(csv), cfg)

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
	results := ParseCSV(s.writeFile(csv), cfg)

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
	results := ParseCSV(s.writeFile(csv), cfg)

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

	results := ParseCSV(s.writeFile(csv), s.cfg)

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
	results := ParseCSV(s.writeFile(csv), cfg)

	s.Len(results, 1)
	s.Equal("alpha", results[0].XAxis)
	s.Equal("beta", results[0].Name)
	s.Equal("gamma", results[0].YAxis)
}

func (s *CSVSuite) TestGroupColumnExcludedFromCharts() {
	s.cfg.Group = []string{"id"}
	csv := "id,sells\n1,10\n2,20\n"

	results := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 2)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
	s.Equal("1", results[0].XAxis)
}

func (s *CSVSuite) TestAnyOneParsesMakesJunkChartColumn() {
	csv := "name,mostlytext\na,hello\nb,42\n"

	results := ParseCSV(s.writeFile(csv), s.cfg)

	// mostlytext qualifies as a chart column (>=1 numeric cell);
	// row a has no numeric cell → dropped, row b kept.
	s.Len(results, 1)
	s.Equal([]string{"mostlytext"}, statTypes(results[0].Stats))
	s.Equal(42.0, *results[0].Stats[0].Value)
}

func (s *CSVSuite) TestNaNAndInfCellsSkipped() {
	csv := "name,v\na,NaN\nb,Inf\nc,3\n"

	results := ParseCSV(s.writeFile(csv), s.cfg)

	// only c has a finite value
	s.Len(results, 1)
	s.Equal(3.0, *results[0].Stats[0].Value)
}

func (s *CSVSuite) TestPureNonNumericColumnIgnored() {
	csv := "label,sells\nfoo,10\nbar,20\n"

	results := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 2)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *CSVSuite) TestBOMStrippedFromFirstHeader() {
	s.cfg.Group = []string{"name"}
	csv := "\ufeffname,sells\nalpha,10\n"

	results := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	s.Equal("alpha", results[0].XAxis)
}

func (s *CSVSuite) TestWhitespaceTrimmedInHeadersAndGroupValues() {
	s.cfg.Group = []string{"name"}
	csv := " name , sells \n alpha , 10 \n"

	results := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	s.Equal("alpha", results[0].XAxis)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
	s.Equal(10.0, *results[0].Stats[0].Value)
}

func (s *CSVSuite) TestRaggedRowsTolerated() {
	csv := "name,sells,stocks\na,10\nb,20,7\n"

	results := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 2)
	// row a missing stocks cell → only sells stat
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
	s.Equal([]string{"sells", "stocks"}, statTypes(results[1].Stats))
}

func (s *CSVSuite) TestDuplicateHeadersSuffixed() {
	csv := "sells,sells\n10,20\n"

	results := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	s.Equal([]string{"sells", "sells (2)"}, statTypes(results[0].Stats))
}

func (s *CSVSuite) TestEmptyHeaderColumnIgnored() {
	csv := "name,,sells\na,99,10\n"
	s.cfg.Group = []string{"name"}

	results := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	// the empty-named column (value 99) is not charted
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *CSVSuite) TestEmptyGroupEntryFilteredOut() {
	s.cfg.Group = []string{"name", ""}
	csv := "name,sells\nalpha,10\n"

	results := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	s.Equal("alpha", results[0].XAxis)
}

func (s *CSVSuite) TestFilterRegexOnGroupLabel() {
	s.cfg.Group = []string{"name"}
	s.cfg.Filter = "keep"
	csv := "name,sells\nkeep_a,10\ndrop_b,20\nkeep_c,30\n"

	results := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 2)
	for _, r := range results {
		s.Contains(r.XAxis, "keep")
	}
}

func (s *CSVSuite) TestNumberUnitScaling() {
	s.cfg.NumberUnit = "M"
	csv := "name,sells\na,2000000\n"

	results := ParseCSV(s.writeFile(csv), s.cfg)

	s.Len(results, 1)
	s.Equal("sells (M)", results[0].Stats[0].Type)
	s.Equal(2.0, *results[0].Stats[0].Value)
}

func (s *CSVSuite) TestLessThanTwoRowsReturnsNil() {
	s.Nil(ParseCSV(s.writeFile("name,sells\n"), s.cfg))
	s.Nil(ParseCSV(s.writeFile(""), s.cfg))
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

func TestCSVFatalSuite(t *testing.T) {
	suite.Run(t, new(CSVFatalSuite))
}
