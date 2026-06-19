package mcp

import (
	"testing"

	"github.com/houruizhe/go-claude-code/pkg/tools"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-mcp", nil)
	if client == nil {
		t.Fatal("NewClient should return a non-nil client")
	}
	if client.Name() != "test-mcp" {
		t.Errorf("Expected name test-mcp, got %s", client.Name())
	}
}

func TestToolWrapper(t *testing.T) {
	wrapper := &ToolWrapper{
		name: "test-tool",
		description: "A test tool",
	}
	if wrapper.Name() != "test-tool" {
		t.Errorf("Expected name test-tool, got %s", wrapper.Name())
	}
}

func TestGetTools(t *testing.T) {
	client := NewClient("test", nil)
	tools := client.GetTools()
	if tools == nil {
		t.Error("GetTools should return non-nil slice")
	}
}

func TestRegisterTools(t *testing.T) {
	client := NewClient("test", nil)
	registry := tools.NewRegistry()
	err := client.RegisterTools(registry)
	if err != nil {
		t.Errorf("RegisterTools should not return error: %v", err)
	}
}

