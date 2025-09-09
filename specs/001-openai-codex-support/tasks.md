# Tasks: OpenAI Codex Native Integration

**Input**: Design documents from `/specs/001-openai-codex-support/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → Extract: Go 1.25, cobra framework, extend context package
2. Load optional design documents:
   → data-model.md: 6 entities → model tasks
   → contracts/: CLI interface → contract test tasks
   → research.md: Technical decisions → implementation approach
3. Generate tasks by category:
   → Setup: Extend existing context command
   → Tests: Contract tests for feature context codex
   → Core: AGENTS.md generation, command file creation
   → Integration: Command file templates
   → Polish: Documentation, validation
4. Apply task rules:
   → Different files = mark [P] for parallel
   → Same file = sequential (no [P])
   → Tests before implementation (TDD)
5. Number tasks sequentially (T001, T002...)
6. Ready for execution
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Phase 3.1: Setup
- [ ] T001 Verify existing context command structure in cmd/feature_context.go
- [ ] T002 Create embedded templates directory structure at internal/codex/templates/
- [ ] T003 [P] Add codex agent type to internal/context/agent_types.go

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**

- [ ] T004 [P] Contract test for 'specify feature context codex' in tests/contract/context_codex_test.go
- [ ] T005 [P] Test AGENTS.md generation in tests/integration/agents_md_test.go
- [ ] T006 [P] Test command file creation in tests/integration/command_files_test.go
- [ ] T007 [P] Test incremental updates to AGENTS.md in tests/integration/agents_update_test.go

## Phase 3.3: Core Implementation (ONLY after tests are failing)

### Data Models
- [ ] T008 [P] CommandFile struct in internal/codex/models.go
- [ ] T009 [P] AgentsConfig struct for AGENTS.md content in internal/codex/models.go
- [ ] T010 [P] CommandTemplate struct in internal/codex/models.go

### Core Functions
- [ ] T011 Generate AGENTS.md content in internal/codex/agents.go
- [ ] T012 Create command file templates in internal/codex/commands.go
- [ ] T013 Embed default command files in internal/codex/templates/commands/
- [ ] T014 Write command files to .codex/commands/ in internal/codex/writer.go

### CLI Integration
- [ ] T015 Add 'codex' case to feature context command in cmd/feature_context.go
- [ ] T016 Implement generateCodexContext function in cmd/feature_context.go
- [ ] T017 Handle --force, --minimal, --no-commands flags in cmd/feature_context.go

## Phase 3.4: Command File Templates

### Core Command Files
- [ ] T018 [P] Create specify.md template in internal/codex/templates/commands/
- [ ] T019 [P] Create plan.md template in internal/codex/templates/commands/
- [ ] T020 [P] Create tasks.md template in internal/codex/templates/commands/
- [ ] T021 [P] Create test.md template in internal/codex/templates/commands/
- [ ] T022 [P] Create implement.md template in internal/codex/templates/commands/
- [ ] T023 [P] Create validate.md template in internal/codex/templates/commands/
- [ ] T024 [P] Create commit.md template in internal/codex/templates/commands/

### Additional Commands
- [ ] T025 [P] Create review.md template in internal/codex/templates/commands/
- [ ] T026 [P] Create help.md template in internal/codex/templates/commands/
- [ ] T027 [P] Create status.md template in internal/codex/templates/commands/

## Phase 3.5: Integration

### File Operations
- [ ] T028 Ensure .codex directory creation with proper permissions
- [ ] T029 Handle existing AGENTS.md merging/updating
- [ ] T030 Implement atomic file writes for safety

### Error Handling
- [ ] T031 Add proper error messages for missing project
- [ ] T032 Handle permission errors gracefully
- [ ] T033 Validate command file templates on write

## Phase 3.6: Polish

### Unit Tests
- [ ] T034 [P] Unit tests for AGENTS.md generation in internal/codex/agents_test.go
- [ ] T035 [P] Unit tests for command file creation in internal/codex/commands_test.go
- [ ] T036 [P] Unit tests for template embedding in internal/codex/templates_test.go

### Documentation
- [ ] T037 [P] Update README.md with Codex integration instructions
- [ ] T038 [P] Add Codex section to CONTRIBUTING.md
- [ ] T039 [P] Create examples/codex-workflow.md

### Validation
- [ ] T040 Run quickstart.md validation checklist
- [ ] T041 Test with actual Codex CLI
- [ ] T042 Verify slash commands work as expected

## Dependencies
- Setup (T001-T003) before tests
- Tests (T004-T007) must fail before implementation
- Models (T008-T010) before core functions (T011-T014)
- Core functions before CLI integration (T015-T017)
- CLI integration before command templates can be tested
- Everything before polish phase

## Parallel Execution Examples

### Launch model creation together (T008-T010):
```bash
Task: "CommandFile struct in internal/codex/models.go"
Task: "AgentsConfig struct in internal/codex/models.go"
Task: "CommandTemplate struct in internal/codex/models.go"
```

### Launch command template creation together (T018-T027):
```bash
Task: "Create specify.md template"
Task: "Create plan.md template"
Task: "Create tasks.md template"
# ... etc
```

## Notes
- Extend existing context command rather than creating new subcommand
- AGENTS.md is Codex's standard file (auto-loaded on startup)
- Command files use same format as Claude commands
- Tests must fail first (TDD requirement)

## Task Count Summary
- Setup: 3 tasks
- Tests: 4 tasks
- Models: 3 tasks
- Core Functions: 4 tasks
- CLI Integration: 3 tasks
- Command Templates: 10 tasks
- Integration: 6 tasks
- Polish: 9 tasks
**Total: 42 tasks**