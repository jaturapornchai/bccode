# BC Ai Account — Documentation Router

> **กฎประหยัด Context (On-Demand Loading)**:  
> AI และนักพัฒนา **ห้ามกวาดอ่านเอกสารทั้งหมดหรือเปิด HANDOFF ล่วงหน้า** เพราะจะสิ้นเปลือง Context Window มหาศาล  
> ให้เปิดอ่านเฉพาะไฟล์ที่ตรงกับขอบเขตงานจริงตามตารางด้านล่างเท่านั้น

---

## แผนที่เลือกอ่านตามประเภทงาน (On-Demand Selection)

| ประเภทงาน | สิ่งที่ต้องอ่าน (อ่านเฉพาะไฟล์นี้) | สิ่งที่ไม่ต้องอ่าน (ห้ามเปิด) |
|---|---|---|
| **งานเล็ก / แก้บั๊ก 1 บรรทัด / คำถามทั่วไป** | ดูโค้ดจริงโดยตรง (`code = truth`) | ไม่ต้องเปิด `docs/` ใด ๆ |
| **งานหน้าจอ / สไตล์ / CSS / UX** | [`docs/skills/ui-scale-polish/SKILL.md`](skills/ui-scale-polish/SKILL.md) | ไม่ต้องอ่าน `kms/` หรือ `handoff/` — **ยกเว้นงานเพิ่ม/แก้เมนูหลัก** ที่ SKILL.md §8 สั่งให้เปิด `kms/19-menu-coverage-market-standard.md` + ADR `kms/decisions/2026-09-08-menu-parity-market-standard.md` + `kms/decisions/2026-09-08-menu-parity-social-sweep.md` ด้วย |
| **งาน Schema / MongoDB / MongoModel** | [`docs/skills/audit-mongomodel-sync/SKILL.md`](skills/audit-mongomodel-sync/SKILL.md) | ไม่ต้องอ่านไฟล์ UI |
| **งานสถาปัตยกรรม / โดเมนเฉพาะเรื่อง** | ดูดัชนีใน [`docs/kms/README.md`](kms/README.md) แล้วเลือก **1 บทความที่ตรงเรื่อง** | ห้ามเปิดบทความอื่นที่ไม่ได้ทำ |
| **ประวัติการแก้ไขระบบ / สิ่งที่ทำไปแล้ว** | [`README.md`](../README.md) ในหัวข้อ Project Activity Log | ไม่ต้องเปิด handoff เก่าถ้าแค่อยากรู้ว่าล่าสุดทำอะไร |
| **ค้นหาตำแหน่งโค้ด / สัญญาณระบบ** | [`docs/kms/00-source-router.md`](kms/00-source-router.md) | ไม่ต้องเปิดบทความยาว |
| **นำทางไฟล์ขนาดใหญ่ (>1,000 บรรทัด)** | [`docs/reference/CODE-MAP.md`](reference/CODE-MAP.md) — auto-generated อาจเก่า: เทียบจำนวนบรรทัดในหัวข้อกับ `wc -l` จริงก่อนเชื่อเลขบรรทัด | ไม่ต้อง grep ทั้งไฟล์ (grep เฉพาะตอนยืนยันตำแหน่งฟังก์ชันเมื่อ CODE-MAP ไม่ตรง) |
| **สถานะงานค้าง / สรุปความเสี่ยงระบบ** | [`docs/handoff/HANDOFF-2026-09-19.md`](handoff/HANDOFF-2026-09-19.md) (ฉบับเดียว — handoff เดิมทั้งหมดลบแล้ว 2026-09-19; ของเก่าดูผ่าน `git show 1660b335:docs/handoff/<file>`) | อ่านเฉพาะเมื่อรับงานต่อจากเซสชันก่อน |
| **การเตรียมพร้อมกู้คืนระบบ (Disaster Recovery)** | [`docs/runbooks/RECOVERY-READINESS.md`](runbooks/RECOVERY-READINESS.md) | อ่านเฉพาะเมื่องานเกี่ยวกับ Backup/Restore |
| **งานเมนูหลัก / เพิ่มเมนูใหม่ / ขอบเขตผลิตภัณฑ์** | [`docs/skills/ui-scale-polish/SKILL.md`](skills/ui-scale-polish/SKILL.md) §8 + §8.1 — เพิ่ม 1 เมนูต้องแก้ครบ 4 จุดพร้อมกัน: `frontend/src/lib/menu-data.ts` → `backend/assets/language/languages.tsv` (13 คอลัมน์) → `frontend/src/lib/menu-icons.ts` → จำนวนใน `frontend/src/lib/menu-icons.test.ts` · ผลเทียบเคียงมาตรฐานที่ [`docs/kms/19-menu-coverage-market-standard.md`](kms/19-menu-coverage-market-standard.md) | ห้ามเพิ่มเมนูเงินเดือน / ภ.ง.ด.1 / ภ.ง.ด.1ก / ไฟล์นำส่งเงินสมทบประกันสังคม (ตัดออกจากขอบเขต 2026-09-08 — กฎใน `AGENTS.md` ชนะเสมอ) · ห้ามเขียนว่า "ครบ 100%" |

---

## โครงสร้างใน `docs/`

- [`kms/`](kms/README.md) — ฐานความรู้ระบบ (บทความหลัก `00`–`19` รวม 20 ไฟล์ + `architecture/` + ADR ใน `decisions/` + `bugs/` + `snippets/`) จัดทำแบบมี citation `path:line`
- [`skills/`](skills/) — Skill ส่วนตัวของลุงจืด (`ui-scale-polish`, `audit-mongomodel-sync`)
- [`handoff/`](handoff/) — รายงานส่งต่องาน **เก็บฉบับล่าสุดฉบับเดียว** (ลุงจืดสั่งลบของเก่า 2026-09-19)
- [`reference/`](reference/CODE-MAP.md) — แผนที่ระบุบรรทัดของไฟล์ขนาดยักษ์
- [`runbooks/`](runbooks/RECOVERY-READINESS.md) — คู่มือเตรียมความพร้อมในการกู้คืนระบบ

---

## กฎเหล็ก: ความเร็วและสุขอนามัย Context (Speed & Context Hygiene)

1. **Surgical Read**: อ่านไฟล์ด้วย StartLine/EndLine เจาะจงเสมอ, ไฟล์ขนาดยักษ์ดูตำแหน่งจาก [`reference/CODE-MAP.md`](reference/CODE-MAP.md) — ถ้าจำนวนบรรทัดใน CODE-MAP ไม่ตรงกับ `wc -l` จริง ให้ `grep -n` ยืนยันตำแหน่ง หรือ regenerate ด้วย `pwsh -NoProfile -File tools/gen-code-map.ps1` (git hook `.githooks/pre-commit` ตรวจให้ — **clone ใหม่ต้องสั่ง `npm run hooks:install` ครั้งหนึ่ง ไม่งั้นไม่มีอะไรตรวจ**; ไม่มี CI ฝั่ง GitHub แล้ว — ลบทิ้ง 2026-09-09 ตามมติ "GitHub เก็บ code อย่างเดียว" ชุดตรวจย้ายมาที่ `npm run verify` / `tools/verify.sh`)
2. **Surgical Patch**: แก้ไขเฉพาะบล็อกที่เปลี่ยนด้วย targeted edit ห้าม rewrite ทั้งไฟล์
3. **Command Output Hygiene**: รันเทสต์เฉพาะไฟล์ที่แตะ (`npm test -- <path>.test.ts`), ใช้ `git status -s`, คุม output ไม่ให้พ่น log ยาว
4. **Context Isolation**: งานสำรวจกว้างขวางให้ใช้ Subagent แยกเพื่อไม่ให้ context หน้าต่างหลักบวม
5. **Prompt Caching Friendly**: ไม่แก้ไขสลับไปมาในโครงสร้าง system config บ่อย เพื่อรักษา cache hit rate 90%
