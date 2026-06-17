package radar

import (
	"path/filepath"
	"testing"

	radarchart "github.com/goptics/vizb/config/charts/radar"
	"github.com/goptics/vizb/testutil"
	"github.com/stretchr/testify/suite"
)

// RadarSuite verifies the radar command omits --scale/--3d-rotate and bakes a
// radar-only selection into the new Settings shape.
type RadarSuite struct {
	suite.Suite
	restoreOsExit func()
}

func (s *RadarSuite) SetupTest() {
	s.restoreOsExit, _ = testutil.TrapOsExitPanic(s.T())
}

func (s *RadarSuite) TearDownTest() {
	s.restoreOsExit()
}

func (s *RadarSuite) TestCommandFlags() {
	cmd := NewCommand()
	s.Equal("radar [target]", cmd.Use)
	s.Nil(cmd.Flags().Lookup("scale"), "radar must not expose --scale")
	s.Nil(cmd.Flags().Lookup("3d-rotate"), "radar must not expose --3d-rotate")
	s.NotNil(cmd.Flags().Lookup("swap"))
	s.NotNil(cmd.Flags().Lookup("sort"))
	s.NotNil(cmd.Flags().Lookup("show-labels"))
}

func (s *RadarSuite) TestBakesRadarOnlySelection() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "y", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)
	s.Equal("radar", ds.Settings[0].ChartType())

	_, ok := ds.Settings[0].(*radarchart.Config)
	s.Require().True(ok, "expected *radarchart.Config, got %T", ds.Settings[0])
}

func (s *RadarSuite) TestRadarCommandNewShape() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "n/y", "--swap", "yn", "-l", "-s", "desc", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)
	s.Equal("radar", ds.Settings[0].ChartType())

	radarCfg, ok := ds.Settings[0].(*radarchart.Config)
	s.Require().True(ok, "expected *radarchart.Config, got %T", ds.Settings[0])
	s.Equal("yn", radarCfg.Swap)
	s.Require().NotNil(radarCfg.ShowLabels)
	s.True(*radarCfg.ShowLabels)
	s.Require().NotNil(radarCfg.Sort)
	s.True(radarCfg.Sort.Enabled)
	s.Equal("desc", radarCfg.Sort.Order)
}

func (s *RadarSuite) TestRadarCommandBadSwapExits() {
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

func TestRadarSuite(t *testing.T) {
	suite.Run(t, new(RadarSuite))
}
