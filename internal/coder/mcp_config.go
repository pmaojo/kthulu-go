package coder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// MCPServerConfig defines an MCP server to connect to
type MCPServerConfig struct {
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Env     []string `json:"env,omitempty"`
}

// DefaultMCPServers returns the default MCP servers to connect to
func DefaultMCPServers() []MCPServerConfig {
	servers := []MCPServerConfig{}

	// Check if kthulu binary exists and add it as MCP server
	kthuluPath, err := findKthuluBinary()
	if err == nil {
		servers = append(servers, MCPServerConfig{
			Name:    "kthulu",
			Command: kthuluPath,
			Args:    []string{"mcp"},
		})
	}

	return servers
}

// findKthuluBinary locates the kthulu binary
func findKthuluBinary() (string, error) {
	// Check if it's in PATH
	if path, err := exec.LookPath("kthulu"); err == nil {
		return path, nil
	}

	// Check current directory
	cwd, err := os.Getwd()
	if err == nil {
		localPath := filepath.Join(cwd, "kthulu")
		if _, err := os.Stat(localPath); err == nil {
			return localPath, nil
		}
	}

	return "", fmt.Errorf("kthulu binary not found")
}

// InitializeMCPServers connects to all configured MCP servers
func InitializeMCPServers(ctx context.Context, manager *MCPManager) []string {
	var connected []string

	for _, server := range DefaultMCPServers() {
		err := manager.Connect(ctx, server.Name, server.Command, server.Args...)
		if err != nil {
			// Log but don't fail - MCP is optional
			continue
		}
		connected = append(connected, server.Name)
	}

	return connected
}
