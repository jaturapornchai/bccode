# BC Ai Account AI Routing Index

Purpose: keep Codex, Claude Code, and other agents fast. Read this file first, then open only the relevant source files.

## Default Workflow
- Identify the task area below.
- Use `rg -n "symbol|label|route"` before opening large files.
- For files over 50 KB, read line ranges or exact functions only.
- Patch the smallest safe scope.
- Verify with focused commands, not whole-repo checks.

## Fast Commands
- Frontend typecheck: `cd frontend; npm run typecheck`
- Focused frontend lint: `cd frontend; npm run lint -- <file>`
- Backend health: `curl.exe --max-time 10 -s -i http://localhost:8888/healthz`
- Narrow search: `rg -n "term" <path>`
- Safe git status: `git status --short -- . ':!clone-skills' ':!frontend/bcaiaccount' ':!frontend/bclms'`

## Large Files
- `frontend/src/app/system-settings/system-settings-screen.tsx` - do not open full file; search exact route/field/function.
- `frontend/src/app/menu/main-menu-screen.tsx` - do not open full file; search exact component or label.
- `frontend/src/app/globals.css` - search class names only.
- `backend/assets/language/languages.tsv` - never open full file; query exact keys only.
- Generated Swagger/docs and lockfiles are not default context.

## Task Routes
- Login/auth UI: `frontend/src/app/login-screen.tsx`, `frontend/src/app/api/auth/**`, backend auth files only when API behavior is involved.
- Workspace/company selection: `frontend/src/app/workspace/workspace-screen.tsx`, `frontend/src/app/api/workspace/[...workspacePath]/route.ts`, `frontend/src/lib/workspace-models.ts`.
- Main menu/dashboard/sidebar: `frontend/src/app/menu/main-menu-screen.tsx`, `frontend/src/app/menu/menu-dashboard-data.ts`, `frontend/src/lib/menu-data.ts`, `frontend/src/lib/menu-icons.ts`.
- System settings screens: `frontend/src/app/system-settings/system-settings-screen.tsx`, `frontend/src/lib/system-setting-screens.ts`, `frontend/src/app/api/system-settings/[[...settingPath]]/route.ts`.
- User management: same as System settings plus `backend/internal/authentication/**` and `backend/internal/shop/**` only when server behavior is involved.
- Permissions/menu access: `frontend/src/lib/menu-data.ts`, `frontend/src/lib/menu-permissions.ts`, `frontend/src/app/system-settings/system-settings-screen.tsx`, related backend permission services when needed.
- Language/i18n: frontend language caller first, then exact key lookup in `backend/assets/language/languages.tsv`; record provisional keys under `backend/prompts/language_requests/`.
- LINE OA/linking: `frontend/src/app/line-oa/**`, auth LINE API routes, backend LINE OA handlers only when API behavior is involved.
- Backend auth/password/shop access: `backend/internal/authentication/authentication_http.go`, `backend/internal/authentication/services/authentication_service.go`, `backend/internal/authentication/models/user.go`, `backend/internal/shop/**`.

## Manual Rule
Do not create or update manuals during normal work. Manual generation is only when explicitly requested or at final project manual phase.

## Escalation
Use broad repo review only when the user asks for whole-system review, security review, production readiness, or architecture changes.
