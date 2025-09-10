package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/euforicio/spec-kit/internal/models"
	"github.com/euforicio/spec-kit/internal/services"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check that all required tools are installed",
	Long: `Check that all required tools are installed and the environment is ready for project initialization.

This command checks:
- Internet connectivity (required for downloading templates)
- Git installation and configuration (optional but recommended)
- AI assistant tools (Claude Code, Gemini CLI)
- System information and recommendations`,
	RunE: runCheck,
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Show banner
	showBanner()

    fmt.Println("🔍 Checking Specify requirements...")

	// Initialize services
	filesystem := services.NewFilesystemService()
	github := services.NewGitHubService()
	environment := services.NewEnvironmentService(filesystem)

	// Detect environment
	env, err := environment.DetectEnvironment()
	if err != nil {
		return fmt.Errorf("failed to detect environment: %w", err)
	}

	// Show system information
	showSystemInfo(env)

	// Check internet connectivity
	checkInternetConnectivity(github)

	// Check required tools
	checkRequiredTools(env)

	// Check optional tools
	checkOptionalTools(env)

	// Show recommendations
	showRecommendations(environment, env)

	// Show final status
	showFinalStatus(env)

	return nil
}

func showSystemInfo(env *models.Environment) {
	fmt.Printf("💻 System Information\n")
	fmt.Printf("   Platform: %s\n", env.GetPlatformDisplayName())
	fmt.Printf("   Working Directory: %s\n", env.WorkingDir)
	fmt.Println()
}

func checkInternetConnectivity(github *services.GitHubService) {
	fmt.Printf("🌐 Internet Connectivity\n")
	
	err := github.CheckConnectivity()
	if err != nil {
		fmt.Printf("   ❌ No internet connection\n")
		fmt.Printf("      %s\n", err.Error())
	} else {
		fmt.Printf("   ✅ Internet connection available\n")
		
		// Get rate limit info
		limit, remaining, resetTime, err := github.GetRateLimitInfo()
		if err == nil {
			fmt.Printf("      GitHub API: %d/%d requests remaining (resets at %s)\n", 
				remaining, limit, resetTime.Format("15:04:05"))
		}
	}
	fmt.Println()
}

func checkRequiredTools(env *models.Environment) {
	fmt.Printf("🔧 Required Tools\n")
	
	hasAllRequired := true
	for tool := range models.RequiredTools {
		status, exists := env.GetToolStatus(tool)
		if !exists || !status.Available {
			hasAllRequired = false
			fmt.Printf("   ❌ %s not found\n", tool)
			if exists {
				fmt.Printf("      Install: %s\n", status.InstallHint)
			}
		} else {
			fmt.Printf("   ✅ %s", tool)
			if status.Version != "" && status.Version != "unknown" {
				fmt.Printf(" (%s)", status.Version)
			}
			fmt.Println()
			
			// Special handling for git
			if tool == "git" && env.GitConfig.Available {
				if env.IsGitConfigured() {
					fmt.Printf("      Configured: %s <%s>\n", env.GitConfig.UserName, env.GitConfig.UserEmail)
				} else {
					fmt.Printf("      ⚠️  Not configured (missing user.name or user.email)\n")
					fmt.Printf("      Run: git config --global user.name \"Your Name\"\n")
					fmt.Printf("      Run: git config --global user.email \"your.email@example.com\"\n")
				}
			}
		}
	}
	
	if hasAllRequired {
		fmt.Printf("   📦 All required tools are available\n")
	}
	fmt.Println()
}

func checkOptionalTools(env *models.Environment) {
	fmt.Printf("🤖 AI Assistant Tools\n")
	
	availableAI := []string{}
	for tool := range models.OptionalTools {
		status, exists := env.GetToolStatus(tool)
		if exists && status.Available {
			availableAI = append(availableAI, tool)
			fmt.Printf("   ✅ %s", getAIAssistantDisplayName(tool))
			if status.Version != "" && status.Version != "unknown" {
				fmt.Printf(" (%s)", status.Version)
			}
			fmt.Println()
		} else {
			fmt.Printf("   ❌ %s not found\n", getAIAssistantDisplayName(tool))
			if exists {
				fmt.Printf("      Install: %s\n", status.InstallHint)
			}
		}
	}
	
	if len(availableAI) == 0 {
		fmt.Printf("   ⚠️  No AI assistant tools found\n")
		fmt.Printf("   Consider installing one for the best experience\n")
	} else {
		fmt.Printf("   🎉 %d AI assistant(s) available: %s\n", 
			len(availableAI), strings.Join(availableAI, ", "))
	}
	fmt.Println()
}

func showRecommendations(environment *services.EnvironmentService, env *models.Environment) {
	recommendations := environment.GetRecommendations(env)
	
	if len(recommendations) > 0 {
		fmt.Printf("💡 Recommendations\n")
		for _, rec := range recommendations {
			fmt.Printf("   • %s\n", rec)
		}
		fmt.Println()
	}
}

func showFinalStatus(env *models.Environment) {
	status := env.GetReadinessStatus()
	
	if env.CanInitializeProjects() {
		fmt.Printf("🎉 Specify CLI is ready to use!\n")
		if status != "Ready" {
			fmt.Printf("   Status: %s\n", status)
		}
	} else {
		fmt.Printf("⚠️  Specify CLI is not ready\n")
		fmt.Printf("   Status: %s\n", status)
		fmt.Printf("   Fix the issues above and run 'specify check' again\n")
	}
	fmt.Println()
}
