package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

// MergeSuite covers the merge subcommand end-to-end via rootCmd.Execute.
type MergeSuite struct {
	suite.Suite
	origOsExit func(int)
	exitCode   int
}

func (s *MergeSuite) SetupTest() {
	s.origOsExit = shared.OsExit
	s.exitCode = 0
	shared.OsExit = func(code int) { s.exitCode = code }
}

func (s *MergeSuite) TearDownTest() {
	shared.OsExit = s.origOsExit
}

func (s *MergeSuite) writeJSON(path string, v any) {
	data, err := json.Marshal(v)
	s.Require().NoError(err)
	s.Require().NoError(os.WriteFile(path, data, 0644))
}

func (s *MergeSuite) TestMergeTwoFilesSkippingInvalid() {
	dir := s.T().TempDir()
	file1 := filepath.Join(dir, "bench1.json")
	file2 := filepath.Join(dir, "bench2.json")
	s.writeJSON(file1, shared.Dataset{Name: "Bench1", Data: []shared.DataPoint{{Name: "Test1", XAxis: "1", YAxis: "100"}}})
	s.writeJSON(file2, shared.Dataset{Name: "Bench2", Data: []shared.DataPoint{{Name: "Test2", XAxis: "2", YAxis: "200"}}})
	invalid := filepath.Join(dir, "invalid.json")
	s.Require().NoError(os.WriteFile(invalid, []byte("{invalid json"), 0644))

	out := filepath.Join(dir, "merged.json")
	rootCmd.SetArgs([]string{"merge", "-o", out, file1, file2, invalid})
	s.Require().NoError(rootCmd.Execute())
	s.Equal(0, s.exitCode)

	parsed := s.readDatasets(out)
	s.Len(parsed, 2)
	names := map[string]bool{}
	for _, b := range parsed {
		names[b.Name] = true
	}
	s.True(names["Bench1"])
	s.True(names["Bench2"])
}

func (s *MergeSuite) TestMergeDirectory() {
	dir := s.T().TempDir()
	s.writeJSON(filepath.Join(dir, "b1.json"), shared.Dataset{Name: "Bench1", Data: []shared.DataPoint{{Name: "Test1"}}})

	out := filepath.Join(dir, "merged_dir.json")
	rootCmd.SetArgs([]string{"merge", "-o", out, dir})
	s.Require().NoError(rootCmd.Execute())
	s.Equal(0, s.exitCode)

	parsed := s.readDatasets(out)
	s.Len(parsed, 1)
	s.Equal("Bench1", parsed[0].Name)
}

func (s *MergeSuite) TestMergeArrayInput() {
	dir := s.T().TempDir()
	s.writeJSON(filepath.Join(dir, "array.json"), []shared.Dataset{
		{Name: "Bench1", Data: []shared.DataPoint{{Name: "Test1", XAxis: "1", YAxis: "100"}}},
		{Name: "Bench2", Data: []shared.DataPoint{{Name: "Test2", XAxis: "2", YAxis: "200"}}},
	})

	out := filepath.Join(dir, "merged_array.json")
	rootCmd.SetArgs([]string{"merge", "-o", out, filepath.Join(dir, "array.json")})
	s.Require().NoError(rootCmd.Execute())
	s.Equal(0, s.exitCode)

	parsed := s.readDatasets(out)
	s.Len(parsed, 2)
}

func (s *MergeSuite) readDatasets(path string) []shared.Dataset {
	content, err := os.ReadFile(path)
	s.Require().NoError(err)
	var parsed []shared.Dataset
	s.Require().NoError(json.Unmarshal(content, &parsed))
	return parsed
}

func TestMergeSuite(t *testing.T) {
	suite.Run(t, new(MergeSuite))
}
