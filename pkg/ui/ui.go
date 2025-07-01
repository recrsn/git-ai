package ui

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/recrsn/git-ai/pkg/git"
)

// PromptForConfirmation asks the user to confirm, edit, or cancel the commit message
func PromptForConfirmation(message string) (string, bool) {
	// Display the generated message with styling
	pterm.DefaultBox.WithTitle("Generated Commit Message").WithTitleBottomRight().Print(message)
	pterm.Println()

	// Create interactive select menu
	options := []string{"Approve", "Edit", "Cancel"}
	selectedOption, err := pterm.DefaultInteractiveSelect.
		WithOptions(options).
		WithDefaultText("What would you like to do?").
		Show()

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
		pterm.Println("Commit cancelled.")
		return "", false
	}

	return "", false
}

// DisplayHeader shows a styled header
func DisplayHeader(text string) {
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgLightMagenta)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Println(text)
	pterm.Println()
}

// DisplaySection shows a section title
func DisplaySection(text string) {
	pterm.DefaultSection.Println(text)
}

// DisplayMessage shows a simple message
func DisplayMessage(text string) {
	pterm.Println(text)
}

// DisplayInfo shows an info message
func DisplayInfo(text string) {
	pterm.Info.Println(text)
}

// DisplayError shows an error message
func DisplayError(text string) {
	pterm.Error.Println(text)
}

// DisplayBox shows text in a box with a title
func DisplayBox(title, content string) {
	pterm.DefaultBox.WithTitle(title).WithTitleBottomRight().Print(content)
	pterm.Println()
}

// PromptForSelection shows a selection menu and returns the selected option
func PromptForSelection(options []string, defaultOption string, promptText string) (string, error) {
	return pterm.DefaultInteractiveSelect.
		WithOptions(options).
		WithDefaultOption(defaultOption).
		WithDefaultText(promptText).
		Show()
}

// PromptForInput shows a text input prompt and returns the entered text
func PromptForInput(promptText string, defaultValue string) (string, error) {
	return pterm.DefaultInteractiveTextInput.
		WithDefaultValue(defaultValue).
		Show(promptText)
}

// PromptForPassword shows a masked text input prompt for passwords
func PromptForPassword(promptText string) (string, error) {
	return pterm.DefaultInteractiveTextInput.
		WithMask("•").
		Show(promptText)
}

// ShowSpinner starts a spinner and returns the spinner instance
func ShowSpinner(text string) (*pterm.SpinnerPrinter, error) {
	return pterm.DefaultSpinner.Start(text)
}

// ExitWithError displays an error message and exits
func ExitWithError(text string) {
	pterm.Error.Println(text)
	os.Exit(1)
}
