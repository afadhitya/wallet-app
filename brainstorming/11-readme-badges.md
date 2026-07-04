# 11 — README Badges

> Depends on: [10-documentation](./10-documentation.md)
> Status: 🔴 pending review | Unblocks: README finalization

---

## Objective

Add professional badges to README.md that show project health at a glance.

---

## Decisions

### D1: Badge Selection
| Badge | Source | Status | Include? |
|-------|--------|--------|----------|
| Go Version | shields.io | ✅ Active | ✅ Yes |
| CI Status | GitHub Actions | ✅ Active | ✅ Yes |
| Code Coverage | Codecov | ✅ Active | ✅ Yes |
| Go Report Card | goreportcard.com | ⚠️ Uses deprecated golint | ❌ No — CI uses golangci-lint |
| License | shields.io | ✅ Active | ✅ Yes |
| Release | GitHub Releases | ✅ Active | ✅ Yes |
| Last Commit | shields.io | ✅ Active | ❌ No — not essential |
| PRs Welcome | shields.io | ✅ Active | ❌ No — not essential |

→ **5 badges** — Go Version, CI Status, Code Coverage, License, Release.

### D2: Badge Style
| Option | Description |
|--------|-------------|
| **A: Flat** | Clean, minimal |
| B: Flat Square | Slightly more modern |
| C: For-the-badge | Large, prominent |

→ **A — Flat style.** Matches minimal project aesthetic.

### D3: Badge Order
| Option | Description |
|--------|-------------|
| A: Build → Coverage → Version → License → Release | Health first |
| **B: Version → Build → Coverage → License → Release** | Identity first |
| C: License → Version → Build → Coverage → Release | Legal first |

→ **B — Version first.** Shows what version users are getting.

---

## Badges

### Markdown

```markdown
# wallet

[![Go Version](https://img.shields.io/github/go-mod/go-version/afadhitya/wallet-app?style=flat)](https://go.dev/)
[![CI](https://github.com/afadhitya/wallet-app/actions/workflows/go-quality.yml/badge.svg?style=flat)](https://github.com/afadhitya/wallet-app/actions/workflows/go-quality.yml)
[![Coverage](https://codecov.io/gh/afadhitya/wallet-app/branch/main/graph/badge.svg?style=flat)](https://codecov.io/gh/afadhitya/wallet-app)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat)](LICENSE)
[![Release](https://img.shields.io/github/v/release/afadhitya/wallet-app?style=flat)](https://github.com/afadhitya/wallet-app/releases)
```

### Visual Preview

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ # wallet                                                                     │
│                                                                              │
│ [Go 1.22] [CI: passing] [Coverage: 100%] [License: MIT] [Release: v1.0.0]   │
│                                                                              │
│ A personal finance CLI for tracking expenses, budgets, and planned payments. │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Badge Details

| Badge | URL Pattern | Data Source |
|-------|-------------|-------------|
| Go Version | `img.shields.io/github/go-mod/go-version/{owner}/{repo}` | `go.mod` go directive |
| CI | `github.com/{owner}/{repo}/actions/workflows/{file}/badge.svg` | GitHub Actions status |
| Coverage | `codecov.io/gh/{owner}/{repo}/branch/main/graph/badge.svg` | Codecov reports |
| License | `img.shields.io/badge/License-MIT-blue.svg` | Static |
| Release | `img.shields.io/github/v/release/{owner}/{repo}` | GitHub Releases |

---

## Setup Required

| Step | Action | When |
|------|--------|------|
| 1 | Codecov integration | After CI setup, upload coverage to Codecov |
| 2 | First release | Create GitHub Release with tag |

**Codecov Setup (CI workflow addition):**

```yaml
- name: Upload coverage
  uses: codecov/codecov-action@v4
  with:
    files: ./coverage.out
```

---

## Dependencies

- Phase 10: `README.md` structure
- Phase 02: GitHub Actions CI workflow
- External: Codecov account (free for open source)

---

## Ready to Review

Check:
- [ ] 5 badges OK?
- [ ] Flat style OK?
- [ ] Version-first order OK?
- [ ] Codecov setup acceptable?
