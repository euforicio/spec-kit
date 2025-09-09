# /plan

## Description
Plan implementation for the current feature specification.

## Usage
When user types: `/plan`

## Prerequisites
- Must be on a feature branch (format: ###-feature-name)
- Specification must exist in specs/###-feature-name/spec.md

## Execute
1. Verify feature branch:
   ```bash
   git branch --show-current
   ```
   - Must match pattern: ###-feature-name
   
2. Run the plan command:
   ```bash
   specify feature plan --json
   ```
   
3. Parse JSON output:
   - Extract: `feature_spec`
   - Extract: `impl_plan`
   - Extract: `specs_dir`
   - Extract: `branch`
   
4. Read specification:
   - Load: {feature_spec}
   - Extract requirements and acceptance criteria
   
5. Read plan template:
   - Load: @templates/plan-template.md
   
6. Fill plan template:
   - Set Input path to {feature_spec}
   - Fill Technical Context based on spec
   - Mark any NEEDS CLARIFICATION items
   - Execute template's main() function
   
7. Generate Phase 0 artifacts:
   - Create research.md with technical decisions
   - Resolve all NEEDS CLARIFICATION items
   
8. Generate Phase 1 artifacts:
   - Create data-model.md from entities
   - Create contracts/ directory with API specs
   - Create quickstart.md with usage guide
   
9. Save all artifacts to {specs_dir}

## Report
Success: "✓ Created implementation plan at {impl_plan}
         ✓ Generated research.md, data-model.md, contracts/, quickstart.md
         Ready for task generation with /tasks"
Error: "Failed to create plan: {error}"

## Examples
User: /plan
Result: ✓ Created implementation plan at specs/001-oauth/plan.md
        ✓ Generated research.md, data-model.md, contracts/, quickstart.md
        Ready for task generation with /tasks

## Natural Language
Patterns that trigger this command:
- "plan this" → /plan
- "create implementation plan" → /plan
- "plan the feature" → /plan
- "how should I implement this" → /plan

## Notes
The plan command analyzes your specification and creates a detailed implementation plan with technical decisions, data models, and API contracts. It follows the spec-driven development methodology and prepares everything needed for task breakdown.