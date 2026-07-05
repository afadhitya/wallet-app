## 1. CLI Documentation Generation

- [x] 1.1 Create `internal/cli/docs.go` — add hidden `docs` command with `markdown` subcommand using `cobra/doc.GenMarkdownTree`, with configurable `--output` flag (default: `docs/cli`)
- [x] 1.2 Register the `docs` command in `internal/cli/root.go` by passing the root command to `newDocsCmd(rootCmd)` and adding it with `rootCmd.AddCommand()`
- [x] 1.3 Add `make docs` target to `Makefile` that runs `go run cmd/wallet/main.go docs markdown` and prints success message
- [x] 1.4 Add `docs/cli/` to `.gitignore` so generated docs are not committed
- [x] 1.5 Run `make lint` and `make test` to verify no regressions
- [x] 1.6 Run `make docs` to verify generation succeeds, and verify hidden commands (including `docs`) are excluded from output

## 2. AI Agent Documentation — Split Files

- [x] 2.1 Create `skill/COMMANDS.md` — domain-grouped command reference with flags, JSON output format, and response structures for each command
- [x] 2.2 Create `skill/ERRORS.md` — all error codes with meaning, cause, recovery action, and common recovery patterns table
- [x] 2.3 Create `skill/EXAMPLES.md` — common multi-step workflows (recording expenses, payday routine, subscription tracking, trip spending, budget management, etc.)
- [x] 2.4 Refactor `skill/SKILL.md` — remove "Command Quick Reference" table and "Error Codes" section, replace with references to `COMMANDS.md` and `ERRORS.md`, keep core principles, intent mapping, rules, common workflows, data model, and multi-currency sections
- [x] 2.5 Update `README.md` — change "Agent Skill (AI Tools)" section to install the entire `skill/` directory (not just `SKILL.md`); replace manual "Commands" table with pointer to `wallet --help` and auto-generated `docs/cli/`
- [x] 2.6 Update `CONTRIBUTING.md` — add instruction that contributors who add or modify CLI commands must run `make docs` to regenerate CLI reference documentation
- [x] 2.7 Refactor `AGENTS.md` — replace duplicated content (build commands table, project structure, coding conventions, commit conventions, code generation) with references to `CONTRIBUTING.md`; retain agent-specific sections (layer responsibilities, key patterns, database, transaction types, amounts, config, pre-commit hook)
- [x] 2.8 Run `make lint` and `make test` to verify no regressions
