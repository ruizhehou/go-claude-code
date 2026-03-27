package config

import (
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	Anthropic AnthropicConfig `json:"anthropic"`
	CLI       CLIConfig       `json:"cli"`
	MCP       MCPConfig       `json:"mcp"`
	Tools     ToolsConfig     `json:"tools"`
}

// AnthropicConfig represents Anthropic API configuration
type AnthropicConfig struct {
	APIKey      string  `json:"apiKey"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"maxTokens"`
	Temperature float64 `json:"temperature"`
}

// CLIConfig represents CLI configuration
type CLIConfig struct {
	Prompt      string `json:"prompt"`
	HistoryFile string `json:"historyFile"`
	LogLevel    string `json:"logLevel"`
}

// MCPConfig represents MCP server configuration
type MCPConfig struct {
	Servers map[string]MCPServerConfig `json:"servers"`
}

// MCPServerConfig represents a single MCP server configuration
type MCPServerConfig struct {
	Command string            `json:"command"` // For stdio transport
	Args    []string          `json:"args"`    // For stdio transport
	URL     string            `json:"url"`     // For HTTP transport
	Env     map[string]string `json:"env"`     // Environment variables
	Enabled bool              `json:"enabled"`
}

// ToolsConfig represents tools configuration
type ToolsConfig struct {
	AllowedPaths []string  `json:"allowedPaths"`
	Shell        ShellConfig `json:"shell"`
}

// ShellConfig represents shell tool configuration
type ShellConfig struct {
	Enabled bool   `json:"enabled"`
	Timeout string `json:"timeout"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Anthropic: AnthropicConfig{
			Model:       "claude-sonnet-4-20250514",
			MaxTokens:   4096,
			Temperature: 0.0,
		},
		CLI: CLIConfig{
			Prompt:      "claude> ",
			HistoryFile: filepath.Join(homeDir, ".claude-code", "history.json"),
			LogLevel:    "info",
		},
		MCP: MCPConfig{
			Servers: make(map[string]MCPServerConfig),
		},
		Tools: ToolsConfig{
			AllowedPaths: []string{homeDir},
			Shell: ShellConfig{
				Enabled: true,
				Timeout: "2m",
			},
		},
	}
}

// GetConfigPath returns the default config file path
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".claude-code")
	return filepath.Join(configDir, "config.json"), nil
}

// GetMCPServersPath returns the MCP servers config file path
func GetMCPServersPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".claude-code")
	return filepath.Join(configDir, "servers.json"), nil
}
