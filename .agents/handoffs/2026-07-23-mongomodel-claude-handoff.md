# Handoff → Claude Code (2026-07-23) — mongomodel (composite keys) + bccode follow-ups

Pick up from here. This session (Kimi CLI) worked on **D:\mongomodel** (the MongoDB diagram tool + MCP server) and the **BC Ai Account** project data inside it. Read `D:\mongomodel\AGENTS.md` first (it documents every feature/trap below), then `D:\bccode\.agents\worklog.md` entries `cont. 15`–`28` for the full history. Communicate with ลุงจืด in Thai, short.

## Environment state (IMPORTANT — not the usual)

- **Docker Desktop is DOWN** on this machine since mid-session: `Error response from daemon: Docker Desktop is unable to start` — `\\.\pipe\dockerBackendApiServer` Access denied, WSL `docker-desktop` distro Stopped, and this shell has **no admin** to restart `com.docker.service`. Fix = reboot machine, or PowerShell **as Administrator** → `Restart-Service com.docker.service`. Several `bccode` containers (mongodb etc.) are down with it.
- **mongomodel currently runs via dev server**: detached `npm run dev` on port **3100** (log `D:\mongomodel\dev-server.log`). Data dir `D:\mongomodel\data` is the same either way, so the app is fully usable.
- **Docker image `mongomodel-mongomodel:latest` is STALE** — it predates ALL of this session's features (key bar, composite keys, unique modes, arrow/animation fixes). When Docker is back: `cd D:\mongomodel && npm run docker:up` (rebuild), and kill the dev server first to free port 3100.

## MCP traps learned the hard way (save yourself an hour)

1. **Any MCP client connected before a schema change keeps the OLD tool schema and silently strips unknown arguments** (zod strips them). E.g. this session's client strips `keygroup`/`keygroupunique` — that's why model edits were done by POSTing JSON-RPC directly to `http://localhost:3100/mcp` (curl/node fetch) or via the UI. **After any server.ts tool-schema change, every AI session must reconnect/restart its MCP.**
2. **The `/mcp` HTTP route pools McpServer instances** — after editing `app/mcp/server.ts`, the running dev server keeps the old bundle. **Restart the dev server** (kill the node owning port 3100, relaunch `npm run dev`).
3. Calling `/mcp` during a dev-server recompile can return transient stale results — just retry.
4. `generate_code`/`update_field` etc. need `diagram` when the collection is not in the current tab (ids: org=`9f0722c3`, login=`4d6e9a1e`, docs=`312297bd`, masters=`efaf857d`, files=`9836076f`).

## Done this session (mongomodel)

- **Key bar on every collection node** (under description): lists all keys (PK `_id` / 🔑 business key / 🌐 session key / ⛓ composite) with their connection handles moved there (same handle ids, old edges unbroken), auto-side rendering, `← N` incoming-ref badges, left-aligned with field rows.
- **Relation direction fixed**: arrowhead now at the child (FK source) pointing into it — parent→child (`markerStart` + `orient: "auto-start-reverse"`, normalized for old persisted edges in `displayEdges`). Dash **animation also reversed** to flow parent→child via inline `animationDirection: "reverse"` merged in `displayEdges` (globals.css override did NOT work — don't retry that path). Same fix in `app/wiki/[project]/graph.tsx`.
- **Composite key groups (`field.keygroup`)**: fields sharing a group id = one key. Group box in key bar `⛓ key ผสม: a + b + c` with per-member handles, ↑↓ reorder (order = field order = compound index order), ✕ remove member / ✕ dissolve group, ⚠ when <2 members. Codegen: compound unique index (mongosh/mongoose), markdown/wiki notes. MCP: `keygroup` everywhere.
- **Group mode (`field.keygroupunique`, default true)**: toggle `ห้ามซ้ำ ⇄ ซ้ำได้` in group header — unique compound vs plain compound index. ⚠ on unique-group members that aren't `required` (duplicate-null trap).
- **BC Ai Account model redesigned via MCP** (all 5 tabs): composite keys everywhere — company `holdingcode+companycode`, branch `holdingcode+companycode+branchcode`, productbarcode `holdingcode+itemcode+barcode`, all 7 doc types `holdingcode+docno`, warehouse/customer/debtor/creditor/employee `holdingcode+code`. New relations: `branchcode→branch.branchcode` ×7 (cross-tab), `doc.creatorcode→employee.code`, `docdetail.docno→doc.docno`. `docdetail` got its missing `holdingcode` (tenant scope). `branchid` renamed → `branchcode` in 6 collections. `holding.field_4` test field gone.
- **names arrays normalized TWICE** (final decision): **canonical = `{code, name}` ONLY** — `isauto`/`isdelete` were stripped from all 42 `*names` arrays project-wide (supersedes the earlier 4-field shape).

## PENDING — pick up these (priority order)

1. **bccode: strip `isauto`/`isdelete` from names shapes to match the new model** — the diagram is now `{code, name}` everywhere, but `D:\bccode\datamodel\*.go` (Go models) and `frontend/src/components/product-barcode/names-editor.tsx` (+ any backend handlers/migrations touching names) still carry `isauto`/`isdelete`. Decide with ลุงจืด whether the runtime model follows the diagram, then implement end-to-end (full-stack ownership rule). Check `names-editor.tsx` props/usages first — removing a flag touches save/load mapping and possibly existing MongoDB documents (Disposable-DB rule: pre-launch, clean rename OK, no migration scaffolding).
2. **Docker recovery + rebuild** — get Docker Desktop running (reboot or admin `Restart-Service com.docker.service`), then `cd D:\mongomodel && npm run docker:up`, verify `:3100` serves the container, kill the temporary dev server. Also restart the other bccode containers that died.
3. **(Undecided — ask ลุงจืด) prefix-dedup in codegen**: skip a single-field FK index when a composite index's prefix already covers it. Known case: `company.createIndex({holdingcode:1})` is redundant with `{holdingcode:1, companycode:1}`. Branch's two FK singles are NOT redundant only because its compound starts with `branchcode` — after the reorder to `holdingcode+companycode+branchcode` they became redundant too. Offered to ลุงจืด, no decision yet.
4. **(Future gap) names template**: nothing enforces the `*names` array shape; new names arrays can drift again (42 arrays were re-normalized manually). Candidates: an "add standard names" helper, or a codegen-level shape check.

## Verification tooling (reusable)

- `D:\mongomodel`: `npx tsc --noEmit` and `npx tsx -e "import {demo} from './app/schema.ts'; demo()"` (schema.ts regression suite — run after ANY schema.ts change).
- Playwright driver: `D:\bccode\frontend\node_modules\playwright-core` + headless chrome at `C:\Users\jatur\AppData\Local\ms-playwright\chromium_headless_shell-1228\chrome-headless-shell-win64\chrome-headless-shell.exe`. Pattern examples in `D:\bccode\tmp\verify-*.mjs` (key bar, drag-connect, composite group, unique-mode, reorder) and `D:\bccode\tmp\test-ui-regression.mjs` (5-point canvas regression: bars/markers/animation/alignment/badge). RF pane drags: mouse direction is INVERTED to node movement; node bodies eat drags — grab empty pane.
- Fresh MCP client for tests: `node D:\bccode\tmp\test-mcp-stdio.mjs` (spawns new stdio server — sees latest tool schema).
