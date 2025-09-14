package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteStdinPipedInputs tests the functionality of writeStdinPipedInputs
func TestWriteStdinPipedInputs(t *testing.T) {
	// Save original stdin
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()

	// Save original stderr and stdout
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	t.Run("Basic functionality with simulated stdin", func(t *testing.T) {
		// Create temp file for output
		tempDir := t.TempDir()
		tempfile := filepath.Join(tempDir, "test_output.txt")

		// Create temp file to simulate stdin input
		stdinFile, err := os.CreateTemp("", "stdin_test")
		require.NoError(t, err)
		defer os.Remove(stdinFile.Name())

		// Write test data to the stdin file
		testData := []string{
			`{"Action":"run","Test":"BenchmarkExample"}`,
			`{"Action":"pass","Test":"BenchmarkExample","Output":"1000 ns/op"}`,
			"BenchmarkAnotherTest-8 \t1000\t2000 ns/op",
		}
		for _, line := range testData {
			stdinFile.WriteString(line + "\n")
		}
		stdinFile.Seek(0, 0) // Reset to beginning

		os.Stdin = stdinFile

		// Execute the function
		writeStdinPipedInputs(tempfile)

		// Read the temp file to verify content was written
		content, err := os.ReadFile(tempfile)
		require.NoError(t, err)

		fileContent := string(content)

		// Verify that content was written correctly
		assert.Contains(t, fileContent, `{"Action":"run","Test":"BenchmarkExample"}`)
		assert.Contains(t, fileContent, `{"Action":"pass","Test":"BenchmarkExample","Output":"1000 ns/op"}`)
		assert.Contains(t, fileContent, "BenchmarkAnotherTest-8")
	})

	t.Run("Empty input", func(t *testing.T) {
		tempDir := t.TempDir()
		tempfile := filepath.Join(tempDir, "empty_output.txt")

		// Create empty stdin file
		stdinFile, err := os.CreateTemp("", "empty_stdin_test")
		require.NoError(t, err)
		defer os.Remove(stdinFile.Name())
		stdinFile.Seek(0, 0)

		os.Stdin = stdinFile

		// Execute the function
		writeStdinPipedInputs(tempfile)

		// Read the temp file - should be empty
		content, err := os.ReadFile(tempfile)
		require.NoError(t, err)
		assert.Empty(t, string(content), "Temp file should be empty")
	})

	t.Run("Large input", func(t *testing.T) {
		tempDir := t.TempDir()
		tempfile := filepath.Join(tempDir, "large_output.txt")

		// Create stdin file with lots of data
		stdinFile, err := os.CreateTemp("", "large_stdin_test")
		require.NoError(t, err)
		defer os.Remove(stdinFile.Name())

		// Generate large amount of test data
		for i := 0; i < 100; i++ {
			line := `{"Action":"run","Test":"BenchmarkTest` + strings.Repeat("X", i%10) + `"}`
			stdinFile.WriteString(line + "\n")
		}
		stdinFile.Seek(0, 0)

		os.Stdin = stdinFile

		// Execute the function
		writeStdinPipedInputs(tempfile)

		// Verify content was written
		content, err := os.ReadFile(tempfile)
		require.NoError(t, err)

		fileContent := string(content)
		assert.Contains(t, fileContent, "BenchmarkTest")

		// Count lines
		lines := strings.Split(strings.TrimSpace(fileContent), "\n")
		assert.Equal(t, 100, len(lines), "Should have written all 100 lines")
	})

	t.Run("JSON with benchmark content", func(t *testing.T) {
		tempDir := t.TempDir()
		tempfile := filepath.Join(tempDir, "json_output.txt")

		stdinFile, err := os.CreateTemp("", "json_stdin_test")
		require.NoError(t, err)
		defer os.Remove(stdinFile.Name())

		jsonData := []string{
			`{"Action":"start"}`,
			`{"Action":"run","Test":"BenchmarkStringBuilder"}`,
			`{"Action":"output","Test":"BenchmarkStringBuilder","Output":"BenchmarkStringBuilder-8   "}`,
			`{"Action":"output","Test":"BenchmarkStringBuilder","Output":"1000000"}`,
			`{"Action":"output","Test":"BenchmarkStringBuilder","Output":"1234 ns/op"}`,
			`{"Action":"pass","Test":"BenchmarkStringBuilder"}`,
		}

		for _, line := range jsonData {
			stdinFile.WriteString(line + "\n")
		}
		stdinFile.Seek(0, 0)

		os.Stdin = stdinFile

		writeStdinPipedInputs(tempfile)

		content, err := os.ReadFile(tempfile)
		require.NoError(t, err)

		fileContent := string(content)
		assert.Contains(t, fileContent, "BenchmarkStringBuilder")
		assert.Contains(t, fileContent, "ns/op")
		assert.Contains(t, fileContent, "1000000")
	})

	t.Run("Mixed JSON and text format", func(t *testing.T) {
		tempDir := t.TempDir()
		tempfile := filepath.Join(tempDir, "mixed_output.txt")

		stdinFile, err := os.CreateTemp("", "mixed_stdin_test")
		require.NoError(t, err)
		defer os.Remove(stdinFile.Name())

		mixedData := []string{
			"goos: linux",
			"goarch: amd64",
			`{"Action":"run","Test":"BenchmarkExample"}`,
			"BenchmarkExample-8    1000000    1234 ns/op",
			`{"Action":"pass","Test":"BenchmarkExample"}`,
			"PASS",
		}

		for _, line := range mixedData {
			stdinFile.WriteString(line + "\n")
		}
		stdinFile.Seek(0, 0)

		os.Stdin = stdinFile

		writeStdinPipedInputs(tempfile)

		content, err := os.ReadFile(tempfile)
		require.NoError(t, err)

		fileContent := string(content)
		assert.Contains(t, fileContent, "goos: linux")
		assert.Contains(t, fileContent, "BenchmarkExample")
		assert.Contains(t, fileContent, "1234 ns/op")
		assert.Contains(t, fileContent, "PASS")
	})
}