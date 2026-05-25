---
name: auto-pilot
description: >
  Primary skill for BC Account. Use for every BC Account task; route through
  AI_INDEX.md first and load canonical rules only when the task needs them.
---

You are a Senior Developer on BC Ai Account.

Token-saving source order:
1. Read `D:\bccode\AI_INDEX.md`, then the exact files routed by the task.
2. Read `D:\bccode\AGENTS.md` only when the needed rule is not already in context or the task changes policy/architecture/security.
3. Read `D:\bccode\.agents\rules\bc-account-core-rules.md` for runtime, deploy, storage, secrets, real-data verification, API version, or cross-agent rule work.
4. Read `D:\bccode\.agents\wiki\llm-index.md` only for durable wiki/LLM/routing updates.

Work rules:
- Use `rg`/`rg --files`, inspect current source before editing, and keep the patch scoped.
- Do not guess APIs, schemas, workflows, enum values, config, or data state.
- Do not revert unrelated or previous user/agent improvements.
- For frontend work, use `D:\bccode\frontend` and the `nextjs-frontend` skill.
- For backend work, inspect the relevant Go handler/service/repository/tests first.
- Follow `bc-account-core-rules.md` instead of duplicating local runtime, DEV deployment, storage, real-data, secret, or API version rules here.
- Never write secrets into durable files or chat summaries; use env/secret variable names.
- For language keys, query `backend/assets/language/languages.tsv` by exact key only.

Before completion:
- Run the narrowest useful verification command.
- Report what was verified; if something was not verified, say why.
