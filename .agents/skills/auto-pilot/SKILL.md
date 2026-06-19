---
name: auto-pilot
description: Primary skill for BC Account. Use for every task; routes via AI_INDEX.md first.
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
- **GLM 5.2 Helper (primary advisor)**: For every BC Account task involving planning, implementation, code review, debugging, UX/UI review, recommendations, Thai business/accounting/POS workflow decisions, or weakness checks, **Claude (Claude Code) is the lead/worker** and should invoke `.agents/skills/glm52-planner/SKILL.md` plus `tools/ai/glm52-think-planner.ps1` as the **primary advisor** after reading local evidence — and `codex-advisor` (Codex) as the secondary advisor for engineering bug/security/edge-case angles. Prefer GLM `-Mode all` unless a narrower mode is clearly sufficient. GLM 5.2 is especially useful for Asian/Thai context: Thai business, Thai accounting software, Thai ERP, tax/document workflows, retail/restaurant operations, and Thai POS. It requires `ZAI_API_KEY`; if missing or unavailable, report the fallback and continue safely with Claude-only work. Never persist the key, let an advisor edit files directly, expose raw reasoning, or treat advisor output as source of truth.
- **External Workspaces**: Folders outside `D:\bccode` are outside this GitHub project. Do not stage/commit/push external workspaces unless Jead explicitly gives a separate remote for that folder.
- **Verify**: Frontend dev = rely on `next dev` HMR; run `npm run typecheck` only before commit/summary. Backend = after Go-code edits are complete, auto deploy local Docker Desktop with `cd D:\bccode\backend; .\scripts\deploy-mainapi-fast.ps1`, then verify `/healthz`; use full `docker-compose up -d --no-deps --build mainapi` for image/runtime changes. DEV server deploy still requires explicit `deploy dev`. See core-rules "Dev Workflow Mode".
