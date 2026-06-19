package tools

import (
	"testing"
)

type mockTool struct {
	name        string
	description string
}

func (m *mockTool) Name() string {
	return m.name
}

func (m *mockTool) Description() string {
	return m.description
}

func (m *mockTool) Parameters() map[string]interface{} {
	return nil
}

func (m *mockTool) Execute(ctx *ExecutionContext, args map[string]interface{}) (interface{}, error) {
	return nil, nil
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry should return non-nil registry")
	}
}

func TestRegister(t *testing.T) {
	registry := NewRegistry()
	tool := &mockTool{name: "test-tool"}
	err := registry.Register(tool)
	if err != nil {
		t.Errorf("Register should not return error: %v", err)
	}
}

func TestRegisterEmptyName(t *testing.T) {
	registry := NewRegistry()
	tool := &mockTool{name: ""}
	err := registry.Register(tool)
	if err == nil {
		t.Error("Register should return error for empty tool name")
	}
}

func TestGet(t *testing.T) {
	registry := NewRegistry()
	tool := &mockTool{name: "test-tool"}
	registry.Register(tool)
	foundTool, ok := registry.Get("test-tool")
	if !ok {
		t.Error("Get should find registered tool")
	}
	if foundTool.Name() != "test-tool" {
		t.Errorf("Expected tool name test-tool, got %s", foundTool.Name())
	}
}

func TestGetNotFound(t *testing.T) {
	registry := NewRegistry()
	_, ok := registry.Get("nonexistent")
	if ok {
		t.Error("Get should return false for nonexistent tool")
	}
}

