# BC Ai Account AI Routing Index

Purpose: keep Codex, Claude Code, and other agents fast. Read this file first, then open only the relevant source files.

## Default Workflow
- For project-wide rules, local runtime, storage, secrets, DEV deployment, wiki/LLM knowledge, or reusable agent context, read `.agents/rules/bc-account-core-rules.md` and `.agents/wiki/llm-index.md` first; keep agent assets portable across Claude Code, Codex/GPT-5.5, and Google Antigravity/Gemini.
- Identify the task area below.
- Use `rg -n "symbol|label|route"` before opening large files.
- For files over 50 KB, read line ranges or exact functions only.
- Patch the smallest safe scope.
- Verify with focused commands, not whole-repo checks.

## Fast Commands
- Frontend typecheck: `cd frontend; npm run typecheck`
- Focused frontend lint: `cd frontend; npm run lint -- <file>`
- Backend local runtime: use Docker Desktop; MainAPI should answer at `http://localhost:8888`; Kafka and Redis must run with MainAPI.
- Backend local deploy after backend changes: `cd backend; docker-compose up -d --no-deps --build mainapi`
- Backend health: `curl.exe --max-time 10 -s -i http://localhost:8888/healthz`
- Real data check: use the selected DEV database/API path; do not rely on mock business data for completion claims.
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
- Login/auth UI: `frontend/src/app/login-screen.tsx`, `frontend/src/app/api/auth/**`, backend auth files only when API behavior is involved.
- Workspace/company selection: `frontend/src/app/workspace/workspace-screen.tsx`, `frontend/src/app/api/workspace/[...workspacePath]/route.ts`, `frontend/src/lib/workspace-models.ts`.
- Main menu/dashboard/sidebar: `frontend/src/app/menu/main-menu-screen.tsx`, `frontend/src/app/menu/menu-dashboard-data.ts`, `frontend/src/lib/menu-data.ts`, `frontend/src/lib/menu-icons.ts`.
- System settings screens: `frontend/src/app/system-settings/system-settings-screen.tsx`, `frontend/src/lib/system-setting-screens.ts`, `frontend/src/app/api/system-settings/[[...settingPath]]/route.ts`.
- Thai company/branch tax structure: `.agents/rules/bc-account-core-rules.md`, `backend/internal/organization/branch/models/branch_code.go`, `backend/internal/organization/branch/services/branch_http_service.go`, `backend/cmd/branch_code_audit/main.go`, `frontend/src/lib/thai-branch-code.ts`, and the system-settings files above.
- Thai address dataset/cascading lookup: `backend/assets/address/thailand-addresses.json`, `backend/internal/goapi/handlers/address_handler.go`, `frontend/src/app/api/address/thailand/route.ts`, `frontend/src/lib/thailand-addresses.ts`, and the system-settings branch address fields.
- Private image upload and preview: `.agents/rules/bc-account-core-rules.md` Image Display Enforcement, `frontend/src/lib/image-upload-proxy.ts`, `frontend/src/app/api/upload/image/route.ts`, `frontend/src/app/system-settings/system-settings-screen.tsx`, `backend/internal/goapi/handlers/image_r2.go`, `backend/internal/goapi/handlers/s3_proxy.go`, and `backend/internal/goapi/handlers/storage_private_test.go`.
- Frontend density/top chrome: `frontend/src/app/menu/main-menu-screen.tsx`, `frontend/src/app/system-settings/system-settings-screen.tsx`, `frontend/src/app/globals.css`, plus `.agents/skills/nextjs-frontend/SKILL.md` and `.agents/skills/formdesign/SKILL.md`.
- User management: same as System settings plus `backend/internal/authentication/**` and `backend/internal/shop/**` only when server behavior is involved.
- Permissions/menu access: `frontend/src/lib/menu-data.ts`, `frontend/src/lib/menu-permissions.ts`, `frontend/src/app/system-settings/system-settings-screen.tsx`, related backend permission services when needed.
- Language/i18n: frontend language caller first, then exact key lookup in `backend/assets/language/languages.tsv`; record provisional keys under `backend/prompts/language_requests/`.
- LINE OA/linking: `frontend/src/app/line-oa/**`, auth LINE API routes, backend LINE OA handlers only when API behavior is involved.
- Backend auth/password/shop access: `backend/internal/authentication/authentication_http.go`, `backend/internal/authentication/services/authentication_service.go`, `backend/internal/authentication/models/user.go`, `backend/internal/shop/**`.
- API version compatibility: inspect the exact frontend API caller/proxy route and backend handler first. Current backend contract baseline is `v1`; future versions such as `v2` must run side-by-side with `v1`, and web/iOS/Android clients must declare required backend version.

## Manual Rule
Do not create or update manuals automatically. Manual generation is only when Jead explicitly requests a specific manual or manual batch.

## Escalation
Use broad repo review only when the user asks for whole-system review, security review, production readiness, or architecture changes.
