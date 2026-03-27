package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the default Anthropic API base URL
	DefaultBaseURL = "https://api.anthropic.com/v1"
	// DefaultModel is the default Claude model to use
	DefaultModel = "claude-sonnet-4-20250514"
	// DefaultMaxTokens is the default max tokens for responses
	DefaultMaxTokens = 4096
	// DefaultTimeout is the default timeout for API requests
	DefaultTimeout = 2 * time.Minute
)

// Client is a client for the Anthropic Claude API
type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	model      string
	maxTokens  int
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithAPIKey sets the API key for the client
func WithAPIKey(apiKey string) ClientOption {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

// WithBaseURL sets the base URL for the client
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithModel sets the default model for the client
func WithModel(model string) ClientOption {
	return func(c *Client) {
		c.model = model
	}
}

// WithMaxTokens sets the default max tokens for the client
func WithMaxTokens(maxTokens int) ClientOption {
	return func(c *Client) {
		c.maxTokens = maxTokens
	}
}

// WithHTTPClient sets the HTTP client for the client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient creates a new Claude API client
func NewClient(options ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL:   DefaultBaseURL,
		model:     DefaultModel,
		maxTokens: DefaultMaxTokens,
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

// StreamChat sends a chat request and returns a channel of events
func (c *Client) StreamChat(ctx context.Context, req *ChatRequest) (<-chan *ChatEvent, error) {
	// Set defaults
	if req.Model == "" {
		req.Model = c.model
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = c.maxTokens
	}
	req.Stream = true

	// Marshal request body
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/messages", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Create event channel
	eventChan := make(chan *ChatEvent, 10)

	// Start streaming
	go c.streamResponse(ctx, resp.Body, eventChan)

	return eventChan, nil
}

// streamResponse reads the SSE stream and sends events to the channel
func (c *Client) streamResponse(ctx context.Context, body io.ReadCloser, eventChan chan<- *ChatEvent) {
	defer close(eventChan)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	var currentContent *ContentBlock
	var currentToolUse *ContentBlock

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			eventChan <- &ChatEvent{Error: ctx.Err()}
			return
		default:
		}

		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for SSE event format
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Extract JSON data
		data := strings.TrimPrefix(line, "data: ")

		// Check for done signal
		if data == "[DONE]" {
			eventChan <- &ChatEvent{Done: true}
			return
		}

		// Parse SSE event
		var sseEvent SSEEvent
		if err := json.Unmarshal([]byte(data), &sseEvent); err != nil {
			continue
		}

		// Handle different event types
		switch sseEvent.Type {
		case "content_block_start":
			if sseEvent.Delta != nil {
				if sseEvent.Delta.Type == "text" {
					currentContent = &ContentBlock{Type: "text", Text: new(string)}
				} else if sseEvent.Delta.Type == "tool_use" {
					currentToolUse = &ContentBlock{Type: "tool_use"}
				}
			}

		case "content_block_delta":
			if sseEvent.Delta != nil {
				if sseEvent.Delta.Type == "text_delta" && currentContent != nil {
					*currentContent.Text += sseEvent.Delta.Text
					eventChan <- &ChatEvent{
						Type:    "text",
						Content: currentContent,
					}
				} else if sseEvent.Delta.Type == "input_json_delta" && currentToolUse != nil {
					if currentToolUse.Input == nil {
						currentToolUse.Input = make(map[string]interface{})
					}
					var partial map[string]interface{}
					if err := json.Unmarshal([]byte(sseEvent.Delta.PartialJSON), &partial); err == nil {
						for k, v := range partial {
							currentToolUse.Input[k] = v
						}
					}
				}
			}

		case "content_block_stop":
			if currentToolUse != nil {
				eventChan <- &ChatEvent{
					Type:    "tool_use",
					Content: currentToolUse,
				}
				currentToolUse = nil
			}

		case "message_delta":
			if sseEvent.Delta != nil && sseEvent.Delta.Type == "stop" {
				eventChan <- &ChatEvent{Done: true}
				return
			}

		case "error":
			if sseEvent.Error != nil {
				eventChan <- &ChatEvent{
					Error: fmt.Errorf("API error: %s", sseEvent.Error.Message),
				}
				return
			}
		}
	}

	// Check for scan errors
	if err := scanner.Err(); err != nil {
		eventChan <- &ChatEvent{Error: fmt.Errorf("stream error: %w", err)}
	}
}

// SendToolResult sends a tool result back to the API
func (c *Client) SendToolResult(ctx context.Context, toolUseID, content string, isError bool) (<-chan *ChatEvent, error) {
	// This would need to be part of a continued conversation
	// For now, we'll need to track conversation state elsewhere
	return nil, fmt.Errorf("SendToolResult requires conversation state management")
}
