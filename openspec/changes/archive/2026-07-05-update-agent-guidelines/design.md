## Context

The wallet app's `skill/SKILL.md` provides comprehensive guidance for AI agents on how to invoke the `wallet` CLI. However, three gaps exist:

1. **CLI `--` separator**: The Cobra CLI uses a `--` separator to end flag parsing. When negative amounts (e.g., `-3612`) are passed as positional arguments, they must appear after `--`; otherwise Cobra parses them as unknown flags. Agents need this documented to avoid constructing incorrect commands.
2. **Database access boundary**: The wallet app has a strict layered architecture. Agents should never bypass the CLI by directly opening the SQLite file or running raw SQL scripts. This boundary should be explicitly stated in the skill file.
3. **Skill installation docs**: Users need to know how to install `skill/SKILL.md` with their AI agentic tools (e.g., Hermes Agent, OpenClaw) so the skill is auto-loaded during relevant sessions.

## Goals / Non-Goals

**Goals:**
- Document the `--` separator syntax in `skill/SKILL.md` with a concrete visual example
- Establish an explicit rule that agents MUST use the `wallet` CLI for data operations, never direct DB manipulation
- Add README instructions for installing the skill file with Hermes Agent and OpenClaw

**Non-Goals:**
- Changing any CLI behavior or command signatures
- Adding programmatic enforcement of the DB access rule
- Modifying AGENTS.md or other project files

## Decisions

### Add both guidelines to the existing "Rules" section in `skill/SKILL.md`

The Rules section (line 222) already contains agent behavior constraints (e.g., "Never auto-create tags, categories, or accounts"). The new guidelines fit naturally here. Both are behavioral rules for agents.

### Use visual diagram for the `--` separator rule

The `--` separator concept is subtle. A visual diagram with annotated parsing ranges makes it immediately clear where flags end and positional args begin.

### Add skill installation to README as a new section under "Installation"

The README already has an "Installation" section. Adding a sub-section for "Agent Skill (AI Tools)" explains how to register the skill file with Hermes Agent and OpenClaw. This keeps discovery close to where users learn about installation overall.

### Keep rules concise

The existing rules are one-liners. New rules should match this style to maintain scannability.

## Risks / Trade-offs

- **Risk**: Agents might ignore the DB access rule if the skill file is not loaded → **Mitigation**: The `wallet` skill has triggers matching finance-related keywords, so it will be auto-loaded in relevant sessions. README installation instructions help users ensure the skill is registered.
