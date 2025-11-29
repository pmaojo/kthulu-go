package injector

import (
	"strings"
	"testing"
)

func TestInjectFunction(t *testing.T) {
	originalSource := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`

	newFunction := `
// NewEndpoint is a new function injected via AST
func NewEndpoint() string {
	return "I am a new endpoint"
}
`

	expectedContains := "func NewEndpoint() string"
	expectedContainsComment := "// NewEndpoint is a new function injected via AST"

	result, err := InjectFunction(originalSource, newFunction)
	if err != nil {
		t.Fatalf("InjectFunction failed: %v", err)
	}

	if !strings.Contains(result, expectedContains) {
		t.Errorf("Result does not contain new function signature. Result:\n%s", result)
	}

	if !strings.Contains(result, expectedContainsComment) {
		t.Errorf("Result does not contain new function comment. Result:\n%s", result)
	}

	// Verify it still contains original code
	if !strings.Contains(result, `fmt.Println("Hello, World!")`) {
		t.Errorf("Result lost original code. Result:\n%s", result)
	}

	t.Logf("Generated Code:\n%s", result)
}

func TestInjectDuplicateFunction(t *testing.T) {
	originalSource := `package main
func Existing() {}
`
	newFunction := `func Existing() {}`

	_, err := InjectFunction(originalSource, newFunction)
	if err == nil {
		t.Error("Expected error when injecting duplicate function, got nil")
	}
}

func TestInjectMethodVsFunction(t *testing.T) {
	originalSource := `package main
type Service struct{}
func (s *Service) Start() {}
`
	// Same name "Start", but no receiver (function vs method) -> Should succeed
	newFunction := `func Start() {}`

	result, err := InjectFunction(originalSource, newFunction)
	if err != nil {
		t.Fatalf("InjectFunction failed for distinct receiver types: %v", err)
	}

	if !strings.Contains(result, "func (s *Service) Start() {}") {
		t.Error("Original method lost")
	}
	// We check for "func Start()" instead of exact brace matching to be safer against formatting
	if !strings.Contains(result, "func Start()") {
		t.Errorf("New function not injected. Result:\n%s", result)
	}
}

func TestInjectDuplicateMethod(t *testing.T) {
	originalSource := `package main
type Service struct{}
func (s *Service) Start() {}
`
	// Same name "Start", same receiver -> Should fail
	newFunction := `func (x *Service) Start() {}` // variable name 'x' shouldn't matter

	_, err := InjectFunction(originalSource, newFunction)
	if err == nil {
		t.Error("Expected error when injecting duplicate method, got nil")
	}
}
