package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// OutputSuite covers the shared output-path helpers.
type OutputSuite struct {
	suite.Suite
}

func (s *OutputSuite) TestResolveOutputFileName() {
	tests := []struct {
		name           string
		outFile        string
		expectedFile   string
		expectTempFile bool
	}{
		{name: "Empty outFile creates temp html file", outFile: "", expectTempFile: true},
		{name: "File with .json extension unchanged", outFile: "output.json", expectedFile: "output.json"},
		{name: "File with .html extension unchanged", outFile: "output.html", expectedFile: "output.html"},
		{name: "File with .JSON extension (uppercase) unchanged", outFile: "output.JSON", expectedFile: "output.JSON"},
		{name: "File without extension gets .html added", outFile: "data", expectedFile: "data.html"},
		{name: "File with .txt extension unchanged", outFile: "report.txt", expectedFile: "report.txt"},
		{name: "Absolute path with .json extension", outFile: "/tmp/output.json", expectedFile: "/tmp/output.json"},
		{name: "File with multiple dots and .json extension", outFile: "my.data.results.json", expectedFile: "my.data.results.json"},
		{name: "Filename ending with dot unchanged", outFile: "test.", expectedFile: "test."},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := ResolveOutputFileName(tt.outFile)
			if tt.expectTempFile {
				s.NotEmpty(result)
				s.True(strings.HasSuffix(result, ".html"), "temp file should end in .html: %q", result)
			} else {
				s.Equal(tt.expectedFile, result)
			}
		})
	}
}

func (s *OutputSuite) TestResolveOutputFileNameTempFileIsCreatedAndUnique() {
	const numFiles = 5
	files := make([]string, numFiles)
	for i := range files {
		files[i] = ResolveOutputFileName("")
	}
	defer func() {
		for _, f := range files {
			os.Remove(f)
		}
	}()

	seen := map[string]bool{}
	for _, f := range files {
		s.False(seen[f], "temp file should be unique: %s", f)
		seen[f] = true
		s.FileExists(f)
		s.True(strings.HasPrefix(f, os.TempDir()) || strings.HasPrefix(mustAbs(s, f), mustAbs(s, os.TempDir())))
	}
}

func mustAbs(s *OutputSuite, p string) string {
	abs, err := filepath.Abs(p)
	s.Require().NoError(err)
	return abs
}

func (s *OutputSuite) TestInferFormatFromExtension() {
	tests := []struct {
		name     string
		outFile  string
		expected string
	}{
		{"json extension", "test.json", "json"},
		{"JSON uppercase", "test.JSON", "json"},
		{"Json mixed case", "test.Json", "json"},
		{"html extension", "test.html", "html"},
		{"txt defaults to html", "test.txt", "html"},
		{"no extension defaults to html", "test", "html"},
		{"path with json", "/path/to/file.json", "json"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expected, InferFormatFromExtension(tt.outFile))
		})
	}
}

func (s *OutputSuite) TestHandleOutputResultWithNamedOutputShowsPath() {
	filename := filepath.Join(s.T().TempDir(), "out.txt")
	s.Require().NoError(os.WriteFile(filename, []byte("<html>Test</html>"), 0644))
	file, err := os.Open(filename)
	s.Require().NoError(err)
	defer file.Close()

	output := s.captureStdout(func() {
		HandleOutputResult(file, "specified_output.html")
	})

	s.Contains(output, "📄 Output file:")
	s.Contains(output, filename)
	s.NotContains(output, "<html>Test</html>")
	s.NotContains(output, "\033[H\033[2J")
}

func (s *OutputSuite) TestHandleOutputResultWithStdoutShowsContent() {
	filename := filepath.Join(s.T().TempDir(), "stdout.txt")
	content := "Test benchmark output\nLine 2"
	s.Require().NoError(os.WriteFile(filename, []byte(content), 0644))
	file, err := os.Open(filename)
	s.Require().NoError(err)
	defer file.Close()

	output := s.captureStdout(func() {
		HandleOutputResult(file, "")
	})

	s.Contains(output, "\033[H\033[2J")
	s.Contains(output, content)
	s.NotContains(output, "📄 Output file:")
}

func (s *OutputSuite) TestConvertToDataset() {
	dir := s.T().TempDir()

	s.Run("valid dataset JSON", func() {
		valid := filepath.Join(dir, "valid.json")
		ds := shared.Dataset{Name: "Test", Data: []shared.DataPoint{{Name: "B1", Stats: []shared.Stat{{Type: "time", Value: 100}}}}}
		data, err := json.Marshal(ds)
		s.Require().NoError(err)
		s.Require().NoError(os.WriteFile(valid, data, 0644))

		result := convertToDataset(valid)
		s.Require().NotNil(result)
		s.Equal("Test", result.Name)
		s.Len(result.Data, 1)
	})

	s.Run("invalid JSON returns nil", func() {
		invalid := filepath.Join(dir, "invalid.json")
		s.Require().NoError(os.WriteFile(invalid, []byte("not json"), 0644))
		s.Nil(convertToDataset(invalid))
	})
}

// captureStdout runs fn with os.Stdout redirected and returns what it printed.
func (s *OutputSuite) captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestOutputSuite(t *testing.T) {
	suite.Run(t, new(OutputSuite))
}
