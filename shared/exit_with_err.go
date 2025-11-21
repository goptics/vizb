package shared

import (
	"fmt"
	"os"

	"github.com/goptics/vizb/pkg/style"
)

// ExitWithError prints an error message to stderr and exits the program with status code 1.
// If err is not nil, it prints both the message and the error details.
// If err is nil, only the message is printed.
func ExitWithError(msg string, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, style.Error.Render(fmt.Sprintf("%s: %v", msg, err)))
	} else {
		fmt.Fprintln(os.Stderr, style.Error.Render(fmt.Sprintf("%s", msg)))
	}

	TempFiles.RemoveAll()
	OsExit(1)
}

// PrintWarning prints a warning message to stderr.
func PrintWarning(msg string) {
	fmt.Fprintln(os.Stderr, style.Warning.Render(msg))
}
