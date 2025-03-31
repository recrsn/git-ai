package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a client for OpenAI-compatible API services
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents a request to the chat completion API
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// ChatCompletionChoice represents a choice returned by the API
type ChatCompletionChoice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// ChatCompletionResponse represents a response from the chat completion API
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Choices []ChatCompletionChoice `json:"choices"`
	Error   *ErrorResponse         `json:"error,omitempty"`
}

// ErrorResponse represents an error returned by the API
type ErrorResponse struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// NewClient creates a new client for OpenAI-compatible API services
func NewClient(endpoint, apiKey string) (*Client, error) {
	return &Client{
		BaseURL: endpoint,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// ChatCompletion sends a chat completion request to the API
func (c *Client) ChatCompletion(model string, messages []Message) (string, error) {
	req := ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", c.BaseURL), bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	resp, err := c.HTTPClient.Do(httpReq)
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

	var respData ChatCompletionResponse
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
