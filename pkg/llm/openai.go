package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIProvider implements the Provider interface for OpenAI-compatible APIs
type OpenAIProvider struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// OpenAIMessage represents a message in OpenAI's format
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIRequest represents a request to the OpenAI chat completion API
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
}

// OpenAIChoice represents a choice returned by the API
type OpenAIChoice struct {
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// OpenAIResponse represents a response from the OpenAI chat completion API
type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Choices []OpenAIChoice `json:"choices"`
	Error   *OpenAIError   `json:"error,omitempty"`
}

// OpenAIError represents an error returned by the API
type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// NewOpenAIProvider creates a new OpenAI-compatible provider
func NewOpenAIProvider(endpoint, apiKey string) (*OpenAIProvider, error) {
	return &OpenAIProvider{
		BaseURL: endpoint,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// ChatCompletion sends a chat completion request to the OpenAI-compatible API
func (p *OpenAIProvider) ChatCompletion(model string, messages []Message) (string, error) {
	// Convert to OpenAI message format
	openaiMessages := make([]OpenAIMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = OpenAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	req := OpenAIRequest{
		Model:       model,
		Messages:    openaiMessages,
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", p.BaseURL), bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.APIKey))

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

	var respData OpenAIResponse
	if err := json.Unmarshal(body, &respData); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w, body: %s", err, string(body))
	}

	if respData.Error != nil {
		return "", fmt.Errorf("API error: %s", respData.Error.Message)
	}

	if len(respData.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	return respData.Choices[0].Message.Content, nil
}
