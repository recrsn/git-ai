package ui

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

// PromptForConfirmation asks the user to confirm, edit, or cancel the commit message
func PromptForConfirmation(message string) (string, bool) {
	fmt.Println("Generated Commit Message:")
	fmt.Println("===========================")
	fmt.Println(message)
	fmt.Println("===========================")

	var action string
	prompt := &survey.Select{
		Message: "What would you like to do?",
		Options: []string{"Approve", "Edit", "Cancel"},
		Default: "Approve",
	}

	err := survey.AskOne(prompt, &action)
	if err == terminal.InterruptErr {
		fmt.Println("Commit cancelled.")
		return "", false
	} else if err != nil {
		fmt.Printf("Error: %v\n", err)
		return "", false
	}

	switch action {
	case "Approve":
		return message, true
	case "Edit":
		var editedMessage string
		textPrompt := &survey.Editor{
			Message:       "Edit commit message:",
			Default:       message,
			AppendDefault: true,
			HideDefault:   false,
		}

		err := survey.AskOne(textPrompt, &editedMessage)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return "", false
		}
		return editedMessage, true
	case "Cancel":
		fmt.Println("Commit cancelled.")
		return "", false
	}

	return "", false
}
