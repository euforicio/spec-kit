# Research: Codex Native Integration

## Command File Format

**Decision**: Use Markdown with structured sections for command instructions  
**Rationale**: Codex understands Markdown well, human-readable, easy to edit  
**Alternatives considered**:
- JSON: Too rigid, harder for humans to edit
- YAML: Parsing complexity for Codex
- Plain text: Lacks structure

## Directory Structure

**Decision**: Mirror .claude structure but with Codex-specific additions  
**Rationale**: Consistency across AI assistants, familiar pattern  
**Structure**:
```
.codex/
├── CODEX_INIT.md       # Initial instructions for Codex
├── commands/           # Command files (shared with Claude)
├── context/           
│   ├── templates/      # Spec templates
│   └── constitution.md # Project principles
└── examples/           # Codex-specific examples
```

## Command Discovery Mechanism

**Decision**: Slash command triggers file lookup in .codex/commands/  
**Rationale**: Simple pattern matching, no complex parsing needed  
**Implementation**:
- `/specify` → reads `.codex/commands/specify.md`
- `/plan` → reads `.codex/commands/plan.md`
- Natural language fallback for unmatched commands

## Template Embedding Strategy

**Decision**: Use Go embed package to bundle default command files  
**Rationale**: Single binary distribution, no external files needed  
**Implementation**:
```go
//go:embed templates/commands/*
var commandTemplates embed.FS
```

## Command File Instructions Format

**Decision**: Structured sections with clear execution steps  
**Rationale**: Codex can follow step-by-step instructions reliably  
**Format**:
```markdown
# /command Name

When user types: `/command <args>`

Execute these steps:
1. Run: `shell command`
2. Parse: Extract JSON fields
3. Read: @file/path
4. Write: Save to path
5. Report: Success message
```

## Natural Language Recognition

**Decision**: Include pattern matching section in CODEX_INIT.md  
**Rationale**: Allows flexibility when users don't use slash commands  
**Patterns**:
- "create a spec for..." → `/specify`
- "plan this" → `/plan`
- "break down tasks" → `/tasks`

## File Access Pattern

**Decision**: Use @ notation for file references (Codex native)  
**Rationale**: Codex already supports this pattern  
**Examples**:
- `@.codex/context/templates/spec-template.md`
- `@specs/001-feature/spec.md`

## Command Execution Flow

**Decision**: Commands execute specify CLI and parse output  
**Rationale**: Leverages existing CLI functionality  
**Flow**:
1. Codex reads command file
2. Executes specify CLI command
3. Parses JSON output
4. Performs additional operations
5. Reports results

## Error Handling

**Decision**: Graceful degradation with helpful messages  
**Rationale**: Guide users to fix issues  
**Scenarios**:
- Missing .codex directory: Suggest running `specify codex init`
- Command not found: Fall back to natural language
- Template missing: Use default from embedded files

## Initialization Process

**Decision**: `specify codex init` creates complete .codex structure  
**Rationale**: One-time setup, everything ready  
**Creates**:
1. Directory structure
2. CODEX_INIT.md with instructions
3. Command files from templates
4. Context files and examples

## Command File Sharing

**Decision**: Same command files work for Claude and Codex  
**Rationale**: Maintain once, use everywhere  
**Compatibility**:
- Use common Markdown format
- Avoid assistant-specific features
- Clear, universal instructions

## Testing Strategy

**Decision**: Test command file readability and execution  
**Rationale**: Ensure Codex can understand and follow instructions  
**Tests**:
1. Unit: Command file parsing
2. Integration: Directory creation
3. E2E: Full Codex workflow simulation

## Security Considerations

**Decision**: Read-only access to command files, no code execution  
**Rationale**: Prevent injection attacks  
**Safeguards**:
- Command files are templates, not executable
- Codex executes only predefined CLI commands
- No arbitrary shell command execution

## Performance Optimization

**Decision**: Load command files on demand, cache in Codex context  
**Rationale**: Fast response, minimal overhead  
**Strategy**:
- Lazy loading of command files
- CODEX_INIT.md loaded once per session
- Templates cached after first use

## Backward Compatibility

**Decision**: Enhance existing specify without breaking changes  
**Rationale**: Smooth adoption path  
**Approach**:
- New `codex` subcommand
- Existing commands unchanged
- Optional feature activation

## Documentation Strategy

**Decision**: Self-documenting command files with examples  
**Rationale**: Documentation lives with code  
**Includes**:
- Usage examples in each command file
- Success/error scenarios
- Common variations

## Summary

All technical decisions have been made to create a seamless Codex-native experience. The system uses familiar patterns (slash commands, @ notation) while maintaining specify's workflow principles. Command files are self-contained, shareable between AI assistants, and bundled with the CLI for easy distribution.