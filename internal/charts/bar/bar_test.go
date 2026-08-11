package bar

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type BarConfigSuite struct {
	suite.Suite
}

func (s *BarConfigSuite) TestBorderRadiusDefaultsNil() {
	cfg := New().(*Config)
	s.Nil(cfg.BorderRadius)
}

func (s *BarConfigSuite) TestBorderRadiusRoundTrip() {
	r := 8
	cfg := &Config{Type: Type, BorderRadius: &r}
	s.Equal("bar", cfg.ChartType())
	s.Require().NotNil(cfg.BorderRadius)
	s.Equal(8, *cfg.BorderRadius)
}

func TestBarConfigSuite(t *testing.T) {
	suite.Run(t, new(BarConfigSuite))
}
