## Context

`skill/COMMANDS.md` is a 625-line reference file listing every wallet CLI command with parameter tables, flag descriptions (required/optional), and JSON response examples. SKILL.md references it as the complete command reference for agents.

Agents can discover parameters via `wallet <command> --help` and JSON output shapes by running commands with `--json`. The verbose tables are unnecessary duplication.

## Goals / Non-Goals

**Goals:**
- Reduce `skill/COMMANDS.md` to a concise command inventory (~60 lines)
- Preserve the domain grouping structure (transactions, accounts, categories, etc.)
- Keep command signatures only (name + positional args pattern)
- Add a header note about `--json` for JSON output
- Remove all parameter tables, flag descriptions, and JSON response examples

**Non-Goals:**
- Changing SKILL.md references to COMMANDS.md (the reference stays, description gets tightened)
- Modifying any Go code or CLI behavior
- Removing any actual commands from the inventory

## Decisions

**Decision: One-line-per-command format grouped by domain**
- Rationale: Agents need a quick scan of available commands; parameters are discovered at runtime via `--help`. Domain grouping preserves the logical structure.

**Decision: Keep header note instead of per-command notes**
- Rationale: The JSON output behavior is global (`--json` flag), so a single note applies universally.

**Decision: Remove all parameter tables and examples**
- Rationale: `--help` output from the CLI is always the authoritative source for parameters. Examples and response shapes become stale; live `--json` output is always current.

## Risks / Trade-offs

- **Risk**: Agents may not discover certain flags that are only documented in COMMANDS.md, not in `--help`. Mitigation: All flags are already documented in Cobra's `--help` output by definition.
- **Risk**: Some response shapes vary across commands. Mitigation: Agents can run any command with `--json` to observe the current output structure.
