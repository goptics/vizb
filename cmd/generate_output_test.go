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

// TestGenerateOutputFileExtended tests additional edge cases and error paths
// in the generateOutputFile function
func TestGenerateOutputFileExtended(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Mock os.Exit
	exitCalled := false
	exitCode := 0
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	osExit = func(code int) {
		exitCalled = true
		exitCode = code
		panic(fmt.Sprintf("os.Exit(%d) called", code))
	}

	// Save original flag state and restore after tests
	origOutputFile := shared.FlagState.OutputFile
	origFormat := shared.FlagState.Format
	defer func() {
		shared.FlagState.OutputFile = origOutputFile
		shared.FlagState.Format = origFormat
	}()

	// Create valid and invalid JSON test files
	validJsonPath := filepath.Join(tempDir, "valid.json")
	emptyResultsPath := filepath.Join(tempDir, "empty_results.json")
	invalidResultsPath := filepath.Join(tempDir, "invalid_results.json")
	readOnlyDirPath := filepath.Join(tempDir, "read_only_dir")

	// Create file with valid benchmark data
	err := os.WriteFile(validJsonPath, []byte(`{"Action":"output","Output":"BenchmarkTest 1 100 ns/op","Test":"BenchmarkTest"}`), 0644)
	require.NoError(t, err)

	// Create file with empty results
	err = os.WriteFile(emptyResultsPath, []byte(`{"Action":"output","Output":"no benchmarks"}`), 0644)
	require.NoError(t, err)

	// Create file with invalid content for ParseBenchmarkResults
	err = os.WriteFile(invalidResultsPath, []byte(`{"Action":"invalid"}`), 0644)
	require.NoError(t, err)

	// Create read-only directory to test file creation errors
	err = os.Mkdir(readOnlyDirPath, 0500) // read-only directory
	require.NoError(t, err)

	tests := []struct {
		name           string
		setupFlags     func()
		inputPath      string
		expectExit     bool
		expectedOutput string
		expectedExit   int
	}{
		{
			name: "Auto-generate output filename with HTML format",
			setupFlags: func() {
				shared.FlagState.Format = "html"
				shared.FlagState.OutputFile = ""
			},
			inputPath:      validJsonPath,
			expectExit:     false,
			expectedOutput: "Generated HTML chart",
		},
		{
			name: "Auto-generate output filename with JSON format",
			setupFlags: func() {
				shared.FlagState.Format = "json"
				shared.FlagState.OutputFile = ""
			},
			inputPath:      validJsonPath,
			expectExit:     false,
			expectedOutput: "Generated JSON",
		},
		{
			name: "Output file without extension",
			setupFlags: func() {
				shared.FlagState.Format = "html"
				shared.FlagState.OutputFile = filepath.Join(tempDir, "output_no_ext")
			},
			inputPath:      validJsonPath,
			expectExit:     false,
			expectedOutput: "Generated HTML chart",
		},
		{
			name: "No benchmark results found",
			setupFlags: func() {
				shared.FlagState.Format = "html"
				shared.FlagState.OutputFile = filepath.Join(tempDir, "empty_output.html")
			},
			inputPath:      emptyResultsPath,
			expectExit:     true,
			expectedOutput: "No benchmark results found",
			expectedExit:   1,
		},
		{
			name: "Error parsing benchmark results",
			setupFlags: func() {
				shared.FlagState.Format = "html"
				shared.FlagState.OutputFile = filepath.Join(tempDir, "error_output.html")
			},
			inputPath:      invalidResultsPath,
			expectExit:     true,
			expectedOutput: "No benchmark results found", // This is the actual message produced
			expectedExit:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the exitCalled flag for each test
			exitCalled = false
			exitCode = 0

			// Set up flags
			tt.setupFlags()

			// Capture stdout and stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

			// Call the function and handle any os.Exit panics
			if tt.expectExit {
				func() {
					defer func() {
						recovered := recover()

						// Close write end of pipe
						w.Close()

						// Read output
						var buf bytes.Buffer
						io.Copy(&buf, r)

						// Restore stdout and stderr
						os.Stdout = oldStdout
						os.Stderr = oldStderr

						// Verify assertions
						if recovered == nil {
							t.Error("Expected os.Exit to be called")
						}
						assert.True(t, exitCalled, "os.Exit should have been called")
						assert.Equal(t, tt.expectedExit, exitCode, "Expected exit code %d, got %d", tt.expectedExit, exitCode)
						assert.Contains(t, buf.String(), tt.expectedOutput)
					}()
					generateOutputFile(tt.inputPath)
				}()
			} else {
				func() {
					defer func() {
						if r := recover(); r != nil {
							t.Errorf("Unexpected panic: %v", r)
						}
					}()
					generateOutputFile(tt.inputPath)
				}()

				// Close write end of pipe
				w.Close()

				// Read output
				var buf bytes.Buffer
				io.Copy(&buf, r)

				// Restore stdout and stderr
				os.Stdout = oldStdout
				os.Stderr = oldStderr

				// Check output
				assert.Contains(t, buf.String(), tt.expectedOutput)

				// For non-empty output file, verify it exists with correct extension
				if shared.FlagState.OutputFile != "" {
					// Check if output file exists, either with original name or with extension
					outputWithExt := shared.FlagState.OutputFile
					if !strings.HasSuffix(outputWithExt, fmt.Sprintf(".%s", shared.FlagState.Format)) {
						outputWithExt += fmt.Sprintf(".%s", shared.FlagState.Format)
					}
					assert.FileExists(t, outputWithExt)
				}
			}
		})
	}

	// Test specific case for output to stdout
	t.Run("Output file to stdout", func(t *testing.T) {
		// Reset the exitCalled flag
		exitCalled = false
		exitCode = 0

		// Set up flags
		shared.FlagState.Format = "html"
		shared.FlagState.OutputFile = ""

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Call the function
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Unexpected panic: %v", r)
				}
			}()
			generateOutputFile(validJsonPath)
		}()

		// Close write end of pipe
		w.Close()

		// Read output
		var buf bytes.Buffer
		io.Copy(&buf, r)

		// Restore stdout
		os.Stdout = oldStdout

		// Verify output contains HTML content (looking for standard HTML tags)
		output := buf.String()
		assert.Contains(t, output, "<!DOCTYPE html>")
		assert.Contains(t, output, "<html")
	})
}
