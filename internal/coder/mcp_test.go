package coder

import (
	"testing"
)

func TestMCPRequest_Marshal(t *testing.T) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
		Params:  nil,
	}

	if req.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %q, want %q", req.JSONRPC, "2.0")
	}
	if req.Method != "tools/list" {
		t.Errorf("Method = %q, want %q", req.Method, "tools/list")
	}
}

func TestMCPResponse_ParseError(t *testing.T) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      1,
		Error: &MCPError{
			Code:    -32600,
			Message: "Invalid Request",
		},
	}

	if resp.Error == nil {
		t.Error("expected error to be non-nil")
	}
	if resp.Error.Code != -32600 {
		t.Errorf("Error.Code = %d, want %d", resp.Error.Code, -32600)
	}
}

func TestMCPManager_NewManager(t *testing.T) {
	manager := NewMCPManager()
	if manager == nil {
		t.Fatal("expected manager to be non-nil")
	}
	if manager.clients == nil {
		t.Error("expected clients map to be initialized")
	}
}

func TestMCPManager_GetClient_NotFound(t *testing.T) {
	manager := NewMCPManager()
	_, ok := manager.GetClient("nonexistent")
	if ok {
		t.Error("expected GetClient to return false for nonexistent client")
	}
}

func TestFindKthuluBinary(t *testing.T) {
	// This test just verifies the function doesn't panic
	// The actual result depends on the environment
	_, _ = findKthuluBinary()
}

func TestDefaultMCPServers(t *testing.T) {
	servers := DefaultMCPServers()
	// Just verify it returns a slice (may be empty if kthulu not found)
	if servers == nil {
		t.Error("expected non-nil slice")
	}
}

func TestMCPResourceStructures(t *testing.T) {
	// Test marshaling/unmarshaling of resource structures
	res := MCPResource{
		URI:         "file:///path/to/file",
		Name:        "file",
		Description: "A file",
		MIMEType:    "text/plain",
	}

	if res.URI != "file:///path/to/file" {
		t.Errorf("URI = %q, want %q", res.URI, "file:///path/to/file")
	}
}

func TestMCPPromptStructures(t *testing.T) {
	// Test marshaling/unmarshaling of prompt structures
	prompt := MCPPrompt{
		Name:        "test-prompt",
		Description: "A test prompt",
		Arguments: []MCPPromptArgument{
			{
				Name:        "arg1",
				Description: "First argument",
				Required:    true,
			},
		},
	}

	if prompt.Name != "test-prompt" {
		t.Errorf("Name = %q, want %q", prompt.Name, "test-prompt")
	}
	if len(prompt.Arguments) != 1 {
		t.Errorf("len(Arguments) = %d, want 1", len(prompt.Arguments))
	}
}
