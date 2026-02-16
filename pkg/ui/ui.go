package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/recrsn/git-ai/pkg/git"
)

var (
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("212")).
			Padding(0, 1)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)
)

// PromptForConfirmation asks the user to confirm, edit, or cancel the commit message
func PromptForConfirmation(message string) (string, bool) {
	// Display the generated message with styling
	DisplayBox("Generated Commit Message", message)

	// Create interactive select menu
	var selectedOption string
	err := huh.NewSelect[string]().
		Title("What would you like to do?").
		Options(
			huh.NewOption("Approve", "Approve"),
			huh.NewOption("Edit", "Edit"),
			huh.NewOption("Cancel", "Cancel"),
		).
		Value(&selectedOption).
		Run()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return "", false
	}

	switch selectedOption {
	case "Approve":
		return message, true
	case "Edit":
		// Use the external editor
		editedMessage, err := git.EditWithExternalEditor(message)
		if err != nil {
			fmt.Printf("Error opening external editor: %v\n", err)
			return "", false
		}
		return editedMessage, true
	case "Cancel":
		fmt.Println("Commit cancelled.")
		return "", false
	}

	return "", false
}

// DisplayHeader shows a styled header
func DisplayHeader(text string) {
	fmt.Println(headerStyle.Render(text))
	fmt.Println()
}

// DisplaySection shows a section title
func DisplaySection(text string) {
	fmt.Println(sectionStyle.Render("▸ " + text))
}

// DisplayMessage shows a simple message
func DisplayMessage(text string) {
	fmt.Println(text)
}

// DisplayInfo shows an info message
func DisplayInfo(text string) {
	fmt.Println(infoStyle.Render("ℹ " + text))
}

// DisplayError shows an error message
func DisplayError(text string) {
	fmt.Println(errorStyle.Render("✗ " + text))
}

// DisplayBox shows text in a box with a title
func DisplayBox(title, content string) {
	titleRendered := titleStyle.Render(title)
	boxContent := titleRendered + "\n\n" + content
	fmt.Println(boxStyle.Render(boxContent))
	fmt.Println()
}

// PromptForSelection shows a selection menu and returns the selected option
func PromptForSelection(options []string, defaultOption string, promptText string) (string, error) {
	var selected string

	// Build huh options
	huhOptions := make([]huh.Option[string], len(options))
	for i, opt := range options {
		huhOptions[i] = huh.NewOption(opt, opt)
	}

	selectField := huh.NewSelect[string]().
		Title(promptText).
		Options(huhOptions...).
		Value(&selected)

	// Set default if provided
	if defaultOption != "" {
		selected = defaultOption
	}

	err := selectField.Run()
	if err != nil {
		return "", err
	}

	return selected, nil
}

// PromptForInput shows a text input prompt and returns the entered text
func PromptForInput(promptText string, defaultValue string) (string, error) {
	var input string
	if defaultValue != "" {
		input = defaultValue
	}

	err := huh.NewInput().
		Title(promptText).
		Value(&input).
		Run()

	if err != nil {
		return "", err
	}

	return input, nil
}

// PromptForPassword shows a masked text input prompt for passwords
func PromptForPassword(promptText string) (string, error) {
	var password string

	err := huh.NewInput().
		Title(promptText).
		EchoMode(huh.EchoModePassword).
		Value(&password).
		Run()

	if err != nil {
		return "", err
	}

	return password, nil
}

// spinnerModel is a simple Bubble Tea model for running a spinner with an operation
type spinnerModel struct {
	spinner   spinner.Model
	message   string
	done      bool
	err       error
	result    interface{}
	operation func() (interface{}, error)
}

func newSpinnerModel(message string, operation func() (interface{}, error)) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return spinnerModel{
		spinner:   s,
		message:   message,
		operation: operation,
	}
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			result, err := m.operation()
			return struct {
				result interface{}
				err    error
			}{result, err}
		},
	)
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case struct {
		result interface{}
		err    error
	}:
		m.done = true
		m.result = msg.result
		m.err = msg.err
		return m, tea.Quit
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		if m.err != nil {
			return errorStyle.Render("✗ Failed: "+m.err.Error()) + "\n"
		}
		return successStyle.Render("✓ Done!") + "\n"
	}
	return m.spinner.View() + " " + m.message
}

// ExitWithError displays an error message and exits
func ExitWithError(text string) {
	DisplayError(text)
	os.Exit(1)
}

// WithSpinner runs an operation with a spinner and handles success/failure
func WithSpinner(message string, operation func() error) error {
	wrappedOp := func() (interface{}, error) {
		err := operation()
		return nil, err
	}

	m := newSpinnerModel(message, wrappedOp)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()

	if err != nil {
		return fmt.Errorf("failed to run spinner: %w", err)
	}

	if sm, ok := finalModel.(spinnerModel); ok {
		return sm.err
	}

	return nil
}

// WithSpinnerResult runs an operation with a spinner and returns both result and error
func WithSpinnerResult[T any](message string, operation func() (T, error)) (T, error) {
	var zero T

	wrappedOp := func() (interface{}, error) {
		result, err := operation()
		return result, err
	}

	m := newSpinnerModel(message, wrappedOp)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()

	if err != nil {
		return zero, fmt.Errorf("failed to run spinner: %w", err)
	}

	if sm, ok := finalModel.(spinnerModel); ok {
		if sm.err != nil {
			return zero, sm.err
		}
		if result, ok := sm.result.(T); ok {
			return result, nil
		}
	}

	return zero, nil
}
