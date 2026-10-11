# ADR 2026-10-11: มิติวิเคราะห์ GL ครบ 6 มิติ ตาม Champ Parity

- **สถานะ:** ตกลงแล้วและนำขึ้น Production แล้ว (Release `r20261011-3`)
- **วันที่:** 2026-10-11
- **การตัดสินใจโดย:** ลุงจืด + Antigravity / Claude
- **เอกสารที่เกี่ยวข้อง:** `docs/kms/champ/design-2026-10-11/gl-dimensions.md`, `docs/kms/champ/09-gl.md`, `docs/kms/20-champ-parity-gap.md`

---

## 1. บริบท (Context)
ต้นแบบ Champ 5.5 Pro มีมิติวิเคราะห์ 7 ตัว (ไม่นับสาขาของบริษัทเหลือ 6 มิติ):
1. **DepartCode** (แผนก) — สูงสุด 10 ตัวอักษร
2. **ProjectCode** (โครงการ) — สูงสุด 10 ตัวอักษร
3. **AllocateCode** (การจัดสรร) — สูงสุด 10 ตัวอักษร
4. **JobCode** (งาน) — สูงสุด 15 ตัวอักษร (ในบรรทัด 20 ตัวอักษร)
5. **PartCode** (หน่วยงาน) — สูงสุด 10 ตัวอักษร
6. **SideCode** (ฝ่าย) — สูงสุด 10 ตัวอักษร

เดิมระบบ BC Ai Account มีเฉพาะ 2 มิติแรก (แผนก และ โครงการ) ทำให้ไม่ครบตาม Champ Parity และไม่สามารถรองรับการปันส่วนต้นทุน งบประมาณ หรือรายงานแยกมิติของธุรกิจขนาดกลางได้

---

## 2. การตัดสินใจเชิงสถาปัตยกรรม (Architectural Decisions)

1. **รวม 6 มิติในหน้าจอกำหนดรหัสต่างๆ จอเดียว (`/gl/dimensions`)**:
   - ใช้รูปแบบ Tab/Action จอเดียว (Champ Parity) เพื่อลดความซ้ำซ้อนของเมนู ไม่แตกเป็น 6 เมนูกระจัดกระจาย
   - ใช้รูปแบบ Master-Detail 2 คอลัมน์ (ซ้ายตารางรายการ ขวาฟอร์มแก้ไข) ตามมาตรฐาน `datacrud` skill

2. **ผังบัญชีกำหนดรหัสที่อนุญาตตามกฎ Champ ข้อ 5**:
   - ติ๊กเปิดธงมิติ + "เลือกทั้งหมด" (`_ALL`) = อนุญาตทุกรหัสของมิตินั้น
   - ติ๊กเปิดธงมิติ + เลือกรหัสเฉพาะ = อนุญาตเฉพาะรหัสที่ระบุ
   - ติ๊กเปิดธงมิติ + ไม่ได้เลือก "เลือกทั้งหมด" และไม่ได้เลือกรหัสใด = **ไม่อนุญาตรหัสใด** (`journal_line_<k>_not_allowed`)
   - ธงมิติในผังบัญชีจะถูก disable พร้อม tooltip เตือน หากสวิตช์มิติของบริษัทยังไม่เปิดใช้งานที่ค่าเริ่มต้นบริษัท (ตรงตาม `GLChartOfAccPage1DetailDlg.cpp`)

3. **Backend & BFF Integration (PostgreSQL Synchronous ACID)**:
   - บันทึกคอลัมน์ `allocatecode`, `jobcode`, `partcode`, `sidecode` ใน `gl_lines`, `gl_records`, และตารางงบประมาณ `gl_budgets`
   - เพิ่ม 4 รายงานทะเบียนมิติกลางใน Report Registry:
     - `gl_allocate_list` (5539) -> `/report/allocates`
     - `gl_job_list` (5540) -> `/report/jobs`
     - `gl_part_list` (5541) -> `/report/parts`
     - `gl_side_list` (5542) -> `/report/sides`
   - เพิ่มตัวกรองและเงื่อนไขการพิมพ์ 6 มิติในรายงานงบทดลอง, แยกประเภท, ข้อมูลรายวัน, และงบประมาณ

4. **TSV และภาษา**:
   - เพิ่มคีย์ภาษาและข้อความครบ 13 คอลัมน์ใน `backend/assets/language/languages.tsv`
   - ใช้ `def.separateLabel` ตรงตัว ป้องกัน key แตกต่างในรายงาน

---

## 3. ผลลัพธ์และหลักฐานการตรวจรับ (Evidence)
- ผ่านการทดสอบ `tools/verify.sh all` ครบ 100% (codemap, frontend 155 ไฟล์ / 1,700 เทสต์, Next.js build, Go backend, PostgreSQL 18)
- ผ่าน UAT อัตโนมัติ (Playwright + PostgreSQL) 1,160 / 1,160 checks
- Deploy ขึ้น Production สำเร็จที่ Release `r20261011-3`
- Playwright Prod Smoke Test ผ่าน 213/213 checks ยืนยัน 6 แท็บ และ 6 Checkbox เงื่อนไขมิติทำงานถูกต้อง
