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
	"github.com/stretchr/testify/suite"
)

// barFromJSON round-trips a map through JSON to produce a *barchart.Config
// whose fields are populated per the map. Mirrors what MigrateDataset does
// for legacy settings: we get a typed Config without manually setting fields.
func (s *UISuite) barFromJSON(payload map[string]any) *barchart.Config {
	raw, err := json.Marshal(payload)
	s.Require().NoError(err)
	cfg, err := config_charts.Decode("bar", raw)
	s.Require().NoError(err)
	c, ok := cfg.(*barchart.Config)
	s.Require().True(ok, "expected *barchart.Config, got %T", cfg)
	return c
}

// UISuite covers the ui/html subcommand end-to-end via rootCmd.Execute.
type UISuite struct {
	suite.Suite
	origOsExit func(int)
	exitCode   int
}

func (s *UISuite) SetupTest() {
	s.origOsExit = shared.OsExit
	s.exitCode = 0
	shared.OsExit = func(code int) { s.exitCode = code }

	// Reset the ui flag bindings so an un-passed flag doesn't inherit a prior
	// test's value (cobra retains bound values between Execute calls).
	uiOpts = uiOptions{Charts: shared.DefaultChartTypes}
}

func (s *UISuite) TearDownTest() {
	shared.OsExit = s.origOsExit
}

func (s *UISuite) writeJSON(path string, v any) {
	data, err := json.Marshal(v)
	s.Require().NoError(err)
	s.Require().NoError(os.WriteFile(path, data, 0644))
}

func (s *UISuite) TestSingleObject() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "bench.json")
	s.writeJSON(input, shared.Dataset{Name: "Bench1", Data: []shared.DataPoint{{Name: "Test1", XAxis: "1", YAxis: "100"}}})
	out := filepath.Join(dir, "output.html")

	rootCmd.SetArgs([]string{"ui", "-o", out, input})
	s.Require().NoError(rootCmd.Execute())
	s.Equal(0, s.exitCode)

	html := s.read(out)
	s.Contains(html, "Bench1")
	s.Contains(html, "Test1")
	s.Contains(html, "<html")
	s.Contains(html, "</html>")
}

func (s *UISuite) TestArrayInput() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "benches.json")
	s.writeJSON(input, []shared.Dataset{
		{Name: "Bench1", Data: []shared.DataPoint{{Name: "Test1", XAxis: "1", YAxis: "100"}}},
		{Name: "Bench2", Data: []shared.DataPoint{{Name: "Test2", XAxis: "2", YAxis: "200"}}},
	})
	out := filepath.Join(dir, "output.html")

	rootCmd.SetArgs([]string{"ui", "-o", out, input})
	s.Require().NoError(rootCmd.Execute())
	s.Equal(0, s.exitCode)

	html := s.read(out)
	s.Contains(html, "Bench1")
	s.Contains(html, "Bench2")
}

func (s *UISuite) TestMissingFileExits() {
	rootCmd.SetArgs([]string{"ui", "/nonexistent/path.json"})
	rootCmd.Execute()
	s.Equal(1, s.exitCode)
}

func (s *UISuite) TestNoArgsExits() {
	rootCmd.SetArgs([]string{"ui"})
	rootCmd.Execute()
	s.Equal(1, s.exitCode)
}

func (s *UISuite) TestAPIFlag() {
	dir := s.T().TempDir()
	out := filepath.Join(dir, "remote.html")

	rootCmd.SetArgs([]string{"ui", "-o", out, "--data-url", "https://example.com/bench.json"})
	s.Require().NoError(rootCmd.Execute())
	s.Equal(0, s.exitCode)

	html := s.read(out)
	s.Contains(html, "https://example.com/bench.json")
	s.Contains(html, "<html")
	s.NotContains(html, `"name":"Bench`)
	s.Equal([]string{"bar", "line", "pie"}, s.extractVIZBCharts(html))
	s.False(s.htmlContains3DChunk(html), "remote UI without --3d must not bundle the 3D chunk")
}

func (s *UISuite) TestAPIFlagWith3D() {
	dir := s.T().TempDir()
	out := filepath.Join(dir, "remote-3d.html")

	rootCmd.SetArgs([]string{"ui", "-o", out, "--data-url", "https://example.com/bench.json", "--3d"})
	s.Require().NoError(rootCmd.Execute())
	s.Equal(0, s.exitCode)

	s.True(s.htmlContains3DChunk(s.read(out)), "remote UI with --3d must bundle the 3D chunk")
}

func (s *UISuite) TestAPIFlagInvalidURLExits() {
	rootCmd.SetArgs([]string{"ui", "--data-url", "not-a-url"})
	rootCmd.Execute()
	s.Equal(1, s.exitCode)
}

func (s *UISuite) TestInvalidJSONExits() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "invalid.json")
	s.Require().NoError(os.WriteFile(input, []byte("not json"), 0644))

	rootCmd.SetArgs([]string{"ui", input})
	rootCmd.Execute()
	s.Equal(1, s.exitCode)
}

func (s *UISuite) TestEmptyFileExits() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "empty.json")
	s.Require().NoError(os.WriteFile(input, []byte(""), 0644))

	rootCmd.SetArgs([]string{"ui", input})
	rootCmd.Execute()
	s.Equal(1, s.exitCode)
}

func (s *UISuite) TestMergedOutput() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "merged.json")
	s.writeJSON(input, []shared.Dataset{{
		Tag:       "v1.0.0",
		Timestamp: "2024-01-01T00:00:00Z",
		Name:      "Sort Benchmarks",
		History:   []shared.HistoryEntry{{Tag: "v0.9.0", Timestamp: "2023-12-01T00:00:00Z"}},
		Data:      []shared.DataPoint{{Name: "v1.0.0", XAxis: "v0.9.0", YAxis: "100"}},
	}})
	out := filepath.Join(dir, "chart.html")

	rootCmd.SetArgs([]string{"ui", "-o", out, input})
	s.Require().NoError(rootCmd.Execute())
	s.Equal(0, s.exitCode)

	html := s.read(out)
	s.Contains(html, "v1.0.0")
	s.Contains(html, "v0.9.0")
	s.Contains(html, "Sort Benchmarks")
}

// TestHtmlAlias verifies the legacy `html` alias still resolves to the ui command.
func (s *UISuite) TestHtmlAlias() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "bench.json")
	s.writeJSON(input, shared.Dataset{Name: "Bench1", Data: []shared.DataPoint{{Name: "Test1", XAxis: "1", YAxis: "100"}}})
	out := filepath.Join(dir, "output.html")

	rootCmd.SetArgs([]string{"html", "-o", out, input})
	s.Require().NoError(rootCmd.Execute())
	s.Equal(0, s.exitCode)

	s.Contains(s.read(out), "Bench1")
}

func (s *UISuite) read(path string) string {
	content, err := os.ReadFile(path)
	s.Require().NoError(err)
	return string(content)
}

// extractVIZBDataArray parses the JSON array assigned to `window.VIZB_DATA` in
// the generated HTML. The data is inlined by the template (see
// pkg/template/generate-ui.go) as a single JS literal; for the ui subcommand
// the value is always a JSON array of datasets.
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

// TestRunUI_FiltersChartsOnExplicitFlag verifies that `-c bar` trims embedded
// settings to match the bundled chart list.
func (s *UISuite) TestRunUI_FiltersChartsOnExplicitFlag() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "multi.json")
	s.writeJSON(input, shared.Dataset{
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
	s.Equal(0, s.exitCode)

	html := s.read(out)
	s.Equal([]string{"bar"}, s.extractVIZBCharts(html))

	datasets := s.extractVIZBDataArray(html)
	s.Require().Len(datasets, 1)
	ds := datasets[0].(map[string]any)
	settings := ds["settings"].([]any)
	s.Require().Len(settings, 1)
	s.Equal("bar", settings[0].(map[string]any)["type"])
}

// TestRunUI_AppliesOverrides verifies that `--chart bar:swap=yxn` on the
// `vizb ui` subcommand is applied to the bar setting baked into the input
// dataset. The override should appear in the embedded VIZB_DATA, not just
// the chunk-pruning list.
func (s *UISuite) TestRunUI_AppliesOverrides() {
	dir := s.T().TempDir()
	input := filepath.Join(dir, "old.json")
	// Dataset with a bar config whose initial swap is "xyn" (different from
	// the override), so the assertion distinguishes override from no-op.
	s.writeJSON(input, shared.Dataset{
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
	s.Equal(0, s.exitCode)

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
