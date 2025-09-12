package models

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/euforicio/spec-kit/internal/template"
)

// AIAssistantDisplayNames maps agent identifiers to their display names.
// This is the single source of truth for supported agents.
var AIAssistantDisplayNames = map[string]string{
	"claude":  "Claude Code",
	"gemini":  "Gemini CLI",
	"copilot": "GitHub Copilot",
	"codex":   "OpenAI Codex",
}

// ListAgents returns the supported AI assistant identifiers (sorted).
func ListAgents() []string {
	keys := slices.Collect(maps.Keys(AIAssistantDisplayNames))
	slices.Sort(keys)
	return keys
}

// IsValidAgent reports whether the identifier is a supported agent.
func IsValidAgent(agent string) bool {
	_, ok := AIAssistantDisplayNames[agent]
	return ok
}

// Precompiled regex for <specify>...</specify> sections to avoid recompilation
var specifySectionRe = regexp.MustCompile(`(?s)<specify>.*?</specify>`)

// Global template processor instance for operations
var templateProcessor = template.NewProcessor()

// processTemplate processes a template string with the given data
func processTemplate(content string, data template.Data) (string, error) {
	return templateProcessor.Process(content, data)
}

// Agent-agnostic AGENTS.md file operations
// This file contains code to support all AI agents' slash command requirements.

// CreateOrUpdateAgentsMD creates or updates AGENTS.md file with template content for any agent
func CreateOrUpdateAgentsMD(
	aiAssistant, projectRoot string,
) (filePath string, created bool, err error) {
	if projectRoot == "" {
		return "", false, fmt.Errorf("invalid project root directory: path cannot be empty")
	}

	// Check if project root exists
	if _, statErr := os.Stat(projectRoot); os.IsNotExist(statErr) {
		return "", false, fmt.Errorf("invalid project root directory: %s", projectRoot)
	}

	agentsPath := filepath.Join(projectRoot, "AGENTS.md")

	// Check if file exists
	if _, statErr := os.Stat(agentsPath); os.IsNotExist(statErr) {
		// File doesn't exist, create it
		content := getAgentsMDTemplate(aiAssistant)
		err = os.WriteFile(agentsPath, []byte(content), 0o644)
		if err != nil {
			return "", false, fmt.Errorf("failed to write AGENTS.md: %w", err)
		}
		return agentsPath, true, nil
	}

	// File exists, read and update it
	content, err := os.ReadFile(agentsPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read existing AGENTS.md: %w", err)
	}

	// Get the specify section content for this agent
	specifySection := getSpecifySectionContent(aiAssistant)

	// Check for malformed sections
	openCount := strings.Count(string(content), "<specify>")
	closeCount := strings.Count(string(content), "</specify>")

	if openCount > 1 || closeCount > 1 {
		return "", false, fmt.Errorf(
			"malformed delimited section: multiple delimited sections found",
		)
	}

	if openCount != closeCount {
		return "", false, fmt.Errorf(
			"malformed delimited section: mismatched opening and closing tags",
		)
	}

	var updatedContent string
	if openCount == 1 {
		// Replace existing section
		updatedContent = specifySectionRe.ReplaceAllString(string(content), specifySection)
	} else {
		// Add new section at the end
		updatedContent = string(content) + "\n" + specifySection + "\n"
	}

	// Write updated content
	err = os.WriteFile(agentsPath, []byte(updatedContent), 0o644)
	if err != nil {
		return "", false, fmt.Errorf("failed to write updated AGENTS.md: %w", err)
	}

	return agentsPath, false, nil
}

// getAgentsMDTemplate returns the complete template content for AGENTS.md from unified template file
func getAgentsMDTemplate(aiAssistant string) string {
	// Use new template location in content/ directory
	templatePath := "templates/content/agents-template.md"

	if content, err := os.ReadFile(templatePath); err == nil {
		// Process template with Go template engine
		data := template.Data{
			AIAssistant: aiAssistant,
		}
		if processedContent, err := processTemplate(string(content), data); err == nil {
			return processedContent
		}
	}

	// If no template file found, return default content with agent-specific paths
	return fmt.Sprintf(
		"# Agent Instructions\n\nThis file contains instructions for AI agents working with the spec-kit project.\n\n<specify>\n## Specify Commands\n\nThe following slash commands are available in this spec-driven development environment.\nFor detailed usage and examples of any command, see the corresponding documentation file in `.%s/commands/<command>.md`.\n\n### Built-in Commands\n\n**`/specify`** - Creates a new feature specification and branch  \nStart the spec-driven development lifecycle by creating a specification from your feature description.\n\n**`/plan`** - Creates an implementation plan from a feature specification  \nSecond phase: convert specification into implementation plan with research, design docs, and contracts.\n\n**`/tasks`** - Breaks down the implementation plan into executable tasks  \nThird phase: generate numbered, ordered tasks for implementation following TDD methodology.\n\n### Command Flow\n1. `/specify <description>` → Creates spec.md and feature branch\n2. `/plan` → Creates plan.md, research.md, contracts/, data-model.md, quickstart.md  \n3. `/tasks` → Creates tasks.md with numbered implementation tasks\n\n### Documentation Structure\n- Each command has detailed documentation at `.%s/commands/<command>.md`\n- Additional commands can be added by creating corresponding documentation files\n- Command documentation includes usage examples, parameters, and expected outputs\n</specify>",
		aiAssistant,
		aiAssistant,
	)
}

// getSpecifySectionContent returns the content for the <specify> delimited section from template file
func getSpecifySectionContent(aiAssistant string) string {
	// Use new template location in content/ directory
	templatePath := "templates/content/agents-template.md"

	if content, err := os.ReadFile(templatePath); err == nil {
		// Extract the <specify>...</specify> section from template content
		if match := specifySectionRe.FindString(string(content)); match != "" {
			// Process template with Go template engine
			data := template.Data{
				AIAssistant: aiAssistant,
			}
			if processedMatch, err := processTemplate(match, data); err == nil {
				return processedMatch
			}
		}
	}

	// If no template file found or no delimited section found, return default section with agent-specific paths
	return fmt.Sprintf(
		"<specify>\n## Specify Commands\n\nThe following slash commands are available in this spec-driven development environment.\nFor detailed usage and examples of any command, see the corresponding documentation file in `.%s/commands/<command>.md`.\n\n### Built-in Commands\n\n**`/specify`** - Creates a new feature specification and branch  \nStart the spec-driven development lifecycle by creating a specification from your feature description.\n\n**`/plan`** - Creates an implementation plan from a feature specification  \nSecond phase: convert specification into implementation plan with research, design docs, and contracts.\n\n**`/tasks`** - Breaks down the implementation plan into executable tasks  \nThird phase: generate numbered, ordered tasks for implementation following TDD methodology.\n\n### Command Flow\n1. `/specify <description>` → Creates spec.md and feature branch\n2. `/plan` → Creates plan.md, research.md, contracts/, data-model.md, quickstart.md  \n3. `/tasks` → Creates tasks.md with numbered implementation tasks\n\n### Documentation Structure\n- Each command has detailed documentation at `.%s/commands/<command>.md`\n- Additional commands can be added by creating corresponding documentation files\n- Command documentation includes usage examples, parameters, and expected outputs\n</specify>",
		aiAssistant,
		aiAssistant,
	)
}

// CreateOrUpdateClaudeMD creates or updates CLAUDE.md file with reference to AGENTS.md
func CreateOrUpdateClaudeMD(projectRoot string) (filePath string, created bool, err error) {
	if projectRoot == "" {
		return "", false, fmt.Errorf("invalid project root directory: path cannot be empty")
	}

	// Check if project root exists
	if _, statErr := os.Stat(projectRoot); os.IsNotExist(statErr) {
		return "", false, fmt.Errorf("invalid project root directory: %s", projectRoot)
	}

	claudePath := filepath.Join(projectRoot, "CLAUDE.md")
	specifySection := "<specify>you MUST follow the RULES in AGENTS.md</specify>"

	// Check if file exists
	if _, statErr := os.Stat(claudePath); os.IsNotExist(statErr) {
		// File doesn't exist, create it
		content := "# Claude Instructions\n\nThis file contains specific instructions for Claude Code.\n\n" + specifySection + "\n"
		err = os.WriteFile(claudePath, []byte(content), 0o644)
		if err != nil {
			return "", false, fmt.Errorf("failed to write CLAUDE.md: %w", err)
		}
		return claudePath, true, nil
	}

	// File exists, read and update it
	content, err := os.ReadFile(claudePath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read existing CLAUDE.md: %w", err)
	}

	// Check for malformed sections
	openCount := strings.Count(string(content), "<specify>")
	closeCount := strings.Count(string(content), "</specify>")

	if openCount > 1 || closeCount > 1 {
		return "", false, fmt.Errorf(
			"malformed delimited section: multiple delimited sections found",
		)
	}

	if openCount != closeCount {
		return "", false, fmt.Errorf(
			"malformed delimited section: mismatched opening and closing tags",
		)
	}

	var updatedContent string
	if openCount == 1 {
		// Replace existing section
		updatedContent = specifySectionRe.ReplaceAllString(string(content), specifySection)
	} else {
		// Add new section at the end
		updatedContent = string(content) + "\n" + specifySection + "\n"
	}

	// Write updated content
	err = os.WriteFile(claudePath, []byte(updatedContent), 0o644)
	if err != nil {
		return "", false, fmt.Errorf("failed to write updated CLAUDE.md: %w", err)
	}

	return claudePath, false, nil
}
