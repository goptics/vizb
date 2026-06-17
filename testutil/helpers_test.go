package testutil

import (
	"fmt"
	"os"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

type HelpersSuite struct {
	suite.Suite
}

func (s *HelpersSuite) TestTrapOsExitPanicRecordsCallAndPanics() {
	restore, exitCalled := TrapOsExitPanic(s.T())
	defer restore()

	s.Panics(func() { shared.OsExit(1) })
	s.True(*exitCalled)
}

func (s *HelpersSuite) TestWriteBenchFileAndReadDataset() {
	dir := s.T().TempDir()
	input := WriteBenchFile(s.T(), dir, "bench.txt", "")
	out := dir + "/out.json"
	WriteJSON(s.T(), out, shared.Dataset{Name: "Test"})

	ds := ReadDataset(s.T(), out)
	s.Equal("Test", ds.Name)
	s.FileExists(input)
}

func (s *HelpersSuite) TestCaptureStdout() {
	out := CaptureStdout(func() { fmt.Print("hello-capture") })
	s.Contains(out, "hello-capture")
}

func (s *HelpersSuite) TestCaptureStderr() {
	out := CaptureStderr(func() { fmt.Fprint(os.Stderr, "warn-capture") })
	s.Contains(out, "warn-capture")
}

func TestHelpersSuite(t *testing.T) {
	suite.Run(t, new(HelpersSuite))
}
