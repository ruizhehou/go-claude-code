package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/houruizhe/go-claude-code/pkg/api"
	"github.com/houruizhe/go-claude-code/pkg/config"
	"github.com/houruizhe/go-claude-code/pkg/tools"
)

// REPL represents the Read-Eval-Print Loop
type REPL struct {
	config   *config.Config
	client   *api.Client
	registry *tools.Registry
	executor *tools.Executor

	messages []api.Message
	reader   *bufio.Reader
}

// NewREPL creates a new REPL
func NewREPL(cfg *config.Config, client *api.Client, registry *tools.Registry, executor *tools.Executor) *REPL {
	return &REPL{
		config:   cfg,
		client:   client,
		registry: registry,
		executor: executor,
		messages: make([]api.Message, 0),
		reader:   bufio.NewReader(os.Stdin),
	}
}

// Run starts the REPL loop
func (r *REPL) Run(ctx context.Context) error {
	// Print welcome message
	fmt.Println("Claude Code - Go Implementation")
	fmt.Println("Type 'exit' or 'quit' to exit, 'clear' to clear history")
	fmt.Println()

	// Get current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}

	// Create execution context
	execCtx := &tools.ExecutionContext{
		WorkingDir: workingDir,
		Env:        make(map[string]string),
	}

	// Main loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Read input
		input, err := r.readLine()
		if err != nil {
			return err
		}

		// Handle empty input
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle special commands
		if r.handleCommand(input) {
			continue
		}

		// Add user message to history
		r.messages = append(r.messages, api.Message{
			Role: "user",
			Content: []api.ContentBlock{
				{Type: "text", Text: &input},
			},
		})

		// Send to API
		if err := r.chat(ctx, execCtx); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}
}

// readLine reads a line from stdin
func (r *REPL) readLine() (string, error) {
	fmt.Print(r.config.CLI.Prompt)
	line, err := r.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(line, "\n"), nil
}

// handleCommand handles special REPL commands
func (r *REPL) handleCommand(input string) bool {
	switch strings.ToLower(input) {
	case "exit", "quit":
		os.Exit(0)
		return true
	case "clear":
		r.messages = make([]api.Message, 0)
		fmt.Println("Conversation history cleared")
		return true
	case "help":
		r.printHelp()
		return true
	case "tools":
		r.printTools()
		return true
	default:
		return false
	}
}

// printHelp prints help information
func (r *REPL) printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  exit, quit  - Exit the program")
	fmt.Println("  clear       - Clear conversation history")
	fmt.Println("  help        - Show this help message")
	fmt.Println("  tools       - List available tools")
}

// printTools prints available tools
func (r *REPL) printTools() {
	fmt.Println("Available tools:")
	for _, tool := range r.registry.List() {
		fmt.Printf("  - %s: %s\n", tool.Name(), tool.Description())
	}
}

// chat sends a message to Claude and handles the response
func (r *REPL) chat(ctx context.Context, execCtx *tools.ExecutionContext) error {
	// Prepare request
	req := &api.ChatRequest{
		Messages:    r.messages,
		Tools:       r.registry.ToAPIDefinitions(),
		Temperature: r.config.Anthropic.Temperature,
	}

	// Stream response
	eventChan, err := r.client.StreamChat(ctx, req)
	if err != nil {
		return err
	}

	// Process events
	var currentMessage api.Message
	currentMessage.Role = "assistant"
	currentMessage.Content = []api.ContentBlock{}
	var currentContent string
	var currentTool *api.ContentBlock

	for event := range eventChan {
		if event.Error != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", event.Error)
			return event.Error
		}

		switch event.Type {
		case "text":
			if event.Content != nil && event.Content.Text != nil {
				fmt.Print(*event.Content.Text)
				currentContent += *event.Content.Text
			}

		case "tool_use":
			if event.Content != nil {
				currentTool = event.Content
				fmt.Printf("\n[Tool: %s]\n", currentTool.Name)
			}

		case "input_json_delta":
			if currentTool != nil && event.Content != nil {
				// Tool arguments are being streamed
			}
		}

		if event.Done {
			break
		}
	}

	// Add text content to message
	if currentContent != "" {
		currentMessage.Content = append(currentMessage.Content, api.ContentBlock{
			Type: "text",
			Text: &currentContent,
		})
	}

	// Add tool use to message
	if currentTool != nil {
		currentMessage.Content = append(currentMessage.Content, *currentTool)

		// Execute tool
		fmt.Println()
		result, err := r.executeTool(ctx, execCtx, currentTool)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Tool execution error: %v\n", err)
		} else {
			// Add tool result to messages and continue conversation
			r.messages = append(r.messages, currentMessage)
			r.messages = append(r.messages, api.Message{
				Role: "user",
				Content: []api.ContentBlock{
					{
						Type:       "tool_result",
						ToolUseID:  currentTool.ID,
						Content:    result,
						IsError:    false,
					},
				},
			})

			// Continue conversation with tool result
			return r.chat(ctx, execCtx)
		}
	}

	// Add assistant message to history
	if len(currentMessage.Content) > 0 {
		r.messages = append(r.messages, currentMessage)
	}

	fmt.Println()
	return nil
}

// executeTool executes a tool
func (r *REPL) executeTool(ctx context.Context, execCtx *tools.ExecutionContext, tool *api.ContentBlock) (string, error) {
	toolCall := &tools.ToolCall{
		ID:        tool.ID,
		Name:      tool.Name,
		Arguments: tool.Input,
	}

	result, err := r.executor.ExecuteToolCall(ctx, toolCall, execCtx)
	if err != nil {
		return "", err
	}

	return result.Content, nil
}
