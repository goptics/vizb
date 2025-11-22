package shared

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// WithSafe executes a function and recovers from any panic, returning it as an error.
// This is useful for testing functions that may panic or call os.Exit.
func WithSafe(name string, fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered inside %s: %v", name, r)
		}
	}()

	fn()

	return err
}

// WithSafeStderr executes a function that may panic, capturing its stderr output.
// Returns the captured output and any error from panic recovery.
// This is useful for testing functions that write to stderr and may call os.Exit.
func WithSafeStderr(name string, fn func()) (output string, err error) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	defer func() {
		// Close write end
		w.Close()

		// Read captured output
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output = buf.String()

		// Restore stderr
		os.Stderr = oldStderr

		// Recover from panic
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("panic recovered inside %s: %v", name, recovered)
		}
	}()

	fn()

	return output, err
}
