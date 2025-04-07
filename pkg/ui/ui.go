package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			MarginBottom(1)

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	selectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("63"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

// PromptModel represents the Bubble Tea model for prompt selection
type PromptModel struct {
	choices  []string
	cursor   int
	selected int
	quitting bool
	message  string
}

// Init initializes the model
func (m PromptModel) Init() tea.Cmd {
	return nil
}

// Update handles updates to the model
func (m PromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			m.selected = -1
			return m, tea.Quit
		case "enter":
			m.selected = m.cursor
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

// View renders the model
func (m PromptModel) View() string {
	s := titleStyle.Render("Generated Commit Message:") + "\n"
	s += "===========================\n"
	s += m.message + "\n"
	s += "===========================\n\n"
	s += "What would you like to do?\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = selectStyle.Render("> ")
		} else {
			cursor = "  "
		}
		s += fmt.Sprintf("%s%s\n", cursor, choice)
	}

	return s
}

// EditorModel represents the Bubble Tea model for text editing
type EditorModel struct {
	textarea textarea.Model
	done     bool
}

// Init initializes the editor model
func (m EditorModel) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles updates to the editor model
func (m EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.done = true
			return m, tea.Quit
		case tea.KeyEsc:
			m.done = true
			return m, tea.Quit
		case tea.KeyCtrlD:
			m.done = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View renders the editor model
func (m EditorModel) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		titleStyle.Render("Edit Commit Message:"),
		m.textarea.View(),
		"(Press ESC or Ctrl+D to finish)"+"      ",
	)
}

// promptForSelection runs a Bubble Tea program to select an option
func promptForSelection(message string) (int, error) {
	model := PromptModel{
		choices:  []string{"Approve", "Edit", "Cancel"},
		cursor:   0,
		selected: -1,
		message:  message,
	}

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return -1, err
	}

	m, ok := finalModel.(PromptModel)
	if !ok {
		return -1, fmt.Errorf("could not cast final model")
	}

	return m.selected, nil
}

// editMessage runs a Bubble Tea program to edit text
func editMessage(message string) (string, error) {
	ta := textarea.New()
	ta.SetValue(message)
	ta.Focus()
	ta.SetHeight(20)
	ta.SetWidth(80)
	ta.Placeholder = "Edit commit message..."
	ta.ShowLineNumbers = false

	model := EditorModel{
		textarea: ta,
	}

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	m, ok := finalModel.(EditorModel)
	if !ok {
		return "", fmt.Errorf("could not cast final model")
	}

	return m.textarea.Value(), nil
}

// PromptForConfirmation asks the user to confirm, edit, or cancel the commit message
func PromptForConfirmation(message string) (string, bool) {
	selected, err := promptForSelection(message)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return "", false
	}

	switch selected {
	case 0: // Approve
		return message, true
	case 1: // Edit
		editedMessage, err := editMessage(message)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return "", false
		}
		return editedMessage, true
	case 2, -1: // Cancel or interrupted
		fmt.Println("Commit cancelled.")
		return "", false
	}

	return "", false
}
