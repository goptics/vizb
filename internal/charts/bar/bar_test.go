package bar

import (
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

type BarConfigSuite struct {
	suite.Suite
}

func (s *BarConfigSuite) TestBorderRadiusDefaultsNil() {
	cfg := New().(*Config)
	s.Nil(cfg.BorderRadius)
}

func (s *BarConfigSuite) TestBorderRadiusField() {
	r := shared.BorderRadius{8}
	cfg := &Config{Type: Type, BorderRadius: &r}
	s.Equal("bar", cfg.ChartType())
	s.Equal(shared.BorderRadius{8}, *cfg.BorderRadius)

	r = shared.BorderRadius{8, 8, 0, 0}
	cfg.BorderRadius = &r
	s.Equal(shared.BorderRadius{8, 8, 0, 0}, *cfg.BorderRadius)
}

func TestBarConfigSuite(t *testing.T) {
	suite.Run(t, new(BarConfigSuite))
}
