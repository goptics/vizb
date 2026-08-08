package bar

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BarConfigSuite struct {
	suite.Suite
}

func (s *BarConfigSuite) TestBarConfigCreation() {
	cfg := New()
	require.NotNil(s.T(), cfg)

	barCfg, ok := cfg.(*Config)
	require.True(s.T(), ok)

	assert.Equal(s.T(), "bar", barCfg.Type)
	assert.Nil(s.T(), barCfg.BorderRadius)
}

func (s *BarConfigSuite) TestConfigChartType() {
	cfg := &Config{Type: "bar"}
	assert.Equal(s.T(), "bar", cfg.ChartType())
}

func TestBarConfigSuite(t *testing.T) {
	suite.Run(t, new(BarConfigSuite))
}