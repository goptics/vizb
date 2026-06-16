package line

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// LineSuite verifies the line command exposes scale + rotate like bar.
type LineSuite struct {
	suite.Suite
}

func (s *LineSuite) TestCommandFlags() {
	cmd := NewCommand()
	s.Equal("line [target]", cmd.Use)
	s.NotNil(cmd.Flags().Lookup("scale"))
	s.NotNil(cmd.Flags().Lookup("rotate"))
	s.NotNil(cmd.Flags().Lookup("swap"))
	s.NotNil(cmd.Flags().Lookup("sort"))
}

func TestLineSuite(t *testing.T) {
	suite.Run(t, new(LineSuite))
}
