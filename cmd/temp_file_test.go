package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCreateBenchTempJsonFile tests both success and error paths of createBenchTempJsonFile
func TestCreateBenchTempJsonFile(t *testing.T) {
	// Test success path - already covered by the stdin tests but included here for completeness
	t.Run("Success Path", func(t *testing.T) {
		// Call the function
		tempFile := createTempFile(tempBenchFilePrefix, "json")
		defer os.Remove(tempFile)

		// Verify temp file was created with correct pattern
		assert.True(t, filepath.Base(tempFile) != "", "Should have created a file with a name")
		assert.Contains(t, tempFile, tempBenchFilePrefix, "Should contain the prefix in name")
		assert.Contains(t, tempFile, ".json", "Should have .json extension")
		assert.FileExists(t, tempFile, "Temp file should exist")
	})

	// Test error path - CreateTemp fails
	t.Run("Error Path", func(t *testing.T) {
		// Save original functions to restore them later
		origCreateTemp := osTempCreate
		origOsExit := osExit
		defer func() {
			osTempCreate = origCreateTemp
			osExit = origOsExit
		}()

		// Mock os.CreateTemp to return an error
		osTempCreate = func(dir, pattern string) (*os.File, error) {
			return nil, errors.New("simulated CreateTemp error")
		}

		// Mock osExit to capture exit code
		exitCalled := false
		exitCode := 0
		osExit = func(code int) {
			exitCalled = true
			exitCode = code
			panic(fmt.Sprintf("os.Exit(%d) called", code)) // Use panic for flow control in tests
		}

		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w
		defer func() { os.Stderr = oldStderr }()

		// Call function and handle expected panic
		assert.Panics(t, func() {
			createTempFile(tempBenchFilePrefix, "json")
		}, "Should panic when os.CreateTemp fails")

		// Close write end of pipe
		w.Close()

		// Read captured stderr output
		var output []byte
		output, _ = io.ReadAll(r)

		// Assert exit was called with correct code
		assert.True(t, exitCalled, "osExit should have been called")
		assert.Equal(t, 1, exitCode, "Exit code should be 1")
		assert.Contains(t, string(output), "Error creating temporary file", "Should show error message")
		assert.Contains(t, string(output), "simulated CreateTemp error", "Should include the error details")
	})
}
