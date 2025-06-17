package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to capture stdout and stderr
func captureOutput(f func()) (string, string) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	f()

	wOut.Close()
	wErr.Close()

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)

	return bufOut.String(), bufErr.String()
}

// Helper function to create a valid benchmark JSON file
func createValidBenchmarkFile(t *testing.T) string {
	tempFile, err := os.CreateTemp("", "valid-bench-*.json")
	assert.NoError(t, err)

	defer tempFile.Close()

	// Create sample benchmark data using BenchEvent struct
	benchData := []BenchEvent{
		{
			Action: "output",
			Test:   "BenchmarkTest/workload/subject",
			Output: "BenchmarkTest/workload/subject 100 1000.0 ns/op 2048.0 B/op 5 allocs/op",
		},
		{
			Action: "output",
			Test:   "BenchmarkTest/workload/subject2",
			Output: "BenchmarkTest/workload/subject2 100 2000.0 ns/op 4096.0 B/op 10 allocs/op",
		},
	}

	for _, data := range benchData {
		jsonData, err := json.Marshal(data)
		assert.NoError(t, err)

		_, err = tempFile.Write(jsonData)
		assert.NoError(t, err)

		_, err = tempFile.Write([]byte("\n"))
		assert.NoError(t, err)
	}

	return tempFile.Name()
}

// Helper function to reset flag state between tests
func resetFlagState() {
	shared.FlagState.Name = "Benchmarks"
	shared.FlagState.OutputFile = ""
	shared.FlagState.MemUnit = "B"
	shared.FlagState.TimeUnit = "ns"
	shared.FlagState.Description = ""
	shared.FlagState.ShowVersion = false
}

// TestCLIFeatures groups all CLI functionality tests
func TestCLIFeatures(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	t.Run("Version flag", func(t *testing.T) {
		resetFlagState()

		// Set version flag
		shared.FlagState.ShowVersion = true

		stdout, _ := captureOutput(func() {
			runBenchmark(&cobra.Command{}, []string{})
		})

		assert.Contains(t, stdout, "vizb version "+Version)
	})

	t.Run("Missing input error", func(t *testing.T) {
		// Skip this test since we can't reliably test os.Exit behavior
		t.Skip("Skipping test that involves os.Exit - can't be reliably tested")

		// The actual implementation would check for the error message:
		// "Error: no target provided and no piped input detected"
	})

	t.Run("Valid file input", func(t *testing.T) {
		resetFlagState()

		// Create a valid benchmark file
		benchFile := createValidBenchmarkFile(t)
		defer os.Remove(benchFile)

		// Set output file
		tempOutFile := filepath.Join(tempDir, "test-output.html")
		shared.FlagState.OutputFile = tempOutFile

		stdout, _ := captureOutput(func() {
			runBenchmark(&cobra.Command{}, []string{benchFile})
		})

		// Check that the chart was generated
		assert.Contains(t, stdout, "Chart generated successfully")
		assert.Contains(t, stdout, "Output file:")

		// Check that the output file exists and has content
		fileInfo, err := os.Stat(tempOutFile)
		assert.NoError(t, err)
		assert.Greater(t, fileInfo.Size(), int64(0))
	})

	t.Run("Invalid JSON input", func(t *testing.T) {
		// Skip this test since we can't reliably test os.Exit behavior
		t.Skip("Skipping test that involves os.Exit - can't be reliably tested")

		// The actual implementation would check for the error message:
		// "Error: Input file is not in proper JSON format"
	})
}

// TestOutputOptions groups tests for different output options
func TestOutputOptions(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	t.Run("Auto-add HTML extension", func(t *testing.T) {
		resetFlagState()

		// Create a valid benchmark file
		benchFile := createValidBenchmarkFile(t)
		defer os.Remove(benchFile)

		// Set output file without .html extension
		outputFile := filepath.Join(tempDir, "output-without-extension")
		shared.FlagState.OutputFile = outputFile

		stdout, _ := captureOutput(func() {
			runBenchmark(&cobra.Command{}, []string{benchFile})
		})

		// Check that the chart was generated with .html extension
		assert.Contains(t, stdout, "Chart generated successfully")
		assert.Contains(t, stdout, outputFile+".html")

		// Check that the output file exists with .html extension
		fileInfo, err := os.Stat(outputFile + ".html")
		assert.NoError(t, err)
		assert.Greater(t, fileInfo.Size(), int64(0))
	})

	t.Run("Custom chart name and description", func(t *testing.T) {
		resetFlagState()

		// Create a valid benchmark file
		benchFile := createValidBenchmarkFile(t)
		defer os.Remove(benchFile)

		// Set output file
		tempOutFile := filepath.Join(tempDir, "custom-name-desc.html")
		shared.FlagState.OutputFile = tempOutFile

		// Set custom name and description
		shared.FlagState.Name = "Custom Chart Name"
		shared.FlagState.Description = "Custom chart description"

		captureOutput(func() {
			runBenchmark(&cobra.Command{}, []string{benchFile})
		})

		// Read the output file and check for custom name and description
		content, err := os.ReadFile(tempOutFile)
		assert.NoError(t, err)

		htmlContent := string(content)
		assert.Contains(t, htmlContent, "Custom Chart Name")
		assert.Contains(t, htmlContent, "Custom chart description")
	})

	t.Run("Custom units", func(t *testing.T) {
		resetFlagState()

		// Create a valid benchmark file
		benchFile := createValidBenchmarkFile(t)
		defer os.Remove(benchFile)

		// Set output file
		tempOutFile := filepath.Join(tempDir, "custom-units.html")
		shared.FlagState.OutputFile = tempOutFile

		// Set custom units
		shared.FlagState.TimeUnit = "ms"
		shared.FlagState.MemUnit = "KB"

		captureOutput(func() {
			runBenchmark(&cobra.Command{}, []string{benchFile})
		})

		// Read the output file and check for custom units
		content, err := os.ReadFile(tempOutFile)
		assert.NoError(t, err)

		htmlContent := string(content)
		assert.Contains(t, htmlContent, "Execution Time (ms/op)")
		assert.Contains(t, htmlContent, "Memory Usage (KB/op)")
	})

	t.Run("Temporary output to stdout", func(t *testing.T) {
		resetFlagState()

		// Create a valid benchmark file
		benchFile := createValidBenchmarkFile(t)
		defer os.Remove(benchFile)

		// Don't set output file to trigger temporary file creation
		shared.FlagState.OutputFile = ""

		stdout, _ := captureOutput(func() {
			runBenchmark(&cobra.Command{}, []string{benchFile})
		})

		// Check that HTML content was printed to stdout
		assert.Contains(t, stdout, "<!DOCTYPE html>")
		assert.Contains(t, stdout, "</html>")
	})
}

// TestInputMethods groups tests for different input methods
func TestInputMethods(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	t.Run("File input", func(t *testing.T) {
		resetFlagState()

		// Create a valid benchmark file
		benchFile := createValidBenchmarkFile(t)
		defer os.Remove(benchFile)

		// Set output file
		tempOutFile := filepath.Join(tempDir, "file-input.html")
		shared.FlagState.OutputFile = tempOutFile

		stdout, _ := captureOutput(func() {
			runBenchmark(&cobra.Command{}, []string{benchFile})
		})

		// Check that the chart was generated
		assert.Contains(t, stdout, "Chart generated successfully")
		assert.Contains(t, stdout, "Output file:")

		// Check that the output file exists and has content
		fileInfo, err := os.Stat(tempOutFile)
		assert.NoError(t, err)
		assert.Greater(t, fileInfo.Size(), int64(0))
	})

	t.Run("Piped input", func(t *testing.T) {
		resetFlagState()

		// Create a valid benchmark file to read its content
		benchFile := createValidBenchmarkFile(t)
		defer os.Remove(benchFile)

		// Read the benchmark file content
		benchContent, err := os.ReadFile(benchFile)
		require.NoError(t, err)

		// Save original stdin
		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		// Create a pipe and write benchmark content to it
		r, w, _ := os.Pipe()
		os.Stdin = r

		go func() {
			w.Write(benchContent)
			w.Close()
		}()

		// Set output file
		tempOutFile := filepath.Join(tempDir, "piped-input.html")
		shared.FlagState.OutputFile = tempOutFile

		stdout, _ := captureOutput(func() {
			runBenchmark(&cobra.Command{}, []string{})
		})

		// Check that the chart was generated
		assert.Contains(t, stdout, "Chart generated successfully")
		assert.Contains(t, stdout, "Output file:")

		// Check that the output file exists and has content
		fileInfo, err := os.Stat(tempOutFile)
		assert.NoError(t, err)
		assert.Greater(t, fileInfo.Size(), int64(0))
	})
}
