package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

// PipelineSuite covers the linear pipeline internals and RunLinear end-to-end.
// The "go" parser is registered transitively via pipeline.go's goparser import.
type PipelineSuite struct {
	suite.Suite
	origOsExit func(int)
}

func (s *PipelineSuite) SetupTest() {
	s.origOsExit = shared.OsExit
}

func (s *PipelineSuite) TearDownTest() {
	shared.OsExit = s.origOsExit
}

func (s *PipelineSuite) writeFile(name, content string) string {
	path := filepath.Join(s.T().TempDir(), name)
	s.Require().NoError(os.WriteFile(path, []byte(content), 0644))
	return path
}

func (s *PipelineSuite) TestPreprocessInputFile() {
	s.Run("Go bench JSON is converted to text", func() {
		jsonFile := s.writeFile("bench.json", `{"Action":"run","Test":"BenchmarkExample"}
{"Action":"pass","Test":"BenchmarkExample","Output":"1000000    1234 ns/op"}`)

		result := preprocessInputFile(jsonFile, "go")
		s.NotEqual(jsonFile, result)
		s.FileExists(result)
		os.Remove(result)
	})

	s.Run("Plain text file passes through", func() {
		textFile := s.writeFile("bench.txt", "BenchmarkExample-8 1000000 1234 ns/op")
		s.Equal(textFile, preprocessInputFile(textFile, "go"))
	})
}

func (s *PipelineSuite) TestPrepareData() {
	cfg := parser.Config{GroupPattern: "y", TimeUnit: "ns", MemUnit: "B"}

	s.Run("valid benchmark results", func() {
		benchFile := s.writeFile("valid.txt", `BenchmarkExample-8    1000000    1234 ns/op    1000 B/op    10 allocs/op
BenchmarkAnother-8    2000000    2345 ns/op    2000 B/op    20 allocs/op`)

		results := prepareData(benchFile, "go", cfg)
		s.NotEmpty(results)
	})

	s.Run("empty results call OsExit", func() {
		shared.OsExit = func(int) { panic("exit") }
		emptyFile := s.writeFile("empty.txt", "")
		s.Panics(func() { prepareData(emptyFile, "go", cfg) })
	})
}

func (s *PipelineSuite) TestWriteOutput() {
	dataSet := &shared.Dataset{Data: []shared.DataPoint{
		{Name: "B1", Stats: []shared.Stat{{Type: "time", Value: 1234}}},
	}}

	s.Run("HTML output is non-empty", func() {
		htmlFile := filepath.Join(s.T().TempDir(), "out.html")
		file, err := os.Create(htmlFile)
		s.Require().NoError(err)
		defer file.Close()

		writeOutput(file, dataSet, "html")
		stat, err := file.Stat()
		s.Require().NoError(err)
		s.Greater(stat.Size(), int64(0))
	})

	s.Run("JSON output round-trips", func() {
		jsonFile := filepath.Join(s.T().TempDir(), "out.json")
		file, err := os.Create(jsonFile)
		s.Require().NoError(err)
		defer file.Close()

		writeOutput(file, dataSet, "json")

		content, err := os.ReadFile(jsonFile)
		s.Require().NoError(err)
		var got shared.Dataset
		s.Require().NoError(json.Unmarshal(content, &got))
		s.Len(got.Data, 1)
	})

	s.Run("unknown format writes nothing", func() {
		invalidFile := filepath.Join(s.T().TempDir(), "out.x")
		file, err := os.Create(invalidFile)
		s.Require().NoError(err)
		defer file.Close()

		writeOutput(file, dataSet, "invalid")
		stat, err := file.Stat()
		s.Require().NoError(err)
		s.Equal(int64(0), stat.Size())
	})
}

func (s *PipelineSuite) TestCheckTargetFile() {
	s.Run("existing file does not exit", func() {
		valid := s.writeFile("valid.txt", "content")
		exitCalled := false
		shared.OsExit = func(int) { exitCalled = true; panic("exit") }
		s.NotPanics(func() { checkTargetFile(valid) })
		s.False(exitCalled)
	})

	s.Run("missing file exits", func() {
		shared.OsExit = func(int) { panic("exit") }
		s.Panics(func() { checkTargetFile(filepath.Join(s.T().TempDir(), "nope.txt")) })
	})
}

func (s *PipelineSuite) TestRunLinearGeneratesOutputFile() {
	benchFile := s.writeFile("input.txt", `BenchmarkExample-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`)

	common := CommonOptions{Parser: "go", GroupPattern: "y", TimeUnit: "ns", MemUnit: "B"}
	sel := []ChartSelection{{Type: "bar"}}

	s.Run("HTML output", func() {
		out := filepath.Join(s.T().TempDir(), "out.html")
		c := common
		c.OutputFile = out
		RunLinear(&cobra.Command{}, []string{benchFile}, c, LinearDefaults{Scale: "linear"}, sel, false)

		s.FileExists(out)
		stat, err := os.Stat(out)
		s.Require().NoError(err)
		s.Greater(stat.Size(), int64(0))
	})

	s.Run("JSON output bakes the chart selection", func() {
		out := filepath.Join(s.T().TempDir(), "out.json")
		c := common
		c.OutputFile = out
		RunLinear(&cobra.Command{}, []string{benchFile}, c, LinearDefaults{Scale: "linear"}, sel, false)

		content, err := os.ReadFile(out)
		s.Require().NoError(err)
		var ds shared.Dataset
		s.Require().NoError(json.Unmarshal(content, &ds))
		s.Equal([]string{"bar"}, ds.Settings.Charts)
	})
}

func (s *PipelineSuite) TestWriteStdinPipedInputs() {
	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	// Silence the progress bar output during the test.
	oldStdout, oldStderr := os.Stdout, os.Stderr
	devnull, _ := os.Open(os.DevNull)
	os.Stdout, os.Stderr = devnull, devnull
	defer func() { os.Stdout, os.Stderr = oldStdout, oldStderr; devnull.Close() }()

	stdinFile, err := os.CreateTemp("", "stdin_test")
	s.Require().NoError(err)
	defer os.Remove(stdinFile.Name())

	for _, line := range []string{
		`{"Action":"run","Test":"BenchmarkExample"}`,
		"BenchmarkAnotherTest-8 \t1000\t2000 ns/op",
	} {
		stdinFile.WriteString(line + "\n")
	}
	stdinFile.Seek(0, 0)
	os.Stdin = stdinFile

	out := filepath.Join(s.T().TempDir(), "out.txt")
	writeStdinPipedInputs(out)

	content, err := os.ReadFile(out)
	s.Require().NoError(err)
	s.Contains(string(content), "BenchmarkAnotherTest")
}

func TestPipelineSuite(t *testing.T) {
	suite.Run(t, new(PipelineSuite))
}
