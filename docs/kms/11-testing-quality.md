# การทดสอบและคุณภาพโค้ด (Testing & Quality)
> ตรวจล่าสุด: 2026-09-07 (commit d93a210d) — ผู้เขียน: AI reader; ทุกข้อเท็จจริงอ้าง path:line · ตรวจซ้ำโดย fact-checker

## 1. ภาพรวม 30 วินาที
- Backend Go มี test file 204 ไฟล์ (นับจาก `git ls-files 'backend/**/*_test.go'`) — กระจุกอยู่ที่ `internal/transaction/transactionconsumer` (37) และ `internal/goapi/handlers` (22); ส่วนใหญ่เป็น unit test ที่ mock DB, มี integration test ที่ต้องมี DB/Kafka จริง 8 ตัวซึ่งถูก gate ด้วย build tag `integration` + env var
- **ไม่มี CI อัตโนมัติแล้ว**: `.github/workflows/ci.yml` (5 jobs) ถูกลบ 2026-09-09 ตามมติลุงจืด "GitHub เก็บ code อย่างเดียว" (ก่อนหน้านั้นทุก run ล้มเหลวเพราะบัญชีถูกล็อกจาก billing — annotation ใน `gh run view 34182717816`: "The job was not started because your account is locked due to a billing issue") → ตัวตรวจเดียวที่เหลือคือ `sh tools/verify.sh` ที่ **คนต้องสั่งเอง** ทุกครั้ง
- Frontend มี 3 ชั้น: vitest unit 44 ไฟล์ (`frontend/vitest.config.ts:6` include `src/**/*.test.ts`), Playwright e2e ใน `frontend/e2e/*.spec.ts` 14 ไฟล์ (`frontend/playwright.config.ts:7`), และ Playwright UAT ที่ root `tests/*.spec.ts` 13 ไฟล์ (`playwright.config.ts:17`) ซึ่งเป็นชุดที่ยิง mongosh ตรวจ MongoDB ทีละ step
- กฎ UAT ของลุงจืด (`AGENTS.md:78-87`) = CRUD ครบ + ตรวจ `appdb` ทีละ step + ลบด้วย id เท่านั้น + seeded random ตาม `tests/uat-crud.spec.ts`

## 2. Go unit tests — layout และวิธีรัน
### 2.1 layout
| กลุ่ม | จำนวนไฟล์ | ตัวอย่าง | หมายเหตุ |
|---|---|---|---|
| `internal/transaction/transactionconsumer/**` | 37 | `backend/internal/transaction/transactionconsumer/stockbalance/stock_balance_transaction_consumer_test.go` | legacy Kafka consumer (DEV_API_MODE=1); 12 sub-package อยู่ใน quarantine (§3) |
| `internal/goapi/handlers/**` | 22 | `backend/internal/goapi/handlers/product_v2_*_test.go` (7 ไฟล์: get/permission/readiness/set/tier/validate/whitelist), `backend/internal/goapi/handlers/kafka/projection_reject_test.go:17,42` | ทดสอบ handler v2 + projection reject/advisory-lock cap |
| `internal/product/productbarcode/**` | 11 | `backend/internal/product/productbarcode/services/productbarcode_outbox_integration_test.go:25` | outbox + reconcile |
| `internal/product/product/**` | 8 | `backend/internal/product/product/outbox/outbox_integration_test.go:21` | transactional outbox |
| `internal/product/projection` | 3 | `backend/internal/product/projection/consumer_test.go:34` (`TestConsumeAcknowledgesRejectedMessagesOnly`) | kafka-go consumer ใหม่ |
| `internal/goapi/mykafkaconsumer` | 1 | `backend/internal/goapi/mykafkaconsumer/product_projection_test.go:38,85` | ack เฉพาะงานสำเร็จ / topic scope |
| `pkg/microservice` (12), `pkg/round` (1), `pkg/serializer` (1), `pkg/stockcalculator` (1) | 15 | `backend/pkg/microservice/persister_mongo_test.go:211` | บางตัว skip เมื่อ `SERVERLESS=serverless` |
| root | 1 | `backend/main_security_test.go:5` (`TestValidReloadConfigSecret`) | — |
| `backend/run/` | 2 | `backend/run/mongodb_test.go:32` (`TestMongodbCreateData`), `backend/run/run_test.go:9` (`TestRun`) | `mongodb_test.go` เป็นไฟล์เดียวในระบบที่ใช้ `testing.Short()` (`git ls-files \| xargs grep -l "testing.Short()"`) |

หลักการ gate ที่ใช้จริงในโค้ด:
- env `SERVERLESS=serverless` ทำให้ test ที่ต้องต่อ DB จริง `t.Skip()` เช่น `backend/internal/authentication/repositories/authentication_repository_test.go:31-33`, `backend/internal/member/member_pg_repository_test.go:20,38`, `backend/pkg/microservice/persister_mongo_test.go:211` — `tools/verify.sh` ตั้ง env นี้ทุก target (`tools/verify.sh:96,129,145`, `backend/.ci/projection.compose.yml:59`)
- build tag `//go:build integration` มี 25 ไฟล์ (รายชื่อจาก `git ls-files | xargs grep -l "^//go:build integration"`) — ชุดตรวจคอมไพล์ทั้งหมดด้วย `-run "^$"` (`tools/verify.sh:105`) แต่ไม่รันจริง เพราะต้องมี Kafka/Mongo/PG ครบ
- integration test ตัวใหม่ skip ตัวเองเมื่อไม่มี env: `BC_OUTBOX_TEST_MONGODB_URI` (`backend/internal/product/product/outbox/outbox_integration_test.go:22-24`), `BC_BARCODE_TEST_POSTGRES_DSN` (`backend/internal/goapi/handlers/kafka/inventory_batch_integration_test.go:18-20`), `BC_PROJECTION_TEST_KAFKA` (`backend/internal/product/projection/consumer_kafka_integration_test.go:39`)

### 2.2 วิธีรันบนเครื่อง dev (Windows) — ต้องผ่าน Docker เท่านั้น
Windows ไม่มี gcc → `confluent-kafka` (CGO) build ไม่ผ่าน (`docs/handoff/HANDOFF-2026-09-06.md:32`) คำสั่งที่ใช้จริง (`docs/handoff/HANDOFF-2026-09-06.md:34`):
```bash
MSYS_NO_PATHCONV=1 docker run --rm -v D:/bccode/backend:/src -w /src -e GOFLAGS=-mod=mod golang:1.26 \
  sh -c "apt-get update -qq >/dev/null 2>&1 && apt-get install -y -qq librdkafka-dev >/dev/null 2>&1; \
         gofmt -l . ; go build ./... && go vet ./internal/goapi/... ./internal/product/... && go test ./internal/goapi/... ./internal/product/..."
```
- image `golang:1.26` ตรงกับ `go 1.26` ใน `backend/go.mod:3`; ต้องติดตั้ง `librdkafka-dev` ก่อนเพราะ CGO ของ confluent-kafka
- ถ้าจะรัน full suite ให้ใช้ `sh tools/verify.sh backend` (ตั้ง `-e SERVERLESS=serverless` และตัด quarantine ให้แล้วที่ `tools/verify.sh:102-104`; ใช้ `tr -d "\r"` เพราะไฟล์ quarantine อาจเป็น CRLF เมื่อ mount จาก Windows — `docs/handoff/HANDOFF-RISKS-2026-09-05.md:101`)
- `backend/Makefile` (125 บรรทัด) มี target test อยู่ 2 ตัวแต่ผูกกับเครื่อง macOS M1 ของทีมเดิม: `run_test_m1` (`backend/Makefile:103-104` — `PKG_CONFIG_PATH=/opt/homebrew/...` + `POSTGRES_HOST=192.168.2.209` รัน `go test --tags dynamic` ไฟล์ journalreport ไฟล์เดียว) และ `run_m1_test_all` (`backend/Makefile:109-110` — `go test --tags dynamic ./...`) — ใช้บน Windows/Docker ไม่ได้ตรงๆ; ที่เหลือเป็น swagger + docker build/push; **ไม่มี target lint/vet/gofmt** และไม่มี `.golangci.yml` ทั้งใน `backend/` และ root (ตรวจด้วย `ls backend/.golangci* .golangci*` → ไม่พบ)

### 2.3 gofmt / vet — ความคาดหวังที่บังคับใช้จริง
| เครื่องมือ | บังคับที่ไหน | สถานะ | อ้างอิง |
|---|---|---|---|
| `gofmt -l .` | เฉพาะคำสั่ง local ใน handoff; `tools/verify.sh` ไม่เช็ค | ทำด้วยมือ | `docs/handoff/HANDOFF-2026-09-06.md:34`; `tools/verify.sh` ไม่มีคำว่า gofmt (ตรวจ `grep -n gofmt tools/verify.sh` → ว่าง) |
| `go vet` | เฉพาะคำสั่ง local, scope `./internal/goapi/... ./internal/product/...` | ทำด้วยมือ | `docs/handoff/HANDOFF-2026-09-06.md:34`, `docs/kms/architecture/product-listing-api-v2-handoff.md:54` |
| กับดัก EOL | ไฟล์เก่าบางไฟล์เป็น CRLF (`product.go`, `bootstrap.go`) ห้าม `gofmt -w` ทั้งไฟล์ | — | `docs/kms/architecture/product-listing-api-v2-handoff.md:62` |
| compile-all | `go test -run "^$" ./cmd/... ./pkg/... ./internal/...` ทั้ง tag ปกติและ `integration` | อยู่ใน `tools/verify.sh backend` (ต้องสั่งเอง ไม่มีอะไรรันให้) | `tools/verify.sh:102,105` |

## 3. Quarantine — 15 package ที่คอมไพล์แต่ไม่รัน
ไฟล์ `backend/.ci/test-quarantine.txt:1-2` ระบุเหตุผล: "legacy business-contract assertions that conflict with current code… Remove a package only after its contract is confirmed and tests pass" `tools/verify.sh backend` อ่านไฟล์นี้ไปกรอง `go list` ออก (`tools/verify.sh:103`) แล้วพิมพ์รายชื่อพร้อมคำเตือนว่า "ผ่านชุดนี้ไม่ได้แปลว่า business contract ของแพ็กเกจข้างบนถูกต้อง" ลง stdout ก่อนรัน (`tools/verify.sh:90-94`) — คำเตือนนี้ย้ายมาจาก step summary ของ workflow เดิมที่ถูกลบ

| # | package (module `smlcloudplatform`) | บรรทัด |
|---|---|---|
| 1 | `internal/transaction/paymentdetail/usecase` | `backend/.ci/test-quarantine.txt:3` |
| 2 | `internal/transaction/transactionconsumer/appurchasereceive` | `:4` |
| 3 | `internal/transaction/transactionconsumer/purchase` | `:5` |
| 4 | `internal/transaction/transactionconsumer/purchaseorder` | `:6` |
| 5 | `internal/transaction/transactionconsumer/purchasereceive` | `:7` |
| 6 | `internal/transaction/transactionconsumer/purchasereturn` | `:8` |
| 7 | `internal/transaction/transactionconsumer/saleinvoice` | `:9` |
| 8 | `internal/transaction/transactionconsumer/saleinvoicereturn` | `:10` |
| 9 | `internal/transaction/transactionconsumer/stockadjustment` | `:11` |
| 10 | `internal/transaction/transactionconsumer/stockpickupproduct` | `:12` |
| 11 | `internal/transaction/transactionconsumer/stockreceiveproduct` | `:13` |
| 12 | `internal/transaction/transactionconsumer/stockreturnproduct` | `:14` |
| 13 | `internal/transaction/transactionconsumer/stocktransfer` | `:15` |
| 14 | `internal/vfgl/accountperiodmaster/services` | `:16` |
| 15 | `internal/vfgl/journal/services` | `:17` |

ทั้ง 15 ตัวเป็น legacy transaction consumer / paymentdetail / GL (vfgl) — ยังไม่ตรวจว่าทุกตัวถูกโหลดเฉพาะ DEV_API_MODE=1; คำถามค้างถึงลุงจืดคือ business contract ของแต่ละตัวคืออะไรจึงจะปลดได้ (`docs/handoff/HANDOFF-2026-09-06.md:71`)

## 4. Integration suite — `backend/.ci/projection.compose.yml`
| service | image | หน้าที่ | อ้างอิง |
|---|---|---|---|
| `mongo` | `mongo:7` + `--replSet rs0` (healthcheck ping) | replica set สำหรับ Mongo transaction ของ outbox | `backend/.ci/projection.compose.yml:3-10` |
| `mongo-init` | `mongo:7` one-shot `rs.initiate` แบบ idempotent (จับ `AlreadyInitialized`) | init replica set | `:11-19` |
| `postgres` | `postgres:18-alpine`, `POSTGRES_HOST_AUTH_METHOD=trust` | projection target | `:20-28` |
| `kafka` | `apache/kafka:4.3.1` KRaft node เดียว, `KAFKA_AUTO_CREATE_TOPICS_ENABLE=false`, heap 256–512m | broker จริงสำหรับ rebalance/fence test | `:29-51` |
| `tests` | `${BC_PROJECTION_GO_IMAGE:-golang:1.26}` profile `test`, mount `..:/src`, env `SERVERLESS`, `BC_OUTBOX_TEST_MONGODB_URI`, `BC_BARCODE_TEST_POSTGRES_DSN`, `BC_PROJECTION_TEST_KAFKA` | รัน `go test -tags=integration … -run '^Test(Product(Outbox\|ServiceOutbox)Integration\|Projection(Kafka\|Rebalance)Integration\|Barcode(Batch\|ServiceOutbox)Integration\|LegacyBarcode(Reconcile\|PrimarySource)Integration)$'` บน 6 package (`:66-68`) | `:52-75` |

8 test ที่ regex ครอบคลุม (ทั้งหมดมี `//go:build integration`):
`TestProductOutboxIntegration` (`backend/internal/product/product/outbox/outbox_integration_test.go:21`), `TestProductServiceOutboxIntegration` (`backend/internal/product/product/services/product_outbox_integration_test.go:24`), `TestProjectionKafkaIntegration` (`backend/internal/goapi/handlers/kafka/projection_kafka_integration_test.go:50`), `TestProjectionRebalanceIntegration` (`backend/internal/product/projection/consumer_kafka_integration_test.go:38`), `TestBarcodeBatchIntegration` (`backend/internal/goapi/handlers/kafka/inventory_batch_integration_test.go:17`), `TestBarcodeServiceOutboxIntegration` (`backend/internal/product/productbarcode/services/productbarcode_outbox_integration_test.go:25`), `TestLegacyBarcodeReconcileIntegration` (`backend/internal/product/productbarcode/repositories/productbarcode_reconcile_integration_test.go:24`), `TestLegacyBarcodePrimarySourceIntegration` (`backend/internal/product/productbarcode/services/productbarcode_projection_source_integration_test.go:21`)

วิธีรัน local (จาก `backend/internal/product/product/outbox/README.md:78-82`, ตรงกับที่ `tools/verify.sh:172` รันให้แล้ว):
```bash
docker compose -p bc-projection-check -f backend/.ci/projection.compose.yml up -d mongo mongo-init postgres kafka
docker compose -p bc-projection-check -f backend/.ci/projection.compose.yml up -d --wait mongo postgres kafka
docker compose -p bc-projection-check -f backend/.ci/projection.compose.yml run --rm --no-deps tests
docker compose -p bc-projection-check -f backend/.ci/projection.compose.yml down -v --remove-orphans
```
ทำไมต้อง `--no-deps`: `docker compose run` (v5.3) จะ restart one-shot dependency (`mongo-init`) ที่ exit ไปแล้ว ทำให้ gate `service_completed_successfully` ล้ม (`docs/handoff/HANDOFF-2026-09-06.md:66`, comment ใน `backend/.ci/projection.compose.yml:17-18`) — จึงต้อง `up -d` ก่อนแล้วค่อย `run --no-deps`; stack นี้ไม่เปิด host port และไม่แตะ volume ของ stack local (`backend/.ci/projection.compose.yml:1`) ผลรอบล่าสุดที่บันทึกไว้: ผ่านทั้ง 8 บน Kafka 4.3.1 + Mongo 7 + PG 18 (`docs/handoff/HANDOFF-RISKS-2026-09-05.md:47`)

## 5. การตรวจอัตโนมัติ — `tools/verify.sh` (แทน GitHub Actions ที่ถูกลบ 2026-09-09)
| target | ทำอะไร | สถานะจริง | อ้างอิง |
|---|---|---|---|
| `backend` | พิมพ์ quarantine → `docker run golang:1.26`: compile-all, `go test -short` เฉพาะ package นอก quarantine, compile tag integration | ใช้ได้ แต่รันเมื่อสั่งเท่านั้น | `tools/verify.sh:87-109` |
| `frontend` / `frontend-build` | `npm run lint` → `npm run typecheck` → `npm test -- --run` (vitest) แล้ว `npm run build` แยก target; env `BCAI_LOCAL_BACKEND_URL` | ใช้ได้ — **ตรวจ 2026-09-09 แล้ว lint ไม่ผ่าน 2 error** (`dashboard-home.tsx:194` hooks ถูกเรียกแบบมีเงื่อนไข, `manage-shortcuts-screen.tsx:133` `Date.now()` ระหว่าง render) | `tools/verify.sh:64-84`; scripts ใน `frontend/package.json:12-15` |
| `outbox` | mongo:7 replica set แบบ `docker run` + postgres:17-alpine → รัน `TestProduct(Outbox\|ServiceOutbox)Integration` และ `TestBarcodeBatchIntegration` เก็บ JSON ผลลัพธ์ | ใช้ได้ แต่ยังไม่ได้รันหลังย้าย | `tools/verify.sh:111-156` |
| `projection` | ใช้ `projection.compose.yml` รัน 5 test ที่เหลือ (`TestProjection(Kafka\|Rebalance)`, `TestBarcodeServiceOutbox`, `TestLegacyBarcode(Reconcile\|PrimarySource)`) เก็บ `backend/projection-test-results.json` | ใช้ได้ แต่ยังไม่ได้รันหลังย้าย | `tools/verify.sh:158-182` |

trigger: **ไม่มี** — ไม่มีอะไรรันให้อัตโนมัติอีกแล้ว ต้องสั่ง `npm run verify` (เร็ว: codemap + frontend) หรือ `npm run verify:all` เอง ก่อน push/deploy; hook เดียวที่ยังทำงานอัตโนมัติคือ `.githooks/pre-commit` ซึ่งตรวจแค่ `docs/reference/CODE-MAP.md` **ข้อสังเกต:** ชุดตรวจนี้ไม่รัน `gofmt`/`go vet`/`golangci-lint` เลย และ `frontend/e2e` + `tests/` (Playwright) ไม่อยู่ใน target ใด — เป็นชุดที่รันมือเท่านั้น (เหมือนตอนเป็น CI) กู้ workflow เดิม: `git show 1799b069:.github/workflows/ci.yml`

## 6. Frontend tests
### 6.1 vitest (unit, 44 ไฟล์)
- config `frontend/vitest.config.ts:3-12`: environment `node`, include `src/**/*.test.ts`, alias `@` → `src`
- รูปแบบหลัก = ทดสอบ Next route handler โดย stub `fetch` ด้วย `vi.fn` แล้วตรวจ header/URL ที่ส่งไป backend เช่น `frontend/src/app/api/product/[[...productPath]]/route.test.ts:1-14`; ไฟล์ security เช่น `frontend/src/app/login-screen.security.test.ts`, `frontend/src/app/menu/main-menu-password.security.test.ts`
- รัน: `cd frontend && npm test` (`frontend/package.json:15` script `test` = `vitest run`)

### 6.2 Playwright e2e ใน `frontend/e2e` (14 spec)
- config `frontend/playwright.config.ts:6-17`: testDir `./e2e`, timeout 60s, ไม่ parallel, retries 0, baseURL `E2E_BASE_URL` (default `http://localhost:3000`), optional `E2E_BROWSER_CHANNEL`, trace `retain-on-failure`; **ไม่จัดการ server เอง** ต้องเปิด `build && start` ก่อน (comment `:3-5`)
- login ผ่านปุ่ม Dev ในหน้าแรก (`frontend/e2e/product-crud.spec.ts:129-136` ปุ่ม `เข้าทดสอบระบบ`); `frontend/e2e/dev-login.spec.ts:3` ตรวจ `/api/auth/dev-login` โดยตรง
- ครอบคลุม CRUD สินค้า/บาร์โค้ด/BOM/product set/warehouse/product group/category/brand/trade partners/branch geo/job-costing channels (ชื่อ test เช่น `frontend/e2e/product-crud.spec.ts:272`, `frontend/e2e/trade-partners-crud.spec.ts:232`, `frontend/e2e/product-barcode-crud.spec.ts:220`)
- **ข้อจำกัด:** spec ในโฟลเดอร์นี้ตรวจผ่าน UI + API เท่านั้น ไม่มีการเรียก `mongosh` (ตรวจ `grep -rn mongosh frontend/e2e` → ว่าง) — แม้ชื่อ test ที่ `frontend/e2e/product-crud.spec.ts:272` จะเขียนว่า "verify API+DB" แต่ "DB" หมายถึงอ่านกลับผ่าน API (comment `:7-8` ระบุว่าเป้าหมายคือ DEV MongoDB Atlas ตาม `backend/bootstrap.json` ไม่ใช่ container `mongodb` local); การยืนยัน Mongo ของ brand ทำแยกด้วย skill ครั้งเดียวเมื่อ 2026-07-01 (`frontend/e2e/master-brand-crud.spec.ts:5-7`) → ไม่ตรงข้อ 2 ของกฎ UAT (`AGENTS.md:83`)
- รัน: `cd frontend && npm run test:e2e` (`frontend/package.json:16` script `test:e2e` = `playwright test`)

### 6.3 Playwright UAT ที่ root `tests/` (13 spec + `auth.setup.ts`)
- config `playwright.config.ts:16-53`: testDir `./tests`, baseURL `PW_BASE_URL` (default `http://127.0.0.1:3000`), reporter `html`, project `setup` รัน `tests/auth.setup.ts` ครั้งเดียวแล้ว project `chromium` ใช้ `storageState` `.auth/user.json` (`:11,49-52`; `.auth/` อยู่ใน `.gitignore:79`)
- `tests/auth.setup.ts:12-33`: กด `Dev Login` → เลือก holding `บ้านเชียง` → สาขา `สาขาทดสอบไทย|TST03` → `สำนักงานใหญ่|00001` แล้ว assert `localStorage.bc_workspace.shop.holdingcode === 'bc001'` — **ผูกกับข้อมูล holding `bc001` ที่ต้องมีอยู่ใน Mongo local**
- spec หลัก `tests/uat-crud.spec.ts` (653 บรรทัด): seed จาก `CRUD_SEED` หรือ `Date.now()%1e9` (`:16`), PRNG `mulberry32` (`:25`), เก็บ seed ลง `metrics` (`:34`) และเขียน `test-results/uat-crud/metrics-crud.json` (`:649`); helper `mongoEval()` ยิง `docker exec mongodb mongosh --quiet appdb --eval` (`:106-111`) + `mongoCount()` (`:112-114`); ลำดับ CRUD-01 Holding → 02 Company → 03 Branch → 04 User → 05 PermissionGroup → 06 read-only screens → 99 report (`:168,305,372,438,581,632,645`) รันแบบ serial retries 2 (`:162`); cleanup ลบเฉพาะ `shopusers` ที่ตรง `holdingcode`+`username` ที่สร้างเอง (`:455,570`)
- spec อื่น (รายชื่อจาก `git ls-files 'tests/*'`): `tests/employee-*.spec.ts` (photo/save-stress/scope-save/uat), `tests/currency-uat.spec.ts`, `tests/login-*.spec.ts` (clear-buttons/dev-uat/header-controls), `tests/logout-expired.spec.ts`, `tests/menu-session-feedback.spec.ts`, `tests/uat.spec.ts`, `tests/example.spec.ts` (ยังไม่ตรวจเนื้อหาทีละไฟล์)
- รัน: `npx playwright test tests/uat-crud.spec.ts` หรือ script root `npm run test:headed|test:ui|test:debug` (`package.json:10-12`) — ต้องมี container ชื่อ `mongodb` (stack local ตอนนี้: `mongodb mongo:7`, `postgres postgres:18-alpine`, `kafka confluentinc/confluent-local`, `mainapi` healthy จาก `docker ps`; ไม่มี container clickhouse ตามที่พักไว้ 2026-09-06)

## 7. Harness `scratch/stagehand-uat` (AI-driven browser UAT)
- ไม่อยู่ใน git (`.gitignore:58` `/scratch/`) — เป็นเครื่องมือ local เท่านั้น; deps `@browserbasehq/stagehand ^3.6.0`, `playwright ^1.61.1`, `zod ^4.4.3` (`scratch/stagehand-uat/package.json`)
- script เช่น `scratch/stagehand-uat/uat-crud-mongo-brand.js:1-12` สร้าง `Stagehand({env:"LOCAL"})` กับ `AISdkClient` ของ DeepSeek (`createDeepSeek(...)("deepseek-chat")`) แล้ว drive หน้า `/masterbrandscreen` ทำ Create→Read→Update→Delete จับ network request จริง; ไฟล์อื่นครอบ warehouse, dimension, product group, channel mappings, variant API
- **ข้อควรระวัง:** ไฟล์ brand **ไม่ได้ฝัง API key ในโค้ด** — บรรทัด 6 อ่าน key จากไฟล์ใต้ `~/.claude/` ด้วย `fs.readFileSync` (ไม่ระบุชื่อไฟล์ที่นี่) แล้วส่งให้ `createDeepSeek({ apiKey })` (`:7`); ยังคงเป็น secret นอก env var จึงควรระวังตอน copy script; และไม่มีไฟล์ใดใน harness นี้เรียก `mongosh` (ตรวจ `grep -l mongosh scratch/stagehand-uat/*.js` → ว่าง) แม้ชื่อไฟล์จะมีคำว่า `mongo` — ขั้นตรวจ Mongo ทำนอก script

## 8. กฎ UAT (AGENTS.md) เทียบกับสิ่งที่ suite ทำได้จริง
| ข้อกำหนด (`AGENTS.md`) | `tests/uat-crud.spec.ts` | `frontend/e2e/*` | stagehand |
|---|---|---|---|
| CRUD ครบ + edge case (`:82`) | ทำ (empty-name guard `:305`, edge cases `:168`) | ทำ (เช่น duplicate barcode `product-barcode-crud.spec.ts:220`) | ทำ |
| ตรวจ MongoDB ทีละ step (`:83`) | ทำผ่าน `mongoEval` (`:106`) | **ไม่ทำ** (ตรวจ API แทน) | **ไม่ทำ** ใน script |
| ข้อมูลอื่นต้องรอด (`:84`) | นับ `role_permission` ก่อน/หลัง (`:611`) | ยังไม่ตรวจ | ยังไม่ตรวจ |
| cleanup ด้วย id/code เท่านั้น (`:85`) | ลบด้วย `holdingcode`+`username` (`:455`) | ลบผ่าน UI/API ของ record ที่สร้าง | ยังไม่ตรวจ |
| seeded random (`:86`) | `CRUD_SEED` + metrics (`:16,34`) | ใช้ `Date.now()` uid (ไม่ reproduce ได้) | `Date.now()` uid (`uat-crud-mongo-brand.js:9`) |

## 9. ช่องว่างความครอบคลุม (packages สำคัญที่ไม่มี `_test.go`)
คำนวณด้วย `find … -type d` แล้วกรอง dir ที่มี `*.go` แต่ไม่มี `*_test.go` (ยืนยันด้วย `ls`):
| package | ทำไมสำคัญ | สถานะ test |
|---|---|---|
| `backend/internal/goapi/workers` | background worker manager (`doc_processor.go:64` `NewWorkerManager`), db log cleaner | ไม่มี test |
| `backend/internal/goapi/mydlq` | DLQ ของ projection (lead ระบุว่า log-only) | ไม่มี test |
| `backend/internal/goapi/mypostgres` (`init_schema.go`, `queue.go`) | สร้าง schema PG ต่อ holding + ตาราง queues | ไม่มี test |
| `backend/internal/goapi/process/build` (`build-product.go`, `build-debtos.go`, `build-erp-user.go`, …) | projection builder ที่ lead ระบุว่า debtor/creditor/erp_user drift — **rg ข้ามโฟลเดอร์นี้เพราะ `.ignore` มี `**/build/` ต้องใช้ `rg --no-ignore`** | ไม่มี test |
| `backend/internal/goapi/handlers/{aichat,approval,datahistory,knowledgebase,lineoa,unified}` | HTTP handler ที่ผู้ใช้เรียกตรง | ไม่มี test |
| `backend/internal/goapi/{config,setupconfig,mydb,mypg,myretry,cache/unified,myclickhouse}` | bootstrap/config/retry/cache | ไม่มี test (มีเฉพาะ `backend/internal/goapi/bootstrap_test.go`) |
| `backend/internal/product/{productgroup,productcategory,unit,promotion,optionpattern,ordertype,color}/services|repositories`, `product/eorder/services` (ไม่มี dir `repositories`), `product/bom/repositories` | master data ที่ frontend e2e แตะทุกวัน | ไม่มี test ฝั่ง Go — พึ่ง e2e อย่างเดียว (ยกเว้น `product/bom/services` มี 1 ไฟล์ `bom_http_service_test.go`) |
| `backend/internal/shop/{branch,employee}/services|repositories`, `backend/internal/shop/shop/repositories` | holding/branch/employee (ใช้ใน `tests/uat-crud.spec.ts` และ `tests/employee-*.spec.ts`) | ไม่มี test ฝั่ง Go |
| `backend/internal/warehouse/repositories` | คลัง/location/bin | มีเฉพาะ `internal/warehouse/services` (3 ไฟล์) |
| `backend/internal/organization/{branch,businesstype,department}/…` | โครงสร้างองค์กร | `organization/branch` มี 4 ไฟล์แต่ `repositories`/`services` ย่อยไม่มี |

สรุปเชิงคุณภาพ: test ฝั่ง Go หนาแน่นที่สุดที่ legacy consumer (แต่ 12 ใน 15 package quarantine อยู่กลุ่มนี้) และเส้นทาง Product/Barcode outbox→projection ใหม่ (unit + 8 integration); ส่วน CRUD master data/organization/shop เกือบทั้งหมดพึ่ง Playwright e2e ที่ไม่อยู่ใน target ใดของ `tools/verify.sh` เลย (ต้องรันมือทั้งหมด) และไม่ตรวจ Mongo

## 10. ช่องว่าง / สิ่งที่ยังไม่ตรวจ
- ยังไม่ตรวจ: ผลรันจริงของ `go test` full suite บนเครื่องนี้รอบ 2026-09-07 (อ้างผลจาก `docs/handoff/HANDOFF-RISKS-2026-09-05.md:47,109` เท่านั้น) — ไม่ได้รัน Docker เพื่อประหยัดเวลาและไม่แตะ stack
- ตรวจแล้ว (fact-checker): "24-worker pool busy-poll ตาราง queues ทุก 100 ms" ตรงกับโค้ด — `backend/internal/goapi/bootstrap.go:136` เรียก `workers.NewWorkerManager(0)` → `backend/internal/goapi/workers/doc_processor.go:65-68` คำนวณ `OptimalWorkerCount(4, 32, 2.0)` = CPU×2 (min 4, max 32); แต่ละ worker เรียก `GetActiveShops` ซึ่งยิง `SELECT DISTINCT holdingcode FROM queues WHERE status = 'pending'` (`backend/internal/goapi/mypostgres/queue.go:228-234`) แล้ว `time.Sleep(100 * time.Millisecond)` เมื่อไม่มีงาน (`doc_processor.go:252,272`); ค่า 24 มาจาก runtime ของ container local: `docker logs mainapi` 2026-09-05 23:27:40 พิมพ์ "Auto-calculated optimal workers: 24 (based on 12 CPU cores)" — บนเครื่องอื่นตัวเลขจะต่างตามจำนวน CPU (ยังไม่ตรวจค่าบน `.202`/DO)
- ยังไม่ตรวจ: เนื้อหา spec ราย ไฟล์ของ `tests/uat.spec.ts`, `tests/employee-*.spec.ts`, `tests/currency-uat.spec.ts` ว่ายังผ่านกับ UI ปัจจุบันหรือไม่ (auth.setup ผูกกับ holding `bc001`/สาขา `TST03` ที่อาจไม่มีใน DB disposable)
- ยังไม่ตรวจ: `backend/http_test/*.http` (REST client แบบ manual, coupon/debtor) และ `backend/loadtest/runtest.go` ยังใช้ได้กับ endpoint ปัจจุบันหรือไม่
- ยังไม่ตรวจ: อีก 17 ไฟล์ที่ติด tag `integration` (25 ไฟล์ทั้งหมด − 8 ที่ target `outbox`/`projection` ของ `tools/verify.sh` รัน) แต่ไม่อยู่ใน regex ของ `tools/verify.sh` — ส่วนใหญ่เป็น `*_realdb_test.go` (เช่น `backend/internal/stockprocess/stockcalculator_realdb_test.go`, `backend/internal/vfgl/journal/repositories/journal_pg_repository_realdb_test.go`) รวมถึง `backend/internal/firebase/firebase_test.go`, `backend/internal/line/line_test.go`, `backend/pkg/microservice/persister_clickhouse_test.go` — ไม่รู้ว่ายังต่อ DB/บริการจริงได้หรือเป็นซากที่คอมไพล์ผ่านอย่างเดียว
- คำถามถึงลุงจืด: (1) จะปลด quarantine ทีละ package หรือทิ้ง legacy consumer ทั้งชุดเมื่อ DEV_API_MODE=1 เลิกใช้? (2) ต้องการให้ `frontend/e2e` เพิ่มขั้นตรวจ `mongosh` ทีละ step ตามกฎ `AGENTS.md:83` หรือยอมรับการตรวจผ่าน API? (3) ~~จะปลดล็อก billing GitHub Actions หรือไม่~~ **ตอบแล้ว 2026-09-09: ไม่จ่าย ไม่ใช้ CI — GitHub เก็บโค้ดอย่างเดียว ตัวตรวจคือ `tools/verify.sh`** (4) ให้เพิ่ม `gofmt -l` + `go vet` เข้า target `backend` ของ `tools/verify.sh` ไหม (ตอนนี้ไม่มี)
