package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/pmaojo/kthulu-go/backend/internal/infrastructure/ai"
	"github.com/pmaojo/kthulu-go/backend/internal/usecase"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/parser"
	"github.com/pmaojo/kthulu-go/backend/internal/adapters/cli/compliance"
)

const ApplyInstruction = "\n\nIMPORTANT: To apply changes, output the code using the following format for each file you want to create or update:\n<<<FILE:path/to/file>>>\n[file content]\n<<<END>>>\n"

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
		prompt := fmt.Sprintf("Generate a Gherkin feature file for: %s. Ensure the filename ends with .feature.", description)

		provider, _ := cmd.Flags().GetString("provider")
		model, _ := cmd.Flags().GetString("model")
		contextFlag, _ := cmd.Flags().GetBool("context")
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

		prompt := fmt.Sprintf("Generate Go step definitions (godog) for the following feature file:\n\n%s\n\nEnsure the filename ends with _test.go.", string(content))

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
  kthulu ai review --fix-security
  kthulu ai review --fix-performance
  kthulu ai review --fix-all
  kthulu ai review --compliance=sox`,
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
  kthulu ai optimize --target=performance
  kthulu ai optimize --target=memory
  kthulu ai optimize --target=scalability
  kthulu ai optimize --benchmark --apply`,
	RunE: func(cmd *cobra.Command, args []string) error {
		target, _ := cmd.Flags().GetString("target")
		benchmark, _ := cmd.Flags().GetBool("benchmark")
		provider, _ := cmd.Flags().GetString("provider")
		model, _ := cmd.Flags().GetString("model")
		apply, _ := cmd.Flags().GetBool("apply")

		return runOptimizeCommand(target, benchmark, provider, model, apply)
	},
}

func init() {
	// AI command flags
	aiCmd.PersistentFlags().String("provider", "openai", "AI provider (openai, anthropic, local)")
	aiCmd.PersistentFlags().String("model", "gpt-4", "AI model to use")
	aiCmd.PersistentFlags().Bool("context", true, "Include project context in prompt")
	aiCmd.PersistentFlags().Bool("apply", false, "Automatically apply generated code")
	aiCmd.PersistentFlags().Bool("mock", false, "Use mock AI client for testing")

	// Register subcommands
	aiCmd.AddCommand(aiFeatureCmd)
	aiCmd.AddCommand(aiStepsCmd)
	aiCmd.AddCommand(reviewCmd)
	aiCmd.AddCommand(optimizeCmd)

	// Review command flags
	reviewCmd.Flags().Bool("fix-security", false, "Fix security vulnerabilities")
	reviewCmd.Flags().Bool("fix-performance", false, "Fix performance issues")
	reviewCmd.Flags().Bool("fix-all", false, "Fix all detected issues")
	reviewCmd.Flags().String("compliance", "", "Check compliance (sox, gdpr, pci)")

	// Optimize command flags
	optimizeCmd.Flags().String("target", "performance", "Optimization target (performance, memory, scalability)")
	optimizeCmd.Flags().Bool("benchmark", false, "Run benchmarks before and after optimization")
}

// createAIClient handles the logic for creating and configuring the MultiProviderClient
func createAIClient(provider, model string) (*ai.MultiProviderClient, error) {
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

	// Attempt to set the requested provider
	if err := multi.SetProvider(provider); err != nil {
		fmt.Printf("⚠️ Provider '%s' not available, falling back to 'mock'. (Set GEMINI_API_KEY or OPENAI_API_KEY?)\n", provider)
		multi.SetProvider("mock")
	}

	return multi, nil
}

// analyzeContext performs advanced analysis and returns a summary string
func analyzeContext() (string, error) {
	// Use advanced integration for context analysis
	integration := parser.NewAdvancedIntegration()
	result, insights, _, err := integration.AnalyzeProjectWithInsights(".")

	if err != nil {
		return "", err
	}

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

	return sb.String(), nil
}

// applyAIChanges parses the AI response and applies file changes
func applyAIChanges(response string) (int, error) {
	// Regex for new block format
	// Matches <<<FILE:path>>> content <<<END>>>
	// Uses ?s to allow . to match newlines
	blockRegex := regexp.MustCompile(`(?s)<<<FILE:(.*?)>>>(.*?)<<<END>>>`)
	matches := blockRegex.FindAllStringSubmatch(response, -1)

	appliedCount := 0

	if len(matches) > 0 {
		for _, match := range matches {
			filename := strings.TrimSpace(match[1])
			content := match[2]

			// Clean path and check for security issues
			filename = filepath.Clean(filename)
			if strings.Contains(filename, "..") || filepath.IsAbs(filename) {
				fmt.Printf("⚠️ Skipping unsafe path: %s\n", filename)
				continue
			}

			// Trim leading/trailing newlines from format extraction
			content = strings.TrimPrefix(content, "\n")
			content = strings.TrimSuffix(content, "\n")
			// Ensure single trailing newline
			content += "\n"

			if filename != "" {
				fmt.Printf("📝 Writing to %s...\n", filename)

				// Ensure directory exists
				dir := filepath.Dir(filename)
				if err := os.MkdirAll(dir, 0755); err != nil {
					fmt.Printf("❌ Failed to create directory %s: %v\n", dir, err)
					continue
				}

				// Write file
				if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
					fmt.Printf("❌ Failed to write file %s: %v\n", filename, err)
					continue
				}
				appliedCount++
			}
		}
	}
	return appliedCount, nil
}

func runAICommand(prompt, provider, model string, includeContext, apply bool, mode string) error {
	fmt.Printf("🤖 AI Assistant (%s/%s)\n", provider, model)
	if mode != "" {
		fmt.Printf("🎯 Mode: %s\n", mode)
	}

	// Append formatting instructions if applying changes
	if apply {
		prompt += ApplyInstruction
	}

	fmt.Printf("💭 Prompt: %s\n", prompt)

	if includeContext {
		fmt.Println("📖 Analyzing project context...")
		contextSummary, err := analyzeContext()
		if err != nil {
			fmt.Printf("⚠️ Warning: Advanced context analysis failed (falling back to basic): %v\n", err)
		} else {
			prompt += contextSummary
			// Disable basic context in usecase to avoid duplication
			includeContext = false
		}
	}

	fmt.Println("🔮 Generating code...")

	ctx := context.Background()

	client, err := createAIClient(provider, model)
	if err != nil {
		return fmt.Errorf("failed to create AI client: %w", err)
	}
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
		appliedCount, err := applyAIChanges(res)
		if err != nil {
			// Currently applyAIChanges doesn't return error but for future proofing
		}

		if appliedCount > 0 {
			fmt.Printf("🎉 Successfully applied %d file(s)!\n", appliedCount)
		} else {
			fmt.Println("⚠️  Could not detect file markers in AI response. Skipping auto-apply.")
			fmt.Println("Tip: AI may not have followed the format instructions. Try again.")
		}

	} else {
		fmt.Println("📋 Preview mode - use --apply to execute changes")
	}

	return nil
}

func runReviewCommand(fixSecurity, fixPerf, fixAll bool, complianceStd string) error {
	fmt.Println("📝 AI Code Review")
	fmt.Println("🔍 Scanning codebase...")

	// Use advanced integration for context analysis
	integration := parser.NewAdvancedIntegration()
	result, insights, _, err := integration.AnalyzeProjectWithInsights(".")
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// 1. Scan Results
	fmt.Println("\n[Codebase Overview]")
	if result != nil {
		fmt.Printf("📦 Modules: %d\n", len(result.Modules))
		metrics := integration.GetProjectMetrics()
		if metrics != nil {
			fmt.Printf("📄 Files: %d\n", metrics.TotalFiles)
			fmt.Printf("🧠 Complexity Score: %.2f\n", metrics.ComplexityScore)
		}
	}

	// 2. Security Analysis
	fmt.Println("\n[Security Analysis]")
	securityIssues := 0
	for _, rec := range insights.Recommendations {
		if strings.Contains(strings.ToLower(rec.Type), "security") || strings.Contains(strings.ToLower(rec.Message), "security") {
			securityIssues++
			fmt.Printf("⚠️  %s (%s): %s\n", rec.Type, rec.Severity, rec.Message)
		}
	}
	if securityIssues == 0 {
		fmt.Println("✅ No immediate security issues detected in static analysis.")
	}

	// 3. Performance Analysis
	fmt.Println("\n[Performance Analysis]")
	perfIssues := 0
	for _, rec := range insights.Recommendations {
		if strings.Contains(strings.ToLower(rec.Type), "performance") || strings.Contains(strings.ToLower(rec.Message), "performance") {
			perfIssues++
			fmt.Printf("⚡ %s (%s): %s\n", rec.Type, rec.Severity, rec.Message)
		}
	}
	if perfIssues == 0 {
		fmt.Println("✅ No immediate performance bottlenecks detected.")
	}

	// 4. Compliance Checks
	if complianceStd != "" {
		fmt.Printf("\n[Compliance Check: %s]\n", complianceStd)
		report, err := compliance.Validate(complianceStd, ".")
		if err != nil {
			fmt.Printf("❌ Compliance check failed: %v\n", err)
		} else {
			if report.Passed {
				fmt.Printf("✅ %s Compliance Passed\n", report.Standard)
			} else {
				fmt.Printf("❌ %s Compliance Failed\n", report.Standard)
			}
			for _, check := range report.Checks {
				icon := "✅"
				if !check.Passed {
					icon = "❌"
				}
				fmt.Printf("  %s %s\n", icon, check.Name)
			}
		}
	}

	// Fix Placeholders
	if fixSecurity || fixAll {
		fmt.Println("\n🔒 Fixing security issues... (Not yet implemented)")
	}

	if fixPerf || fixAll {
		fmt.Println("⚡ Fixing performance issues... (Not yet implemented)")
	}

	return nil
}

func runOptimizeCommand(target string, benchmark bool, provider, model string, apply bool) error {
	fmt.Printf("⚡ Optimizing for %s\n", target)

	var baselineOutput string
	if benchmark {
		fmt.Println("📊 Running baseline benchmarks...")
		cmd := exec.Command("go", "test", "-bench=.", "./...")
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("⚠️ Benchmark failed: %v\n", err)
			baselineOutput = "Benchmarks failed or not present."
		} else {
			baselineOutput = string(output)
			fmt.Printf("Baseline Results:\n%s\n", baselineOutput)
		}
	}

	fmt.Println("🔍 Analyzing code patterns...")
	contextSummary, err := analyzeContext()
	if err != nil {
		fmt.Printf("⚠️ Analysis failed: %v\n", err)
		// Continue without analysis if it fails, or maybe just return?
		// We'll proceed with basic context handling in Suggest
	}

	// Construct Prompt
	prompt := fmt.Sprintf("Analyze the project and provide optimization suggestions targeting: %s.\n", target)
	if baselineOutput != "" {
		prompt += fmt.Sprintf("\nBaseline Benchmark Results:\n%s\n", baselineOutput)
	}
	prompt += "\nIdentify bottlenecks and rewrite code to improve performance/memory/scalability as requested."

	if contextSummary != "" {
		prompt += contextSummary
	}

	if apply {
		prompt += ApplyInstruction
	}

	fmt.Printf("💭 Prompt: Optimize for %s\n", target)
	fmt.Println("🔮 Generating optimizations...")

	client, err := createAIClient(provider, model)
	if err != nil {
		return fmt.Errorf("failed to create AI client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	uc := usecase.NewAIUseCase(client)
	// We already added context summary to prompt, so includeContext=false
	res, err := uc.Suggest(ctx, prompt, false, ".")
	if err != nil {
		return fmt.Errorf("optimization suggestion failed: %w", err)
	}

	fmt.Println("\n=== Optimization Plan ===")
	fmt.Println(res)
	fmt.Println("=========================")

	if apply {
		fmt.Println("✅ Applying optimizations...")
		count, _ := applyAIChanges(res)
		if count > 0 {
			fmt.Printf("Applied %d optimizations.\n", count)
			if benchmark {
				fmt.Println("📊 Running optimized benchmarks...")
				cmd := exec.Command("go", "test", "-bench=.", "./...")
				output, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Printf("⚠️ Benchmark failed: %v\n", err)
				} else {
					fmt.Printf("Optimized Results:\n%s\n", string(output))
					// Ideally we would compare baselineOutput and new output here
					fmt.Println("Compare the results above with baseline to verify improvement.")
				}
			}
		} else {
			fmt.Println("No changes applied.")
		}
	} else {
		fmt.Println("📋 Preview mode - use --apply to execute changes")
	}

	return nil
}
