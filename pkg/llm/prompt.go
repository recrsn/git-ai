package llm

import (
	"bytes"
	_ "embed"
	"github.com/recrsn/git-ai/pkg/logger"
	"strings"
	"text/template"
)

// Embedded prompt files at compile time
//
//go:embed prompts/commit_system.txt
var commitSystemPromptTemplate string

//go:embed prompts/commit_user.txt
var commitUserPromptTemplate string

// PromptData contains the data to be inserted into the prompt template
type PromptData struct {
	Diff                    string
	ChangedFiles            string
	RecentCommits           string
	UseConventional         bool
	CommitsWithDescriptions bool
}

// GetSystemPrompt returns the system prompt for commit message generation
func GetSystemPrompt(useConventionalCommits bool, commitsWithDescriptions bool) string {
	// Define template functions
	funcMap := template.FuncMap{
		"trimSpace": strings.TrimSpace,
	}

	// Parse the template
	tmpl, err := template.New("systemPrompt").Funcs(funcMap).Parse(commitSystemPromptTemplate)
	if err != nil {
		// Fallback to a simple prompt if template parsing fails
		logger.Error("Error parsing system prompt template: %v", err)
		logger.Warn("Falling back to default system prompt")
		return "Generate a clear, concise commit message for the changes."
	}

	// Prepare data for the template
	data := struct {
		UseConventional         bool
		CommitsWithDescriptions bool
	}{
		UseConventional:         useConventionalCommits,
		CommitsWithDescriptions: commitsWithDescriptions,
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		// Fallback to a simple prompt if execution fails
		logger.Error("Error executing system prompt template: %v", err)
		logger.Warn("Falling back to default system prompt")
		return "Generate a clear, concise commit message for the changes."
	}

	return buf.String()
}

// GetUserPrompt generates a user prompt with the given data
func GetUserPrompt(diff, changedFiles, recentCommits string) (string, error) {
	// Format changed files as a list
	formattedChangedFiles := formatAsList(changedFiles)

	// Format recent commits as a list
	formattedRecentCommits := formatAsList(recentCommits)

	// Prepare data for template
	data := PromptData{
		Diff:          diff,
		ChangedFiles:  formattedChangedFiles,
		RecentCommits: formattedRecentCommits,
	}

	// Define template functions
	funcMap := template.FuncMap{
		"trimSpace": strings.TrimSpace,
	}

	// Parse and execute the template
	tmpl, err := template.New("commit").Funcs(funcMap).Parse(commitUserPromptTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Helper function to format a newline-separated string as a list
func formatAsList(input string) string {
	lines := strings.Split(input, "\n")
	var result strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result.WriteString("- ")
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return result.String()
}
