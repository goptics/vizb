package parser

import (
	"testing"

	"github.com/goptics/vizb/shared"
)

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

func TestParseMixedMode(t *testing.T) {
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

func TestParseValueModeSkipsIncompleteRowsAndSetsMetric(t *testing.T) {
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
	results := ParseValueMode(rows, cfg)
	if len(results) != 1 {
		t.Fatalf("want 1 row, got %d", len(results))
	}
	if results[0].Metric != "0.5" {
		t.Fatalf("metric = %q", results[0].Metric)
	}
}

func TestParseValueModeRejectsMissingMetricColumn(t *testing.T) {
	restore, exitCalled := shared.TrapOsExitPanic(t)
	defer restore()

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
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		ParseValueMode(rows, cfg)
	}()
	if !*exitCalled {
		t.Fatal("expected OsExit")
	}
}

func TestParseSelectStatModeMergedRow(t *testing.T) {
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
	results := ParseSelectStatMode(rows, cfg)
	if len(results) != 1 || results[0].XAxis != "Asia" || len(results[0].Stats) != 2 {
		t.Fatalf("unexpected merged stat point: %+v", results)
	}
}

func TestParseSelectStatModeEmptyResultsExits(t *testing.T) {
	restore, exitCalled := shared.TrapOsExitPanic(t)
	defer restore()

	cfg := Config{
		SelectViews: []SelectView{
			{Columns: []ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "tax", AxisKey: "y"}}},
		},
	}
	rows := []RowReader{
		mockRowReader{cells: map[string]string{"region": ""}},
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		ParseSelectStatMode(rows, cfg)
	}()
	if !*exitCalled {
		t.Fatal("expected OsExit")
	}
}

func TestParseMixedModeSkipsUnknownAxisKey(t *testing.T) {
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

func TestParseValueModeRejectsNonNumericMetric(t *testing.T) {
	restore, exitCalled := shared.TrapOsExitPanic(t)
	defer restore()

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
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		ParseValueMode(rows, cfg)
	}()
	if !*exitCalled {
		t.Fatal("expected OsExit")
	}
}

func TestParseSelectStatModeNonMerge(t *testing.T) {
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
	results := ParseSelectStatMode(rows, cfg)
	if len(results) != 2 {
		t.Fatalf("want 2 points, got %d", len(results))
	}
}

func TestDispatchSelectModePropagatesAxisType(t *testing.T) {
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
	results := DispatchSelectMode(rows, &cfg, kindFn)
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
