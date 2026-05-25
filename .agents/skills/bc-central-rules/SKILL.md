---
name: bc-central-rules
description: Use when working on BC Ai Account rules, skills, development workflow, local runtime, DEV deployment, database/storage configuration, or cross-agent instructions.
---

# BC Central Rules

Central sources:
- Canonical rules: `D:\bccode\AGENTS.md` and `D:\bccode\.agents\rules\bc-account-core-rules.md`
- Token-saving routing: `D:\bccode\AI_INDEX.md` and `D:\bccode\.agents\wiki\llm-index.md`
- Applies to Claude Code, Codex/GPT-5.5, and Google Antigravity/Gemini.

Before changing any rule, skill, wiki/LLM page, prompt, handoff, checklist, workflow, runtime config, deploy instruction, database/storage instruction, or agent entrypoint:
1. Read `D:\bccode\AGENTS.md`.
2. Read `D:\bccode\.agents\rules\bc-account-core-rules.md`.
3. Read `D:\bccode\.agents\wiki\llm-index.md` when the change affects durable knowledge, routing, source maps, or reusable LLM context.
4. Keep the change portable across all three agent systems.
5. Do not duplicate canonical runtime/deploy/storage/version details in skills or wiki pages. Link back to the central rule file.
6. Do not write secrets into durable files. Use env/secret variable names instead.
7. For private image/storage rules, keep the durable rule in `bc-account-core-rules.md` and add only short routing pointers in `AI_INDEX.md`, wiki, or task-specific skills.

If a local skill or wiki page conflicts with the central rule file, update that downstream file to point back to the central rule instead of creating a new rule variant.
