package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
)

func TestHtmlCmd_SingleObject(t *testing.T) {
	oldOsExit := shared.OsExit
	defer func() { shared.OsExit = oldOsExit }()

	var exitCode int
	shared.OsExit = func(code int) { exitCode = code }

	tmpDir, err := os.MkdirTemp("", "vizb-html-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	bench := shared.Benchmark{
		Name: "Bench1",
		Data: []shared.BenchmarkData{
			{Name: "Test1", XAxis: "1", YAxis: "100"},
		},
	}
	inputFile := filepath.Join(tmpDir, "bench.json")
	writeJSON(t, inputFile, bench)

	outFile := filepath.Join(tmpDir, "output.html")
	shared.FlagState.OutputFile = outFile

	rootCmd.SetArgs([]string{"html", inputFile})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, 0, exitCode)

	content, err := os.ReadFile(outFile)
	assert.NoError(t, err)

	htmlStr := string(content)
	assert.Contains(t, htmlStr, "Bench1")
	assert.Contains(t, htmlStr, "Test1")
	assert.Contains(t, htmlStr, "<html")
	assert.Contains(t, htmlStr, "</html>")
}

func TestHtmlCmd_ArrayInput(t *testing.T) {
	oldOsExit := shared.OsExit
	defer func() { shared.OsExit = oldOsExit }()

	var exitCode int
	shared.OsExit = func(code int) { exitCode = code }

	tmpDir, err := os.MkdirTemp("", "vizb-html-array-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	benches := []shared.Benchmark{
		{Name: "Bench1", Data: []shared.BenchmarkData{{Name: "Test1", XAxis: "1", YAxis: "100"}}},
		{Name: "Bench2", Data: []shared.BenchmarkData{{Name: "Test2", XAxis: "2", YAxis: "200"}}},
	}
	inputFile := filepath.Join(tmpDir, "benches.json")
	writeJSON(t, inputFile, benches)

	outFile := filepath.Join(tmpDir, "output.html")
	shared.FlagState.OutputFile = outFile

	rootCmd.SetArgs([]string{"html", inputFile})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, 0, exitCode)

	content, err := os.ReadFile(outFile)
	assert.NoError(t, err)

	htmlStr := string(content)
	assert.Contains(t, htmlStr, "Bench1")
	assert.Contains(t, htmlStr, "Bench2")
	assert.Contains(t, htmlStr, "Test1")
	assert.Contains(t, htmlStr, "Test2")
}

func TestHtmlCmd_MissingFile(t *testing.T) {
	oldOsExit := shared.OsExit
	defer func() { shared.OsExit = oldOsExit }()

	exitCode := 0
	shared.OsExit = func(code int) { exitCode = code }

	rootCmd.SetArgs([]string{"html", "/nonexistent/path.json"})
	rootCmd.Execute()
	assert.Equal(t, 1, exitCode)
}

func TestHtmlCmd_NoArgs(t *testing.T) {
	oldOsExit := shared.OsExit
	defer func() { shared.OsExit = oldOsExit }()

	exitCode := 0
	shared.OsExit = func(code int) { exitCode = code }

	shared.FlagState.API = ""
	rootCmd.SetArgs([]string{"html"})
	rootCmd.Execute()
	assert.Equal(t, 1, exitCode)
}

func TestHtmlCmd_APIFlag(t *testing.T) {
	oldOsExit := shared.OsExit
	defer func() { shared.OsExit = oldOsExit }()

	var exitCode int
	shared.OsExit = func(code int) { exitCode = code }

	tmpDir, err := os.MkdirTemp("", "vizb-html-api-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	outFile := filepath.Join(tmpDir, "remote.html")
	shared.FlagState.OutputFile = outFile
	shared.FlagState.API = "https://example.com/bench.json"
	defer func() { shared.FlagState.API = "" }()

	rootCmd.SetArgs([]string{"html", "--api", "https://example.com/bench.json"})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, 0, exitCode)

	content, err := os.ReadFile(outFile)
	assert.NoError(t, err)

	htmlStr := string(content)
	assert.Contains(t, htmlStr, "https://example.com/bench.json")
	assert.Contains(t, htmlStr, "<html")
	assert.NotContains(t, htmlStr, `"name":"Bench`)
}

func TestHtmlCmd_APIFlag_InvalidURL(t *testing.T) {
	oldOsExit := shared.OsExit
	defer func() { shared.OsExit = oldOsExit }()

	exitCode := 0
	shared.OsExit = func(code int) { exitCode = code }

	shared.FlagState.API = "not-a-url"
	defer func() { shared.FlagState.API = "" }()

	rootCmd.SetArgs([]string{"html", "--api", "not-a-url"})
	rootCmd.Execute()
	assert.Equal(t, 1, exitCode)
}

func TestHtmlCmd_InvalidJSON(t *testing.T) {
	oldOsExit := shared.OsExit
	defer func() { shared.OsExit = oldOsExit }()

	exitCode := 0
	shared.OsExit = func(code int) { exitCode = code }

	tmpDir, err := os.MkdirTemp("", "vizb-html-invalid-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	inputFile := filepath.Join(tmpDir, "invalid.json")
	os.WriteFile(inputFile, []byte("not json"), 0644)

	rootCmd.SetArgs([]string{"html", inputFile})
	rootCmd.Execute()
	assert.Equal(t, 1, exitCode)
}

func TestHtmlCmd_StdoutFallback(t *testing.T) {
	oldOsExit := shared.OsExit
	defer func() { shared.OsExit = oldOsExit }()

	var exitCode int
	shared.OsExit = func(code int) { exitCode = code }

	tmpDir, err := os.MkdirTemp("", "vizb-html-stdout-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	bench := shared.Benchmark{
		Name: "Bench1",
		Data: []shared.BenchmarkData{
			{Name: "Test1", XAxis: "1", YAxis: "100"},
		},
	}
	inputFile := filepath.Join(tmpDir, "bench.json")
	writeJSON(t, inputFile, bench)

	shared.FlagState.OutputFile = ""

	rootCmd.SetArgs([]string{"html", inputFile})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, 0, exitCode)
}

func TestHtmlCmd_EmptyFile(t *testing.T) {
	oldOsExit := shared.OsExit
	defer func() { shared.OsExit = oldOsExit }()

	exitCode := 0
	shared.OsExit = func(code int) { exitCode = code }

	tmpDir, err := os.MkdirTemp("", "vizb-html-empty-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	inputFile := filepath.Join(tmpDir, "empty.json")
	os.WriteFile(inputFile, []byte(""), 0644)

	rootCmd.SetArgs([]string{"html", inputFile})
	rootCmd.Execute()
	assert.Equal(t, 1, exitCode)
}

func TestHtmlCmd_MergedOutput(t *testing.T) {
	oldOsExit := shared.OsExit
	defer func() { shared.OsExit = oldOsExit }()

	var exitCode int
	shared.OsExit = func(code int) { exitCode = code }

	tmpDir, err := os.MkdirTemp("", "vizb-html-merged-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	benches := []shared.Benchmark{
		{
			Tag:       "v1.0.0",
			Timestamp: "2024-01-01T00:00:00Z",
			Name:      "Sort Benchmarks",
			History: []shared.HistoryEntry{
				{Tag: "v0.9.0", Timestamp: "2023-12-01T00:00:00Z"},
			},
			Data: []shared.BenchmarkData{
				{Name: "v1.0.0", XAxis: "v0.9.0", YAxis: "100"},
			},
		},
	}
	inputFile := filepath.Join(tmpDir, "merged.json")
	writeJSON(t, inputFile, benches)

	outFile := filepath.Join(tmpDir, "chart.html")
	shared.FlagState.OutputFile = outFile

	rootCmd.SetArgs([]string{"html", inputFile})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, 0, exitCode)

	content, err := os.ReadFile(outFile)
	assert.NoError(t, err)

	htmlStr := string(content)
	assert.Contains(t, htmlStr, "v1.0.0")
	assert.Contains(t, htmlStr, "v0.9.0")
	assert.Contains(t, htmlStr, "Sort Benchmarks")
}
