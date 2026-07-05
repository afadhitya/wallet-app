# 13 — Multi-Currency Total & Negative Balance

> Depends on: [01-data-model](./01-data-model.md), [03-core-crud](./03-core-crud.md), [07-multi-currency](./07-multi-currency.md), [12-account-management](./12-account-management.md)
> Status: 🔴 pending review | Unblocks: implementation

---

## Objective

Fix two issues found during testing:
1. `wallet account list` total sums raw balances without currency conversion
2. Negative balance should be explicitly allowed (credit cards, loans, debt tracking)

---

## Decisions

### D1: Multi-Currency Total
| Option | Description |
|--------|-------------|
| **A: Convert to base currency** | Single total in IDR using exchange rates |
| B: Show per-currency totals | Separate total per currency |

→ **A — Convert all to base currency.** Quick net worth check.

### D2: Negative Balance
| Option | Description |
|--------|-------------|
| **A: Allow freely** | No check, negative is valid |
| B: Allow with warning | Warn but allow |
| C: Block at zero | Prevent going below 0 |

→ **A — Allow freely.** Credit cards, loans, debt tracking are valid use cases.

---

## Issue 1: Multi-Currency Total

### Current Behavior

```go
// internal/cli/account.go — runAccountList()
totalBalance += acc.Balance  // BUG: sums raw balances regardless of currency
```

If BCA = 15,000,000 IDR and USD Savings = 1,000 USD:
- Current: Total = 15,001,000 (WRONG)
- Expected: Total = 30,800,000 IDR (15M + 1000×15800)

### New Behavior

1. Load base currency + rates from `pkg/config.RateConfig`
2. For each account:
   - If `account.Currency == baseCurrency` → use raw balance
   - Else → `convertedBalance = account.Balance × rates[account.Currency]`
3. Sum all converted balances
4. Display total with base currency label

### Output Example

```
ID   Name                      Type         Currency Balance          Status
---- ------------------------- ------------ -------- --------------- ------
1    BCA Checking              checking     IDR      Rp 15.000.000   active
2    USD Savings               savings      USD      $ 1.000         active
3    GoPay                     ewallet      IDR      Rp 250.000      active
---- ------------------------- ------------ -------- --------------- ------
                                        Total (IDR): Rp 31.050.000
```

### Conversion Formula

```
if account.Currency == baseCurrency:
    convertedBalance = account.Balance
else:
    rate = rates[account.Currency]  // e.g., USD = 15800
    convertedBalance = account.Balance × rate
    // 1000 USD × 15800 = 15,800,000 IDR
```

### Edge Cases

| Case | Behavior |
|------|----------|
| Rate not found | Show account balance normally, exclude from total, print warning: `Warning: No rate for USD, excluded from total` |
| Same currency | No conversion, use raw balance |
| Negative balance | Include in total (can make total negative) |
| All accounts same currency | Still show `(IDR)` label for clarity |

### Code Changes

**File:** `internal/cli/account.go` — `runAccountList()`

```go
// Before
totalBalance += acc.Balance

// After
rateCfg, err := config.LoadRates()
if err != nil {
    return formatError(cmd, fmt.Errorf("load rates: %w", err))
}

var totalBalance int64
var missingRates []string
for _, acc := range accounts {
    // ... existing display code ...
    
    if acc.Currency == rateCfg.BaseCurrency {
        totalBalance += acc.Balance
    } else if rate, ok := rateCfg.Rates[acc.Currency]; ok {
        totalBalance += acc.Balance * rate
    } else {
        missingRates = append(missingRates, acc.Currency)
    }
}

// Print warning if any rates missing
if len(missingRates) > 0 {
    fmt.Fprintf(stdout, "\n⚠️  No rate for %s, excluded from total\n", strings.Join(missingRates, ", "))
}

fmt.Fprintf(stdout, "Total (%s): %s\n", rateCfg.BaseCurrency, formatAmount(totalBalance))
```

---

## Issue 2: Negative Balance

### Current State (Already Works)

| Component | Handles Negative? | Notes |
|-----------|-------------------|-------|
| Schema | ✅ | `INTEGER NOT NULL DEFAULT 0` — no constraint |
| `formatAmount()` | ✅ | Shows `-Rp 50.000` |
| Transfer | ✅ | Warns but allows (doesn't block) |
| `wallet adjust` | ✅ | Can set any value including negative |
| Balance recalculation | ✅ | `GetAccountBalance` SUM works with negatives |

### What Changes

**No code changes needed.** Negative balance already works.

**Documentation update:** Explicitly state in spec/docs that negative balance is allowed and expected for:
- Credit cards (outstanding balance)
- Loans (remaining debt)
- Overdraft accounts
- Manual adjustments to track debt

---

## Testing

### Multi-Currency Total

| Test | Input | Expected |
|------|-------|----------|
| Same currency | 2 IDR accounts | Sum of raw balances |
| Mixed currency | IDR + USD accounts | USD converted × rate |
| Missing rate | IDR + XYZ (no rate) | IDR only, warning for XYZ |
| Negative mixed | IDR -5M + USD 1000 | (-5M) + (1000×15800) = 10.8M |
| All negative | IDR -5M + USD -100 | (-5M) + (-100×15800) = -6.58M |

### Negative Balance

| Test | Input | Expected |
|------|-------|----------|
| Expense exceeds balance | Balance 50K, expense 100K | Balance becomes -50K |
| Adjust to negative | `wallet adjust bca -500000` | Balance = -500,000 |
| Transfer with insufficient | Balance 50K, transfer 100K | Warning (not error), balance goes negative |

---

## Dependencies

- Phase 01: `accounts` table (balance column)
- Phase 07: `pkg/config.RateConfig` for exchange rates
- Phase 12: `wallet account list` command

---

## Ready to Review

Check:
- [ ] Conversion logic correct?
- [ ] Edge cases covered?
- [ ] Negative balance approach OK?
- [ ] Warning message helpful?
