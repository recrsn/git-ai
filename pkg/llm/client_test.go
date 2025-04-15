package llm

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChatCompletionSuccess(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("Expected path /chat/completions, got %s", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Errorf("Expected Authorization: Bearer test-api-key, got %s", r.Header.Get("Authorization"))
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
		}

		// Verify request body
		var req ChatCompletionRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("Failed to unmarshal request: %v", err)
		}
		if req.Model != "gpt-4o" {
			t.Errorf("Expected model gpt-4o, got %s", req.Model)
		}
		if len(req.Messages) != 1 || req.Messages[0].Role != "user" || req.Messages[0].Content != "Hello" {
			t.Errorf("Unexpected messages: %+v", req.Messages)
		}

		// Return successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := ChatCompletionResponse{
			ID:      "chatcmpl-123",
			Object:  "chat.completion",
			Created: 1677652288,
			Choices: []ChatCompletionChoice{
				{
					Message: Message{
						Role:    "assistant",
						Content: "Hello, how can I assist you today?",
					},
					FinishReason: "stop",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client using test server URL
	client, _ := NewClient(server.URL, "test-api-key")

	// Test the ChatCompletion function
	messages := []Message{
		{
			Role:    "user",
			Content: "Hello",
		},
	}
	content, err := client.ChatCompletion("gpt-4o", messages)
	if err != nil {
		t.Fatalf("ChatCompletion failed: %v", err)
	}

	expectedContent := "Hello, how can I assist you today?"
	if content != expectedContent {
		t.Errorf("Expected content '%s', got '%s'", expectedContent, content)
	}
}

func TestChatCompletionInvalidJSON(t *testing.T) {
	// Setup test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test-api-key")
	messages := []Message{
		{
			Role:    "user",
			Content: "Hello",
		},
	}
	_, err := client.ChatCompletion("gpt-4o", messages)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to unmarshal response") {
		t.Errorf("Expected unmarshal error, got: %v", err)
	}
}

func TestChatCompletionNon200Response(t *testing.T) {
	// Setup test server that returns a non-200 response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Bad request"}`))
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test-api-key")
	messages := []Message{
		{
			Role:    "user",
			Content: "Hello",
		},
	}
	_, err := client.ChatCompletion("gpt-4o", messages)
	if err == nil {
		t.Error("Expected error for non-200 response, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected status code: 400") {
		t.Errorf("Expected status code error, got: %v", err)
	}
}

func TestChatCompletionAPIError(t *testing.T) {
	// Setup test server that returns an API error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"created": 1677652288,
			"choices": [],
			"error": {
				"message": "Invalid API key",
				"type": "invalid_request_error",
				"code": "invalid_api_key"
			}
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test-api-key")
	messages := []Message{
		{
			Role:    "user",
			Content: "Hello",
		},
	}
	_, err := client.ChatCompletion("gpt-4o", messages)
	if err == nil {
		t.Error("Expected error for API error, got nil")
	}
	if !strings.Contains(err.Error(), "API error: Invalid API key") {
		t.Errorf("Expected API error, got: %v", err)
	}
}

func TestChatCompletionNoChoices(t *testing.T) {
	// Setup test server that returns no choices
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"created": 1677652288,
			"choices": []
		}`))
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test-api-key")
	messages := []Message{
		{
			Role:    "user",
			Content: "Hello",
		},
	}
	_, err := client.ChatCompletion("gpt-4o", messages)
	if err == nil {
		t.Error("Expected error for no choices, got nil")
	}
	if !strings.Contains(err.Error(), "no choices returned") {
		t.Errorf("Expected no choices error, got: %v", err)
	}
}

func TestChatCompletionRequestError(t *testing.T) {
	// Create client with invalid URL to force request error
	client, _ := NewClient("http://invalid-url-that-does-not-exist.example", "test-api-key")
	messages := []Message{
		{
			Role:    "user",
			Content: "Hello",
		},
	}
	_, err := client.ChatCompletion("gpt-4o", messages)
	if err == nil {
		t.Error("Expected error for request error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to send request") {
		t.Errorf("Expected request error, got: %v", err)
	}
}

type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestChatCompletionReadResponseError(t *testing.T) {
	// Setup test server that returns a response with a body that errors on read
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// The actual response doesn't matter as we'll mock the error
	}))
	defer server.Close()

	// Create a custom client with a transport that returns a reader that always errors
	originalClient := &http.Client{
		Transport: &errorRoundTripper{},
	}

	client := &Client{
		BaseURL:    server.URL,
		APIKey:     "test-api-key",
		HTTPClient: originalClient,
	}

	messages := []Message{
		{
			Role:    "user",
			Content: "Hello",
		},
	}
	_, err := client.ChatCompletion("gpt-4o", messages)
	if err == nil {
		t.Error("Expected error for read response error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read response body") {
		t.Errorf("Expected read response error, got: %v", err)
	}
}

// Custom round tripper that returns a response with a body that errors on read
type errorRoundTripper struct{}

func (e *errorRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(errorReader{}),
	}, nil
}
