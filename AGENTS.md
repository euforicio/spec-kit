# Repository Guidelines

## Project Structure & Module Organization
- `cmd/specify/`: CLI entrypoint (`main.go`).
- `internal/cli/`: Cobra commands (`root`, `init`, `check`, `feature`).
- `internal/services/`: File, git, environment, template, project helpers.
- `internal/models/` and `internal/ui/`: Data structures and small TUI utilities.
- `templates/`: Spec, plan, and task templates used by the CLI.
- `specs/`: Spec-driven plans/research for features.
- `tests/contract/`: Black-box CLI and API contract tests.
- `Makefile`, `.golangci.yml`, `.goreleaser.yml`: build, lint, release config.

## Build, Test, and Development Commands
- `make build`: Build `./bin/specify` for the host OS.
- `./bin/specify --help`: Run the CLI locally.
- `make test`: Run all Go tests with race + coverage.
- `make test-contract`: Run only contract tests in `tests/contract`.
- `go test -short ./...`: Skip networked tests (GitHub API checks).
- `make lint` / `make lint-fix`: Lint and auto-fix where possible.
- `make fmt` / `make vet` / `make check`: Format, vet, and run full checks.
- `make dev`: Hot reload with Air (install via `go install github.com/cosmtrek/air@latest`).

## Coding Style & Naming Conventions
- Go 1.25+. Standard Go formatting enforced (`gofmt`, `goimports`, `gofumpt`).
- Lint via `golangci-lint` (see `.golangci.yml` for enabled rules and thresholds).
- Packages: lowercase, no underscores; exported identifiers use CamelCase.
- Keep functions small and testable; prefer table-driven tests.
- Error handling: wrap with context; avoid naked returns.

## Testing Guidelines
- Framework: Go `testing` with `testify` (`assert`, `require`).
- Layout: contract tests live in `tests/contract` and build a temp CLI binary.
- Names: files end with `_test.go`; functions `TestXxx`.
- Concurrency: use `t.Parallel()` where safe; run with `-race` (default in `make test`).
- Coverage: `make coverage` generates `coverage.html` from `coverage.out`.

## Commit & Pull Request Guidelines
- Commits: imperative mood, concise subject (≤72 chars), explain the why.
- Branches: use focused branches (e.g., `feat/`, `fix/`, `chore/`).
- PRs: clear description, scope of change, tests added/updated, and docs updates (`README.md`, `spec-driven.md`, or templates) when user-facing behavior changes.
- Validate locally with `make check` and a quick `./bin/specify --help` run before requesting review.

## Security & Configuration
- Do not commit secrets. Networked tests hit the public GitHub API; use `-short` to skip in constrained environments.

<!-- specify-codex-start -->
## For OpenAI Codex

### Command System

When you see a message starting with "/", check .codex/commands/ for the matching command file.
Examples:
- `/specify "<args>"` → Read .codex/commands/specify.md
- `/plan "<args>"` → Read .codex/commands/plan.md
- `/tasks "<args>"` → Read .codex/commands/tasks.md

### Available Commands

- `/specify` - Create a new feature specification
- `/plan` - Plan implementation for current feature
- `/tasks` - Generate task breakdown
- `/test` - Write tests following TDD
- `/implement` - Implement code to pass tests
- `/validate` - Validate against specification
- `/commit` - Create git commit

### Natural Language Patterns

You can also use natural language:
- "create a spec for" → /specify
- "plan this" → /plan
- "break down tasks" → /tasks
- "write tests for" → /test
- "implement" → /implement
- "validate" → /validate
- "commit changes" → /commit

### Workflow

Follow the specify workflow:
1. Create specification (`/specify`)
2. Plan implementation (`/plan`)
3. Generate tasks (`/tasks`)
4. Write tests (`/test`)
5. Implement code (`/implement`)
6. Validate (`/validate`)
7. Commit changes (`/commit`)


<!-- specify-codex-end -->