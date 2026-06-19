# Agent Handoffs

Optional coordination notes for parallel / cross-session work. Claude (Claude Code) is the lead/full-stack worker and may pick up any layer; GLM 5.2 (primary advisor) and Codex (secondary advisor + image gen) contribute advice — see core-rules "AI Capability & Instant Upgrades". Handoffs are NOT role locks; they only carry context so the next agent doesn't re-derive intent.

## Why
When work spans sessions/agents, or one chunk is done and another remains (e.g. frontend done, backend pending), drop a handoff so any agent can continue without re-reading the whole history. The handoff is the contract; it keeps work parallel and fast.

## Flow
Any agent may write a handoff to request follow-up work in any layer. There is no fixed model→model direction.

1. An agent finishes a chunk → creates `{YYYY-MM-DD}-{kebab-feature}.md` from the template below, filling everything except the pending phases.
2. The next agent reads it + the real source code → implements the remaining layer(s) → ticks the checklist and updates the phase status.
3. A reviewer (any agent) checks the diff against this file, plans next steps, marks done. Any agent may edit any layer when end-to-end work is needed.

## Rules
- One file per feature. English. Plain Markdown. Keep it short.
- Every ask must be source-linked (`path:line`) — no vague "fix the API".
- Do not store secrets, tokens, or full connection strings.
- When done end-to-end, set all phases to `DONE` (or `N/A`). Stale handoffs can be deleted once merged.

## Template
Copy the block below into a new `{YYYY-MM-DD}-{kebab-feature}.md`.

```markdown
# Handoff: <feature name>

- Feature: <short name>
- Author: <agent / session>
- Date: <YYYY-MM-DD>
- Scope: <what's done / what's left>
- Status: frontend=<DONE|TODO|N/A> | backend=<DONE|TODO|N/A> | review=<DONE|TODO|N/A>

## 1. What changed (done)
What changed and why. List touched files (`frontend/...` or `backend/...`).

## 2. Remaining work required
Exactly what the next layer must provide. Source-link existing code (`path:line`).
- [ ] Model field(s): ...
- [ ] Endpoint(s): method, path, request, response
- [ ] Validation / business rule: ...
- [ ] MCP tool / migration (if any): ...

## 3. API contract
Request and response shapes the frontend codes against (so backend matches exactly).

## 4. Affected files
- Done: `...`
- TODO: `...`

## 5. Verification checklist
- [ ] `cd frontend; npm run typecheck` (before commit/summary — dev relies on `next dev` HMR)
- [ ] `cd backend; go build ./...` for touched packages (full Docker rebuild ONLY when Jead says `rebuild`/`deploy`)
- [ ] Real DEV data / API check (no mock for business behavior)
- [ ] Light/dark + mobile check for any UI

## 6. Notes for reviewer
Risks, open questions, follow-up phases.
```
