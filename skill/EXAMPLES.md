# Wallet CLI — Common Workflows

> Ready-to-use command sequences for common multi-step tasks. All commands use `--json` for structured output.

---

## Recording Expenses

### Single expense
```bash
wallet add expense 50000 "Lunch at Warung" -c Restaurant -a "BCA Checking" --json
```

### Categorized expense with tag
```bash
wallet add expense 350000 "Shinkansen" -c Travel -a bca -t japan-2026 --json
```

### Expense on a past date
```bash
wallet add expense 200000 "Dinner" -c Restaurant -a bca -d 2026-07-01 --json
```

---

## Payday Routine

```bash
# Record salary income
wallet add income 15000000 "Monthly Salary" -c Salary -a "BCA Checking" --json

# Check budget status
wallet budget check --all --json

# See bills due in next 30 days
wallet bill due --next 30 --json

# Forecast next month
wallet forecast -n 1 --json
```

---

## Monthly Checkup

```bash
# Get monthly report
wallet report -m july --json

# Check all budgets
wallet budget check --all --json

# See upcoming bills
wallet bill due --json
```

---

## Subscription Tracking

### Set up subscriptions
```bash
wallet bill add "Netflix" 149000 -c Subscriptions -a bca --monthly --day 15 --json
wallet bill add "Spotify" 54990 -c Subscriptions -a bca --monthly --day 1 --json
wallet bill add "iCloud" 15000 -c Subscriptions -a bca --monthly --day 5 --json
```

### Review subscriptions
```bash
wallet bill list --json
```

### See upcoming subscription charges
```bash
wallet forecast bills -n 3 --json
```

### Pay all due bills
```bash
# Get due bill IDs
wallet bill due --json

# Pay each bill (parse IDs from due output)
wallet bill pay 1 --json
wallet bill pay 2 --json
```

---

## Trip Spending with a Tag

```bash
# Create a trip tag
wallet tag add "japan-2026" --json

# Record trip expenses
wallet add expense 350000 "Shinkansen Tokyo-Kyoto" -c Travel -a bca -t japan-2026 --json
wallet add expense 120000 "Ramen at Ichiran" -c Restaurant -a gopay -t japan-2026 --json
wallet add expense 50000 "Temple entrance" -c Entertainment -a cash -t japan-2026 --json

# Review trip spending
wallet list -t japan-2026 --json

# Get trip report
wallet report -c Travel --json
```

---

## Budget Management

### Set up monthly budgets
```bash
# Food budget across multiple categories
wallet budget set "Food & Dining" 2500000 -c Restaurant -c Groceries -c "Coffee & Snacks" --period monthly --notify 80 --json

# Entertainment budget
wallet budget set "Entertainment" 1000000 -c Entertainment -c "Movies & Shows" --period monthly --notify 80 --json

# Trip budget (one-time, with tags)
wallet budget set "Japan Trip" 10000000 -t japan-2026 --period one_time --from 2026-07-01 --to 2026-07-31 --json
```

### Check and manage budgets
```bash
# Check all budgets
wallet budget check --all --json

# Check specific budget
wallet budget check -b 1 --json

# List all budgets
wallet budget list --json

# Edit a budget amount
wallet budget edit 1 --amount 3000000 --json

# Remove a budget
wallet budget rm 1 --json
```

---

## Multi-Currency Setup

### Configure exchange rates
```bash
wallet rate add USD 15800 --json
wallet rate add EUR 17200 --json
wallet rate add JPY 105 --json
wallet rate add SGD 11800 --json
```

### Create foreign currency account
```bash
wallet account add "USD Savings" --type savings --currency USD --json
```

### Transfer between currencies
```bash
wallet add transfer 100000 --from "BCA Checking" --to "USD Savings" --json
```

---

## Account Setup and Maintenance

### Initialize wallet (first time)
```bash
wallet init --json
```

### Add new accounts
```bash
wallet account add "GoPay" --type ewallet --currency IDR --json
wallet account add "Cash" --type cash --currency IDR --json
```

### Check account balances
```bash
wallet account list --json
```

### Adjust account balance (reconcile)
```bash
wallet adjust "BCA Checking" 5000000 "Reconcile with bank statement" --json
```

---

## Transaction Management

### List recent transactions
```bash
wallet list -n 10 --json
```

### List by category
```bash
wallet list -c Restaurant --json
```

### List by tag
```bash
wallet list -t japan-2026 --json
```

### List by type
```bash
wallet list --type expense --json
```

### Edit a transaction
```bash
wallet edit 42 --amount 75000 --desc "Actually it was more expensive" --json
```

### Archive a transaction (soft delete)
```bash
wallet rm 42 --force --json
```

---

## Tag Management

```bash
# Create tags
wallet tag add "grocery-weekly" --json
wallet tag add "reimbursable" --json

# List all tags
wallet tag list --json

# Remove a tag
wallet tag rm "grocery-weekly" --json
```

---

## Category Management

```bash
# List all categories
wallet category list --json

# Add a custom category (under existing parent)
wallet category add "Coffee Shops" -p "Food & Dining" --json

# Edit a category name
wallet category edit 5 -n "Cafes & Bakeries" --json
```

---

## Forecasting

```bash
# 3-month balance projection
wallet forecast -n 3 --json

# Specific account projection
wallet forecast -n 3 -a "BCA Checking" --json

# Upcoming bill schedule only
wallet forecast bills -n 3 --json
```

---

## Report Export

```bash
# Monthly summary
wallet report -m july --json

# Breakdown by category
wallet report --by category --json

# Export to CSV
wallet report -m july --export csv --json
```
