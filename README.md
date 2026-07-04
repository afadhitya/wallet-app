# Wallet App

A CLI-first personal finance tracker written in Go. Track expenses, income, and transfers across multiple accounts with budgets, recurring bills, multi-currency support, and AI-friendly JSON output.

## Features

- **Transaction tracking** — Record expenses, income, transfers, and balance adjustments
- **Multi-account** — Manage checking, savings, cash, credit cards, and e-wallets
- **Category hierarchy** — 32 built-in categories with 2-level parent-child organization
- **Tags** — Freeform labels for flexible transaction organization
- **Budget engine** — Set recurring or one-time budgets with spending alerts
- **Planned payments** — Schedule recurring bills with daily/weekly/monthly/yearly/custom RRULE patterns
- **Forecasting** — Project future balances and upcoming bill schedules
- **Reports** — Generate monthly summaries with category, account, and tag breakdowns; export to CSV
- **Multi-currency** — Convert between currencies with configurable exchange rates
- **AI-native** — `--json` flag for structured machine-readable output
- **Offline-first** — Single SQLite database, no network required
- **MIT licensed** — Free to use and modify

## Installation

### From source

```sh
git clone https://github.com/afadhitya/wallet-app.git
cd wallet-app
make build
make install
```

### Pre-built binaries

Download the latest binary from the [releases page](https://github.com/afadhitya/wallet-app/releases).

### Prerequisites

- Go 1.25+
- [sqlc](https://sqlc.dev) (for development only: `brew install sqlc` or `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)

## Quick Start

### 1. Initialize

```sh
wallet init
```

Creates the SQLite database and seeds default categories. One account is created by default.

### 2. Add an expense

```sh
wallet add expense 50000 "Lunch at Warung" -c Restaurant -a "BCA Checking"
```

Amounts are in the smallest currency unit — `50000` means Rp 50,000.

### 3. Add income

```sh
wallet add income 10000000 "Monthly salary" -c Salary -a "BCA Checking"
```

### 4. List transactions

```sh
wallet list                    # Current month
wallet list -m 2026-01         # Specific month
wallet list -c Restaurant      # By category
wallet list --type expense     # By type
```

### 5. Transfer between accounts

```sh
wallet add transfer 100000 --from "BCA Checking" --to "GoPay"
```

### 6. Set a budget

```sh
wallet budget set "Food" 2000000 -c Restaurant -c Groceries --period monthly --notify 80
wallet budget check              # Check all active budgets
```

### 7. Schedule a recurring bill

```sh
wallet bill add "Rent" 2500000 -c "Bills & Utilities" -a "BCA Checking" --monthly --day 1
wallet bill due                  # See bills due this month
wallet bill pay 1                # Pay a bill (creates expense transaction)
```

### 8. Generate reports

```sh
wallet report                    # Current month summary
wallet report --by category      # Breakdown by category
wallet report --export csv       # Export to CSV
```

### 9. Forecast future balances

```sh
wallet forecast                  # Next month projection
wallet forecast -n 3             # 3-month projection
wallet forecast bills            # Upcoming bill schedule
```

## Configuration

Configuration is stored at `~/.config/wallet/config.toml`. The app works without it — defaults are used when the file is missing.

```toml
[database]
path = "~/.local/share/wallet/wallet.db"

[display]
currency = "IDR"
date_format = "2006-01-02"
first_day_of_week = "monday"

[ai]
json = true

[defaults]
account = ""
```

### Exchange rates

Exchange rates are stored at `~/.config/wallet/rates.toml`:

```toml
base_currency = "IDR"

[rates]
USD = 15800
SGD = 11800
EUR = 17200
JPY = 105
MYR = 3400
```

Manage rates with `wallet rate {list,add,set,rm}`.

## Commands

| Command | Description |
|---------|-------------|
| `wallet init` | Initialize the wallet database and create config files |
| `wallet add expense <amount> <desc>` | Record an expense |
| `wallet add income <amount> <desc>` | Record income |
| `wallet add transfer <amount>` | Transfer between accounts |
| `wallet list` | List and filter transactions |
| `wallet edit <id>` | Edit a transaction |
| `wallet rm <id>` | Archive a transaction |
| `wallet adjust <account> <target> <desc>` | Set account balance to a target value |
| `wallet category {list,add,edit,rm}` | Manage categories |
| `wallet tag {list,add,rm}` | Manage tags |
| `wallet budget {set,list,check,edit,rm}` | Manage budgets |
| `wallet bill {add,list,due,pay,skip,pause,resume,edit,rm}` | Manage planned payments |
| `wallet report` | Generate financial reports |
| `wallet forecast {bills}` | Forecast balances and bills |
| `wallet rate {list,add,set,rm}` | Manage exchange rates |

All commands support `--help` for detailed usage. Use `--json` for structured JSON output.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, conventions, and the pull request process.

## License

MIT — see [LICENSE](LICENSE).
