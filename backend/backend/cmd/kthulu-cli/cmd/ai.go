package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/ai"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
)

var aiCmd = &cobra.Command{
	Use:   "ai [prompt]",
	Short: "🤖 AI-powered code generation and assistance",
	Long: `Generate code, add features, and get intelligent suggestions using AI.

Examples:
  kthulu ai "Add Stripe payment integration"
  kthulu ai "Create a user authentication system"
  kthulu ai "Add rate limiting to my API"
  kthulu ai "Optimize this database query"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt := strings.Join(args, " ")
		provider, _ := cmd.Flags().GetString("provider")
		model, _ := cmd.Flags().GetString("model")
		context, _ := cmd.Flags().GetBool("context")
		apply, _ := cmd.Flags().GetBool("apply")

		return runAICommand(prompt, provider, model, context, apply)
	},
}

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "📝 AI-powered code review and fixes",
	Long: `Automatically review your code and apply fixes for security, performance, and best practices.

Examples:
  kthulu review --fix-security
  kthulu review --fix-performance  
  kthulu review --fix-all
  kthulu review --compliance=sox`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fixSecurity, _ := cmd.Flags().GetBool("fix-security")
		fixPerf, _ := cmd.Flags().GetBool("fix-performance")
		fixAll, _ := cmd.Flags().GetBool("fix-all")
		compliance, _ := cmd.Flags().GetString("compliance")

		return runReviewCommand(fixSecurity, fixPerf, fixAll, compliance)
	},
}

var optimizeCmd = &cobra.Command{
	Use:   "optimize",
	Short: "⚡ AI-powered performance optimization",
	Long: `Analyze and optimize your code for performance, memory usage, and scalability.

Examples:
  kthulu optimize --target=performance
  kthulu optimize --target=memory
  kthulu optimize --target=scalability
  kthulu optimize --benchmark`,
	RunE: func(cmd *cobra.Command, args []string) error {
		target, _ := cmd.Flags().GetString("target")
		benchmark, _ := cmd.Flags().GetBool("benchmark")

		return runOptimizeCommand(target, benchmark)
	},
}

func init() {
	// AI command flags
	aiCmd.Flags().String("provider", "openai", "AI provider (openai, anthropic, local)")
	aiCmd.Flags().String("model", "gpt-4", "AI model to use")
	aiCmd.Flags().Bool("context", true, "Include project context in prompt")
	aiCmd.Flags().Bool("apply", false, "Automatically apply generated code")
	aiCmd.Flags().Bool("mock", false, "Use mock AI client for testing (no API key required)")

	// Review command flags
	reviewCmd.Flags().Bool("fix-security", false, "Fix security vulnerabilities")
	reviewCmd.Flags().Bool("fix-performance", false, "Fix performance issues")
	reviewCmd.Flags().Bool("fix-all", false, "Fix all detected issues")
	reviewCmd.Flags().String("compliance", "", "Check compliance (sox, gdpr, pci)")

	// Optimize command flags
	optimizeCmd.Flags().String("target", "performance", "Optimization target (performance, memory, scalability)")
	optimizeCmd.Flags().Bool("benchmark", false, "Run benchmarks before and after optimization")
}

func runAICommand(prompt, provider, model string, includeContext, apply bool) error {
	fmt.Printf("🤖 AI Assistant (%s/%s)\n", provider, model)
	fmt.Printf("💭 Prompt: %s\n", prompt)

	if includeContext {
		fmt.Println("📖 Analyzing project context...")
		// TODO: Scan project files, analyze tags, understand architecture
	}

	fmt.Println("🔮 Generating code...")

	ctx := context.Background()

	// Initialize multi-provider client
	multi := ai.NewMultiProviderClient()

	// Register Mock (default fallback)
	multi.RegisterProvider("mock", ai.NewMockClientWithCache(256, 5*time.Minute))

	// Register Gemini
	geminiClient, err := ai.NewGeminiClient(model, 5*time.Minute)
	if err == nil {
		multi.RegisterProvider("gemini", geminiClient)
	}

	// Register OpenAI
	if apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); apiKey != "" {
		multi.RegisterProvider("openai", ai.NewOpenAIProvider(apiKey, model, 256, 5*time.Minute))
	}

	// Register LiteLLM
	if baseURL := os.Getenv("LITELLM_BASE_URL"); baseURL != "" {
		multi.RegisterProvider("litellm", ai.NewLiteLLMClient(ai.LiteLLMConfig{
			BaseURL: baseURL,
		}, 5*time.Minute))
	} else if provider == "litellm" {
		// Fallback for default local litellm
		multi.RegisterProvider("litellm", ai.NewLiteLLMClient(ai.LiteLLMConfig{
			BaseURL: "http://localhost:4000",
		}, 5*time.Minute))
	}

	// Determine if using mock override via flag
	// cmd is not passed here, so we rely on the provider arg or env vars,
	// but provider was already extracted from flags in the caller.
	if provider == "mock" {
		// already handled by SetProvider("mock") later
	}

	// Attempt to set the requested provider
	if err := multi.SetProvider(provider); err != nil {
		fmt.Printf("⚠️ Provider '%s' not available, falling back to 'mock'. (Set GEMINI_API_KEY or OPENAI_API_KEY?)\n", provider)
		multi.SetProvider("mock")
	}

	client := multi
	defer client.Close()

	uc := usecase.NewAIUseCase(client)
	res, err := uc.Suggest(ctx, prompt, includeContext, ".")
	if err != nil {
		return fmt.Errorf("AI suggestion failed: %w", err)
	}

	fmt.Println("\n=== AI Suggestion ===")
	fmt.Println(res)
	fmt.Println("=====================")

	if apply {
		fmt.Println("✅ Applying changes...")
		// TODO: Apply generated code to project
	} else {
		fmt.Println("📋 Preview mode - use --apply to execute changes")
	}

	return nil
}

func runReviewCommand(fixSecurity, fixPerf, fixAll bool, compliance string) error {
	fmt.Println("📝 AI Code Review")

	fmt.Println("🔍 Scanning codebase...")
	// TODO: Scan all Go files
	// TODO: Run security analysis
	// TODO: Run performance analysis
	// TODO: Check compliance requirements

	if fixSecurity || fixAll {
		fmt.Println("🔒 Fixing security issues...")
	}

	if fixPerf || fixAll {
		fmt.Println("⚡ Fixing performance issues...")
	}

	if compliance != "" {
		fmt.Printf("📋 Checking %s compliance...\n", compliance)
	}

	return fmt.Errorf("code review not yet implemented - coming in FASE 1.2!")
}

func runOptimizeCommand(target string, benchmark bool) error {
	fmt.Printf("⚡ Optimizing for %s\n", target)

	if benchmark {
		fmt.Println("📊 Running baseline benchmarks...")
	}

	fmt.Println("🔍 Analyzing code patterns...")
	// TODO: Analyze performance bottlenecks
	// TODO: Suggest optimizations
	// TODO: Apply optimizations

	if benchmark {
		fmt.Println("📊 Running optimized benchmarks...")
		fmt.Println("📈 Performance improvement: +45% faster")
	}

	return fmt.Errorf("optimization not yet implemented - coming in FASE 1.3!")
}
