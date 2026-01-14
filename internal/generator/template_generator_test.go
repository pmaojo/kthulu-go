package generator

import (
	"testing"

	"github.com/pmaojo/kthulu-go/internal/resolver"
	"github.com/stretchr/testify/assert"
)

func TestGenerateProject_Templates(t *testing.T) {
	// Setup
	mockResolver := &resolver.DependencyResolver{} // You might need to mock this properly if it has logic
	gen := NewTemplateGenerator(mockResolver)

	tests := []struct {
		name         string
		config       *GeneratorConfig
		wantFiles    []string
		wantDirs     []string
		wantTemplate string
	}{
		{
			name: "CLI Template",
			config: &GeneratorConfig{
				ProjectName:  "my-cli",
				TemplateType: "cli",
				OutputPath:   "/tmp/my-cli",
			},
			wantFiles: []string{
				"cmd/my-cli/main.go",
				"internal/cli/root.go",
				"go.mod",
				"README.md",
			},
			wantDirs: []string{
				"cmd/my-cli",
				"internal/cli",
			},
		},
		{
			name: "MCP Template",
			config: &GeneratorConfig{
				ProjectName:  "my-mcp",
				TemplateType: "mcp",
				OutputPath:   "/tmp/my-mcp",
			},
			wantFiles: []string{
				"cmd/my-mcp/main.go",
				"internal/tools/tools.go",
				"go.mod",
				"README.md",
			},
			wantDirs: []string{
				"cmd/my-mcp",
				"internal/tools",
			},
		},
		{
			name: "Server Template (Default)",
			config: &GeneratorConfig{
				ProjectName:  "my-server",
				TemplateType: "server",
				OutputPath:   "/tmp/my-server",
			},
			wantFiles: []string{
				"cmd/server/main.go",
				"go.mod",
				"README.md",
			},
			wantDirs: []string{
				"cmd/server",
				"internal/core",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock resolver response for empty dependencies
			// Since we can't easily mock the struct method without an interface,
			// we rely on the fact that ResolveDependencies with empty features returns empty plan
			// (Assuming the real implementation does that)

			structure, err := gen.GenerateProject(tt.config)
			assert.NoError(t, err)
			assert.NotNil(t, structure)

			// Check directories
			for _, dir := range tt.wantDirs {
				found := false
				for _, d := range structure.Directories {
					if d == dir {
						found = true
						break
					}
				}
				assert.True(t, found, "Directory %s not found in generated structure", dir)
			}

			// Check files
			for _, file := range tt.wantFiles {
				found := false
				for _, f := range structure.Files {
					if f.Path == file {
						found = true
						break
					}
				}
				assert.True(t, found, "File %s not found in generated structure", file)
			}
		})
	}
}
