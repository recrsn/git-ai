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
		name:     "OpenAI",
		endpoint: "https://api.openai.com/v1",
		availableModels: []string{
			"gpt-5",
			"gpt-5-mini",
			"gpt-5-nano",
			"custom",
		},
		defaultModel: "gpt-5-mini",
	},
	AnthropicProvider: {
		name:     "Anthropic",
		endpoint: "https://api.anthropic.com",
		availableModels: []string{
			"claude-sonnet-4-5",
			"claude-haiku-4-5",
			"claude-sonnet-4",
			"claude-3-7-sonnet-latest",
			"claude-3-5-haiku-latest",
			"custom",
		},
		defaultModel: "claude-haiku-4-5",
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
