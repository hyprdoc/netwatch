package ui

import "github.com/charmbracelet/lipgloss"

// Column widths for the table
const (
	LocalAddrWidth  = 22
	RemoteAddrWidth = 22
	StateWidth      = 15
	PidWidth        = 8
	ProcWidth       = 20
)

// Styles for the UI components
var (
	// titleStyle is used for the application title (currently unused)
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1).
			MarginBottom(1)

	// headerStyle is used for the table header row
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	// rowStyle is used for even-numbered rows
	rowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	// altRowStyle is used for odd-numbered rows
	altRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D0D0D0")).
			Background(lipgloss.Color("#2a2a2a"))

	// selectedRowStyle is used for the currently selected row
	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#5A5AFF")).
				Bold(true)

	// State-specific styles
	establishedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true)

	listenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00BFFF")).
			Bold(true)

	closingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)

	// errorStyle is used for error messages
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Padding(1)

	// helpStyle is used for the help text at the bottom
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)
