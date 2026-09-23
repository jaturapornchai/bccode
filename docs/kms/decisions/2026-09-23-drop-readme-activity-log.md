---
date: 2026-09-23
status: accepted  # proposed | accepted | deprecated | superseded
tags: [bc-account, docs, process, git-hooks]
supersedes: กฎ "ต้องบันทึกประวัติการแก้ไขใน README.md ทุกครั้ง" (AGENTS.md, 2026-09-09)
---

# เลิก Activity Log ใน README — ประวัติการแก้ไขดูจาก git log + ล้างกฎ Mongo/Redis ที่ปลดระวางแล้วออกจาก AGENTS.md

## Context

- กฎ 2026-09-09 บังคับให้ทุกงานที่แก้โค้ดต้องเพิ่มรายการใน "Project Activity Log" ของ `README.md` และ `.githooks/pre-commit` (Check 1) ปฏิเสธ commit ที่ stage ไฟล์ใต้ `frontend/src` หรือ `backend` โดยไม่มี `README.md`
- README บวมถึง 1,817 บรรทัด — ลุงจืดเขียน README ใหม่เหลือ 134 บรรทัดใน commit `96c73cbb` (2026-09-21) และตัด Activity Log ออก ทำให้กฎ + hook ขัดกับสภาพจริง (ทุก commit โค้ดต้องใช้ `SKIP_README_CHECK=1` หรือเขียน log กลับเข้า README)
- การอ่านโปรเจ็กต์ใหม่ 2026-09-23 พบว่า `AGENTS.md` ยังมีหัวข้อที่ขัดกับ "คำสั่งล่าสุด: PostgreSQL เท่านั้น" (2-Tier Mongo, UAT ตรวจใน Mongo, รูปภาพใน Mongo, topology ที่มี redis, 12 ภาษา, ส่ง DeepSeek แปล)

## Decision

ลุงจืดเลือกตัวเลือก [1] (2026-09-23):

1. ตัดกฎ Activity Log ใน README → แทนด้วยกฎ "ประวัติการแก้ไขดูจาก git log" — commit subject อังกฤษแบบ conventional + **body สรุปภาษาไทย** (แก้อะไร ได้อะไร evidence); เหตุผลการตัดสินใจอยู่ใน ADR, บั๊กอยู่ใน `docs/kms/bugs/`
2. ลบ Check 1 (README guard) ออกจาก `.githooks/pre-commit`; Check คำต้องห้ามและ CODE-MAP คงเดิม
3. ล้างหัวข้อที่ปลดระวางใน `AGENTS.md`:
   - ลบหัวข้อ "สถาปัตยกรรม 2-Tier — MongoDB เก็บย่อ + PostgreSQL ประมวลผล" ทั้งหัวข้อ (หลักการ "ประมวลผลใน PostgreSQL" ยังอยู่ในกฎ Pure PostgreSQL + Backend-First)
   - UAT: ตรวจทีละ step ใน PostgreSQL ของ holding (`psql` ใน container `postgres`) แทน mongosh
   - รูปภาพ: ห้ามเก็บ binary ในฐานข้อมูล เก็บแค่ URI; ไฟล์อยู่ใน MinIO + thumbnail (ตาม `mydocs/specs/rules.md`)
   - Topology: เป้าหมาย `postgres` + `minio` + `mainapi` + `frontend`; ระบุสถานะจริงว่า prod ยังรัน `redis` สำหรับ session login (หนี้ที่ต้องถอด)
   - i18n: ระบุว่าเปิดแค่ th + en; ลบคำว่า "ส่ง batch ให้ DeepSeek แปล"
   - DevOps: ลบ Mongo ออกจาก preflight backup และ prepared queries

## Alternatives

- **[2] ย้าย log ไป `CHANGELOG.md` แยก ไม่ผูก hook** — ไม่เลือก: เป็นไฟล์ที่ต้องดูแลเพิ่มโดยไม่มีตัวบังคับ และซ้ำกับ git log
- **[3] คืน Activity Log เข้า README** — ไม่เลือก: ย้อนการตัดสินใจของลุงจืดเมื่อ 2026-09-21 และ README จะบวมกลับ ทุก AI ต้องอ่านไฟล์ยาวขึ้นเรื่อย ๆ

## Consequences

- ✅ ไม่มีกฎขัดกันเอง; commit โค้ดไม่ต้องแตะ README อีก; AGENTS.md สั้นลง (506 → 480 บรรทัด)
- ⚠️ หน้า GitHub ไม่มีประวัติภาษาไทยอ่านง่ายแล้ว — ต้องพึ่ง body ของ commit ภาษาไทย (ไม่มี hook บังคับ)
- ⚠️ โค้ดยังมี Mongo/Kafka/Redis ค้างอยู่ (Mongo driver, `internal/vfgl`, Kafka consumer, Redis session) — AGENTS.md บอกชัดว่าเป็นหนี้ ไม่ใช่แบบอย่าง; skill `.agents/skills/audit-mongomodel-sync` ยังอยู่และต้องตัดสินแยก
- ทุก clone ต้องรัน `npm run hooks:install` ใหม่เพื่อให้ hook ใน `.git/hooks/` ตรงกับ `.githooks/`
