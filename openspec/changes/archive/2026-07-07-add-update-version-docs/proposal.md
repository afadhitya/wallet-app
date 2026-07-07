## Why

The `wallet version` and `wallet update` CLI commands exist in the codebase but are not documented anywhere visible to users or AI agents — users discover them only via `wallet --help`, and AI agents using the skill files have no knowledge of these commands at all.

## What Changes

- Add a "System" section to `skill/COMMANDS.md` documenting `wallet version` (with `--check` flag) and `wallet update` (with `--force` flag)
- Add 5 update-related error codes to `skill/ERRORS.md`: `UPDATE_NETWORK_ERROR`, `UPDATE_PERMISSION_ERROR`, `UPDATE_ALREADY_LATEST`, `UPDATE_FAILED`, `UPDATE_CHECKSUM_MISMATCH`
- Add a mention of `version` and `update` commands in the README's Commands section

## Capabilities

### New Capabilities
- `<none>`: This is purely a documentation change with no new features.

### Modified Capabilities
- `ai-agent-documentation`: COMMANDS.md gains a System section for `version` and `update` commands, and ERRORS.md gains 5 new error codes, expanding the documented command inventory and error reference available to AI agents.

## Impact

- `skill/COMMANDS.md` — new System section with 2 commands
- `skill/ERRORS.md` — 5 new error code rows in the reference table and recovery patterns
- `README.md` — one additional bullet or line mentioning these commands
- No code changes, no API changes, no dependency changes
