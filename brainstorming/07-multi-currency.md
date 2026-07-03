# 07 — Multi-Currency

> Depends on: [01-data-model](./01-data-model.md), [03-core-crud](./03-core-crud.md)
> Status: 🔴 pending review | Unblocks: 08-ai-native-layer

---

## Objective

Implement multi-currency support — track transactions in different currencies, convert to base currency at transaction time, and report in base currency.

---

## Design Decisions

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| M1 | Rate source | Config file (`~/.config/wallet/rates.toml`) | Manual, user-controlled, offline-friendly |
| M2 | When to convert | At transaction time | Rate locked, fast reports, already in schema |

---

## Configuration

### rates.toml

```toml
# ~/.config/wallet/rates.toml
# Exchange rates relative to base currency (IDR)
# 1 USD = 15700 IDR

[base]
currency = "IDR"

[rates]
USD = 15700
EUR = 17200
SGD = 11800
JPY = 105
MYR = 3400
AUD = 10500
GBP = 20000
```

**Behavior:**
- Base currency defaults to IDR (configurable)
- Rates are "how many base units per 1 foreign unit"
- If rate not found for a currency, error on transaction creation
- User updates rates manually by editing config file

---

## Commands

### `wallet rate list`

Show configured exchange rates.

```
$ wallet rate list
┌──────────┬──────────────┬─────────────┐
│ Currency │ Rate (→ IDR) │ 1 IDR =     │
├──────────┼──────────────┼─────────────┤
│ USD      │ 15.700       │ 0.0000637   │
│ EUR      │ 17.200       │ 0.0000581   │
│ SGD      │ 11.800       │ 0.0000847   │
│ JPY      │ 105          │ 0.0095238   │
└──────────┴──────────────┴─────────────┘
  Base currency: IDR
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--json` | JSON output |

---

### `wallet rate set`

Override a rate in config.

```
$ wallet rate set USD 15800
✓ Updated: 1 USD = 15.800 IDR
```

**Behavior:**
- Updates `~/.config/wallet/rates.toml`
- Does NOT retroactively update existing transactions
- Only affects future transactions

---

### `wallet rate add`

Add a new currency pair.

```
$ wallet rate add KRW 12
✓ Added: 1 KRW = 12 IDR
```

---

### `wallet rate rm`

Remove a currency pair.

```
$ wallet rate rm KRW
✓ Removed: KRW
```

---

## Transaction Integration

### Adding foreign currency transaction

```
$ wallet add expense 100 "Coffee in Singapore" -c food -a dbs-sgd -t travel
✓ Recorded: S$100 (Rp1.180.000) — Coffee in Singapore (Food & Dining) [DBS SGD]
  Rate: 1 SGD = 11.800 IDR
```

**Behavior:**
1. Detect account currency (from `accounts.currency`)
2. If account currency ≠ base currency:
   a. Load rate from `rates.toml`
   b. Calculate `base_amount = amount * rate`
   c. Store: `amount=100, currency=SGD, base_amount=1180000, base_currency=IDR`
3. If account currency = base currency:
   a. Store: `amount=X, currency=IDR, base_amount=NULL, base_currency=NULL`

---

### Listing foreign currency transactions

```
$ wallet list --account dbs-sgd
┌────┬────────────┬──────────────────────┬───────────┬──────────────┐
│ ID │ Date       │ Description          │ Amount    │ IDR Equiv    │
├────┼────────────┼──────────────────────┼───────────┼──────────────┤
│ 42 │ 2026-07-01 │ Coffee in Singapore  │ S$100     │ Rp1.180.000  │
│ 43 │ 2026-07-02 │ Lunch at hawker      │ S$15      │ Rp177.000    │
└────┴────────────┴──────────────────────┴───────────┴──────────────┘
  Total: S$115 (Rp1.357.000)
```

---

### Reporting with mixed currencies

```
$ wallet report --month july
┌─────────────────────────────────────────────────────────────┐
│ Monthly Report — July 2026                                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ Income:       Rp15.000.000                                  │
│   Salary:     Rp15.000.000                                  │
│                                                             │
│ Expenses:     Rp8.500.000                                   │
│   Food:       Rp2.500.000                                   │
│   Transport:  Rp1.200.000                                   │
│   Travel:     Rp1.357.000 (S$115 via SGD)                  │
│   Bills:      Rp3.443.000                                   │
│                                                             │
│ Net:          +Rp6.500.000                                  │
│                                                             │
│ Balances:                                                   │
│   BCA (IDR):  Rp12.000.000                                  │
│   DBS (SGD):  S$500 (Rp5.900.000)                          │
│   Total:      Rp17.900.000                                  │
└─────────────────────────────────────────────────────────────┘
```

**Note:** All amounts in report are in base currency (IDR). Original currency shown in parentheses. Adjustments (`type='adjustment'`) are excluded from income/expense totals.

---

## Service Layer

### CurrencyService

```go
type CurrencyService struct {
    config *config.Config
}

func (s *CurrencyService) GetRate(from, to string) (float64, error)
func (s *CurrencyService) Convert(amount int64, from, to string) (int64, error)
func (s *CurrencyService) ListRates() map[string]float64
func (s *CurrencyService) SetRate(currency string, rate float64) error
func (s *CurrencyService) AddRate(currency string, rate float64) error
func (s *CurrencyService) RemoveRate(currency string) error
func (s *CurrencyService) BaseCurrency() string
```

---

## Error Handling

| Error | Message | Exit code |
|-------|---------|-----------|
| Rate not found | `Exchange rate for KRW not configured. Add it with: wallet rate add KRW <rate>` | 1 |
| Invalid rate | `Rate must be a positive number.` | 1 |
| Config file missing | `Rates config not found. Run 'wallet init' to create default rates.` | 1 |

---

## Open Questions

| # | Question | Status |
|---|----------|--------|
| OQ1 | Should `wallet rate list` also show rates from exchange_rates table? | → TBD |
| OQ2 | Allow override rate per transaction? (e.g., `--rate 15800`) | → TBD |
