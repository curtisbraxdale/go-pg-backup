package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model defines the application's state.
type Model struct {
	inputs        []textinput.Model
	step          int
	focusOnInput  bool
	focusedButton int // 0: back, 1: next/submit
	submitted     bool
	quitting      bool

	// Backup state
	backupInProgress bool
	backupFinished   bool
	backupError      error
	outputPath       string
	backupMessage    string
}

// NewModel initializes the model with the required text inputs.
func NewModel() Model {
	inputs := make([]textinput.Model, 5)
	prompts := []string{
		"Database Host",
		"Database User",
		"Database Password",
		"Database Name",
		"Backup Directory",
	}
	placeholders := []string{
		"localhost",
		"postgres",
		"password",
		"mydatabase",
		"/path/to/backup",
	}

	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Prompt = prompts[i] + ": "
		inputs[i].Placeholder = placeholders[i]
		inputs[i].CharLimit = 256
		inputs[i].Width = 50
		inputs[i].PromptStyle = pinkTextPrompt
		inputs[i].TextStyle = whiteText
		if i == 2 { // Password
			inputs[i].EchoMode = textinput.EchoPassword
			inputs[i].EchoCharacter = '•'
		}
	}

	inputs[0].Focus()

	return Model{
		inputs:        inputs,
		step:          0,
		focusOnInput:  true,
		focusedButton: 0,
	}
}

// Init kicks off the event loop.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// nextStep moves the form to the next input.
func (m *Model) nextStep() {
	if m.step < len(m.inputs)-1 {
		m.inputs[m.step].Blur()
		m.step++
		m.inputs[m.step].Focus()
	}
}

// prevStep moves the form to the previous input.
func (m *Model) prevStep() {
	if m.step > 0 {
		m.inputs[m.step].Blur()
		m.step--
		m.inputs[m.step].Focus()
	}
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quitting {
		return m, tea.Quit
	}

	// Handle backup messages
	switch msg := msg.(type) {
	case PgDumpStartedMsg:
		m.backupInProgress = true
		m.backupMessage = "Backup started..."
		return m, nil
	case PgDumpFinishedMsg:
		m.backupInProgress = false
		m.backupFinished = true
		m.backupError = msg.Err
		m.outputPath = msg.OutputPath
		if msg.Err != nil {
			m.backupMessage = fmt.Sprintf("Backup failed: %v", msg.Err)
		} else {
			m.backupMessage = "Backup completed successfully!"
		}
		m.quitting = true  // Set quitting to true to show final summary
		return m, tea.Quit // Quit after showing the final message
	case PgDumpProgressMsg:
		m.backupMessage = string(msg)
		return m, nil
	}

	if m.backupInProgress || m.backupFinished {
		// If backup is running or finished, only allow quitting
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.quitting = true
				return m, tea.Quit
			}
		}
		return m, nil
	}

	currentInput := &m.inputs[m.step]

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit

		// Change focus between input and buttons
		case tea.KeyUp, tea.KeyDown:
			m.focusOnInput = !m.focusOnInput
			if m.focusOnInput {
				currentInput.Focus()
			} else {
				currentInput.Blur()
			}
			return m, nil

		// Cycle through buttons
		case tea.KeyLeft, tea.KeyRight, tea.KeyTab, tea.KeyShiftTab:
			if !m.focusOnInput {
				m.focusedButton = 1 - m.focusedButton // Toggle between 0 and 1
			}
			return m, nil

		case tea.KeyEnter:
			if m.focusOnInput {
				if m.step == len(m.inputs)-1 {
					m.submitted = true
					return m, RunPgDumpCmd(m)
				}
				m.nextStep()
			} else {
				// A button is focused
				if m.focusedButton == 0 { // Back button
					m.prevStep()
				} else { // Next/Submit button
					if m.step == len(m.inputs)-1 {
						m.submitted = true
						return m, RunPgDumpCmd(m)
					}
					m.nextStep()
				}
				// Return focus to the input after a button press
				m.focusOnInput = true
				m.inputs[m.step].Focus()
			}
			return m, nil
		}
	}

	// Handle character input and blinking
	var cmd tea.Cmd
	if m.focusOnInput {
		*currentInput, cmd = currentInput.Update(msg)
	}
	return m, cmd
}

// View renders the UI.
func (m Model) View() string {
	if m.quitting {
		if m.submitted {
			var b strings.Builder
			if m.backupError != nil {
				b.WriteString(summaryStyle.Render("Backup Failed!"))
				b.WriteString("\n\n")
				b.WriteString(cancelledStyle.Render(m.backupMessage))
			} else {
				b.WriteString(summaryStyle.Render("Backup Successful!"))
				b.WriteString("\n\n")
				b.WriteString(greenTextPrompt.Render(m.backupMessage))
				if m.outputPath != "" {
					b.WriteString(fmt.Sprintf("\nBackup file: %s", greenTextValue.Render(m.outputPath)))
				}
			}
			b.WriteString("\n\nPress any key to exit.")
			return b.String()
		}
		return cancelledStyle.Render("Backup wizard cancelled.") + "\n"
	}

	if m.backupInProgress {
		var b strings.Builder
		b.WriteString(welcomeStyle.Render("Welcome to the PostgreSQL Backup Wizard!"))
		b.WriteString("\n\n")
		b.WriteString(lipgloss.JoinVertical(lipgloss.Left,
			"Backup in progress...",
			greyText.Render(m.backupMessage),
		))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("ctrl+c: cancel"))
		return b.String()
	}

	if m.submitted {
		var b strings.Builder
		b.WriteString(summaryStyle.Render("Backup configuration summary:"))
		b.WriteString("\n\n")
		for i := range m.inputs {
			var value string
			if i == 2 { // Password
				value = strings.Repeat("•", len(m.inputs[i].Value()))
			} else {
				value = m.inputs[i].Value()
			}
			line := fmt.Sprintf("%s %s", greenTextPrompt.Render(m.inputs[i].Prompt), greenTextValue.Render(value))
			b.WriteString(line)
			b.WriteRune('\n')
		}
		b.WriteString("\n")
		b.WriteString("Starting backup...")
		return b.String()
	}

	var b strings.Builder

	// Title
	b.WriteString(welcomeStyle.Render("Welcome to the PostgreSQL Backup Wizard!"))
	b.WriteString("\n\n")

	// Step Indicator
	var steps []string
	for i := 1; i <= len(m.inputs); i++ {
		stepStr := fmt.Sprintf("Step %d", i)
		if i-1 <= m.step {
			steps = append(steps, greenTextPrompt.Render(stepStr))
		} else {
			steps = append(steps, greyText.Render(stepStr))
		}
	}
	b.WriteString(strings.Join(steps, " -> "))
	b.WriteString("\n\n")

	// Staged answers
	for i := 0; i < m.step; i++ {
		var value string
		if i == 2 { // Password
			value = strings.Repeat("•", len(m.inputs[i].Value()))
		} else {
			value = m.inputs[i].Value()
		}
		line := fmt.Sprintf("%s %s", greenTextPrompt.Render(m.inputs[i].Prompt), greenTextValue.Render(value))
		b.WriteString(line)
		b.WriteRune('\n')
	}
	if m.step > 0 {
		b.WriteRune('\n')
	}

	// Current input
	b.WriteString(m.inputs[m.step].View())
	b.WriteString("\n\n")

	// Buttons
	var backButton, nextButton string
	backStyle, nextStyle := blurredButton, blurredButton

	if !m.focusOnInput {
		if m.focusedButton == 0 {
			backStyle = focusedButton
		} else {
			nextStyle = focusedButton
		}
	}

	if m.step > 0 {
		backButton = backStyle.Render("[ Back ]")
	} else {
		// Keep alignment
		backButton = strings.Repeat(" ", 8)
	}

	if m.step == len(m.inputs)-1 {
		nextButton = nextStyle.Render("[ Submit ]")
	} else {
		nextButton = nextStyle.Render("[ Next ]")
	}

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, backButton, " ", nextButton))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("up/down: toggle focus • left/right: switch buttons • enter: select • ctrl+c: quit"))

	return b.String()
}
