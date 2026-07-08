## Why

The agent skill in `skill/SKILL.md` lacks guidance on destructive operations. An AI agent may run `wallet rm`, `wallet adjust`, or similar mutation commands without confirming with the user. For a personal finance tool, undoing such operations is difficult or impossible. This change ensures agents always ask before deleting or altering financial data.

## What Changes

- Add one bullet point to the **Rules** section of `skill/SKILL.md` requiring agent confirmation before executing destructive commands (`rm`, `archive`, `adjust`, `budget rm`, `bill rm`, etc.)
- Read-only queries and data creation commands are explicitly exempt from confirmation

## Capabilities

### New Capabilities

- `skill-destructive-confirmation`: Require AI agents to ask for user confirmation before performing destructive wallet operations (delete, archive, adjust). Read queries and data creation are exempt.

### Modified Capabilities

_None. This is a documentation-only change — no existing capability requirements change._

## Impact

- `skill/SKILL.md` — one new bullet in the Rules section
- No code changes, no API changes, no database changes, no new tests required
