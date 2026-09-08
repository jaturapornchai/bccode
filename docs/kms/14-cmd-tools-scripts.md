# เครื่องมือบรรทัดคำสั่งและสคริปต์ (backend/cmd, scripts, tools, scratch, prompts, cluster)
> ตรวจล่าสุด: 2026-09-07 (commit d93a210d) — ผู้เขียน: AI reader; ทุกข้อเท็จจริงอ้าง path:line · ตรวจซ้ำโดย fact-checker

## 1. ภาพรวม: binary จริงมีตัวเดียว ที่เหลือคือ legacy หรือ one-off

- **binary ที่ deploy จริงคือ root `backend/main.go`** (749 บรรทัด) — `backend/Dockerfile:21` build `main.go` เป็น `go-app`, `backend/docker-compose.yml:60-62` ใช้ `Dockerfile` นี้เป็น service `mainapi`, overlay local ใช้ `Dockerfile.local` (`backend/docker-compose.local.yml:104-106`, build `main.go` ที่ `backend/Dockerfile.local:18`)
- โหมดทำงานคุมด้วย `DEV_API_MODE` (`backend/main.go:238`): `""`/`"2"` = HTTP + goapi ใต้ `/goapi` (`backend/main.go:256`, `:568-575`), `"3"` = migration อย่างเดียว (`backend/main.go:585`), `""`/`"1"` = Kafka consumers (`backend/main.go:654`). prod compose ใช้ image เดียวกัน 3 บทบาท: `migrate` (`deploy/account/compose.yml:208`, `DEV_API_MODE: "3"` บรรทัด 212), `mainapi` (`:234`, ไม่ตั้ง DEV_API_MODE → HTTP+consumers), `worker` (`:271`, `DEV_API_MODE: "1"` บรรทัด 275); local compose ตั้ง `"2"` (`backend/docker-compose.yml:80`)
- **ไม่มี Dockerfile/compose/CI ที่ build `backend/cmd/*` ตัวใดเข้า deploy ปัจจุบัน** — CI root มีแค่ `.github/workflows/ci.yml` (ไฟล์เดียวใน `.github/workflows/`) ซึ่ง *compile* `./cmd/...` ด้วย `go test -run "^$"` (`.github/workflows/ci.yml:36-39`) แต่ไม่ build image; Actions ถูก billing-lock (ตาม brief ของ lead) การ compile ล่าสุดที่บันทึกไว้คือ `docs/handoff/HANDOFF-RISKS-2026-09-05.md:108`
- Dockerfile ชุดเก่าใน `backend/cmd/*/Dockerfile` มี 8 ไฟล์: 6 ไฟล์ใช้ `golang:1.17-alpine3.15` (เช่น `backend/cmd/app/Dockerfile:6`), ยกเว้น `backend/cmd/journalconsumerdelete/Dockerfile:6` = `golang:1.18.0-alpine3.15` และ `backend/cmd/mediauploadservice/Dockerfile:6` = `golang:1.20.2-alpine3.17` และถูกอ้างจาก `backend/Makefile` เท่านั้น (`backend/Makefile:5-44`); target บางตัวชี้ dir ที่ไม่มีแล้ว: `cmd/inventoryservice` (`:22,25`), `cmd/inventoryimportservice` (`:29`), `cmd/uploadmediaservice` (`:122`) — ตรวจแล้ว 3 dir นี้ไม่มีใน repo
- มี workflow ซ้อนอยู่ที่ `backend/.github/workflows/` 9 ไฟล์ (build_api_member/consumer/migration ฯลฯ, แตะล่าสุด 2026-05-22 commit fbc6729f) อ้าง `Dockerfile-member`/`Dockerfile-consumer`/`Dockerfile-migration` (`backend/.github/workflows/build_api_member.yaml:35`, `build_consumer.yaml:35`, `build_migration.yaml:34`) — อยู่นอก root `.github/` จึงไม่ถูก GitHub รัน (พฤติกรรม GitHub ทั่วไป; ยังไม่ตรวจว่าเคย fork/ย้ายมาอย่างไร)

## 2. บัญชี `backend/cmd/*` (30 dir)

สถานะ: **LIVE** = ถูก build/รันใน pipeline ปัจจุบัน, **one-off** = เครื่องมือรันมือ (ยังคอมไพล์ได้), **LEGACY** = microservice ยุค smlcloudplatform ที่ root `main.go` แทนที่แล้ว, **DEAD** = ไม่มี logic/ชี้ของที่ไม่มี. คอลัมน์ "แตะล่าสุด" มาจาก `git log -1 -- backend/cmd/<dir>`.

| module | หน้าที่ | entry / invoked by | สถานะ | แตะล่าสุด | อ้างอิง |
|---|---|---|---|---|---|
| `goapi` | รัน goapi แบบ standalone (ไม่มี prefix `/goapi`), PORT default 8888, CORS จาก `CORS_ALLOWED_ORIGINS` | `backend/Dockerfile.goapi:15` build `./cmd/goapi/` (healthcheck `/version` `:42`) — **ไม่มี compose/CI/deploy ใดอ้าง Dockerfile.goapi** (grep ทั้ง repo) | LEGACY-ready (build ได้ แต่ไม่ได้ใช้) | 2026-08-23 | `backend/cmd/goapi/main.go:21-27,36,61-64,69-72` |
| `app` | monolith รุ่นก่อน root main.go: register HTTP ทุก module (product/transaction/vfgl/restaurant…) | `backend/cmd/app/Dockerfile:20`; `backend/cmd/app/Makefile:1-2` (`go run main.go` + JAEGER env) — ไม่มี compose อ้าง | LEGACY | 2026-09-02 | `backend/cmd/app/main.go:9-120` |
| `authenticationservice` | HTTP auth อย่างเดียว (login/dev-login/demo-login/googlelogin/refresh) | `backend/cmd/authenticationservice/Dockerfile:24`; `backend/Makefile:6,9`; cluster compose image `smlsoft/smlcloudplatform:authen-1.0` (`backend/cluster/docker-compose.yml:5`) | LEGACY | 2026-09-02 | `backend/cmd/authenticationservice/main.go:10-37` |
| `ws` | auth HTTP + `journal.NewJournalWs` (websocket) + Prometheus/Jaeger | `backend/cmd/ws/Makefile:1-2` เท่านั้น | LEGACY | 2026-09-02 | `backend/cmd/ws/main.go:9-47` |
| `shopservice` | HTTP shop module | `backend/cmd/shopservice/Dockerfile:20`; `backend/Makefile:18` | LEGACY | 2026-08-23 | `backend/cmd/shopservice/main.go:8-33` |
| `masterdataservice` | HTTP productcategory + member | `backend/cmd/masterdataservice/Dockerfile:20`; `backend/Makefile:33` | LEGACY | 2026-08-23 | `backend/cmd/masterdataservice/main.go:9-30` |
| `memberservice` | HTTP member | `backend/cmd/memberservice/Dockerfile:20`; `backend/Makefile:44` | LEGACY | 2026-08-23 | `backend/cmd/memberservice/main.go:8-24` |
| `member` | LINE member API (`linemember:` token prefix, public `/holding/*`,`/shop/*`,`/member/line`) | `backend/Dockerfile-member:20`; nested workflow `backend/.github/workflows/build_api_member.yaml:35` | LEGACY | 2026-08-23 | `backend/cmd/member/main.go:9-39` |
| `imageuploadservice` | HTTP images (file persister local) | `backend/cmd/imageuploadservice/Dockerfile:20`; `backend/Makefile:37,41` | LEGACY | 2026-08-23 | `backend/cmd/imageuploadservice/main.go:8-33` |
| `mediauploadservice` | HTTP media upload | `backend/cmd/mediauploadservice/Dockerfile:20` | LEGACY | 2026-08-23 | `backend/cmd/mediauploadservice/main.go:8-22` |
| `transactionservice` | HTTP purchase module อย่างเดียว | ไม่มี Dockerfile/Makefile อ้าง | LEGACY | 2026-08-23 | `backend/cmd/transactionservice/main.go:8-24` |
| `sync` | `migration.StartMigrateModel` แล้วเปิด HTTP เปล่า ๆ | ไม่มีใครอ้าง | LEGACY | 2026-08-23 | `backend/cmd/sync/main.go:8-25` |
| `migrate` | เรียก `migration.StartMigrateModel(ms, cfg)` แล้วจบ (ไม่เช็ค error return) | ไม่มีใครอ้าง; prod ใช้ `DEV_API_MODE=3` ของ root แทน | LEGACY | 2026-03-08 | `backend/cmd/migrate/main.go:7-16` |
| `migrationapi` | HTTP `POST /migrationtools/{journalimport,shopimport,chartimport}` | ไม่มีใครอ้าง | LEGACY | 2026-06-04 | `backend/cmd/migrationapi/main.go:7-20`, `api/migrationapi.go:38-41` |
| `transaction_consumer` | Kafka consumer PO/PR/RFQ/PurchaseReceive เมื่อ `DEV_API_MODE` = 1/2; โหลด `.env` ด้วย godotenv (comment ระบุยังไม่ย้ายมา bootstrap.json) | ไม่มี Dockerfile อ้าง (Dockerfile-consumer build root `main.go` `backend/Dockerfile-consumer:20`) | LEGACY | 2026-07-09 | `backend/cmd/transaction_consumer/main.go:15-29,52,68-76` |
| `productbarcodeconsumer` | migrate ตาราง productbarcode + register consumer | ไม่มีใครอ้าง | LEGACY | 2026-03-08 | `backend/cmd/productbarcodeconsumer/main.go:7-22` |
| `cons` | สร้าง microservice แล้ว `ms.Start()`; consumer ทุกตัวถูก comment | ไม่มีใครอ้าง | DEAD | 2026-03-08 | `backend/cmd/cons/main.go:11-24` |
| `journalconsume` | เหมือน `cons` (consumer ถูก comment) | ไม่มีใครอ้าง | DEAD | 2026-03-08 | `backend/cmd/journalconsume/main.go:6-18` |
| `journalconsumerdelete` | อ่าน `CONSUMER_GROUP_NAME` แล้ว `ms.Start()`; consumer ถูก comment | `backend/cmd/journalconsumerdelete/Dockerfile:20` (ไม่มี compose อ้าง) | DEAD | 2026-03-08 | `backend/cmd/journalconsumerdelete/main.go:7-23` |
| `inventorysearchconsumer` | `NewMicroservice` + `ms.Start()` ไม่มี logic | ไม่มีใครอ้าง | DEAD | 2026-03-08 | `backend/cmd/inventorysearchconsumer/main.go:6-15` |
| `glservice` | ไม่มี `main.go` — มีแค่ `postgl.http` (REST client ยิง `POST http://localhost:8080/journal`) และ `chartofaccount.json` ตัวอย่าง | รันมือด้วย REST client | DEAD (fixture) | 2026-03-08 | `backend/cmd/glservice/postgl.http:4` |
| `test` | ทดลอง `JournalPgRepository` กับ `mock.NewPersisterPostgresqlConfig()` และ JSON journal hard-code | รันมือ | DEAD (scratch) | 2026-06-06 | `backend/cmd/test/main.go:12-23` |
| `dbtest` | ทดสอบต่อ PostgreSQL (db = holdingcode hard-code) + ClickHouse จาก `bootstrap.json` | `go run ./cmd/dbtest` (อ่าน `./bootstrap.json` หรือ `../bootstrap.json`) | one-off / บางส่วน DEAD (ClickHouse container ถูกถอด 2026-09-06) | 2026-06-06 | `backend/cmd/dbtest/main.go:41-44,58,112-116` |
| `datatransfer` | คัดลอกข้อมูล holding ระหว่าง Mongo 2 ก้อน (`--holdingcode`, `--toholdingcode`, `--confirm`) | `go run cmd/datatransfer/main.go ...` (`backend/cmd/datatransfer/README.md:12`) | one-off | 2026-06-06 | `backend/cmd/datatransfer/main.go:15-17,50-57`; env keys `backend/internal/datatransfer/config.go:13,17,27,34` |
| `removeshopdata` | ลบข้อมูลทั้ง holding (shopuser, employee, barcode, category, …) หลังถาม y/n | `go run` มือ, flag `--holdingcode` (help text ยังเขียนว่า "to transfer") | one-off (ทำลายข้อมูล — R0) | 2026-06-06 | `backend/cmd/removeshopdata/main.go:64,67-76,90-94,99-205` |
| `stable_identity_migration` | backfill `users.uid`, `shopusers.useruid`, `poapprovalsettings/poapprovalstatus.approver_useruid` + สร้าง index; dry-run เว้นแต่ `-apply` | `go run ./cmd/stable_identity_migration -config bootstrap.json [-apply]`; env `MONGODB_URI`/`MONGO_DB_NAME` มาก่อน | one-off (rerun-safe ตาม ROLLBACK.md:3) | 2026-06-06 | `main.go:29-33,51-60,78-79`; `ROLLBACK.md:1-3` |
| `storage_name_migration` | rename collection/table เป็น lowercase no-underscore ใน Mongo/Postgres/ClickHouse (`--target`, `--apply`) | `go run ./cmd/storage_name_migration --target all ...` (`README.md:20-30`) | one-off (ทำไปแล้ว 2026-05-20 ตาม `MIGRATION_HISTORY.md:5`) | 2026-06-06 | `main.go:30-42`; `main_test.go:5,29` |
| `branch_code_audit` | audit `organizationbranches` (code ผิดรูป, ซ้ำ, ไม่มีสำนักงานใหญ่, VAT ไม่มี taxid) ออก JSON; `BRANCH_AUDIT_DISCOVER` = ไล่ทุก db | รันมือ; env `MONGODB_DEV_URI`/`MONGODB_URI`, `MONGODB_DEV_DB` (default `dev`) | one-off (read-only) | 2026-06-06 | `backend/cmd/branch_code_audit/main.go:63-66,88,97-100` |
| `r2_private_smoke` | smoke test S3/MinIO/R2: put → get → ยิง unsigned GET ต้องถูกปฏิเสธ | รันมือ; env `S3_ENDPOINT`/`S3_ACCESS_KEY_ID`/`S3_SECRET_ACCESS_KEY`/`S3_BUCKET_NAME` หรือ legacy `R2_*` + `BC_R2_SMOKE_HOLDING_CODE`; fallback `bootstrap.json`/`config/bootstrap.json` | one-off (มี unit test 5 ตัว) | 2026-09-03 | `backend/cmd/r2_private_smoke/main.go:116-131,225-226`; `main_test.go:8-75` |
| `codex_cleanup_devdata` | **dir ว่างเปล่า** (untracked, ไม่มีไฟล์; `ls -la` ว่าง, ไม่อยู่ใน `git ls-files`) | — | DEAD (ซากโฟลเดอร์) | — | `backend/cmd/codex_cleanup_devdata/` |

หลักฐาน build วันนี้ (native Windows, ไม่ผ่าน kafka CGO): `go vet ./cmd/dbtest ./cmd/branch_code_audit ./cmd/stable_identity_migration ./cmd/r2_private_smoke ./cmd/storage_name_migration` exit 0; `go test ./cmd/storage_name_migration ./cmd/r2_private_smoke` → `ok` ทั้งคู่ (fact-checker รันซ้ำ 2026-09-07 ได้ผลเดียวกัน). module ที่ import `pkg/microservice` (kafka) ยังไม่ compile ซ้ำวันนี้ (ต้องใช้ container ตาม memory `backend-go-docker-test-command`).

ข้อสังเกตจากโค้ด:
- `backend/cmd/stable_identity_migration/ROLLBACK.md:52` สั่ง `go run .\cmd\stableidentitymigration` แต่ dir จริงชื่อ `stable_identity_migration` (ตรวจแล้ว path ใน doc ไม่มีอยู่; ไฟล์ยาว 53 บรรทัด)
- `datatransfer/README.md:5-8` ตรงกับโค้ด: source fallback `MONGODB_UAT_URI/DB`, destination fallback ไปที่ env ปัจจุบัน (`backend/internal/datatransfer/config.go:13-17,27-34`)
- `dbtest` hard-code holdingcode `3E0aX0qsmeRr26TjCk3kRz5vBdv` (`main.go:58`) และอ่าน password จาก `bootstrap.json` — ใช้กับ prod ไม่ได้ตรง ๆ

## 3. `scripts/**` (root) — seed/UAT ผ่าน frontend origin

| ไฟล์ | หน้าที่ | invoked by | สถานะ | อ้างอิง |
|---|---|---|---|---|
| `scripts/seed-demo.mjs` + `scripts/demo-data.json` | seed holding `demo` (3 บริษัท/แผนก/ชุดสิทธิ์/พนักงาน/สินค้า/ลูกค้า/ผู้ขาย) สำหรับปุ่ม "ทดลองใช้ระบบ (Demo)"; idempotent (409 = ข้าม); ยิงผ่าน `${SEED_BASE}/backend` (Next rewrite → mainapi) | `node scripts/seed-demo.mjs` / `SEED_BASE=https://... node ...` | LIVE (ใช้กับ demo login; อ้างใน `docs/skills/ui-scale-polish/references/case-studies-and-gotchas.md:636`, `docs/kms/architecture/product-listing-api-v2-handoff.md:73`) | `scripts/seed-demo.mjs:1-17`; `scripts/demo-data.json:1-5` |
| `scripts/seed-barcodes.mjs` | seed 20 barcode `UATBC01..20` พร้อมรูปจาก URL ภายนอก (z-cdn.chatglm.cn) → `/api/product-barcode/image` → MinIO + thumb; BASE hard-code `127.0.0.1:3000`, seed 20260831 | `node scripts/seed-barcodes.mjs` | one-off (2026-08-31) | `scripts/seed-barcodes.mjs:1-11` |
| `scripts/seed-barcodes-fix.mjs` | รอบสอง: ผูกรูปกลับเข้า barcode 20 ตัว, ตรวจผ่าน `docker exec mongodb mongosh appdb` | `node scripts/seed-barcodes-fix.mjs` | one-off | `scripts/seed-barcodes-fix.mjs:1-18,21` |
| `scripts/test-google-login.js` | Playwright headless เปิด prod แล้วรอปุ่ม Google GSI render, เก็บ console log | `node scripts/test-google-login.js [url]` (default `https://account.bcaicloud.com/`) | one-off / ใช้ซ้ำได้ | `scripts/test-google-login.js:1-6` |

## 4. `tools/**` — เครื่องมือช่วย dev

| ไฟล์ | หน้าที่ | invoked by | สถานะ | อ้างอิง |
|---|---|---|---|---|
| `tools/gen-code-map.ps1` | สร้าง `docs/reference/CODE-MAP.md` index ไฟล์ ≥ 1000 บรรทัด (function/component/section-comment); มีโหมด `-Check` ที่ regenerate ในหน่วยความจำแล้วเทียบกับไฟล์ที่ commit ไว้ (ข้ามบรรทัด `> Generated:`) → exit 1 พร้อมรายการแถวที่เลื่อน — ใช้โดย `.githooks/pre-commit` (ตัวจริงที่ทำงาน) และ CI job `code-map-check` (`.github/workflows/ci.yml`) ที่**ยังรันไม่ได้ — Actions ล็อกเรื่อง billing ตั้งแต่ 2026-09-02**; repo root มาจากตำแหน่งสคริปต์ จึงรันได้ทั้ง Windows และ Linux CI | `pwsh -NoProfile -File tools/gen-code-map.ps1` / `... -Check` | LIVE (docs/reference/CODE-MAP.md:3 ระบุว่า auto-generated จากไฟล์นี้) | `tools/gen-code-map.ps1:1-12` |
| `tools/kung_test.py`, `kung_loop.py`, `kung_small.py`, `kung_test.sh`, `kung_analyze.sh`, `kung_analyze.go` | test runner ของ chatbot "น้องกุ้ง" ยิง `http://localhost:8888/goapi/api/v1/chatbot/chat-agent-v2-sync` | รันมือ | **DEAD** — goapi ปัจจุบัน register แค่ `/api/v1/chatbot/chat-gemini` และ `/analyze-document` (`backend/internal/goapi/bootstrap.go:542-544`); ไม่มี `chat-agent` ใน `.go` ใด (rg --no-ignore) | `tools/kung_test.py:22`; `tools/kung_test.sh:4-5` |
| `tools/__pycache__/*.pyc` (2 ไฟล์) | ไบต์โค้ด Python ถูก commit ทั้งที่ `.gitignore:47` มี `__pycache__/` | — | DEAD (ควรลบออกจาก index) | `git ls-files tools/__pycache__` |
| `tools/playwright-mcp.config.json` | ตั้ง `chromiumSandbox: false` ให้ Playwright MCP | Playwright MCP (ปิดอยู่ตาม `~/.claude.json` — ยังไม่ตรวจ) | unused ตอนนี้ | `tools/playwright-mcp.config.json:1-7` |
| `tools/smooth-mouse.js` | overlay cursor ให้ demo headed browser ผ่าน `browser_run_code_unsafe` | Playwright MCP | unused ตอนนี้ | `tools/smooth-mouse.js:1-12` |
| `tools/test-mcp-chrome-launch.js` | PoC launch Chrome persistent profile (hard-code `C:/Users/jatur/AppData/Local/bcai-chrome-profile`) | `node tools/test-mcp-chrome-launch.js` | one-off (ผูกเครื่องเดียว) | `tools/test-mcp-chrome-launch.js:1-8` |
| `tools/seed/seed-access-setup-dev.ps1` | seed บริษัท/สาขา/สิทธิ์/กลุ่ม/ผู้ใช้/อนุมัติ จำนวนมาก; ดึง token จาก `docker exec redis redis-cli --scan auth-*` | `pwsh -File ... -HoldingCode X -Username Y` | one-off dev (ต้องมี redis container local) | `tools/seed/seed-access-setup-dev.ps1:1-14,18-28` |

## 5. `scratch/**` — โฟลเดอร์ทิ้ง แต่มี 18 ไฟล์ติด git

- `.gitignore:57-58` ประกาศ `/scratch/` เป็น throwaway "may contain live bearer/refresh tokens — never commit" แต่ `git ls-files scratch | wc -l` = 18 (commit ก่อน/ทั้งที่มี rule; แตะล่าสุด 2026-06-06)
- tracked: `add-types.js`, `modify-*.js` (5 ไฟล์), `restore-product-screen.js` = string-replace patch แบบครั้งเดียวบน `frontend/src/app/menu/product-screen.tsx` / `product-barcode-screen.tsx` (`scratch/add-types.js:1-3`, `scratch/modify-barcode-search.js:1-3`) → **DEAD** (โค้ดปลายทางเปลี่ยนไปแล้ว รันซ้ำจะไม่ match); `inspect_companies.go`, `inspect_shops.go`, `inspect_unit.go`, `simulate_list.go`, `test_find.go` = `package main` อ่าน Mongo ตรง (หลาย `package main` ใน dir เดียว build ไม่ได้ ต้อง `go run <ไฟล์>` ทีละตัว); `seed-product-category-dev.ps1`, `seed-product-variant-master-dev.ps1`, `seed-solao-restaurant-dev.ps1` = seed ผ่าน `http://localhost:8888` (default HoldingCode `"xxx"` `scratch/seed-product-category-dev.ps1:2-6`); `verify_barcode_first_flow.py`, `verify_product_barcode_relation.py` = E2E ผ่าน `127.0.0.1:8888`; `aliexpress-product-foundation-reference.json` 25,559 bytes reference schema
- untracked (ถูก ignore ถูกต้อง): `bc_auth.json` (key `token`,`refresh` = live token — ห้าม commit), `login-*.mjs` (Playwright screenshot/audit หน้า login), `seed/`, `stagehand-uat/` (harness ตาม memory), `verify-somtam/` (dir ว่าง — `ls -a` ไม่มีไฟล์)
- คำแนะนำ: ยัง "มีประโยชน์" เฉพาะ `stagehand-uat/` และ `seed-*-dev.ps1` เป็นตัวอย่าง; ที่เหลือลบจาก index ได้ (ต้องให้ลุงจืดตัดสิน — นอกขอบเขตบทความนี้)

## 6. `backend/prompts/**` — AI prompt / language request assets

| ไฟล์ | เนื้อหา | สถานะ | อ้างอิง |
|---|---|---|---|
| `backend/prompts/api_requests/chatbot_agent.md` | สเปก endpoint `POST /goapi/api/v1/chatbot/chat-agent` (ReAct + tools) ให้ frontend ใช้แทน `chat-gemini` | **STALE** — route ไม่มีใน `bootstrap.go:542-544`; เหลือแค่ข้อความใน prompt `backend/internal/goapi/handlers/aichat/query_generator.go:30,78` ที่ยังพูดถึง chat-agent | `chatbot_agent.md:1-8` |
| `backend/prompts/language_requests/README.md` | กติกา: จด key ภาษาที่ขาดไว้ก่อน แล้วค่อย batch เข้า `backend/assets/language/languages.tsv` | LIVE (process) | `README.md:1-5` |
| `language_requests/access-control-menu-rename.md` | rename เมนู "การเข้าถึง" → 5 key ใหม่; หัวเรื่องระบุ `STATUS: KEYS LANDED IN TSV` | เสร็จแล้ว (เก็บเป็นประวัติ) | `access-control-menu-rename.md:1-6` |
| `language_requests/user-self-permission-cannot-edit.md` | key `self_permission_cannot_edit` สำหรับ system-settings-screen | ยังค้าง — `grep -c self_permission_cannot_edit backend/assets/language/languages.tsv` = 0 (key ยังไม่ลง TSV) | `user-self-permission-cannot-edit.md:1-8` |
| (ไม่มีไฟล์) `language_requests/product-barcode-screen.md` | ถูกอ้างจาก `frontend/src/lib/product-barcode/language.ts:9` แต่ **ไม่มีในโฟลเดอร์** (ls มี 3 ไฟล์) | dangling reference | `frontend/src/lib/product-barcode/language.ts:5-9` |

## 7. `backend/cluster/**` — compose/k8s ยุค smlcloudplatform

- `backend/cluster/docker/storage/README.md:1-5` มีแค่คำสั่ง `sudo chown 1001:1001 mongodb_data` (owner สำหรับ bitnami image); `backend/cluster/docker/storage/docker-compose.yml:4-29` = `bitnami/redis:5.0`, `bitnami/mongodb:5.0` (container `mongodbdev`), `postgres:14` + service `backend` (`:39`)
- compose อื่นใน `backend/cluster/docker/services/*` ดึง image `smlsoft/smlcloudplatform:{authen,imageuploadservice,inventory,masterdata,member,shop,swagger}` (`services/*/docker-compose.yml:5`) และ `systemmoniter` = mongo-express/redisinsight/kafka-ui; `backend/cluster/k8s/*.yaml`, `seaweedfs/s3.json` มีอยู่แต่ไม่มีอะไรใน deploy ปัจจุบันอ้าง (แตะล่าสุด 2026-05-25)
- สถานะ: **LEGACY ทั้งชุด** — stack จริงคือ `backend/docker-compose.yml` (local) และ `deploy/account/compose.yml` (prod) ตาม §1

## 8. ช่องว่าง / สิ่งที่ยังไม่ตรวจ

1. **ยังไม่ compile `cmd/*` ที่ import `pkg/microservice`** (app, authenticationservice, ws, shopservice, … transaction_consumer) บนเครื่องนี้วันนี้ — อาศัย `docs/handoff/HANDOFF-RISKS-2026-09-05.md:108` (compile `./cmd/...` ผ่านใน container 2026-09-05) เท่านั้น
2. `backend/.github/workflows/*` 9 ไฟล์: ยังไม่ตรวจว่าเคยรันได้จริงจาก repo ไหน (อยู่นอก root `.github/` ของ repo นี้) — ถ้าไม่ใช้แล้วควรลบเพื่อลดความสับสน
3. `tools/playwright-mcp.config.json`, `smooth-mouse.js`: MCP playwright ถูกปิดใน `~/.claude.json` ตาม memory แต่ยังไม่เปิดไฟล์ตรวจใน session นี้
4. `user-self-permission-cannot-edit.md`: ตรวจแล้ว (fact-checker) — key ยังไม่อยู่ใน `backend/assets/language/languages.tsv` (grep = 0) ดู §6
5. ไม่ได้รัน runtime evidence จาก docker stack สำหรับบทความนี้ (เครื่องมือทั้งหมดเป็น offline/one-off) — `dbtest` และ `storage_name_migration --target clickhouse` จะล้มเหลวแน่หลัง ClickHouse ถูกถอด 2026-09-06 (สรุปจากโค้ด `backend/cmd/dbtest/main.go:112-116`, ยังไม่รันจริง)

คำถามถึงลุงจืด:
- จะเก็บ `backend/cmd/*` กลุ่ม LEGACY/DEAD (cons, journalconsume, journalconsumerdelete, inventorysearchconsumer, glservice, test, app, *service) ไว้ทำไม — ถ้าไม่มีแผน microservice แยก ควรลบเป็น batch เดียว (R1: กระทบ `go test ./cmd/...` ใน CI น้อยลงด้วยซ้ำ)
- `Dockerfile.goapi` (goapi standalone) ตั้งใจให้เป็นทางเลือก deploy หรือเป็นซาก?
- `chatbot_agent.md` + `tools/kung_*`: route chat-agent ถูกถอดจาก `bootstrap.go` ใน commit 57d70c21 (2026-06-25 "refactor(aichat): remove dead-weight agentic chat") ส่วน commit b4cc4140 (2026-06-26, แตะ `backend/prompts` ล่าสุด) ลบแค่ `mcp_guide.md` + `mcp_apikey_permission_preset.md` ไม่ได้ลบ `chatbot_agent.md` → ควรลบ prompt + tools ชุดนี้พร้อมกันไหม
- 18 ไฟล์ใน `scratch/` และ `tools/__pycache__/*.pyc` ที่ติด git — อนุญาตให้ `git rm --cached` ไหม
- `removeshopdata` เป็นเครื่องมือทำลายข้อมูลระดับ holding ไม่มี dry-run — ควรใส่ `--dry-run` หรือย้ายไปอยู่ใต้ goapi cleanup service (`backend/main.go:565` มี `cleanupService.StartCleanupScheduler` อยู่แล้ว)?
