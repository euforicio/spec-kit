# [PROJECT NAME] Development Guidelines

Auto-generated from all feature plans. Last updated: [DATE]

## Active Technologies
[EXTRACTED FROM ALL PLAN.MD FILES]

## Project Structure
```
[ACTUAL STRUCTURE FROM PLANS]
```

## Commands

### Built-in Commands

The following slash commands are available in this spec-driven development environment.
For detailed usage and examples of any command, see the corresponding documentation file in `.{{.AIAssistant}}/commands/<command>.md`.

**`/specify`** - Creates a new feature specification and branch  
Start the spec-driven development lifecycle by creating a specification from your feature description.

**`/plan`** - Creates an implementation plan from a feature specification  
Second phase: convert specification into implementation plan with research, design docs, and contracts.

**`/tasks`** - Breaks down the implementation plan into executable tasks  
Third phase: generate numbered, ordered tasks for implementation following TDD methodology.

### Command Flow
1. `/specify <description>` → Creates spec.md and feature branch
2. `/plan` → Creates plan.md, research.md, contracts/, data-model.md, quickstart.md  
3. `/tasks` → Creates tasks.md with numbered implementation tasks

- Each command has detailed documentation at `.{{.AIAssistant}}/commands/<command>.md`
- Additional commands can be added by creating corresponding documentation files
- Command documentation includes usage examples, parameters, and expected outputs

[ONLY COMMANDS FOR ACTIVE TECHNOLOGIES]

## Code Style
[LANGUAGE-SPECIFIC, ONLY FOR LANGUAGES IN USE]

## Recent Changes
[LAST 3 FEATURES AND WHAT THEY ADDED]
