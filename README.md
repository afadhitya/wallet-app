# Wallet App

[![Go Version](https://img.shields.io/github/go-mod/go-version/afadhitya/wallet-app?style=flat)](https://go.dev/)
[![CI](https://github.com/afadhitya/wallet-app/actions/workflows/go-quality.yml/badge.svg)](https://github.com/afadhitya/wallet-app/actions/workflows/go-quality.yml)
[![Coverage](https://codecov.io/gh/afadhitya/wallet-app/branch/main/graph/badge.svg)](https://codecov.io/gh/afadhitya/wallet-app)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat)](LICENSE)
[![Release](https://img.shields.io/github/v/release/afadhitya/wallet-app?style=flat)](https://github.com/afadhitya/wallet-app/releases)

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

### Agent Skill (AI Tools)

Register the entire `skill/` directory with your AI agentic tool so it auto-detects wallet-related queries and uses the correct CLI commands. The directory contains:

- `SKILL.md` — Core principles, intent mapping, and behavioral rules
- `COMMANDS.md` — Domain-grouped command reference with flags and JSON response structures
- `ERRORS.md` — Error codes with meanings, causes, and recovery actions
- `EXAMPLES.md` — Ready-to-use command sequences for common workflows

**Hermes Agent:** Copy the skill directory to your Hermes skills directory:

```sh
mkdir -p ~/.hermes/skills/wallet
cp -r skill/* ~/.hermes/skills/wallet/
```

**OpenClaw:** Copy the skill directory to your OpenClaw skills directory:

```sh
mkdir -p ~/.openclaw/skills/wallet
cp -r skill/* ~/.openclaw/skills/wallet/
```

Once registered, your AI agent will recognize wallet-related queries (expenses, income, budgets, bills, forecasts) and invoke the correct `wallet` CLI commands automatically.

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

Run `wallet --help` for the full command list, or `wallet <command> --help` for details on a specific command. Add `--json` for structured machine-readable output.

Auto-generated CLI reference docs for every command are available in `docs/cli/` (run `make docs` to generate).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, conventions, and the pull request process.

## License

MIT — see [LICENSE](LICENSE).
