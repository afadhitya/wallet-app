## Why

The wallet CLI has grown to 50+ commands across multiple domains, but CLI reference docs must be maintained manually and fall out of sync with code. Additionally, the existing `skill/SKILL.md` mixes command reference, error codes, and workflow examples into a single monolithic file, making it harder for AI agents to find specific information efficiently.

## What Changes

- **Auto-generate CLI reference docs** from Cobra command definitions using `cobra/doc`, ensuring docs stay in sync with code
- **Add a hidden `wallet docs markdown` command** that generates Markdown files from the Cobra command tree
- **Add `make docs` Makefile target** to trigger documentation generation
- **Split `skill/SKILL.md` into focused files**: `COMMANDS.md` (command reference grouped by domain), `ERRORS.md` (error codes and recovery), `EXAMPLES.md` (common workflows)
- **Move command reference out of `skill/SKILL.md`** and reference the split files instead, keeping SKILL.md focused on principles, rules, and intent mapping
- **Update README.md agent skill installation** to install the entire `skill/` directory (not just `SKILL.md`)
- **Replace README.md Commands table** with a reference to auto-generated `docs/cli/` instead of a manually maintained command list
- **Update CONTRIBUTING.md** to instruct contributors to run `make docs` after adding or modifying CLI commands
- **Deduplicate AGENTS.md** by referencing CONTRIBUTING.md for build commands, project structure, coding conventions, commit conventions, and code generation workflows instead of maintaining duplicate copies

## Capabilities

### New Capabilities

- `cli-doc-generation`: Auto-generate Markdown CLI reference documentation from Cobra command definitions via a hidden `wallet docs markdown` command and `make docs` Makefile target. Generated docs go to `docs/cli/`, excluded from git, regenerated on release.
- `ai-agent-documentation`: Split AI agent documentation into focused, domain-grouped files under `skill/` — `COMMANDS.md` for command reference, `ERRORS.md` for error codes and recovery patterns, `EXAMPLES.md` for common workflows. Refactor `SKILL.md` to focus on core principles and reference the split files.

### Modified Capabilities

- `agent-guidelines`: SKILL.md content will be refactored - command reference and error code sections moved to separate files, SKILL.md updated to reference them. README agent skill installation instructions updated to install the entire `skill/` directory. Core rules and principles remain in SKILL.md.
- `project-documentation`: CONTRIBUTING.md updated to instruct contributors to run `make docs` after adding or modifying CLI commands. AGENTS.md refactored to reference CONTRIBUTING.md for duplicated content (build commands, project structure, coding conventions, commit conventions, code generation) instead of maintaining copies.

## Impact

- **New file**: `internal/cli/docs.go` — hidden `docs` command with `markdown` subcommand
- **Modified file**: `internal/cli/root.go` — register the new docs command
- **Modified file**: `Makefile` — add `docs` target
- **Modified file**: `.gitignore` — add `docs/cli/` to ignore list
- **New files**: `skill/COMMANDS.md`, `skill/ERRORS.md`, `skill/EXAMPLES.md`
- **Modified file**: `skill/SKILL.md` — refactored to reference split files
- **Modified file**: `README.md` — agent skill installation updated to copy entire `skill/` directory
- **Modified file**: `CONTRIBUTING.md` — instruct contributors to run `make docs` after CLI changes
- **Modified file**: `AGENTS.md` — deduplicate by referencing CONTRIBUTING.md for build commands, project structure, coding conventions, commit conventions, and code generation
- **New dependency**: `github.com/spf13/cobra/doc` (already in module tree via cobra)
