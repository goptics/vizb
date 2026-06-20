package scatter

import (
	"path/filepath"
	"testing"

	scatterchart "github.com/goptics/vizb/config/charts/scatter"
	"github.com/goptics/vizb/testutil"
	"github.com/stretchr/testify/suite"
)

// ScatterSuite verifies the scatter command exposes scale + 3d-rotate like bar and
// bakes a scatter-only selection into the new Settings shape.
type ScatterSuite struct {
	suite.Suite
	restoreOsExit func()
}

func (s *ScatterSuite) SetupTest() {
	s.restoreOsExit, _ = testutil.TrapOsExitPanic(s.T())
}

func (s *ScatterSuite) TearDownTest() {
	s.restoreOsExit()
}

func (s *ScatterSuite) TestCommandFlags() {
	cmd := NewCommand()
	s.Equal("scatter [target]", cmd.Use)
	s.NotNil(cmd.Flags().Lookup("scale"))
	s.NotNil(cmd.Flags().Lookup("3d-rotate"))
	s.NotNil(cmd.Flags().Lookup("swap"))
	s.NotNil(cmd.Flags().Lookup("sort"))
	s.NotNil(cmd.Flags().Lookup("show-labels"))
	s.NotNil(cmd.Flags().Lookup("axes"))
}

func (s *ScatterSuite) TestBakesScatterOnlySelection() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "y", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)
	s.Equal("scatter", ds.Settings[0].ChartType())

	scatterCfg, ok := ds.Settings[0].(*scatterchart.Config)
	s.Require().True(ok, "expected *scatterchart.Config, got %T", ds.Settings[0])
	s.Equal("linear", scatterCfg.Scale)
}

func (s *ScatterSuite) TestScatterCommandNewShape() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "n/x/y", "--swap", "yxn", "-l", "-s", "desc", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)
	s.Equal("scatter", ds.Settings[0].ChartType())

	scatterCfg, ok := ds.Settings[0].(*scatterchart.Config)
	s.Require().True(ok, "expected *scatterchart.Config, got %T", ds.Settings[0])
	s.Equal("yxn", scatterCfg.Swap)
	s.Require().NotNil(scatterCfg.ShowLabels)
	s.True(*scatterCfg.ShowLabels)
	s.Require().NotNil(scatterCfg.Sort)
	s.True(scatterCfg.Sort.Enabled)
	s.Equal("desc", scatterCfg.Sort.Order)
}

func (s *ScatterSuite) TestScatterCommandThreeDWithoutXYWarns() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "x", "--3d", input})

	stderr := testutil.CaptureStderr(func() {
		s.Require().NoError(cmd.Execute())
	})
	s.Contains(stderr, "Warning")
	s.Contains(stderr, "--3d requires both x and y")

	ds := testutil.ReadDataset(s.T(), out)
	scatterCfg, ok := ds.Settings[0].(*scatterchart.Config)
	s.Require().True(ok)
	s.Require().NotNil(scatterCfg.ThreeD)
	s.True(*scatterCfg.ThreeD)
}

func (s *ScatterSuite) TestScatterCommandBadSwapExits() {
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

func TestScatterSuite(t *testing.T) {
	suite.Run(t, new(ScatterSuite))
}
