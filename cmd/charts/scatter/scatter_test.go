package scatter

import (
	"os"
	"path/filepath"
	"testing"

	scatterchart "github.com/goptics/vizb/config/charts/scatter"
	_ "github.com/goptics/vizb/pkg/parser/csv"
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

func (s *ScatterSuite) TestScatterCommandAutoValueTwoCols() {
	dir := s.T().TempDir()
	csv := filepath.Join(dir, "data.csv")
	s.Require().NoError(os.WriteFile(csv, []byte("price,latency\n100,12\n200,18\n"), 0644))
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "auto", csv})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Axes, 2)
	s.Equal("value", ds.Axes[0].Type)
	s.Equal("x", ds.Axes[0].Key)
	s.Equal("value", ds.Axes[1].Type)
	s.Equal("y", ds.Axes[1].Key)
}

func (s *ScatterSuite) TestScatterCommandAutoValueThreeCols() {
	dir := s.T().TempDir()
	csv := filepath.Join(dir, "data.csv")
	s.Require().NoError(os.WriteFile(csv, []byte("x,y,z\n1,2,3\n4,5,6\n"), 0644))
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "auto", csv})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Axes, 3)
	s.Equal("x", ds.Axes[0].Key)
	s.Equal("y", ds.Axes[1].Key)
	s.Equal("z", ds.Axes[2].Key)
	s.Len(ds.Data, 2)
}

func (s *ScatterSuite) TestScatterCommandWithThreeDVisualMapFlag() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "n/x/y", "--3d", "--3d-visualmap=false", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	scatterCfg, ok := ds.Settings[0].(*scatterchart.Config)
	s.Require().True(ok)
	s.Require().NotNil(scatterCfg.ThreeD)
	s.True(*scatterCfg.ThreeD)
	s.Require().NotNil(scatterCfg.ThreeDVisualMap)
	s.False(*scatterCfg.ThreeDVisualMap)
}

func (s *ScatterSuite) TestScatterCommandThreeDVisualMapWithoutThreeD() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "n/x/y", "--3d-visualmap", input})
	s.Require().NoError(cmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	scatterCfg, ok := ds.Settings[0].(*scatterchart.Config)
	s.Require().True(ok)
	s.Nil(scatterCfg.ThreeD)
	s.Require().NotNil(scatterCfg.ThreeDVisualMap)
	s.True(*scatterCfg.ThreeDVisualMap)
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
