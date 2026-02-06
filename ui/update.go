package ui

import (
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
			// Manual refresh
			return m, fetchConnections(m.netPath)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case []conn.Connection:
		// Update connections when new data arrives
		m.connections = msg
		m.err = nil

	case errMsg:
		m.err = msg.err

	case tickMsg:
		// Periodic refresh
		return m, tea.Batch(
			fetchConnections(m.netPath),
			tickEvery(),
		)
	}

	return m, nil
}
