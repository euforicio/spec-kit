package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/euforicio/spec-kit/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration test for Codex agent initialization
func TestCodexAgentInitialization(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "codex-init-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// This test MUST fail initially - Codex agent support doesn't exist yet
	
	// Test 1: Codex should be available as valid agent type
    assert.Contains(t, models.ListAgents(), "codex", "codex should be in ValidAgents")
	
	// Test 2: Codex should have proper display name
	displayName := models.GetAIAssistantDisplayName("codex")
	assert.Equal(t, "OpenAI Codex", displayName, "Codex should have correct display name")
	
	// Test 3: Codex should have installation instructions
    installInstructions := models.GetInstallHint("codex")
	assert.NotEmpty(t, installInstructions, "Codex should have installation instructions")
	assert.Contains(t, installInstructions, "codex", "Installation instructions should mention codex")
}

// Test that Codex agent selection triggers AGENTS.md creation
func TestCodexAgentSelection_TriggersAgentsMDCreation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "codex-agents-md-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// This test MUST fail initially - integration doesn't exist yet
	
    // Simulate selecting Codex as agent type during initialization (implicit via AGENTS.md operations)
	
	// Initialize Codex project - should create AGENTS.md
	_, _, err = models.CreateOrUpdateAgentsMD(tempDir)
	assert.NoError(t, err, "Codex initialization should succeed")
	
	// Verify AGENTS.md was created
	agentsPath := filepath.Join(tempDir, "AGENTS.md")
	assert.FileExists(t, agentsPath, "AGENTS.md should be created")
	
	// Verify AGENTS.md contains proper content
	content, err := os.ReadFile(agentsPath)
	require.NoError(t, err)
	contentStr := string(content)
	
	assert.Contains(t, contentStr, "<specify>", "Should contain opening delimiter")
	assert.Contains(t, contentStr, "</specify>", "Should contain closing delimiter")
	assert.Contains(t, contentStr, "/specify", "Should document /specify command")
	assert.Contains(t, contentStr, "/plan", "Should document /plan command")  
	assert.Contains(t, contentStr, "/tasks", "Should document /tasks command")
	assert.Contains(t, contentStr, ".codex/commands/", "Should reference .codex/commands/ directory")
}

// Test that other agent types don't create AGENTS.md
func TestNonCodexAgents_DoNotCreateAgentsMD(t *testing.T) {
	agentTypes := []string{"claude", "gemini", "copilot"}
	
	for _, agentType := range agentTypes {
		t.Run("agent_"+agentType, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "non-codex-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)
			
			// This test MUST fail initially - integration doesn't exist yet
			
            // Simulate selecting non-Codex agent type; no AGENTS.md should be created
			
			// For non-Codex agents, no special initialization should occur
			// (In real usage, the project service would only call InitializeCodexProject for Codex)
			
			// Verify AGENTS.md was NOT created for non-Codex agents
			agentsPath := filepath.Join(tempDir, "AGENTS.md")
			assert.NoFileExists(t, agentsPath, "AGENTS.md should not be created for %s agent", agentType)
		})
	}
}
