package parser

import (
	"fmt"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

type AxesSpecSuite struct {
	suite.Suite
}

func (s *AxesSpecSuite) TestResolveAxesTypesMixed() {
	t := s.T()
	cfg := Config{Axes: []ColumnSpec{
		{Source: "region", AxisKey: "x"},
		{Source: "latency", AxisKey: "y"},
		{Source: "sales", AxisKey: "z"},
	}}
	kindFn := func(source, axisKey string) (string, error) {
		if source == "region" {
			return "category", nil
		}
		return "value", nil
	}
	if err := ResolveAxesTypes(&cfg, kindFn); err != nil {
		t.Fatal(err)
	}
	if !IsMixedAxes(cfg) {
		t.Fatal("expected mixed mode")
	}
	axes := MixedAxes(cfg)
	if len(axes) != 3 || axes[0].Type != "" || axes[1].Type != "value" || axes[2].Type != "value" {
		t.Fatalf("unexpected mixed axes: %+v", axes)
	}
}

func (s *AxesSpecSuite) TestResolveAxesTypesAllValue() {
	t := s.T()
	cfg := Config{Axes: []ColumnSpec{
		{Source: "x", AxisKey: "x"},
		{Source: "y", AxisKey: "y"},
	}}
	kindFn := func(_, _ string) (string, error) { return "value", nil }
	if err := ResolveAxesTypes(&cfg, kindFn); err != nil {
		t.Fatal(err)
	}
	if IsMixedAxes(cfg) {
		t.Fatal("expected pure value mode")
	}
}

func (s *AxesSpecSuite) TestResolveAxesTypesRejectsCategoryNotOnX() {
	t := s.T()
	cfg := Config{Axes: []ColumnSpec{
		{Source: "latency", AxisKey: "x"},
		{Source: "region", AxisKey: "y"},
	}}
	kindFn := func(source, _ string) (string, error) {
		if source == "region" {
			return "category", nil
		}
		return "value", nil
	}
	if err := ResolveAxesTypes(&cfg, kindFn); err == nil {
		t.Fatal("expected error when category is not on x")
	}
}

func (s *AxesSpecSuite) TestResolveAxesTypesRejectsMultipleCategories() {
	t := s.T()
	cfg := Config{Axes: []ColumnSpec{
		{Source: "region", AxisKey: "x"},
		{Source: "product", AxisKey: "y"},
		{Source: "latency", AxisKey: "z"},
	}}
	kindFn := func(source, _ string) (string, error) {
		if source == "latency" {
			return "value", nil
		}
		return "category", nil
	}
	if err := ResolveAxesTypes(&cfg, kindFn); err == nil {
		t.Fatal("expected error for multiple categoricals")
	}
}

func (s *AxesSpecSuite) TestDatasetAxesForSelectViewMixed() {
	t := s.T()
	view := []ColumnSpec{
		{Source: "region", AxisKey: "x", Label: "Region", AxisType: "category"},
		{Source: "latency", AxisKey: "y", AxisType: "value"},
	}
	results := []shared.DataPoint{{XAxis: "Asia", YAxis: "12"}}
	axes := DatasetAxesForSelectView(view, results)
	if len(axes) != 2 || axes[0].Type != "" || axes[1].Type != "value" {
		t.Fatalf("unexpected mixed select axes: %+v", axes)
	}
}

func (s *AxesSpecSuite) TestDatasetAxesForSelectViewValue() {
	t := s.T()
	view := []ColumnSpec{
		{Source: "x", AxisKey: "x", AxisType: "value"},
		{Source: "y", AxisKey: "y", AxisType: "value"},
	}
	results := []shared.DataPoint{{XAxis: "1", YAxis: "2"}}
	axes := DatasetAxesForSelectView(view, results)
	if len(axes) != 2 || axes[0].Type != "value" || axes[1].Type != "value" {
		t.Fatalf("unexpected value select axes: %+v", axes)
	}
}

func (s *AxesSpecSuite) TestValueAxes() {
	t := s.T()
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

func (s *AxesSpecSuite) TestValueAxesThreeColumns() {
	t := s.T()
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

func (s *AxesSpecSuite) TestResolveAxesTypesEmptyAxes() {
	t := s.T()
	cfg := Config{}
	err := ResolveAxesTypes(&cfg, func(_, _ string) (string, error) {
		return "", fmt.Errorf("kindFn should not run")
	})
	if err != nil {
		t.Fatal(err)
	}
}

func (s *AxesSpecSuite) TestResolveAxesTypesKindFnError() {
	t := s.T()
	cfg := Config{Axes: []ColumnSpec{{Source: "x", AxisKey: "x"}}}
	err := ResolveAxesTypes(&cfg, func(_, _ string) (string, error) {
		return "", fmt.Errorf("kind error")
	})
	if err == nil || err.Error() != "kind error" {
		t.Fatalf("got %v", err)
	}
}

func (s *AxesSpecSuite) TestDatasetAxesForSelectViewInfersMixed() {
	t := s.T()
	view := []ColumnSpec{
		{Source: "region", AxisKey: "x"},
		{Source: "latency", AxisKey: "y"},
	}
	results := []shared.DataPoint{{XAxis: "Asia", YAxis: "12"}}
	axes := DatasetAxesForSelectView(view, results)
	if len(axes) != 2 || axes[0].Type != "" || axes[1].Type != "value" {
		t.Fatalf("unexpected inferred mixed axes: %+v", axes)
	}
}

func (s *AxesSpecSuite) TestDatasetAxesForSelectViewInfersMixedAfterEmptyXAxis() {
	t := s.T()
	view := []ColumnSpec{
		{Source: "region", AxisKey: "x"},
		{Source: "latency", AxisKey: "y"},
	}
	results := []shared.DataPoint{
		{XAxis: "", YAxis: "12"},
		{XAxis: "Asia", YAxis: "12"},
	}
	axes := DatasetAxesForSelectView(view, results)
	if len(axes) != 2 || axes[0].Type != "" || axes[1].Type != "value" {
		t.Fatalf("unexpected inferred mixed axes: %+v", axes)
	}
}

func (s *AxesSpecSuite) TestDatasetAxesForSelectViewInfersValue() {
	t := s.T()
	view := []ColumnSpec{
		{Source: "x", AxisKey: "x"},
		{Source: "y", AxisKey: "y"},
	}
	results := []shared.DataPoint{{XAxis: "1.5", YAxis: "2"}}
	axes := DatasetAxesForSelectView(view, results)
	if len(axes) != 2 || axes[0].Type != "value" || axes[1].Type != "value" {
		t.Fatalf("unexpected inferred value axes: %+v", axes)
	}
}

func (s *AxesSpecSuite) TestDatasetAxesForSelectViewSkipsInferenceWhenEmptyResults() {
	t := s.T()
	view := []ColumnSpec{
		{Source: "region", AxisKey: "x"},
		{Source: "latency", AxisKey: "y"},
	}
	axes := DatasetAxesForSelectView(view, nil)
	if len(axes) != 2 || axes[0].Type != "value" || axes[1].Type != "value" {
		t.Fatalf("expected value axes when results empty: %+v", axes)
	}
}

func (s *AxesSpecSuite) TestValueAxesExplicitAxisKeyAndCap() {
	t := s.T()
	cfg := Config{Axes: []ColumnSpec{
		{Source: "a", AxisKey: "y"},
		{Source: "b", AxisKey: "x"},
		{Source: "c", AxisKey: "z"},
		{Source: "d", AxisKey: "w"},
	}}
	axes := ValueAxes(cfg)
	if len(axes) != 3 {
		t.Fatalf("want 3 axes, got %d", len(axes))
	}
	if axes[0].Key != "y" || axes[1].Key != "x" || axes[2].Key != "z" {
		t.Fatalf("unexpected axis keys: %+v", axes)
	}
}

func TestAxesSpecSuite(t *testing.T) {
	suite.Run(t, new(AxesSpecSuite))
}
