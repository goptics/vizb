package cliout

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"
	"unsafe"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
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

func (s *ClioutSuite) TestWarnfFormatsMessage() {
	out := s.capture(func() {
		Warnf("ignored flag %s for %s", "--select", "go")
	})
	s.Contains(out, "ignored flag --select for go")
	s.Contains(out, ">")
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

func (s *ClioutSuite) TestPairAccentEmptyValueAndAccent() {
	// Empty value: key only, no colon.
	s.Equal("Reading data", PairAccent("Reading data", "", valueAccent))
	s.Equal("Reading data", PairAccent("Reading data", "", ""))

	// Empty accent defaults to valueAccent; Ascii profile keeps text plain.
	var sink bytes.Buffer
	Configure(&sink)
	defer Configure(nil)

	s.Equal("key: val", PairAccent("key", "val", ""))
	s.Equal("key: val", Pair("key", "val"))
}

func (s *ClioutSuite) TestParserAndFormatAccents() {
	s.Equal("#00ADD8", ParserAccent("go"))
	s.Equal("#F0DB4F", ParserAccent("js:vitest"))
	s.Equal("#F0DB4F", ParserAccent("js")) // bare js prefix
	s.Equal("#DEA584", ParserAccent("rs:criterion"))
	s.Equal("#DEA584", ParserAccent("rs")) // bare rs prefix
	s.Equal("#217346", ParserAccent("csv"))
	s.Equal(accentJSON, ParserAccent("json"))
	s.Equal(valueAccent, ParserAccent("unknown"))
	s.Equal(accentHTML, FormatAccent("html"))
	s.Equal(AccentCount, FormatAccent("json"))
	s.Equal(AccentCount, FormatAccent("JSON")) // case-insensitive
	s.Equal(valueAccent, FormatAccent("svg"))  // default branch
	s.Equal(valueAccent, FormatAccent(""))
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

func (s *ClioutSuite) TestColorProfileEnvNoColor() {
	// Any *os.File hits EnvNoColor before the TTY check; force NO_COLOR so
	// the Ascii return is from that branch (deterministic, no real TTY).
	f, err := os.CreateTemp(s.T().TempDir(), "cliout-color-*")
	s.Require().NoError(err)
	defer f.Close()

	s.T().Setenv("NO_COLOR", "1")
	s.T().Setenv("CLICOLOR_FORCE", "")
	s.Equal(termenv.Ascii, ColorProfile(f))
}

func (s *ClioutSuite) TestColorProfileNonTTYFileIsAscii() {
	// *os.File that is not a terminal: EnvNoColor false → IsTerminal false → Ascii.
	f, err := os.CreateTemp(s.T().TempDir(), "cliout-nontty-*")
	s.Require().NoError(err)
	defer f.Close()

	s.T().Setenv("NO_COLOR", "")
	s.T().Setenv("CLICOLOR", "1")
	s.T().Setenv("CLICOLOR_FORCE", "")
	s.False(term.IsTerminal(int(f.Fd())))
	s.Equal(termenv.Ascii, ColorProfile(f))
}

func (s *ClioutSuite) TestColorProfileCOLORTERMOnTTY() {
	// COLORTERM / ANSI256 branches require a real terminal. Use a PTY when
	// available; skip cleanly if the platform cannot open one.
	master, slave, err := openTestPTY()
	if err != nil {
		s.T().Skipf("no PTY available: %v", err)
	}
	defer master.Close()
	defer slave.Close()
	s.Require().True(term.IsTerminal(int(slave.Fd())))

	// Clear NO_COLOR so the TTY path is reachable.
	s.T().Setenv("NO_COLOR", "")
	s.T().Setenv("CLICOLOR", "1")
	s.T().Setenv("CLICOLOR_FORCE", "1")

	s.T().Setenv("COLORTERM", "truecolor")
	s.Equal(termenv.TrueColor, ColorProfile(slave))

	s.T().Setenv("COLORTERM", "24bit")
	s.Equal(termenv.TrueColor, ColorProfile(slave))

	s.T().Setenv("COLORTERM", "")
	s.Equal(termenv.ANSI256, ColorProfile(slave))
}

func (s *ClioutSuite) TestColorProfileStderrWriterProbesStderr() {
	// stderrWriter is not *os.File; ColorProfile should fall through to os.Stderr.
	// Non-TTY stderr (typical under go test) yields Ascii.
	s.Equal(ColorProfile(os.Stderr), ColorProfile(stderrWriter{}))
}

func (s *ClioutSuite) TestAccentUsesConfiguredProfile() {
	// Force Ascii via a non-file Configure sink; Accent must stay plain.
	var sink bytes.Buffer
	Configure(&sink)
	defer Configure(nil)

	s.Equal("42", Accent("42", AccentCount))
	s.Equal(termenv.Ascii, accentProfile)
}

func (s *ClioutSuite) TestAccentEmptyShortCircuit() {
	// Empty string returns immediately regardless of profile.
	prevProfile := accentProfile
	prevRender := accentRender
	defer func() {
		accentProfile = prevProfile
		accentRender = prevRender
	}()

	accentProfile = termenv.TrueColor
	accentRender = lipgloss.NewRenderer(&bytes.Buffer{})
	accentRender.SetColorProfile(termenv.TrueColor)

	s.Equal("", Accent("", AccentCount))
	s.Equal("", Accent("", BrandGreen))
}

func (s *ClioutSuite) TestAccentColorPath() {
	// Same-package access: force a non-Ascii profile so Accent emits ANSI.
	prevProfile := accentProfile
	prevRender := accentRender
	defer func() {
		accentProfile = prevProfile
		accentRender = prevRender
	}()

	var sink bytes.Buffer
	accentProfile = termenv.TrueColor
	accentRender = lipgloss.NewRenderer(&sink)
	accentRender.SetColorProfile(termenv.TrueColor)

	got := Accent("42", AccentCount)
	s.Contains(got, "42")
	s.Contains(got, "\x1b[")
	s.NotEqual("42", got)
}

func (s *ClioutSuite) TestAccentNilRendererShortCircuit() {
	prevProfile := accentProfile
	prevRender := accentRender
	defer func() {
		accentProfile = prevProfile
		accentRender = prevRender
	}()

	accentProfile = termenv.TrueColor
	accentRender = nil
	s.Equal("plain", Accent("plain", AccentCount))
}

func TestClioutSuite(t *testing.T) {
	suite.Run(t, new(ClioutSuite))
}

// openTestPTY opens a Linux pseudo-terminal pair for ColorProfile TTY branches.
// Returns an error when the host cannot allocate a PTY (skip, no flake).
func openTestPTY() (master, slave *os.File, err error) {
	master, err = os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}
	var unlock int32
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, master.Fd(), uintptr(unix.TIOCSPTLCK), uintptr(unsafe.Pointer(&unlock)))
	if errno != 0 {
		_ = master.Close()
		return nil, nil, errno
	}
	var n uint32
	_, _, errno = unix.Syscall(unix.SYS_IOCTL, master.Fd(), uintptr(unix.TIOCGPTN), uintptr(unsafe.Pointer(&n)))
	if errno != 0 {
		_ = master.Close()
		return nil, nil, errno
	}
	slave, err = os.OpenFile("/dev/pts/"+strconv.Itoa(int(n)), os.O_RDWR, 0)
	if err != nil {
		_ = master.Close()
		return nil, nil, err
	}
	return master, slave, nil
}
