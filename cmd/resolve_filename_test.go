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
		expectedFile   string
		expectTempFile bool
	}{
		{
			name:           "Empty outFile creates temp html file",
			outFile:        "",
			expectTempFile: true,
		},
		{
			name:         "File with .json extension unchanged",
			outFile:      "output.json",
			expectedFile: "output.json",
		},
		{
			name:         "File with .html extension unchanged",
			outFile:      "output.html",
			expectedFile: "output.html",
		},
		{
			name:         "File with .JSON extension (uppercase) unchanged",
			outFile:      "output.JSON",
			expectedFile: "output.JSON",
		},
		{
			name:         "File without extension gets .html added",
			outFile:      "data",
			expectedFile: "data.html",
		},
		{
			name:         "File with .txt extension unchanged",
			outFile:      "report.txt",
			expectedFile: "report.txt",
		},
		{
			name:         "Absolute path with .json extension",
			outFile:      "/tmp/output.json",
			expectedFile: "/tmp/output.json",
		},
		{
			name:         "Relative path with .html extension",
			outFile:      "./reports/bench.html",
			expectedFile: "./reports/bench.html",
		},
		{
			name:         "File with multiple dots and .json extension",
			outFile:      "my.data.results.json",
			expectedFile: "my.data.results.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultFile := resolveOutputFileName(tt.outFile)

			if tt.expectTempFile {
				assert.NotEmpty(t, resultFile, "Temp file path should not be empty")
				assert.True(t, strings.HasSuffix(resultFile, ".html"),
					"Temp file should have .html extension: got %q", resultFile)
			} else {
				assert.Equal(t, tt.expectedFile, resultFile, "Output filename should match expected")
			}
		})
	}
}

// TestResolveOutputFileNameBoundaryConditions tests edge cases and boundary conditions
func TestResolveOutputFileNameBoundaryConditions(t *testing.T) {
	t.Run("Very long filename without extension", func(t *testing.T) {
		longName := strings.Repeat("a", 200)
		resultFile := resolveOutputFileName(longName)
		expected := longName + ".html"
		assert.Equal(t, expected, resultFile)
	})

	t.Run("Very long filename with extension", func(t *testing.T) {
		longName := strings.Repeat("a", 200) + ".json"
		resultFile := resolveOutputFileName(longName)
		assert.Equal(t, longName, resultFile)
	})

	t.Run("Filename with Unicode characters", func(t *testing.T) {
		unicodeName := "测试文件.json"
		resultFile := resolveOutputFileName(unicodeName)
		assert.Equal(t, unicodeName, resultFile)
	})

	t.Run("Filename with spaces", func(t *testing.T) {
		nameWithSpaces := "my test file"
		resultFile := resolveOutputFileName(nameWithSpaces)
		expected := nameWithSpaces + ".html"
		assert.Equal(t, expected, resultFile)
	})

	t.Run("Filename ending with dot", func(t *testing.T) {
		resultFile := resolveOutputFileName("test.")
		// Filepath.Ext returns "." for "test.", which is not empty, so file is unchanged
		assert.Equal(t, "test.", resultFile)
	})

	t.Run("Complex path with .json extension", func(t *testing.T) {
		complexPath := "/home/user/project/data/results.final.json"
		resultFile := resolveOutputFileName(complexPath)
		assert.Equal(t, complexPath, resultFile)
	})

	t.Run("Complex path with other extension", func(t *testing.T) {
		complexPath := "/home/user/project/data/results.final.txt"
		resultFile := resolveOutputFileName(complexPath)
		assert.Equal(t, complexPath, resultFile)
	})

	t.Run("Windows-style paths without extension", func(t *testing.T) {
		windowsPath := "C:\\Users\\test\\output"
		resultFile := resolveOutputFileName(windowsPath)
		assert.Equal(t, windowsPath+".html", resultFile)
	})

	t.Run("Relative paths with dots", func(t *testing.T) {
		relativePath := "../output/data"
		resultFile := resolveOutputFileName(relativePath)
		assert.Equal(t, relativePath+".html", resultFile)

		relativePath = "./current.dir/file.json"
		resultFile = resolveOutputFileName(relativePath)
		assert.Equal(t, relativePath, resultFile)
	})
}

// TestResolveOutputFileNameTempFileValidation tests temp file creation validation
func TestResolveOutputFileNameTempFileValidation(t *testing.T) {
	t.Run("Temp file creation defaults to html", func(t *testing.T) {
		resultFile := resolveOutputFileName("")

		assert.NotEmpty(t, resultFile, "Temp file path should not be empty")
		assert.True(t, strings.HasSuffix(resultFile, ".html"),
			"Temp file should have .html extension")

		// Verify the file was actually created
		_, err := os.Stat(resultFile)
		assert.NoError(t, err, "Temp file should exist after creation")

		// Clean up
		os.Remove(resultFile)
	})

	t.Run("Multiple temp file calls create unique files", func(t *testing.T) {
		const numFiles = 5
		tempFiles := make([]string, numFiles)

		// Create multiple temp files
		for i := 0; i < numFiles; i++ {
			tempFiles[i] = resolveOutputFileName("")
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
		resultFile := resolveOutputFileName("")
		defer os.Remove(resultFile)

		// Get system temp directory
		systemTempDir := os.TempDir()

		// Verify temp file is in system temp directory or subdirectory
		absResult, err := filepath.Abs(resultFile)
		assert.NoError(t, err)

		absSystemTemp, err := filepath.Abs(systemTempDir)
		assert.NoError(t, err)

		// Check if temp file path starts with system temp directory
		assert.True(t, strings.HasPrefix(absResult, absSystemTemp),
			"Temp file should be in system temp directory. File: %s, TempDir: %s", absResult, absSystemTemp)
	})
}

// TestInferFormatFromExtension tests the format inference function
func TestInferFormatFromExtension(t *testing.T) {
	tests := []struct {
		name           string
		outFile        string
		expectedFormat string
	}{
		{"json extension", "test.json", "json"},
		{"JSON uppercase", "test.JSON", "json"},
		{"Json mixed case", "test.Json", "json"},
		{"html extension", "test.html", "html"},
		{"HTML uppercase", "test.HTML", "html"},
		{"txt extension defaults to html", "test.txt", "html"},
		{"xml extension defaults to html", "test.xml", "html"},
		{"no extension defaults to html", "test", "html"},
		{"path with json", "/path/to/file.json", "json"},
		{"path with html", "/path/to/file.html", "html"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferFormatFromExtension(tt.outFile)
			assert.Equal(t, tt.expectedFormat, result)
		})
	}
}
