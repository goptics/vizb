package cliout

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/muesli/termenv"
	"github.com/stretchr/testify/suite"
)

type ClioutSuite struct {
	suite.Suite
}

func (s *ClioutSuite) capture(fn func()) string {
	old := os.Stderr
	r, w, err := os.Pipe()
	s.Require().NoError(err)
	os.Stderr = w
	Configure(nil)

	fn()

	s.Require().NoError(w.Close())
	os.Stderr = old
	Configure(nil)

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String()
}

func (s *ClioutSuite) TestInfoWarnErrorUseChevronMarks() {
	out := s.capture(func() {
		Info("hello")
		Warn("careful")
		Error("boom")
	})

	s.Contains(out, ">")
	s.Contains(out, "hello")
	s.Contains(out, "careful")
	s.Contains(out, "boom")
	s.NotContains(out, "INFO")
	s.NotContains(out, "WARN")
	s.NotContains(out, "ERROR")
	s.Contains(out, "> hello")
	s.NotContains(out, "\x1b[48") // no background level pills
}

func (s *ClioutSuite) TestInfoPairKeyColonValue() {
	out := s.capture(func() {
		InfoPair("Reading data", "precision.csv")
		InfoPairAccent("Auto-detected parser", "csv", ParserAccent("csv"))
	})

	s.Contains(out, "Reading data: precision.csv")
	s.Contains(out, "Auto-detected parser: csv")
	s.Equal("Reading data: x.csv", Pair("Reading data", "x.csv"))
}

func (s *ClioutSuite) TestParserAndFormatAccents() {
	s.Equal("#00ADD8", ParserAccent("go"))
	s.Equal("#F0DB4F", ParserAccent("js:vitest"))
	s.Equal("#DEA584", ParserAccent("rs:criterion"))
	s.Equal("#217346", ParserAccent("csv"))
	s.Equal(accentJSON, ParserAccent("json"))
	s.Equal(valueAccent, ParserAccent("unknown"))
	s.Equal(accentHTML, FormatAccent("html"))
	s.Equal(AccentCount, FormatAccent("json"))
	s.Equal(BrandGreen, BrandGreen)
}

func (s *ClioutSuite) TestLogClearsInPlaceActivityLine() {
	old := os.Stderr
	r, w, err := os.Pipe()
	s.Require().NoError(err)
	os.Stderr = w
	Configure(nil)

	_, _ = fmt.Fprint(w, "\r\033[K◐ Mapping")
	Info("Auto-grouped by column: order_date")

	s.Require().NoError(w.Close())
	os.Stderr = old
	Configure(nil)

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	out := buf.String()
	s.Contains(out, "Auto-grouped by column: order_date")
	s.NotContains(out, "Mapping>")
}

func (s *ClioutSuite) TestColorProfileNonFileIsAscii() {
	s.Equal(termenv.Ascii, ColorProfile(&bytes.Buffer{}))
}

func TestClioutSuite(t *testing.T) {
	suite.Run(t, new(ClioutSuite))
}
