package json

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
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

func (s *JSONSuite) TestAxesValueModeTwoColumns() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "price"}, {Source: "latency"}}
	js := `[{"name":"a","price":100,"latency":12},{"name":"b","price":200,"latency":8}]`

	results := ParseJSON(s.writeFile(js), s.cfg)

	s.Len(results, 2)
	s.Equal("100", results[0].XAxis)
	s.Equal("12", results[0].YAxis)
	s.Equal("", results[0].ZAxis)
	s.Empty(results[0].Stats)
}

func (s *JSONSuite) TestAxesValueModeNumericStringField() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}
	js := `[{"x":"1.5","y":2}]`

	results := ParseJSON(s.writeFile(js), s.cfg)

	s.Len(results, 1)
	s.Equal("1.5", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
}

func (s *JSONSuite) TestHybridModeGroupPlusAxesField() {
	s.cfg.Group = []string{"region", "category"}
	s.cfg.GroupPattern = "x,y"
	s.cfg.Axes = []parser.ColumnSpec{{Source: "latency", Label: "Latency (ms)"}}
	js := `[{"region":"US","category":"Widget","latency":12},{"region":"EU","category":"Gadget","latency":8}]`

	results := ParseJSON(s.writeFile(js), s.cfg)

	s.Len(results, 2)
	s.Equal("US", results[0].XAxis)
	s.Equal("Widget", results[0].YAxis)
	s.Empty(results[0].ZAxis)
	s.Len(results[0].Stats, 1)
	s.Equal("Latency (ms)", results[0].Stats[0].Type)
	s.Equal(12.0, *results[0].Stats[0].Value)
}

func (s *JSONSuite) TestAxesValueModeThreeColumns() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "x"}, {Source: "y"}, {Source: "z"}}
	js := `[{"x":1,"y":2,"z":3},{"x":4,"y":5,"z":6}]`

	results := ParseJSON(s.writeFile(js), s.cfg)

	s.Len(results, 2)
	s.Equal("1", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
	s.Equal("3", results[0].ZAxis)
	s.Equal("6", results[1].ZAxis)
}

func (s *JSONSuite) TestAxesValueModeSkipsIncompleteRow() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}
	js := `[{"x":1,"y":2},{"x":3}]`

	results := ParseJSON(s.writeFile(js), s.cfg)

	s.Len(results, 1)
	s.Equal("1", results[0].XAxis)
}

func (s *JSONSuite) TestHybridModeSkipsMissingZField() {
	s.cfg.Group = []string{"region", "category"}
	s.cfg.GroupPattern = "x,y"
	s.cfg.Axes = []parser.ColumnSpec{{Source: "latency"}}
	js := `[{"region":"US","category":"Widget","latency":12},{"region":"EU","category":"Gadget"}]`

	results := ParseJSON(s.writeFile(js), s.cfg)

	s.Len(results, 1)
	s.Equal("US", results[0].XAxis)
}

func (s *JSONSuite) TestHybridModeFilterRegex() {
	s.cfg.Group = []string{"region", "category"}
	s.cfg.GroupPattern = "x,y"
	s.cfg.Axes = []parser.ColumnSpec{{Source: "latency"}}
	s.cfg.Filter = "US"
	js := `[{"region":"US","category":"Widget","latency":12},{"region":"EU","category":"Gadget","latency":8}]`

	results := ParseJSON(s.writeFile(js), s.cfg)

	s.Len(results, 1)
	s.Equal("US", results[0].XAxis)
}

func (s *JSONSuite) TestHybridModeZLabelFallsBackToSource() {
	s.cfg.Group = []string{"region", "category"}
	s.cfg.GroupPattern = "x,y"
	s.cfg.Axes = []parser.ColumnSpec{{Source: "latency"}}
	js := `[{"region":"US","category":"Widget","latency":12}]`

	results := ParseJSON(s.writeFile(js), s.cfg)

	s.Len(results, 1)
	s.Equal("latency", results[0].Stats[0].Type)
}

func (s *JSONSuite) TestStringifyNumericAndString() {
	s.Equal("3.5", stringify(3.5))
	s.Equal("alpha", stringify("alpha"))
	s.Equal("", stringify(true))
}

func (s *JSONSuite) TestLeafNumberRejectsNonNumeric() {
	_, ok := leafNumber(true)
	s.False(ok)
	_, ok = leafNumber(nil)
	s.False(ok)
	_, ok = leafNumber(math.NaN())
	s.False(ok)
}

func (s *JSONSuite) TestDecodeElementSkipsNonObjectRows() {
	dec := json.NewDecoder(strings.NewReader(`[1, {"x":1}, [2,3]]`))
	tok, err := dec.Token()
	s.Require().NoError(err)
	open, ok := tok.(json.Delim)
	s.Require().True(ok)
	s.Equal(json.Delim('['), open)

	leaves, err := decodeElement(dec)
	s.Require().NoError(err)
	s.Nil(leaves)

	leaves, err = decodeElement(dec)
	s.Require().NoError(err)
	s.Len(leaves, 1)
	s.Equal("x", leaves[0].key)

	leaves, err = decodeElement(dec)
	s.Require().NoError(err)
	s.Nil(leaves)
}

func (s *JSONSuite) TestHybridModeNumericStringZField() {
	s.cfg.Group = []string{"region", "category"}
	s.cfg.GroupPattern = "x,y"
	s.cfg.Axes = []parser.ColumnSpec{{Source: "latency"}}
	js := `[{"region":"US","category":"Widget","latency":"3.5"}]`

	results := ParseJSON(s.writeFile(js), s.cfg)

	s.Len(results, 1)
	s.Equal(3.5, *results[0].Stats[0].Value)
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

func (s *JSONFatalSuite) TestAxesValueModeMissingFieldErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "missing"}, {Source: "y"}}
	path := s.writeFile(`[{"x":1,"y":2}]`)

	s.PanicsWithValue("exit", func() { ParseJSON(path, s.cfg) })
}

func (s *JSONFatalSuite) TestAxesValueModeNonNumericFieldErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "name"}, {Source: "y"}}
	path := s.writeFile(`[{"name":"alpha","y":2}]`)

	s.PanicsWithValue("exit", func() { ParseJSON(path, s.cfg) })
}

func (s *JSONFatalSuite) TestHybridModeNonNumericZFieldErrors() {
	s.cfg.Group = []string{"region", "category"}
	s.cfg.GroupPattern = "x,y"
	s.cfg.Axes = []parser.ColumnSpec{{Source: "label"}}
	path := s.writeFile(`[{"region":"US","category":"Widget","label":"foo"}]`)

	s.PanicsWithValue("exit", func() { ParseJSON(path, s.cfg) })
}

func (s *JSONFatalSuite) TestHybridModeMissingZFieldErrors() {
	s.cfg.Group = []string{"region", "category"}
	s.cfg.GroupPattern = "x,y"
	s.cfg.Axes = []parser.ColumnSpec{{Source: "missing"}}
	path := s.writeFile(`[{"region":"US","category":"Widget","latency":12}]`)

	s.PanicsWithValue("exit", func() { ParseJSON(path, s.cfg) })
}

func TestJSONFatalSuite(t *testing.T) {
	suite.Run(t, new(JSONFatalSuite))
}
