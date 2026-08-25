package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/goptics/vizb/pkg/cliout"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/suite"
)

type DownloadBarSuite struct {
	suite.Suite
}

func (s *DownloadBarSuite) TestNonTTYIsSilent() {
	var buf bytes.Buffer
	bar := NewDownloadBar(&buf)
	src := strings.NewReader("hello")
	s.Same(src, bar.Wrap(src, 5, "v1.0.0"))
	s.False(bar.IsTTY())
	s.NoError(bar.Finish())
	s.Empty(buf.String())
}

func (s *DownloadBarSuite) TestTTYPaintsThenFinishClears() {
	var buf bytes.Buffer
	bar := NewDownloadBar(&buf)
	bar.tty = true
	wrapped := bar.Wrap(bytes.NewReader(bytes.Repeat([]byte("a"), 100)), 100, "v1.1.0")
	_, err := io.Copy(io.Discard, wrapped)
	s.Require().NoError(err)
	out := buf.String()
	s.Contains(out, "Downloading v1.1.0")
	s.Contains(out, "100%")
	s.Contains(out, "#")
	s.Equal("1.0 KB", formatBytes(1024))
	s.Equal("1.0 MB", formatBytes(1<<20))

	closer, ok := wrapped.(io.Closer)
	s.Require().True(ok)
	s.NoError(closer.Close())
	s.Contains(buf.String(), "\033[?25h")
	s.NoError(bar.Finish())
}

func (s *DownloadBarSuite) TestFillFollowsColor() {
	payload := bytes.Repeat([]byte("a"), 50)

	var off bytes.Buffer
	plain := NewDownloadBar(&off)
	plain.tty, plain.color = true, false
	_, err := io.Copy(io.Discard, plain.Wrap(bytes.NewReader(payload), 50, "v1.2.0"))
	s.Require().NoError(err)
	s.Contains(off.String(), "#")
	s.NotContains(off.String(), "█")

	var on bytes.Buffer
	color := NewDownloadBar(&on)
	color.tty, color.color = true, true
	r := lipgloss.NewRenderer(&on)
	r.SetColorProfile(termenv.TrueColor)
	color.style = lipgloss.NewStyle().
		Foreground(lipgloss.Color(cliout.BrandGreen)).
		Bold(true).
		Renderer(r)
	_, err = io.Copy(io.Discard, color.Wrap(bytes.NewReader(payload), 50, "v1.2.0"))
	s.Require().NoError(err)
	s.Contains(on.String(), "█")
	s.Contains(on.String(), "\x1b[")
}

func TestDownloadBarSuite(t *testing.T) {
	suite.Run(t, new(DownloadBarSuite))
}
