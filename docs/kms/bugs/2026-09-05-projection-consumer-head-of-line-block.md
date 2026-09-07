---
date: 2026-09-05
severity: high
component: [backend, kafka, postgres, ci]
tags: [bc-account, go, kafka, postgres, outbox]
fixed: true
---

# Symptom

หลัง commit `ef0b62e7` (transactional outbox + projection fences) consumer ใหม่ของ Product/Barcode topics ใช้ fetch → handler → sync commit และ "ไม่ข้าม head ที่ล้ม" ทำให้:

1. bulk Barcode ระดับ holding (`POST /product/barcode/bulk` สร้าง doc ที่ `businesscode` ว่าง) หรือ JSON ที่ parse ไม่ได้ → handler ตอบ "holdingcode, businesscode and barcode are required" ทุกครั้ง → reader ปิด/เปิดใหม่ทุก 2–30 วิ แล้ว replay offset เดิมตลอดกาล (ก่อนหน้านี้แค่ log แล้ว auto-commit ทิ้ง)
2. legacy service ลง Barcode topics ด้วย kafka-go reader ใน consumer group เดียวกับ librdkafka members ของ topic อื่น → JoinGroup member metadata คนละรูปแบบ (kafka-go ไม่รับ trailing OwnedPartitions ของ librdkafka v1.9.2) → rebalance วนหรือ kafka-go ไม่ได้ partition
3. GoAPI ลง `when-product-barcode-created/updated` ซ้ำใน `biapi-warehouse-consumer` (handler เดียวกับ `biapi-inventory-consumer`) → เขียน PG ซ้ำ และ head ที่ค้างทำ warehouse readers โดน rebalance ไปด้วย
4. bulk message 5000 barcode → `pg_advisory_xact_lock` 5000 ตัวใน transaction เดียว เกิน lock table default (64 × 100 ≈ 6400) → "out of shared memory" ค้างทั้ง partition
5. CI job `backend-projection-kafka-integration`: `docker compose run tests` restart `mongo-init` ที่ exit ไปแล้ว → `rs.initiate` ซ้ำล้ม (AlreadyInitialized) → gate `service_completed_successfully` แดงก่อนรัน test

พบโดย adversarial review workflow (7 มิติ × 3 lenses, 49 agents) ไม่ใช่จาก production

## Root Cause

- consumer loop ใหม่ปฏิบัติต่อ "ข้อความที่ไม่มีวันสำเร็จ" (identity/parse) เหมือน infra error (retry) — ไม่มีการจำแนก
- สมมติว่า kafka-go และ librdkafka แชร์ group ได้ — จริงแต่ metadata format ไม่ compatible
- registration ซ้ำใน manager.go มีมาก่อน แต่เดิม handler เป็น fire-and-forget จึงไม่เจ็บ
- ไม่มี cap จำนวน row lock ต่อ transaction
- compose one-shot init ไม่ idempotent

## Fix

commit ถัดจาก `42e11748` บน `dev` (ดู `git log --oneline -3`):

- `internal/product/projection/consumer.go`: `ErrRejected` — handler ห่อ error ที่ไม่มีวันสำเร็จด้วย `%w`; `Consume` ack offset นั้นแล้วไปต่อ; error อื่นยังถือ offset
- `mykafkaconsumer/product_projection.go`: `rejectProjectionMessage` log topic/partition/offset (ไม่ log payload); `handlers/kafka/{inventory,product,projection_reconcile}.go` + legacy `productbarcode_consumer{,_service}.go` ห่อ identity/parse error เป็น rejection
- `pkg/microservice/barcode_projection_consumer.go`: `barcodeProjectionGroup(group) = group+"-projection"` (offset disposable pre-launch)
- `handlers/kafka/manager.go`: ลบ 2 registration ซ้ำใน warehouse group
- `projection_reconcile.go`: `maxRowLocksPerTransaction = 256`; เกินนั้นใช้ exclusive company lock (เหมือน rebuild → ไม่ deadlock)
- `.ci/projection.compose.yml`: `rs.initiate` ใน try/catch AlreadyInitialized; `ci.yml` + README ใช้ `run --rm --no-deps tests`
- แถม (ต่ำ): `outbox.runTransaction` retry เฉพาะ `TransientTransactionError` ≤ 5 ครั้ง; Resync ข้ามสินค้าที่ถูกลบระหว่าง loop

## Regression test

- `internal/product/projection/consumer_test.go` — rejected ack, infra error หยุดโดยไม่ commit
- `internal/goapi/handlers/kafka/projection_reject_test.go` — identity/parse errors เป็น `ErrRejected`; sqlmock ยืนยัน lock cap
- `internal/goapi/mykafkaconsumer/product_projection_test.go` — mode "missing holding"/"rejected handler" ต้อง ack ต่อ
- `internal/product/product/outbox/outbox_test.go` — transient retry / non-transient ไม่ retry / cancel หยุด
- `pkg/microservice/barcode_projection_consumer_test.go` — group แยกและคงที่
- E2E: isolated compose suite (8 integration tests) ผ่าน; poison message จริงเข้า `when-product-barcode-bulk-created` บน local Kafka → log ⛔ แล้ว lag=0; browser UAT Product/Barcode CRUD บน demo/C02 ตรวจ Mongo+outbox+PG ทีละ step ผ่าน

ดู [[2026-09-05]] (daily) และ `docs/handoff/HANDOFF-RISKS-2026-09-05.md` ใน repo
