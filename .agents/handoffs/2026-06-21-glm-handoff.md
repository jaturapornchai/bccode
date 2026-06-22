# Handoff → next agent (GLM) — 2026-06-21

Continue from this session. Read `AGENTS.md` + `.agents/rules/bc-account-core-rules.md` first; the new rules below are already written into them.

## Done this session
Backend changes are **deployed to DEV `192.168.2.202`** (mainapi healthy, healthz 200). Frontend changes are **`tsc` + `next build` clean but NOT deployed** to Jead's running frontend (port 3000) — see Pending #1.

1. **select-holding "holdingcode invalid" fixed** (backend, deployed) — root cause: `PersisterMongo.FindOne` swallows "no documents" (returns zero struct + nil error), so `findShopUser`/`processUserLogin`/`UpdateFavoriteShop` in `internal/authentication/services/authentication_service.go` skipped the `useruid`→`username` fallback when the token uid drifted from the membership's stored useruid. Fix: fall back when `shopUser.ID == primitive.NilObjectID`, not only on `err != nil`. → encoded in go-expert + bc-account-expert skills.
2. **`FindOne`-swallow audit** (17-agent) — found 2 more latent instances, both in the report-execute service (`reportquerym`/`reportqueryc/services/reportquery_http_service.go`), guarded with a `findDoc.Code == ""` not-found check. **Caveat:** those handlers are registered only in `cmd/app` (port 8080); the deployed mainapi (root `main.go`/8888) does NOT register `/reportm/execute` or `/report/execute`, so they're unreachable in the running build — the guards are defense-in-depth.
3. **Global toast system** (frontend, NOT deployed) — new `frontend/src/lib/toast.ts` + `frontend/src/components/toast-viewport.tsx`, mounted in `app/layout.tsx`. 13 screens migrated off per-screen inline notice banners to `pushNotice`/`toast.*`. → see new "Global Toast Notification Rule".
4. **workspace "เลือกบริษัท" empty-state button** (frontend, NOT deployed) — was permanently disabled when 0 holdings; now `handlePrimarySetup()` routes to `/holding` when no holding, else opens setup. Context-aware label.
5. **Images = PNG/JPG only** — frontend accept/validation/encoders (NOT deployed) + **backend `image_r2.go ImageUploadHandler` authoritative guard (DEPLOYED)**: ext `jpg/jpeg/png` + `http.DetectContentType` must be image/jpeg|png. Verified live: PNG/JPEG→200, WebP→400, webp-renamed-`.jpg`→400. Stored S3 objects are now PNG/JPG only; thumbnails kept (small JPG/PNG). → see new "Uploaded Image Format Iron Rule".

## New rules to follow (already in AGENTS.md + core-rules)
- **Uploaded Image Format Iron Rule** — user uploads PNG/JPG only (FE accept+guard, FE encoders PNG/JPG-no-WebP, BE ext+content-type guard), thumbnails kept, S3 stores PNG/JPG only. Static `.webp` page-background assets are exempt.
- **Global Toast Notification Rule** — use `@/lib/toast` (`toast.*` / `pushNotice`) + `<ToastViewport/>`; never build per-screen inline notice banners.
- **Mongo `FindOne` not-found gotcha** (go-expert, go-api-handler) — `FindOne` swallows no-match; test the decoded result for emptiness, not `err`.

## Pending (for GLM)
1. **Deploy the frontend** (R1 — confirm with Jead first). All frontend work above is local-verified only. Either `cd frontend && BCAI_LOCAL_BACKEND_URL=http://192.168.2.202:8888 npm run build && npm run start` on Jead's machine, or rebuild the on-prem Docker frontend (`Dockerfile.onprem`, container `bc-frontend`, port 3000 on `.202` — see memory `public-https-cloudflare-tunnel`). NOTE: backend already rejects WebP, so even the old frontend is format-safe; deploy is needed for the toast UI, empty-state fix, and FE accept filters.
2. **Old `.webp` objects already in S3/MinIO** predate the PNG/JPG rule and were NOT migrated. If Jead wants them converted, write a one-off backfill (list `app-images` bucket → re-encode `.webp` → PNG/JPG → update referencing MongoDB metadata). Optional.
3. **Rotate `bootstrap.json` secrets** (DB password + groq/openrouter/cloudflare API keys) — flagged in `2026-06-21-security-audit.md`; only Jead can (external accounts).
4. If `/reportm/execute` or `/report/execute` are ever wired into mainapi (`root main.go`), the not-found guards are already in place.

## Environment reminders
- Mongo on `.202` is a single-node replica set `rs0` (transactions). If the container is recreated it MUST re-run with `--replSet rs0 --keyFile /data/db/keyfile`. Admin mongosh from host needs `&directConnection=true`. (See memory `onprem-infra-and-run`.)
- Backend deploy = tar (exclude `bootstrap.json` to preserve the server's rs0 URI) → ssh `.202` → `docker compose up -d --build --no-deps --force-recreate mainapi` → healthz 200.
- ClickHouse `altcoin` DB on `.202` = separate project — never touch.
