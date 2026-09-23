---
date: 2026-09-23
status: accepted  # proposed | accepted | deprecated | superseded
tags: [bc-account, postgres, go, architecture]
supersedes: หนี้ "ยังมีของค้าง (Mongo driver, route internal/vfgl, Kafka consumer, Redis session)" ใน AGENTS.md
---

# ถอด MongoDB, Kafka, Redis และ ClickHouse ออกจากระบบ — PostgreSQL ตัวเดียว

## Context

- ลุงจืดสั่ง (2026-09-23): "ถอด mongodb, kafka, redis, clickhouse ออกจากระบบนี้ด้วย ทดสอบด้วย" — ต่อยอดกฎ "100% Pure Single PostgreSQL Engine" ใน `AGENTS.md` และสเปก `mydocs/specs/gl/spec.md` ("ไม่มี MongoDB, ClickHouse, Kafka หรือ Redis")
- ก่อนงานนี้ prod ยังรัน `redis` เพราะ session login พึ่ง `ms.Cacher`; โค้ดยังมี Mongo driver ในทุก repository ของ ERP, Kafka consumer, ClickHouse/OpenSearch client และ `cmd/*` microservice ยุคเก่า
- ช่วง dev ข้อมูลทิ้งได้ (กฎ "เดินหน้าอย่างเดียว") จึงไม่ทำ migration ข้อมูลจาก Mongo

## Decision

1. **ลบ dependency ทั้งหมด** ออกจาก `backend/go.mod`: `go.mongodb.org/mongo-driver`, `smlsoft/mongopagination`, `segmentio/kafka-go`, `go-redis/redis/v8`, `ClickHouse/clickhouse-go/v2`, `opensearch-go`, `go-elasticsearch` — `go build ./... && go vet ./...` ผ่าน, ไม่มีชื่อเหล่านี้ใน `go.mod`/`go.sum`
2. **Session/token → PostgreSQL** ตาราง `cache_entries` ในฐานกลาง `bcai_projection` (`backend/pkg/microservice/cacher.go`) ใช้ร่วมทั้ง mainapi และ goapi; งานล้าง session หมดอายุทุก 10 นาทีใน `backend/main.go`
3. **ตรวจสิทธิ์สดทุกคำขอจาก PostgreSQL** (`pkg/microservice/live_authorization.go`): ผู้ใช้ถูกปิด / ถูกถอดจากกลุ่มกิจการ / บริษัท-สาขาถูกปิด → token เดิมใช้ไม่ได้ทันที ต่อสายด้วย `controlDB` ใน `backend/main.go` ทั้ง mainapi และ `/goapi`
4. **Auth, holding, บริษัท, สาขา, สมาชิก, พนักงาน, สิทธิ์** อ่าน/เขียน PostgreSQL ผ่าน `internal/centraldb` + repository ใหม่ (`internal/shop/shop_repository.go`, `scopes_postgres.go`, `authentication/models/role_postgres.go`, `organization/*`)
5. **จอที่ API เดิมอยู่บน MongoDB ถูกลบ backend ทิ้ง** (เอกสารซื้อขาย `/transaction/*`, สินค้า/บาร์โค้ด/คลัง, ลูกหนี้-เจ้าหนี้ master, ธนาคาร, โปรโมชัน, LINE OA ฯลฯ) — เมนูยังอยู่ครบตาม Champ แต่ขึ้น "รอพัฒนา" ผ่าน `isMenuBackendRetired` ใน `frontend/src/lib/menu-screen-status.ts` จนกว่าจะมี API บน PostgreSQL; ลุงจืดยืนยัน (2026-09-23) ว่า "ระบบอื่นๆ ยังไม่ต้องทำ" — เน้นให้ GL ใช้ได้ครบก่อน
6. **ลบ `backend/cmd/*` microservice เก่า, `cluster/`, Dockerfile/Makefile ยุค Kafka/k8s, swagger เก่า** และ e2e/UAT ที่ตรวจผ่าน `mongosh`/`redis-cli` (spec ของจอที่ถูกปลดลบทิ้ง; spec ของจอที่ยังอยู่เปลี่ยนไปตรวจ `psql` ผ่าน `tests/support/pg.ts`)
7. **Deploy**: `tools/fast-deploy.py` สำรองเฉพาะ Postgres + config (ตัด `mongodump`); `deploy/account/compose.yml` ไม่มี redis — container `redis` เดิมบน prod ลบหลัง deploy

## Alternatives

- **คง Redis ไว้เฉพาะ session** — เร็วกว่าเล็กน้อยแต่ต้องดูแลอีก 1 service และขัดสเปก; session ต่อคำขอเป็น PK lookup บน PG ซึ่งเร็วพอ → ไม่เลือก
- **พอร์ตจอ ERP ทุกจอจาก Mongo มา PG ในรอบเดียว** — งานใหญ่หลายสัปดาห์ และลุงจืดสั่งให้โฟกัส GL ก่อน → ไม่เลือก ใช้สถานะ "รอพัฒนา" แทน

## Consequences

- เมนูที่ "รอพัฒนา" เพิ่มจาก 20 เป็น 118 (98 จอ backend ถูกปลด) — ต้องสร้าง API บน PostgreSQL ทีละโมดูลเมื่อลุงจืดสั่ง
- ทุกคำขอที่ login มี query ตรวจสิทธิ์สด 2–4 คำสั่ง (timeout 2 วินาที) — แลกกับการเพิกถอนสิทธิ์มีผลทันที
- เทส integration ของ PG: `BC_GL_TEST_POSTGRES_DSN=... go test -tags=integration ./pkg/... ./internal/...` ผ่าน ยกเว้น 3 เคสเดิมที่ต้องการฐานเฉพาะ (`pkg/microservice/persister_test.go` ต่อ `localhost:5432` ตายตัว, `TestTaxWithholdingReportFromGL`/`TestTaxVatQueriesRunOnRealSchema` ต้องการฐาน `rungrueng` ที่ seed แล้ว)
