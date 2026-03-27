package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Executor handles tool execution
type Executor struct {
	registry *Registry
	timeout  time.Duration
}

// NewExecutor creates a new tool executor
func NewExecutor(registry *Registry, timeout time.Duration) *Executor {
	return &Executor{
		registry: registry,
		timeout:  timeout,
	}
}

// Execute executes a tool by name with the given arguments
func (e *Executor) Execute(ctx context.Context, name string, args map[string]interface{}, execCtx *ExecutionContext) (interface{}, error) {
	// Get tool from registry
	tool, exists := e.registry.Get(name)
	if !exists {
		return nil, fmt.Errorf("tool '%s' not found", name)
	}

	// Create execution context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Execute tool
	resultChan := make(chan toolResult, 1)
	go func() {
		result, err := tool.Execute(execCtx, args)
		resultChan <- toolResult{result: result, err: err}
	}()

	// Wait for result or timeout
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("tool execution timed out after %v", e.timeout)
	case res := <-resultChan:
		if res.err != nil {
			return nil, fmt.Errorf("tool execution failed: %w", res.err)
		}
		return res.result, nil
	}
}

// ExecuteToolCall executes a tool call from an API response
func (e *Executor) ExecuteToolCall(ctx context.Context, toolCall *ToolCall, execCtx *ExecutionContext) (*ToolResult, error) {
	result, err := e.Execute(ctx, toolCall.Name, toolCall.Arguments, execCtx)
	if err != nil {
		return &ToolResult{
			ToolUseID: toolCall.ID,
			Content:    err.Error(),
			IsError:    true,
		}, nil
	}

	// Convert result to string
	content, err := e.formatResult(result)
	if err != nil {
		return &ToolResult{
			ToolUseID: toolCall.ID,
			Content:    fmt.Sprintf("failed to format result: %v", result),
			IsError:    true,
		}, nil
	}

	return &ToolResult{
		ToolUseID: toolCall.ID,
		Content:    content,
		IsError:    false,
	}, nil
}

// formatResult formats a tool result as a string
func (e *Executor) formatResult(result interface{}) (string, error) {
	if result == nil {
		return "", nil
	}

	// Try to convert to JSON
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", result), nil
	}
	return string(data), nil
}

// ExecuteBatch executes multiple tools in parallel
func (e *Executor) ExecuteBatch(ctx context.Context, toolCalls []*ToolCall, execCtx *ExecutionContext) ([]*ToolResult, error) {
	results := make([]*ToolResult, len(toolCalls))
	errChan := make(chan error, len(toolCalls))
	doneChan := make(chan int, len(toolCalls))

	for i, toolCall := range toolCalls {
		go func(index int, tc *ToolCall) {
			result, err := e.ExecuteToolCall(ctx, tc, execCtx)
			if err != nil {
				errChan <- err
				return
			}
			results[index] = result
			doneChan <- index
		}(i, toolCall)
	}

	// Wait for all tools to complete
	completed := 0
	for completed < len(toolCalls) {
		select {
		case err := <-errChan:
			return nil, err
		case <-doneChan:
			completed++
		}
	}

	return results, nil
}

// ToolCall represents a tool call from the API
type ToolCall struct {
	ID       string
	Name     string
	Arguments map[string]interface{}
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolUseID string
	Content   string
	IsError   bool
}

// ToAPIMessage converts a tool result to an API message content block
func (r *ToolResult) ToAPIMessage() map[string]interface{} {
	return map[string]interface{}{
		"type":        "tool_result",
		"tool_use_id": r.ToolUseID,
		"content":     r.Content,
		"is_error":    r.IsError,
	}
}

type toolResult struct {
	result interface{}
	err    error
}
