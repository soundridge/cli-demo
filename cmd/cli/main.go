package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/boke/cli-demo/internal/tui"
)

func main() {
	// Bubble Tea itself will error when not attached to a TTY; we rely on that.

	p := tea.NewProgram(tui.New(0, 0), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
