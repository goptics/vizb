package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TabularSuite struct {
	suite.Suite
}

type mockRowReader struct {
	cells   map[string]string
	numeric map[string]float64
	headers []string
	flag    string
}

func (m mockRowReader) Cell(source string) (string, bool) {
	v, ok := m.cells[source]
	return v, ok
}

func (m mockRowReader) Numeric(source string) (float64, bool) {
	v, ok := m.numeric[source]
	return v, ok
}

func (m mockRowReader) AvailableColumns() []string { return m.headers }
func (m mockRowReader) FlagLabel() string          { return m.flag }

func (s *TabularSuite) TestParseMixedMode() {
	t := s.T()
	cfg := Config{Axes: []ColumnSpec{
		{Source: "region", AxisKey: "x", AxisType: "category"},
		{Source: "latency", AxisKey: "y", AxisType: "value"},
		{Source: "sales", AxisKey: "z", AxisType: "value"},
	}}
	rows := []RowReader{
		mockRowReader{
			cells:   map[string]string{"region": "Asia"},
			numeric: map[string]float64{"latency": 12, "sales": 100},
		},
		mockRowReader{
			cells: map[string]string{"region": ""},
		},
	}
	results := ParseMixedMode(rows, cfg)
	if len(results) != 1 {
		t.Fatalf("want 1 complete row, got %d", len(results))
	}
	if results[0].XAxis != "Asia" || results[0].YAxis != "12" || results[0].ZAxis != "100" {
		t.Fatalf("unexpected point: %+v", results[0])
	}
}

func (s *TabularSuite) TestParseValueModeSkipsIncompleteRowsAndSetsMetric() {
	t := s.T()
	cfg := Config{
		Axes: []ColumnSpec{
			{Source: "x", AxisKey: "x"},
			{Source: "y", AxisKey: "y"},
		},
		MetricColumn: "noise",
	}
	rows := []RowReader{
		mockRowReader{
			cells:   map[string]string{"noise": "0.5"},
			numeric: map[string]float64{"x": 1, "y": 2, "noise": 0.5},
			headers: []string{"x", "y", "noise"},
		},
		mockRowReader{
			numeric: map[string]float64{"x": 3},
			headers: []string{"x", "y", "noise"},
		},
	}
	results, err := ParseValueMode(rows, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 row, got %d", len(results))
	}
	if results[0].Metric != "0.5" {
		t.Fatalf("metric = %q", results[0].Metric)
	}
}

func (s *TabularSuite) TestParseValueModeRejectsMissingMetricColumn() {
	cfg := Config{
		Axes:         []ColumnSpec{{Source: "x", AxisKey: "x"}, {Source: "y", AxisKey: "y"}},
		MetricColumn: "noise",
	}
	rows := []RowReader{
		mockRowReader{
			headers: []string{"x", "y"},
			flag:    "--select",
			numeric: map[string]float64{"x": 1, "y": 2},
		},
	}
	_, err := ParseValueMode(rows, cfg)
	s.ErrorContains(err, `metric column "noise" not found`)
}

func (s *TabularSuite) TestParseSelectStatModeMergedRow() {
	t := s.T()
	cfg := Config{
		SelectViews: []SelectView{
			{Columns: []ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "tax", AxisKey: "y"}}},
			{Columns: []ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "sales", AxisKey: "y"}}},
		},
	}
	rows := []RowReader{
		mockRowReader{
			cells:   map[string]string{"region": "Asia"},
			numeric: map[string]float64{"tax": 12, "sales": 100},
		},
	}
	results, err := ParseSelectStatMode(rows, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].XAxis != "Asia" || len(results[0].Stats) != 2 {
		t.Fatalf("unexpected merged stat point: %+v", results)
	}
}

func (s *TabularSuite) TestParseSelectStatModeEmptyResultsReturnsError() {
	cfg := Config{
		SelectViews: []SelectView{
			{Columns: []ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "tax", AxisKey: "y"}}},
		},
	}
	rows := []RowReader{
		mockRowReader{cells: map[string]string{"region": ""}},
	}
	_, err := ParseSelectStatMode(rows, cfg)
	s.ErrorContains(err, "no dataset found")
}

func (s *TabularSuite) TestParseMixedModeSkipsUnknownAxisKey() {
	t := s.T()
	cfg := Config{Axes: []ColumnSpec{
		{Source: "region", AxisKey: "x", AxisType: "category"},
		{Source: "missing", AxisKey: "w", AxisType: "value"},
	}}
	rows := []RowReader{
		mockRowReader{cells: map[string]string{"region": "Asia"}},
	}
	if got := ParseMixedMode(rows, cfg); len(got) != 0 {
		t.Fatalf("want 0 rows, got %+v", got)
	}
}

func (s *TabularSuite) TestParseValueModeRejectsNonNumericMetric() {
	cfg := Config{
		Axes:         []ColumnSpec{{Source: "x", AxisKey: "x"}, {Source: "y", AxisKey: "y"}},
		MetricColumn: "noise",
	}
	rows := []RowReader{
		mockRowReader{
			cells:   map[string]string{"noise": "loud"},
			numeric: map[string]float64{"x": 1, "y": 2},
			headers: []string{"x", "y", "noise"},
			flag:    "--select",
		},
	}
	_, err := ParseValueMode(rows, cfg)
	s.ErrorContains(err, `metric column "noise" is not numeric`)
}

func (s *TabularSuite) TestTabularParsersReturnErrors() {
	rows := []RowReader{mockRowReader{
		cells:   map[string]string{"metric": "4"},
		numeric: map[string]float64{"x": 1, "y": 2, "metric": 4},
		headers: []string{"x", "y", "metric"},
		flag:    "--axes",
	}}
	points, err := ParseValueMode(rows, Config{
		Axes:         []ColumnSpec{{Source: "x", AxisKey: "x"}, {Source: "y", AxisKey: "y"}},
		MetricColumn: "metric",
	})
	s.Require().NoError(err)
	s.Require().Len(points, 1)
	s.Equal("4", points[0].Metric)

	_, err = ParseValueMode(rows, Config{MetricColumn: "missing"})
	s.ErrorContains(err, `metric column "missing" not found`)

	_, err = ParseSelectStatMode([]RowReader{mockRowReader{}}, Config{SelectViews: []SelectView{{
		Columns: []ColumnSpec{{Source: "region"}, {Source: "sales"}},
	}}})
	s.ErrorContains(err, "no dataset found")
}

func (s *TabularSuite) TestDispatchSelectMode() {
	cfg := Config{Mode: ModeValue, SelectViews: []SelectView{{
		Columns: []ColumnSpec{{Source: "x", AxisKey: "x"}, {Source: "y", AxisKey: "y"}},
	}}}
	rows := []RowReader{mockRowReader{numeric: map[string]float64{"x": 1, "y": 2}}}
	points, err := DispatchSelectMode(rows, &cfg, func(string, string) (string, error) { return "value", nil })
	s.Require().NoError(err)
	s.Require().Len(points, 1)
	s.Equal(ModeValue, cfg.Mode)

	_, err = DispatchSelectMode(rows, &cfg, func(string, string) (string, error) {
		return "", fmt.Errorf("invalid axis")
	})
	s.ErrorContains(err, "invalid axis")
}

func (s *TabularSuite) TestParseSelectStatModeNonMerge() {
	t := s.T()
	cfg := Config{
		SelectViews: []SelectView{
			{Columns: []ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "tax", AxisKey: "y"}}},
			{Columns: []ColumnSpec{{Source: "product", AxisKey: "x"}, {Source: "sales", AxisKey: "y"}}},
		},
	}
	rows := []RowReader{
		mockRowReader{
			cells:   map[string]string{"region": "Asia", "product": "Widget"},
			numeric: map[string]float64{"tax": 12, "sales": 100},
		},
	}
	results, err := ParseSelectStatMode(rows, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("want 2 points, got %d", len(results))
	}
}

func (s *TabularSuite) TestDispatchSelectModePropagatesAxisType() {
	t := s.T()
	cfg := Config{
		Mode: ModeValue,
		SelectViews: []SelectView{
			{Columns: []ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "latency", AxisKey: "y"}}},
		},
	}
	rows := []RowReader{
		mockRowReader{
			cells:   map[string]string{"region": "Asia"},
			numeric: map[string]float64{"latency": 12},
		},
	}
	kindFn := func(source, axisKey string) (string, error) {
		if source == "region" {
			return "category", nil
		}
		return "value", nil
	}
	results, err := DispatchSelectMode(rows, &cfg, kindFn)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].XAxis != "Asia" {
		t.Fatalf("unexpected results: %+v", results)
	}
	if cfg.Mode != ModeMixed {
		t.Fatalf("mode = %v, want ModeMixed", cfg.Mode)
	}
	if cfg.SelectViews[0].Columns[0].AxisType != "category" || cfg.SelectViews[0].Columns[1].AxisType != "value" {
		t.Fatalf("axis types not propagated: %+v", cfg.SelectViews[0].Columns)
	}
}

func TestTabularSuite(t *testing.T) {
	suite.Run(t, new(TabularSuite))
}
