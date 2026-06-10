package deployment

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FlyioDeployer deploys a Kthulu application to Fly.io using the fly CLI.
type FlyioDeployer struct {
	region      string
	projectRoot string
	appName     string
}

// NewFlyioDeployer creates a new FlyioDeployer for the given region.
func NewFlyioDeployer(region string) *FlyioDeployer {
	wd, _ := os.Getwd()
	appName := appNameFromGoMod(wd)
	if appName == "" {
		appName = filepath.Base(wd)
	}
	return &FlyioDeployer{
		region:      region,
		projectRoot: wd,
		appName:     appName,
	}
}

// Deploy checks prerequisites, generates fly.toml if absent, deploys, and
// prints the app URL on success.
func (d *FlyioDeployer) Deploy() error {
	// 1. Check fly CLI is installed.
	flyBin, err := exec.LookPath("fly")
	if err != nil {
		fmt.Println("❌ fly CLI not found.")
		fmt.Println("   Install it with:")
		fmt.Println("   curl -L https://fly.io/install.sh | sh")
		return fmt.Errorf("fly CLI not installed")
	}

	// 2. Ensure fly.toml exists.
	flyToml := filepath.Join(d.projectRoot, "fly.toml")
	if _, err := os.Stat(flyToml); os.IsNotExist(err) {
		fmt.Println("⚠️  fly.toml not found — generating a minimal one...")
		if err := d.generateFlyToml(flyToml); err != nil {
			return fmt.Errorf("failed to generate fly.toml: %w", err)
		}
		fmt.Println("✅ fly.toml generated")
	} else {
		fmt.Println("✅ fly.toml found")
	}

	// 3. Run fly deploy --remote-only and stream output.
	fmt.Println("🚀 Running fly deploy --remote-only...")
	deployCmd := exec.Command(flyBin, "deploy", "--remote-only")
	deployCmd.Dir = d.projectRoot

	deployStdout, _ := deployCmd.StdoutPipe()
	deployStderr, _ := deployCmd.StderrPipe()

	if err := deployCmd.Start(); err != nil {
		return fmt.Errorf("fly deploy failed to start: %w", err)
	}

	// Stream stdout
	go func() {
		scanner := bufio.NewScanner(deployStdout)
		for scanner.Scan() {
			fmt.Printf("\033[36m[FLY] %s\033[0m\n", scanner.Text())
		}
	}()

	// Stream stderr
	go func() {
		scanner := bufio.NewScanner(deployStderr)
		for scanner.Scan() {
			fmt.Printf("\033[33m[FLY] %s\033[0m\n", scanner.Text())
		}
	}()

	if err := deployCmd.Wait(); err != nil {
		return fmt.Errorf("fly deploy failed: %w", err)
	}

	fmt.Println("✅ fly deploy succeeded")

	// 4. On success, run fly status and print the app URL.
	statusCmd := exec.Command(flyBin, "status")
	statusCmd.Dir = d.projectRoot
	statusOut, err := statusCmd.Output()
	if err != nil {
		fmt.Printf("⚠️  fly status failed: %v\n", err)
	} else {
		fmt.Println(string(statusOut))
		fmt.Printf("🌐 App URL: https://%s.fly.dev\n", d.appName)
	}

	return nil
}

// generateFlyToml writes a minimal fly.toml to the given path.
func (d *FlyioDeployer) generateFlyToml(path string) error {
	region := d.region
	if region == "" {
		region = "iad"
	}

	content := fmt.Sprintf(`app = "%s"
primary_region = "%s"

[build]
  dockerfile = "Dockerfile"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 0
`, d.appName, region)

	return os.WriteFile(path, []byte(content), 0644)
}

// appNameFromGoMod reads the module name from go.mod and returns the last
// path segment (e.g. "github.com/foo/bar" → "bar").
func appNameFromGoMod(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			modulePath := strings.TrimPrefix(line, "module ")
			modulePath = strings.TrimSpace(modulePath)
			parts := strings.Split(modulePath, "/")
			if len(parts) > 0 {
				return parts[len(parts)-1]
			}
		}
	}
	return ""
}
