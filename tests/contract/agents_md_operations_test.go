package contract

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/euforicio/spec-kit/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Contract test for AGENTS.md operations based on agents-md-operations.md contract
func TestAgentsMDOperations_CreateOrUpdateAgentsMD(t *testing.T) {
	tests := []struct {
		name            string
		projectRoot     string
		existingContent string
		expectedCreated bool
		expectError     bool
		expectedErrMsg  string
	}{
		{
			name:            "create new AGENTS.md file",
			projectRoot:     createTempDir(t),
			existingContent: "",
			expectedCreated: true,
			expectError:     false,
		},
		{
			name:            "update existing AGENTS.md file",
			projectRoot:     createTempDirWithAgentsMD(t, "# Existing Content\nSome content here\n"),
			existingContent: "# Existing Content\nSome content here\n",
			expectedCreated: false,
			expectError:     false,
		},
		{
			name:           "invalid project root",
			projectRoot:    "/nonexistent/path",
			expectError:    true,
			expectedErrMsg: "invalid project root directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test MUST fail initially - CreateOrUpdateAgentsMD doesn't exist yet
			filePath, created, err := models.CreateOrUpdateAgentsMD(tt.projectRoot)
			
			if tt.expectError {
				assert.Error(t, err, "Expected error")
				assert.Contains(t, strings.ToLower(err.Error()), tt.expectedErrMsg, "Error message should contain expected text")
				return
			}
			
			assert.NoError(t, err, "Should not have error")
			assert.Equal(t, tt.expectedCreated, created, "Created flag mismatch")
			
			expectedPath := filepath.Join(tt.projectRoot, "AGENTS.md")
			assert.Equal(t, expectedPath, filePath, "File path mismatch")
			
			// Verify file exists
			assert.FileExists(t, filePath, "AGENTS.md should exist")
			
			// Read and verify content
			content, err := os.ReadFile(filePath)
			require.NoError(t, err, "Should be able to read AGENTS.md")
			
			contentStr := string(content)
			
			// Should contain delimited section
			assert.Contains(t, contentStr, "<specify>", "Should contain opening delimiter")
			assert.Contains(t, contentStr, "</specify>", "Should contain closing delimiter")
			
			// Should contain slash command documentation
			assert.Contains(t, contentStr, "/specify", "Should contain /specify command")
			assert.Contains(t, contentStr, "/plan", "Should contain /plan command")
			assert.Contains(t, contentStr, "/tasks", "Should contain /tasks command")
			
			// Should reference .codex/commands/ directory
			assert.Contains(t, contentStr, ".codex/commands/", "Should reference .codex/commands/ directory")
			
			// If updating existing file, should preserve content outside delimiters
			if !tt.expectedCreated && tt.existingContent != "" {
				assert.Contains(t, contentStr, "# Existing Content", "Should preserve existing content")
			}
			
			// Cleanup
			defer os.RemoveAll(tt.projectRoot)
		})
	}
}

// Test delimited section detection and replacement
func TestAgentsMDOperations_DelimitedSectionHandling(t *testing.T) {
	tempDir := createTempDir(t)
	defer os.RemoveAll(tempDir)
	
	// Create file with existing delimited section
	existingContent := `# My Custom AGENTS.md

Custom instructions here.

<specify>
Old specify content that should be replaced
</specify>

More custom content after.
`
	
	agentsPath := filepath.Join(tempDir, "AGENTS.md")
	err := os.WriteFile(agentsPath, []byte(existingContent), 0644)
	require.NoError(t, err)
	
	// This test MUST fail initially - CreateOrUpdateAgentsMD doesn't exist yet
	filePath, created, err := models.CreateOrUpdateAgentsMD(tempDir)
	
	assert.NoError(t, err, "Should not have error")
	assert.False(t, created, "Should be update, not create")
	assert.Equal(t, agentsPath, filePath)
	
	// Read updated content
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	contentStr := string(content)
	
	// Should preserve content outside delimiters
	assert.Contains(t, contentStr, "# My Custom AGENTS.md", "Should preserve header")
	assert.Contains(t, contentStr, "Custom instructions here.", "Should preserve custom content")
	assert.Contains(t, contentStr, "More custom content after.", "Should preserve trailing content")
	
	// Should replace content within delimiters
	assert.NotContains(t, contentStr, "Old specify content", "Should replace old delimited content")
	assert.Contains(t, contentStr, "/specify", "Should contain new specify documentation")
}

// Helper functions
func createTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "agents-md-test-*")
	require.NoError(t, err)
	return dir
}

func createTempDirWithAgentsMD(t *testing.T, content string) string {
	t.Helper()
	dir := createTempDir(t)
	agentsPath := filepath.Join(dir, "AGENTS.md")
	err := os.WriteFile(agentsPath, []byte(content), 0644)
	require.NoError(t, err)
	return dir
}