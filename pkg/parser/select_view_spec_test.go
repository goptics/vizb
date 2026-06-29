package parser

import (
	"testing"

	"github.com/goptics/vizb/shared"
)

func TestParseSelectViewFlagTwoColumns(t *testing.T) {
	view, err := ParseSelectViewFlag("region,latency")
	if err != nil {
		t.Fatal(err)
	}
	specs := view.Columns
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
	view, err := ParseSelectViewFlag("region{Region},latency{Latency (ms)},sales")
	if err != nil {
		t.Fatal(err)
	}
	specs := view.Columns
	if len(specs) != 3 || specs[0].Label != "Region" || specs[2].AxisKey != "z" {
		t.Fatalf("unexpected specs: %+v", specs)
	}
}

func TestParseSelectViewFlagExplicitSyntax(t *testing.T) {
	view, err := ParseSelectViewFlag("x:region,y:latency,z:sales")
	if err != nil {
		t.Fatal(err)
	}
	specs := view.Columns
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

func TestParseSelectViewFlagTrailingParenTypeLabel(t *testing.T) {
	view, err := ParseSelectViewFlag("region,latency (Latency by Region)")
	if err != nil {
		t.Fatal(err)
	}
	if view.TypeLabel != "Latency by Region" {
		t.Fatalf("TypeLabel = %q", view.TypeLabel)
	}
	if len(view.Columns) != 2 || view.Columns[1].Source != "latency" {
		t.Fatalf("unexpected columns: %+v", view.Columns)
	}
}

func TestParseSelectViewFlagAxisLabelAndParenTypeLabel(t *testing.T) {
	view, err := ParseSelectViewFlag("region{Region},latency (Custom Title)")
	if err != nil {
		t.Fatal(err)
	}
	if view.TypeLabel != "Custom Title" {
		t.Fatalf("TypeLabel = %q", view.TypeLabel)
	}
	if view.Columns[0].Label != "Region" {
		t.Fatalf("axis label = %q", view.Columns[0].Label)
	}
}

func TestParseSelectViewFlagParenOverridesMetricBraceLabel(t *testing.T) {
	view, err := ParseSelectViewFlag("region,latency{Legacy} (New Title)")
	if err != nil {
		t.Fatal(err)
	}
	if view.TypeLabel != "New Title" {
		t.Fatalf("TypeLabel = %q", view.TypeLabel)
	}
	if got := SelectStatType(view); got != "New Title" {
		t.Fatalf("SelectStatType = %q", got)
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

func TestParseSelectViewFlagRejectsEmptyParenTitle(t *testing.T) {
	if _, err := ParseSelectViewFlag("region,latency ()"); err == nil {
		t.Fatal("want error for empty ()")
	}
}

func TestHasSelect(t *testing.T) {
	if HasSelect(Config{}) {
		t.Fatal("expected false for empty config")
	}
	if !HasSelect(Config{Select: []ColumnSpec{{Source: "a"}}}) {
		t.Fatal("expected true for grouped select")
	}
	if !HasSelect(Config{SelectViews: []SelectView{{Columns: []ColumnSpec{{Source: "a"}, {Source: "b"}}}}}) {
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

func TestResolveMode(t *testing.T) {
	if ResolveMode(Config{}) != ModeAuto {
		t.Fatal("empty config should be ModeAuto")
	}
	cfg := Config{SelectViews: []SelectView{{Columns: []ColumnSpec{{Source: "a"}, {Source: "b"}}}}}
	if m := ResolveMode(cfg); m != ModeValue {
		t.Fatalf("single solo view should be ModeValue, got %d", m)
	}
	cfg.SelectViews = append(cfg.SelectViews, SelectView{Columns: []ColumnSpec{{Source: "a"}, {Source: "c"}}})
	if m := ResolveMode(cfg); m != ModeMultiStat {
		t.Fatalf("two solo views should be ModeMultiStat, got %d", m)
	}
	cfg.Group = []string{"region"}
	cfg.Select = []ColumnSpec{{Source: "price"}}
	if m := ResolveMode(cfg); m != ModeGrouped {
		t.Fatalf("grouped + select should be ModeGrouped, got %d", m)
	}
}

func TestSelectStatType(t *testing.T) {
	view := SelectView{Columns: []ColumnSpec{{Source: "region"}, {Source: "latency"}}}
	if got := SelectStatType(view); got != "latency by region" {
		t.Fatalf("got %q", got)
	}
	view = SelectView{Columns: []ColumnSpec{{Source: "region", Label: "Geo"}, {Source: "latency"}}}
	if got := SelectStatType(view); got != "latency by Geo" {
		t.Fatalf("got %q", got)
	}
	view = SelectView{Columns: []ColumnSpec{{Source: "region"}, {Source: "latency", Label: "Custom"}}}
	if got := SelectStatType(view); got != "Custom" {
		t.Fatalf("got %q", got)
	}
	view = SelectView{
		Columns:   []ColumnSpec{{Source: "region"}, {Source: "latency"}},
		TypeLabel: "Latency by Region",
	}
	if got := SelectStatType(view); got != "Latency by Region" {
		t.Fatalf("got %q", got)
	}
}

func TestValidateMultiSelectStatViews(t *testing.T) {
	if err := ValidateMultiSelectStatViews([]SelectView{{Columns: []ColumnSpec{{Source: "a"}, {Source: "b"}}}}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMultiSelectStatViews([]SelectView{{Columns: []ColumnSpec{{Source: "a"}, {Source: "b"}, {Source: "c"}}}}); err == nil {
		t.Fatal("want error for 3-column view")
	}
}

func TestMultiSelectSharedDim(t *testing.T) {
	shared := []SelectView{
		{Columns: []ColumnSpec{{Source: "region"}, {Source: "tax"}}},
		{Columns: []ColumnSpec{{Source: "region"}, {Source: "amount"}}},
	}
	if !MultiSelectSharedDim(shared) {
		t.Fatal("expected shared dim")
	}
	mixed := []SelectView{
		{Columns: []ColumnSpec{{Source: "region"}, {Source: "tax"}}},
		{Columns: []ColumnSpec{{Source: "product"}, {Source: "sales"}}},
	}
	if MultiSelectSharedDim(mixed) {
		t.Fatal("expected different dims")
	}
}

func TestMultiSelectStatAxesUsesDimLabel(t *testing.T) {
	views := []SelectView{{
		Columns: []ColumnSpec{{Source: "region", Label: "Region", AxisKey: "x"}, {Source: "tax", AxisKey: "y"}},
	}}
	axes := MultiSelectStatAxes(views)
	if len(axes) != 1 || axes[0].Key != "x" || axes[0].Label != "Region" {
		t.Fatalf("got %+v, want x/Region", axes)
	}
}

func TestMultiSelectStatAxesFallsBackToSource(t *testing.T) {
	views := []SelectView{{
		Columns: []ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "tax", AxisKey: "y"}},
	}}
	axes := MultiSelectStatAxes(views)
	if axes[0].Label != "region" {
		t.Fatalf("got label %q", axes[0].Label)
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

func TestParseSelectViewFlagQuotedParenInColumn(t *testing.T) {
	view, err := ParseSelectViewFlag(`"region (EU)",latency (Latency by Region)`)
	if err != nil {
		t.Fatal(err)
	}
	if view.TypeLabel != "Latency by Region" {
		t.Fatalf("TypeLabel = %q", view.TypeLabel)
	}
	if len(view.Columns) != 2 || view.Columns[0].Source != "region (EU)" {
		t.Fatalf("unexpected columns: %+v", view.Columns)
	}
}

func TestParseSelectViewFlagInvalidAxisKeyTreatedAsColumnName(t *testing.T) {
	view, err := ParseSelectViewFlag("w:region,latency")
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Columns) != 2 || view.Columns[0].Source != "w:region" || view.Columns[0].AxisKey != "x" {
		t.Fatalf("unexpected specs: %+v", view.Columns)
	}
}

func TestParseSelectViewFlagExplicitMissingY(t *testing.T) {
	if _, err := ParseSelectViewFlag("x:region,z:latency"); err == nil {
		t.Fatal("want missing y error")
	}
}

func TestMultiSelectSharedDimEmptyViews(t *testing.T) {
	if MultiSelectSharedDim(nil) {
		t.Fatal("expected false for empty views")
	}
}

func TestAppendMultiSelectStatPointNonMergeSkipsFailedRead(t *testing.T) {
	views := []SelectView{
		{Columns: []ColumnSpec{{Source: "region", AxisKey: "x"}, {Source: "tax", AxisKey: "y"}}},
		{Columns: []ColumnSpec{{Source: "product", AxisKey: "x"}, {Source: "sales", AxisKey: "y"}}},
	}
	var results []shared.DataPoint
	AppendMultiSelectStatPoint(&results, views, "", false, func(view SelectView) (MultiSelectRowStat, bool) {
		if view.Columns[0].Source == "product" {
			return MultiSelectRowStat{}, false
		}
		return MultiSelectRowStat{DimVal: "Asia", Value: 12}, true
	})
	if len(results) != 1 || results[0].XAxis != "Asia" {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestModeIsGrouped(t *testing.T) {
	if !ModeGrouped.IsGrouped() {
		t.Fatal("ModeGrouped should report grouped")
	}
	for _, m := range []Mode{ModeAuto, ModeValue, ModeMixed, ModeMultiStat} {
		if m.IsGrouped() {
			t.Fatalf("mode %v should not be grouped", m)
		}
	}
}
