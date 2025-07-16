package tui

import tea "github.com/charmbracelet/bubbletea"

// application state
type Model struct {
	title string
}

// initial model
func NewModel() Model {
	return Model{
		title: "Welcome to the go=pg-backup wizard!",
	}
}

// kicks off event loop
func (m Model) Init() tea.Cmd {
	return nil
}

// handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	return m.title
}
