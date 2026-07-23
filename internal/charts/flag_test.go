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

func (s *ChartFlagSuite) TestValidateLabelModeValue() {
	for _, value := range []string{"none", "VALUE", "percentage"} {
		s.NoError(charts.ValidateLabelModeValue(value))
	}
	s.Error(charts.ValidateLabelModeValue("percent"))
}

func (s *ChartFlagSuite) TestValidateSymbolSizeValue() {
	t := s.T()
	require.NoError(t, charts.ValidateSymbolSizeValue("12"))
	assert.Error(t, charts.ValidateSymbolSizeValue("nope"))
	assert.Error(t, charts.ValidateSymbolSizeValue("0"))
}

func TestChartFlagSuite(t *testing.T) {
	suite.Run(t, new(ChartFlagSuite))
}
