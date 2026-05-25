# BC Ai Account Core Rules

This is the central cross-agent rule file for Claude Code, Codex/GPT-5.5, and Google Antigravity/Gemini.

Related central entrypoints:
- Central skill: `D:\bccode\.agents\skills\bc-central-rules\SKILL.md`
- Central wiki/LLM index: `D:\bccode\.agents\wiki\llm-index.md`

## Precedence And Token Budget
- Canonical rule order is: `D:\bccode\AGENTS.md` -> this file -> `D:\bccode\AI_INDEX.md` / `.agents\wiki\llm-index.md` -> task-specific `SKILL.md` -> active source/tests/runtime evidence.
- This file is the single durable reference for local runtime, DEV deployment, storage, secret handling, real-data verification, and API version compatibility.
- Skills and wiki/LLM pages should route to this file instead of duplicating these facts. If a duplicate conflicts with this file, update the duplicate to point here.
- Keep reusable agent assets short, English, plain Markdown, and progressive-disclosure based. Load only the task-specific source paths or reference pages needed for the current request.
- K3s/Kubernetes is not part of the active local or DEV workflow. Do not load or maintain cluster instructions unless Jead explicitly asks for production scaling or Kubernetes work.

## Cross-Agent Compatibility
- Keep rules, skills, wiki/LLM knowledge, prompts, handoffs, checklists, and workflows in plain Markdown.
- Use explicit Windows paths and explicit commands with working directories.
- Keep verification source-backed through source, diff, command output, logs, tests, or browser checks.
- Do not depend on one vendor's hidden memory, connector-only feature, slash command, or runtime-only directive for core project behavior.
- If a tool-specific instruction is unavoidable, label the target tool and state the limitation clearly. Do not hide missing tool capability behind automatic replacement behavior.
- New or updated wiki/LLM pages must start from `D:\bccode\.agents\wiki\llm-index.md` and link to source evidence instead of becoming a large duplicated manual.
- **Skill Upgrade Rule**: If a code change modifies a reusable system pattern, layout contract, or business rule covered by `D:\bccode\.agents\skills\`, update the matching skill in the same task. Keep the skill as a short rule pointer, not a duplicated manual.

## Wiki / LLM Knowledge
- Use `D:\bccode\.agents\wiki\llm-index.md` as the shared routing layer for LLM-readable project knowledge.
- Keep wiki/LLM pages short, source-first, and token-saving.
- Prefer routing pages, source maps, and checklists over long narrative documents.
- Every non-obvious claim in wiki/LLM pages should point to active repo files, local docs, tests, commands, or runtime evidence.
- Do not store secrets, customer data, credentials, or full connection strings in wiki/LLM pages.
- If a wiki/LLM page becomes broad, split it into smaller linked pages and keep `llm-index.md` as the first stop.

## Development Environment
- Frontend runs locally on the host machine for speed, usually from `D:\bccode\frontend`.
- Backend runs on Docker Desktop for local development, using MainAPI as the single entrypoint on `http://localhost:8888`.
- Local backend startup must use Docker Desktop/Compose or an equivalent Docker run path that mounts local secret/config files at runtime.
- Backend code/config changes must be deployed to Docker Desktop automatically before completion. Rebuild and recreate only the affected service by default: `cd D:\bccode\backend; docker-compose up -d --no-deps --build mainapi`, then verify `/healthz` or the changed backend route.
- Do not require host Go/CGO/librdkafka setup for normal local backend runs unless the task is specifically backend compiler/toolchain work.
- Kafka and Redis are mandatory for backend runtime. Local Docker Desktop backend must run both services and configure MainAPI to reach them through the Docker network, usually `KAFKA_SERVER_URL=kafka:29092` and `REDIS_CACHE_URI=redis:6379`.
- Do not start replacement local database containers unless Jead explicitly requests isolated local testing.

## Data And Storage
- DEV MongoDB uses MongoDB Atlas.
- MongoDB credentials must come from local env/secret files, for example `MONGODB_DEV_URI` and `MONGODB_DEV_DB`.
- System images/files are stored on Cloudflare storage, for example Cloudflare R2.
- Cloudflare upload credentials must come from local env/secret files as explicit R2 runtime config: `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, and `R2_BUCKET_NAME`.
- A standalone Cloudflare API token is not enough for this project's image/file upload path. If any R2 runtime value is missing, show the missing env/config error and stop.
- DEV PostgreSQL and ClickHouse run on the DEV server `45.144.166.112`.
- PostgreSQL and ClickHouse credentials must come from local env/secret files.

## No Fallback Enforcement
- Runtime code, tools, screens, reports, uploads, language rendering, and agent workflows must not silently substitute missing config, missing data, unavailable APIs, old endpoints, mock data, derived credentials, or legacy storage.
- If a required source is missing, invalid, unauthorized, unreachable, or unverified, fail that operation and return a visible error that includes the safe real reason, such as the missing env var, missing language key, missing route, missing tenant/branch context, or rejected permission.
- Compatibility paths are allowed only when they are explicit, versioned, documented, and tested. They must not run as hidden automatic fallback.
- Cloudflare image/file upload must error when the Cloudflare/R2 runtime config is incomplete. Do not reroute to Azure Blob, SeaweedFS, local disk, derived token credentials, standalone Cloudflare API tokens, or mock storage.

## Image Display Enforcement
- Image/file upload fields must render an actual preview in read-only/detail screens and keep the stored URL/path visible for audit/debug.
- Protected GoAPI/R2 paths (`/goapi/s3/file/...` or `/s3/file/...`) are authenticated private object streams. Frontend preview/edit components must fetch them with the current bearer token and render a browser-local object URL; do not set protected paths directly as `<img>`, Next `<Image>`, or CSS background URLs.
- Private image performance must come from short-lived in-memory object URL reuse inside the current frontend session. Do not make R2 public, use direct R2 URLs, write image blobs to persistent browser storage, or enable shared/persistent caching without explicit approval.
- Broken or unavailable images must show an explicit failed preview state and the original stored value. Do not hide the field or silently substitute another image.

## Real Data / No Mock Enforcement
- Do not guess data state, API behavior, schema behavior, tenant access, or workflow results.
- For debugging and verification, query the real selected DEV data source through the approved runtime path: MongoDB Atlas, server-hosted PostgreSQL, server-hosted ClickHouse, Kafka, Redis, or the real backend API.
- Do not use mock, fake, dummy, demo, sample, placeholder, or invented business data to claim that business behavior is complete or correct.
- Unit-level mocks are allowed only for isolated technical failures and pure code paths; they do not replace real DEV data verification for business behavior.
- If real data access is blocked, report what was checked and mark the result unverified instead of inventing an answer.

## Thai Branch Code Rule
- For Thailand tax/VAT branch numbering, head office (`สำนักงานใหญ่`) is branch code `00000`.
- `00001` is the first branch office code, not the head office.
- Default company setup and automatic head-office branch creation must use `00000` unless Jead explicitly provides a different country/legal numbering rule.
- Persist Thai tax branch codes as normalized five-digit strings. Accept UI aliases such as `สำนักงานใหญ่`, `สนญ`, `HQ`, and `HO` only as input aliases that normalize to `00000`; accept numeric input such as `1` or `01` only after padding to `00001`.
- Backend branch create/update/import endpoints must enforce this normalization and reject non-numeric or longer-than-five-digit branch codes. Frontend normalization is a UX aid only, not the source of truth.
- A company must keep at least one branch. The head-office branch (`00000`) must not be deleted through normal branch CRUD.
- Company records stay minimal: company name and company address. Branch records own tax/legal/document settings such as tax ID, VAT status/rate/type, document company names, branch names, contact/address, currency, timezone, and business flags.
- Use the read-only audit command at `D:\bccode\backend\cmd\branch_code_audit\main.go` before any branch-code data migration. It must read MongoDB credentials from environment variables only and must not write data.

## Thai Address Dataset Rule
- Thailand province, district, subdistrict, and postal-code UI must use backend asset `D:\bccode\backend\assets\address\thailand-addresses.json` through GoAPI `/api/address/thailand`.
- Do not bundle the full Thailand address JSON in `frontend/public`; the frontend must load it from the configured Backend URL/proxy to keep frontend builds small.
- The dataset is generated from `thailand-geography-data/thailand-geography-json` and must keep the MIT license notice at `D:\bccode\backend\assets\address\thailand-addresses.LICENSE.txt`.
- Store official codes, not display names, in address fields: `country_code=TH`, `province_code`, `district_code`, `sub_district_code`, and `zip_code`.
- Thai address UX must support both directions: province -> district -> subdistrict -> postal code, and postal code -> filtered province/district/subdistrict choices. Auto-fill only when a postal-code match is unambiguous; otherwise keep filtered choices visible for user confirmation.
- Postal code must be recomputed when province, district, or subdistrict changes: district selection may fill postal code only when the district has one unique code or the existing postal code still matches that district; subdistrict selection must set the exact subdistrict postal code; invalid stale postal codes must be cleared.

## API Version Compatibility
- Backend API contracts must be versioned. The current baseline is `v1`; future breaking changes must use `v2`, `v3`, and so on without silently breaking `v1`.
- Supported backend API versions must run side-by-side in the same deployed backend. When `v2` is introduced, `v1` and `v2` must both remain callable at the same time to support older frontend web, iOS, and Android clients.
- Version selection must happen per request through explicit URL versioning and/or request metadata, not through a single global backend switch that disables older clients.
- Keep `v1` backward-compatible for clients that cannot update immediately, including web deployments, iOS apps, and Android apps.
- Frontend web, iOS, and Android clients must declare the backend contract version they require. Preferred request metadata is `X-BC-Required-Backend-Version`, `X-BC-Client-Platform`, and `X-BC-Client-Version`.
- Backend responses for version-aware endpoints should expose the active backend API version and supported versions, so clients can block or warn before calling incompatible APIs.
- Deploy or release work must verify the compatibility matrix: frontend web version, iOS version, Android version, required backend version, and deployed backend supported versions.
- Do not remove, rename, or change the behavior of a `v1` contract until every dependent client version is verified as migrated or a compatibility adapter exists.

## DEV Deployment
- Deploy to DEV only when Jead explicitly says `deploy dev`.
- `deploy dev` means build and deploy both backend and frontend together.
- DEV server IP: `45.144.166.112`.
- DEV SSH user: `root`.
- DEV SSH password must come from a local secret store or env var such as `BCAICLOUD_DEV_SSH_PASSWORD`; never write it into repo files.
- Frontend URL: `https://dev.bcaicloud.com`.
- Backend URL: `https://dev.bcaicloud.com/backend`.
- Public routing must go through Caddy on `dev.bcaicloud.com`; agents must not require direct public access to internal frontend/backend ports such as `3000` or `8888`.
- The public HTTP entrypoint is port `80`; HTTPS certificate handling is Caddy's responsibility for the public URLs.

## Secret Handling
- Never commit server passwords, MongoDB Atlas connection strings with credentials, Cloudflare tokens, API keys, access keys, refresh tokens, or bearer tokens.
- Never paste secrets into docs, skills, rules, prompts, logs, commits, PR text, or chat summaries.
- If Jead provides a secret in chat, use it only for the current runtime task when necessary, then refer to it by env var name in durable project files.
- Before commit/deploy work, check the changed files for accidental secrets when relevant.

## Display and Information Completeness
- **No Truncation / No Omission**: Displayed information, text, metadata, IDs, and labels must not be cut off, truncated (e.g. using `text-overflow: ellipsis` or overflow: hidden to hide text), or omitted. All details must be fully visible and wrapped properly to fit the layout.

## Frontend Density Contract
- BC Ai Account screens are dense business work surfaces. Preserve vertical space for records, tables, forms, and detail panes.
- Keep above-the-fold chrome compact: global topbar/search/preferences, category/menu row, open-tab strip, screen header, and search/action toolbar should use compact controls, small padding/gaps, and low tab/header heights.
- Prefer wrap-first compact rows over tall stacked chrome. If many controls are required, wrap them tightly or use accessible icon-only utility controls instead of expanding the top area.
- Do not reintroduce large hero-like headers, decorative spacing, `py-3`/larger section padding, `h-10`/`h-11` utility controls, or tall open-tab cards in these zones unless Jead explicitly asks for a roomier layout.
- Verify UI density changes with browser evidence when practical: desktop screenshot/bounding boxes for top chrome, plus a narrower viewport check when responsive behavior changed.

## Brainstorming and Planning Process
- **Brainstorm & Plan First Rule**: Before editing, creating, or implementing any features or UI, brainstorm the absolute best, most beautiful, and easiest-to-use UX/UI and technical solutions. Outline and review the implementation plan, write/update code incrementally, and perform thorough testing to verify the changes before claiming completion.
