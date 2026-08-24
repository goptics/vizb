package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/goptics/vizb/pkg/cliout"
	"github.com/muesli/termenv"
)

const downloadBarWidth = 24

// DownloadBar is a TTY-only determinate download progress line on stderr.
type DownloadBar struct {
	mu        sync.Mutex
	w         io.Writer
	tty       bool
	color     bool
	finished  bool
	painted   bool
	lastPaint time.Time
	written   int64
	total     int64
	label     string
	style     lipgloss.Style
}

// NewDownloadBar paints download progress on w when w is a TTY.
func NewDownloadBar(w io.Writer) *DownloadBar {
	profile := cliout.ColorProfile(w)
	b := &DownloadBar{
		w:     w,
		tty:   writerIsTerminal(w),
		color: profile != termenv.Ascii,
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color(cliout.BrandGreen)).
			Bold(true),
	}
	if b.color {
		r := lipgloss.NewRenderer(w)
		r.SetColorProfile(profile)
		b.style = b.style.Renderer(r)
	}
	return b
}

// IsTTY reports whether the bar will paint (stderr is a terminal).
func (b *DownloadBar) IsTTY() bool {
	return b.tty
}

// Wrap reports read progress for r. Non-TTY writers and unknown totals return r unchanged.
func (b *DownloadBar) Wrap(r io.Reader, total int64, label string) io.Reader {
	if !b.tty || total <= 0 {
		return r
	}
	b.mu.Lock()
	b.total = total
	b.label = label
	b.written = 0
	b.mu.Unlock()
	return &progressReader{bar: b, r: r}
}

// Finish clears the progress line when it was painted. Idempotent.
func (b *DownloadBar) Finish() error {
	b.mu.Lock()
	if b.finished {
		b.mu.Unlock()
		return nil
	}
	b.finished = true
	painted := b.painted
	b.mu.Unlock()
	if painted {
		_, _ = fmt.Fprint(b.w, "\r\033[K"+ansiShowCursor)
	}
	return nil
}

type progressReader struct {
	bar *DownloadBar
	r   io.Reader
}

func (p *progressReader) Read(buf []byte) (int, error) {
	n, err := p.r.Read(buf)
	if n > 0 {
		p.bar.add(int64(n))
	}
	if err == io.EOF {
		p.bar.paint(true)
	}
	return n, err
}

func (p *progressReader) Close() error {
	return p.bar.Finish()
}

func (b *DownloadBar) add(n int64) {
	b.mu.Lock()
	b.written += n
	complete := b.written >= b.total
	b.mu.Unlock()
	b.paint(complete)
}

func (b *DownloadBar) paint(final bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !final && b.written < b.total && !b.lastPaint.IsZero() && time.Since(b.lastPaint) < spinnerTick {
		return
	}

	_, _ = fmt.Fprint(b.w, ansiHideCursor)
	_, _ = fmt.Fprintf(b.w, "\r\033[K%s", b.lineLocked())
	b.painted = true
	b.lastPaint = time.Now()
}

func (b *DownloadBar) lineLocked() string {
	var sb strings.Builder
	sb.WriteString("Downloading")
	if b.label != "" {
		sb.WriteByte(' ')
		sb.WriteString(b.label)
	}
	sb.WriteString("  ")
	sb.WriteString(b.renderBarLocked())
	sb.WriteString("  ")
	pct := int(float64(b.written) / float64(b.total) * 100)
	if b.written >= b.total {
		pct = 100
	}
	sb.WriteString(strconv.Itoa(pct))
	sb.WriteString("%  ")
	sb.WriteString(formatBytes(b.written))
	sb.WriteByte('/')
	sb.WriteString(formatBytes(b.total))
	return sb.String()
}

func (b *DownloadBar) renderBarLocked() string {
	filled := min(downloadBarWidth, int(float64(downloadBarWidth)*float64(b.written)/float64(b.total)))
	empty := downloadBarWidth - filled
	if b.color {
		return b.style.Render(strings.Repeat("█", filled)) + strings.Repeat("░", empty)
	}
	return strings.Repeat("#", filled) + strings.Repeat(" ", empty)
}

func formatBytes(n int64) string {
	const (
		kb = 1024.0
		mb = 1024.0 * 1024.0
	)
	switch {
	case n >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(n)/mb)
	case n >= 1024:
		return fmt.Sprintf("%.1f KB", float64(n)/kb)
	default:
		return fmt.Sprintf("%d B", n)
	}
}
