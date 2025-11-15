package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/parser"
	"github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli/internal/security"
	"github.com/pmaojo/kthulu-go/backend/internal/secure"
)

var secureCmd = &cobra.Command{
	Use:   "secure",
	Short: "Scan project dependencies for known vulnerabilities",
	RunE: func(cmd *cobra.Command, args []string) error {
		patch, _ := cmd.Flags().GetBool("patch")
		vulns, err := secure.Scan(cmd.Context())
		if err != nil {
			return err
		}
		if len(vulns) == 0 {
			fmt.Println("no high severity vulnerabilities found")
			return nil
		}
		for _, v := range vulns {
			fmt.Printf("%s@%s\t%s\t%s\n", v.Module, v.Version, v.ID, v.Severity)
			if patch {
				if err := secure.Patch(v.Module, "latest"); err != nil {
					return err
				}
			}
		}
		if patch && os.Getenv("CI") != "" {
			_ = exec.Command("git", "add", "go.mod", "go.sum").Run()
			_ = exec.Command("git", "commit", "-m", "chore: update dependencies").Run()
		}
		return nil
	},
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze [path]",
	Short: "🔍 Advanced semantic analysis of Kthulu projects",
	Long: `Performs comprehensive semantic analysis including:
  • Architectural pattern detection (DDD, Repository, Service)
  • Dependency graph modeling and circular detection  
  • Security tag validation and insights
  • Performance optimization recommendations`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := "."
		if len(args) > 0 {
			targetPath = args[0]
		}

		// Usar la integración avanzada que incluye todos los features
		integration := parser.NewAdvancedIntegration()

		// Analizar proyecto con insights completos
		fmt.Printf("🔍 Analyzing project at: %s\n\n", targetPath)

		result, insights, depGraph, err := integration.AnalyzeProjectWithInsights(targetPath)
		if err != nil {
			return fmt.Errorf("analysis failed: %w", err)
		}

		// Mostrar resultados básicos
		fmt.Printf("📊 Analysis Results:\n")
		fmt.Printf("  Total tags found: %d\n", len(result.Tags))
		fmt.Printf("  Modules discovered: %d\n", len(result.Modules))
		fmt.Printf("  Dependencies: %d\n", len(result.Dependencies))

		if len(result.Modules) > 0 {
			fmt.Printf("\n📦 Modules:\n")
			count := 0
			for _, module := range result.Modules {
				if count >= 5 { // Limitar output
					fmt.Printf("  ... and %d more\n", len(result.Modules)-5)
					break
				}
				fmt.Printf("  • %s (%d files)\n", module.Name, len(module.Files))
				count++
			}
		}

		// Mostrar patrones arquitectónicos
		if insights != nil && len(insights.Patterns) > 0 {
			fmt.Printf("\n🏗️ Architectural Patterns:\n")
			count := 0
			for _, pattern := range insights.Patterns {
				if count >= 5 {
					break
				}
				fmt.Printf("  • %s: %d occurrences (confidence: %.1f%%)\n",
					pattern.Name, pattern.Occurrences, pattern.Confidence*100)
				count++
			}
		}

		// Mostrar recomendaciones
		recommendations := integration.GetRecommendations()
		if len(recommendations) > 0 {
			fmt.Printf("\n💡 Top Recommendations:\n")
			for i, rec := range recommendations {
				if i >= 5 { // Mostrar solo las 5 más importantes
					break
				}
				fmt.Printf("  • %s: %s\n", strings.Title(rec.Type), rec.Message)
			}
		}

		// Detectar dependencias circulares
		cycles := integration.GetCircularDependencies()
		if len(cycles) > 0 {
			fmt.Printf("\n⚠️ Circular Dependencies Detected:\n")
			for i, cycle := range cycles {
				if i >= 3 { // Limitar a 3 ciclos
					break
				}
				fmt.Printf("  %d. %s\n", i+1, strings.Join(cycle, " → "))
			}
		}

		// Mostrar métricas del proyecto
		metrics := integration.GetProjectMetrics()
		if metrics != nil {
			fmt.Printf("\n📈 Project Metrics:\n")
			fmt.Printf("  Total files: %d\n", metrics.TotalFiles)
			fmt.Printf("  Module count: %d\n", metrics.ModuleCount)
			fmt.Printf("  Complexity score: %.2f\n", metrics.ComplexityScore)
			if metrics.TaggedFiles > 0 {
				coverage := float64(metrics.TaggedFiles) / float64(metrics.TotalFiles) * 100
				fmt.Printf("  Tag coverage: %.1f%%\n", coverage)
			}
		}

		// Información del grafo de dependencias
		if depGraph != nil && len(depGraph.Nodes) > 0 {
			fmt.Printf("\n🕸️ Dependency Graph:\n")
			fmt.Printf("  Nodes: %d\n", len(depGraph.Nodes))
			fmt.Printf("  Edges: %d\n", len(depGraph.Edges))
			if len(depGraph.Cycles) > 0 {
				fmt.Printf("  Circular dependencies: %d\n", len(depGraph.Cycles))
			}
		}

		return nil
	},
}

var generateSecurityCmd = &cobra.Command{
	Use:   "generate [path]",
	Short: "🛡️ Generate enterprise security infrastructure from @kthulu:security tags",
	Long: `Automatically generates comprehensive security infrastructure including:
  • RBAC policies and roles from @kthulu:security tags
  • Authentication and authorization middleware
  • Security configuration files
  • Audit logging infrastructure
  • Security compliance reports and recommendations`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := "."
		if len(args) > 0 {
			targetPath = args[0]
		}

		// Create security generator
		generator := security.NewSecurityGenerator()

		fmt.Printf("🛡️ Generating enterprise security infrastructure at: %s\n\n", targetPath)

		// Generate security infrastructure
		result, err := generator.GenerateSecurityInfrastructure(targetPath)
		if err != nil {
			return fmt.Errorf("security generation failed: %w", err)
		}

		// Display results
		fmt.Printf("✅ Security Generation Complete!\n\n")
		fmt.Printf("📊 Generation Summary:\n")
		fmt.Printf("  Policies generated: %d\n", result.PoliciesGenerated)
		fmt.Printf("  Roles generated: %d\n", result.RolesGenerated)
		fmt.Printf("  Middleware generated: %t\n", result.MiddlewareGenerated)
		fmt.Printf("  Config generated: %t\n", result.ConfigGenerated)
		fmt.Printf("  Processing time: %v\n", result.ProcessingTime)

		if len(result.GeneratedFiles) > 0 {
			fmt.Printf("\n📁 Generated Files:\n")
			for _, file := range result.GeneratedFiles {
				fmt.Printf("  • %s\n", file)
			}
		}

		if result.SecurityReport != nil {
			report := result.SecurityReport
			fmt.Printf("\n🔒 Security Analysis:\n")
			fmt.Printf("  Security tags found: %d\n", report.TotalSecurityTags)
			fmt.Printf("  Coverage percentage: %.1f%%\n", report.CoveragePercentage)
			fmt.Printf("  Risk level: %s\n", report.RiskLevel)

			if len(report.ComplianceStatus) > 0 {
				fmt.Printf("\n📋 Compliance Status:\n")
				for standard, compliant := range report.ComplianceStatus {
					status := "❌ NON-COMPLIANT"
					if compliant {
						status = "✅ COMPLIANT"
					}
					fmt.Printf("  %s: %s\n", standard, status)
				}
			}
		}

		if len(result.Recommendations) > 0 {
			fmt.Printf("\n💡 Security Recommendations:\n")
			for i, rec := range result.Recommendations {
				if i >= 5 { // Limit to top 5 recommendations
					break
				}
				priority := "🔵"
				if rec.Priority == "HIGH" {
					priority = "🔴"
				} else if rec.Priority == "MEDIUM" {
					priority = "🟡"
				}
				fmt.Printf("  %s %s: %s\n", priority, rec.Title, rec.Description)
				fmt.Printf("    Action: %s\n", rec.Action)
			}
		}

		fmt.Printf("\n🚀 Next Steps:\n")
		fmt.Printf("  1. Review generated security policies\n")
		fmt.Printf("  2. Integrate RBAC middleware in your router\n")
		fmt.Printf("  3. Configure security settings in config/security.json\n")
		fmt.Printf("  4. Test authorization with different user roles\n")
		fmt.Printf("  5. Review compliance status and address gaps\n")

		return nil
	},
}

func init() {
	secureCmd.Flags().Bool("patch", false, "attempt to patch vulnerable modules")
	secureCmd.AddCommand(analyzeCmd)
	secureCmd.AddCommand(generateSecurityCmd)
}
