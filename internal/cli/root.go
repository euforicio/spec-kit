package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	buildTime = "unknown"
	commit    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "specify",
	Short: "Specify CLI - Setup tool for Specify projects",
	Long: `Specify CLI helps you set up spec-driven development projects with your preferred AI assistant.

This tool downloads the latest templates from GitHub and initializes projects with the appropriate
configuration for Claude Code, Gemini CLI, or GitHub Copilot.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle version flag
		if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
			showVersion()
			return
		}

		// Show banner and help when no subcommand is provided
		showBanner()
		fmt.Println("Run 'specify --help' for usage information")
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")

	// Add commands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(featureCmd)
}

func showVersion() {
	fmt.Printf("specify version %s\n", version)
	fmt.Printf("Built: %s\n", buildTime)
	fmt.Printf("Commit: %s\n", commit)
}

func showBanner() {
	banner := `
███████╗██████╗ ███████╗ ██████╗██╗███████╗██╗   ██╗
██╔════╝██╔══██╗██╔════╝██╔════╝██║██╔════╝╚██╗ ██╔╝
███████╗██████╔╝█████╗  ██║     ██║█████╗   ╚████╔╝ 
╚════██║██╔═══╝ ██╔══╝  ██║     ██║██╔══╝    ╚██╔╝  
███████║██║     ███████╗╚██████╗██║██║        ██║   
╚══════╝╚═╝     ╚══════╝ ╚═════╝╚═╝╚═╝        ╚═╝   
`
	fmt.Printf("%s\n", banner)
	fmt.Println("Spec-Driven Development Toolkit")
	fmt.Println()
}
