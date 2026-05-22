package shared

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTmpFileManager(t *testing.T) {
	manager := NewTmpFileManager()
	assert.Empty(t, manager.files)
}

func TestTmpFilesManager_Store(t *testing.T) {
	manager := NewTmpFileManager()

	manager.Store("file1.txt")
	assert.Len(t, manager.files, 1)
	assert.Equal(t, "file1.txt", manager.files[0])

	manager.Store("file2.txt", "file3.txt")
	assert.Len(t, manager.files, 3)
	assert.Equal(t, []string{"file1.txt", "file2.txt", "file3.txt"}, manager.files)
}

func TestTmpFilesManager_RemoveAll(t *testing.T) {
	tempDir := t.TempDir()

	file1 := filepath.Join(tempDir, "test1.txt")
	file2 := filepath.Join(tempDir, "test2.txt")
	file3 := filepath.Join(tempDir, "test3.txt")

	for _, file := range []string{file1, file2, file3} {
		f, err := os.Create(file)
		require.NoError(t, err, "Failed to create test file")
		f.Close()
	}

	for _, file := range []string{file1, file2, file3} {
		_, err := os.Stat(file)
		require.False(t, os.IsNotExist(err), "Test file should exist")
	}

	manager := NewTmpFileManager()
	manager.Store(file1, file2, file3)

	manager.RemoveAll()

	for _, file := range []string{file1, file2, file3} {
		_, err := os.Stat(file)
		assert.True(t, os.IsNotExist(err), "File should have been deleted")
	}

	assert.Empty(t, manager.files)
}

func TestTmpFilesManager_RemoveAll_NonExistentFiles(t *testing.T) {
	manager := NewTmpFileManager()
	manager.Store("nonexistent1.txt", "nonexistent2.txt")

	manager.RemoveAll()

	assert.Empty(t, manager.files)
}

func TestGlobalTempFiles(t *testing.T) {
	originalFiles := make([]string, len(TempFiles.files))
	copy(originalFiles, TempFiles.files)

	defer func() {
		TempFiles.files = originalFiles
	}()

	initialLen := len(TempFiles.files)
	TempFiles.Store("global_test.txt")

	assert.Len(t, TempFiles.files, initialLen+1)

	TempFiles.RemoveAll()

	assert.Empty(t, TempFiles.files)
}

func TestTmpFilesManager_EmptyStore(t *testing.T) {
	manager := NewTmpFileManager()

	manager.Store()

	assert.Empty(t, manager.files)
}

func TestTmpFilesManager_Integration(t *testing.T) {
	tempDir := t.TempDir()

	manager := NewTmpFileManager()

	files := make([]string, 3)
	for i := 0; i < 3; i++ {
		file := filepath.Join(tempDir, "integration_test_"+string(rune('1'+i))+".txt")
		f, err := os.Create(file)
		require.NoError(t, err, "Failed to create test file")
		f.Close()
		files[i] = file
	}

	manager.Store(files[0])
	manager.Store(files[1], files[2])

	assert.Len(t, manager.files, 3)

	for _, file := range files {
		_, err := os.Stat(file)
		assert.False(t, os.IsNotExist(err), "File should exist before RemoveAll")
	}

	manager.RemoveAll()

	for _, file := range files {
		_, err := os.Stat(file)
		assert.True(t, os.IsNotExist(err), "File should be deleted after RemoveAll")
	}

	assert.Empty(t, manager.files)
}
