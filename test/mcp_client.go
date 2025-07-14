package test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// MCPClient is a simple MCP client for testing
type MCPClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	mu     sync.Mutex
	nextID int64
}

// MCPRequest represents an MCP JSON-RPC request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents an MCP JSON-RPC response
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewMCPClient creates a new MCP client connected to the scout-mcp server
func NewMCPClient(serverPath string, args ...string) (*MCPClient, error) {
	cmd := exec.Command(serverPath, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	client := &MCPClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		nextID: 1,
	}

	return client, nil
}

// Initialize sends the MCP initialization handshake
func (c *MCPClient) Initialize(ctx context.Context) error {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      c.getNextID(),
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "scout-mcp-test",
				"version": "1.0.0",
			},
		},
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("initialization error: %s", resp.Error.Message)
	}

	// Send initialized notification
	notif := MCPRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}

	return c.sendNotification(notif)
}

// CallTool calls an MCP tool with the given name and arguments
func (c *MCPClient) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (*MCPResponse, error) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      c.getNextID(),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}

	return c.sendRequest(ctx, req)
}

// ListTools requests the list of available tools
func (c *MCPClient) ListTools(ctx context.Context) (*MCPResponse, error) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      c.getNextID(),
		Method:  "tools/list",
	}

	return c.sendRequest(ctx, req)
}

// Close shuts down the MCP client and server
func (c *MCPClient) Close() error {
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.stdout != nil {
		c.stdout.Close()
	}
	if c.stderr != nil {
		c.stderr.Close()
	}

	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
		c.cmd.Wait()
	}

	return nil
}

// getNextID returns the next request ID
func (c *MCPClient) getNextID() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.nextID
	c.nextID++
	return id
}

// sendRequest sends a request and waits for a response
func (c *MCPClient) sendRequest(ctx context.Context, req MCPRequest) (*MCPResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request
	if _, err := c.stdin.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Read response
	scanner := bufio.NewScanner(c.stdout)

	// Set timeout
	done := make(chan bool, 1)
	var resp MCPResponse
	var scanErr error

	go func() {
		if scanner.Scan() {
			scanErr = json.Unmarshal(scanner.Bytes(), &resp)
		} else {
			scanErr = fmt.Errorf("failed to read response")
		}
		done <- true
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("request timeout")
	case <-done:
		if scanErr != nil {
			return nil, scanErr
		}
		return &resp, nil
	}
}

// sendNotification sends a notification (no response expected)
func (c *MCPClient) sendNotification(req MCPRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	_, err = c.stdin.Write(append(data, '\n'))
	return err
}
