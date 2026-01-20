package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/pmaojo/kthulu-go/internal/blueprint"
	"github.com/pmaojo/kthulu-go/internal/usecase"
)

var aiScaffoldCmd = &cobra.Command{
	Use:   "scaffold [description]",
	Short: "🏗️  AI-powered project scaffolding",
	Long: `Generate a complete project configuration plan from a natural language description.
    
The AI will analyze your description and generate a 'kthulu-plan.yaml' file
that contains optimal architectural choices, features, and database selections.
You can then use this plan to create your project.

Examples:
  kthulu ai scaffold "A clone of Uber for dog walking with Stripe payments"
  kthulu ai scaffold "A simple blog with markdown support and admin panel"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		description := strings.Join(args, " ")
		provider, _ := cmd.Flags().GetString("provider")
		model, _ := cmd.Flags().GetString("model")
		output, _ := cmd.Flags().GetString("output")

		if output == "" {
			output = "kthulu-plan.yaml"
		}

		return runAIScaffold(description, provider, model, output)
	},
}

func init() {
	aiScaffoldCmd.Flags().String("output", "kthulu-plan.yaml", "Output file for the plan")
	aiCmd.AddCommand(aiScaffoldCmd)
}

func runAIScaffold(description, provider, model, outputFile string) error {
	// Fallback to Env if flags not set
	if provider == "openai" && os.Getenv("KTHULU_AI_PROVIDER") != "" && !isFlagSet(aiCmd, "provider") {
		provider = os.Getenv("KTHULU_AI_PROVIDER")
	}
	if model == "gpt-4" && os.Getenv("KTHULU_AI_MODEL") != "" && !isFlagSet(aiCmd, "model") {
		model = os.Getenv("KTHULU_AI_MODEL")
	}

	fmt.Printf("\n🏗️  Scaffolding project plan for: \"%s\"\n", description)
	fmt.Printf("🤖 Using AI Provider: %s (%s)\n", provider, model)

	// 1. Construct the prompt
	schemaPrompt := `
You are an expert software architect acting as a scaffolding engine for the Kthulu Framework.
Your goal is to convert a natural language description of a software project into a structured YAML plan.

Output ONLY the YAML content. Do not include markdown formatting like '''yaml or '''.

The YAML must strictly adhere to the following schema:

name: string (project name, kebab-case)
description: string
template: string (one of: microservice, monolith, api-gateway, fintech, ecommerce, saas)
features: [list of strings]
  # Available features: user, auth, organization, contact, product, invoice, payment, verifactu, audit, oauthsso, notification, calendar, inventory, realtime
database: string (postgres, sqlite, mysql)
frontend: string (react, templ, fyne, none)
auth: string (jwt, oauth, both)
modules:
  [module_name]:
    fields: [list of strings in "name:type" format, e.g., "title:string", "price:float"]

If the user mentions specific entities (like "Dog", "Car", "Post"), add them as modules with inferred fields.
If the user mentions capabilities (like "payments", "login"), add the corresponding features.
If the user mentions a specific database or frontend, use those values.

User Description:
` + description

	// 2. Create AI Client
	client, err := createAIClient(provider, model)
	if err != nil {
		return fmt.Errorf("failed to create AI client: %w", err)
	}
	defer client.Close()

	// 3. Request Generation
	fmt.Print("🔮 Analyzing requirements and designing architecture... ")
	ctx := context.Background()
	uc := usecase.NewAIUseCase(client)

	yamlContent, err := uc.Suggest(ctx, schemaPrompt, false, "")
	if err != nil {
		fmt.Println("❌")
		return fmt.Errorf("AI generation failed: %w", err)
	}
	fmt.Println("✅")

	// 4. Validate and Sanitize YAML
	cleanYaml := performBasicYamlCleanup(yamlContent)

	// Verify it parses
	var bp blueprint.ProjectBlueprint
	if err := yaml.Unmarshal([]byte(cleanYaml), &bp); err != nil {
		fmt.Printf("⚠️  Generated YAML might be invalid, but saving anyway for inspection.\nError: %v\n", err)
	}

	// 5. Save to file
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create directory for plan: %w", err)
	}
	if err := os.WriteFile(outputFile, []byte(cleanYaml), 0644); err != nil {
		return fmt.Errorf("failed to save plan: %w", err)
	}

	fmt.Println("\n✅ Plan generated successfully!")
	fmt.Printf("📄 Saved to: %s\n", outputFile)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	colorPrintPlan(cleanYaml)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("\n🚀 To create this project, run:\n")
	fmt.Printf("   kthulu create %s --from-plan %s\n", bp.Name, outputFile)

	return nil
}

func isFlagSet(cmd *cobra.Command, name string) bool {
	f := cmd.Flag(name)
	return f != nil && f.Changed
}

func colorPrintPlan(yamlStr string) {
	lines := strings.Split(yamlStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			fmt.Printf("\033[36m%s\033[0m:%s\n", parts[0], parts[1])
		} else {
			fmt.Println(line)
		}
	}
}

func performBasicYamlCleanup(content string) string {
	content = strings.TrimSpace(content)
	// Remove markdown code blocks (any language)
	re := regexp.MustCompile("(?s)^```(?:yaml)?\n?(.*?)\n?```$")
	if matches := re.FindStringSubmatch(content); len(matches) > 1 {
		content = matches[1]
	}
	return strings.TrimSpace(content)
}
