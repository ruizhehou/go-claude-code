package builtin

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/houruizhe/go-claude-code/pkg/tools"
)

// BashTool executes shell commands
type BashTool struct {
	timeout time.Duration
}

// NewBashTool creates a new Bash tool
func NewBashTool(timeout time.Duration) *BashTool {
	return &BashTool{timeout: timeout}
}

// Name returns the name of the tool
func (t *BashTool) Name() string {
	return "bash"
}

// Description returns the description of the tool
func (t *BashTool) Description() string {
	return "Execute a bash command in the terminal. Use this for running git commands, npm, docker, and other terminal operations. DO NOT use this for file operations - use the file tools instead."
}

// Parameters returns the JSON schema for the tool's parameters
func (t *BashTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"command": map[string]interface{}{
			"type":        "string",
			"description": "The bash command to execute",
		},
		"timeout": map[string]interface{}{
			"type":        "integer",
			"description": "Optional timeout in milliseconds (max 600000)",
		},
	}
}

// Execute executes the tool
func (t *BashTool) Execute(ctx *tools.ExecutionContext, args map[string]interface{}) (interface{}, error) {
	// Parse command
	command, ok := args["command"].(string)
	if !ok || command == "" {
		return nil, fmt.Errorf("command is required")
	}

	// Parse timeout
	timeout := t.timeout
	if timeoutMs, ok := args["timeout"].(float64); ok {
		timeout = time.Duration(timeoutMs) * time.Millisecond
		if timeout > 10*time.Minute {
			timeout = 10 * time.Minute
		}
	}

	// Create command context
	cmdCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Determine shell
	shell := "/bin/bash"
	if _, err := os.Stat(shell); os.IsNotExist(err) {
		shell = "/bin/sh"
	}

	// Create command
	cmd := exec.CommandContext(cmdCtx, shell, "-c", command)

	// Set working directory
	if ctx.WorkingDir != "" {
		cmd.Dir = ctx.WorkingDir
	}

	// Set environment variables
	if ctx.Env != nil {
		env := os.Environ()
		for k, v := range ctx.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = env
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	// Prepare result
	result := map[string]interface{}{
		"exitCode": 0,
		"stdout":   stdout.String(),
		"stderr":   stderr.String(),
		"duration": duration.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result["exitCode"] = exitErr.ExitCode()
		} else {
			result["exitCode"] = -1
		}
		result["error"] = err.Error()
	}

	// Truncate output if too long
	maxOutputLength := 30000
	if len(result["stdout"].(string)) > maxOutputLength {
		result["stdout"] = result["stdout"].(string)[:maxOutputLength] + "\n... (truncated)"
	}
	if len(result["stderr"].(string)) > maxOutputLength {
		result["stderr"] = result["stderr"].(string)[:maxOutputLength] + "\n... (truncated)"
	}

	return result, nil
}

// IsDangerousCommand checks if a command is potentially dangerous
func (t *BashTool) IsDangerousCommand(command string) bool {
	lowerCmd := strings.ToLower(command)

	// List of dangerous commands
	dangerous := []string{
		"rm -rf /",
		"rm -rf /*",
		"mkfs",
		"dd if=/dev/zero",
		":(){ :|:& };:",
		"chmod -R 777 /",
		"chown -R root:root /",
		"shutdown",
		"reboot",
		"poweroff",
		"halt",
	}

	for _, d := range dangerous {
		if strings.Contains(lowerCmd, d) {
			return true
		}
	}

	return false
}
