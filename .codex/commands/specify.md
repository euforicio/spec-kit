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
   - Extract: `spec_file`
   - Extract: `feature_num`
   
3. Checkout new branch:
   ```bash
   git checkout {branch_name}
   ```
   
4. Read template:
   - Load: @templates/spec-template.md (if exists)
   
5. Fill template:
   - Replace [FEATURE NAME] with {user_input}
   - Replace [DATE] with current date
   - Replace [###-feature-name] with {branch_name}
   - Mark uncertainties with [NEEDS CLARIFICATION]
   
6. Save specification:
   - Write updated content to: {spec_file}
   
7. Validate specification:
   ```bash
   specify feature check
   ```

## Report
Success: "✓ Created feature '{branch_name}' with spec at {spec_file}"
Error: "Failed to create feature: {error}"

## Examples
User: /specify "user authentication with OAuth"
Result: ✓ Created feature '001-user-auth-oauth' with spec at specs/001-user-auth-oauth/spec.md

User: /specify "payment processing"
Result: ✓ Created feature '002-payment-processing' with spec at specs/002-payment-processing/spec.md

## Natural Language
Patterns that trigger this command:
- "create a spec for {x}" → /specify "{x}"
- "start a new feature {x}" → /specify "{x}"
- "specify {x}" → /specify "{x}"
- "I want to build {x}" → /specify "{x}"

## Notes
This creates a new git branch and specification file. The spec will need to be edited to resolve any [NEEDS CLARIFICATION] markers before planning. The feature number is automatically assigned based on existing features.