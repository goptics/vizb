package pie

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	piechart "github.com/goptics/vizb/config/charts/pie"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// PieSuite verifies the pie command omits --scale/--3d-rotate (non-linear chart)
// while keeping the shared chart flags, and bakes a pie-only selection into
// the new Settings shape.
type PieSuite struct {
	suite.Suite
	origOsExit func(int)
}

func (s *PieSuite) SetupTest() {
	s.origOsExit = shared.OsExit
}

func (s *PieSuite) TearDownTest() {
	shared.OsExit = s.origOsExit
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

func (s *PieSuite) TestPieCommand_NewShape() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "bench.txt")
	s.Require().NoError(os.WriteFile(input, []byte("BenchmarkExample-8 1000000 1234 ns/op"), 0644))
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "n/y", "--swap", "yn", "-l", "-s", "desc", input})
	s.Require().NoError(cmd.Execute())

	content, err := os.ReadFile(out)
	s.Require().NoError(err)
	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(content, &ds))
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
	// pie config has no Scale / AutoRotate fields — confirmed at compile time.
}

func TestPieSuite(t *testing.T) {
	suite.Run(t, new(PieSuite))
}
