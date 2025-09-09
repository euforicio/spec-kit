package integration_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/euforicio/spec-kit/internal/codex"
)

func TestAgentsMdGeneration(t *testing.T) {
	t.Run("should generate AGENTS.md with Codex section", func(t *testing.T) {
		
		// Create a mock plan content
		planContent := `# Implementation Plan
		
## Technical Context
**Language/Version**: Go 1.25
**Primary Dependencies**: cobra, embed
**Testing**: go test
**Project Type**: single
`
		
		// Generate AGENTS.md content
		generator := codex.NewAgentsGenerator()
		content, err := generator.Generate(planContent, nil)
		
		// This should fail initially (TDD)
		require.NoError(t, err)
		assert.NotEmpty(t, content)
		
		// Verify structure
		assert.Contains(t, content, "# AGENTS.md")
		assert.Contains(t, content, "## For OpenAI Codex")
		assert.Contains(t, content, "### Command System")
		assert.Contains(t, content, "### Available Commands")
		assert.Contains(t, content, "### Natural Language Patterns")
		assert.Contains(t, content, "### Workflow")
		assert.Contains(t, content, "## Project Context")
		
		// Verify command instructions
		assert.Contains(t, content, "When you see a message starting with \"/\"")
		assert.Contains(t, content, ".codex/commands/")
		assert.Contains(t, content, "/specify")
		assert.Contains(t, content, "/plan")
		assert.Contains(t, content, "/tasks")
	})
	
	t.Run("should merge with existing AGENTS.md content", func(t *testing.T) {
		existingContent := `# AGENTS.md

This file provides context for AI assistants.

## For GitHub Copilot

Some existing Copilot configuration.

## Project Context

Existing project information.
`
		
		generator := codex.NewAgentsGenerator()
		merged, err := generator.Generate("", []byte(existingContent))
		
		require.NoError(t, err)
		
		// Check existing content is preserved
		assert.Contains(t, merged, "## For GitHub Copilot")
		assert.Contains(t, merged, "Some existing Copilot configuration")
		
		// Check Codex section is added
		assert.Contains(t, merged, "## For OpenAI Codex")
		
		// Ensure no duplication
		codexCount := strings.Count(merged, "## For OpenAI Codex")
		assert.Equal(t, 1, codexCount, "Codex section should appear only once")
	})
	
	t.Run("should extract tech stack from plan", func(t *testing.T) {
		planContent := `# Implementation Plan

## Technical Context
**Language/Version**: Python 3.11
**Primary Dependencies**: FastAPI, SQLAlchemy, Pytest
**Storage**: PostgreSQL
**Testing**: pytest
**Project Type**: web
`
		
		generator := codex.NewAgentsGenerator()
		content, err := generator.Generate(planContent, nil)
		
		require.NoError(t, err)
		
		// Check tech stack is included
		assert.Contains(t, content, "Python 3.11")
		assert.Contains(t, content, "FastAPI")
		assert.Contains(t, content, "PostgreSQL")
	})
	
	t.Run("should include command descriptions", func(t *testing.T) {
		generator := codex.NewAgentsGenerator()
		content, err := generator.Generate("", nil)
		
		require.NoError(t, err)
		
		// Check command descriptions
		expectedCommands := map[string]string{
			"/specify":   "Create a new feature specification",
			"/plan":      "Plan implementation for current feature",
			"/tasks":     "Generate task breakdown",
			"/test":      "Write tests following TDD",
			"/implement": "Implement code to pass tests",
			"/validate":  "Validate against specification",
			"/commit":    "Create git commit",
		}
		
		for cmd, desc := range expectedCommands {
			assert.Contains(t, content, cmd)
			assert.Contains(t, content, desc)
		}
	})
	
	t.Run("should include natural language patterns", func(t *testing.T) {
		generator := codex.NewAgentsGenerator()
		content, err := generator.Generate("", nil)
		
		require.NoError(t, err)
		
		// Check patterns
		patterns := []string{
			"create a spec for",
			"plan this",
			"break down tasks",
			"write tests for",
			"implement",
		}
		
		for _, pattern := range patterns {
			assert.Contains(t, content, pattern)
		}
	})
}