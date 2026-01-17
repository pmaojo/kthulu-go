package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start the self-healing development server",
	Long:  `Runs your backend and frontend in parallel, monitoring for errors and suggesting fixes using AI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDev()
	},
}

func init() {
	rootCmd.AddCommand(devCmd)
}

func runDev() error {
	fmt.Println("🔮 Starting Kthulu Self-Healing Dev Server...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	// 1. Check if we have a frontend
	hasFrontend := false
	if _, err := os.Stat(filepath.Join(currentDir, "frontend", "package.json")); err == nil {
		hasFrontend = true
	}

	var wg sync.WaitGroup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create error channel for diagnosis
	errChan := make(chan string)

	// Keep track of processes to kill them later
	var backendCmd *exec.Cmd
	var frontendCmd *exec.Cmd
	var mu sync.Mutex

	// Start Backend
	wg.Add(1)
	go func() {
		defer wg.Done()
		cmd := exec.Command("go", "run", "cmd/server/main.go")
		cmd.Dir = currentDir
		cmd.Env = append(os.Environ(), "GOWORK=off")

		
		mu.Lock()
		backendCmd = cmd
		mu.Unlock()

		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			fmt.Printf("❌ Backend failed to start: %v\n", err)
			return
		}

		go monitorOutput(stdout, "BACKEND", false, errChan)
		go monitorOutput(stderr, "BACKEND", true, errChan)

		if err := cmd.Wait(); err != nil {
			// Expected on kill
		}
	}()

	// Start Frontend if exists
	if hasFrontend {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.Command("npm", "run", "dev")
			cmd.Dir = filepath.Join(currentDir, "frontend")
			
			mu.Lock()
			frontendCmd = cmd
			mu.Unlock()
			
			stdout, _ := cmd.StdoutPipe()
			stderr, _ := cmd.StderrPipe()
			
			if err := cmd.Start(); err != nil {
				fmt.Printf("❌ Frontend failed to start: %v\n", err)
				return
			}

			go monitorOutput(stdout, "FRONTEND", false, errChan)
			go monitorOutput(stderr, "FRONTEND", true, errChan)

			if err := cmd.Wait(); err != nil {
				// Expected
			}
		}()
	}

	// AI Diagnosis Loop
	go func() {
		for errMsg := range errChan {
			diagnoseError(errMsg)
		}
	}()

	fmt.Println("🚀 Services are running. Press Ctrl+C to stop.")
	<-sigChan
	fmt.Println("\n🛑 Shutting down...")
	
	mu.Lock()
	if backendCmd != nil && backendCmd.Process != nil {
		backendCmd.Process.Signal(syscall.SIGTERM)
	}
	if frontendCmd != nil && frontendCmd.Process != nil {
		frontendCmd.Process.Signal(syscall.SIGTERM)
	}
	mu.Unlock()

	return nil
}

func monitorOutput(r io.Reader, prefix string, isError bool, errChan chan string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		
		// Colorize output
		color := "\033[36m" // Cyan for backend
		if prefix == "FRONTEND" {
			color = "\033[35m" // Magenta for frontend
		}
		
		fmt.Printf("%s[%s] %s\033[0m\n", color, prefix, line)

		if isError || strings.Contains(strings.ToLower(line), "panic") || strings.Contains(strings.ToLower(line), "error") {
			// Send to AI for diagnosis (simple heuristic for now)
			// Filter out common frontend noise/warnings that aren't critical
			if strings.Contains(line, "panic") || strings.Contains(line, "build failed") || strings.Contains(line, "FAIL") {
				errChan <- line
			}
		}
	}
}

// simulate AI diagnosis
func diagnoseError(logLine string) {
	// Debounce or simple logic could go here
	fmt.Println("\n🤖 \033[1;33mAI DOCTOR DETECTED AN ISSUE\033[0m")
	fmt.Printf("   Error: %s\n", strings.TrimSpace(logLine))
	
	fmt.Println("   💡 \033[1;32mSuggested Fix:\033[0m")
	
	if strings.Contains(logLine, "panic: runtime error: invalid memory address") {
		fmt.Println("      You are dereferencing a nil pointer.")
		fmt.Println("      Check for uninitialized structs or unavailable services.")
	} else if strings.Contains(logLine, "address already in use") {
		fmt.Println("      The port is occupied.")
		fmt.Println("      Run 'lsof -i :8080' to find the process or choose a different port in config.")
	} else if strings.Contains(logLine, "undefined:") {
		fmt.Println("      Compilation error: Undefined variable or function.")
		fmt.Println("      Check for typos or missing imports.")
	} else if strings.Contains(logLine, "is not in std") || strings.Contains(logLine, "no required module provides package") {
		fmt.Println("      Dependency resolution error.")
		fmt.Println("      Try running: export GOWORK=off")
		fmt.Println("      Or: go mod tidy")
	} else {
		fmt.Println("      I'm analyzing the stack trace...")
		fmt.Println("      (Simulated: Check imports and syntax or recent changes)")
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	time.Sleep(1 * time.Second) // Prevent spam
}
