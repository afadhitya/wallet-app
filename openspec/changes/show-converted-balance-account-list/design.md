## Context

The `wallet account list` command currently displays a table with columns `ID | Name | Type | Currency | Balance | Status`. The "Balance" column shows each account's raw balance in its own currency. The total row at the bottom converts all balances to the base currency, but individual rows lack a base-currency equivalent, forcing users to mentally convert.

The conversion logic and rate config are already available at the CLI level — `runAccountList` already calls `svc.ListRates()` and computes `totalBalance` using the same per-row conversion. This change adds a "Converted" column to surface that per-row conversion in the table.

## Goals / Non-Goals

**Goals:**
- Add a "Converted" column showing each non-base-currency account's balance converted to base currency
- Base-currency accounts show `-` (no conversion needed)
- Accounts with missing rates show `-` (cannot convert)
- Keep the existing "Balance" column unchanged
- Update tests to cover all column display scenarios

**Non-Goals:**
- Changing `formatAmount` to support non-IDR base currencies (pre-existing limitation)
- Modifying JSON output format (JSON already returns raw data; conversion is a display concern)
- Adding new currency conversion service methods (existing `svc.ListRates()` is sufficient)
- Changing the total row behavior

## Decisions

### 1. Perform conversion entirely in the CLI layer

**Decision:** Compute converted balance inline in `runAccountList` using `baseCurrency` and `rates` already fetched via `svc.ListRates()`.

**Rationale:** The conversion logic for the total already does this per-row. Adding a column display for it requires no new service methods. The CLI is the right layer for display formatting.

**Alternatives considered:**
- *New service method returning enriched structs* — Overkill. The data is already available; a service method would add an unnecessary indirection for a pure display concern.
- *New SQL query joining exchange_rates* — The rates come from a TOML file, not the DB's `exchange_rates` table. SQL approach would be inconsistent with the rate system.

### 2. Column width: 15 characters for "Converted"

**Decision:** Use `%-15s` width matching the existing "Balance" column.

**Rationale:** Converted amounts have the same magnitude as raw balances (they're also int64 amounts formatted with thousand separators). Consistent width keeps the table aligned.

### 3. Base-currency and missing-rate accounts display `-`

**Decision:** Show `-` in the Converted column for accounts where conversion is either unnecessary (already base currency) or impossible (missing rate).

**Rationale:** Blank cells are ambiguous. `-` is a clear "not applicable / unavailable" indicator. This is a common convention in financial tables (balance sheets, ledgers).

Alternatives considered:
- *Show same as raw balance for base-currency accounts* — Redundant, adds visual noise.
- *Blank/empty string* — Ambiguous; could be mistaken for a rendering bug.

### 4. Reuse `formatAmount` for converted values

**Decision:** Use the existing `formatAmount()` function to format the converted balance.

**Rationale:** The converted balance is always in base currency, and `formatAmount` is already used for the total row (also in base currency). The `Rp` prefix hardcoding is a pre-existing limitation (see Non-Goals).

## Risks / Trade-offs

- **[Low] Table may exceed 80-char width on narrow terminals** → The table already has 6 columns; adding a 7th makes it wider. Users with narrow terminals may see line wrapping. Most modern terminals are 120+ chars wide, and the total width (~95 chars) is well within that.
- **[Low] `formatAmount` hardcodes `Rp` prefix** → If base currency is not IDR, the Converted column and total row both show wrong prefix. This is a pre-existing issue tracked separately.

## Migration Plan

No migration needed. This is a display-only change:
1. Add the column to the table header and data rows
2. Existing JSON output is unaffected
3. No database, config, or API changes

## Open Questions

None.
