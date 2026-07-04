## Context

The project is a personal finance CLI in Go with SQLite, currently under development. Core data model, project skeleton, and CRUD operations are designed but implementation is ongoing. The project currently lacks any documentation files at the repo root — no README, LICENSE, CONTRIBUTING.md, or AGENTS.md.

Stakeholders: users (need to understand what the tool does and how to use it), contributors (need conventions and setup instructions), AI agents (need architecture and pattern context).

## Goals / Non-Goals

**Goals:**
- Provide a standard OSS README covering features, install, usage, and configuration
- Establish MIT license with proper copyright attribution
- Give contributors clear setup, conventions, and PR process guidance
- Equip AI agents with architecture overview, patterns, and build commands

**Non-Goals:**
- User-facing documentation website (future)
- Auto-generated or tool-synced documentation
- API reference docs (CLI is self-documenting via `--help`)

## Decisions

1. **License: MIT** — Permissive, simple, widely understood. Matches open-source intent for a personal tool.

2. **README structure: Standard OSS** — Title, features, install, quick start, config, commands table, contributing link, license. Covers what users need without over-engineering.

3. **AGENTS.md scope: Full context** — Architecture overview, conventions, patterns, file structure. AI agents need to understand patterns not just commands.

4. **Single capability: `project-documentation`** — All four files (README, LICENSE, CONTRIBUTING, AGENTS.md) are tightly coupled and reference each other. Grouping as one capability avoids partial-file states.

## Risks / Trade-offs

- **AGENTS.md may drift from code** → Mitigation: AGENTS.md captures stable patterns and architecture, not line-level detail. Review during significant refactors.
- **Documentation references planned features not yet built** → Mitigation: README and AGENTS.md describe intended state; features will be implemented before public release.
