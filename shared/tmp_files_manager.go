package shared

import (
	"os"
)

type TmpFilesManager struct {
	files []string
}

var TempFiles = NewTmpFileManager()

// NewTmpFileManager creates and returns a new instance of TmpFilesManager
// with an empty files slice ready for storing temporary file paths.
func NewTmpFileManager() TmpFilesManager {
	return TmpFilesManager{
		files: make([]string, 0),
	}
}

func (tfm *TmpFilesManager) Store(args ...string) {
	tfm.files = append(tfm.files, args...)
}

func (tfm *TmpFilesManager) RemoveAll() {
	for _, filePath := range tfm.files {
		os.Remove(filePath)
	}

	tfm.files = make([]string, 0)
}
