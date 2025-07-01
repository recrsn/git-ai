# Git AI Project Guidelines

## Build and Test Commands
- Build: `go build`
- Run: `go run main.go`
- Test all: `go test ./...`
- Test specific: `go test -run TestFunctionName`
- Lint: `golint ./...`
- Vet: `go vet ./...`
- Format: `gofmt -s -w .`
- Install as git extension: `ln -s "$(pwd)/git-ai" /usr/local/bin/git-ai`

## Code Style Guidelines
- Follow standard Go conventions from [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Import order: standard library, third-party, local packages
- Error handling: check errors immediately, use descriptive error messages
- Naming: CamelCase for exported names, camelCase for unexported names
- Comments: package doc with `// Package name ...`, function doc with `// FuncName ...`
- Types: prefer explicit types to interface{}, use meaningful type names
- Prefer early returns over nested conditions
- Max line length: 100 characters
- Use context for cancellation and timeouts when appropriate
- Use cobra for command-line interfaces
- Use Bubble Tea (github.com/charmbracelet/bubbletea) for interactive terminal UIs
- Use playwright when trying to lookup information on the web
