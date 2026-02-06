package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	rowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	altRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D0D0D0")).
			Background(lipgloss.Color("#2a2a2a"))

	establishedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true)

	listenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00BFFF")).
			Bold(true)

	closingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Padding(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

// View renders the UI
func (m Model) View() string {
	var b strings.Builder

	// Error display
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press 'q' to quit • Press 'r' to retry"))
		return b.String()
	}

	// No connections
	if len(m.connections) == 0 {
		b.WriteString("No active connections found...\n\n")
		b.WriteString(helpStyle.Render("Press 'q' to quit • Press 'r' to refresh"))
		return b.String()
	}

	// Column widths
	const (
		localAddrWidth  = 22
		remoteAddrWidth = 22
		stateWidth      = 15
		pidWidth        = 8
		procWidth       = 20
	)

	// Table header
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s",
		localAddrWidth, "Local Address",
		remoteAddrWidth, "Remote Address",
		stateWidth, "State",
		pidWidth, "PID",
		procWidth, "Process",
	)
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// Table rows
	for i, c := range m.connections {
		localAddr := fmt.Sprintf("%s:%s", c.LocalIp, c.LocalPort)
		remoteAddr := fmt.Sprintf("%s:%s", c.RemoteIp, c.RemotePort)

		// Truncate if too long
		if len(localAddr) > localAddrWidth {
			localAddr = localAddr[:localAddrWidth-3] + "..."
		}
		if len(remoteAddr) > remoteAddrWidth {
			remoteAddr = remoteAddr[:remoteAddrWidth-3] + "..."
		}
		if len(c.Proc) > procWidth {
			c.Proc = c.Proc[:procWidth-3] + "..."
		}

		// Apply state-specific styling
		stateText := c.State
		var stateStyled string
		switch c.State {
		case "ESTABLISHED":
			stateStyled = establishedStyle.Render(c.State)
		case "LISTEN":
			stateStyled = listenStyle.Render(c.State)
		case "CLOSE", "CLOSE_WAIT", "CLOSING", "TIME_WAIT":
			stateStyled = closingStyle.Render(c.State)
		default:
			stateStyled = c.State
		}

		// Build row with proper spacing
		// We need to account for the fact that styled text has ANSI codes
		// So we pad based on the original text length, not the styled length
		statePadding := stateWidth - len(stateText)

		row := fmt.Sprintf("%-*s %-*s %s%*s %-*s %-*s",
			localAddrWidth, localAddr,
			remoteAddrWidth, remoteAddr,
			stateStyled, statePadding, "",
			pidWidth, c.PID,
			procWidth, c.Proc,
		)

		// Alternate row colors
		if i%2 == 0 {
			b.WriteString(rowStyle.Render(row))
		} else {
			b.WriteString(altRowStyle.Render(row))
		}
		b.WriteString("\n")
	}

	// Footer with connection count and help
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(
		fmt.Sprintf("Total connections: %d • Press 'q' to quit • Press 'r' to refresh • Auto-refresh: 2s",
			len(m.connections)),
	))

	return b.String()
}
