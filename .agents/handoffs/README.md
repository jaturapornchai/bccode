# Agent Handoffs

Cross-agent handoff contracts for the Gemini → Codex → Claude workflow.
See role rules in `D:\bccode\.agents\rules\bc-account-core-rules.md` ("AI Models Collaboration Rules").

## Why
Gemini edits frontend only. When a frontend change needs backend/model support, Gemini drops a handoff here so **Codex** can adapt the Go model/backend and **Claude** can review/plan — without re-deriving the frontend intent. The handoff is the contract; it keeps the three agents parallel and fast.

## Flow
Default direction is **Gemini → Codex → Claude**, but any agent may write a handoff to request work in another agent's lane (e.g. Codex → Gemini for a UI tweak, or anyone → Claude for review/plan). Same roles apply to all three — see core-rules "AI Models Collaboration Rules".

1. **Gemini** finishes frontend work → creates `{YYYY-MM-DD}-{kebab-feature}.md` from the template below. Fills everything except the Codex/Claude status.
2. **Codex** reads it + the real frontend code → implements/adjusts Go models, services, handlers, MCP tools, migrations → ticks the backend checklist and updates `Status: Codex`.
3. **Claude** reviews the diff against this file, plans next steps, updates `Status: Claude`. Claude may edit any layer when end-to-end work is needed.

## Rules
- One file per feature. English. Plain Markdown. Keep it short.
- Every backend ask must be source-linked (`path:line`) — no vague "fix the API".
- Do not store secrets, tokens, or full connection strings.
- When done end-to-end, set all three `Status` lines to `DONE` (or `N/A`). Stale handoffs can be deleted once merged.

## Template
Copy the block below into a new `{YYYY-MM-DD}-{kebab-feature}.md`.

```markdown
# Handoff: <feature name>

- Feature: <short name>
- Author: Gemini
- Date: <YYYY-MM-DD>
- Scope: frontend-done / backend-todo
- Status: Gemini=DONE | Codex=TODO | Claude=TODO

## 1. Frontend change (done by Gemini)
What changed in the UI and why. List touched files (`frontend/...`).

## 2. Backend / model changes required (for Codex)
Exactly what the Go model / backend must provide. Source-link existing code (`backend/...:line`).
- [ ] Model field(s): ...
- [ ] Endpoint(s): method, path, request, response
- [ ] Validation / business rule: ...
- [ ] MCP tool / migration (if any): ...

## 3. API contract expected by the frontend
Request and response shapes the frontend already codes against (so backend matches exactly).

## 4. Affected files
- Frontend (done): `frontend/...`
- Backend (todo): `backend/...`

## 5. Verification checklist
- [ ] `cd backend; go build ./...` (or Docker build for kafka/CGO paths)
- [ ] `cd frontend; npm run typecheck`
- [ ] Real DEV data / API check (no mock for business behavior)
- [ ] Light/dark + mobile check for any UI

## 6. Notes for Claude (review + plan)
Risks, open questions, follow-up phases.
```
