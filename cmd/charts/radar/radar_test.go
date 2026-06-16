package radar

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	radarchart "github.com/goptics/vizb/config/charts/radar"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// RadarSuite verifies the radar command omits --scale/--rotate and bakes a
// radar-only selection into the new Settings shape.
type RadarSuite struct {
	suite.Suite
	origOsExit func(int)
}

func (s *RadarSuite) SetupTest() {
	s.origOsExit = shared.OsExit
}

func (s *RadarSuite) TearDownTest() {
	shared.OsExit = s.origOsExit
}

func (s *RadarSuite) TestCommandFlags() {
	cmd := NewCommand()
	s.Equal("radar [target]", cmd.Use)
	s.Nil(cmd.Flags().Lookup("scale"), "radar must not expose --scale")
	s.Nil(cmd.Flags().Lookup("rotate"), "radar must not expose --rotate")
	s.NotNil(cmd.Flags().Lookup("swap"))
	s.NotNil(cmd.Flags().Lookup("sort"))
}

func (s *RadarSuite) TestRadarCommand_NewShape() {
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

func TestRadarSuite(t *testing.T) {
	suite.Run(t, new(RadarSuite))
}
