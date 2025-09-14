package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResolveOutputFileNameComprehensive tests the resolveOutputFileName function comprehensively
func TestResolveOutputFileNameComprehensive(t *testing.T) {
	tests := []struct {
		name           string
		outFile        string
		format         string
		expectedSuffix string
		expectTempFile bool
	}{
		{
			name:           "Empty outFile creates temp file for html",
			outFile:        "",
			format:         "html",
			expectedSuffix: ".html",
			expectTempFile: true,
		},
		{
			name:           "Empty outFile creates temp file for json",
			outFile:        "",
			format:         "json",
			expectedSuffix: ".json",
			expectTempFile: true,
		},
		{
			name:           "File with correct extension unchanged",
			outFile:        "output.html",
			format:         "html",
			expectedSuffix: ".html",
			expectTempFile: false,
		},
		{
			name:           "File without extension gets extension added",
			outFile:        "data",
			format:         "json",
			expectedSuffix: ".json",
			expectTempFile: false,
		},
		{
			name:           "File with wrong extension gets new extension appended",
			outFile:        "report.txt",
			format:         "html",
			expectedSuffix: ".html",
			expectTempFile: false,
		},
		{
			name:           "Case insensitive extension matching - uppercase file",
			outFile:        "Chart.HTML",
			format:         "html",
			expectedSuffix: ".HTML",
			expectTempFile: false,
		},
		{
			name:           "Case insensitive extension matching - uppercase format",
			outFile:        "data.json",
			format:         "JSON",
			expectedSuffix: ".JSON",
			expectTempFile: false,
		},
		{
			name:           "Multiple dots in filename",
			outFile:        "my.data.file",
			format:         "html",
			expectedSuffix: ".html",
			expectTempFile: false,
		},
		{
			name:           "Absolute path",
			outFile:        "/tmp/output",
			format:         "json",
			expectedSuffix: ".json",
			expectTempFile: false,
		},
		{
			name:           "Relative path with correct extension",
			outFile:        "./reports/bench.html",
			format:         "html",
			expectedSuffix: ".html",
			expectTempFile: false,
		},
		{
			name:           "Empty format adds dot",
			outFile:        "test",
			format:         "",
			expectedSuffix: ".",
			expectTempFile: false,
		},
		{
			name:           "Format with special characters",
			outFile:        "backup",
			format:         "tar.gz",
			expectedSuffix: ".tar.gz",
			expectTempFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveOutputFileName(tt.outFile, tt.format)

			if tt.expectTempFile {
				// For temp files, we can't predict the exact name, but we can verify:
				// 1. It's not empty
				// 2. It has the correct extension
				assert.NotEmpty(t, result, "Temp file path should not be empty")
				assert.True(t, strings.HasSuffix(result, tt.expectedSuffix),
					"Temp file should have correct extension: expected suffix %q in %q", tt.expectedSuffix, result)
			} else {
				// For regular files, verify the expected behavior
				if tt.outFile == "" {
					// This case should have been handled as temp file
					assert.Fail(t, "Unexpected case: empty outFile should create temp file")
				} else {
					// Check if extension should be added
					expectedResult := tt.outFile
					outFileLower := strings.ToLower(tt.outFile)

					// The actual function checks: strings.HasSuffix(strings.ToLower(outFile), "."+format)
					// So it lowercases the file but not the format
					if !strings.HasSuffix(outFileLower, "."+tt.format) {
						if tt.format != "" {
							expectedResult = tt.outFile + "." + tt.format
						} else {
							expectedResult = tt.outFile + "."
						}
					}
					assert.Equal(t, expectedResult, result, "Output filename should match expected")
				}

				assert.True(t, strings.HasSuffix(result, tt.expectedSuffix),
					"Result should end with expected suffix: %q, got: %q", tt.expectedSuffix, result)
			}
		})
	}
}

// TestResolveOutputFileNameBoundaryConditions tests edge cases and boundary conditions
func TestResolveOutputFileNameBoundaryConditions(t *testing.T) {
	t.Run("Very long filename", func(t *testing.T) {
		longName := strings.Repeat("a", 200)
		result := resolveOutputFileName(longName, "html")
		expected := longName + ".html"
		assert.Equal(t, expected, result)
		assert.True(t, strings.HasSuffix(result, ".html"))
	})

	t.Run("Filename with Unicode characters", func(t *testing.T) {
		unicodeName := "测试文件"
		result := resolveOutputFileName(unicodeName, "json")
		expected := unicodeName + ".json"
		assert.Equal(t, expected, result)
		assert.True(t, strings.HasSuffix(result, ".json"))
	})

	t.Run("Filename with spaces", func(t *testing.T) {
		nameWithSpaces := "my test file"
		result := resolveOutputFileName(nameWithSpaces, "html")
		expected := nameWithSpaces + ".html"
		assert.Equal(t, expected, result)
	})

	t.Run("Format with dots", func(t *testing.T) {
		result := resolveOutputFileName("backup", "tar.gz")
		assert.Equal(t, "backup.tar.gz", result)
		assert.True(t, strings.HasSuffix(result, ".tar.gz"))
	})

	t.Run("Multiple consecutive dots in filename", func(t *testing.T) {
		result := resolveOutputFileName("test..file", "html")
		assert.Equal(t, "test..file.html", result)
	})

	t.Run("Filename ending with dot", func(t *testing.T) {
		result := resolveOutputFileName("test.", "html")
		// Should add .html after the existing dot
		assert.Equal(t, "test..html", result)
	})

	t.Run("Format matching with different cases", func(t *testing.T) {
		// File has lowercase extension, format is uppercase
		result := resolveOutputFileName("file.html", "HTML")
		// The current implementation doesn't normalize the format parameter to lowercase
		// So it will add the uppercase extension since it doesn't match
		assert.Equal(t, "file.html.HTML", result, "Current implementation adds extension when format case differs")

		// File has uppercase extension, format is lowercase
		result = resolveOutputFileName("FILE.JSON", "json")
		// The function does case-insensitive matching: strings.ToLower("FILE.JSON") gives "file.json"
		// and it checks if "file.json" ends with ".json" which it does, so no extension is added
		assert.Equal(t, "FILE.JSON", result, "Function does case-insensitive matching")
	})

	t.Run("Complex path with extension matching", func(t *testing.T) {
		complexPath := "/home/user/project/data/results.final.json"
		result := resolveOutputFileName(complexPath, "json")
		assert.Equal(t, complexPath, result, "Should not modify when extension matches")

		result = resolveOutputFileName(complexPath, "html")
		assert.Equal(t, complexPath+".html", result, "Should add extension when no match")
	})

	t.Run("Windows-style paths", func(t *testing.T) {
		windowsPath := "C:\\Users\\test\\output"
		result := resolveOutputFileName(windowsPath, "html")
		assert.Equal(t, windowsPath+".html", result)
	})

	t.Run("Relative paths with dots", func(t *testing.T) {
		relativePath := "../output/data"
		result := resolveOutputFileName(relativePath, "json")
		assert.Equal(t, relativePath+".json", result)

		relativePath = "./current.dir/file.html"
		result = resolveOutputFileName(relativePath, "html")
		assert.Equal(t, relativePath, result, "Should not modify when extension matches")
	})
}

// TestResolveOutputFileNameTempFileValidation tests temp file creation validation
func TestResolveOutputFileNameTempFileValidation(t *testing.T) {
	t.Run("Temp file creation with different formats", func(t *testing.T) {
		formats := []string{"html", "json", "xml", "txt", "csv"}

		for _, format := range formats {
			t.Run("Format_"+format, func(t *testing.T) {
				result := resolveOutputFileName("", format)

				// Verify temp file characteristics
				assert.NotEmpty(t, result, "Temp file path should not be empty")
				assert.True(t, strings.HasSuffix(result, "."+format),
					"Temp file should have correct extension for format %s", format)

				// Verify the file was actually created
				_, err := os.Stat(result)
				assert.NoError(t, err, "Temp file should exist after creation")

				// Clean up
				os.Remove(result)
			})
		}
	})

	t.Run("Multiple temp file calls create unique files", func(t *testing.T) {
		const numFiles = 5
		tempFiles := make([]string, numFiles)

		// Create multiple temp files
		for i := 0; i < numFiles; i++ {
			tempFiles[i] = resolveOutputFileName("", "test")
		}

		// Clean up
		defer func() {
			for _, file := range tempFiles {
				os.Remove(file)
			}
		}()

		// Verify all files are unique
		fileSet := make(map[string]bool)
		for i, file := range tempFiles {
			assert.False(t, fileSet[file], "File %d should be unique: %s", i, file)
			fileSet[file] = true

			// Verify file exists
			_, err := os.Stat(file)
			assert.NoError(t, err, "Temp file %d should exist: %s", i, file)
		}
	})

	t.Run("Temp file in system temp directory", func(t *testing.T) {
		result := resolveOutputFileName("", "html")
		defer os.Remove(result)

		// Get system temp directory
		systemTempDir := os.TempDir()

		// Verify temp file is in system temp directory or subdirectory
		absResult, err := filepath.Abs(result)
		assert.NoError(t, err)

		absSystemTemp, err := filepath.Abs(systemTempDir)
		assert.NoError(t, err)

		// Check if temp file path starts with system temp directory
		assert.True(t, strings.HasPrefix(absResult, absSystemTemp),
			"Temp file should be in system temp directory. File: %s, TempDir: %s", absResult, absSystemTemp)
	})
}
