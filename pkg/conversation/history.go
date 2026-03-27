package conversation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/houruizhe/go-claude-code/pkg/api"
)

// History manages conversation history
type History struct {
	messages []api.Message
	filePath string
	mu       sync.RWMutex
}

// NewHistory creates a new conversation history
func NewHistory(filePath string) *History {
	return &History{
		messages: make([]api.Message, 0),
		filePath: filePath,
	}
}

// Load loads conversation history from disk
func (h *History) Load() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.filePath == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(h.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	// Read file
	data, err := os.ReadFile(h.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's OK
		}
		return fmt.Errorf("failed to read history file: %w", err)
	}

	// Parse JSON
	if err := json.Unmarshal(data, &h.messages); err != nil {
		return fmt.Errorf("failed to parse history: %w", err)
	}

	return nil
}

// Save saves conversation history to disk
func (h *History) Save() error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.filePath == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(h.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(h.messages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	// Write file
	if err := os.WriteFile(h.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// Add adds a message to the history
func (h *History) Add(message api.Message) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.messages = append(h.messages, message)
}

// AddUserMessage adds a user message to the history
func (h *History) AddUserMessage(content string) {
	h.Add(api.Message{
		Role: "user",
		Content: []api.ContentBlock{
			{Type: "text", Text: &content},
		},
	})
}

// AddAssistantMessage adds an assistant message to the history
func (h *History) AddAssistantMessage(content string) {
	h.Add(api.Message{
		Role: "assistant",
		Content: []api.ContentBlock{
			{Type: "text", Text: &content},
		},
	})
}

// GetMessages returns all messages
func (h *History) GetMessages() []api.Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return a copy to avoid race conditions
	messages := make([]api.Message, len(h.messages))
	copy(messages, h.messages)
	return messages
}

// GetLastN returns the last N messages
func (h *History) GetLastN(n int) []api.Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if n <= 0 {
		return []api.Message{}
	}

	start := len(h.messages) - n
	if start < 0 {
		start = 0
	}

	// Return a copy to avoid race conditions
	messages := make([]api.Message, len(h.messages)-start)
	copy(messages, h.messages[start:])
	return messages
}

// Clear clears the conversation history
func (h *History) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.messages = make([]api.Message, 0)
}

// Count returns the number of messages
func (h *History) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.messages)
}

// Prune removes old messages to keep the history within a limit
func (h *History) Prune(maxMessages int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.messages) <= maxMessages {
		return
	}

	// Keep the most recent messages
	h.messages = h.messages[len(h.messages)-maxMessages:]
}

// EstimateTokenCount estimates the token count of the conversation
func (h *History) EstimateTokenCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Rough estimate: 1 token ≈ 4 characters
	totalChars := 0
	for _, msg := range h.messages {
		for _, block := range msg.Content {
			if block.Text != nil {
				totalChars += len(*block.Text)
			}
		}
	}

	return totalChars / 4
}

// GetTimestamp returns the timestamp of the last message
func (h *History) GetTimestamp() time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.messages) == 0 {
		return time.Time{}
	}

	return time.Now() // In a real implementation, we'd store timestamps
}

// Export exports the conversation history to a file
func (h *History) Export(filePath string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.MarshalIndent(h.messages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}
