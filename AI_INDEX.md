# BC Ai Account AI Routing Index

Purpose: keep Codex, Claude Code, and other agents fast. Read this file first, then open only the relevant source files.

## Default Workflow
- Start every task with a concise plan before running command sequences, debugging, editing, deploying, committing, or pushing. Scale the plan to the task; even simple fixes need a short plan.
- After Jead gives a command, keep working toward an end-to-end result in the same turn: implement, adjust related datamodel/UX/UI/rules/models when appropriate, verify with real evidence, and report blockers only when the task is R0, missing required information that cannot be discovered locally, or blocked by runtime/tool limits. Test like a non-technical user would use the screen and go as deep as the current scope, time, and tools allow. If the change is risky or rollback safety matters, auto-push to GitHub after targeted verification plus secret/diff checks, unless remote divergence, secret risk, or an R0 blocker is found.
- For project-wide rules, local runtime, storage, secrets, DEV deployment, wiki/LLM knowledge, or reusable agent context, read `.agents/rules/bc-account-core-rules.md` and `.agents/wiki/llm-index.md` first; keep agent assets portable across Claude Code, Codex/GPT-5.5, and Google Antigravity/Gemini.
- Identify the task area below.
- Use `rg -n "symbol|label|route"` before opening large files.
- For files over 50 KB, read line ranges or exact functions only.
- Automatically use suitable available plugins/tools/skills for the task, such as Browser/Playwright for UI checks and GitHub/Drive plugins for those domains. Do not install new plugins without Jead's approval, and keep source/runtime evidence as the source of truth.
- Patch the smallest safe scope.
- Verify with focused commands, not whole-repo checks.

## Agent Fast Execution Contract
- Applies to Codex, Claude Code, Gemini/Antigravity, and any agent working in this repo.
- Default to targeted checks. Do not run expensive whole-repo commands such as `go test ./...`, broad browser automation, or full-repo scans unless Jead explicitly asks, the touched scope genuinely requires it, or targeted checks cannot provide useful evidence.
- Docs/rules-only changes: use `git diff --check` plus staged secret scanning. Do not run frontend/backend build or test commands.
- Frontend code changes: fast-iteration dev mode — rely on `next dev` HMR; run `cd frontend; npm run typecheck` only before commit/summary, not every edit. Use browser verification only when UI behavior, layout, routing, or visual output changed.
- Backend code changes: write correct code; run touched-package tests only if needed. After backend Go-code edits are complete, auto deploy `mainapi` on local Docker Desktop with the fast local path (`cd backend; .\scripts\deploy-mainapi-fast.ps1`) and verify `/healthz`. Use full image rebuild (`docker-compose up -d --no-deps --build mainapi`) when Dockerfile, dependencies, runtime assets, compose, config, or image contents changed. DEV server deploy still needs `deploy dev`. Avoid repo-wide backend tests by default because this repo has known CGO/Kafka/env-sensitive noisy packages.
- Long commands must be visible: state what is running, update Jead about every 30 seconds, and if a command exceeds roughly 2 minutes, report whether to continue, narrow, or stop based on evidence.
- For meaningful changes, push the whole project after targeted verification and secret checks. Keep commits moving; do not wait on irrelevant broad checks.
- Folders outside `D:\bccode` are outside this GitHub project. Do not include external workspaces in `push to github` for this repo unless Jead explicitly provides a separate remote for that folder.

## Fast Commands
- Frontend typecheck: `cd frontend; npm run typecheck`
- Focused frontend lint: `cd frontend; npm run lint -- <file>`
- Backend local runtime: use Docker Desktop; MainAPI should answer at `http://localhost:8888`; Kafka and Redis must run with MainAPI.
- Operational CRUD data flow: write/read MongoDB first; propagate writes through `MongoDB -> Kafka -> PostgreSQL -> ClickHouse`. Use PostgreSQL/ClickHouse only from explicit rebuild/sync/projection/BI workers.
- Database naming: new or changed MongoDB/PostgreSQL/ClickHouse contracts, database-facing function names, MCP tool/action names, API route/query/body keys tied to database fields, table/collection/index names, columns/fields, and variables/constants that represent database identifiers must be lowercase with no underscore, such as `holdingcode`, `guidfixed`, and `businesscode`.
- Frontend CRUD mutations: create/edit/delete screens call MongoDB-backed operational APIs only; Kafka/projection fan-out is backend responsibility.
- Data-list row click: select and show read-only detail only. Edit mode requires the pencil/edit action; amber/orange row highlight is editing-only.
- Model-sensitive system settings/access work: read active `D:\bccode` source, `.agents/rules/bc-account-core-rules.md`, and the relevant runtime/API code before changing API fields, CRUD identity, tenant scope, or cross-record references. Do not depend on external model-document folders.
- Shared frontend widgets: when the same UX appears in multiple menus/screens, reuse or create a central component/field renderer first so labels, validation, search, save/load mapping, and empty states stay consistent.
- Holding access scopes: users, screen permissions, permission groups, user permission assignments, and approval rights are Holding-owned under `holdingcode`; configure applicability with `scopetype`, `businesscode`, `branchcode`, and `allbranches`.
- User access audit: route `/useraccessaudit` is a read-only report in the access setup flow; it summarizes user access, screen permissions, groups, approvals, and can export to PDF through print.
- Company access selectors: general master-data `businesscodes` is company-level only. Do not render or save branch selections in normal CRUD/master-data company access fields. Read legacy `companyguids` only as a compatibility alias.
- Backend local Docker Desktop deploy (auto after backend edits): for Go-code-only changes use `cd backend; .\scripts\deploy-mainapi-fast.ps1`; for image/runtime changes use `docker-compose up -d --no-deps --build mainapi`; always verify `/healthz`. DEV server deploy still needs `deploy dev`.
- Backend health: `curl.exe --max-time 10 -s -i http://localhost:8888/healthz`
- Real data check: use the selected DEV database/API path; do not rely on mock business data for completion claims.
- DEV seed data: use `.agents/skills/dev-data-seeder/SKILL.md`; resolve the active `holdingcode` first, then seed through real DEV APIs and verify via the same screen API path.
- Product category groups: `groupnumber` is a usage/device/channel group. For `/productcategorygroupselectscreen` and `/productcategorylist`, build/verify a complete tree with `parentguid` and `parentguidall`; attach `codelist` to sellable leaf categories, not one flat root per menu type.
- Narrow search: `rg -n "term" <path>`
- Focused git status: `git status --short -- <exact-path-or-module>`
- Changed-file count only: `git diff --name-only -- <exact-path-or-module> | Measure-Object -Line`

## Large Files
- `frontend/src/app/system-settings/system-settings-screen.tsx` - do not open full file; search exact route/field/function.
- `frontend/src/app/menu/main-menu-screen.tsx` - do not open full file; search exact component or label.
- `frontend/src/app/globals.css` - search class names only.
- `backend/assets/language/languages.tsv` - never open full file; query exact keys only.
- Generated Swagger/docs, manuals, Playwright snapshots/logs, duplicate skill packs, backend runtime logs, and lockfiles are not default context.

## Task Routes
- Thai SME accounting/business-domain workflows: `.agents/rules/bc-account-core-rules.md` Thai SME Business Domain, `.agents/skills/bc-account-expert/SKILL.md`, then exact frontend/backend/legacy source for the affected module.
- Login/auth UI: `frontend/src/app/login-screen.tsx`, `frontend/src/app/api/auth/**`, backend auth files only when API behavior is involved.
- Workspace/company selection: `frontend/src/app/workspace/workspace-screen.tsx`, `frontend/src/app/api/workspace/[...workspacePath]/route.ts`, `frontend/src/lib/workspace-models.ts`.
- Main menu/dashboard/sidebar: `frontend/src/app/menu/main-menu-screen.tsx`, `frontend/src/app/menu/menu-dashboard-data.ts`, `frontend/src/lib/menu-data.ts`, `frontend/src/lib/menu-icons.ts`.
- System settings screens: `frontend/src/app/system-settings/system-settings-screen.tsx`, `frontend/src/lib/system-setting-screens.ts`, `frontend/src/app/api/system-settings/[[...settingPath]]/route.ts`.
- Thai company/branch tax structure: `.agents/rules/bc-account-core-rules.md`, `backend/internal/organization/branch/models/branchcode.go`, `backend/internal/organization/branch/services/branch_http_service.go`, `backend/cmd/branchcode_audit/main.go`, `frontend/src/lib/thai-branch-code.ts`, and the system-settings files above.
- Thai address dataset/cascading lookup: `backend/assets/address/thailand-addresses.json`, `backend/internal/goapi/handlers/address_handler.go`, `frontend/src/app/api/address/thailand/route.ts`, `frontend/src/lib/thailand-addresses.ts`, and the system-settings branch address fields.
- Private image upload and preview: `.agents/rules/bc-account-core-rules.md` Image Display Enforcement, `frontend/src/lib/image-upload-proxy.ts`, `frontend/src/app/api/upload/image/route.ts`, `frontend/src/app/system-settings/system-settings-screen.tsx`, `backend/internal/goapi/handlers/image_r2.go`, `backend/internal/goapi/handlers/s3_proxy.go`, and `backend/internal/goapi/handlers/storage_private_test.go`.
- Product category screen: `frontend/src/app/system-settings/product-category-tree-view.tsx`, `frontend/src/app/system-settings/system-settings-screen.tsx`, `frontend/src/lib/system-setting-screens.ts`, active backend model `backend/internal/product/productcategory/models/productcategory.go`, and legacy Flutter reference `D:\bcdev\frontend\bcaiaccount\lib\screens\config\product_category_screen.dart`.
- Product/product set/barcode DEV sample data: `.agents/skills/dev-data-seeder/SKILL.md`, then verify `frontend/src/app/menu/product-screen.tsx`, `frontend/src/app/api/product/[[...productPath]]/route.ts`, and backend `/product` + `/product/barcode` routes.
- Frontend density/top chrome: `frontend/src/app/menu/main-menu-screen.tsx`, `frontend/src/app/system-settings/system-settings-screen.tsx`, `frontend/src/app/globals.css`, plus `.agents/skills/nextjs-frontend/SKILL.md` and `.agents/skills/formdesign/SKILL.md`.
- User management: same as System settings plus `backend/internal/authentication/**` and `backend/internal/shop/**` only when server behavior is involved.
- Permissions/menu access: `frontend/src/lib/menu-data.ts`, `frontend/src/lib/menu-permissions.ts`, `frontend/src/app/system-settings/system-settings-screen.tsx`, related backend permission services when needed.
- Language/i18n: frontend language caller first, then exact key lookup in `backend/assets/language/languages.tsv`; record provisional keys under `backend/prompts/language_requests/`.
- LINE OA/linking: `frontend/src/app/line-oa/**`, auth LINE API routes, backend LINE OA handlers only when API behavior is involved.
- Backend auth/password/shop access: `backend/internal/authentication/authentication_http.go`, `backend/internal/authentication/services/authentication_service.go`, `backend/internal/authentication/models/user.go`, `backend/internal/shop/**`.
- API version compatibility: inspect the exact frontend API caller/proxy route and backend handler first. Current backend contract baseline is `v1`; future versions such as `v2` must run side-by-side with `v1`, and web/iOS/Android clients must declare required backend version.

## Manual Rule
Do not create, update, regenerate, or restore manuals automatically. Manual generation is only when Jead explicitly requests a specific manual or manual batch.

## Escalation
Use broad repo review only when the user asks for whole-system review, security review, production readiness, or architecture changes.
