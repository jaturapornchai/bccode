# BC Ai Account Project Rules

## Communication
- Communicate with Jead in Thai.
- Address the user as `ลุงจืด` unless asked otherwise.
- Keep answers concise and backed by source, diff, command output, logs, or tests.

## Source Of Truth
- Read active project files first: `AGENTS.md`, `CLAUDE.md`, README, source code, tests, and local docs.
- Do not rely on deleted shared skill folders.
- Do not guess APIs, schemas, enum values, SQL, workflows, or config.

## AI Truth Verification Rule
- AI must not guess, fabricate, invent, embellish, or fill missing details as fact.
- Project facts must be backed by active repo source, local docs, tests, runtime output, logs, diff evidence, database/query evidence, or explicit user-provided data.
- External/current facts must be verified on the internet before being stated as fact. Use multiple reliable sources where practical, prefer official/primary sources, and cross-check important claims against at least two credible sources.
- If a fact cannot be verified, state clearly that it is unverified and do not present it as true.
- Distinguish confirmed evidence, assumptions, and recommendations. Never mix assumptions into factual summaries.

## License / Copyright Rule
- Treat third-party code, templates, UI kits, design systems, images, icons, fonts, screenshots, generated assets, sample code, and documentation as legally sensitive until their license and usage rights are verified.
- Use external ERP/UI systems such as SAP Fiori, Odoo, ERPNext, React-admin, Refine, shadcn/ui, and similar products only as references for general UX patterns, workflows, and floorplans. Do not copy their code, assets, branding, screenshots, visual identity, or screen layouts one-to-one.
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

## Context Budget Rule
- Keep Codex/agent context small by default. Start from the requested module, current source, tests, and narrow runtime evidence; do not scan or load the whole repo unless explicitly requested.
- Start each coding task from `AI_INDEX.md` when available. Use it as the routing map, then open only the listed files for the task area.
- Use `rg` and `rg --files` so root `.ignore` is respected. Do not use broad recursive `Get-ChildItem`/full-tree reads unless a task specifically requires inventory.
- Do not read generated or bulky files into context by default: `backend/docs/docs.go`, `backend/docs/swagger.json`, `backend/docs/swagger.yaml`, `backend/api/swagger/swagger.json`, `backend/assets/fonts/**`, `backend/tdict-std.txt`, lockfiles, screenshots/images, `manual/*.json`, build outputs, dependency folders, or legacy Flutter/reference trees.
- For `backend/assets/language/languages.tsv`, never open the full file. Query exact keys only, for example `rg -n "^permission_link\t|^new_item\t" backend/assets/language/languages.tsv`.
- During fast UI iteration, do not edit `languages.tsv` for every small label change. Use stable language keys with local fallback text, then record missing or provisional keys under `backend/prompts/language_requests/` for later batch translation.
- For API behavior, read handlers/services/tests first. Use generated Swagger only when the task is specifically about OpenAPI docs.
- Use project-local root skills under `D:\bccode\.agents` only. Ignore duplicate backend skill bundles under `.claude`, `.cline`, `.roo`, `.kiro`, and `.kilocode` unless explicitly requested.
- Prefer line-range reads and symbol/function-level inspection for files over 50 KB. Files currently known to be large include `frontend/src/app/system-settings/system-settings-screen.tsx` and `frontend/src/app/menu/main-menu-screen.tsx`.
- For routine git checks, avoid dumping legacy deletion noise. Prefer `git status --short -- . ':!clone-skills' ':!frontend/bcaiaccount' ':!frontend/bclms'` unless the task is specifically about those legacy trees.

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

## ERP Language Source Of Truth Rule
- Supported UI languages for the whole system are exactly: Thai (`th`), English (`en`), Chinese (`cn`), Japanese (`ja`), Korean (`ko`), Lao (`lo`), Myanmar/Burmese (`my`), Khmer (`km`), Vietnamese (`vi`), Malay (`ms`), Indonesian (`id`), and Filipino (`fil`).
- Every new visible UI label, placeholder, hint text, helper text, tooltip, option label, validation message, status, dialog text, manual text, report label, and repeated ERP/business term must eventually support all 12 languages above. Prefer backend `backend/assets/language/languages.tsv` as the source; during UI iteration, a stable key with a local Thai/English fallback is allowed if the missing key is recorded for batch translation.
- Backend `backend/assets/language/languages.tsv` is the single source of truth for ERP UI labels, common field labels, report names, report headers, report columns, status text, and repeated business terms.
- Default language is Thai (`th`). If no language is selected, invalid, or missing from session/config, normalize to `th`.
- Do not hardcode visible business/menu/screen/report text directly in frontend or report code. Use a stable language id/key and resolve the displayed text from backend `languages.tsv` when available; temporary fallback text is allowed only with a tracked missing-language request.
- When adding or changing any screen, menu, report, placeholder, hint/helper text, option, status, or validation text during development, do not force an immediate `languages.tsv` edit. Add or update `languages.tsv` only when the key is already nearby, the change is release-ready, or the user explicitly asks for translations. Otherwise record the key, fallback text, screen/module, and reason under `backend/prompts/language_requests/`.
- Frontend and backend reports must use the same language keys from `languages.tsv`. Do not create independent frontend-only or report-only dictionaries for business labels unless they are generated/cache layers from this source.
- Frontend must load the selected language from the backend language API (`/goapi/api/language/:lang` through the configured Backend URL/proxy) and re-render immediately when the user changes language.
- Backend report generation must accept/pass `language_code`, normalize legacy aliases such as `zh -> cn`, `jp -> ja`, `kr -> ko`, and resolve labels from the shared language system.
- If a visible translation is wrong or incomplete in a released or report-critical path, fix the backend language key and the caller's language id wiring, not a one-off frontend string.
- Missing translation fallback may use Thai/English fallback text during development to keep the UI usable. Before release, batch missing keys into `languages.tsv`; report-critical output should not ship with provisional fallback text.
- Report screens and report output must display the same translated field names, filters, titles, and columns as the frontend screen for the selected language.
- Year/calendar labels must use full translated words in every supported UI language, not Thai-only abbreviations such as `พ.ศ.` / `ค.ศ.`, except where space is intentionally constrained and a localized abbreviation is explicitly desired.

## K3s Production Scaling Rule
- Local development may keep using Docker Desktop, but high-concurrency production deployment for BC Ai Account must use the K3s deployment path under `backend/cluster/k3s`.
- Docker Desktop may use k3d with `backend/cluster/k3s/overlays/k3d` for local smoke tests only; do not treat k3d/Docker Desktop as production HA.
- Do not claim support for 10,000 concurrent screens until load testing proves it with real API mix, database latency, object storage, Kafka, report generation, and frontend static asset behavior.
- K3s production must be high-availability: multiple server nodes, external or embedded etcd/datastore with SSD-backed storage, encrypted secrets at rest, and no hardcoded credentials in manifests or repository files.
- Stateless app services such as `mainapi` must run as multiple replicas behind Service/Ingress and HPA. Stateful services such as PostgreSQL, MongoDB, Kafka, ClickHouse, Redis, and object storage must be deployed as HA managed services or dedicated clustered components, not single-node demo containers.
- Production config must come from Kubernetes Secrets/ConfigMaps or a proper config service. Do not rely on mutable per-pod `bootstrap.json` writes for multi-replica deployments.

## Frontend Responsive Rule
- Design every screen mobile-first.
- Every screen must work on common mobile, tablet, and notebook viewport sizes before it is considered done.
- Layouts must avoid fixed desktop-only widths, horizontal overflow, overlapping text, and controls that are too small to tap.
- Use responsive constraints, flexible grids, and readable touch targets; verify with browser/device viewport checks when changing UI.
- Use dense, information-first UX/UI globally through CSS variables/tokens first: keep margin, padding, gaps, textbox/input/control height, form spacing, table/list spacing, and decorative whitespace as small as practical so each screen shows maximum useful data, while preserving readable text and usable touch targets. Prefer global CSS density rules over one-off per-screen spacing.
- Use full-width, wrap-first UX/UI: screen containers, sections, cards, forms, fields, lists, tables, dialogs, and action rows must default to `width: 100%` and `max-width: 100%`, then use flex/grid wrapping to fit the viewport. Avoid fixed widths except for icon-only controls and intentionally fixed-format widgets.

## Frontend Theme Rule
- Every screen must support both light and dark themes.
- Every screen must include a visible icon control for switching light/dark theme.
- Each screen must have only one light/dark theme switch icon, and it must be placed in the top-most screen controls/header.
- The theme control must persist the user preference and apply a root theme marker such as `data-theme` so CSS changes immediately.
- Use CSS variables/tokens for background, foreground/text color, border, shadow, focus, and state colors; do not hardcode one-off screen colors unless there is a documented reason.
- Prefer `color-scheme`, root theme markers, and `prefers-color-scheme` fallback support, and keep components readable in both modes.
- Combobox, dropdown, menu, and select-like controls must use themed CSS for the trigger and the opened list; avoid native select popups when their option panel cannot reliably follow the active theme.
- Language selection must use a themed dialog box with national flags from the Flutter reference assets when available; do not use a combobox for language selection.
- When changing UI, verify light and dark theme rendering for the affected screen.

## Frontend Manual Rule
- Hard rule: do not auto-create, auto-update, regenerate, or bulk-generate manuals during normal feature work, bug fixes, migrations, UI changes, or screen creation.
- Create or update files under `D:\bccode\manual` only when Jead explicitly asks to create a manual for a screen/menu, or when Jead says the project is ready for final manual generation at the end of the project.
- When a manual is requested, write it for non-technical end users with step-by-step workflow, field/button explanations, expected results, common mistakes, troubleshooting, limitations, and next steps.
- Existing manual pages and manual links may remain, but new screens do not need a manual/help link unless a manual exists or Jead asks for that screen's manual.
- Manual pages that are created on request must still support configured UI languages and light/dark theme.

## No Business Hardcode Rule
- Do not hardcode company, branch, tenant, currency, timezone, UTC offset, working time, holiday, tax, warehouse, customer, supplier, or business-specific defaults in reusable code.
- BC Ai Account is multi-company and multi-branch. Defaults must come from persisted company/branch settings, backend language/config data, user input, environment config, or an explicit seed/setup flow.
- If a fallback is technically required, keep it generic, visible, and overrideable; do not let fallback values silently become business rules.
- Working days, holidays, local time, timezone, and UTC offset must be scoped by company and branch where the workflow depends on location.
- Departments, working days, and holidays must be scoped by branch where branches can differ. Use the selected workspace branch when saving/listing these screens, and store branch timezone/calendar context with date-time values.
- Persist every backend/API date-time/timestamp value as UTC+0. UI may display local company/branch time, Buddhist Era, or Christian Era, but storage/API payloads must carry UTC+0 values plus explicit timezone/UTC-offset context when local display is needed.
- Do not derive business time from the client machine timezone. Use selected company/branch timezone settings.

## Verification
- Before saying a change is done, run the narrowest useful verification command for the changed area.
- For Next.js work, prefer commands from `D:\bccode\frontend\package.json` once the app is scaffolded.
- For backend/API contract work, use the relevant Go tests or documented HTTP/MCP health checks.
