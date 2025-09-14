package shared

import (
	"fmt"
	"os"
)

func ExitWithError(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %s: %v\n", msg, err)
	} else {
		fmt.Fprintf(os.Stderr, "❌ %s\n", msg)
	}

	TempFiles.RemoveAll()
	OsExit(1)
}
