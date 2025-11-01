package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AnthropicProvider implements the Provider interface for Anthropic's Messages API
type AnthropicProvider struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// AnthropicMessage represents a message in Anthropic's format
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicRequest represents a request to Anthropic's Messages API
type AnthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []AnthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature,omitempty"`
	System      string             `json:"system,omitempty"`
}

// AnthropicContent represents content in the response
type AnthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AnthropicResponse represents a response from Anthropic's Messages API
type AnthropicResponse struct {
	ID      string             `json:"id"`
	Type    string             `json:"type"`
	Role    string             `json:"role"`
	Content []AnthropicContent `json:"content"`
	Model   string             `json:"model"`
	Error   *AnthropicError    `json:"error,omitempty"`
}

// AnthropicError represents an error from Anthropic's API
type AnthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(endpoint, apiKey string) (*AnthropicProvider, error) {
	return &AnthropicProvider{
		BaseURL: endpoint,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// ChatCompletion sends a chat completion request to Anthropic's Messages API
func (p *AnthropicProvider) ChatCompletion(model string, messages []Message) (string, error) {
	// Separate system message from conversation messages
	var systemPrompt string
	var anthropicMessages []AnthropicMessage

	for _, msg := range messages {
		if msg.Role == "system" {
			// Anthropic uses a separate system field
			systemPrompt = msg.Content
		} else {
			anthropicMessages = append(anthropicMessages, AnthropicMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	req := AnthropicRequest{
		Model:       model,
		Messages:    anthropicMessages,
		MaxTokens:   4096,
		Temperature: 0.7,
	}

	if systemPrompt != "" {
		req.System = systemPrompt
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/messages", p.BaseURL), bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.HTTPClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var respData AnthropicResponse
	if err := json.Unmarshal(body, &respData); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w, body: %s", err, string(body))
	}

	if respData.Error != nil {
		return "", fmt.Errorf("API error: %s", respData.Error.Message)
	}

	if len(respData.Content) == 0 {
		return "", fmt.Errorf("no content returned")
	}

	// Concatenate all text content blocks
	var result string
	for _, content := range respData.Content {
		if content.Type == "text" {
			result += content.Text
		}
	}

	return result, nil
}
