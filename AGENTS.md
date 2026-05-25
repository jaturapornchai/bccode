# BC Ai Account Project Rules

## Communication
- Communicate with Jead in Thai.
- Address the user as `ลุงจืด` unless asked otherwise.
- Keep answers concise and backed by source, diff, command output, logs, or tests.
- **English Rule for Agent Assets**: Create and maintain all new guidelines, rules, planning artifacts (`implementation_plan.md`, `task.md`, `walkthrough.md`), instructions, and custom skills in English to save tokens and improve execution speed.
- Any new or updated skill, rule, agent prompt, handoff, checklist, workflow document, or reusable AI instruction that may consume recurring context must be written in English unless Jead explicitly requests Thai output for end users.


## Source Of Truth
- Read active project files first: `AGENTS.md`, `CLAUDE.md`, README, source code, tests, and local docs.
- Do not rely on deleted shared skill folders.
- Do not guess APIs, schemas, enum values, SQL, workflows, or config.
- **Conflict Rule**: `AGENTS.md` and `D:\bccode\.agents\rules\bc-account-core-rules.md` are the canonical rule sources. `AI_INDEX.md`, `.agents\wiki\llm-index.md`, and local skills are routing layers. If a skill/wiki/LLM page duplicates and conflicts with the canonical rules, update that downstream file to link back to the canonical source instead of creating another variant.
- **Token Budget Rule For Agent Assets**: Keep skills, wiki/LLM pages, prompts, handoffs, checklists, and workflow docs router-first and short. Put durable project facts in the canonical rule/source file once, then link to it. Do not copy long runtime, deployment, storage, version, or security blocks into every skill.
- **Skill Upgrade Rule**: If any code change modifies, refines, or affects a system pattern, layout contract, or business rule that relates to an existing skill (found under `D:\bccode\.agents\skills\`), you MUST proactively update and upgrade the matching `SKILL.md` file immediately. This ensures that the agent's core skills stay synchronized, prevents regressions, and enables continuous, incremental development without reverting to old behaviors.

## AI Truth Verification Rule
- AI must not guess, fabricate, invent, embellish, or fill missing details as fact.
- Project facts must be backed by active repo source, local docs, tests, runtime output, logs, diff evidence, database/query evidence, or explicit user-provided data.
- External/current facts must be verified on the internet before being stated as fact. Use multiple reliable sources where practical, prefer official/primary sources, and cross-check important claims against at least two credible sources.
- If a fact cannot be verified, state clearly that it is unverified and do not present it as true.
- Distinguish confirmed evidence, assumptions, and recommendations. Never mix assumptions into factual summaries.

## No Fallback Rule
- Do not silently substitute missing config, API routes, storage providers, database values, language keys, business defaults, credentials, tenant/branch context, or user/company data.
- If a required value is missing, invalid, unauthorized, unreachable, or unverified, stop that operation and show a clear error with the real reason and the missing source/key/env/route when it is safe to reveal.
- Do not route image/file upload to old storage, mock data, legacy endpoints, local labels, hardcoded defaults, or derived credentials when the configured source is unavailable.
- Migration compatibility must be explicit and tested; it must not be hidden as automatic fallback behavior.

## License / Copyright Rule
- Treat third-party code, templates, UI kits, design systems, images, icons, fonts, screenshots, generated assets, sample code, and documentation as legally sensitive until their license and usage rights are verified.
- Use external business/UI systems such as SAP Fiori, Odoo, ERPNext, React-admin, Refine, shadcn/ui, and similar products only as references for general UX patterns, workflows, and floorplans. Do not copy their code, assets, branding, screenshots, visual identity, or screen layouts one-to-one.
- New UI must be implemented as original BC Ai Account work using project-owned code and approved dependencies. Inspiration is allowed; direct reproduction is not.
- Before adding a dependency, template, component registry item, icon pack, font, image, or generated asset, verify the license from an official or primary source and record enough evidence in the task summary or local docs when the choice affects distribution.
- Prefer permissive licenses that are normally suitable for commercial/customer installation, such as MIT, Apache-2.0, BSD, and ISC, after verifying the actual project license.
- GPL, AGPL, LGPL, copyleft, source-available, commercial, trial, no-license, unclear-license, or trademark-heavy materials require explicit review before use. Do not copy GPL/AGPL project code into BC Ai Account unless the distribution obligation is intentionally accepted and documented.
- Keep required copyright notices, license files, and attribution when a license requires it. Do not remove upstream license headers from copied permitted code.
- Do not use third-party logos, product names as branding, screenshots, proprietary icons, trademarked illustrations, or customer/vendor assets unless the project has explicit permission or the asset is already owned/approved by the project.
- For AI-generated images or assets, use them only when generation rights and commercial usage are acceptable, and avoid imitating a living artist, proprietary brand style, or identifiable copyrighted work.
- If license status is unclear, block usage and choose an original implementation or a verified permissive alternative.

## AI Coding Rules (Pareto-style)
- Be concise, technical, actionable, and answer in Thai.
- **Brainstorm & Plan First Rule**: Before editing, creating, or implementing any features or UI, brainstorm the absolute best, most beautiful, and easiest-to-use UX/UI and technical solutions. Outline and review the implementation plan, write/update code incrementally, and perform thorough testing to verify the changes before claiming completion.
- No magic: if infra, code, API, schema, or file path is not verified, state the assumption or ask first.
- Verify before saying done. Use tests, build, curl, logs, browser check, or diff evidence; if verification is not possible, say why.
- Use web/docs for external facts, especially facts that may change such as versions, pricing, recent releases, external API behavior, legal/licensing, standards, product/service capabilities, and public company/provider claims.
- Surface risky assumptions before acting. For irreversible R0 work such as dropping data, force-pushing main, production deploy contracts, or sending real orders, stop and ask. For R1 work such as API/schema changes or production restarts, explain impact first. For R2 work such as local UI/file edits/tests, proceed.
- Keep the smallest safe change: patch first, refactor only when systemic, rewrite only with clear need and approval.
- Prevent scope drift. Fix only the requested issue; flag unrelated findings separately.
- Security is mandatory: no hardcoded secrets, no PII in logs, validate inputs at boundaries, use parameterized SQL, and grep for `sk-`, `Bearer `, and `password=` before commit/deploy when relevant.
- Match existing project style. Names should describe intent, one function should have one responsibility, and comments should explain why, not what.
- Optimize performance only with evidence. Avoid N+1 queries, `SELECT *`, and loading whole tables when pagination/filtering is suitable.
- Feature docs must include objective, workflow, config, dependency, usage example, and limitation. Do not duplicate changelog/version history.
- For substantial task closeouts, always end with: `✅ Pros`, `⚠️ Cons / Risks`, and `💡 Recommendations`.

## No Mock Data Rule
- Do not create or rely on mock, fake, dummy, demo, sample, placeholder, or invented business data when testing workflows, screens, API behavior, business logic, permissions, approvals, reports, tenant isolation, or database behavior.
- Use real data from the selected DEV database for development verification. For UAT or PRO, read or write data only when the task explicitly allows that environment and the action is safe under the environment rules.
- If a test needs temporary records, create them in the real DEV database through the real application/API flow, mark them with a clear test prefix, verify the behavior, and clean them up before completion.
- Do not replace real integration checks with mocked API responses for feature completion claims. Mocking low-level technical failures is allowed only for isolated unit tests where no business data or business behavior is being validated.
- When debugging data issues, query or search the real selected DEV data source through the approved path: backend API, MongoDB Atlas, server-hosted PostgreSQL, server-hosted ClickHouse, Kafka, or Redis. If real data access is blocked, say so and mark the answer unverified.
- Final reports must distinguish real database verification from unit tests or static checks.

## Context Budget Rule
- Keep Codex/agent context small by default. Start from the requested module, current source, tests, and narrow runtime evidence; do not scan or load the whole repo unless explicitly requested.
- Start each coding task from `AI_INDEX.md` when available. Use it as the routing map, then open only the listed files for the task area.
- For rule-sensitive work, read `D:\bccode\.agents\rules\bc-account-core-rules.md` as the single runtime/deployment/storage/version reference; do not load every skill just to re-read duplicated rules.
- Use `rg` and `rg --files` so root `.ignore` is respected. Do not use broad recursive `Get-ChildItem`/full-tree reads unless a task specifically requires inventory.
- Do not read generated or bulky files into context by default: `backend/docs/**`, `backend/api/swagger/**`, `backend/assets/fonts/**`, `backend/tdict-std.txt`, lockfiles, screenshots/images, `.playwright-mcp/**`, `manual/*.json`, build outputs, runtime logs, dependency folders, duplicate skill packs, or legacy Flutter/reference trees, except when the task is explicitly a frontend migration/clone/reference-screen task.
- For `backend/assets/language/languages.tsv`, never open the full file. Query exact keys only, for example `rg -n "^permission_link\t|^new_item\t" backend/assets/language/languages.tsv`.
- During fast UI iteration, do not edit `languages.tsv` for every small label change. Use stable language keys, show an explicit missing-translation error when a key is absent, and record missing or provisional keys under `backend/prompts/language_requests/` for later batch translation.
- For API behavior, read handlers/services/tests first. Use generated Swagger only when the task is specifically about OpenAPI docs.
- Use project-local root skills under `D:\bccode\.agents` only. Ignore duplicate skill bundles under `backend/.agents`, `.claude`, `.cline`, `.roo`, `.kiro`, `.kilocode`, and `bcai-claude-skills` unless explicitly requested.
- Prefer line-range reads and symbol/function-level inspection for files over 50 KB. Files currently known to be large include `frontend/src/app/system-settings/system-settings-screen.tsx` and `frontend/src/app/menu/main-menu-screen.tsx`.
- For routine git checks, avoid broad `git status` or `git diff` in a dirty tree. Use exact paths or module scopes, report counts when the output would exceed 50 lines, and never paste long status/diff output into chat unless requested.

## AI Agent Speed Rule
- Default to fast, narrow execution: route by `AI_INDEX.md`, inspect exact files/functions, patch the smallest scope, and verify with focused commands.
- Do not open full large files by default. For files over 50 KB, use `rg -n`, line ranges, symbol search, or component extraction before reading more.
- Do not run whole-repo scans, broad lint/test commands, or generated-doc reads unless the user asks for full-system review or the narrow check cannot answer the task.
- For normal UI/bug work, prefer medium reasoning. Escalate to high/xhigh only for architecture, security, cross-service data model changes, or unclear root-cause debugging.
- Keep handoffs short and actionable: path, function/component, command, result. Do not paste large source blocks into chat.

## Frontend Migration Rule
- Use `D:\bccode` from `https://github.com/jaturapornchai/bccode` as the active target workspace.
- Use `D:\bcdev` from `https://github.com/jaturapornchai/bcdev` as the legacy/reference workspace.
- Legacy frontend reference/template is Flutter under `D:\bcdev\frontend`, especially `D:\bcdev\frontend\bcaiaccount`.
- Do not use the old `jaturapornchai/bcai` remote as the source for `D:\bcdev` or `D:\bccode`.
- Target frontend implementation is Next.js under `D:\bccode\frontend`.
- Do not recreate or copy Flutter/Dart/native platform scaffolds into `D:\bccode\frontend`.
- When migrating a screen or workflow, read the Flutter source first for behavior, labels, layout intent, validation, and API usage, then implement the equivalent in Next.js.
- When Jead asks to clone, migrate, redesign, or fix a legacy BC screen, always use the old Flutter screen and any user-provided screenshot/reference screen as the primary visual and workflow reference before editing Next.js.
- The Next.js screen should preserve the Flutter screen's user-facing intent: menu hierarchy, route purpose, labels, core actions, form fields, validation behavior, empty/loading/error states, permission behavior, and data/API flow. Improve implementation quality and responsive behavior, but do not invent a different workflow unless Jead explicitly approves it.
- If the matching Flutter screen cannot be found, state what was searched and treat the implementation as blocked or provisional instead of guessing.
- Preserve BC Ai Account business rules and multi-tenant constraints while migrating.
- If a required backend API is missing or unclear, create/update an API spec prompt under `backend/prompts/api_requests/` instead of guessing backend behavior.

## Multi-Tenant Rule
- `tenant_id` is the canonical tenant boundary for BC Ai Account going forward.
- `tenant_id` means one company/business/legal entity/workspace, not the owner user.
- For existing production data, `tenant_id` is a logical alias whose value must be the existing core `shopid`. Do not create a new tenant identifier for old data.
- One owner/user can own or access many `tenant_id` values through a membership/role table. After login, selecting a company selects the active `tenant_id`.
- For owner-level overview across many companies, use `company_group_id` above many `tenant_id` values; do not weaken per-tenant authorization.
- Existing core storage uses `shopid` as the physical tenant key in MongoDB, PostgreSQL, Kafka payloads, and ClickHouse rows. Some GoAPI/MCP/AI/approval DTOs use `shop_id`; inspect the module before assuming. Do not add `tenant_id` columns or rename existing fields only for naming consistency.
- New code should speak `tenant_id` at the API/service boundary and map to the module's physical key: usually `shopid`, sometimes `shop_id` in newer GoAPI/MCP modules.
- Every API, query, cache key, object-storage path, Kafka message, report job, background worker, and audit log that touches customer data must carry and validate `tenant_id`.
- Backend must derive and authorize `tenant_id` from JWT/session/workspace membership. Frontend may pass `tenant_id` for context, but backend must not trust it without permission checks.
- `branch_id` / branch code is a sub-scope under `tenant_id`, never a replacement for `tenant_id`.
- Department, work day, and holiday data are branch-scoped because branches can have different departments, schedules, holidays, timezones, and calendars. New UI/API paths for these screens must carry the active branch context (`branchcode`, `branchguid`, and/or `branch_key`) under the active `tenant_id`.
- Department records may be treated as company-wide only when an explicit all-branches/company-wide flag exists. Do not silently save new department, work day, or holiday records without branch context.
- New standalone tables/collections should include `tenant_id` plus useful composite indexes. Existing legacy tables/collections should keep their real physical key, usually `shopid`, unless there is a functional reason to migrate.
- Cross-tenant queries are forbidden by default. Admin/support cross-tenant access must be explicit, audited, and read-only unless approved.
- High-scale MongoDB -> Kafka -> PostgreSQL -> Kafka -> ClickHouse design must follow `backend/architecture/high-scale-multitenant-bi.md`.
- Admin, owner, company, tenant, and branch access must follow `backend/architecture/admin-access-control.md`: first admin/bootstrap can assign admin emails, owner emails map to `company_group_id`, and every data request must resolve authorized `tenant_id`/`branch_id` from backend policy.

## Environment Database Isolation Rule
- DEV, UAT, and PRO must use separate database instances, databases, credentials, object storage scopes, Kafka topics/groups, cache prefixes/databases, ClickHouse databases, and configuration values.
- DEV means active development, UAT means user acceptance testing, and PRO means real production usage. Data in these environments must not be treated as shared or interchangeable.
- Runtime code must never silently mix DEV/UAT/PRO data sources. Any environment-to-environment data movement must be an explicit, audited, development-only tool.
- MongoDB location is configurable because this is a private system. DEV/UAT/PRO must still use separate MongoDB URIs, databases, credentials, and access scopes, whether the location is MongoDB Atlas, a private MongoDB server, or another approved MongoDB deployment.
- MongoDB rollout starts with fresh DEV/UAT/PRO databases. Do not migrate or upload old MongoDB data unless Jead explicitly approves a separate migration task with source, target, backup, and rollback plan.
- Copying data from UAT or PRO into DEV is allowed only when the running environment is DEV/development/local and the target environment is DEV. Copying from DEV into UAT/PRO, UAT into PRO, PRO into UAT, or any write into UAT/PRO from a copy tool is forbidden by default.
- Production mode must disable development copy tools even if the route is reachable or a user has a valid token.
- DEV mode should emit detailed console/server logs for environment, route, source, target, tenant/shop id, counts, and safe masked connection metadata to make debugging easy. PRO must avoid noisy logs and must not log secrets, tokens, passwords, or full connection strings.
- Environment names in APIs/config should be normalized to `dev`, `uat`, and `pro`. Legacy aliases such as `development`, `local`, `production`, and `prod` may be accepted only after explicit normalization.

## Development and Deployment Rules
- Frontend runs locally on the host machine for speed, usually from `D:\bccode\frontend`.
- Backend runs on Docker Desktop for local development, using MainAPI as the single entrypoint on `http://localhost:8888`.
- After changing backend code, backend Docker config, or backend runtime config, automatically rebuild and recreate the affected Docker Desktop service before claiming completion. Default command: `cd D:\bccode\backend; docker-compose up -d --no-deps --build mainapi`, then verify `http://localhost:8888/healthz` or the changed route.
- Kafka and Redis are mandatory backend runtime services. Keep them running with MainAPI and use Docker-network addresses such as `KAFKA_SERVER_URL=kafka:29092` and `REDIS_CACHE_URI=redis:6379`.
- DEV MongoDB uses MongoDB Atlas through local env/secret files only.
- System images/files are stored on Cloudflare storage through local env/secret files only.
- Cloudflare upload runtime must use explicit R2 config: `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, and `R2_BUCKET_NAME`. A standalone Cloudflare API token must not be treated as upload-ready config.
- DEV PostgreSQL and ClickHouse run on server `45.144.166.112` with credentials from local env/secret files only.
- Deploy to DEV only when Jead explicitly says `deploy dev`.
- `deploy dev` means build and deploy both backend and frontend together.
- DEV public URLs are frontend `https://dev.bcaicloud.com` and backend `https://dev.bcaicloud.com/backend`.
- DEV public routing must go through Caddy on `dev.bcaicloud.com`; the public HTTP entrypoint is port `80`, and internal service ports must not become public workflow requirements.
- Never write server passwords, MongoDB Atlas connection strings with credentials, or Cloudflare tokens into repo files, docs, skills, prompts, logs, commits, PR text, or chat summaries.

## API Version Compatibility Rule
- Backend API contracts must be versioned. The current baseline is `v1`; future breaking changes must add `v2`, `v3`, and so on.
- Supported backend API versions must work side-by-side in the same deployed backend. When `v2` exists, `v1` and `v2` must both continue serving requests at the same time for old and new clients.
- Version selection must be explicit per request through URL versioning and/or request metadata. Do not use a single global backend mode that turns off older API versions.
- Keep `v1` backward-compatible for clients that cannot update immediately, including frontend web, iOS, and Android.
- Frontend web, iOS, and Android clients must declare the backend contract version they require. Preferred request metadata is `X-BC-Required-Backend-Version`, `X-BC-Client-Platform`, and `X-BC-Client-Version`.
- Backend responses for version-aware endpoints should expose the active backend API version and supported versions, so clients can block or warn before calling incompatible APIs.
- Release and deploy checks must include a compatibility matrix: frontend web version, iOS version, Android version, required backend version, and deployed backend supported versions.
- Do not remove, rename, or change the behavior of a `v1` contract until every dependent client version is verified as migrated or a compatibility adapter exists.

## Language Source Of Truth Rule
- Supported UI languages for the whole system are exactly: Thai (`th`), English (`en`), Chinese (`cn`), Japanese (`ja`), Korean (`ko`), Lao (`lo`), Myanmar/Burmese (`my`), Khmer (`km`), Vietnamese (`vi`), Malay (`ms`), Indonesian (`id`), and Filipino (`fil`).
- Every new visible UI label, placeholder, hint text, helper text, tooltip, option label, validation message, status, dialog text, manual text, report label, and repeated business term must eventually support all 12 languages above. Prefer backend `backend/assets/language/languages.tsv` as the source; during UI iteration, a missing key must show an explicit missing-translation error and be recorded for batch translation.
- Backend `backend/assets/language/languages.tsv` is the single source of truth for UI labels, common field labels, report names, report headers, report columns, status text, and repeated business terms.
- Default language is Thai (`th`). If no language is selected, invalid, or missing from session/config, normalize to `th`.
- Do not hardcode visible business/menu/screen/report text directly in frontend or report code. Use a stable language id/key and resolve the displayed text from backend `languages.tsv`; if the key is missing, show an explicit missing-translation error instead of temporary text.
- When adding or changing any screen, menu, report, placeholder, hint/helper text, option, status, or validation text during development, do not force an immediate `languages.tsv` edit. Add or update `languages.tsv` only when the key is already nearby, the change is release-ready, or the user explicitly asks for translations. Otherwise record the key, intended text, screen/module, and reason under `backend/prompts/language_requests/`.
- Frontend and backend reports must use the same language keys from `languages.tsv`. Do not create independent frontend-only or report-only dictionaries for business labels unless they are generated/cache layers from this source.
- Frontend must load the selected language from the backend language API (`/goapi/api/language/:lang` through the configured Backend URL/proxy) and re-render immediately when the user changes language.
- Backend report generation must accept/pass `language_code`, normalize legacy aliases such as `zh -> cn`, `jp -> ja`, `kr -> ko`, and resolve labels from the shared language system.
- If a visible translation is wrong or incomplete in a released or report-critical path, fix the backend language key and the caller's language id wiring, not a one-off frontend string.
- Missing translations must not use Thai/English substitute text. Show the missing language key/error visibly during development, then batch missing keys into `languages.tsv` before release.
- Report screens and report output must display the same translated field names, filters, titles, and columns as the frontend screen for the selected language.
- Year/calendar labels must use full translated words in every supported UI language, not Thai-only abbreviations such as `พ.ศ.` / `ค.ศ.`, except where space is intentionally constrained and a localized abbreviation is explicitly desired.

## Image Display Rule
- Fields whose type or meaning is image/file upload must display the actual image preview in read-only/detail views, not only the stored URL/path.
- Private GoAPI/R2 file paths such as `/goapi/s3/file/...` require bearer auth. Frontend previews must load them with authenticated `fetch`, convert the response to a browser-local object URL, and render that object URL. Do not use the protected path directly as `<img>`/CSS image source because browser image requests cannot attach the auth header.
- For private images, optimize speed with session-memory object URL caching only. Do not make R2 buckets public, expose direct R2 URLs, store image blobs in `localStorage`/IndexedDB, or enable persistent/shared cache unless Jead explicitly approves that security trade-off.
- Keep the stored image URL/path visible and fully readable beside or below the preview for debugging and auditability.
- If the image cannot be loaded, show a visible broken/missing-image state with the original path; do not hide the value or silently substitute another image.

## Production Scaling Rule
- Current local development and DEV deployment are Docker/Caddy based; K3s/Kubernetes is not part of the active default workflow.
- Do not load, create, or follow K3s/Kubernetes manifests or cluster docs unless Jead explicitly asks for production scaling or Kubernetes work.
- If Kubernetes is reintroduced later, keep it in a separate architecture document and do not let it override the local Docker Desktop or DEV Caddy deployment rules.
- Do not claim support for 10,000 concurrent screens until load testing proves it with real API mix, database latency, object storage, Kafka, report generation, and frontend static asset behavior.

## Frontend Responsive Rule
- Design every screen mobile-first.
- Every screen must work on common mobile, tablet, and notebook viewport sizes before it is considered done.
- Layouts must avoid fixed desktop-only widths, horizontal overflow, overlapping text, and controls that are too small to tap.
- Use responsive constraints, flexible grids, and readable touch targets; verify with browser/device viewport checks when changing UI.
- Use dense, information-first UX/UI globally through CSS variables/tokens first: keep margin, padding, gaps, textbox/input/control height, form spacing, table/list spacing, and decorative whitespace as small as practical so each screen shows maximum useful data, while preserving readable text and usable touch targets. Prefer global CSS density rules over one-off per-screen spacing.
- Use full-width, wrap-first UX/UI: screen containers, sections, cards, forms, fields, lists, tables, dialogs, and action rows must default to `width: 100%` and `max-width: 100%`, then use flex/grid wrapping to fit the viewport. Avoid fixed widths except for icon-only controls and intentionally fixed-format widgets.
- Above-the-fold chrome must stay compact. The global topbar, category/menu row, open-tab strip, screen header, and search/action toolbar are data-visibility surfaces, not hero/banner space. Keep desktop controls around compact heights (`h-8`/about 32px where practical), use small padding/gaps, keep tabs and headers low, and wrap controls into tight rows instead of increasing vertical height. Do not reintroduce large `py-3`, `h-10`/`h-11`, decorative header cards, or tall tab strips in these zones unless Jead explicitly requests it.
- Layout heights and widths must automatically expand and adjust in realtime like Flutter's `Expanded`/flex layout system. Avoid hardcoded viewport height clippings (like static `max-h-[52dvh]`) or inner scrollbars on containers (like checkbox grids or checklist matrices) inside detail views and forms. Instead, let nested components expand to their natural content height, and let the outer pane scroll container (which recalculates viewport height in realtime) handle scrolling to eliminate confusing double scrollbars.

## Frontend Theme Rule
- Every screen must support both light and dark themes.
- Every screen must include a visible icon control for switching light/dark theme.
- Each screen must have only one light/dark theme switch icon, and it must be placed in the top-most screen controls/header.
- The theme control must persist the user preference and apply a root theme marker such as `data-theme` so CSS changes immediately.
- Use CSS variables/tokens for background, foreground/text color, border, shadow, focus, and state colors; do not hardcode one-off screen colors unless there is a documented reason.
- Prefer `color-scheme`, root theme markers, and `prefers-color-scheme` only for initial theme detection, and keep components readable in both modes.
- Combobox, dropdown, menu, and select-like controls must use themed CSS for the trigger and the opened list; avoid native select popups when their option panel cannot reliably follow the active theme.
- Language selection must use a themed dialog box with national flags from the Flutter reference assets when available; do not use a combobox for language selection.
- When changing UI, verify light and dark theme rendering for the affected screen.

## Frontend Manual Rule
- Hard rule: do not auto-create, auto-update, regenerate, or bulk-generate manuals during normal feature work, bug fixes, migrations, UI changes, or screen creation.
- Create or update files under `D:\bccode\manual` only when Jead explicitly asks for that manual. Do not assume the final project phase requires manual generation; wait for Jead's direct instruction.
- When a manual is requested, write it for non-technical end users with step-by-step workflow, field/button explanations, expected results, common mistakes, troubleshooting, limitations, and next steps.
- Existing manual pages and manual links may remain, but new screens do not need a manual/help link unless a manual exists or Jead asks for that screen's manual.
- Manual pages that are created on request must still support configured UI languages and light/dark theme.

## No Business Hardcode Rule
- Do not hardcode company, branch, tenant, currency, timezone, UTC offset, working time, holiday, tax, warehouse, customer, supplier, or business-specific defaults in reusable code.
- BC Ai Account is multi-company and multi-branch. Defaults must come from persisted company/branch settings, backend language/config data, user input, environment config, or an explicit seed/setup flow.
- Thailand tax/VAT branch numbering is a legal default: `สำนักงานใหญ่` / head office uses branch code `00000`; `00001` is the first branch office, not the head office.
- Thai branch codes must be normalized as five-digit strings in persisted data and tax/document output. Input aliases such as `สำนักงานใหญ่`, `สนญ`, `HQ`, and `HO` may normalize to `00000`; numeric input such as `1` or `01` normalizes to `00001`; non-numeric or longer-than-five-digit codes must be rejected for Thai tax branches.
- A company must always keep at least one branch, and normal CRUD must not delete the head-office branch `00000`.
- Thai address selection must use the backend licensed dataset at `backend/assets/address/thailand-addresses.json` served by GoAPI `/api/address/thailand`; frontend must call it through the Backend URL/proxy, not bundle the JSON in `frontend/public`. Store official area codes in `province_code`, `district_code`, and `sub_district_code`; store five-digit postal code in `zip_code`.
- If a required business value/config/data source is unavailable, show an error and stop the operation; do not substitute a generic value or let replacement values become business rules.
- Working days, holidays, local time, timezone, and UTC offset must be scoped by company and branch where the workflow depends on location.
- Departments, working days, and holidays must be scoped by branch where branches can differ. Use the selected workspace branch when saving/listing these screens, and store branch timezone/calendar context with date-time values.
- Persist every backend/API date-time/timestamp value as UTC+0. UI may display local company/branch time, Buddhist Era, or Christian Era, but storage/API payloads must carry UTC+0 values plus explicit timezone/UTC-offset context when local display is needed.
- Do not derive business time from the client machine timezone. Use selected company/branch timezone settings.

## Verification
- Before saying a change is done, run the narrowest useful verification command for the changed area.
- For Next.js work, prefer commands from `D:\bccode\frontend\package.json` once the app is scaffolded.
- For backend/API contract work, use the relevant Go tests or documented HTTP/MCP health checks.
