## Why

The project lacks essential documentation files that set expectations for users, contributors, and AI agents. Without a README, users can't understand what the project does or how to use it. Without AGENTS.md, AI tooling lacks the context needed to work effectively on this codebase.

## What Changes

- Add **README.md** with project overview, features, installation, quick start, configuration reference, command table, and contributing/license links
- Add **LICENSE** (MIT) with copyright attribution
- Add **CONTRIBUTING.md** with development setup, project structure, conventions, PR process, and code generation instructions
- Add **AGENTS.md** with architecture overview, build/test commands, key patterns, database and config details for AI agent context

## Capabilities

### New Capabilities
- `project-documentation`: Standard OSS documentation suite covering README, LICENSE (MIT), CONTRIBUTING, and AGENTS.md

### Modified Capabilities
<!-- No existing capabilities are changing -->

## Impact

- New files at repo root: `README.md`, `LICENSE`, `CONTRIBUTING.md`, `AGENTS.md`
- No code, API, or dependency changes
- Documentation references existing project structure from `brainstorming/01-data-model.md`, `02-project-skeleton.md`, and `03-core-crud.md`
