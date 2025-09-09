# Implementation Plan: OpenAI Codex Native Integration

**Branch**: `001-openai-codex-support` | **Date**: 2025-09-09 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/Users/euforic/go/src/github.com/euforicio/spec-kit/specs/001-openai-codex-support/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
4. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
5. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, or `GEMINI.md` for Gemini CLI).
6. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
7. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
8. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
Extend the existing `specify feature context` command to support Codex as an AI assistant type, generating or updating AGENTS.md in the project root (Codex's standard context file) and creating .codex/commands/ directory with command files. This enables developers to work entirely within Codex using slash commands while maintaining specify's workflow.

## Technical Context
**Language/Version**: Go 1.25  
**Primary Dependencies**: cobra (CLI framework), embed (for bundled files)  
**Storage**: File system (.codex directory structure)  
**Testing**: Go testing framework, testify  
**Target Platform**: MacOS/Linux/Windows CLI
**Project Type**: single - CLI enhancement  
**Performance Goals**: <100ms for command file loading  
**Constraints**: Must work with existing Codex CLI, no external API calls  
**Scale/Scope**: Single-user development tool, ~20 command files

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Simplicity**:
- Projects: 1 (CLI enhancement with tests)
- Using framework directly? Yes (cobra, no wrappers)
- Single data model? Yes (command files and config)
- Avoiding patterns? Yes (direct file operations)

**Architecture**:
- EVERY feature as library? Yes (extends existing context package)
- Libraries listed: 
  - `context`: Extended to support codex agent type
  - `commands`: Command file templates for .codex directory
- CLI per library: `specify feature context codex`
- Library docs: llms.txt format planned? Yes

**Testing (NON-NEGOTIABLE)**:
- RED-GREEN-Refactor cycle enforced? Yes
- Git commits show tests before implementation? Yes
- Order: Contract→Integration→E2E→Unit strictly followed? Yes
- Real dependencies used? Yes (actual file system)
- Integration tests for: new libraries, contract changes, shared schemas? Yes
- FORBIDDEN: Implementation before test, skipping RED phase

**Observability**:
- Structured logging included? Yes (slog)
- Frontend logs → backend? N/A (CLI only)
- Error context sufficient? Yes

**Versioning**:
- Version number assigned? 0.2.0
- BUILD increments on every change? Yes
- Breaking changes handled? Yes (backward compatible)

## Project Structure

### Documentation (this feature)
```
specs/[###-feature]/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
# Option 1: Single project (DEFAULT)
src/
├── models/
├── services/
├── cli/
└── lib/

tests/
├── contract/
├── integration/
└── unit/

# Option 2: Web application (when "frontend" + "backend" detected)
backend/
├── src/
│   ├── models/
│   ├── services/
│   └── api/
└── tests/

frontend/
├── src/
│   ├── components/
│   ├── pages/
│   └── services/
└── tests/

# Option 3: Mobile + API (when "iOS/Android" detected)
api/
└── [same as backend above]

ios/ or android/
└── [platform-specific structure]
```

**Structure Decision**: Option 1 (Single project - CLI enhancement)

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:
   ```
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:
   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Generate contract tests** from contracts:
   - One test file per endpoint
   - Assert request/response schemas
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:
   - Each story → integration test scenario
   - Quickstart test = story validation steps

5. **Update agent file incrementally** (O(1) operation):
   - Run `specify feature context [claude|gemini|copilot]` for your AI assistant
   - If exists: Add only NEW tech from current plan
   - Preserve manual additions between markers
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency
   - Output to repository root

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, agent-specific file

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- Load `/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs:
  - Each CLI command (init, validate, list, add, update) → contract test [P]
  - Each entity (CommandFile, ContextDirectory, etc.) → model task [P]
  - Each command file → template creation task
  - Integration tests for Codex workflow
- Implementation tasks to make tests pass

**Ordering Strategy**:
1. Setup: Directory structure, embedded files
2. Tests: Contract tests for CLI commands (TDD)
3. Models: Entity implementations [P]
4. Core: Initialize, validate, command loading
5. Commands: CLI subcommand implementations
6. Templates: Command file templates
7. Integration: Full workflow tests

**Task Breakdown Preview**:
- Setup tasks: 3-4 (structure, dependencies)
- Contract tests: 5 (one per CLI command) [P]
- Model tasks: 6 (entities from data-model.md) [P]
- Implementation: 8-10 (core functionality)
- Command files: 10-12 (templates for Codex)
- Integration: 3-4 (workflow validation)

**Estimated Output**: 35-40 numbered, ordered tasks in tasks.md

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |


## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [x] Complexity deviations documented (none needed)

---
*Based on Constitution v2.1.1 - See `memory/constitution.md`*
