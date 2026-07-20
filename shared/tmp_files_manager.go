package shared

import (
	"os"
	"sync"
)

type TmpFilesManager struct {
	mu    sync.Mutex
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
	tfm.mu.Lock()
	defer tfm.mu.Unlock()
	tfm.files = append(tfm.files, args...)
}

func (tfm *TmpFilesManager) RemoveAll() {
	tfm.mu.Lock()
	defer tfm.mu.Unlock()

	for _, filePath := range tfm.files {
		os.Remove(filePath)
	}
	tfm.files = make([]string, 0)
}
