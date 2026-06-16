package heatmap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	heatmapchart "github.com/goptics/vizb/config/charts/heatmap"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// HeatmapSuite verifies the heatmap command omits --scale/--rotate and bakes
// a heatmap-only selection into the new Settings shape.
type HeatmapSuite struct {
	suite.Suite
	origOsExit func(int)
}

func (s *HeatmapSuite) SetupTest() {
	s.origOsExit = shared.OsExit
}

func (s *HeatmapSuite) TearDownTest() {
	shared.OsExit = s.origOsExit
}

func (s *HeatmapSuite) TestCommandFlags() {
	cmd := NewCommand()
	s.Equal("heatmap [target]", cmd.Use)
	s.Nil(cmd.Flags().Lookup("scale"), "heatmap must not expose --scale")
	s.Nil(cmd.Flags().Lookup("rotate"), "heatmap must not expose --rotate")
	s.NotNil(cmd.Flags().Lookup("swap"))
	s.NotNil(cmd.Flags().Lookup("sort"))
}

func (s *HeatmapSuite) TestHeatmapCommand_NewShape() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "bench.txt")
	s.Require().NoError(os.WriteFile(input, []byte("BenchmarkExample-8 1000000 1234 ns/op"), 0644))
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "y", "--swap", "yn", "-l", "-s", "desc", input})
	s.Require().NoError(cmd.Execute())

	content, err := os.ReadFile(out)
	s.Require().NoError(err)
	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(content, &ds))
	s.Require().Len(ds.Settings, 1)
	s.Equal("heatmap", ds.Settings[0].ChartType())

	heatCfg, ok := ds.Settings[0].(*heatmapchart.Config)
	s.Require().True(ok, "expected *heatmapchart.Config, got %T", ds.Settings[0])
	s.Equal("yn", heatCfg.Swap)
	s.Require().NotNil(heatCfg.ShowLabels)
	s.True(*heatCfg.ShowLabels)
	s.Require().NotNil(heatCfg.Sort)
	s.True(heatCfg.Sort.Enabled)
	s.Equal("desc", heatCfg.Sort.Order)
}

func TestHeatmapSuite(t *testing.T) {
	suite.Run(t, new(HeatmapSuite))
}
