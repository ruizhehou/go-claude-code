package tools

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/houruizhe/go-claude-code/pkg/api"
)

// Tool represents a tool that can be executed
type Tool interface {
	// Name returns the name of the tool
	Name() string

	// Description returns a description of what the tool does
	Description() string

	// Parameters returns the JSON schema for the tool's parameters
	Parameters() map[string]interface{}

	// Execute executes the tool with the given arguments
	Execute(ctx *ExecutionContext, args map[string]interface{}) (interface{}, error)
}

// ExecutionContext provides context for tool execution
type ExecutionContext struct {
	WorkingDir string
	Env         map[string]string
	User        string
	// Add more context as needed
}

// Registry manages available tools
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register registers a tool in the registry
func (r *Registry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool '%s' already registered", name)
	}

	r.tools[name] = tool
	return nil
}

// Unregister removes a tool from the registry
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tools, name)
}

// Get returns a tool by name
func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, exists := r.tools[name]
	return tool, exists
}

// List returns all registered tools
func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Names returns all registered tool names
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered tools
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

// ToAPIDefinitions converts registered tools to API definitions
func (r *Registry) ToAPIDefinitions() []api.ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	definitions := make([]api.ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		def := api.ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: api.InputSchema{
				Type:       "object",
				Properties: make(map[string]api.PropertySchema),
			},
		}

		// Convert parameters to PropertySchema
		for key, param := range tool.Parameters() {
			if paramMap, ok := param.(map[string]interface{}); ok {
				prop := api.PropertySchema{
					Type:        "string",
					Description: "",
				}
				if t, ok := paramMap["type"].(string); ok {
					prop.Type = t
				}
				if desc, ok := paramMap["description"].(string); ok {
					prop.Description = desc
				}
				def.InputSchema.Properties[key] = prop
			}
		}

		definitions = append(definitions, def)
	}
	return definitions
}

// ToJSON returns a JSON representation of the registry
func (r *Registry) ToJSON() (string, error) {
	tools := r.List()
	data, err := json.MarshalIndent(tools, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
