package json

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

// JSONSuite exercises ParseJSON with a per-test parser.Config.
type JSONSuite struct {
	suite.Suite
	cfg parser.Config
}

func (s *JSONSuite) SetupTest() {
	s.cfg = parser.Config{GroupPattern: "x"}
}

func (s *JSONSuite) writeFile(content string) string {
	path := filepath.Join(s.T().TempDir(), "data.json")
	s.Require().NoError(os.WriteFile(path, []byte(content), 0644))
	return path
}

func (s *JSONSuite) TestExplicitColsSelectsAndOrders() {
	s.cfg.Select = []parser.ColumnSpec{{Source: "stocks"}, {Source: "sells"}}
	j := `[{"name":"a","sells":10,"stocks":5},{"name":"b","sells":20,"stocks":7}]`

	results := ParseJSON(s.writeFile(j), s.cfg)

	s.Equal([]string{"stocks", "sells"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestExplicitColsRename() {
	s.cfg.Select = []parser.ColumnSpec{{Source: "sells", Label: "Revenue"}}
	j := `[{"name":"a","sells":10}]`

	results := ParseJSON(s.writeFile(j), s.cfg)

	s.Equal([]string{"Revenue"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestExplicitColsNestedKey() {
	s.cfg.Select = []parser.ColumnSpec{{Source: "mem.alloc"}}
	j := `[{"name":"a","mem":{"alloc":3}}]`

	results := ParseJSON(s.writeFile(j), s.cfg)

	s.Equal([]string{"mem.alloc"}, statTypes(results[0].Stats))
	s.Equal(3.0, *results[0].Stats[0].Value)
}

func (s *JSONSuite) TestNumericFieldsBecomeChartsNoGroup() {
	j := `[{"name":"a","sells":10,"stocks":5,"date":"2024-01"},{"name":"b","sells":20,"stocks":7,"date":"2025-02"}]`

	results := ParseJSON(s.writeFile(j), s.cfg)

	s.Len(results, 2)
	s.Equal([]string{"sells", "stocks"}, statTypes(results[0].Stats))
	s.Equal(10.0, *results[0].Stats[0].Value)
	s.Equal(5.0, *results[0].Stats[1].Value)
	s.Empty(results[0].XAxis)
}

func (s *JSONSuite) TestFirstSeenColumnOrderPreserved() {
	j := `[{"zeta":1,"alpha":2,"mid":3}]`

	results := ParseJSON(s.writeFile(j), s.cfg)

	s.Len(results, 1)
	s.Equal([]string{"zeta", "alpha", "mid"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestNumericStringParsed() {
	j := `[{"name":"a","sells":"42"}]`

	results := ParseJSON(s.writeFile(j), s.cfg)

	s.Len(results, 1)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
	s.Equal(42.0, *results[0].Stats[0].Value)
}

func (s *JSONSuite) TestNestedObjectFlattenedToDottedKeys() {
	j := `[{"name":"a","mem":{"alloc":5,"bytes":100}}]`

	results := ParseJSON(s.writeFile(j), s.cfg)

	s.Len(results, 1)
	s.Equal([]string{"mem.alloc", "mem.bytes"}, statTypes(results[0].Stats))
	s.Equal(5.0, *results[0].Stats[0].Value)
}

func (s *JSONSuite) TestArrayValuedFieldSkipped() {
	j := `[{"name":"a","sells":10,"tags":[1,2,3]}]`

	results := ParseJSON(s.writeFile(j), s.cfg)

	s.Len(results, 1)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestBoolAndNullSkipped() {
	j := `[{"name":"a","sells":10,"active":true,"note":null}]`

	results := ParseJSON(s.writeFile(j), s.cfg)

	s.Len(results, 1)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestHeterogeneousRowsMissingKeyIsGap() {
	j := `[{"name":"a","sells":10},{"name":"b","stocks":7}]`

	results := ParseJSON(s.writeFile(j), s.cfg)

	s.Len(results, 2)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
	s.Equal([]string{"stocks"}, statTypes(results[1].Stats))
}

func (s *JSONSuite) TestMixedTypePerKeyNumericWhereParseable() {
	// v is a number in row 1, non-numeric string in row 2
	j := `[{"v":3},{"v":"foo"}]`

	results := ParseJSON(s.writeFile(j), s.cfg)

	// v qualifies as a chart column (>=1 numeric); row 2 has no stats → dropped
	s.Len(results, 1)
	s.Equal(3.0, *results[0].Stats[0].Value)
}

func (s *JSONSuite) TestGroupSingleFieldToXAxis() {
	s.cfg.Group = []string{"name"}
	j := `[{"name":"alpha","sells":10},{"name":"beta","sells":20}]`

	results := ParseJSON(s.writeFile(j), s.cfg)

	s.Len(results, 2)
	s.Equal("alpha", results[0].XAxis)
	s.Equal("beta", results[1].XAxis)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestGroupBracketValueSplitDateCategory() {
	cfg, err := parser.ResolveGroupConfig(parser.Config{
		Group:        []string{"date", "category"},
		GroupPattern: "[x-y-n],z",
	})
	s.Require().NoError(err)

	j := `[{"date":"2022-2-30","category":"Widget","sales":100}]`
	results := ParseJSON(s.writeFile(j), cfg)

	s.Len(results, 1)
	s.Equal("2022", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
	s.Equal("30", results[0].Name)
	s.Equal("Widget", results[0].ZAxis)
}

func (s *JSONSuite) TestGroupMultiFieldRoutedByPattern() {
	s.cfg.Group = []string{"name", "date"}
	s.cfg.GroupPattern = "name,x"
	j := `[{"name":"alpha","sells":10,"date":"2024-01"}]`

	results := ParseJSON(s.writeFile(j), s.cfg)

	s.Len(results, 1)
	s.Equal("alpha", results[0].Name)
	s.Equal("2024-01", results[0].XAxis)
}

func (s *JSONSuite) TestGroupOnNumericFieldStringified() {
	s.cfg.Group = []string{"id"}
	j := `[{"id":7,"sells":10}]`

	results := ParseJSON(s.writeFile(j), s.cfg)

	s.Len(results, 1)
	s.Equal("7", results[0].XAxis)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestFilterRegexOnGroupLabel() {
	s.cfg.Group = []string{"name"}
	s.cfg.Filter = "keep"
	j := `[{"name":"keep_a","sells":10},{"name":"drop_b","sells":20},{"name":"keep_c","sells":30}]`

	results := ParseJSON(s.writeFile(j), s.cfg)

	s.Len(results, 2)
	for _, r := range results {
		s.Contains(r.XAxis, "keep")
	}
}

func (s *JSONSuite) TestNumberUnitScaling() {
	s.cfg.NumberUnit = "M"
	j := `[{"name":"a","sells":2000000}]`

	results := ParseJSON(s.writeFile(j), s.cfg)

	s.Len(results, 1)
	s.Equal("sells (M)", results[0].Stats[0].Type)
	s.Equal(2.0, *results[0].Stats[0].Value)
}

func (s *JSONSuite) TestNonArrayInputReturnsNil() {
	s.Nil(ParseJSON(s.writeFile(`{"name":"a","sells":10}`), s.cfg))
	s.Nil(ParseJSON(s.writeFile(`[]`), s.cfg))
	s.Nil(ParseJSON(s.writeFile(``), s.cfg))
}

func TestJSONSuite(t *testing.T) {
	suite.Run(t, new(JSONSuite))
}

// JSONFatalSuite covers the fatal (os.Exit) paths by trapping shared.OsExit.
type JSONFatalSuite struct {
	suite.Suite
	cfg        parser.Config
	origOsExit func(int)
}

func (s *JSONFatalSuite) SetupTest() {
	s.cfg = parser.Config{GroupPattern: "x"}
	s.origOsExit = shared.OsExit
	shared.OsExit = func(int) { panic("exit") }
}

func (s *JSONFatalSuite) TearDownTest() {
	shared.OsExit = s.origOsExit
}

func (s *JSONFatalSuite) writeFile(content string) string {
	path := filepath.Join(s.T().TempDir(), "data.json")
	s.Require().NoError(os.WriteFile(path, []byte(content), 0644))
	return path
}

func (s *JSONFatalSuite) TestMissingGroupFieldIsFatal() {
	s.cfg.Group = []string{"nope"}
	path := s.writeFile(`[{"name":"a","sells":10}]`)

	s.PanicsWithValue("exit", func() { ParseJSON(path, s.cfg) })
}

func (s *JSONFatalSuite) TestNoNumericFieldsIsFatal() {
	path := s.writeFile(`[{"name":"a","label":"foo"}]`)

	s.PanicsWithValue("exit", func() { ParseJSON(path, s.cfg) })
}

func (s *JSONFatalSuite) TestExplicitColsMissingColumnErrors() {
	s.cfg.Select = []parser.ColumnSpec{{Source: "missing"}}
	path := s.writeFile(`[{"name":"a","sells":10}]`)

	s.PanicsWithValue("exit", func() { ParseJSON(path, s.cfg) })
}

func (s *JSONFatalSuite) TestExplicitColsNonNumericErrors() {
	s.cfg.Select = []parser.ColumnSpec{{Source: "name"}}
	path := s.writeFile(`[{"name":"alpha","sells":10}]`)

	s.PanicsWithValue("exit", func() { ParseJSON(path, s.cfg) })
}

func TestJSONFatalSuite(t *testing.T) {
	suite.Run(t, new(JSONFatalSuite))
}
