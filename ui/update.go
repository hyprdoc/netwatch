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
			m.cursor++
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
			fetchConnections(m.netPath),
			tickEvery(),
		)
	}

	return m, nil
}
