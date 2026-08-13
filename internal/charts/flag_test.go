package charts_test

import (
	"testing"

	"github.com/goptics/vizb/internal/charts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ChartFlagSuite struct {
	suite.Suite
}

func (s *ChartFlagSuite) TestValidateScaleValue() {
	t := s.T()
	require.NoError(t, charts.ValidateScaleValue("linear"))
	require.NoError(t, charts.ValidateScaleValue("LOG"))
	assert.Error(t, charts.ValidateScaleValue("sqrt"))
}

func (s *ChartFlagSuite) TestValidateSymbolSizeValue() {
	t := s.T()
	require.NoError(t, charts.ValidateSymbolSizeValue("12"))
	assert.Error(t, charts.ValidateSymbolSizeValue("nope"))
	assert.Error(t, charts.ValidateSymbolSizeValue("0"))
}

func (s *ChartFlagSuite) TestValidateBorderRadiusValue() {
	t := s.T()
	require.NoError(t, charts.ValidateBorderRadiusValue("0"))
	require.NoError(t, charts.ValidateBorderRadiusValue("8"))
	require.NoError(t, charts.ValidateBorderRadiusValue("100"))
	require.NoError(t, charts.ValidateBorderRadiusValue("8,8,0,0"))
	require.NoError(t, charts.ValidateBorderRadiusValue("1,2"))
	require.NoError(t, charts.ValidateBorderRadiusValue("1,2,3"))
	require.NoError(t, charts.ValidateBorderRadiusValue("1, 2, 3, 4"))
	assert.Error(t, charts.ValidateBorderRadiusValue("-5"))
	assert.Error(t, charts.ValidateBorderRadiusValue("abc"))
	assert.Error(t, charts.ValidateBorderRadiusValue("8.5"))
	assert.Error(t, charts.ValidateBorderRadiusValue(""))
	assert.Error(t, charts.ValidateBorderRadiusValue("8,8,0,0,1"))
	assert.Error(t, charts.ValidateBorderRadiusValue("8,-1,0,0"))
	assert.Error(t, charts.ValidateBorderRadiusValue("8,8.5,0,0"))
	assert.Error(t, charts.ValidateBorderRadiusValue(","))
}

func (s *ChartFlagSuite) TestEncodeBorderRadius() {
	t := s.T()
	assert.Equal(t, []int{8}, charts.EncodeBorderRadius("8"))
	assert.Equal(t, []int{0}, charts.EncodeBorderRadius("0"))
	assert.Equal(t, []int{8, 8, 0, 0}, charts.EncodeBorderRadius("8,8,0,0"))
	assert.Equal(t, []int{1, 2}, charts.EncodeBorderRadius("1,2"))
	assert.Equal(t, []int{1, 2, 3, 4}, charts.EncodeBorderRadius("1, 2, 3, 4"))
	// Non-string input is returned unchanged.
	assert.Equal(t, 42, charts.EncodeBorderRadius(42))
}

func TestChartFlagSuite(t *testing.T) {
	suite.Run(t, new(ChartFlagSuite))
}
