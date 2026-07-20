package shared

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ExitWithErrorSuite struct {
	suite.Suite
}

func (s *ExitWithErrorSuite) TestExitWithError() {
	tests := []struct {
		name           string
		msg            string
		err            error
		expectedOutput string
	}{
		{
			name:           "Error with message and error",
			msg:            "Failed to process file",
			err:            errors.New("file not found"),
			expectedOutput: "Failed to process file: file not found\n",
		},
		{
			name:           "Error with message only",
			msg:            "Invalid configuration",
			err:            nil,
			expectedOutput: "Invalid configuration\n",
		},
		{
			name:           "Empty message with error",
			msg:            "",
			err:            errors.New("unexpected error"),
			expectedOutput: ": unexpected error\n",
		},
		{
			name:           "Empty message and nil error",
			msg:            "",
			err:            nil,
			expectedOutput: "\n",
		},
		{
			name:           "Long message with complex error",
			msg:            "Failed to initialize benchmark processing pipeline",
			err:            fmt.Errorf("wrapped error: %w", errors.New("original cause")),
			expectedOutput: "Failed to initialize benchmark processing pipeline: wrapped error: original cause\n",
		},
		{
			name:           "Message with special characters",
			msg:            "Error processing file 'test.json'",
			err:            errors.New("permission denied"),
			expectedOutput: "Error processing file 'test.json': permission denied\n",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			restore, exitCalled := TrapOsExitPanic(s.T())
			defer restore()

			output, err := WithSafeStderr("ExitWithError", func() {
				ExitWithError(tt.msg, tt.err)
			})

			s.Error(err, "Expected OsExit to be called (should panic)")
			s.True(*exitCalled, "OsExit should have been called")
			s.Equal(tt.expectedOutput, output, "stderr output should match expected")
		})
	}
}

func (s *ExitWithErrorSuite) TestExitWithErrorConcurrent() {
	origOsExit := OsExit
	exitCalls := make(chan struct{}, 5)
	OsExit = func(int) {
		exitCalls <- struct{}{}
		panic("exit")
	}
	defer func() { OsExit = origOsExit }()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	done := make(chan bool, 5)
	for i := range 5 {
		go func(id int) {
			_ = WithSafe(fmt.Sprintf("goroutine-%d", id), func() {
				ExitWithError(fmt.Sprintf("Error from goroutine %d", id), errors.New("test error"))
			})
			done <- true
		}(i)
	}

	for range 5 {
		<-done
	}

	w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stderr = oldStderr

	s.NotEmpty(exitCalls, "Should have received at least one OsExit call")
	s.NotEmpty(buf.String(), "stderr should contain error messages")
}

func (s *ExitWithErrorSuite) TestExitWithErrorStderrFailure() {
	restore, exitCalled := TrapOsExitPanic(s.T())
	defer restore()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	w.Close()
	os.Stderr = w

	err := WithSafe("ExitWithError", func() {
		ExitWithError("Test message", errors.New("test error"))
	})

	r.Close()
	os.Stderr = oldStderr

	s.Error(err, "Expected OsExit to be called (should panic)")
	s.True(*exitCalled, "OsExit should be called even if stderr write fails")
}

func (s *ExitWithErrorSuite) TestPrintWarning() {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	PrintWarning("disk nearly full")

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	s.Contains(buf.String(), "disk nearly full")
}

func TestExitWithErrorSuite(t *testing.T) {
	suite.Run(t, new(ExitWithErrorSuite))
}
