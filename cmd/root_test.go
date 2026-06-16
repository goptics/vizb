package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	barchart "github.com/goptics/vizb/config/charts/bar"
	linechart "github.com/goptics/vizb/config/charts/line"
	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

// osExit mirrors the historical test hook; tests panic instead of exiting.
var osExit = shared.OsExit
var originalOsExit = os.Exit

// TestMain replaces os.Exit so an exit-on-error path doesn't kill the test run.
func TestMain(m *testing.M) {
	osExit = func(code int) {
		panic(fmt.Sprintf("os.Exit(%d) was called", code))
	}

	code := m.Run()

	osExit = originalOsExit
	osExit(code)
}

// RootSuite covers the root command's runBenchmark + option validation. SetupTest
// resets the global rootOpts so cases don't leak state into one another (a payoff
// of removing the old shared.FlagState global).
type RootSuite struct {
	suite.Suite
	origOsExit func(int)
}

func (s *RootSuite) SetupTest() {
	rootOpts = rootOptions{}
	rootOpts.Parser = "auto"
	rootOpts.GroupPattern = "x"
	rootOpts.MemUnit = "B"
	rootOpts.TimeUnit = "ns"
	rootOpts.Charts = allChartTypes
	s.origOsExit = shared.OsExit
}

func (s *RootSuite) TearDownTest() {
	shared.OsExit = s.origOsExit
}

func (s *RootSuite) TestValidateRootOptionsWarnsAndDefaults() {
	rootOpts.MemUnit = "invalid"
	out := s.captureStderr(func() { validateRootOptions() })
	s.Equal("B", rootOpts.MemUnit)
	s.Contains(out, "Invalid memory unit")
}

func (s *RootSuite) TestRunBenchmarkValidFileInput() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "valid.txt")
	s.Require().NoError(os.WriteFile(input, []byte(`BenchmarkTest-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`), 0644))

	rootOpts.OutputFile = filepath.Join(dir, "out.html")

	out := s.captureStdout(func() {
		runBenchmark(&cobra.Command{}, []string{input})
	})

	s.FileExists(rootOpts.OutputFile)
	s.Contains(out, "Generated")
}

func (s *RootSuite) TestRunBenchmarkNoArgsNoStdinExits() {
	exitCalled := false
	shared.OsExit = func(int) { exitCalled = true; panic("exit") }

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	s.captureStdout(func() {
		s.Panics(func() { runBenchmark(cmd, []string{}) })
	})
	s.True(exitCalled)
}

func (s *RootSuite) captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestRunBenchmark_GlobalSortApplied runs the root command with the global
// --sort=desc flag, and asserts that every chart in the output has the sort
// applied. Verifies the global -s/--sort flag becomes a default applied to
// every chart (per spec section 4).
func (s *RootSuite) TestRunBenchmark_GlobalSortApplied() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "valid.txt")
	s.Require().NoError(os.WriteFile(input, []byte(`BenchmarkTest-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`), 0644))

	rootOpts.Charts = []string{"bar", "line"}
	rootOpts.Sort = "desc"
	rootOpts.OutputFile = filepath.Join(dir, "out.json")

	out := s.captureStdout(func() {
		runBenchmark(&cobra.Command{}, []string{input})
	})

	s.Contains(out, "Generated")

	content, err := os.ReadFile(rootOpts.OutputFile)
	s.Require().NoError(err)
	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(content, &ds))
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

func (s *RootSuite) captureStderr(fn func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = old }()

	fn()

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestRootSuite(t *testing.T) {
	suite.Run(t, new(RootSuite))
}
