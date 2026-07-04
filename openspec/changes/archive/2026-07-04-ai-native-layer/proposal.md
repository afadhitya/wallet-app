## Why

The wallet app is intended to be operated by both humans and AI agents, but commands need a consistent machine-readable contract before agents can safely invoke them and parse results. Adding an AI-native CLI layer makes every command usable through structured JSON while preserving human-friendly table output by default.

## What Changes

- Add a global `--json` output mode for every wallet command.
- Standardize successful JSON responses with a `success`, `data`, and `meta` envelope.
- Standardize command failures with structured JSON error codes, messages, and optional suggestions when `--json` is used.
- Add command-specific JSON payloads for transaction listing, budget checks, bill due lists, forecasts, and other wallet commands.
- Add a version-controlled `skill/SKILL.md` agent skill that maps natural language finance requests to wallet CLI commands with `--json`.
- Preserve existing text/table output as the default behavior when `--json` is not passed.

## Capabilities

### New Capabilities
- `ai-native-cli`: Provide structured JSON CLI responses and an agent skill wrapper for AI-assisted wallet usage.

### Modified Capabilities
- `core-crud`: Existing CRUD commands gain a consistent JSON output mode and structured JSON errors.
- `budget-engine`: Budget commands gain a consistent JSON output mode and structured JSON errors.
- `planned-payments`: Bill and planned-payment commands gain a consistent JSON output mode and structured JSON errors.
- `forecasting`: Forecast commands gain a consistent JSON output mode and structured JSON errors.

## Impact

- Affects Cobra command wiring and output helpers across the wallet CLI.
- Adds shared JSON response and error formatting helpers.
- Adds or updates command tests for text output preservation, JSON envelopes, JSON error cases, linting, and coverage expectations.
- Adds `skill/SKILL.md` in the repository so AI agents can invoke the wallet app through stable CLI commands.
