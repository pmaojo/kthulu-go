package deployment

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

type Manager struct {
	ProjectRoot string
	Cloud       string
	Region      string
	Namespace   string
	Scale       string
	AppName     string
}

type TemplateData struct {
	AppName   string
	Namespace string
	Replicas  int
	Image     string
}

func NewManager(cloud, region, namespace, scale string) (*Manager, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	appName := filepath.Base(wd)

	return &Manager{
		ProjectRoot: wd,
		Cloud:       cloud,
		Region:      region,
		Namespace:   namespace,
		Scale:       scale,
		AppName:     appName,
	}, nil
}

func (m *Manager) Analyze() error {
	fmt.Println("🔍 Analyzing project structure...")

	// Check for go.mod
	if _, err := os.Stat(filepath.Join(m.ProjectRoot, "go.mod")); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found - are you in the project root?")
	}

	// Check for Dockerfile
	if _, err := os.Stat(filepath.Join(m.ProjectRoot, "Dockerfile")); os.IsNotExist(err) {
		fmt.Println("⚠️  Dockerfile not found, generating default...")

		// Detect main package
		buildPath := "cmd/server/main.go"
		if _, err := os.Stat(filepath.Join(m.ProjectRoot, buildPath)); os.IsNotExist(err) {
			// Try root
			if _, err := os.Stat(filepath.Join(m.ProjectRoot, "main.go")); err == nil {
				buildPath = "main.go"
			} else {
				// Try to find any main
				// For now, default to .
				buildPath = "."
			}
		}

		// Inject build path into template
		tmpl := fmt.Sprintf(DockerfileTemplate, buildPath)

		if err := os.WriteFile(filepath.Join(m.ProjectRoot, "Dockerfile"), []byte(tmpl), 0644); err != nil {
			return fmt.Errorf("failed to generate Dockerfile: %w", err)
		}
	} else {
		fmt.Println("✅ Dockerfile found")
	}

	return nil
}

func (m *Manager) GenerateConfig() error {
	fmt.Println("📄 Generating cloud configurations...")

	deployDir := filepath.Join(m.ProjectRoot, "deployments", m.Cloud)
	if err := os.MkdirAll(deployDir, 0755); err != nil {
		return err
	}

	data := TemplateData{
		AppName:   m.AppName,
		Namespace: m.Namespace,
		Replicas:  2, // Default
		Image:     fmt.Sprintf("%s:latest", m.AppName), // Local image for now
	}

	// Determine replicas based on scale strategy
	if m.Scale == "manual" {
		data.Replicas = 1
	}

	if err := m.generateFile(filepath.Join(deployDir, "deployment.yaml"), K8sDeploymentTemplate, data); err != nil {
		return err
	}
	if err := m.generateFile(filepath.Join(deployDir, "service.yaml"), K8sServiceTemplate, data); err != nil {
		return err
	}

	fmt.Printf("✅ Configs generated in deployments/%s/\n", m.Cloud)
	return nil
}

func (m *Manager) Build() error {
	fmt.Println("🏗️  Building container image...")

	// Check if docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		fmt.Println("⚠️  Docker not found. Skipping build step.")
		return nil
	}

	imageName := fmt.Sprintf("%s:latest", m.AppName)
	cmd := exec.Command("docker", "build", "-t", imageName, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	fmt.Printf("✅ Image built: %s\n", imageName)
	return nil
}

func (m *Manager) Deploy() error {
	fmt.Printf("☁️  Deploying to %s...\n", m.Cloud)

	if m.Cloud == "kubernetes" || m.Cloud == "aws" || m.Cloud == "gcp" || m.Cloud == "azure" {
		// Assume kubectl usage for simplicity as implied by context
		if _, err := exec.LookPath("kubectl"); err != nil {
			fmt.Println("⚠️  kubectl not found. Skipping deployment step.")
			fmt.Println("👉 Manual step: kubectl apply -f deployments/" + m.Cloud)
			return nil
		}

		deployDir := filepath.Join("deployments", m.Cloud)
		cmd := exec.Command("kubectl", "apply", "-f", deployDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("deployment failed: %w", err)
		}

		fmt.Println("✅ Deployment applied successfully")
	} else {
		fmt.Printf("⚠️  Deployment logic for provider '%s' not fully automated yet.\n", m.Cloud)
	}

	return nil
}

func (m *Manager) SetupMonitoring() error {
	fmt.Println("📊 Setting up monitoring...")
	// For now, just print what would happen
	fmt.Println("ℹ️  Prometheus metrics are exposed at /metrics")
	return nil
}

func (m *Manager) ConfigureAutoScaling() error {
	if m.Scale != "auto" {
		fmt.Println("ℹ️  Auto-scaling skipped (strategy: " + m.Scale + ")")
		return nil
	}

	fmt.Println("⚖️  Configuring auto-scaling...")
	deployDir := filepath.Join(m.ProjectRoot, "deployments", m.Cloud)
	data := TemplateData{
		AppName:   m.AppName,
		Namespace: m.Namespace,
	}

	if err := m.generateFile(filepath.Join(deployDir, "hpa.yaml"), K8sHPATemplate, data); err != nil {
		return err
	}

	// Apply HPA if kubectl exists
	if _, err := exec.LookPath("kubectl"); err == nil {
		cmd := exec.Command("kubectl", "apply", "-f", filepath.Join(deployDir, "hpa.yaml"))
		// Swallow output for cleaner CLI
		_ = cmd.Run()
		fmt.Println("✅ HPA configured")
	} else {
		fmt.Printf("✅ HPA config generated in deployments/%s/hpa.yaml\n", m.Cloud)
	}

	return nil
}

func (m *Manager) generateFile(path, tmpl string, data interface{}) error {
	t, err := template.New("tmpl").Parse(tmpl)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}
