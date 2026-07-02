# Research: Wallet by BudgetBakers

> Compiled: 2026-07-02 | Source: budgetbakers.com, Google Play, Finder UK, Beebom, WealthyPot, Finny

---

## Overview

**Wallet by BudgetBakers** is a personal & family finance manager with smart budgeting, expense tracking, bank sync, and planned payments. Launched by BudgetBakers s.r.o. (Czech Republic). Available on iOS, Android, and Web.

- **Play Store**: 4.6★ (377K+ votes), Editor's Choice
- **Business model**: Freemium (free with ads + manual entry; Premium for bank sync)
- **Pricing (2026)**: ~$3.99/mo or ~$23.99/yr (Premium)

---

## Feature Map

### 1. Smart Budgets
- AI-powered category assignment
- Spending limits per category with progress bars
- Proactive overspending alerts
- Multiple budget types
- Budget overview showing category vs. limit (e.g., "Food & Dining $450 / $500")
- Month-over-month comparisons with tips ("You spent 15% more on dining out")

### 2. Bank Synchronization (Premium)
- **15,000+** bank connections worldwide (via third-party aggregator — likely Salt Edge / Plaid)
- Real-time transaction sync + auto-categorization
- Balance updates
- Bank-level encryption
- Multiple accounts per bank (checking, savings, credit cards)
- Pending transactions may not sync (bank-dependent)

### 3. Expense Tracking
- Automatic categorization with ML/AI
- Custom categories + subcategories
- Spending trends over time
- Percentage breakdown by category
- Personalized recommendations/insights
- Manual transaction entry (free tier)

### 4. Planned Payments (Bill Tracker)
- Recurring payment detection (auto-detects patterns from history)
- Upcoming payment calendar
- Cash flow forecast ("expected balance after upcoming payments")
- Subscription/bill tracking
- Unused subscription detection ("Eliminate Waste")
- Due date reminders/notifications

### 5. Multi-Currency
- Track accounts in different currencies
- Budget limits in multiple currencies
- Exchange rate handling

### 6. Sharing & Collaboration
- Share selected accounts with spouse/family/friends/colleagues
- Collaborative budget management
- Permissions control per shared account

### 7. Reports & Visualization
- Charts and graphs by category/period
- Income vs. expense breakdown
- Net worth tracking
- Export: CSV, PDF
- Templates for recurring entries

### 8. Labels & Organization
- Custom labels/tags on transactions
- Notes on transactions
- Templates for frequent transactions

### 9. Investments (limited)
- Basic investment account tracking
- Manual portfolio value updates

### 10. Cross-Platform Sync
- iOS ↔ Android ↔ Web
- Real-time sync across devices
- Single account login

---

## Architecture Observations (Inferred)

| Layer | Likelihood | Notes |
|-------|-----------|-------|
| **Backend** | Python/Node.js? | Unknown — no public tech stack. Likely REST API + WebSocket for sync |
| **Mobile** | Native (Kotlin/Swift) | Separate iOS & Android apps, not cross-platform |
| **Web** | React or Angular | "Open web app" button suggests a PWA or SPA |
| **Database** | PostgreSQL | Multi-tenant SaaS pattern |
| **Bank Sync** | Salt Edge / Plaid / Yodlee | Third-party aggregator for 15K+ banks |
| **AI/ML** | In-house or API | "AI-powered category assignment" & spending insights |
| **Auth** | Email + OAuth (Google/Apple) | Standard SaaS auth |
| **Sync** | Real-time | Cross-device sync with conflict resolution |

---

## Competitors

| App | Open Source | Self-Hosted | Bank Sync | Key Differentiator |
|-----|-------------|-------------|-----------|-------------------|
| **Wallet (BudgetBakers)** | ❌ | ❌ | ✅ 15K+ banks | Planned payments, AI categorization |
| **YNAB** | ❌ | ❌ | ✅ | Zero-based budgeting methodology |
| **Firefly III** | ✅ | ✅ | Via importers | Double-entry bookkeeping, self-hosted |
| **Actual Budget** | ✅ | ✅ | Via GoCardless | Envelope budgeting, local-first |
| **Money Manager** | ❌ | ❌ | ❌ | Simple, manual, popular in Asia |
| **Monely** | ❌ | ❌ | ❌ | Lower price, ad-free all plans |

---

## What Makes Wallet Stand Out

1. **Planned Payments engine** — auto-detects recurring patterns and forecasts future cash flow
2. **AI categorization** — learns from user behavior, suggests categories
3. **15,000+ bank connections** — wide coverage via aggregator
4. **Polished UI/UX** — consistently praised in reviews for beautiful design
5. **Family sharing** — selected account sharing with granular control
6. **All-platform** — web + iOS + Android, real-time sync

---

## Pain Points (from reviews)

- Bank sync sometimes misses pending transactions
- Premium subscription required for bank sync
- Limited investment tracking (manual only)
- No Indonesian bank support via Salt Edge/Plaid (relevance for local market)
- No open-source / self-hosted option
- Data locked in proprietary cloud

---

## Key Takeaways for Our Wallet App

**Opportunities:**
1. **Self-hosted / local-first** — data privacy, no subscription lock-in
2. **Indonesian bank integration** — leverage local APIs (BCA, Mandiri, etc.)
3. **Offline-first** — work without internet, sync when available
4. **Open source** — community contributions, customization
5. **Better investment tracking** — auto-fetch prices, reksadana, saham Indonesia
6. **Flexible architecture** — plugin system for bank connectors

**Risks:**
- Bank sync is the hardest part (regulatory, security, maintenance)
- Competing with polished UX of established apps
- Indonesian bank APIs are fragmented/non-standard
