package shared

import (
	"testing"
)

func TestParseChartSpecs(t *testing.T) {
	xynAxes := []Axis{
		{Key: "x"},
		{Key: "y"},
		{Key: "name"},
	}
	allCharts := []string{"bar", "line", "pie", "heatmap"}
	barLine := []string{"bar", "line"}

	tests := []struct {
		name    string
		specs   []string
		charts  []string
		axes    []Axis
		wantErr bool
		check   func(t *testing.T, got map[string]ChartSettings)
	}{
		{
			name:   "valid: swap and sort",
			specs:  []string{"bar:swap=yxn,sort=asc"},
			charts: allCharts,
			axes:   xynAxes,
			check: func(t *testing.T, got map[string]ChartSettings) {
				s, ok := got["bar"]
				if !ok {
					t.Fatal("expected bar entry")
				}
				if s.Swap != "yxn" {
					t.Errorf("Swap: got %q, want %q", s.Swap, "yxn")
				}
				if s.Sort == nil || !s.Sort.Enabled || s.Sort.Order != "asc" {
					t.Errorf("Sort: got %+v, want &Sort{Enabled:true, Order:asc}", s.Sort)
				}
			},
		},
		{
			name:   "valid: bare labels flag",
			specs:  []string{"bar:labels"},
			charts: allCharts,
			axes:   xynAxes,
			check: func(t *testing.T, got map[string]ChartSettings) {
				s := got["bar"]
				if s.ShowLabels == nil || !*s.ShowLabels {
					t.Errorf("ShowLabels: got %v, want &true", s.ShowLabels)
				}
			},
		},
		{
			name:   "valid: bare rotate flag",
			specs:  []string{"bar:rotate"},
			charts: allCharts,
			axes:   xynAxes,
			check: func(t *testing.T, got map[string]ChartSettings) {
				s := got["bar"]
				if s.AutoRotate == nil || !*s.AutoRotate {
					t.Errorf("AutoRotate: got %v, want &true", s.AutoRotate)
				}
			},
		},
		{
			name:   "valid: scale=log",
			specs:  []string{"bar:scale=log"},
			charts: allCharts,
			axes:   xynAxes,
			check: func(t *testing.T, got map[string]ChartSettings) {
				s := got["bar"]
				if s.Scale != "log" {
					t.Errorf("Scale: got %q, want %q", s.Scale, "log")
				}
			},
		},
		{
			name:   "valid: multiple specs same type merged",
			specs:  []string{"bar:sort=asc", "bar:scale=log"},
			charts: allCharts,
			axes:   xynAxes,
			check: func(t *testing.T, got map[string]ChartSettings) {
				s := got["bar"]
				if s.Sort == nil || s.Sort.Order != "asc" {
					t.Errorf("Sort: got %+v, want asc", s.Sort)
				}
				if s.Scale != "log" {
					t.Errorf("Scale: got %q, want log", s.Scale)
				}
			},
		},
		{
			name:    "invalid: pie:scale=log",
			specs:   []string{"pie:scale=log"},
			charts:  allCharts,
			axes:    xynAxes,
			wantErr: true,
		},
		{
			name:    "invalid: heatmap:rotate",
			specs:   []string{"heatmap:rotate"},
			charts:  allCharts,
			axes:    xynAxes,
			wantErr: true,
		},
		{
			name:    "invalid: bad sort value",
			specs:   []string{"bar:sort=invalid"},
			charts:  allCharts,
			axes:    xynAxes,
			wantErr: true,
		},
		{
			name:    "invalid: swap not a permutation",
			specs:   []string{"bar:swap=abc"},
			charts:  allCharts,
			axes:    xynAxes,
			wantErr: true,
		},
		{
			name:    "invalid: unknown key",
			specs:   []string{"bar:unknown=val"},
			charts:  allCharts,
			axes:    xynAxes,
			wantErr: true,
		},
		{
			name:    "invalid: type not in --charts",
			specs:   []string{"xyz:sort=asc"},
			charts:  barLine,
			axes:    xynAxes,
			wantErr: true,
		},
		{
			name:   "valid: empty specs returns nil",
			specs:  []string{},
			charts: allCharts,
			axes:   xynAxes,
			check: func(t *testing.T, got map[string]ChartSettings) {
				if got != nil {
					t.Errorf("expected nil map, got %v", got)
				}
			},
		},
		{
			name:    "invalid: malformed (no colon)",
			specs:   []string{"barswap=yxn"},
			charts:  allCharts,
			axes:    xynAxes,
			wantErr: true,
		},
		{
			name:   "valid: swap with no axes accepts non-empty string",
			specs:  []string{"bar:swap=anything"},
			charts: allCharts,
			axes:   []Axis{},
			check: func(t *testing.T, got map[string]ChartSettings) {
				if got["bar"].Swap != "anything" {
					t.Errorf("Swap: got %q, want %q", got["bar"].Swap, "anything")
				}
			},
		},
		{
			name:   "valid: sort normalised to lower",
			specs:  []string{"bar:sort=ASC"},
			charts: allCharts,
			axes:   xynAxes,
			check: func(t *testing.T, got map[string]ChartSettings) {
				s := got["bar"]
				if s.Sort == nil || s.Sort.Order != "asc" {
					t.Errorf("Sort.Order: got %+v, want asc", s.Sort)
				}
			},
		},
		{
			name:   "valid: scale normalised to lower",
			specs:  []string{"bar:scale=LOG"},
			charts: allCharts,
			axes:   xynAxes,
			check: func(t *testing.T, got map[string]ChartSettings) {
				if got["bar"].Scale != "log" {
					t.Errorf("Scale: got %q, want log", got["bar"].Scale)
				}
			},
		},
		{
			name:    "invalid: empty rest after colon",
			specs:   []string{"bar:"},
			charts:  allCharts,
			axes:    xynAxes,
			wantErr: true,
		},
		{
			name:   "valid: labels=true sets ShowLabels to true",
			specs:  []string{"bar:labels=true"},
			charts: allCharts,
			axes:   xynAxes,
			check: func(t *testing.T, got map[string]ChartSettings) {
				s := got["bar"]
				if s.ShowLabels == nil || !*s.ShowLabels {
					t.Errorf("ShowLabels: got %v, want &true", s.ShowLabels)
				}
			},
		},
		{
			name:   "valid: labels=false sets ShowLabels to false",
			specs:  []string{"bar:labels=false"},
			charts: allCharts,
			axes:   xynAxes,
			check: func(t *testing.T, got map[string]ChartSettings) {
				s := got["bar"]
				if s.ShowLabels == nil || *s.ShowLabels {
					t.Errorf("ShowLabels: got %v, want &false", s.ShowLabels)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseChartSpecs(tt.specs, tt.charts, tt.axes)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseChartSpecs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
