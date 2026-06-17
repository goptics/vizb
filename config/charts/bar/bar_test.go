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

	partial := &Config{Swap: "n"}
	got = Materialise(Flags{Swap: "xyn", Scale: "log", ShowLabels: true, AutoRotate: true}, partial)
	s.Equal("n", got.Swap)
	s.Equal("log", got.Scale)
	s.Require().NotNil(got.ShowLabels)
	s.True(*got.ShowLabels)
	s.Require().NotNil(got.AutoRotate)
	s.True(*got.AutoRotate)

	got = Materialise(Flags{}, nil)
	s.Equal("", got.Swap)
	s.Equal("linear", got.Scale)
	s.Nil(got.ShowLabels)
	s.Nil(got.AutoRotate)
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

func TestBarSuite(t *testing.T) {
	suite.Run(t, new(BarSuite))
}
