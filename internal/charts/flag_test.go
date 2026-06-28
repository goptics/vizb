package charts_test

import (
	"testing"

	"github.com/goptics/vizb/internal/charts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateScaleValue(t *testing.T) {
	require.NoError(t, charts.ValidateScaleValue("linear"))
	require.NoError(t, charts.ValidateScaleValue("LOG"))
	assert.Error(t, charts.ValidateScaleValue("sqrt"))
}

func TestValidateSymbolSizeValue(t *testing.T) {
	require.NoError(t, charts.ValidateSymbolSizeValue("12"))
	assert.Error(t, charts.ValidateSymbolSizeValue("nope"))
	assert.Error(t, charts.ValidateSymbolSizeValue("0"))
}
