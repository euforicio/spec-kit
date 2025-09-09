package services

import (
	"fmt"
	"path/filepath"

	"github.com/euforicio/spec-kit/internal/codex"
	"github.com/euforicio/spec-kit/internal/models"
)

// ProjectService handles project initialization workflows
type ProjectService struct {
	environment *EnvironmentService
	template    *TemplateService
	filesystem  *FilesystemService
}

// ProjectInitOptions contains options for project initialization
type ProjectInitOptions struct {
	Name              string
	Path              string
	AIAssistant       string
	IsHere            bool
	NoGit             bool
	IgnoreAgentTools  bool
	Force             bool
}

// ProjectInitResult contains the result of project initialization
type ProjectInitResult struct {
	Project     *models.Project
	Template    *models.Template
	Environment *models.Environment
	GitRepo     bool
	Warnings    []string
}

// NewProjectService creates a new project service instance
func NewProjectService(environment *EnvironmentService, template *TemplateService, filesystem *FilesystemService) *ProjectService {
	return &ProjectService{
		environment: environment,
		template:    template,
		filesystem:  filesystem,
	}
}

// InitializeProject initializes a new project with the specified options
func (p *ProjectService) InitializeProject(options ProjectInitOptions) (*ProjectInitResult, error) {
	result := &ProjectInitResult{
		Warnings: make([]string, 0),
	}

	// Step 1: Detect environment
	env, err := p.environment.DetectEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to detect environment: %w", err)
	}
	result.Environment = env

	// Step 2: Validate prerequisites
	if err := p.validatePrerequisites(env, options); err != nil {
		return nil, err
	}

	// Step 3: Create and validate project
	project, err := p.createProject(options)
	if err != nil {
		return nil, err
	}
	result.Project = project

	// Step 4: Validate project path
	if err := p.validateProjectPath(project); err != nil {
		return nil, err
	}

	// Step 5: Download and extract template
	template, err := p.downloadTemplate(project)
	if err != nil {
		return nil, err
	}
	result.Template = template

	// Step 6: Initialize git repository (if requested and available)
	gitInitialized, warning := p.initializeGit(project, env, options.NoGit)
	result.GitRepo = gitInitialized
	if warning != "" {
		result.Warnings = append(result.Warnings, warning)
	}

	// Step 7: Setup AI assistant specific files
	if err := p.setupAIAssistantFiles(project); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to setup AI assistant files: %v", err))
	}

	// Step 8: Validate final result
	if err := p.validateResult(project); err != nil {
		return nil, err
	}

	return result, nil
}

// validatePrerequisites checks if all prerequisites are met for project initialization
func (p *ProjectService) validatePrerequisites(env *models.Environment, options ProjectInitOptions) error {
	// Check internet connectivity
	if !env.HasInternet {
		return &models.EnvironmentError{
			Type:    models.InternetNotAvailable,
			Message: "internet connection is required to download templates",
			Hint:    "check your network connection and try again",
		}
	}

	// Check AI assistant tools (unless ignored)
	if !options.IgnoreAgentTools {
		if err := p.validateAITools(env, options.AIAssistant); err != nil {
			return err
		}
	}

	return nil
}

// validateAITools validates that the required AI assistant tools are available
func (p *ProjectService) validateAITools(env *models.Environment, aiAssistant string) error {
	switch aiAssistant {
	case "claude":
		if !env.IsToolAvailable("claude") {
			return &models.EnvironmentError{
				Type: models.ToolNotFound,
				Tool: "claude",
				Message: "Claude CLI is required for Claude Code projects",
				Hint: "Install from: https://docs.anthropic.com/en/docs/claude-code/setup",
			}
		}
	case "gemini":
		if !env.IsToolAvailable("gemini") {
			return &models.EnvironmentError{
				Type: models.ToolNotFound,
				Tool: "gemini",
				Message: "Gemini CLI is required for Gemini projects",
				Hint: "Install from: https://github.com/google-gemini/gemini-cli",
			}
		}
	case "copilot":
		// GitHub Copilot doesn't require a CLI tool - it's available in IDEs
		// No validation needed
	}

	return nil
}

// createProject creates and validates a new project instance
func (p *ProjectService) createProject(options ProjectInitOptions) (*models.Project, error) {
	// Determine project path
	projectPath := options.Path
	if options.IsHere {
		workingDir, err := p.filesystem.GetWorkingDirectory()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		projectPath = workingDir
		options.Name = filepath.Base(workingDir)
	} else {
		// Make path absolute
		if !filepath.IsAbs(projectPath) {
			workingDir, err := p.filesystem.GetWorkingDirectory()
			if err != nil {
				return nil, fmt.Errorf("failed to get working directory: %w", err)
			}
			projectPath = filepath.Join(workingDir, options.Path)
		}
	}

	// Create project model
	project, err := models.NewProject(options.Name, projectPath, options.AIAssistant, options.IsHere)
	if err != nil {
		return nil, err
	}

	return project, nil
}

// validateProjectPath validates the project path and directory state
func (p *ProjectService) validateProjectPath(project *models.Project) error {
	return p.environment.ValidateProjectPath(project.Path, project.IsHere)
}

// downloadTemplate downloads and extracts the template for the project
func (p *ProjectService) downloadTemplate(project *models.Project) (*models.Template, error) {
	template, err := p.template.DownloadAndExtract(project.AIAssistant, project.Path, project.IsHere)
	if err != nil {
		// Clean up on failure (unless using --here)
		if !project.IsHere {
			p.filesystem.RemoveDirectory(project.Path)
		}
		return nil, err
	}

	// Validate extracted template
	if err := p.template.ValidateExtractedTemplate(project.Path); err != nil {
		// Clean up on failure (unless using --here)
		if !project.IsHere {
			p.filesystem.RemoveDirectory(project.Path)
		}
		return nil, err
	}

	return template, nil
}

// initializeGit initializes a git repository if conditions are met
func (p *ProjectService) initializeGit(project *models.Project, env *models.Environment, noGit bool) (bool, string) {
	// Skip if --no-git flag is set
	if noGit {
		return false, ""
	}

	// Check if git is available
	if !env.IsToolAvailable("git") {
		return false, "Git not found - skipping repository initialization"
	}

	// Check if already in a git repository
	if p.environment.IsInGitRepository(project.Path) {
		project.HasGit = true
		return true, "Existing git repository detected"
	}

	// Initialize new git repository
	if err := p.environment.InitializeGitRepository(project.Path); err != nil {
		return false, fmt.Sprintf("Failed to initialize git repository: %v", err)
	}

	project.HasGit = true
	return true, ""
}

// validateResult performs final validation on the initialized project
func (p *ProjectService) validateResult(project *models.Project) error {
	// Check that project directory exists and is accessible
	exists, err := p.filesystem.DirectoryExists(project.Path)
	if err != nil {
		return fmt.Errorf("failed to validate project directory: %w", err)
	}

	if !exists {
		return &models.ProjectError{
			Type:    models.ProjectPathInvalid,
			Path:    project.Path,
			Message: "project directory does not exist after initialization",
		}
	}

	// Check that directory is not empty
	isEmpty, err := p.filesystem.IsDirectoryEmpty(project.Path)
	if err != nil {
		return fmt.Errorf("failed to check project directory contents: %w", err)
	}

	if isEmpty {
		return &models.ProjectError{
			Type:    models.ProjectPathInvalid,
			Path:    project.Path,
			Message: "project directory is empty after initialization",
		}
	}

	return nil
}

// GetNextSteps returns recommended next steps for the user
func (p *ProjectService) GetNextSteps(result *ProjectInitResult) []string {
	var steps []string

	// Step 1: Navigate to project (if not using --here)
	if !result.Project.IsHere {
		steps = append(steps, fmt.Sprintf("cd %s", result.Project.Name))
	}

	// Step 2: AI assistant specific instructions
	switch result.Project.AIAssistant {
	case "claude":
		steps = append(steps,
			"Open in Visual Studio Code and start using / commands with Claude Code",
			"Type / in any file to see available commands",
			"Use /specify to create specifications",
			"Use /plan to create implementation plans",
			"Use /tasks to generate tasks",
		)
	case "gemini":
		steps = append(steps,
			"Use / commands with Gemini CLI",
			"Run gemini /specify to create specifications",
			"Run gemini /plan to create implementation plans",
			"See GEMINI.md for all available commands",
		)
	case "copilot":
		steps = append(steps,
			"Open in Visual Studio Code and use /specify, /plan, /tasks commands with GitHub Copilot",
		)
	case "codex":
		steps = append(steps,
			"Open project with OpenAI Codex",
			"Use / commands within Codex (e.g., /specify, /plan, /tasks)",
			"See AGENTS.md for all available commands and patterns",
			"Run 'specify feature create \"your feature\"' when ready to start development",
		)
	}

	// Step 3: Constitution
	steps = append(steps, "Update CONSTITUTION.md with your project's non-negotiable principles")

	return steps
}

// CheckProjectHealth performs health checks on an existing project
func (p *ProjectService) CheckProjectHealth(projectPath string) error {
	// Check if directory exists
	exists, err := p.filesystem.DirectoryExists(projectPath)
	if err != nil {
		return fmt.Errorf("failed to check project directory: %w", err)
	}

	if !exists {
		return &models.ProjectError{
			Type:    models.ProjectPathInvalid,
			Path:    projectPath,
			Message: "project directory does not exist",
		}
	}

	// Check if directory is readable and writable
	if err := p.filesystem.IsWritable(projectPath); err != nil {
		return err
	}

	// Additional health checks could be added here:
	// - Check for required files
	// - Validate project structure
	// - Check git repository health
	// - Validate AI assistant configuration

	return nil
}

// EstimateProjectSize estimates the total size a project will take after initialization
func (p *ProjectService) EstimateProjectSize(aiAssistant string) (int64, error) {
	template, err := p.template.GetTemplateInfo(aiAssistant)
	if err != nil {
		return 0, err
	}

	// Template size is a good estimate (ZIP is compressed, but git repo adds overhead)
	return template.Size, nil
}

// ListAvailableTemplates returns information about all available templates
func (p *ProjectService) ListAvailableTemplates() (map[string]*models.Template, error) {
	return p.template.GetAvailableTemplates()
}

// setupAIAssistantFiles creates AI assistant specific files during project initialization
func (p *ProjectService) setupAIAssistantFiles(project *models.Project) error {
	switch project.AIAssistant {
	case "codex":
		return p.setupCodexFiles(project)
	default:
		// Other AI assistants handled by template files
		return nil
	}
}

// setupCodexFiles creates AGENTS.md and .codex/commands/ for Codex integration
func (p *ProjectService) setupCodexFiles(project *models.Project) error {
	// Create codex service
	codexService := codex.NewService()
	
	// Generate basic AGENTS.md content (no plan content yet)
	agentsContent, err := codexService.GenerateAGENTS("", nil)
	if err != nil {
		return fmt.Errorf("failed to generate AGENTS.md content: %w", err)
	}
	
	// Write AGENTS.md to project directory
	if err := codexService.WriteAGENTS(agentsContent, project.Path); err != nil {
		return fmt.Errorf("failed to write AGENTS.md: %w", err)
	}
	
	// Create command files
	if err := codexService.WriteCommandFiles(false, project.Path); err != nil {
		return fmt.Errorf("failed to write command files: %w", err)
	}
	
	return nil
}