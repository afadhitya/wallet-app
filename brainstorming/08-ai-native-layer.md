# 08 — AI-Native Layer

> Depends on: [02-project-skeleton](./02-project-skeleton.md), [03-core-crud](./03-core-crud.md)
> Status: 🔴 pending review | Unblocks: 09-reports

---

## Objective

Make the wallet app AI-native — designed for Hermes skill to interact via CLI with structured JSON output. Every command supports `--json` for machine parsing.

---

## Design Decisions

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| A1 | JSON output | `--json` flag on every command | Explicit, Hermes adds flag, user sees table by default |
| A2 | Skill scope | All commands wrapped | Skill is simple wrapper, no complex logic |

---

## JSON Output Contract

### Standard JSON envelope

Every command with `--json` outputs a consistent structure:

```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "command": "list",
    "timestamp": "2026-07-02T10:30:00Z"
  }
}
```

### Error JSON

```json
{
  "success": false,
  "error": {
    "code": "CATEGORY_NOT_FOUND",
    "message": "Category 'foo' not found. Did you mean 'food'?",
    "suggestion": "food"
  }
}
```

---

## Command JSON Examples

### `wallet list --json`

```json
{
  "success": true,
  "data": {
    "transactions": [
      {
        "id": 42,
        "date": "2026-07-01",
        "type": "expense",
        "amount": 35000,
        "currency": "IDR",
        "base_amount": null,
        "base_currency": null,
        "description": "Lunch at Warung",
        "notes": null,
        "category": {
          "id": 5,
          "name": "Restaurant",
          "parent": "Food & Dining"
        },
        "account": {
          "id": 1,
          "name": "BCA"
        },
        "tags": ["lunch", "work"],
        "is_planned": false
      }
    ],
    "total": 285000,
    "count": 15
  }
}
```

### `wallet budget check --json`

```json
{
  "success": true,
  "data": {
    "budgets": [
      {
        "id": 1,
        "name": "Monthly Food",
        "limit": 2000000,
        "spent": 1250000,
        "remaining": 750000,
        "percent_used": 62.5,
        "status": "ok",
        "period_start": "2026-07-01",
        "period_end": "2026-07-31"
      }
    ]
  }
}
```

### `wallet bill due --json`

```json
{
  "success": true,
  "data": {
    "due": [
      {
        "id": 2,
        "name": "Gym Membership",
        "amount": 500000,
        "currency": "IDR",
        "due_date": "2026-07-05",
        "overdue_days": 1,
        "category": "Fitness"
      }
    ],
    "total_due": 500000,
    "count": 1
  }
}
```

### `wallet forecast --json`

```json
{
  "success": true,
  "data": {
    "horizon_months": 1,
    "forecasts": [
      {
        "month": "2026-08",
        "starting_balance": 15000000,
        "projected_income": 5000000,
        "projected_expense": 2649000,
        "ending_balance": 17351000,
        "planned_payments": [
          {"name": "Netflix", "amount": 149000, "date": "2026-08-15"},
          {"name": "Gym", "amount": 500000, "date": "2026-08-05"},
          {"name": "Rent", "amount": 2000000, "date": "2026-08-01"}
        ]
      }
    ]
  }
}
```

---

## Hermes Skill

### Skill file: `~/.hermes/skills/productivity/wallet/SKILL.md`

```yaml
---
name: wallet
description: "Personal finance tracker — record expenses, check budgets, manage bills, forecast cash flow. Use when user mentions money, expenses, budget, bills, or financial tracking."
tags: [finance, budget, money, expenses, bills]
---
```

### Trigger words
- "expense", "income", "spent", "earned", "paid"
- "budget", "balance", "report"
- "bill", "subscription", "recurring"
- "forecast", "cash flow", "upcoming"
- "wallet"

### Skill behavior

1. Parse user intent → map to wallet command
2. Execute `wallet <command> <args> --json`
3. Parse JSON response
4. Format friendly response for user

### Example interactions

**User:** "Catat pengeluaran makan siang 35rb"
**Hermes:** `wallet add expense 35000 "Lunch" -c food --json` → "Tercatat: Rp35.000 — Makan siang (Food & Dining)"

**User:** "Berapa sisa budget food bulan ini?"
**Hermes:** `wallet budget check --budget "Monthly Food" --json` → "Budget Food: Rp750.000 tersisa dari Rp2.000.000 (62.5% terpakai)"

**User:** "Tagihan apa yang jatuh tempo minggu ini?"
**Hermes:** `wallet bill due --this-week --json` → "2 tagihan: Gym Rp500.000 (terlambat 1 hari), Netflix Rp149.000 (15 Jul)"

**User:** "Prediksi cash flow bulan depan"
**Hermes:** `wallet forecast --json` → "Agustus: Saldo akhir Rp17.351.000. 3 tagihan jatuh tempo: Netflix, Gym, Rent."

---

## CLI Implementation

### Cobra `--json` flag

```go
// cmd/wallet/root.go
var jsonOutput bool

func init() {
    rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
}

func outputJSON(data interface{}) {
    result := map[string]interface{}{
        "success": true,
        "data":    data,
        "meta": map[string]interface{}{
            "command":   cmd.Name(),
            "timestamp": time.Now().UTC().Format(time.RFC3339),
        },
    }
    json.NewEncoder(os.Stdout).Encode(result)
}

func outputError(code, message string, suggestion string) {
    result := map[string]interface{}{
        "success": false,
        "error": map[string]interface{}{
            "code":       code,
            "message":    message,
            "suggestion": suggestion,
        },
    }
    json.NewEncoder(os.Stdout).Encode(result)
    os.Exit(1)
}
```

### Command pattern

```go
// cmd/wallet/list.go
func runList(cmd *cobra.Command, args []string) {
    txns, err := txnService.List(ctx, filters)
    if err != nil {
        outputError("LIST_FAILED", err.Error(), "")
    }

    if jsonOutput {
        outputJSON(map[string]interface{}{
            "transactions": txns,
            "total":        calcTotal(txns),
            "count":        len(txns),
        })
        return
    }

    // Table output
    printTable(txns)
}
```

---

## Error Codes

| Code | Meaning |
|------|---------|
| `CATEGORY_NOT_FOUND` | Category doesn't exist |
| `ACCOUNT_NOT_FOUND` | Account doesn't exist |
| `TAG_NOT_FOUND` | Tag doesn't exist |
| `BUDGET_NOT_FOUND` | Budget doesn't exist |
| `BILL_NOT_FOUND` | Bill doesn't exist |
| `INVALID_AMOUNT` | Amount must be positive |
| `INVALID_DATE` | Date format invalid |
| `RATE_NOT_FOUND` | Exchange rate not configured |
| `ALREADY_PAID` | Bill already paid this period |
| `PAUSED_BILL` | Bill is paused |
| `DB_ERROR` | Database error |

---

## Open Questions

| # | Question | Status |
|---|----------|--------|
| OQ1 | Should Hermes skill auto-create tags if user mentions new ones? | → TBD |
| OQ2 | Support `--batch` for multiple transactions at once? | → TBD |
