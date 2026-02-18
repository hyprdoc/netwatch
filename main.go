package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xruc/netwatch/ui"

	"net/http"
	_ "net/http/pprof"
)

func main() {

	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	m := ui.NewModel()
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(), // Use alternate screen buffer
	)

	log.Println("Starting netwatch...")
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
