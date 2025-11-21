package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/shared/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Save the real os.Exit function in a variable
var osExit = shared.OsExit
var originalOsExit = os.Exit

// Override os.Exit for testing
func TestMain(m *testing.M) {
	// Replace the os.Exit function with a custom version for testing
	// that doesn't actually exit the process
	osExit = func(code int) {
		// In tests, don't actually exit
		panic(fmt.Sprintf("os.Exit(%d) was called", code))
	}

	// Run tests
	code := m.Run()

	// Restore the real os.Exit
	osExit = originalOsExit

	// Exit with the code from the test run
	osExit(code)
}

// TestValidateFlags tests that flag validation works correctly
func TestValidateFlags(t *testing.T) {
	// Save original state to restore after tests
	origMemUnit := shared.FlagState.MemUnit
	origTimeUnit := shared.FlagState.TimeUnit
	origAllocUnit := shared.FlagState.NumberUnit
	origFormat := shared.FlagState.Format

	defer func() {
		// Restore original flag values
		shared.FlagState.MemUnit = origMemUnit
		shared.FlagState.TimeUnit = origTimeUnit
		shared.FlagState.NumberUnit = origAllocUnit
		shared.FlagState.Format = origFormat
	}()

	tests := []struct {
		name              string
		setupFlags        func()
		expectedMemUnit   string
		expectedTimeUnit  string
		expectedAllocUnit string
		expectedFormat    string
		expectedOutput    string
	}{
		{
			name: "Valid flags",
			setupFlags: func() {
				shared.FlagState.MemUnit = "b"
				shared.FlagState.TimeUnit = "ns"
				shared.FlagState.NumberUnit = ""
				shared.FlagState.Format = "html"
			},
			expectedMemUnit:   "b",
			expectedTimeUnit:  "ns",
			expectedAllocUnit: "",
			expectedFormat:    "html",
			expectedOutput:    "",
		},
		{
			name: "Invalid memory unit",
			setupFlags: func() {
				shared.FlagState.MemUnit = "invalid"
				shared.FlagState.TimeUnit = "ns"
				shared.FlagState.NumberUnit = ""
				shared.FlagState.Format = "html"
			},
			expectedMemUnit:   "b",
			expectedTimeUnit:  "ns",
			expectedAllocUnit: "",
			expectedFormat:    "html",
			expectedOutput:    "Warning: Invalid memory unit 'invalid'. Using default 'b'",
		},
		{
			name: "Invalid time unit",
			setupFlags: func() {
				shared.FlagState.MemUnit = "b"
				shared.FlagState.TimeUnit = "invalid"
				shared.FlagState.NumberUnit = ""
				shared.FlagState.Format = "html"
			},
			expectedMemUnit:   "b",
			expectedTimeUnit:  "ns",
			expectedAllocUnit: "",
			expectedFormat:    "html",
			expectedOutput:    "Warning: Invalid time unit 'invalid'. Using default 'ns'",
		},
		{
			name: "Invalid alloc unit",
			setupFlags: func() {
				shared.FlagState.MemUnit = "b"
				shared.FlagState.TimeUnit = "ns"
				shared.FlagState.NumberUnit = "invalid"
				shared.FlagState.Format = "html"
			},
			expectedMemUnit:   "b",
			expectedTimeUnit:  "ns",
			expectedAllocUnit: "",
			expectedFormat:    "html",
			expectedOutput:    "Warning: Invalid number unit 'INVALID'. Using default ''",
		},
		{
			name: "Invalid format",
			setupFlags: func() {
				shared.FlagState.MemUnit = "b"
				shared.FlagState.TimeUnit = "ns"
				shared.FlagState.NumberUnit = ""
				shared.FlagState.Format = "invalid"
			},
			expectedMemUnit:   "b",
			expectedTimeUnit:  "ns",
			expectedAllocUnit: "",
			expectedFormat:    "html",
			expectedOutput:    "Warning: Invalid format 'invalid'. Using default 'html'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the flags
			tt.setupFlags()

			// Capture stderr output
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Call validation function to trigger validation and warning messages
			utils.ApplyValidationRules(flagValidationRules)

			// Close write end of pipe to get all output
			w.Close()
			// Read output data
			var buf bytes.Buffer
			io.Copy(&buf, r)
			// Restore stderr
			os.Stderr = oldStderr

			// Check that the flags were updated correctly after validation
			assert.Equal(t, tt.expectedMemUnit, shared.FlagState.MemUnit)
			assert.Equal(t, tt.expectedTimeUnit, shared.FlagState.TimeUnit)
			assert.Equal(t, tt.expectedAllocUnit, shared.FlagState.NumberUnit)
			assert.Equal(t, tt.expectedFormat, shared.FlagState.Format)

			// Check the stderr output if expected
			if tt.expectedOutput != "" {
				assert.Contains(t, buf.String(), tt.expectedOutput)
			}
		})
	}
}

// TestCheckTargetFile tests the JSON file processing functionality
func TestCheckTargetFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Mock shared.OsExit since checkTargetFile calls shared.ExitWithError
	exitCalled := false
	originalOsExitFunc := shared.OsExit
	shared.OsExit = func(code int) {
		exitCalled = true
		panic(fmt.Sprintf("shared.OsExit(%d) called", code)) // Use panic for flow control in tests
	}
	defer func() { shared.OsExit = originalOsExitFunc }()

	// Create valid and invalid JSON test files
	validJsonPath := filepath.Join(tempDir, "valid.json")
	invalidJsonPath := filepath.Join(tempDir, "invalid.json")
	nonExistentPath := filepath.Join(tempDir, "nonexistent.json")

	// Create a valid JSON test file
	validEvent := shared.BenchEvent{Action: "output", Output: "BenchmarkTest 1 100 ns/op"}
	validEventBytes, _ := json.Marshal(validEvent)
	err := os.WriteFile(validJsonPath, append(validEventBytes, '\n'), 0644)
	require.NoError(t, err)

	// Create an invalid JSON test file with .json extension to trigger JSON validation
	err = os.WriteFile(invalidJsonPath, []byte("this is not json"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name          string
		inputPath     string
		expectExit    bool
		expectedError string
	}{
		{
			name:          "Valid JSON file",
			inputPath:     validJsonPath,
			expectExit:    false,
			expectedError: "",
		},
		{
			name:          "Non-existent file",
			inputPath:     nonExistentPath,
			expectExit:    true,
			expectedError: "does not exist",
		},
		{
			name:          "Invalid JSON file",
			inputPath:     invalidJsonPath,
			expectExit:    true,
			expectedError: "not in proper JSON format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the exitCalled flag for each test
			exitCalled = false

			// Prepare stderr capture
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Call the function and handle any os.Exit panics
			if tt.expectExit {
				func() {
					defer func() {
						recovered := recover()
						// Close the write end of pipe to get all output
						w.Close()

						// Read the stderr output
						var buf bytes.Buffer
						io.Copy(&buf, r)

						// Restore stderr
						os.Stderr = oldStderr

						// Check assertions
						if recovered == nil {
							t.Error("Expected shared.OsExit to be called")
						}
						assert.True(t, exitCalled, "shared.OsExit should have been called")
						assert.Contains(t, buf.String(), tt.expectedError)
					}()
					checkTargetFile(tt.inputPath)
				}()
			} else {
				checkTargetFile(tt.inputPath)

				// Close write end of pipe
				w.Close()
				os.Stderr = oldStderr

				assert.False(t, exitCalled, "shared.OsExit should not have been called")
			}
		})
	}
}

// TestRunBenchmark tests the main runBenchmark function
func TestRunBenchmark(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create a valid text benchmark file (avoid JSON conversion issues)
	validTextPath := filepath.Join(tempDir, "valid.txt")
	textContent := `BenchmarkTest-8    1000000    1234 ns/op    1000 B/op    10 allocs/op`
	err := os.WriteFile(validTextPath, []byte(textContent), 0644)
	require.NoError(t, err)

	// Save original flag state
	origOutputFile := shared.FlagState.OutputFile
	origFormat := shared.FlagState.Format
	defer func() {
		shared.FlagState.OutputFile = origOutputFile
		shared.FlagState.Format = origFormat
	}()

	// Mock shared.OsExit since runBenchmark uses shared.ExitWithError
	exitCalled := false
	oldOsExit := shared.OsExit
	defer func() { shared.OsExit = oldOsExit }()
	shared.OsExit = func(code int) {
		exitCalled = true
		panic(fmt.Sprintf("shared.OsExit(%d) called", code)) // Use panic for flow control in tests
	}

	tests := []struct {
		name           string
		argsFunc       func() []string
		setupStdin     func() (restoreFunc func())
		expectExit     bool
		expectedOutput string
		setupFlags     func()
	}{
		{
			name: "Valid file input",
			argsFunc: func() []string {
				return []string{validTextPath}
			},
			setupStdin:     func() func() { return func() {} },
			expectExit:     false,
			expectedOutput: "Generated",
			setupFlags: func() {
				shared.FlagState.Format = "html"
				shared.FlagState.OutputFile = filepath.Join(tempDir, "out.html")
			},
		},
		{
			name: "No args and no stdin",
			argsFunc: func() []string {
				return []string{}
			},
			setupStdin:     func() func() { return func() {} },
			expectExit:     true,
			expectedOutput: "no target provided",
			setupFlags:     func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset exit flag
			exitCalled = false

			// Set up flags
			tt.setupFlags()

			// Setup stdin if needed
			restore := tt.setupStdin()
			defer restore()

			// Capture stdout and stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

			// Create a cobra command for testing
			cmd := &cobra.Command{}
			args := tt.argsFunc()

			// Run the function and catch any os.Exit calls
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
							t.Error("Expected shared.OsExit to be called")
						}
						assert.True(t, exitCalled, "shared.OsExit should have been called")
						assert.Contains(t, buf.String(), tt.expectedOutput)
					}()
					runBenchmark(cmd, args)
				}()
			} else {
				runBenchmark(cmd, args)

				// Close write end of pipe
				w.Close()

				// Read output
				var buf bytes.Buffer
				io.Copy(&buf, r)

				// Restore stdout and stderr
				os.Stdout = oldStdout
				os.Stderr = oldStderr

				// Validate assertions
				assert.Contains(t, buf.String(), tt.expectedOutput)
			}
		})
	}
}
