package models

import (
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

// Precompiled regex for <specify>...</specify> sections to avoid recompilation
var specifySectionRe = regexp.MustCompile(`(?s)<specify>.*?</specify>`)

// Codex-specific AGENTS.md file operations
// This file contains temporary code to support OpenAI Codex's slash command requirements.
// TODO: Remove this file when Codex supports slash commands natively.

// CreateOrUpdateAgentsMD creates or updates AGENTS.md file with static template content
func CreateOrUpdateAgentsMD(projectRoot string) (filePath string, created bool, err error) {
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
		filePath, err = CreateAgentsMD(projectRoot)
		if err != nil {
			return "", false, fmt.Errorf("failed to create AGENTS.md: %w", err)
		}
		return filePath, true, nil
	}

	// File exists, update it
	filePath, _, err = UpdateAgentsMD(projectRoot)
	if err != nil {
		return "", false, fmt.Errorf("failed to update AGENTS.md: %w", err)
	}
	return filePath, false, nil
}

// CreateAgentsMD creates a new AGENTS.md file with static template content
func CreateAgentsMD(projectRoot string) (string, error) {
	agentsPath := filepath.Join(projectRoot, "AGENTS.md")
	
	// Get static template content
	content := getAgentsMDTemplate()
	
	// Write file
	err := os.WriteFile(agentsPath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write AGENTS.md: %w", err)
	}
	
	return agentsPath, nil
}

// UpdateAgentsMD updates existing AGENTS.md file with static template content
func UpdateAgentsMD(projectRoot string) (string, bool, error) {
	agentsPath := filepath.Join(projectRoot, "AGENTS.md")
	
	// Read existing content
	content, err := os.ReadFile(agentsPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read existing AGENTS.md: %w", err)
	}
	
	// Update content with new delimited section
	updatedContent, err := replaceDelimitedSection(string(content))
	if err != nil {
		return "", false, fmt.Errorf("malformed delimited section: %w", err)
	}
	
	// Write updated content
	err = os.WriteFile(agentsPath, []byte(updatedContent), 0644)
	if err != nil {
		return "", false, fmt.Errorf("failed to write updated AGENTS.md: %w", err)
	}
	
	return agentsPath, true, nil
}

// replaceDelimitedSection replaces or adds the <specify>...</specify> section
func replaceDelimitedSection(content string) (string, error) {
    specifySection := getSpecifySectionContent()
    
    // Check if delimited section exists
    
    // Check for malformed sections
    openCount := strings.Count(content, "<specify>")
    closeCount := strings.Count(content, "</specify>")
	
	if openCount > 1 || closeCount > 1 {
		return "", fmt.Errorf("multiple delimited sections found")
	}
	
	if openCount != closeCount {
		return "", fmt.Errorf("mismatched opening and closing tags")
	}
	
    if openCount == 1 {
        // Replace existing section
        return specifySectionRe.ReplaceAllString(content, specifySection), nil
    }
	
	// Add new section at the end
	return content + "\n" + specifySection + "\n", nil
}

// getAgentsMDTemplate returns the complete static template content for AGENTS.md from template file
func getAgentsMDTemplate() string {
	templatePath := filepath.Join("templates", "agents-md-content.md")
	content, err := os.ReadFile(templatePath)
	if err != nil {
		// If template file is not found, return a minimal fallback
		return fmt.Sprintf("# Agent Instructions\n\nThis file contains instructions for AI agents working with the spec-kit project.\n\nError: Could not load template from %s: %v", templatePath, err)
	}
	return string(content)
}

// getSpecifySectionContent returns the content for the <specify> delimited section from template file
func getSpecifySectionContent() string {
	templatePath := filepath.Join("templates", "agents-md-content.md")
	content, err := os.ReadFile(templatePath)
	if err != nil {
		// If template file is not found, return a minimal error section
		return fmt.Sprintf("<specify>\nError: Could not load template from %s: %v\n</specify>", templatePath, err)
	}
	
    // Extract the <specify>...</specify> section from template content
    match := specifySectionRe.FindString(string(content))
	if match == "" {
		// If no delimited section found, return error section
		return "<specify>\nError: No <specify> section found in template file\n</specify>"
	}
	
	return match
}
