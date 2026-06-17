package cmd

import (
	"os"
	"path/filepath"
	"testing"

	barchart "github.com/goptics/vizb/config/charts/bar"
	linechart "github.com/goptics/vizb/config/charts/line"
	piechart "github.com/goptics/vizb/config/charts/pie"
	"github.com/goptics/vizb/testutil"
	"github.com/stretchr/testify/suite"
)

// RootSuite covers the root command end-to-end via rootCmd.Execute.
type RootSuite struct {
	suite.Suite
	restoreOsExit func()
}

func (s *RootSuite) SetupTest() {
	ResetTestState()
	s.restoreOsExit, _ = testutil.TrapOsExitPanic(s.T())
}

func (s *RootSuite) TearDownTest() {
	s.restoreOsExit()
}

func (s *RootSuite) TestCommandFlags() {
	s.NotNil(rootCmd.Flags().Lookup("charts"))
	s.NotNil(rootCmd.Flags().Lookup("sort"))
	s.NotNil(rootCmd.Flags().Lookup("parser"))
	s.Equal(defaultChartTypes, rootOpts.Charts)
}

func (s *RootSuite) TestValidateRootOptionsWarnsAndDefaults() {
	rootOpts.MemUnit = "invalid"
	out := testutil.CaptureStderr(func() { validateRootOptions() })
	s.Equal("B", rootOpts.MemUnit)
	s.Contains(out, "Invalid memory unit")
}

func (s *RootSuite) TestRunBenchmarkValidFileInput() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "valid.txt",
		`BenchmarkTest-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`)
	out := filepath.Join(dir, "out.html")

	outStr := testutil.CaptureStdout(func() {
		rootCmd.SetArgs([]string{"-o", out, input})
		s.Require().NoError(rootCmd.Execute())
	})

	s.FileExists(out)
	s.Contains(outStr, "Generated")
}

func (s *RootSuite) TestRunBenchmarkNoArgsNoStdinExits() {
	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Close()
	defer func() { os.Stdin = oldStdin }()

	rootCmd.SetArgs(nil)
	s.Panics(func() { _ = rootCmd.Execute() })
	s.True(*exitCalled)
}

func (s *RootSuite) TestRunBenchmarkGlobalSortApplied() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "valid.txt",
		`BenchmarkTest-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`)
	out := filepath.Join(dir, "out.json")

	outStr := testutil.CaptureStdout(func() {
		rootCmd.SetArgs([]string{"-o", out, "-c", "bar,line", "-s", "desc", input})
		s.Require().NoError(rootCmd.Execute())
	})

	s.Contains(outStr, "Generated")

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 2)

	for i, c := range ds.Settings {
		switch c := c.(type) {
		case *barchart.Config:
			s.Require().NotNil(c.Sort, "settings[%d] (bar) sort should be set", i)
			s.True(c.Sort.Enabled)
			s.Equal("desc", c.Sort.Order)
		case *linechart.Config:
			s.Require().NotNil(c.Sort, "settings[%d] (line) sort should be set", i)
			s.True(c.Sort.Enabled)
			s.Equal("desc", c.Sort.Order)
		default:
			s.Failf("unexpected chart type", "settings[%d] type=%T", i, c)
		}
	}
}

func (s *RootSuite) TestBakesDefaultCharts() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "valid.txt",
		`BenchmarkTest-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`)
	out := filepath.Join(dir, "out.json")

	rootCmd.SetArgs([]string{"-o", out, input})
	s.Require().NoError(rootCmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 3)
	s.Equal("bar", ds.Settings[0].ChartType())
	s.Equal("line", ds.Settings[1].ChartType())
	s.Equal("pie", ds.Settings[2].ChartType())

	_, ok := ds.Settings[0].(*barchart.Config)
	s.Require().True(ok)
	_, ok = ds.Settings[1].(*linechart.Config)
	s.Require().True(ok)
	_, ok = ds.Settings[2].(*piechart.Config)
	s.Require().True(ok)
}

func TestRootSuite(t *testing.T) {
	suite.Run(t, new(RootSuite))
}
