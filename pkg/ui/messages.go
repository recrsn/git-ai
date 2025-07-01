package ui

import (
	"fmt"
)

// PrintMessage prints an unformatted message to stdout (for user-facing output)
// User messages always go to stdout with decoration
func PrintMessage(message string) {
	fmt.Printf("🤖 %s\n", message)
}

// PrintMessagef prints a formatted message to stdout (for user-facing output)
// User messages always go to stdout with decoration
func PrintMessagef(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Printf("🤖 %s\n", message)
}

// PrintSuccess prints a success message to stdout with decoration
func PrintSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

// PrintError prints an error message to stdout with decoration (for user-facing errors)
func PrintError(message string) {
	fmt.Printf("❌ %s\n", message)
}

func PrintErrorf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Printf("❌ %s\n", message)
}
