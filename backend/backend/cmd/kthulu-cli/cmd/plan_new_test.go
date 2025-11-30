package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestPlanCommand(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)

	cmd := newPlanCmd()
	cmd.SetArgs([]string{"my-test-app", "--template=microservice", "--features=auth"})

	err := cmd.Execute()
	assert.NoError(t, err)

	planPath := filepath.Join(tempDir, "kthulu-plan.yaml")
	assert.FileExists(t, planPath)

	data, err := os.ReadFile(planPath)
	assert.NoError(t, err)

	var blueprint ProjectBlueprint
	err = yaml.Unmarshal(data, &blueprint)
	assert.NoError(t, err)

	assert.Equal(t, "my-test-app", blueprint.Name)
	assert.Equal(t, "microservice", blueprint.Template)
	assert.Contains(t, blueprint.Features, "auth")
	assert.Equal(t, "sqlite", blueprint.Database)
}
