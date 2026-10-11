# BC Ai Account — Codex Router (ฉบับย่อ)

Full version (reasons, examples, history): `docs/kms/rules/agents-full.md` — open only the section you need

## Router (`D:\bccode`)
- KMS `docs/kms/` (`00`-`22` no `12`; `README.md` index; `decisions/` ADR; `bugs/`; `architecture/`; `snippets/`) = what code does, not business rules; unclear requirement→ask ลุงจืด, no guessing; code=truth (docs≠code→fix docs in same commit)
- Read only needed files (map `docs/README.md`); no sweeping folders/opening handoff early; `docs/kms/00-source-router.md` only to locate code evidence

## Rules/skills in one place — ADR `2026-09-20-single-source-ai-rules-skills.md`
- Rules=`AGENTS.md` (Claude: `CLAUDE.md` `@AGENTS.md`; Gemini CLI: `.gemini/settings.json` `context.fileName`); skills=`.agents/skills/<name>/SKILL.md` (Claude: junction `.claude/skills` via `npm run ai:link`)
- No per-tool rule files (`GEMINI.md`, `.agents/rules/*`, `.cursorrules`, `.codex/AGENTS.md`, rules in `CLAUDE.md`)→pointer/import only; no skill copies in real `.claude/skills/`, `.gemini/skills/`, `docs/skills/`, `~/.agents/skills/`
- Machine rules=`~/.agents/AGENTS.md` (copies: `node ~/.agents/sync-rules.mjs`); no project rules in global
- Read SKILL.md from disk, no cache: UI→`ui-scale-polish`; accounting/tax→`thai-accounting-tax`+`docs/kms/21-thai-tax-form-references.md` (official refs only, no guessing); skill edits commit with the work; lessons→`docs/kms/` one topic/file, cite `file:line`, list in `docs/kms/README.md`

## mydocs/ read-only
- `mydocs/`=ลุงจืด specs (`mydocs/specs/rules.md` …): re-read when changed; no AI may create/edit/delete files/folders inside; suggestions→chat/docs outside; GL: md `mydocs/specs/gl/`, SQL `mydocs/datamodels/gl/`

## Auto deploy
- Done+verify pass (test/tsc/build 100%)→deploy prod (https://account.bcaicloud.com/) unasked
- Frontend-only→no `mainapi` rebuild (retag old image); `docker save`|`ssh -C root@159.223.43.229 "docker load"`, no `images.tar`
- Preflight backup `release.env.before`+DB; atomic `release.env` switch+`docker compose up -d --no-deps frontend` (or mainapi); check HTTP 200/401+smoke-test real URL

## Champ (`D:\project-champ`) = template
ADR `2026-10-09-champ-parity-purpose-additive-extras.md`, `2026-10-11-champ-processing-workflow-template.md`, `2026-09-19-champ-parity-menu-cut.md`
- Behave like Champ: menus/screens/functions/fields+order/buttons/shortcuts/wording/steps/workflow/logic/calc results equivalent to Champ (`champ/champ/menuconfigxml/menuconfig.xml`, `BC5Account.rc`, C++/MFC, DDL, reports); no cuts unless ordered in writing (e.g. payroll); read Champ code before designing screen/menu/field/API
- Processing/workflow always=Champ: before designing/changing any processing (record/check/approve, post/GL, settle, stock, cancel/reverse, period lock/close, year-end, cost, reports) trace real Champ code (`champ-lib/BC5PROCESS` e.g. `gltlsprocess.cpp`/`glautoprocess.cpp`, the screen `.cpp`, `champ/champ/Script/SQLSERVER_Script.sql`) for step order, status/flags, checks+warnings, affected tables/balances, reversal, then follow it; cite Champ `file:line` in design/ADR/commit body; differ only for equal-result technique (e.g. synchronous `BEGIN…COMMIT` vs batch), law, or owner order, and report it; not found/unreadable→ask ลุงจืด, never invent workflow
- Champ lacks it+no order+not required by Thai accounting/tax law→never add a menu/screen/main function yourself; Champ duplicate screens→one screen with Tab/Action; menus follow `menuconfig.xml`; `goods-inspection`, `low-stock-alert`, `sale-reservation-flow` are Champ, keep
- Small extras OK unasked if additive (Champ steps/results unchanged), optional, no new top-level menu, clearly useful (e.g. search/filter, Smart Paste, Export); report at end (removable); unsure=big; big (new menu/screen/module, workflow change, new mandatory field, formula change, new capability e.g. LINE OA/e-Tax/AI assistant)→propose pros/cons, approval first

## Backend-first+REST+MCP
- Business logic/accounting-tax calc/aggregation/period close/year-end/cost allocation/statements-reports→Backend (Go+PostgreSQL) 100%; Frontend=presentation/input/basic validation, no business logic/calc engine
- Every transaction/function/query: REST API+schema+error contract; Backend MCP Server (Read/Write/Knowledge & Schema), usable via MCP and REST

## PostgreSQL only — ADR `2026-09-23-remove-mongo-kafka-redis-clickhouse.md`
- Never re-add Kafka/Zookeeper/Worker/MongoDB/Redis/ClickHouse; pure Go `CGO_ENABLED=0`, no `librdkafka`/CGO
- Accounting transactions (GL, journals, invoices, receipts, master data)=synchronous ACID `BEGIN ... COMMIT` via `sql.Tx`, no eventual consistency/queue; pagination always `LIMIT $1 OFFSET $2`+`ORDER BY`
- Prod: `postgres` (`postgres:18-alpine`, `deploy/account/compose.yml`), `minio`; session/token in table `cache_entries`; screens whose old API was Mongo="รอพัฒนา"

## PG names lowercase
- All names (database, schema, table, column, index, FK, PK, sequence)=`[a-z0-9_]`, e.g. `bcai_projection`, `gl_records`; no uppercase/camelCase/quoted identifiers (`holding_code` not `"holdingCode"`); Go `db:`/`json:` tags+TS API/DTO=same snake_case

## No payroll
- No menu/screen/spec/API for payroll, ภ.ง.ด.1/ภ.ง.ด.1ก, social-security files (สปส./กท.20 ก)
- In scope: ภ.ง.ด.2 (menu `vat-pnd2`→`/report/vatpnd2` in `frontend/src/lib/menu-data.ts`, never delete), ภ.ง.ด.3/53, 50 ทวิ, employee cash advances, employee register `/employee`
- Ref `docs/kms/19-menu-coverage-market-standard.md`, ADR `2026-09-08-menu-parity-market-standard.md`; never write "ครบ 100%"

## GL scope — ADR `2026-09-23-tax-inside-general-ledger.md`
- All tax menus under "บัญชีแยกประเภท"; GL-only user must work right away, other systems later; accounting-office system (receive, check, reconcile, close), not front-office/other-system workflow; main spec `mydocs/specs/gl/spec.md`
- One journal↔AR/AP/bank Statement many-to-many (allocations in link table); AR/AP/Statement=audit detail, not payment orders

## DevOps+Thai accountant
- Never trust HTTP 200/401/build alone; check real payload (no empty JSON/white screen/stale chunks); after deploy `docker compose ps`+`docker logs --tail 50`; compare local vs prod (`Asia/Bangkok`, env, `X-Forwarded-For`); investigate anomalies with evidence
- Assess blast radius; backup before switch; keep old release ≥72 h for rollback; evidence per function (test, DDL, API response, log)
- Defensive: `?? ""`/`|| []` fallbacks; save enabled on real edit+dirty guard; rounding/÷0/negatives/`trim()`; catch `ChunkLoadError`; OWASP, HttpOnly Secure cookies, parameterized SQL 100%, sanitized input
- Accounting: TFRS, Revenue Code, DBD; Debit=Credit 100%; VAT (ภ.พ.30/36) separate base+non-creditable input VAT; WHT (ภ.ง.ด.2/3/53, 50 ทวิ) real rates; GL audit trail per transaction; log cancel/edit reasons; never guess law/tax/formulas, check law/Revenue Dept notices/Champ/SML
- Fast keying (Tab/Enter/Numpad, Smart Paste); Thai errors that say how to fix

## Present as `[1]`,`[2]`,`[3]`
- End with status+2-3 next steps as `[1]`,`[2]`,`[3]`, no long prose (reply `1`/`1,2,3`); each=Action+problem/value+Scope (files/modules); best first marked `(แนะนำ)`; group Bug/Fix, Enhancement UX, Value Add

## Realistic Thai SME data
- All data (seed/test/master/journals) real-world-like, real Thai business/branch/transaction names; no "demo", "ข้อมูลตัวอย่าง", "(สาธิต)", "Mock", "Dummy", "Test" in user-visible data
- Chart of accounts multi-level (TFRS/DBD), never flat; JSON `"level"`+`"parentaccountcode"`: L1 `10000`-`50000`, L2 e.g. `11000`, L3 control e.g. `11100`, L4 posting e.g. `11110`; L1-L3 `allowposting:false`, L4 `true`

## No competitor brands
- Never name FlowAccount, PEAK Account (or any outside party) in code (comments, file/var names, routes, UI, commit messages) or `docs/`; use neutral words e.g. "มาตรฐานโปรแกรมบัญชีไทยทั่วไป"; `.githooks/pre-commit` rejects staged hits

## Language: key+th/en
- No hard-coded Thai/any-language text: snake_case key (`gl_post_journal`)+`backendText(dictionary,key,fallback)` (`frontend/src/lib/backend-language.ts`); text in `backend/assets/language/languages.tsv` (13 cols `key th en cn ja km ko lo my vi ms id fil`); new text=one full row in same commit (incomplete→raw key, `docs/kms/17-dev-gotchas.md`); th=source, en=real translation, others=English placeholder, never empty
- Screens: prop `language`+`useBackendLanguage(language,backendUrl)`, never `localStorage` over the prop; shared components (`confirm-dialog`, `numeric-input`) take labels as props; no `t(th,en)`; Backend: key+`language.Text(key,lang)` per `Accept-Language`/`lang`, not `fmt.Errorf("ข้อความไทย")`
- Only th+en: `ACTIVE_LANGUAGE_CODES = ["th", "en"]` (`frontend/src/lib/i18n.ts`); no 3rd language/translation round/tsv column change until ordered; keep 12-language skeleton (`src/locales/*.json`) hidden; accept th+en (Thai left in en=bug)

## UX/UI
**40+** (users=Thai 40+; beats beauty)
- Font ≥0.9rem (key text never <0.8rem); line-height ≥1.45; no mid-word Thai breaks, long tokens `overflow-wrap: anywhere`
- Primary buttons ≥2.6em (≈44px+); key actions never icon-only (Thai text or `aria-label`+`title`); WCAG AA 4.5:1 (large/UI 3:1); state never by color alone
- One CTA per screen; delete/overwrite→Thai dialog explaining effect+validate+dirty guard; plain Thai feedback, no tech codes; no sudden auto-scroll; transitions ≤300ms, honor `prefers-reduced-motion`

**Premium** (match login/holding)
- Accent from `--primary` via `color-mix`, no hard-coded colors; light/dark equally good; cards radius 16-28px, hairline border, soft primary-tinted shadow; hover lift ≤1px; focus ring 3-4px
- Motion: ambient glow ≥15s+stagger ≤0.5s only; no animated `filter: blur()`; hero=real Thai business photo webp ≥1600px, no SVG/clipart; no `overflow: hidden` on panels with popovers
- Accept once at end of batch: screenshots light+dark×1600/1280/1024/768-portrait+hover/focus/disabled/error, no console errors; dark via real theme toggle
- CSS skin pass=one block at end of `frontend/src/app/globals.css` (comment date+reason); edit via byte-preserving Node (mixed EOL), not Edit tool

**Upgrade skill**: read `ui-scale-polish` before UI work; shared standards (button icons, `!pl-10`, button height, palette, density)→update `.agents/skills/ui-scale-polish/SKILL.md` at once (pattern, anti-pattern, root cause, file:line), commit with code; AGENTS.md beats skill; never copy old un-upgraded code

## UAT: CRUD+PostgreSQL
- Full CRUD+edge cases (empty, duplicate, special chars, double-click); `psql` (container `postgres`, that holding DB) right after each operation: Create→row/values right; Update→same row changed (no new row/orphan); Delete→gone; never batch-check at end
- One command `npm run uat` (`tools/uat/`): Demo login→GL+ERP CRUD with psql after every step→UI smoke of every implemented screen (console/API errors, screenshots)→cleanup by exact id; `--rebuild` rebuilds local mainapi, `--target=prod` = read-only UI smoke on RUNGRUENG/01
- Other records must survive; clean test data by target id/code only, no broad regex; seeded randomness per `tests/uat-crud.spec.ts`; report pass/fail+evidence

## Demo button (all permissions)
- "ทดลองใช้ระบบ(Demo)" on in every env during dev: `/api/auth/demo-login`→backend `/demo-login`; env `BCAI_DEMO_LOGIN_ENABLED=true`+`BCAI_DEMO_USERNAME` (default `demo`)
- Demo=owner of holding `rungrueng` (`permission_sets`=`["*"]`, `access_scopes`=`{}` in `holding_members` of `bcai_projection`); screen refusing Demo=permission-metadata bug→fix, never use another account
- After UI/API changes test yourself (Browser pane/Playwright): Demo→`rungrueng`→company `01`→branch `00000`→real network payload, console errors, screenshot; never ask ลุงจืด to test; never type real passwords (Demo only)
- Before go-live: `BCAI_DEMO_LOGIN_ENABLED=false` or Demo to its own holding; never in a holding with real customer data

## Images: S3/MinIO
- No image binary/base64 in PostgreSQL, only URI/object key; upload `POST /api/upload/image`→`/goapi/image/upload`→PutObject to bucket from env `S3_*`; always thumbnail `<field>thumb`, list screens use it

## Speed+tokens
- No whole-file reads >300 lines unless needed; >1,000 lines→`docs/reference/CODE-MAP.md` first (stale if `wc -l` differs→`grep -n` or `pwsh -NoProfile -File tools/gen-code-map.ps1`)
- Patch in place, no whole-file rewrites; no full test suite (`npm test -- <path>.test.ts`); output >50 lines→`| head -n 30`; exploring >5 files→subagent, read-only→`subagent_type:"Explore"` (loads no rules), 5-10 line summary

## No CI (GitHub stores code only)
- `.github/workflows/ci.yml` deleted for good (restore from `1799b069`); no new `.github/` workflows or CI proposals without asking ลุงจืด first; never move `backend/.github/workflows/*.yaml` to root
- Every clone `npm run hooks:install` (never `git config core.hooksPath`: kills `post-commit`); `.githooks/pre-commit` runs `tools/gen-code-map.ps1 -Check` on files ≥950 lines/in map (skip once `SKIP_CODE_MAP_CHECK=1 git commit ...`); `.githooks/pre-push` runs `tools/verify.sh codemap`+`frontend` when `frontend/` changed (skip `SKIP_VERIFY=1 git push`)
- `npm run verify` before every push, `npm run verify:all` (+backend) before deploy; never write "CI จะจับให้"; verify not run=say so; edit test commands only in `tools/verify.sh` + update `docs/kms/11-testing-quality.md` §5 in the same commit

## History=git log, not README
- No Activity Log in `README.md` (edit only when system overview changes); commit line 1 English conventional (`feat(gl): ...`)+Thai body (what changed, result, tests); reasons→ADR; bugs→`docs/kms/bugs/`

## Auto commit
- Done+verify pass→commit unasked (before summary/deploy); unrelated work→separate commits; verify failed/unfinished→never commit as done, say so
- `git add <path>` task files only, never `git add -A`/`git add .`; check `git status -s`+`git diff --cached --stat`; grep diff for `sk-`/`Bearer `/`password=`/`api_key=`→hit=stop+ask
- Push only when ordered; never `--no-verify`; hook fails→fix+new commit, never `--amend`; in worktree (`.claude/worktrees/*`) commit on its branch, not `dev`; report hash+branch+uncommitted

## Opus lead+5.5 children
- Orchestration Rule in `~/.agents/AGENTS.md` (5.5 family; fable only when ordered): `opus`=accounting/tax/money (Rule 20)/security/auth/permissions/verify; `sonnet`=explore/UI/unit+test; `haiku`=mechanical + short Thai drafts (UI microcopy, `languages.tsv` th/en rows, menu descriptions; Lead checks before commit), never tax/law/accounting explanations, accounting KMS or completeness audits (Champ parity, secret scan, flag removal)
- Develop first, verify once: mid-work only `go build ./...`/`go test ./internal/<pkg>/...`/`npx tsc --noEmit`/`npm test -- <file>`, no per-step agent reviews; end of batch: `npm run verify:all`+one combined `git diff` review (money/tax/permissions→opus, else sonnet)+one Demo UAT (psql per operation still)+one screenshot round→commit→deploy; not done before that passes
- Children=draft: Opus re-runs verify/tests/PostgreSQL queries; children never commit/push/deploy/touch `mydocs/`; no other-vendor AI unless ordered

## Dev: forward only
- PostgreSQL in every env (local/.202/production) disposable: no migration/backfill/dual-read-write/compat shim; change schema/API contract directly; code creates structure (`CREATE ... IF NOT EXISTS`)→rebuild+deploy; no manual DDL on running DB
- New ready→delete old in same change; no feature flag/old table-route-wrapper kept "เผื่อไว้"
- Change DB/fields/screens/code unasked; excludes git push/force-push/other R0; data wiped→report what+how much (keep `tools/fast-deploy.py` pre-deploy backup)
- Go-live (ลุงจืด declares it or prod tenant has real customer data)→stops for real data: real migration (backup, counts before/after, rollback path), stop+ask each time; unsure→ask ลุงจืด, never guess
