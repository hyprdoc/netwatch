package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xruc/netwatch/conn"
)

// FilterMode represents the connection filter state
type FilterMode int

const (
	FilterAll FilterMode = iota
	FilterLocal
	FilterPublic
)

type Model struct {
	protocol    string
	connections []conn.Connection // Current list of connections
	err         error             // Last error, if any
	width       int               // Terminal width
	height      int               // Terminal height
	netPath     string            // Path to /proc/net/*
	filterMode  FilterMode        // Current filter mode
	cursor      int               // Currently selected row
}

// tickMsg is sent periodically to refresh connection data
type tickMsg time.Time

// errMsg wraps errors for the update loop
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// NewModel creates a new UI model with the specified network path
func NewModel() Model {

	conns, err := conn.FetchAllConnections()
	if err != nil {
		return Model{
			err:         fmt.Errorf("[NewModel] failed to fetch connections error: %w", err),
			connections: []conn.Connection{},
		}
	}

	return Model{
		connections: conns,
	}
}

// Init initializes the model and starts the refresh ticker
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchConnections(),
		tickEvery(),
	)
}

// fetchConnections returns a command that fetches network connections
func fetchConnections() tea.Cmd {
	return func() tea.Msg {
		conns, err := conn.FetchAllConnections()
		if err != nil {
			return errMsg{err}
		}
		return conns
	}
}

// tickEvery returns a command that sends tick messages periodically
func tickEvery() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
