package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/euforicio/spec-kit/internal/codex"
)

func TestCommandFileCreation(t *testing.T) {
	t.Run("should create command files in .codex/commands", func(t *testing.T) {
		tempDir := t.TempDir()
		
		creator := codex.NewCommandCreator()
		err := creator.CreateCommands(tempDir, codex.CommandSetStandard)
		
		// This should fail initially (TDD)
		require.NoError(t, err)
		
		// Check directory structure
		codexDir := filepath.Join(tempDir, ".codex", "commands")
		assert.DirExists(t, codexDir)
		
		// Check core command files exist
		expectedFiles := []string{
			"specify.md",
			"plan.md",
			"tasks.md",
			"test.md",
			"implement.md",
			"validate.md",
			"commit.md",
		}
		
		for _, file := range expectedFiles {
			filePath := filepath.Join(codexDir, file)
			assert.FileExists(t, filePath, "Command file %s should exist", file)
			
			// Verify file is not empty
			content, err := os.ReadFile(filePath)
			require.NoError(t, err)
			assert.NotEmpty(t, content)
		}
	})
	
	t.Run("should create valid command file structure", func(t *testing.T) {
		tempDir := t.TempDir()
		
		creator := codex.NewCommandCreator()
		err := creator.CreateCommands(tempDir, codex.CommandSetStandard)
		require.NoError(t, err)
		
		// Read and verify a command file structure
		specifyPath := filepath.Join(tempDir, ".codex", "commands", "specify.md")
		content, err := os.ReadFile(specifyPath)
		require.NoError(t, err)
		
		contentStr := string(content)
		
		// Check required sections
		assert.Contains(t, contentStr, "# /specify")
		assert.Contains(t, contentStr, "## Description")
		assert.Contains(t, contentStr, "## Usage")
		assert.Contains(t, contentStr, "## Execute")
		assert.Contains(t, contentStr, "## Report")
		
		// Check command execution steps
		assert.Contains(t, contentStr, "specify feature create")
		assert.Contains(t, contentStr, "Parse JSON output")
		assert.Contains(t, contentStr, "Success:")
		assert.Contains(t, contentStr, "Error:")
	})
	
	t.Run("should support minimal command set", func(t *testing.T) {
		tempDir := t.TempDir()
		
		creator := codex.NewCommandCreator()
		err := creator.CreateCommands(tempDir, codex.CommandSetMinimal)
		require.NoError(t, err)
		
		codexDir := filepath.Join(tempDir, ".codex", "commands")
		entries, err := os.ReadDir(codexDir)
		require.NoError(t, err)
		
		// Minimal set should have fewer commands
		assert.LessOrEqual(t, len(entries), 5)
		
		// But must have core commands
		assert.FileExists(t, filepath.Join(codexDir, "specify.md"))
		assert.FileExists(t, filepath.Join(codexDir, "plan.md"))
		assert.FileExists(t, filepath.Join(codexDir, "tasks.md"))
	})
	
	t.Run("should not overwrite existing custom commands", func(t *testing.T) {
		tempDir := t.TempDir()
		codexDir := filepath.Join(tempDir, ".codex", "commands")
		
		// Create directory and custom command
		require.NoError(t, os.MkdirAll(codexDir, 0755))
		customPath := filepath.Join(codexDir, "custom.md")
		customContent := "# /custom\nMy custom command"
		require.NoError(t, os.WriteFile(customPath, []byte(customContent), 0644))
		
		// Create commands
		creator := codex.NewCommandCreator()
		err := creator.CreateCommands(tempDir, codex.CommandSetStandard)
		require.NoError(t, err)
		
		// Check custom command still exists
		content, err := os.ReadFile(customPath)
		require.NoError(t, err)
		assert.Equal(t, customContent, string(content))
	})
	
	t.Run("should include natural language patterns in commands", func(t *testing.T) {
		tempDir := t.TempDir()
		
		creator := codex.NewCommandCreator()
		err := creator.CreateCommands(tempDir, codex.CommandSetStandard)
		require.NoError(t, err)
		
		// Check specify command has patterns
		specifyPath := filepath.Join(tempDir, ".codex", "commands", "specify.md")
		content, err := os.ReadFile(specifyPath)
		require.NoError(t, err)
		
		contentStr := string(content)
		assert.Contains(t, contentStr, "## Natural Language")
		assert.Contains(t, contentStr, "create a spec for")
		assert.Contains(t, contentStr, "start a new feature")
	})
	
	t.Run("should handle force overwrite", func(t *testing.T) {
		tempDir := t.TempDir()
		
		// Create initial commands
		creator := codex.NewCommandCreator()
		err := creator.CreateCommands(tempDir, codex.CommandSetStandard)
		require.NoError(t, err)
		
		// Modify a command file
		specifyPath := filepath.Join(tempDir, ".codex", "commands", "specify.md")
		require.NoError(t, os.WriteFile(specifyPath, []byte("modified"), 0644))
		
		// Force overwrite
		err = creator.CreateCommandsForce(tempDir, codex.CommandSetStandard)
		require.NoError(t, err)
		
		// Check file was overwritten
		content, err := os.ReadFile(specifyPath)
		require.NoError(t, err)
		assert.NotEqual(t, "modified", string(content))
		assert.Contains(t, string(content), "# /specify")
	})
}