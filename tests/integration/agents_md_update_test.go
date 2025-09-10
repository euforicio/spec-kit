package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/euforicio/spec-kit/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration test for AGENTS.md file update scenarios
func TestAgentsMDUpdate_ExistingFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agents-md-update-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create existing AGENTS.md with custom content
	agentsPath := filepath.Join(tempDir, "AGENTS.md")
	existingContent := `# My Custom Agent Instructions

This is my custom content that should be preserved.

## Custom Section
Important custom instructions here.

<specify>
Old specify content that should be replaced.
</specify>

## More Custom Content
Additional instructions that should remain.
`

	err = os.WriteFile(agentsPath, []byte(existingContent), 0644)
	require.NoError(t, err)

	// This test MUST fail initially - UpdateAgentsMD function doesn't exist yet
	filePath, updated, err := models.UpdateAgentsMD(tempDir)
	
	assert.NoError(t, err, "AGENTS.md update should succeed")
	assert.True(t, updated, "Should report file was updated")
	assert.Equal(t, agentsPath, filePath, "Returned path should match expected")
	
	// Verify file still exists
	assert.FileExists(t, agentsPath, "AGENTS.md should still exist")
	
	// Read updated content
	content, err := os.ReadFile(agentsPath)
	require.NoError(t, err)
	contentStr := string(content)
	
	// Should preserve custom content outside delimited section
	assert.Contains(t, contentStr, "# My Custom Agent Instructions", "Should preserve custom header")
	assert.Contains(t, contentStr, "This is my custom content that should be preserved.", "Should preserve custom intro")
	assert.Contains(t, contentStr, "## Custom Section", "Should preserve custom section header")
	assert.Contains(t, contentStr, "Important custom instructions here.", "Should preserve custom section content")
	assert.Contains(t, contentStr, "## More Custom Content", "Should preserve trailing custom section")
	assert.Contains(t, contentStr, "Additional instructions that should remain.", "Should preserve trailing custom content")
	
	// Should replace content within delimited section
	assert.NotContains(t, contentStr, "Old specify content that should be replaced.", "Should replace old delimited content")
	assert.Contains(t, contentStr, "/specify", "Should have new /specify documentation")
	assert.Contains(t, contentStr, "/plan", "Should have new /plan documentation")
	assert.Contains(t, contentStr, "/tasks", "Should have new /tasks documentation")
	assert.Contains(t, contentStr, ".codex/commands/", "Should reference .codex/commands/ directory")
}

// Test updating file with no existing delimited section
func TestAgentsMDUpdate_NoExistingDelimitedSection(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agents-md-no-delimited-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create AGENTS.md without any delimited section
	agentsPath := filepath.Join(tempDir, "AGENTS.md")
	existingContent := `# My Agent Instructions

Custom content without any specify section.

## Custom Instructions
Do this and that.
`

	err = os.WriteFile(agentsPath, []byte(existingContent), 0644)
	require.NoError(t, err)

	// This test MUST fail initially - UpdateAgentsMD function doesn't exist yet
	filePath, updated, err := models.UpdateAgentsMD(tempDir)
	
	assert.NoError(t, err, "AGENTS.md update should succeed")
	assert.True(t, updated, "Should report file was updated")
	
	// Read updated content
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	contentStr := string(content)
	
	// Should preserve existing content
	assert.Contains(t, contentStr, "# My Agent Instructions", "Should preserve existing header")
	assert.Contains(t, contentStr, "Custom content without any specify section.", "Should preserve existing content")
	assert.Contains(t, contentStr, "## Custom Instructions", "Should preserve existing section")
	assert.Contains(t, contentStr, "Do this and that.", "Should preserve existing instructions")
	
	// Should add new delimited section
	assert.Contains(t, contentStr, "<specify>", "Should add opening delimiter")
	assert.Contains(t, contentStr, "</specify>", "Should add closing delimiter")
	assert.Contains(t, contentStr, "/specify", "Should add specify documentation")
}

// Test updating file with malformed delimited section
func TestAgentsMDUpdate_MalformedDelimitedSection(t *testing.T) {
	testCases := []struct {
		name            string
		existingContent string
		expectError     bool
	}{
		{
			name: "missing closing tag",
			existingContent: `# Instructions

<specify>
Some content but no closing tag.

## More content
`,
			expectError: true,
		},
		{
			name: "missing opening tag",
			existingContent: `# Instructions

Some content before.
</specify>

## More content
`,
			expectError: true,
		},
		{
			name: "multiple opening tags",
			existingContent: `# Instructions

<specify>
First section
<specify>
Second section
</specify>
`,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "agents-md-malformed-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			agentsPath := filepath.Join(tempDir, "AGENTS.md")
			err = os.WriteFile(agentsPath, []byte(tc.existingContent), 0644)
			require.NoError(t, err)

			// This test MUST fail initially - UpdateAgentsMD function doesn't exist yet
			_, _, err = models.UpdateAgentsMD(tempDir)
			
			if tc.expectError {
				assert.Error(t, err, "Should fail with malformed delimited section")
				assert.Contains(t, err.Error(), "malformed", "Error should mention malformed section")
			} else {
				assert.NoError(t, err, "Should handle malformed section gracefully")
			}
		})
	}
}

// Test updating with file permission issues
func TestAgentsMDUpdate_PermissionError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tempDir, err := os.MkdirTemp("", "agents-md-update-permission-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create read-only AGENTS.md file
	agentsPath := filepath.Join(tempDir, "AGENTS.md")
	err = os.WriteFile(agentsPath, []byte("# Read Only File"), 0444)
	require.NoError(t, err)

	// Restore permissions for cleanup
	defer func() {
		os.Chmod(agentsPath, 0644)
	}()

	// This test MUST fail initially - UpdateAgentsMD function doesn't exist yet
	_, _, err = models.UpdateAgentsMD(tempDir)
	
	assert.Error(t, err, "Should fail with permission error")
	assert.Contains(t, err.Error(), "permission denied", "Error should mention permission denied")
}

// Test content integrity after update
func TestAgentsMDUpdate_ContentIntegrity(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agents-md-integrity-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create complex existing content
	agentsPath := filepath.Join(tempDir, "AGENTS.md")
	existingContent := `# Complex Agent Instructions

## Section 1
Content before delimited section.

### Subsection 1.1
- Item 1
- Item 2
- Item 3

<specify>
Old content to be replaced.
</specify>

### Subsection 1.2
More content after delimited section.

## Section 2
Final section content.

**Bold text** and *italic text*.

` + "`Code blocks should be preserved`" + `

> Block quotes should be preserved
`

	err = os.WriteFile(agentsPath, []byte(existingContent), 0644)
	require.NoError(t, err)

	// This test MUST fail initially - UpdateAgentsMD function doesn't exist yet
	_, _, err = models.UpdateAgentsMD(tempDir)
	require.NoError(t, err)

	// Read and verify content
	content, err := os.ReadFile(agentsPath)
	require.NoError(t, err)
	contentStr := string(content)

	// Count sections to ensure structure is preserved
	assert.Equal(t, 1, strings.Count(contentStr, "# Complex Agent Instructions"), "Should have exactly one main header")
	assert.Equal(t, 2, strings.Count(contentStr, "## Section"), "Should preserve section headers")
	assert.Equal(t, 2, strings.Count(contentStr, "### Subsection"), "Should preserve subsection headers")
	
	// Check markdown formatting preservation
	assert.Contains(t, contentStr, "**Bold text**", "Should preserve bold formatting")
	assert.Contains(t, contentStr, "*italic text*", "Should preserve italic formatting")
	assert.Contains(t, contentStr, "`Code blocks should be preserved`", "Should preserve code formatting")
	assert.Contains(t, contentStr, "> Block quotes should be preserved", "Should preserve block quotes")
	
	// Check list preservation
	assert.Contains(t, contentStr, "- Item 1", "Should preserve list items")
	assert.Contains(t, contentStr, "- Item 2", "Should preserve list items")
	assert.Contains(t, contentStr, "- Item 3", "Should preserve list items")
}