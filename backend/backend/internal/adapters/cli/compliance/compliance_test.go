package compliance_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/compliance"
	"github.com/stretchr/testify/assert"
)

func TestSOXValidator(t *testing.T) {
	// Setup temporary test directory
	tempDir, err := os.MkdirTemp("", "sox-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Fail all checks
	report, err := compliance.Validate("sox", tempDir)
	assert.NoError(t, err)
	assert.False(t, report.Passed)
	assert.Equal(t, "SOX", report.Standard)
	assert.False(t, report.Checks[0].Passed, "Audit Logging should fail")
	assert.False(t, report.Checks[1].Passed, "Change Management should fail")
	assert.False(t, report.Checks[2].Passed, "Access Control should fail")

	// Test 2: Pass checks
	os.Mkdir(filepath.Join(tempDir, "migrations"), 0755)
	os.WriteFile(filepath.Join(tempDir, "audit.go"), []byte("package main\nfunc AuditLogger(){}"), 0644)
	os.WriteFile(filepath.Join(tempDir, "rbac.go"), []byte("package main\n// RBAC middleware"), 0644)

	report, err = compliance.Validate("sox", tempDir)
	assert.NoError(t, err)
	assert.True(t, report.Passed)
}

func TestGDPRValidator(t *testing.T) {
	// Setup temporary test directory
	tempDir, err := os.MkdirTemp("", "gdpr-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Fail all checks
	report, err := compliance.Validate("gdpr", tempDir)
	assert.NoError(t, err)
	assert.False(t, report.Passed)

	// Test 2: Pass checks
	os.WriteFile(filepath.Join(tempDir, "PRIVACY.md"), []byte("# Privacy"), 0644)
	os.WriteFile(filepath.Join(tempDir, "crypto.go"), []byte("package main\nimport \"crypto/aes\""), 0644)
	os.WriteFile(filepath.Join(tempDir, "user.go"), []byte("package main\nfunc DeleteUser(){}"), 0644)

	report, err = compliance.Validate("gdpr", tempDir)
	assert.NoError(t, err)
	assert.True(t, report.Passed)
}
