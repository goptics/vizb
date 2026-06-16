package heatmap

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// HeatmapSuite verifies the heatmap command omits --scale/--rotate.
type HeatmapSuite struct {
	suite.Suite
}

func (s *HeatmapSuite) TestCommandFlags() {
	cmd := NewCommand()
	s.Equal("heatmap [target]", cmd.Use)
	s.Nil(cmd.Flags().Lookup("scale"), "heatmap must not expose --scale")
	s.Nil(cmd.Flags().Lookup("rotate"), "heatmap must not expose --rotate")
	s.NotNil(cmd.Flags().Lookup("swap"))
	s.NotNil(cmd.Flags().Lookup("sort"))
}

func TestHeatmapSuite(t *testing.T) {
	suite.Run(t, new(HeatmapSuite))
}
