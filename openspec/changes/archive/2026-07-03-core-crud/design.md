## Context

The repository already has a Cobra root command, an embedded SQLite migration system, sqlc configuration, generated DB package location, and an empty `internal/service` package. The initial database schema includes accounts, categories, tags, transactions, transaction tags, budgets, planned payments, and exchange rates, with seeded category data in the first migration.

This change turns the skeleton into the first usable CLI by wiring command handlers to services backed by sqlc queries. It focuses on local SQLite operation, integer minor-unit amounts, ID-or-name lookup for user-facing references, and deterministic text or JSON output.

## Goals / Non-Goals

**Goals:**

- Provide usable CLI commands for wallet initialization, transaction CRUD, category CRUD, tag CRUD, and balance adjustment.
- Keep business rules in services and keep Cobra handlers focused on parsing flags, invoking services, and rendering output.
- Maintain account balances consistently when transactions are created, edited, soft deleted, transferred, or adjusted.
- Preserve scriptability with explicit flags, stable errors, non-zero exits for validation failures, and `--json` output.
- Cover service behavior with in-memory SQLite tests and command behavior with CLI integration tests.

**Non-Goals:**

- Budget, bill, report, forecast, recurring payment, or exchange-rate behavior beyond not breaking the existing schema.
- Interactive transaction entry, automatic tag creation, or recursive category hierarchy.
- Multi-currency conversion logic beyond persisting existing amount and currency fields.
- External synchronization, cloud storage, or multi-user support.

## Decisions

1. Use service methods as the transaction boundary for CRUD operations.

   Service methods will validate inputs, resolve names or IDs, execute database writes, update tag links, and recalculate affected balances. This keeps CLI handlers thin and avoids duplicating rules across text and JSON output paths. The alternative was to implement most logic in Cobra handlers, but that would make service-level unit testing weaker and future interfaces harder to add.

2. Recalculate balances from persisted transactions after each write.

   Account balance updates will derive from non-archived transactions instead of applying only deltas. This is slightly more database work, but it is safer for edits, transfer changes, soft deletes, and adjustment corrections. The alternative was delta-based updates, which is faster but easier to corrupt when a transaction changes account, type, or amount.

3. Model transfers as a single transaction with `account_id` as source and `transfer_to_id` as destination.

   This matches the existing schema and keeps transfer listing/editing simple. Balance recalculation must treat the source as negative and the destination as positive. The alternative was two linked transaction rows, which would require schema support and synchronization logic that is not needed for this phase.

4. Treat balance adjustment as a transaction type plus explicit account balance reconciliation.

   `wallet adjust` computes the difference between current and target balance, records an `adjustment` transaction with the absolute difference and direction implied by the recalculation logic, then brings the account to the target amount. This preserves an audit trail without counting adjustments as income or expense reports later.

5. Use explicit category and tag management commands.

   Tags are not auto-created during transaction entry; missing tags fail with a creation hint. Categories are full CRUD with soft deletion/archive semantics where supported by schema updates or compatible query fields. The alternative was implicit creation for faster entry, but explicit creation avoids typos becoming permanent classification data.

6. Keep output rendering small and deterministic.

   Text output should be human-friendly and stable enough for integration tests. JSON output should return structured entities or error payloads without table formatting. The alternative was adding a richer table rendering dependency immediately, but minimal formatting is enough for the first usable CLI and reduces dependency risk.

7. Treat CI coverage failures as implementation gaps, with narrow documented exclusions.

   The project quality gate requires 100% Go unit test coverage for included, meaningful application behavior. After implementation, any package, branch, helper, error path, or command path reported as uncovered by `go tool cover -func=coverage.out` must be covered by targeted tests, intentionally refactored so unreachable code is removed, or explicitly excluded when it is generated code or hard-to-test OS/infrastructure failure handling. The only acceptable uncovered remainder for this change is the documented 0.9% of CLI init/mkdir/rm/tag infrastructure error branches where deterministic tests would require brittle OS-level fault injection. The alternative was to accept lower coverage broadly, but that would weaken the project quality gate beyond this narrow infrastructure exception.

8. Exclude generated query code from coverage and keep handwritten helpers covered.

   Sqlc-generated `internal/gen` code should be compiled and checked for staleness, but excluded from coverage totals. Handwritten packages such as `internal/testdb`, `internal/cli`, `internal/service`, `internal/db`, and `pkg/config` remain in coverage unless a specific branch is documented as OS/infrastructure failure handling that is not practical to exercise deterministically.

9. Document OS/infrastructure coverage exclusions at the coverage-command boundary.

   The remaining hard-to-test branches are CLI initialization and filesystem/removal/tag infrastructure error paths. These should be excluded by a small, explicit coverage filtering step or package/file/block policy that is auditable in CI, not by silently lowering the threshold for all code. The coverage report should still fail if new uncovered application logic appears outside the documented exclusion set.

## Risks / Trade-offs

- Balance recalculation logic can mishandle transfer direction -> Mitigation: add focused tests for income, expense, transfer source/destination, edit, delete, and adjustment cases.
- The existing schema does not currently include `categories.updated_at` or `categories.is_archived` -> Mitigation: add the smallest compatible migration/query changes needed for soft-delete category behavior, or document a hard-delete fallback only if schema changes are intentionally deferred.
- ID-or-name lookup can be ambiguous when names overlap -> Mitigation: exact ID matches win for numeric input, otherwise exact case-insensitive name match is required before any suggestion is shown.
- CLI integration tests may be brittle if they assert decorative table borders -> Mitigation: assert stable content, exit codes, and JSON payloads rather than every formatting character.
- `wallet init` touches user paths by default -> Mitigation: expose injectable paths or environment/config overrides in tests so integration tests use temporary directories.
- Coverage can remain below 100% even when core behavior works -> Mitigation: inspect the CI coverage profile, add focused tests for uncovered application logic, and document any narrow generated-code or OS/infrastructure exclusions before considering the change complete.
- Generated sqlc packages can appear as uncovered even when indirectly used -> Mitigation: compile and staleness-check `internal/gen`, but exclude it from coverage totals.
- OS-level CLI init/mkdir/rm/tag infrastructure error branches are brittle to test deterministically -> Mitigation: keep successful and validation paths covered, document the exact excluded branches, and fail CI on any uncovered behavior outside that list.

## Migration Plan

1. Add or update sqlc query files under `internal/query` and regenerate `internal/gen`.
2. Add minimal migration changes only if required for category soft delete or timestamp updates.
3. Implement services and command handlers incrementally, keeping each command covered by tests.
4. Run `go fmt`, sqlc generation, `go test ./...`, lint, and coverage checks.
5. If GitHub Actions or local coverage reports less than 100%, use the coverage profile to identify uncovered code, add targeted tests for application behavior, and explicitly filter only generated code or documented OS/infrastructure failure branches that are accepted as untestable.

Rollback is local: remove the new command/service/query changes and any new migration from a development database before release. Once a migration is shipped, rollback requires a follow-up migration instead of editing history.

## Open Questions

None.
