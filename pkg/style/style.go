package style

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/goptics/vizb/internal/specparse"
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
	if _, ok := themeCatalog[strings.ToLower(trimmed)]; ok {
		return strings.ToLower(trimmed)
	}
	// Prefer shared tokenizer when the value is a structured theme spec.
	if strings.Contains(trimmed, ":") {
		if parsed, err := specparse.Parse(trimmed, themeSpecParseOptions(trimmed)); err == nil {
			return normalizeStructuredSpec(parsed)
		}
		// Best-effort fallback for structured-looking input that fails tokenization.
		if name, props, ok := splitStructured(trimmed); ok {
			return normalizeStructured(name, props)
		}
	}
	// Bare hex (or anything comma-separated): trim each segment.
	colors := strings.Split(trimmed, ",")
	for i := range colors {
		colors[i] = strings.TrimSpace(colors[i])
	}
	return strings.Join(colors, ",")
}

func normalizeStructuredSpec(parsed specparse.Spec) string {
	parts := make([]string, 0, len(parsed.Props))
	for _, prop := range parsed.Props {
		key := prop.Key
		switch strings.ToLower(key) {
		case "colors":
			key = "colors"
		case "visualmapcolors":
			key = "visualMapColors"
		}
		if !prop.HasValue {
			parts = append(parts, key)
			continue
		}
		segments := strings.Split(prop.Value, ",")
		for i := range segments {
			segments[i] = strings.TrimSpace(segments[i])
		}
		parts = append(parts, fmt.Sprintf("%s=%s", key, strings.Join(segments, ",")))
	}
	return parsed.Prefix + ":" + strings.Join(parts, ";")
}

// normalizeStructured is a best-effort semicolon re-serializer used when
// specparse.Parse fails but the input still looks structured.
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
