package bar

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// BarSuite verifies the bar command exposes its full flag set and bakes a
// bar-only selection end-to-end.
type BarSuite struct {
	suite.Suite
	origOsExit func(int)
}

func (s *BarSuite) SetupTest() {
	s.origOsExit = shared.OsExit
}

func (s *BarSuite) TearDownTest() {
	shared.OsExit = s.origOsExit
}

func (s *BarSuite) TestCommandFlags() {
	cmd := NewCommand()
	s.Equal("bar [target]", cmd.Use)
	// bar supports scale + rotate (3D) in addition to the shared chart flags.
	s.NotNil(cmd.Flags().Lookup("scale"))
	s.NotNil(cmd.Flags().Lookup("rotate"))
	s.NotNil(cmd.Flags().Lookup("swap"))
	s.NotNil(cmd.Flags().Lookup("sort"))
	s.NotNil(cmd.Flags().Lookup("show-labels"))
}

func (s *BarSuite) TestBakesBarOnlySelection() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "bench.txt")
	s.Require().NoError(os.WriteFile(input, []byte("BenchmarkExample-8 1000000 1234 ns/op"), 0644))
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "y", input})
	s.Require().NoError(cmd.Execute())

	content, err := os.ReadFile(out)
	s.Require().NoError(err)
	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(content, &ds))
	s.Equal([]string{"bar"}, ds.Settings.Charts)
}

func TestBarSuite(t *testing.T) {
	suite.Run(t, new(BarSuite))
}
