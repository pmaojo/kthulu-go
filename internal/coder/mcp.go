package coder

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// MCPClient connects to MCP servers via stdio
type MCPClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	reader *bufio.Reader
	mu     sync.Mutex
	nextID int
}

// MCPRequest is a JSON-RPC 2.0 request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse is a JSON-RPC 2.0 response
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError represents an error in the response
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCPTool represents a tool exposed by an MCP server
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// MCPToolsResult is the response from tools/list
type MCPToolsResult struct {
	Tools []MCPTool `json:"tools"`
}

// MCPCallResult is the response from tools/call
type MCPCallResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// MCPContent is a content block in tool results
type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// NewMCPClient creates a new MCP client connected to a server
func NewMCPClient(ctx context.Context, command string, args ...string) (*MCPClient, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP server: %w", err)
	}

	client := &MCPClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		reader: bufio.NewReader(stdout),
		nextID: 1,
	}

	// Initialize the connection
	if err := client.initialize(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}

// initialize sends the initialize request to the MCP server
func (c *MCPClient) initialize() error {
	_, err := c.call("initialize", map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "kthulu-coder",
			"version": "1.0.0",
		},
	})
	if err != nil {
		return fmt.Errorf("initialize failed: %w", err)
	}

	// Send initialized notification
	return c.notify("notifications/initialized", nil)
}

// ListTools returns all tools available from the MCP server
func (c *MCPClient) ListTools() ([]MCPTool, error) {
	result, err := c.call("tools/list", nil)
	if err != nil {
		return nil, err
	}

	var toolsResult MCPToolsResult
	if err := json.Unmarshal(result, &toolsResult); err != nil {
		return nil, fmt.Errorf("failed to parse tools: %w", err)
	}

	return toolsResult.Tools, nil
}

// CallTool invokes a tool on the MCP server
func (c *MCPClient) CallTool(name string, arguments map[string]interface{}) (*MCPCallResult, error) {
	result, err := c.call("tools/call", map[string]interface{}{
		"name":      name,
		"arguments": arguments,
	})
	if err != nil {
		return nil, err
	}

	var callResult MCPCallResult
	if err := json.Unmarshal(result, &callResult); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	return &callResult, nil
}

// call sends a JSON-RPC request and waits for response
func (c *MCPClient) call(method string, params interface{}) (json.RawMessage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := c.nextID
	c.nextID++

	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Send request with newline delimiter
	if _, err := c.stdin.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resp MCPResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("MCP error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	return resp.Result, nil
}

// notify sends a JSON-RPC notification (no response expected)
func (c *MCPClient) notify(method string, params interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	req := struct {
		JSONRPC string      `json:"jsonrpc"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params,omitempty"`
	}{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, err = c.stdin.Write(append(data, '\n'))
	return err
}

// Close shuts down the MCP client
func (c *MCPClient) Close() error {
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

// MCPManager manages multiple MCP server connections
type MCPManager struct {
	clients map[string]*MCPClient
	mu      sync.RWMutex
}

// NewMCPManager creates a new MCP manager
func NewMCPManager() *MCPManager {
	return &MCPManager{
		clients: make(map[string]*MCPClient),
	}
}

// Connect starts an MCP server and adds it to the manager
func (m *MCPManager) Connect(ctx context.Context, name, command string, args ...string) error {
	client, err := NewMCPClient(ctx, command, args...)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.clients[name] = client
	m.mu.Unlock()

	return nil
}

// GetClient returns a client by name
func (m *MCPManager) GetClient(name string) (*MCPClient, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	client, ok := m.clients[name]
	return client, ok
}

// ListAllTools returns all tools from all connected servers
func (m *MCPManager) ListAllTools() (map[string][]MCPTool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string][]MCPTool)
	for name, client := range m.clients {
		tools, err := client.ListTools()
		if err != nil {
			continue // Skip failed servers
		}
		result[name] = tools
	}
	return result, nil
}

// Close shuts down all MCP connections
func (m *MCPManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, client := range m.clients {
		client.Close()
	}
	m.clients = make(map[string]*MCPClient)
}
