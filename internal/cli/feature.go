package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/euforicio/spec-kit/internal/services"
)

var featureCmd = &cobra.Command{
	Use:   "feature",
	Short: "Manage spec-driven development features",
	Long: `Feature management commands for spec-driven development workflow.
	
This command group provides tools to create, plan, and manage features
following the spec-driven development methodology.`,
}

var featureCreateCmd = &cobra.Command{
	Use:   "create <description>",
	Short: "Create a new feature branch and directory structure",
	Long: `Create a new feature with numbered branch, directory structure, and template.

This command will:
1. Find the next available feature number
2. Create a new branch named with the number and description
3. Create the feature directory structure in specs/
4. Copy the spec template if available

Examples:
  specify feature create "user authentication system"
  specify feature create "add payment processing"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runFeatureCreate,
}

var featurePlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Setup implementation plan structure for current branch",
	Long: `Setup implementation plan structure for the current feature branch.

This command will:
1. Verify you're on a valid feature branch
2. Create the feature directory if it doesn't exist
3. Copy the plan template if available
4. Return paths for LLM use

Must be run from a feature branch (format: 001-feature-name).`,
	RunE: runFeaturePlan,
}

var featureCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check feature prerequisites and find available documents",
	Long: `Check that implementation plan exists and find optional design documents.

This command will:
1. Verify you're on a valid feature branch
2. Check that the feature directory exists
3. Verify the implementation plan exists
4. List available design documents

Must be run from a feature branch with existing plan.md.`,
	RunE: runFeatureCheck,
}

var featureContextCmd = &cobra.Command{
	Use:   "context [agent]",
	Short: "Update agent context files based on feature plan",
	Long: `Update AI agent context files based on the current feature plan.

Supported agents: claude, gemini, copilot
If no agent is specified, updates all existing agent context files.

This command will:
1. Read the current feature plan
2. Extract technology information
3. Update the appropriate agent context files
4. Preserve manual additions in context files

Examples:
  specify feature context           # Update all existing files
  specify feature context claude   # Update only CLAUDE.md
  specify feature context gemini   # Update only GEMINI.md`,
	Args: cobra.MaximumNArgs(1),
	RunE: runFeatureContext,
}

var featurePathsCmd = &cobra.Command{
	Use:   "paths",
	Short: "Get paths for current feature branch",
	Long: `Get all relevant paths for the current feature branch without creating anything.

This command outputs:
- Repository root
- Current branch name
- Feature directory
- Feature spec file
- Implementation plan file
- Tasks file

Must be run from a valid feature branch.`,
	RunE: runFeaturePaths,
}


func init() {
	// Add feature subcommands
	featureCmd.AddCommand(featureCreateCmd)
	featureCmd.AddCommand(featurePlanCmd)
	featureCmd.AddCommand(featureCheckCmd)
	featureCmd.AddCommand(featureContextCmd)
	featureCmd.AddCommand(featurePathsCmd)

	// Add flags
	featureCreateCmd.Flags().Bool("json", false, "Output results in JSON format")
	featurePlanCmd.Flags().Bool("json", false, "Output results in JSON format")
	featureCheckCmd.Flags().Bool("json", false, "Output results in JSON format")
}

func runFeatureCreate(cmd *cobra.Command, args []string) error {
	description := strings.Join(args, " ")
	
	filesystem := services.NewFilesystemService()
	git := services.NewGitService()
	feature := services.NewFeatureService(filesystem, git)

	result, err := feature.CreateFeature(description)
	if err != nil {
		return fmt.Errorf("failed to create feature: %w", err)
	}

	jsonOutput, err := cmd.Flags().GetBool("json")
	if err != nil {
		return fmt.Errorf("failed to get 'json' flag: %w", err)
	}
	if jsonOutput {
		if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
			return fmt.Errorf("failed to write json output: %w", err)
		}
	} else {
		fmt.Printf("BRANCH_NAME: %s\n", result.BranchName)
		fmt.Printf("SPEC_FILE: %s\n", result.SpecFile)
		fmt.Printf("FEATURE_NUM: %s\n", result.FeatureNum)
	}

	return nil
}

func runFeaturePlan(cmd *cobra.Command, args []string) error {
	filesystem := services.NewFilesystemService()
	git := services.NewGitService()
	feature := services.NewFeatureService(filesystem, git)

	result, err := feature.SetupPlan()
	if err != nil {
		return fmt.Errorf("failed to setup plan: %w", err)
	}

	jsonOutput, err := cmd.Flags().GetBool("json")
	if err != nil {
		return fmt.Errorf("failed to get 'json' flag: %w", err)
	}
	if jsonOutput {
		if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
			return fmt.Errorf("failed to write json output: %w", err)
		}
	} else {
		fmt.Printf("FEATURE_SPEC: %s\n", result.FeatureSpec)
		fmt.Printf("IMPL_PLAN: %s\n", result.ImplPlan)
		fmt.Printf("SPECS_DIR: %s\n", result.SpecsDir)
		fmt.Printf("BRANCH: %s\n", result.Branch)
	}

	return nil
}

func runFeatureCheck(cmd *cobra.Command, args []string) error {
	filesystem := services.NewFilesystemService()
	git := services.NewGitService()
	feature := services.NewFeatureService(filesystem, git)

	result, err := feature.CheckPrerequisites()
	if err != nil {
		return fmt.Errorf("failed to check prerequisites: %w", err)
	}

	jsonOutput, err := cmd.Flags().GetBool("json")
	if err != nil {
		return fmt.Errorf("failed to get 'json' flag: %w", err)
	}
	if jsonOutput {
		if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
			return fmt.Errorf("failed to write json output: %w", err)
		}
	} else {
		fmt.Printf("FEATURE_DIR:%s\n", result.FeatureDir)
		fmt.Println("AVAILABLE_DOCS:")
		for _, doc := range result.AvailableDocs {
			fmt.Printf("  ✓ %s\n", doc)
		}
	}

	return nil
}

func runFeatureContext(cmd *cobra.Command, args []string) error {
	agentType := ""
	if len(args) > 0 {
		agentType = args[0]
	}

	filesystem := services.NewFilesystemService()
	git := services.NewGitService()
	feature := services.NewFeatureService(filesystem, git)

	// Validate agent type using service
	if err := feature.ValidateAgentType(agentType); err != nil {
		return err
	}

	result, err := feature.UpdateContext(agentType)
	if err != nil {
		return fmt.Errorf("failed to update context: %w", err)
	}

	fmt.Printf("=== Updating agent context files for feature %s ===\n", result.Branch)
	
	for _, update := range result.Updates {
		fmt.Printf("✅ %s context file updated successfully\n", update.Agent)
	}

	if len(result.Summary) > 0 {
		fmt.Println("\nSummary of changes:")
		for _, change := range result.Summary {
			fmt.Printf("- %s\n", change)
		}
	}

	return nil
}

func runFeaturePaths(cmd *cobra.Command, args []string) error {
	filesystem := services.NewFilesystemService()
	git := services.NewGitService()
	feature := services.NewFeatureService(filesystem, git)

	result, err := feature.GetPaths()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	fmt.Printf("REPO_ROOT: %s\n", result.RepoRoot)
	fmt.Printf("BRANCH: %s\n", result.Branch)
	fmt.Printf("FEATURE_DIR: %s\n", result.FeatureDir)
	fmt.Printf("FEATURE_SPEC: %s\n", result.FeatureSpec)
	fmt.Printf("IMPL_PLAN: %s\n", result.ImplPlan)
	fmt.Printf("TASKS: %s\n", result.Tasks)

	return nil
}

