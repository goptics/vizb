package cli

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/suite"
)

type SpinnerSuite struct {
	suite.Suite
}

func (s *SpinnerSuite) TestFinishOnNonTTYIsNoop() {
	var buf bytes.Buffer
	sp := NewGreenSpinner(&buf)
	sp.Describe("BenchmarkX · 3 records")
	s.NoError(sp.Finish())
	s.False(sp.tty)
	s.Empty(buf.String())
}

func (s *SpinnerSuite) TestDescribeAndPhrases() {
	var buf bytes.Buffer
	sp := NewGreenSpinner(&buf)
	sp.Describe("BenchmarkX · 3 records")
	sp.mu.Lock()
	s.Equal("BenchmarkX · 3 records", sp.detail)
	s.NotEmpty(sp.phrase)
	s.NotContains(sp.phrase, "…")
	sp.mu.Unlock()
	s.NoError(sp.Finish())
}

func (s *SpinnerSuite) TestPhaseConstructors() {
	var buf bytes.Buffer
	pipe := NewGreenSpinner(&buf)
	parse := NewParseSpinner(&buf)
	agg := NewAggregateSpinner(&buf)

	s.Equal(time.Duration(0), pipe.delay)
	s.Equal(parseSpinnerDelay, parse.delay)
	s.Equal(parseSpinnerDelay, agg.delay)
	s.Contains(parse.phrases, "Parsing")
	s.Contains(agg.phrases, "Aggregating")
	s.NotContains(parse.phrases, "Aggregating")
	s.NotContains(agg.phrases, "Parsing")
	for _, list := range [][]string{activityPhrases, parsePhrases, aggregatePhrases} {
		for _, p := range list {
			s.NotContains(p, " ", "phrase must be a single word: %q", p)
		}
	}

	s.NoError(pipe.Finish())
	s.NoError(parse.Finish())
	s.NoError(agg.Finish())
}

func (s *SpinnerSuite) TestDelayedSpinnerSilentBeforeElapsed() {
	d := &discardWriter{}
	sp := startSpinner(d, 5*time.Second, parsePhrases)
	time.Sleep(50 * time.Millisecond)
	s.NoError(sp.Finish())
	s.Equal(0, d.n)
}

func (s *SpinnerSuite) TestGradientTextShiftsWithPhase() {
	r := lipgloss.NewRenderer(&bytes.Buffer{})
	r.SetColorProfile(termenv.TrueColor)
	a := gradientText("Collecting", 0, r)
	b := gradientText("Collecting", 3, r)
	s.Contains(a, "C")
	s.NotEqual(a, b)
	s.Contains(a, "\x1b[")
	s.Equal("", gradientText("", 0, nil))
}

func (s *SpinnerSuite) TestFinishIdempotent() {
	var buf bytes.Buffer
	sp := NewGreenSpinner(&buf)
	s.NoError(sp.Finish())
	s.NoError(sp.Finish())
}

type discardWriter struct{ n int }

func (d *discardWriter) Write(p []byte) (int, error) {
	d.n += len(p)
	return len(p), nil
}

func (s *SpinnerSuite) TestWriterIsTerminalNonFile() {
	s.False(writerIsTerminal(io.Discard))
}

func TestSpinnerSuite(t *testing.T) {
	suite.Run(t, new(SpinnerSuite))
}
