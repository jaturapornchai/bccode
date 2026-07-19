---
name: auto-pilot
description: Primary skill for BC Ai Account. Use for every task; routes via AI_INDEX.md first.
---

## 1. Context & Routing Order
- **Route**: Always read [AI_INDEX.md](file:///D:/bccode/AI_INDEX.md) and task files first.
- **Reference**: Read [AGENTS.md](file:///D:/bccode/AGENTS.md) and [bc-account-core-rules.md](file:///D:/bccode/.agents/rules/bc-account-core-rules.md) only if needed.
- **Durable Facts**: Keep knowledge in central rules, not copied to local skills.

## 2. Coding & Quality
- **Patches**: Small, focused edits. No unrelated refactoring.
- **No Mocks**: Use DEV database/API; no guessed configurations/schemas.
- **Languages**: Query `languages.tsv` by exact key regex only (e.g. `rg -n "^key\t"`).
- **Backend/Frontend**: Target `frontend/` (Next.js) or `backend/` (Go) scope.
- **Multi-model orchestration (Opus 4.8 leads)**: Opus splits the task, dispatches subtasks to the advisor/subagent pool (GLM 5.2, Kimi K3 Max, DeepSeek, ChatGPT 5.6, Claude Fable, Claude Sonnet), evaluates the returned work, and re-assigns — grounding every decision in real source/runtime evidence, applying, verifying, and owning the outcome. See core-rules "Multi-Model Orchestration".
- **External Workspaces**: Folders outside `D:\bccode` are outside this GitHub project. Do not stage/commit/push external workspaces unless Jead explicitly gives a separate remote for that folder.
- **Verify**: Frontend dev = rely on `next dev` HMR; run `npm run typecheck` only before commit/summary. Backend = after Go-code edits are complete, auto deploy local Docker Desktop with `cd D:\bccode\backend; .\scripts\deploy-mainapi-fast.ps1`, then verify `/healthz`; use full `docker-compose up -d --no-deps --build mainapi` for image/runtime changes. DEV server deploy still requires explicit `deploy dev`. See core-rules "Dev Workflow Mode".
