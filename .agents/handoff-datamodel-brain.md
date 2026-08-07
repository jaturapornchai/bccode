# Handoff: Data Model Wiki + 3D Brain Graph (2026-07-14)

> From: Claude Code session → To: Codex (or any agent). Read `AGENTS.md` + `.agents/rules/bc-account-core-rules.md` first. Everything below is DONE and VERIFIED unless listed under Backlog.

## What exists now

| Piece | Location | Notes |
|---|---|---|
| Go MongoDB models (source of truth) | `backend/internal/goapi/models/*.go` | 20 files, 112 structs. Junction `D:\bccode\datamodel` points here (gitignored). |
| Obsidian wiki (Thai docs, generated from the Go structs) | `D:\bccode\datamodelwikillm` | 113 md files: `Home.md` + `collections/` (5) + `types/` (107). `[[wikilinks]]` between structs. |
| 3D graph viewer (third-party, AGPL) | `D:\bccode\jarvisui` (gitignored) | Clone of Prompt-Surfer/obsidian-jarvis-ui. Reads the wiki as a vault: 113 nodes / 223 links. Run via `start.bat` → API :3002 + Vite :5173. Vault path persisted in `~/.jarvis-config.yaml`. |
| Frontend menu screen | `frontend/src/app/menu/datamodel-graph-screen.tsx` | Route `/datamodelgraph`, menu item `data-model-graph` in `frontend/src/lib/menu-data.ts` (settings section). Screen = health-check + iframe to `http://localhost:5173`, fallback message tells user to run start.bat. |

Removed (do not resurrect): `react-force-graph-2d` dep, `frontend/scripts/gen-datamodel-graph.mjs`, `frontend/public/datamodel-graph.json` — the old 2D canvas implementation was replaced by the Jarvis embed.

## Verified state
- `tsc --noEmit` exit 0; `next build` exit 0 (31/31 pages).
- End-to-end in browser: login (dev button) → menu ตั้งค่า → "โครงสร้างข้อมูล (สมอง)" → iframe shows 3D graph, hover tooltips show Thai descriptions from the wiki, Brain layout works, FPS 60.
- Jarvis server log confirms vault read: `113 nodes, 223 links`.

## Data classification contract (2026-07-14)

- `datamodelwikillm/collections/` contains only MongoDB root documents with source/repository evidence. Keep tags `datamodel`, `mongodb`, and `mongo-root`; `#mongodb` is the filter for documents stored as their own collection.
- `datamodelwikillm/types/` contains general/supporting types with tags `datamodel` and `general-type`. Some may be embedded inside a root document, but none should be presented as an independent MongoDB collection without direct repository evidence.
- Jarvis graph parsing must normalize Windows path separators and accept CRLF frontmatter. Otherwise `collections`/`types` lose their folder grouping and newly generated notes lose their tags.

## Machine gotchas (Windows, this box)
- **npm run / npx shims are broken** ("The system cannot find the path specified"). Run binaries directly: `node node_modules/next/dist/bin/next build`, `./node_modules/.bin/tsc --noEmit`, `node node_modules/vite/bin/vite.js`, `node node_modules/tsx/dist/cli.mjs`.
- **jarvisui npm install requires `--ignore-scripts`** (sharp/libvips postinstall fails). Only side effect: semantic search disabled.
- **Frontend iteration**: `next dev` HMR is the default (fast-iteration mode); use `next build` + `next start` for production-like verification. If Turbopack hangs on this machine, kill the stale port-3000 server and fall back to build+start for that session.
- `frontend/src/app/globals.css` has mixed CRLF/CR/LF — do not let an editor rewrite whole-file EOLs.
- Wiki folder `D:\bccode\s\` is read-only scope-of-work docs — never write there.

## Backlog (in priority order)
1. **Extend wiki coverage to mainapi models** — `backend/internal/product/*/models/*.go` (product, productbarcode, bom, eorder, ~15+ structs) are NOT in the wiki yet. Same note format as existing ones (Thai, frontmatter, field table, `[[links]]`), place under `datamodelwikillm/types/` or `collections/` as appropriate; update `Home.md`. Jarvis picks up new .md files automatically (vault watcher).
2. **Wiki regen story** — the wiki is a snapshot; if Go structs change it drifts. Options: a small generator script (Go or Node) or a documented manual process. Keep it simple; the old gen script was deleted with the 2D graph, do not bring back the JSON pipeline.
3. **Jarvis auto-start (optional)** — Task Scheduler entry or similar so `start.bat` runs at boot; only if Jead asks.
4. **Handler-local structs** — collections listed in `Home.md` with "handler-local" structs (lineoa, approval, datahistory...) have no notes. Add if coverage is wanted.

## Rules that bit during this work
- Verify before "done": run typecheck + build + open the real screen in a browser.
- Testing note: CDP-driven `type` does NOT fire React onChange on controlled inputs — use native value setter + `dispatchEvent(new Event('input', {bubbles:true}))` when automating forms, or Playwright.
- DB/Kafka data is disposable pre-launch (see disposable-database rule) — model/schema changes need no migration.
