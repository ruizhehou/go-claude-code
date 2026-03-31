package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Load loads configuration from file, environment variables, and command-line flags
func Load(configPath string) (*Config, error) {
	// Start with default config
	cfg := DefaultConfig()

	// Load from file if specified
	if configPath != "" {
		if err := loadFromFile(cfg, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
	} else {
		// Try default config path
		defaultPath, err := GetConfigPath()
		if err == nil {
			if _, err := os.Stat(defaultPath); err == nil {
				if err := loadFromFile(cfg, defaultPath); err != nil {
					return nil, fmt.Errorf("failed to load config from default path: %w", err)
				}
			}
		}
	}

	// Load MCP servers from separate file
	mcpPath, err := GetMCPServersPath()
	if err == nil {
		if _, err := os.Stat(mcpPath); err == nil {
			if err := loadMCPServers(cfg, mcpPath); err != nil {
				return nil, fmt.Errorf("failed to load MCP servers: %w", err)
			}
		}
	}

	// Override with environment variables
	loadFromEnv(cfg)

	return cfg, nil
}

// loadFromFile loads configuration from a JSON file
func loadFromFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, cfg)
}

// loadMCPServers loads MCP server configuration
func loadMCPServers(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var mcpConfig struct {
		Servers map[string]MCPServerConfig `json:"mcpServers"`
	}

	if err := json.Unmarshal(data, &mcpConfig); err != nil {
		return err
	}

	cfg.MCP.Servers = mcpConfig.Servers
	return nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(cfg *Config) {
	// Provider
	if provider := os.Getenv("CLAUDE_PROVIDER"); provider != "" {
		cfg.Provider = provider
	}

	// Anthropic API key
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		cfg.Anthropic.APIKey = apiKey
	}

	// Anthropic Model
	if model := os.Getenv("CLAUDE_MODEL"); model != "" {
		cfg.Anthropic.Model = model
	}

	// Anthropic Max tokens
	if maxTokens := os.Getenv("CLAUDE_MAX_TOKENS"); maxTokens != "" {
		var mt int
		if _, err := fmt.Sscanf(maxTokens, "%d", &mt); err == nil {
			cfg.Anthropic.MaxTokens = mt
		}
	}

	// Anthropic Temperature
	if temp := os.Getenv("CLAUDE_TEMPERATURE"); temp != "" {
		var t float64
		if _, err := fmt.Sscanf(temp, "%f", &t); err == nil {
			cfg.Anthropic.Temperature = t
		}
	}

	// Volcengine API key
	if apiKey := os.Getenv("VOLCENGINE_API_KEY"); apiKey != "" {
		cfg.Volcengine.APIKey = apiKey
	}

	// Volcengine Model
	if model := os.Getenv("VOLCENGINE_MODEL"); model != "" {
		cfg.Volcengine.Model = model
	}

	// Volcengine Max tokens
	if maxTokens := os.Getenv("VOLCENGINE_MAX_TOKENS"); maxTokens != "" {
		var mt int
		if _, err := fmt.Sscanf(maxTokens, "%d", &mt); err == nil {
			cfg.Volcengine.MaxTokens = mt
		}
	}

	// Volcengine Temperature
	if temp := os.Getenv("VOLCENGINE_TEMPERATURE"); temp != "" {
		var t float64
		if _, err := fmt.Sscanf(temp, "%f", &t); err == nil {
			cfg.Volcengine.Temperature = t
		}
	}

	// Volcengine Base URL
	if baseURL := os.Getenv("VOLCENGINE_BASE_URL"); baseURL != "" {
		cfg.Volcengine.BaseURL = baseURL
	}

	// Log level
	if logLevel := os.Getenv("CLAUDE_LOG_LEVEL"); logLevel != "" {
		cfg.CLI.LogLevel = logLevel
	}

	// Allowed paths
	if allowedPaths := os.Getenv("CLAUDE_ALLOWED_PATHS"); allowedPaths != "" {
		cfg.Tools.AllowedPaths = strings.Split(allowedPaths, ":")
	}

	// Shell enabled
	if shellEnabled := os.Getenv("CLAUDE_SHELL_ENABLED"); shellEnabled != "" {
		cfg.Tools.Shell.Enabled = strings.ToLower(shellEnabled) == "true"
	}
}

// Save saves configuration to a file
func Save(cfg *Config, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Validate validates the configuration
func Validate(cfg *Config) error {
	// Check provider
	switch cfg.Provider {
	case "anthropic":
		if cfg.Anthropic.APIKey == "" {
			return fmt.Errorf("anthropic API key is required (set ANTHROPIC_API_KEY environment variable)")
		}
		if cfg.Anthropic.Model == "" {
			return fmt.Errorf("anthropic model is required")
		}
		if cfg.Anthropic.MaxTokens <= 0 {
			return fmt.Errorf("max tokens must be positive")
		}
		if cfg.Anthropic.Temperature < 0 || cfg.Anthropic.Temperature > 1 {
			return fmt.Errorf("temperature must be between 0 and 1")
		}
	case "volcengine":
		if cfg.Volcengine.APIKey == "" {
			return fmt.Errorf("volcengine API key is required (set VOLCENGINE_API_KEY environment variable)")
		}
		if cfg.Volcengine.Model == "" {
			return fmt.Errorf("volcengine model is required")
		}
		if cfg.Volcengine.MaxTokens <= 0 {
			return fmt.Errorf("max tokens must be positive")
		}
		if cfg.Volcengine.Temperature < 0 || cfg.Volcengine.Temperature > 1 {
			return fmt.Errorf("temperature must be between 0 and 1")
		}
	default:
		return fmt.Errorf("invalid provider '%s', must be 'anthropic' or 'volcengine'", cfg.Provider)
	}

	// Check MCP servers
	for name, server := range cfg.MCP.Servers {
		if server.Command == "" && server.URL == "" {
			return fmt.Errorf("MCP server '%s' must have either command or URL", name)
		}
	}

	return nil
}
