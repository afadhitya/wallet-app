## ADDED Requirements

### Requirement: Agent asks for missing mandatory parameters

The `skill/SKILL.md` SHALL state that AI agents MUST ask the user when mandatory parameters (category, account, etc.) are missing from a command, rather than assuming or guessing values. Agents MAY suggest options (e.g., "Which category? Some options: Groceries, Food & Dining, Household"), but the final choice MUST come from the user.

#### Scenario: Agent receives a command with missing required parameters

- **WHEN** an AI agent receives a user request that maps to a wallet command requiring parameters the user did not provide (e.g., category, account, amount)
- **THEN** the agent SHALL ask the user to specify those parameters rather than guessing or silently using a default value

#### Scenario: Agent suggests options to help user decide

- **WHEN** an AI agent asks the user for a missing parameter
- **THEN** the agent MAY list available options (categories, accounts, tags, etc.) to help the user choose, but MUST wait for the user's explicit response before proceeding

#### Scenario: Agent assumes missing parameter values

- **WHEN** an AI agent fills in a missing parameter with a guessed or assumed value without asking the user
- **THEN** this SHALL be treated as a violation of the SKILL.md guidelines
