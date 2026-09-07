# Handoff: จุดเสี่ยงและงานปรับปรุง BC Ai Account

## อัปเดต 6 กันยายน 2026 — audit บทบาท MongoDB / PostgreSQL / ClickHouse + พัก ClickHouse บนเครื่อง dev

**คำถามลุงจืด:** 3 store ทำงานสัมพันธ์กันถูกต้องไหม (Mongo = เก็บทุกอย่าง, PG = เพิ่มความเร็ว/ประมวลผล, ClickHouse = dimension DB) และ ClickHouse จำเป็นไหม
**วิธีตรวจ:** scout + workflow 6 mapper (verify ถูกตัดที่ 67/180 เพราะ session limit) → Claude verify จุดชี้ขาดเองด้วยมือบน local stack (.202 ssh timeout — ไม่ได้ดู on-prem)

### ผลสรุป
- **MongoDB = SoT จริง** — ทุกหน้าจอ frontend อ่าน Mongo (รวม `/goapi/api/product/barcode/list` ที่อ่าน Mongo ไม่ใช่ PG) ข้อยกเว้นที่เขียน PG อย่างเดียว (manual-close PO, search alias, `/goapi/inventory/*`) UI ไม่ได้เรียก
- **PostgreSQL = projection ที่เขียนอย่างเดียว แทบไม่มี reader และครึ่งหนึ่งพังเงียบ**
  - ใช้ได้: product / productbarcode / ic_warehouse (biapi-* consumers) แต่ demo/C03 มี barcode ค้างใน PG 9 แถวที่ Mongo ไม่มี
  - พัง: debtor / creditor / erp_user — consumer INSERT คอลัมน์ `name0, createdat, updatedat` (และ position/department/approval_role/max_approval_amount) ที่ DDL จาก GORM model ไม่มี → ทดสอบยิง `when-debtor-created` จริง ได้ `column "name0" of relation "debtor" does not exist` → DLQ เป็น log-only (`mydlq/dlq.go:91`) offset commit ต่อ → PG ลูกหนี้ 0 แถว ทั้งที่ Mongo มี 38 (`handlers/kafka/debtor.go:110`, `creditor.go:111`, `employee.go:119`)
  - ไม่มี reader: frontend ไม่เรียก endpoint ที่อ่าน PG เลย; GL report (`vfgl/journalreport`) อ่าน PG แต่ UI ไม่ใช้; `stockwaitprocess` ไม่มีใคร drain (ProcessStockCostAll เรียกได้จาก rebuild ที่ไม่ได้ register เท่านั้น); `docwaitprocess` ใน holding test ค้าง transflag 12/44 (drainer จัดการเฉพาะ transflag 6)
  - สิ้นเปลือง: worker pool 24 ตัว poll ตาราง `queues` ว่างทุก 100 ms (`workers/doc_processor.go:239`) → 133M index scans / 136M commits
  - prod: `worker` (DEV_API_MODE=1) รัน legacy gorm consumers เขียน schema อีกชุดลง PG เดียวกับ goapi consumers; `ReportPostHandler` (rebuild ทุกคำสั่ง) ไม่ได้ register route
- **ClickHouse = ไม่มีบทบาทจริง**
  - goapi ปิดตั้งแต่ commit `dcceb83f` (2026-05-31): `myclickhouse.ClickHouseFastConnect/CreateClickHouseConnection` คืน "clickhouse is disabled", `InsertDocumentToClickHouse`/`SoftDeleteDocClickHouse` no-op, `connectClickHouse` bypass, `performBackgroundTask` ว่าง (17+3 call sites ทำงานเปล่า)
  - legacy ที่ยัง register และจะพังเมื่อเรียก: `/productimport/*` (13 routes ใช้ CH เป็น SoT ของ staging), `/stockbalanceimport/*` (9), `/product/barcode2` (hard-code holding `productbarcode_http_service.go:1886`) — ไม่มี DDL ตาราง CH ใน repo, UI ไม่เรียก; `reportqueryc` ไม่ได้ register
  - runtime local: 0 ตาราง, 7 วันมี 2 query, container RSS 5.2 GB; prod compose จอง 1 GB / 768 MB และ migrate/mainapi/worker `depends_on: clickhouse service_healthy`

### ที่ทำไปแล้ว (2026-09-06)
- **พัก ClickHouse บนเครื่อง dev** ตามคำสั่งลุงจืด ("ปิดไว้ก่อน ai coding จะได้ไม่ทำงานหนัก"): stop+rm container `clickhouse`, ลบ service/depends_on/volume ออกจาก `backend/docker-compose.local.yml` (volume `backend_clickhouse-data` ยังอยู่, 0 ตาราง); force-recreate `mainapi` แบบ cold start → healthy ใน 10 วิ, ไม่มี error, consumer groups Stable 3 members, RAM ประหยัด ≈ 5 GB
- **ไม่ได้แตะ**: โค้ด Go, `bootstrap.local.json` (block `clickhouse` ยังอยู่ → env `CH_SERVER_ADDRESS=clickhouse:9000` ชี้ host ที่ไม่มี; ไม่กระทบ startup เพราะ clickhouse-go Open ไม่ dial), prod compose, .202

### ค้างให้ลุงจืดตัดสิน
1. ถอด ClickHouse ถาวร (R1): compose local/prod + `provision-server.sh` 9 keys + bootstrap/setupconfig mapping + frontend `setup-config.ts`/`settings-screen.tsx` + goapi CH layer (~2.8k บรรทัด) + `reportqueryc` (1.1k) + 3 module legacy (ลบ หรือย้าย staging ไป Mongo/PG) + `go.mod` clickhouse-go/ch-go + `architecture/high-scale-multitenant-bi.md`
2. Mongo→PG: แก้ schema drift debtor/creditor/erp_user หรือถอด consumer ทิ้ง; DLQ → ตาราง PG + metric; ลบ/แก้ busy-poll 24 workers; prod หยุด `worker` mode 1 จนกว่าจะเหลือ projection ชุดเดียว; ล้าง barcode ค้าง demo/C03 (9 แถว) ผ่าน `/product/resync`
3. ยังไม่รู้: on-prem .202 (ssh timeout), prod `ENABLE_KAFKA`/bootstrap ที่ `/var/lib/bcai-account/config`, มี client นอก web เรียก `/productimport` หรือไม่ (ดู `shopuseraccesslogs`)


อัปเดต: 5 กันยายน 2026 | Workspace: `D:\bccode`

## อัปเดตรอบต่อ (5 กันยายน 2026 ช่วงบ่าย) — checkpoint, adversarial review, แก้ 8 จุด, UAT จริง

สิ่งที่ทำต่อจาก handoff ด้านบน (ลุงจืดสั่ง "ทำต่อเลย"):

1. **Checkpoint** — commit working tree 3 ก้อน (`caaf5c59` frontend, `ef0b62e7` backend outbox, `42e11748` chore) แล้ว push `origin/dev`; backend compile/unit/integration-compile และ frontend typecheck + vitest 44 files/320 tests ผ่านก่อน push
2. **Adversarial review workflow** (7 มิติ × 3 skeptic lenses, 49 agents) ยืนยัน 8 ข้อ ปฏิเสธ 6 ข้อ — แก้ครบ 8 ข้อแล้ว:
   - **[สูง] head-of-line block**: bulk Barcode ระดับ holding (`businesscode` ว่าง) หรือ JSON ที่ parse ไม่ได้ ทำให้ consumer ใหม่ replay offset เดิมตลอดกาล → เพิ่ม `projection.ErrRejected`; identity/parse error เป็น rejection → log topic/partition/offset (ไม่ log payload) แล้ว ack; infra/source error ยังถือ offset เหมือนเดิม (ทั้ง GoAPI และ legacy transport)
   - **[สูง] legacy group ผสม kafka-go + librdkafka**: JoinGroup metadata คนละรูปแบบ → Barcode readers ใช้ group `<CONSUMER_GROUP_ID>-projection` แยกต่างหาก (offset disposable ตามกฎ pre-launch)
   - **[กลาง] GoAPI อ่าน barcode topics ซ้ำใน `biapi-warehouse-consumer`** (เขียน PG ซ้ำ + rebalance storm ปน warehouse) → ลบ registration ซ้ำ เหลือ `biapi-inventory-consumer` กลุ่มเดียว
   - **[กลาง] 5000 advisory locks/transaction** (lock table PG default ≈6400) → batch > 256 codes ใช้ exclusive company lock แทน row locks (เหมือน rebuild)
   - **[สูง] CI job Kafka พัง**: `docker compose run tests` รัน `mongo-init` ซ้ำแล้ว `rs.initiate` ล้ม → init idempotent + `--no-deps` (ci.yml + README recipe) ทดสอบ rerun mongo-init exit 0 แล้ว
   - **[ต่ำ] WriteConflict บน aggregate เดียวกัน**: outbox commit retry เฉพาะ error ที่มี label `TransientTransactionError` สูงสุด 5 ครั้ง
   - **[ต่ำ] Resync ล้มทั้งชุดเมื่อสินค้าถูกลบระหว่าง loop** → ข้ามตัวนั้น (`errResyncProductRemoved`)
   - unit tests ใหม่: `projection/consumer_test.go`, `handlers/kafka/projection_reject_test.go` (sqlmock lock cap), `outbox_test.go` (transient retry), `barcode_projection_consumer_test.go` (group), แก้ `product_projection_test.go`
3. **หลักฐาน**: gofmt/vet สะอาด, full unit suite (ตัด quarantine 15 แพ็กเกจ) ผ่าน, isolated compose suite บน Kafka 4.3.1 + Mongo 7 + PG 18 ผ่านทั้ง 8 integration tests, mainapi local rebuild แล้ว UAT ผ่าน browser บน `demo/C02`: Product Create→Update→Delete และ Barcode Create→Update→Delete ตรวจ MongoDB (doc เดิม/soft-delete) + `outboxevents` (v1..v3 PUBLISHED ทุกตัว, pending 0) + PostgreSQL (row สร้าง/แก้/ลบ, `product_projection_fences` v3 deleted=true) ทีละ step, ข้อมูลเดิม 15/15 ไม่กระทบ, consumer lag 0 ทุก group; poison test ส่ง message ไม่มี businesscode เข้า `when-product-barcode-bulk-created` จริง → log ⛔ แล้วข้าม, message ถัดไปประมวลผลปกติ
4. **ยังไม่ทำ / พบระหว่างทาง (ไม่แก้ เพราะนอกขอบเขต)**: `handlers/kafka.go:191` log format ผิด (`$0%!(EXTRA ...)`) และ echo payload ทั้งก้อน; dialog ลบสินค้าใช้ข้อความ "ลบบาร์โค้ด"; แถวใน unit picker ไม่มี accessible name; tenant PG `test` มี schema เก่า (product ไม่มี holding_code/businesscode) ต้อง rebuild ตามกฎ disposable; `pkg/microservice/persister_mongo_test.go` มี vet warning เก่า 15 จุด; คำถาม 4 ข้อท้ายไฟล์ยังรอลุงจืดตอบเหมือนเดิม

## สถานะตามลำดับงานของลุงจืด

แก้ source และทดสอบบนระบบแยกแล้วตามรายการนี้ รักษางาน UI/config/language/skill ที่มีอยู่ก่อน ไม่ได้ commit, deploy, restart ระบบจริง หรือ migrate ข้อมูลบัญชี

| งาน | ผลที่ทำแล้ว | สิ่งที่ยังต้องทำ |
| --- | --- | --- |
| 1. Product version fence / legacy Barcode / real Kafka | เพิ่ม fence ข้าม Product topics, primary-source reconciliation, legacy Barcode writer/ack; company Barcode CRUD + Unit อัตโนมัติใช้ outbox; real Kafka 2 consumers/2 partitions + MongoDB 7 + PostgreSQL 18 ผ่าน | ยังไม่ตรวจหลาย broker/full service deployment และ legacy source writers นอก company CRUD |
| 2. Exact accounting numbers | ระบุเส้นทาง float เดิมและเกณฑ์ตรวจ; outbox รักษารูปแบบ payload เดิม; version ส่งเป็น decimal string | รอยืนยัน workflow แรก, currency, precision/scale, rounding mode/จุดปัดเศษ และ trusted source ก่อนเปลี่ยนสัญญาข้อมูล |
| 3. Quarantine / browser CRUD UAT | CI รายงาน quarantine และแก้ตัวกรองให้รองรับ CRLF; unit suite ที่ตัด 15 แพ็กเกจเดิมผ่าน; service CRUD query DB ทีละ step ผ่าน | ยังไม่ปลด quarantine หรือทำ browser CRUD UAT เพราะ business contracts ยังไม่ยืนยัน |
| 4. Backup / restore drill | มี runbook อ้างอิง compose/provision จริง เพิ่มการกู้ fences/offsets/outbox ให้สอดคล้องกัน | รอข้อมูล backup ภายนอก, source/target, format, RPO/RTO; ยังไม่ได้ restore |
| 5. MongoModel technical workflow/index | revision 1423: technical workflow draft 18 steps/22 transitions, queue index และคำอธิบาย; แก้ field references เดิม 6 จุดแล้ว ทุก 12 workflows lint 0 issues | เครื่องมือแทน partialFilter/index name ไม่ได้; Unit model ยังต่างจาก implementation |

นี่เป็นหลักฐานทางเทคนิค ไม่ใช่การรับรองธุรกรรมบัญชีทั้งระบบหรือ production readiness

## Event delivery และ projection ที่แก้

ปัญหาเดิม: Product service เขียน MongoDB ก่อน MQ ทำให้ DB สำเร็จแต่ API แจ้งล้ม และบาง event อาจหาย ส่วนข้อความ create/update/delete คนละ topic รวมถึง legacy writer สามารถมาถึงผิดลำดับได้

1. Product Create/Update/Delete/Resync บันทึก `outboxevents` subtype `productprojection` ใน MongoDB transaction เดียวกับการเปลี่ยน Product/linked Barcode มี version, lease/heartbeat และ retry
2. Dispatcher เติม `_projection={aggregateuid,eventuid,version}` ให้ Product topics โดย version เป็น decimal JSON string ไม่เสีย precision เหนือ 2^53; Barcode array shape เดิมยังอยู่
3. GoAPI ทั้ง 9 Product/Barcode topics ใช้ fetch → handler → synchronous offset commit และเปิด reader ใหม่หลังล้มเหลว โดยไม่ fetch ข้อความถัดไปใน loop ที่ล้ม
4. Legacy Microservice Barcode registrations ใช้วงจร acknowledgement เดียวกัน มี lifecycle cancellation และ retry; selector รองรับ Barcode 6 topic names แต่ legacy service เดิมลงทะเบียน single create/update/delete กับ bulk create อยู่ 4 เส้นทาง
5. Handler ถือ PostgreSQL advisory lock แยก holding/business/code ก่อนอ่าน MongoDB primary แล้วเขียน metadata ปัจจุบัน แทนการเอา snapshot เก่ามาทับ ทั้ง GoAPI และ legacy Barcode source lookup บังคับ primary แม้ connection ตั้ง secondary
6. Product บันทึก version/deleted ลง `product_projection_fences` พร้อม projection ใน transaction เดียวกัน รุ่นเก่าหรือซ้ำข้าม topics ถูกข้าม ส่วน legacy ไม่มี version และการลบแล้วสร้าง code เดิมด้วย GUID ใหม่แก้ด้วย current-source reconciliation
7. Product company rebuild ถือ exclusive company lock ก่อนอ่าน primary; event handlers ถือ shared company lock + exclusive row lock และ Barcode batch ล็อก code ตามลำดับเดียวกัน
8. Legacy Barcode metadata upsert ระบุคอลัมน์ชัดเจนและไม่เขียนทับ balanceqty/balanceamount/averagecost ที่มีอยู่; helper DELETE/COPY เดิมยัง atomic
9. Company Barcode Create/Update/Delete บันทึก intent ใน transaction เดียวกับ source ใช้ aggregateuid prefix `barcode:` แยกลำดับจาก Product; Product ที่สร้างร่วมเป็น unversioned current-source signal เพื่อไม่ให้ version ของ Barcode รบกวน Product fence
10. Company Barcode Create รวม Unit ที่สร้างอัตโนมัติไว้ใน transaction ด้วย ใช้ normalization/name mapping เดิมจาก shared Unit builder และบันทึก Unit/Product/Barcode messages ตามลำดับ ทดสอบ outbox insert failure แล้วย้อนกลับครบ 3 collections ไม่เหลือ orphan
11. Exported GoAPI Barcode insert/update/bulk/delete helpers และ `handlers.ProductBarcodeBuild` ส่งผ่าน canonical reconciliation แล้ว; ลบ private snapshot-writing helpers ที่ไม่มี caller อื่นและ ClickHouse placeholders ที่ว่าง

ไฟล์หลัก:

- `backend/internal/product/product/outbox/` — transaction intent, dispatch metadata, retry และ integration tests
- `backend/internal/product/projection/` — aggregate identity, exact version metadata, shared locks และ acknowledged consumer loop
- `backend/internal/goapi/handlers/kafka/projection_reconcile.go` — current-source reconciliation + PostgreSQL fence
- `backend/internal/goapi/mykafkaconsumer/product_projection.go` — GoAPI topic acknowledgement
- `backend/pkg/microservice/barcode_projection_consumer.go` — legacy Barcode transport
- `backend/internal/product/productbarcode/repositories/productbarcode_pg_repository.go` — legacy GORM lock/read/upsert transaction
- `backend/internal/product/productbarcode/services/productbarcode_projection_source.go` — บังคับ primary เฉพาะ consumer lookup
- `backend/internal/goapi/process/build/build-product.go` — company rebuild lock ก่อน source snapshot

API success ของ Product และ company Barcode CRUD หมายถึง MongoDB commit แล้ว PostgreSQL อาจตามภายหลัง การส่งยังเป็น at-least-once; PUBLISHED ยืนยันเพียง broker รับ ไม่ใช่ทุก consumer เขียน DB เสร็จ

Resync HTTP ยัง rebuild PostgreSQL แล้ว queue active snapshots และตอบ `queued` พร้อม legacy `published:0` การ rebuild กับ queue ทั้งชุดไม่ได้เป็น cross-database transaction

ดู configuration, คำสั่งทดสอบ, การตรวจคิว และ rollback ใน [Product outbox README](backend/internal/product/product/outbox/README.md)

## จุดแก้เพิ่มเติมที่พบระหว่างตรวจ

- `backend/pkg/microservice/persister.go`: เดิมทิ้ง error จาก GORM Transaction และคืน nil เสมอ แก้ให้ส่ง begin/callback/commit failure กลับ พร้อม regression tests
- `backend/internal/shop/shopuser_service.go`: ฟังก์ชัน SaveUserPermissionShop เดิม resolve edit target แม้เป็น create ที่ editusername ว่าง ทำให้ create branch เข้าไม่ถึง แก้ให้ resolve เฉพาะเมื่อมี edit id; unit test ยืนยัน create สำเร็จ, ส่ง write error กลับ และ edit เป้าหมายที่ไม่มีอยู่ยังถูกปฏิเสธ ไม่เปลี่ยน role policy ทั้งนี้ HTTP ปัจจุบันใช้ SaveUserFullProfile อยู่แล้ว
- `.github/workflows/ci.yml`: เพิ่ม real Kafka integration job แยกจาก outbox job เดิม พร้อม artifacts/cleanup; ตัด CR จาก quarantine file ก่อนเปรียบเทียบ package names เพื่อให้ผล Windows mount ตรงกับ CI
- งานก่อนหน้า: auth tests แยก development/test/production ก่อน import และตรวจ Secure ตาม environment; Product proxy ส่งต่อ queued count; AI_INDEX แก้ทางเข้าที่อ้าง docs ที่ลบแล้ว

## หลักฐานทดสอบ

ผลที่รันจริงใน workspace:

- Backend compile ทุก package ใน `./cmd/...`, `./pkg/...`, `./internal/...` รวม integration-tag compile ผ่าน; build root main entry point ด้วย `go build -tags musl -o /tmp/bc-risk-main .` ผ่าน
- Full backend unit suite ตัด 15 แพ็กเกจ quarantine เดิม ผ่าน; targeted GoAPI และ Barcode/Unit tests หลังเปลี่ยน helper ผ่าน
- MongoDB แยก: transaction rollback, retry/order, restart-before-ack, competing outbox workers, partial delivery และ unrelated outbox exclusion ผ่าน
- Product service CRUD: query MongoDB หลัง Create/Update/Delete, duplicate/invalid input, _id เดิม, tenant isolation, soft-delete audit และ active-only Resync ผ่าน
- Company Barcode service CRUD: query หลังแต่ละ step, duplicate/immutable identity rejection, stock guard ระดับ holding เดิม, soft-delete/อีกบริษัทไม่เปลี่ยน, broker failure ไม่ทำ intent หาย และ outbox failure rollback Unit/Product/Barcode ผ่าน
- Real Kafka rebalance: 2 consumer members / 2 partitions, สมาชิกออกจากกลุ่มหลัง handler fail, อีกสมาชิก replay offset ที่ยังไม่ commit และรับครบ 12 unique offsets ผ่าน
- Real Kafka 4.3.1 + PostgreSQL 18: ส่งผ่าน Confluent producer/acknowledged kafka-go loop; บังคับ consume delete ก่อน create/update, duplicate replay, code reuse GUID ใหม่ และ stale legacy Barcode signals ผ่าน
- PostgreSQL constraint rejection: projection และ version fence rollback, ไม่ commit offset; เปิด consumer group เดิมใหม่แล้ว replay สำเร็จหลังแก้สาเหตุ
- Legacy GORM writer บน PostgreSQL 18: source error ไม่ถูกกลืน, source callback ไม่ถูกเรียกก่อนรับ shared lock, replay รักษาค่า NUMERIC `12.3400 / 56.7800 / 4.6000`, ลบเฉพาะบริษัทเป้าหมาย ผ่าน
- Real MongoDB primary regression: client ตั้ง SECONDARY-only แต่ replica set มี primary ตัวเดียว; consumer อ่านสำเร็จผ่าน primary override, soft-delete และบริษัทอื่นยังถูกต้อง
- DELETE/COPY helper: rollback/replay/tenant isolation และ NUMERIC totals ผ่านทั้งการตรวจ PG17 รอบก่อนและ PG18 รอบนี้
- actionlint, Compose configuration และ git diff --check ผ่าน
- Frontend รอบก่อน: 44 files / 320 tests และ typecheck ผ่าน; รอบนี้ไม่แก้ UI หรือรัน visual UAT ใหม่

ชุดแยกใช้ `backend/.ci/projection.compose.yml` ไม่มี host ports หรือ production volumes และ tests สร้าง/ลบเฉพาะชื่อ database/schema/topic ของรอบนั้น ตรวจ Compose project labels แล้วลบ container/network/volumes ของรอบทดสอบเรียบร้อย และคืน tracked log ที่ tests สร้างแล้ว CI workflow ยังไม่ได้ push ให้ GitHub รัน

ข้อจำกัด: ชุดนี้เป็น service/integration tests ไม่ใช่ browser UAT ใน appdb, ไม่ใช่ full deployed legacy-service test, ยังไม่ทดสอบ broker failover, ClickHouse หรือ backup restore

## ความเสี่ยงที่ยังเหลือและเงื่อนไข rollout

- ทุก writer ที่แข่งกันต้องใช้ protocol lock/source-read เดียวกัน ก่อนถือว่าป้องกัน stale writes ครบ ไม่รวม manual Barcode/admin/stock rebuild ทุกตัวในรอบนี้
- Legacy Barcode writers ที่ไม่ระบุ businesscode และ raw import/batch บางเส้นทางยังมี DB-before-MQ gap; company CRUD และ Unit ที่สร้างร่วมถูกย้ายเข้า outbox แล้ว การอ่าน current source ไม่สร้าง signal ที่ไม่เคยส่งให้เอง
- Legacy Microservice topics อื่นยังใช้ auto-commit/handler error behavior เดิม ไม่ได้ขยายการแก้ทั้งระบบโดยไม่มี contract
- Projection ต้องเข้าถึง MongoDB primary และ PostgreSQL มีสิทธิ์สร้าง fence table; เมื่อ source/DB ล้ม handler จะ retry และคิวอาจค้าง
- Poison message/aggregate head จะไม่ถูกข้ามอัตโนมัติ ยังไม่มี dead-letter routing หรือ production alert integration
- Metadata เดิมยังมี float ของราคา/จำนวน ไม่ใช่ exact-accounting migration และ tests การรักษา NUMERIC ไม่รับรองตัวเลขเดิมทั้งระบบ
- ห้าม purge outbox history เพราะ next version ยัง derive จากประวัติ ต้องมี durable version floor ก่อนทำ retention
- Rollout consumer/writer ที่เกี่ยวข้องพร้อมกัน และรักษา group ID/offsets; rollback ต้องรักษา outbox + fences + offsets ร่วมกัน การย้อนเฉพาะ consumer เก่าจะสูญเสีย fence/lock guarantee
- ยังไม่ย้าย production, reset offsets, replay live events หรือแก้ข้อมูลย้อนหลัง

## MongoModel

ใช้ skill `audit-mongomodel-sync` และ live MCP อ่านก่อนแก้: baseline 1408 → 1412 → 1416 และรอบนี้ถึง 1423 เพิ่ม queue index และ draft workflow `product_projection_delivery` ชื่อ “ส่ง Product และ Barcode เข้าตารางอ่านอย่างปลอดภัย” พร้อม dataAccess ที่อ้าง collection/field IDs จริง อัปเดต company Barcode/Unit transaction และ rebalance evidence ตาม source

Description/model error-level lint ผ่าน; ทุก 12 workflows มี 0 issues แล้ว พบว่า 6 errors เดิมใน `system_setup_steps` ใช้ field name แทน field ID ที่มีอยู่ จึงแก้เฉพาะ references โดยคง approved status, business text และ transitions เดิม ไม่ต้องเปลี่ยนกติกาสิทธิ์

เครื่องมือ collection index ยังแทน partialFilter/index names ไม่ได้ จึงบันทึก exact specification ในคำอธิบายและ source พร้อมรักษา generic unique index เดิม ไม่อ้างว่าสอดคล้องเชิงโครงสร้างครบแล้ว การอัปเดต model ไม่แตะ live database

## ข้อมูลที่รอและงานรับช่วง

คำถามที่ส่งไว้ยังไม่ได้คำตอบ จึงคงงานที่พึ่งข้อมูลนั้นไว้:

1. ยืนยัน workflow แรก, currency, precision/scale ของเงิน/จำนวน/ต้นทุน/FX, rounding mode/จุดปัดเศษ และ trusted source จากนั้นไล่ request → calculation → MongoDB → Kafka → PostgreSQL/report พร้อม round-trip, 0.1+0.2, rounding boundaries, debit=credit, idempotency และ exact reconciliation
2. ยืนยัน business contracts ของ 15 quarantined packages แล้วปลดทีละ package; เพิ่ม browser Create → DB query → Read → Update → DB query → Delete → DB query พร้อม seeded data และ cleanup ตาม id
3. ระบุ backup platform/config reference, source/isolated target, format, RPO/RTO และหลักฐาน backup ล่าสุด โดยไม่ส่ง secrets; ทำ drill ตาม [Recovery readiness](docs/runbooks/RECOVERY-READINESS.md)
4. ปิด structural partial-index gap เมื่อเครื่องมือรองรับ; ยืนยัน Unit contract เพราะแบบใช้ `unit_of_measure/code/businesscode` แต่ implementation ใช้ `units/unitcode` ระดับ holding ไม่ควรย้ายขอบเขต tenant โดยเดา

ตรวจ runtime แบบ read-only พบ mainapi เริ่ม 2026-09-03 และ frontend build artifact วันที่ 2026-09-02 ซึ่งเก่ากว่า source รอบนี้ จึงยังไม่มี browser UAT ที่รับรอง source ล่าสุด ไม่ได้ restart runtime เดิมหรือใช้ผลจาก binary เก่ามารับรอง diff ใหม่

`AGENTS.md` ระบุว่า docs ธุรกิจเดิมถูกลบแล้ว `docs/kms/00-source-router.md` ใช้ค้น implementation เท่านั้น จึงไม่ใช้ตัวเลขหรือ test ที่มีอยู่เป็นการอนุมัติ business contract ใหม่
