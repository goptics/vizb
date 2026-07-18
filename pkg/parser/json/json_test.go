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

func parseJSONFile(t testing.TB, path string, cfg parser.Config) ([]shared.DataPoint, parser.Config, error) {
	t.Helper()
	input, err := os.Open(path)
	if err != nil {
		return nil, cfg, err
	}
	defer input.Close()
	return ParseJSON(input, cfg)
}

func parseJSONFileError(t testing.TB, path string, cfg parser.Config) error {
	t.Helper()
	_, _, err := parseJSONFile(t, path, cfg)
	return err
}

func mustParseJSONFile(t testing.TB, path string, cfg parser.Config) ([]shared.DataPoint, parser.Config) {
	t.Helper()
	points, effectiveCfg, err := parseJSONFile(t, path, cfg)
	if err != nil {
		t.Fatalf("ParseJSON returned an error: %v", err)
	}
	return points, effectiveCfg
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

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Equal([]string{"stocks", "sells"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestExplicitColsRename() {
	s.cfg.Select = []parser.ColumnSpec{{Source: "sells", Label: "Revenue"}}
	j := `[{"name":"a","sells":10}]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Equal([]string{"Revenue"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestExplicitColsNestedKey() {
	s.cfg.Select = []parser.ColumnSpec{{Source: "mem.alloc"}}
	j := `[{"name":"a","mem":{"alloc":3}}]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Equal([]string{"mem.alloc"}, statTypes(results[0].Stats))
	s.Equal(3.0, *results[0].Stats[0].Value)
}

func (s *JSONSuite) TestNumericFieldsBecomeChartsNoGroup() {
	j := `[{"name":"a","sells":10,"stocks":5,"date":"2024-01"},{"name":"b","sells":20,"stocks":7,"date":"2025-02"}]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Len(results, 2)
	s.Equal([]string{"sells", "stocks"}, statTypes(results[0].Stats))
	s.Equal(10.0, *results[0].Stats[0].Value)
	s.Equal(5.0, *results[0].Stats[1].Value)
	s.Empty(results[0].XAxis)
}

func (s *JSONSuite) TestFirstSeenColumnOrderPreserved() {
	j := `[{"zeta":1,"alpha":2,"mid":3}]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Len(results, 1)
	s.Equal([]string{"zeta", "alpha", "mid"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestNumericStringParsed() {
	j := `[{"name":"a","sells":"42"}]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Len(results, 1)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
	s.Equal(42.0, *results[0].Stats[0].Value)
}

func (s *JSONSuite) TestNestedObjectFlattenedToDottedKeys() {
	j := `[{"name":"a","mem":{"alloc":5,"bytes":100}}]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Len(results, 1)
	s.Equal([]string{"mem.alloc", "mem.bytes"}, statTypes(results[0].Stats))
	s.Equal(5.0, *results[0].Stats[0].Value)
}

func (s *JSONSuite) TestArrayValuedFieldSkipped() {
	j := `[{"name":"a","sells":10,"tags":[1,2,3]}]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Len(results, 1)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestHeaderMatrixUsesFirstRowAsColumns() {
	s.cfg.Group = []string{"region"}
	j := `[["region","sales"],["West",10],["East",20]]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Require().Len(results, 2)
	s.Equal("West", results[0].XAxis)
	s.Equal("East", results[1].XAxis)
	s.Equal([]string{"sales"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestMixedFirstMatrixRowUsesSyntheticColumns() {
	j := `[["x",2],[3,4]]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Require().Len(results, 2)
	s.Equal([]string{"y"}, statTypes(results[0].Stats))
	s.Equal([]string{"x", "y"}, statTypes(results[1].Stats))
}

func (s *JSONSuite) TestNumericLookingStringMatrixHeadersStayHeaders() {
	j := `[["10","20"],[1,2]]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Require().Len(results, 1)
	s.Equal([]string{"10", "20"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestMatrixHeaderEmptyAndDuplicateNames() {
	j := `[["","sales","sales"],[99,10,20]]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Require().Len(results, 1)
	s.Equal([]string{"sales", "sales (2)"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestEmptyFirstMatrixRowUsesSyntheticNamesForLaterRows() {
	j := `[[],[1,2]]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Require().Len(results, 1)
	s.Equal([]string{"x", "y"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestMatrixRaggedRowsAndSkippedCellsAreGaps() {
	j := `[["a","b","c"],[1],[2,3,4],[null,5,{"nested":true}]]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Require().Len(results, 3)
	s.Equal([]string{"a"}, statTypes(results[0].Stats))
	s.Equal([]string{"a", "b", "c"}, statTypes(results[1].Stats))
	s.Equal([]string{"b"}, statTypes(results[2].Stats))
}

func (s *JSONSuite) TestMatrixBoolNullAndNestedValuesAreSkipped() {
	j := `[["flag","empty","arr","obj","value"],[true,null,[1,2],{"nested":1},5]]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Require().Len(results, 1)
	s.Equal([]string{"value"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestNoHeaderMatrixSkipsMixedTopLevelElements() {
	j := `[[1,2],99,{"skip":true},[3,4]]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Require().Len(results, 2)
	s.Equal([]string{"x", "y"}, statTypes(results[0].Stats))
	s.Equal([]string{"x", "y"}, statTypes(results[1].Stats))
}

func (s *JSONSuite) TestNoHeaderMatrixNamesColumnsAfterMetric() {
	j := `[[1,2,3,4,5,6]]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Require().Len(results, 1)
	s.Equal([]string{"x", "y", "z", "metric", "col5", "col6"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestMatrixEmptyAndHeaderOnlyReturnNil() {
	pts, _ := mustParseJSONFile(s.T(), s.writeFile(`[]`), s.cfg)
	s.Nil(pts)

	pts, _ = mustParseJSONFile(s.T(), s.writeFile(`[["x","y"]]`), s.cfg)
	s.Nil(pts)
}

func (s *JSONSuite) TestDecodeTopLevelRowsMalformedInputsReturnErrors() {
	cases := []struct {
		name  string
		input string
	}{
		{name: "first token", input: `[tru`},
		{name: "object body", input: `[{"x":`},
		{name: "object tail", input: `[{"x":1},{"y":`},
		{name: "matrix first row", input: `[[1`},
		{name: "matrix header data row", input: `[["x"],[`},
		{name: "matrix data token", input: `[[1], tru`},
		{name: "matrix data row", input: `[[1],[2`},
		{name: "matrix skipped object", input: `[[1],{"skip":`},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			dec := json.NewDecoder(strings.NewReader(tc.input))
			tok, err := dec.Token()
			s.Require().NoError(err)
			s.Equal(json.Delim('['), tok)

			_, _, _, err = decodeTopLevelRows(dec)
			s.Error(err)
		})
	}
}

func (s *JSONSuite) TestDecodeTopLevelRowsScalarFirstFallsBackToObjectRows() {
	dec := json.NewDecoder(strings.NewReader(`[1,{"x":2}]`))
	tok, err := dec.Token()
	s.Require().NoError(err)
	s.Equal(json.Delim('['), tok)

	rows, colOrder, seenCol, err := decodeTopLevelRows(dec)
	s.Require().NoError(err)
	s.Equal([]string{"x"}, colOrder)
	s.True(seenCol["x"])
	s.Require().Len(rows, 1)
	s.Equal(2.0, rows[0]["x"])
}

func (s *JSONSuite) TestDecodeMatrixCellErrorAndUnexpectedClose() {
	dec := json.NewDecoder(strings.NewReader(""))
	_, err := decodeMatrixCell(dec)
	s.Error(err)

	dec = json.NewDecoder(strings.NewReader(`[tru`))
	tok, err := dec.Token()
	s.Require().NoError(err)
	s.Equal(json.Delim('['), tok)
	_, err = decodeMatrixRowBody(dec)
	s.Error(err)

	dec = json.NewDecoder(strings.NewReader(`[]`))
	tok, err = dec.Token()
	s.Require().NoError(err)
	s.Equal(json.Delim('['), tok)

	cell, err := decodeMatrixCell(dec)
	s.NoError(err)
	s.False(cell.ok)
}

func (s *JSONSuite) TestJSONKindFnSkipsRowsMissingAxisField() {
	rows := []map[string]any{
		{"y": 2.0},
		{"x": 1.0, "y": 3.0},
	}
	kind := jsonKindFn(rows, map[string]bool{"x": true, "y": true}, []string{"x", "y"}, "--select")

	got, err := kind("x", "x")

	s.NoError(err)
	s.Equal("value", got)
}

func (s *JSONSuite) TestResolveGroupKeysSkipsEmptyNames() {
	keys, set, err := resolveGroupKeys([]string{"x"}, map[string]bool{"x": true}, []string{"", "x"})
	s.Require().NoError(err)

	s.Equal([]string{"x"}, keys)
	s.True(set["x"])
}

func (s *JSONSuite) TestLowLevelObjectDecodersReturnErrors() {
	dec := json.NewDecoder(strings.NewReader(""))
	_, err := decodeElement(dec)
	s.Error(err)

	dec = json.NewDecoder(strings.NewReader(`{"x":1`))
	tok, err := dec.Token()
	s.Require().NoError(err)
	s.Equal(json.Delim('{'), tok)
	_, err = decodeObjectBody(dec, "")
	s.Error(err)

	dec = json.NewDecoder(strings.NewReader(`{`))
	tok, err = dec.Token()
	s.Require().NoError(err)
	s.Equal(json.Delim('{'), tok)
	_, err = decodeObjectBody(dec, "")
	s.Error(err)

	dec = json.NewDecoder(strings.NewReader(`{"`))
	tok, err = dec.Token()
	s.Require().NoError(err)
	s.Equal(json.Delim('{'), tok)
	_, err = decodeObjectBody(dec, "")
	s.Error(err)

	dec = json.NewDecoder(strings.NewReader(""))
	_, err = decodeValue(dec, "x")
	s.Error(err)

	dec = json.NewDecoder(strings.NewReader(`[]`))
	tok, err = dec.Token()
	s.Require().NoError(err)
	s.Equal(json.Delim('['), tok)
	leaves, err := decodeValue(dec, "x")
	s.NoError(err)
	s.Nil(leaves)

	dec = json.NewDecoder(strings.NewReader(`[[1]]`))
	tok, err = dec.Token()
	s.Require().NoError(err)
	s.Equal(json.Delim('['), tok)
	s.NoError(skipContainerBody(dec))
}

func (s *JSONSuite) TestBoolAndNullSkipped() {
	j := `[{"name":"a","sells":10,"active":true,"note":null}]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Len(results, 1)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestHeterogeneousRowsMissingKeyIsGap() {
	j := `[{"name":"a","sells":10},{"name":"b","stocks":7}]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Len(results, 2)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
	s.Equal([]string{"stocks"}, statTypes(results[1].Stats))
}

func (s *JSONSuite) TestMixedTypePerKeyNumericWhereParseable() {
	// v is a number in row 1, non-numeric string in row 2
	j := `[{"v":3},{"v":"foo"}]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	// v qualifies as a chart column (>=1 numeric); row 2 has no stats → dropped
	s.Len(results, 1)
	s.Equal(3.0, *results[0].Stats[0].Value)
}

func (s *JSONSuite) TestGroupSingleFieldToXAxis() {
	s.cfg.Group = []string{"name"}
	j := `[{"name":"alpha","sells":10},{"name":"beta","sells":20}]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

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
	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), cfg)

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

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Len(results, 1)
	s.Equal("alpha", results[0].Name)
	s.Equal("2024-01", results[0].XAxis)
}

func (s *JSONSuite) TestGroupOnNumericFieldStringified() {
	s.cfg.Group = []string{"id"}
	j := `[{"id":7,"sells":10}]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Len(results, 1)
	s.Equal("7", results[0].XAxis)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *JSONSuite) TestFilterRegexOnGroupLabel() {
	s.cfg.Group = []string{"name"}
	s.cfg.Filter = "keep"
	j := `[{"name":"keep_a","sells":10},{"name":"drop_b","sells":20},{"name":"keep_c","sells":30}]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Len(results, 2)
	for _, r := range results {
		s.Contains(r.XAxis, "keep")
	}
}

func (s *JSONSuite) TestInvalidFilterReturnsError() {
	s.cfg.Group = []string{"name"}
	s.cfg.Filter = "["

	err := parseJSONFileError(s.T(), s.writeFile(`[{"name":"keep","sells":10}]`), s.cfg)
	s.ErrorContains(err, "invalid filter regex")
}

func (s *JSONSuite) TestQuietAutoDetect() {
	s.cfg.AutoGroup = true
	s.cfg.ChartTypes = []string{"bar"}
	s.cfg.QuietAutoDetect = true

	results, effectiveCfg := mustParseJSONFile(
		s.T(),
		s.writeFile(`[{"name":"alpha","sells":10},{"name":"beta","sells":20}]`),
		s.cfg,
	)

	s.Require().Len(results, 2)
	s.Equal("alpha", results[0].XAxis)
	s.Equal([]string{"name"}, effectiveCfg.Group)
}

func (s *JSONSuite) TestNumberUnitScaling() {
	s.cfg.NumberUnit = "M"
	j := `[{"name":"a","sells":2000000}]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Len(results, 1)
	s.Equal("sells (M)", results[0].Stats[0].Type)
	s.Equal(2.0, *results[0].Stats[0].Value)
}

func (s *JSONSuite) TestNonArrayInputReturnsNil() {
	pts, _ := mustParseJSONFile(s.T(), s.writeFile(`{"name":"a","sells":10}`), s.cfg)
	s.Nil(pts)
	pts, _ = mustParseJSONFile(s.T(), s.writeFile(`[]`), s.cfg)
	s.Nil(pts)
	pts, _ = mustParseJSONFile(s.T(), s.writeFile(``), s.cfg)
	s.Nil(pts)
}

func (s *JSONSuite) TestParseJSONReturnsResultsAndErrors() {
	results, cfg, err := ParseJSON(strings.NewReader(`[{"name":"alpha","sells":10}]`), parser.Config{
		GroupPattern: "x",
		Group:        []string{"name"},
	})
	s.Require().NoError(err)
	s.Equal([]string{"name"}, cfg.Group)
	s.Require().Len(results, 1)
	s.Equal("alpha", results[0].XAxis)

	_, _, err = ParseJSON(strings.NewReader(`[{"name":"alpha","sells":10}]`), parser.Config{
		GroupPattern: "x",
		Group:        []string{"missing"},
	})
	s.ErrorContains(err, `group field "missing" not found`)

	_, _, err = ParseJSON(strings.NewReader(`[{"name":`), parser.Config{GroupPattern: "x"})
	s.ErrorContains(err, "read JSON")

	results, _, err = ParseJSON(strings.NewReader(`[{"name":"alpha","sells":10}]`), parser.Config{
		AutoGroup:  true,
		ChartTypes: []string{"scatter"},
	})
	s.Require().NoError(err)
	s.Len(results, 1)
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

// JSONErrorSuite covers parser failures returned to callers.
type JSONErrorSuite struct {
	suite.Suite
	cfg parser.Config
}

func (s *JSONErrorSuite) SetupTest() {
	s.cfg = parser.Config{GroupPattern: "x"}
}

func (s *JSONErrorSuite) writeFile(content string) string {
	path := filepath.Join(s.T().TempDir(), "data.json")
	s.Require().NoError(os.WriteFile(path, []byte(content), 0644))
	return path
}

func (s *JSONErrorSuite) TestMissingGroupFieldReturnsError() {
	s.cfg.Group = []string{"nope"}
	path := s.writeFile(`[{"name":"a","sells":10}]`)

	s.Error(parseJSONFileError(s.T(), path, s.cfg))
}

func (s *JSONErrorSuite) TestInvalidGroupConfigReturnsError() {
	s.cfg.Group = []string{"a", "b"}
	s.cfg.GroupPattern = "x"
	path := s.writeFile(`[{"a":"A","b":"B","sales":10}]`)

	s.Error(parseJSONFileError(s.T(), path, s.cfg))
}

func (s *JSONErrorSuite) TestNoNumericFieldsReturnsError() {
	path := s.writeFile(`[{"name":"a","label":"foo"}]`)

	s.Error(parseJSONFileError(s.T(), path, s.cfg))
}

func (s *JSONErrorSuite) TestAutoDetectGroupConfigErrorReturnsError() {
	s.cfg.AutoGroup = true
	s.cfg.ChartTypes = []string{"bar"}
	path := s.writeFile(`[{"a/b":"cat","sales":10}]`)

	s.Error(parseJSONFileError(s.T(), path, s.cfg))
}

func (s *JSONErrorSuite) TestMalformedMatrixReturnsError() {
	path := s.writeFile(`[["x"],[`)

	s.Error(parseJSONFileError(s.T(), path, s.cfg))
}

func (s *JSONErrorSuite) TestGroupTabularRowErrorReturnsError() {
	s.cfg.Group = []string{"a"}
	s.cfg.GroupRegex = ".*"
	path := s.writeFile(`[{"a":"A","sales":10}]`)

	s.Error(parseJSONFileError(s.T(), path, s.cfg))
}

func (s *JSONErrorSuite) TestExplicitColsMissingColumnErrors() {
	s.cfg.Select = []parser.ColumnSpec{{Source: "missing"}}
	path := s.writeFile(`[{"name":"a","sells":10}]`)

	s.Error(parseJSONFileError(s.T(), path, s.cfg))
}

func (s *JSONErrorSuite) TestExplicitColsNonNumericErrors() {
	s.cfg.Select = []parser.ColumnSpec{{Source: "name"}}
	path := s.writeFile(`[{"name":"alpha","sells":10}]`)

	s.Error(parseJSONFileError(s.T(), path, s.cfg))
}

func (s *JSONErrorSuite) TestValueModeMissingAxisFieldErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "missing"}, {Source: "y"}}
	path := s.writeFile(`[{"x":1,"y":2}]`)

	s.Error(parseJSONFileError(s.T(), path, s.cfg))
}

func (s *JSONErrorSuite) TestValueModeNonNumericAxisFieldErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "name"}, {Source: "y"}}
	path := s.writeFile(`[{"name":"alpha","y":2}]`)

	s.Error(parseJSONFileError(s.T(), path, s.cfg))
}

func (s *JSONErrorSuite) TestValueModeMetricFieldMissingErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}
	s.cfg.MetricColumn = "m"
	path := s.writeFile(`[{"x":1,"y":2}]`)

	s.Error(parseJSONFileError(s.T(), path, s.cfg))
}

func (s *JSONErrorSuite) TestValueModeMetricFieldNonNumericErrors() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}
	s.cfg.MetricColumn = "label"
	path := s.writeFile(`[{"x":1,"y":2,"label":"foo"}]`)

	s.Error(parseJSONFileError(s.T(), path, s.cfg))
}

func (s *JSONErrorSuite) TestValueModeSkipsRowWithBadMetric() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}
	s.cfg.MetricColumn = "m"
	path := s.writeFile(`[{"x":1,"y":2,"m":3},{"x":4,"y":5,"m":"bad"},{"x":6,"y":7,"m":8}]`)

	results, _ := mustParseJSONFile(s.T(), path, s.cfg)
	s.Len(results, 2)
	s.Equal("3", results[0].Metric)
	s.Equal("8", results[1].Metric)
}

func (s *JSONAutoValueSuite) TestSelectSkipsAutoDetect() {
	// Solo --select (SelectViews) disables auto-value inference and routes value mode.
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "x", AxisKey: "x"}, {Source: "y", AxisKey: "y"}}},
	}
	path := s.writeFile(`[{"x":1,"y":2,"z":3,"w":4}]`)

	results, _ := mustParseJSONFile(s.T(), path, s.cfg)
	s.Require().Len(results, 1)
	s.Equal("1", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
	s.Empty(results[0].Stats)
}

func (s *JSONErrorSuite) TestSelectMixedModeMapsCategoryXAndValueY() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "latency", AxisKey: "y"}}},
	}
	path := s.writeFile(`[
		{"region":"Asia","latency":12,"sales":100},
		{"region":"EU","latency":11,"sales":60}
	]`)

	results, _ := mustParseJSONFile(s.T(), path, s.cfg)
	s.Require().Len(results, 2)
	s.Equal("Asia", results[0].XAxis)
	s.Equal("12", results[0].YAxis)
	s.Empty(results[0].Stats)
}

func (s *JSONErrorSuite) TestSelectColumnNotFoundReturnsError() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "missing", AxisKey: "x"}, {Source: "latency", AxisKey: "y"}}},
	}
	path := s.writeFile(`[{"region":"Asia","latency":12}]`)
	s.Error(parseJSONFileError(s.T(), path, s.cfg))
}

func (s *JSONErrorSuite) TestSelectNonNumericYFieldReturnsError() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "label", AxisKey: "y"}}},
	}
	path := s.writeFile(`[{"region":"Asia","label":"fast"}]`)
	s.Error(parseJSONFileError(s.T(), path, s.cfg))
}

func (s *JSONErrorSuite) TestSelectEmptyFieldReturnsError() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "latency", AxisKey: "y"}}},
	}
	path := s.writeFile(`[{"region":"","latency":12}]`)
	s.Error(parseJSONFileError(s.T(), path, s.cfg))
}

func (s *JSONErrorSuite) TestSelectValueModeAllNumeric() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "x", AxisKey: "x"}, {Source: "y", AxisKey: "y"}}},
	}
	path := s.writeFile(`[{"x":1,"y":2},{"x":3,"y":4}]`)

	results, _ := mustParseJSONFile(s.T(), path, s.cfg)
	s.Require().Len(results, 2)
	s.Equal("1", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
	s.Empty(results[0].Stats)
}

func (s *JSONErrorSuite) TestSelectValueModeNoHeaderMatrix() {
	s.cfg.SelectViews = []parser.SelectView{
		{Columns: []parser.ColumnSpec{{Source: "x", AxisKey: "x"}, {Source: "y", AxisKey: "y"}}},
	}
	path := s.writeFile(`[[1,2],[3,4]]`)

	results, _ := mustParseJSONFile(s.T(), path, s.cfg)
	s.Require().Len(results, 2)
	s.Equal("1", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
	s.Empty(results[0].Stats)
}

func TestJSONErrorSuite(t *testing.T) {
	suite.Run(t, new(JSONErrorSuite))
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
	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("West", results[0].XAxis)
	s.Equal("East", results[1].XAxis)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *JSONAutoGroupSuite) TestHeaderMatrixCategoricalColumnBecomesXAxis() {
	j := `[["region","sales"],["West",10],["East",20]]`
	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("West", results[0].XAxis)
	s.Equal("East", results[1].XAxis)
	s.Equal([]string{"sales"}, statTypes(results[0].Stats))
}

func (s *JSONAutoGroupSuite) TestNestedFlattenedFieldChosen() {
	// nested object flattened to "addr.city"; along with region (also categorical)
	// the most-unique categorical wins. Both have 2 distinct here; leftmost in
	// first-seen order wins → region (appears first).
	j := `[{"region":"West","addr":{"city":"NY"},"sells":10},{"region":"East","addr":{"city":"LA"},"sells":20}]`
	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.NotEmpty(results[0].XAxis)
	s.Equal([]string{"sells"}, statTypes(results[0].Stats))
}

func (s *JSONAutoGroupSuite) TestAllNumericAutoValues() {
	// all numeric → auto-value-mode: first 2 cols become x,y value axes
	j := `[{"id":1,"sells":10},{"id":2,"sells":20},{"id":3,"sells":30}]`
	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)
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
	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)
	s.Require().NotEmpty(results)
	for _, r := range results {
		s.NotEmpty(r.XAxis)
		s.Empty(r.YAxis)
	}
}

func (s *JSONAutoGroupSuite) TestExplicitGroupDisablesAutoGroup() {
	s.cfg.Group = []string{"region"}
	j := `[{"region":"West","sells":10},{"region":"East","sells":20}]`
	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("West", results[0].XAxis)
	s.Empty(results[0].YAxis)
}

func (s *JSONAutoGroupSuite) TestAxesDisablesAutoGroup() {
	s.cfg.Axes = []parser.ColumnSpec{{Source: "sells"}}
	j := `[{"region":"West","sells":10},{"region":"East","sells":20}]`
	_, _ = mustParseJSONFile(s.T(), s.writeFile(j), s.cfg) // no panic; value mode handled elsewhere
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
	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("10", results[0].XAxis)
	s.Equal("5", results[0].YAxis)
	s.Empty(results[0].Stats)
}

func (s *JSONAutoValueSuite) TestNoHeaderMatrixTwoNumericColumnsAutoValue() {
	j := `[[1,2],[3,4],[5,6]]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Require().Len(results, 3)
	s.Equal("1", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
	s.Empty(results[0].Stats)
}

func (s *JSONAutoValueSuite) TestThreeNumericFields() {
	j := `[{"price":10,"latency":5,"mem":100},{"price":20,"latency":7,"mem":200}]`
	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("10", results[0].XAxis)
	s.Equal("5", results[0].YAxis)
	s.Equal("100", results[0].ZAxis)
	s.Empty(results[0].Stats)
}

func (s *JSONAutoValueSuite) TestNoHeaderMatrixFourNumericColumnsUsesMetric() {
	j := `[[1,2,3,4],[5,6,7,8]]`

	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)

	s.Require().Len(results, 2)
	s.Equal("1", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
	s.Equal("3", results[0].ZAxis)
	s.Equal("4", results[0].Metric)
	s.Empty(results[0].Stats)
}

func (s *JSONAutoValueSuite) TestNestedMatrixViaJSONPathAutoValue() {
	source := s.writeFile(`{"payload":{"rows":[[1,2],[3,4]]}}`)
	selected, err := SelectPath(source, ".payload.rows")
	s.Require().NoError(err)

	results, _ := mustParseJSONFile(s.T(), s.writeFile(string(selected)), s.cfg)

	s.Require().Len(results, 2)
	s.Equal("1", results[0].XAxis)
	s.Equal("2", results[0].YAxis)
	s.Empty(results[0].Stats)
}

func (s *JSONAutoValueSuite) TestFourNumericFieldsTakeFirstThree() {
	j := `[{"a":1,"b":2,"c":3,"d":4},{"a":5,"b":6,"c":7,"d":8}]`
	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)
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
	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Equal("West", results[0].XAxis)
	s.NotEmpty(results[0].Stats)
}

func (s *JSONAutoValueSuite) TestOneNumericFieldFallsBackToFlat() {
	j := `[{"price":10},{"price":20}]`
	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Empty(results[0].XAxis)
	s.NotEmpty(results[0].Stats)
}

func (s *JSONAutoValueSuite) TestOneColumnMatrixFallsBackToFlat() {
	j := `[[10],[20]]`
	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Empty(results[0].XAxis)
	s.Equal([]string{"x"}, statTypes(results[0].Stats))
}

func (s *JSONAutoValueSuite) TestPieChartFallsBackToFlat() {
	s.cfg.ChartTypes = []string{"pie"}
	j := `[{"price":10,"latency":5},{"price":20,"latency":7}]`
	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Empty(results[0].XAxis)
	s.NotEmpty(results[0].Stats)
}

func (s *JSONAutoValueSuite) TestRadarChartFallsBackToFlat() {
	s.cfg.ChartTypes = []string{"radar"}
	j := `[{"price":10,"latency":5},{"price":20,"latency":7}]`
	results, _ := mustParseJSONFile(s.T(), s.writeFile(j), s.cfg)
	s.Require().Len(results, 2)
	s.Empty(results[0].XAxis)
	s.NotEmpty(results[0].Stats)
}

func TestJSONAutoValueSuite(t *testing.T) {
	suite.Run(t, new(JSONAutoValueSuite))
}
