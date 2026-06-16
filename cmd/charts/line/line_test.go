package line

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	linechart "github.com/goptics/vizb/config/charts/line"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// LineSuite verifies the line command exposes scale + rotate like bar and
// bakes a line-only selection into the new Settings shape.
type LineSuite struct {
	suite.Suite
	origOsExit func(int)
}

func (s *LineSuite) SetupTest() {
	s.origOsExit = shared.OsExit
}

func (s *LineSuite) TearDownTest() {
	shared.OsExit = s.origOsExit
}

func (s *LineSuite) TestCommandFlags() {
	cmd := NewCommand()
	s.Equal("line [target]", cmd.Use)
	s.NotNil(cmd.Flags().Lookup("scale"))
	s.NotNil(cmd.Flags().Lookup("rotate"))
	s.NotNil(cmd.Flags().Lookup("swap"))
	s.NotNil(cmd.Flags().Lookup("sort"))
}

func (s *LineSuite) TestLineCommand_NewShape() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "bench.txt")
	s.Require().NoError(os.WriteFile(input, []byte("BenchmarkExample-8 1000000 1234 ns/op"), 0644))
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "y", "--swap", "yxn", "-l", "-s", "desc", input})
	s.Require().NoError(cmd.Execute())

	content, err := os.ReadFile(out)
	s.Require().NoError(err)
	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(content, &ds))
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

func TestLineSuite(t *testing.T) {
	suite.Run(t, new(LineSuite))
}
