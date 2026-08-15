package cli

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/goptics/vizb/pkg/cliout"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

// Spinner frames chosen for optical vertical alignment with Latin text in
// common mono fonts. Braille (⠋⠙…) sits high on the cell and looks offset
// next to status copy; these share the Latin midline better.
var spinnerFrames = []string{"●", "◕", "◑", "◒", "◐", "◓", "○", "◉"}

// activityPhrases rotate on the spinner line while stdin is being read.
// Single words keep the live line dense and easy to scan.
var activityPhrases = []string{
	"Collecting",
	"Ingesting",
	"Gathering",
	"Reading",
	"Scooping",
	"Absorbing",
	"Harvesting",
	"Sifting",
	"Catching",
	"Sniffing",
	"Wiring",
	"Folding",
}

// parsePhrases rotate while decoding input into data points.
var parsePhrases = []string{
	"Parsing",
	"Decoding",
	"Building",
	"Shaping",
	"Mapping",
	"Reading",
}

// aggregatePhrases rotate while collapsing / summing grouped rows.
var aggregatePhrases = []string{
	"Aggregating",
	"Summing",
	"Collapsing",
	"Rolling",
	"Merging",
	"Grouping",
}

// greenGradientStops is a mid-density band around brand green.
var greenGradientStops = []string{
	"#2A8A5C",
	"#329A68",
	cliout.BrandGreen,
	"#48B07E",
	"#58C08C",
	"#68D09A",
	"#78E0A8",
	"#68D09A",
	"#58C08C",
	"#48B07E",
	cliout.BrandGreen,
	"#329A68",
}

// parseSpinnerDelay is how long a phase may block before the activity line
// appears. Fast work stays silent (no flash).
const parseSpinnerDelay = 250 * time.Millisecond

const (
	spinnerTick      = 100 * time.Millisecond
	phraseEveryTicks = 20 // 20 * 100ms = 2.0s
	ansiHideCursor   = "\033[?25l"
	ansiShowCursor   = "\033[?25h"
	ansiCursorUp     = "\033[1A"
)

// GreenSpinner is a stderr activity spinner. Implements ProgressBar so
// DataProgressManager can drive a second-row detail while the verb phrase
// rotates and shimmers green on its own row.
type GreenSpinner struct {
	mu       sync.Mutex
	w        io.Writer
	detail   string
	phrase   string
	phrases  []string
	frame    int
	tick     int
	rows     int // 1 or 2; last painted height
	stop     chan struct{}
	done     chan struct{}
	active   bool // painting allowed (after Delay)
	finished bool
	tty      bool
	color    bool
	delay    time.Duration
	style    lipgloss.Style
	render   *lipgloss.Renderer
	rng      *rand.Rand
}

// NewGreenSpinner starts an immediate green spinner for stdin pipe loading.
func NewGreenSpinner(w io.Writer) *GreenSpinner {
	return startSpinner(w, 0, activityPhrases)
}

// NewParseSpinner starts a delayed green spinner for the parse phase.
func NewParseSpinner(w io.Writer) *GreenSpinner {
	return startSpinner(w, parseSpinnerDelay, parsePhrases)
}

// NewAggregateSpinner starts a delayed green spinner for the aggregation phase.
func NewAggregateSpinner(w io.Writer) *GreenSpinner {
	return startSpinner(w, parseSpinnerDelay, aggregatePhrases)
}

func startSpinner(w io.Writer, delay time.Duration, phrases []string) *GreenSpinner {
	if w == nil {
		w = os.Stderr
	}

	profile := cliout.ColorProfile(w)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := &GreenSpinner{
		w:       w,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
		tty:     writerIsTerminal(w),
		color:   profile != termenv.Ascii,
		rng:     rng,
		phrases: phrases,
		delay:   delay,
		phrase:  phrases[rng.Intn(len(phrases))],
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cliout.BrandGreen)).
			Bold(true),
	}
	if s.color {
		s.render = lipgloss.NewRenderer(w)
		s.render.SetColorProfile(profile)
		s.style = s.style.Renderer(s.render)
	}
	if s.tty {
		go s.loop()
	} else {
		close(s.done)
	}
	return s
}

// Describe sets the optional second-row detail (benchmark name, record count).
func (s *GreenSpinner) Describe(detail string) {
	s.mu.Lock()
	s.detail = detail
	s.mu.Unlock()
	if s.tty {
		s.paint()
	}
}

// Finish stops the spinner and clears the activity line(s) when they were visible.
func (s *GreenSpinner) Finish() error {
	s.mu.Lock()
	if s.finished {
		s.mu.Unlock()
		return nil
	}
	s.finished = true
	wasActive := s.active
	s.active = false
	s.mu.Unlock()

	close(s.stop)
	<-s.done
	if wasActive {
		s.mu.Lock()
		rows := s.rows
		s.mu.Unlock()
		// After a 2-row paint the cursor is parked on the activity row.
		if rows == 2 {
			_, _ = fmt.Fprint(s.w, "\r\033[K\n\033[K"+ansiCursorUp)
		} else {
			_, _ = fmt.Fprint(s.w, "\r\033[K")
		}
		_, _ = fmt.Fprint(s.w, ansiShowCursor)
	}
	return nil
}

func (s *GreenSpinner) loop() {
	defer close(s.done)

	if s.delay > 0 {
		select {
		case <-s.stop:
			return
		case <-time.After(s.delay):
		}
	}

	s.mu.Lock()
	if s.finished {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.mu.Unlock()
	s.paint()

	t := time.NewTicker(spinnerTick)
	defer t.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-t.C:
			s.mu.Lock()
			if s.finished || !s.active {
				s.mu.Unlock()
				return
			}
			s.frame = (s.frame + 1) % len(spinnerFrames)
			s.tick++
			if s.tick%phraseEveryTicks == 0 {
				s.rotatePhraseLocked()
			}
			s.mu.Unlock()
			s.paint()
		}
	}
}

func (s *GreenSpinner) rotatePhraseLocked() {
	next := s.rng.Intn(len(s.phrases))
	if s.phrases[next] == s.phrase {
		next = (next + 1) % len(s.phrases)
	}
	s.phrase = s.phrases[next]
}

func (s *GreenSpinner) paint() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.active {
		return
	}

	_, _ = fmt.Fprint(s.w, ansiHideCursor)

	// Width(2) keeps glyph + phrase baseline aligned in mono fonts.
	glyph := s.style.Width(2).Render(spinnerFrames[s.frame])
	phrase := s.phrase
	if s.color {
		phrase = gradientText(phrase, s.tick, s.render)
	}
	_, _ = fmt.Fprintf(s.w, "\r\033[K%s%s", glyph, phrase)

	if s.detail != "" {
		mark := ">"
		if s.color {
			mark = s.style.Render(">")
		}
		// Park on the activity row so the cursor never sits on "> …" between ticks.
		_, _ = fmt.Fprintf(s.w, "\n\033[K%s %s%s", mark, s.detail, ansiCursorUp)
		s.rows = 2
		return
	}

	if s.rows == 2 {
		// Cursor is already on row 1. Clear leftover row 2 and stay put.
		_, _ = fmt.Fprint(s.w, "\n\033[K"+ansiCursorUp)
	}
	s.rows = 1
}

// gradientStopIndex maps rune i at animation phase to a stop. Subtracting
// phase walks the highlight toward higher indices (left → right).
func gradientStopIndex(i, phase, n int) int {
	return ((i-phase)%n + n) % n
}

func gradientText(text string, phase int, r *lipgloss.Renderer) string {
	if text == "" {
		return text
	}
	runes := []rune(text)
	n := len(greenGradientStops)
	var b strings.Builder
	b.Grow(len(runes) * 12)
	for i, ch := range runes {
		st := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(greenGradientStops[gradientStopIndex(i, phase, n)]))
		if r != nil {
			st = st.Renderer(r)
		}
		b.WriteString(st.Render(string(ch)))
	}
	return b.String()
}

func writerIsTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}
