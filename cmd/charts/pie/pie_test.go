package pie

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// PieSuite verifies the pie command omits --scale/--rotate (non-linear chart)
// while keeping the shared chart flags.
type PieSuite struct {
	suite.Suite
}

func (s *PieSuite) TestCommandFlags() {
	cmd := NewCommand()
	s.Equal("pie [target]", cmd.Use)
	s.Nil(cmd.Flags().Lookup("scale"), "pie must not expose --scale")
	s.Nil(cmd.Flags().Lookup("rotate"), "pie must not expose --rotate")
	s.NotNil(cmd.Flags().Lookup("swap"))
	s.NotNil(cmd.Flags().Lookup("sort"))
	s.NotNil(cmd.Flags().Lookup("show-labels"))
}

func TestPieSuite(t *testing.T) {
	suite.Run(t, new(PieSuite))
}
