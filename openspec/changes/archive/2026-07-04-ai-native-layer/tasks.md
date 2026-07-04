## 1. Shared JSON Output Infrastructure

- [x] 1.1 Add a persistent root `--json` flag that is available to every wallet command.
- [x] 1.2 Implement shared success response rendering with `success`, `data`, and `meta.command`/`meta.timestamp` fields.
- [x] 1.3 Implement shared JSON error rendering with stable error codes, messages, and optional suggestions.
- [x] 1.4 Add or update error classification helpers for validation, not-found, paused/already-paid bill, invalid date, invalid amount, exchange-rate, and database failures.

## 2. Command JSON Coverage

- [x] 2.1 Update core CRUD commands to render created/updated/deleted/listed transaction, category, tag, account, transfer, and adjustment results through the shared JSON envelope.
- [x] 2.2 Update budget commands to render list, check, set, edit, and removal results through the shared JSON envelope.
- [x] 2.3 Update bill and planned-payment commands to render add, list, due, pay, skip, pause, resume, edit, and removal results through the shared JSON envelope.
- [x] 2.4 Update forecast commands to render balance and bill forecast results through the shared JSON envelope.
- [x] 2.5 Preserve existing text, table, prose, empty-state, and warning output when `--json` is not supplied.

## 3. Agent Skill

- [x] 3.1 Add `skill/SKILL.md` with wallet skill metadata, finance trigger words, and behavior guidance for AI agents.
- [x] 3.2 Document command-mapping examples for expense entry, budget checks, bill due queries, and forecasts using `--json`.
- [x] 3.3 Document that agents must parse the JSON envelope and must not auto-create missing tags implicitly.

## 4. Tests

- [x] 4.1 Add focused unit tests for shared JSON success rendering, metadata generation, error rendering, and error-code classification.
- [x] 4.2 Add CLI tests for representative core CRUD JSON success and error responses while confirming text output remains unchanged without `--json`.
- [x] 4.3 Add CLI tests for representative budget JSON success and error responses.
- [x] 4.4 Add CLI tests for representative bill/planned-payment JSON success and error responses.
- [x] 4.5 Add CLI tests for forecast JSON success and error responses.
- [x] 4.6 Add tests or documentation checks for the repository-local wallet skill file.

## 5. Verification

- [x] 5.1 Run the full Go test suite and fix any failing tests.
- [x] 5.2 Run the project linter and fix reported issues.
- [x] 5.3 Check unit test coverage against the repository coverage gate.
- [x] 5.4 If a path is impractical to test directly, document the exclusion rationale and ensure it is limited to generated-code or deterministic OS/infrastructure failure handling, not business logic or normal JSON rendering.
- [x] 5.5 Run OpenSpec validation for the `ai-native-layer` change.
