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

	result, err := InjectFunction(originalSource, newFunction, nil)
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
}

func TestInjectFunctionWithImports(t *testing.T) {
	originalSource := `package main
func main() {}
`
	newFunction := `func Now() time.Time { return time.Now() }`
	imports := []string{"time"}

	result, err := InjectFunction(originalSource, newFunction, imports)
	if err != nil {
		t.Fatalf("InjectFunction failed: %v", err)
	}

	if !strings.Contains(result, `import "time"`) && !strings.Contains(result, `import (
	"time"
)`) {
		// Just check for "time" anywhere in imports block is tricky with simple strings
		// but astutil usually adds it clearly.
		if !strings.Contains(result, "\"time\"") {
			t.Errorf("Result does not contain imported package 'time'. Result:\n%s", result)
		}
	}
}

func TestInjectDuplicateFunction(t *testing.T) {
	originalSource := `package main
func Existing() {}
`
	newFunction := `func Existing() {}`

	_, err := InjectFunction(originalSource, newFunction, nil)
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

	result, err := InjectFunction(originalSource, newFunction, nil)
	if err != nil {
		t.Fatalf("InjectFunction failed for distinct receiver types: %v", err)
	}

	if !strings.Contains(result, "func (s *Service) Start() {}") {
		t.Error("Original method lost")
	}
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

	_, err := InjectFunction(originalSource, newFunction, nil)
	if err == nil {
		t.Error("Expected error when injecting duplicate method, got nil")
	}
}

func TestInjectMethodReceiverCollision(t *testing.T) {
	// Case 1: Existing method has value receiver, new method has pointer receiver
	originalSource := `package main
type Service struct{}
func (s Service) Start() {}
`
	newFunction := `func (s *Service) Start() {}`

	_, err := InjectFunction(originalSource, newFunction, nil)
	if err == nil {
		t.Error("Expected error when injecting method with pointer receiver when value receiver exists, got nil")
	}

	// Case 2: Existing method has pointer receiver, new method has value receiver
	originalSource2 := `package main
type Service struct{}
func (s *Service) Stop() {}
`
	newFunction2 := `func (s Service) Stop() {}`

	_, err = InjectFunction(originalSource2, newFunction2, nil)
	if err == nil {
		t.Error("Expected error when injecting method with value receiver when pointer receiver exists, got nil")
	}
}

func TestInjectStructField(t *testing.T) {
	originalSource := `package main
type User struct {
	ID int
}
`
	result, err := InjectStructField(originalSource, "User", "Email", "string", `json:"email"`)
	if err != nil {
		t.Fatalf("InjectStructField failed: %v", err)
	}

	if !strings.Contains(result, "Email string `json:\"email\"`") {
		t.Errorf("Result does not contain new field. Result:\n%s", result)
	}
}

func TestInjectStructTag(t *testing.T) {
	originalSource := `package main
type User struct {
	ID int
}
`
	result, err := InjectStructTag(originalSource, "User", "ID", `json:"id" gorm:"primaryKey"`)
	if err != nil {
		t.Fatalf("InjectStructTag failed: %v", err)
	}

	if !strings.Contains(result, "ID int `json:\"id\" gorm:\"primaryKey\"`") {
		t.Errorf("Result does not contain updated tag. Result:\n%s", result)
	}
}
