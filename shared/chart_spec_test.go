package shared

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// ChartSpecSuite covers --chart spec parsing and the shared swap validator.
type ChartSpecSuite struct {
	suite.Suite
	xynAxes   []Axis
	allCharts []string
}

func (s *ChartSpecSuite) SetupTest() {
	s.xynAxes = []Axis{{Key: "x"}, {Key: "y"}, {Key: "name"}}
	s.allCharts = []string{"bar", "line", "pie", "heatmap", "radar"}
}

func (s *ChartSpecSuite) TestParseChartSpecs() {
	tests := []struct {
		name    string
		specs   []string
		charts  []string
		axes    []Axis
		wantErr bool
		check   func(got map[string]ChartSettings)
	}{
		{
			name:   "valid: swap and sort",
			specs:  []string{"bar:swap=yxn,sort=asc"},
			charts: s.allCharts,
			axes:   s.xynAxes,
			check: func(got map[string]ChartSettings) {
				bar, ok := got["bar"]
				s.Require().True(ok, "expected bar entry")
				s.Equal("yxn", bar.Swap)
				s.Require().NotNil(bar.Sort)
				s.True(bar.Sort.Enabled)
				s.Equal("asc", bar.Sort.Order)
			},
		},
		{
			name:   "valid: bare labels flag",
			specs:  []string{"bar:labels"},
			charts: s.allCharts,
			axes:   s.xynAxes,
			check: func(got map[string]ChartSettings) {
				s.Require().NotNil(got["bar"].ShowLabels)
				s.True(*got["bar"].ShowLabels)
			},
		},
		{
			name:   "valid: bare rotate flag",
			specs:  []string{"bar:rotate"},
			charts: s.allCharts,
			axes:   s.xynAxes,
			check: func(got map[string]ChartSettings) {
				s.Require().NotNil(got["bar"].AutoRotate)
				s.True(*got["bar"].AutoRotate)
			},
		},
		{
			name:   "valid: scale=log",
			specs:  []string{"bar:scale=log"},
			charts: s.allCharts,
			axes:   s.xynAxes,
			check: func(got map[string]ChartSettings) {
				s.Equal("log", got["bar"].Scale)
			},
		},
		{
			name:   "valid: multiple specs same type merged",
			specs:  []string{"bar:sort=asc", "bar:scale=log"},
			charts: s.allCharts,
			axes:   s.xynAxes,
			check: func(got map[string]ChartSettings) {
				s.Require().NotNil(got["bar"].Sort)
				s.Equal("asc", got["bar"].Sort.Order)
				s.Equal("log", got["bar"].Scale)
			},
		},
		{name: "invalid: pie:scale=log", specs: []string{"pie:scale=log"}, charts: s.allCharts, axes: s.xynAxes, wantErr: true},
		{name: "invalid: heatmap:rotate", specs: []string{"heatmap:rotate"}, charts: s.allCharts, axes: s.xynAxes, wantErr: true},
		{name: "invalid: bad sort value", specs: []string{"bar:sort=invalid"}, charts: s.allCharts, axes: s.xynAxes, wantErr: true},
		{name: "invalid: swap not a permutation", specs: []string{"bar:swap=abc"}, charts: s.allCharts, axes: s.xynAxes, wantErr: true},
		{name: "invalid: unknown key", specs: []string{"bar:unknown=val"}, charts: s.allCharts, axes: s.xynAxes, wantErr: true},
		{name: "invalid: type not in --charts", specs: []string{"xyz:sort=asc"}, charts: []string{"bar", "line"}, axes: s.xynAxes, wantErr: true},
		{
			name:   "valid: empty specs returns nil",
			specs:  []string{},
			charts: s.allCharts,
			axes:   s.xynAxes,
			check: func(got map[string]ChartSettings) {
				s.Nil(got)
			},
		},
		{name: "invalid: malformed (no colon)", specs: []string{"barswap=yxn"}, charts: s.allCharts, axes: s.xynAxes, wantErr: true},
		{
			name:   "valid: swap with no axes accepts non-empty string",
			specs:  []string{"bar:swap=anything"},
			charts: s.allCharts,
			axes:   []Axis{},
			check: func(got map[string]ChartSettings) {
				s.Equal("anything", got["bar"].Swap)
			},
		},
		{
			name:   "valid: sort normalised to lower",
			specs:  []string{"bar:sort=ASC"},
			charts: s.allCharts,
			axes:   s.xynAxes,
			check: func(got map[string]ChartSettings) {
				s.Require().NotNil(got["bar"].Sort)
				s.Equal("asc", got["bar"].Sort.Order)
			},
		},
		{
			name:   "valid: scale normalised to lower",
			specs:  []string{"bar:scale=LOG"},
			charts: s.allCharts,
			axes:   s.xynAxes,
			check: func(got map[string]ChartSettings) {
				s.Equal("log", got["bar"].Scale)
			},
		},
		{name: "invalid: empty rest after colon", specs: []string{"bar:"}, charts: s.allCharts, axes: s.xynAxes, wantErr: true},
		{
			name:   "valid: radar:sort=asc",
			specs:  []string{"radar:sort=asc"},
			charts: s.allCharts,
			axes:   s.xynAxes,
			check: func(got map[string]ChartSettings) {
				radar, ok := got["radar"]
				s.Require().True(ok, "expected radar entry")
				s.Require().NotNil(radar.Sort)
				s.True(radar.Sort.Enabled)
				s.Equal("asc", radar.Sort.Order)
			},
		},
		{name: "invalid: radar:scale=log", specs: []string{"radar:scale=log"}, charts: s.allCharts, axes: s.xynAxes, wantErr: true},
		{name: "invalid: radar:rotate", specs: []string{"radar:rotate"}, charts: s.allCharts, axes: s.xynAxes, wantErr: true},
		{
			name:   "valid: labels=true sets ShowLabels to true",
			specs:  []string{"bar:labels=true"},
			charts: s.allCharts,
			axes:   s.xynAxes,
			check: func(got map[string]ChartSettings) {
				s.Require().NotNil(got["bar"].ShowLabels)
				s.True(*got["bar"].ShowLabels)
			},
		},
		{
			name:   "valid: labels=false sets ShowLabels to false",
			specs:  []string{"bar:labels=false"},
			charts: s.allCharts,
			axes:   s.xynAxes,
			check: func(got map[string]ChartSettings) {
				s.Require().NotNil(got["bar"].ShowLabels)
				s.False(*got["bar"].ShowLabels)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := ParseChartSpecs(tt.specs, tt.charts, tt.axes)
			if tt.wantErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			if tt.check != nil {
				tt.check(got)
			}
		})
	}
}

func (s *ChartSpecSuite) TestValidateSwap() {
	s.Run("empty swap is always valid", func() {
		s.NoError(ValidateSwap("", s.xynAxes))
	})
	s.Run("valid permutation passes", func() {
		s.NoError(ValidateSwap("yxn", s.xynAxes))
	})
	s.Run("non-permutation fails", func() {
		s.Error(ValidateSwap("abc", s.xynAxes))
	})
	s.Run("no axes accepts any non-empty swap", func() {
		s.NoError(ValidateSwap("anything", nil))
	})
}

func TestChartSpecSuite(t *testing.T) {
	suite.Run(t, new(ChartSpecSuite))
}
