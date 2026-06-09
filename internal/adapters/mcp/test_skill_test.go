package mcp_test

import (
	"context"
	"testing"

	"github.com/pmaojo/kthulu-go/internal/adapters/mcp"
	"github.com/stretchr/testify/require"
)

func TestGoTestServiceRunTests(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "go.mod", "module example.com/tiny\n\ngo 1.24\n")
	writeTestFile(t, dir, "tiny.go", "package tiny\n\nfunc Add(a, b int) int { return a + b }\n")
	writeTestFile(t, dir, "tiny_test.go", `package tiny

import "testing"

func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Fatal("math is broken")
	}
}
`)

	service := mcp.NewGoTestService()
	result, err := service.RunTests(context.Background(), dir, mcp.GoTestArgs{Packages: ".", Verbose: true})
	require.NoError(t, err)
	require.Contains(t, result, "TestAdd")
	require.Contains(t, result, "PASS")
	require.Contains(t, result, "Exit status: ok")
}

func TestGoTestServiceReportsFailuresAsOutput(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "go.mod", "module example.com/broken\n\ngo 1.24\n")
	writeTestFile(t, dir, "broken_test.go", `package broken

import "testing"

func TestAlwaysFails(t *testing.T) {
	t.Fatal("intentional failure")
}
`)

	service := mcp.NewGoTestService()
	result, err := service.RunTests(context.Background(), dir, mcp.GoTestArgs{Packages: "."})
	require.NoError(t, err, "test failures should be reported as output, not handler errors")
	require.Contains(t, result, "intentional failure")
	require.Contains(t, result, "Exit status: exit status 1")
}
