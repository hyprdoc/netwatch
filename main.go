package main

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xruc/netwatch/conn"
)

type model struct {
	message string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("\n %s \n", m.message)
}

func main() {

	// m := model{
	// 	message: "Hooo Hoooo",
	// }

	// p := tea.NewProgram(m)

	// if _, err := p.Run(); err != nil {
	// 	fmt.Println("error :", err)
	// 	os.Exit(1)
	// }
	fmt.Println("LocalAddr\t\tRemoteAddr\t\tState\t\tInode")

	prevLines := 0

	for {
		conns, err := conn.FetchConnections("/proc/net/tcp")
		if err != nil {
			log.Println("error :", err)
			return
		}

		// move up only the lines we printed last time
		if prevLines > 0 {
			fmt.Printf("\033[%dA", prevLines)
		}

		linesPrinted := 0

		for _, c := range conns {
			fmt.Printf(
				"\033[2K%s:%s\t\t%s:%s\t\t%s\t\t%s\n",
				c.LocalIp, c.LocalPort,
				c.RemoteIp, c.RemotePort,
				c.State, c.Inode,
			)
			linesPrinted++
		}

		prevLines = linesPrinted

		time.Sleep(5 * time.Second)
	}

}
