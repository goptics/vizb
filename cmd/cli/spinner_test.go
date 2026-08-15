package cli

import (
	"bytes"
	"io"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sys/unix"
)

type SpinnerSuite struct {
	suite.Suite
}

func (s *SpinnerSuite) TestFinishOnNonTTYIsNoop() {
	var buf bytes.Buffer
	sp := NewGreenSpinner(&buf)
	sp.Describe("Collected 3 records - BenchmarkX")
	s.NoError(sp.Finish())
	s.False(sp.tty)
	s.Empty(buf.String())
}

func (s *SpinnerSuite) TestDescribeAndPhrases() {
	var buf bytes.Buffer
	sp := NewGreenSpinner(&buf)
	sp.Describe("Collected 3 records - BenchmarkX")
	sp.mu.Lock()
	s.Equal("Collected 3 records - BenchmarkX", sp.detail)
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

func (s *SpinnerSuite) TestGradientMovesLeftToRight() {
	n := len(greenGradientStops)
	s.Equal(0, gradientStopIndex(0, 0, n))
	// The color on rune 0 at phase 0 is on rune 1 at phase 1 (band walks right).
	s.Equal(gradientStopIndex(0, 0, n), gradientStopIndex(1, 1, n))
	s.Equal(gradientStopIndex(0, 0, n), gradientStopIndex(2, 2, n))
	s.NotEqual(gradientStopIndex(0, 0, n), gradientStopIndex(0, 1, n))
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

func (s *SpinnerSuite) TestNewGreenSpinnerNilWriterDefaultsToStderr() {
	// nil writer → os.Stderr; must not panic even when stderr is not a TTY.
	sp := NewGreenSpinner(nil)
	s.NotNil(sp)
	s.NotNil(sp.w)
	s.NoError(sp.Finish())
}

// forcedTTYSpinner builds a spinner that looks like a live TTY without
// spawning startSpinner's real loop (no hung goroutines).
func (s *SpinnerSuite) forcedTTYSpinner(w io.Writer, color bool) *GreenSpinner {
	if w == nil {
		w = &bytes.Buffer{}
	}
	sp := &GreenSpinner{
		w:       w,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
		tty:     true,
		color:   color,
		active:  true,
		phrases: []string{"Parsing", "Decoding", "Building"},
		phrase:  "Parsing",
		rng:     rand.New(rand.NewSource(1)),
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3BA272")).
			Bold(true),
	}
	if color {
		r := lipgloss.NewRenderer(w)
		r.SetColorProfile(termenv.TrueColor)
		sp.render = r
		sp.style = sp.style.Renderer(r)
	}
	// Non-loop path: close done immediately so Finish can wait without hang.
	close(sp.done)
	return sp
}

func (s *SpinnerSuite) TestPaintActiveWritesLine() {
	var buf bytes.Buffer
	sp := s.forcedTTYSpinner(&buf, false)
	sp.detail = "Collected 2 records - BenchmarkX"
	sp.paint()
	out := buf.String()
	s.Contains(out, "\r\033[K")
	s.Contains(out, "Parsing")
	s.Contains(out, "Collected 2 records - BenchmarkX")
	s.Contains(out, "\n")
	s.Contains(out, "> ")
	s.NotContains(out, "Parsing · ")
	s.Contains(out, "\033[?25l")
	s.True(bytes.HasSuffix([]byte(out), []byte("\033[1A")), "cursor must park on the activity row")
	s.NoError(sp.Finish())
	s.Contains(buf.String(), "\033[?25h")
}

func (s *SpinnerSuite) TestPaintWithoutDetailIsSingleRow() {
	var buf bytes.Buffer
	sp := s.forcedTTYSpinner(&buf, false)
	sp.paint()
	out := buf.String()
	s.Contains(out, "\r\033[K")
	s.Contains(out, "Parsing")
	s.NotContains(out, "\n")
	s.NotContains(out, ">")
	s.NoError(sp.Finish())
}

func (s *SpinnerSuite) TestPaintClearsSecondRowWhenDetailRemoved() {
	var buf bytes.Buffer
	sp := s.forcedTTYSpinner(&buf, false)
	sp.detail = "Collected 1 records - BenchmarkX"
	sp.paint()
	sp.detail = ""
	sp.paint()
	out := buf.String()
	s.Contains(out, "\033[1A")
	s.NoError(sp.Finish())
}

func (s *SpinnerSuite) TestPaintInactiveIsNoop() {
	var buf bytes.Buffer
	sp := s.forcedTTYSpinner(&buf, false)
	sp.active = false
	sp.paint()
	s.Empty(buf.String())
	s.NoError(sp.Finish())
}

func (s *SpinnerSuite) TestPaintWithColorUsesGradient() {
	var buf bytes.Buffer
	sp := s.forcedTTYSpinner(&buf, true)
	sp.detail = "3 records"
	sp.paint()
	out := buf.String()
	// Gradient paints each rune with its own SGR sequence, so the bare word
	// "Parsing" is not contiguous; assert ANSI + a frame glyph instead.
	s.Contains(out, "\x1b[")
	s.Contains(out, "\r\033[K")
	s.Contains(out, "P")
	s.Contains(out, "g") // last letter of Parsing
	s.NoError(sp.Finish())
}

func (s *SpinnerSuite) TestDescribeOnTTYPaints() {
	var buf bytes.Buffer
	sp := s.forcedTTYSpinner(&buf, false)
	sp.Describe("detail-suffix")
	s.Equal("detail-suffix", sp.detail)
	s.Contains(buf.String(), "detail-suffix")
	s.NoError(sp.Finish())
}

func (s *SpinnerSuite) TestFinishClearsWhenWasActive() {
	var buf bytes.Buffer
	// Build spinner with open done channel; Finish closes stop and waits on done.
	// Simulate loop exiting by closing done after stop.
	sp := &GreenSpinner{
		w:       &buf,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
		tty:     true,
		active:  true,
		phrases: []string{"Reading"},
		phrase:  "Reading",
		rng:     rand.New(rand.NewSource(1)),
		style:   lipgloss.NewStyle(),
	}
	// Mimic a short-lived loop: wait for stop, then signal done.
	go func() {
		<-sp.stop
		close(sp.done)
	}()
	s.NoError(sp.Finish())
	s.Contains(buf.String(), "\r\033[K")
	s.NotContains(buf.String(), "\033[1A")
	// Second Finish is idempotent.
	s.NoError(sp.Finish())
}

func (s *SpinnerSuite) TestFinishClearsTwoRows() {
	var buf bytes.Buffer
	sp := &GreenSpinner{
		w:       &buf,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
		tty:     true,
		active:  true,
		rows:    2,
		phrases: []string{"Reading"},
		phrase:  "Reading",
		rng:     rand.New(rand.NewSource(1)),
		style:   lipgloss.NewStyle(),
	}
	go func() {
		<-sp.stop
		close(sp.done)
	}()
	s.NoError(sp.Finish())
	out := buf.String()
	s.Contains(out, "\033[1A")
	s.Equal(2, bytes.Count([]byte(out), []byte("\033[K")))
}

func (s *SpinnerSuite) TestRotatePhraseLockedSkipsCurrent() {
	sp := &GreenSpinner{
		phrases: []string{"A", "B"},
		phrase:  "A",
		rng:     rand.New(rand.NewSource(0)), // deterministic
	}
	// Force many rotations; never stay on the same phrase when len>1.
	for i := 0; i < 20; i++ {
		prev := sp.phrase
		sp.rotatePhraseLocked()
		s.NotEqual(prev, sp.phrase, "iteration %d", i)
		s.Contains(sp.phrases, sp.phrase)
	}
}

func (s *SpinnerSuite) TestLoopPaintsThenStops() {
	var buf bytes.Buffer
	sp := &GreenSpinner{
		w:       &buf,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
		tty:     true,
		color:   false,
		active:  false,
		delay:   0,
		phrases: []string{"Scooping", "Folding"},
		phrase:  "Scooping",
		rng:     rand.New(rand.NewSource(42)),
		style:   lipgloss.NewStyle(),
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		sp.loop()
	}()

	// Wait until paint has run at least once.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if buf.Len() > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	s.Greater(buf.Len(), 0, "loop should paint at least once")

	// Advance past a phrase rotation tick if possible, then stop cleanly.
	time.Sleep(spinnerTick + 20*time.Millisecond)
	close(sp.stop)
	wg.Wait()

	out := buf.String()
	s.Contains(out, "\r\033[K")
	s.Contains(out, "Scooping")
}

func (s *SpinnerSuite) TestLoopRespectsDelayAndStopBeforeActive() {
	var buf bytes.Buffer
	sp := &GreenSpinner{
		w:       &buf,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
		tty:     true,
		delay:   5 * time.Second,
		phrases: []string{"Parsing"},
		phrase:  "Parsing",
		rng:     rand.New(rand.NewSource(1)),
		style:   lipgloss.NewStyle(),
	}

	done := make(chan struct{})
	go func() {
		sp.loop()
		close(done)
	}()

	// Stop before delay elapses — loop should exit without painting.
	time.Sleep(20 * time.Millisecond)
	close(sp.stop)
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		s.Fail("loop did not exit after stop during delay")
	}
	s.Empty(buf.String())
	s.False(sp.active)
}

func (s *SpinnerSuite) TestLoopExitsWhenFinishedBeforeActive() {
	var buf bytes.Buffer
	sp := &GreenSpinner{
		w:        &buf,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
		tty:      true,
		delay:    0,
		finished: true, // set before loop becomes active
		phrases:  []string{"Parsing"},
		phrase:   "Parsing",
		rng:      rand.New(rand.NewSource(1)),
		style:    lipgloss.NewStyle(),
	}

	sp.loop() // synchronous; should return immediately
	s.Empty(buf.String())
	s.False(sp.active)
}

func (s *SpinnerSuite) TestLoopRotatesPhraseOnTickCadence() {
	var buf bytes.Buffer
	sp := &GreenSpinner{
		w:       &buf,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
		tty:     true,
		delay:   0,
		phrases: []string{"Alpha", "Beta", "Gamma"},
		phrase:  "Alpha",
		// Start just below rotation threshold so the next tick rotates.
		tick:  phraseEveryTicks - 1,
		rng:   rand.New(rand.NewSource(7)),
		style: lipgloss.NewStyle(),
	}

	go sp.loop()

	// Wait for active paint, then one more tick for rotation.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		sp.mu.Lock()
		active := sp.active
		sp.mu.Unlock()
		if active {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Allow at least one ticker fire.
	time.Sleep(spinnerTick + 50*time.Millisecond)

	sp.mu.Lock()
	phrase := sp.phrase
	tick := sp.tick
	sp.mu.Unlock()

	close(sp.stop)
	<-sp.done

	s.GreaterOrEqual(tick, phraseEveryTicks)
	s.Contains([]string{"Alpha", "Beta", "Gamma"}, phrase)
}

func (s *SpinnerSuite) TestStartSpinnerColorPath() {
	// Force a color profile via a custom writer that is not a TTY — color is
	// false for buffers. Exercise the color branch by constructing via
	// startSpinner then enabling color+render manually and painting.
	var buf bytes.Buffer
	sp := startSpinner(&buf, 0, []string{"Mapping", "Shaping"})
	s.False(sp.tty)
	s.False(sp.color)

	// Promote to "color TTY" for paint coverage without a real terminal.
	sp.mu.Lock()
	sp.tty = true
	sp.color = true
	sp.active = true
	r := lipgloss.NewRenderer(&buf)
	r.SetColorProfile(termenv.ANSI256)
	sp.render = r
	sp.style = sp.style.Renderer(r)
	sp.mu.Unlock()

	sp.paint()
	s.Contains(buf.String(), "\x1b[")
	s.NoError(sp.Finish())
}

func (s *SpinnerSuite) TestStartSpinnerOnPTYEnablesColorAndLoop() {
	master, slave, err := openTestPTY()
	if err != nil {
		s.T().Skipf("no PTY available: %v", err)
	}
	defer master.Close()
	defer slave.Close()

	s.T().Setenv("NO_COLOR", "")
	s.T().Setenv("CLICOLOR", "1")
	s.T().Setenv("CLICOLOR_FORCE", "1")
	s.T().Setenv("COLORTERM", "truecolor")

	sp := startSpinner(slave, 0, []string{"Wiring", "Folding"})
	s.True(sp.tty)
	s.True(sp.color)
	s.NotNil(sp.render)

	// Wait for at least one paint, then finish cleanly (stops loop).
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		sp.mu.Lock()
		active := sp.active
		sp.mu.Unlock()
		if active {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	s.NoError(sp.Finish())
}

func (s *SpinnerSuite) TestLoopExitsOnInactiveDuringTick() {
	var buf bytes.Buffer
	sp := &GreenSpinner{
		w:       &buf,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
		tty:     true,
		delay:   0,
		phrases: []string{"Sifting"},
		phrase:  "Sifting",
		rng:     rand.New(rand.NewSource(3)),
		style:   lipgloss.NewStyle(),
	}

	go sp.loop()

	// Wait until active, then clear active without closing stop so the next
	// tick takes the finished/!active return path.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		sp.mu.Lock()
		active := sp.active
		sp.mu.Unlock()
		if active {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	sp.mu.Lock()
	sp.active = false
	sp.finished = true
	sp.mu.Unlock()

	select {
	case <-sp.done:
	case <-time.After(2 * time.Second):
		close(sp.stop) // fail-safe so we do not hang the suite
		s.Fail("loop did not exit after active cleared")
		return
	}
	// stop is still open; loop already exited — close for cleanliness.
	close(sp.stop)
}

func TestSpinnerSuite(t *testing.T) {
	suite.Run(t, new(SpinnerSuite))
}

// openTestPTY opens a Linux PTY pair for startSpinner TTY/color branches.
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
