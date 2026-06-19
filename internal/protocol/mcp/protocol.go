package mcp

const (
	// ProtocolVersion is the MCP protocol version
	ProtocolVersion = "2024-11-05"

	// JSON-RPC methods
	MethodInitialize      = "initialize"
	MethodInitialized     = "notifications/initialized"
	MethodShutdown        = "shutdown"
	MethodListTools       = "tools/list"
	MethodCallTool        = "tools/call"
	MethodListPrompts     = "prompts/list"
	MethodGetPrompt       = "prompts/get"
	MethodListResources   = "resources/list"
	MethodReadResource    = "resources/read"
	MethodListResourceTemplates = "resources/templates/list"
)

// Request represents a JSON-RPC request
type Request struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Method  string        `json:"method"`
	Params  interface{}   `json:"params,omitempty"`
}

// Response represents a JSON-RPC response
type Response struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *Error        `json:"error,omitempty"`
}

// Error represents a JSON-RPC error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Notification represents a JSON-RPC notification
type Notification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// InitializeParams represents parameters for the initialize request
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ClientCapabilities     `json:"capabilities"`
	ClientInfo      ImplementationInfo     `json:"clientInfo"`
}

// ClientCapabilities represents client capabilities
type ClientCapabilities struct {
	Roots    *RootsCapability    `json:"roots,omitempty"`
	Sampling *SamplingCapability `json:"sampling,omitempty"`
}

// RootsCapability represents the roots capability
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability represents the sampling capability
type SamplingCapability struct{}

// ImplementationInfo represents implementation information
type ImplementationInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult represents the result of the initialize request
type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ServerCapabilities     `json:"capabilities"`
	ServerInfo      ImplementationInfo     `json:"serverInfo"`
}

// ServerCapabilities represents server capabilities
type ServerCapabilities struct {
	Tools         *ToolsCapability         `json:"tools,omitempty"`
	Prompts       *PromptsCapability       `json:"prompts,omitempty"`
	Resources     *ResourcesCapability     `json:"resources,omitempty"`
}

// ToolsCapability represents the tools capability
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability represents the prompts capability
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability represents the resources capability
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// ListToolsResult represents the result of listing tools
type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// CallToolParams represents parameters for calling a tool
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// CallToolResult represents the result of calling a tool
type CallToolResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content represents content in a tool result
type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Prompt represents an MCP prompt
type Prompt struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListPromptsResult represents the result of listing prompts
type ListPromptsResult struct {
	Prompts []Prompt `json:"prompts"`
}

// GetPromptParams represents parameters for getting a prompt
type GetPromptParams struct {
	Name string `json:"name"`
}

// GetPromptResult represents the result of getting a prompt
type GetPromptResult struct {
	Prompt Prompt `json:"prompt"`
}

// Resource represents an MCP resource
type Resource struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListResourcesResult represents the result of listing resources
type ListResourcesResult struct {
	Resources []Resource `json:"resources"`
}

// ReadResourceParams represents parameters for reading a resource
type ReadResourceParams struct {
	Name string `json:"name"`
}

// ResourceContent represents the content of a resource
type ResourceContent struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// ReadResourceResult represents the result of reading a resource
type ReadResourceResult struct {
	Content ResourceContent `json:"content"`
}

// ResourceTemplate represents a resource template
type ResourceTemplate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListResourceTemplatesResult represents the result of listing resource templates
type ListResourceTemplatesResult struct {
	Templates []ResourceTemplate `json:"templates"`
}