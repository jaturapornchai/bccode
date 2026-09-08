---
date: 2026-09-06
status: accepted  # proposed | accepted | deprecated | superseded
tags: [bc-account, architecture, clickhouse, postgres, docker]
---

# พัก ClickHouse บนเครื่อง dev (ยังไม่ถอดถาวร)

## Context

audit บทบาท 3 store (Mongo = SoT, PG = speed layer, ClickHouse = dimension DB) วันที่ 2026-09-05/06 พบว่า ClickHouse ไม่มีบทบาทจริง:

- goapi ปิดตั้งแต่ commit `dcceb83f` (2026-05-31): `myclickhouse.ClickHouseFastConnect` คืน "clickhouse is disabled", `InsertDocumentToClickHouse`/`SoftDeleteDocClickHouse` no-op, `connectClickHouse` bypass
- legacy ที่ยัง register แต่พังเมื่อเรียก (ไม่มี DDL ตาราง CH ใน repo, UI ไม่เรียก): `/productimport/*`, `/stockbalanceimport/*`, `/product/barcode2`
- local: 0 ตาราง, 7 วันมี 2 query, แต่ container `clickhouse/clickhouse-server:latest` กิน RSS 5.2 GB (มากกว่า mainapi+mongo+postgres รวมกัน)
- prod compose: จอง 1 GB / 768 MB และ migrate/mainapi/worker `depends_on: clickhouse service_healthy`

ลุงจืดสั่ง "ปิดระบบ clickhouse ไว้ก่อน ai coding จะได้ไม่ทำงานหนัก"

## Decision

พัก (dormant) เฉพาะเครื่อง dev:

- stop + rm container `clickhouse`; ลบ service / `depends_on` / volume ออกจาก `backend/docker-compose.local.yml` (volume `backend_clickhouse-data` ยังอยู่)
- ไม่แตะโค้ด Go, `bootstrap.local.json`, prod compose, .202

## Alternatives

- **ถอดถาวร (A)** — ต้องแก้ compose prod + provision script + bootstrap/setupconfig + frontend setup UI + ลบ goapi CH layer ~2.8k บรรทัด + reportqueryc + ตัดสินใจ 3 module legacy → R1 รอลุงจืดสั่งแยก
- **ฟื้นตามแบบ (C)** — ไม่มี reader/BI requirement และข้อมูลระดับร้อย–พันแถว; PG พอจนถึงหลายสิบล้านแถว

## Consequences

- ✅ RAM เครื่อง dev ลด ≈ 5 GB; `up -d` ไม่ปลุก CH อีก; mainapi cold start โดยไม่มี CH → healthy ใน 10 วิ, ไม่มี error, consumer groups Stable
- ⚠️ 3 route legacy ยัง 500 เมื่อเรียก (เหมือนเดิม); env `CH_SERVER_ADDRESS=clickhouse:9000` ชี้ host ที่ไม่มี (ไม่กระทบเพราะ clickhouse-go `Open` ไม่ dial)
- 🔁 ย้อนกลับ: `git revert` commit compose แล้ว `up -d` (volume ยังอยู่)

เกี่ยวข้อง: [[2026-09-05-projection-consumer-head-of-line-block]], `docs/handoff/HANDOFF-RISKS-2026-09-05.md` (section 6 ก.ย.)
