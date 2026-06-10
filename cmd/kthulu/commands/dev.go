package commands

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	errChan := make(chan string)

	var backendCmd *exec.Cmd
	var frontendCmd *exec.Cmd
	var mu sync.Mutex

	// Start Backend
	wg.Add(1)
	go func() {
		defer wg.Done()

		if _, err := os.Stat(filepath.Join(currentDir, "internal", "views")); err == nil {
			if err := runTemplGenerate(currentDir); err != nil {
				fmt.Printf("⚠️  Templ generation failed: %v\n", err)
			}
		}

		cmd := exec.Command("go", "run", "./cmd/server")
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

		color := "\033[36m" // Cyan for backend
		if prefix == "FRONTEND" {
			color = "\033[35m" // Magenta for frontend
		}

		fmt.Printf("%s[%s] %s\033[0m\n", color, prefix, line)

		if isError || strings.Contains(strings.ToLower(line), "panic") || strings.Contains(strings.ToLower(line), "error") {
			if strings.Contains(line, "panic") || strings.Contains(line, "build failed") || strings.Contains(line, "FAIL") {
				errChan <- line
			}
		}
	}
}

// diagnoseError prints a diagnosis for a detected panic or build failure.
// It tries the OpenAI API when OPENAI_API_KEY is set; otherwise falls back
// to fast heuristic suggestions so the output is always useful.
func diagnoseError(logLine string) {
	fmt.Println("\n🤖 \033[1;33mAI DOCTOR DETECTED AN ISSUE\033[0m")
	fmt.Printf("   Error: %s\n", strings.TrimSpace(logLine))
	fmt.Println("   💡 \033[1;32mSuggested Fix:\033[0m")

	if suggestion := aiDiagnose(logLine); suggestion != "" {
		for _, line := range strings.Split(strings.TrimSpace(suggestion), "\n") {
			fmt.Println("      " + line)
		}
	} else {
		heuristicDiagnose(logLine)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	time.Sleep(1 * time.Second) // Prevent spam
}

// aiDiagnose calls the OpenAI chat completions API and returns a short
// diagnosis. Returns "" if no API key is set or the call fails.
func aiDiagnose(logLine string) string {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return ""
	}

	prompt := fmt.Sprintf(
		"I am running a Go application that just produced this error or panic log line:\n\n%s\n\n"+
			"Give me a concise diagnosis (2-4 lines): what caused it and the most likely fix. "+
			"Be specific, no preamble.",
		strings.TrimSpace(logLine),
	)

	body, _ := json.Marshal(map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  256,
		"temperature": 0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || len(result.Choices) == 0 {
		return ""
	}
	return result.Choices[0].Message.Content
}

// heuristicDiagnose prints a fast pattern-matched suggestion when no AI key
// is available.
func heuristicDiagnose(logLine string) {
	switch {
	case strings.Contains(logLine, "panic: runtime error: invalid memory address"):
		fmt.Println("      Nil pointer dereference.")
		fmt.Println("      Check for uninitialized structs or services that weren't started.")
	case strings.Contains(logLine, "address already in use"):
		fmt.Println("      Port is already occupied.")
		fmt.Println("      Run: lsof -i :8080  (or change the port in configs/app.yaml)")
	case strings.Contains(logLine, "undefined:"):
		fmt.Println("      Undefined variable or function — check for typos or missing imports.")
	case strings.Contains(logLine, "is not in std"), strings.Contains(logLine, "no required module provides package"):
		fmt.Println("      Missing dependency. Try: export GOWORK=off && go mod tidy")
	case strings.Contains(logLine, "build failed"), strings.Contains(logLine, "FAIL"):
		fmt.Println("      Build failure — run: go build ./... to see full compiler output.")
	default:
		fmt.Println("      Set OPENAI_API_KEY for AI-powered diagnosis.")
		fmt.Println("      Or run: kthulu debug  for the full interactive runtime monitor.")
	}
}

