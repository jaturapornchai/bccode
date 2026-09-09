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

### 2026-09-09 — ลบกลุ่มเมนู "ร้านอาหาร/คาเฟ่" ทั้งหมด 7 เมนูออกจากระบบ
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**:
  1. **ลบกลุ่มเมนูร้านอาหาร/คาเฟ่ออกจากระบบ**:
     - ลบกลุ่มเมนู `ร้านอาหาร/คาเฟ่` (`restaurant-setup`) ทั้งหมด 7 เมนู ได้แก่ `โซน` (`/zonegroupselectscreen`), `โต๊ะ` (`/tablegroupselectscreen`), `ผังโต๊ะ` (`/tablemapgroupselectscreen`), `ครัว` (`/kitchengroupselectscreen`), `ตั้งค่าเครื่องสั่งอาหาร` (`/ordertemplatsetting`), `ตั้งค่าการสั่งอาหาร` (`/ordersetting`), `สั่งอาหารด้วย QR` (`/qrcodeordergroupselectscreen`) ออกจาก `frontend/src/lib/menu-data.ts`
     - ลบแมปปิ้งไอคอนทั้ง 7 เส้นทางใน `frontend/src/lib/menu-icons.ts`
     - อัปเดตจำนวนเมนูระบบใน `menu-icons.test.ts` จาก 213 เหลือ 206 เมนู
     - อัปเดต unit tests ใน `menu-data.test.ts` เอา `restaurant-setup` ออกจาก `groupOrder`
  2. **อัปเดต CODE-MAP**: ซิงค์แผนผังโค้ดระบบ `docs/reference/CODE-MAP.md` ให้ตรงกับขนาดและบรรทัดของไฟล์หลังตัดโค้ด
- **ไฟล์สำคัญ**:
  - `frontend/src/lib/menu-data.ts`
  - `frontend/src/lib/menu-icons.ts`
  - `frontend/src/lib/menu-icons.test.ts`
  - `frontend/src/lib/menu-data.test.ts`
  - `docs/reference/CODE-MAP.md`
- **ผลการทดสอบ (Evidence)**:
  - Unit tests: `menu-icons.test.ts`, `menu-data.test.ts`, `system-setting-screens.test.ts` ผ่าน 100% (32/32 tests)
  - Typecheck: `tsc --noEmit` ผ่าน 0 errors
  - Codemap check: `pwsh -NoProfile -File tools/gen-code-map.ps1 -Check` ซิงค์ถูกต้อง (45 files indexed)

### 2026-09-09 — ลบเมนู "รุ่นสินค้า" (Model) และตัดการเชื่อมโยงจากระบบอื่นอย่างสมบูรณ์
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**:
  1. **ลบเมนูและไอคอนรุ่นสินค้า**:
     - ลบรายการเมนู `รุ่นสินค้า` (`/mastermodelscreen`) ออกจากกลุ่มรายละเอียดประกอบสินค้าใน `frontend/src/lib/menu-data.ts`
     - ลบแมปปิ้งไอคอน `"/mastermodelscreen": "design"` ใน `frontend/src/lib/menu-icons.ts`
     - อัปเดตจำนวนเมนูระบบใน `menu-icons.test.ts` จาก 214 เหลือ 213 เมนู
  2. **ถอดคอนฟิกหน้าตั้งค่าระบบ**:
     - ลบคอนฟิก `master_model_screen` ออกจาก `frontend/src/lib/system-setting-screens.ts`
  3. **ตัดการเชื่อมโยงจากหน้าจอสินค้า (Product)**:
     - ลบฟิลด์เลือก `model` ออกจากแถบจัดหมวดหมู่สินค้าใน `frontend/src/app/menu/tab-product-classification.tsx`
     - ลบฟิลด์ `model` ออกจาก `classificationFields`, `clearFields` และการแสดงผลรายละเอียดสินค้าใน `frontend/src/app/menu/product-screen.tsx`
  4. **ตัดการเชื่อมโยง API Proxy Master Picker**:
     - ลบ endpoint mapping `model: "/aicloud/model"` ออกจาก `frontend/src/app/api/product-barcode/master/[master]/route.ts`
     - ลบ `| "model"` ออกจากประเภท `MasterName` ใน `frontend/src/lib/product-barcode/api.ts`
  5. **อัปเดต CODE-MAP**: ซิงค์แผนผังโค้ดระบบ `docs/reference/CODE-MAP.md` ให้ตรงกับขนาดและบรรทัดของไฟล์หลังตัดโค้ด
- **ไฟล์สำคัญ**:
  - `frontend/src/lib/menu-data.ts`
  - `frontend/src/lib/menu-icons.ts`
  - `frontend/src/lib/menu-icons.test.ts`
  - `frontend/src/lib/system-setting-screens.ts`
  - `frontend/src/app/menu/tab-product-classification.tsx`
  - `frontend/src/app/menu/product-screen.tsx`
  - `frontend/src/app/api/product-barcode/master/[master]/route.ts`
  - `frontend/src/lib/product-barcode/api.ts`
  - `docs/reference/CODE-MAP.md`
- **ผลการทดสอบ (Evidence)**:
  - Unit tests: `menu-icons.test.ts`, `menu-data.test.ts`, `system-setting-screens.test.ts` ผ่าน 100% (32/32 tests)
  - Typecheck: `tsc --noEmit` ผ่าน 0 errors
  - Codemap check: `pwsh -NoProfile -File tools/gen-code-map.ps1 -Check` ซิงค์ถูกต้อง (45 files indexed)

### 2026-09-09 — ลบเมนู "คุณลักษณะสินค้า" และกลุ่มเมนู "สี ไซซ์ และตัวเลือก" ออกจากระบบ
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**:
  1. **ลบเมนูและกลุ่มเมนูออกจากระบบ**:
     - ลบรายการเมนู `คุณลักษณะสินค้า` (`/mastercategoryscreen`) ออกจากกลุ่มรายละเอียดประกอบสินค้าใน `frontend/src/lib/menu-data.ts`
     - ลบกลุ่มเมนู `สี ไซซ์ และตัวเลือก` (`product-sku-options`) ทั้งกลุ่ม ซึ่งประกอบด้วย `สีสินค้า` (`/productcolor`), `ไซซ์/ขนาดสินค้า` (`/productsize`), `ชุดตัวเลือกสินค้า` (`/productvariantmatrix`) ออกจาก `frontend/src/lib/menu-data.ts`
     - ลบแมปปิ้งไอคอนทั้ง 4 เส้นทางใน `frontend/src/lib/menu-icons.ts`
     - อัปเดตจำนวนเมนูระบบใน `menu-icons.test.ts` จาก 218 เหลือ 214 เมนู
     - อัปเดต unit tests ใน `menu-data.test.ts` ให้สอดคล้องกับโครงสร้างเมนูใหม่
  2. **ถอดคอนฟิกหน้าตั้งค่าระบบ**:
     - ลบคอนฟิก `master_category_screen` ออกจาก `frontend/src/lib/system-setting-screens.ts`
  3. **อัปเดต CODE-MAP**: ซิงค์แผนผังโค้ดระบบ `docs/reference/CODE-MAP.md` ให้ตรงกับขนาดและบรรทัดของไฟล์หลังตัดโค้ด
- **ไฟล์สำคัญ**:
  - `frontend/src/lib/menu-data.ts`
  - `frontend/src/lib/menu-icons.ts`
  - `frontend/src/lib/menu-icons.test.ts`
  - `frontend/src/lib/menu-data.test.ts`
  - `frontend/src/lib/system-setting-screens.ts`
  - `docs/reference/CODE-MAP.md`
- **ผลการทดสอบ (Evidence)**:
  - Unit tests: `menu-icons.test.ts`, `menu-data.test.ts`, `system-setting-screens.test.ts` ผ่าน 100% (32/32 tests)
  - Typecheck: `tsc --noEmit` ผ่าน 0 errors
  - Codemap check: `pwsh -NoProfile -File tools/gen-code-map.ps1 -Check` ซิงค์ถูกต้อง (45 files indexed)

### 2026-09-09 — ลบเมนู "รูปทรงสินค้า", "ระดับสินค้า", "เกรดสินค้า", "มิติสินค้า" และตัดการเชื่อมโยงจากระบบอื่นอย่างสมบูรณ์
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**:
  1. **ลบ 4 เมนูและไอคอนออกจากระบบ**:
     - ลบรายการเมนู `ขนาด/มิติสินค้า` (`/productdimension`), `เกรดสินค้า` (`/mastergradescreen`), `ระดับสินค้า` (`/masterclassscreen`), `รูปทรงสินค้า` (`/masterdesignscreen`) ออกจากกลุ่มข้อมูลหลักใน `frontend/src/lib/menu-data.ts`
     - ลบแมปปิ้งไอคอนทั้ง 4 เส้นทางใน `frontend/src/lib/menu-icons.ts`
     - อัปเดตจำนวนเมนูระบบใน `menu-icons.test.ts` จาก 222 เหลือ 218 เมนู
  2. **ถอดคอนฟิกหน้าตั้งค่าระบบ**:
     - ลบคอนฟิก `productdimension`, `master_class_screen`, `master_design_screen`, `master_grade_screen` ออกจาก `frontend/src/lib/system-setting-screens.ts`
  3. **ตัดการเชื่อมโยงจากหน้าจอสินค้า (Product)**:
     - ลบฟิลด์เลือก `class`, `design`, `grade` ออกจากแถบจัดหมวดหมู่สินค้าใน `frontend/src/app/menu/tab-product-classification.tsx`
     - ลบฟิลด์ `class`, `design`, `grade` ออกจาก `classificationFields`, `clearFields` และการแสดงผลรายละเอียดสินค้าใน `frontend/src/app/menu/product-screen.tsx`
  4. **ตัดการเชื่อมโยง API Proxy Master Picker**:
     - ลบ endpoint mapping `class`, `design`, `grade` ออกจาก `frontend/src/app/api/product-barcode/master/[master]/route.ts`
     - ลบ `| "class" | "design" | "grade"` ออกจากประเภท `MasterName` ใน `frontend/src/lib/product-barcode/api.ts`
  5. **อัปเดต CODE-MAP**: ซิงค์แผนผังโค้ดระบบ `docs/reference/CODE-MAP.md` ให้ตรงกับขนาดและบรรทัดของไฟล์หลังตัดโค้ด
- **ไฟล์สำคัญ**:
  - `frontend/src/lib/menu-data.ts`
  - `frontend/src/lib/menu-icons.ts`
  - `frontend/src/lib/menu-icons.test.ts`
  - `frontend/src/lib/system-setting-screens.ts`
  - `frontend/src/app/menu/tab-product-classification.tsx`
  - `frontend/src/app/menu/product-screen.tsx`
  - `frontend/src/app/api/product-barcode/master/[master]/route.ts`
  - `frontend/src/lib/product-barcode/api.ts`
  - `docs/reference/CODE-MAP.md`
- **ผลการทดสอบ (Evidence)**:
  - Unit tests: `menu-icons.test.ts`, `menu-data.test.ts`, `system-setting-screens.test.ts` ผ่าน 100% (33/33 tests)
  - Typecheck: `tsc --noEmit` ผ่าน 0 errors
  - Codemap check: `pwsh -NoProfile -File tools/gen-code-map.ps1 -Check` ซิงค์ถูกต้อง (45 files indexed)

### 2026-09-09 — ลบเมนู "รูปแบบสินค้า" และตัดการเชื่อมโยงจากระบบอื่นอย่างสมบูรณ์
- **ประเภท**: `[UI/UX]` `[Cleanup]`
- **สิ่งที่ทำ**:
  1. **ลบเมนูและไอคอนรูปแบบสินค้า**: ลบรายการเมนู `รูปแบบสินค้า` (`/masterpatternscreen`) ออกจากกลุ่มข้อมูลหลักใน `frontend/src/lib/menu-data.ts`, ลบไอคอนใน `frontend/src/lib/menu-icons.ts` และอัปเดตจำนวนเมนูใน `menu-icons.test.ts` จาก 223 เหลือ 222 เมนู
  2. **ถอดคอนฟิกหน้าตั้งค่าระบบ**: ลบคอนฟิก `master_pattern_screen` ออกจาก `frontend/src/lib/system-setting-screens.ts` ไม่ให้เข้าถึงผ่าน `/[systemSetting]`
  3. **ตัดการเชื่อมโยงจากหน้าสินค้า (Product)**:
     - ลบช่องเลือก `pattern` ออกจากแถบจัดหมวดหมู่สินค้าใน `frontend/src/app/menu/tab-product-classification.tsx`
     - ลบฟิลด์ `pattern` ออกจาก `classificationFields`, `clearFields` และการแสดงผลรายละเอียดสินค้าใน `frontend/src/app/menu/product-screen.tsx`
  4. **ตัดการเชื่อมโยง API Proxy Master Picker**:
     - ลบ endpoint mapping `pattern: "/aicloud/pattern"` ออกจาก `frontend/src/app/api/product-barcode/master/[master]/route.ts`
     - ลบ `| "pattern"` ออกจากประเภท `MasterName` ใน `frontend/src/lib/product-barcode/api.ts`
  5. **อัปเดต CODE-MAP**: ซิงค์แผนผังโค้ดระบบ `docs/reference/CODE-MAP.md` ให้ตรงกับขนาดและบรรทัดของไฟล์หลังตัดโค้ด
- **ไฟล์สำคัญ**:
  - `frontend/src/lib/menu-data.ts`
  - `frontend/src/lib/menu-icons.ts`
  - `frontend/src/lib/menu-icons.test.ts`
  - `frontend/src/lib/system-setting-screens.ts`
  - `frontend/src/app/menu/tab-product-classification.tsx`
  - `frontend/src/app/menu/product-screen.tsx`
  - `frontend/src/app/api/product-barcode/master/[master]/route.ts`
  - `frontend/src/lib/product-barcode/api.ts`
  - `docs/reference/CODE-MAP.md`
- **ผลการทดสอบ (Evidence)**:
  - Unit tests: `menu-icons.test.ts`, `menu-data.test.ts`, `system-setting-screens.test.ts` ผ่าน 100% (33/33 tests)
  - Typecheck: `tsc --noEmit` ผ่าน 0 errors
  - Codemap check: `pwsh -NoProfile -File tools/gen-code-map.ps1 -Check` ซิงค์ถูกต้อง (45 files indexed)

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
