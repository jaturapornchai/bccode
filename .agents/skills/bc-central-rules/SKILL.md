---
name: bc-central-rules
description: Use when working on rules, skills, workflow, deployment, database, or storage.
---

## Rules Portability & Priorities
- **Canonical Files**: [AGENTS.md](file:///D:/bccode/AGENTS.md) and [bc-account-core-rules.md](file:///D:/bccode/.agents/rules/bc-account-core-rules.md).
- **Rule Changes**: Read canonical files first. Keep every change model-agnostic so any agent (Claude Code / ZCode / Codex / any) reads it identically — no model names in a role sense, no advisor/lead/helper language.
- **Wrong Screen / Width Fixes**: If Jead says `ผิดจอ`, re-confirm `D:\bccode` route/source evidence before edits. Full-width ERP screens should use the central rule contract `w-full max-w-none min-w-0` and avoid hiding work-surface overflow.
- **Tokens Budget**: Link to central rules. Do not copy long technical details to downstream skills.
- **No Secrets**: Never write credentials or tokens into Git, wiki, or prompt assets.
- **Internet Knowledge Search**: When local source/docs/rules do not cover a question (version compat, install steps, deprecated API, unfamiliar error, external-service behavior), search the web (WebSearch) or read the official doc (webReader) before guessing. Cite source URL, prefer official/vendor/GitHub-issue pages, and verify findings against the project source/runtime before applying. See `AGENTS.md` "Execution guidance" → Internet Knowledge Search.
