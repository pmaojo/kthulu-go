package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "🛠️  Start development server with Self-Healing",
	Long: `Runs your backend in development mode with AI-powered error analysis.
    
If the application crashes or panics, Kthulu will intercept the log,
analyze the stack trace using AI, and automatically fix the issue in real-time.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		entrypoint, _ := cmd.Flags().GetString("entrypoint")
		provider, _ := cmd.Flags().GetString("provider")
		model, _ := cmd.Flags().GetString("model")
		return runDevServer(entrypoint, provider, model)
	},
}

func init() {
	devCmd.Flags().String("entrypoint", "cmd/server/main.go", "Go entrypoint file")
	devCmd.Flags().String("provider", "openai", "AI provider for healing")
	devCmd.Flags().String("model", "gpt-4", "AI model for healing")
	rootCmd.AddCommand(devCmd)
}

func runDevServer(entrypoint, provider, model string) error {
	color.Green("🛠️  Starting Kthulu Dev Server...")
	color.HiBlack("   Entrypoint: %s", entrypoint)
	color.HiBlack("   Healer:     %s (%s)", provider, model)

	// Check if entrypoint exists
	if _, err := os.Stat(entrypoint); os.IsNotExist(err) {
		// Try to fallback
		if _, err := os.Stat("main.go"); err == nil {
			entrypoint = "main.go"
			color.Yellow("⚠️  Entrypoint not found, using main.go")
		} else {
			return fmt.Errorf("entrypoint %s not found", entrypoint)
		}
	}

	for {
		shouldRestart := runProcessAndMonitor(entrypoint, provider, model)
		if !shouldRestart {
			break
		}
		color.Yellow("\n🔄 Restarting server in 2 seconds...\n")
		time.Sleep(2 * time.Second)
	}

	return nil
}

func runProcessAndMonitor(entrypoint, provider, model string) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", entrypoint)
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		color.Red("❌ Failed to start process: %v", err)
		return false
	}

	handleInterrupts(cmd)

	var wg sync.WaitGroup
	wg.Add(2)

	go scanStdout(&wg, stdoutPipe)

	// Channel to signal that a panic has been captured
	panicChan := make(chan string, 1)
	
	// Buffer and Mutex for thread-safe access
	panicBuffer := &strings.Builder{}
	var bufferMu sync.Mutex

	go scanStderr(&wg, stderrPipe, panicBuffer, &bufferMu, panicChan)

	// Channel to signal process exit
	doneChan := make(chan error, 1)
	go func() {
		doneChan <- cmd.Wait()
	}()

	select {
	case logOutput := <-panicChan:
		// Panic detected! Kill process immediately to stop the bleeding/loop
		color.HiRed("\n🚨 DETECTED CRITICAL ERROR! INITIATING SELF-HEALING... 🚨\n")
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait() // Wait for process to exit
		}

		performSelfHealing(logOutput, provider, model)
		return true // Trigger restart

	case err := <-doneChan:
		// Process exited on its own (maybe crashed or stopped)
		wg.Wait() // Wait for logs to finish processing

		if err != nil {
			// Check if we captured a panic but the timer didn't fire yet
			bufferMu.Lock()
			captured := panicBuffer.String()
			bufferMu.Unlock()

			if len(captured) > 0 && isCriticalError(captured) {
				performSelfHealing(captured, provider, model)
				return true
			}
			color.Red("❌ Process exited with error: %v", err)
		}

		return false // Exit loop on normal termination or unknown error
	}
}

func handleInterrupts(cmd *exec.Cmd) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if _, ok := <-sigChan; ok {
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		}
	}()
}

func scanStdout(wg *sync.WaitGroup, pipe io.ReadCloser) {
	defer wg.Done()
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

func scanStderr(wg *sync.WaitGroup, pipe io.ReadCloser, buffer *strings.Builder, mu *sync.Mutex, panicChan chan<- string) {
	defer wg.Done()
	scanner := bufio.NewScanner(pipe)
	var capture bool
	var once sync.Once

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintln(os.Stderr, line)

		if isCriticalError(line) {
			capture = true
			// Trigger panic handling after a short delay to allow stack trace to accumulate
			once.Do(func() {
				time.AfterFunc(500*time.Millisecond, func() {
					mu.Lock()
					logs := buffer.String()
					mu.Unlock()
					panicChan <- logs
				})
			})
		}

		if capture {
			mu.Lock()
			buffer.WriteString(line + "\n")
			mu.Unlock()
		}
	}
}

func isCriticalError(line string) bool {
	return strings.Contains(line, "panic:") || 
		strings.Contains(line, "fatal error:") || 
		strings.Contains(line, "Error:") ||
		strings.Contains(line, "Panic recovered")
}

func performSelfHealing(logOutput, provider, model string) {
	fmt.Println("🤖 Kthulu AI is diagnosing and fixing the crash...")

	prompt := fmt.Sprintf(`
You are a Golang expert debugger.
The running application just crashed with the following stderr output.
Explain what went wrong in 1 sentence and provide the corrected code.

STDERR:
%s

%s
`, logOutput, ApplyInstruction) // Add ApplyInstruction to prompt

	// Create AI Client
	client, err := createAIClient(provider, model)
	if err != nil {
		color.Red("⚠️  Could not start AI debugger: %v", err)
		return
	}
	defer client.Close()

	uc := usecase.NewAIUseCase(client)
	// We include context to help the AI understand where the error might be
	diagnosis, err := uc.Suggest(context.Background(), prompt, true, ".")
	
	if err != nil {
		color.Red("⚠️  AI Diagnosis failed: %v", err)
		return
	}

	color.HiCyan("\n🏥 AI DIAGNOSIS & FIX:")
	fmt.Println(diagnosis)
	fmt.Println("\n---------------------------------------------------")

	// Apply the fix
	color.HiGreen("⚡ Applying fixes...")
	count, err := applyAIChanges(diagnosis)
	if err != nil {
		color.Red("❌ Failed to apply changes: %v", err)
	} else if count > 0 {
		color.HiGreen("✅ Applied %d fix(es). Restarting server...", count)
	} else {
		color.Yellow("⚠️  No fixes were applied (AI did not return file blocks).")
	}
}
