package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	config_charts "github.com/goptics/vizb/config/charts"
	barchart "github.com/goptics/vizb/config/charts/bar"
	linechart "github.com/goptics/vizb/config/charts/line"
	"github.com/goptics/vizb/pkg/parser"
	_ "github.com/goptics/vizb/pkg/parser/csv"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

// PipelineSuite covers the linear pipeline internals and RunLinear end-to-end.
// The "go" parser is registered transitively via pipeline.go's goparser import.
type PipelineSuite struct {
	suite.Suite
	restoreOsExit func()
}

func (s *PipelineSuite) SetupTest() {
	s.restoreOsExit, _ = testutil.TrapOsExitPanic(s.T())
}

func (s *PipelineSuite) TearDownTest() {
	s.restoreOsExit()
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
		restore, exitCalled := testutil.TrapOsExitPanic(s.T())
		defer restore()
		emptyFile := s.writeFile("empty.txt", "")
		s.Panics(func() { prepareData(emptyFile, "go", cfg) })
		s.True(*exitCalled)
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
		s.NotPanics(func() { checkTargetFile(valid) })
	})

	s.Run("missing file exits", func() {
		restore, exitCalled := testutil.TrapOsExitPanic(s.T())
		defer restore()
		s.Panics(func() { checkTargetFile(filepath.Join(s.T().TempDir(), "nope.txt")) })
		s.True(*exitCalled)
	})
}

func (s *PipelineSuite) TestRunLinearGeneratesOutputFile() {
	benchFile := s.writeFile("input.txt", `BenchmarkExample-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`)

	common := CommonOptions{Parser: "go", GroupPattern: "y", TimeUnit: "ns", MemUnit: "B"}
	barCfg := barchart.Materialise(barchart.Flags{Scale: "linear"}, nil)
	configs := []config_charts.ChartConfig{barCfg}

	s.Run("HTML output", func() {
		out := filepath.Join(s.T().TempDir(), "out.html")
		c := common
		c.OutputFile = out
		RunLinear(&cobra.Command{}, []string{benchFile}, c, configs, false)

		s.FileExists(out)
		stat, err := os.Stat(out)
		s.Require().NoError(err)
		s.Greater(stat.Size(), int64(0))
	})

	s.Run("JSON output bakes the chart selection", func() {
		out := filepath.Join(s.T().TempDir(), "out.json")
		c := common
		c.OutputFile = out
		RunLinear(&cobra.Command{}, []string{benchFile}, c, configs, false)

		content, err := os.ReadFile(out)
		s.Require().NoError(err)
		var ds shared.Dataset
		s.Require().NoError(json.Unmarshal(content, &ds))
		s.Require().Len(ds.Settings, 1)
		s.Equal("bar", ds.Settings[0].ChartType())
		typed, ok := ds.Settings[0].(*barchart.Config)
		s.Require().True(ok, "expected *barchart.Config, got %T", ds.Settings[0])
		s.Equal("linear", typed.Scale)
	})
}

func (s *PipelineSuite) TestRunSingleChartEmptyConfigs() {
	dir := s.T().TempDir()
	out := filepath.Join(dir, "out.json")
	common := CommonOptions{Parser: "go", GroupPattern: "y", TimeUnit: "ns", MemUnit: "B", OutputFile: out}

	RunSingleChart(&cobra.Command{}, []string{}, common, nil)

	_, err := os.Stat(out)
	s.True(os.IsNotExist(err), "empty configs should be a no-op (no file written)")
}

func (s *PipelineSuite) TestRunLinearDatasetPassthrough() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "baked.json")
	testutil.WriteJSON(s.T(), input, shared.Dataset{
		Name: "Baked",
		Settings: []config_charts.ChartConfig{
			&linechart.Config{Type: "line", Scale: "linear"},
		},
		Data: []shared.DataPoint{{Name: "P1", XAxis: "1", YAxis: "100"}},
	})
	out := filepath.Join(dir, "out.json")

	barCfg := barchart.Materialise(barchart.Flags{Scale: "log"}, nil)
	common := CommonOptions{Parser: "go", GroupPattern: "y", TimeUnit: "ns", MemUnit: "B", OutputFile: out}

	oldStdout, oldStderr := os.Stdout, os.Stderr
	devnull, _ := os.Open(os.DevNull)
	os.Stdout, os.Stderr = devnull, devnull
	defer func() { os.Stdout, os.Stderr = oldStdout, oldStderr; devnull.Close() }()

	RunLinear(&cobra.Command{}, []string{input}, common, []config_charts.ChartConfig{barCfg}, true)

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)
	s.Equal("bar", ds.Settings[0].ChartType())
	typed, ok := ds.Settings[0].(*barchart.Config)
	s.Require().True(ok)
	s.Equal("log", typed.Scale)
}

func (s *PipelineSuite) TestRunLinearPreservesDatasetOnRoot() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "baked.json")
	testutil.WriteJSON(s.T(), input, shared.Dataset{
		Name: "Baked",
		Settings: []config_charts.ChartConfig{
			barchart.Materialise(barchart.Flags{Scale: "linear"}, nil),
			&linechart.Config{Type: "line", Scale: "log"},
		},
		Data: []shared.DataPoint{{Name: "P1", XAxis: "1", YAxis: "100"}},
	})
	out := filepath.Join(dir, "out.json")

	barCfg := barchart.Materialise(barchart.Flags{Scale: "log"}, nil)
	common := CommonOptions{Parser: "go", GroupPattern: "y", TimeUnit: "ns", MemUnit: "B", OutputFile: out}

	oldStdout, oldStderr := os.Stdout, os.Stderr
	devnull, _ := os.Open(os.DevNull)
	os.Stdout, os.Stderr = devnull, devnull
	defer func() { os.Stdout, os.Stderr = oldStdout, oldStderr; devnull.Close() }()

	RunLinear(&cobra.Command{}, []string{input}, common, []config_charts.ChartConfig{barCfg}, false)

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 2)
	s.Equal("bar", ds.Settings[0].ChartType())
	s.Equal("line", ds.Settings[1].ChartType())
	lineCfg, ok := ds.Settings[1].(*linechart.Config)
	s.Require().True(ok)
	s.Equal("log", lineCfg.Scale)
}

func (s *PipelineSuite) TestRunLinearAutoParser() {
	csvFile := s.writeFile("data.csv", "name,value\na,10\nb,20\n")
	out := filepath.Join(s.T().TempDir(), "out.json")
	common := CommonOptions{
		Parser:       "auto",
		GroupPattern: "x",
		TimeUnit:     "ns",
		MemUnit:      "B",
		OutputFile:   out,
	}
	barCfg := barchart.Materialise(barchart.Flags{Scale: "linear"}, nil)

	oldStdout, oldStderr := os.Stdout, os.Stderr
	devnull, _ := os.Open(os.DevNull)
	os.Stdout, os.Stderr = devnull, devnull
	defer func() { os.Stdout, os.Stderr = oldStdout, oldStderr; devnull.Close() }()

	RunLinear(&cobra.Command{}, []string{csvFile}, common, []config_charts.ChartConfig{barCfg}, false)

	s.FileExists(out)
}

func (s *PipelineSuite) TestPrepareDataAggregatesCSV() {
	csvFile := s.writeFile("grouped.csv", "name,sells,date\nalpha,10,2024-01\nalpha,20,2024-01\nbeta,5,2025-02\n")
	cfg := parser.Config{GroupPattern: "name/x", Group: []string{"name", "date"}}

	oldStdout, oldStderr := os.Stdout, os.Stderr
	devnull, _ := os.Open(os.DevNull)
	os.Stdout, os.Stderr = devnull, devnull
	defer func() { os.Stdout, os.Stderr = oldStdout, oldStderr; devnull.Close() }()

	results := prepareData(csvFile, "csv", cfg)
	s.Len(results, 2)
	s.Equal("alpha", results[0].Name)
	s.Equal("2024-01", results[0].XAxis)
	s.Equal(30.0, results[0].Stats[0].Value)
}

func (s *PipelineSuite) TestPrepareDataUnknownParserExits() {
	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	s.Panics(func() {
		prepareData(s.writeFile("x.csv", "a,b\n1,2"), "nope", parser.Config{})
	})
	s.True(*exitCalled)
}

func (s *PipelineSuite) TestResolveInputStdin() {
	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	oldStdout, oldStderr := os.Stdout, os.Stderr
	devnull, _ := os.Open(os.DevNull)
	os.Stdout, os.Stderr = devnull, devnull
	defer func() { os.Stdout, os.Stderr = oldStdout, oldStderr; devnull.Close() }()

	stdinFile, err := os.CreateTemp("", "stdin_linear")
	s.Require().NoError(err)
	defer os.Remove(stdinFile.Name())

	s.Require().NoError(os.WriteFile(stdinFile.Name(), []byte(
		"BenchmarkExample-8    1000000    1234 ns/op    1000 B/op    10 allocs/op\n",
	), 0644))
	f, err := os.Open(stdinFile.Name())
	s.Require().NoError(err)
	os.Stdin = f
	defer f.Close()

	out := filepath.Join(s.T().TempDir(), "out.json")
	common := CommonOptions{Parser: "go", GroupPattern: "y", TimeUnit: "ns", MemUnit: "B", OutputFile: out}
	barCfg := barchart.Materialise(barchart.Flags{Scale: "linear"}, nil)

	RunLinear(&cobra.Command{}, nil, common, []config_charts.ChartConfig{barCfg}, false)
	s.FileExists(out)
}

func (s *PipelineSuite) TestResolveInputNoArgsShowsHelp() {
	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	// /dev/null is a character device → not treated as piped stdin.
	devnull, err := os.Open(os.DevNull)
	s.Require().NoError(err)
	os.Stdin = devnull
	defer devnull.Close()

	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	cmd := &cobra.Command{Use: "test"}
	s.Panics(func() { resolveInput(cmd, nil) })
	s.True(*exitCalled)
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
