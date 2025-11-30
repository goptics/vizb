package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckTargetFile is already implemented in root_test.go - no need to duplicate

func TestPreprocessInputFile(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("JSON file gets preprocessed", func(t *testing.T) {
		jsonFile := filepath.Join(tempDir, "bench.json")

		// Create valid JSON benchmark content
		jsonContent := `{"Action":"run","Test":"BenchmarkExample"}
{"Action":"pass","Test":"BenchmarkExample","Output":"1000000    1234 ns/op"}`

		err := os.WriteFile(jsonFile, []byte(jsonContent), 0644)
		require.NoError(t, err)

		result := preprocessInputFile(jsonFile)

		// Should return a different file (preprocessed)
		assert.NotEqual(t, jsonFile, result)
		assert.FileExists(t, result)

		// Clean up the preprocessed file
		os.Remove(result)
	})

	t.Run("Non-JSON file returns same path", func(t *testing.T) {
		textFile := filepath.Join(tempDir, "bench.txt")

		textContent := `BenchmarkExample-8    1000000    1234 ns/op`
		err := os.WriteFile(textFile, []byte(textContent), 0644)
		require.NoError(t, err)

		result := preprocessInputFile(textFile)

		// Should return the same file
		assert.Equal(t, textFile, result)
	})

	t.Run("Invalid file returns same path", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "does_not_exist.txt")

		result := preprocessInputFile(nonExistentFile)

		// Should return the same path even if file doesn't exist
		assert.Equal(t, nonExistentFile, result)
	})
}

func TestParseResults(t *testing.T) {
	// Save original shared.OsExit
	originalOsExit := shared.OsExit
	defer func() { shared.OsExit = originalOsExit }()

	tempDir := t.TempDir()

	t.Run("Valid benchmark results", func(t *testing.T) {
		benchFile := filepath.Join(tempDir, "valid_bench.txt")

		benchContent := `BenchmarkExample-8    1000000    1234 ns/op    1000 B/op    10 allocs/op
BenchmarkAnother-8    2000000    2345 ns/op    2000 B/op    20 allocs/op`

		err := os.WriteFile(benchFile, []byte(benchContent), 0644)
		require.NoError(t, err)

		results := parseResults(benchFile)

		assert.NotEmpty(t, results, "Should parse benchmark results")
		assert.True(t, len(results) > 0, "Should have at least one result")
	})

	t.Run("Empty results causes exit", func(t *testing.T) {
		emptyFile := filepath.Join(tempDir, "empty_bench.txt")

		err := os.WriteFile(emptyFile, []byte(""), 0644)
		require.NoError(t, err)

		exitCalled := false
		shared.OsExit = func(code int) {
			exitCalled = true
			panic(fmt.Sprintf("shared.OsExit(%d) called", code))
		}

		assert.Panics(t, func() {
			parseResults(emptyFile)
		})
		assert.True(t, exitCalled, "Should call osExit when no results found")
	})

	t.Run("Invalid benchmark file", func(t *testing.T) {
		invalidFile := filepath.Join(tempDir, "invalid_bench.txt")

		invalidContent := `This is not a benchmark file
Just some random text`

		err := os.WriteFile(invalidFile, []byte(invalidContent), 0644)
		require.NoError(t, err)

		exitCalled := false
		shared.OsExit = func(code int) {
			exitCalled = true
			panic(fmt.Sprintf("shared.OsExit(%d) called", code))
		}

		assert.Panics(t, func() {
			parseResults(invalidFile)
		})
		assert.True(t, exitCalled, "Should call osExit when no valid benchmarks found")
	})
}

func TestWriteOutput(t *testing.T) {
	// Save original osExit
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	tempDir := t.TempDir()

	// Create sample benchmark results
	sampleResults := []shared.BenchmarkResult{
		{
			Name:  "BenchmarkExample",
			XAxis: "8",
			YAxis: "Example",
			Stats: []shared.Stat{
				{Type: "time", Value: 1234},
				{Type: "memory", Value: 1000},
				{Type: "allocs", Value: 10},
			},
		},
		{
			Name:  "BenchmarkAnother",
			XAxis: "8",
			YAxis: "Another",
			Stats: []shared.Stat{
				{Type: "time", Value: 2345},
				{Type: "memory", Value: 2000},
				{Type: "allocs", Value: 20},
			},
		},
	}

	t.Run("HTML output", func(t *testing.T) {
		htmlFile := filepath.Join(tempDir, "output.html")
		file, err := os.Create(htmlFile)
		require.NoError(t, err)
		defer file.Close()

		assert.NotPanics(t, func() {
			writeOutput(file, sampleResults, "html")
		})

		// Verify file was written
		stat, err := file.Stat()
		require.NoError(t, err)
		assert.True(t, stat.Size() > 0, "HTML file should not be empty")
	})

	t.Run("JSON output", func(t *testing.T) {
		jsonFile := filepath.Join(tempDir, "output.json")
		file, err := os.Create(jsonFile)
		require.NoError(t, err)
		defer file.Close()

		assert.NotPanics(t, func() {
			writeOutput(file, sampleResults, "json")
		})

		// Verify JSON content
		file.Seek(0, 0)
		content, err := os.ReadFile(jsonFile)
		require.NoError(t, err)

		var bench shared.Benchmark
		err = json.Unmarshal(content, &bench)
		assert.NoError(t, err, "Should produce valid JSON")
		assert.Len(t, bench.Data, 2, "Should have 2 benchmark results")
	})

	t.Run("Invalid format", func(t *testing.T) {
		invalidFile := filepath.Join(tempDir, "output.invalid")
		file, err := os.Create(invalidFile)
		require.NoError(t, err)
		defer file.Close()

		// Invalid format should not cause panic, but file should remain empty
		assert.NotPanics(t, func() {
			writeOutput(file, sampleResults, "invalid_format")
		})

		stat, err := file.Stat()
		require.NoError(t, err)
		assert.Equal(t, int64(0), stat.Size(), "Invalid format should not write content")
	})

	t.Run("Empty results", func(t *testing.T) {
		emptyFile := filepath.Join(tempDir, "empty_output.json")
		file, err := os.Create(emptyFile)
		require.NoError(t, err)
		defer file.Close()

		emptyResults := []shared.BenchmarkResult{}

		assert.NotPanics(t, func() {
			writeOutput(file, emptyResults, "json")
		})

		// Verify empty JSON array was written
		content, err := os.ReadFile(emptyFile)
		require.NoError(t, err)
		// The output should be a Benchmark struct with empty Data
		var bench shared.Benchmark
		err = json.Unmarshal(content, &bench)
		assert.NoError(t, err)
		assert.Empty(t, bench.Data, "Should have empty Data")
	})
}

func TestGenerateOutputFile(t *testing.T) {
	// Save original flag state and shared.OsExit
	origOutputFile := shared.FlagState.OutputFile
	origFormat := shared.FlagState.Format
	origOsExit := shared.OsExit
	defer func() {
		shared.FlagState.OutputFile = origOutputFile
		shared.FlagState.Format = origFormat
		shared.OsExit = origOsExit
	}()

	// Mock shared.OsExit to prevent actual exits during tests
	shared.OsExit = func(code int) {
		panic(fmt.Sprintf("shared.OsExit(%d) called", code))
	}

	tempDir := t.TempDir()

	t.Run("Generate HTML output file", func(t *testing.T) {
		// Create valid benchmark input file
		benchFile := filepath.Join(tempDir, "input_bench.txt")
		benchContent := `BenchmarkExample-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`

		err := os.WriteFile(benchFile, []byte(benchContent), 0644)
		require.NoError(t, err)

		shared.FlagState.OutputFile = filepath.Join(tempDir, "output.html")
		shared.FlagState.Format = "html"

		assert.NotPanics(t, func() {
			generateOutputFile(benchFile)
		})

		// Verify output file was created
		assert.FileExists(t, shared.FlagState.OutputFile)

		// Verify file has content
		stat, err := os.Stat(shared.FlagState.OutputFile)
		require.NoError(t, err)
		assert.True(t, stat.Size() > 0, "Output file should not be empty")
	})

	t.Run("Generate JSON output file", func(t *testing.T) {
		benchFile := filepath.Join(tempDir, "input_json.txt")
		benchContent := `BenchmarkJSON-8    500000    2468 ns/op    500 B/op    5 allocs/op`

		err := os.WriteFile(benchFile, []byte(benchContent), 0644)
		require.NoError(t, err)

		shared.FlagState.OutputFile = filepath.Join(tempDir, "output.json")
		shared.FlagState.Format = "json"

		assert.NotPanics(t, func() {
			generateOutputFile(benchFile)
		})

		// Verify JSON output
		assert.FileExists(t, shared.FlagState.OutputFile)
		content, err := os.ReadFile(shared.FlagState.OutputFile)
		require.NoError(t, err)

		var bench shared.Benchmark
		err = json.Unmarshal(content, &bench)
		assert.NoError(t, err, "Should produce valid JSON")
	})

	t.Run("Temp file output", func(t *testing.T) {
		benchFile := filepath.Join(tempDir, "temp_input.txt")
		benchContent := `BenchmarkTemp-8    750000    1357 ns/op    750 B/op    7 allocs/op`

		err := os.WriteFile(benchFile, []byte(benchContent), 0644)
		require.NoError(t, err)

		// Empty output file should create temp file
		shared.FlagState.OutputFile = ""
		shared.FlagState.Format = "html"

		// Capture stdout to prevent noisy output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Read from pipe in a separate goroutine to prevent deadlock
		// if the output exceeds the pipe buffer size
		done := make(chan struct{})
		go func() {
			io.Copy(io.Discard, r)
			close(done)
		}()

		assert.NotPanics(t, func() {
			generateOutputFile(benchFile)
		})

		// Restore stdout
		w.Close()
		os.Stdout = oldStdout
		<-done
	})
}

// Integration test for the full workflow
func TestOutputWorkflowIntegration(t *testing.T) {
	tempDir := t.TempDir()

	// Save original flag state and shared.OsExit
	origOutputFile := shared.FlagState.OutputFile
	origFormat := shared.FlagState.Format
	origOsExit := shared.OsExit
	defer func() {
		shared.FlagState.OutputFile = origOutputFile
		shared.FlagState.Format = origFormat
		shared.OsExit = origOsExit
	}()

	// Mock shared.OsExit to prevent actual exits during tests
	shared.OsExit = func(code int) {
		panic(fmt.Sprintf("shared.OsExit(%d) called", code))
	}

	t.Run("End-to-end text workflow", func(t *testing.T) {
		// Create simple text benchmark input that doesn't require JSON conversion
		txtFile := filepath.Join(tempDir, "bench.txt")
		txtContent := `BenchmarkWorkflow-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`

		err := os.WriteFile(txtFile, []byte(txtContent), 0644)
		require.NoError(t, err)

		shared.FlagState.OutputFile = filepath.Join(tempDir, "workflow_output.json")
		shared.FlagState.Format = "json"

		assert.NotPanics(t, func() {
			generateOutputFile(txtFile)
		})

		// Verify the complete workflow worked
		assert.FileExists(t, shared.FlagState.OutputFile)
		content, err := os.ReadFile(shared.FlagState.OutputFile)
		require.NoError(t, err)

		var bench shared.Benchmark
		err = json.Unmarshal(content, &bench)
		assert.NoError(t, err)
		assert.True(t, len(bench.Data) > 0, "Should have processed benchmark results")
	})
}
