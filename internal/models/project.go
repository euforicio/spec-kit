package models

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Project represents a spec-driven development project being initialized.
type Project struct {
	Name        string    `json:"name"`         // Project name (directory name)
	Path        string    `json:"path"`         // Absolute path to project directory
	AIAssistant string    `json:"ai_assistant"` // Selected AI assistant (claude, gemini, copilot)
	IsHere      bool      `json:"is_here"`      // Whether initializing in current directory
	HasGit      bool      `json:"has_git"`      // Whether project has/should have git repository
	CreatedAt   time.Time `json:"created_at"`   // When project was initialized
}

// ProjectState represents the current state of project initialization
type ProjectState int

const (
	ProjectStateInitial ProjectState = iota
	ProjectStateValidated
	ProjectStateTemplateSelected
	ProjectStateCreated
	ProjectStateGitInitialized
)

// String returns a string representation of the project state
func (ps ProjectState) String() string {
	switch ps {
	case ProjectStateInitial:
		return "initial"
	case ProjectStateValidated:
		return "validated"
	case ProjectStateTemplateSelected:
		return "template_selected"
	case ProjectStateCreated:
		return "created"
	case ProjectStateGitInitialized:
		return "git_initialized"
	default:
		return "unknown"
	}
}

// ValidAIAssistants contains the list of supported AI assistants
var ValidAIAssistants = []string{"claude", "gemini", "copilot", "codex"}

// NewProject creates a new Project instance with validation
func NewProject(name, path, aiAssistant string, isHere bool) (*Project, error) {
	project := &Project{
		Name:        name,
		Path:        path,
		AIAssistant: aiAssistant,
		IsHere:      isHere,
		HasGit:      false, // Will be determined later
		CreatedAt:   time.Now(),
	}

	if err := project.Validate(); err != nil {
		return nil, err
	}

	return project, nil
}

// Validate checks if the project configuration is valid
func (p *Project) Validate() error {
	// Validate project name (unless using --here)
	if !p.IsHere {
		if err := p.validateName(); err != nil {
			return err
		}
	}

	// Validate path
	if err := p.validatePath(); err != nil {
		return err
	}

	// Validate AI assistant
	if err := p.validateAIAssistant(); err != nil {
		return err
	}

	return nil
}

// validateName validates the project name for filesystem compatibility
func (p *Project) validateName() error {
	if p.Name == "" {
		return &ProjectError{
			Type:    ProjectNameInvalid,
			Path:    p.Path,
			Message: "project name cannot be empty",
		}
	}

	// Check for invalid characters in project name
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(p.Name, char) {
			return &ProjectError{
				Type:    ProjectNameInvalid,
				Path:    p.Path,
				Message: fmt.Sprintf("project name contains invalid character: %s", char),
			}
		}
	}

	// Check for reserved names (platform-specific)
	reservedNames := []string{"con", "prn", "aux", "nul", "com1", "com2", "com3", 
		"com4", "com5", "com6", "com7", "com8", "com9", "lpt1", "lpt2", "lpt3", 
		"lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9"}
	
	for _, reserved := range reservedNames {
		if strings.EqualFold(p.Name, reserved) {
			return &ProjectError{
				Type:    ProjectNameInvalid,
				Path:    p.Path,
				Message: fmt.Sprintf("project name '%s' is reserved", p.Name),
			}
		}
	}

	return nil
}

// validatePath validates the project path
func (p *Project) validatePath() error {
	if p.Path == "" {
		return &ProjectError{
			Type:    ProjectPathInvalid,
			Path:    p.Path,
			Message: "project path cannot be empty",
		}
	}

	// Ensure path is absolute
	if !filepath.IsAbs(p.Path) {
		return &ProjectError{
			Type:    ProjectPathInvalid,
			Path:    p.Path,
			Message: "project path must be absolute",
		}
	}

	return nil
}

// validateAIAssistant validates the AI assistant selection
func (p *Project) validateAIAssistant() error {
	if p.AIAssistant == "" {
		return &ProjectError{
			Type:    ProjectNameInvalid, // Reusing error type for validation
			Path:    p.Path,
			Message: "AI assistant must be specified",
		}
	}

	// Check if AI assistant is in the valid list
	for _, valid := range ValidAIAssistants {
		if p.AIAssistant == valid {
			return nil
		}
	}

	return &ProjectError{
		Type:    ProjectNameInvalid,
		Path:    p.Path,
		Message: fmt.Sprintf("invalid AI assistant '%s', must be one of: %s", 
			p.AIAssistant, strings.Join(ValidAIAssistants, ", ")),
	}
}

// GetDisplayName returns the display name for the project
func (p *Project) GetDisplayName() string {
	if p.IsHere {
		return filepath.Base(p.Path)
	}
	return p.Name
}

// GetAIAssistantDisplayName returns the full display name for the AI assistant
func (p *Project) GetAIAssistantDisplayName() string {
	switch p.AIAssistant {
	case "claude":
		return "Claude Code"
	case "gemini":
		return "Gemini CLI"
	case "copilot":
		return "GitHub Copilot"
	default:
		return p.AIAssistant
	}
}

// ShouldInitializeGit returns whether git should be initialized for this project
func (p *Project) ShouldInitializeGit() bool {
	// Don't initialize git if already in a git repository
	// This will be determined by the environment service
	return !p.HasGit
}