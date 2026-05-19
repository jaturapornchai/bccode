---
name: auto-pilot
description: >
  Primary skill for BC Account. Use for every BC Account task, but keep this
  skill lightweight and read AGENTS.md as the full source of project rules.
---

You are a Senior Developer on BC Ai Account.

Before coding:
- Read `D:\bccode\AGENTS.md` for the full rules.
- Read `D:\bccode\AI_INDEX.md` when present, then route to only the files listed for the task.
- Read only the source/docs/tests needed for the requested module.
- Use `rg`/`rg --files`; root `.ignore` must keep searches out of generated,
  dependency, binary, manual, and legacy folders.
- Do not guess APIs, schemas, workflows, enum values, or config.

While working:
- Make the smallest safe scoped change.
- For frontend work, use `D:\bccode\frontend` and the `nextjs-frontend` skill.
- For backend work, inspect the relevant Go handler/service/repository/tests first.
- For language keys, query `backend/assets/language/languages.tsv` by exact key only.
- During fast UI iteration, prefer stable keys with fallback text and record missing keys under `backend/prompts/language_requests/`; batch-edit `languages.tsv` later unless release readiness or the user requires translations now.

Before completion:
- Run the narrowest useful verification command.
- Report what was verified; if something was not verified, say why.
