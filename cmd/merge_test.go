package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
)

func TestMergeCmd(t *testing.T) {
	// Mock OsExit
	oldOsExit := shared.OsExit
	defer func() { shared.OsExit = oldOsExit }()

	var exitCode int
	shared.OsExit = func(code int) {
		exitCode = code
	}

	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "vizb-merge-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create two dummy benchmark files
	bench1 := shared.Benchmark{
		Name: "Bench1",
		Data: []shared.BenchmarkData{
			{Name: "Test1", XAxis: "1", YAxis: "100"},
		},
	}
	bench2 := shared.Benchmark{
		Name: "Bench2",
		Data: []shared.BenchmarkData{
			{Name: "Test2", XAxis: "2", YAxis: "200"},
		},
	}

	file1 := filepath.Join(tmpDir, "bench1.json")
	file2 := filepath.Join(tmpDir, "bench2.json")

	writeJSON(t, file1, bench1)
	writeJSON(t, file2, bench2)

	// Create an invalid file
	file3 := filepath.Join(tmpDir, "invalid.json")
	os.WriteFile(file3, []byte("{invalid json"), 0644)

	// Output file
	outFile := filepath.Join(tmpDir, "merged.html")

	// Reset FlagState
	shared.FlagState.OutputFile = outFile

	// Execute via rootCmd
	rootCmd.SetArgs([]string{"merge", file1, file2, file3})
	err = rootCmd.Execute()

	// Check if OsExit was called with error
	if exitCode != 0 {
		t.Errorf("OsExit called with code %d", exitCode)
	}

	assert.NoError(t, err)

	// Verify output file exists
	_, err = os.Stat(outFile)
	assert.NoError(t, err)

	content, _ := os.ReadFile(outFile)
	htmlStr := string(content)
	assert.Contains(t, htmlStr, "Test1")
	assert.Contains(t, htmlStr, "Test2")
}

func TestMergeCmd_Directory(t *testing.T) {
	// Mock OsExit
	oldOsExit := shared.OsExit
	defer func() { shared.OsExit = oldOsExit }()

	var exitCode int
	shared.OsExit = func(code int) {
		exitCode = code
	}

	tmpDir, err := os.MkdirTemp("", "vizb-merge-dir-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	bench1 := shared.Benchmark{
		Name: "Bench1",
		Data: []shared.BenchmarkData{{Name: "Test1"}},
	}
	writeJSON(t, filepath.Join(tmpDir, "b1.json"), bench1)

	outFile := filepath.Join(tmpDir, "merged_dir.html")

	// Reset FlagState
	shared.FlagState.OutputFile = outFile
	shared.FlagState.Name = "Benchmarks"
	shared.FlagState.Charts = []string{"bar", "line", "pie"}

	rootCmd.SetArgs([]string{"merge", tmpDir})
	err = rootCmd.Execute()

	if exitCode != 0 {
		t.Errorf("OsExit called with code %d", exitCode)
	}
	assert.NoError(t, err)

	content, _ := os.ReadFile(outFile)
	assert.Contains(t, string(content), "Test1")
}

func writeJSON(t *testing.T, path string, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}
