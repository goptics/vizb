package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

// OptionsSuite covers option validation, parser-config mapping, and chart
// selection assembly.
type OptionsSuite struct {
	suite.Suite
}

// captureStderr runs fn with os.Stderr redirected and returns what it printed
// (validation warnings go to stderr).
func (s *OptionsSuite) captureStderr(fn func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = old }()

	fn()

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func (s *OptionsSuite) TestCommonValidateNormalisesUnits() {
	o := &CommonOptions{MemUnit: "kb", TimeUnit: "ns", NumberUnit: "m", GroupPattern: "x", Parser: "auto"}
	o.Validate()
	s.Equal("KB", o.MemUnit)
	s.Equal("M", o.NumberUnit)
}

func (s *OptionsSuite) TestCommonValidateWarnsAndDefaultsInvalid() {
	o := &CommonOptions{MemUnit: "invalid", TimeUnit: "ns", GroupPattern: "x", Parser: "auto"}
	out := s.captureStderr(func() { o.Validate() })
	s.Equal("B", o.MemUnit)
	s.Contains(out, "Invalid memory unit")
}

func (s *OptionsSuite) TestCommonValidateRejectsUnknownParser() {
	o := &CommonOptions{TimeUnit: "ns", MemUnit: "B", GroupPattern: "x", Parser: "nope"}
	s.captureStderr(func() { o.Validate() })
	s.Equal("auto", o.Parser)
}

func (s *OptionsSuite) TestLinearValidateNormalisesSort() {
	o := &LinearOptions{Sort: "ASC"}
	o.CommonOptions = CommonOptions{TimeUnit: "ns", MemUnit: "B", GroupPattern: "x", Parser: "auto"}
	o.Validate()
	s.Equal("asc", o.Sort)
}

func (s *OptionsSuite) TestParseConfigMapsFields() {
	o := &CommonOptions{
		GroupPattern: "n/x", GroupRegex: "re", Group: []string{"a", "b"},
		Filter: "keep", MemUnit: "KB", TimeUnit: "us", NumberUnit: "M",
	}
	cfg := o.ParseConfig()
	s.Equal("n/x", cfg.GroupPattern)
	s.Equal("re", cfg.GroupRegex)
	s.Equal([]string{"a", "b"}, cfg.Group)
	s.Equal("keep", cfg.Filter)
	s.Equal("KB", cfg.MemUnit)
	s.Equal("us", cfg.TimeUnit)
	s.Equal("M", cfg.NumberUnit)
}

func (s *OptionsSuite) TestValidateScale() {
	s.Run("log is accepted", func() {
		scale := "LOG"
		ValidateScale(&scale)
		s.Equal("log", scale)
	})
	s.Run("invalid falls back to linear", func() {
		scale := "bogus"
		s.captureStderr(func() { ValidateScale(&scale) })
		s.Equal("linear", scale)
	})
}

func TestOptionsSuite(t *testing.T) {
	suite.Run(t, new(OptionsSuite))
}
