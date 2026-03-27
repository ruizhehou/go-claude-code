package mcp

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// StdioTransport implements MCP transport via stdio
type StdioTransport struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.Reader
	stderr io.Reader

	mu     sync.Mutex
	closed bool
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(command string, args []string, env map[string]string) *StdioTransport {
	cmd := exec.Command(command, args...)

	// Set environment variables
	if env != nil {
		cmd.Env = append(cmd.Env, envToSlice(env)...)
	}

	return &StdioTransport{
		cmd: cmd,
	}
}

// Connect connects to the MCP server via stdio
func (t *StdioTransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	stdin, err := t.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := t.cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := t.cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	t.stdin = stdin
	t.stdout = stdout
	t.stderr = stderr

	// Start the process
	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	return nil
}

// Read reads from stdout
func (t *StdioTransport) Read(p []byte) (n int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return 0, fmt.Errorf("transport is closed")
	}

	return t.stdout.Read(p)
}

// Write writes to stdin
func (t *StdioTransport) Write(p []byte) (n int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return 0, fmt.Errorf("transport is closed")
	}

	return t.stdin.Write(p)
}

// Close closes the transport
func (t *StdioTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true

	// Close pipes
	if t.stdin != nil {
		t.stdin.Close()
	}
	if t.stdout != nil {
		// Can't close stdout as it's from a pipe
	}
	if t.stderr != nil {
		// Can't close stderr as it's from a pipe
	}

	// Kill the process if it's still running
	if t.cmd.Process != nil {
		t.cmd.Process.Kill()
	}

	// Wait for the process to exit
	t.cmd.Wait()

	return nil
}

// envToSlice converts a map to environment variable slice
func envToSlice(env map[string]string) []string {
	slice := make([]string, 0, len(env))
	for k, v := range env {
		slice = append(slice, fmt.Sprintf("%s=%s", k, v))
	}
	return slice
}
