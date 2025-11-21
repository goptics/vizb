package shared

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMain sets up and tears down test environment
func TestMain(m *testing.M) {
	// Save original OsExit
	originalOsExit := OsExit

	// Replace OsExit with a test version that doesn't actually exit
	OsExit = func(code int) {
		panic(fmt.Sprintf("OsExit(%d) was called", code))
	}

	// Run tests
	code := m.Run()

	// Restore original OsExit
	OsExit = originalOsExit

	// Exit with the test result code
	originalOsExit(code)
}

func TestExitWithError(t *testing.T) {
	tests := []struct {
		name           string
		msg            string
		err            error
		expectedOutput string
		expectedCode   int
	}{
		{
			name:           "Error with message and error",
			msg:            "Failed to process file",
			err:            errors.New("file not found"),
			expectedOutput: "Failed to process file: file not found\n",
			expectedCode:   1,
		},
		{
			name:           "Error with message only",
			msg:            "Invalid configuration",
			err:            nil,
			expectedOutput: "Invalid configuration\n",
			expectedCode:   1,
		},
		{
			name:           "Empty message with error",
			msg:            "",
			err:            errors.New("unexpected error"),
			expectedOutput: ": unexpected error\n",
			expectedCode:   1,
		},
		{
			name:           "Empty message and nil error",
			msg:            "",
			err:            nil,
			expectedOutput: "\n",
			expectedCode:   1,
		},
		{
			name:           "Long message with complex error",
			msg:            "Failed to initialize benchmark processing pipeline",
			err:            fmt.Errorf("wrapped error: %w", errors.New("original cause")),
			expectedOutput: "Failed to initialize benchmark processing pipeline: wrapped error: original cause\n",
			expectedCode:   1,
		},
		{
			name:           "Message with special characters",
			msg:            "Error processing file 'test.json'",
			err:            errors.New("permission denied"),
			expectedOutput: "Error processing file 'test.json': permission denied\n",
			expectedCode:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Track if OsExit was called with correct code
			exitCalled := false
			exitCode := -1

			// Override OsExit for this test
			originalOsExit := OsExit
			OsExit = func(code int) {
				exitCalled = true
				exitCode = code
				panic(fmt.Sprintf("OsExit(%d) was called", code))
			}

			// Execute the function and catch the panic
			func() {
				defer func() {
					recovered := recover()

					// Close write end and read output
					w.Close()
					var buf bytes.Buffer
					io.Copy(&buf, r)

					// Restore stderr and OsExit
					os.Stderr = oldStderr
					OsExit = originalOsExit

					// Verify OsExit was called
					if recovered == nil {
						t.Error("Expected OsExit to be called (should panic)")
					}
					assert.True(t, exitCalled, "OsExit should have been called")
					assert.Equal(t, tt.expectedCode, exitCode, "OsExit should be called with code 1")

					// Verify stderr output
					actualOutput := buf.String()
					assert.Equal(t, tt.expectedOutput, actualOutput, "stderr output should match expected")
				}()

				ExitWithError(tt.msg, tt.err)
			}()
		})
	}
}

// TestExitWithErrorThreadSafety tests concurrent calls to ExitWithError
func TestExitWithErrorConcurrent(t *testing.T) {
	// This test ensures ExitWithError works correctly under concurrent access
	originalOsExit := OsExit
	defer func() { OsExit = originalOsExit }()

	// Track calls
	calls := make(chan int, 10)
	OsExit = func(code int) {
		calls <- code
		panic(fmt.Sprintf("OsExit(%d) was called", code))
	}

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Run multiple goroutines
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() {
				recover() // Catch the panic from OsExit
				done <- true
			}()
			ExitWithError(fmt.Sprintf("Error from goroutine %d", id), errors.New("test error"))
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Clean up
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stderr = oldStderr

	// Verify we got some calls (race condition means we might not get exactly 5)
	close(calls)
	callCount := 0
	for range calls {
		callCount++
	}
	assert.Greater(t, callCount, 0, "Should have received at least one OsExit call")

	// Verify stderr contains some output
	output := buf.String()
	assert.NotEmpty(t, output, "stderr should contain error messages")
}

// TestExitWithErrorStderrFailure tests behavior when stderr write fails
func TestExitWithErrorStderrFailure(t *testing.T) {
	originalOsExit := OsExit
	defer func() { OsExit = originalOsExit }()

	// Track OsExit calls
	exitCalled := false
	OsExit = func(code int) {
		exitCalled = true
		panic(fmt.Sprintf("OsExit(%d) was called", code))
	}

	// Replace stderr with a closed pipe to cause write failure
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	w.Close() // Close write end immediately to cause write error
	os.Stderr = w

	defer func() {
		recovered := recover()
		r.Close()
		os.Stderr = oldStderr

		// Even if stderr write fails, OsExit should still be called
		if recovered == nil {
			t.Error("Expected OsExit to be called (should panic)")
		}
		assert.True(t, exitCalled, "OsExit should be called even if stderr write fails")
	}()

	ExitWithError("Test message", errors.New("test error"))
}
