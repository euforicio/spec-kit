# Data Model: Codex Native Integration

## Core Entities

### CommandFile
Represents a command instruction file for Codex to execute.

**Fields:**
- `Name` (string): Command name (e.g., "specify", "plan")
- `Path` (string): File path relative to .codex/commands/
- `Content` (string): Markdown content with instructions
- `Triggers` ([]string): Slash commands that activate this file
- `Patterns` ([]string): Natural language patterns that map to this command
- `RequiredArgs` ([]string): Arguments the command expects
- `OptionalArgs` ([]string): Optional arguments

**Validation Rules:**
- Name must be alphanumeric with hyphens
- Content must have required sections (When, Execute, Report)
- Path must exist when loaded

### ContextDirectory
Collection of templates and configuration for Codex.

**Fields:**
- `Path` (string): Root path (.codex/)
- `Templates` (map[string]Template): Available templates
- `Commands` ([]CommandFile): Loaded command files
- `InitFile` (string): Path to CODEX_INIT.md
- `Examples` ([]Example): Example workflows

**Validation Rules:**
- Path must be writable
- InitFile must exist after initialization
- At least one command file required

### InitializationConfig
Settings for creating the .codex directory.

**Fields:**
- `ProjectRoot` (string): Project root directory
- `IncludeExamples` (bool): Whether to include example files
- `CommandSet` (string): Which commands to include (minimal, standard, full)
- `AIAssistant` (string): Target assistant (codex, claude, both)
- `OverwriteExisting` (bool): Whether to overwrite existing files

**Validation Rules:**
- ProjectRoot must exist
- CommandSet must be valid option
- AIAssistant must be supported

### CommandRegistry
Mapping of commands to their handlers.

**Fields:**
- `Commands` (map[string]CommandFile): Command name to file mapping
- `Aliases` (map[string]string): Alternative names for commands
- `NaturalPatterns` (map[string]string): Natural language to command mapping
- `LoadedAt` (time.Time): When registry was loaded

**Validation Rules:**
- No duplicate command names
- Aliases must map to existing commands
- Patterns must be unique

### Template
Reusable template file for specifications and plans.

**Fields:**
- `Name` (string): Template identifier
- `Path` (string): File path relative to context/templates/
- `Content` (string): Template markdown content
- `Placeholders` ([]string): Variables to replace
- `Type` (string): Template type (spec, plan, task, etc.)

**Validation Rules:**
- Content must be valid Markdown
- Placeholders must follow [PLACEHOLDER] format
- Type must be recognized

### Example
Example workflow or usage pattern.

**Fields:**
- `Name` (string): Example identifier
- `Description` (string): What this example demonstrates
- `Content` (string): Example content/workflow
- `Tags` ([]string): Categories for filtering

**Validation Rules:**
- Content must be executable/followable
- At least one tag required

## Enumerations

### CommandSet
```go
type CommandSet string

const (
    CommandSetMinimal  CommandSet = "minimal"  // Core commands only
    CommandSetStandard CommandSet = "standard" // Common workflow commands
    CommandSetFull     CommandSet = "full"     // All available commands
)
```

### TemplateType
```go
type TemplateType string

const (
    TemplateTypeSpec  TemplateType = "spec"
    TemplateTypePlan  TemplateType = "plan"
    TemplateTypeTask  TemplateType = "task"
    TemplateTypeOther TemplateType = "other"
)
```

### AIAssistant
```go
type AIAssistant string

const (
    AssistantCodex  AIAssistant = "codex"
    AssistantClaude AIAssistant = "claude"
    AssistantBoth   AIAssistant = "both"
)
```

## Relationships

1. **ContextDirectory** → **CommandFile** (1:many)
   - A context directory contains multiple command files

2. **ContextDirectory** → **Template** (1:many)
   - A context directory contains multiple templates

3. **CommandRegistry** → **CommandFile** (1:1)
   - Registry maps to specific command files

4. **ContextDirectory** → **Example** (1:many)
   - Context can include multiple examples

## File Structure

The data model maps to this file structure:
```
.codex/
├── CODEX_INIT.md                 # Initialization instructions
├── commands/                      # Command files
│   ├── specify.md
│   ├── plan.md
│   ├── tasks.md
│   ├── test.md
│   └── implement.md
├── context/
│   ├── templates/                # Reusable templates
│   │   ├── spec-template.md
│   │   ├── plan-template.md
│   │   └── tasks-template.md
│   └── constitution.md           # Project principles
└── examples/                      # Usage examples
    ├── full-workflow.md
    └── quick-start.md
```

## Operations

### Initialize
```go
func Initialize(config InitializationConfig) error
```
Creates the .codex directory structure with all files.

### LoadCommand
```go
func LoadCommand(name string) (*CommandFile, error)
```
Loads a specific command file from disk.

### RegisterCommands
```go
func RegisterCommands(dir string) (*CommandRegistry, error)
```
Scans directory and builds command registry.

### ExecuteCommand
```go
func ExecuteCommand(command string, args []string) (string, error)
```
Processes a command and returns instructions for Codex.

## Constraints

1. **Atomicity**: Directory creation must be all-or-nothing
2. **Idempotency**: Re-initialization should be safe
3. **Compatibility**: Command files must work for multiple AI assistants
4. **Versioning**: Track command file format version for upgrades
5. **Size**: Keep command files under 10KB for quick loading