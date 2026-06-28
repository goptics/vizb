package cli

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	config_charts "github.com/goptics/vizb/config/charts"
	barchart "github.com/goptics/vizb/config/charts/bar"
	linechart "github.com/goptics/vizb/config/charts/line"
	scatterchart "github.com/goptics/vizb/config/charts/scatter"
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

func (s *PipelineSuite) TestApplyJSONPathEnvelope() {
	envelope := s.writeFile("env.json", `{"data":[{"impl":"a","ops":120},{"impl":"b","ops":80}]}`)

	extracted := applyJSONPath(envelope, ".data")
	s.FileExists(extracted)

	cfg := parser.Config{GroupPattern: "x", Group: []string{"impl"}, JSONPath: ".data"}
	results := prepareData(extracted, "json", cfg)
	s.Len(results, 2)
	s.ElementsMatch([]string{"a", "b"}, []string{results[0].XAxis, results[1].XAxis})
}

func (s *PipelineSuite) TestPrepareDataWarnsJSONPathIgnoredForNonJSONParser() {
	benchFile := s.writeFile("valid.txt", `BenchmarkExample-8    1000000    1234 ns/op`)
	cfg := parser.Config{GroupPattern: "y", TimeUnit: "ns", MemUnit: "B", JSONPath: ".data"}

	errOut := testutil.CaptureStderr(func() {
		results := prepareData(benchFile, "go", cfg)
		s.NotEmpty(results)
	})
	s.Contains(errOut, "--json-path is only supported for the json parser")
}

func (s *PipelineSuite) TestPrepareDataWarnsSelectIgnoredForGoParser() {
	benchFile := s.writeFile("valid.txt", `BenchmarkExample-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`)
	cfg := parser.Config{
		GroupPattern: "y",
		TimeUnit:     "ns",
		MemUnit:      "B",
		Select:       []parser.ColumnSpec{{Source: "price"}},
	}

	errOut := testutil.CaptureStderr(func() {
		results := prepareData(benchFile, "go", cfg)
		s.NotEmpty(results)
	})
	s.Contains(errOut, "--select is only supported for csv/json parsers")
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
		{Name: "B1", Stats: []shared.Stat{{Type: "time", Value: shared.F64(1234)}}},
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

	meta := RunMeta{Parser: "go"}
	cfg := parser.Config{GroupPattern: "y", TimeUnit: "ns", MemUnit: "B"}
	barCfg := &barchart.Config{Type: "bar", Scale: "linear"}
	configs := []config_charts.ChartConfig{barCfg}

	s.Run("HTML output", func() {
		out := filepath.Join(s.T().TempDir(), "out.html")
		m := meta
		m.OutputFile = out
		RunLinear(&cobra.Command{}, []string{benchFile}, m, cfg, configs, false)

		s.FileExists(out)
		stat, err := os.Stat(out)
		s.Require().NoError(err)
		s.Greater(stat.Size(), int64(0))
	})

	s.Run("JSON output bakes the chart selection", func() {
		out := filepath.Join(s.T().TempDir(), "out.json")
		m := meta
		m.OutputFile = out
		RunLinear(&cobra.Command{}, []string{benchFile}, m, cfg, configs, false)

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
	meta := RunMeta{Parser: "go", OutputFile: out}

	RunSingleChart(&cobra.Command{}, []string{}, meta, parser.Config{GroupPattern: "y", TimeUnit: "ns", MemUnit: "B"}, nil)

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

	barCfg := &barchart.Config{Type: "bar", Scale: "log"}
	meta := RunMeta{Parser: "go", OutputFile: out}
	cfg := parser.Config{GroupPattern: "y", TimeUnit: "ns", MemUnit: "B"}

	oldStdout, oldStderr := os.Stdout, os.Stderr
	devnull, _ := os.Open(os.DevNull)
	os.Stdout, os.Stderr = devnull, devnull
	defer func() { os.Stdout, os.Stderr = oldStdout, oldStderr; devnull.Close() }()

	RunLinear(&cobra.Command{}, []string{input}, meta, cfg, []config_charts.ChartConfig{barCfg}, true)

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
			&barchart.Config{Type: "bar", Scale: "linear"},
			&linechart.Config{Type: "line", Scale: "log"},
		},
		Data: []shared.DataPoint{{Name: "P1", XAxis: "1", YAxis: "100"}},
	})
	out := filepath.Join(dir, "out.json")

	barCfg := &barchart.Config{Type: "bar", Scale: "log"}
	meta := RunMeta{Parser: "go", OutputFile: out}
	cfg := parser.Config{GroupPattern: "y", TimeUnit: "ns", MemUnit: "B"}

	oldStdout, oldStderr := os.Stdout, os.Stderr
	devnull, _ := os.Open(os.DevNull)
	os.Stdout, os.Stderr = devnull, devnull
	defer func() { os.Stdout, os.Stderr = oldStdout, oldStderr; devnull.Close() }()

	RunLinear(&cobra.Command{}, []string{input}, meta, cfg, []config_charts.ChartConfig{barCfg}, false)

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
	meta := RunMeta{Parser: "auto", OutputFile: out}
	cfg := parser.Config{GroupPattern: "x", TimeUnit: "ns", MemUnit: "B"}
	barCfg := &barchart.Config{Type: "bar", Scale: "linear"}

	oldStdout, oldStderr := os.Stdout, os.Stderr
	devnull, _ := os.Open(os.DevNull)
	os.Stdout, os.Stderr = devnull, devnull
	defer func() { os.Stdout, os.Stderr = oldStdout, oldStderr; devnull.Close() }()

	RunLinear(&cobra.Command{}, []string{csvFile}, meta, cfg, []config_charts.ChartConfig{barCfg}, false)

	s.FileExists(out)
}

func (s *PipelineSuite) TestFormatAggregationGroup() {
	tests := []struct {
		name string
		cfg  parser.Config
		want string
	}{
		{
			name: "two columns with name and x dimensions",
			cfg:  parser.Config{GroupPattern: "name,x", Group: []string{"name", "date"}},
			want: "by columns: name, date (name: name, x: date)",
		},
		{
			name: "three columns with name x y dimensions",
			cfg:  parser.Config{GroupPattern: "name,x,y", Group: []string{"region", "product", "month"}},
			want: "by columns: region, product, month (name: region, x: product, y: month)",
		},
		{
			name: "single column singular phrasing",
			cfg:  parser.Config{GroupPattern: "x", Group: []string{"name"}},
			want: "by column: name (x: name)",
		},
		{
			name: "curly axis labels override column names",
			cfg:  parser.Config{GroupPattern: "y{Region},x{Product}", Group: []string{"region", "product"}},
			want: "by columns: region, product (y: Region, x: Product)",
		},
		{
			name: "columns only when no group pattern",
			cfg:  parser.Config{Group: []string{"name", "date"}},
			want: "by columns: name, date",
		},
		{
			name: "regex axes without labels use key only",
			cfg:  parser.Config{GroupRegex: `(?P<name>.*)/(?P<x>.*)`, Group: []string{"benchmark"}},
			want: "by column: benchmark (name, x)",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, formatAggregationGroup(tt.cfg))
		})
	}
}

func (s *PipelineSuite) TestPrepareDataAggregatesCSV() {
	csvFile := s.writeFile("grouped.csv", "name,sells,date\nalpha,10,2024-01\nalpha,20,2024-01\nbeta,5,2025-02\n")
	cfg := parser.Config{GroupPattern: "name,x", Group: []string{"name", "date"}}

	r, w, err := os.Pipe()
	s.Require().NoError(err)
	oldStdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	results := prepareData(csvFile, "csv", cfg)
	w.Close()

	output, err := io.ReadAll(r)
	s.Require().NoError(err)
	s.Contains(string(output), "by columns: name, date (name: name, x: date)")

	s.Len(results, 2)
	s.Equal("alpha", results[0].Name)
	s.Equal("2024-01", results[0].XAxis)
	s.Equal(30.0, *results[0].Stats[0].Value)
}

func (s *PipelineSuite) TestPrepareDataAxesRejectsNonTabularParser() {
	cfg := parser.Config{Axes: []parser.ColumnSpec{{Source: "x"}, {Source: "y"}}}
	s.Panics(func() { prepareData("ignored.txt", "go", cfg) })
}

func (s *PipelineSuite) TestAssembleDatasetUsesAutoValueAxesFromData() {
	// Auto-group path: Stats empty + axes populated → value-type axes
	results := []shared.DataPoint{{XAxis: "100", YAxis: "12", ZAxis: "5", Stats: []shared.Stat{}}}
	cfg := parser.Config{AutoGroup: true}
	ds := assembleDataset(results, RunMeta{Name: "T"}, nil, cfg)

	s.Len(ds.Axes, 3)
	for _, ax := range ds.Axes {
		s.Equal("value", ax.Type)
	}
	s.Equal("x", ds.Axes[0].Key)
	s.Equal("y", ds.Axes[1].Key)
	s.Equal("z", ds.Axes[2].Key)
}

func (s *PipelineSuite) TestAssembleDatasetUsesCategoryAxesFromData() {
	// Auto-group path: Stats populated → category-type axes
	results := []shared.DataPoint{{XAxis: "US", YAxis: "Widget", Stats: []shared.Stat{{Type: "sells", Value: shared.F64(10)}}}}
	cfg := parser.Config{AutoGroup: true}
	ds := assembleDataset(results, RunMeta{Name: "T"}, nil, cfg)

	s.Len(ds.Axes, 2)
	for _, ax := range ds.Axes {
		s.Equal("", ax.Type)
	}
}

func (s *PipelineSuite) TestAssembleDatasetAutoEnablesVisualMapForValueMetric() {
	results := []shared.DataPoint{
		{XAxis: "0", YAxis: "0", ZAxis: "0", Metric: "4", Stats: []shared.Stat{}},
	}
	cfg := parser.Config{AutoGroup: true}
	configs := []config_charts.ChartConfig{
		&scatterchart.Config{Type: "scatter"},
	}
	ds := assembleDataset(results, RunMeta{Name: "Noise"}, configs, cfg)

	s.Require().Len(ds.Settings, 1)
	sc := ds.Settings[0].(*scatterchart.Config)
	s.Require().NotNil(sc.ThreeDVisualMap)
	s.True(*sc.ThreeDVisualMap)
	s.Require().NotNil(sc.ThreeD)
	s.True(*sc.ThreeD)
	s.Require().Len(ds.Axes, 4)
	s.Equal("metric", ds.Axes[3].Key)
	s.Equal("value", ds.Axes[3].Label)
}

func (s *PipelineSuite) TestAssembleDatasetAutoEnables3DForBarAndLine() {
	results := []shared.DataPoint{
		{XAxis: "1", YAxis: "2", ZAxis: "3", Stats: []shared.Stat{}},
	}
	cfg := parser.Config{AutoGroup: true}
	configs := []config_charts.ChartConfig{
		&barchart.Config{Type: "bar"},
		&linechart.Config{Type: "line"},
	}
	ds := assembleDataset(results, RunMeta{Name: "Grid"}, configs, cfg)

	s.Require().Len(ds.Settings, 2)
	bc := ds.Settings[0].(*barchart.Config)
	s.Require().NotNil(bc.ThreeD)
	s.True(*bc.ThreeD)
	s.Nil(bc.ThreeDVisualMap)

	lc := ds.Settings[1].(*linechart.Config)
	s.Require().NotNil(lc.ThreeD)
	s.True(*lc.ThreeD)
	s.Nil(lc.ThreeDVisualMap)
}

func (s *PipelineSuite) TestAssembleDatasetAppendMetricAxisFromConfig() {
	results := []shared.DataPoint{{XAxis: "0", YAxis: "0", ZAxis: "0", Stats: []shared.Stat{}}}
	cfg := parser.Config{AutoGroup: true, MetricColumn: "noise"}
	ds := assembleDataset(results, RunMeta{Name: "T"}, nil, cfg)

	s.Require().Len(ds.Axes, 4)
	s.Equal("metric", ds.Axes[3].Key)
	s.Equal("noise", ds.Axes[3].Label)
}

func (s *PipelineSuite) TestAssembleDatasetCategoryAxesSkipAuto3D() {
	results := []shared.DataPoint{
		{XAxis: "US", YAxis: "Widget", ZAxis: "Q1", Stats: []shared.Stat{{Type: "sells", Value: shared.F64(10)}}},
	}
	cfg := parser.Config{AutoGroup: true}
	configs := []config_charts.ChartConfig{
		&scatterchart.Config{Type: "scatter"},
	}
	ds := assembleDataset(results, RunMeta{Name: "Grouped"}, configs, cfg)

	sc := ds.Settings[0].(*scatterchart.Config)
	s.Nil(sc.ThreeD)
	s.Nil(sc.ThreeDVisualMap)
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
	meta := RunMeta{Parser: "go", OutputFile: out}
	cfg := parser.Config{GroupPattern: "y", TimeUnit: "ns", MemUnit: "B"}
	barCfg := &barchart.Config{Type: "bar", Scale: "linear"}

	RunLinear(&cobra.Command{}, nil, meta, cfg, []config_charts.ChartConfig{barCfg}, false)
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
		_, err := stdinFile.WriteString(line + "\n")
		s.Require().NoError(err)
	}
	_, err = stdinFile.Seek(0, 0)
	s.Require().NoError(err)
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
