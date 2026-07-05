## Context

The wallet CLI has 50+ Cobra commands across multiple domains. Currently there is no auto-generated CLI reference documentation. The `skill/SKILL.md` file serves as a monolithic AI agent guide mixing command reference, error codes, principles, and workflow examples. This makes it harder for agents to efficiently locate specific information and requires manual maintenance to keep command tables in sync with code.

## Goals / Non-Goals

**Goals:**
- Auto-generate CLI reference docs from Cobra command definitions to eliminate manual maintenance
- Split AI agent documentation into focused, purpose-specific files for faster lookup
- Keep generated docs out of version control (regenerate on demand)
- Use existing `cobra/doc` package — no new external dependency needed

**Non-Goals:**
- Man pages or other doc formats (Markdown only)
- Auto-generating AI agent docs from code (COMMANDS.md, etc. remain manual)
- CI auto-generation on every commit (manual `make docs` for now)
- Including hidden/developer commands in generated docs

## Decisions

### D1: Markdown-only generation via `cobra/doc.GenMarkdownTree`
The `cobra` library provides `doc.GenMarkdownTree()`, `doc.GenManTree()`, etc. We use only Markdown — it's the same format GitHub, docs sites, and AI agents consume. Man pages add complexity without a clear audience.

### D2: Hidden `wallet docs markdown` command
The generation command is a hidden Cobra subcommand with a `--output` flag for the output directory. Hidden commands don't show in `--help` output but remain executable. `cobra/doc.GenMarkdownTree` respects the `Hidden: true` flag — hidden commands (including `docs` itself) won't generate `.md` files.

### D3: `make docs` as the primary trigger
A `make docs` target runs `go run cmd/wallet/main.go docs markdown`. This requires a working Go build but no binary installation. CI jobs will also call `make docs` on tagged releases to publish docs.

### D4: `docs/cli/` in `.gitignore`
Generated docs are excluded from git. They're ephemeral — regenerated on demand from the authoritative source (the Go code). This prevents PRs from having stale generated docs.

### D5: Three split AI agent files
The `skill/` directory gets three new files: `COMMANDS.md`, `ERRORS.md`, `EXAMPLES.md`. This follows the brainstorming doc's recommended structure. Compared to alternatives:
- Single monolithic file → harder to scan, harder to maintain
- MORE granular files (one per domain) → too many files, harder to navigate
- Three files is the sweet spot: reference (commands), diagnostics (errors), tutorials (examples)

### D6: Domain-grouped command reference
`COMMANDS.md` groups by domain (Transaction, Account, Category, Tag, Budget, Bill, Forecast, Report, Rate) rather than alphabetically. AI agents reason in terms of "I need to work with accounts" — domain grouping matches that mental model.

### D7: SKILL.md retains core principles and intent mapping
The refactored `SKILL.md` keeps the sections agents need most: core principles, JSON envelope format, intent mapping table, rules, and data model. Command Quick Reference and Error Codes tables move to the split files, replaced with references.

## Risks / Trade-offs

- **Generated docs may have incomplete descriptions**: If Cobra command `Long` fields are sparse, generated docs will be thin. → Mitigation: review generated output and fill in `Long` descriptions where needed (existing commands already have good short/long descriptions).
- **Split files could get out of sync**: `COMMANDS.md` is manually written, not auto-generated. If commands change, `COMMANDS.md` must be manually updated. → Mitigation: `COMMANDS.md` is a curated reference for AI agents (focused on flags, JSON output, agent-relevant details), not an exhaustive dump. Major command changes are infrequent.
- **`make docs` requires Go toolchain**: Developers without Go installed can't generate docs. → Acceptable — this is a Go project and all developers have Go.
