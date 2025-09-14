package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
)

func TestTempFileCreation(t *testing.T) {
	t.Run("Create temp file with prefix and extension", func(t *testing.T) {
		tempFile := shared.MustCreateTempFile("test-prefix", "json")
		defer os.Remove(tempFile)

		// Verify temp file was created
		assert.FileExists(t, tempFile, "Temp file should exist")

		// Verify filename contains prefix
		basename := filepath.Base(tempFile)
		assert.Contains(t, basename, "test-prefix", "Filename should contain prefix")
		assert.Contains(t, basename, ".json", "Filename should have correct extension")
	})

	t.Run("Create temp file with empty prefix", func(t *testing.T) {
		tempFile := shared.MustCreateTempFile("", "txt")
		defer os.Remove(tempFile)

		assert.FileExists(t, tempFile, "Temp file should exist even with empty prefix")
		assert.Contains(t, filepath.Base(tempFile), ".txt", "Should have correct extension")
	})

	t.Run("Create temp file with empty extension", func(t *testing.T) {
		tempFile := shared.MustCreateTempFile("test", "")
		defer os.Remove(tempFile)

		assert.FileExists(t, tempFile, "Temp file should exist even with empty extension")
		basename := filepath.Base(tempFile)
		assert.Contains(t, basename, "test", "Should contain prefix")
	})

	t.Run("Multiple temp files are unique", func(t *testing.T) {
		tempFiles := make([]string, 5)

		for i := 0; i < 5; i++ {
			tempFiles[i] = shared.MustCreateTempFile("unique", "tmp")
		}

		// Clean up
		defer func() {
			for _, file := range tempFiles {
				os.Remove(file)
			}
		}()

		// Verify all files are unique
		fileSet := make(map[string]bool)
		for _, file := range tempFiles {
			assert.False(t, fileSet[file], "All temp files should be unique")
			fileSet[file] = true
			assert.FileExists(t, file, "Each temp file should exist")
		}
	})

	t.Run("Special characters in prefix and extension", func(t *testing.T) {
		tempFile := shared.MustCreateTempFile("test-with_chars.123", "ext.json")
		defer os.Remove(tempFile)

		assert.FileExists(t, tempFile, "Should handle special characters")
		basename := filepath.Base(tempFile)
		assert.Contains(t, basename, "test-with_chars.123", "Should preserve special characters in prefix")
	})
}