package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

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
	uiOpts = uiOptions{Charts: []string{"bar", "line", "pie", "heatmap"}}
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

func TestUISuite(t *testing.T) {
	suite.Run(t, new(UISuite))
}
