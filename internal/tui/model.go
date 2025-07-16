package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewState int

const (
	mainMenu viewState = iota
	restoreChoiceMenu
	backupForm
	restoreForm
)

// Model defines the application's state.
type Model struct {
	// View management
	currentView       viewState
	mainMenuChoice    int // 0: backup, 1: restore
	restoreMenuChoice int // 0: existing db, 1: new db

	// Form state
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

	// Restore state
	restoreInProgress bool
	restoreFinished   bool
	restoreError      error
	restoreMessage    string
	restoreNewDB      bool
}

// NewModel initializes the model with the required text inputs.
func NewModel() Model {
	return Model{
		currentView:    mainMenu,
		mainMenuChoice: 0,
		focusOnInput:   true,
	}
}

func setupBackupInputs() []textinput.Model {
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
	return inputs
}

func setupRestoreInputs() []textinput.Model {
	inputs := make([]textinput.Model, 5)
	prompts := []string{
		"Database Host",
		"Database User",
		"Database Password",
		"Database Name",
		"Backup File Path",
	}
	placeholders := []string{
		"localhost",
		"postgres",
		"password",
		"mydatabase_restored",
		"/path/to/backup.sql",
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
	return inputs
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

	// Global messages
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
			m.quitting = true
			return m, tea.Quit
		}
	// Backup messages
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
		m.quitting = true
		return m, tea.Quit
	case PgDumpProgressMsg:
		m.backupMessage = string(msg)
		return m, nil
	// Restore messages
	case PgRestoreStartedMsg:
		m.restoreInProgress = true
		m.restoreMessage = "Restore started..."
		return m, nil
	case PgRestoreFinishedMsg:
		m.restoreInProgress = false
		m.restoreFinished = true
		m.restoreError = msg.Err
		if msg.Err != nil {
			m.restoreMessage = fmt.Sprintf("Restore failed: %v", msg.Err)
		} else {
			m.restoreMessage = "Restore completed successfully!"
		}
		m.quitting = true
		return m, tea.Quit
	case PgRestoreProgressMsg:
		m.restoreMessage = string(msg)
		return m, nil
	}

	if m.backupInProgress || m.restoreInProgress {
		return m, nil
	}

	switch m.currentView {
	case mainMenu:
		return m.updateMainMenu(msg)
	case restoreChoiceMenu:
		return m.updateRestoreChoiceMenu(msg)
	case backupForm, restoreForm:
		return m.updateForm(msg)
	}

	return m, nil
}

func (m Model) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp, tea.KeyDown:
			m.mainMenuChoice = 1 - m.mainMenuChoice // Toggle
		case tea.KeyEnter:
			if m.mainMenuChoice == 0 { // Backup
				m.currentView = backupForm
				m.inputs = setupBackupInputs()
			} else { // Restore
				m.currentView = restoreChoiceMenu
			}
		}
	}
	return m, nil
}

func (m Model) updateRestoreChoiceMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp, tea.KeyDown:
			m.restoreMenuChoice = 1 - m.restoreMenuChoice // Toggle
		case tea.KeyEnter:
			m.restoreNewDB = m.restoreMenuChoice == 1 // 1 is "Create new database"
			m.currentView = restoreForm
			m.inputs = setupRestoreInputs()
		case tea.KeyEsc: // Go back to main menu
			m.currentView = mainMenu
		}
	}
	return m, nil
}

func (m Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	currentInput := &m.inputs[m.step]

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp, tea.KeyDown:
			m.focusOnInput = !m.focusOnInput
			if m.focusOnInput {
				currentInput.Focus()
			} else {
				currentInput.Blur()
			}
			return m, nil
		case tea.KeyLeft, tea.KeyRight, tea.KeyTab, tea.KeyShiftTab:
			if !m.focusOnInput {
				m.focusedButton = 1 - m.focusedButton // Toggle
			}
			return m, nil
		case tea.KeyEnter:
			if m.focusOnInput {
				if m.step == len(m.inputs)-1 {
					m.submitted = true
					if m.currentView == backupForm {
						return m, RunPgDumpCmd(m)
					}
					return m, RunPgRestoreCmd(m)
				}
				m.nextStep()
			} else {
				if m.focusedButton == 0 { // Back
					if m.step == 0 {
						m.currentView = mainMenu
						m.step = 0 // Reset form state
					} else {
						m.prevStep()
					}
				} else { // Next/Submit
					if m.step == len(m.inputs)-1 {
						m.submitted = true
						if m.currentView == backupForm {
							return m, RunPgDumpCmd(m)
						}
						return m, RunPgRestoreCmd(m)
					}
					m.nextStep()
				}
				m.focusOnInput = true
				m.inputs[m.step].Focus()
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	if m.focusOnInput {
		*currentInput, cmd = currentInput.Update(msg)
	}
	return m, cmd
}

// View renders the UI.
func (m Model) View() string {
	if m.quitting {
		return m.viewSummary()
	}

	if m.backupInProgress {
		return m.viewProgress("Backup in progress...", m.backupMessage)
	}
	if m.restoreInProgress {
		return m.viewProgress("Restore in progress...", m.restoreMessage)
	}

	if m.submitted {
		return m.viewPreSubmit()
	}

	switch m.currentView {
	case mainMenu:
		return m.viewMainMenu()
	case restoreChoiceMenu:
		return m.viewRestoreChoiceMenu()
	case backupForm, restoreForm:
		return m.viewForm()
	default:
		return "Something went wrong."
	}
}

func (m Model) viewSummary() string {
	if !m.submitted {
		return cancelledStyle.Render("Wizard cancelled.") + "\n"
	}

	var b strings.Builder
	var err error
	var msg, title string

	if m.backupFinished {
		err = m.backupError
		msg = m.backupMessage
		title = "Backup"
	} else if m.restoreFinished {
		err = m.restoreError
		msg = m.restoreMessage
		title = "Restore"
	}

	if err != nil {
		b.WriteString(summaryStyle.Render(fmt.Sprintf("%s Failed!", title)))
		b.WriteString("\n\n")
		b.WriteString(cancelledStyle.Render(msg))
	} else {
		b.WriteString(summaryStyle.Render(fmt.Sprintf("%s Successful!", title)))
		b.WriteString("\n\n")
		b.WriteString(greenTextPrompt.Render(msg))
		if m.outputPath != "" {
			b.WriteString(fmt.Sprintf("\nBackup file: %s", greenTextValue.Render(m.outputPath)))
		}
	}
	b.WriteString("\n\nPress any key to exit.")
	return b.String()
}

func (m Model) viewProgress(title, message string) string {
	var b strings.Builder
	b.WriteString(welcomeStyle.Render("PostgreSQL Backup & Restore Wizard"))
	b.WriteString("\n\n")
	b.WriteString(lipgloss.JoinVertical(lipgloss.Left,
		title,
		greyText.Render(message),
	))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("ctrl+c: cancel"))
	return b.String()
}

func (m Model) viewPreSubmit() string {
	var b strings.Builder
	title := "Backup"
	if m.currentView == restoreForm {
		title = "Restore"
	}
	b.WriteString(summaryStyle.Render(fmt.Sprintf("%s configuration summary:", title)))
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
	b.WriteString(fmt.Sprintf("Starting %s...", strings.ToLower(title)))
	return b.String()
}

func (m Model) viewMainMenu() string {
	var b strings.Builder
	b.WriteString(welcomeStyle.Render("Welcome to the PostgreSQL Backup & Restore Wizard!"))
	b.WriteString("\n\n")
	b.WriteString("What would you like to do?\n\n")

	backup := "[ ] Create a new backup"
	restore := "[ ] Restore from a backup file"

	if m.mainMenuChoice == 0 {
		backup = focusedButton.Render("[x] Create a new backup")
	} else {
		restore = focusedButton.Render("[x] Restore from a backup file")
	}

	b.WriteString(lipgloss.JoinVertical(lipgloss.Left, backup, restore))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("up/down: select • enter: confirm • ctrl+c: quit"))
	return b.String()
}

func (m Model) viewRestoreChoiceMenu() string {
	var b strings.Builder
	b.WriteString(welcomeStyle.Render("Restore Database"))
	b.WriteString("\n\n")
	b.WriteString("Choose a restore option:\n\n")

	existing := "[ ] Restore to an existing database"
	newDB := "[ ] Create a new database and restore into it"

	if m.restoreMenuChoice == 0 {
		existing = focusedButton.Render("[x] Restore to an existing database")
	} else {
		newDB = focusedButton.Render("[x] Create a new database and restore into it")
	}

	b.WriteString(lipgloss.JoinVertical(lipgloss.Left, existing, newDB))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("up/down: select • enter: confirm • esc: back • ctrl+c: quit"))
	return b.String()
}

func (m Model) viewForm() string {
	var b strings.Builder

	title := "Backup"
	if m.currentView == restoreForm {
		title = "Restore"
	}

	b.WriteString(welcomeStyle.Render(fmt.Sprintf("PostgreSQL %s Wizard", title)))
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

	backButton = backStyle.Render("[ Back ]")

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
