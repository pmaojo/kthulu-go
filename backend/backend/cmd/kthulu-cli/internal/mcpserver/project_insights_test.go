package mcpserver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/parser"
)

func TestProjectInsightsServiceBuildOverview(t *testing.T) {
	projectDir := writeProjectFixture(t)
	service := NewProjectInsightsService(parser.NewTagParser(nil))

	overview, err := service.BuildOverview(projectDir)
	if err != nil {
		t.Fatalf("BuildOverview returned error: %v", err)
	}

	if !strings.Contains(overview, "Modules: 1") {
		t.Fatalf("expected module count in overview, got: %s", overview)
	}
	if !strings.Contains(overview, "Dependencies: 1") {
		t.Fatalf("expected dependency count in overview, got: %s", overview)
	}
	if !strings.Contains(overview, "Tags: 3") {
		t.Fatalf("expected tag count in overview, got: %s", overview)
	}
}

func TestProjectInsightsServiceDescribeModules(t *testing.T) {
	projectDir := writeProjectFixture(t)
	service := NewProjectInsightsService(parser.NewTagParser(nil))

	description, err := service.DescribeModules(projectDir)
	if err != nil {
		t.Fatalf("DescribeModules returned error: %v", err)
	}

	if !strings.Contains(description, "payments") {
		t.Fatalf("expected module name in description: %s", description)
	}
	if !strings.Contains(description, "Files: 1") {
		t.Fatalf("expected file count in description: %s", description)
	}
	if !strings.Contains(description, "module.go") {
		t.Fatalf("expected module file listing in description: %s", description)
	}
}

func TestProjectInsightsServiceDescribeTags(t *testing.T) {
	projectDir := writeProjectFixture(t)
	service := NewProjectInsightsService(parser.NewTagParser(nil))

	description, err := service.DescribeTags(projectDir)
	if err != nil {
		t.Fatalf("DescribeTags returned error: %v", err)
	}

	for _, expected := range []string{"module", "dependency", "service"} {
		if !strings.Contains(description, expected) {
			t.Fatalf("expected tag type %s in description: %s", expected, description)
		}
	}
}

func TestProjectInsightsServiceDescribeDependencies(t *testing.T) {
	projectDir := writeProjectFixture(t)
	service := NewProjectInsightsService(parser.NewTagParser(nil))

	description, err := service.DescribeDependencies(projectDir)
	if err != nil {
		t.Fatalf("DescribeDependencies returned error: %v", err)
	}

	if !strings.Contains(description, "payments -> core") {
		t.Fatalf("expected dependency relationship in description: %s", description)
	}
}

func writeProjectFixture(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	fileOne := `package payments

// @kthulu:module:payments
// @kthulu:dependency:core
// @kthulu:service:charge

type Processor struct{}
`

	fileTwo := `package payments

func Charge(amount int) int {
    return amount
}
`

	if err := os.WriteFile(filepath.Join(dir, "module.go"), []byte(fileOne), 0o600); err != nil {
		t.Fatalf("failed to write module fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "charge.go"), []byte(fileTwo), 0o600); err != nil {
		t.Fatalf("failed to write charge fixture: %v", err)
	}

	return dir
}
