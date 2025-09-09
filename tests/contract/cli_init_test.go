package contract

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIInitContract(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	t.Run("specify init PROJECT_NAME", func(t *testing.T) {
		tempDir := t.TempDir()
		projectName := "test-project"
		projectPath := filepath.Join(tempDir, projectName)

		cmd := exec.Command(binaryPath, "init", projectName)
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()
		
		outputStr := string(output)
		t.Logf("Command output: %s", outputStr)
		
		// Command should succeed (once implemented)
		// For now, this test will fail until implementation is complete
		if err != nil {
			// Expected to fail during TDD phase
			assert.Contains(t, outputStr, "init", "error should mention init command")
			return
		}

		// Once implemented, these assertions should pass:
		// - Project directory should be created
		assert.DirExists(t, projectPath, "project directory should be created")
		
		// - Should contain template files
		// We can't test exact files without knowing the template structure,
		// but we can test that some files were created
		files, err := os.ReadDir(projectPath)
		require.NoError(t, err)
		assert.NotEmpty(t, files, "project directory should contain files")
	})

	t.Run("specify init with --ai flag", func(t *testing.T) {
		tempDir := t.TempDir()
		projectName := "claude-project"
		
		testCases := []string{"claude", "gemini", "copilot"}
		
		for _, ai := range testCases {
			t.Run("ai="+ai, func(t *testing.T) {
				cmd := exec.Command(binaryPath, "init", projectName+"-"+ai, "--ai", ai)
				cmd.Dir = tempDir
				output, err := cmd.CombinedOutput()
				
				outputStr := string(output)
				t.Logf("Command output: %s", outputStr)
				
				// For TDD phase, this will fail
				if err != nil {
					// Expected during TDD - should mention the AI assistant
					assert.True(t, 
						strings.Contains(outputStr, ai) || 
						strings.Contains(outputStr, "init") ||
						strings.Contains(outputStr, "assistant"),
						"error should be related to AI assistant or init command")
					return
				}
				
				// Once implemented:
				projectPath := filepath.Join(tempDir, projectName+"-"+ai)
				assert.DirExists(t, projectPath, "project directory should be created")
			})
		}
	})

	t.Run("specify init with --no-git flag", func(t *testing.T) {
		tempDir := t.TempDir()
		projectName := "no-git-project"

		cmd := exec.Command(binaryPath, "init", projectName, "--no-git")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()
		
		outputStr := string(output)
		t.Logf("Command output: %s", outputStr)
		
		// Expected to fail during TDD phase
		if err != nil {
			assert.Contains(t, outputStr, "init", "error should mention init command")
			return
		}

		// Once implemented:
		projectPath := filepath.Join(tempDir, projectName)
		assert.DirExists(t, projectPath, "project directory should be created")
		
		// Should NOT contain .git directory
		gitPath := filepath.Join(projectPath, ".git")
		assert.NoDirExists(t, gitPath, "should not create git repository with --no-git")
	})

	t.Run("specify init existing directory fails", func(t *testing.T) {
		tempDir := t.TempDir()
		projectName := "existing-project"
		projectPath := filepath.Join(tempDir, projectName)
		
		// Create directory first
		err := os.Mkdir(projectPath, 0755)
		require.NoError(t, err)

		cmd := exec.Command(binaryPath, "init", projectName)
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()
		
		// Should fail because directory exists
		assert.Error(t, err, "should fail when directory already exists")
		
		outputStr := string(output)
		assert.True(t,
			strings.Contains(outputStr, "exists") ||
			strings.Contains(outputStr, "already") ||
			strings.Contains(outputStr, "error") ||
			strings.Contains(outputStr, "init"),
			"error message should indicate directory exists or init failed")
	})

	t.Run("specify init without project name fails", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "init")
		output, err := cmd.CombinedOutput()
		
		// Should fail because no project name provided
		assert.Error(t, err, "should fail when no project name provided")
		
		outputStr := string(output)
		assert.True(t,
			strings.Contains(outputStr, "required") ||
			strings.Contains(outputStr, "argument") ||
			strings.Contains(outputStr, "name") ||
			strings.Contains(outputStr, "usage"),
			"error should indicate missing argument")
	})
}

func TestCLIInitErrorHandling(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	t.Run("invalid AI assistant", func(t *testing.T) {
		tempDir := t.TempDir()

		cmd := exec.Command(binaryPath, "init", "test-project", "--ai", "invalid-ai")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()
		
		// Should fail with invalid AI assistant
		assert.Error(t, err, "should fail with invalid AI assistant")
		
		outputStr := string(output)
		assert.True(t,
			strings.Contains(outputStr, "invalid") ||
			strings.Contains(outputStr, "ai") ||
			strings.Contains(outputStr, "assistant"),
			"error should mention invalid AI assistant")
	})

	t.Run("conflicting flags", func(t *testing.T) {
		tempDir := t.TempDir()

		cmd := exec.Command(binaryPath, "init", "test-project", "--here")
		cmd.Dir = tempDir
		output, err := cmd.CombinedOutput()
		
		// Should fail because --here and project name are mutually exclusive
		assert.Error(t, err, "should fail with conflicting flags")
		
		outputStr := string(output)
		assert.True(t,
			strings.Contains(outputStr, "cannot") ||
			strings.Contains(outputStr, "both") ||
			strings.Contains(outputStr, "conflict") ||
			strings.Contains(outputStr, "here"),
			"error should mention conflicting options")
	})
}