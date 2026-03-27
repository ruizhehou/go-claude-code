package mcp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// HTTPTransport implements MCP transport via HTTP
type HTTPTransport struct {
	url    string
	client *http.Client

	mu     sync.Mutex
	closed bool
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(url string) *HTTPTransport {
	return &HTTPTransport{
		url: url,
		client: &http.Client{
			Transport: &http.Transport{
				DisableCompression: false,
			},
		},
	}
}

// Connect connects to the MCP server via HTTP
func (t *HTTPTransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	// HTTP is connectionless, so just verify the URL is accessible
	// We could send a ping here, but for now we'll just return
	return nil
}

// Read reads from the HTTP response (not used for HTTP transport)
func (t *HTTPTransport) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("HTTP transport does not support Read")
}

// Write writes to the HTTP request (not used for HTTP transport)
func (t *HTTPTransport) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("HTTP transport does not support Write")
}

// SendRequest sends an HTTP request and returns the response
func (t *HTTPTransport) SendRequest(ctx context.Context, data []byte) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil, fmt.Errorf("transport is closed")
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", t.url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	// Read response
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return buf.Bytes(), nil
}

// Close closes the transport
func (t *HTTPTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.closed = true
	return nil
}

// HTTPReadWriteCloser implements io.ReadWriteCloser for HTTP transport
type HTTPReadWriteCloser struct {
	transport *HTTPTransport
	reader    *bufio.Reader
	data      []byte
	mu        sync.Mutex
}

// NewHTTPReadWriteCloser creates a new HTTP read-write closer
func NewHTTPReadWriteCloser(transport *HTTPTransport) *HTTPReadWriteCloser {
	return &HTTPReadWriteCloser{
		transport: transport,
	}
}

// Read reads from the HTTP response buffer
func (rwc *HTTPReadWriteCloser) Read(p []byte) (n int, err error) {
	rwc.mu.Lock()
	defer rwc.mu.Unlock()

	if len(rwc.data) == 0 {
		return 0, io.EOF
	}

	n = copy(p, rwc.data)
	rwc.data = rwc.data[n:]
	return n, nil
}

// Write sends data via HTTP and stores the response for reading
func (rwc *HTTPReadWriteCloser) Write(p []byte) (n int, err error) {
	resp, err := rwc.transport.SendRequest(context.Background(), p)
	if err != nil {
		return 0, err
	}

	rwc.mu.Lock()
	rwc.data = resp
	rwc.reader = bufio.NewReader(bytes.NewReader(resp))
	rwc.mu.Unlock()

	return len(p), nil
}

// Close closes the HTTP read-write closer
func (rwc *HTTPReadWriteCloser) Close() error {
	return rwc.transport.Close()
}
