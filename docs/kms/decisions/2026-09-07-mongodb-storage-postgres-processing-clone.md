---
date: 2026-09-07
status: accepted  # proposed | accepted | deprecated | superseded
tags: [bc-account, architecture, mongodb, postgresql, data-store, 2-tier, outbox, projection]
---

# สถาปัตยกรรม 2-Tier Data Store: MongoDB เก็บย่อ (Storage) + PostgreSQL ประมวลผลเร็วแบบครบจบ (Processing Engine)

## Context

ระบบ BC Ai Account มีการใช้งานทั้ง MongoDB และ PostgreSQL:
- MongoDB (`appdb`) ทำหน้าที่เป็นตัวรับการเขียน (Write Store / Intake) ผ่าน REST API และ Document Model
- PostgreSQL (per-holding DB `<holdingcode>`) มีตาราง projection และ costing/inventory engine บางส่วน แต่เดิมขาดความชัดเจนว่าบทบาทของ PostgreSQL จะใช้ทำอะไรแน่ (ระบุไว้ในคำถามค้างข้อ 6 ของ `docs/kms/18-decisions-and-agreements.md` และ `docs/kms/02-data-stores.md:121`)

วันที่ 2026-09-07 ลุงจืดได้ตั้งกฎสถาปัตยกรรมหลักเพื่อกำหนดบทบาทของทั้ง 2 store ให้ชัดเจนและเด็ดขาด:
> "ข้อมูลใน mongodb ต้อง ไปสร้าง clone ใน pgsql ด้วย เพราะ mongodb จะเอาไว้เก็บข้อมูล แต่การประมวลผลทั้งหมด จะอยู่ใน pgsql เพื่อความเร็ว mongodb ต้องประหยัดขนาดของข้อมูลด้วย pgsql ต้องมีรายละเอียดครบ เพราะตอนใช้ pgsql จะได้ไม่ต้องมาเชื่อม mongodb อีก"

## Decision

กำหนดสถาปัตยกรรม **2-Tier Data Store Pattern**:

1. **Single-Direction Clone (MongoDB → PostgreSQL)**:
   - ข้อมูล Master Data, Configuration, เอกสาร Transactions ทั้งหมดที่เขียนลง MongoDB ต้องถูก clone/project ไปยัง PostgreSQL per-holding DB เสมอผ่าน Event-driven pipeline (Transactional Outbox + Kafka + Consumer)
2. **MongoDB = Compact Storage Layer**:
   - หน้าที่: รับการเขียน (CRUD), เก็บรักษาข้อมูลดิบ และสถานะเอกสาร
   - หลักการ: ประหยัดขนาดพื้นที่จัดเก็บ (Slim / Compact Payload), ห้ามเก็บข้อมูลบวมซ้ำซ้อน, ห้ามเก็บ binary/base64 (เก็บเพียง URI ของ S3/MinIO)
3. **PostgreSQL = High-Speed Processing Engine**:
   - หน้าที่: เป็นเครื่องจักรคำนวณและประมวลผลทั้งหมดของระบบ (Stock Costing, Balance, FIFO, Ledger, Accounting, Sales/Purchase Summary, Search, Reporting & Analytics)
   - ใช้ประโยชน์จาก SQL, Foreign Keys, Compound Index, Partitioning, Aggregation เพื่อความเร็วสูงสุด
4. **PostgreSQL Self-Contained (รายละเอียดครบถ้วน ไม่ย้อนกลับมาต่อ MongoDB)**:
   - ข้อมูลใน PostgreSQL ต้องมี Schema และ Payload ที่ Denormalize / Enriched ครบถ้วน (รวมชื่อภาษาต่างๆ, หน่วยนับ, ข้อมูล Master ที่จำเป็นต่อการคำนวณและแสดงผลรายงาน)
   - ขณะที่ Processing Engine หรือ Reporting ทำงานบน PostgreSQL **ต้องทำงานจบในตัว 100% โดยไม่ต้องเชื่อมต่อหรือข้ามกลับมาดึงข้อมูลจาก MongoDB อีกเด็ดขาด** (Zero Cross-DB Join / Lookup at runtime)

## Consequences

- ✅ **Clear Separation of Concerns**: แบ่งหน้าที่ชัดเจน Mongo = Intake/Document Archive, PG = Compute/Query Engine
- ✅ **Maximum Query & Calculation Speed**: งานบัญชี รายงาน และการตัดสต็อกประมวลผลบน Relational Engine แท้ๆ ของ PostgreSQL รวดเร็วกว่าการทำ Map-Reduce หรือ Aggregation ซับซ้อนบน MongoDB ขนาดใหญ่
- ✅ **Cost & Storage Optimization**: MongoDB ข้อมูลขนาดกะทัดรัด ประหยัด RAM/Disk Index cache
- ✅ **Independent Processing**: เมื่อระบบรันงาน batch/report ขนาดใหญ่บน PostgreSQL จะไม่แย่ง I/O หรือ Lock กับ MongoDB ที่กำลังรับการสร้างเอกสารหน้าร้าน
- ⚠️ **Sync Pipeline Responsibility**: ต้องซ่อมแซมและบำรุงรักษา Consumer / Outbox pipeline (เช่น ซ่อมตาราง `debtor`, `creditor`, `doc*` ที่เคยมี DDL drift) ให้ clone ข้อมูลลง PG ได้อย่างแม่นยำและเป็น Real-time/Near-real-time
- ⚠️ **Schema Migration on PG**: เมื่อมีการเพิ่มฟิลด์ใน Mongo และจำเป็นต่อการประมวลผล ต้องขยายคอลัมน์ใน PostgreSQL ให้สอดคล้องกัน

## Related
- `AGENTS.md` (กฎ สถาปัตยกรรม 2-Tier Data Store)
- `docs/kms/02-data-stores.md`
- `docs/kms/18-decisions-and-agreements.md`
