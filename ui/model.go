package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xruc/netwatch/conn"
)

// Model represents the application state for the Bubble Tea UI
type Model struct {
	connections []conn.Connection
	err         error
	width       int
	height      int
	netPath     string
}

// tickMsg is sent periodically to refresh connection data
type tickMsg time.Time

// errMsg wraps errors for the update loop
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// NewModel creates a new UI model with the specified network path
func NewModel(netPath string) Model {
	return Model{
		netPath:     netPath,
		connections: []conn.Connection{},
	}
}

// Init initializes the model and starts the refresh ticker
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchConnections(m.netPath),
		tickEvery(),
	)
}

// fetchConnections returns a command that fetches network connections
func fetchConnections(netPath string) tea.Cmd {
	return func() tea.Msg {
		conns, err := conn.FetchConnections(netPath)
		if err != nil {
			return errMsg{err}
		}
		return conns
	}
}

// tickEvery returns a command that sends tick messages periodically
func tickEvery() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
