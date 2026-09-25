# เครื่องมือบรรทัดคำสั่งและสคริปต์ (backend/cmd, scripts, tools, scratch, prompts)
> ตรวจล่าสุด: 2026-09-25 — ผู้เขียน: AI reader; ทุกข้อเท็จจริงอ้าง path:line ตรวจกับซอร์สจริงในรอบนี้
> MongoDB, Kafka, Redis, ClickHouse, `backend/cmd/*` microservice เก่า, `backend/cluster/` (compose/k8s) และ Dockerfile/Makefile ยุคเก่าถูกถอดออกเมื่อ 2026-09-23 — ADR [`decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md`](decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md) — **ห้ามเพิ่มกลับ**; สคริปต์ที่ยังพึ่งของเหล่านี้ถูกระบุเป็น **DEAD** ด้านล่าง

## 1. ภาพรวม: binary จริงมีตัวเดียว

- **binary ที่ deploy จริงคือ root `backend/main.go`** (192 บรรทัด) — `backend/Dockerfile:22` build `main.go` เป็น `go-app`, `backend/docker-compose.yml:21-65` ใช้ `Dockerfile` นี้เป็น service `mainapi`, overlay local ใช้ `Dockerfile.local` (`backend/docker-compose.local.yml:90-106`, build `main.go` ที่ `backend/Dockerfile.local:18`)
- รันเป็น process เดียว (HTTP API + goapi ใต้ `/goapi/*`) ไม่มีการแยกโหมด — env `DEV_API_MODE` ยังถูกตั้ง (`backend/Dockerfile:42`, `backend/docker-compose.yml:39`, `deploy/account/provision-server.sh:85`) แต่ไม่มีโค้ดอ่าน (รายละเอียด `10-infra-deploy.md` §1, §6); prod compose มี service backend ตัวเดียวคือ `mainapi`
- **`backend/cmd/` เหลือ dir เดียวคือ `glseed/`** (`ls backend/cmd/`) — ไม่มี Dockerfile ใต้ `backend/cmd/` และ `backend/Makefile` ไม่มีแล้ว
- `backend/.github/workflows/` ยังมี 9 ไฟล์ค้างจาก repo ยุคเก่า (แตะล่าสุด 2026-05-22 commit `fbc6729f`) — GitHub ไม่รันเพราะไม่ได้อยู่ที่ root `.github/`; 5 ไฟล์อ้าง Dockerfile ที่ไม่มีแล้ว (`build_api_member.yaml:35`, `build_deploy_api_member_dev.yaml:35` → `Dockerfile-member`; `build_consumer.yaml:35`, `build_deploy_consumer_dev.yaml:35` → `Dockerfile-consumer`; `build_migration.yaml:34` → `Dockerfile-migration`) และไฟล์ `build_deploy_*_dev.yaml` push image ไป `ghcr.io/smlsoft/*` แล้ว `git push origin main` (เช่น `build_deploy_api_dev.yaml:54-58`) — AGENTS.md ห้ามย้ายขึ้น root

## 2. `backend/cmd/*`

| module | หน้าที่ | invoked by | สถานะ | อ้างอิง |
|---|---|---|---|---|
| `glseed` | เติมข้อมูลหลักฐานลูกหนี้/เจ้าหนี้/ธนาคาร/งบประมาณ/การเชื่อมบัญชี/กลุ่มบัญชีสินค้า/ใบสำคัญที่ผ่านบัญชีแล้ว ให้ครบทุกจอของ GL — หาเลขบัญชีจากคุณสมบัติในผังบัญชีจริงของบริษัท (ไม่ยึดรหัสบัญชี เพราะแต่ละบริษัทมีผังต่างกัน) เขียนผ่าน service layer เดียวกับ backend (`store.Execute`) ไม่ยิง SQL ตรง — dry-run เป็นค่าเริ่มต้น | `GLSEED_DSN='postgres://...' go run ./cmd/glseed -company 01 -fiscal 2569 [-apply]` (flag เต็ม: `-branch`, `-holding`, `-actor`, `-acc-<บทบาท>`, `-wht-new-code`, `-wht-new-parent`) | LIVE (PostgreSQL) | `backend/cmd/glseed/main.go:1-13, 72-86` (619 บรรทัด) |

`tools/verify.sh backend` compile + test `./cmd/...` ด้วย จึงครอบคลุม `glseed` (`tools/verify.sh:143-146`)

## 3. `scripts/**` (root) — seed/UAT ผ่าน frontend origin

| ไฟล์ | หน้าที่ | invoked by | สถานะ | อ้างอิง |
|---|---|---|---|---|
| `scripts/seed-demo.mjs` + `scripts/demo-data.json` | seed holding `demo` (3 บริษัท/สาขา/แผนก/ชุดสิทธิ์/พนักงาน/กลุ่มสินค้า/สินค้า/ลูกค้า/ผู้ขาย) สำหรับปุ่ม "ทดลองใช้ระบบ (Demo)"; idempotent (409 = ข้าม); ยิงผ่าน `${SEED_BASE}/backend` (Next rewrite → mainapi) | `node scripts/seed-demo.mjs` / `SEED_BASE=https://... node ...` | **ใช้ได้บางส่วน** — ขั้น holding/บริษัท/สาขา/ประเภทธุรกิจ/ชุดสิทธิ์/พนักงาน/สิทธิ์ผู้ใช้ยังมี route; ขั้นแผนก (`/organization/department`), กลุ่มสินค้า/บาร์โค้ด (`/product/*`) และลูกหนี้/เจ้าหนี้ (`/debtaccount/*`) ไม่มี route ใน backend แล้ว (ถอด 2026-09-23) → นับเป็น failed และจบด้วย exit code 2 (`seed-demo.mjs:190`) | `scripts/seed-demo.mjs:1-17, 94, 160-183`; `.agents/skills/ui-scale-polish/references/case-studies-and-gotchas.md:648` |
| `scripts/seed-gl-tax-rungrueng-2569.mjs` + `.data.mjs` | ข้อมูลบัญชี+ภาษี ก.ค.–ก.ย. 2569 ของ `rungrueng/01/00000`: ใบรายวัน 31 ใบ (docno `-01NN`) พร้อมคู่ค้า/VAT/หัก ณ ที่จ่าย แล้วผ่านรายการ + แบบยื่น 7 ฉบับ (ภ.พ.30/ภ.ง.ด.3/53 ก.ค.–ส.ค., ภ.ง.ด.2 ส.ค.); dry-run เป็นค่าเริ่มต้น, idempotent ตาม docno/แบบ+งวด, เลือกสมุดตาม booktype ครั้งเดียวแล้วจำใน manifest `tmp/sample-data/`; `verify` อ่านรายงาน VAT/WHT + ดาวน์โหลดไฟล์ยื่นด้วยสื่อ/PDF แบบ/50 ทวิ | `node scripts/seed-gl-tax-rungrueng-2569.mjs [--apply]` / `... verify` (env `SEED_BASE`, `SEED_MANIFEST`, `SEED_MEDIA_REF_NO`) | LIVE (ใช้ seed prod 2026-09-24; dry-run local+prod และ verify prod ผ่าน 2026-09-25) | `scripts/seed-gl-tax-rungrueng-2569.mjs:1-9,89-110`; `docs/kms/22-sample-data-gl-tax-2569.md`; provenance `docs/examples/gl-tax-rungrueng-01-2569-20260924.json` |
| `scripts/test-google-login.js` | Playwright headless เปิด prod แล้วรอปุ่ม Google GSI render, เก็บ console log | `node scripts/test-google-login.js [url]` (default `https://account.bcaicloud.com/`) | one-off / ใช้ซ้ำได้ | `scripts/test-google-login.js:1-6` |

## 4. `tools/**` — เครื่องมือช่วย dev

| ไฟล์ | หน้าที่ | invoked by | สถานะ | อ้างอิง |
|---|---|---|---|---|
| `tools/gen-code-map.ps1` | สร้าง `docs/reference/CODE-MAP.md` index ไฟล์ ≥ 1000 บรรทัด (function/component/section-comment); มีโหมด `-Check` ที่ regenerate ในหน่วยความจำแล้วเทียบกับไฟล์ที่ commit ไว้ (ข้ามบรรทัด `> Generated:`) → exit 1 พร้อมรายการแถวที่เลื่อน — ใช้โดย `.githooks/pre-commit` และ `tools/verify.sh codemap`; repo root มาจากตำแหน่งสคริปต์ จึงรันได้ทั้ง Windows และ Linux | `pwsh -NoProfile -File tools/gen-code-map.ps1` / `... -Check` | LIVE | `tools/gen-code-map.ps1:1-12, 22` |
| `tools/verify.sh` | **ตัวแทน GitHub CI ที่ถูกลบ 2026-09-09** — target เดี่ยว 5 ตัว: `codemap`, `frontend`, `frontend-build`, `backend` (compile + unit test ใน container `golang:1.26`), `postgres` (integration tests กับ container `postgres:18-alpine` ชั่วคราว ลบทิ้งหลังจบ — `t_postgres` บรรทัด 154-175) + ชุดรวม `fast` = codemap + frontend, `all` = ทั้ง 5 ตัว (บรรทัด 189-203); `backend/.ci/` เหลือแค่ `test-quarantine.txt` | `npm run verify` / `npm run verify:all` / `sh tools/verify.sh <target>` | LIVE | `tools/verify.sh:34, 128-215`; ADR `docs/kms/decisions/2026-09-09-github-storage-only.md` |
| `tools/install-hooks.mjs` | คัดลอก `.githooks/*` (`pre-commit`, `pre-merge-commit`, `pre-push`) เข้า `.git/hooks/` (ไม่ใช้ `core.hooksPath` เพราะมันทับ hook เดิมทั้งโฟลเดอร์) — ต้องรันซ้ำทุกครั้งที่ `.githooks/` เปลี่ยน | `npm run hooks:install` | LIVE | `tools/install-hooks.mjs:1-11` |
| `.githooks/pre-push` | กันของเสียขึ้น GitHub แทน CI: รัน `tools/verify.sh codemap` เสมอ + `frontend` เมื่อช่วง commit ที่ push แตะ `frontend/`; ถ้าแตะ `backend/` จะเตือนให้รัน `npm run verify:all` เอง (ไม่รันให้เพราะช้า) | อัตโนมัติตอน `git push` (ข้าม: `SKIP_VERIFY=1`) | LIVE (ทดสอบ 2026-09-09: ผ่าน/บล็อก/ลบ branch/branch ใหม่ ครบ) | `.githooks/pre-push:1-18` |
| `tools/kung_test.py`, `kung_loop.py`, `kung_small.py`, `kung_test.sh`, `kung_analyze.sh`, `kung_analyze.go` | test runner ของ chatbot "น้องกุ้ง" ยิง `http://localhost:8888/goapi/api/v1/chatbot/chat-agent-v2-sync` | รันมือ | **DEAD** — ไม่มี route `/api/v1/chatbot/*` ใดใน backend แล้ว (grep `chatbot` ใน `backend/**/*.go` ไม่เจอ; route ชุดสุดท้ายถูกถอดใน commit `550d5489` 2026-09-23) | `tools/kung_test.py:22`; `tools/kung_test.sh:5` |
| `tools/__pycache__/*.pyc` (2 ไฟล์) | ไบต์โค้ด Python ถูก commit ทั้งที่ `.gitignore:49` มี `__pycache__/` | — | DEAD (ควรลบออกจาก index) | `git ls-files tools/__pycache__` |
| `tools/playwright-mcp.config.json` | ตั้ง `chromiumSandbox: false` ให้ Playwright MCP | Playwright MCP (ปิดอยู่ตาม `~/.claude.json` — ยังไม่ตรวจ) | unused ตอนนี้ | `tools/playwright-mcp.config.json:1-7` |
| `tools/smooth-mouse.js` | overlay cursor ให้ demo headed browser ผ่าน `browser_run_code_unsafe` | Playwright MCP | unused ตอนนี้ | `tools/smooth-mouse.js:1-12` |
| `tools/test-mcp-chrome-launch.js` | PoC launch Chrome persistent profile (hard-code `C:/Users/jatur/AppData/Local/bcai-chrome-profile`) | `node tools/test-mcp-chrome-launch.js` | one-off (ผูกเครื่องเดียว) | `tools/test-mcp-chrome-launch.js:1-8` |
| `tools/seed/seed-access-setup-dev.ps1` | seed บริษัท/สาขา/สิทธิ์/กลุ่ม/ผู้ใช้/อนุมัติ จำนวนมาก | `pwsh -File ... -HoldingCode X -Username Y` | **DEAD** — `Get-SelectedToken` หา token จากแคชที่ถูกถอดแล้ว 2026-09-23 (session/token ปัจจุบันอยู่ในตาราง PostgreSQL `cache_entries` — `backend/pkg/microservice/cacher.go:45-53`) สคริปต์จึงหยุดตั้งแต่ขั้นหา token (บรรทัด 149) | `tools/seed/seed-access-setup-dev.ps1:1-15, 19-30, 149` |

## 5. `scratch/**` — โฟลเดอร์ทิ้ง แต่มี 18 ไฟล์ติด git

- `.gitignore:64-65` ประกาศ `/scratch/` เป็น throwaway "may contain live bearer/refresh tokens — never commit" แต่ `git ls-files scratch | wc -l` = 18 (commit ก่อน/ทั้งที่มี rule; แตะล่าสุด 2026-06-06)
- tracked ทั้ง 18 ไฟล์ใช้กับระบบปัจจุบันไม่ได้แล้ว:
  - `add-types.js`, `modify-*.js` (5 ไฟล์), `restore-product-screen.js` = string-replace patch ครั้งเดียวบนจอสินค้า (`scratch/add-types.js:1-3`, `scratch/modify-barcode-search.js:1-3`) → **DEAD** (โค้ดปลายทางเปลี่ยนไปแล้ว)
  - `inspect_companies.go`, `inspect_shops.go`, `inspect_unit.go`, `simulate_list.go`, `test_find.go` = `package main` ที่ import driver ของฐานข้อมูลที่ถูกถอด (`scratch/inspect_companies.go:11-13`) ซึ่งไม่มีใน `backend/go.mod` แล้ว → **DEAD** (compile ไม่ผ่าน)
  - `seed-product-category-dev.ps1`, `seed-product-variant-master-dev.ps1`, `seed-solao-restaurant-dev.ps1` = หา token จากแคชที่ถูกถอดแล้วแบบเดียวกับ `tools/seed/seed-access-setup-dev.ps1` และยิง `/product/*` หรือ `/goapi/atlas/*` ที่ไม่มี route แล้ว → **DEAD**
  - `verify_barcode_first_flow.py`, `verify_product_barcode_relation.py` = E2E ยิง `/product/*` ที่ `127.0.0.1:8888` (route ไม่มีแล้ว) → **DEAD**
  - `aliexpress-product-foundation-reference.json` = ไฟล์ reference schema (ไม่ใช่โค้ด)
- untracked (ถูก ignore ถูกต้อง): `bc_auth.json` (live token — ห้าม commit), `login-*.mjs` (Playwright screenshot/audit หน้า login), `seed/`, `stagehand-uat/` (harness ที่เรียก LLM ภายนอก — ขัดกฎห้ามใช้ AI ภายนอกตั้งแต่ 2026-09-23), `verify-somtam/` (dir ว่าง)
- คำแนะนำ: ไม่มีไฟล์ tracked ตัวใดที่ยังใช้งานได้ — ลบออกจาก index ได้ทั้งชุด (ต้องให้ลุงจืดตัดสิน)

## 6. `backend/prompts/**` — AI prompt / language request assets

| ไฟล์ | เนื้อหา | สถานะ | อ้างอิง |
|---|---|---|---|
| `backend/prompts/api_requests/chatbot_agent.md` | สเปก endpoint `POST /goapi/api/v1/chatbot/chat-agent` (ReAct + tools) | **STALE** — ไม่มี route `/api/v1/chatbot/*` ใดใน backend แล้ว (ดูแถว `tools/kung_*` ใน §4) | `chatbot_agent.md:1-8` |
| `backend/prompts/language_requests/README.md` | กติกา: จด key ภาษาที่ขาดไว้ก่อน แล้วค่อย batch เข้า `backend/assets/language/languages.tsv` | LIVE (process) | `README.md:1-5` |
| `language_requests/access-control-menu-rename.md` | rename เมนู "การเข้าถึง" → 5 key ใหม่; หัวเรื่องระบุ `STATUS: KEYS LANDED IN TSV` | เสร็จแล้ว (เก็บเป็นประวัติ) | `access-control-menu-rename.md:1` |
| `language_requests/user-self-permission-cannot-edit.md` | key `self_permission_cannot_edit` สำหรับ system-settings-screen | ยังค้าง — `grep -c self_permission_cannot_edit backend/assets/language/languages.tsv` = 0 (ตรวจ 2026-09-25) | `user-self-permission-cannot-edit.md:1-8` |

## 7. ช่องว่าง / สิ่งที่ยังไม่ตรวจ

1. `backend/.github/workflows/*` 9 ไฟล์ค้างจาก repo ยุคเก่า (§1) — ควรลบทั้งชุด (R1, รอลุงจืดตัดสิน)
2. สคริปต์ที่ยังไม่ถูกบันทึกในบทความนี้: `scripts/seed-gl-local.mjs`, `scripts/seed-gl-rungrueng-2569.py`, `scripts/generate_thai_chart_of_accounts.mjs`, `tools/fast-deploy.py` (มีใน `10-infra-deploy.md` §9–§10), `tools/ai-link.mjs`, `tools/audit-auth-fetch.mjs`, `tools/i18n-sweep.mjs`, `tools/probe-endpoints.mjs`, `tools/seed-demo.mjs`, `tools/rdform/`, `tools/restart-docker.bat` — ยังไม่ได้เปิดตรวจ
3. `scripts/seed-demo.mjs` — ขั้นที่ยิง route ที่ถูกถอด (§3) ควรตัดออกหรือรอสร้าง API บน PostgreSQL ใหม่ (รอลุงจืดตัดสิน)
4. `tools/seed/seed-access-setup-dev.ps1` — ต้องเปลี่ยนวิธีได้ token (เช่น ผ่านปุ่ม Demo login) หรือลบทิ้ง
5. `chatbot_agent.md` + `tools/kung_*` — route ไม่มีแล้ว ควรลบพร้อมกันไหม (รอลุงจืดตัดสิน)
6. 18 ไฟล์ใน `scratch/` และ `tools/__pycache__/*.pyc` ที่ติด git — ขออนุญาต `git rm --cached` (§4–§5)
7. `tools/playwright-mcp.config.json`, `smooth-mouse.js`: MCP playwright ถูกปิดใน `~/.claude.json` ตาม memory แต่ยังไม่ได้ตรวจการใช้งานจริง
