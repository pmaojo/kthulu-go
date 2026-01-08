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
analyze the stack trace using AI, and suggest a fix immediately.`,
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
	
	panicBuffer := &strings.Builder{}
	capturePanic := false
	go scanStderr(&wg, stderrPipe, panicBuffer, &capturePanic)

	wg.Wait()
	_ = cmd.Wait()
	
	if capturePanic && panicBuffer.Len() > 0 {
		analyzeErrorLoop(panicBuffer.String(), provider, model)
		return true
	}

	return false
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

func scanStderr(wg *sync.WaitGroup, pipe io.ReadCloser, buffer *strings.Builder, capture *bool) {
	defer wg.Done()
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintln(os.Stderr, line)

		if isCriticalError(line) {
			*capture = true
			color.HiRed("\n🚨 DETECTED CRITICAL ERROR! ANALYZING... 🚨\n")
		}

		if *capture {
			buffer.WriteString(line + "\n")
		}
	}
}

func isCriticalError(line string) bool {
	return strings.Contains(line, "panic:") || 
		strings.Contains(line, "fatal error:") || 
		strings.Contains(line, "Error:")
}

func analyzeErrorLoop(logOutput, provider, model string) {
	fmt.Println("🤖 Kthulu AI is diagnosing the crash...")

	prompt := fmt.Sprintf(`
You are a Golang expert debugger.
The running application just crashed with the following stderr output.
Explain what went wrong in 1 sentence and provide the fix code block.

STDERR:
%s
`, logOutput)

	// Create AI Client
	client, err := createAIClient(provider, model)
	if err != nil {
		color.Red("⚠️  Could not start AI debugger: %v", err)
		return
	}
	defer client.Close()

	uc := usecase.NewAIUseCase(client)
	diagnosis, err := uc.Suggest(context.Background(), prompt, true, ".") // Include local context!
	
	if err != nil {
		color.Red("⚠️  AI Diagnosis failed: %v", err)
		return
	}

	color.HiCyan("\n🏥 AI DIAGNOSIS:")
	fmt.Println(diagnosis)
	fmt.Println("\n---------------------------------------------------")
	color.HiGreen("Tip: Apply the fix above and saving will trigger restart.")
}
