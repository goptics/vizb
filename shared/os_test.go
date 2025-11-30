package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMustCreateTempFile(t *testing.T) {
	// Save original functions
	originalOsTempCreate := OsTempCreate
	originalOsExit := OsExit

	defer func() {
		OsTempCreate = originalOsTempCreate
		OsExit = originalOsExit
	}()

	t.Run("Successful temp file creation", func(t *testing.T) {
		tests := []struct {
			name      string
			prefix    string
			extension string
		}{
			{
				name:      "Standard prefix and extension",
				prefix:    "test",
				extension: "txt",
			},
			{
				name:      "Empty prefix",
				prefix:    "",
				extension: "json",
			},
			{
				name:      "Empty extension",
				prefix:    "benchmark",
				extension: "",
			},
			{
				name:      "Long prefix and extension",
				prefix:    "very-long-benchmark-prefix",
				extension: "html",
			},
			{
				name:      "Special characters in prefix",
				prefix:    "test-file_123",
				extension: "log",
			},
			{
				name:      "Numeric extension",
				prefix:    "backup",
				extension: "001",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Call the function
				filename := MustCreateTempFile(tt.prefix, tt.extension)

				// Verify file was created
				assert.NotEmpty(t, filename, "Filename should not be empty")

				// Check file exists
				_, err := os.Stat(filename)
				assert.NoError(t, err, "File should exist")

				// Verify filename pattern
				_ = fmt.Sprintf("%s*", tt.prefix) // Pattern for reference

				basename := filepath.Base(filename)
				if tt.prefix != "" {
					assert.True(t, strings.HasPrefix(basename, tt.prefix),
						"Filename should start with prefix: %s, got: %s", tt.prefix, basename)
				}

				if tt.extension != "" {
					assert.True(t, strings.HasSuffix(basename, "."+tt.extension),
						"Filename should end with extension: .%s, got: %s", tt.extension, basename)
				}

				// Clean up
				os.Remove(filename)
			})
		}
	})

	t.Run("Temp file creation failure", func(t *testing.T) {
		// Mock OsTempCreate to return an error
		OsTempCreate = func(dir, pattern string) (*os.File, error) {
			return nil, fmt.Errorf("permission denied")
		}

		// Mock OsExit to track calls
		exitCalled := false
		exitCode := -1
		OsExit = func(code int) {
			exitCalled = true
			exitCode = code
			panic(fmt.Sprintf("OsExit(%d) was called", code))
		}

		// Test function should panic due to OsExit
		assert.Panics(t, func() {
			MustCreateTempFile("test", "txt")
		}, "Function should panic when temp file creation fails")

		// Verify OsExit was called
		assert.True(t, exitCalled, "OsExit should be called on error")
		assert.Equal(t, 1, exitCode, "Should exit with code 1")
	})

	t.Run("File handle management", func(t *testing.T) {
		// Ensure we use the original functions for this test
		OsTempCreate = originalOsTempCreate
		OsExit = originalOsExit

		// Test that the file handle is properly closed
		filename := MustCreateTempFile("handle-test", "tmp")
		defer os.Remove(filename)

		// Try to open the file to verify it's not locked
		file, err := os.OpenFile(filename, os.O_RDWR, 0666)
		require.NoError(t, err, "Should be able to open the created temp file")

		// Write to it to verify it's accessible
		_, err = file.WriteString("test content")
		assert.NoError(t, err, "Should be able to write to temp file")

		file.Close()
	})

	t.Run("Multiple temp files uniqueness", func(t *testing.T) {
		// Ensure we use the original functions for this test
		OsTempCreate = originalOsTempCreate
		OsExit = originalOsExit

		// Create multiple temp files and ensure they're unique
		const numFiles = 10
		filenames := make([]string, numFiles)

		for i := 0; i < numFiles; i++ {
			filenames[i] = MustCreateTempFile("unique", "test")
		}

		// Clean up
		defer func() {
			for _, filename := range filenames {
				os.Remove(filename)
			}
		}()

		// Verify all filenames are unique
		filenameSet := make(map[string]bool)
		for _, filename := range filenames {
			assert.False(t, filenameSet[filename], "All filenames should be unique, found duplicate: %s", filename)
			filenameSet[filename] = true
		}
	})
}

func TestMustCreateFile(t *testing.T) {
	// Save original functions
	originalOsExit := OsExit
	defer func() { OsExit = originalOsExit }()

	t.Run("Successful file creation", func(t *testing.T) {
		tempDir := t.TempDir()

		tests := []struct {
			name     string
			filename string
		}{
			{
				name:     "Simple filename",
				filename: filepath.Join(tempDir, "test.txt"),
			},
			{
				name:     "Filename with spaces",
				filename: filepath.Join(tempDir, "test file.html"),
			},
			{
				name:     "Filename with special characters",
				filename: filepath.Join(tempDir, "test-file_123.json"),
			},
			{
				name:     "Long filename",
				filename: filepath.Join(tempDir, "very-long-filename-with-many-characters-for-testing.log"),
			},
			{
				name:     "Nested directory",
				filename: filepath.Join(tempDir, "subdir", "nested.txt"),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Create parent directory if needed
				parentDir := filepath.Dir(tt.filename)
				os.MkdirAll(parentDir, 0755)

				// Call the function
				file := MustCreateFile(tt.filename)
				assert.NotNil(t, file, "File handle should not be nil")

				// Verify file was created
				_, err := os.Stat(tt.filename)
				assert.NoError(t, err, "File should exist")

				// Verify file is writable
				_, err = file.WriteString("test content")
				assert.NoError(t, err, "Should be able to write to file")

				// Clean up
				file.Close()
				os.Remove(tt.filename)
			})
		}
	})

	t.Run("File creation failure scenarios", func(t *testing.T) {
		tests := []struct {
			name        string
			setupFunc   func() (string, func()) // returns filepath and cleanup func
			expectPanic bool
		}{
			{
				name: "Invalid directory path",
				setupFunc: func() (string, func()) {
					// Try to create file in non-existent directory without creating it
					return "/non/existent/directory/test.txt", func() {}
				},
				expectPanic: true,
			},
			{
				name: "Permission denied",
				setupFunc: func() (string, func()) {
					// Create a read-only directory
					tempDir := t.TempDir()
					readOnlyDir := filepath.Join(tempDir, "readonly")
					os.Mkdir(readOnlyDir, 0400) // read-only
					return filepath.Join(readOnlyDir, "test.txt"), func() {
						os.Chmod(readOnlyDir, 0755) // restore permissions for cleanup
					}
				},
				expectPanic: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				filepath, cleanup := tt.setupFunc()
				defer cleanup()

				// Mock OsExit to track calls
				exitCalled := false
				OsExit = func(code int) {
					exitCalled = true
					panic(fmt.Sprintf("OsExit(%d) was called", code))
				}

				if tt.expectPanic {
					assert.Panics(t, func() {
						MustCreateFile(filepath)
					}, "Function should panic when file creation fails")
					assert.True(t, exitCalled, "OsExit should be called on error")
				} else {
					file := MustCreateFile(filepath)
					assert.NotNil(t, file, "File should be created successfully")
					file.Close()
					os.Remove(filepath)
				}
			})
		}
	})

	t.Run("Overwrite existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		filename := filepath.Join(tempDir, "existing.txt")

		// Create initial file
		initialFile, err := os.Create(filename)
		require.NoError(t, err)
		initialFile.WriteString("initial content")
		initialFile.Close()

		// Use MustCreateFile to overwrite
		newFile := MustCreateFile(filename)
		assert.NotNil(t, newFile, "Should create new file handle")

		// Write new content
		_, err = newFile.WriteString("new content")
		assert.NoError(t, err)
		newFile.Close()

		// Verify content was overwritten
		content, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.Equal(t, "new content", string(content), "File content should be overwritten")
	})

	t.Run("File handle properties", func(t *testing.T) {
		tempDir := t.TempDir()
		filename := filepath.Join(tempDir, "properties.txt")

		file := MustCreateFile(filename)
		defer file.Close()

		// Verify file is writable
		n, err := file.Write([]byte("test"))
		assert.NoError(t, err)
		assert.Equal(t, 4, n, "Should write 4 bytes")

		// Verify file position
		pos, err := file.Seek(0, 1) // Get current position
		assert.NoError(t, err)
		assert.Equal(t, int64(4), pos, "File position should be at 4")

		// Verify we can seek
		pos, err = file.Seek(0, 0) // Seek to beginning
		assert.NoError(t, err)
		assert.Equal(t, int64(0), pos, "Should be able to seek to beginning")
	})
}

// TestOsVariableAssignment tests that the os function variables are properly assigned
func TestOsVariableAssignment(t *testing.T) {
	// Test that OsExit is assigned to os.Exit by default
	assert.NotNil(t, OsExit, "OsExit should be assigned")

	// Test that OsTempCreate is assigned to os.CreateTemp by default
	assert.NotNil(t, OsTempCreate, "OsTempCreate should be assigned")

	// Test that we can reassign them (for mocking)
	originalOsExit := OsExit
	originalOsTempCreate := OsTempCreate

	OsExit = func(code int) { /* mock exit */ }

	mockTempCreateCalled := false
	OsTempCreate = func(dir, pattern string) (*os.File, error) {
		mockTempCreateCalled = true
		return originalOsTempCreate(dir, pattern)
	}

	// Create a temp file to test the mock
	filename := MustCreateTempFile("mock-test", "tmp")
	defer os.Remove(filename)

	assert.True(t, mockTempCreateCalled, "Mock OsTempCreate should be called")

	// Restore original functions
	OsExit = originalOsExit
	OsTempCreate = originalOsTempCreate
}

// TestEdgeCases tests edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	t.Run("Empty strings", func(t *testing.T) {
		// Test MustCreateTempFile with empty strings
		filename := MustCreateTempFile("", "")
		defer os.Remove(filename)
		assert.NotEmpty(t, filename, "Should create temp file even with empty prefix and extension")

		// Verify file exists
		_, err := os.Stat(filename)
		assert.NoError(t, err, "File should exist")
	})

	t.Run("Very long strings", func(t *testing.T) {
		tempDir := t.TempDir()

		// Test with long prefix (but reasonable for filesystem)
		longPrefix := strings.Repeat("a", 50)
		filename := MustCreateTempFile(longPrefix, "txt")
		defer os.Remove(filename)

		basename := filepath.Base(filename)
		assert.True(t, strings.HasPrefix(basename, longPrefix), "Should handle long prefix")

		// Test MustCreateFile with reasonably long filename
		longFilename := filepath.Join(tempDir, strings.Repeat("b", 100)+".txt")
		file := MustCreateFile(longFilename)
		file.Close()
		defer os.Remove(longFilename)

		_, err := os.Stat(longFilename)
		assert.NoError(t, err, "Should create file with long name")
	})
}
