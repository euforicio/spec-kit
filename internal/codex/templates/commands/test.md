# /test

## Description
Write tests following TDD principles for a specific task.

## Usage
When user types: `/test <task-id>` or `/test <description>`

## Prerequisites
- Tasks.md must exist with numbered tasks
- Must follow TDD: tests must fail first

## Execute
1. Identify the task:
   - If task ID provided (e.g., T006), find in tasks.md
   - If description provided, find matching task
   
2. Read task details:
   - Extract task description
   - Extract target file path
   - Identify test type (unit, integration, contract)
   
3. Create test file:
   - Determine test file path (add _test suffix)
   - Create directory if needed
   
4. Write failing test:
   - Import necessary packages
   - Create test function
   - Write assertions that will fail
   - Include clear test descriptions
   
5. Run test to verify it fails:
   ```bash
   go test {test_file_path} -v
   ```
   - Must see FAIL status (RED phase of TDD)
   
6. Commit the failing test:
   ```bash
   git add {test_file_path}
   git commit -m "test: add failing test for {task_description}"
   ```

## Report
Success: "✓ Created failing test at {test_file_path}
         ✓ Test fails as expected (TDD RED phase)
         ✓ Committed test
         Ready to implement with /implement {task_id}"
Error: "Failed to create test: {error}"

## Examples
User: /test T006
Result: ✓ Created failing test at tests/contract/context_codex_test.go
        ✓ Test fails as expected (TDD RED phase)
        ✓ Committed test
        Ready to implement with /implement T006

User: /test "AGENTS.md generation"
Result: ✓ Created failing test at tests/integration/agents_md_test.go
        ✓ Test fails as expected (TDD RED phase)
        ✓ Committed test

## Natural Language
Patterns that trigger this command:
- "write test for {x}" → /test "{x}"
- "create test for task {x}" → /test {x}
- "test {x}" → /test "{x}"

## Notes
Tests MUST fail initially - this is the RED phase of Red-Green-Refactor. Never skip this step. The test defines the expected behavior before implementation exists.