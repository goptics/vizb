package parser

import "testing"

func TestParseSelectViewFlagTwoColumns(t *testing.T) {
	specs, err := ParseSelectViewFlag("region,latency")
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 2 {
		t.Fatalf("want 2 specs, got %d", len(specs))
	}
	if specs[0].AxisKey != "x" || specs[0].Source != "region" {
		t.Fatalf("axis[0] wrong: %+v", specs[0])
	}
	if specs[1].AxisKey != "y" || specs[1].Source != "latency" {
		t.Fatalf("axis[1] wrong: %+v", specs[1])
	}
}

func TestParseSelectViewFlagThreeColumnsWithLabel(t *testing.T) {
	specs, err := ParseSelectViewFlag("region{Region},latency{Latency (ms)},sales")
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 3 || specs[0].Label != "Region" || specs[2].AxisKey != "z" {
		t.Fatalf("unexpected specs: %+v", specs)
	}
}

func TestParseSelectViewFlagExplicitSyntax(t *testing.T) {
	specs, err := ParseSelectViewFlag("x:region,y:latency,z:sales")
	if err != nil {
		t.Fatal(err)
	}
	want := []struct{ key, src string }{
		{"x", "region"},
		{"y", "latency"},
		{"z", "sales"},
	}
	for i, w := range want {
		if specs[i].AxisKey != w.key || specs[i].Source != w.src {
			t.Fatalf("spec[%d] = %+v, want key=%s src=%s", i, specs[i], w.key, w.src)
		}
	}
}

func TestParseSelectViewFlagRejectsArity(t *testing.T) {
	if _, err := ParseSelectViewFlag("region"); err == nil {
		t.Fatal("want error for 1 column")
	}
	if _, err := ParseSelectViewFlag("a,b,c,d"); err == nil {
		t.Fatal("want error for 4 columns")
	}
	if _, err := ParseSelectViewFlag(""); err == nil {
		t.Fatal("want error for empty")
	}
}

func TestParseSelectViewFlagRejectsDuplicateColumn(t *testing.T) {
	if _, err := ParseSelectViewFlag("region,region"); err == nil {
		t.Fatal("want duplicate column error")
	}
}

func TestParseSelectViewFlagRejectsIncompleteExplicitSyntax(t *testing.T) {
	if _, err := ParseSelectViewFlag("x:region,latency"); err == nil {
		t.Fatal("want mixed explicit/implicit error")
	}
	if _, err := ParseSelectViewFlag("y:latency,z:sales"); err == nil {
		t.Fatal("want missing x: error")
	}
}

func TestHasSelect(t *testing.T) {
	if HasSelect(Config{}) {
		t.Fatal("expected false for empty config")
	}
	if !HasSelect(Config{Select: []ColumnSpec{{Source: "a"}}}) {
		t.Fatal("expected true for grouped select")
	}
	if !HasSelect(Config{SelectViews: [][]ColumnSpec{{{Source: "a"}, {Source: "b"}}}}) {
		t.Fatal("expected true for select views")
	}
}

func TestIsExplicitGrouping(t *testing.T) {
	if IsExplicitGrouping(Config{}) {
		t.Fatal("expected false for empty config")
	}
	if IsExplicitGrouping(Config{GroupPattern: "x"}) {
		t.Fatal("expected false for default pattern")
	}
	if !IsExplicitGrouping(Config{Group: []string{"region"}}) {
		t.Fatal("expected true for --group")
	}
	if !IsExplicitGrouping(Config{GroupRegex: ".*"}) {
		t.Fatal("expected true for --group-regex")
	}
	if !IsExplicitGrouping(Config{GroupPattern: "x,y"}) {
		t.Fatal("expected true for custom pattern")
	}
}

func TestIsSelectAxisMode(t *testing.T) {
	cfg := Config{
		SelectViews: [][]ColumnSpec{{{Source: "a"}, {Source: "b"}}},
	}
	if !IsSelectAxisMode(cfg) {
		t.Fatal("expected solo select axis mode")
	}
	cfg.GroupPattern = "x"
	if !IsSelectAxisMode(cfg) {
		t.Fatal("expected solo select axis mode with default pattern")
	}
	cfg.Group = []string{"region"}
	if IsSelectAxisMode(cfg) {
		t.Fatal("expected false when grouped")
	}
	if IsSelectAxisMode(Config{Select: []ColumnSpec{{Source: "price"}}}) {
		t.Fatal("expected false for grouped numeric select")
	}
}

func TestSelectViewDatasetName(t *testing.T) {
	view := []ColumnSpec{
		{Source: "region", Label: "Region", AxisKey: "x"},
		{Source: "latency", AxisKey: "y"},
	}
	if got := SelectViewDatasetName(view, 0); got != "Region × latency" {
		t.Fatalf("got %q, want Region × latency", got)
	}
	if got := SelectViewDatasetName(nil, 2); got != "View 3" {
		t.Fatalf("got %q, want View 3", got)
	}
}
