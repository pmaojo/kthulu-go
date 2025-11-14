package usecase

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"backend/internal/ai"
)

// AIUseCase provides AI powered features
type AIUseCase struct {
	client ai.Client
	// simple ttl for context snapshot freshness
	contextTTL time.Duration
}

// NewAIUseCase constructs AIUseCase
func NewAIUseCase(client ai.Client) *AIUseCase {
	return &AIUseCase{client: client, contextTTL: 5 * time.Minute}
}

// Suggest generates a suggestion for a given prompt, optionally including project context
func (a *AIUseCase) Suggest(ctx context.Context, prompt string, includeContext bool, projectPath string) (string, error) {
	fullPrompt := prompt
	if includeContext {
		summary, err := a.buildProjectSummary(projectPath)
		if err == nil && summary != "" {
			fullPrompt = fmt.Sprintf("%s\n\nProject context:\n%s", prompt, summary)
		}
	}

	res, err := a.client.GenerateText(ctx, fullPrompt)
	if err != nil {
		return "", err
	}
	return res, nil
}

// buildProjectSummary collects a lightweight summary of the project
func (a *AIUseCase) buildProjectSummary(projectPath string) (string, error) {
	var sb strings.Builder
	// count modules
	modulesDir := filepath.Join(projectPath, "internal", "modules")
	count := 0
	_ = filepath.WalkDir(modulesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && path != modulesDir {
			count++
		}
		return nil
	})

	sb.WriteString(fmt.Sprintf("Modules discovered: %d\n", count))

	// read README if present
	readmePath := filepath.Join(projectPath, "README.md")
	if data, err := os.ReadFile(readmePath); err == nil {
		text := string(data)
		if len(text) > 800 {
			text = text[:800]
		}
		sb.WriteString("README: \n")
		sb.WriteString(text)
	}

	return sb.String(), nil
}
