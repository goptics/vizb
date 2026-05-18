package shared

import (
	"fmt"
	"os"
	"path/filepath"
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
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		ExitWithError("Error creating parent directories", err)
	}

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
