package api

import "encoding/json"

// Message represents a message in the conversation
type Message struct {
	Role    string       `json:"role"`
	Content []ContentBlock `json:"content"`
}

// ContentBlock represents a block of content in a message
type ContentBlock struct {
	Type string `json:"type"`
	// For text blocks
	Text *string `json:"text,omitempty"`
	// For tool_use blocks
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Input       map[string]interface{} `json:"input,omitempty"`
	// For tool_result blocks
	ToolUseID   string `json:"tool_use_id,omitempty"`
	Content     string `json:"content,omitempty"`
	IsError     bool   `json:"is_error,omitempty"`
}

// ToolDefinition defines a tool that can be called by Claude
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"input_schema"`
}

// InputSchema defines the JSON schema for tool inputs
type InputSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]PropertySchema `json:"properties"`
	Required   []string                  `json:"required,omitempty"`
}

// PropertySchema defines a property in the input schema
type PropertySchema struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// ChatRequest represents a request to the Claude API
type ChatRequest struct {
	Model         string           `json:"model"`
	Messages      []Message        `json:"messages"`
	MaxTokens     int              `json:"max_tokens"`
	Temperature   float64          `json:"temperature,omitempty"`
	Tools         []ToolDefinition `json:"tools,omitempty"`
	Stream        bool             `json:"stream"`
	System        string           `json:"system,omitempty"`
}

// ChatResponse represents a response from the Claude API
type ChatResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Role         string `json:"role"`
	Content      []ContentBlock `json:"content"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence,omitempty"`
	Usage        Usage  `json:"usage"`
}

// Usage represents token usage information
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// SSEEvent represents a Server-Sent Event from the Claude API
type SSEEvent struct {
	Type    string          `json:"type"`
	Index   int             `json:"index,omitempty"`
	Delta   *ContentDelta   `json:"delta,omitempty"`
	Message *ChatResponse   `json:"message,omitempty"`
	Usage   *Usage          `json:"usage,omitempty"`
	Error   *ErrorResponse  `json:"error,omitempty"`
}

// ContentDelta represents a delta in content during streaming
type ContentDelta struct {
	Type         string `json:"type"`
	Text         string `json:"text,omitempty"`
	PartialJSON  string `json:"partial_json,omitempty"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ChatEvent represents an event during streaming
type ChatEvent struct {
	Type    string
	Content *ContentBlock
	Error   error
	Done    bool
}

// UnmarshalJSON implements custom JSON unmarshaling for SSEEvent
func (e *SSEEvent) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	e.Type, _ = raw["type"].(string)

	switch e.Type {
	case "content_block_start":
		if contentBlock, ok := raw["content_block"].(map[string]interface{}); ok {
			blockType, _ := contentBlock["type"].(string)
			cb := &ContentBlock{Type: blockType}
			if blockType == "text" {
				if text, ok := contentBlock["text"].(string); ok {
					cb.Text = &text
				}
			} else if blockType == "tool_use" {
				cb.ID, _ = contentBlock["id"].(string)
				cb.Name, _ = contentBlock["name"].(string)
				if input, ok := contentBlock["input"].(map[string]interface{}); ok {
					cb.Input = input
				}
			}
			e.Delta = &ContentDelta{Type: blockType}
		}
	case "content_block_delta":
		if delta, ok := raw["delta"].(map[string]interface{}); ok {
			e.Delta = &ContentDelta{}
			if deltaType, ok := delta["type"].(string); ok {
				e.Delta.Type = deltaType
			}
			if text, ok := delta["text"].(string); ok {
				e.Delta.Text = text
			}
			if partialJSON, ok := delta["partial_json"].(string); ok {
				e.Delta.PartialJSON = partialJSON
			}
		}
	case "message_start":
		if message, ok := raw["message"].(map[string]interface{}); ok {
			e.Message = &ChatResponse{}
			if id, ok := message["id"].(string); ok {
				e.Message.ID = id
			}
			if role, ok := message["role"].(string); ok {
				e.Message.Role = role
			}
		}
	case "message_delta":
		if delta, ok := raw["delta"].(map[string]interface{}); ok {
			e.Delta = &ContentDelta{}
			if stopReason, ok := delta["stop_reason"].(string); ok {
				e.Delta.Type = stopReason
			}
		}
	case "message_stop":
		// No additional data needed
	case "error":
		if errorData, ok := raw["error"].(map[string]interface{}); ok {
			e.Error = &ErrorResponse{
				Type:    errorData["type"].(string),
				Message: errorData["message"].(string),
			}
		}
	}

	return nil
}
