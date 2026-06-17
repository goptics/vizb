package shared

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type OsSuite struct {
	suite.Suite
	originalOsExit       func(int)
	originalOsTempCreate func(string, string) (*os.File, error)
}

func (s *OsSuite) SetupTest() {
	s.originalOsExit = OsExit
	s.originalOsTempCreate = OsTempCreate
}

func (s *OsSuite) TearDownTest() {
	OsExit = s.originalOsExit
	OsTempCreate = s.originalOsTempCreate
}

func (s *OsSuite) TestMustCreateTempFile() {
	s.Run("Successful temp file creation", func() {
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
			s.Run(tt.name, func() {
				filename := MustCreateTempFile(tt.prefix, tt.extension)

				s.NotEmpty(filename, "Filename should not be empty")

				_, err := os.Stat(filename)
				s.NoError(err, "File should exist")

				basename := filepath.Base(filename)
				if tt.prefix != "" {
					s.True(strings.HasPrefix(basename, tt.prefix),
						"Filename should start with prefix: %s, got: %s", tt.prefix, basename)
				}

				if tt.extension != "" {
					s.True(strings.HasSuffix(basename, "."+tt.extension),
						"Filename should end with extension: .%s, got: %s", tt.extension, basename)
				}

				os.Remove(filename)
			})
		}
	})

	s.Run("Temp file creation failure", func() {
		OsTempCreate = func(dir, pattern string) (*os.File, error) {
			return nil, fmt.Errorf("permission denied")
		}
		defer func() { OsTempCreate = s.originalOsTempCreate }()

		restore, exitCalled := TrapOsExitPanic(s.T())
		defer restore()

		s.Panics(func() {
			MustCreateTempFile("test", "txt")
		}, "Function should panic when temp file creation fails")

		s.True(*exitCalled, "OsExit should be called on error")
	})

	s.Run("File handle management", func() {
		filename := MustCreateTempFile("handle-test", "tmp")
		defer os.Remove(filename)

		file, err := os.OpenFile(filename, os.O_RDWR, 0666)
		s.Require().NoError(err, "Should be able to open the created temp file")

		_, err = file.WriteString("test content")
		s.NoError(err, "Should be able to write to temp file")

		file.Close()
	})

	s.Run("Multiple temp files uniqueness", func() {
		const numFiles = 10
		filenames := make([]string, numFiles)

		for i := 0; i < numFiles; i++ {
			filenames[i] = MustCreateTempFile("unique", "test")
		}

		defer func() {
			for _, filename := range filenames {
				os.Remove(filename)
			}
		}()

		filenameSet := make(map[string]bool)
		for _, filename := range filenames {
			s.False(filenameSet[filename], "All filenames should be unique, found duplicate: %s", filename)
			filenameSet[filename] = true
		}
	})

	s.Run("Create temp file with prefix and extension", func() {
		tempFile := MustCreateTempFile("test-prefix", "json")
		defer os.Remove(tempFile)

		s.FileExists(tempFile)
		basename := filepath.Base(tempFile)
		s.Contains(basename, "test-prefix")
		s.Contains(basename, ".json")
	})

	s.Run("Create temp file with empty prefix", func() {
		tempFile := MustCreateTempFile("", "txt")
		defer os.Remove(tempFile)

		s.FileExists(tempFile)
		s.Contains(filepath.Base(tempFile), ".txt")
	})

	s.Run("Create temp file with empty extension", func() {
		tempFile := MustCreateTempFile("test", "")
		defer os.Remove(tempFile)

		s.FileExists(tempFile)
		s.Contains(filepath.Base(tempFile), "test")
	})

	s.Run("Special characters in prefix and extension", func() {
		tempFile := MustCreateTempFile("test-with_chars.123", "ext.json")
		defer os.Remove(tempFile)

		s.FileExists(tempFile)
		s.Contains(filepath.Base(tempFile), "test-with_chars.123")
	})
}

func (s *OsSuite) TestMustCreateFile() {
	s.Run("Successful file creation", func() {
		tempDir := s.T().TempDir()

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
			s.Run(tt.name, func() {
				parentDir := filepath.Dir(tt.filename)
				os.MkdirAll(parentDir, 0755)

				file := MustCreateFile(tt.filename)
				s.NotNil(file, "File handle should not be nil")

				_, err := os.Stat(tt.filename)
				s.NoError(err, "File should exist")

				_, err = file.WriteString("test content")
				s.NoError(err, "Should be able to write to file")

				file.Close()
				os.Remove(tt.filename)
			})
		}
	})

	s.Run("File creation failure scenarios", func() {
		tests := []struct {
			name        string
			setupFunc   func() (string, func())
			expectPanic bool
		}{
			{
				name: "Invalid directory path",
				setupFunc: func() (string, func()) {
					return "/non/existent/directory/test.txt", func() {}
				},
				expectPanic: true,
			},
			{
				name: "Permission denied",
				setupFunc: func() (string, func()) {
					tempDir := s.T().TempDir()
					readOnlyDir := filepath.Join(tempDir, "readonly")
					os.Mkdir(readOnlyDir, 0400)
					return filepath.Join(readOnlyDir, "test.txt"), func() {
						os.Chmod(readOnlyDir, 0755)
					}
				},
				expectPanic: true,
			},
		}

		for _, tt := range tests {
			s.Run(tt.name, func() {
				filepath, cleanup := tt.setupFunc()
				defer cleanup()

				restore, exitCalled := TrapOsExitPanic(s.T())
				defer restore()

				if tt.expectPanic {
					s.Panics(func() {
						MustCreateFile(filepath)
					}, "Function should panic when file creation fails")
					s.True(*exitCalled, "OsExit should be called on error")
				} else {
					file := MustCreateFile(filepath)
					s.NotNil(file, "File should be created successfully")
					file.Close()
					os.Remove(filepath)
				}
			})
		}
	})

	s.Run("Overwrite existing file", func() {
		tempDir := s.T().TempDir()
		filename := filepath.Join(tempDir, "existing.txt")

		initialFile, err := os.Create(filename)
		s.Require().NoError(err)
		initialFile.WriteString("initial content")
		initialFile.Close()

		newFile := MustCreateFile(filename)
		s.NotNil(newFile, "Should create new file handle")

		_, err = newFile.WriteString("new content")
		s.NoError(err)
		newFile.Close()

		content, err := os.ReadFile(filename)
		s.Require().NoError(err)
		s.Equal("new content", string(content), "File content should be overwritten")
	})

	s.Run("File handle properties", func() {
		tempDir := s.T().TempDir()
		filename := filepath.Join(tempDir, "properties.txt")

		file := MustCreateFile(filename)
		defer file.Close()

		n, err := file.Write([]byte("test"))
		s.NoError(err)
		s.Equal(4, n, "Should write 4 bytes")

		pos, err := file.Seek(0, 1)
		s.NoError(err)
		s.Equal(int64(4), pos, "File position should be at 4")

		pos, err = file.Seek(0, 0)
		s.NoError(err)
		s.Equal(int64(0), pos, "Should be able to seek to beginning")
	})
}

func (s *OsSuite) TestOsVariableAssignment() {
	s.NotNil(OsExit, "OsExit should be assigned")
	s.NotNil(OsTempCreate, "OsTempCreate should be assigned")

	mockTempCreateCalled := false
	OsExit = func(code int) { /* mock exit */ }
	OsTempCreate = func(dir, pattern string) (*os.File, error) {
		mockTempCreateCalled = true
		return s.originalOsTempCreate(dir, pattern)
	}

	filename := MustCreateTempFile("mock-test", "tmp")
	defer os.Remove(filename)

	s.True(mockTempCreateCalled, "Mock OsTempCreate should be called")
}

func (s *OsSuite) TestEdgeCases() {
	s.Run("Empty strings", func() {
		filename := MustCreateTempFile("", "")
		defer os.Remove(filename)
		s.NotEmpty(filename, "Should create temp file even with empty prefix and extension")

		_, err := os.Stat(filename)
		s.NoError(err, "File should exist")
	})

	s.Run("Very long strings", func() {
		tempDir := s.T().TempDir()

		longPrefix := strings.Repeat("a", 50)
		filename := MustCreateTempFile(longPrefix, "txt")
		defer os.Remove(filename)

		basename := filepath.Base(filename)
		s.True(strings.HasPrefix(basename, longPrefix), "Should handle long prefix")

		longFilename := filepath.Join(tempDir, strings.Repeat("b", 100)+".txt")
		file := MustCreateFile(longFilename)
		file.Close()
		defer os.Remove(longFilename)

		_, err := os.Stat(longFilename)
		s.NoError(err, "Should create file with long name")
	})
}

func TestOsSuite(t *testing.T) {
	suite.Run(t, new(OsSuite))
}
