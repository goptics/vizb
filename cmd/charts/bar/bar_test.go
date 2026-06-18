package bar

import (
	"path/filepath"
	"testing"

	barchart "github.com/goptics/vizb/config/charts/bar"
	"github.com/goptics/vizb/testutil"
	"github.com/stretchr/testify/suite"
)

// BarSuite verifies the bar command exposes its full flag set and bakes a
// bar-only selection end-to-end.
type BarSuite struct {
	suite.Suite
	restoreOsExit func()
}

func (s *BarSuite) SetupTest() {
	s.restoreOsExit, _ = testutil.TrapOsExitPanic(s.T())
}

func (s *BarSuite) TearDownTest() {
	s.restoreOsExit()
}

func (s *BarSuite) TestCommandFlags() {
	cmd := NewCommand()
	s.Equal("bar [target]", cmd.Use)
	s.NotNil(cmd.Flags().Lookup("scale"))
	s.NotNil(cmd.Flags().Lookup("swap"))
	s.NotNil(cmd.Flags().Lookup("sort"))
	s.NotNil(cmd.Flags().Lookup("show-labels"))
	s.NotNil(cmd.Flags().Lookup("3d"))
	s.NotNil(cmd.Flags().Lookup("3d-rotate"))
	s.NotNil(cmd.Flags().Lookup("3d-visualmap"))
}

func (s *BarSuite) TestBakesBarOnlySelection() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "y", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)
	s.Equal("bar", ds.Settings[0].ChartType())

	barCfg, ok := ds.Settings[0].(*barchart.Config)
	s.Require().True(ok, "expected *barchart.Config, got %T", ds.Settings[0])
	s.Equal("linear", barCfg.Scale)
}

func (s *BarSuite) TestBarCommandNewShape() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "n/x/y", "--swap", "yxn", "-l", "-s", "desc", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)
	s.Equal("bar", ds.Settings[0].ChartType())

	barCfg, ok := ds.Settings[0].(*barchart.Config)
	s.Require().True(ok, "expected *barchart.Config, got %T", ds.Settings[0])
	s.Equal("yxn", barCfg.Swap)
	s.Require().NotNil(barCfg.ShowLabels)
	s.True(*barCfg.ShowLabels)
	s.Require().NotNil(barCfg.Sort)
	s.True(barCfg.Sort.Enabled)
	s.Equal("desc", barCfg.Sort.Order)
}

func (s *BarSuite) TestBarCommandWithThreeDFlag() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "n/x/y", "--3d", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)

	barCfg, ok := ds.Settings[0].(*barchart.Config)
	s.Require().True(ok)
	s.Require().NotNil(barCfg.ThreeD)
	s.True(*barCfg.ThreeD)
	s.Nil(barCfg.ThreeDVisualMap)
}

func (s *BarSuite) TestBarCommandWithThreeDVisualMapFlag() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "n/x/y", "--3d", "--3d-visualmap=false", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)

	barCfg, ok := ds.Settings[0].(*barchart.Config)
	s.Require().True(ok)
	s.Require().NotNil(barCfg.ThreeD)
	s.True(*barCfg.ThreeD)
	s.Require().NotNil(barCfg.ThreeDVisualMap)
	s.False(*barCfg.ThreeDVisualMap)
}

func (s *BarSuite) TestBarCommandThreeDWithoutXYWarns() {
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
	barCfg, ok := ds.Settings[0].(*barchart.Config)
	s.Require().True(ok)
	s.Require().NotNil(barCfg.ThreeD)
	s.True(*barCfg.ThreeD)
}

func (s *BarSuite) TestBarCommandThreeDWithXYZNoWarn() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "n/x/y/z", "--3d", input})

	stderr := testutil.CaptureStderr(func() {
		s.Require().NoError(cmd.Execute())
	})
	s.NotContains(stderr, "--3d requires both x and y")

	ds := testutil.ReadDataset(s.T(), out)
	barCfg, ok := ds.Settings[0].(*barchart.Config)
	s.Require().True(ok)
	s.Require().NotNil(barCfg.ThreeD)
	s.True(*barCfg.ThreeD)
}

func (s *BarSuite) TestBarCommandThreeDVisualMapWithoutThreeD() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "n/x/y", "--3d-visualmap", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)

	barCfg, ok := ds.Settings[0].(*barchart.Config)
	s.Require().True(ok)
	s.Nil(barCfg.ThreeD)
	s.Require().NotNil(barCfg.ThreeDVisualMap)
	s.True(*barCfg.ThreeDVisualMap)
}

func (s *BarSuite) TestBarCommandBadSwapExits() {
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

func TestBarSuite(t *testing.T) {
	suite.Run(t, new(BarSuite))
}
