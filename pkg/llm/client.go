package llm

import (
	"fmt"
)

// Client represents a unified client that delegates to provider-specific implementations
type Client struct {
	provider Provider
}

// NewClient creates a new LLM client with the appropriate provider based on the endpoint
// Deprecated: Use NewClientWithProvider to explicitly specify the provider
func NewClient(endpoint, apiKey string) (*Client, error) {
	return NewClientWithProvider("openai", endpoint, apiKey)
}

// NewClientWithProvider creates a new LLM client with an explicit provider type
// providerType must be one of: "anthropic", "openai", "ollama", "other"
// endpoint can override the default endpoint for the provider
func NewClientWithProvider(providerType, endpoint, apiKey string) (*Client, error) {
	var provider Provider
	var err error

	if providerType == "" {
		return nil, fmt.Errorf("provider type is required")
	}

	switch providerType {
	case "anthropic":
		provider, err = NewAnthropicProvider(endpoint, apiKey)
	case "openai", "ollama", "other":
		// OpenAI provider works for OpenAI, Ollama, and other OpenAI-compatible APIs
		provider, err = NewOpenAIProvider(endpoint, apiKey)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerType)
	}

	if err != nil {
		return nil, err
	}

	return &Client{
		provider: provider,
	}, nil
}

// ChatCompletion sends a chat completion request via the configured provider
func (c *Client) ChatCompletion(model string, messages []Message) (string, error) {
	return c.provider.ChatCompletion(model, messages)
}
