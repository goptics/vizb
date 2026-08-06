// Package cliout configures charmbracelet/log for vizb's human-facing CLI
// diagnostics: leveled text on stderr, no timestamps, no caller frames.
package cliout

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

// Brand / level colors (logo palette + clear warn yellow).
const (
	// BrandGreen is the logo green for info `>` marks and activity spinners.
	BrandGreen  = "#3BA272"
	markYellow  = "#F1FA8C" // warn `>` and warn body
	markRed     = "#EE6666" // error `>` and error body
	valueAccent = "#5470C6" // Pair values only
	// AccentCount tints numbers in status prose (e.g. aggregation counts).
	AccentCount = "#F5A623"
	accentHTML  = "#E34F26" // HTML5
	accentJSON  = AccentCount
)

// clearLine wipes the current terminal row (carriage return + erase to end).
// Spinners paint in place with \r and no trailing \n; without this, the next
// log record is appended on the same line ("◐ Mapping fields> Auto-grouped…").
const clearLine = "\r\033[K"

// Rendering state set by Configure; Accent reuses this instead of re-probing
// os.Stderr so test buffers and forced profiles stay consistent with the logger.
var (
	accentProfile termenv.Profile
	accentRender  *lipgloss.Renderer
)

// stderrWriter always writes to the current os.Stderr so test helpers that
// swap os.Stderr still observe log output.
//
// Each write clears any in-place activity line first so `>` logs land on their
// own row; the spinner's next tick redraws below that log with \r.
type stderrWriter struct{}

func (stderrWriter) Write(p []byte) (int, error) {
	if _, err := io.WriteString(os.Stderr, clearLine); err != nil {
		return 0, err
	}
	return os.Stderr.Write(p)
}

func init() {
	Configure(nil)
}

// Configure sets the process-wide default logger used by Info/Warn/Error.
// Pass a custom writer from tests when you need an isolated buffer; production
// uses a writer that always targets the current os.Stderr.
//
// Levels render as a single colored `>` mark (no INFO/WARN/ERROR badges).
// NO_COLOR / CLICOLOR=0 / non-TTY disable color.
func Configure(w io.Writer) {
	if w == nil {
		w = stderrWriter{}
	}
	profile := ColorProfile(w)
	accentProfile = profile
	// Renderer target: live stderr for the production wrapper; otherwise w so
	// Accent matches the configured sink (Ascii on pipes/buffers).
	renderW := io.Writer(w)
	if _, ok := w.(stderrWriter); ok {
		renderW = os.Stderr
	}
	accentRender = lipgloss.NewRenderer(renderW)
	accentRender.SetColorProfile(profile)

	logger := log.NewWithOptions(w, log.Options{
		Level:           log.InfoLevel,
		ReportTimestamp: false,
		ReportCaller:    false,
		Formatter:       log.TextFormatter,
	})
	logger.SetStyles(vizbStyles())
	logger.SetColorProfile(profile)
	log.SetDefault(logger)
}

func levelMark(color string) lipgloss.Style {
	return lipgloss.NewStyle().
		SetString(">").
		Bold(true).
		Foreground(lipgloss.Color(color))
}

func vizbStyles() *log.Styles {
	s := log.DefaultStyles()
	s.Levels[log.InfoLevel] = levelMark(BrandGreen)
	s.Levels[log.WarnLevel] = levelMark(markYellow)
	s.Levels[log.ErrorLevel] = levelMark(markRed)
	s.Message = lipgloss.NewStyle()
	return s
}

// ColorProfile decides whether to emit ANSI colors for w.
//
// Rules:
//  1. NO_COLOR set or CLICOLOR=0 (without CLICOLOR_FORCE) → Ascii
//  2. Writer is not a TTY (non-*os.File, or non-terminal file) → Ascii
//  3. Else ANSI256, or TrueColor when COLORTERM advertises it
//
// stderrWriter is not an *os.File; Configure probes the live os.Stderr instead.
func ColorProfile(w io.Writer) termenv.Profile {
	f, ok := w.(*os.File)
	if !ok {
		if _, isStderrWriter := w.(stderrWriter); isStderrWriter {
			f = os.Stderr
		} else {
			return termenv.Ascii
		}
	}
	if termenv.NewOutput(f).EnvNoColor() {
		return termenv.Ascii
	}
	if !term.IsTerminal(int(f.Fd())) {
		return termenv.Ascii
	}
	switch os.Getenv("COLORTERM") {
	case "truecolor", "24bit":
		return termenv.TrueColor
	}
	return termenv.ANSI256
}

// Pair formats "key: value" with only the value colored (default brand blue).
func Pair(key, value string) string {
	return PairAccent(key, value, valueAccent)
}

// PairAccent is Pair with an explicit value color (hex, e.g. parser brands).
func PairAccent(key, value, accent string) string {
	if value == "" {
		return key
	}
	if accent == "" {
		accent = valueAccent
	}
	return key + ": " + Accent(value, accent)
}

// ParserAccent maps a registered parser key to a language/format brand color
// used when printing "Auto-detected parser: …".
//
//	go          → Go cyan
//	js:*        → JavaScript yellow
//	rs:*        → Rust orange
//	csv         → spreadsheet green
//	json        → JSON amber
func ParserAccent(parser string) string {
	switch {
	case parser == "go":
		return "#00ADD8"
	case parser == "csv":
		return "#217346"
	case parser == "json":
		return accentJSON
	case parser == "js" || strings.HasPrefix(parser, "js:"):
		return "#F0DB4F"
	case parser == "rs" || strings.HasPrefix(parser, "rs:"):
		return "#DEA584"
	default:
		return valueAccent
	}
}

// FormatAccent maps an output format ("html" | "json") to its brand color for
// path values like "Output file: …".
func FormatAccent(format string) string {
	switch strings.ToLower(format) {
	case "json":
		return accentJSON
	case "html":
		return accentHTML
	default:
		return valueAccent
	}
}

// Accent tints s with hex when Configure's profile allows color; plain otherwise.
func Accent(s, color string) string {
	if s == "" || accentProfile == termenv.Ascii || accentRender == nil {
		return s
	}
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(color)).
		Renderer(accentRender).
		Render(s)
}

// Info logs an informational status message (plain body; green `>` mark).
func Info(msg string) {
	log.Info(msg)
}

// InfoPair logs "key: value" with only value colored (default brand blue).
func InfoPair(key, value string) {
	log.Info(Pair(key, value))
}

// InfoPairAccent logs "key: value" with the value in an explicit hex color.
func InfoPairAccent(key, value, accent string) {
	log.Info(PairAccent(key, value, accent))
}

// Warn logs a warning (yellow `>` and yellow body).
func Warn(msg string) {
	log.Warn(Accent(msg, markYellow))
}

// Warnf logs a formatted warning with yellow body.
func Warnf(format string, args ...any) {
	log.Warn(Accent(fmt.Sprintf(format, args...), markYellow))
}

// Error logs an error without exiting (red `>` and red body). Callers that must
// exit should use shared.ExitWithError so temp-file cleanup still runs.
func Error(msg string) {
	log.Error(Accent(msg, markRed))
}
