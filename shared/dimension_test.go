package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DimensionSuite struct {
	suite.Suite
}

func (s *DimensionSuite) TestDimensionAxisKey() {
	t := s.T()
	assert.Equal(t, "name", DimensionName.AxisKey())
	assert.Equal(t, "x", DimensionXAxis.AxisKey())
	assert.Equal(t, "y", DimensionYAxis.AxisKey())
	assert.Equal(t, "z", DimensionZAxis.AxisKey())
	assert.Equal(t, "name", Dimension("unknown").AxisKey())
}

func TestDimensionSuite(t *testing.T) {
	suite.Run(t, new(DimensionSuite))
}
