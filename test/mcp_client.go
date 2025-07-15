package test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
func NewMCPClient(serverPath string, args ...string) (client *MCPClient, err error) {
	var cmd *exec.Cmd
	var stdin io.WriteCloser
	var stdout io.ReadCloser
	var stderr io.ReadCloser

	cmd = exec.Command(serverPath, args...)

	stdin, err = cmd.StdinPipe()
	if err != nil {
		err = fmt.Errorf("failed to create stdin pipe: %w", err)
		goto end
	}

	stdout, err = cmd.StdoutPipe()
	if err != nil {
		err = fmt.Errorf("failed to create stdout pipe: %w", err)
		goto end
	}

	stderr, err = cmd.StderrPipe()
	if err != nil {
		err = fmt.Errorf("failed to create stderr pipe: %w", err)
		goto end
	}

	err = cmd.Start()
	if err != nil {
		err = fmt.Errorf("failed to start server: %w", err)
		goto end
	}

	client = &MCPClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		nextID: 1,
	}

end:
	return client, err
}

// Initialize sends the MCP initialization handshake
func (c *MCPClient) Initialize(ctx context.Context) (err error) {
	var req MCPRequest
	var resp *MCPResponse
	var notif MCPRequest

	req = MCPRequest{
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

	resp, err = c.sendRequest(ctx, req)
	if err != nil {
		err = fmt.Errorf("initialization failed: %w", err)
		goto end
	}

	if resp.Error != nil {
		err = fmt.Errorf("initialization error: %s", resp.Error.Message)
		goto end
	}

	// Send initialized notification
	notif = MCPRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}

	err = c.sendNotification(notif)

end:
	return err
}

// CallTool calls an MCP tool with the given name and arguments
func (c *MCPClient) CallTool(ctx context.Context, toolName string, arguments map[string]interface{}) (result *MCPResponse, err error) {
	var req MCPRequest

	req = MCPRequest{
		JSONRPC: "2.0",
		ID:      c.getNextID(),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}

	result, err = c.sendRequest(ctx, req)

	return result, err
}

// ListTools requests the list of available tools
func (c *MCPClient) ListTools(ctx context.Context) (result *MCPResponse, err error) {
	var req MCPRequest

	req = MCPRequest{
		JSONRPC: "2.0",
		ID:      c.getNextID(),
		Method:  "tools/list",
	}

	result, err = c.sendRequest(ctx, req)

	return result, err
}

// Close shuts down the MCP client and server
func (c *MCPClient) Close() (err error) {
	if c.stdin != nil {
		_ = c.stdin.Close()
	}
	if c.stdout != nil {
		_ = c.stdout.Close()
	}
	if c.stderr != nil {
		_ = c.stderr.Close()
	}

	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
		_ = c.cmd.Wait()
	}

	return err
}

// getNextID returns the next request ID
func (c *MCPClient) getNextID() (id int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	id = c.nextID
	c.nextID++

	return id
}

// sendRequest sends a request and waits for a response
func (c *MCPClient) sendRequest(ctx context.Context, req MCPRequest) (response *MCPResponse, err error) {
	var data []byte
	var errChan chan error
	var resp MCPResponse
	var scanner *bufio.Scanner

	data, err = json.Marshal(req)
	if err != nil {
		err = fmt.Errorf("failed to marshal request: %w", err)
		goto end
	}

	// Send request
	_, err = c.stdin.Write(append(data, '\n'))
	if err != nil {
		err = fmt.Errorf("failed to write request: %w", err)
		goto end
	}

	// Read response
	scanner = bufio.NewScanner(c.stdout)

	// Set timeout
	errChan = make(chan error, 1)

	go func() {
		if scanner.Scan() {
			err := json.Unmarshal(scanner.Bytes(), &resp)
			errChan <- err
		} else {
			errChan <- fmt.Errorf("failed to read response")
		}
	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
		goto end
	case <-time.After(5 * time.Second):
		err = fmt.Errorf("request timeout")
		goto end
	case err = <-errChan:
		if err != nil {
			goto end
		}
		response = &resp
	}

end:
	return response, err
}

// sendNotification sends a notification (no response expected)
func (c *MCPClient) sendNotification(req MCPRequest) (err error) {
	var data []byte

	data, err = json.Marshal(req)
	if err != nil {
		err = fmt.Errorf("failed to marshal notification: %w", err)
		goto end
	}

	_, err = c.stdin.Write(append(data, '\n'))

end:
	return err
}
