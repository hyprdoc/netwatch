package ui

import (
	"log"
	"net/netip"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xruc/netwatch/conn"
)

// Update handles messages and updates the model state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "r":
			return m, fetchConnections() // manual refresh
		case "l":
			// Toggle filter mode: all -> local -> public -> all
			switch m.filterMode {
			case FilterAll:
				m.filterMode = FilterLocal
			case FilterLocal:
				m.filterMode = FilterPublic
			case FilterPublic:
				m.filterMode = FilterAll
			}
			// Reset cursor when filter changes
			m.cursor = 0
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			// We'll check bounds in the view based on filtered connections
			if m.cursor < len(m.connections)-1 {
				m.cursor++
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case []conn.Connection:
		// Update connections when new data arrives
		m.connections = msg
		m.err = nil
		// Keep cursor in bounds
		if m.cursor >= len(msg) {
			m.cursor = len(msg) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}

	case errMsg:
		m.err = msg.err

	case tickMsg:
		// Periodic refresh
		return m, tea.Batch(
			fetchConnections(),
			tickEvery(),
		)
	}

	m.connections = m.filterConnections()

	return m, nil
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

		log.Println(c.RemoteIp, "is local ?", isLocal)

		if m.filterMode == FilterLocal && isLocal {
			filtered = append(filtered, c)
		} else if m.filterMode == FilterPublic && !isLocal {
			filtered = append(filtered, c)
		}
	}

	return filtered
}

// isLocalAddress checks if an IP address is local/loopback/private
func isLocalAddress(addr string) bool {
	ip, err := netip.ParseAddr(addr)
	if err != nil {
		return false
	}
	return isWildcardLocal(ip.String()) || ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsMulticast()
}

func isWildcardLocal(local string) bool {
	local = strings.TrimSpace(local)
	if local == "" {
		return false
	}

	// Common ss / netstat patterns
	if local == "0.0.0.0:0" || local == "0.0.0.0:" ||
		local == ":::0" || local == ":::" {
		return true
	}

	// Anything starting with 0.0.0.0: or ::: (followed by port)
	if strings.HasPrefix(local, "0.0.0.0") ||
		strings.HasPrefix(local, "::") {
		return true
	}

	// Rare variants like [::]:port (some tools show brackets)
	// if strings.HasPrefix(local, "[::]:") {
	//     return true
	// }

	return false
}
