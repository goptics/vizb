package cmd

import (
	"os"
	"path/filepath"
	"testing"

	config_charts "github.com/goptics/vizb/config/charts"
	barchart "github.com/goptics/vizb/config/charts/bar"
	linechart "github.com/goptics/vizb/config/charts/line"
	piechart "github.com/goptics/vizb/config/charts/pie"
	scatterchart "github.com/goptics/vizb/config/charts/scatter"
	"github.com/goptics/vizb/shared"
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
	s.Equal(defaultChartTypes, rootCharts)
}

func (s *RootSuite) TestValidateRootOptionsWarnsAndDefaults() {
	s.Require().NoError(rootCmd.Flags().Set("mem-unit", "invalid"))
	out := testutil.CaptureStderr(func() { validateRootOptions(rootCmd) })
	s.Equal("B", rootBag.String("mem-unit"))
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
	rootCharts = nil

	var outStr string
	errStr := testutil.CaptureStderr(func() {
		outStr = testutil.CaptureStdout(func() {
			rootCmd.SetArgs([]string{"-o", out, "-c", "bar", "-c", "line", "-s", "desc", input})
			s.Require().NoError(rootCmd.Execute())
		})
	})

	s.Contains(outStr, "Generated")
	s.Contains(errStr, "--sort is deprecated")

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

func (s *RootSuite) TestRunBenchmarkStdinPipe() {
	origStdin := os.Stdin
	defer func() { os.Stdin = origStdin }()

	stdinFile, err := os.CreateTemp("", "root_stdin")
	s.Require().NoError(err)
	defer os.Remove(stdinFile.Name())
	s.Require().NoError(os.WriteFile(stdinFile.Name(), []byte(
		"BenchmarkExample-8    1000000    1234 ns/op    1000 B/op    10 allocs/op\n",
	), 0644))
	f, err := os.Open(stdinFile.Name())
	s.Require().NoError(err)
	os.Stdin = f
	defer f.Close()

	dir := s.T().TempDir()
	out := filepath.Join(dir, "out.json")

	outStr := testutil.CaptureStdout(func() {
		rootCmd.SetArgs([]string{"-o", out})
		s.Require().NoError(rootCmd.Execute())
	})

	s.FileExists(out)
	s.Contains(outStr, "Generated")
	ds := testutil.ReadDataset(s.T(), out)
	s.NotEmpty(ds.Data)
}

func (s *RootSuite) TestRunBenchmarkDatasetPassthrough() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "baked.json")
	testutil.WriteJSON(s.T(), input, shared.Dataset{
		Name: "Baked",
		Settings: []config_charts.ChartConfig{
			&barchart.Config{Type: "bar", Scale: "log"},
			&linechart.Config{Type: "line", Scale: "log"},
		},
		Data: []shared.DataPoint{{Name: "P1", XAxis: "1", YAxis: "100"}},
	})
	out := filepath.Join(dir, "out.json")
	rootCharts = nil

	rootCmd.SetArgs([]string{"-o", out, "-c", "bar", input})
	s.Require().NoError(rootCmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 2)
	s.Equal("bar", ds.Settings[0].ChartType())
	s.Equal("line", ds.Settings[1].ChartType())
	lineCfg, ok := ds.Settings[1].(*linechart.Config)
	s.Require().True(ok)
	s.Equal("log", lineCfg.Scale)
}

func (s *RootSuite) TestRunBenchmarkAutoParser() {
	dir := s.T().TempDir()
	csv := filepath.Join(dir, "data.csv")
	s.Require().NoError(os.WriteFile(csv, []byte("name,sells\nalpha,10\nbeta,20\n"), 0644))
	out := filepath.Join(dir, "out.json")

	rootCmd.SetArgs([]string{"-o", out, "-P", "auto", csv})
	s.Require().NoError(rootCmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.NotEmpty(ds.Data)
}

func (s *RootSuite) TestRunBenchmarkChartOverride() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "valid.txt",
		`BenchmarkTest-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`)
	out := filepath.Join(dir, "out.json")
	rootCharts = nil

	rootCmd.SetArgs([]string{"-o", out, "-c", "bar", "--chart", "bar:swap=yxn", input})
	s.Require().NoError(rootCmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)
	barCfg, ok := ds.Settings[0].(*barchart.Config)
	s.Require().True(ok)
	s.Equal("yxn", barCfg.Swap)
}

func (s *RootSuite) TestRunBenchmarkScatterChart() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "valid.txt",
		`BenchmarkTest-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`)
	out := filepath.Join(dir, "out.json")
	rootCharts = nil

	rootCmd.SetArgs([]string{"-o", out, "-c", "scatter", input})
	s.Require().NoError(rootCmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)
	s.Equal("scatter", ds.Settings[0].ChartType())

	scatterCfg, ok := ds.Settings[0].(*scatterchart.Config)
	s.Require().True(ok)
	s.Equal("linear", scatterCfg.Scale)
}

func (s *RootSuite) TestRunBenchmarkScatterChartOverride() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "valid.txt",
		`BenchmarkTest-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`)
	out := filepath.Join(dir, "out.json")
	rootCharts = nil

	rootCmd.SetArgs([]string{"-o", out, "-c", "scatter", "--chart", "scatter:scale=log", input})
	s.Require().NoError(rootCmd.Execute())

	ds := testutil.ReadDataset(s.T(), out)
	s.Require().Len(ds.Settings, 1)

	scatterCfg, ok := ds.Settings[0].(*scatterchart.Config)
	s.Require().True(ok)
	s.Equal("log", scatterCfg.Scale)
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

func (s *RootSuite) TestExecuteRunsRootCommand() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "valid.txt", "")
	out := filepath.Join(dir, "out.json")

	rootCmd.SetArgs([]string{"-o", out, input})
	Execute()
	s.FileExists(out)
}

func (s *RootSuite) TestRunBenchmarkSortDeprecationWarning() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "valid.txt",
		`BenchmarkTest-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`)
	out := filepath.Join(dir, "out.json")

	var outStr string
	errStr := testutil.CaptureStderr(func() {
		outStr = testutil.CaptureStdout(func() {
			rootCmd.SetArgs([]string{"-o", out, "-s", "asc", input})
			s.Require().NoError(rootCmd.Execute())
		})
	})

	s.Contains(errStr, "--sort is deprecated")
	s.Contains(errStr, "--chart")
	s.Contains(outStr, "Generated")
	s.FileExists(out)
}

func (s *RootSuite) TestRunBenchmarkShowLabelsDeprecationWarning() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "valid.txt",
		`BenchmarkTest-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`)
	out := filepath.Join(dir, "out.json")

	var outStr string
	errStr := testutil.CaptureStderr(func() {
		outStr = testutil.CaptureStdout(func() {
			rootCmd.SetArgs([]string{"-o", out, "-l", input})
			s.Require().NoError(rootCmd.Execute())
		})
	})

	s.Contains(errStr, "--show-labels is deprecated")
	s.Contains(errStr, "--chart")
	s.Contains(outStr, "Generated")
	s.FileExists(out)
}

func (s *RootSuite) TestRunBenchmarkNoDeprecationWarningWhenFlagsUnset() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "valid.txt",
		`BenchmarkTest-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`)
	out := filepath.Join(dir, "out.json")

	var outStr string
	errStr := testutil.CaptureStderr(func() {
		outStr = testutil.CaptureStdout(func() {
			rootCmd.SetArgs([]string{"-o", out, input})
			s.Require().NoError(rootCmd.Execute())
		})
	})

	s.NotContains(errStr, "deprecated")
	s.Contains(outStr, "Generated")
	s.FileExists(out)
}

func (s *RootSuite) TestChartSubcommandSortNoDeprecationWarning() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "valid.txt",
		`BenchmarkTest-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`)
	out := filepath.Join(dir, "out.json")

	errStr := testutil.CaptureStderr(func() {
		rootCmd.SetArgs([]string{"bar", "-o", out, "-s", "asc", input})
		s.Require().NoError(rootCmd.Execute())
	})

	s.NotContains(errStr, "deprecated")
	s.FileExists(out)
}

func TestRootSuite(t *testing.T) {
	suite.Run(t, new(RootSuite))
}

// RootAutoGroupSuite covers the csv/json auto-grouping end-to-end via
// rootCmd.Execute, mirroring how a user runs `vizb data.csv`.
type RootAutoGroupSuite struct {
	suite.Suite
	restoreOsExit func()
}

func (s *RootAutoGroupSuite) SetupTest() {
	ResetTestState()
	s.restoreOsExit, _ = testutil.TrapOsExitPanic(s.T())
}

func (s *RootAutoGroupSuite) TearDownTest() {
	s.restoreOsExit()
}

func (s *RootAutoGroupSuite) writeCSV(dir, content string) string {
	p := filepath.Join(dir, "data.csv")
	s.Require().NoError(os.WriteFile(p, []byte(content), 0644))
	return p
}

func (s *RootAutoGroupSuite) writeJSON(dir, content string) string {
	p := filepath.Join(dir, "data.json")
	s.Require().NoError(os.WriteFile(p, []byte(content), 0644))
	return p
}

func (s *RootAutoGroupSuite) TestCSVAutoGroupsXAxisNoFlags() {
	dir := s.T().TempDir()
	input := s.writeCSV(dir, "region,sells\nWest,10\nEast,20\nNorth,30\n")
	out := filepath.Join(dir, "out.json")

	outStr := testutil.CaptureStdout(func() {
		rootCmd.SetArgs([]string{"-c", "bar", "-o", out, input})
		s.Require().NoError(rootCmd.Execute())
	})

	s.FileExists(out)
	s.Contains(outStr, "Auto-grouped by column")
	ds := testutil.ReadDataset(s.T(), out)
	// Auto-group picked the region column → every data point carries its
	// region value on xAxis. (Axis.Label stays empty without {label} slots.)
	s.Require().NotEmpty(ds.Axes)
	var hasX bool
	for _, ax := range ds.Axes {
		if ax.Key == "x" {
			hasX = true
		}
	}
	s.True(hasX)
	s.NotEmpty(ds.Data)
	for _, dp := range ds.Data {
		s.NotEmpty(dp.XAxis)
	}
}

func (s *RootAutoGroupSuite) TestCSVAutoGroup3DPicksSingleColumnAndWarns() {
	dir := s.T().TempDir()
	csv := "region,product,sells\nWest,A,10\nEast,B,20\nNorth,C,30\nSouth,D,40\nCentral,E,50\nWest,F,60\nEast,G,70\n"
	input := s.writeCSV(dir, csv)
	out := filepath.Join(dir, "out.json")

	var stdout string
	stderr := testutil.CaptureStderr(func() {
		stdout = testutil.CaptureStdout(func() {
			rootCmd.SetArgs([]string{"-c", "bar", "--chart", "bar:3d", "-o", out, input})
			s.Require().NoError(rootCmd.Execute())
		})
	})

	s.FileExists(out)
	s.Contains(stdout, "Auto-grouped by column")
	s.Contains(stderr, "Warning")
	s.Contains(stderr, "--3d requires both x and y")

	ds := testutil.ReadDataset(s.T(), out)
	s.NotEmpty(ds.Data)
	for _, dp := range ds.Data {
		s.NotEmpty(dp.XAxis)
		s.Empty(dp.YAxis)
	}
}

func (s *RootAutoGroupSuite) TestJSONAutoGroupsXAxisNoFlags() {
	dir := s.T().TempDir()
	input := s.writeJSON(dir, `[{"region":"West","sells":10},{"region":"East","sells":20}]`)
	out := filepath.Join(dir, "out.json")

	_ = testutil.CaptureStdout(func() {
		rootCmd.SetArgs([]string{"-c", "bar", "-o", out, input})
		s.Require().NoError(rootCmd.Execute())
	})

	s.FileExists(out)
	ds := testutil.ReadDataset(s.T(), out)
	s.Require().NotEmpty(ds.Axes)
	for _, dp := range ds.Data {
		s.NotEmpty(dp.XAxis)
	}
}

func (s *RootAutoGroupSuite) TestExplicitGroupOverridesAutoGroup() {
	dir := s.T().TempDir()
	input := s.writeCSV(dir, "region,product,sells\nWest,A,10\nEast,B,20\n")
	out := filepath.Join(dir, "out.json")

	outStr := testutil.CaptureStdout(func() {
		rootCmd.SetArgs([]string{"-c", "bar", "-g", "product", "-o", out, input})
		s.Require().NoError(rootCmd.Execute())
	})

	s.FileExists(out)
	s.NotContains(outStr, "Auto-grouped")
	ds := testutil.ReadDataset(s.T(), out)
	// Explicit -g product → xAxis values are the product column values.
	xAxisVals := make([]string, 0, len(ds.Data))
	for _, dp := range ds.Data {
		xAxisVals = append(xAxisVals, dp.XAxis)
	}
	s.ElementsMatch([]string{"A", "B"}, xAxisVals)
}

func (s *RootAutoGroupSuite) TestBenchmarkParserUnchangedNoGroup() {
	dir := s.T().TempDir()
	input := testutil.WriteBenchFile(s.T(), dir, "valid.txt",
		`BenchmarkTest-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`)
	out := filepath.Join(dir, "out.json")

	outStr := testutil.CaptureStdout(func() {
		rootCmd.SetArgs([]string{"-c", "bar", "-o", out, input})
		s.Require().NoError(rootCmd.Execute())
	})

	s.FileExists(out)
	s.NotContains(outStr, "Auto-grouped")
}

func TestRootAutoGroupSuite(t *testing.T) {
	suite.Run(t, new(RootAutoGroupSuite))
}
