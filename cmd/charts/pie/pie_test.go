package pie

import (
	"path/filepath"
	"testing"

	piechart "github.com/goptics/vizb/config/charts/pie"
	"github.com/goptics/vizb/testutil"
	"github.com/stretchr/testify/suite"
)

// PieSuite verifies the pie command omits --scale/--3d-rotate (non-linear chart)
// while keeping the shared chart flags, and bakes a pie-only selection into
// the new Settings shape.
type PieSuite struct {
	suite.Suite
	restoreOsExit func()
}

func (s *PieSuite) SetupTest() {
	s.restoreOsExit, _ = testutil.TrapOsExitPanic(s.T())
}

func (s *PieSuite) TearDownTest() {
	s.restoreOsExit()
}

func (s *PieSuite) TestCommandFlags() {
	cmd := NewCommand()
	s.Equal("pie [target]", cmd.Use)
	s.Nil(cmd.Flags().Lookup("scale"), "pie must not expose --scale")
	s.Nil(cmd.Flags().Lookup("3d-rotate"), "pie must not expose --3d-rotate")
	s.NotNil(cmd.Flags().Lookup("swap"))
	s.NotNil(cmd.Flags().Lookup("sort"))
	s.NotNil(cmd.Flags().Lookup("show-labels"))
}

func (s *PieSuite) TestBakesPieOnlySelection() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "y", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)
	s.Equal("pie", ds.Settings[0].ChartType())

	_, ok := ds.Settings[0].(*piechart.Config)
	s.Require().True(ok, "expected *piechart.Config, got %T", ds.Settings[0])
}

func (s *PieSuite) TestPieCommandNewShape() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "n/y", "--swap", "yn", "-l", "-s", "desc", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)
	s.Equal("pie", ds.Settings[0].ChartType())

	pieCfg, ok := ds.Settings[0].(*piechart.Config)
	s.Require().True(ok, "expected *piechart.Config, got %T", ds.Settings[0])
	s.Equal("yn", pieCfg.Swap)
	s.Require().NotNil(pieCfg.ShowLabels)
	s.True(*pieCfg.ShowLabels)
	s.Require().NotNil(pieCfg.Sort)
	s.True(pieCfg.Sort.Enabled)
	s.Equal("desc", pieCfg.Sort.Order)
}

func (s *PieSuite) TestPieCommandBadSwapExits() {
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

func TestPieSuite(t *testing.T) {
	suite.Run(t, new(PieSuite))
}
