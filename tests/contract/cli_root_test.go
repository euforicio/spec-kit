package contract

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIRootContract(t *testing.T) {
	// Build the CLI binary for testing
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	t.Run("specify --version", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "--version")
		output, err := cmd.CombinedOutput()
		
		require.NoError(t, err, "command should succeed")
		outputStr := string(output)
		
		// Should contain version information
		assert.Contains(t, outputStr, "specify", "output should contain program name")
		assert.Contains(t, outputStr, "version", "output should contain version info")
	})

	t.Run("specify -v", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "-v")
		output, err := cmd.CombinedOutput()
		
		require.NoError(t, err, "short version flag should succeed")
		outputStr := string(output)
		
		// Should contain version information
		assert.Contains(t, outputStr, "specify", "output should contain program name")
		assert.Contains(t, outputStr, "version", "output should contain version info")
	})

	t.Run("specify --help", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "--help")
		output, err := cmd.CombinedOutput()
		
		require.NoError(t, err, "help command should succeed")
		outputStr := string(output)
		
		// Should contain usage information
		assert.Contains(t, outputStr, "specify", "help should contain program name")
		assert.Contains(t, outputStr, "Usage:", "help should contain usage section")
		assert.Contains(t, outputStr, "init", "help should mention init command")
		assert.Contains(t, outputStr, "check", "help should mention check command")
	})

	t.Run("specify -h", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "-h")
		output, err := cmd.CombinedOutput()
		
		require.NoError(t, err, "short help flag should succeed")
		outputStr := string(output)
		
		// Should contain usage information
		assert.Contains(t, outputStr, "specify", "help should contain program name")
		assert.Contains(t, outputStr, "Usage:", "help should contain usage section")
	})

	t.Run("specify (no args)", func(t *testing.T) {
		cmd := exec.Command(binaryPath)
		output, _ := cmd.CombinedOutput()
		
		// Should either show help or banner, and exit successfully or with code 1
		outputStr := string(output)
		
		// Should show something helpful to the user
		assert.True(t, 
			strings.Contains(outputStr, "specify") || 
			strings.Contains(outputStr, "Usage:") ||
			strings.Contains(outputStr, "SPECIFY"),
			"no-args invocation should show helpful information")
	})
}

func TestCLIExitCodes(t *testing.T) {
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	t.Run("valid commands exit 0", func(t *testing.T) {
		testCases := [][]string{
			{"--help"},
			{"-h"},
			{"--version"},
			{"-v"},
		}

		for _, args := range testCases {
			t.Run(strings.Join(args, " "), func(t *testing.T) {
				cmd := exec.Command(binaryPath, args...)
				err := cmd.Run()
				
				assert.NoError(t, err, "valid command should exit with code 0")
			})
		}
	})

	t.Run("invalid commands exit non-zero", func(t *testing.T) {
		testCases := [][]string{
			{"--invalid-flag"},
			{"invalid-command"},
			{"init"}, // This should fail without proper args
		}

		for _, args := range testCases {
			t.Run(strings.Join(args, " "), func(t *testing.T) {
				cmd := exec.Command(binaryPath, args...)
				err := cmd.Run()
				
				assert.Error(t, err, "invalid command should exit with non-zero code")
			})
		}
	})
}

// buildTestBinary builds the CLI binary for contract testing
func buildTestBinary(t *testing.T) string {
	t.Helper()
	
	tmpDir := t.TempDir()
	binaryPath := tmpDir + "/specify-test"
	
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/specify")
	cmd.Dir = "../../"
	
	err := cmd.Run()
	require.NoError(t, err, "failed to build test binary")
	
	return binaryPath
}