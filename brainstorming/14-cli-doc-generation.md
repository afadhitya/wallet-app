# 14 — CLI Documentation Generation

> Depends on: [02-project-skeleton](./02-project-skeleton.md), [10-documentation](./10-documentation.md)
> Status: 🔴 pending review | Unblocks: implementation

---

## Objective

Auto-generate CLI reference documentation from Cobra commands. Keeps docs synced with code, zero manual maintenance.

---

## Decisions

### D1: Doc Format
| Option | Description |
|--------|-------------|
| **A: Markdown** | For GitHub/docs site, easy to read |
| B: Man pages | Traditional Unix, `man wallet` |
| C: Both | Generate both formats |

→ **A — Markdown only.** Target audience is developers/AI agents, not terminal man page users.

### D2: Generation Trigger
| Option | Description |
|--------|-------------|
| **A: Makefile target** | `make docs` — manual run |
| B: CI auto-generate | Generate on every commit |
| C: Pre-commit hook | Generate before commit |

→ **A — Makefile target.** Run manually when commands change. CI generates on release.

### D3: Hidden Commands
| Option | Description |
|--------|-------------|
| A: Include all | Generate docs for hidden commands too |
| **B: Exclude hidden** | Skip hidden commands like `docs` itself |

→ **B — Exclude hidden.** `wallet docs` is internal, shouldn't appear in user docs.

---

## Implementation

### 1. Add `docs` command

```go
// internal/cli/docs.go
package cli

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/spf13/cobra/doc"
)

func newDocsCmd(rootCmd *cobra.Command) *cobra.Command {
    var outputDir string

    cmd := &cobra.Command{
        Use:    "docs",
        Short:  "Generate CLI documentation",
        Hidden: true, // Internal command
    }

    mdCmd := &cobra.Command{
        Use:   "markdown",
        Short: "Generate markdown documentation",
        RunE: func(cmd *cobra.Command, args []string) error {
            if outputDir == "" {
                outputDir = "docs/cli"
            }
            
            // Create output directory
            if err := os.MkdirAll(outputDir, 0755); err != nil {
                return fmt.Errorf("create output dir: %w", err)
            }
            
            // Generate markdown
            if err := doc.GenMarkdownTree(rootCmd, outputDir); err != nil {
                return fmt.Errorf("generate docs: %w", err)
            }
            
            fmt.Printf("✓ Generated CLI docs in %s\n", outputDir)
            return nil
        },
    }

    mdCmd.Flags().StringVarP(&outputDir, "output", "o", "docs/cli", "Output directory")

    cmd.AddCommand(mdCmd)
    return cmd
}
```

### 2. Register command in root.go

```go
// internal/cli/root.go
func NewRootCmd() *cobra.Command {
    rootCmd := &cobra.Command{
        Use:   "wallet",
        Short: "A personal finance CLI",
    }
    
    // ... existing commands ...
    
    // Add docs command (pass rootCmd for generation)
    rootCmd.AddCommand(newDocsCmd(rootCmd))
    
    return rootCmd
}
```

### 3. Makefile target

```makefile
.PHONY: docs
docs:
	@go run cmd/wallet/main.go docs markdown
	@echo "Docs generated in docs/cli/"
```

### 4. .gitignore addition

```gitignore
# Generated docs (regenerate with make docs)
docs/cli/
```

---

## Generated File Structure

```
docs/cli/
├── wallet.md                    # Root command
├── wallet_init.md               # wallet init
├── wallet_add.md                # wallet add
├── wallet_add_expense.md        # wallet add expense
├── wallet_add_income.md         # wallet add income
├── wallet_add_transfer.md       # wallet add transfer
├── wallet_list.md               # wallet list
├── wallet_edit.md               # wallet edit
├── wallet_rm.md                 # wallet rm
├── wallet_adjust.md             # wallet adjust
├── wallet_account.md            # wallet account
├── wallet_account_add.md        # wallet account add
├── wallet_account_list.md       # wallet account list
├── wallet_account_edit.md       # wallet account edit
├── wallet_account_archive.md    # wallet account archive
├── wallet_category.md           # wallet category
├── wallet_tag.md                # wallet tag
├── wallet_budget.md             # wallet budget
├── wallet_bill.md               # wallet bill
├── wallet_forecast.md           # wallet forecast
├── wallet_report.md             # wallet report
└── wallet_rate.md               # wallet rate
```

---

## Example Generated Doc

```markdown
## wallet add expense

Add an expense transaction

### Synopsis

Record a new expense with amount, description, and optional category/account/tags.

```
wallet add expense <amount> <description> [flags]
```

### Examples

```
wallet add expense 35000 "Lunch at Warung" -c food -a bca
wallet add expense 150000 "Groceries" -c groceries -t weekly
```

### Flags

```
  -a, --account string    Account name or ID (default from config)
  -c, --category string   Category name or ID
  -d, --date string       Transaction date (default "today")
      --description string   Alias for positional description
      --help              help for expense
      --json              JSON output
      --notes string      Additional notes
  -t, --tag string        Tag name (repeatable)
```

### SEE ALSO

* [wallet add](wallet_add.md)	 - Add a transaction
```

---

## Testing

| Test | Expected |
|------|----------|
| `make docs` | Generates files in `docs/cli/` |
| `wallet docs markdown -o /tmp/docs` | Generates in custom dir |
| Hidden commands | `wallet docs.md` should NOT exist |
| All visible commands | Every command has a `.md` file |

---

## Dependencies

- Phase 02: Cobra CLI framework
- Phase 10: README references generated docs

---

## Ready to Review

Check:
- [ ] Markdown-only approach OK?
- [ ] Makefile target approach OK?
- [ ] Hidden commands excluded OK?
- [ ] Output directory structure OK?
