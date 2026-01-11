package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "✨ Generate code from natural language prompts",
	Long:  `Describe what you want to build, and Kthulu will translate it into CLI commands.`,
	Example: `  kthulu generate --prompt "I need a blog module"
  kthulu generate --prompt "create a user handler"
  kthulu generate --prompt "add an admin dashboard"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		prompt, _ := cmd.Flags().GetString("prompt")
		if prompt == "" {
			return fmt.Errorf("please provide a prompt using --prompt")
		}
		return runGenerate(prompt)
	},
}

func init() {
	generateCmd.Flags().StringP("prompt", "p", "", "Natural language description of what to build")
	rootCmd.AddCommand(generateCmd)
}

func runGenerate(prompt string) error {
	fmt.Printf("🧠 Analyzing prompt: \"%s\"\n", prompt)

	// Simulated NLP / Intent Recognition
	// In a real version, this would call an LLM API
	command, err := parseIntent(prompt)
	if err != nil {
		return fmt.Errorf("could not understand prompt: %v", err)
	}

	fmt.Printf("🤖 Translated to: %s\n", command)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Execute the command
	parts := strings.Fields(command)
	// assuming parts[0] is 'kthulu'
	if parts[0] != "kthulu" {
		return fmt.Errorf("unsafe command generation")
	}

	// We'll execute it as a subprocess to keep it simple and reusing existing CLIs
	// Skip 'kthulu' executable name
	cmdArgs := parts[1:]
	
	// If it's an 'add' command, we might want to auto-confirm if not present
	// Check if -y is present, if not append it for smoother AI experience? 
	// Or maybe ask user confirmation here?
	// Let's ask via standard input/confirmation simulated here or just run it.
	// Since the user explicitly asked to generate, let's assume -y for "add" commands if safe.
	// However, `kthulu add` has prompts. Let's pass what we have.
	
	finalCmd := exec.Command(os.Args[0], cmdArgs...)
	finalCmd.Stdout = os.Stdout
	finalCmd.Stderr = os.Stderr
	finalCmd.Stdin = os.Stdin

	return finalCmd.Run()
}

func parseIntent(prompt string) (string, error) {
	prompt = strings.ToLower(prompt)

	if strings.Contains(prompt, "admin") && (strings.Contains(prompt, "dashboard") || strings.Contains(prompt, "panel")) {
		return "kthulu add admin", nil
	}

	if strings.Contains(prompt, "module") {
		// "add a blog module" -> extract "blog"
		words := strings.Fields(prompt)
		var name string
		for i, w := range words {
			if w == "module" && i > 0 {
				// "blog module"
				if words[i-1] != "a" && words[i-1] != "add" && words[i-1] != "create" {
					name = words[i-1]
				}
			}
			if w == "module" && i < len(words)-1 {
				// "module blog"
				name = words[i+1]
			}
		}
		
		if name != "" {
			cmd := fmt.Sprintf("kthulu add module %s", name)
			if strings.Contains(prompt, "admin") {
				cmd += " --admin"
			}
			return cmd, nil
		}
	}

	if strings.Contains(prompt, "handler") {
		// "user handler"
		words := strings.Fields(prompt)
		for i, w := range words {
			if w == "handler" && i > 0 {
				name := words[i-1]
				// Capitalize for component name
				name = strings.Title(name) + "Handler" // simplistic
				return fmt.Sprintf("kthulu add component handler %s", name), nil
			}
		}
	}

	return "", fmt.Errorf("I'm not smart enough yet to understand that request. Try 'add a <name> module'")
}
