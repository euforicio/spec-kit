# /tasks

## Description
Generate task breakdown from the implementation plan.

## Usage
When user types: `/tasks`

## Prerequisites
- Must be on a feature branch
- Implementation plan must exist (plan.md)
- Design documents should be available (research.md, data-model.md, contracts/)

## Execute
1. Run the check command:
   ```bash
   specify feature check --json
   ```
   
2. Parse JSON output:
   - Extract: `feature_dir`
   - Extract: `available_docs`
   
3. Load plan and design documents:
   - Read: {feature_dir}/plan.md
   - Read: {feature_dir}/data-model.md (if exists)
   - Read: {feature_dir}/contracts/* (if exists)
   - Read: {feature_dir}/research.md (if exists)
   
4. Load task template:
   - Load: @templates/tasks-template.md
   
5. Generate tasks based on documents:
   - Setup tasks: Project initialization, dependencies
   - Test tasks [P]: One per contract, one per integration scenario
   - Model tasks [P]: One per entity from data-model.md
   - Service tasks: Core functionality implementation
   - Integration tasks: Connections, middleware, logging
   - Polish tasks [P]: Unit tests, documentation
   
6. Apply task rules:
   - Mark parallel tasks with [P] (different files)
   - Order by TDD: Tests before implementation
   - Number sequentially: T001, T002, etc.
   
7. Save tasks:
   - Write to: {feature_dir}/tasks.md
   
8. Summary:
   - Count total tasks
   - Identify parallel execution opportunities
   - Show task categories

## Report
Success: "✓ Generated {total} tasks in tasks.md
         - {setup_count} setup tasks
         - {test_count} test tasks (TDD)
         - {impl_count} implementation tasks
         - {polish_count} polish tasks
         Tasks marked [P] can run in parallel"
Error: "Failed to generate tasks: {error}"

## Examples
User: /tasks
Result: ✓ Generated 42 tasks in tasks.md
        - 3 setup tasks
        - 8 test tasks (TDD)
        - 24 implementation tasks
        - 7 polish tasks
        Tasks marked [P] can run in parallel

## Natural Language
Patterns that trigger this command:
- "break down tasks" → /tasks
- "create tasks" → /tasks
- "generate task list" → /tasks
- "what tasks do I need" → /tasks

## Notes
Tasks follow TDD principles - tests must be written and fail before implementation. Tasks marked with [P] can be executed in parallel as they work on different files. The task list provides a complete roadmap for implementing the feature.