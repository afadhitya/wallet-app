## Context

The wallet app already supports local SQLite-backed transaction entry, category/account/tag management, transaction listing, locked base-currency conversion for foreign-currency transactions, text output, and AI-native JSON envelopes. Existing specs also require mixed-currency report totals to use the configured base currency while preserving original-currency context where available.

Reports are the final MVP phase described in `brainstorming/09-reports-export.md`. They should reuse existing transaction data rather than introduce new persistence tables. The TUI portion is intentionally skipped; this change focuses on CLI reports, JSON, and CSV export.

## Goals / Non-Goals

**Goals:**
- Implement `wallet report` monthly summaries with income, expenses, net, transfers, and transaction count.
- Support period filtering by `--month` or explicit `--from` and `--to`, plus optional `--account` filtering.
- Support breakdowns with `--by category`, `--by account`, and `--by tag`.
- Support `--json` output through the shared success envelope.
- Support CSV export with `--export csv` and optional `--output`.
- Use locked `base_amount` values for report totals when transactions are in non-base currencies.
- Add service, sqlc, CLI, rendering, validation, export, lint, unit test, and coverage verification work.

**Non-Goals:**
- TUI report screens.
- Annual `--year` summaries.
- Export formats other than CSV.
- New exchange-rate lookup during report generation; reports use transaction-time locked values.
- Database schema changes unless implementation discovers a missing index that is necessary for acceptable local performance.

## Decisions

1. Add a dedicated report service over sqlc aggregate queries.
   - Rationale: report behavior crosses transactions, categories, accounts, tags, transfers, and currency fields. Keeping aggregation in a service gives the CLI a stable typed API and keeps rendering separate from calculation.
   - Alternative considered: put all calculation in CLI commands. That would make JSON, text, CSV, and tests harder to keep consistent.

2. Use base-currency-equivalent amounts for totals and percentages.
   - Rationale: existing multi-currency requirements say mixed-currency reports are primarily in configured base currency. Each transaction already stores original amount/currency and optional locked base amount/base currency at entry time.
   - Alternative considered: sum original amounts by account currency. That is useful context for account rows, but it cannot produce a single report total across currencies.

3. Treat transfers as their own report total, excluded from income, expenses, net, and breakdown percentages.
   - Rationale: transfers move money between owned accounts and should not inflate income or spending. The brainstorming report displays transfers separately.
   - Alternative considered: include outgoing transfers as expenses. That would misrepresent cash flow for internal account movement.

4. Support account filtering across summaries, breakdowns, and exports.
   - Rationale: users need both whole-wallet reports and account-specific reports. Applying one filter model across report methods avoids inconsistent results.
   - Alternative considered: account filter only on monthly summary. That would make breakdown/export output surprising when combined with `--account`.

5. Implement category breakdown as parent-aware expense rows.
   - Rationale: categories are hierarchical and the brainstorming example shows parent categories with optional child rows. The service can return row metadata that renderers can group or flatten depending on output mode.
   - Alternative considered: only aggregate by leaf category. That is simpler but loses the approved parent-category rollup behavior.

6. Include untagged expenses in tag breakdowns.
   - Rationale: a tag report should account for all expense spending in the period, including transactions with no tag, so percentages add up to the expense total.
   - Alternative considered: omit untagged rows. That hides part of spending and makes percentages less useful.

7. Export transaction rows, not rendered summary tables.
   - Rationale: CSV is most useful as portable raw report data. The approved CSV format lists transaction fields: date, type, amount, currency, base amount, category, account, description, and tags.
   - Alternative considered: export the current rendered report table. That creates different CSV schemas for each `--by` mode and is less universal.

## Risks / Trade-offs

- Category parent rollups can double count if parent and child rows are both summed by consumers -> Clearly document row semantics and compute percentages against total expenses.
- Multi-currency account rows can be confusing when account balances use original currency but report totals use base currency -> Include account currency/original-currency context where available while keeping totals in base currency.
- CSV output can become unstable if tags are unordered -> Sort tags consistently before writing tag lists.
- Month-name parsing can be ambiguous across years -> Resolve month names to the current year and support `YYYY-MM` for explicit year selection.
- SQL aggregates may exclude transactions without optional relationships -> Use left joins where optional data is expected, and cover no-category/no-tag paths in tests when those states are valid.
- Some rendering branches may be difficult to unit test exactly because of table formatting -> Prefer testing service data and stable CLI substrings; document and exclude only genuinely impractical branches from coverage if needed.

## Migration Plan

No schema migration is expected. Implementation should add report sqlc queries, regenerate generated code, add report service methods, add the CLI command and output rendering, add CSV export support, and extend tests. Rollback is code-only: remove report command registration, service code, queries, generated query methods, and tests.

## Open Questions

None.
