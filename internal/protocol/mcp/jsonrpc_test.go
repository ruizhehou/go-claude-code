package mcp

import (
	"bytes"
	"testing"
)

func TestNewJSONRPC(t *testing.T) {
	buf := &bytes.Buffer{}
	j := NewJSONRPC(buf)
	if j == nil {
		t.Fatal("NewJSONRPC should return non-nil instance")
	}
}

func TestSendNotification(t *testing.T) {
	buf := &bytes.Buffer{}
	j := NewJSONRPC(buf)
	err := j.SendNotification("test", nil)
	if err != nil {
		t.Errorf("SendNotification should not return error: %v", err)
	}
}

func TestClose(t *testing.T) {
	buf := &bytes.Buffer{}
	j := NewJSONRPC(buf)
	err := j.Close()
	if err != nil {
		t.Errorf("Close should not return error: %v", err)
	}
}

