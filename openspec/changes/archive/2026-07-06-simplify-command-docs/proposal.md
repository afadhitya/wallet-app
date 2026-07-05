## Why

`skill/COMMANDS.md` contains verbose parameter tables, flag descriptions, and JSON response examples for every wallet CLI command. Since agents can discover parameters via `wallet <command> --help`, these detailed references are redundant. They bloat the file to 625 lines, increasing token consumption and obscuring the command inventory.

## What Changes

- Reduce `skill/COMMANDS.md` to a concise command inventory: each command on one line with its signature only
- Strip all parameter tables, flag descriptions, required/optional markings, and JSON response examples
- Add a single note at the top: all commands accept `--json` and output is JSON format
- Agents can discover parameters via `wallet <command> --help` at runtime

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `agent-guidelines`: Command reference embedded in conversation context is simplified to bare command signatures

## Impact

- `skill/COMMANDS.md`: Reduced from ~625 lines to ~60 lines (~90% reduction); faster agent processing and lower context usage
- No code or behavior changes
