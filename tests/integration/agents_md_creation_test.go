package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/euforicio/spec-kit/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration test for AGENTS.md file creation scenarios
func TestAgentsMDCreation_NewFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agents-md-creation-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Verify file doesn't exist initially
	agentsPath := filepath.Join(tempDir, "AGENTS.md")
	assert.NoFileExists(t, agentsPath, "AGENTS.md should not exist initially")

	// This test MUST fail initially - CreateAgentsMD function doesn't exist yet
	filePath, err := models.CreateAgentsMD(tempDir)
	
	assert.NoError(t, err, "AGENTS.md creation should succeed")
	assert.Equal(t, agentsPath, filePath, "Returned path should match expected")
	
	// Verify file was created
	assert.FileExists(t, agentsPath, "AGENTS.md should be created")
	
	// Verify file content
	content, err := os.ReadFile(agentsPath)
	require.NoError(t, err)
	contentStr := string(content)
	
	// Should have basic structure
	assert.Contains(t, contentStr, "# Agent Instructions", "Should have header")
	assert.Contains(t, contentStr, "<specify>", "Should have opening delimiter")
	assert.Contains(t, contentStr, "</specify>", "Should have closing delimiter")
	
	// Should document slash commands
	assert.Contains(t, contentStr, "/specify", "Should document /specify command")
	assert.Contains(t, contentStr, "/plan", "Should document /plan command")
	assert.Contains(t, contentStr, "/tasks", "Should document /tasks command")
	
	// Should reference .codex/commands/ directory structure
	assert.Contains(t, contentStr, ".codex/commands/<command>.md", "Should reference generic command structure")
	assert.Contains(t, contentStr, "Documentation Structure", "Should have documentation structure section")
	
	// Should explain command flow
	assert.Contains(t, contentStr, "Command Flow", "Should explain command flow")
	assert.Contains(t, contentStr, "Built-in Commands", "Should have built-in commands section")
}

// Test that AGENTS.md has proper markdown structure
func TestAgentsMDCreation_ValidMarkdown(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agents-md-markdown-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// This test MUST fail initially - CreateAgentsMD function doesn't exist yet
	filePath, err := models.CreateAgentsMD(tempDir)
	require.NoError(t, err)
	
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	contentStr := string(content)
	
	// Should be valid markdown with proper structure
	lines := strings.Split(contentStr, "\n")
	
	// Should have markdown headers
	hasH1 := false
	hasH2 := false
	hasH3 := false
	
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			hasH1 = true
		}
		if strings.HasPrefix(line, "## ") {
			hasH2 = true
		}
		if strings.HasPrefix(line, "### ") {
			hasH3 = true
		}
	}
	
	assert.True(t, hasH1, "Should have H1 headers")
	assert.True(t, hasH2, "Should have H2 headers") 
	assert.True(t, hasH3, "Should have H3 headers")
	
	// Should have proper formatting
	assert.Contains(t, contentStr, "**`/specify`**", "Should have formatted command names")
	assert.Contains(t, contentStr, "- Each command has detailed documentation", "Should have documentation structure info")
	assert.Contains(t, contentStr, "- Additional commands can be added", "Should explain extensibility")
}

// Test AGENTS.md creation with file permission errors
func TestAgentsMDCreation_PermissionError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}
	
	// Create directory with no write permissions
	tempDir, err := os.MkdirTemp("", "agents-md-permission-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Remove write permission
	err = os.Chmod(tempDir, 0444)
	require.NoError(t, err)
	
	// Restore permissions for cleanup
	defer func() {
		os.Chmod(tempDir, 0755)
	}()

	// This test MUST fail initially - CreateAgentsMD function doesn't exist yet
	_, err = models.CreateAgentsMD(tempDir)
	
	assert.Error(t, err, "Should fail with permission error")
	assert.Contains(t, err.Error(), "permission denied", "Error should mention permission denied")
}

// Test AGENTS.md creation in non-existent directory
func TestAgentsMDCreation_InvalidDirectory(t *testing.T) {
	invalidDir := "/nonexistent/directory/path"
	
	// This test MUST fail initially - CreateAgentsMD function doesn't exist yet
	_, err := models.CreateAgentsMD(invalidDir)
	
	assert.Error(t, err, "Should fail with invalid directory")
	assert.Contains(t, strings.ToLower(err.Error()), "no such file or directory", "Error should mention missing directory")
}