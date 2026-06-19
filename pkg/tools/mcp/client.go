package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/houruizhe/go-claude-code/internal/protocol/mcp"
	"github.com/houruizhe/go-claude-code/pkg/tools"
)

// Client represents an MCP client
type Client struct {
	name      string
	rpc       *mcp.JSONRPC
	transport Transport
	tools     map[string]*ToolWrapper
	mu        sync.RWMutex
}

// ToolWrapper wraps an MCP tool to implement the Tool interface
type ToolWrapper struct {
	name        string
	description string
	inputSchema map[string]interface{}
	client      *Client
}

// NewClient creates a new MCP client
func NewClient(name string, transport Transport) *Client {
	return &Client{
		name:      name,
		transport: transport,
		tools:     make(map[string]*ToolWrapper),
	}
}

// Connect connects to the MCP server
func (c *Client) Connect(ctx context.Context) error {
	// Connect via transport
	if err := c.transport.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect transport: %w", err)
	}

	// Create JSON-RPC client
	c.rpc = mcp.NewJSONRPC(c.transport)

	// Send initialize request
	initParams := mcp.InitializeParams{
		ProtocolVersion: mcp.ProtocolVersion,
		Capabilities: mcp.ClientCapabilities{
			Roots: &mcp.RootsCapability{
				ListChanged: false,
			},
		},
		ClientInfo: mcp.ImplementationInfo{
			Name:    "go-claude-code",
			Version: "0.1.0",
		},
	}

	resp, err := c.rpc.SendRequest(mcp.MethodInitialize, initParams)
	if err != nil {
		c.Close()
		return fmt.Errorf("initialize failed: %w", err)
	}

	// Parse initialize result
	var initResult mcp.InitializeResult
	if err := unmarshalResult(resp.Result, &initResult); err != nil {
		c.Close()
		return fmt.Errorf("failed to parse initialize result: %w", err)
	}

	// Send initialized notification
	if err := c.rpc.SendNotification(mcp.MethodInitialized, nil); err != nil {
		c.Close()
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}

	// List available tools
	if err := c.listTools(); err != nil {
		c.Close()
		return fmt.Errorf("failed to list tools: %w", err)
	}

	return nil
}

// listTools lists available tools from the server
func (c *Client) listTools() error {
	resp, err := c.rpc.SendRequest(mcp.MethodListTools, nil)
	if err != nil {
		return err
	}

	var result mcp.ListToolsResult
	if err := unmarshalResult(resp.Result, &result); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, tool := range result.Tools {
		wrapper := &ToolWrapper{
			name:        tool.Name,
			description: tool.Description,
			inputSchema: tool.InputSchema,
			client:      c,
		}
		c.tools[tool.Name] = wrapper
	}

	return nil
}

// GetTool returns a tool by name
func (c *Client) GetTool(name string) (*ToolWrapper, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	tool, ok := c.tools[name]
	return tool, ok
}

// GetTools returns all tools
func (c *Client) GetTools() []*ToolWrapper {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tools := make([]*ToolWrapper, 0, len(c.tools))
	for _, tool := range c.tools {
		tools = append(tools, tool)
	}
	return tools
}

// CallTool calls a tool on the MCP server
func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	params := mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	}

	resp, err := c.rpc.SendRequest(mcp.MethodCallTool, params)
	if err != nil {
		return nil, err
	}

	var result mcp.CallToolResult
	if err := unmarshalResult(resp.Result, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Close closes the MCP client
func (c *Client) Close() error {
	// Send shutdown notification
	if c.rpc != nil {
		c.rpc.SendNotification(mcp.MethodShutdown, nil)
	}

	// Close transport
	if c.transport != nil {
		c.transport.Close()
	}

	// Close RPC
	if c.rpc != nil {
		c.rpc.Close()
	}

	return nil
}

// Name returns the client name
func (c *Client) Name() string {
	return c.name
}

// ToolWrapper methods

// Name returns the tool name
func (t *ToolWrapper) Name() string {
	return t.name
}

// Description returns the tool description
func (t *ToolWrapper) Description() string {
	return t.description
}

// Parameters returns the tool parameters schema
func (t *ToolWrapper) Parameters() map[string]interface{} {
	return t.inputSchema
}

// Execute executes the tool via the MCP client
func (t *ToolWrapper) Execute(ctx *tools.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	result, err := t.client.CallTool(context.Background(), t.name, args)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Helper function to unmarshal result
func unmarshalResult(result interface{}, v interface{}) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
