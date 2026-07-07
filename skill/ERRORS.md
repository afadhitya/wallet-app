# Wallet CLI — Error Codes & Recovery

> Reference for all error codes returned by the wallet CLI. Every error response follows the JSON envelope:
> ```json
> {"success": false, "error": {"code": "...", "message": "...", "suggestion": "..."}}
> ```
> Always parse `code` for programmatic handling, relay `suggestion` to the user.

## Error Codes Reference

| Code | Meaning | Typical Cause | Recovery Action |
|------|---------|---------------|-----------------|
| `CATEGORY_NOT_FOUND` | Category does not exist | Referenced a category that was never created or was archived | List categories with `wallet category list --json` and suggest closest match or `wallet category add <name>` |
| `ACCOUNT_NOT_FOUND` | Account does not exist | Referenced an account not created in init, or archived | Check setup with `wallet account list --json`, suggest `wallet init` or `wallet account add` |
| `TAG_NOT_FOUND` | Tag does not exist | Referenced a tag that wasn't created | Suggest `wallet tag add <name>` (never auto-create tags) |
| `TRANSACTION_NOT_FOUND` | Transaction ID does not exist | Wrong ID or already archived | List transactions with `wallet list --json` to find the correct ID |
| `BUDGET_NOT_FOUND` | Budget ID does not exist | Wrong ID or budget was deactivated | List budgets with `wallet budget list --all --json` |
| `PLANNED_PAYMENT_NOT_FOUND` | Bill ID does not exist | Wrong ID or bill was removed | List bills with `wallet bill list --all --json` |
| `INVALID_AMOUNT` | Amount is zero or negative | User provided non-positive value | Ask for a positive amount in smallest currency units |
| `INVALID_DATE` | Bad date format | Date not in YYYY-MM-DD format, or month name not recognized | Use YYYY-MM-DD format; month names: "january", "july", etc. |
| `INVALID_INPUT` | Generic validation failure | Various reasons | Read the `message` field for specifics. Common: missing required flag, invalid enum value. |
| `VALIDATION_ERROR` | Validation logic failure | Input violates business rules (e.g., duplicate name) | Read `message` and `suggestion` fields |
| `BILL_PAUSED` | Bill is currently paused | Attempted to pay a paused bill | Tell user to run `wallet bill resume <id>` first |
| `EXCHANGE_RATE_NOT_FOUND` | No rate configured for a currency | Cross-currency operation without a configured rate | Suggest `wallet rate add <currency> <rate>` or tell user the rate is missing |
| `EXCHANGE_RATE_CONFIG_MISSING` | No rate config file exists | First run, no rates.toml created yet | Run `wallet init` to create the config |
| `EXCHANGE_RATE_INVALID` | Rate is negative or zero | Provided an invalid rate value | Rate must be a positive integer representing 1 unit in base currency minor units |
| `UPDATE_NETWORK_ERROR` | Network failure during update | No internet connection or GitHub API unreachable | Check internet connection and retry |
| `UPDATE_PERMISSION_ERROR` | Cannot write updated binary | No write permission for the binary path | Run with appropriate permissions or reinstall |
| `UPDATE_ALREADY_LATEST` | Already running the latest version | Current version matches latest release | No action needed; force with `--force` if reinstall is desired |
| `UPDATE_FAILED` | Update process failed | Download or extraction failed | Check the error message for details; may require manual reinstall |
| `UPDATE_CHECKSUM_MISMATCH` | Binary checksum does not match | Corrupted download or malicious file | Retry update; checksum verification failed |
| `DB_ERROR` | Database connection or query failure | Corrupted database, permission issue, or not initialized | Suggest `wallet init`; check file permissions |
| `INTERNAL_ERROR` | Unexpected internal error | Bug or unhandled edge case | Report the `message` to the user; the specific error detail is included |

---

## Common Recovery Patterns

| Situation | Pattern |
|-----------|---------|
| Unknown category | `wallet category list --json` → find closest match or `wallet category add <name>` |
| Unknown tag | `wallet tag add <name> --json` (do NOT auto-create without asking) |
| Unknown account | `wallet account list --json` → if empty, run `wallet init` |
| Missing exchange rate | `wallet rate add <currency> <rate> --json` → retry original command |
| Paused bill blocking payment | `wallet bill resume <id> --json` → retry `wallet bill pay <id>` |
| Wrong transaction ID | `wallet list -n 20 --json` → find correct ID → retry |
| Database error | `wallet init --json` → retry (if fresh setup), otherwise report |
| Duplicate name error | `VALIDATION_ERROR` with message — read `suggestion`, ask user for a different name |
| Update network error | `wallet update --json` → check internet connection → retry |
| Update permission error | `wallet update --json` → reinstall with `go install` or `make install` |
| Update already latest | No action needed; use `wallet update --force` to reinstall |
| Update failed | `wallet update --json` → read `message` for details → retry or manually reinstall |
| Update checksum mismatch | `wallet update --json` → retry; if persistent, manually download from releases |

---

## Error Handling in Agent Code

```
1. Parse stdout for JSON envelope
2. If `success` is false:
   a. Parse stderr for `error.code`, `error.message`, `error.suggestion`
   b. Match `code` against this table
   c. Apply the recovery action
   d. Relay `suggestion` to the user
3. If `success` is true:
   a. Read `data` and format a human-readable response
   b. Never dump raw JSON to the user
```
