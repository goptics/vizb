package scatter

import (
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// ScatterSuite covers scatter config Materialise precedence and field invariants.
type ScatterSuite struct {
	suite.Suite
}

func (s *ScatterSuite) TestMaterialiseScatterPrecedence() {
	tr := true
	fa := false

	override := &Config{Swap: "yxn", Scale: "log", ShowLabels: &tr, ThreeDRotate: &tr}
	got := Materialise(Flags{Swap: "xyn", Scale: "linear", ShowLabels: false, ThreeDRotate: false}, override)
	s.Equal("yxn", got.Swap)
	s.Equal("log", got.Scale)
	s.Require().NotNil(got.ShowLabels)
	s.True(*got.ShowLabels)
	s.Require().NotNil(got.ThreeDRotate)
	s.True(*got.ThreeDRotate)

	got = Materialise(Flags{Swap: "xyn", Scale: "linear", ShowLabels: true, ThreeDRotate: true}, nil)
	s.Equal("xyn", got.Swap)
	s.Equal("linear", got.Scale)
	s.Require().NotNil(got.ShowLabels)
	s.True(*got.ShowLabels)
	s.Require().NotNil(got.ThreeDRotate)
	s.True(*got.ThreeDRotate)

	got = Materialise(Flags{}, nil)
	s.Equal("linear", got.Scale)
	s.Nil(got.ShowLabels)
	s.Nil(got.ThreeDRotate)

	got = Materialise(Flags{ShowLabels: true}, &Config{ShowLabels: &fa})
	s.Require().NotNil(got.ShowLabels)
	s.False(*got.ShowLabels)

	overrideSort := &shared.Sort{Enabled: true, Order: "desc"}
	got = Materialise(Flags{}, &Config{Sort: overrideSort})
	s.Require().NotNil(got.Sort)
	s.True(got.Sort.Enabled)
	s.Equal("desc", got.Sort.Order)

	got = Materialise(Flags{Sort: "asc"}, nil)
	s.Require().NotNil(got.Sort)
	s.True(got.Sort.Enabled)
	s.Equal("asc", got.Sort.Order)
}

func (s *ScatterSuite) TestScatterConfigHasScaleAndThreeDRotate() {
	got := Materialise(Flags{Scale: "log", ThreeDRotate: true}, nil)
	s.Equal("log", got.Scale)
	s.Require().NotNil(got.ThreeDRotate)
	s.True(*got.ThreeDRotate)
}

func (s *ScatterSuite) TestMaterialiseThreeDAndVisualMap() {
	got := Materialise(Flags{ThreeD: true}, nil)
	s.Require().NotNil(got.ThreeD)
	s.True(*got.ThreeD)
	s.Nil(got.ThreeDVisualMap)

	falseVal := false
	got = Materialise(Flags{ThreeD: true, ThreeDVisualMap: &falseVal}, nil)
	s.Require().NotNil(got.ThreeDVisualMap)
	s.False(*got.ThreeDVisualMap)

	trueVal := true
	got = Materialise(Flags{ThreeDVisualMap: &trueVal}, nil)
	s.Nil(got.ThreeD)
	s.Require().NotNil(got.ThreeDVisualMap)
	s.True(*got.ThreeDVisualMap)

	overrideThreeD := false
	overrideVisualMap := true
	got = Materialise(Flags{ThreeD: true, ThreeDVisualMap: &falseVal}, &Config{
		ThreeD:          &overrideThreeD,
		ThreeDVisualMap: &overrideVisualMap,
		Stat:            &shared.StatConfig{Enabled: true, Math: []string{"mean"}},
	})
	s.Require().NotNil(got.ThreeD)
	s.False(*got.ThreeD)
	s.Require().NotNil(got.ThreeDVisualMap)
	s.True(*got.ThreeDVisualMap)
	s.Require().NotNil(got.Stat)
	s.True(got.Stat.Enabled)
	s.Equal([]string{"mean"}, got.Stat.Math)
}

func (s *ScatterSuite) TestMaterialiseVisualMap() {
	got := Materialise(Flags{}, nil)
	s.Nil(got.VisualMap)

	trueVal := true
	got = Materialise(Flags{VisualMap: &trueVal}, nil)
	s.Require().NotNil(got.VisualMap)
	s.True(*got.VisualMap)

	falseVal := false
	got = Materialise(Flags{VisualMap: &trueVal}, &Config{VisualMap: &falseVal})
	s.Require().NotNil(got.VisualMap)
	s.False(*got.VisualMap)
}

func (s *ScatterSuite) TestChartConfigInterface() {
	cfg := Config{Type: Type, Stat: &shared.StatConfig{Enabled: true, Math: []string{"median"}}}
	s.Equal("scatter", cfg.ChartType())
	s.True(cfg.StatEnabled())
	s.Equal([]string{"median"}, cfg.StatMath())

	empty := Config{}
	s.Equal("scatter", empty.ChartType())
	s.False(empty.StatEnabled())
	s.Nil(empty.StatMath())
}

func TestScatterSuite(t *testing.T) {
	suite.Run(t, new(ScatterSuite))
}
