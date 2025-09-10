package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/euforicio/spec-kit/internal/models"
)

// EnvironmentService handles environment detection and validation
type EnvironmentService struct {
	filesystem *FilesystemService
}

// NewEnvironmentService creates a new environment service instance
func NewEnvironmentService(filesystem *FilesystemService) *EnvironmentService {
	return &EnvironmentService{
		filesystem: filesystem,
	}
}

// DetectEnvironment detects and returns the current environment information
func (e *EnvironmentService) DetectEnvironment() (*models.Environment, error) {
	workingDir, err := e.filesystem.GetWorkingDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	env := models.NewEnvironment(workingDir)

	// Detect platform (already set by NewEnvironment using runtime.GOOS)

	// Check internet connectivity
	env.HasInternet = e.checkInternetConnectivity()

	// Check required tools
	e.checkRequiredTools(env)

	// Check optional tools
	e.checkOptionalTools(env)

	// Check git configuration
	e.checkGitConfiguration(env)

	// Validate the environment
	if err := env.Validate(); err != nil {
		return nil, fmt.Errorf("environment validation failed: %w", err)
	}

	return env, nil
}

// checkInternetConnectivity tests internet connectivity
func (e *EnvironmentService) checkInternetConnectivity() bool {
	// Test multiple endpoints for reliability
	testURLs := []string{
		"https://api.github.com",
		"https://www.google.com",
		"https://www.cloudflare.com",
	}

	for _, url := range testURLs {
		if e.testConnectivity(url) {
			return true
		}
	}

	return false
}

// testConnectivity tests connectivity to a specific URL
func (e *EnvironmentService) testConnectivity(url string) bool {
	// Use curl if available, otherwise skip
	cmd := exec.Command("curl", "-s", "--max-time", "5", "--head", url)
	err := cmd.Run()
	return err == nil
}

// checkRequiredTools checks for required development tools
func (e *EnvironmentService) checkRequiredTools(env *models.Environment) {
	for tool, installHint := range models.RequiredTools {
		version, available := e.checkTool(tool)
		env.SetToolStatus(tool, available, version, installHint)
	}
}

// checkOptionalTools checks for optional development tools
func (e *EnvironmentService) checkOptionalTools(env *models.Environment) {
	for tool, installHint := range models.OptionalTools {
		version, available := e.checkTool(tool)
		env.SetToolStatus(tool, available, version, installHint)
	}
}

// checkTool checks if a specific tool is available and returns its version
func (e *EnvironmentService) checkTool(tool string) (version string, available bool) {
	// Try to find the tool in PATH
	_, err := exec.LookPath(tool)
	if err != nil {
		return "", false
	}

	// Try to get version information
	version = e.getToolVersion(tool)
	return version, true
}

// getToolVersion attempts to get version information for a tool
func (e *EnvironmentService) getToolVersion(tool string) string {
	// Common version flags to try
	versionFlags := []string{"--version", "-v", "version", "-V"}

	for _, flag := range versionFlags {
		cmd := exec.Command(tool, flag)
		output, err := cmd.CombinedOutput()
		if err == nil {
			// Parse the first line of output and clean it up
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 {
				version := strings.TrimSpace(lines[0])
				// Remove common prefixes
				version = strings.TrimPrefix(version, tool+" ")
				version = strings.TrimPrefix(version, tool+" version ")
				version = strings.TrimPrefix(version, "version ")
				return version
			}
		}
	}

	return "unknown"
}

// checkGitConfiguration checks git installation and configuration
func (e *EnvironmentService) checkGitConfiguration(env *models.Environment) {
	version, available := e.checkTool("git")
	if !available {
		env.SetGitConfig(false, "", "", "")
		return
	}

	// Get git user configuration
	userName := e.getGitConfig("user.name")
	userEmail := e.getGitConfig("user.email")

	env.SetGitConfig(true, version, userName, userEmail)
}

// getGitConfig gets a git configuration value
func (e *EnvironmentService) getGitConfig(key string) string {
	cmd := exec.Command("git", "config", "--global", "--get", key)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// IsInGitRepository checks if the current directory is inside a git repository
func (e *EnvironmentService) IsInGitRepository(path string) bool {
	if path == "" {
		var err error
		path, err = e.filesystem.GetWorkingDirectory()
		if err != nil {
			return false
		}
	}

	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = path
	err := cmd.Run()
	return err == nil
}

// InitializeGitRepository initializes a new git repository
func (e *EnvironmentService) InitializeGitRepository(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		return &models.EnvironmentError{
			Type:    models.ToolNotFound,
			Tool:    "git",
			Message: "git is not installed or not found in PATH",
			Hint:    "Install git from https://git-scm.com/downloads",
		}
	}

	// Initialize repository
	cmd := exec.Command("git", "init")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Add all files
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}

	// Create initial commit
	cmd = exec.Command("git", "commit", "-m", "Initial commit from Specify template")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	return nil
}

// CheckDiskSpace checks available disk space at the given path
func (e *EnvironmentService) CheckDiskSpace(path string) (free, total uint64, err error) {
	// This is a simplified version - would need platform-specific implementation
	// For now, return reasonable defaults
	return 1 << 30, 10 << 30, nil // 1GB free, 10GB total
}

// GetSystemInfo returns system information
func (e *EnvironmentService) GetSystemInfo() map[string]string {
	info := make(map[string]string)

	info["os"] = runtime.GOOS
	info["arch"] = runtime.GOARCH
	info["go_version"] = runtime.Version()
	info["num_cpu"] = fmt.Sprintf("%d", runtime.NumCPU())

	// Get hostname
	if hostname, err := os.Hostname(); err == nil {
		info["hostname"] = hostname
	}

	// Get user info
	if user := os.Getenv("USER"); user != "" {
		info["user"] = user
	} else if user := os.Getenv("USERNAME"); user != "" {
		info["user"] = user
	}

	// Get shell
	if shell := os.Getenv("SHELL"); shell != "" {
		info["shell"] = shell
	}

	return info
}

// ValidateProjectPath validates that a path is suitable for project creation
func (e *EnvironmentService) ValidateProjectPath(path string, isHere bool) error {
	// Check if path exists
	exists, err := e.filesystem.DirectoryExists(path)
	if err != nil {
		return fmt.Errorf("failed to check path existence: %w", err)
	}

	if !isHere && exists {
		return &models.ProjectError{
			Type:    models.ProjectAlreadyExists,
			Path:    path,
			Message: "directory already exists",
		}
	}

	// Check parent directory permissions
	parentDir := path
	if !exists {
		parentDir = filepath.Dir(path)
	}

	if err := e.filesystem.IsWritable(parentDir); err != nil {
		return err
	}

	return nil
}

// GetRecommendations returns environment setup recommendations
func (e *EnvironmentService) GetRecommendations(env *models.Environment) []string {
	var recommendations []string

	// Check internet connectivity
	if !env.HasInternet {
		recommendations = append(recommendations,
			"Internet connection is required to download templates")
	}

	// Check git
	if !env.IsToolAvailable("git") {
		recommendations = append(recommendations,
			"Install git for version control: https://git-scm.com/downloads")
	} else if !env.IsGitConfigured() {
		recommendations = append(recommendations,
			"Configure git with your name and email: git config --global user.name \"Your Name\"")
	}

	// Check AI tools
	hasAnyAI := slices.ContainsFunc(models.ListAgents(), func(ai string) bool { return env.IsToolAvailable(ai) })
	if !hasAnyAI {
		recommendations = append(recommendations,
			"Consider installing an AI assistant (Claude Code, Gemini CLI, GitHub Copilot, or OpenAI Codex) for the best experience")
	}

	return recommendations
}

// WaitForConnectivity waits for internet connectivity to be restored
func (e *EnvironmentService) WaitForConnectivity(timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		if e.checkInternetConnectivity() {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return &models.EnvironmentError{
		Type:    models.InternetNotAvailable,
		Message: "timeout waiting for internet connectivity",
		Hint:    "Check your network connection and try again",
	}
}
