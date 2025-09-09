package models

import (
	"fmt"
	"runtime"
)

// Environment represents the user's local development environment and tool availability.
type Environment struct {
	Tools       map[string]ToolStatus `json:"tools"`        // Available tools and their status
	Platform    string               `json:"platform"`     // Operating system (linux, darwin, windows)
	HasInternet bool                 `json:"has_internet"` // Internet connectivity status
	GitConfig   GitConfig            `json:"git_config"`   // Git configuration details
	WorkingDir  string               `json:"working_dir"`  // Current working directory
}

// ToolStatus represents the status of a development tool
type ToolStatus struct {
	Available   bool   `json:"available"`    // Whether tool is found in PATH
	Version     string `json:"version"`      // Tool version (if available)
	InstallHint string `json:"install_hint"` // Installation instructions
}

// GitConfig represents git configuration information
type GitConfig struct {
	Available bool   `json:"available"`  // Git command available
	UserName  string `json:"user_name"`  // Configured user.name
	UserEmail string `json:"user_email"` // Configured user.email
	Version   string `json:"version"`    // Git version
}

// SupportedPlatforms contains the list of supported operating systems
var SupportedPlatforms = []string{"linux", "darwin", "windows"}

// RequiredTools contains the list of tools we check for
var RequiredTools = map[string]string{
	"git": "https://git-scm.com/downloads",
}

// OptionalTools contains AI assistant tools and their install hints
var OptionalTools = map[string]string{
	"claude": "Install from: https://docs.anthropic.com/en/docs/claude-code/setup",
	"gemini": "Install from: https://github.com/google-gemini/gemini-cli", 
}

// NewEnvironment creates a new Environment instance
func NewEnvironment(workingDir string) *Environment {
	return &Environment{
		Tools:       make(map[string]ToolStatus),
		Platform:    runtime.GOOS,
		HasInternet: false, // Will be determined by connectivity check
		GitConfig:   GitConfig{},
		WorkingDir:  workingDir,
	}
}

// Validate checks if the environment configuration is valid
func (e *Environment) Validate() error {
	// Validate platform
	if err := e.validatePlatform(); err != nil {
		return err
	}

	// Validate working directory
	if err := e.validateWorkingDir(); err != nil {
		return err
	}

	// Validate git config if git is available
	if e.GitConfig.Available {
		if err := e.validateGitConfig(); err != nil {
			return err
		}
	}

	return nil
}

// validatePlatform validates the operating system platform
func (e *Environment) validatePlatform() error {
	if e.Platform == "" {
		return &EnvironmentError{
			Type:    ToolNotFound,
			Message: "platform cannot be empty",
		}
	}

	// Check if platform is supported
	for _, supported := range SupportedPlatforms {
		if e.Platform == supported {
			return nil
		}
	}

	return &EnvironmentError{
		Type:    ToolNotFound,
		Message: fmt.Sprintf("unsupported platform '%s', supported platforms: %v", e.Platform, SupportedPlatforms),
	}
}

// validateWorkingDir validates the working directory
func (e *Environment) validateWorkingDir() error {
	if e.WorkingDir == "" {
		return &EnvironmentError{
			Type:    ToolNotFound,
			Message: "working directory cannot be empty",
		}
	}

	return nil
}

// validateGitConfig validates git configuration
func (e *Environment) validateGitConfig() error {
	if e.GitConfig.Version == "" {
		return &EnvironmentError{
			Type:    ToolVersionUnsupported,
			Tool:    "git",
			Message: "git version cannot be empty when git is available",
		}
	}

	return nil
}

// SetToolStatus sets the status for a specific tool
func (e *Environment) SetToolStatus(tool string, available bool, version, installHint string) {
	e.Tools[tool] = ToolStatus{
		Available:   available,
		Version:     version,
		InstallHint: installHint,
	}
}

// GetToolStatus returns the status for a specific tool
func (e *Environment) GetToolStatus(tool string) (ToolStatus, bool) {
	status, exists := e.Tools[tool]
	return status, exists
}

// IsToolAvailable returns whether a specific tool is available
func (e *Environment) IsToolAvailable(tool string) bool {
	status, exists := e.Tools[tool]
	return exists && status.Available
}

// SetGitConfig sets the git configuration
func (e *Environment) SetGitConfig(available bool, version, userName, userEmail string) {
	e.GitConfig = GitConfig{
		Available: available,
		Version:   version,
		UserName:  userName,
		UserEmail: userEmail,
	}
}

// IsGitConfigured returns whether git is properly configured
func (e *Environment) IsGitConfigured() bool {
	return e.GitConfig.Available && e.GitConfig.UserName != "" && e.GitConfig.UserEmail != ""
}

// GetMissingRequiredTools returns a list of required tools that are not available
func (e *Environment) GetMissingRequiredTools() []string {
	var missing []string
	for tool := range RequiredTools {
		if !e.IsToolAvailable(tool) {
			missing = append(missing, tool)
		}
	}
	return missing
}

// GetMissingOptionalTools returns a list of optional tools that are not available
func (e *Environment) GetMissingOptionalTools() []string {
	var missing []string
	for tool := range OptionalTools {
		if !e.IsToolAvailable(tool) {
			missing = append(missing, tool)
		}
	}
	return missing
}

// GetPlatformDisplayName returns a user-friendly platform name
func (e *Environment) GetPlatformDisplayName() string {
	switch e.Platform {
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	default:
		return e.Platform
	}
}

// CanInitializeProjects returns whether the environment is ready for project initialization
func (e *Environment) CanInitializeProjects() bool {
	// At minimum, we need internet connectivity
	// Git is optional but recommended
	return e.HasInternet
}

// GetReadinessStatus returns a summary of environment readiness
func (e *Environment) GetReadinessStatus() string {
	if !e.HasInternet {
		return "Internet connection required"
	}

	missingRequired := e.GetMissingRequiredTools()
	if len(missingRequired) > 0 {
		return fmt.Sprintf("Missing required tools: %v", missingRequired)
	}

	missingOptional := e.GetMissingOptionalTools()
	if len(missingOptional) > 0 {
		return fmt.Sprintf("Ready (consider installing: %v)", missingOptional)
	}

	return "Ready"
}

// GetInstallHint returns installation instructions for a specific tool
func (e *Environment) GetInstallHint(tool string) string {
	// Check in tool status first
	if status, exists := e.Tools[tool]; exists {
		return status.InstallHint
	}

	// Check in required tools
	if hint, exists := RequiredTools[tool]; exists {
		return hint
	}

	// Check in optional tools
	if hint, exists := OptionalTools[tool]; exists {
		return hint
	}

	return fmt.Sprintf("Installation instructions not available for tool: %s", tool)
}