package llm

// Provider defines the interface for LLM API providers
type Provider interface {
	// ChatCompletion sends a chat completion request and returns the response
	ChatCompletion(model string, messages []Message) (string, error)
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
