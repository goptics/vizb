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
