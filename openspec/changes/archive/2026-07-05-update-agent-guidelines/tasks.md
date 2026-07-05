## 1. Add CLI `--` separator rule to skill/SKILL.md

- [x] 1.1 Add a rule to the Rules section: "**Flags before `--`, positional args after.** When passing negative amounts or other values that look like flags, use `--` to separate flags from positional args."
- [x] 1.2 Include the visual diagram example showing `wallet adjust "Bunga Bank" --json -- -3612 "Initial balance"` with annotation

## 2. Add database access boundary rule to skill/SKILL.md

- [x] 2.1 Add a rule to the Rules section: "**Never touch the database directly.** Do not open the SQLite file, write raw SQL, or create scripts that manipulate database data. Always use the `wallet` CLI for all data operations (inserts, queries, updates, deletes)."

## 3. Add skill installation instructions to README.md

- [x] 3.1 Add a sub-section under "Installation" titled "Agent Skill (AI Tools)" explaining how to register `skill/SKILL.md`
- [x] 3.2 Include instructions for Hermes Agent and OpenClaw
- [x] 3.3 Note that registering the skill enables AI tools to auto-detect wallet-related queries and use correct CLI commands

## 4. Verification

- [x] 4.1 Review `skill/SKILL.md` to confirm both rules are present in the Rules section
- [x] 4.2 Review `README.md` to confirm skill installation instructions are present under Installation
- [x] 4.3 Verify existing content in both files is otherwise unchanged
