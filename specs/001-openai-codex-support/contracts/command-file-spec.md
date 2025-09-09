# Command File Specification

## Purpose
Define the structure and requirements for command files that Codex reads and executes.

## File Structure

### Required Sections

#### 1. Header
```markdown
# /command-name
```
- Must start with `#` and `/`
- Command name must be lowercase with hyphens
- This becomes the trigger for the command

#### 2. Description
```markdown
## Description
Brief explanation of what this command does and when to use it.
```
- Required for documentation
- Should be 1-2 sentences
- Used by `specify codex list`

#### 3. Usage
```markdown
## Usage
When user types: `/command <required-arg> [optional-arg]`
```
- Shows the expected command format
- Use `<>` for required arguments
- Use `[]` for optional arguments

#### 4. Execute
```markdown
## Execute
1. Step description
   - Sub-step details
2. Another step
   ```bash
   shell command to run
   ```
3. Parse output
   - Extract: `field_name` from JSON
4. Continue...
```
- Numbered steps that Codex follows
- Can include shell commands in code blocks
- Use `Extract:` for parsing instructions
- Use `@file/path` for file references

#### 5. Report
```markdown
## Report
Success: "Created {thing} at {path}"
Error: "Failed to {action}: {error}"
```
- Templates for status messages
- Use `{}` for variable substitution

### Optional Sections

#### Examples
```markdown
## Examples
```
User: /command "argument"
Result: Success message

User: /command --flag value
Result: Different outcome
```

#### Natural Language
```markdown
## Natural Language
Patterns that trigger this command:
- "create a spec for {x}" → /command "{x}"
- "do something with {x}" → /command "{x}"
```

#### Prerequisites
```markdown
## Prerequisites
- Current directory must be git repository
- Feature branch must exist
- Spec file must be present
```

#### Notes
```markdown
## Notes
Additional context or warnings for Codex to consider.
```

## Variable Substitution

Commands can use these variables:

- `{project_root}`: Repository root directory
- `{current_branch}`: Active git branch
- `{spec_dir}`: Current spec directory
- `{timestamp}`: Current timestamp
- `{user_input}`: The user's command arguments

## File References

Use @ notation for files:
- `@.codex/context/templates/spec-template.md`: Absolute from .codex
- `@specs/current/spec.md`: Relative to project root
- `@{spec_dir}/plan.md`: Using variables

## Shell Command Execution

```markdown
Run command:
```bash
specify feature create "{user_input}"
```
Parse as JSON:
- branch_name
- spec_path
```

## Conditional Logic

```markdown
If file exists @{spec_dir}/plan.md:
  - Read existing plan
  - Append new sections
Else:
  - Create new plan from template
```

## Error Handling

```markdown
On error:
  - If "file not found": Suggest creating it first
  - If "permission denied": Request elevated permissions
  - Otherwise: Show error and stop
```

## Validation Rules

1. **Structure**: Must have required sections in order
2. **Commands**: Shell commands must be in code blocks
3. **Variables**: All {} variables must be defined
4. **References**: @ references must be valid paths
5. **Steps**: Execute section must have numbered steps

## Example Complete Command File

```markdown
# /specify

## Description
Create a new feature specification from a description.

## Usage
When user types: `/specify "<feature-description>"`

## Prerequisites
- Must be in a git repository
- specify CLI must be installed

## Execute
1. Run the specify command:
   ```bash
   specify feature create "{user_input}"
   ```
   
2. Parse JSON output:
   - Extract: `branch_name`
   - Extract: `spec_path`
   
3. Checkout new branch:
   ```bash
   git checkout {branch_name}
   ```
   
4. Read template:
   - Load: @.codex/context/templates/spec-template.md
   
5. Fill template:
   - Replace [FEATURE NAME] with {user_input}
   - Replace [DATE] with {timestamp}
   - Mark uncertainties with [NEEDS CLARIFICATION]
   
6. Save specification:
   - Write to: {spec_path}
   
7. Validate specification:
   ```bash
   specify feature check
   ```

## Report
Success: "✓ Created feature '{branch_name}' with spec at {spec_path}"
Error: "Failed to create feature: {error}"

## Examples
User: /specify "user authentication with OAuth"
Result: ✓ Created feature '001-user-auth-oauth' with spec at specs/001-user-auth-oauth/spec.md

## Natural Language
Patterns that trigger this command:
- "create a spec for {x}" → /specify "{x}"
- "start a new feature {x}" → /specify "{x}"
- "specify {x}" → /specify "{x}"

## Notes
This creates a new git branch and specification file. The spec will need to be edited to resolve any [NEEDS CLARIFICATION] markers before planning.
```

## Testing Command Files

Command files should be tested for:

1. **Parseability**: Can be read and sections extracted
2. **Executability**: Steps can be followed by Codex
3. **Completeness**: All required sections present
4. **Validity**: Shell commands are valid
5. **Safety**: No destructive operations without confirmation