package codex

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed templates/commands/*
var commandTemplates embed.FS

// CommandCreator handles creation of command files
type CommandCreator struct {
	templates embed.FS
}

// NewCommandCreator creates a new command file creator
func NewCommandCreator() *CommandCreator {
	return &CommandCreator{
		templates: commandTemplates,
	}
}

// CreateCommands creates command files in the .codex/commands directory
func (c *CommandCreator) CreateCommands(projectRoot string, commandSet CommandSet) error {
	return c.createCommands(projectRoot, commandSet, false)
}

// CreateCommandsForce creates command files, overwriting existing ones
func (c *CommandCreator) CreateCommandsForce(projectRoot string, commandSet CommandSet) error {
	return c.createCommands(projectRoot, commandSet, true)
}

func (c *CommandCreator) createCommands(projectRoot string, commandSet CommandSet, force bool) error {
	// Create .codex/commands directory
	commandsDir := filepath.Join(projectRoot, ".codex", "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return fmt.Errorf("failed to create commands directory: %w", err)
	}
	
	// Get commands for the specified set
	commands := GetCommandsForSet(commandSet)
	
	// Create each command file
	for _, cmdName := range commands {
		cmdPath := filepath.Join(commandsDir, cmdName+".md")
		
		// Skip if file exists and not forcing
		if !force {
			if _, err := os.Stat(cmdPath); err == nil {
				continue // File exists, skip
			}
		}
		
		// Read template content
		templatePath := fmt.Sprintf("templates/commands/%s.md", cmdName)
		content, err := c.templates.ReadFile(templatePath)
		if err != nil {
			// If template doesn't exist, use a generic template
			content = c.generateGenericCommand(cmdName)
		}
		
		// Write command file
		if err := os.WriteFile(cmdPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write command file %s: %w", cmdName, err)
		}
	}
	
	return nil
}

// generateGenericCommand generates a generic command template
func (c *CommandCreator) generateGenericCommand(name string) []byte {
	template := fmt.Sprintf(`# /%s

## Description
%s command for specify workflow.

## Usage
When user types: ` + "`/%s <args>`" + `

## Execute
1. Run the specify command:
   ` + "```bash" + `
   specify %s
   ` + "```" + `
   
2. Process the output
   
3. Report results

## Report
Success: "✓ %s completed successfully"
Error: "Failed to execute %s: {error}"

## Examples
User: /%s
Result: Command executed successfully

## Natural Language
Patterns that trigger this command:
- "%s"
- "run %s"
`, name, name, name, name, name, name, name, name, name)
	
	return []byte(template)
}