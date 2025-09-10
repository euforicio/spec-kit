package contract

import (
	"testing"

	"github.com/euforicio/spec-kit/internal/models"
	"github.com/stretchr/testify/assert"
)

// Contract test for agent validation based on agent-validation.md contract
func TestAgentValidation_ValidateAgentType(t *testing.T) {
	tests := []struct {
		name          string
		agentType     string
		expectedValid bool
		expectedName  string
		expectedError string
	}{
		{
			name:          "valid claude agent",
			agentType:     "claude",
			expectedValid: true,
			expectedName:  "Claude Code",
			expectedError: "",
		},
		{
			name:          "valid gemini agent",
			agentType:     "gemini",
			expectedValid: true,
			expectedName:  "Gemini CLI",
			expectedError: "",
		},
		{
			name:          "valid copilot agent",
			agentType:     "copilot",
			expectedValid: true,
			expectedName:  "GitHub Copilot",
			expectedError: "",
		},
		{
			name:          "valid codex agent",
			agentType:     "codex",
			expectedValid: true,
			expectedName:  "OpenAI Codex",
			expectedError: "",
		},
		{
			name:          "empty agent type",
			agentType:     "",
			expectedValid: false,
			expectedName:  "",
			expectedError: "Agent type cannot be empty",
		},
		{
			name:          "unknown agent type",
			agentType:     "unknown",
			expectedValid: false,
			expectedName:  "",
			expectedError: "Unsupported agent type: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test MUST fail initially - ValidateAgentType doesn't exist yet
			isValid, displayName, errorMsg := models.ValidateAgentType(tt.agentType)

			assert.Equal(t, tt.expectedValid, isValid, "Validation result mismatch")
			assert.Equal(t, tt.expectedName, displayName, "Display name mismatch")
			assert.Equal(t, tt.expectedError, errorMsg, "Error message mismatch")
		})
	}
}

// Test that "codex" is in ValidAgents list
func TestAgentValidation_CodexInValidTypes(t *testing.T) {
	// This test MUST fail initially - "codex" not yet added to ValidAgents
	assert.Contains(t, models.ListAgents(), "codex", "codex should be in ValidAgents")
}

// Test that "codex" is in ValidAgents list
func TestAgentValidation_CodexInValidAIAssistants(t *testing.T) {
	// This test MUST fail initially - "codex" not yet added to ValidAgents
	assert.Contains(t, models.ListAgents(), "codex", "codex should be in ValidAgents")
}
