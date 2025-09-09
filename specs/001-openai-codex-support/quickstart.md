# Quickstart: Codex Native Integration

## Prerequisites

1. **Install specify CLI**
   ```bash
   # Verify specify is installed
   specify --version
   
   # Should show version 0.2.0 or higher
   ```

2. **Have OpenAI Codex ready**
   ```bash
   # Verify Codex CLI is available
   which codex
   
   # Test Codex works
   codex "Hello, are you ready?"
   ```

## Initial Setup

### 1. Generate Codex Context

```bash
# Navigate to your project
cd /path/to/your/project

# Generate AGENTS.md and command files
specify feature context codex

# Expected output:
# ✓ Created/Updated AGENTS.md
# ✓ Created .codex/commands/ directory
# ✓ Installed 12 command files
# ✓ Added project context
```

### 2. Verify Installation

```bash
# Check generated files
ls -la AGENTS.md .codex/

# Should see:
# AGENTS.md (in project root)
# .codex/commands/ (with command files)

# View AGENTS.md
head -30 AGENTS.md

# Should show Codex section with command instructions
```

## Using Codex with Specify

### 1. Start Codex

```bash
# Start Codex in your project directory
# (Codex automatically reads AGENTS.md on startup if present)
codex

# Codex responds:
# "I've loaded AGENTS.md and understand this is a specify project.
#  Command system detected in .codex/commands/.
#  Available commands:
#  /specify - Create specifications
#  /plan - Plan implementations
#  /tasks - Generate tasks
#  ...
#  Ready to help with spec-driven development!"
```

### 2. Create Your First Feature

In Codex, type:
```
/specify "user authentication system"
```

Codex will:
1. Read `.codex/commands/specify.md`
2. Execute `specify feature create "user authentication system"`
3. Create branch `001-user-authentication`
4. Generate `specs/001-user-authentication/spec.md`
5. Report success with file location

### 3. Plan the Implementation

Continue in Codex:
```
/plan
```

Codex will:
1. Read the current spec
2. Create `research.md` with technical decisions
3. Generate `data-model.md` with entities
4. Create `contracts/` with API specs
5. Generate `quickstart.md` for testing

### 4. Generate Tasks

Still in Codex:
```
/tasks
```

Codex will:
1. Analyze the plan and contracts
2. Generate `tasks.md` with numbered tasks
3. Mark parallel tasks with [P]
4. Order by TDD principles

### 5. Start Implementation

In Codex:
```
/test T001
```

Codex will:
1. Read task T001 from tasks.md
2. Write failing test first (RED phase)
3. Commit the test
4. Guide you to implement (GREEN phase)

## Complete Workflow Example

Here's a full session in Codex:

```
# Start Codex (AGENTS.md loaded automatically)
Codex: Loaded AGENTS.md. Specify workflow ready with 12 commands.

User: /specify "payment processing with Stripe"
Codex: ✓ Created feature '001-payment-stripe' at specs/001-payment-stripe/spec.md
       Found 3 items needing clarification:
       - Payment methods to support?
       - Refund policy?
       - Currency handling?

User: Support cards and ACH, 30-day refunds, USD only
Codex: ✓ Updated specification with clarifications

User: /plan
Codex: ✓ Created research.md with Stripe SDK analysis
       ✓ Generated data-model.md with Payment, Transaction entities
       ✓ Created contracts/payment-api.yaml
       ✓ Completed plan.md

User: /tasks
Codex: ✓ Generated 42 tasks in tasks.md
       - 5 setup tasks
       - 8 test tasks (marked [P])
       - 15 implementation tasks
       - 14 integration tasks

User: Show me the first test task
Codex: T006 [P]: Contract test for POST /payments in tests/contract/payments_test.go
       Should test payment creation endpoint with Stripe mock

User: /test T006
Codex: Creating failing test for payment creation...
       [Shows test code]
       ✓ Test written and failing as expected
       Ready to implement when you run: /implement T006
```

## Natural Language Alternative

You can also use natural language instead of slash commands:

```
User: Create a spec for user authentication
Codex: [Recognizes intent, executes /specify command]

User: Plan this feature
Codex: [Recognizes intent, executes /plan command]

User: Break this down into tasks
Codex: [Recognizes intent, executes /tasks command]
```

## Available Commands

View all commands:
```bash
specify codex list
```

Or in Codex:
```
/help
```

Core commands:
- `/specify` - Create feature specifications
- `/plan` - Plan implementations
- `/tasks` - Generate task breakdowns
- `/test` - Write tests (TDD)
- `/implement` - Write implementation code
- `/validate` - Check against spec
- `/commit` - Create git commits

## Troubleshooting

### Codex doesn't recognize commands

```bash
# Verify .codex exists
ls .codex/

# Validate structure
specify codex validate --fix

# Reinitialize if needed
specify codex init --force
```

### Command not working as expected

```bash
# Check command file
cat .codex/commands/specify.md

# Update to latest
specify codex update
```

### Natural language not mapping

In Codex:
```
/help specify
```

Shows the natural language patterns for that command.

## Advanced Usage

### Custom Commands

Add your own command:
```bash
specify codex add review --template advanced
```

Edit `.codex/commands/review.md` to define behavior.

### Command Chaining

In Codex:
```
Create and plan a feature for OAuth login
```

Codex recognizes this as:
1. `/specify "OAuth login"`
2. `/plan`

### Workspace-Specific Context

Add project-specific guidance:
```bash
echo "Always use PostgreSQL for storage" >> .codex/context/project-rules.md
```

Reference in Codex:
```
Check @.codex/context/project-rules.md for our standards
```

## Validation Checklist

- [ ] `AGENTS.md` exists in project root
- [ ] `.codex/commands/` directory contains command files
- [ ] Codex loads AGENTS.md automatically on startup
- [ ] `/specify` command creates a spec
- [ ] `/plan` generates artifacts
- [ ] `/tasks` creates task list
- [ ] Natural language patterns work
- [ ] Commands execute specify CLI correctly
- [ ] File references with @ notation work
- [ ] Command files are readable by Codex

## Next Steps

1. **Explore Examples**: Check `.codex/examples/` for workflows
2. **Customize**: Add project-specific commands
3. **Integrate**: Add to your team's workflow documentation
4. **Contribute**: Share useful commands with the community