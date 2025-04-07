package config

// LLM providers
const (
	OpenAIProvider    = "openai"
	AnthropicProvider = "anthropic"
	OllamaProvider    = "ollama"
	OtherProvider     = "other"
)

type providerInfo struct {
	name            string
	endpoint        string
	availableModels []string
	defaultModel    string
}

var availableProviders = []string{
	OpenAIProvider,
	AnthropicProvider,
	OllamaProvider,
	OtherProvider,
}

var providers = map[string]providerInfo{
	OpenAIProvider: {
		name:            "OpenAI",
		endpoint:        "https://api.openai.com/v1",
		availableModels: []string{"gpt-3.5-turbo", "gpt-4", "gpt-4o-mini", "gpt-4o", "gpt-4-turbo", "custom"},
		defaultModel:    "gpt-4o-mini",
	},
	AnthropicProvider: {
		name:            "Anthropic",
		endpoint:        "https://api.anthropic.com/v1",
		availableModels: []string{"claude-3-7-sonnet-latest", "claude-3-5-sonnet-latest", "claude-3-5-haiku-latest", "custom"},
		defaultModel:    "claude-3-5-haiku-latest",
	},
	OllamaProvider: {
		name:            "Ollama",
		endpoint:        "http://localhost:11434/v1",
		availableModels: []string{"llama3", "mistral", "codellama", "custom"},
		defaultModel:    "llama3",
	},
	OtherProvider: {
		name:            "Other",
		availableModels: []string{"custom"},
		defaultModel:    "custom",
	},
}
