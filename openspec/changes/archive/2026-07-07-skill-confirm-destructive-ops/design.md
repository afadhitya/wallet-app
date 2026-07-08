## Context

The `skill/SKILL.md` file defines rules for AI agents using the `wallet` CLI. Currently it lacks guidance on destructive operations — an agent may run `wallet rm`, `wallet adjust`, or similar commands without confirming with the user. For a personal finance tool, undoing such mutations is difficult or impossible.

## Goals / Non-Goals

**Goals:**
- Add a single rule to `skill/SKILL.md` requiring agent confirmation before destructive operations
- Cover all destructive commands: `rm`, `archive`, `adjust`, `budget rm`, `bill rm`, `category rm`, `tag rm`, `rate rm`
- Explicitly exempt read-only queries and data creation from confirmation

**Non-Goals:**
- No code changes to the wallet CLI itself
- No changes to `EXAMPLES.md` or other documentation files
- No automated tests (documentation-only change)
- No changes to how the wallet CLI prompts users (the CLI already has its own interactive prompts; this rule is for agent behavior)

## Decisions

### Placement in Rules section
Insert the new bullet after the `--force`/`rm` rule and the `--` flag separator rule, before "Never touch the database directly." This groups it with other operational safety rules.

### Confirmation requirement: "show a brief summary"
The rule instructs agents to show what will be affected (account name, amount, transaction description) before asking for confirmation. This prevents blind confirmation where the user doesn't know what they're approving.

### No intent mapping table changes
The intent mapping table maps user phrases to commands — it's not the right place for agent behavior rules. Keeping the rule in the Rules section is cleaner.

## Risks / Trade-offs

- **Risk**: Over-confirmation fatigue if users trigger many destructive ops. → **Mitigation**: Read queries and data creation are explicitly exempt. Destructive ops are rare in normal usage.
- **Risk**: Rule is advisory, not enforced. An agent could still skip confirmation. → **Mitigation**: This is a documentation/skill change; the wallet CLI's own `--force` flag behavior is unchanged.
