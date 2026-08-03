package style

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Error style (Red, Bold)
	Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5555")).
		Bold(true)

	// Warning style (Yellow)
	Warning = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F1FA8C"))

	// Success style (Green)
	Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B"))

	// Info style (Cyan) - used for general info and progress bar
	Info = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8BE9FD"))
)

var (
	// builtInThemes is the set of known built-in theme names (for NormalizeTheme).
	// Full color catalogs live in themeCatalog (themes.go).
	builtInThemes = map[string]struct{}{
		"default": {}, "vintage": {}, "meadow": {}, "westeros": {}, "essos": {},
		"wonderland": {}, "walden": {}, "chalk": {}, "infographic": {},
		"macarons": {}, "roma": {}, "shine": {}, "purple-passion": {},
	}
	hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{3}(?:[0-9a-fA-F]{3})?$`)
)

// ValidateTheme accepts a built-in name, a structured theme spec, or a custom
// palette containing at least two comma-separated #rgb/#rrggbb colors.
func ValidateTheme(value string) error {
	_, err := ParseThemeSpec(value)
	return err
}

// NormalizeTheme canonicalizes built-in names, structured specs, and custom
// palette whitespace for soft-flag validation.
func NormalizeTheme(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return trimmed
	}
	if _, ok := builtInThemes[strings.ToLower(trimmed)]; ok {
		return strings.ToLower(trimmed)
	}
	if name, props, ok := splitStructured(trimmed); ok {
		return normalizeStructured(name, props)
	}
	// Bare hex (or anything comma-separated): trim each segment.
	colors := strings.Split(trimmed, ",")
	for i := range colors {
		colors[i] = strings.TrimSpace(colors[i])
	}
	return strings.Join(colors, ",")
}

func normalizeStructured(name, props string) string {
	var parts []string
	// Preserve property order from input while normalizing values.
	for _, part := range strings.Split(props, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, val, found := strings.Cut(part, "=")
		if !found {
			parts = append(parts, part)
			continue
		}
		key = strings.TrimSpace(key)
		// Canonical key casing for known props.
		switch strings.ToLower(key) {
		case "colors":
			key = "colors"
		case "visualmapcolors":
			key = "visualMapColors"
		}
		segments := strings.Split(val, ",")
		for i := range segments {
			segments[i] = strings.TrimSpace(segments[i])
		}
		parts = append(parts, fmt.Sprintf("%s=%s", key, strings.Join(segments, ",")))
	}
	return name + ":" + strings.Join(parts, ";")
}
