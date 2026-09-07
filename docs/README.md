# BC Ai Account — Documentation Router

> **กฎประหยัด Context (On-Demand Loading)**:  
> AI และนักพัฒนา **ห้ามกวาดอ่านเอกสารทั้งหมดหรือเปิด HANDOFF ล่วงหน้า** เพราะจะสิ้นเปลือง Context Window มหาศาล  
> ให้เปิดอ่านเฉพาะไฟล์ที่ตรงกับขอบเขตงานจริงตามตารางด้านล่างเท่านั้น

---

## แผนที่เลือกอ่านตามประเภทงาน (On-Demand Selection)

| ประเภทงาน | สิ่งที่ต้องอ่าน (อ่านเฉพาะไฟล์นี้) | สิ่งที่ไม่ต้องอ่าน (ห้ามเปิด) |
|---|---|---|
| **งานเล็ก / แก้บั๊ก 1 บรรทัด / คำถามทั่วไป** | ดูโค้ดจริงโดยตรง (`code = truth`) | ไม่ต้องเปิด `docs/` ใด ๆ |
| **งานหน้าจอ / สไตล์ / CSS / UX** | [`docs/skills/ui-scale-polish/SKILL.md`](skills/ui-scale-polish/SKILL.md) | ไม่ต้องอ่าน `kms/` หรือ `handoff/` |
| **งาน Schema / MongoDB / MongoModel** | [`docs/skills/audit-mongomodel-sync/SKILL.md`](skills/audit-mongomodel-sync/SKILL.md) | ไม่ต้องอ่านไฟล์ UI |
| **งานสถาปัตยกรรม / โดเมนเฉพาะเรื่อง** | ดูดัชนีใน [`docs/kms/README.md`](kms/README.md) แล้วเลือก **1 บทความที่ตรงเรื่อง** | ห้ามเปิดบทความอื่นที่ไม่ได้ทำ |
| **ค้นหาตำแหน่งโค้ด / สัญญาณระบบ** | [`docs/kms/00-source-router.md`](kms/00-source-router.md) | ไม่ต้องเปิดบทความยาว |
| **นำทางไฟล์ขนาดใหญ่ (>1,000 บรรทัด)** | [`docs/reference/CODE-MAP.md`](reference/CODE-MAP.md) | ไม่ต้อง grep ทั้งไฟล์ |
| **สถานะงานค้าง / สรุปความเสี่ยงระบบ** | [`docs/handoff/HANDOFF-2026-09-06.md`](handoff/HANDOFF-2026-09-06.md) | อ่านเฉพาะเมื่อลุงจืดสั่ง/ถามความเสี่ยง |
| **การเตรียมพร้อมกู้คืนระบบ (Disaster Recovery)** | [`docs/runbooks/RECOVERY-READINESS.md`](runbooks/RECOVERY-READINESS.md) | อ่านเฉพาะเมื่องานเกี่ยวกับ Backup/Restore |
| **ศึกษา Benchmark โปรแกรมบัญชี (FlowAccount / PEAK)** | [`docs/features-flowaccount-peak/README.md`](features-flowaccount-peak/README.md) | ไม่ต้องอ่านไฟล์โค้ด |

---

## โครงสร้างใน `docs/`

- [`kms/`](kms/README.md) — ฐานความรู้ระบบ (18 บทความ + ADR + Bugs) จัดทำแบบมี citation `path:line`
- [`features-flowaccount-peak/`](features-flowaccount-peak/README.md) — บทวิเคราะห์และเปรียบเทียบคุณสมบัติเชิงลึก FlowAccount vs PEAK Account
- [`skills/`](skills/) — Skill ส่วนตัวของลุงจืด (`ui-scale-polish`, `audit-mongomodel-sync`)
- [`handoff/`](handoff/) — รายงานส่งต่องานระหว่างเซสชัน
- [`reference/`](reference/CODE-MAP.md) — แผนที่ระบุบรรทัดของไฟล์ขนาดยักษ์
- [`runbooks/`](runbooks/RECOVERY-READINESS.md) — คู่มือเตรียมความพร้อมในการกู้คืนระบบ

---

## กฎเหล็ก: ความเร็วและสุขอนามัย Context (Speed & Context Hygiene)

1. **Surgical Read**: อ่านไฟล์ด้วย StartLine/EndLine เจาะจงเสมอ, ไฟล์ขนาดยักษ์ดูตำแหน่งจาก [`reference/CODE-MAP.md`](reference/CODE-MAP.md)
2. **Surgical Patch**: แก้ไขเฉพาะบล็อกที่เปลี่ยนด้วย targeted edit ห้าม rewrite ทั้งไฟล์
3. **Command Output Hygiene**: รันเทสต์เฉพาะไฟล์ที่แตะ (`npm test -- <path>.test.ts`), ใช้ `git status -s`, คุม output ไม่ให้พ่น log ยาว
4. **Context Isolation**: งานสำรวจกว้างขวางให้ใช้ Subagent แยกเพื่อไม่ให้ context หน้าต่างหลักบวม
5. **Prompt Caching Friendly**: ไม่แก้ไขสลับไปมาในโครงสร้าง system config บ่อย เพื่อรักษา cache hit rate 90%

