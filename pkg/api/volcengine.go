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
)

const (
	// DefaultVolcengineBaseURL is the default Volcengine API base URL
	DefaultVolcengineBaseURL = "https://ark.cn-beijing.volces.com/api/v3"
	// DefaultVolcengineModel is the default Volcengine model
	DefaultVolcengineModel = "ep-20241125174025-l8jx8"
)

// VolcengineClient is a client for the Volcengine API (Doubao)
type VolcengineClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	model      string
	maxTokens  int
}

// VolcengineClientOption is a function that configures a VolcengineClient
type VolcengineClientOption func(*VolcengineClient)

// WithVolcengineAPIKey sets the API key for the Volcengine client
func WithVolcengineAPIKey(apiKey string) VolcengineClientOption {
	return func(c *VolcengineClient) {
		c.apiKey = apiKey
	}
}

// WithVolcengineBaseURL sets the base URL for the Volcengine client
func WithVolcengineBaseURL(baseURL string) VolcengineClientOption {
	return func(c *VolcengineClient) {
		c.baseURL = baseURL
	}
}

// WithVolcengineModel sets the default model for the Volcengine client
func WithVolcengineModel(model string) VolcengineClientOption {
	return func(c *VolcengineClient) {
		c.model = model
	}
}

// WithVolcengineMaxTokens sets the default max tokens for the Volcengine client
func WithVolcengineMaxTokens(maxTokens int) VolcengineClientOption {
	return func(c *VolcengineClient) {
		c.maxTokens = maxTokens
	}
}

// NewVolcengineClient creates a new Volcengine API client
func NewVolcengineClient(options ...VolcengineClientOption) *VolcengineClient {
	c := &VolcengineClient{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL:   DefaultVolcengineBaseURL,
		model:     DefaultVolcengineModel,
		maxTokens: DefaultMaxTokens,
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

// VolcengineChatRequest represents a request to the Volcengine API
type VolcengineChatRequest struct {
	Model       string              `json:"model"`
	Messages    []VolcengineMessage `json:"messages"`
	MaxTokens   int                 `json:"max_tokens"`
	Temperature float64             `json:"temperature,omitempty"`
	Tools       []ToolDefinition    `json:"tools,omitempty"`
	Stream      bool                `json:"stream"`
}

// VolcengineMessage represents a message in the Volcengine conversation
type VolcengineMessage struct {
	Role    string                   `json:"role"`
	Content []VolcengineContentBlock `json:"content"`
}

// VolcengineContentBlock represents a content block in a Volcengine message
type VolcengineContentBlock struct {
	Type       string               `json:"type"`
	Text       string               `json:"text,omitempty"`
	ToolCallID string               `json:"tool_call_id,omitempty"`
	ToolCalls  []VolcengineToolCall `json:"tool_calls,omitempty"`
}

// VolcengineToolCall represents a tool call in Volcengine
type VolcengineToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function VolcengineFunctionCall `json:"function"`
}

// VolcengineFunctionCall represents a function call
type VolcengineFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// StreamChat sends a chat request and returns a channel of events
func (c *VolcengineClient) StreamChat(ctx context.Context, req *ChatRequest) (<-chan *ChatEvent, error) {
	// Convert to Volcengine request format
	volcReq := c.convertRequest(req)

	// Marshal request body
	reqBody, err := json.Marshal(volcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
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

// convertRequest converts a ChatRequest to VolcengineChatRequest
func (c *VolcengineClient) convertRequest(req *ChatRequest) *VolcengineChatRequest {
	// Set defaults
	model := c.model
	if req.Model != "" {
		model = req.Model
	}

	maxTokens := c.maxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	// Convert messages
	messages := make([]VolcengineMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = VolcengineMessage{
			Role:    msg.Role,
			Content: c.convertContentBlocks(msg.Content),
		}
	}

	return &VolcengineChatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
		Tools:       req.Tools,
		Stream:      true,
	}
}

// convertContentBlocks converts ContentBlocks to VolcengineContentBlocks
func (c *VolcengineClient) convertContentBlocks(blocks []ContentBlock) []VolcengineContentBlock {
	result := make([]VolcengineContentBlock, len(blocks))
	for i, block := range blocks {
		if block.Type == "text" && block.Text != nil {
			result[i] = VolcengineContentBlock{
				Type: "text",
				Text: *block.Text,
			}
		} else if block.Type == "tool_result" {
			result[i] = VolcengineContentBlock{
				Type:       "tool_result",
				ToolCallID: block.ToolUseID,
				Text:       block.Content,
			}
		}
	}
	return result
}

// streamResponse reads the SSE stream and sends events to the channel
func (c *VolcengineClient) streamResponse(ctx context.Context, body io.ReadCloser, eventChan chan<- *ChatEvent) {
	defer close(eventChan)
	defer body.Close()

	scanner := bufio.NewScanner(body)
	var currentContent string
	var currentToolCalls []VolcengineToolCall

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
		var sseEvent struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			Model   string `json:"model"`
			Choices []struct {
				Index int `json:"index"`
				Delta struct {
					Content   string               `json:"content,omitempty"`
					ToolCalls []VolcengineToolCall `json:"tool_calls,omitempty"`
					Role      string               `json:"role,omitempty"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(data), &sseEvent); err != nil {
			continue
		}

		if len(sseEvent.Choices) > 0 {
			choice := sseEvent.Choices[0]

			// Handle text content
			if choice.Delta.Content != "" {
				currentContent += choice.Delta.Content
				eventChan <- &ChatEvent{
					Type: "text",
					Content: &ContentBlock{
						Type: "text",
						Text: &choice.Delta.Content,
					},
				}
			}

			// Handle tool calls
			if len(choice.Delta.ToolCalls) > 0 {
				currentToolCalls = append(currentToolCalls, choice.Delta.ToolCalls...)
			}

			// Check for finish
			if choice.FinishReason != "" {
				// Send tool calls if any
				if len(currentToolCalls) > 0 {
					for _, tc := range currentToolCalls {
						var args map[string]interface{}
						if tc.Function.Arguments != "" {
							json.Unmarshal([]byte(tc.Function.Arguments), &args)
						}
						eventChan <- &ChatEvent{
							Type: "tool_use",
							Content: &ContentBlock{
								Type:  "tool_use",
								ID:    tc.ID,
								Name:  tc.Function.Name,
								Input: args,
							},
						}
					}
				}
				eventChan <- &ChatEvent{Done: true}
				return
			}
		}
	}

	// Check for scan errors
	if err := scanner.Err(); err != nil {
		eventChan <- &ChatEvent{Error: fmt.Errorf("stream error: %w", err)}
	}
}
