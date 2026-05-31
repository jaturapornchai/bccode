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
4. **ONE concern per turn** — don't bundle unrelated changes.
5. **Output diff, not full file** when editing existing code.
6. **Schema-first outputs** (JSON schema) — more token-efficient than prose.
7. **Clear thought preservation** if turn is simple Q&A (saves multi-turn input growth).
8. **Push means whole project** — when Jead says `push to github`, `push to GitHub`, or equivalent without an explicit narrower scope, stage all project changes (`git add -A` from repo root), run the narrow verification/secret checks that fit the change, commit, and push the current branch to GitHub. Do not limit the push to only the current task's files unless Jead explicitly says so.
9. **Auto-push meaningful changes** — after any moderate, multi-file, rule/schema/model, backend, frontend workflow, or otherwise important change, automatically stage the whole project, run the appropriate narrow verification and secret checks, commit, and push the current branch to GitHub. Do not wait for a separate push request unless the change is trivial/read-only or an R0/divergence/secret blocker requires asking first.
10. **Agent Fast Execution Contract** — Codex, Claude Code, Gemini/Antigravity, and other agents must prefer targeted evidence over broad slow checks. Docs/rules-only changes need only `git diff --check` plus secret scan. Frontend code changes rely on `next dev` HMR — run `npm run typecheck` only before commit/summary, not every edit. Backend changes: write correct code, run touched-package tests only if needed, and do NOT rebuild Docker until Jead says `rebuild`/`deploy`. Do not run `go test ./...`, broad browser automation, or full-repo scans unless Jead explicitly asks or targeted checks are insufficient. Announce long commands, update every ~30 seconds, and after ~2 minutes report whether to continue, narrow, or stop.

## 💾 Caching strategy (all models with prompt caching)
- Stable content (system prompt, project context, source, docs) → place at TOP so it can be cached.
- Dynamic question/task → place at BOTTOM.
- Keep the stable prefix unchanged across turns to maximize cache hits (saves most input cost).

## Quality (NO MAGIC)
1. Never invent file path, API, lib — grep/glob first.
2. Never claim "done" without test/build output pasted.
3. Unsure → "ต้อง verify: <specific>" — don't guess.
4. R0 (drop db, force-push, `git pull`/`reset --hard` over local, deploy, real $) → STOP + ask.
5. **Backend rebuild on request ONLY**: Do NOT auto-rebuild Docker when editing `backend/`. Backend is batch mode — write correct code, accumulate changes, and rebuild+verify `/healthz` ONLY when Jead says `rebuild`/`deploy`: `cd backend; docker-compose up -d --no-deps --build mainapi`. (See core-rules "Dev Workflow Mode".)
6. **Rule & Skill & Database Model Upgrade**: If a change, fix, or debug finding alters system patterns, database models, or logic covered by rules/skills/database schemas, update the rules/skills/models (`d:\bccode\.agents\rules\bc-account-core-rules.md`, `d:\bccode\.agents\skills\`, or `d:\bccode\AGENTS.md`) immediately so the AI grows smarter over time.


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
- **Data Store Roles**: MongoDB is the authoritative operational source for all business CRUD, documents, master data, and user-entered data. Cloudflare R2/S3 stores all images/files; MongoDB stores only metadata and private file paths. PostgreSQL is a relational processing/projection engine for postings, balances, tax/VAT, AR/AP, GL, and strict relational calculations. ClickHouse is the BI/analytics/reporting store fed from processed facts. Do not use PostgreSQL or ClickHouse as the direct CRUD source of truth.
- **Default Data Source Rule**: Unless Jead gives an explicit special instruction for a projection, relational calculation, migration, or BI/reporting flow, every operational create/read/update/delete/list/detail workflow must read and write MongoDB first. PostgreSQL and ClickHouse may only consume or expose projections derived from MongoDB/system processing; they must not become the primary user-facing source for business master data such as companies, branches, settings, products, users, or transactions.
- **Compact Data Lists**: หน้าจอรายการข้อมูล (data list/table) ทุกหน้าจอต้องใช้ `text-xs` เป็น font size มาตรฐานเดียวกันทั้ง header และ row, ใช้ padding น้อยที่สุด (`px-2 py-1`) เพื่อแสดงข้อมูลได้มากที่สุดในพื้นที่จำกัด ห้ามใช้ font size ต่างกันระหว่าง column หรือระหว่าง header กับ row และเวลาอยู่ในสถานะแก้ไข (Editing) พื้นหลังของแถวนั้นต้องเปลี่ยนเป็นสีส้มไฮไลท์ (`bg-amber-100/70` / `dark:bg-amber-950/40`) พร้อมใช้ปุ่มไอคอนจัดการที่เป็นปุ่มขอบมีสีตามโทนการทำงาน (แก้ไขเป็นสีฟ้า primary, ลบเป็นสีแดง destructive)
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
