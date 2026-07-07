## Context

The wallet CLI has `version` and `update` subcommands (`internal/cli/version.go`, `internal/cli/update.go`) that are registered in the root command. The `version` command shows the current version and supports `--check` to compare against the latest GitHub release. The `update` command downloads and replaces the binary with the latest release, supporting `--force`.

Currently these commands and their error codes are absent from all user-facing and agent-facing documentation:
- `skill/COMMANDS.md` has no System/Version section
- `skill/ERRORS.md` is missing 5 update-related error codes (`UPDATE_NETWORK_ERROR`, `UPDATE_PERMISSION_ERROR`, `UPDATE_ALREADY_LATEST`, `UPDATE_FAILED`, `UPDATE_CHECKSUM_MISMATCH`)
- `README.md` has no mention of these commands

## Goals / Non-Goals

**Goals:**
- Add a "System" section to `skill/COMMANDS.md` documenting `wallet version [--check]` and `wallet update [--force]`
- Add 5 update-related error codes to `skill/ERRORS.md` reference table and recovery patterns
- Add a brief mention of these commands in the README Commands section

**Non-Goals:**
- No code changes to the CLI commands themselves
- No new commands or features
- No changes to `skill/EXAMPLES.md` or `skill/SKILL.md`

## Decisions

**Placement in COMMANDS.md**: Added as a new "System" section at the end of the existing domain groups (after "Init"). This is consistent with how Rate and Init are grouped as separate sections. Using "System" as the section name groups version/update logically as system-level commands distinct from financial operations.

**Placement in README.md**: Added to the "Quick Start → 1. Initialize" section or the "Commands" section as a brief mention that these commands exist. No dedicated section needed — they are auxiliary commands, not core workflow commands.

**Error codes for ERRORS.md**: The 5 update error codes come from `internal/cli/helpers.go` lines 93-97. They map to sentinel errors defined in `pkg/update/updater.go`. Error messages and recovery actions should mirror the pattern of existing entries in ERRORS.md.

## Risks / Trade-offs

- **Minimal risk** — purely additive documentation changes to three Markdown files. No code, no config, no database changes.
