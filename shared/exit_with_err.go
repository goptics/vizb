package shared

import (
	"fmt"
	"os"
)

// ExitWithError prints an error message to stderr and exits the program with status code 1.
// If err is not nil, it prints both the message and the error details.
// If err is nil, only the message is printed.
func ExitWithError(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %s: %v\n", msg, err)
	} else {
		fmt.Fprintf(os.Stderr, "❌ %s\n", msg)
	}

	TempFiles.RemoveAll()
	OsExit(1)
}
