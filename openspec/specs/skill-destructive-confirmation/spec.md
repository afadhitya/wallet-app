# Skill: skill-destructive-confirmation

## Purpose

Instruct AI agents to ask for user confirmation before executing destructive wallet operations that permanently delete, archive, or alter financial data. The agent must show a brief summary of what will be affected and wait for explicit user approval before proceeding.

## Requirements

### Requirement: Agent confirms before destructive operations
The agent skill SHALL instruct AI agents to ask for user confirmation before executing any destructive wallet operation. Destructive operations include commands that permanently delete, archive, or alter financial data. The agent MUST show a brief summary of what will be affected and wait for explicit user approval before proceeding.

#### Scenario: Agent confirms before deleting a transaction
- **WHEN** a user asks to delete or remove a transaction (e.g., "delete transaction #5")
- **THEN** the agent SHALL show the transaction details (amount, description, account) and ask for confirmation before running `wallet rm`

#### Scenario: Agent confirms before archiving an account
- **WHEN** a user asks to archive an account
- **THEN** the agent SHALL show the account name and ask for confirmation before running `wallet account archive`

#### Scenario: Agent confirms before adjusting an account balance
- **WHEN** a user asks to adjust an account balance
- **THEN** the agent SHALL show the account name, target amount, and description before running `wallet adjust`

#### Scenario: Agent confirms before deleting a bill
- **WHEN** a user asks to delete a planned payment
- **THEN** the agent SHALL show the bill name and amount and ask for confirmation before running `wallet bill rm`

#### Scenario: Agent confirms before deleting a budget
- **WHEN** a user asks to delete a budget
- **THEN** the agent SHALL show the budget name and ask for confirmation before running `wallet budget rm`

#### Scenario: Agent confirms before deleting a category
- **WHEN** a user asks to delete a category
- **THEN** the agent SHALL show the category name and ask for confirmation before running `wallet category rm`

#### Scenario: Agent confirms before deleting a tag
- **WHEN** a user asks to delete a tag
- **THEN** the agent SHALL show the tag name and ask for confirmation before running `wallet tag rm`

#### Scenario: Agent confirms before deleting an exchange rate
- **WHEN** a user asks to delete an exchange rate
- **THEN** the agent SHALL show the currency code and ask for confirmation before running `wallet rate rm`

#### Scenario: No confirmation needed for read-only queries
- **WHEN** a user asks for a read-only query (list, report, forecast, bill due, budget check, account list, category list, tag list, rate list)
- **THEN** the agent SHALL execute the command without asking for confirmation

#### Scenario: No confirmation needed for data creation
- **WHEN** a user asks to create data (add transaction, set budget, add bill, add category, add tag, add rate)
- **THEN** the agent SHALL execute the command without asking for confirmation

#### Scenario: No confirmation needed for non-destructive mutations
- **WHEN** a user asks to edit, pay, skip, pause, or resume a bill
- **THEN** the agent SHALL execute the command without asking for confirmation

#### Scenario: Batch operations require scope description
- **WHEN** a user asks for a batch destructive operation (e.g., "delete all transactions")
- **THEN** the agent SHALL describe the scope of what will be deleted before asking for confirmation
