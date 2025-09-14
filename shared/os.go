package shared

import (
	"fmt"
	"os"
)

var OsExit = os.Exit
var OsTempCreate = os.CreateTemp

func MustCreateTempFile(prefix, extension string) string {
	temp, err := OsTempCreate("", fmt.Sprintf("%s-*.%s", prefix, extension))

	if err != nil {
		ExitWithError("Error creating temporary file", err)
	}

	defer temp.Close()

	return temp.Name()
}

func MustCreateFile(filePath string) *os.File {
	f, err := os.Create(filePath)
	if err != nil {
		ExitWithError("Error creating file", err)
	}

	return f
}

func MustOpenFile(filePath string) *os.File {
	f, err := os.Open(filePath)
	if err != nil {
		ExitWithError("Error opening file", err)
	}

	return f
}
