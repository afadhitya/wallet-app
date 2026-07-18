## Context

The `skill/SKILL.md` file defines rules and principles for AI agents using the `wallet` CLI. It currently has a rule "Never auto-create tags, categories, or accounts" and another "Confirm before destructive operations." However, there is no explicit rule telling agents to ask users for missing mandatory parameters (category, account, etc.) instead of guessing or assuming values.

Current behavior: agents may silently pick the first available category, account, or other required parameter when the user omits it in a command, leading to incorrect data being recorded.

## Goals / Non-Goals

**Goals:**
- Add an explicit rule to `skill/SKILL.md` requiring agents to ask for missing mandatory parameters
- Include concrete examples showing acceptable (asking with suggestions) vs. unacceptable (assuming) behavior
- Update the `agent-guidelines` spec with a corresponding requirement

**Non-Goals:**
- Changing any CLI commands or flags
- Adding interactive prompts to the `wallet` CLI itself
- Enforcing this rule at the tool level — it remains a convention for agents

## Decisions

**Decision 1: Place the rule in Core Principles vs. Rules section**

The rule fits naturally as a new Core Principle (item #6 or similar) since it's a foundational behavior guideline. The Core Principles section already covers "Never auto-create" (principle #3), making "ask for missing params" a logical neighbor.

**Decision 2: Allow suggestions but not assumptions**

The rule distinguishes between suggesting options (acceptable) and assuming values (unacceptable). This gives agents a helpful middle ground: they can list available resources to help the user choose, but must not proceed without explicit confirmation.

**Alternative considered:** A blanket "never guess" rule without the suggestion allowance. Rejected — listing options is helpful UX and doesn't risk data pollution since the user still makes the final choice.

**Decision 3: Include concrete examples in SKILL.md**

The rule includes both a bad example (agent creates bill with assumed category/account) and a good example (agent asks and waits for user response), matching the pattern used in the flag-argument separator documentation.

## Risks / Trade-offs

- **Risk: More verbose agent interactions.** Agents will ask questions instead of silently proceeding. → Mitigation: The suggestion allowance means agents can provide context that helps the user answer quickly.

- **Risk: Some agents may ignore this guideline.** SKILL.md is advisory for AI agents, not mechanically enforced. → Mitigation: Clear, prominent placement and concrete examples maximize adherence. This is consistent with all other SKILL.md rules.
