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

	override := &Config{Swap: "yxn", Scale: "log", ShowLabels: &tr, AutoRotate: &tr}
	got := Materialise(Flags{Swap: "xyn", Scale: "linear", ShowLabels: false, AutoRotate: false}, override)
	s.Equal("yxn", got.Swap)
	s.Equal("log", got.Scale)
	s.Require().NotNil(got.ShowLabels)
	s.True(*got.ShowLabels)
	s.Require().NotNil(got.AutoRotate)
	s.True(*got.AutoRotate)

	got = Materialise(Flags{Swap: "xyn", Scale: "linear", ShowLabels: true, AutoRotate: true}, nil)
	s.Equal("xyn", got.Swap)
	s.Equal("linear", got.Scale)
	s.Require().NotNil(got.ShowLabels)
	s.True(*got.ShowLabels)
	s.Require().NotNil(got.AutoRotate)
	s.True(*got.AutoRotate)

	got = Materialise(Flags{}, nil)
	s.Equal("linear", got.Scale)
	s.Nil(got.ShowLabels)
	s.Nil(got.AutoRotate)

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

func (s *LineSuite) TestLineConfigHasScaleAndAutoRotate() {
	got := Materialise(Flags{Scale: "log", AutoRotate: true}, nil)
	s.Equal("log", got.Scale)
	s.Require().NotNil(got.AutoRotate)
	s.True(*got.AutoRotate)
}

func TestLineSuite(t *testing.T) {
	suite.Run(t, new(LineSuite))
}
