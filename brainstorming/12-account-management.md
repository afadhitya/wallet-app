# 12 — Account Management

> Depends on: [01-data-model](./01-data-model.md), [02-project-skeleton](./02-project-skeleton.md)
> Status: 🔴 pending review | Unblocks: implementation

---

## Objective

Complete the account management CLI commands. Phase 03 defines the service layer and SQL queries, but the CLI commands are missing.

---

## Commands

### `wallet account add`

Add a new account.

```
$ wallet account add "BCA" --type checking --currency IDR
✓ Account created: BCA (checking) — IDR
```

**Flags:**

| Flag | Short | Required | Default | Description |
|------|-------|----------|---------|-------------|
| `--type` | `-t` | No | checking | Account type |
| `--currency` | `-c` | No | IDR | ISO 4217 currency code |
| `--json` | | No | false | JSON output |

**Account Types:**
- `checking` — bank checking account
- `savings` — bank savings account
- `cash` — physical cash
- `credit_card` — credit card
- `ewallet` — e-wallet (GoPay, OVO, Dana, etc.)

**Validation:**
- Name must be unique (case-insensitive)
- Type must be one of the allowed types
- Currency must be valid ISO 4217 code

---

### `wallet account list`

List all accounts with balances.

```
$ wallet account list
┌────┬────────────────┬──────────┬───────┬──────────────┬─────────────┐
│ ID │ Name           │ Type     │ Curr  │ Balance      │ Status      │
├────┼────────────────┼──────────┼───────┼──────────────┼─────────────┤
│  1 │ BCA Checking   │ checking │ IDR   │ Rp15.000.000 │ active      │
│  2 │ GoPay          │ ewallet  │ IDR   │ Rp250.000    │ active      │
│  3 │ Cash           │ cash     │ IDR   │ Rp500.000    │ active      │
│  4 │ Old Account    │ checking │ IDR   │ Rp0          │ archived    │
└────┴────────────────┴──────────┴───────┴──────────────┴─────────────┘
                                    Total: Rp15.750.000
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--all` | `-a` | Include archived accounts |
| `--json` | | JSON output |

**Default:** Active accounts only.

---

### `wallet account edit`

Edit an existing account.

```
$ wallet account edit 1 --name "BCA Main"
✓ Updated account #1: BCA Main
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--name` | Account name |
| `--type` | Account type |
| `--sort-order` | Display order |
| `--json` | JSON output |

**Validation:**
- Name must be unique if changed
- Cannot change currency (would break balance)

---

### `wallet account archive`

Archive (soft delete) an account.

```
$ wallet account archive 4
✓ Archived account #4: Old Account
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--force` | Skip confirmation |
| `--json` | JSON output |

**Behavior:**
- Sets `is_archived = 1`
- Account disappears from default `wallet account list`
- Account still visible with `wallet account list --all`
- Transactions linked to this account are preserved
- Warn if account has non-zero balance

---

## Validation Summary

| Command | Checks |
|---------|--------|
| `account add` | Unique name, valid type, valid currency |
| `account list` | None (read-only) |
| `account edit` | Unique name (if changed), account exists |
| `account archive` | Account exists, warn if non-zero balance |

---

## Error Handling

| Error | Message | Exit code |
|-------|---------|-----------|
| Duplicate name | `Account 'BCA' already exists.` | 1 |
| Invalid type | `Invalid type 'foo'. Must be: checking, savings, cash, credit_card, ewallet` | 1 |
| Invalid currency | `Invalid currency 'XYZ'. Use ISO 4217 code.` | 1 |
| Account not found | `Account #99 not found.` | 1 |
| Non-zero balance | `Warning: BCA has balance Rp15.000.000. Archive anyway?` | 0 (prompt) |

---

## Dependencies

- Phase 01: `accounts` table schema
- Phase 02: Cobra CLI framework
- Phase 03: `AccountService` + `accounts.sql` queries (already defined)

---

## Ready to Review

Check:
- [ ] Commands cover all CRUD operations?
- [ ] Flags reasonable?
- [ ] Validation rules correct?
- [ ] Error messages helpful?
