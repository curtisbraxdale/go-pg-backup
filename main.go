package main

import (
	"fmt"
	"os"

	"github.com/curtisbraxdale/go-pg-backup/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := tui.NewModel()
	p := tea.NewProgram(m) // Initialize your Bubble Tea model

	if _, err := p.Run(); err != nil {
		fmt.Printf("Uh oh, something went wrong: %v\n", err)
		os.Exit(1)
	}
}
