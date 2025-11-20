package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleOutputResultBehavior tests the handleOutputResult function behavior
func TestHandleOutputResultBehavior(t *testing.T) {
	tempDir := t.TempDir()

	// Save original flag state
	originalOutputFile := shared.FlagState.OutputFile
	defer func() { shared.FlagState.OutputFile = originalOutputFile }()

	t.Run("Output file specified - shows file path", func(t *testing.T) {
		shared.FlagState.OutputFile = "specified_output.html"

		// Create temp file with content
		filename := filepath.Join(tempDir, "test_output.txt")
		testContent := "<html>Test Content</html>"
		err := os.WriteFile(filename, []byte(testContent), 0644)
		require.NoError(t, err)
		defer os.Remove(filename)

		// Open file
		file, err := os.Open(filename)
		require.NoError(t, err)
		defer file.Close()

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Execute handleOutputResult
		HandleOutputResult(file)

		// Read stdout
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()

		// Should show file path message, not content
		assert.Contains(t, output, "📄 Output file:", "Should show output file message")
		assert.Contains(t, output, filename, "Should show actual file name")
		assert.NotContains(t, output, testContent, "Should not print file content when output file is specified")
		assert.NotContains(t, output, "\033[H\033[2J", "Should not clear screen when output file is specified")
	})

	t.Run("No output file - shows content with screen clear", func(t *testing.T) {
		shared.FlagState.OutputFile = ""

		// Create temp file with content
		filename := filepath.Join(tempDir, "stdout_test.txt")
		testContent := "Test benchmark output content\nLine 2\nLine 3"
		err := os.WriteFile(filename, []byte(testContent), 0644)
		require.NoError(t, err)
		defer os.Remove(filename)

		// Open file
		file, err := os.Open(filename)
		require.NoError(t, err)
		defer file.Close()

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Execute handleOutputResult
		HandleOutputResult(file)

		// Read stdout
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()

		// Should clear screen and show content
		assert.Contains(t, output, "\033[H\033[2J", "Should clear screen when no output file")
		assert.Contains(t, output, testContent, "Should print file content to stdout")
		assert.NotContains(t, output, "📄 Output file:", "Should not show file path message when printing to stdout")
	})

	t.Run("Empty file content", func(t *testing.T) {
		shared.FlagState.OutputFile = ""

		// Create empty file
		filename := filepath.Join(tempDir, "empty_test.txt")
		err := os.WriteFile(filename, []byte(""), 0644)
		require.NoError(t, err)
		defer os.Remove(filename)

		file, err := os.Open(filename)
		require.NoError(t, err)
		defer file.Close()

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		HandleOutputResult(file)

		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()
		assert.Contains(t, output, "\033[H\033[2J", "Should clear screen even for empty files")
		// The output should only contain the clear sequence, no additional content
	})

	t.Run("Large file content handling", func(t *testing.T) {
		shared.FlagState.OutputFile = ""

		// Create file with large content
		filename := filepath.Join(tempDir, "large_test.txt")
		largeContent := strings.Repeat("Large content line\n", 1000)
		err := os.WriteFile(filename, []byte(largeContent), 0644)
		require.NoError(t, err)
		defer os.Remove(filename)

		file, err := os.Open(filename)
		require.NoError(t, err)
		defer file.Close()

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		HandleOutputResult(file)

		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()
		assert.Contains(t, output, "Large content line", "Should handle large files")
		assert.Contains(t, output, "\033[H\033[2J", "Should clear screen for large files")
	})

	t.Run("File with special characters", func(t *testing.T) {
		shared.FlagState.OutputFile = ""

		// Create file with special characters
		filename := filepath.Join(tempDir, "special_test.txt")
		specialContent := "Content with émojis 🎉 and spëcial chàracters ñ"
		err := os.WriteFile(filename, []byte(specialContent), 0644)
		require.NoError(t, err)
		defer os.Remove(filename)

		file, err := os.Open(filename)
		require.NoError(t, err)
		defer file.Close()

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		HandleOutputResult(file)

		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		os.Stdout = oldStdout

		output := buf.String()
		assert.Contains(t, output, specialContent, "Should handle special characters correctly")
	})
}

// TestHandleOutputResultErrorConditions tests error handling in handleOutputResult
func TestHandleOutputResultErrorConditions(t *testing.T) {
	// Save original state
	originalOutputFile := shared.FlagState.OutputFile
	originalOsExit := osExit
	defer func() {
		shared.FlagState.OutputFile = originalOutputFile
		osExit = originalOsExit
	}()

	// Force content reading mode
	shared.FlagState.OutputFile = ""

	// Mock osExit to track calls
	osExit = func(code int) {
		panic(fmt.Sprintf("osExit(%d) called", code))
	}
}

// TestWriteStdinPipedInputsErrorScenarios tests error handling in writeStdinPipedInputs
func TestWriteStdinPipedInputsErrorScenarios(t *testing.T) {
	// Save original shared.OsExit since MustOpenFile uses shared.ExitWithError
	originalOsExit := shared.OsExit
	defer func() { shared.OsExit = originalOsExit }()

	t.Run("Invalid temp file path causes exit", func(t *testing.T) {
		// Create a temp directory structure, then create the directory we'll try to use
		tempDir := t.TempDir()
		removedDir := filepath.Join(tempDir, "removed_dir")

		// Create the directory first
		err := os.MkdirAll(removedDir, 0755)
		require.NoError(t, err)

		// Now remove it to create a non-existent path scenario
		err = os.RemoveAll(removedDir)
		require.NoError(t, err)

		invalidPath := filepath.Join(removedDir, "temp.file")

		// Mock shared.OsExit to track calls
		exitCalled := false
		exitCode := -1
		shared.OsExit = func(code int) {
			exitCalled = true
			exitCode = code
			panic(fmt.Sprintf("shared.OsExit(%d) called", code))
		}

		assert.Panics(t, func() {
			writeStdinPipedInputs(invalidPath)
		}, "Should panic when cannot open temp file")

		assert.True(t, exitCalled, "shared.OsExit should be called when file cannot be opened")
		assert.Equal(t, 1, exitCode, "Should exit with code 1")
	})

	t.Run("Permission denied on temp file causes exit", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a file but make it inaccessible
		restrictedFile := filepath.Join(tempDir, "restricted.tmp")
		err := os.WriteFile(restrictedFile, []byte(""), 0644)
		require.NoError(t, err)

		// Make file unreadable and unwritable
		err = os.Chmod(restrictedFile, 0000)
		require.NoError(t, err)
		defer func() {
			os.Chmod(restrictedFile, 0644) // Restore for cleanup
			os.Remove(restrictedFile)
		}()

		exitCalled := false
		shared.OsExit = func(code int) {
			exitCalled = true
			panic(fmt.Sprintf("shared.OsExit(%d) called", code))
		}

		assert.Panics(t, func() {
			writeStdinPipedInputs(restrictedFile)
		}, "Should panic on permission denied")

		assert.True(t, exitCalled, "shared.OsExit should be called on permission error")
	})
}

// TestConstants validates that constants are properly defined and used
func TestConstantsValidation(t *testing.T) {
	t.Run("Constants have expected values", func(t *testing.T) {
		assert.Equal(t, "vizb-benchmark", shared.TempBenchFilePrefix,
			"tempBenchFilePrefix should have expected value")

	})

	t.Run("Constants have reasonable length", func(t *testing.T) {
		assert.True(t, len(shared.TempBenchFilePrefix) > 0 && len(shared.TempBenchFilePrefix) < 50,
			"tempBenchFilePrefix should have reasonable length")
	})
}

// TestGenerateOutputFileFlow tests the overall flow without complex dependencies
func TestGenerateOutputFileFlow(t *testing.T) {
	t.Run("resolveOutputFileName integration", func(t *testing.T) {
		// Test the filename resolution works correctly
		testCases := []struct {
			input    string
			format   string
			expected func(string) bool // validation function
		}{
			{
				input:  "custom",
				format: "json",
				expected: func(result string) bool {
					return result == "custom.json"
				},
			},
			{
				input:  "report.html",
				format: "html",
				expected: func(result string) bool {
					return result == "report.html"
				},
			},
		}

		for i, tc := range testCases {
			t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
				result := resolveOutputFileName(tc.input, tc.format)
				assert.True(t, tc.expected(result),
					"Case %d failed: input=%q, format=%q, result=%q", i, tc.input, tc.format, result)

				// Clean up temp files
				if tc.input == "" {
					os.Remove(result)
				}
			})
		}
	})
}
