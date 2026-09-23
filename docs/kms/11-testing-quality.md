# การทดสอบและคุณภาพโค้ด (Testing & Quality)
> ตรวจล่าสุด: 2026-09-23 (หลังถอด MongoDB/Kafka/Redis/ClickHouse — ADR `decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md`)

## 1. ภาพรวม 30 วินาที
- Backend Go เป็น Pure Go (`CGO_ENABLED=0`) ต่อ PostgreSQL ตัวเดียว — test file 225 ไฟล์ (`git ls-files '*_test.go'` ใน `backend/`), 21 ไฟล์ติด tag `//go:build integration`
- **ไม่มี CI อัตโนมัติ** (`.github/workflows/ci.yml` ลบ 2026-09-09 — ADR `decisions/2026-09-09-github-storage-only.md`) ตัวตรวจจริงคือ `tools/verify.sh` ที่คนสั่งเอง (§5)
- Frontend: vitest 95 ไฟล์ใน `frontend/src`, Playwright e2e 10 spec ใน `frontend/e2e/`, UAT ที่ root `tests/` (§6)
- กฎ UAT (`AGENTS.md`) = CRUD ครบ + ตรวจ PostgreSQL ทีละ step + ลบด้วย id/code เท่านั้น + seeded random

## 2. Go tests — วิธีรัน
### 2.1 unit (ไม่ต้องมีฐานข้อมูล)
```bash
sh tools/verify.sh backend
```
รันใน `golang:1.26` (Docker): compile-all ทั้ง tag ปกติและ `integration` แล้ว `go test -short` ทุก package (`tools/verify.sh:128`) — Windows รัน `go build ./... && go vet ./... && go test ./...` ตรง ๆ ได้แล้วเพราะไม่มี CGO (`librdkafka` ถูกถอด)

### 2.2 integration (PostgreSQL จริง)
test ที่ต้องต่อฐานจะ `t.Skip()` เมื่อไม่มี env:
| env | ใช้กับ | หมายเหตุ |
|---|---|---|
| `BC_GL_TEST_POSTGRES_DSN` | `internal/generalledger` (+ `httpapi`), `internal/fixedasset`, `internal/organization/rolepermission`, `pkg/microservice` | ฐานเปล่าก็ได้ — test สร้าง schema เอง (`generalledger.EnsureSchema`) |
| `GL_AUTH_TEST_DSN` | test auth/สิทธิ์ที่ต่อ `bcai_projection` | ฐานเปล่าก็ได้ |
| `BC_TAXFORM_TEST_POSTGRES_DSN` | `internal/goapi/handlers/tax_form_integration_test.go` (แบบยื่นภาษี: บันทึกฉบับ, prefill จาก GL, เครดิต ภ.ง.ด.50/51) | ฐานเปล่าก็ได้ — test สร้างข้อมูลเองแล้วลบตาม company code (IT01–IT03); `verify.sh postgres` ตั้งค่านี้ให้ (แยกจาก `BC_TAX_TEST_POSTGRES_DSN` 2026-09-23 เพราะตัวนั้นต้องใช้ฐานที่ seed แล้ว) |
| `BC_TAX_TEST_POSTGRES_DSN` | `internal/goapi/handlers/tax_withholding_integration_test.go` | ต้องเป็นฐาน holding ที่ seed ใบสำคัญ WHT ด้วย `backend/cmd/glseed` แล้ว — `verify.sh` ไม่ตั้งค่านี้จึงข้าม |

### 2.3 gofmt / vet
| เครื่องมือ | บังคับที่ไหน | สถานะ |
|---|---|---|
| `go vet ./...` | ทำด้วยมือก่อน commit งาน backend | ผ่าน 2026-09-23 |
| `gofmt -l .` | ทำด้วยมือ — `verify.sh` ไม่เช็ค | ค้าง: `generalledger/models.go`, `statement_template_test.go`, `httpapi/error_contract_test.go` |
| กับดัก EOL | ไฟล์เก่าบางไฟล์เป็น CRLF ห้าม `gofmt -w` ทั้งไฟล์โดยไม่ดู diff | — |

## 3. Quarantine
`backend/.ci/test-quarantine.txt` ว่างแล้ว (2026-09-23) — package legacy ที่เคยถูกกักถูกลบไปพร้อมโค้ด Mongo/Kafka; `verify.sh backend` รันทุก package ที่เหลือ ถ้าจะกักใหม่ต้องเขียนเหตุผลในไฟล์นั้น

## 4. Integration suite — `tools/verify.sh postgres`
เปิด `postgres:18-alpine` ชั่วคราวชื่อ `bc-pg-ci` (trust auth) → รัน `go test -tags=integration -count=1 ./pkg/... ./internal/...` ใน `golang:1.26` ด้วย `--network container:bc-pg-ci` ตั้ง `BC_GL_TEST_POSTGRES_DSN` + `GL_AUTH_TEST_DSN` + `BC_TAXFORM_TEST_POSTGRES_DSN` → ลบ container ทิ้งหลังจบ (`tools/verify.sh:154`)
- ไม่มี MongoDB replica set / Kafka broker / `projection.compose.yml` อีกแล้ว (ลบ 2026-09-23)
- รันจริง 2026-09-23 **ผ่าน** (~50 วินาที)

## 5. การตรวจอัตโนมัติ — `tools/verify.sh` (แทน GitHub Actions ที่ถูกลบ 2026-09-09)
| target | ทำอะไร | สถานะจริง | อ้างอิง |
|---|---|---|---|
| `codemap` | ตรวจ `docs/reference/CODE-MAP.md` ตรงกับซอร์ส | ผ่าน 2026-09-23 | `tools/verify.sh:61` |
| `frontend` / `frontend-build` | `npm run lint` → `npm run typecheck` → vitest แล้ว `npm run build` แยก target | ผ่าน 2026-09-23 (710 test) | `tools/verify.sh:87,119` |
| `backend` | compile-all (ปกติ + `integration`) + `go test -short` ทุก package | ผ่าน 2026-09-23 | `tools/verify.sh:128` |
| `postgres` | integration กับ PostgreSQL 18 ชั่วคราว (§4) | ผ่าน 2026-09-23 | `tools/verify.sh:154` |

`npm run verify` = codemap + frontend (ก่อน push ทุกครั้ง), `npm run verify:all` = ทุก target (ก่อน deploy)

trigger: **ไม่มี** — ไม่มีอะไรรันให้อัตโนมัติหลัง push; hook อัตโนมัติมีแค่ `.githooks/pre-commit` (CODE-MAP + คำต้องห้าม) และ `.githooks/pre-push` (codemap + frontend เมื่อแตะ `frontend/`) — ต้อง `npm run hooks:install` ครั้งหนึ่งต่อ clone

## 6. Frontend tests
### 6.1 vitest (unit, 95 ไฟล์)
- config `frontend/vitest.config.ts:3-12`: environment `node`, include `src/**/*.test.ts`, alias `@` → `src`
- รูปแบบหลัก = ทดสอบ Next route handler โดย stub `fetch` ด้วย `vi.fn` แล้วตรวจ header/URL ที่ส่งไป backend เช่น `frontend/src/app/api/product/[[...productPath]]/route.test.ts:1-14`; ไฟล์ security เช่น `frontend/src/app/login-screen.security.test.ts`, `frontend/src/app/menu/main-menu-password.security.test.ts`
- รัน: `cd frontend && npm test` (`frontend/package.json:15` script `test` = `vitest run`)

### 6.2 Playwright e2e ใน `frontend/e2e` (10 spec)
- config `frontend/playwright.config.ts:6-17`: testDir `./e2e`, timeout 60s, ไม่ parallel, retries 0, baseURL `E2E_BASE_URL` (default `http://localhost:3000`), optional `E2E_BROWSER_CHANNEL`, trace `retain-on-failure`; **ไม่จัดการ server เอง** ต้องเปิด `build && start` ก่อน (comment `:3-5`)
- login ผ่านปุ่ม Dev ในหน้าแรก; `frontend/e2e/dev-login.spec.ts:3` ตรวจ `/api/auth/dev-login` โดยตรง
- ครอบคลุม GL (ผังบัญชี error ไทย, รายละเอียดใบสำคัญ, review, drill รายงาน), เมนู (Champ upgrade/consistency/tree), MCP token, ที่อยู่สาขา — spec CRUD สินค้า/บาร์โค้ด/คลัง/ธนาคาร/คู่ค้า และ GL UAT ที่ตรวจผ่าน `mongosh` ถูกลบ 2026-09-23 พร้อมการถอด MongoDB (ADR `docs/kms/decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md`)
- UAT ที่ root `tests/` ตรวจ PostgreSQL ผ่าน `tests/support/pg.ts` (`docker exec <PG_CONTAINER> psql`, ค่าเริ่มต้น `postgres`/`bcai_projection`) ตามกฎ UAT ตรวจทีละ step
- รัน: `cd frontend && npm run test:e2e` (`frontend/package.json:16` script `test:e2e` = `playwright test`)

### 6.3 Playwright UAT ที่ root `tests/` (12 spec + `auth.setup.ts`)
- config `playwright.config.ts:16-53`: testDir `./tests`, baseURL `PW_BASE_URL` (default `http://127.0.0.1:3000`), reporter `html`, project `setup` รัน `tests/auth.setup.ts` ครั้งเดียวแล้ว project `chromium` ใช้ `storageState` `.auth/user.json` (`:11,49-52`; `.auth/` อยู่ใน `.gitignore:79`)
- `tests/auth.setup.ts:12-33`: กด `Dev Login` → เลือก holding `บ้านเชียง` → สาขา `สาขาทดสอบไทย|TST03` → `สำนักงานใหญ่|00001` แล้ว assert `localStorage.bc_workspace.shop.holdingcode === 'bc001'` — **ผูกกับข้อมูล holding `bc001` ที่ต้องมีอยู่ใน Mongo local**
- spec หลัก `tests/uat-crud.spec.ts` (653 บรรทัด): seed จาก `CRUD_SEED` หรือ `Date.now()%1e9` (`:16`), PRNG `mulberry32` (`:25`), เก็บ seed ลง `metrics` (`:34`) และเขียน `test-results/uat-crud/metrics-crud.json` (`:649`); helper `memberCount()` (`:107`) นับแถว `holding_members` ผ่าน `tests/support/pg.ts` + `mongoCount()` (`:112-114`); ลำดับ CRUD-01 Holding → 02 Company → 03 Branch → 04 User → 05 PermissionGroup → 06 read-only screens → 99 report (`:168,305,372,438,581,632,645`) รันแบบ serial retries 2 (`:162`); cleanup ลบเฉพาะ `shopusers` ที่ตรง `holdingcode`+`username` ที่สร้างเอง (`:455,570`)
- spec อื่น (รายชื่อจาก `git ls-files 'tests/*'`): `tests/employee-*.spec.ts` (photo/save-stress/scope-save/uat), `tests/login-*.spec.ts` (clear-buttons/dev-uat/header-controls), `tests/logout-expired.spec.ts`, `tests/menu-session-feedback.spec.ts`, `tests/uat.spec.ts`, `tests/example.spec.ts` (ยังไม่ตรวจเนื้อหาทีละไฟล์)
- รัน: `npx playwright test tests/uat-crud.spec.ts` หรือ script root `npm run test:headed|test:ui|test:debug` (`package.json:10-12`) — ต้องมี container PostgreSQL (`PG_CONTAINER`, ค่าเริ่มต้น `postgres`) ให้ `tests/support/pg.ts` ใช้ `docker exec psql` ได้

## 7. UAT หน้าจอด้วยปุ่ม Demo
- AI ทดสอบเองผ่านปุ่ม "ทดลองใช้ระบบ (Demo)" (Browser pane หรือ Playwright) → holding `rungrueng` → บริษัท `01` → สาขา `00000` แล้วตรวจ payload/console/screenshot (กฎใน `AGENTS.md`)
- harness `scratch/stagehand-uat` (ขับด้วย DeepSeek) **ปลดระวาง 2026-09-23** — ห้ามใช้ AI ภายนอก ใช้ Playwright/Browser pane แทน

## 8. กฎ UAT (AGENTS.md) เทียบกับสิ่งที่ suite ทำได้จริง
| ข้อกำหนด | `tests/*.spec.ts` (root) | `frontend/e2e/*` |
|---|---|---|
| CRUD ครบ + edge case | ทำ (`uat-crud.spec.ts`, `employee-*.spec.ts`) | ทำเฉพาะจอที่ยังมี backend (GL, องค์กร, ตั้งค่า) |
| ตรวจ PostgreSQL ทีละ step | ทำผ่าน `tests/support/pg.ts` (`docker exec <PG_CONTAINER> psql`) | ตรวจผ่าน API เป็นหลัก |
| cleanup ด้วย id/code เท่านั้น | ทำ | ลบ record ที่ตัวเองสร้าง |
| seeded random | `CRUD_SEED` ใน `uat-crud.spec.ts` | ยังใช้ `Date.now()` uid |

## 9. ช่องว่าง / สิ่งที่ยังไม่ตรวจ
- ภาษีซื้อ/ภาษีขาย/ภ.พ.30 อ่านจาก `details.vats` ของใบสำคัญ GL — integration test `TestVatReportsReadRecordedVat` + `TestPostgresVatRecordsForPeriod` (tag `integration`, ต้องมี `BC_GL_TEST_POSTGRES_DSN`); ภ.พ.36 ยังไม่มีข้อมูลต้นทาง (`bugs/2026-09-23-vat-report-reads-missing-erp-tables.md`)
- `TestTaxWithholdingReportFromGL` รันได้เฉพาะเมื่อตั้ง `BC_TAX_TEST_POSTGRES_DSN` เป็นฐานที่ seed แล้ว — `verify.sh` ข้าม
- `tests/employee-scope-save.spec.ts:114` มี type error เดิม (ไม่กระทบ `npm run verify` เพราะ root `tests/` ไม่อยู่ใน typecheck ของ frontend)
- gofmt ค้าง 3 ไฟล์ (§2.3)
