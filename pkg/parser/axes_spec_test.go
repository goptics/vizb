package parser

import "testing"

func TestParseAxesFlagTwoColumns(t *testing.T) {
	specs, err := ParseAxesFlag("price,latency")
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 2 || specs[0].Source != "price" || specs[1].Source != "latency" {
		t.Fatalf("unexpected specs: %+v", specs)
	}
}

func TestParseAxesFlagThreeColumnsWithLabel(t *testing.T) {
	specs, err := ParseAxesFlag("price{Price (USD)},latency,mem")
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 3 || specs[0].Label != "Price (USD)" {
		t.Fatalf("unexpected specs: %+v", specs)
	}
}

func TestParseAxesFlagRejectsArity(t *testing.T) {
	if _, err := ParseAxesFlag("price"); err == nil {
		t.Fatal("expected error for 1 column")
	}
	if _, err := ParseAxesFlag("a,b,c,d"); err == nil {
		t.Fatal("expected error for 4 columns")
	}
	if _, err := ParseAxesFlag(""); err == nil {
		t.Fatal("expected error for empty")
	}
}

func TestParseAxesFlagRejectsDuplicate(t *testing.T) {
	if _, err := ParseAxesFlag("price,price"); err == nil {
		t.Fatal("expected duplicate-column error")
	}
}

func TestValueAxes(t *testing.T) {
	cfg := Config{Axes: []ColumnSpec{{Source: "price"}, {Source: "latency", Label: "Lat"}}}
	axes := ValueAxes(cfg)
	if len(axes) != 2 {
		t.Fatalf("want 2 axes, got %d", len(axes))
	}
	if axes[0].Key != "x" || axes[0].Label != "price" || axes[0].Type != "value" {
		t.Fatalf("axis[0] wrong: %+v", axes[0])
	}
	if axes[1].Key != "y" || axes[1].Label != "Lat" || axes[1].Type != "value" {
		t.Fatalf("axis[1] wrong: %+v", axes[1])
	}
}

func TestIsHybridMode(t *testing.T) {
	hybrid := Config{
		Group:        []string{"region", "category"},
		GroupPattern: "x,y",
		Axes:         []ColumnSpec{{Source: "latency"}},
	}
	if !IsHybridMode(hybrid) {
		t.Fatal("expected hybrid mode")
	}
	if IsHybridMode(Config{Axes: []ColumnSpec{{Source: "x"}, {Source: "y"}}}) {
		t.Fatal("pure value mode should not be hybrid")
	}
	if IsHybridMode(Config{Group: []string{"a"}, GroupPattern: "x"}) {
		t.Fatal("group-only should not be hybrid")
	}
}

func TestValueAxesThreeColumns(t *testing.T) {
	cfg := Config{Axes: []ColumnSpec{
		{Source: "x", Label: "X"},
		{Source: "y", Label: "Y"},
		{Source: "z", Label: "Z"},
	}}
	axes := ValueAxes(cfg)
	if len(axes) != 3 {
		t.Fatalf("want 3 axes, got %d", len(axes))
	}
	if axes[0].Label != "X" || axes[1].Label != "Y" || axes[2].Key != "z" || axes[2].Label != "Z" {
		t.Fatalf("unexpected axes: %+v", axes)
	}
}

func TestHybridAxesFallsBackToSourceLabel(t *testing.T) {
	cfg := Config{
		Group:        []string{"region", "category"},
		GroupPattern: "x,y",
		Axes:         []ColumnSpec{{Source: "latency"}},
	}
	axes := HybridAxes(cfg)
	if len(axes) != 3 {
		t.Fatalf("want 3 axes, got %d", len(axes))
	}
	if axes[2].Label != "latency" {
		t.Fatalf("expected source fallback label, got %+v", axes[2])
	}
}

func TestHybridAxesIgnoresExtraGroupDims(t *testing.T) {
	cfg := Config{
		Group:        []string{"region", "category"},
		GroupPattern: "x,y,z",
		Axes:         []ColumnSpec{{Source: "latency"}},
	}
	axes := HybridAxes(cfg)
	if len(axes) != 3 {
		t.Fatalf("want 3 axes, got %d: %+v", len(axes), axes)
	}
	if axes[0].Key != "x" || axes[1].Key != "y" || axes[2].Key != "z" {
		t.Fatalf("unexpected axis keys: %+v", axes)
	}
}

func TestHybridAxes(t *testing.T) {
	cfg := Config{
		Group:        []string{"region", "category"},
		GroupPattern: "x,y",
		Axes:         []ColumnSpec{{Source: "latency", Label: "Latency (ms)"}},
	}
	axes := HybridAxes(cfg)
	if len(axes) != 3 {
		t.Fatalf("want 3 axes, got %d: %+v", len(axes), axes)
	}
	if axes[0].Key != "x" || axes[0].Label != "region" || axes[0].Type != "" {
		t.Fatalf("axis[0] wrong: %+v", axes[0])
	}
	if axes[1].Key != "y" || axes[1].Label != "category" || axes[1].Type != "" {
		t.Fatalf("axis[1] wrong: %+v", axes[1])
	}
	if axes[2].Key != "z" || axes[2].Label != "Latency (ms)" || axes[2].Type != "value" {
		t.Fatalf("axis[2] wrong: %+v", axes[2])
	}
}
