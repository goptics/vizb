package line

import (
	"path/filepath"
	"testing"

	linechart "github.com/goptics/vizb/config/charts/line"
	"github.com/goptics/vizb/testutil"
	"github.com/stretchr/testify/suite"
)

// LineSuite verifies the line command exposes scale + 3d-rotate like bar and
// bakes a line-only selection into the new Settings shape.
type LineSuite struct {
	suite.Suite
	restoreOsExit func()
}

func (s *LineSuite) SetupTest() {
	s.restoreOsExit, _ = testutil.TrapOsExitPanic(s.T())
}

func (s *LineSuite) TearDownTest() {
	s.restoreOsExit()
}

func (s *LineSuite) TestCommandFlags() {
	cmd := NewCommand()
	s.Equal("line [target]", cmd.Use)
	s.NotNil(cmd.Flags().Lookup("scale"))
	s.NotNil(cmd.Flags().Lookup("3d-rotate"))
	s.NotNil(cmd.Flags().Lookup("swap"))
	s.NotNil(cmd.Flags().Lookup("sort"))
	s.NotNil(cmd.Flags().Lookup("show-labels"))
	s.Nil(cmd.Flags().Lookup("axes"))
}

func (s *LineSuite) TestLineCommandRejectsUnknownAxesFlag() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "--axes", "x,y", input})
	err := cmd.Execute()
	s.Require().Error(err)
	s.Contains(err.Error(), "unknown flag")
}

func (s *LineSuite) TestBakesLineOnlySelection() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "y", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)
	s.Equal("line", ds.Settings[0].ChartType())

	lineCfg, ok := ds.Settings[0].(*linechart.Config)
	s.Require().True(ok, "expected *linechart.Config, got %T", ds.Settings[0])
	s.Equal("linear", lineCfg.Scale)
}

func (s *LineSuite) TestLineCommandNewShape() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "n/x/y", "--swap", "yxn", "-l", "-s", "desc", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)
	s.Equal("line", ds.Settings[0].ChartType())

	lineCfg, ok := ds.Settings[0].(*linechart.Config)
	s.Require().True(ok, "expected *linechart.Config, got %T", ds.Settings[0])
	s.Equal("yxn", lineCfg.Swap)
	s.Require().NotNil(lineCfg.ShowLabels)
	s.True(*lineCfg.ShowLabels)
	s.Require().NotNil(lineCfg.Sort)
	s.True(lineCfg.Sort.Enabled)
	s.Equal("desc", lineCfg.Sort.Order)
}

func (s *LineSuite) TestLineCommandThreeDWithoutXYWarns() {
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
	lineCfg, ok := ds.Settings[0].(*linechart.Config)
	s.Require().True(ok)
	s.Require().NotNil(lineCfg.ThreeD)
	s.True(*lineCfg.ThreeD)
}

func (s *LineSuite) TestLineCommandBadSwapExits() {
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

func TestLineSuite(t *testing.T) {
	suite.Run(t, new(LineSuite))
}
