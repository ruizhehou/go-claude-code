package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// JSONRPC represents a JSON-RPC 2.0 client/server
type JSONRPC struct {
	enc *json.Encoder
	dec *json.Decoder

	mu       sync.Mutex
	pending  map[interface{}]chan *Response
	nextID   int
}

// NewJSONRPC creates a new JSON-RPC instance
func NewJSONRPC(rw io.ReadWriter) *JSONRPC {
	return &JSONRPC{
		enc:      json.NewEncoder(rw),
		dec:      json.NewDecoder(rw),
		pending:  make(map[interface{}]chan *Response),
		nextID:   1,
	}
}

// SendRequest sends a JSON-RPC request and waits for the response
func (j *JSONRPC) SendRequest(method string, params interface{}) (*Response, error) {
	j.mu.Lock()

	id := j.nextID
	j.nextID++

	respChan := make(chan *Response, 1)
	j.pending[id] = respChan

	req := &Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	if err := j.enc.Encode(req); err != nil {
		delete(j.pending, id)
		j.mu.Unlock()
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	j.mu.Unlock()

	// Wait for response
	resp := <-respChan

	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	return resp, nil
}

// SendNotification sends a JSON-RPC notification (no response expected)
func (j *JSONRPC) SendNotification(method string, params interface{}) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	notif := &Notification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	if err := j.enc.Encode(notif); err != nil {
		return fmt.Errorf("failed to encode notification: %w", err)
	}

	return nil
}

// ReadMessage reads a JSON-RPC message from the stream
func (j *JSONRPC) ReadMessage() (*Response, error) {
	var raw map[string]json.RawMessage
	if err := j.dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode message: %w", err)
	}

	// Check if it's a response (has id)
	if idBytes, ok := raw["id"]; ok {
		var id interface{}
		if err := json.Unmarshal(idBytes, &id); err != nil {
			return nil, fmt.Errorf("failed to unmarshal id: %w", err)
		}

		j.mu.Lock()
		respChan, ok := j.pending[id]
		if ok {
			delete(j.pending, id)
		}
		j.mu.Unlock()

		if !ok {
			return nil, fmt.Errorf("unexpected response id: %v", id)
		}

		// Parse response
		resp := &Response{ID: id}
		if errBytes, ok := raw["error"]; ok {
			var err Error
			if err := json.Unmarshal(errBytes, &err); err != nil {
				return nil, fmt.Errorf("failed to unmarshal error: %w", err)
			}
			resp.Error = &err
		} else if resultBytes, ok := raw["result"]; ok {
			if err := json.Unmarshal(resultBytes, &resp.Result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal result: %w", err)
			}
		}

		respChan <- resp
		return resp, nil
	}

	// Handle notifications (no id)
	return nil, nil
}

// ReceiveNotification receives a JSON-RPC notification from the stream
func (j *JSONRPC) ReceiveNotification() (*Notification, error) {
	var raw map[string]json.RawMessage
	if err := j.dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode notification: %w", err)
	}

	// Check if it's a notification (no id)
	if _, ok := raw["id"]; ok {
		return nil, fmt.Errorf("expected notification but got response")
	}

	// Parse notification
	var notif Notification
	if methodBytes, ok := raw["method"]; ok {
		if err := json.Unmarshal(methodBytes, &notif.Method); err != nil {
			return nil, fmt.Errorf("failed to unmarshal method: %w", err)
		}
	}

	if paramsBytes, ok := raw["params"]; ok && len(paramsBytes) > 0 {
		if err := json.Unmarshal(paramsBytes, &notif.Params); err != nil {
			return nil, fmt.Errorf("failed to unmarshal params: %w", err)
		}
	}

	notif.JSONRPC = "2.0"
	return &notif, nil
}

// Close closes the JSON-RPC connection
func (j *JSONRPC) Close() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	// Close all pending channels
	for _, ch := range j.pending {
		close(ch)
	}
	j.pending = make(map[interface{}]chan *Response)

	return nil
}