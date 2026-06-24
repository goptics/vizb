package parser

import "testing"

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
