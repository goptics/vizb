package style

import "github.com/charmbracelet/lipgloss"

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
