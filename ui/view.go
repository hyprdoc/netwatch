package ui

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/xruc/netwatch/conn"
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

	// Filter connections based on mode
	filteredConns := m.connections
	slices.SortFunc(filteredConns, func(a, b conn.Connection) int {
		return cmp.Compare(a.Proc, b.Proc)
	})

	// No connections
	if len(filteredConns) == 0 {
		filterMsg := ""
		switch m.filterMode {
		case FilterLocal:
			filterMsg = " (filtering: local only)"
		case FilterPublic:
			filterMsg = " (filtering: public only)"
		}
		b.WriteString(fmt.Sprintf("No connections found%s...\n\n", filterMsg))
		b.WriteString(helpStyle.Render("Press 'q' to quit • Press 'r' to refresh • Press 'l' to toggle filter"))
		return b.String()
	}

	// Table header
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s",
		ProcWidth, "Process",
		LocalAddrWidth, "Local Address",
		RemoteAddrWidth, "Remote Address",
		StateWidth, "State",
	)
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// Keep cursor in bounds for filtered connections
	cursor := m.cursor
	if cursor >= len(filteredConns) {
		cursor = len(filteredConns) - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	// Table rows
	for i, c := range filteredConns {
		localAddr := fmt.Sprintf("%s:%s", c.LocalIp, c.LocalPort)
		remoteAddr := fmt.Sprintf("%s:%s", c.RemoteIp, c.RemotePort)

		// Truncate if too long
		if len(localAddr) > LocalAddrWidth {
			localAddr = localAddr[:LocalAddrWidth-3] + "..."
		}
		if len(remoteAddr) > RemoteAddrWidth {
			remoteAddr = remoteAddr[:RemoteAddrWidth-3] + "..."
		}
		if len(c.Proc) > ProcWidth {
			c.Proc = c.Proc[:ProcWidth-3] + "..."
		}

		// Apply state-specific styling
		stateText := c.State

		// Build row - if selected, don't apply state colors (they'll be overridden anyway)
		var row string
		if i == cursor {
			// For selected row, use plain text for state (no color styling)
			row = fmt.Sprintf("%-*s %-*s %-*s %-*s",
				ProcWidth, c.Proc+"("+c.PID+")",
				LocalAddrWidth, localAddr,
				RemoteAddrWidth, remoteAddr,
				StateWidth, stateText,
			)
			b.WriteString(selectedRowStyle.Render(row))
		} else {
			// For non-selected rows, apply state-specific styling
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
			statePadding := StateWidth - len(stateText)

			row = fmt.Sprintf("%-*s %-*s %-*s %s%*s",
				ProcWidth, c.Proc+"("+c.PID+")",
				LocalAddrWidth, localAddr,
				RemoteAddrWidth, remoteAddr,
				stateStyled, statePadding, "",
			)

			// Apply alternating row colors
			if i%2 == 0 {
				b.WriteString(rowStyle.Render(row))
			} else {
				b.WriteString(altRowStyle.Render(row))
			}
		}
		b.WriteString("\n")
	}

	// Footer with connection count and help
	b.WriteString("\n")

	// Filter status
	filterStatus := "all"
	switch m.filterMode {
	case FilterLocal:
		filterStatus = "local only"
	case FilterPublic:
		filterStatus = "public only"
	}

	b.WriteString(helpStyle.Render(
		fmt.Sprintf("Showing: %d/%d (%s) • ↑↓/j/k navigate • 'l' filter • 'r' refresh • 'q' quit • Auto-refresh: 2s",
			len(filteredConns), len(m.connections), filterStatus),
	))

	return b.String()
}
