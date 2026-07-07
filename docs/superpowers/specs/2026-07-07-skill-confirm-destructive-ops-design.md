# Design: Add confirmation rule for destructive operations to skill/SKILL.md

## Summary

Add a rule to `skill/SKILL.md` instructing AI agents to ask for user confirmation before performing destructive wallet operations (delete, archive, adjust). This is a documentation-only change addressing issue #39.

## Motivation

The agent skill in `skill/SKILL.md` lacks guidance on destructive operations. An AI agent may run `wallet rm`, `wallet adjust`, or similar commands without confirming with the user. For a personal finance tool, undoing such mutations is difficult or impossible.

## Design

### What changes

One new bullet point in the **Rules** section of `skill/SKILL.md`, inserted before the existing "Never touch the database directly" rule.

```markdown
- **Confirm destructive operations.** Before running any command that permanently deletes, archives, or alters data (`rm`, `archive`, `adjust`, `budget rm`, `bill rm`), ask the user to confirm. Show a brief summary of what will be affected (account name, amount, transaction description) and wait for explicit approval. For batch operations (e.g., "delete all"), describe the scope before proceeding. Read-only queries (`list`, `report`, `forecast`, `bill due`) and data creation (`add`, `set`) do not need confirmation.
```

### Destructive commands covered

- `wallet rm <id> --force` — permanently delete a transaction
- `wallet bill rm <id>` — delete a planned payment
- `wallet account archive <id>` — archive an account
- `wallet adjust <account> <amount>` — change a balance
- `wallet budget rm <id>` — delete a budget
- `wallet category rm <id>` — delete a category
- `wallet tag rm <name>` — delete a tag
- `wallet rate rm <currency>` — delete an exchange rate

### Exceptions (no confirmation needed)

- Read queries: `list`, `report`, `forecast`, `bill due`, `budget check`, `account list`, `category list`, `tag list`, `rate list`
- Data creation: `add`, `set`, `edit`, `pay`, `skip`, `pause`, `resume`

### Placement

Inserted into the existing Rules block, positioned after the `--force`/`rm` rule and the `--` flag separator rule, before "Never touch the database directly." This groups it with other operational safety rules.

### Files changed

- `skill/SKILL.md` — add one bullet point to the Rules section

No other files modified. No code changes. No tests required.

## Alternatives considered

1. **Rule + update EXAMPLES.md** — show confirmation in destructive workflow examples. Rejected: adds clutter; the rule covers all cases generically.
2. **Rule + annotate intent mapping table** — mark destructive commands in the table. Rejected: table is for intent-to-command mapping, not agent behavior rules.

## Verification

- Read the Rules section of SKILL.md. Confirm the new bullet appears in the correct position.
- No automated tests needed (documentation change).
