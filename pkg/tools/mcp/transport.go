package mcp

import (
	"context"
	"io"
)

// Transport represents an MCP transport
type Transport interface {
	// Connect connects to the MCP server
	Connect(ctx context.Context) error

	// Close closes the connection
	Close() error

	// Read reads data from the transport
	Read(p []byte) (n int, err error)

	// Write writes data to the transport
	Write(p []byte) (n int, err error)
}

// ioReadWriteCloser wraps an io.ReadWriter with a Close method
type ioReadWriteCloser struct {
	io.ReadWriter
	closeFunc func() error
}

// Close closes the underlying connection
func (rwc *ioReadWriteCloser) Close() error {
	if rwc.closeFunc != nil {
		return rwc.closeFunc()
	}
	return nil
}
