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

func (s *JSONFatalSuite) TestValueModeMissingAxisFieldErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "missing"}, {Source: "y"}}
	path := s.writeFile(`[{"x":1,"y":2}]`)

	s.PanicsWithValue("exit", func() { ParseJSON(path, s.cfg) })
}

func (s *JSONFatalSuite) TestValueModeNonNumericAxisFieldErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "name"}, {Source: "y"}}
	path := s.writeFile(`[{"name":"alpha","y":2}]`)

	s.PanicsWithValue("exit", func() { ParseJSON(path, s.cfg) })
}

func (s *JSONFatalSuite) TestValueModeMetricFieldMissingErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}
	s.cfg.MetricColumn = "m"
	path := s.writeFile(`[{"x":1,"y":2}]`)

	s.PanicsWithValue("exit", func() { ParseJSON(path, s.cfg) })
}

func (s *JSONFatalSuite) TestValueModeMetricFieldNonNumericErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}
	s.cfg.MetricColumn = "label"
	path := s.writeFile(`[{"x":1,"y":2,"label":"foo"}]`)

	s.PanicsWithValue("exit", func() { ParseJSON(path, s.cfg) })
}

func (s *JSONFatalSuite) TestValueModeSkipsRowWithBadMetric() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}
	s.cfg.MetricColumn = "m"
	path := s.writeFile(`[{"x":1,"y":2,"m":3},{"x":4,"y":5,"m":"bad"},{"x":6,"y":7,"m":8}]`)

	results := ParseJSON(path, s.cfg)
	s.Len(results, 2)
	s.Equal("3", results[0].Metric)
	s.Equal("8", results[1].Metric)
}

func (s *JSONAutoValueSuite) TestSelectScopesAutoDetect() {
	s.cfg.Select = []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}
	path := s.writeFile(`[{"x":1,"y":2,"z":3,"w":4}]`)

	results := ParseJSON(path, s.cfg)
	s.Require().Len(results, 1)
	s.Equal("1", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
	s.Empty(results[0].ZAxis)
}

func TestJSONFatalSuite(t *testing.T) {
	suite.Run(t, new(JSONFatalSuite))
}

// JSONAutoGroupSuite exercises ParseJSON with cfg.AutoGroup set, simulating
// the pipeline's "no grouping configured" case.
type JSONAutoGroupSuite struct {
	suite.Suite
	cfg parser.Config
}

func (s *JSONAutoGroupSuite) SetupTest() {
	s.cfg = parser.Config{GroupPattern: "x", AutoGroup: true, ChartTypes: []string{"scatter"}}
}

func (s *JSONAutoGroupSuite) writeFile(content string) string {
	path := filepath.Join(s.T().TempDir(), "data.json")
	s.Require().NoError(os.WriteFile(path, []byte(content), 0644))
	return path
}

func (s *JSONAutoGroupSuite) TestCategoricalFieldBecomesXAxis() {
	j := `[{"region":"West","sells":10},{"region":"East","sells":20}]`
	results := ParseJSON(s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("West", results[0].XAxis)
	s.Equal("East", results[1].XAxis)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *JSONAutoGroupSuite) TestNestedFlattenedFieldChosen() {
	// nested object flattened to "addr.city"; along with region (also categorical)
	// the most-unique categorical wins. Both have 2 distinct here; leftmost in
	// first-seen order wins → region (appears first).
	j := `[{"region":"West","addr":{"city":"NY"},"sells":10},{"region":"East","addr":{"city":"LA"},"sells":20}]`
	results := ParseJSON(s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.NotEmpty(results[0].XAxis)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *JSONAutoGroupSuite) TestAllNumericAutoValues() {
	// all numeric → auto-value-mode: first 2 cols become x,y value axes
	j := `[{"id":1,"sells":10},{"id":2,"sells":20},{"id":3,"sells":30}]`
	results := ParseJSON(s.writeFile(j), s.cfg)
	s.Require().Len(results, 3)
	s.Equal("1", results[0].XAxis)
	s.Equal("10", results[0].YAxis)
	s.Empty(results[0].Stats)
}

func (s *JSONAutoGroupSuite) TestAutoGroupPicksSingleFieldEvenWithMultipleCategoricals() {
	// product 7 distinct > region 5 → xAxis=product only
	j := `[{"region":"West","product":"A","sells":10},{"region":"East","product":"B","sells":20},` +
		`{"region":"North","product":"C","sells":30},{"region":"South","product":"D","sells":40},` +
		`{"region":"Central","product":"E","sells":50},{"region":"West","product":"F","sells":60},` +
		`{"region":"East","product":"G","sells":70}]`
	results := ParseJSON(s.writeFile(j), s.cfg)
	s.Require().NotEmpty(results)
	for _, r := range results {
		s.NotEmpty(r.XAxis)
		s.Empty(r.YAxis)
	}
}

func (s *JSONAutoGroupSuite) TestExplicitGroupDisablesAutoGroup() {
	s.cfg.Group = []string{"region"}
	j := `[{"region":"West","sells":10},{"region":"East","sells":20}]`
	results := ParseJSON(s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("West", results[0].XAxis)
	s.Empty(results[0].YAxis)
}

func (s *JSONAutoGroupSuite) TestAxesDisablesAutoGroup() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "sells"}}
	j := `[{"region":"West","sells":10},{"region":"East","sells":20}]`
	_ = ParseJSON(s.writeFile(j), s.cfg) // no panic; value mode handled elsewhere
}

func TestJSONAutoGroupSuite(t *testing.T) {
	suite.Run(t, new(JSONAutoGroupSuite))
}

type JSONAutoValueSuite struct {
	suite.Suite
	cfg           parser.Config
	restoreOsExit func()
}

func (s *JSONAutoValueSuite) SetupTest() {
	s.restoreOsExit, _ = testutil.TrapOsExitPanic(s.T())
	s.cfg = parser.Config{GroupPattern: "x", AutoGroup: true, ChartTypes: []string{"scatter"}}
}

func (s *JSONAutoValueSuite) TearDownTest() {
	s.restoreOsExit()
}

func (s *JSONAutoValueSuite) writeFile(content string) string {
	path := filepath.Join(s.T().TempDir(), "data.json")
	s.Require().NoError(os.WriteFile(path, []byte(content), 0644))
	return path
}

func (s *JSONAutoValueSuite) TestTwoNumericFields() {
	j := `[{"price":10,"latency":5},{"price":20,"latency":7}]`
	results := ParseJSON(s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("10", results[0].XAxis)
	s.Equal("5", results[0].YAxis)
	s.Empty(results[0].Stats)
}

func (s *JSONAutoValueSuite) TestThreeNumericFields() {
	j := `[{"price":10,"latency":5,"mem":100},{"price":20,"latency":7,"mem":200}]`
	results := ParseJSON(s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("10", results[0].XAxis)
	s.Equal("5", results[0].YAxis)
	s.Equal("100", results[0].ZAxis)
	s.Empty(results[0].Stats)
}

func (s *JSONAutoValueSuite) TestFourNumericFieldsTakeFirstThree() {
	j := `[{"a":1,"b":2,"c":3,"d":4},{"a":5,"b":6,"c":7,"d":8}]`
	results := ParseJSON(s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("1", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
	s.Equal("3", results[0].ZAxis)
	s.Equal("4", results[0].Metric)
	s.Empty(results[0].Stats)
}

func (s *JSONAutoValueSuite) TestAutoGroupTakesPriority() {
	// categorical exists → auto-group fires
	j := `[{"region":"West","price":10},{"region":"East","price":20}]`
	results := ParseJSON(s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("West", results[0].XAxis)
	s.NotEmpty(results[0].Stats)
}

func (s *JSONAutoValueSuite) TestOneNumericFieldFallsBackToFlat() {
	j := `[{"price":10},{"price":20}]`
	results := ParseJSON(s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Empty(results[0].XAxis)
	s.NotEmpty(results[0].Stats)
}

func (s *JSONAutoValueSuite) TestPieChartFallsBackToFlat() {
	s.cfg.ChartTypes = []string{"pie"}
	j := `[{"price":10,"latency":5},{"price":20,"latency":7}]`
	results := ParseJSON(s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Empty(results[0].XAxis)
	s.NotEmpty(results[0].Stats)
}

func (s *JSONAutoValueSuite) TestRadarChartFallsBackToFlat() {
	s.cfg.ChartTypes = []string{"radar"}
	j := `[{"price":10,"latency":5},{"price":20,"latency":7}]`
	results := ParseJSON(s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Empty(results[0].XAxis)
	s.NotEmpty(results[0].Stats)
}

func TestJSONAutoValueSuite(t *testing.T) {
	suite.Run(t, new(JSONAutoValueSuite))
}
