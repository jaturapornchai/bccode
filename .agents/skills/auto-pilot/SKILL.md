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
- **Verify**: Frontend dev = rely on `next dev` HMR; run `npm run typecheck` only before commit/summary. Backend = rebuild Docker ONLY when Jead says `rebuild`/`deploy` (batch mode). See core-rules "Dev Workflow Mode".
