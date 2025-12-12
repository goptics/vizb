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
			expectedMemUnit:   "B",
			expectedTimeUnit:  "ns",
			expectedAllocUnit: "",
			expectedFormat:    "html",
			expectedOutput:    "Warning: Invalid memory unit 'invalid'.",
		},
		{
			name: "Invalid time unit",
			setupFlags: func() {
				shared.FlagState.MemUnit = "B"
				shared.FlagState.TimeUnit = "invalid"
				shared.FlagState.NumberUnit = ""
				shared.FlagState.Format = "html"
			},
			expectedMemUnit:   "B",
			expectedTimeUnit:  "ns",
			expectedAllocUnit: "",
			expectedFormat:    "html",
			expectedOutput:    "Warning: Invalid time unit 'invalid'.",
		},
		{
			name: "Invalid alloc unit",
			setupFlags: func() {
				shared.FlagState.MemUnit = "B"
				shared.FlagState.TimeUnit = "ns"
				shared.FlagState.NumberUnit = "invalid"
				shared.FlagState.Format = "html"
			},
			expectedMemUnit:   "B",
			expectedTimeUnit:  "ns",
			expectedAllocUnit: "",
			expectedFormat:    "html",
			expectedOutput:    "Warning: Invalid number unit 'INVALID'.",
		},
		{
			name: "Invalid format",
			setupFlags: func() {
				shared.FlagState.MemUnit = "B"
				shared.FlagState.TimeUnit = "ns"
				shared.FlagState.NumberUnit = ""
				shared.FlagState.Format = "invalid"
			},
			expectedMemUnit:   "B",
			expectedTimeUnit:  "ns",
			expectedAllocUnit: "",
			expectedFormat:    "html",
			expectedOutput:    "Warning: Invalid format 'invalid'.",
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

// TestCheckTargetFile tests the file existence check
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

	// Create a valid file
	validPath := filepath.Join(tempDir, "valid.txt")
	err := os.WriteFile(validPath, []byte("content"), 0644)
	require.NoError(t, err)

	nonExistentPath := filepath.Join(tempDir, "nonexistent.txt")

	tests := []struct {
		name          string
		inputPath     string
		expectExit    bool
		expectedError string
	}{
		{
			name:          "Existing file",
			inputPath:     validPath,
			expectExit:    false,
			expectedError: "",
		},
		{
			name:          "Non-existent file",
			inputPath:     nonExistentPath,
			expectExit:    true,
			expectedError: "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the exitCalled flag for each test
			exitCalled = false

			// Call the function and handle any os.Exit panics
			if tt.expectExit {
				// Use WithSafeStderr to execute and capture output
				output, err := shared.WithSafeStderr("checkTargetFile", func() {
					checkTargetFile(tt.inputPath)
				})

				// Check assertions
				if err == nil {
					t.Error("Expected shared.OsExit to be called")
				}
				assert.True(t, exitCalled, "shared.OsExit should have been called")
				assert.Contains(t, output, tt.expectedError)
			} else {
				// Prepare stderr capture for non-exit case
				oldStderr := os.Stderr
				_, w, _ := os.Pipe()
				os.Stderr = w

				checkTargetFile(tt.inputPath)

				// Close write end of pipe
				w.Close()
				os.Stderr = oldStderr

				assert.False(t, exitCalled, "shared.OsExit should not have been called")
			}
		})
	}
}

func TestConvertToBenchmark(t *testing.T) {
	tempDir := t.TempDir()

	// Mock shared.OsExit
	originalOsExit := shared.OsExit
	defer func() { shared.OsExit = originalOsExit }()

	shared.OsExit = func(code int) {
		panic(fmt.Sprintf("shared.OsExit(%d) called", code))
	}

	t.Run("Valid benchmark JSON", func(t *testing.T) {
		validFile := filepath.Join(tempDir, "valid_bench.json")
		bench := shared.Benchmark{
			Name: "Test Benchmark",
			Data: []shared.BenchmarkData{
				{Name: "Bench1", Stats: []shared.Stat{{Type: "time", Value: 100}}},
			},
		}
		data, err := json.Marshal(bench)
		require.NoError(t, err)
		err = os.WriteFile(validFile, data, 0644)
		require.NoError(t, err)

		result := convertToBenchmark(validFile)
		require.NotNil(t, result)
		assert.Equal(t, "Test Benchmark", result.Name)
		assert.Len(t, result.Data, 1)
	})

	t.Run("Invalid JSON content", func(t *testing.T) {
		invalidFile := filepath.Join(tempDir, "invalid.json")
		err := os.WriteFile(invalidFile, []byte("invalid json"), 0644)
		require.NoError(t, err)

		result := convertToBenchmark(invalidFile)
		assert.Nil(t, result)
	})

	t.Run("File read error", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "does_not_exist.json")

		assert.Panics(t, func() {
			convertToBenchmark(nonExistentFile)
		})
	})
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
				// Capture both stdout and stderr
				oldStdout := os.Stdout
				oldStderr := os.Stderr
				r, w, _ := os.Pipe()
				os.Stdout = w
				os.Stderr = w

				// Use WithSafe to handle panic
				err := shared.WithSafe("runBenchmark", func() {
					runBenchmark(cmd, args)
				})

				// Close write end and read output
				w.Close()
				var buf bytes.Buffer
				io.Copy(&buf, r)

				// Restore stdout and stderr
				os.Stdout = oldStdout
				os.Stderr = oldStderr

				// Verify assertions
				if err == nil {
					t.Error("Expected shared.OsExit to be called")
				}
				assert.True(t, exitCalled, "shared.OsExit should have been called")
				assert.Contains(t, buf.String(), tt.expectedOutput)
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
