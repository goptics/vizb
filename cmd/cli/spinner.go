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
)

// GreenSpinner is a stderr activity spinner. Implements ProgressBar so
// DataProgressManager can drive a detail suffix while the verb phrase rotates
// and the line shimmers green.
type GreenSpinner struct {
	mu       sync.Mutex
	w        io.Writer
	detail   string
	phrase   string
	phrases  []string
	frame    int
	tick     int
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

// Describe sets the optional detail suffix (benchmark name, record count).
func (s *GreenSpinner) Describe(detail string) {
	s.mu.Lock()
	s.detail = detail
	s.mu.Unlock()
	if s.tty {
		s.paint()
	}
}

// Finish stops the spinner and clears the activity line when it was visible.
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
		_, _ = fmt.Fprint(s.w, "\r\033[K")
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
	// Width(2) keeps glyph + phrase baseline aligned in mono fonts.
	glyph := s.style.Width(2).Render(spinnerFrames[s.frame])
	line := s.phrase
	if s.detail != "" {
		line = s.phrase + " · " + s.detail
	}
	if s.color {
		line = gradientText(line, s.tick, s.render)
	}
	_, _ = fmt.Fprintf(s.w, "\r\033[K%s%s", glyph, line)
}

func gradientText(text string, phase int, r *lipgloss.Renderer) string {
	if text == "" {
		return text
	}
	var b strings.Builder
	b.Grow(len(text) * 12)
	i := 0
	for _, ch := range text {
		stop := (i + phase) % len(greenGradientStops)
		st := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(greenGradientStops[stop]))
		if r != nil {
			st = st.Renderer(r)
		}
		b.WriteString(st.Render(string(ch)))
		i++
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
