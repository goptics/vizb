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
	bench1 := shared.Dataset{
		Name: "Bench1",
		Data: []shared.DataPoint{
			{Name: "Test1", XAxis: "1", YAxis: "100"},
		},
	}
	bench2 := shared.Dataset{
		Name: "Bench2",
		Data: []shared.DataPoint{
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
	outFile := filepath.Join(tmpDir, "merged.json")

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

	content, err := os.ReadFile(outFile)
	assert.NoError(t, err)

	var parsed []shared.Dataset
	assert.NoError(t, json.Unmarshal(content, &parsed))
	assert.Len(t, parsed, 2)

	names := make(map[string]bool)
	for _, b := range parsed {
		names[b.Name] = true
	}
	assert.True(t, names["Bench1"])
	assert.True(t, names["Bench2"])
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

	bench1 := shared.Dataset{
		Name: "Bench1",
		Data: []shared.DataPoint{{Name: "Test1"}},
	}
	writeJSON(t, filepath.Join(tmpDir, "b1.json"), bench1)

	outFile := filepath.Join(tmpDir, "merged_dir.json")

	// Reset FlagState
	shared.FlagState.OutputFile = outFile
	shared.FlagState.Name = "Comparisons"
	shared.FlagState.Charts = []string{"bar", "line", "pie"}

	rootCmd.SetArgs([]string{"merge", tmpDir})
	err = rootCmd.Execute()

	if exitCode != 0 {
		t.Errorf("OsExit called with code %d", exitCode)
	}
	assert.NoError(t, err)

	content, err := os.ReadFile(outFile)
	assert.NoError(t, err)

	var parsed []shared.Dataset
	assert.NoError(t, json.Unmarshal(content, &parsed))
	assert.Len(t, parsed, 1)
	assert.Equal(t, "Bench1", parsed[0].Name)
}

func TestMergeCmd_ArrayInput(t *testing.T) {
	oldOsExit := shared.OsExit
	defer func() { shared.OsExit = oldOsExit }()

	var exitCode int
	shared.OsExit = func(code int) { exitCode = code }

	tmpDir, err := os.MkdirTemp("", "vizb-merge-array-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Single file with an array of 2 benchmarks
	benches := []shared.Dataset{
		{Name: "Bench1", Data: []shared.DataPoint{{Name: "Test1", XAxis: "1", YAxis: "100"}}},
		{Name: "Bench2", Data: []shared.DataPoint{{Name: "Test2", XAxis: "2", YAxis: "200"}}},
	}
	writeJSON(t, filepath.Join(tmpDir, "array.json"), benches)

	outFile := filepath.Join(tmpDir, "merged_array.json")
	shared.FlagState.OutputFile = outFile

	rootCmd.SetArgs([]string{"merge", filepath.Join(tmpDir, "array.json")})
	err = rootCmd.Execute()

	if exitCode != 0 {
		t.Errorf("OsExit called with code %d", exitCode)
	}
	assert.NoError(t, err)

	content, err := os.ReadFile(outFile)
	assert.NoError(t, err)

	var parsed []shared.Dataset
	assert.NoError(t, json.Unmarshal(content, &parsed))
	assert.Len(t, parsed, 2)

	names := make(map[string]bool)
	for _, b := range parsed {
		names[b.Name] = true
	}
	assert.True(t, names["Bench1"])
	assert.True(t, names["Bench2"])
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

func TestMergeCmd_JSONOutput(t *testing.T) {
	oldOsExit := shared.OsExit
	defer func() { shared.OsExit = oldOsExit }()

	var exitCode int
	shared.OsExit = func(code int) { exitCode = code }

	tmpDir, err := os.MkdirTemp("", "vizb-merge-json-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	bench1 := shared.Dataset{
		Name: "Bench1",
		Data: []shared.DataPoint{
			{Name: "Test1", XAxis: "1", YAxis: "100"},
		},
	}
	writeJSON(t, filepath.Join(tmpDir, "b1.json"), bench1)

	outFile := filepath.Join(tmpDir, "merged.json")
	shared.FlagState.OutputFile = outFile

	rootCmd.SetArgs([]string{"merge", filepath.Join(tmpDir, "b1.json")})
	err = rootCmd.Execute()

	if exitCode != 0 {
		t.Errorf("OsExit called with code %d", exitCode)
	}
	assert.NoError(t, err)

	content, err := os.ReadFile(outFile)
	assert.NoError(t, err)

	var parsed []shared.Dataset
	assert.NoError(t, json.Unmarshal(content, &parsed))
	assert.Len(t, parsed, 1)
	assert.Equal(t, "Bench1", parsed[0].Name)
}
