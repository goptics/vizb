package radar

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// RadarSuite verifies the radar command omits --scale/--rotate.
type RadarSuite struct {
	suite.Suite
}

func (s *RadarSuite) TestCommandFlags() {
	cmd := NewCommand()
	s.Equal("radar [target]", cmd.Use)
	s.Nil(cmd.Flags().Lookup("scale"), "radar must not expose --scale")
	s.Nil(cmd.Flags().Lookup("rotate"), "radar must not expose --rotate")
	s.NotNil(cmd.Flags().Lookup("swap"))
	s.NotNil(cmd.Flags().Lookup("sort"))
}

func TestRadarSuite(t *testing.T) {
	suite.Run(t, new(RadarSuite))
}
