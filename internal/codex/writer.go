package codex

import (
	"fmt"
	"os"
	"path/filepath"
)

// Writer handles writing files for Codex integration
type Writer struct {
	projectRoot string
}

// NewWriter creates a new file writer
func NewWriter(projectRoot string) *Writer {
	return &Writer{
		projectRoot: projectRoot,
	}
}

// WriteAGENTSmd writes or updates the AGENTS.md file
func (w *Writer) WriteAGENTSmd(content string) error {
	agentsPath := filepath.Join(w.projectRoot, "AGENTS.md")
	
	// Write atomically by writing to temp file first
	tempPath := agentsPath + ".tmp"
	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	
	// Rename temp file to actual file (atomic on most systems)
	if err := os.Rename(tempPath, agentsPath); err != nil {
		// Clean up temp file if rename fails
		os.Remove(tempPath)
		return fmt.Errorf("failed to update AGENTS.md: %w", err)
	}
	
	return nil
}

// WriteCommandFiles writes command files to .codex/commands/
func (w *Writer) WriteCommandFiles(commands []CommandFile, force bool) error {
	commandsDir := filepath.Join(w.projectRoot, ".codex", "commands")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return fmt.Errorf("failed to create commands directory: %w", err)
	}
	
	for _, cmd := range commands {
		cmdPath := filepath.Join(commandsDir, cmd.Path)
		
		// Check if file exists and skip if not forcing
		if !force {
			if _, err := os.Stat(cmdPath); err == nil {
				continue // File exists, skip
			}
		}
		
		// Write command file
		if err := w.writeFile(cmdPath, cmd.Content); err != nil {
			return fmt.Errorf("failed to write command %s: %w", cmd.Name, err)
		}
	}
	
	return nil
}

// writeFile writes content to a file atomically
func (w *Writer) writeFile(path, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Write to temp file first
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	
	// Rename to actual file
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename file: %w", err)
	}
	
	return nil
}

// EnsureCodexDirectory ensures .codex directory structure exists
func (w *Writer) EnsureCodexDirectory() error {
	codexDir := filepath.Join(w.projectRoot, ".codex")
	commandsDir := filepath.Join(codexDir, "commands")
	
	// Create directories
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return fmt.Errorf("failed to create .codex directory structure: %w", err)
	}
	
	return nil
}