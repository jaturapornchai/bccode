# Project IRON LAW (model-agnostic — Claude, Codex, Gemini follow identically; loaded every turn, ≤2000 tok)

## Communication
- Reply in Thai, action-first. No "ผมจะ..." preamble. Address Jead as `ลุงจืด`.
- Code comments in English only.
- No closing summary if diff already shows result.

## AI Capability & Instant Upgrades (ความสามารถและการอัปเกรดระบบกฎ)
- **AI Capability**: ทุก AI Agent มีความสามารถทำงานทดแทนกันได้หมดในทุกส่วนของระบบ (Full-stack: ทั้ง Frontend, Backend, Database และ MCP tools) โดยไม่มีการแบ่งแยกหน้าที่ตามโมเดล (ไม่มีการแบ่งแยกเฉพาะ Gemini = Frontend หรือ Codex = Backend อีกต่อไป)
- **การปรับปรุงกฎและทักษะทันที**: หากมีการอัปเดตโค้ด ปรับปรุงตรรกะ หรือระบบใด ๆ ตามคำสั่งของลุงจืด ให้ผู้พัฒนา/AI ทำการปรับปรุงกฎ (Rules), ทักษะ (Skills) หรือองค์ความรู้ (KM) ของระบบให้สอดคล้องเสมอทันที เพื่อให้ระบบความรู้ของ AI ทันสมัยและไม่กลับไปเขียนหรือแก้เป็นแบบเดิม


## 🧠 REASONING DEPTH (all models — map to your own knob)
Match reasoning depth to task complexity. The LEVEL below is the shared contract; each model maps it to its own mechanism: **Claude** = effort level · **Codex** = `model_reasoning_effort` · **Gemini** = `thinking_level`.

| Task type | depth | When |
|---|---|---|
| Q&A, lookup, format | **minimal** | "What is X?", rename, comment |
| Boilerplate, CRUD, simple fix | **low** | /qcrud, single-file edit |
| Complex coding, multi-file | **medium** (default) | refactor, debug, /qui |
| Architecture, novel logic, multi-file rewrite | **high** | /plan, design decision |

Default is **medium**. If quality feels low, escalate one level explicitly. Do NOT inflate the prompt with verbose chain-of-thought — use the reasoning knob instead.

## 🚫 DO NOT change sampling defaults
- temperature, top_p, top_k → **never modify** (keep each model's defaults)
- Use the reasoning-depth knob instead of CoT prompting

## Speed & Token
1. **READ minimum** — only files you need. Never scan whole dir.
2. **NO inline completion** for new features → use Manager View.
3. **PLAN FIRST EVERY TIME** — before any task, command sequence, debugging, implementation, edit, deploy, commit, or push, output a concise plan first (≤10 lines for normal work). This applies even to single-file/simple fixes; for trivial read-only Q&A, the plan may be one short sentence. หลังจากวางแผน ให้ลุยทำต่อทันทีจนเสร็จโดยไม่ต้องรออนุมัติ เว้นแต่งานเป็น R0 หรือมีคำถามที่จำเป็นต้องหยุดถามก่อน (After planning, proceed with changes immediately without waiting for approval unless the action is R0 or a blocking clarification is required).
4. **Finish The Given Task** — เมื่อได้รับคำสั่งจาก Jead แล้ว ต้องพยายามทำให้เสร็จ end-to-end ในเทิร์นเดียว รวมถึงแก้โค้ด/กฎ/datamodel/UX/UI ที่เกี่ยวข้องและ verify ด้วยหลักฐานจริง ห้ามหยุดที่ข้อเสนอหรือรายงานครึ่งทาง ยกเว้นติด R0, ข้อมูลจำเป็นที่หาเองไม่ได้, หรือข้อจำกัด runtime ที่ทำต่อไม่ได้จริง ต้องทดลองใช้งานแบบผู้ใช้ไม่เก่งคอมพิวเตอร์และทดสอบให้ลึกที่สุดเท่าที่ขอบเขต/เวลา/เครื่องมืออนุญาต ถ้างานมีความเสี่ยงหรือกลัวว่าแก้ผิด ให้ push GitHub อัตโนมัติหลัง targeted verification และ secret/diff checks เพื่อมีจุด rollback ยกเว้นพบ remote divergence, secret risk, หรือ R0 blocker
5. **ONE concern per turn** — don't bundle unrelated changes.
6. **Output diff, not full file** when editing existing code.
7. **Schema-first outputs** (JSON schema) — more token-efficient than prose.
8. **Clear thought preservation** if turn is simple Q&A (saves multi-turn input growth).
9. **Push means whole project** — when Jead says `push to github`, `push to GitHub`, or equivalent without an explicit narrower scope, stage all project changes (`git add -A` from repo root), run the narrow verification/secret checks that fit the change, commit, and push the current branch to GitHub. Do not limit the push to only the current task's files unless Jead explicitly says so.
10. **Auto-push meaningful changes** — after any moderate, multi-file, rule/schema/model, backend, frontend workflow, or otherwise important change, automatically stage the whole project, run the appropriate narrow verification and secret checks, commit, and push the current branch to GitHub. Do not wait for a separate push request unless the change is trivial/read-only or an R0/divergence/secret blocker requires asking first.
11. **External Workspaces Are Not Project Git** — Folders outside `D:\bccode` are not part of this GitHub project. Do not stage, commit, push, or create GitHub assumptions for any external workspace unless Jead explicitly provides a separate repository/remote for that folder.
12. **Agent Fast Execution Contract** — Codex, Claude Code, Gemini/Antigravity, and other agents must prefer targeted evidence over broad slow checks. Docs/rules-only changes need only `git diff --check` plus secret scan. Frontend code changes rely on `next dev` HMR — run `npm run typecheck` only before commit/summary, not every edit. Backend changes: write correct code, run touched-package tests only if needed, and do NOT rebuild Docker until Jead says `rebuild`/`deploy`. Do not run `go test ./...`, broad browser automation, or full-repo scans unless Jead explicitly asks or targeted checks are insufficient. Announce long commands, update every ~30 seconds, and after ~2 minutes report whether to continue, narrow, or stop.
13. **Auto Plugin Selection Rule** — Codex and other agents should automatically choose and use the most appropriate available plugin, connector, MCP tool, or skill for the task (for example Browser/Playwright for UI checks, GitHub for GitHub work, Drive/Docs/Sheets/Slides for those file types, and project-local skills for BC Account). Do not install new plugins/connectors without Jead's approval, and never replace source/runtime verification with plugin assumptions.

## 💾 Caching strategy (all models with prompt caching)
- Stable content (system prompt, project context, source, docs) → place at TOP so it can be cached.
- Dynamic question/task → place at BOTTOM.
- Keep the stable prefix unchanged across turns to maximize cache hits (saves most input cost).

## Quality (NO MAGIC)
1. **No-Guess Evidence Rule**: AI must not guess. Before answering, designing, debugging, or editing, inspect the relevant source code, local docs, schema/model, API handler, runtime output, logs, or real UI/API behavior first. If evidence is missing, say exactly what still needs verification.
2. Never invent file path, API, lib — grep/glob first.
3. Never claim "done" without test/build output pasted.
4. Unsure → "ต้อง verify: <specific>" — don't guess.
5. R0 (drop db, force-push, `git pull`/`reset --hard` over local, deploy, real $) → STOP + ask.
6. **Backend Auto Docker Desktop Deploy Iron Rule**: After backend Go-code edits are complete, prefer the fast local path `cd backend; .\scripts\deploy-mainapi-fast.ps1`, then verify `/healthz`. Use full image rebuild `docker-compose up -d --no-deps --build mainapi` when Dockerfile, dependencies, runtime assets, compose, config, or image contents changed. Do not wait for a separate `rebuild` command. DEV server deploy still requires an explicit `deploy dev`. (See core-rules "Dev Workflow Mode".)
7. **Rule & Skill & Database Model Upgrade**: If a change, fix, or debug finding alters system patterns, database models, or logic covered by rules/skills/database schemas, update the rules/skills/models (`d:\bccode\.agents\rules\bc-account-core-rules.md`, `d:\bccode\.agents\skills\`, or `d:\bccode\AGENTS.md`) immediately so the AI grows smarter over time.


## Code Rules
- Match existing style. Don't impose new pattern.
- YAGNI — no abstraction before 2nd real use.
- Delete > refactor > add (Pareto: smallest diff wins).
- Function ≤50 lines, single responsibility.
- 3 duplicate lines OK > wrong abstraction.
- **Multilingual Fields & NamesEditor**: สำหรับช่องกรอกชื่อข้อมูลหลายภาษาใน Next.js frontend ให้เรียกใช้คอมโพเนนต์ส่วนกลาง `<NamesEditor>` เสมอ (ที่ `D:\bccode\frontend\src\components\product-barcode\names-editor.tsx`) เพื่อแสดง UI ที่สวยงามแบบ 2 คอลัมน์พร้อมธงชาติ ไม่เขียนช่องกรอกแมนนวลเองแยกในแต่ละหน้าจอ
- **Radio Buttons vs Combo Box (Select)**: หากฟิลด์ข้อมูลมีตัวเลือกคงที่จำนวนน้อย (ไม่เกิน 4 ตัวเลือก) ให้เลือกใช้ Radio buttons (`RadioOptionGroup` / ปุ่มตัวเลือกวิทยุ) แทน Dropdown (`CustomSelect` / Combo Box) เสมอ เพื่อให้ผู้ใช้เห็นตัวเลือกทั้งหมดได้ทันทีและลดจำนวนการคลิก
- **No Native Alert/Confirm**: ห้ามใช้กล่องแจ้งเตือน/ยืนยันของเว็บเบราว์เซอร์ดั้งเดิม (`window.alert`, `window.confirm`) ใน Next.js component หรือ frontend screens เป็นอันขาด ให้เลือกใช้คอมโพเนนต์หรือ UI แจ้งเตือนแบบ Custom ที่สวยงาม (เช่น ฟังก์ชัน `confirm` จาก hook `useConfirmDialog` หรือ Dialog/Modal ของระบบ) เสมอเพื่อให้สอดคล้องกับดีไซน์ที่พรีเมียม
- **Backend-centric logic**: Backend is the main logic worker. Frontend is only for rendering and basic CRUD. In the future, the backend will be directly commanded by an AI agent via MCP.
- **Database model flexibility**: สามารถปรับเปลี่ยนและแก้ไข database model/schema ได้ตามต้องการ เพื่อซัพพอร์ตการทำงานและการแสดงผลของ UX/UI ให้เป็นไปตามคำสั่ง
- **No Auto Manual Rule**: ห้ามสร้าง แก้ไข regenerate หรือ restore ไฟล์ใน `manual/` อัตโนมัติจากการแก้หน้าจอ/โมเดล/กฏใด ๆ Manual เป็นงาน opt-in เท่านั้น ต้องมีคำสั่งชัดเจนจาก Jead ว่าต้องการสร้าง manual ตัวไหนหรือชุดไหนก่อนจึงทำได้
- **Lower Snake Case Naming Iron Rule**: New or changed operational identifiers must be lowercase `snake_case`: MongoDB collection names, document fields, embedded fields, JSON keys, API query/body keys, Kafka event fields, PostgreSQL/ClickHouse projection columns, index names, schema/model constants, and any string identifier used as a persisted/API contract. Do not introduce camelCase, PascalCase, mixedCase, spaces, hyphens, or non-ASCII identifier names in new contracts. Existing legacy names may remain only as compatibility paths; rename them through explicit migration/backfill, dual-read/dual-write, and rollback notes.
- **Current ERP Ideation Mode**: During the current ERP discovery/build phase, prioritize MongoDB-first operational models and allow full model, database model/schema, backend, API, frontend, UX, and UI adjustments when they make the ERP better for Thai businesses. Keep changes source-backed, testable, and aligned with MongoDB as operational truth.
- **Local Model Authority**: ERP model/schema/API contract decisions are designed from the active `D:\bccode` source, runtime evidence, and local project rules. Do not depend on external model-document folders. New/changed operational records use `holding_code` for tenant scope, `guid_fixed` for CRUD identity, and business/reference codes such as `business_code`, `branch_code`, `business_codes`, and `unit_code` for user-facing references.
- **Mandatory Real Verification Rule**: Every model, database model, backend, frontend, UX, or UI change must be verified with the narrowest real evidence that proves it works: command output, tests/typecheck, backend health/API call, logs, or browser/UI test using real DEV data. Do not claim completion from assumptions or mock-only checks.
- **Shared CRUD Workbench Rule**: Most ERP screens are CRUD workflows and should reuse shared CRUD/workbench widgets whenever the interaction model is similar: list/search/filter on the left, inline detail/add/edit form on the right, common save/delete/confirm/dirty-state behavior, compact list styling, and shared multilingual/company-access editors. Create screen-specific UI only when the business workflow genuinely differs. After every create, update, or delete, refresh the affected list/detail/cache/workspace state immediately so related UI shows the latest MongoDB-backed data without manual reload.
- **Shared Frontend Widget Rule**: When the same UX pattern appears in multiple frontend menus or screens, build or reuse a central widget/component/field renderer before writing page-specific UI. Keep labels, validation, search behavior, empty states, save/load mapping, and accessibility consistent through that shared widget. Page-specific UI is allowed only when the workflow is genuinely different.
- **User-Facing Plain Language Rule**: Frontend menus, page titles, buttons, form labels, helper text, empty states, validation messages, and report labels must use everyday business terms that non-technical Thai SME users understand immediately. Avoid internal/technical terms such as schema, matrix, mapping, payload, raw JSON, GUID, collection, route, slug, or integration jargon in normal screens. Technical words are allowed only on explicit admin/debug/import screens and must be paired with plain-language explanation. Prefer labels such as `ชุดตัวเลือกสินค้า`, `เชื่อม Shopee`, and `จัดหมวดสินค้า`.
- **No Ellipsis For User-Facing Context Rule**: User-facing context and identity text must display in full and wrap naturally, not be hidden behind `...`. This includes Holding names/codes, company names/codes, branch names/codes, user names, document numbers, business/reference codes, selected workspace labels, report headers, and important badges. Do not use `truncate`, `line-clamp`, fixed-height badges, or CSS `text-overflow: ellipsis` for these values. Dense data tables may keep columns compact only when the full value is also visible in the row expansion/detail or an adjacent wrap area.
- **Data List Selection vs Editing Rule**: Clicking a data-list row selects it and opens read-only detail only. Editing must require an explicit edit action such as the pencil button. Selected/read-only rows must not use amber/orange editing highlights; reserve amber/orange only for the active editing row.
- **Data Store Roles**: MongoDB is the only authoritative operational source for all business CRUD, documents, master data, settings, transactions, and user-entered data. Cloudflare R2/S3 stores all images/files; MongoDB stores only metadata and private file paths.
- **MongoDB Aggregate Embed Rule**: Prefer storing child/subdocument data inside the parent MongoDB document when the child is part of the same business aggregate and is normally loaded, saved, and audited with the parent. Avoid relation-style reads and extra collection fetches for children that belong to the parent workflow; design documents for fast user-facing reads. Keep separate collections only for independent master data, very large/unbounded child lists, cross-aggregate references, or data that has its own lifecycle/permissions.
- **Immutable GUID Identity Rule**: Every persisted business record must get an immutable GUID at first creation (`guid_fixed` / `guid`) and that GUID must never change for the lifetime of the record. CRUD detail/update/delete IDs, selected row keys, audit links, Kafka event identity, projection identity, and sync identity must use this immutable GUID. User-facing codes, names, document numbers, SKU, barcode, and similar fields exist for users to see, search, print, and communicate; they are not the persisted record identity.
- **Business Reference Code Rule**: Business reference fields between records must store the relevant user/business code, not the target record GUID, when the domain expects code-based references. Example: product or barcode unit references must store `unit_code` (รหัสหน่วยนับ), not the product unit `guid_fixed`. The referenced master record still has its own immutable `guid_fixed`; the relation value is the business code.
- **Business Code Uppercase Rule**: User-facing business/reference/display code values must be normalized to uppercase before validation, duplicate checks, search, save, and downstream sync. This covers fields such as `code`, `<thing>_code`, `business_code`, `business_codes[]`, company code, product code, unit code, document code, SKU, and equivalent business codes. Do not apply this to persisted/API field names (still lower `snake_case`), `holding_code` (must stay lowercase), `lang_code`, email/user login text, or GUID values.
- **Product Unit Access Model**: Product unit master data (`units`) stores company-level availability in `business_codes`. Empty or missing `business_codes` means available to every company in the tenant. Products, barcodes, and business documents reference units by `unit_code`, while CRUD identity and sync identity stay on immutable `guid_fixed`.
- **Product Stock, Cost, Marketplace Stock, and Dimension Price Rule**: Accounting stock belongs to product-level stock records and must support product total balance, warehouse-level balance, and storage-location-level balance. Costing must calculate from accounting stock only: normally one cost per product; when the explicit warehouse-cost option `cost_by_warehouse` is enabled, calculate cost per warehouse first and aggregate back to product cost. Do not calculate inventory cost by marketplace dimensions such as color or size. Marketplace stock is a separate availability projection that may expose product-level available balance and product-dimension available balance such as red, XL, black 128GB, SIM package, network, lot, or serial dimensions. Selling prices may be defined at product, barcode/SKU, price-level, marketplace, and detailed dimension level; create/use a separate dimension price table/model when the normal product/barcode price array cannot represent marketplace variants clearly. Barcode/SKU rows are sellable identifiers for scanning, units, prices, marketplace mapping, images, and lookup only; they must not own authoritative accounting `qty`, `balance_qty`, average cost, or stock ledger balance. Existing barcode stock/balance fields are legacy/projection/display compatibility fields only and must not be used as the source for stock deduction or costing.
- **Product Category Group Semantics**: Product category groups (`group_number` 1-20) are usage/device/channel groups, not flat menu-type categories. Example: group 1 can be ordering/tablet, group 2 cashier/POS, group 3 kitchen/KDS, group 4 delivery. Each group must contain a complete category tree using `parent_guid` and `parentguidall`; root nodes describe the usage group, children/subchildren describe menu/business categories, and `codelist` should be attached to selling leaf categories or explicitly sellable category nodes. Do not seed or design product categories as one flat root category per menu type unless Jead explicitly asks for that.
- **Iron MongoDB Source Rule**: Unless Jead gives an explicit special instruction for a projection, rebuild, sync, relational calculation, migration, or BI/reporting flow, every operational create/read/update/delete/list/detail workflow must read and write MongoDB first. User-facing screens, APIs, tools, and agents must not use PostgreSQL or ClickHouse as the source for companies, branches, settings, products, users, permissions, approvals, transactions, or other operational data.
- **Operational CRUD Pipeline Iron Rule**: Every operational CRUD write must flow `MongoDB -> Kafka -> PostgreSQL -> ClickHouse`. CRUD handlers write MongoDB first, then publish a Kafka event or durable MongoDB outbox event. PostgreSQL consumers build rebuildable relational projections from those events, and ClickHouse consumes processed facts/projections for BI. CRUD handlers must not write PostgreSQL or ClickHouse directly unless the task is explicitly a rebuild/sync/projection/BI worker.
- **Frontend CRUD Mutation Rule**: Next.js/frontend create, edit, and delete actions must call MongoDB-backed operational APIs only. Frontend CRUD screens must not write PostgreSQL or ClickHouse, must not call projection/BI endpoints as mutation sources, and must refresh normal list/detail state from MongoDB-backed APIs after a mutation. Kafka/projection fan-out is backend responsibility.
- **Rebuildable Projection Stores**: PostgreSQL is a relational processing/projection engine for postings, balances, tax/VAT, AR/AP, GL, strict relational calculations, and integration-ready relational outputs. ClickHouse is the BI/analytics/reporting store fed from processed facts/projections. PostgreSQL and ClickHouse are rebuild/sync targets derived from MongoDB/system events for other systems to consume; they are not operational truth. If PostgreSQL or ClickHouse conflicts with MongoDB, MongoDB wins and the sync/rebuild pipeline must be fixed.
- **Compact Data Lists**: หน้าจอรายการข้อมูล (data list/table) ทุกหน้าจอต้องใช้ `text-xs` เป็น font size มาตรฐานเดียวกันทั้ง header และ row, ใช้ padding น้อยที่สุด (`px-2 py-1`) เพื่อแสดงข้อมูลได้มากที่สุดในพื้นที่จำกัด ห้ามใช้ font size ต่างกันระหว่าง column หรือระหว่าง header กับ row และเวลาอยู่ในสถานะแก้ไข (Editing) พื้นหลังของแถวนั้นต้องเปลี่ยนเป็นสีส้มไฮไลท์ (`bg-amber-100/70` / `dark:bg-amber-950/40`) พร้อมใช้ปุ่มไอคอนจัดการที่เป็นปุ่มขอบมีสีตามโทนการทำงาน (แก้ไขเป็นสีฟ้า primary, ลบเป็นสีแดง destructive)
- **Holding Access Scope Rule**: Users, screen permissions, permission groups, user permission assignments, and approval rights are Holding-owned records under `holding_code`. Their applicability must be configured with explicit scope entries such as `access_scopes[]`, `scope_rules[]`, or `approval_rules[]` using `scope_type` = `holding | company | branch`. `scope_type=holding` applies to the whole Holding; `scope_type=company` requires uppercase `business_code` and applies to all branches in that company; `scope_type=branch` requires uppercase `business_code` plus zero-padded 5-digit `branch_code`. Use the shared Holding scope widget: a whole-Holding checkbox, searchable company add-to-list, selected-company branch panel, branch search add-to-list, and per-company `all_branches` checkbox. Do not expose raw JSON.
- **Company-Level Master Data Access Selectors**: General master-data fields such as `business_codes` are company-level only. Do not show, edit, or save branch selections inside these normal company access editors. Branch-specific permissions belong only to the Holding access scope model above, or to explicit branch/legal/calendar/operation screens. Read legacy `company_guids` only as a compatibility alias.
- **Workspace Company & Branch Selection**: เมื่อเลือกบริษัทในหน้าเลือก Workspace ระบบต้องทำการตรวจสอบ หากยังไม่มีข้อมูลสาขาในระบบเลย ให้ทำการเพิ่มสาขาสำนักงานใหญ่ (`00000`) ให้อัตโนมัติ และต้องบังคับส่งผู้ใช้งานเข้าสู่ขั้นตอนเลือกสาขาเสมอ เพื่อให้ผู้ใช้งานคลิกเลือกสาขาด้วยตนเอง แม้จะมีเพียงสาขาเดียวก็ตาม ห้ามข้ามขั้นตอนนี้อัตโนมัติ
- **Semantic Page Backgrounds**: Every major frontend page must use a page-specific, meaningful background image that reflects that screen's business purpose. Use optimized `.webp` assets only. The image style must be photorealistic, premium-camera quality, physically plausible in lighting/perspective/scale/shadows/materials, include presentable professional Thai people when the page context can naturally include people, and have clear dimensional depth/depth of field. Support light and dark themes without reducing text contrast or dense work-surface usability. New page background images must be generated by Codex only; other agents must not generate image assets themselves and should request/use Codex-created WebP assets.



## Stack (EDIT TO MATCH YOUR PROJECT)
- Language: Go, TypeScript
- Framework: Next.js, Go API
- Stores: MongoDB operational data, Cloudflare R2/S3 images/files, PostgreSQL relational projections, ClickHouse BI
- UI: shadcn/ui, Tailwind CSS
- Test: go test, npm run typecheck

## ⚙️ Execution guidance (all models)
- **Big/architectural change** → escalate reasoning depth to high + split into subtasks.
- **Bash output** → re-read literally, check exit code (not just last line), watch syntax edge cases.
- **Novel logic** (non-memorized) → ground with an example or source before generating.
- **Multimodal input** (screenshots/diagrams/schemas) → use as primary reference when provided.
- **Tool use** → delegate non-trivial work to MCP/tools and batch parallel calls; don't reason inline.
- **Plugin selection** → choose suitable available plugins/tools automatically based on the task; ask only before installing new plugins/connectors or when a tool action is R0/risky.
- **Capability gaps** → if a model lacks a feature (e.g. computer use, image segmentation), state it and fall back per availability — never fake the result.

## Forbidden
- ❌ Chain-of-thought verbose ("Let me think step by step...")
- ❌ Modifying temperature/top_p/top_k
- ❌ `SELECT *` for 2-3 cols
- ❌ String-concat SQL
- ❌ Hardcoded secrets
- ❌ `git pull` / `fetch`+`reset --hard` / `checkout origin/…` that overwrites newer local code (local = source of truth; remote may be older → R0, ask first)
- ❌ Reading *.gen.*, /vendor, /dist, /node_modules
- ❌ Generic AI UI: stock card grid + serif heading + accent bar
