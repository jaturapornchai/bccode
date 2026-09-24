---
date: 2026-09-24
severity: critical
component: [frontend]
tags: [bc-account, security, ssrf, settings]
fixed: true
---

# Symptom

The BFF route `frontend/src/app/api/setup/[...setupPath]/route.ts` (POST, used only by the `/settings` "ศูนย์ตั้งค่าระบบ" screen) had no session check. Anyone who could reach the public URL could:

1. **Pass `verify-password`** with an empty password, `12345` or `admin`.
2. **Read `config/get` / `config/get-raw`** without a password. The response included internal hostnames (`mainapi`, `postgres`, `minio`), the PostgreSQL user and database name, and the S3 access key id. Secrets were masked as `***`.
3. **Overwrite `config/save`** and **`change-password`** without a password.
4. **Use `test-connection` as a blind SSRF and port scanner.** It made the frontend server `fetch` any URI, or open a raw TCP socket (`node:net`) to any `host:port`, and it returned success/failure plus latency. Example targets: `postgres:5432`, `minio:9000`, or the DigitalOcean metadata service `169.254.169.254`.

`/api/storage/health` (also used only by `/settings`) had the same class of problem. Its SSRF guard blocked numeric private IPs only, so a docker hostname such as `http://postgres:5432` was fetched and its status or error message was returned.

## Root Cause

- Commit `07c9c21c` (2026-08-23) had closed the setup API with a 410 response: "until there is Control Plane authentication separate from the tenant".
- Commit `e8cc6537` (2026-09-19) re-opened it with a hard-coded default password so the `/settings` screen would stop showing that error. It also added the socket and URL probes. That commit was deployed to production (`r20260919-setup-probe`).
- The screen itself had no real purpose. **Nothing read the config it saved.** `.setup-config.json` was written inside the frontend container, and no service loads it. The Backend URL that the screen stored in `localStorage` was deleted by the login page on every load (`login-screen.tsx`, `runtimeBackendUrlForCurrentPage()`). Infra settings are configured by `deploy/account/compose.yml` and the env/bootstrap at deploy time. Champ, a Windows desktop app, needs a DB-connection dialog; the web version does not.

## Fix

The whole feature was deleted rather than put behind an OWNER session, because there is no working function worth protecting (forward-only rule, no fallback):

- Deleted `frontend/src/app/settings/{page,settings-screen}.tsx`, `frontend/src/app/api/setup/[...setupPath]/route{,.test}.ts`, `frontend/src/app/api/storage/health/route{,.test}.ts`, `frontend/src/lib/setup-config{,.test}.ts` and `frontend/manual/settings.json`.
- Removed the gear link to `/settings` from `AppHeaderControls` (and its `showSettings` prop). Also removed the "ไปตั้งค่า →" link from the login connection-error box.
- Cleaned up everything left dead by the deletion:
  - 160 CSS rules in `globals.css` that only matched classes rendered by the deleted screen. They were removed with a postcss script, and every remaining rule was checked to be byte-identical in body.
  - 74 rows in `languages.tsv` whose keys no other file uses.
  - The `NEXT_PUBLIC_DEFAULT_BACKEND_URL` build-arg.
  - The Playwright HC-04 test and the `/settings` UAT step.
- `/api/backend/check` stays. It only probes the configured `serverGoApiBase()` and ignores user-supplied hosts.

## Regression Test

- `frontend/src/app/api/removed-setup-api.test.ts` fails if `src/app/settings`, `src/app/api/setup` or `src/app/api/storage/health` comes back, and if any BFF `route.ts` imports `node:net`.
- `tests/login-header-controls.spec.ts` HC-01 now expects 4 header controls on the login card (font, palette, theme, language).
- If a future screen genuinely needs infra settings or connection tests, it must be a backend endpoint behind an authenticated OWNER session. It must test only the configured service URLs, never a host/port from the request body.

## Regression (2026-09-24 → 2026-09-25)

- The fix was first deployed from its branch as `r20260924-sec-setup-1` but was never merged into `dev`. The next deploy from `dev` (`r20260924-1`) brought `/api/setup/*` back on prod, where it stayed live until the merge into `dev` was deployed on 2026-09-25 (GET returned 405 = the POST handler was deployed again).
- **Lesson:** never run `tools/fast-deploy.py` from a branch or worktree whose commits are not in `dev`. Merge into `dev` first; otherwise the next deploy from `dev` silently reverts it. Check `/etc/bcai-account/release.env` and `/opt/bcai-account/releases/*/release.env.before` on the server for the release chain.
