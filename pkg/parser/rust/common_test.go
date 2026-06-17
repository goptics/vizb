package rust

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type CommonSuite struct {
	suite.Suite
}

func (s *CommonSuite) TestToNsMultiplierAllUnits() {
	s.Equal(1.0, toNsMultiplier("ns"))
	s.Equal(1000.0, toNsMultiplier("µs"))
	s.Equal(1000.0, toNsMultiplier("μs"))
	s.Equal(1_000_000.0, toNsMultiplier("ms"))
	s.Equal(1_000_000_000.0, toNsMultiplier("s"))
	s.Equal(1.0, toNsMultiplier("unknown"))
}

func TestCommonSuite(t *testing.T) {
	suite.Run(t, new(CommonSuite))
}
