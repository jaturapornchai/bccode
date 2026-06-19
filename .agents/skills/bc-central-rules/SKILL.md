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
