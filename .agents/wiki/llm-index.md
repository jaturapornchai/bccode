# BC Ai Account LLM Wiki Index

This is the shared LLM-readable wiki entrypoint, identical for every agent that opens this repo (Claude Code / ZCode / Codex / any). **Opus 4.8 leads** the advisor/subagent pool (GLM 5.2, Kimi K3 Max, DeepSeek, ChatGPT 5.6, Claude Fable, Claude Sonnet) — see core-rules "Multi-Model Orchestration". Whichever CLI is running uses these same routes.

## Read Order
1. `D:\bccode\AI_INDEX.md` for task routing.
2. `D:\bccode\AGENTS.md` and `D:\bccode\.agents\rules\bc-account-core-rules.md` only for rule-sensitive work.
3. The task-specific `SKILL.md` only when its trigger matches.
4. Only the exact source files/tests/runtime evidence routed by the task.

## Purpose
- Route agents to the right source files without loading the whole repo.
- Keep durable LLM context short, source-first, and reusable across agent systems.
- Prevent per-agent rule variants from drifting — every agent reads one identical rule set.

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
- DEV sample/seed data routing: `D:\bccode\.agents\skills\dev-data-seeder\SKILL.md`; resolve the selected `holdingcode` first and seed through real DEV APIs only.
- Private image upload/preview routing: `D:\bccode\.agents\rules\bc-account-core-rules.md` Image Display Enforcement, then `D:\bccode\AI_INDEX.md` private image upload route for exact frontend/backend files.
- Image upload FORMAT routing (PNG/JPG only, no WebP): `D:\bccode\.agents\rules\bc-account-core-rules.md` "Uploaded Image Format Iron Rule". Authoritative backend guard is `backend/internal/goapi/handlers/image_r2.go ImageUploadHandler` (ext `jpg/jpeg/png` + `http.DetectContentType` image/jpeg|png). Frontend accept/validation/encoders in `frontend/src/app/system-settings/system-settings-screen.tsx`, `menu/product-screen.tsx`, `components/product-barcode/barcode-form.tsx`. Thumbnails must stay (small JPG/PNG). Static `.webp` page-background assets are exempt.
- Notification/toast routing: all screens use the global toast `frontend/src/lib/toast.ts` (`toast.*` / `pushNotice`) rendered by `frontend/src/components/toast-viewport.tsx` mounted in `app/layout.tsx`; never per-screen inline notice banners. See `D:\bccode\.agents\rules\bc-account-core-rules.md` "Global Toast Notification Rule".
- Mongo not-found / `FindOne`-swallow routing: `PersisterMongo.FindOne` returns zero struct + nil error on no-match (unlike `FindByID`). Existence/fallback/duplicate decisions must test the decoded result for emptiness, not `err`. See go-expert / go-api-handler skills (canonical fix: select-holding `holdingcode invalid`, `internal/authentication/services/authentication_service.go`).
- Frontend density and above-the-fold chrome rules: `D:\bccode\.agents\rules\bc-account-core-rules.md`, then `D:\bccode\.agents\skills\nextjs-frontend\SKILL.md` and `D:\bccode\.agents\skills\formdesign\SKILL.md`.
- Responsive priority (Tablet/Desktop-first, mobile last): `D:\bccode\.agents\rules\bc-account-core-rules.md` "Responsive Priority Iron Rule" and `D:\bccode\AGENTS.md` "Premium UX/UI Standard" — design and verify iPad (`md:768px`) → notebook (`lg/xl`) → desktop (`2xl`) → mobile (`<768px`) last. Do NOT follow any older "mobile-first / 375px first" instruction.
- Frontend raw JSON UX fixes: route to `D:\bccode\.agents\rules\bc-account-core-rules.md` "No raw JSON user input rule" and `D:\bccode\.agents\skills\nextjs-frontend\SKILL.md` "No Raw JSON User Input"; inspect the active renderer/config first and replace normal business textareas with structured editors.
- Frontend leading-icon input/search overlap fixes: route to `D:\bccode\.agents\rules\bc-account-core-rules.md` "Frontend Density Contract" and `D:\bccode\.agents\skills\nextjs-frontend\SKILL.md` "Leading Icon Inputs"; inspect active `frontend/src` source and verify no `Search` icon input relies on plain `pl-8`/`pl-9`.
- Marketplace SKU dimension availability routing: inspect `D:\bccode\frontend\src\lib\product-barcode\types.ts`, `D:\bccode\frontend\src\app\menu\tab-product-marketplace.tsx`, and matching backend product/productbarcode models. The new durable contract is `marketplaceskumappings[].marketplacedimensionstocks[]` for per-channel dimension availability projection, not accounting stock; legacy underscore keys are compatibility only.
- Marketplace scope + field mapping (Shopee/Lazada/TikTok/AliExpress): system scope rule in `D:\bccode\.agents\rules\bc-account-core-rules.md` "System Scope & Marketplace Support"; field-by-field mapping of the canonical `MarketplaceProductMap`/`MarketplaceSKUMap` contract to each marketplace's official API fields (+ Shopee example, official doc URLs) in `D:\bccode\.agents\wiki\marketplace-field-mapping.md`. Marketplace mapping is product/SKU-level only — do NOT add parallel marketplace structures on templates/option-sets. Official docs mandatory, no fabricated field names.
- Add-copy workflow routing: inspect `D:\bccode\frontend\src\app\system-settings\system-settings-screen.tsx` (`openCreateCopy` and toolbar buttons). Editable CRUD/list screens should show `คัดลอก` beside `เพิ่ม` when a selected row can be duplicated; it opens create mode and pre-fills the form from the selected row.
- Master-data list toolbar routing: list screens should include search, `ตัวกรอง`, `รูป` when image data exists, `เลือกเพื่อลบ`, selected-row multi-delete, and total count. If a right-side detail panel already has edit/delete actions, remove per-row edit/delete buttons from the list.
- Screen/menu label routing: verify page titles, detail headers, aria labels, and toolbar context against the active route/menu before finishing UI work. Shared product/barcode dictionaries must not leak screen-level labels across `/product`, `/productbarcode`, and `/productset`.
- Screen header workspace context routing: individual screen headers must not repeat selected company/branch badges because the global workspace bar already shows them. Keep company/branch UI only when it is actual workflow data, a filter, an access scope selector, or legal/branch setup.
- Product list datalist routing: `/product` datalist rows must come from the PostgreSQL `product` read model/projection first, then enrich MongoDB detail only for full edit/detail fields. Do not use `productbarcode` as the product list identity/source; it is only the barcode/SKU/unit/price child relation. Product list columns should include `ประเภทหน่วยนับ` and product-level `ยอดคงเหลือ` formatted by auto-packing (`ลัง x โหล x ชิ้น`) from product/accounting stock relation. Projection queries must use source-verified columns/tables or defensive defaults for optional display fields.
- Mobile/computer product routing: use `/productserialregistry` for `ทะเบียนเลขเครื่อง` and `/channelprice` for `ราคาตามช่องทางขาย`. These are structured MongoDB-backed CRUD screens (local DEV containers per Environment Topology; Atlas retired) for mobile phone, SIM, computer, and serialized electronics workflows; do not replace them with raw JSON fields.
- Database naming reset routing: new or changed MongoDB/PostgreSQL/ClickHouse contracts, database-facing function names, API route/query/body keys tied to database fields, table/collection/index names, fields, query keys, API payload keys tied to database fields, and variables/constants representing database identifiers must be lowercase (underscore allowed; snake_case OK — only uppercase/camelCase is a violation). Use `D:\bccode\.agents\rules\bc-account-core-rules.md` "Lowercase Database Naming Iron Rule (underscore allowed)" before schema/API work.
- Accounting number routing: for ERP Core, Accounting, Finance, Stock, Sales, Purchase, AR/AP, GL, VAT/Tax, Payment, Cost, Report, and Analytics, read `D:\bccode\.agents\rules\bc-account-core-rules.md` "ERP Accounting Decimal Iron Rule" and `D:\bccode\.agents\wiki\accounting-number-rules.md` before changing money, price, cost, decimal quantity, exchange rate, stock value, debit/credit, report total, or projection schemas.
- Thai company/branch tax structure and head-office branch numbering: `D:\bccode\.agents\rules\bc-account-core-rules.md`, then `D:\bccode\AI_INDEX.md` for exact backend/frontend source paths.
- Timezone routing (set 2026-06-22): databases (MongoDB/PostgreSQL/ClickHouse) store timestamps in UTC+0; backend writes/reads UTC; frontend converts UTC → active branch `timezone` for display via `frontend/src/lib/date-time.ts`. Read `D:\bccode\.agents\rules\bc-account-core-rules.md` "Timezone Iron Rule" (+ "Project Date Display Contract"); skills go-expert/db-expert (UTC storage) + nextjs-frontend (branch-tz display).
- Plan-first routing (set 2026-06-22): after any task/command, state a brief plan (goal/steps/blast-radius/smallest-safe approach) BEFORE editing; confirm first for ambiguous scope or R0/R1. See `D:\bccode\.agents\rules\bc-account-core-rules.md` "Plan-First Rule" + `D:\bccode\AGENTS.md` Quality.
- Thai tax / VAT / WHT / e-Tax / GL correctness (research-first + cache): `D:\bccode\.agents\rules\bc-account-core-rules.md` "Tax & Accounting Correctness", then read `D:\bccode\.agents\skills\bc-account-expert\tax-legal-cache.md` (verified findings) BEFORE re-searching; for anything uncached, research `rd.go.th` + Thai tax law + competitor ERP and append a dated entry to the cache.
- Cross-agent rule editing: `D:\bccode\.agents\skills\bc-central-rules\SKILL.md`.
- Opus 4.8 orchestrates and owns the whole task, dispatching subtasks to the advisor/subagent pool and verifying every result itself: `D:\bccode\.agents\rules\bc-account-core-rules.md` "Multi-Model Orchestration" + "AI Capability & Instant Upgrades". Handoff files (`D:\bccode\.agents\handoffs\README.md`) are optional cross-session coordination notes, not role locks.
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
2. Keep it plain Markdown and model-agnostic so every agent (Claude Code / ZCode / Codex / any) reads it identically.
3. Link back to this index or the central rule file.
4. Verify with a targeted `rg` sweep before completion.
