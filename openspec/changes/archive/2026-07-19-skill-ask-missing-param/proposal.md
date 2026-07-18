## Why

The agent skill in `skill/SKILL.md` doesn't explicitly tell AI agents what to do when mandatory parameters are missing. Agents might guess or assume values (e.g., picking a category or account on behalf of the user), use previously seen data as defaults, or make decisions that should be left to the user. The wallet is a personal finance tool — data accuracy matters, and auto-assuming parameters can silently pollute financial data.

## What Changes

- Add a new rule to the **Core Principles** or **Rules** section of `skill/SKILL.md`:
  > **Ask for missing required parameters.** If the user doesn't provide a mandatory parameter (category, account, etc.), ask them. Do not assume or guess any input values. It's okay to *suggest* options (e.g., "Which category? Some options: Groceries, Food & Dining, Household"), but the final choice must come from the user.
- Include concrete examples of bad (assuming) vs. good (asking with suggestions) behavior.

## Capabilities

### New Capabilities
<!-- No new capabilities introduced; this is a refinement of existing agent behavior guidelines. -->

### Modified Capabilities
- `agent-guidelines`: Adding a requirement that agents MUST ask for missing mandatory parameters instead of assuming or guessing values. It's acceptable to suggest options, but the final choice must come from the user.

## Impact

- `skill/SKILL.md`: Add new rule and examples to Core Principles or Rules section.
- `openspec/specs/agent-guidelines/spec.md`: Add new requirement with scenario for missing parameter handling.
