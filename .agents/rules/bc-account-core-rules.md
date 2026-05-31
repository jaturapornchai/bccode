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
- **Skill, Rule & Database Model Upgrade Rule**: If a code change, system behavior, database model, schema change, or debug findings modify a reusable system pattern, layout contract, business rule, or database mapping covered by the rules or skills, the agent MUST immediately update the matching rule, skill, or database model in the same task. This ensures the rules/skills/models remain accurate, up-to-date, and get smarter over time. Keep them as short rule pointers, not duplicated manuals.

## AI Capability & Instant Upgrades
**Central rule — binding on all agents equally.** All three entrypoints route here: `D:\bccode\CLAUDE.md`, `D:\bccode\GEMINI.md`, and `D:\bccode\AGENTS.md` all read this file.

- **AI Capability**: Every AI agent is capable of working full-stack across the entire project (including Frontend, Backend/Go API, Database, and MCP tools) without any model-based division of labor.
- **Rule & Skill Upgrades**: Whenever a code change, business logic change, or system behavior is implemented based on Jead's instruction, the developer/AI agent MUST immediately update the corresponding rules, skills, database models, or KM (Knowledge Management) files to keep the system up to date and prevent reversion to obsolete behaviors.

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
- Backend runs in batch mode. Do NOT auto-rebuild the Docker container while editing `backend/`. Rebuild (`cd D:\bccode\backend; docker-compose up -d --no-deps --build mainapi`) and verify `/healthz` ONLY when Jead explicitly says `rebuild` or `deploy`. Accumulate backend edits and rebuild in batches, not per change. See "Dev Workflow Mode" below.
- Do not require host Go/CGO/librdkafka setup for normal local backend runs unless the task is specifically backend compiler/toolchain work.
- Kafka and Redis are mandatory for backend runtime. Local Docker Desktop backend must run both services and configure MainAPI to reach them through the Docker network, usually `KAFKA_SERVER_URL=kafka:29092` and `REDIS_CACHE_URI=redis:6379`.
- Do not start replacement local database containers unless Jead explicitly requests isolated local testing.

## Dev Workflow Mode (Frontend Fast / Backend Batch)
- **Current phase**: the frontend is under active design/development; the backend is stable and changed only as needed. This mode is binding on all agents (Claude Code, Codex, Gemini/Antigravity) and overrides any older "must auto-rebuild on backend edit" or "typecheck after every frontend edit" instruction anywhere in the repo.
- **Frontend = fast iteration**: move quickly, brainstorm and propose alternative design/UX ideas, and rely on `next dev` HMR to preview changes in the browser. Run `npm run typecheck` only before a commit or when summarizing a chunk of work — not after every edit. Do not block idea iteration on full verification.
- **Backend = minimal + batch**: change backend only when necessary and keep edits small. Do NOT auto-rebuild the Docker container per change. Write correct code, accumulate backend edits, then rebuild (`cd backend; docker-compose up -d --no-deps --build mainapi`) and verify `/healthz` ONLY when Jead says `rebuild`/`deploy`. Summarize backend changes in rounds so Jead can review before each rebuild.

## Data And Storage
- DEV MongoDB uses MongoDB Atlas.
- MongoDB credentials must come from local env/secret files, for example `MONGODB_DEV_URI` and `MONGODB_DEV_DB`.
- System images/files are stored on Cloudflare storage, for example Cloudflare R2.
- Cloudflare upload credentials must come from local env/secret files as explicit R2 runtime config: `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, and `R2_BUCKET_NAME`.
- A standalone Cloudflare API token is not enough for this project's image/file upload path. If any R2 runtime value is missing, show the missing env/config error and stop.
- DEV PostgreSQL and ClickHouse run on the DEV server `45.144.166.112`.
- PostgreSQL and ClickHouse credentials must come from local env/secret files.
- Data store roles:
  - MongoDB is the authoritative operational source for all business CRUD, documents, master data, and user-entered data, including products, barcodes, units, categories, debtors/creditors, branches, settings, and transactions.
  - Cloudflare R2/S3 stores all images/files/binary objects. MongoDB stores only metadata, ownership context, and private file paths. Do not store image blobs in MongoDB, PostgreSQL, or ClickHouse.
  - PostgreSQL is a relational processing/projection engine for posted results, balances, stock costing, tax/VAT, AR/AP, GL, auditable relational calculations, and strict relational lookup outputs. It is not the direct CRUD source of truth.
  - ClickHouse is the BI/analytics/reporting store fed from processed facts/projections. It is read-only for BI/report consumers and must not become an operational CRUD store.

## No Fallback Enforcement
- Runtime code, tools, screens, reports, uploads, language rendering, and agent workflows must not silently substitute missing config, missing data, unavailable APIs, old endpoints, mock data, derived credentials, or legacy storage.
- If a required source is missing, invalid, unauthorized, unreachable, or unverified, fail that operation and return a visible error that includes the safe real reason, such as the missing env var, missing language key, missing route, missing tenant/branch context, or rejected permission.
- Compatibility paths are allowed only when they are explicit, versioned, documented, and tested. They must not run as hidden automatic fallback.
- Cloudflare image/file upload must error when the Cloudflare/R2 runtime config is incomplete. Do not reroute to Azure Blob, SeaweedFS, local disk, derived token credentials, standalone Cloudflare API tokens, or mock storage.

## Multilingual Data Enforcement
- Multilingual business data editors must render inputs from the active company language configuration (`settings.languageconfigs`) in that configured order. Normal data forms must not provide their own add-language button; language add/remove/reorder belongs only to the Active Languages screen. **Always use the shared `<NamesEditor>` component (located at `D:\bccode\frontend\src\components\product-barcode\names-editor.tsx`) to render localized multi-language text fields. This ensures consistent grid layouts, country flags, and a premium look and feel across all CRUD screens (including Product Master and Barcode Screen).**
- **Radio Buttons vs Combo Box (Select)**: หากฟิลด์ข้อมูลมีตัวเลือกคงที่จำนวนน้อย (ไม่เกิน 4 ตัวเลือก) ให้เลือกใช้ Radio buttons (`RadioOptionGroup` / ปุ่มตัวเลือกวิทยุ) แทน Dropdown (`CustomSelect` / Combo Box) เสมอ เพื่อให้ผู้ใช้เห็นตัวเลือกทั้งหมดได้ทันทีและลดจำนวนการคลิก เช่น ประเภทสินค้า ประเภทวัตถุดิบ ประเภทภาษี หรือตัวเลือกเปิด/ปิดการคิดคะแนนสะสม
- Removing or reducing active languages changes field visibility only. Existing stored values for hidden languages must be preserved on load and save unless that hidden language becomes active and the user explicitly edits it.
- Do not silently auto-fill, translate, or fallback one business data language value from another language. Missing language values must remain missing or be surfaced as the screen's explicit missing state.
- When active languages or workspace configurations are updated, the change must immediately take effect across the entire app without requiring a manual page refresh. Screens like `MainMenuScreen`, `SystemSettingsScreen`, and `ProductBarcodeScreen` must listen to the custom `bc-workspace-changed` event and `storage` event to reload the active languages session immediately. Form inputs, name editors, and language selectors must update and adjust their multilingual fields immediately and reactively.

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

## Thai SME Business Domain
- BC Ai Account supports Thai SME business operations, not generic admin CRUD.
- Agents working on this project must behave as Thai SME accounting and business-domain experts across accounting, marketing, sales, purchasing, trading/distribution, restaurant operations, light manufacturing, general ledger, inventory accounting, accounts receivable, accounts payable, tax/VAT-aware workflows, company/branch operations, reporting, and auditability.
- Business modules must be designed around real operating documents and lifecycle flows, such as quotation, sale order, invoice, receipt, purchase request, purchase order, bill, payment, stock receipt, stock issue, transfer, stock count, restaurant sale, production/BOM consumption, finished goods receipt, debtor/creditor aging, and GL posting.
- Do not implement accounting, inventory, tax, AR/AP, sales, purchase, restaurant, production, or reporting features as isolated data-entry screens without preserving document flow, numbering, tax/VAT treatment, stock impact, accounting impact, permissions, reports, and traceability.
- When a change touches a business process, inspect the current source, legacy Flutter behavior when relevant, existing data contracts, and real DEV behavior before implementation. Surface business impact and compatibility risk before changing schemas, APIs, postings, stock movement, tax logic, or reporting behavior.
- Product BOM/recipe setup is an independent recipe master: create a recipe code and recipe names first, then add product barcode units as ingredients or reference other recipe codes as sub-recipes. Do not make the recipe parent an existing product/barcode.

## Thai Branch Code Rule
- For Thailand tax/VAT branch numbering, head office (`สำนักงานใหญ่`) is branch code `00000`.
- `00001` is the first branch office code, not the head office.
- Default company setup and automatic head-office branch creation must use `00000` unless Jead explicitly provides a different country/legal numbering rule.
- Persist Thai tax branch codes as normalized five-digit strings. Accept UI aliases such as `สำนักงานใหญ่`, `สนญ`, `HQ`, and `HO` only as input aliases that normalize to `00000`; accept numeric input such as `1` or `01` only after padding to `00001`.
- Backend branch create/update/import endpoints must enforce this normalization and reject non-numeric or longer-than-five-digit branch codes. Frontend normalization is a UX aid only, not the source of truth.
- A company must keep at least one branch. The head-office branch (`00000`) must not be deleted through normal branch CRUD.
- **Workspace Company & Branch Selection**: When a company is selected in the workspace selection screen, the system must check the branch list. If no branch exists, it must automatically create the head-office branch (`00000` with Thai name `สำนักงานใหญ่`). The user must always be routed to the branch selection screen even if there is only a single branch available. Under no circumstances should the system automatically bypass the branch selection step.
- Company records stay minimal: company name and company address. Branch records own tax/legal/document settings such as tax ID, VAT status/rate/type, document company names, branch names, contact/address, currency, timezone, and business flags.
- Use the read-only audit command at `D:\bccode\backend\cmd\branch_code_audit\main.go` before any branch-code data migration. It must read MongoDB credentials from environment variables only and must not write data.

## Thai Address Dataset Rule
- Thailand province, district, subdistrict, and postal-code UI must use backend asset `D:\bccode\backend\assets\address\thailand-addresses.json` through GoAPI `/api/address/thailand`.
- Do not bundle the full Thailand address JSON in `frontend/public`; the frontend must load it from the configured Backend URL/proxy to keep frontend builds small.
- The dataset is generated from `thailand-geography-data/thailand-geography-json` and must keep the MIT license notice at `D:\bccode\backend\assets\address\thailand-addresses.LICENSE.txt`.
- Store official codes, not display names, in address fields: `country_code=TH`, `province_code`, `district_code`, `sub_district_code`, and `zip_code`.
- Thai address UX must support both directions: province -> district -> subdistrict -> postal code, and postal code -> filtered province/district/subdistrict choices. Auto-fill only when a postal-code match is unambiguous; otherwise keep filtered choices visible for user confirmation.
- Postal code must be recomputed when province, district, or subdistrict changes: district selection may fill postal code only when the district has one unique code or the existing postal code still matches that district; subdistrict selection must set the exact subdistrict postal code; invalid stale postal codes must be cleared.

## Tax & Accounting Correctness (Research-First)
- **Uncertainty is the trigger.** Whenever you are NOT sure about any domain / business / accounting / tax / legal / regulatory knowledge — even outside an active tax feature — do NOT guess and do NOT skip it. Go find the authoritative answer first so the program behaves correctly, then cache it (see below). Thai tax/accounting is the primary case, but the same rule covers any business-domain knowledge you are unsure about. Authoritative answers come from: competitor products, the Revenue Department (`rd.go.th`), and Thai tax law — plus official docs/specs for the specific domain.
- Thai tax / VAT / WHT / e-Tax / GL / statutory-report features must be CORRECT by Thai law and real market practice — never guessed. Before building or changing any such feature, ALWAYS research authoritative sources first:
  1. **กรมสรรพากร / Revenue Department (`rd.go.th`)** — VAT rate & rules, full tax-invoice required fields (Revenue Code §86/4), WHT rates & forms (PND 1/2/3/53/54), e-Tax Invoice & e-Receipt specs, filing deadlines.
  2. **Thai tax law** — Revenue Code, Royal Decrees, ministerial regulations. Rates/thresholds change (e.g. the 7% VAT extension is renewed by Royal Decree) — verify the CURRENT value, never trust memory.
  3. **Competitor Thai accounting/ERP software** (FlowAccount, PEAK, Express, BusinessPlus, SML, Xero TH, etc.) — to match how documents, fields, and flows are modeled in the real market so users get expected behavior.
- Use IRON web-first (WebSearch/WebFetch); cite source URL + date. If a value cannot be verified, mark it unverified and STOP — no guessed rate, required field, or form layout (HONEST UNCERTAINTY).
- **Cache every verified finding so the next task does not re-research.** Append dated, source-linked entries to `D:\bccode\.agents\skills\bc-account-expert\tax-legal-cache.md`. Keep the skill `SKILL.md` pointer short; detail goes in the cache file. This is mandatory follow-through, not optional.

## API Version Compatibility
- Backend API contracts must be versioned. The current baseline is `v1`; future breaking changes must use `v2`, `v3`, and so on without silently breaking `v1`.
- Supported backend API versions must run side-by-side in the same deployed backend. When `v2` is introduced, `v1` and `v2` must both remain callable at the same time to support older frontend web, iOS, and Android clients.
- Version selection must happen per request through explicit URL versioning and/or request metadata, not through a single global backend switch that disables older clients.
- Keep `v1` backward-compatible for clients that cannot update immediately, including web deployments, iOS apps, and Android apps.
- Frontend web, iOS, and Android clients must declare the backend contract version they require. Preferred request metadata is `X-BC-Required-Backend-Version`, `X-BC-Client-Platform`, and `X-BC-Client-Version`.
- Backend responses for version-aware endpoints should expose the active backend API version and supported versions, so clients can block or warn before calling incompatible APIs.
- Deploy or release work must verify the compatibility matrix: frontend web version, iOS version, Android version, required backend version, and deployed backend supported versions.
- Do not remove, rename, or change the behavior of a `v1` contract until every dependent client version is verified as migrated or a compatibility adapter exists.

## Backend-Centric Logic & MCP Architecture
- **Backend-centric logic**: The backend is the main logic worker and the source of truth for all business processes, validation, and calculations. The frontend is restricted to rendering UI and performing basic CRUD operations.
- **AI Agent Command via MCP**: In the future, the backend will be directly commanded and operated by an AI agent using the Model Context Protocol (MCP). Design all backend services, handlers, and workflows to be fully self-contained, API-driven, and accessible directly by AI agents without relying on any frontend state or logic.


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

## Git / Source Safety
- The local working tree is the **source of truth**. NEVER overwrite newer local code with older code from GitHub. The remote may be OLDER than local; pulling it destroys newer work.
- **R0 — STOP and ask Jead first** before any command that replaces local files with remote state: `git pull`, `git fetch` + `git reset --hard origin/...`, `git checkout origin/<ref> -- <path>`, `git merge`/`git rebase` that discards local work, `git stash drop/clear`, or re-cloning over the repo.
- Read-only inspection of the remote is allowed when it does NOT touch the working tree: `git fetch` alone, `git log origin/...`, `git diff origin/...`, `git status`. Only overwrite/merge/reset actions are forbidden.
- Push (local → remote) only when Jead explicitly asks. Default is to keep local ahead of GitHub. If local and remote diverge, surface it and ask — do not auto-resolve by pulling.
- When Jead says `push to github`, `push to GitHub`, or equivalent without explicitly narrowing the scope, treat it as a request to push the whole project: stage all repo changes with `git add -A` from the repository root, run appropriate diff/secret verification, commit, and push the current branch. Do not switch to task-only partial staging unless Jead explicitly requests a limited scope.
- After any moderate, multi-file, rule/schema/model, backend, frontend workflow, or otherwise important change, automatically push the whole project to GitHub: stage all repo changes with `git add -A`, run the narrow verification and secret checks appropriate to the change, commit, and push the current branch. Trivial read-only work does not need a push. R0 actions, remote divergence, or suspected secrets still require stopping and asking first.

## Agent Fast Execution Contract
- Applies to Codex, Claude Code, Gemini/Antigravity, and any other agent working in `D:\bccode`.
- Prefer targeted evidence over broad slow checks. Avoid full-repo scans, broad browser automation, and `go test ./...` unless Jead explicitly asks, the touched scope genuinely requires it, or narrower checks cannot provide useful evidence.
- Docs/rules-only changes: run `git diff --check` and staged secret scanning only. Do not spend time on frontend/backend build or test commands for documentation-only changes.
- Frontend changes: frontend is in fast-iteration dev mode — rely on `next dev` HMR to preview in the browser. Run `cd frontend; npm run typecheck` only before commit or when summarizing, NOT after every edit. Use browser verification only for UI behavior, layout, routing, or visual changes.
- Backend changes: write correct code and run tests for touched packages only if needed. Do NOT rebuild the Docker container until Jead says `rebuild`/`deploy`; accumulate changes and rebuild in batches. Treat repo-wide backend test failures from known CGO/Kafka/env-sensitive packages as noisy unless the touched scope points there.
- Long-running commands must be visible. State the command and reason before running it, update Jead about every 30 seconds, and if it exceeds roughly 2 minutes, report whether continuing, narrowing, or stopping is the fastest safe path.
- For meaningful changes, commit and push the whole project after targeted verification and secret checks. Do not block the push on irrelevant broad checks.

## Display and Information Completeness
- **No Truncation / No Omission**: Displayed information, text, metadata, IDs, and labels must not be cut off, truncated (e.g. using `text-overflow: ellipsis` or overflow: hidden to hide text), or omitted. All details must be fully visible and wrapped properly to fit the layout.
- **Project Date Display Contract**: Frontend business screens must render dates and timestamps through `D:\bccode\frontend\src\lib\date-time.ts` helpers (`formatDefaultDate`, `formatDefaultDateTime`, and `resolveWorkspaceDateTimeDisplayOptions`) using the active workspace timezone and year type. Do not call browser-default `Date.toLocaleString()`, `Date.toLocaleDateString()`, or ad hoc `Intl.DateTimeFormat` in screen components for business dates.

## Frontend Density Contract
- BC Ai Account screens are dense business work surfaces. Preserve vertical space for records, tables, forms, and detail panes.
- Keep above-the-fold chrome compact: global topbar/search/preferences, category/menu row, open-tab strip, screen header, and search/action toolbar should use compact controls, small padding/gaps, and low tab/header heights.
- Prefer wrap-first compact rows over tall stacked chrome. If many controls are required, wrap them tightly or use accessible icon-only utility controls instead of expanding the top area.
- Do not reintroduce large hero-like headers, decorative spacing, `py-3`/larger section padding, `h-10`/`h-11` utility controls, or tall open-tab cards in these zones unless Jead explicitly asks for a roomier layout.
- Verify UI density changes with browser evidence when practical: desktop screenshot/bounding boxes for top chrome, plus a narrower viewport check when responsive behavior changed.
- **No Native Alert/Confirm**: ห้ามใช้กล่องแจ้งเตือน/ยืนยันของเว็บเบราว์เซอร์ดั้งเดิม (`window.alert`, `window.confirm`) ใน Next.js component หรือ frontend screens เป็นอันขาด ให้เลือกใช้คอมโพเนนต์หรือ UI แจ้งเตือนแบบ Custom ที่สวยงาม (เช่น ฟังก์ชัน `confirm` จาก hook `useConfirmDialog` หรือ Dialog/Modal ของระบบ) เสมอเพื่อให้สอดคล้องกับดีไซน์ที่พรีเมียม

## Frontend Semantic Background Contract
- Every major frontend page must include a page-specific background image that clearly communicates the page's business meaning, not a generic decorative gradient or abstract filler.
- Store and reference page backgrounds as optimized `.webp` assets. Do not use PNG/JPEG for new page backgrounds unless Jead explicitly asks for a source/reference asset; keep shipped UI backgrounds WebP.
- Background images must look photorealistic and premium-camera captured: physically plausible lighting, perspective, scale, shadows, reflections, and material behavior, include presentable professional Thai people when the page context can naturally include people, with visible dimensional depth and controlled depth of field.
- Backgrounds must support both light and dark themes. Use separate theme assets or theme-safe overlays when needed, while preserving text contrast, dense business layout usability, and compact work-surface spacing.
- The background must be relevant to the screen's domain, for example Thai SMEs, Thai factories, Thai shops, accounting, inventory, purchasing, sales, production, or settings depending on the page.
- Do not set protected authenticated R2/GoAPI image paths directly as CSS backgrounds. Project-local public backgrounds are allowed; private images must follow the authenticated object-URL preview rule.
- Only Codex may generate new page background images for this project. Claude, Gemini, Antigravity, or any non-Codex agent must not generate image assets themselves; they may request/assign Codex to create the WebP backgrounds, then wire already-approved assets into the UI.

- **Mandatory Plan Before Work Rule**: Before any task, command sequence, debugging, implementation, edit, deploy, commit, or push, the agent MUST output a concise plan first. This applies to every task, including single-file changes and simple fixes; for trivial read-only Q&A, the plan may be one short sentence. The plan should state objective, scope/files when known, risk level, steps, and verification path scaled to task size. หลังจากวางแผนเสร็จแล้ว ให้ดำเนินการแก้ไขและพัฒนาต่อให้เสร็จสิ้นทันทีโดยไม่ต้องหยุดรอคำอนุมัติ เว้นแต่งานเป็น R0 หรือมีคำถามที่จำเป็นต้องหยุดถามก่อน (After planning, proceed immediately without waiting for user approval unless the action is R0 or a blocking clarification is required). Verify with source, diff, command output, logs, tests, or browser evidence before claiming completion.
