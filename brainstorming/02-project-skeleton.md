# 02 — Project Skeleton

> Depends on: [01-data-model](./01-data-model.md)
> Status: ✅ design approved | Unblocks: 03-core-crud

---

## Objective

Initialize the Go project with proper module structure, CLI framework, configuration, build tooling, and SQLite integration — the foundation everything else builds on.

---

## Design Decisions

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| P1 | Module path | `github.com/afadhitya/wallet-app` | Standard Go module, installable |
| P2 | CLI framework | `spf13/cobra` | Most popular Go CLI, subcommands, completion |
| P3 | SQLite driver | `modernc.org/sqlite` | Pure Go (no CGO), simpler cross-compilation |
| P4 | Config format | `TOML` in `~/.config/wallet/config.toml` | Human-readable, standard XDG path |
| P5 | DB location | `~/.local/share/wallet/wallet.db` | XDG data dir, single file |
| P6 | Build | `Makefile` + `go build` | Simple, no goreleaser needed yet |
| P7 | Testing | stdlib `testing` + `testify/assert` | Standard Go testing toolkit |
| P8 | Migrations | Embedded SQL files + version tracking | Simple, no external migration tool |
| P9 | Code generation | `sqlc` for repository layer | Type-safe SQL queries, no ORM overhead |
| P10 | CI | GitHub Actions, 100% test coverage gate | Enforced on PR to main, coverage.html on failure |

---

## Project Structure

```
wallet-app/
├── cmd/
│   └── wallet/
│       └── main.go              # Entry point
├── internal/
│   ├── db/
│   │   ├── db.go               # Connection, migrations
│   │   └── migrations/
│   │       └── 001_initial.sql  # Full schema
│   ├── query/                   # sqlc SQL input files
│   │   ├── accounts.sql
│   │   ├── categories.sql
│   │   ├── tags.sql
│   │   ├── transactions.sql
│   │   ├── budgets.sql
│   │   └── planned_payments.sql
│   ├── gen/                     # sqlc-generated Go code
│   │   ├── db.go               # Querier interface
│   │   ├── models.go           # Data structs
│   │   └── *.sql.go            # Query implementations
│   ├── service/                 # Business logic
│   │   ├── account_svc.go
│   │   ├── transaction_svc.go
│   │   ├── budget_svc.go
│   │   └── planned_payment_svc.go
│   └── cli/                     # Cobra commands
│       ├── root.go             # wallet
│       ├── init.go             # wallet init
│       ├── add.go              # wallet add (expense/income/transfer)
│       ├── list.go             # wallet list
│       ├── category.go         # wallet category
│       ├── tag.go              # wallet tag
│       ├── budget.go           # wallet budget
│       ├── bill.go             # wallet bill (planned payments)
│       ├── report.go           # wallet report
│       └── forecast.go         # wallet forecast
├── sqlc.yaml                    # sqlc configuration
├── go.mod
├── go.sum
├── Makefile
├── .gitignore
└── brainstorming/               # Design docs (existing)
```

### Package rationale

| Package | Purpose | Visibility |
|---------|---------|------------|
| `cmd/wallet` | Single binary entry point | — |
| `internal/db` | Connection pool, migration runner, schema queries | Internal only |
| `internal/query` | Raw SQL files (input to sqlc) | Internal only |
| `internal/gen` | sqlc-generated models + query functions | Internal only |
| `internal/service` | Business rules, validation, multi-step operations | Internal only |
| `internal/cli` | Cobra command tree, arg parsing, output formatting | Internal only |
| `pkg/config` | Config struct, TOML loading, defaults | Public (minimal) |

**Why sqlc?** Generates type-safe Go code from SQL queries. Write SQL → get Go structs + functions. No ORM magic, no reflection. Single-user app doesn't need the overhead of GORM.

---

## CLI Command Tree

```
wallet
├── init                    # Create DB + seed data
├── add
│   ├── expense             # wallet add expense 35000 -c food -a bca "Lunch"
│   ├── income              # wallet add income 5000000 -c salary -a bca "Gaji Juli"
│   └── transfer            # wallet add transfer 100000 --from bca --to gopay
├── list                    # wallet list --month july --category food
├── category
│   ├── list                # wallet category list
│   ├── add                 # wallet category add "Coffee" --parent "Food & Dining"
│   └── edit                # wallet category edit 3 --icon ☕
├── tag
│   ├── list                # wallet tag list
│   ├── add                 # wallet tag add "japan-2026"
│   └── rm                  # wallet tag rm "japan-2026"
├── budget
│   ├── list                # wallet budget list
│   ├── set                 # wallet budget set "Monthly Food" 2000000 -c food -c transport
│   └── check               # wallet budget check --all
├── bill                    # Planned payments
│   ├── list                # wallet bill list
│   ├── add                 # wallet bill add "Netflix" 149000 --monthly --day 15
│   └── due                 # wallet bill due --this-week
├── report                  # wallet report --month july
└── forecast                # wallet forecast --next-month
```

### CLI conventions
- `--json` flag on every command for machine-readable output (Hermes skill target)
- `--account` → `-a`, `--category` → `-c`, `--tag` → `-t`
- Amounts: `35000` = Rp35.000 (minor units), or `35k` shorthand
- Dates: `today`, `yesterday`, `2026-07-02`, `july`

---

## Configuration

```toml
# ~/.config/wallet/config.toml
[database]
path = "~/.local/share/wallet/wallet.db"

[display]
currency = "IDR"
date_format = "2006-01-02"
first_day_of_week = "monday"

[ai]
json_output = true          # default --json for Hermes

[defaults]
account = "bca"             # default account for add expense
```

Loaded via `pkg/config`:

```go
type Config struct {
    Database DatabaseConfig `toml:"database"`
    Display  DisplayConfig  `toml:"display"`
    AI       AIConfig       `toml:"ai"`
    Defaults DefaultsConfig `toml:"defaults"`
}
```

---

## Build Tooling

### Makefile

```makefile
.PHONY: build run test clean install

BINARY=wallet
MODULE=github.com/afadhitya/wallet-app

build:
	go build -o bin/$(BINARY) ./cmd/wallet

run: build
	./bin/$(BINARY)

test:
	go test ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/

install: build
	cp bin/$(BINARY) /usr/local/bin/

deps:
	go mod tidy
	go mod verify
```

### .gitignore

```
/bin/
/coverage.out
/coverage.html
*.db
*.db-journal
*.db-wal
```

---

## Database Initialization

### Migration strategy

Single-file migration for MVP — just run the full schema.

```go
// internal/db/db.go
func Open(path string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_foreign_keys=on")
    if err != nil {
        return nil, err
    }
    return db, nil
}

func Migrate(db *sql.DB) error {
    // Check version table
    // Run embedded SQL files in order
    // Update version
}
```

### `001_initial.sql`

- Embedded via `embed.FS`
- Creates all 10 tables from Phase 01
- Seeds 24 default categories
- WAL mode, foreign keys ON

---

## Dependencies

```
go.mod:
  github.com/spf13/cobra      v1.8+      # CLI framework
  modernc.org/sqlite           v1.30+     # Pure Go SQLite driver
  github.com/BurntSushi/toml   v1.4+      # TOML config parsing
  github.com/stretchr/testify  v1.9+      # Test assertions (dev only)

sqlc (dev tool):
  github.com/sqlc-dev/sqlc     v1.27+     # SQL → Go code generator
```

### sqlc.yaml

```yaml
version: "2"
sql:
  - engine: "sqlite"
    queries: "internal/query/"
    schema: "internal/db/migrations/"
    gen:
      go:
        package: "gen"
        out: "internal/gen/"
        sql_package: "database/sql"
        emit_json_tags: true
        emit_db_tags: true
        emit_empty_slices: true
        emit_result_struct_pointers: true
        emit_interface: true
```

**Why `modernc.org/sqlite` over `mattn/go-sqlite3`?**
- Pure Go — no CGO, no C compiler needed
- Cross-compiles to ARM (Raspberry Pi, etc.) trivially
- Slightly slower but single-user app won't notice

---

## Hermes Skill Integration Points

Designed for future Phase 08 skill:

```
wallet add expense 35000 -c food -t lunch "Nasi Goreng" --json
wallet list --month this --json
wallet budget check --all --json
wallet bill due --this-week --json
wallet forecast --next-month --json
```

`--json` outputs structured data Hermes can parse directly. No screen-scraping needed.

---

## Resolved Questions

| # | Question | Resolution |
|---|----------|------------|
| OQ1 | TUI (bubbletea/textual)? | ❌ Skip for now |
| OQ2 | Binary name? | `wallet` |
| OQ3 | Hand-written structs or sqlc? | sqlc-generated |
