---
date: 2026-09-25
status: accepted  # proposed | accepted | deprecated | superseded
tags: [bc-account, postgres, docs, skills]
---

# ล้างเอกสารและ skill ยุค MongoDB/Kafka/Redis/ClickHouse — ให้ตรงกับระบบ PostgreSQL อย่างเดียว

## Context

- โค้ดถอด MongoDB, Kafka, Redis, ClickHouse ออกหมดแล้วเมื่อ 2026-09-23 ([ADR](2026-09-23-remove-mongo-kafka-redis-clickhouse.md)) แต่ skill และบทความ `docs/kms/` ยังสอนวิธีทำงานกับของพวกนี้อยู่ (ราว 500 จุดใน kms, 4 skill) ทำให้ AI ตัวถัดไปเสี่ยงทำตามของที่ไม่มีแล้ว
- ลุงจืดสั่ง (2026-09-25): "ตรวจตัวอื่นๆด้วย ที่ล้าสมัย เพราะ mongodb, radis, k8, kafka, clickhouse ไม่ได้ใช้แล้ว ใช้แต่ pgsql" แล้วเลือก: ขอบเขต = skills + บทความ kms (ไม่แตะ ADR/bugs), ลบ skill MongoModel, เปลี่ยน UAT skill เป็น PostgreSQL อย่างเดียว

## Decision

1. **Skills**: `datacrud` §2 เหลือแหล่งข้อมูลเดียว (BFF → Go → PostgreSQL) + roster ครบ 14 จอ; `ui-scale-polish` ตัดวิธีทำยุค Kafka/Redis/mongosh/MongoModel; **ลบ `.agents/skills/audit-mongomodel-sync/`** และตัดตัวชี้ใน `AGENTS.md` / `docs/README.md`
2. **UAT skill ระดับเครื่อง** `~/.claude/skills/uat-crud-mongo` → `uat-crud-postgres` ตรวจด้วย `psql` ทีละขั้น (อยู่นอก repo)
3. **`.mcp.json`**: ลบ server `mongodb` และ `mongomodel`
4. **docs/kms**: เขียนบทความ 01–11, 13–18, 20 และ architecture ใหม่ตามโค้ดจริง; **ลบ** `12-kafka-messaging.md`, `architecture/high-scale-multitenant-bi.md` (แบบ BI บน ClickHouse), `architecture/product-listing-api-v2.md` + `-handoff.md` (backend สินค้าถูกลบแล้ว), `snippets/picklangname.md` (PG เลิกตั้งแต่ 2026-06-11, ส่วน ClickHouse ถูกถอด) แล้วแก้ดัชนีให้ตรง
5. ของยุคเก่าเหลือได้แค่ประโยค "ถอดออกแล้ว 2026-09-23" หรือ "ห้ามเพิ่มกลับ" เท่านั้น
6. `docs/kms/decisions/` และ `docs/kms/bugs/` ไม่แก้ เพราะเป็นประวัติ ลิงก์ในนั้นที่ชี้ไปไฟล์ที่ลบแล้วจึงเสียได้ (ดูย้อนหลังด้วย `git show <commit>:<path>`)

## Alternatives

- **คง skill MongoModel ไว้** — ไม่เลือก เพราะ schema จริงตอนนี้คือ SQL ใน `mydocs/datamodels/` และ MCP ถูกถอดแล้ว
- **ทำ UAT skill ให้รองรับทั้ง PG และ Mongo** — ไม่เลือกตามคำสั่งลุงจืด (PostgreSQL อย่างเดียว)
- **ใส่หมายเหตุ "เลิกใช้" ไว้บน ADR เก่า** — ไม่เลือก เพราะ ADR เป็นบันทึก ณ เวลานั้น

## Consequences

- session/AI อื่นที่ยังเปิดอยู่ หรือ worktree เก่า (`.claude/worktrees/*`) ยังเห็นเอกสารชุดเก่า ต้อง sync กับ `dev` ก่อนเชื่อ kms
- ระหว่างตรวจพบโค้ดค้าง/พังนอกขอบเขตงานนี้ (ยังไม่แก้ รอลุงจืดตัดสินใจ): `docs/runbooks/RECOVERY-READINESS.md` ยังเขียนถึง Mongo/Kafka/Redis, `tools/seed/seed-access-setup-dev.ps1` ยังพึ่ง redis, `deploy/account/rotate-postgres-password.sh` อ้าง service `worker`/`migrate` ที่ไม่มีแล้ว, `scratch/*.go` import mongo-driver (compile ไม่ผ่าน), `backend/.github/workflows/*` ยุค consumer, dead code ใน goapi (`gemini/`, `ragflow/`, `mypostgres/queue.go`, `process/process_status.go`), frontend ยังรอซ้ำ `GL_PROJECTION_PENDING` (`frontend/src/lib/general-ledger-api.ts:124`) และ BFF สินค้า/approval ที่ backend ไม่มีแล้ว
