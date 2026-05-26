# BC Ai Account LLM Wiki Index

This is the shared LLM-readable wiki entrypoint for Claude Code, Codex/GPT-5.5, and Google Antigravity/Gemini.

## Read Order
1. `D:\bccode\AI_INDEX.md` for task routing.
2. `D:\bccode\AGENTS.md` and `D:\bccode\.agents\rules\bc-account-core-rules.md` only for rule-sensitive work.
3. The task-specific `SKILL.md` only when its trigger matches.
4. Only the exact source files/tests/runtime evidence routed by the task.

## Purpose
- Route agents to the right source files without loading the whole repo.
- Keep durable LLM context short, source-first, and reusable across agent systems.
- Prevent separate Claude/Codex/Gemini rule variants from drifting.

## Knowledge Rules
- Do not guess project facts. Verify with active repo files, local docs, tests, command output, logs, browser checks, or runtime evidence.
- Do not store secrets, credentials, tokens, private keys, customer data, or full connection strings in wiki pages.
- Do not duplicate large source explanations. Link to source paths and summarize only the decision-critical facts.
- If a page grows too broad, split it into narrow pages and keep this index as the router.

## Current Core Routing
- Frontend task routing: `D:\bccode\AI_INDEX.md`, then `D:\bccode\frontend\...`.
- Backend task routing: `D:\bccode\AI_INDEX.md`, then `D:\bccode\backend\CLAUDE.md`, then exact Go handlers/services/tests.
- Thai SME accounting/business-domain work: `D:\bccode\.agents\rules\bc-account-core-rules.md` Thai SME Business Domain, then `D:\bccode\.agents\skills\bc-account-expert\SKILL.md`, then active source/legacy reference routed by `D:\bccode\AI_INDEX.md`.
- Runtime, DEV deployment, storage, secrets, real-data verification, and API version rules: `D:\bccode\.agents\rules\bc-account-core-rules.md`.
- Private image upload/preview routing: `D:\bccode\.agents\rules\bc-account-core-rules.md` Image Display Enforcement, then `D:\bccode\AI_INDEX.md` private image upload route for exact frontend/backend files.
- Frontend density and above-the-fold chrome rules: `D:\bccode\.agents\rules\bc-account-core-rules.md`, then `D:\bccode\.agents\skills\nextjs-frontend\SKILL.md` and `D:\bccode\.agents\skills\formdesign\SKILL.md`.
- Thai company/branch tax structure and head-office branch numbering: `D:\bccode\.agents\rules\bc-account-core-rules.md`, then `D:\bccode\AI_INDEX.md` for exact backend/frontend source paths.
- Cross-agent rule editing: `D:\bccode\.agents\skills\bc-central-rules\SKILL.md`.
- Legacy Flutter reference only when needed for migrated screens: `D:\bcdev\frontend\bcaiaccount`.

## Runtime Map
- Do not duplicate runtime/deploy/storage facts in wiki pages. Read `D:\bccode\.agents\rules\bc-account-core-rules.md` for the current values.
- K3s/Kubernetes is not active by default; only route there when Jead explicitly asks for production scaling or Kubernetes work.

## Verification And Version Routing
- Business behavior must be verified against real selected DEV data or real APIs. Mocked/unit checks alone cannot prove business data behavior.
- Backend API contracts start at `v1`. Future breaking changes must add `v2`, `v3`, and so on while keeping old clients compatible.
- Supported backend API versions must run side-by-side; adding `v2` must not disable `v1`.
- Web, iOS, and Android clients must declare the required backend version, preferably through `X-BC-Required-Backend-Version`, plus client platform/version metadata.
- Version-sensitive work should route through `D:\bccode\AI_INDEX.md`, then inspect the exact frontend API caller and backend route/handler before changing contracts.

## Update Rule
When creating or updating any rule, skill, wiki page, LLM prompt, handoff, checklist, workflow document, or reusable agent instruction:
1. Keep it in English unless Jead explicitly asks for Thai end-user output.
2. Keep it plain Markdown and portable across Claude Code, Codex/GPT-5.5, and Google Antigravity/Gemini.
3. Link back to this index or the central rule file.
4. Verify with a targeted `rg` sweep before completion.
