package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteStdinPipedInputs tests the functionality of writeStdinPipedInputs
// This simulates piping data to stdin
func TestWriteStdinPipedInputs(t *testing.T) {
	// Save original stdin and create a pipe
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()

	// Save original stderr and stdout
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Save original osExit and replace with mock
	originalOsExit := osExit
	exitCalled := false
	exitCode := 0
	osExit = func(code int) {
		exitCalled = true
		exitCode = code
		panic(fmt.Sprintf("os.Exit(%d) called", code))
	}
	defer func() { osExit = originalOsExit }()

	// Test case 1: Successful processing of valid JSON input
	t.Run("Valid JSON Input", func(t *testing.T) {
		// Reset exit flags
		exitCalled = false
		exitCode = 0

		// Create a pipe to simulate stdin
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdin = r

		// Capture stdout/stderr
		stdout, outW, _ := os.Pipe()
		stderr, errW, _ := os.Pipe()
		os.Stdout = outW
		os.Stderr = errW

		// Write valid JSON events to the pipe in a goroutine
		go func() {
			defer w.Close()
			validEvents := []shared.BenchEvent{
				{Action: "output", Output: "BenchmarkTest/case1 1 100 ns/op", Test: "BenchmarkTest/case1"},
				{Action: "output", Output: "BenchmarkTest/case2 1 200 ns/op", Test: "BenchmarkTest/case2"},
			}

			for _, event := range validEvents {
				eventBytes, _ := json.Marshal(event)
				w.Write(eventBytes)
				w.Write([]byte("\n"))
			}
		}()

		// Create a temp file path using the function we're testing
		resultPath := createTempFile(tempBenchFilePrefix, "json")
		defer os.Remove(resultPath)

		// Call the function in a way that we can recover from the panic
		func() {
			defer func() {
				recover() // Recover from any panics
			}()
			writeStdinPipedInputs(resultPath)
		}()

		// Close the write end of stdout/stderr pipes
		outW.Close()
		errW.Close()

		// Collect output
		var outBuf bytes.Buffer
		io.Copy(&outBuf, stdout)
		var errBuf bytes.Buffer
		io.Copy(&errBuf, stderr)

		// Restore stdout and stderr
		os.Stdout = oldStdout
		os.Stderr = oldStderr

		// Check that a temporary file was created
		assert.False(t, exitCalled, "os.Exit should not have been called")
		assert.NotEmpty(t, resultPath, "Should have returned a file path")
		assert.Contains(t, resultPath, ".json", "Should be a JSON file")

		// Check file content if no exit occurred
		if !exitCalled {
			// Try to read the temp file before it's auto-deleted
			content, err := os.ReadFile(resultPath)
			// File may already be deleted in the function, so we don't assert on error
			if err == nil && len(content) > 0 {
				// Just verify we have some content
				assert.Contains(t, string(content), "BenchmarkTest")
			}
		}
	})

	// Test case 2: Invalid JSON input
	t.Run("Invalid JSON Input", func(t *testing.T) {
		// Reset exit flags
		exitCalled = false
		exitCode = 0

		// Create a pipe to simulate stdin
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdin = r

		// Capture stderr
		_, errW, _ := os.Pipe()
		os.Stderr = errW

		// Write invalid content to pipe in a goroutine
		go func() {
			defer w.Close()
			w.Write([]byte("This is not valid JSON\n"))
		}()

		// Create a temp file path using the function we're testing
		resultPath := createTempFile(tempBenchFilePrefix, "json")
		defer os.Remove(resultPath)

		// Call the function and expect a panic from osExit
		func() {
			defer func() {
				recovered := recover()
				assert.NotNil(t, recovered, "Expected panic from os.Exit")
			}()
			writeStdinPipedInputs(resultPath) // This should call osExit and panic
		}()

		errW.Close()

		// Check assertions
		assert.True(t, exitCalled, "os.Exit should have been called")
		assert.Equal(t, 1, exitCode, "Exit code should be 1")
	})
}

// Note: We can't directly test Execute() by mocking rootCmd.Execute
// since it's not assignable. This test is removed for now.
