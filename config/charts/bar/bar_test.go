package bar

import (
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// BarSuite covers bar config registry, decode, and Materialise precedence.
type BarSuite struct {
	suite.Suite
}

func (s *BarSuite) TestMaterialiseBarPrecedence() {
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

	partial := &Config{Swap: "n"}
	got = Materialise(Flags{Swap: "xyn", Scale: "log", ShowLabels: true, ThreeDRotate: true}, partial)
	s.Equal("n", got.Swap)
	s.Equal("log", got.Scale)
	s.Require().NotNil(got.ShowLabels)
	s.True(*got.ShowLabels)
	s.Require().NotNil(got.ThreeDRotate)
	s.True(*got.ThreeDRotate)

	got = Materialise(Flags{}, nil)
	s.Equal("", got.Swap)
	s.Equal("linear", got.Scale)
	s.Nil(got.ShowLabels)
	s.Nil(got.ThreeDRotate)
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

	got = Materialise(Flags{ThreeD: true}, nil)
	s.Require().NotNil(got.ThreeD)
	s.True(*got.ThreeD)
	s.Require().NotNil(got.ThreeDVisualMap)
	s.True(*got.ThreeDVisualMap)

	falseVal := false
	got = Materialise(Flags{ThreeD: true, ThreeDVisualMap: &falseVal}, nil)
	s.Require().NotNil(got.ThreeDVisualMap)
	s.False(*got.ThreeDVisualMap)

	trueVal := true
	got = Materialise(Flags{ThreeDVisualMap: &trueVal}, nil)
	s.Nil(got.ThreeD)
	s.Require().NotNil(got.ThreeDVisualMap)
	s.True(*got.ThreeDVisualMap)
}

func TestBarSuite(t *testing.T) {
	suite.Run(t, new(BarSuite))
}
