package shared

import (
	"fmt"

	"github.com/goptics/vizb/pkg/cliout"
)

// ExitWithError prints an error message to stderr and exits the program with status code 1.
// If err is not nil, it prints both the message and the error details.
// If err is nil, only the message is printed.
//
// Does not use log.Fatal so temp-file cleanup always runs before OsExit.
func ExitWithError(msg string, err error) {
	if err != nil {
		cliout.Error(fmt.Sprintf("%s: %v", msg, err))
	} else {
		cliout.Error(msg)
	}

	TempFiles.RemoveAll()
	OsExit(1)
}
