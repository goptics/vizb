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
	builtInThemes = map[string]struct{}{
		"default": {}, "vintage": {}, "meadow": {}, "westeros": {}, "essos": {},
		"wonderland": {}, "walden": {}, "chalk": {}, "infographic": {},
		"macarons": {}, "roma": {}, "shine": {}, "purple-passion": {},
	}
	hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{3}(?:[0-9a-fA-F]{3})?$`)
)

// ValidateTheme accepts a built-in name or a custom palette containing at
// least two comma-separated #rgb/#rrggbb colors.
func ValidateTheme(value string) error {
	if _, ok := builtInThemes[strings.ToLower(value)]; ok {
		return nil
	}
	colors := strings.Split(value, ",")
	if len(colors) < 2 {
		return fmt.Errorf("expected a built-in theme or at least two comma-separated hex colors")
	}
	for _, color := range colors {
		if !hexColorPattern.MatchString(strings.TrimSpace(color)) {
			return fmt.Errorf("invalid hex color %q", strings.TrimSpace(color))
		}
	}
	return nil
}

// NormalizeTheme canonicalizes built-in names and custom palette whitespace.
func NormalizeTheme(value string) string {
	trimmed := strings.TrimSpace(value)
	if _, ok := builtInThemes[strings.ToLower(trimmed)]; ok {
		return strings.ToLower(trimmed)
	}
	colors := strings.Split(trimmed, ",")
	for i := range colors {
		colors[i] = strings.TrimSpace(colors[i])
	}
	return strings.Join(colors, ",")
}
