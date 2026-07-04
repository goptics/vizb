package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDimensionAxisKey(t *testing.T) {
	assert.Equal(t, "name", DimensionName.AxisKey())
	assert.Equal(t, "x", DimensionXAxis.AxisKey())
	assert.Equal(t, "y", DimensionYAxis.AxisKey())
	assert.Equal(t, "z", DimensionZAxis.AxisKey())
	assert.Equal(t, "name", Dimension("unknown").AxisKey())
}
