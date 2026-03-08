# Manual Writing — กฏการเขียนคู่มือ

## หลักการ: คู่มือต้องอธิบายให้ละเอียด

คู่มือ (manual) คือสิ่งที่ user อ่านเพื่อใช้งาน → ต้องละเอียดพอที่ user เข้าใจได้โดยไม่ต้องถามใคร

## กฏการเขียนคู่มือ

### 1. ทุก field ต้องอธิบาย
- **ชื่อ field** + ความหมาย
- **ค่า default** (ถ้ามี)
- **ตัวเลือก/ช่วงค่า** ที่เป็นไปได้
- **ตัวอย่าง** การใช้งานจริง
- **ผลกระทบ** ถ้าตั้งค่าผิด

### 2. ต้อง cross-reference กับ source code
- ดู **Go backend model** → field ครบไหม
- ดู **Flutter frontend** → UI แสดง field อะไรบ้าง
- ดู **MongoDB data** → ค่าจริงเป็นยังไง
- ถ้า field มีใน source code แต่ไม่มีในคู่มือ → **ต้องเพิ่ม**

### 3. โครงสร้างคู่มือ
- แบ่ง section ชัดเจน ตาม feature group
- มี summary/overview ตอนต้น
- มี tips/คำแนะนำ ตอนท้าย
- ใช้ icon/emoji ช่วยจัดหมวด (เช่น ⚙️ 📋 💡)

### 4. ภาษาที่ใช้
- เขียนเป็น**ภาษาไทย** เข้าใจง่าย
- ศัพท์เทคนิคให้วงเล็บภาษาอังกฤษ เช่น "สกุลเงินหลัก (Base Currency)"
- น้ำเสียงเป็นมิตร ชัดเจน ไม่ใช้ศัพท์ยากเกินไป

### 5. Website manual page format (Next.js)
- ใช้ Tailwind CSS + responsive design
- มี breadcrumb navigation
- มี sidebar สำหรับ jump ไปแต่ละ section
- Screenshot/illustration ถ้ามี
- Update `lib/search-index.json` ให้ chatbot ค้นเจอ

## Checklist ก่อน publish คู่มือ
- [ ] ครบทุก field ที่มีใน backend/frontend
- [ ] ทุก field มีคำอธิบาย + ตัวอย่าง
- [ ] ค่า default ถูกต้อง (ตรง source code)
- [ ] ภาษาไทยอ่านเข้าใจง่าย
- [ ] Build ผ่าน ไม่มี error
