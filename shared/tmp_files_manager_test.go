package shared

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTmpFileManager(t *testing.T) {
	manager := NewTmpFileManager()

	if len(manager.files) != 0 {
		t.Errorf("NewTmpFileManager() should create empty files slice, got %d files", len(manager.files))
	}
}

func TestTmpFilesManager_Store(t *testing.T) {
	manager := NewTmpFileManager()

	// Test storing single file
	manager.Store("file1.txt")
	if len(manager.files) != 1 {
		t.Errorf("Expected 1 file after Store, got %d", len(manager.files))
	}
	if manager.files[0] != "file1.txt" {
		t.Errorf("Expected file1.txt, got %s", manager.files[0])
	}

	// Test storing multiple files at once
	manager.Store("file2.txt", "file3.txt")
	if len(manager.files) != 3 {
		t.Errorf("Expected 3 files after storing multiple, got %d", len(manager.files))
	}

	expected := []string{"file1.txt", "file2.txt", "file3.txt"}
	for i, expected := range expected {
		if manager.files[i] != expected {
			t.Errorf("Expected %s at index %d, got %s", expected, i, manager.files[i])
		}
	}
}

func TestTmpFilesManager_RemoveAll(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tempDir, "test1.txt")
	file2 := filepath.Join(tempDir, "test2.txt")
	file3 := filepath.Join(tempDir, "test3.txt")

	// Create actual files
	for _, file := range []string{file1, file2, file3} {
		f, err := os.Create(file)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
		f.Close()
	}

	// Verify files exist
	for _, file := range []string{file1, file2, file3} {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Fatalf("Test file %s should exist", file)
		}
	}

	// Store files in manager
	manager := NewTmpFileManager()
	manager.Store(file1, file2, file3)

	// Remove all files
	manager.RemoveAll()

	// Verify files are deleted
	for _, file := range []string{file1, file2, file3} {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Errorf("File %s should have been deleted", file)
		}
	}

	// Verify manager files slice is reset
	if len(manager.files) != 0 {
		t.Errorf("Manager files slice should be empty after RemoveAll, got %d files", len(manager.files))
	}
}

func TestTmpFilesManager_RemoveAll_NonExistentFiles(t *testing.T) {
	manager := NewTmpFileManager()
	manager.Store("nonexistent1.txt", "nonexistent2.txt")

	// Should not panic when removing non-existent files
	manager.RemoveAll()

	// Verify manager files slice is reset
	if len(manager.files) != 0 {
		t.Errorf("Manager files slice should be empty after RemoveAll, got %d files", len(manager.files))
	}
}

func TestGlobalTempFiles(t *testing.T) {
	// Save original state
	originalFiles := make([]string, len(TempFiles.files))
	copy(originalFiles, TempFiles.files)

	// Clean up after test
	defer func() {
		TempFiles.files = originalFiles
	}()

	// Test global TempFiles variable
	initialLen := len(TempFiles.files)
	TempFiles.Store("global_test.txt")

	if len(TempFiles.files) != initialLen+1 {
		t.Errorf("Global TempFiles should have %d files, got %d", initialLen+1, len(TempFiles.files))
	}

	TempFiles.RemoveAll()

	if len(TempFiles.files) != 0 {
		t.Errorf("Global TempFiles should be empty after RemoveAll, got %d files", len(TempFiles.files))
	}
}

func TestTmpFilesManager_EmptyStore(t *testing.T) {
	manager := NewTmpFileManager()

	// Test storing no files
	manager.Store()

	if len(manager.files) != 0 {
		t.Errorf("Expected 0 files after Store(), got %d", len(manager.files))
	}
}

func TestTmpFilesManager_Integration(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create manager and test full workflow
	manager := NewTmpFileManager()

	// Create and store multiple files
	files := make([]string, 3)
	for i := 0; i < 3; i++ {
		file := filepath.Join(tempDir, "integration_test_"+string(rune('1'+i))+".txt")
		f, err := os.Create(file)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		f.Close()
		files[i] = file
	}

	// Store files one by one and in batch
	manager.Store(files[0])
	manager.Store(files[1], files[2])

	if len(manager.files) != 3 {
		t.Errorf("Expected 3 files stored, got %d", len(manager.files))
	}

	// Verify all files exist
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("File %s should exist before RemoveAll", file)
		}
	}

	// Remove all
	manager.RemoveAll()

	// Verify all files are deleted
	for _, file := range files {
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Errorf("File %s should be deleted after RemoveAll", file)
		}
	}

	// Verify manager is reset
	if len(manager.files) != 0 {
		t.Errorf("Manager should be empty after RemoveAll, got %d files", len(manager.files))
	}
}
