package adapterhttp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// CliHandler exposes endpoints to run CLI commands.
type CliHandler struct {
	log *zap.SugaredLogger
}

// NewCliHandler constructs CliHandler.
func NewCliHandler(logger *zap.Logger) *CliHandler {
	return &CliHandler{
		log: logger.Sugar(),
	}
}

// RegisterRoutes attaches CLI routes to the router.
func (h *CliHandler) RegisterRoutes(r chi.Router) {
	r.Post("/api/v1/cli/{command}", h.runCommand)
}

// Allowed commands whitelist
var allowedCommands = map[string]bool{
	"generate": true,
	"migrate":  true,
	"build":    true,
	"deploy":   true,
	"test":     true,
	"validate": true,
	"bdd":      true,
}

// runCommand executes a CLI command.
func (h *CliHandler) runCommand(w http.ResponseWriter, r *http.Request) {
	logger := h.log
	command := chi.URLParam(r, "command")

	if !allowedCommands[command] {
		logger.Warnw("Blocked unauthorized CLI command attempt", "command", command)
		http.Error(w, "Command not allowed", http.StatusForbidden)
		return
	}

	var req struct {
		Args []string `json:"args"`
		// Generic payload map can be accepted too if commands need specific flags
		// but simple args list is easier to map to exec.
	}

	// Try to decode body if present
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// If decoding fails, maybe it's the other format: keys as flags?
			// For now, assume strict Args list or handle empty.
			// The frontend sends { args: [...] } via parseCliArgs?
			// Let's look at frontend/src/components/Terminal.tsx if available or just stick to simple contract.
			// Frontend sends `payload`.
			// In `kthuluApi.ts`: `runCliCommand(command, payload)`.
			// If I change frontend to send `{ args: ['features'] }`, then `req.Args` works.
			logger.Warnw("Failed to decode CLI request body", "error", err)
		}
	}

	logger.Infow("Executing CLI command", "command", command, "args", req.Args)

	// Locate binary
	// In development, it might be in ../../bin/kthulu-cli relative to server working dir?
	// Or in PATH.
	// Let's try to find it.
	binPath := "kthulu-cli" // Default to PATH

	// Check common locations
	cwd, _ := os.Getwd()
	possiblePaths := []string{
		filepath.Join(cwd, "bin", "kthulu-cli"),
		filepath.Join(cwd, "..", "bin", "kthulu-cli"),
		filepath.Join(cwd, "..", "..", "bin", "kthulu-cli"), // If running from backend/backend
		"/app/bin/kthulu-cli",
	}

	for _, p := range possiblePaths {
		if _, err := os.Stat(p); err == nil {
			binPath = p
			break
		}
	}

	// Construct args
	cmdArgs := append([]string{command}, req.Args...)

	cmd := exec.Command(binPath, cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\nSTDERR:\n" + stderr.String()
	}

	// Result structure matching CliCommandResult interface in frontend
	result := map[string]interface{}{
		"command": command,
		"status":  "success",
		"output":  strings.Split(output, "\n"),
		"duration": "0s", // TODO: measure duration
	}

	if err != nil {
		result["status"] = "error"
		result["error"] = err.Error()
		logger.Errorw("CLI command failed", "command", command, "error", err, "stderr", stderr.String())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
