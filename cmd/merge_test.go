package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/goptics/vizb/testutil"
	"github.com/stretchr/testify/suite"
)

// MergeSuite covers the merge subcommand end-to-end via rootCmd.Execute.
type MergeSuite struct {
	suite.Suite
	restoreOsExit func()
}

func (s *MergeSuite) SetupTest() {
	ResetTestState()
	s.restoreOsExit, _ = testutil.TrapOsExitPanic(s.T())
}

func (s *MergeSuite) TearDownTest() {
	s.restoreOsExit()
}

func (s *MergeSuite) TestMergeTwoFilesSkippingInvalid() {
	dir := s.T().TempDir()
	file1 := filepath.Join(dir, "bench1.json")
	file2 := filepath.Join(dir, "bench2.json")
	testutil.WriteJSON(s.T(), file1, shared.Dataset{Name: "Bench1", Data: []shared.DataPoint{{Name: "Test1", XAxis: "1", YAxis: "100"}}})
	testutil.WriteJSON(s.T(), file2, shared.Dataset{Name: "Bench2", Data: []shared.DataPoint{{Name: "Test2", XAxis: "2", YAxis: "200"}}})
	invalid := filepath.Join(dir, "invalid.json")
	s.Require().NoError(os.WriteFile(invalid, []byte("{invalid json"), 0644))

	out := filepath.Join(dir, "merged.json")
	rootCmd.SetArgs([]string{"merge", "-o", out, file1, file2, invalid})
	s.Require().NoError(rootCmd.Execute())

	parsed := s.readDatasets(out)
	s.Require().Len(parsed, 2)
	names := map[string]bool{}
	for _, b := range parsed {
		names[b.Name] = true
	}
	s.True(names["Bench1"])
	s.True(names["Bench2"])
}

func (s *MergeSuite) TestMergeDirectory() {
	dir := s.T().TempDir()
	testutil.WriteJSON(s.T(), filepath.Join(dir, "b1.json"), shared.Dataset{Name: "Bench1", Data: []shared.DataPoint{{Name: "Test1"}}})

	out := filepath.Join(dir, "merged_dir.json")
	rootCmd.SetArgs([]string{"merge", "-o", out, dir})
	s.Require().NoError(rootCmd.Execute())

	parsed := s.readDatasets(out)
	s.Require().Len(parsed, 1)
	s.Equal("Bench1", parsed[0].Name)
}

func (s *MergeSuite) TestMergeArrayInput() {
	dir := s.T().TempDir()
	testutil.WriteJSON(s.T(), filepath.Join(dir, "array.json"), []shared.Dataset{
		{Name: "Bench1", Data: []shared.DataPoint{{Name: "Test1", XAxis: "1", YAxis: "100"}}},
		{Name: "Bench2", Data: []shared.DataPoint{{Name: "Test2", XAxis: "2", YAxis: "200"}}},
	})

	out := filepath.Join(dir, "merged_array.json")
	rootCmd.SetArgs([]string{"merge", "-o", out, filepath.Join(dir, "array.json")})
	s.Require().NoError(rootCmd.Execute())

	parsed := s.readDatasets(out)
	s.Require().Len(parsed, 2)
}

func (s *MergeSuite) TestMergeNoArgsExits() {
	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	rootCmd.SetArgs([]string{"merge"})
	s.Panics(func() { _ = rootCmd.Execute() })
	s.True(*exitCalled)
}

func (s *MergeSuite) TestMergeEmptyDirectoryExits() {
	dir := s.T().TempDir()

	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	rootCmd.SetArgs([]string{"merge", "-o", filepath.Join(dir, "out.json"), dir})
	s.Panics(func() { _ = rootCmd.Execute() })
	s.True(*exitCalled)
}

func (s *MergeSuite) TestMergeDefaultTempOutput() {
	dir := s.T().TempDir()
	file1 := filepath.Join(dir, "bench1.json")
	testutil.WriteJSON(s.T(), file1, shared.Dataset{Name: "Bench1", Data: []shared.DataPoint{{Name: "T1"}}})

	outStr := testutil.CaptureStdout(func() {
		rootCmd.SetArgs([]string{"merge", file1})
		s.Require().NoError(rootCmd.Execute())
	})

	s.Contains(outStr, "Generated merged JSON")
}

func (s *MergeSuite) TestMergeOutputWithoutExtension() {
	dir := s.T().TempDir()
	file1 := filepath.Join(dir, "bench1.json")
	testutil.WriteJSON(s.T(), file1, shared.Dataset{Name: "Bench1", Data: []shared.DataPoint{{Name: "T1"}}})
	out := filepath.Join(dir, "merged")

	rootCmd.SetArgs([]string{"merge", "-o", out, file1})
	s.Require().NoError(rootCmd.Execute())

	s.FileExists(out + ".json")
}

func (s *MergeSuite) TestMergeDirectoryWithNoJSONExits() {
	dir := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0644))

	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	rootCmd.SetArgs([]string{"merge", "-o", filepath.Join(dir, "out.json"), dir})
	s.Panics(func() { _ = rootCmd.Execute() })
	s.True(*exitCalled)
}

func (s *MergeSuite) TestMergeSkipsInaccessiblePath() {
	dir := s.T().TempDir()
	file1 := filepath.Join(dir, "bench1.json")
	testutil.WriteJSON(s.T(), file1, shared.Dataset{Name: "Bench1", Data: []shared.DataPoint{{Name: "T1"}}})
	out := filepath.Join(dir, "merged.json")

	stderr := testutil.CaptureStderr(func() {
		rootCmd.SetArgs([]string{"merge", "-o", out, file1, "/nonexistent/bench.json"})
		s.Require().NoError(rootCmd.Execute())
	})

	s.Contains(stderr, "Warning")
	s.Contains(stderr, "cannot access")
	parsed := s.readDatasets(out)
	s.Require().Len(parsed, 1)
}

func (s *MergeSuite) TestMergeNoValidFilesExits() {
	dir := s.T().TempDir()
	invalid := filepath.Join(dir, "invalid.json")
	s.Require().NoError(os.WriteFile(invalid, []byte("{bad"), 0644))

	restore, exitCalled := testutil.TrapOsExitPanic(s.T())
	defer restore()

	rootCmd.SetArgs([]string{"merge", "-o", filepath.Join(dir, "out.json"), invalid})
	s.Panics(func() { _ = rootCmd.Execute() })
	s.True(*exitCalled)
}

func (s *MergeSuite) TestMergeDatasetsCoreErrorExits() {
	s.Panics(func() {
		mergeDatasets([]shared.Dataset{{Name: "Bench"}}, shared.Dimension("invalid"))
	})
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
