# BC Ai Account — ระบบบัญชีและการเงินอัจฉริยะ

ระบบ ERP และบัญชีสำหรับธุรกิจไทย รองรับการทำงานแบบ Multi-Tenant (Holding / Company / Branch) ประมวลผลรวดเร็วด้วยสถาปัตยกรรมแบบ 2-Tier (MongoDB Storage + PostgreSQL Processing Engine) พร้อมการออกแบบ UX/UI ที่เป็นมิตรกับคนไทยอายุ 40+ ใช้งานง่าย ชัดเจน และปลอดภัย

---

## 📋 บันทึกประวัติการพัฒนาและแก้ไขระบบ (Project Activity Log)

> **กฎเหล็กของระบบ**: ทุกครั้งที่มีการแก้ไขโค้ด, เพิ่มฟีเจอร์, แก้บั๊ก, ปรับ UI หรือคอนฟิก **ต้องเพิ่มบันทึกรายการในส่วนนี้เสมอ** (เรียงลำดับจากล่าสุดอยู่บนสุด) และ commit ไปพร้อมกับโค้ดใน commit เดียวกันเสมอ

### [แม่แบบการบันทึก (Template)]
<!--
### YYYY-MM-DD — <หัวข้อการแก้ไขสั้นกระชับ>
- **ประเภท**: `[Feature]` / `[Fix]` / `[UI/UX]` / `[Refactor]` / `[Deploy]` / `[Docs]`
- **สิ่งที่ทำ**:
  1. <รายละเอียดภาษาไทยชัดเจน คนอายุ 40+ อ่านแล้วเข้าใจทันที>
- **ไฟล์สำคัญ**:
  - `<path/to/file>`
- **ผลการทดสอบ (Evidence)**:
  - <ผลการทดสอบ เช่น ผ่าน vitest ... tests, typecheck 0 errors, curl 200 OK>
-->

### 2026-09-09 — ตั้งกฎห้ามอ้างอิงบุคคลภายนอก (ซอฟต์แวร์คู่แข่ง) และชำระล้างเอกสารทั้งระบบ พร้อม Pre-commit Guard
- **ประเภท**: `[Docs]` `[Tooling]` `[Compliance]`
- **สิ่งที่ทำ**:
  1. **ตั้งกฎ Zero Reference Policy ใน AGENTS.md**: สั่งเด็ดขาดห้ามมีชื่อ ยี่ห้อ หรือการอ้างอิงถึงซอฟต์แวร์ภายนอกในโค้ด, คอมเมนต์, ชื่อไฟล์, ตัวแปร, หน้าจอ UI, commit message, และเอกสารทุกชนิด เพื่อป้องกันปัญหาลิขสิทธิ์และเครื่องหมายการค้า โดยให้ใช้คำกลาง ("มาตรฐานโปรแกรมบัญชีในตลาด") แทน
  2. **ลบโฟลเดอร์เอกสารวิจัยคู่แข่งเดิม**: ลบโฟลเดอร์เอกสารวิจัยเดิม 2 โฟลเดอร์และ handoff เก่า รวม 23 ไฟล์ออกจาก repository
  3. **ชำระล้างเอกสารทั้งระบบ**: เปลี่ยนชื่อไฟล์และปรับถ้อยคำใน `docs/kms/` (บทความ 19, ADRs, ดัชนี), `docs/README.md`, `docs/skills/ui-scale-polish/SKILL.md` และ `docs/handoff/` ให้เป็นคำกลางทั้งหมด
  4. **เพิ่มระบบตรวจจับอัตโนมัติ (Git Pre-commit Guard)**: อัปเดต `.githooks/pre-commit` ให้สแกนทุกไฟล์ที่ staged หากพบคำต้องห้ามจะสกัดและปฏิเสธ commit ทันที
- **ไฟล์สำคัญ**:
  - `AGENTS.md`
  - `.githooks/pre-commit`
  - `docs/kms/19-menu-coverage-market-standard.md`
  - `docs/kms/decisions/2026-09-08-menu-parity-market-standard.md`
  - `docs/README.md`
- **ผลการทดสอบ (Evidence)**:
  - สแกนทั้ง repository: ปลอดคำต้องห้าม 100%
  - ทดสอบ Pre-commit Guard: สกัดคำต้องห้ามสำเร็จทุกกรณี (`ExitCode: 1`)

### 2026-09-09 — ตั้งกฎและระบบ Activity Log ใน README.md พร้อม Git Pre-commit Hook
- **ประเภท**: `[Docs]` `[Tooling]`
- **สิ่งที่ทำ**:
  1. สร้างไฟล์ `README.md` ที่ root ของโปรเจกต์ เพื่อเป็นหน้าแรกของ Repository บน GitHub และเป็นจุดบันทึกประวัติงานหลัก
  2. เพิ่มหัวข้อกฎใน `AGENTS.md`: บังคับให้ AI ทุกตัว (Gemini, Claude, Codex) ต้องบันทึกสิ่งที่แก้ลงใน `README.md` และ commit พร้อมโค้ดทุกครั้ง
  3. เพิ่มตัวตรวจจับใน Git Pre-commit Hook (`.githooks/pre-commit`): หากมีการ stage โค้ดใน `frontend/src` หรือ `backend` แต่ไม่มี `README.md` ระบบจะแจ้งเตือนและปฏิเสธ commit เพื่อป้องกันการลืม
- **ไฟล์สำคัญ**:
  - `README.md`
  - `AGENTS.md`
  - `.githooks/pre-commit`
- **ผลการทดสอบ (Evidence)**:
  - ทดสอบ Pre-commit Hook ดักจับกรณีไม่มี README.md ได้ถูกต้อง
  - `npm run hooks:install` อัปเดต hook ลง `.git/hooks/` สำเร็จ

### 2026-09-09 — ย้าย "จัดหมวดสินค้า" ไปข้อมูลหลัก, บาร์โค้ดในหมวด, Tree View UX & Deploy Production
- **ประเภท**: `[UI/UX]` `[Feature]` `[Deploy]`
- **สิ่งที่ทำ**:
  1. **ย้ายเมนูจัดหมวดสินค้า**: ย้ายจากกลุ่ม "ตั้งค่าระบบ" (`/defaults`) ไปไว้ที่ "ข้อมูลหลัก › สินค้าและบาร์โค้ด" (`/product-category`) ต่อจากเมนูบาร์โค้ด
  2. **เปลี่ยนเป็นเพิ่มบาร์โค้ด**: ปรับระบบจัดการรายการในหมวดสินค้า จากเดิมที่เลือกสินค้า ให้เป็นการเลือกและค้นหา "บาร์โค้ด" เข้าหมวดแทน ผ่าน `POST /api/product-barcode/list`
  3. **ยกระดับ Tree View & Row Actions**: ปรับดีไซน์ Tree View ให้ตรงกับหน้า Group Tree View โดยมีปุ่มเลือกกลุ่มแบบเต็มจอ (Full-width CSS Grid) และมี Row Actions (แก้ไข, ลบ, จัดลำดับ, จัดการบาร์โค้ด) ประจำแถว
  4. **อัปเดต UI Skill**: บันทึกแบบแผน CSS Grid Full-width Selector และ Anti-pattern ลงใน `docs/skills/ui-scale-polish/SKILL.md` (§8.4)
  5. **Deploy ขึ้น Cloud Production**: Build frontend Docker image และ deploy ไปยังเซิร์ฟเวอร์ DigitalOcean `https://account.bcaicloud.com/` พร้อม push ขึ้น GitHub branch `dev`
- **ไฟล์สำคัญ**:
  - `frontend/src/lib/menu-data.ts`
  - `frontend/src/app/menu/product-category-screen.tsx`
  - `docs/skills/ui-scale-polish/SKILL.md`
- **ผลการทดสอบ (Evidence)**:
  - Frontend Typecheck: 0 errors
  - Vitest: 45 test files, 324/324 tests passed
  - Pre-push hook & Code map check: ผ่าน
  - Production Health Check: `https://account.bcaicloud.com/` ตอบ 200 OK, Google Sign-in ใช้งานได้ปกติ

---

## 📚 แผนที่เอกสารและการเรียนรู้ระบบ

โปรเจกต์นี้มีเอกสารและคลังความรู้ที่บันทึกไว้อย่างเป็นระบบในโฟลเดอร์ `docs/`:

- **[คู่มือการเลือกอ่านเอกสาร (`docs/README.md`)](docs/README.md)**: แผนที่ On-Demand Context สำหรับเลือกอ่านเอกสารเฉพาะที่ตรงกับงาน
- **[คลังความรู้ระบบ (`docs/kms/README.md`)](docs/kms/README.md)**: รวบรวมสถาปัตยกรรมระบบ 20 บทความ (`00`–`19`), การตัดสินใจทางเทคนิค (ADR), และประวัติบั๊ก
- **[ทักษะและมาตรฐาน UI/UX (`docs/skills/ui-scale-polish/SKILL.md`)](docs/skills/ui-scale-polish/SKILL.md)**: มาตรฐานการออกแบบสำหรับผู้ใช้คนไทยอายุ 40+, สี Palette, และแบบแผน UI
- **[มาตรฐานการจัดการฐานข้อมูล (`docs/skills/audit-mongomodel-sync/SKILL.md`)](docs/skills/audit-mongomodel-sync/SKILL.md)**: กฎการเชื่อมประสานระหว่าง MongoModel และ PostgreSQL

---

## 🛠️ คำสั่งที่ใช้บ่อยในการพัฒนา (Developer Commands)

```bash
# ติดตั้ง Git Hooks ประจำเครื่อง (ต้องรันหลังจาก clone หรือแก้ไข .githooks/)
npm run hooks:install

# รันโหมดพัฒนา Frontend (Next.js)
npm run dev:frontend

# ตรวจสอบความถูกต้องของโค้ดแบบเร็ว (Code map + Frontend lint/typecheck/vitest)
npm run verify

# ตรวจสอบความถูกต้องแบบเต็มระบบ (รวม Backend integration suites)
npm run verify:all

# อัปเดตแผนที่ระบุบรรทัดของไฟล์ขนาดใหญ่ (CODE-MAP)
npm run codemap
```
