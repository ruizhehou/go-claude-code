package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/houruizhe/go-claude-code/pkg/api"
	"github.com/houruizhe/go-claude-code/pkg/cli"
	"github.com/houruizhe/go-claude-code/pkg/config"
	"github.com/houruizhe/go-claude-code/pkg/tools"
	"github.com/houruizhe/go-claude-code/pkg/tools/builtin"
)

// APIClient is an interface for API clients
type APIClient interface {
	StreamChat(ctx context.Context, req *api.ChatRequest) (<-chan *api.ChatEvent, error)
}

func main() {
	// Parse command line arguments
	configPath := ""
	if len(os.Args) > 1 {
		for i := 1; i < len(os.Args); i++ {
			if os.Args[i] == "--config" && i+1 < len(os.Args) {
				configPath = os.Args[i+1]
				break
			}
		}
	}

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := config.Validate(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration validation failed: %v\n", err)
		os.Exit(1)
	}

	// Create API client based on provider
	var apiClient APIClient
	switch cfg.Provider {
	case "anthropic":
		apiClient = api.NewClient(
			api.WithAPIKey(cfg.Anthropic.APIKey),
			api.WithModel(cfg.Anthropic.Model),
			api.WithMaxTokens(cfg.Anthropic.MaxTokens),
		)
		fmt.Printf("Using Anthropic API (model: %s)\n", cfg.Anthropic.Model)
	case "volcengine":
		apiClient = api.NewVolcengineClient(
			api.WithVolcengineAPIKey(cfg.Volcengine.APIKey),
			api.WithVolcengineModel(cfg.Volcengine.Model),
			api.WithVolcengineMaxTokens(cfg.Volcengine.MaxTokens),
			api.WithVolcengineBaseURL(cfg.Volcengine.BaseURL),
		)
		fmt.Printf("Using Volcengine API (model: %s)\n", cfg.Volcengine.Model)
	default:
		fmt.Fprintf(os.Stderr, "Unknown provider: %s\n", cfg.Provider)
		os.Exit(1)
	}

	// Create tool registry
	registry := tools.NewRegistry()

	// Register built-in tools
	shellTimeout := parseDuration(cfg.Tools.Shell.Timeout)
	registry.Register(builtin.NewBashTool(shellTimeout))
	registry.Register(builtin.NewFileReadTool(cfg.Tools.AllowedPaths))
	registry.Register(builtin.NewFileWriteTool(cfg.Tools.AllowedPaths))
	registry.Register(builtin.NewFileEditTool(cfg.Tools.AllowedPaths))

	// Create tool executor
	executor := tools.NewExecutor(registry, shellTimeout)

	// Create REPL
	repl := cli.NewREPL(cfg, apiClient, registry, executor)

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Run REPL
	if err := repl.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// parseDuration parses a duration string (e.g., "2m", "30s")
func parseDuration(s string) time.Duration {
	if s == "" {
		return 2 * time.Minute
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return 2 * time.Minute
	}

	return d
}
