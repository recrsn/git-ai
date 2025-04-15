package llm

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"
)

// Embedded prompt files at compile time
//
//go:embed prompts/commit_system.txt
var commitSystemPromptTemplate string

//go:embed prompts/commit_user.txt
var commitUserPromptTemplate string

//go:embed prompts/branch_system.txt
var branchSystemPromptTemplate string

//go:embed prompts/branch_user.txt
var branchUserPromptTemplate string

// CommitPromptData contains the data to be inserted into the commit prompt template
type CommitPromptData struct {
	Diff                    string
	ChangedFiles            string
	RecentCommits           string
	UseConventional         bool
	CommitsWithDescriptions bool
}

// BranchPromptData contains the data to be inserted into the branch prompt template
type BranchPromptData struct {
	Request        string
	LocalBranches  string
	RemoteBranches string
}

// GetSystemPrompt returns the system prompt for commit message generation
func GetSystemPrompt(useConventionalCommits bool, commitsWithDescriptions bool) (string, error) {
	// Define template functions
	funcMap := template.FuncMap{
		"trimSpace": strings.TrimSpace,
	}

	// Parse the template
	tmpl, err := template.New("systemPrompt").Funcs(funcMap).Parse(commitSystemPromptTemplate)
	if err != nil {
		return "", fmt.Errorf("error parsing system prompt template: %w", err)
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
		return "", fmt.Errorf("error executing system prompt template: %w", err)
	}

	return buf.String(), nil
}

// GetUserPrompt generates a user prompt with the given data
func GetUserPrompt(diff, changedFiles, recentCommits string) (string, error) {
	// Format changed files as a list
	formattedChangedFiles := formatAsList(changedFiles)

	// Format recent commits as a list
	formattedRecentCommits := formatAsList(recentCommits)

	// Prepare data for template
	data := CommitPromptData{
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

// GetBranchSystemPrompt returns the system prompt for branch name generation
func GetBranchSystemPrompt() string {
	return branchSystemPromptTemplate
}

// GetBranchUserPrompt generates a user prompt for branch name generation
func GetBranchUserPrompt(request string, localBranches, remoteBranches []string) (string, error) {
	// Format branches as a list
	formattedLocalBranches := formatAsList(strings.Join(localBranches, "\n"))
	formattedRemoteBranches := formatAsList(strings.Join(remoteBranches, "\n"))

	// Prepare data for template
	data := BranchPromptData{
		Request:        request,
		LocalBranches:  formattedLocalBranches,
		RemoteBranches: formattedRemoteBranches,
	}

	// Define template functions
	funcMap := template.FuncMap{
		"trimSpace": strings.TrimSpace,
	}

	// Parse and execute the template
	tmpl, err := template.New("branch").Funcs(funcMap).Parse(branchUserPromptTemplate)
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
