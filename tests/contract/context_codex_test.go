package contract_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeatureContextCodex(t *testing.T) {
	// Setup: Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Initialize a git repo and create a feature branch
	setupTestRepo(t, tempDir)
	
	// Change to the test directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	require.NoError(t, os.Chdir(tempDir))
	
	t.Run("should generate AGENTS.md file", func(t *testing.T) {
		// Run the command
		cmd := exec.Command("specify", "feature", "context", "codex")
		output, err := cmd.CombinedOutput()
		
		// For TDD: This should fail initially
		require.NoError(t, err, "Command failed: %s", string(output))
		
		// Check AGENTS.md was created
		agentsPath := filepath.Join(tempDir, "AGENTS.md")
		assert.FileExists(t, agentsPath, "AGENTS.md should be created")
		
		// Read and verify content
		content, err := os.ReadFile(agentsPath)
		require.NoError(t, err)
		
		// Check for Codex-specific content
		assert.Contains(t, string(content), "## For OpenAI Codex")
		assert.Contains(t, string(content), "### Command System")
		assert.Contains(t, string(content), ".codex/commands/")
	})
	
	t.Run("should create .codex/commands directory", func(t *testing.T) {
		// Check directory exists
		codexDir := filepath.Join(tempDir, ".codex", "commands")
		assert.DirExists(t, codexDir, ".codex/commands directory should exist")
		
		// Check for command files
		entries, err := os.ReadDir(codexDir)
		require.NoError(t, err)
		assert.NotEmpty(t, entries, "Command files should be created")
		
		// Verify at least core commands exist
		expectedCommands := []string{"specify.md", "plan.md", "tasks.md"}
		for _, cmd := range expectedCommands {
			cmdPath := filepath.Join(codexDir, cmd)
			assert.FileExists(t, cmdPath, "Command file %s should exist", cmd)
		}
	})
	
	t.Run("should update existing AGENTS.md without overwriting", func(t *testing.T) {
		// Create an existing AGENTS.md with custom content
		agentsPath := filepath.Join(tempDir, "AGENTS.md")
		customContent := "# AGENTS.md\n\n## Custom Section\nThis should be preserved\n"
		require.NoError(t, os.WriteFile(agentsPath, []byte(customContent), 0644))
		
		// Run the command again
		cmd := exec.Command("specify", "feature", "context", "codex")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Command failed: %s", string(output))
		
		// Read updated content
		content, err := os.ReadFile(agentsPath)
		require.NoError(t, err)
		
		// Check both custom and Codex content exist
		assert.Contains(t, string(content), "## Custom Section")
		assert.Contains(t, string(content), "This should be preserved")
		assert.Contains(t, string(content), "## For OpenAI Codex")
	})
	
	t.Run("should handle --force flag", func(t *testing.T) {
		// Run with --force flag
		cmd := exec.Command("specify", "feature", "context", "codex", "--force")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Command failed: %s", string(output))
		
		// Verify files were recreated
		assert.Contains(t, string(output), "Created/Updated AGENTS.md")
	})
	
	t.Run("should handle --minimal flag", func(t *testing.T) {
		// Clean up commands directory
		codexDir := filepath.Join(tempDir, ".codex", "commands")
		os.RemoveAll(codexDir)
		
		// Run with --minimal flag
		cmd := exec.Command("specify", "feature", "context", "codex", "--minimal")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Command failed: %s", string(output))
		
		// Check only core commands exist
		entries, err := os.ReadDir(codexDir)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(entries), 5, "Should have minimal command set")
	})
	
	t.Run("should handle --no-commands flag", func(t *testing.T) {
		// Clean up
		codexDir := filepath.Join(tempDir, ".codex")
		os.RemoveAll(codexDir)
		
		// Run with --no-commands flag
		cmd := exec.Command("specify", "feature", "context", "codex", "--no-commands")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Command failed: %s", string(output))
		
		// Check AGENTS.md exists but not commands directory
		agentsPath := filepath.Join(tempDir, "AGENTS.md")
		assert.FileExists(t, agentsPath)
		
		commandsDir := filepath.Join(tempDir, ".codex", "commands")
		assert.NoDirExists(t, commandsDir, "Commands directory should not be created")
	})
}

func setupTestRepo(t *testing.T, dir string) {
	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
	
	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
	
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
	
	// Create initial commit
	readmePath := filepath.Join(dir, "README.md")
	require.NoError(t, os.WriteFile(readmePath, []byte("# Test Project"), 0644))
	
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
	
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
	
	// Create and checkout feature branch
	cmd = exec.Command("git", "checkout", "-b", "001-test-feature")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
	
	// Create specs directory and plan.md
	specsDir := filepath.Join(dir, "specs", "001-test-feature")
	require.NoError(t, os.MkdirAll(specsDir, 0755))
	
	planPath := filepath.Join(specsDir, "plan.md")
	planContent := `# Implementation Plan

## Technical Context
**Language/Version**: Go 1.25
**Primary Dependencies**: cobra, testify
**Testing**: go test
**Project Type**: single
`
	require.NoError(t, os.WriteFile(planPath, []byte(planContent), 0644))
}