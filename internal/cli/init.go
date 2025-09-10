package cli

import (
    "fmt"
    "path/filepath"
    "strings"

	"github.com/spf13/cobra"

	"github.com/euforicio/spec-kit/internal/models"
	"github.com/euforicio/spec-kit/internal/services"
	"github.com/euforicio/spec-kit/internal/ui"
)

var initCmd = &cobra.Command{
	Use:   "init [PROJECT_NAME]",
	Short: "Initialize a new Specify project",
	Long: `Initialize a new Specify project from the latest template.

This command will:
1. Check that required tools are installed (git is optional)
2. Let you choose your AI assistant (Claude Code, Gemini CLI, or GitHub Copilot)
3. Download the appropriate template from GitHub
4. Extract the template to a new project directory or current directory
5. Initialize a fresh git repository (if not --no-git and no existing repo)
6. Optionally set up AI assistant commands

Examples:
  specify init my-project
  specify init my-project --ai claude
  specify init my-project --ai gemini
  specify init my-project --ai copilot --no-git
  specify init --ignore-agent-tools my-project
  specify init --here --ai claude
  specify init --here`,
	RunE: runInit,
}

var (
	aiAssistant      string
	ignoreAgentTools bool
	noGit            bool
	here             bool
	force            bool
)

func init() {
    initCmd.Flags().StringVar(&aiAssistant, "ai", "", "AI assistant to use: claude, gemini, copilot, or codex")
	initCmd.Flags().BoolVar(&ignoreAgentTools, "ignore-agent-tools", false, "Skip checks for AI agent tools like Claude Code")
	initCmd.Flags().BoolVar(&noGit, "no-git", false, "Skip git repository initialization")
	initCmd.Flags().BoolVar(&here, "here", false, "Initialize project in the current directory instead of creating a new one")
	initCmd.Flags().BoolVar(&force, "force", false, "Force initialization even if directory is not empty")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Show banner
	showBanner()

	// Validate arguments
	projectName := ""
	if len(args) > 0 {
		projectName = args[0]
	}

	if here && projectName != "" {
		return fmt.Errorf("cannot specify both project name and --here flag")
	}

	if !here && projectName == "" {
		return fmt.Errorf("must specify either a project name or use --here flag")
	}

	// Initialize services
	filesystem := services.NewFilesystemService()
	github := services.NewGitHubService()
	template := services.NewTemplateService(github, filesystem)
	environment := services.NewEnvironmentService(filesystem)
	project := services.NewProjectService(environment, template, filesystem)

	// Create project options
	options := services.ProjectInitOptions{
		Name:             projectName,
		AIAssistant:      aiAssistant,
		IsHere:           here,
		NoGit:            noGit,
		IgnoreAgentTools: ignoreAgentTools,
		Force:            force,
	}

	// Determine project path
	if here {
		workingDir, err := filesystem.GetWorkingDirectory()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		options.Path = workingDir
		options.Name = filepath.Base(workingDir)
	} else {
		options.Path = projectName
	}

	// Check if directory exists and handle accordingly
	if !here {
		exists, err := filesystem.DirectoryExists(options.Path)
		if err != nil {
			return fmt.Errorf("failed to check directory existence: %w", err)
		}
		if exists {
			return fmt.Errorf("directory '%s' already exists", projectName)
		}
	} else {
		// Check if current directory is empty
		isEmpty, err := filesystem.IsDirectoryEmpty(options.Path)
		if err != nil {
			return fmt.Errorf("failed to check directory contents: %w", err)
		}
		if !isEmpty && !force {
			fmt.Printf("Warning: Current directory is not empty.\n")
			fmt.Printf("Template files will be merged with existing content and may overwrite existing files.\n")
			
			if !ui.Confirm("Do you want to continue?") {
				fmt.Println("Operation cancelled")
				return nil
			}
		}
	}

	// AI assistant selection
	if options.AIAssistant == "" {
		var err error
		options.AIAssistant, err = selectAIAssistant()
		if err != nil {
			return err
		}
	} else {
        // Validate provided AI assistant
        if !models.IsValidAgent(options.AIAssistant) {
            return fmt.Errorf("invalid AI assistant '%s', must be one of: %s", 
                options.AIAssistant, strings.Join(models.ListAgents(), ", "))
        }
	}

	// Show project information
	fmt.Printf("\nInitializing Specify Project\n")
	if here {
		fmt.Printf("Location: Current directory (%s)\n", filepath.Base(options.Path))
	} else {
		fmt.Printf("Project: %s\n", projectName)
	}
	fmt.Printf("AI Assistant: %s\n", getAIAssistantDisplayName(options.AIAssistant))
	fmt.Println()

	// Initialize project with progress tracking
	tracker := ui.NewProgressTracker("Initialize Project")
	tracker.Start()

	tracker.AddStep("precheck", "Check environment")
	tracker.AddStep("download", "Download template")
	tracker.AddStep("extract", "Extract template")
	tracker.AddStep("git", "Initialize git repository")
	tracker.AddStep("finalize", "Finalize setup")

	tracker.StartStep("precheck")
	result, err := project.InitializeProject(options)
	if err != nil {
		tracker.FailStep("precheck", err.Error())
		tracker.Stop()
		return err
	}
	tracker.CompleteStep("precheck")

	tracker.StartStep("download")
	tracker.CompleteStep("download")

	tracker.StartStep("extract")
	tracker.CompleteStep("extract")

	tracker.StartStep("git")
	if result.GitRepo {
		tracker.CompleteStep("git")
	} else {
		tracker.SkipStep("git", "skipped")
	}

	tracker.StartStep("finalize")
	tracker.CompleteStep("finalize")

	tracker.Stop()

	// Show success message
	fmt.Printf("\n✅ Project initialized successfully!\n\n")

	// Show warnings if any
	if len(result.Warnings) > 0 {
		fmt.Println("⚠️  Warnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("   %s\n", warning)
		}
		fmt.Println()
	}

	// Show next steps
	nextSteps := project.GetNextSteps(result)
	if len(nextSteps) > 0 {
		fmt.Println("📋 Next steps:")
		for i, step := range nextSteps {
			fmt.Printf("   %d. %s\n", i+1, step)
		}
		fmt.Println()
	}

	return nil
}

func selectAIAssistant() (string, error) {
	fmt.Println("Select your AI assistant:")
	fmt.Println("1. Claude Code")
	fmt.Println("2. Gemini CLI") 
	fmt.Println("3. GitHub Copilot")
	fmt.Println("4. OpenAI Codex")
	fmt.Println()

	choice := ui.PromptSelect("Choose (1-4): ", []string{"1", "2", "3", "4"})

	switch choice {
	case "1":
		return "claude", nil
	case "2":
		return "gemini", nil
	case "3":
		return "copilot", nil
	case "4":
		return "codex", nil
	default:
		return "", fmt.Errorf("invalid choice: %s", choice)
	}
}

func getAIAssistantDisplayName(aiAssistant string) string {
	return models.GetAIAssistantDisplayName(aiAssistant)
}
