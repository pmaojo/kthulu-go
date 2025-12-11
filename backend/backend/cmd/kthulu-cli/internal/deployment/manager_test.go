package deployment

import (
	"os"
	"testing"
)

func TestManager(t *testing.T) {
	// Create a temporary directory for the test project
	tmpDir, err := os.MkdirTemp("", "kthulu-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Switch to temp dir
	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change dir: %v", err)
	}
	defer os.Chdir(originalWd)

	// Create go.mod
	if err := os.WriteFile("go.mod", []byte("module test"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Initialize manager
	mgr, err := NewManager("kubernetes", "us-east-1", "default", "auto")
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test Analyze
	if err := mgr.Analyze(); err != nil {
		t.Errorf("Analyze failed: %v", err)
	}

	// Verify Dockerfile was created
	if _, err := os.Stat("Dockerfile"); os.IsNotExist(err) {
		t.Error("Dockerfile was not created")
	}

	// Test GenerateConfig
	if err := mgr.GenerateConfig(); err != nil {
		t.Errorf("GenerateConfig failed: %v", err)
	}

	// Verify configs were created
	if _, err := os.Stat("deployments/kubernetes/deployment.yaml"); os.IsNotExist(err) {
		t.Error("deployment.yaml was not created")
	}

	// Test ConfigureAutoScaling
	if err := mgr.ConfigureAutoScaling(); err != nil {
		t.Errorf("ConfigureAutoScaling failed: %v", err)
	}

	// Verify HPA
	if _, err := os.Stat("deployments/kubernetes/hpa.yaml"); os.IsNotExist(err) {
		t.Error("hpa.yaml was not created")
	}
}
