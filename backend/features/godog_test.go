package features

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"gopkg.in/yaml.v3"
)

var (
	binaryPath string
)

type planContext struct {
	cmdOutput string
	cmdError  error
	workDir   string
	planFile  string
}

// ProjectBlueprint mirrors the struct in cmd/plan.go (duplicated for test independence)
type ProjectBlueprint struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Template    string   `yaml:"template"`
	Features    []string `yaml:"features"`
	Database    string   `yaml:"database"`
	Frontend    string   `yaml:"frontend"`
	Auth        string   `yaml:"auth"`
	Modules     []string `yaml:"modules,omitempty"`
}

func (p *planContext) iRunThePlanCommandWith(args string) error {
	argList := strings.Fields(args)

	// Create a temporary directory for execution to avoid polluting source
	tmpDir, err := os.MkdirTemp("", "kthulu-bdd-*")
	if err != nil {
		return err
	}
	p.workDir = tmpDir
	p.planFile = filepath.Join(tmpDir, "kthulu-plan.yaml")

	// Run the compiled binary
	// We need to pass the arguments: plan <args>
	fullArgs := append([]string{"plan"}, argList...)

	cmd := exec.Command(binaryPath, fullArgs...)
	cmd.Dir = p.workDir // Execute in temp dir

	out, err := cmd.CombinedOutput()
	p.cmdOutput = string(out)
	p.cmdError = err

	return nil
}

func (p *planContext) aFileNamedShouldExist(filename string) error {
	if p.cmdError != nil {
		return fmt.Errorf("command failed: %v\nOutput: %s", p.cmdError, p.cmdOutput)
	}

	path := filepath.Join(p.workDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file %s was not created", filename)
	}
	return nil
}

func (p *planContext) thePlanShouldHaveName(expectedName string) error {
	plan, err := p.readPlan()
	if err != nil {
		return err
	}
	if plan.Name != expectedName {
		return fmt.Errorf("expected name %s, got %s", expectedName, plan.Name)
	}
	return nil
}

func (p *planContext) thePlanShouldHaveTemplate(expectedTemplate string) error {
	plan, err := p.readPlan()
	if err != nil {
		return err
	}
	if plan.Template != expectedTemplate {
		return fmt.Errorf("expected template %s, got %s", expectedTemplate, plan.Template)
	}
	return nil
}

func (p *planContext) thePlanShouldHaveDatabase(expectedDB string) error {
	plan, err := p.readPlan()
	if err != nil {
		return err
	}
	if plan.Database != expectedDB {
		return fmt.Errorf("expected database %s, got %s", expectedDB, plan.Database)
	}
	return nil
}

func (p *planContext) thePlanShouldHaveFeature(expectedFeature string) error {
	plan, err := p.readPlan()
	if err != nil {
		return err
	}

	found := false
	for _, f := range plan.Features {
		if f == expectedFeature {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("feature %s not found in plan. Features: %v", expectedFeature, plan.Features)
	}
	return nil
}

func (p *planContext) readPlan() (*ProjectBlueprint, error) {
	data, err := os.ReadFile(p.planFile)
	if err != nil {
		return nil, err
	}
	var plan ProjectBlueprint
	if err := yaml.Unmarshal(data, &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	p := &planContext{}

	// Clean up after scenario
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if p.workDir != "" {
			os.RemoveAll(p.workDir)
		}
		return ctx, nil
	})

	ctx.Step(`^I run the plan command with "([^"]*)"$`, p.iRunThePlanCommandWith)
	ctx.Step(`^a file named "([^"]*)" should exist$`, p.aFileNamedShouldExist)
	ctx.Step(`^the plan should have name "([^"]*)"$`, p.thePlanShouldHaveName)
	ctx.Step(`^the plan should have template "([^"]*)"$`, p.thePlanShouldHaveTemplate)
	ctx.Step(`^the plan should have database "([^"]*)"$`, p.thePlanShouldHaveDatabase)
	ctx.Step(`^the plan should have feature "([^"]*)"$`, p.thePlanShouldHaveFeature)
}

func TestMain(m *testing.M) {
	// Build the binary once
	tmpDir, err := os.MkdirTemp("", "kthulu-build-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath = filepath.Join(tmpDir, "kthulu-cli")

	// Assuming running from backend/backend/features
	cmdPath, err := filepath.Abs("../cmd/kthulu-cli")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get absolute path: %v\n", err)
		os.Exit(1)
	}

	buildCmd := exec.Command("go", "build", "-o", binaryPath, cmdPath)
	if out, err := buildCmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build binary: %v\nOutput: %s\n", err, out)
		os.Exit(1)
	}

	opts := godog.Options{
		Format:    "pretty",
		Paths:     []string{"."},
		Randomize: time.Now().UTC().UnixNano(),
	}

	status := godog.TestSuite{
		Name:                "kthulu-cli-bdd",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}
