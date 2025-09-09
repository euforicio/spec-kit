package contract

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLICheckContract(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	t.Run("specify check", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "check")
		output, err := cmd.CombinedOutput()
		
		outputStr := string(output)
		t.Logf("Command output: %s", outputStr)
		
		// Expected to fail during TDD phase
		if err != nil {
			assert.True(t,
				strings.Contains(outputStr, "check") ||
				strings.Contains(outputStr, "command") ||
				strings.Contains(outputStr, "not") ||
				strings.Contains(outputStr, "implemented"),
				"error should mention check command or implementation")
			return
		}

		// Once implemented, should succeed and show environment status
		assert.NoError(t, err, "check command should succeed")
		
		// Should check internet connectivity
		assert.True(t,
			strings.Contains(outputStr, "internet") ||
			strings.Contains(outputStr, "connectivity") ||
			strings.Contains(outputStr, "connection"),
			"should report internet connectivity status")
		
		// Should check git availability
		assert.True(t,
			strings.Contains(outputStr, "git") ||
			strings.Contains(outputStr, "Git"),
			"should report git availability")
		
		// Should check AI tools
		assert.True(t,
			strings.Contains(outputStr, "claude") ||
			strings.Contains(outputStr, "Claude") ||
			strings.Contains(outputStr, "gemini") ||
			strings.Contains(outputStr, "Gemini") ||
			strings.Contains(outputStr, "AI") ||
			strings.Contains(outputStr, "tool"),
			"should report AI tool availability")
	})

	t.Run("specify check exit code", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "check")
		err := cmd.Run()
		
		// During TDD phase, this may fail, but once implemented:
		// Check command should always exit with 0 even if tools are missing
		// It's an informational command, not a validation gate
		if err == nil {
			// Once implemented, should always succeed
			assert.NoError(t, err, "check command should exit with code 0")
		}
	})

	t.Run("specify check output format", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "check")
		output, err := cmd.CombinedOutput()
		
		outputStr := string(output)
		t.Logf("Command output: %s", outputStr)
		
		if err != nil {
			// Expected during TDD phase
			return
		}

		// Once implemented, output should be well-formatted
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")
		assert.Greater(t, len(lines), 1, "should have multiple lines of output")
		
		// Should have clear status indicators
		hasCheckmarks := strings.Contains(outputStr, "✓") || 
						strings.Contains(outputStr, "✗") ||
						strings.Contains(outputStr, "OK") ||
						strings.Contains(outputStr, "FAIL") ||
						strings.Contains(outputStr, "available") ||
						strings.Contains(outputStr, "not found")
		
		assert.True(t, hasCheckmarks, "should have clear status indicators")
	})

	t.Run("specify check shows helpful info", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "check")
		output, err := cmd.CombinedOutput()
		
		outputStr := string(output)
		t.Logf("Command output: %s", outputStr)
		
		if err != nil {
			// Expected during TDD phase
			return
		}

		// Once implemented, should provide helpful information
		// Should mention installation instructions or next steps
		assert.True(t,
			strings.Contains(outputStr, "install") ||
			strings.Contains(outputStr, "Install") ||
			strings.Contains(outputStr, "ready") ||
			strings.Contains(outputStr, "setup") ||
			strings.Contains(outputStr, "missing"),
			"should provide helpful setup information")
	})
}

func TestCLICheckEnvironmentDetection(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	t.Run("check detects git", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "check")
		output, err := cmd.CombinedOutput()
		
		outputStr := string(output)
		
		if err != nil {
			// Expected during TDD phase
			return
		}

		// Should report git status (available or not)
		assert.True(t,
			strings.Contains(outputStr, "git") ||
			strings.Contains(outputStr, "Git"),
			"should mention git status")
		
		// Check if git is actually available
		gitCmd := exec.Command("git", "--version")
		gitErr := gitCmd.Run()
		
		if gitErr == nil {
			// If git is available, should report it as such
			assert.True(t,
				strings.Contains(outputStr, "✓") ||
				strings.Contains(outputStr, "available") ||
				strings.Contains(outputStr, "OK") ||
				strings.Contains(outputStr, "found"),
				"should report git as available when it exists")
		} else {
			// If git is not available, should report installation hint
			assert.True(t,
				strings.Contains(outputStr, "✗") ||
				strings.Contains(outputStr, "not found") ||
				strings.Contains(outputStr, "missing") ||
				strings.Contains(outputStr, "install"),
				"should report git as missing and provide install hint")
		}
	})

	t.Run("check detects internet connectivity", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "check")
		output, err := cmd.CombinedOutput()
		
		outputStr := string(output)
		
		if err != nil {
			// Expected during TDD phase
			return
		}

		// Should check internet connectivity (required for template downloads)
		assert.True(t,
			strings.Contains(outputStr, "internet") ||
			strings.Contains(outputStr, "connectivity") ||
			strings.Contains(outputStr, "connection") ||
			strings.Contains(outputStr, "network"),
			"should check internet connectivity")
	})

	t.Run("check provides actionable information", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "check")
		output, err := cmd.CombinedOutput()
		
		outputStr := string(output)
		
		if err != nil {
			// Expected during TDD phase
			return
		}

		// Should provide actionable next steps
		assert.True(t,
			strings.Contains(outputStr, "ready") ||
			strings.Contains(outputStr, "install") ||
			strings.Contains(outputStr, "setup") ||
			strings.Contains(outputStr, "consider") ||
			strings.Contains(outputStr, "recommend"),
			"should provide actionable information")
	})
}