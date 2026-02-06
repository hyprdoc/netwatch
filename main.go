package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xruc/netwatch/ui"
)

func main() {
	// Create the UI model with the network path
	m := ui.NewModel("/proc/net/tcp")

	// Create the Bubble Tea program
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(), // Use alternate screen buffer
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
