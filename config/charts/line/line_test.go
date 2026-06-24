package line

import (
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// LineSuite covers line config Materialise precedence and field invariants.
type LineSuite struct {
	suite.Suite
}

func (s *LineSuite) TestMaterialiseLinePrecedence() {
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

func (s *LineSuite) TestLineConfigHasScaleAndThreeDRotate() {
	got := Materialise(Flags{Scale: "log", ThreeDRotate: true}, nil)
	s.Equal("log", got.Scale)
	s.Require().NotNil(got.ThreeDRotate)
	s.True(*got.ThreeDRotate)
}

func (s *LineSuite) TestMaterialiseThreeDVisualMapOverride() {
	falseVal := false
	trueVal := true
	got := Materialise(Flags{ThreeD: true, ThreeDVisualMap: &falseVal}, &Config{
		ThreeDVisualMap: &trueVal,
	})
	s.Require().NotNil(got.ThreeDVisualMap)
	s.True(*got.ThreeDVisualMap)
}

func (s *LineSuite) TestSwapString() {
	s.Equal("yxn", Config{Swap: "yxn"}.SwapString())
}

func TestLineSuite(t *testing.T) {
	suite.Run(t, new(LineSuite))
}
