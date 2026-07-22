package shared

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TmpFilesManagerSuite struct {
	suite.Suite
}

func (s *TmpFilesManagerSuite) TestNewTmpFileManager() {
	manager := NewTmpFileManager()
	s.Empty(manager.files)
}

func (s *TmpFilesManagerSuite) TestTmpFilesManagerStore() {
	manager := NewTmpFileManager()

	manager.Store("file1.txt")
	s.Len(manager.files, 1)
	s.Equal("file1.txt", manager.files[0])

	manager.Store("file2.txt", "file3.txt")
	s.Len(manager.files, 3)
	s.Equal([]string{"file1.txt", "file2.txt", "file3.txt"}, manager.files)
}

func (s *TmpFilesManagerSuite) TestTmpFilesManagerRemoveAll() {
	tempDir := s.T().TempDir()

	file1 := filepath.Join(tempDir, "test1.txt")
	file2 := filepath.Join(tempDir, "test2.txt")
	file3 := filepath.Join(tempDir, "test3.txt")

	for _, file := range []string{file1, file2, file3} {
		f, err := os.Create(file)
		s.Require().NoError(err, "Failed to create test file")
		f.Close()
	}

	for _, file := range []string{file1, file2, file3} {
		_, err := os.Stat(file)
		s.Require().False(os.IsNotExist(err), "Test file should exist")
	}

	manager := NewTmpFileManager()
	manager.Store(file1, file2, file3)

	manager.RemoveAll()

	for _, file := range []string{file1, file2, file3} {
		_, err := os.Stat(file)
		s.True(os.IsNotExist(err), "File should have been deleted")
	}

	s.Empty(manager.files)
}

func (s *TmpFilesManagerSuite) TestTmpFilesManagerRemoveAllNonExistentFiles() {
	manager := NewTmpFileManager()
	manager.Store("nonexistent1.txt", "nonexistent2.txt")

	manager.RemoveAll()

	s.Empty(manager.files)
}

func (s *TmpFilesManagerSuite) TestTmpFilesManagerRemoveAllDirectory() {
	tempDir := filepath.Join(s.T().TempDir(), "managed")
	s.Require().NoError(os.MkdirAll(filepath.Join(tempDir, "nested"), 0o755))
	s.Require().NoError(os.WriteFile(filepath.Join(tempDir, "nested", "update.tmp"), []byte("temporary"), 0o600))

	manager := NewTmpFileManager()
	manager.Store(tempDir)
	manager.RemoveAll()

	s.NoDirExists(tempDir)
	s.Empty(manager.files)
}

func (s *TmpFilesManagerSuite) TestGlobalTempFiles() {
	originalFiles := make([]string, len(TempFiles.files))
	copy(originalFiles, TempFiles.files)

	defer func() {
		TempFiles.files = originalFiles
	}()

	initialLen := len(TempFiles.files)
	TempFiles.Store("global_test.txt")

	s.Len(TempFiles.files, initialLen+1)

	TempFiles.RemoveAll()

	s.Empty(TempFiles.files)
}

func (s *TmpFilesManagerSuite) TestTmpFilesManagerEmptyStore() {
	manager := NewTmpFileManager()

	manager.Store()

	s.Empty(manager.files)
}

func (s *TmpFilesManagerSuite) TestTmpFilesManagerConcurrentAccess() {
	manager := NewTmpFileManager()
	var wg sync.WaitGroup

	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			manager.Store("nonexistent.txt")
			manager.RemoveAll()
		}()
	}
	wg.Wait()
	manager.RemoveAll()

	s.Empty(manager.files)
}

func (s *TmpFilesManagerSuite) TestTmpFilesManagerIntegration() {
	tempDir := s.T().TempDir()

	manager := NewTmpFileManager()

	files := make([]string, 3)
	for i := 0; i < 3; i++ {
		file := filepath.Join(tempDir, "integration_test_"+string(rune('1'+i))+".txt")
		f, err := os.Create(file)
		s.Require().NoError(err, "Failed to create test file")
		f.Close()
		files[i] = file
	}

	manager.Store(files[0])
	manager.Store(files[1], files[2])

	s.Len(manager.files, 3)

	for _, file := range files {
		_, err := os.Stat(file)
		s.False(os.IsNotExist(err), "File should exist before RemoveAll")
	}

	manager.RemoveAll()

	for _, file := range files {
		_, err := os.Stat(file)
		s.True(os.IsNotExist(err), "File should be deleted after RemoveAll")
	}

	s.Empty(manager.files)
}

func TestTmpFilesManagerSuite(t *testing.T) {
	suite.Run(t, new(TmpFilesManagerSuite))
}
