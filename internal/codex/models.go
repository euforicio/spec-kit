package codex

// CommandSet represents the set of commands to install
type CommandSet string

const (
	// CommandSetMinimal includes only core commands
	CommandSetMinimal CommandSet = "minimal"
	// CommandSetStandard includes common workflow commands
	CommandSetStandard CommandSet = "standard"
	// CommandSetFull includes all available commands
	CommandSetFull CommandSet = "full"
)

// CommandFile represents a command instruction file for Codex
type CommandFile struct {
	Name         string   // Command name (e.g., "specify", "plan")
	Path         string   // File path relative to .codex/commands/
	Content      string   // Markdown content with instructions
	Triggers     []string // Slash commands that activate this file
	Patterns     []string // Natural language patterns that map to this command
	Description  string   // Brief description of the command
	RequiredArgs []string // Arguments the command expects
	OptionalArgs []string // Optional arguments
}

// AgentsConfig represents the configuration for AGENTS.md content
type AgentsConfig struct {
	ExistingContent []byte            // Existing AGENTS.md content to merge
	PlanContent     string            // Plan.md content to extract tech info from
	ProjectContext  map[string]string // Extracted project context
	Commands        []CommandFile     // Available commands to document
}

// CommandTemplate represents a template for generating command files
type CommandTemplate struct {
	Name        string // Template name
	Content     string // Template content
	CommandSet  CommandSet // Which command set this belongs to
}

// CommandDescription holds command metadata for AGENTS.md
type CommandDescription struct {
	Command     string   // The slash command (e.g., "/specify")
	Description string   // What the command does
	Patterns    []string // Natural language patterns
}

// StandardCommands defines the standard command descriptions
var StandardCommands = []CommandDescription{
	{
		Command:     "/specify",
		Description: "Create a new feature specification",
		Patterns:    []string{"create a spec for", "start a new feature", "specify"},
	},
	{
		Command:     "/plan",
		Description: "Plan implementation for current feature",
		Patterns:    []string{"plan this", "create implementation plan", "plan the feature"},
	},
	{
		Command:     "/tasks",
		Description: "Generate task breakdown",
		Patterns:    []string{"break down tasks", "create tasks", "generate tasks"},
	},
	{
		Command:     "/test",
		Description: "Write tests following TDD",
		Patterns:    []string{"write tests for", "create test for", "test"},
	},
	{
		Command:     "/implement",
		Description: "Implement code to pass tests",
		Patterns:    []string{"implement", "write code for", "create implementation"},
	},
	{
		Command:     "/validate",
		Description: "Validate against specification",
		Patterns:    []string{"validate", "check implementation", "verify"},
	},
	{
		Command:     "/commit",
		Description: "Create git commit",
		Patterns:    []string{"commit changes", "create commit", "save changes"},
	},
}

// MinimalCommands defines the minimal command set
var MinimalCommands = []string{"specify", "plan", "tasks"}

// GetCommandsForSet returns the commands for a given set
func GetCommandsForSet(set CommandSet) []string {
	switch set {
	case CommandSetMinimal:
		return MinimalCommands
	case CommandSetFull:
		// Return all command names
		var all []string
		for _, cmd := range StandardCommands {
			name := cmd.Command[1:] // Remove leading slash
			all = append(all, name)
		}
		// Add additional commands for full set
		all = append(all, "review", "help", "status")
		return all
	default: // CommandSetStandard
		var standard []string
		for _, cmd := range StandardCommands {
			name := cmd.Command[1:] // Remove leading slash
			standard = append(standard, name)
		}
		return standard
	}
}