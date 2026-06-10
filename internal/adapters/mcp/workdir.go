package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	mcp_golang "github.com/metoro-io/mcp-golang"
)

// sessionWorkdir holds the session-wide working directory override. MCP
// servers are started with a fixed --working-dir; agents that scaffold a
// project mid-session (kthulu create) need to retarget every subsequent
// tool call at the new project without restarting the server.
var sessionWorkdir struct {
	mu   sync.RWMutex
	path string
}

// resolveWorkdir returns the session override when set, otherwise the
// directory the tool was configured with. Every tool resolves its working
// directory through this function at call time.
func resolveWorkdir(configured string) string {
	sessionWorkdir.mu.RLock()
	defer sessionWorkdir.mu.RUnlock()
	if sessionWorkdir.path != "" {
		return sessionWorkdir.path
	}
	return configured
}

// setSessionWorkdir validates and records the session working directory.
// An empty path clears the override.
func setSessionWorkdir(base, path string) (string, error) {
	if path == "" {
		sessionWorkdir.mu.Lock()
		sessionWorkdir.path = ""
		sessionWorkdir.mu.Unlock()
		return base, nil
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(resolveWorkdir(base), path)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("directory not found: %s", abs)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", abs)
	}
	sessionWorkdir.mu.Lock()
	sessionWorkdir.path = abs
	sessionWorkdir.mu.Unlock()
	return abs, nil
}

// describeWorkdir reports the effective working directory and whether it
// looks like a kthulu project, so agents can orient themselves.
func describeWorkdir(configured string) string {
	dir := resolveWorkdir(configured)
	desc := fmt.Sprintf("Working directory: %s", dir)
	if _, err := os.Stat(filepath.Join(dir, "kthulu-plan.yaml")); err == nil {
		desc += "\nProject: kthulu project (kthulu-plan.yaml present)"
	} else if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		desc += "\nProject: Go module (go.mod present, no kthulu-plan.yaml)"
	} else {
		desc += "\nProject: no Go module detected here"
	}
	if entries, err := os.ReadDir(dir); err == nil {
		names := make([]string, 0, len(entries))
		for i, e := range entries {
			if i >= 25 {
				names = append(names, "...")
				break
			}
			name := e.Name()
			if e.IsDir() {
				name += "/"
			}
			names = append(names, name)
		}
		desc += fmt.Sprintf("\nEntries: %v", names)
	}
	return desc
}

// WorkdirSetArgs are the arguments for the workdir_set tool.
type WorkdirSetArgs struct {
	Path string `json:"path" jsonschema:"required,description=Directory to use as the working directory for all subsequent tool calls. Relative paths resolve against the current working directory. Empty resets to the server default."`
}

// WorkdirGetArgs are the arguments for the workdir_get tool.
type WorkdirGetArgs struct{}

// WorkdirTools returns the get/set working-directory tools.
func WorkdirTools(configured string) []RegisteredTool {
	return []RegisteredTool{
		{
			Name:        "workdir_get",
			Description: "Show the current working directory used by all tools, whether it is a kthulu project, and its top-level entries. Use this to orient yourself.",
			Handler: func(ctx context.Context, args WorkdirGetArgs) (*mcp_golang.ToolResponse, error) {
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(describeWorkdir(configured))), nil
			},
		},
		{
			Name:        "workdir_set",
			Description: "Change the working directory for all subsequent tool calls (e.g. after 'create' scaffolds a new project, point the session at it). Pass an empty path to reset to the server default.",
			Handler: func(ctx context.Context, args WorkdirSetArgs) (*mcp_golang.ToolResponse, error) {
				dir, err := setSessionWorkdir(configured, args.Path)
				if err != nil {
					return nil, err
				}
				msg := fmt.Sprintf("✅ Working directory set to %s\n\n%s", dir, describeWorkdir(configured))
				return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(msg)), nil
			},
		},
	}
}
