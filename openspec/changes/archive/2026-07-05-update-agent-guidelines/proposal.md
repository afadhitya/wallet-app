## Why

AI agents using the wallet CLI need explicit guidance on CLI flag/argument ordering (the `--` separator) and a firm rule against direct database manipulation. Without these, agents may construct incorrect CLI commands or resort to dangerous direct SQL/data scripting. Additionally, users need clear instructions for installing the skill file so their AI tools can leverage it.

## What Changes

- Add documentation about the `--` separator syntax to `skill/SKILL.md`: flags must appear **before** `--`, and all arguments after `--` become positional (e.g., `wallet adjust "Bunga Bank" --json -- -3612 "Initial balance"`)
- Add an explicit rule to `skill/SKILL.md` that AI agents **must not** touch the database directly or create scripts that manipulate database data — they must only use the `wallet` CLI
- Add installation instructions to `README.md` explaining how to register `skill/SKILL.md` with AI agentic tools (e.g., Hermes Agent, OpenClaw)

## Capabilities

### New Capabilities
- `agent-guidelines`: CLI usage patterns, data access boundaries, and skill installation instructions for AI agents using the wallet CLI

### Modified Capabilities
_None_

## Impact

- **Affected files**: `skill/SKILL.md` — two new rules added to the Rules section; `README.md` — new section for skill installation
- **No code changes**, no API changes, no dependency changes
