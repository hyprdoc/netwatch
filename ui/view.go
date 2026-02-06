package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xruc/netwatch/conn"
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

	// Filter connections based on mode
	filteredConns := m.filterConnections()

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
	for i, c := range filteredConns {
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

	// Filter status
	filterStatus := "all"
	switch m.filterMode {
	case FilterLocal:
		filterStatus = "local only"
	case FilterPublic:
		filterStatus = "public only"
	}

	b.WriteString(helpStyle.Render(
		fmt.Sprintf("Showing: %d/%d (%s) • 'q' quit • 'r' refresh • 'l' toggle filter • Auto-refresh: 2s",
			len(filteredConns), len(m.connections), filterStatus),
	))

	return b.String()
}

// filterConnections returns connections based on the current filter mode
func (m Model) filterConnections() []conn.Connection {
	if m.filterMode == FilterAll {
		return m.connections
	}

	filtered := make([]conn.Connection, 0)
	for _, c := range m.connections {
		// Only check the remote address to determine if connection is local or public
		isLocal := isLocalAddress(c.RemoteIp)

		if m.filterMode == FilterLocal && isLocal {
			filtered = append(filtered, c)
		} else if m.filterMode == FilterPublic && !isLocal {
			filtered = append(filtered, c)
		}
	}

	return filtered
}

// isLocalAddress checks if an IP address is local/loopback/private
func isLocalAddress(ip string) bool {
	// Loopback and unspecified addresses
	if ip == "127.0.0.1" || ip == "::1" || ip == "0.0.0.0" {
		return true
	}

	// Loopback range (127.x.x.x)
	if strings.HasPrefix(ip, "127.") {
		return true
	}

	// Private IP ranges
	if strings.HasPrefix(ip, "10.") {
		return true
	}

	if strings.HasPrefix(ip, "192.168.") {
		return true
	}

	// 172.16.0.0 - 172.31.255.255
	if strings.HasPrefix(ip, "172.") {
		parts := strings.Split(ip, ".")
		if len(parts) >= 2 {
			second := parts[1]
			// Check if second octet is between 16 and 31
			for i := 16; i <= 31; i++ {
				if second == fmt.Sprintf("%d", i) {
					return true
				}
			}
		}
	}

	// Link-local (169.254.x.x)
	if strings.HasPrefix(ip, "169.254.") {
		return true
	}

	// Multicast (224.x.x.x - 239.x.x.x)
	if strings.HasPrefix(ip, "224.") || strings.HasPrefix(ip, "225.") ||
		strings.HasPrefix(ip, "226.") || strings.HasPrefix(ip, "227.") ||
		strings.HasPrefix(ip, "228.") || strings.HasPrefix(ip, "229.") ||
		strings.HasPrefix(ip, "230.") || strings.HasPrefix(ip, "231.") ||
		strings.HasPrefix(ip, "232.") || strings.HasPrefix(ip, "233.") ||
		strings.HasPrefix(ip, "234.") || strings.HasPrefix(ip, "235.") ||
		strings.HasPrefix(ip, "236.") || strings.HasPrefix(ip, "237.") ||
		strings.HasPrefix(ip, "238.") || strings.HasPrefix(ip, "239.") {
		return true
	}

	return false
}
