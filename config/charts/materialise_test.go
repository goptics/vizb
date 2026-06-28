package charts_test

import (
	"testing"

	"github.com/goptics/vizb/config/charts"
	barchart "github.com/goptics/vizb/config/charts/bar"
	_ "github.com/goptics/vizb/config/charts/heatmap"
	_ "github.com/goptics/vizb/config/charts/line"
	piechart "github.com/goptics/vizb/config/charts/pie"
	_ "github.com/goptics/vizb/config/charts/radar"
	scatterchart "github.com/goptics/vizb/config/charts/scatter"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// MaterialiseSuite covers the generic charts.Materialise: defaults < seed <
// override precedence and per-chart field handling.
type MaterialiseSuite struct {
	suite.Suite
}

func sortSeed(order string) map[string]any {
	return map[string]any{"enabled": true, "order": order}
}

func (s *MaterialiseSuite) materialise(chartType string, seed map[string]any, override charts.ChartConfig) charts.ChartConfig {
	cfg, err := charts.Materialise(chartType, seed, override)
	s.Require().NoError(err)
	return cfg
}

func (s *MaterialiseSuite) TestScaleDefault() {
	// No seed, no override → scale falls back to the descriptor default.
	got := s.materialise("bar", nil, nil).(*barchart.Config)
	s.Equal("linear", got.Scale)
	s.Nil(got.ShowLabels)
	s.Nil(got.Sort)
}

func (s *MaterialiseSuite) TestSeedThenOverride() {
	seed := map[string]any{"swap": "xyn", "scale": "linear", "sort": sortSeed("asc")}
	override := &barchart.Config{Swap: "yxn", Scale: "log"}
	got := s.materialise("bar", seed, override).(*barchart.Config)

	s.Equal("yxn", got.Swap)  // override wins
	s.Equal("log", got.Scale) // override wins
	s.Require().NotNil(got.Sort)
	s.Equal("asc", got.Sort.Order) // seed survives where override is empty
}

func (s *MaterialiseSuite) TestOverrideCanSetFalse() {
	tr := true
	fa := false
	seed := map[string]any{"showLabels": true}
	got := s.materialise("bar", seed, &barchart.Config{ShowLabels: &fa}).(*barchart.Config)
	s.Require().NotNil(got.ShowLabels)
	s.False(*got.ShowLabels) // override false beats seed true

	got = s.materialise("bar", map[string]any{"threeD": true}, &barchart.Config{ThreeDRotate: &tr}).(*barchart.Config)
	s.Require().NotNil(got.ThreeD)
	s.True(*got.ThreeD)
	s.Require().NotNil(got.ThreeDRotate)
	s.True(*got.ThreeDRotate)
}

func (s *MaterialiseSuite) TestScatterSymbolAndVisualMap() {
	size := 11.0
	seed := map[string]any{"symbol": "pin", "symbolSize": size, "visualMap": true}
	got := s.materialise("scatter", seed, nil).(*scatterchart.Config)
	s.Equal("pin", got.Symbol)
	s.Require().NotNil(got.SymbolSize)
	s.Equal(11.0, *got.SymbolSize)
	s.Require().NotNil(got.VisualMap)
	s.True(*got.VisualMap)
}

func (s *MaterialiseSuite) TestStatSeed() {
	seed := map[string]any{"stat": shared.MaterialiseStatFlags([]string{"mean"})}
	got := s.materialise("scatter", seed, nil).(*scatterchart.Config)
	s.Require().NotNil(got.Stat)
	s.True(got.Stat.Enabled)
	s.Equal([]string{"mean"}, got.Stat.Math)
}

func (s *MaterialiseSuite) TestPieDropsInapplicableSeed() {
	// Pie has no Scale field; a "scale" key in the seed is dropped by Decode.
	got := s.materialise("pie", map[string]any{"scale": "log", "sort": sortSeed("desc")}, nil).(*piechart.Config)
	s.Require().NotNil(got.Sort)
	s.Equal("desc", got.Sort.Order)
	s.Equal("pie", got.ChartType())
}

func (s *MaterialiseSuite) TestUnknownChartType() {
	_, err := charts.Materialise("graph", nil, nil)
	s.Error(err)
}

func TestMaterialiseSuite(t *testing.T) {
	suite.Run(t, new(MaterialiseSuite))
}
