# CLI Interface Contract: Codex Integration

## Command Structure

### Extended Command: `specify feature context`
Now supports codex as an agent type alongside claude, gemini, and copilot

### Usage

#### `specify feature context codex`
Generate or update AGENTS.md and create .codex/commands/ directory

**Usage:**
```bash
specify feature context codex [flags]
```

**Flags:**
- `--force`: Overwrite existing AGENTS.md and .codex directory
- `--minimal`: Install only core commands
- `--no-commands`: Skip creating .codex/commands/ directory

**Output:**
```
Generating Codex context...
✓ Created/Updated AGENTS.md
✓ Created .codex/commands/ directory
✓ Installed 12 command files
✓ Added project context

Codex is ready! Just start Codex and it will automatically load AGENTS.md
```

**Exit Codes:**
- 0: Success
- 1: Not in a specify project
- 2: Permission denied
- 3: Invalid flags

#### Additional Subcommands (Optional Future Features)

The following commands could be added as the feature matures:

- `specify feature context codex --validate`: Validate existing setup
- `specify feature context codex --update`: Update to latest command templates
- `specify feature context codex --list`: List available commands

## Command File Format

Each command file in `.codex/commands/` follows this structure:

```markdown
# /command-name

## Description
Brief description of what this command does

## Usage
When user types: `/command <args>`

## Execute
1. Run: `specify command --json`
2. Parse: Extract field from JSON
3. Read: @.codex/context/template.md
4. Process: Apply logic
5. Write: Save to path
6. Report: Success message

## Examples
/command "example usage"

## Natural Language
Patterns that trigger this command:
- "do something" → /command
```

## Environment Variables

- `CODEX_DIR`: Override default .codex location
- `SPECIFY_AI_ASSISTANT`: Default assistant for init

## Configuration

Configuration stored in `.codex/config.yaml`:

```yaml
version: "1.0"
assistant: both
command_set: standard
include_examples: true
custom_commands:
  - review
  - deploy
```

## AGENTS.md Structure

The generated or updated AGENTS.md file contains:

```markdown
# AGENTS.md

This file provides context for AI assistants when working with this repository.

## For OpenAI Codex

### Command System

When you see a message starting with "/", check .codex/commands/ for the matching command file.
Examples:
- `/specify "feature"` → Read .codex/commands/specify.md
- `/plan` → Read .codex/commands/plan.md
- `/tasks` → Read .codex/commands/tasks.md

### Available Commands
[List of commands with descriptions]

### Natural Language Patterns
[Patterns that map to commands]

### Workflow
[Specify workflow: spec → plan → tasks → implement]

## Project Context
[Recent changes and current tech stack from plan.md]

## [Other AI Assistant Sections if present]
```

## Integration with Codex

When Codex reads AGENTS.md, it learns:

1. **Command System**: How to interpret slash commands
2. **File Access**: Using @ notation for file references
3. **Workflow**: The specify development process
4. **Natural Language**: Pattern matching for commands

Example Codex session:
```
# Codex starts and automatically loads AGENTS.md
Codex: I see this is a specify project with command system support.
       Available commands loaded from .codex/commands/.

User: /specify "user authentication"
Codex: [Reads .codex/commands/specify.md and executes instructions]
       Created feature branch: 001-user-authentication
       Generated specification at: specs/001-user-authentication/spec.md

User: /plan
Codex: [Reads .codex/commands/plan.md and executes instructions]
       Created implementation plan with research and design artifacts
```

## Error Handling

All commands should handle these scenarios:

1. **Missing .codex**: Suggest running `specify codex init`
2. **Invalid command**: Show available commands with `specify codex list`
3. **Missing templates**: Use embedded defaults
4. **Permission errors**: Clear error message with fix suggestions

## Testing

Contract tests verify:

1. All commands produce expected output format
2. Error codes match documentation
3. Flags work as specified
4. Command files are valid Markdown
5. Natural language patterns map correctly