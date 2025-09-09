package codex

import (
	"fmt"
	"regexp"
	"strings"
)

// AgentsGenerator handles generation and updating of AGENTS.md
type AgentsGenerator struct {
	commands []CommandDescription
}

// NewAgentsGenerator creates a new AGENTS.md generator
func NewAgentsGenerator() *AgentsGenerator {
	return &AgentsGenerator{
		commands: StandardCommands,
	}
}

// Generate creates or updates AGENTS.md content
func (g *AgentsGenerator) Generate(planContent string, existingContent []byte) (string, error) {
	var builder strings.Builder
	
	// Parse existing content if provided
	existingStr := string(existingContent)
	
	// If we have existing content, merge with it
	if existingStr != "" {
		return g.mergeContent(existingStr, planContent)
	}
	
	// Create new AGENTS.md
	builder.WriteString("# AGENTS.md\n\n")
	builder.WriteString("This file provides context for AI assistants when working with this repository.\n\n")
	
	// Add Codex section
	builder.WriteString(g.generateCodexSection())
	
	// Add project context
	builder.WriteString(g.generateProjectContext(planContent))
	
	return builder.String(), nil
}

// mergeContent merges new Codex content with existing AGENTS.md using delimited sections
func (g *AgentsGenerator) mergeContent(existing, planContent string) (string, error) {
	const (
		codexStartMarker = "<!-- specify-codex-start -->"
		codexEndMarker   = "<!-- specify-codex-end -->"
	)
	
	// Look for existing Codex section
	startIdx := strings.Index(existing, codexStartMarker)
	if startIdx != -1 {
		// Find end marker
		endIdx := strings.Index(existing, codexEndMarker)
		if endIdx == -1 {
			return "", fmt.Errorf("found codex start marker but no end marker")
		}
		endIdx += len(codexEndMarker)
		
		// Replace existing section
		newCodex := g.generateDelimitedCodexSection()
		updated := existing[:startIdx] + newCodex + existing[endIdx:]
		return updated, nil
	}
	
	// No existing section, append new delimited section
	result := strings.TrimSpace(existing)
	if result != "" {
		result += "\n\n"
	}
	result += g.generateDelimitedCodexSection()
	
	return result, nil
}

// generateDelimitedCodexSection creates a Codex section wrapped in HTML comments for easy identification
func (g *AgentsGenerator) generateDelimitedCodexSection() string {
	codexContent := g.generateCodexSection()
	return fmt.Sprintf("<!-- specify-codex-start -->\n%s\n<!-- specify-codex-end -->", codexContent)
}

// generateCodexSection generates the Codex-specific content
func (g *AgentsGenerator) generateCodexSection() string {
	var builder strings.Builder
	
	builder.WriteString("## For OpenAI Codex\n\n")
	
	// Command System
	builder.WriteString("### Command System\n\n")
	builder.WriteString("When you see a message starting with \"/\", check .codex/commands/ for the matching command file.\n")
	builder.WriteString("Examples:\n")
	for _, cmd := range g.commands[:3] { // Show first 3 as examples
		name := cmd.Command[1:] // Remove leading slash
		builder.WriteString(fmt.Sprintf("- `%s \"<args>\"` → Read .codex/commands/%s.md\n", cmd.Command, name))
	}
	builder.WriteString("\n")
	
	// Available Commands
	builder.WriteString("### Available Commands\n\n")
	for _, cmd := range g.commands {
		builder.WriteString(fmt.Sprintf("- `%s` - %s\n", cmd.Command, cmd.Description))
	}
	builder.WriteString("\n")
	
	// Natural Language Patterns
	builder.WriteString("### Natural Language Patterns\n\n")
	builder.WriteString("You can also use natural language:\n")
	for _, cmd := range g.commands {
		if len(cmd.Patterns) > 0 {
			builder.WriteString(fmt.Sprintf("- \"%s\" → %s\n", cmd.Patterns[0], cmd.Command))
		}
	}
	builder.WriteString("\n")
	
	// Workflow
	builder.WriteString("### Workflow\n\n")
	builder.WriteString("Follow the specify workflow:\n")
	builder.WriteString("1. Create specification (`/specify`)\n")
	builder.WriteString("2. Plan implementation (`/plan`)\n")
	builder.WriteString("3. Generate tasks (`/tasks`)\n")
	builder.WriteString("4. Write tests (`/test`)\n")
	builder.WriteString("5. Implement code (`/implement`)\n")
	builder.WriteString("6. Validate (`/validate`)\n")
	builder.WriteString("7. Commit changes (`/commit`)\n\n")
	
	return builder.String()
}

// generateProjectContext generates the project context section
func (g *AgentsGenerator) generateProjectContext(planContent string) string {
	var builder strings.Builder
	
	builder.WriteString("## Project Context\n\n")
	
	if planContent != "" {
		// Extract tech info from plan
		techInfo := g.extractTechInfo(planContent)
		
		if techInfo["language"] != "" {
			builder.WriteString(fmt.Sprintf("- **Language**: %s\n", techInfo["language"]))
		}
		if techInfo["dependencies"] != "" {
			builder.WriteString(fmt.Sprintf("- **Dependencies**: %s\n", techInfo["dependencies"]))
		}
		if techInfo["storage"] != "" {
			builder.WriteString(fmt.Sprintf("- **Storage**: %s\n", techInfo["storage"]))
		}
		if techInfo["testing"] != "" {
			builder.WriteString(fmt.Sprintf("- **Testing**: %s\n", techInfo["testing"]))
		}
		if techInfo["projectType"] != "" {
			builder.WriteString(fmt.Sprintf("- **Project Type**: %s\n", techInfo["projectType"]))
		}
	} else {
		builder.WriteString("*Project context will be updated from plan.md*\n")
	}
	
	builder.WriteString("\n")
	return builder.String()
}

// extractTechInfo extracts technology information from plan content
func (g *AgentsGenerator) extractTechInfo(planContent string) map[string]string {
	info := make(map[string]string)
	
	// Define patterns to extract
	patterns := map[string]*regexp.Regexp{
		"language":     regexp.MustCompile(`\*\*Language/Version\*\*:\s*(.+)`),
		"dependencies": regexp.MustCompile(`\*\*Primary Dependencies\*\*:\s*(.+)`),
		"storage":      regexp.MustCompile(`\*\*Storage\*\*:\s*(.+)`),
		"testing":      regexp.MustCompile(`\*\*Testing\*\*:\s*(.+)`),
		"projectType":  regexp.MustCompile(`\*\*Project Type\*\*:\s*(.+)`),
	}
	
	for key, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(planContent); len(matches) > 1 {
			info[key] = strings.TrimSpace(matches[1])
		}
	}
	
	return info
}