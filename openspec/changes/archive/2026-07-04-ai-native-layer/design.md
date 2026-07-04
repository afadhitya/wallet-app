## Context

The wallet CLI already supports human-oriented command output and several specs mention JSON output on individual commands. The AI-native layer needs to turn that into a consistent cross-command contract so an agent can always request `--json`, parse the same top-level response shape, inspect metadata, and handle failures from stable error codes.

The repository also needs a version-controlled agent skill so Hermes or another AI agent can map natural language finance requests to wallet CLI invocations without embedding application-specific logic outside the project.

## Goals / Non-Goals

**Goals:**
- Add one global `--json` flag that every command can use.
- Return successful JSON as a stable envelope containing `success`, `data`, and `meta`.
- Return JSON errors as a stable envelope containing `success: false` and an `error` object.
- Preserve current table/prose output when `--json` is absent.
- Add a repository-local `skill/SKILL.md` with wallet command mapping guidance for AI agents.
- Verify the implementation with linting and unit/CLI coverage; document any intentionally excluded hard-to-test paths.

**Non-Goals:**
- Add a network API or background service.
- Change wallet domain behavior such as transaction persistence, budget calculation, bill scheduling, or forecast algorithms except where JSON rendering exposes existing results.
- Add agent-side business logic beyond command selection, `--json` execution, response parsing, and friendly response formatting.
- Support automatic tag/category creation from the agent skill.

## Decisions

- Implement JSON as shared CLI rendering helpers.
  - Rationale: response shape, timestamps, command names, and error formatting must stay consistent across commands.
  - Alternative considered: each command hand-assembles JSON. That would likely drift and make agent parsing brittle.

- Keep `--json` as a persistent Cobra flag on the root command.
  - Rationale: agents can append the same flag to any wallet command and existing command-specific flags remain unchanged.
  - Alternative considered: separate `wallet json <command>` wrapper. That would add command indirection and diverge from normal CLI usage.

- Use command result structs or maps as `data` while keeping the envelope generic.
  - Rationale: command payloads vary, but the top-level parse path should be identical for all commands.
  - Alternative considered: a single universal data schema. That would either be too sparse or force unrelated commands into awkward fields.

- Convert command errors through centralized error rendering when `--json` is active.
  - Rationale: AI agents need structured codes such as `CATEGORY_NOT_FOUND`, `INVALID_AMOUNT`, and `DB_ERROR`, not only prose messages.
  - Alternative considered: encode only the raw Go error string. That would be less actionable and harder to branch on safely.

- Store the agent skill under `skill/SKILL.md` in the repository.
  - Rationale: the skill stays versioned with the CLI contract and can be installed or symlinked by different AI providers.
  - Alternative considered: manage the skill only in a user-local agent directory. That would make it easy for the skill to drift from the app.

## Risks / Trade-offs

- Incomplete command coverage could leave some commands emitting text under `--json` -> add CLI tests that exercise representative commands across core CRUD, budgets, bills, and forecasts.
- Existing command-specific JSON payload expectations may change when wrapped in an envelope -> update tests and specs around the envelope as the stable external contract.
- Error code mapping can become inconsistent if callers return untyped errors -> centralize common error classification and add focused tests for validation, not-found, and database-style failures where feasible.
- Some OS-level failure branches may be hard to test deterministically -> keep business logic covered and document any coverage exclusions with rationale.
- Agent skill examples can become stale as commands evolve -> keep examples short, command-based, and covered by the same documented CLI contract.
