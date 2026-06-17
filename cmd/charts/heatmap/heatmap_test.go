package heatmap

import (
	"path/filepath"
	"testing"

	heatmapchart "github.com/goptics/vizb/config/charts/heatmap"
	"github.com/goptics/vizb/testutil"
	"github.com/stretchr/testify/suite"
)

// HeatmapSuite verifies the heatmap command omits --scale/--3d-rotate and bakes
// a heatmap-only selection into the new Settings shape.
type HeatmapSuite struct {
	suite.Suite
	restoreOsExit func()
}

func (s *HeatmapSuite) SetupTest() {
	s.restoreOsExit, _ = testutil.TrapOsExitPanic(s.T())
}

func (s *HeatmapSuite) TearDownTest() {
	s.restoreOsExit()
}

func (s *HeatmapSuite) TestCommandFlags() {
	cmd := NewCommand()
	s.Equal("heatmap [target]", cmd.Use)
	s.Nil(cmd.Flags().Lookup("scale"), "heatmap must not expose --scale")
	s.Nil(cmd.Flags().Lookup("3d-rotate"), "heatmap must not expose --3d-rotate")
	s.NotNil(cmd.Flags().Lookup("swap"))
	s.NotNil(cmd.Flags().Lookup("sort"))
	s.NotNil(cmd.Flags().Lookup("show-labels"))
}

func (s *HeatmapSuite) TestBakesHeatmapOnlySelection() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "y", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)
	s.Equal("heatmap", ds.Settings[0].ChartType())

	_, ok := ds.Settings[0].(*heatmapchart.Config)
	s.Require().True(ok, "expected *heatmapchart.Config, got %T", ds.Settings[0])
}

func (s *HeatmapSuite) TestHeatmapCommandNewShape() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "n/y", "--swap", "yn", "-l", "-s", "desc", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
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

func (s *HeatmapSuite) TestHeatmapCommandBadSwapExits() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "y", "--swap", "xyz", input})
	s.Panics(func() { _ = cmd.Execute() })
	s.True(*exitCalled, "expected shared.OsExit to be invoked for bad --swap")
}

func TestHeatmapSuite(t *testing.T) {
	suite.Run(t, new(HeatmapSuite))
}
