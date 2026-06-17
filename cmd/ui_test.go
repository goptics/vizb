package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	config_charts "github.com/goptics/vizb/config/charts"
	barchart "github.com/goptics/vizb/config/charts/bar"
	linechart "github.com/goptics/vizb/config/charts/line"
	"github.com/goptics/vizb/pkg/template"
	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/testutil"
	"github.com/stretchr/testify/suite"
)

// UISuite covers the ui/html subcommand end-to-end via rootCmd.Execute.
type UISuite struct {
	suite.Suite
	restoreOsExit func()
}

func (s *UISuite) SetupTest() {
	ResetTestState()
	s.restoreOsExit, _ = testutil.TrapOsExitPanic(s.T())
}

func (s *UISuite) TearDownTest() {
	s.restoreOsExit()
}

func (s *UISuite) barFromJSON(payload map[string]any) *barchart.Config {
	raw, err := json.Marshal(payload)
	s.Require().NoError(err)
	cfg, err := config_charts.Decode("bar", raw)
	s.Require().NoError(err)
	c, ok := cfg.(*barchart.Config)
	s.Require().True(ok, "expected *barchart.Config, got %T", cfg)
	return c
}

func (s *UISuite) TestSingleObject() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "bench.json")
	testutil.WriteJSON(s.T(), input, shared.Dataset{Name: "Bench1", Data: []shared.DataPoint{{Name: "Test1", XAxis: "1", YAxis: "100"}}})
	out := filepath.Join(dir, "output.html")

	rootCmd.SetArgs([]string{"ui", "-o", out, input})
	s.Require().NoError(rootCmd.Execute())

	html := s.read(out)
	s.Contains(html, "Bench1")
	s.Contains(html, "Test1")
	s.Contains(html, "<html")
	s.Contains(html, "</html>")
}

func (s *UISuite) TestArrayInput() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "benches.json")
	testutil.WriteJSON(s.T(), input, []shared.Dataset{
		{Name: "Bench1", Data: []shared.DataPoint{{Name: "Test1", XAxis: "1", YAxis: "100"}}},
		{Name: "Bench2", Data: []shared.DataPoint{{Name: "Test2", XAxis: "2", YAxis: "200"}}},
	})
	out := filepath.Join(dir, "output.html")

	rootCmd.SetArgs([]string{"ui", "-o", out, input})
	s.Require().NoError(rootCmd.Execute())

	html := s.read(out)
	s.Contains(html, "Bench1")
	s.Contains(html, "Bench2")
}

func (s *UISuite) TestMissingFileExits() {
	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	rootCmd.SetArgs([]string{"ui", "/nonexistent/path.json"})
	s.Panics(func() { _ = rootCmd.Execute() })
	s.True(*exitCalled)
}

func (s *UISuite) TestNoArgsExits() {
	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	rootCmd.SetArgs([]string{"ui"})
	s.Panics(func() { _ = rootCmd.Execute() })
	s.True(*exitCalled)
}

func (s *UISuite) TestUIRemoteDataURL() {
	dir := s.T().TempDir()
	out := filepath.Join(dir, "remote.html")

	rootCmd.SetArgs([]string{"ui", "-o", out, "--data-url", "https://example.com/bench.json"})
	s.Require().NoError(rootCmd.Execute())

	html := s.read(out)
	s.Contains(html, "https://example.com/bench.json")
	s.Contains(html, "<html")
	s.NotContains(html, `"name":"Bench`)
	s.Equal([]string{"bar", "line", "pie"}, s.extractVIZBCharts(html))
	s.False(s.htmlContains3DChunk(html), "remote UI without --3d must not bundle the 3D chunk")
}

func (s *UISuite) TestUIRemoteDataURLWith3D() {
	dir := s.T().TempDir()
	out := filepath.Join(dir, "remote-3d.html")

	rootCmd.SetArgs([]string{"ui", "-o", out, "--data-url", "https://example.com/bench.json", "--3d"})
	s.Require().NoError(rootCmd.Execute())

	s.True(s.htmlContains3DChunk(s.read(out)), "remote UI with --3d must bundle the 3D chunk")
}

func (s *UISuite) TestUIRemoteDataURLInvalidExits() {
	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	rootCmd.SetArgs([]string{"ui", "--data-url", "not-a-url"})
	s.Panics(func() { _ = rootCmd.Execute() })
	s.True(*exitCalled)
}

func (s *UISuite) TestInvalidJSONExits() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "invalid.json")
	s.Require().NoError(os.WriteFile(input, []byte("not json"), 0644))

	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	rootCmd.SetArgs([]string{"ui", input})
	s.Panics(func() { _ = rootCmd.Execute() })
	s.True(*exitCalled)
}

func (s *UISuite) TestEmptyFileExits() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "empty.json")
	s.Require().NoError(os.WriteFile(input, []byte(""), 0644))

	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	rootCmd.SetArgs([]string{"ui", input})
	s.Panics(func() { _ = rootCmd.Execute() })
	s.True(*exitCalled)
}

func (s *UISuite) TestMergedOutput() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "merged.json")
	testutil.WriteJSON(s.T(), input, []shared.Dataset{{
		Tag:       "v1.0.0",
		Timestamp: "2024-01-01T00:00:00Z",
		Name:      "Sort Benchmarks",
		History:   []shared.HistoryEntry{{Tag: "v0.9.0", Timestamp: "2023-12-01T00:00:00Z"}},
		Data:      []shared.DataPoint{{Name: "v1.0.0", XAxis: "v0.9.0", YAxis: "100"}},
	}})
	out := filepath.Join(dir, "chart.html")

	rootCmd.SetArgs([]string{"ui", "-o", out, input})
	s.Require().NoError(rootCmd.Execute())

	html := s.read(out)
	s.Contains(html, "v1.0.0")
	s.Contains(html, "v0.9.0")
	s.Contains(html, "Sort Benchmarks")
}

func (s *UISuite) TestHTMLAlias() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "bench.json")
	testutil.WriteJSON(s.T(), input, shared.Dataset{Name: "Bench1", Data: []shared.DataPoint{{Name: "Test1", XAxis: "1", YAxis: "100"}}})
	out := filepath.Join(dir, "output.html")

	rootCmd.SetArgs([]string{"html", "-o", out, input})
	s.Require().NoError(rootCmd.Execute())

	s.Contains(s.read(out), "Bench1")
}

func (s *UISuite) read(path string) string {
	content, err := os.ReadFile(path)
	s.Require().NoError(err)
	return string(content)
}

func (s *UISuite) extractVIZBDataArray(html string) []any {
	const prefix = "window.VIZB_DATA = "
	start := strings.Index(html, prefix)
	s.Require().NotEqual(-1, start, "expected window.VIZB_DATA in HTML")
	start += len(prefix)
	end := strings.Index(html[start:], ";")
	s.Require().NotEqual(-1, end, "expected ';' after window.VIZB_DATA")
	var data []any
	s.Require().NoError(json.Unmarshal([]byte(html[start:start+end]), &data))
	return data
}

func (s *UISuite) extractVIZBCharts(html string) []string {
	const prefix = "window.VIZB_CHARTS = "
	start := strings.Index(html, prefix)
	s.Require().NotEqual(-1, start, "expected window.VIZB_CHARTS in HTML")
	start += len(prefix)
	end := strings.Index(html[start:], ";")
	s.Require().NotEqual(-1, end, "expected ';' after window.VIZB_CHARTS")
	var charts []string
	s.Require().NoError(json.Unmarshal([]byte(html[start:start+end]), &charts))
	return charts
}

func (s *UISuite) htmlContains3DChunk(html string) bool {
	root3D := template.VizbChartRoots["3d"]
	s.Require().NotEmpty(root3D, "generated VizbChartRoots must contain 3d")
	return strings.Contains(html, `"`+root3D+`"`)
}

func (s *UISuite) TestRunUIFiltersChartsOnExplicitFlag() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "multi.json")
	testutil.WriteJSON(s.T(), input, shared.Dataset{
		Name: "Test",
		Settings: []config_charts.ChartConfig{
			s.barFromJSON(map[string]any{"type": "bar", "scale": "linear"}),
			&linechart.Config{Type: "line", Scale: "linear"},
		},
		Data: []shared.DataPoint{{Name: "Test1", XAxis: "1", YAxis: "100"}},
	})
	out := filepath.Join(dir, "output.html")

	rootCmd.SetArgs([]string{"ui", "-o", out, "-c", "bar", input})
	s.Require().NoError(rootCmd.Execute())

	html := s.read(out)
	s.Equal([]string{"bar"}, s.extractVIZBCharts(html))

	datasets := s.extractVIZBDataArray(html)
	s.Require().Len(datasets, 1)
	ds := datasets[0].(map[string]any)
	settings := ds["settings"].([]any)
	s.Require().Len(settings, 1)
	s.Equal("bar", settings[0].(map[string]any)["type"])
}

func (s *UISuite) TestRunUIAppliesOverrides() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "old.json")
	testutil.WriteJSON(s.T(), input, shared.Dataset{
		Name: "Test",
		Axes: []shared.Axis{{Key: "x"}, {Key: "y"}, {Key: "name"}},
		Settings: []config_charts.ChartConfig{s.barFromJSON(map[string]any{
			"type":  "bar",
			"swap":  "xyn",
			"scale": "linear",
		})},
		Data: []shared.DataPoint{{Name: "Test1", XAxis: "1", YAxis: "100"}},
	})
	out := filepath.Join(dir, "output.html")

	rootCmd.SetArgs([]string{"ui", "-o", out, "--chart", "bar:swap=yxn", input})
	s.Require().NoError(rootCmd.Execute())

	html := s.read(out)
	datasets := s.extractVIZBDataArray(html)
	s.Require().Len(datasets, 1)

	ds, ok := datasets[0].(map[string]any)
	s.Require().True(ok, "expected dataset object, got %T", datasets[0])

	settings, ok := ds["settings"].([]any)
	s.Require().True(ok, "expected settings array in VIZB_DATA, got %T", ds["settings"])
	s.Require().Len(settings, 1)

	bar, ok := settings[0].(map[string]any)
	s.Require().True(ok, "expected bar config object, got %T", settings[0])
	s.Equal("yxn", bar["swap"], "override should replace baked swap value")
}

func TestUISuite(t *testing.T) {
	suite.Run(t, new(UISuite))
}
