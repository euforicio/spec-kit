package integration_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/euforicio/spec-kit/internal/codex"
)

func TestAgentsMdIncrementalUpdate(t *testing.T) {
	t.Run("should update existing AGENTS.md without losing content", func(t *testing.T) {
		existingContent := `# AGENTS.md

This file provides context for AI assistants when working with this repository.

## For GitHub Copilot

### Configuration
Custom Copilot settings here.

## Project Context

- Language: Python 3.10
- Framework: Django
- Database: MySQL

## Custom Section

This is a manually added section that should be preserved.
`
		
		planContent := `# Implementation Plan

## Technical Context
**Language/Version**: Go 1.25
**Primary Dependencies**: cobra, embed
**Testing**: go test
`
		
		generator := codex.NewAgentsGenerator()
		updated, err := generator.Generate(planContent, []byte(existingContent))
		
		// This should fail initially (TDD)
		require.NoError(t, err)
		
		// Check all sections are present
		assert.Contains(t, updated, "## For GitHub Copilot")
		assert.Contains(t, updated, "Custom Copilot settings here")
		assert.Contains(t, updated, "## For OpenAI Codex")
		assert.Contains(t, updated, "## Custom Section")
		assert.Contains(t, updated, "This is a manually added section")
		
		// Check project context is updated with new info
		assert.Contains(t, updated, "Go 1.25")
		assert.Contains(t, updated, "cobra")
	})
	
	t.Run("should update existing Codex section", func(t *testing.T) {
		existingContent := `# AGENTS.md

## For OpenAI Codex

### Old Content
This should be replaced.

## Project Context
Old project info.
`
		
		generator := codex.NewAgentsGenerator()
		updated, err := generator.Generate("", []byte(existingContent))
		
		require.NoError(t, err)
		
		// Old content should be replaced
		assert.NotContains(t, updated, "### Old Content")
		assert.NotContains(t, updated, "This should be replaced")
		
		// New content should be present
		assert.Contains(t, updated, "### Command System")
		assert.Contains(t, updated, "### Available Commands")
		
		// Should not duplicate the section
		codexCount := strings.Count(updated, "## For OpenAI Codex")
		assert.Equal(t, 1, codexCount)
	})
	
	t.Run("should preserve order of sections", func(t *testing.T) {
		existingContent := `# AGENTS.md

## Section A
Content A

## For GitHub Copilot
Copilot content

## Section B
Content B

## Project Context
Project info
`
		
		generator := codex.NewAgentsGenerator()
		updated, err := generator.Generate("", []byte(existingContent))
		
		require.NoError(t, err)
		
		// Find positions of sections
		posA := strings.Index(updated, "## Section A")
		posCopilot := strings.Index(updated, "## For GitHub Copilot")
		posCodex := strings.Index(updated, "## For OpenAI Codex")
		posB := strings.Index(updated, "## Section B")
		posContext := strings.Index(updated, "## Project Context")
		
		// Codex should be inserted after Copilot, before Section B
		assert.Less(t, posA, posCopilot)
		assert.Less(t, posCopilot, posCodex)
		assert.Less(t, posCodex, posB)
		assert.Less(t, posB, posContext)
	})
	
	t.Run("should handle empty existing file", func(t *testing.T) {
		generator := codex.NewAgentsGenerator()
		updated, err := generator.Generate("", []byte(""))
		
		require.NoError(t, err)
		
		// Should create full structure
		assert.Contains(t, updated, "# AGENTS.md")
		assert.Contains(t, updated, "## For OpenAI Codex")
		assert.Contains(t, updated, "## Project Context")
	})
	
	t.Run("should update project context from plan", func(t *testing.T) {
		existingContent := `# AGENTS.md

## Project Context
- Language: Unknown
- Dependencies: None
`
		
		planContent := `# Implementation Plan

## Technical Context
**Language/Version**: TypeScript 5.0
**Primary Dependencies**: Next.js, React, Prisma
**Storage**: PostgreSQL
**Testing**: Jest, React Testing Library
`
		
		generator := codex.NewAgentsGenerator()
		updated, err := generator.Generate(planContent, []byte(existingContent))
		
		require.NoError(t, err)
		
		// Old context should be updated
		assert.NotContains(t, updated, "Language: Unknown")
		assert.NotContains(t, updated, "Dependencies: None")
		
		// New context should be present
		assert.Contains(t, updated, "TypeScript 5.0")
		assert.Contains(t, updated, "Next.js")
		assert.Contains(t, updated, "PostgreSQL")
		assert.Contains(t, updated, "Jest")
	})
	
	t.Run("should maintain markdown formatting", func(t *testing.T) {
		generator := codex.NewAgentsGenerator()
		updated, err := generator.Generate("", nil)
		
		require.NoError(t, err)
		
		// Check proper markdown structure
		lines := strings.Split(updated, "\n")
		
		// Should start with main heading
		assert.True(t, strings.HasPrefix(lines[0], "# "))
		
		// Should have proper spacing
		for i := 0; i < len(lines)-1; i++ {
			if strings.HasPrefix(lines[i], "## ") {
				// Section headings should have blank line before (except first)
				if i > 0 && lines[i-1] != "" {
					assert.Empty(t, lines[i-1], "Should have blank line before section heading")
				}
			}
		}
	})
}