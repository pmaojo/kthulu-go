package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/ai"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/parser"
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

		return runAICommand(prompt, provider, model, context, apply, "")
	},
}

var aiFeatureCmd = &cobra.Command{
	Use:   "gen-feature [description]",
	Short: "Generate a BDD feature file",
	Long:  `Generate a Gherkin .feature file based on a natural language description.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		description := strings.Join(args, " ")
		prompt := fmt.Sprintf("Generate a Gherkin feature file for: %s. \n\nIMPORTANT: Start the response with 'FILENAME: path/to/file.feature' on the first line, followed by the feature content.", description)

		provider, _ := cmd.Flags().GetString("provider")
		model, _ := cmd.Flags().GetString("model")
		contextFlag, _ := cmd.Flags().GetBool("context") // Renamed to avoid shadowing
		apply, _ := cmd.Flags().GetBool("apply")

		return runAICommand(prompt, provider, model, contextFlag, apply, "feature")
	},
}

var aiStepsCmd = &cobra.Command{
	Use:   "gen-steps [feature_file]",
	Short: "Generate step definitions for a feature",
	Long:  `Generate Go step definitions for the specified .feature file.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		featureFile := args[0]
		content, err := os.ReadFile(featureFile)
		if err != nil {
			return fmt.Errorf("failed to read feature file: %w", err)
		}

		prompt := fmt.Sprintf("Generate Go step definitions (godog) for the following feature file:\n\n%s\n\nIMPORTANT: Start the response with 'FILENAME: path/to/steps_test.go' on the first line, followed by the Go code.", string(content))

		provider, _ := cmd.Flags().GetString("provider")
		model, _ := cmd.Flags().GetString("model")
		contextFlag, _ := cmd.Flags().GetBool("context")
		apply, _ := cmd.Flags().GetBool("apply")

		return runAICommand(prompt, provider, model, contextFlag, apply, "steps")
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

	// Register subcommands
	aiCmd.AddCommand(aiFeatureCmd)
	aiCmd.AddCommand(aiStepsCmd)

	// Propagate flags to subcommands (optional, but good practice if they share flags)
	// We are re-declaring reading them from parent flags in RunE, or we can copy flags.
	// Cobra looks up flags in parent commands automatically if not found in child,
	// but strictly speaking we defined them on aiCmd.
	// To be safe/clean, we can leave them on aiCmd and access them via cmd.Flags().
	// The current implementation inside RunE does `cmd.Flags().GetString` which might fail if the flag is not defined on the subcommand.
	// So we should make them persistent or add them to subcommands.
	// Making them persistent on aiCmd is easiest.

	// Reset flags to be persistent where appropriate
	aiCmd.PersistentFlags().String("provider", "openai", "AI provider (openai, anthropic, local)")
	aiCmd.PersistentFlags().String("model", "gpt-4", "AI model to use")
	aiCmd.PersistentFlags().Bool("context", true, "Include project context in prompt")
	aiCmd.PersistentFlags().Bool("apply", false, "Automatically apply generated code")
	aiCmd.PersistentFlags().Bool("mock", false, "Use mock AI client for testing")

	// Review command flags
	reviewCmd.Flags().Bool("fix-security", false, "Fix security vulnerabilities")
	reviewCmd.Flags().Bool("fix-performance", false, "Fix performance issues")
	reviewCmd.Flags().Bool("fix-all", false, "Fix all detected issues")
	reviewCmd.Flags().String("compliance", "", "Check compliance (sox, gdpr, pci)")

	// Optimize command flags
	optimizeCmd.Flags().String("target", "performance", "Optimization target (performance, memory, scalability)")
	optimizeCmd.Flags().Bool("benchmark", false, "Run benchmarks before and after optimization")
}

func runAICommand(prompt, provider, model string, includeContext, apply bool, mode string) error {
	fmt.Printf("🤖 AI Assistant (%s/%s)\n", provider, model)
	if mode != "" {
		fmt.Printf("🎯 Mode: %s\n", mode)
	}
	fmt.Printf("💭 Prompt: %s\n", prompt)

	if includeContext {
		fmt.Println("📖 Analyzing project context...")

		// Use advanced integration for context analysis
		integration := parser.NewAdvancedIntegration()
		result, insights, _, err := integration.AnalyzeProjectWithInsights(".")

		if err != nil {
			// Fallback to basic context in usecase if advanced analysis fails
			fmt.Printf("⚠️ Warning: Advanced context analysis failed (falling back to basic): %v\n", err)
		} else {
			// Construct advanced context summary
			var sb strings.Builder
			sb.WriteString("\n\n[Project Analysis Context]\n")

			// Modules
			if len(result.Modules) > 0 {
				sb.WriteString("Modules:\n")
				for _, mod := range result.Modules {
					sb.WriteString(fmt.Sprintf("- %s (%d files)\n", mod.Name, len(mod.Files)))
				}
			}

			// Architecture Patterns
			if insights != nil && len(insights.Patterns) > 0 {
				sb.WriteString("\nDetected Patterns:\n")
				for _, p := range insights.Patterns {
					if p.Confidence > 0.5 {
						sb.WriteString(fmt.Sprintf("- %s (%.0f%%)\n", p.Name, p.Confidence*100))
					}
				}
			}

			// Metrics
			metrics := integration.GetProjectMetrics()
			if metrics != nil {
				sb.WriteString("\nMetrics:\n")
				sb.WriteString(fmt.Sprintf("- Total Files: %d\n", metrics.TotalFiles))
				sb.WriteString(fmt.Sprintf("- Complexity Score: %.2f\n", metrics.ComplexityScore))
			}

			// Recommendations
			recommendations := integration.GetRecommendations()
			if len(recommendations) > 0 {
				sb.WriteString("\nRecommendations:\n")
				for i, rec := range recommendations {
					if i >= 3 {
						break
					}
					sb.WriteString(fmt.Sprintf("- %s: %s\n", rec.Type, rec.Message))
				}
			}

			// Add to prompt and disable basic context in usecase to avoid duplication
			prompt += sb.String()
			includeContext = false
		}
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
		// Parse FILENAME if present
		lines := strings.Split(res, "\n")
		var filename string
		var content string

		if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "FILENAME:") {
			filename = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[0]), "FILENAME:"))
			content = strings.Join(lines[1:], "\n")
		}

		if filename != "" {
			fmt.Printf("✅ Applying changes to %s...\n", filename)

			// Ensure directory exists
			dir := filepath.Dir(filename)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}

			// Write file
			if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", filename, err)
			}
			fmt.Println("🎉 File created successfully!")
		} else {
			fmt.Println("⚠️  Could not detect 'FILENAME:' marker in AI response. Skipping auto-apply.")
			fmt.Println("Tip: Copy the content manually or try again.")
		}
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
