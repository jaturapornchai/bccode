---
name: bc-central-rules
description: Use when working on rules, skills, workflow, deployment, database, or storage.
---

## Rules Portability & Priorities
- **Canonical Files**: [AGENTS.md](file:///D:/bccode/AGENTS.md) and [bc-account-core-rules.md](file:///D:/bccode/.agents/rules/bc-account-core-rules.md).
- **Rule Changes**: Read canonical files first. Keep every change consistent across all entrypoints so any CLI reads the same rules. **Opus 4.8 is the orchestrator/lead** over the advisor/subagent pool (GLM 5.2, Kimi K3 Max, DeepSeek, ChatGPT 5.6, Claude Fable, Claude Sonnet) — see core-rules "Multi-Model Orchestration".
- **Wrong Screen / Width Fixes**: If Jead says `ผิดจอ`, re-confirm `D:\bccode` route/source evidence before edits. Full-width ERP screens should use the central rule contract `w-full max-w-none min-w-0` and avoid hiding work-surface overflow.
- **Tokens Budget**: Link to central rules. Do not copy long technical details to downstream skills.
- **No Secrets**: Never write credentials or tokens into Git, wiki, or prompt assets.
- **Internet Knowledge Search**: When local source/docs/rules do not cover a question, follow `AGENTS.md` "Internet Knowledge Search": start external discovery at Perplexity, then verify against a cited primary official/vendor/GitHub source; use WebSearch/webReader only as fallback when Perplexity is unavailable or insufficient.
