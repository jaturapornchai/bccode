# Handoff → ZCode (2026-06-28)

Pick up from here. Read `AGENTS.md` → `.agents/rules/bc-account-core-rules.md` first, then this file. Communicate with ลุงจืด in Thai, short.

## Current state (verified working this session)

### DEV = fully-local stack (running now)
Per the new Environment Topology rule (core-rules): **dev runs EVERYTHING locally; deploy runs everything on `.202`. Two separate systems.**
- **Frontend:** localhost:3000 (`cd frontend && npx next build && npx next start`; never `npm run dev`/Turbopack). `.env.local` now points `BCAI_LOCAL_BACKEND_URL=http://localhost:8888` (gitignored).
- **Backend + all stores:** local Docker Desktop. Start: `cd backend && docker-compose -f docker-compose.yml -f docker-compose.local.yml up -d` (note: `docker compose` v2 NOT available here → use `docker-compose` v1; do NOT pass `--build` on the combined up — build mainapi separately because base Dockerfile uses BuildKit `--mount` and buildx is missing → use the gitignored `Dockerfile.local`).
- `docker-compose.local.yml` (committed) adds mongo(rs0)/postgres/clickhouse/minio to the existing redis/kafka/mainapi. `bootstrap.local.json`, `custom_config.local.json`, `Dockerfile.local` are gitignored (per-machine local creds, no Atlas/R2).
- **Seeded dev data** (local mongo, disposable): user `jaturapornchai@gmail.com`/`smlsoft`, holding `test`, 1 company + 1 branch (00000). The "เข้าทดสอบระบบ (dev)" button logs in with this user. If you `down -v`, re-seed: `POST localhost:8888/register-username {username,password,name}` then login → `POST /create-holding` (Bearer token) `{holdingcode:"test",name:"Test",names:[{code:"th",name:"Test",isauto:false}]}`.
- Verified end-to-end: backend healthz 200, frontend "เชื่อมต่อ Server สำเร็จ", dev login → /holding shows holding "test".

### DEPLOY = `.202` (already done this session, for the team to play)
- `deploy dev` = build+ship frontend (bc-frontend Docker, Dockerfile.onprem, :3000) + backend (mainapi Docker) + all stores, all on `192.168.2.202`. ssh `smlsoft@192.168.2.202` (key-based).
- Backend deploy: tar EXCLUDING `bootstrap.json`/`custom_config.json` (server keeps its own — verify md5 `7dbfef35` unchanged) → extract `/home/smlsoft/bc-backend` → `docker compose up -d --build --no-deps --force-recreate mainapi`.
- Frontend deploy: tar → scp `/home/smlsoft/bc-frontend` → `docker build -f Dockerfile.onprem --build-arg NEXT_PUBLIC_GOOGLE_CLIENT_ID=501250317679-05usebjtla636rm1an1dcdv3d0aeb4v1.apps.googleusercontent.com` → `docker rm -f bc-frontend` + `docker run -d` (port 3000, network `bc-backend_app-network`, env PORT/BCAI_LOCAL_BACKEND_URL=http://192.168.2.202:8888/GOOGLE_CLIENT_ID/NODE_ENV, --restart unless-stopped).
- Current `.202` already has the latest backend (documentformats struct live) + frontend (drawer/sidebar/wide-form UI).
- Public domain `app.bcaicloud.com` is RETIRED (DNS gone, new domain TBD). Use LAN `192.168.2.202:3000`.

## Done this session (committed + pushed to origin/dev)
- Multi-format document-number builder per branch (`documentformats` 13 fields, backend + frontend) replacing single prefix. Has add/remove formats, year (BE/CE × 2/4 digit), month/day/separator/runlength/resetmode/startnumber/enabled/isdefault, duplicate-prefix guard, live preview.
- Builder moved into a slide-over **drawer** (button "จัดการรูปแบบเลขที่เอกสาร") — branch form shows a compact summary.
- Collapsible icon-rail for the 8-step settings stepper (localStorage-persisted).
- Wide-form layout (narrow org tree, form takes the rest); iPad-first responsive rule set.
- Rules set: Disposable Database (all envs pre-launch incl. Kafka), Viewport iPad-first, Environment Topology (dev local / deploy .202), proactive advisor consults.

## PENDING — pick up these (priority order)

### 1. [PRIMARY] Login page UX/UI for iPad mini + iPad Air
ลุงจืด wants the login page fixed for iPad, **especially iPad mini and iPad Air**. See the detailed spec below.

Files: `src/app/login-screen.tsx` (`<main class="login-shell">` = `.brand-panel` hero + form card), `src/app/globals.css` (`.login-shell` ~L528, `.brand-panel` ~L542, `.brand-hero` ~L686, `.login-card` ~L869, login `@media` ~L886-960).

**⚠️ globals.css has MIXED line-endings (CRLF/CR/LF).** Do NOT whole-file-Edit/IDE-format it — that reformats EOL and bloats the diff. Patch byte-preserving (Node: read buffer → replace substring → write) or edit only small blocks.

Target viewports to test: iPad mini 768×1024 & 1024×768; iPad Air 820×1180 & 1180×820; desktop ≥1280.

Problem: `.login-shell` is grid `minmax(0,620px) minmax(360px,460px)` (hero+form), but `@media 901-1180px` sets `.brand-panel{display:none}` → on iPad Air landscape (1180) and iPad mini landscape (1024) the hero disappears and a ~460px form floats centered with huge side whitespace.

Fix direction (agreed with ลุงจืด, synthesized from GLM+DeepSeek — both converged):
- Change the hero-hide breakpoint from "901-1180" to **`< 1024px` only**.
  - **≥1024px** (iPad mini landscape, iPad Air landscape, desktop): always show **2-panel hero+form**. Give the form room — e.g. hero ~45-55% : form 420-460px, or asymmetric 40:60. No more narrow floating form.
  - **768-1023px** (iPad mini/Air portrait): single centered form, but bump `max-width` 460px → **520-560px**; add top spacing + logo/tagline above the card for vertical balance.
  - **<768px** (phone — not a priority): full-width + ~24px padding, just usable.
- Form width: 2-panel mode 420-460px; single portrait mode 520-560px.
- Optional premium polish on wide screens: full-height hero (min-height:100vh, vertically centered), faint logo watermark behind the form, increase in-card padding 40-56px rather than widening.

Rule: iPad-first (iPad and up is the target; phone just-usable). Verify by `next build` + open the running localhost:3000 at each viewport (Playwright/preview), confirm hero shows ≥1024 and the portrait form is 520-560px and not cramped.

### 2. [optional] Faster backend dev loop — `air` hot-reload
ลุงจืด chose "air hot-reload" to cut the 10-30s rebuild to ~2-5s. Not yet set up. Add `air` (or CompileDaemon) in a dev container: mount source + reuse the go-build-cache/go-mod volumes, incremental CGO rebuild + auto-restart (librdkafka is already cached so it's not the bottleneck). Keep `deploy-mainapi-fast.ps1` as fallback.

### 3. [low] Stale-doc cleanup in core-rules (flagged, ask ลุงจืด before sweeping)
`bc-account-core-rules.md` still mentions Cloudflare R2 (~L68-70) although R2 was replaced by MinIO, and "MongoDB Atlas" lingers in the reset-rule/debugging lines. Out of scope of the topology change — flag, don't silently rewrite.

## Key rules to honor (full list in core-rules)
- iPad-first viewport (≥768px primary; use horizontal space, multi-column; phone just-usable).
- Disposable DB: all stores (Mongo/PG/CH/Kafka) disposable across all envs pre-launch — change schema/contracts directly, no migration/backfill.
- Topology: dev fully local, deploy fully .202 (above).
- VERIFY before "done": run build/typecheck/healthz/UI, show evidence.
- globals.css mixed-EOL → byte-preserving patch only.
- Consult GLM/DeepSeek proactively on design/UX/trade-offs (`~/.claude/tools/glm-ask.py`, `deepseek-ask.py`).
- No secrets in commits; bootstrap/.env.local/Dockerfile.local are gitignored.

## Repo
Branch `dev`, all work pushed to `origin/dev` (latest `01c9777a`). Working tree clean except the running dev stack + gitignored local config.
