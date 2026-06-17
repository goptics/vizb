package bar

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	barchart "github.com/goptics/vizb/config/charts/bar"
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
	// bar supports scale + 3d-rotate in addition to the shared chart flags.
	s.NotNil(cmd.Flags().Lookup("scale"))
	s.NotNil(cmd.Flags().Lookup("3d-rotate"))
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
	s.Require().Len(ds.Settings, 1)
	s.Equal("bar", ds.Settings[0].ChartType())

	barCfg, ok := ds.Settings[0].(*barchart.Config)
	s.Require().True(ok, "expected *barchart.Config, got %T", ds.Settings[0])
	s.Equal("linear", barCfg.Scale)
}

func (s *BarSuite) TestBarCommand_NewShape() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "bench.txt")
	s.Require().NoError(os.WriteFile(input, []byte("BenchmarkExample-8 1000000 1234 ns/op"), 0644))
	out := filepath.Join(dir, "out.json")

	cmd := NewCommand()
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "n/x/y", "--swap", "yxn", "-l", "-s", "desc", input})
	s.Require().NoError(cmd.Execute())

	content, err := os.ReadFile(out)
	s.Require().NoError(err)
	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(content, &ds))
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

func (s *BarSuite) TestBarCommand_BadSwapExits() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "bench.txt")
	s.Require().NoError(os.WriteFile(input, []byte("BenchmarkExample-8 1000000 1234 ns/op"), 0644))
	out := filepath.Join(dir, "out.json")

	exitCalled := false
	shared.OsExit = func(int) { exitCalled = true; panic("exit") }

	cmd := NewCommand()
	// -p y produces axes "y"; "xyz" is not a permutation of "y" → ValidateSwap errors.
	cmd.SetArgs([]string{"-o", out, "-P", "go", "-p", "y", "--swap", "xyz", input})
	s.Panics(func() { _ = cmd.Execute() })
	s.True(exitCalled, "expected shared.OsExit to be invoked for bad --swap")
}

func TestBarSuite(t *testing.T) {
	suite.Run(t, new(BarSuite))
}
