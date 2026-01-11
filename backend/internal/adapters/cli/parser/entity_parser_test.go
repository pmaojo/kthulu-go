package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntityParser(t *testing.T) {
	// Create a temporary Go file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_entity.go")

	content := `package domain

import "time"

type TestEntity struct {
	ID        uint      ` + "`json:\"id\"`" + `
	Name      string    ` + "`json:\"name\" validate:\"required\"`" + `
	Age       int       ` + "`json:\"age\"`" + `
	CreatedAt time.Time ` + "`json:\"createdAt\"`" + `
	Active    bool
}
`
	err := os.WriteFile(filePath, []byte(content), 0644)
	assert.NoError(t, err)

	p := NewEntityParser(tmpDir)
	entities, err := p.ParseFile(filePath)

	assert.NoError(t, err)
	assert.Len(t, entities, 1)

	entity := entities[0]
	assert.Equal(t, "TestEntity", entity.Name)
	assert.Len(t, entity.Fields, 5)

	// Check Fields
	fields := make(map[string]FieldDefinition)
	for _, f := range entity.Fields {
		fields[f.Name] = f
	}

	assert.Equal(t, "id", fields["ID"].JSONName)
	assert.Equal(t, "uint", fields["ID"].Type)

	assert.Equal(t, "name", fields["Name"].JSONName)
	assert.Equal(t, "string", fields["Name"].Type)
	assert.Contains(t, fields["Name"].ValidationRules, "required")

	assert.Equal(t, "age", fields["Age"].JSONName)
	assert.Equal(t, "int", fields["Age"].Type)

	assert.Equal(t, "time.Time", fields["CreatedAt"].Type)
	assert.Equal(t, "bool", fields["Active"].Type)
}
