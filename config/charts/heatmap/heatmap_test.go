package heatmap

import (
	"reflect"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// HeatmapSuite covers heatmap config Materialise precedence and field invariants.
type HeatmapSuite struct {
	suite.Suite
}

func (s *HeatmapSuite) TestMaterialiseHeatmapPrecedence() {
	tr := true
	fa := false

	override := &Config{Swap: "yxn", ShowLabels: &tr}
	got := Materialise(Flags{Swap: "xyn", ShowLabels: false}, override)
	s.Equal("yxn", got.Swap)
	s.Require().NotNil(got.ShowLabels)
	s.True(*got.ShowLabels)

	got = Materialise(Flags{Swap: "xyn", ShowLabels: true}, nil)
	s.Equal("xyn", got.Swap)
	s.Require().NotNil(got.ShowLabels)
	s.True(*got.ShowLabels)

	got = Materialise(Flags{}, nil)
	s.Equal("", got.Swap)
	s.Equal("heatmap", got.Type)
	s.Nil(got.ShowLabels)
	s.Nil(got.Sort)

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

func (s *HeatmapSuite) TestHeatmapConfigNoScaleOrAutoRotate() {
	typ := reflect.TypeOf(Config{})
	_, hasScale := typ.FieldByName("Scale")
	_, hasAutoRotate := typ.FieldByName("AutoRotate")
	s.False(hasScale, "heatmap Config should not have a Scale field")
	s.False(hasAutoRotate, "heatmap Config should not have an AutoRotate field")
}

func TestHeatmapSuite(t *testing.T) {
	suite.Run(t, new(HeatmapSuite))
}
