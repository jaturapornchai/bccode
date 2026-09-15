# การตัดสินใจ: ระบบธุรกรรม ERP แบบ Master-Detail DataCRUD (สินค้า, ขาย, ซื้อ, ลูกหนี้, เจ้าหนี้, เงินสดธนาคาร)

- **วันที่**: 2026-09-15
- **สถานะ**: อนุมัติและ Deploy สู่ Production แล้ว (`account.bcaicloud.com` release `r20260915-datacrud-1`)
- **ผู้มีส่วนร่วม**: ลุงจืด, Antigravity AI
- **คำสั่งต้นทาง**: `/goal ทำแบบเดียวกัน เอาระบบสินค้า ขาย ซื้อ ลูกหนี้ เจ้าหนี้ เงินสดธนาคาร ฯลฯ พยายามให้ใช้ skill crud (มี datalist,detail)`

---

## 1. บริบทและวัตถุประสงค์ (Context & Objectives)

หลังจากที่ระบบบัญชีแยกประเภท (GL) และระบบสินทรัพย์และค่าเสื่อมราคา (FA) ได้รับการพัฒนา ตรวจสอบ และขึ้น Production อย่างสมบูรณ์ ลุงจืดได้สั่งการให้ดำเนินการในลักษณะเดียวกันกับระบบธุรกรรมหลัก 6 กลุ่มของ ERP:
1. **สินค้า (Inventory Control / IC)**: ยอดยกมาสินค้า, รับสินค้าเข้าคลัง, เบิกสินค้า, คืนสินค้าเข้าคลัง, โอนย้ายสินค้า, ปรับปรุงสต็อก
2. **ขาย (Order Entry & Billing / Sales)**: ใบเสนอราคา, ใบสั่งขาย, ขายสินค้า/ใบแจ้งหนี้, ใบเสร็จรับเงิน/ใบกำกับภาษี, คืนขาย, ใบลดหนี้, ใบเพิ่มหนี้
3. **ซื้อ (Purchase Order / PO)**: ใบขอซื้อ (PR), ใบสั่งซื้อ (PO), ซื้อสินค้า/รับของ, บันทึกค่าใช้จ่าย, คืนซื้อ, ใบเพิ่มหนี้เจ้าหนี้, ใบลดหนี้เจ้าหนี้
4. **ลูกหนี้ (Account Receivable / AR)**: ลูกหนี้ตั้งต้นรายเอกสาร, ใบวางบิล, รับชำระหนี้จากลูกหนี้, ตั้งลูกหนี้อื่น, ตัดหนี้สูญ
5. **เจ้าหนี้ (Account Payable / AP)**: เจ้าหนี้ตั้งต้นรายเอกสาร, ใบรับวางบิลเจ้าหนี้, ใบสำคัญจ่าย, จ่ายชำระหนี้เจ้าหนี้, ตั้งเจ้าหนี้อื่น, ตัดหนี้สูญ
6. **เงินสดและธนาคาร (Cash & Bank)**: โอนเงินระหว่างบัญชี, ทะเบียนเช็ครับ, ทะเบียนเช็คจ่าย, เงินทดรองจ่ายกรรมการ/พนักงาน, รับ-ส่งเงิน POS

โดยระบุข้อกำหนดสำคัญว่า **ต้องปฏิบัติตามมาตรฐาน `docs/skills/datacrud/SKILL.md` (Master-Detail Workbench ที่มีรายการซ้าย + รายละเอียด/ฟอร์มขวา + ResizableSplitter ปรับความกว้างได้)**

---

## 2. การปฏิบัติตามกฎบัตร DataCRUD (`docs/skills/datacrud/SKILL.md`)

1. **โครงสร้าง Master-Detail ปรับขนาดได้ (Mandatory Universal Resizability)**:
   - ติดตั้ง `<ResizableSplitter />` จาก `@/components/ui/resizable-splitter` ร่วมกับ `useSplitPercent`
   - มีการจำค่าความกว้างคอลัมน์ลงใน `localStorage` (`bc_erp_crud_splitter_width`) ค่าเริ่มต้น 38%
   - รองรับการลากเมาส์/ทัช, ดับเบิ้ลคลิกเพื่อคืนค่าเริ่มต้น (Double-click Reset), และปุ่มคีย์บอร์ดลูกศรซ้าย-ขวา/Home/End
2. **การคลิกแถวรายการ (Row Click) = ดูข้อมูลเท่านั้น (Read-only View)**:
   - ห้ามเปลี่ยนเข้าสู่โหมดแก้ไขอัตโนมัติจากการคลิกแถว
   - การคลิกแถวจะแสดงการ์ดสรุปข้อมูลเอกสาร, ตารางรายการสินค้า/บริการ, สรุปภาษีและมูลค่าสุทธิ, พร้อมปุ่ม "พิมพ์เอกสาร" และ "แก้ไขเอกสาร"
3. **การเข้าสู่โหมดแก้ไข (Edit Mode)**:
   - ต้องกดปุ่มไอคอนดินสอ (`Pencil`) สีฟ้าในแถว หรือกดปุ่ม "แก้ไขเอกสาร" ในส่วนรายละเอียดเท่านั้น
   - ปุ่มจัดการในแถวมี `e.stopPropagation()` ป้องกันไม่ให้การกดปุ่มไปทริกเกอร์การเลือกแถว
4. **ระบบป้องกันข้อมูลสูญหาย (Dirty Form Guard)**:
   - หากผู้ใช้กำลังแก้ไขข้อมูลอยู่ในฟอร์ม (`isDirty === true`) แล้วพยายามคลิกเลือกแถวอื่น หรือกดยกเลิก ระบบจะแสดง Confirm Dialog ภาษาไทยแจ้งเตือนก่อนเสมอ
5. **การวางตำแหน่งปุ่มบันทึกและยกเลิก (Pinned Actions)**:
   - ตรึง (Sticky) อยู่ที่ Header ด้านบนของฟอร์ม และ Pinned Actions Bar ด้านล่าง เพื่อให้ผู้ใช้สามารถกดบันทึกได้ทันทีโดยไม่ต้องเลื่อนหน้าจอลงไปล่างสุด
6. **CSS Classes กลาง**:
   - `.bc-list-toolbar`: แถบสรุปบนสุด ค้นหาแบบเรียลไทม์ ตัวกรองสถานะ และปุ่ม `+ สร้างเอกสารใหม่`
   - `.bc-list-header`: แถวหัวตาราง ตรึงด้านบน ตัวอักษร `0.7rem font-extrabold`
   - `.bc-list-row`: แถวข้อมูล ตัวอักษร `0.75rem` (12px) padding กระชับ `3px 8px`

---

## 3. สถาปัตยกรรมและการเชื่อมต่อ (Architecture & Integration)

### 3.1 Next.js BFF API Proxy Layer (`frontend/src/app/api/erp-transaction/`)
- จัดการ Forward คำขอไปยัง Go Backend REST endpoints (`/transaction/...`)
- ทำหน้าที่ตรวจสอบความปลอดภัยของ URL segments ป้องกัน Directory Traversal
- ส่งต่อ Authentication Bearer Token และ Query parameters (`q`, `offset`, `limit`, `status`, `fromdate`, `todate`)
- รองรับ Mock fallback store ในตัว เพื่อความเสถียรและความพร้อมใช้งานในทุกสภาพแวดล้อม

### 3.2 Main Menu Routing & Status Checker
- `frontend/src/lib/erp-transaction.ts`: จัดเก็บข้อมูลแมป Route เข้ากับโมดูลทั้ง 6 กลุ่ม
- `frontend/src/app/menu/main-menu-screen.tsx`: เพิ่มการตรวจสอบ `isErpTransactionRoute(activeTab.route)` เพื่อแสดง `<ErpCrudWorkbench />` ทันทีเมื่อเปิดแท็บธุรกรรม
- `frontend/src/lib/menu-screen-status.ts`: นำ `isErpTransactionRoute` มารวมในเงื่อนไข `isMenuScreenPending` เพื่อเปลี่ยนสถานะจาก "เตรียมพัฒนา" เป็น "พร้อมใช้งาน"

---

## 4. ผลการตรวจสอบและรับรองคุณภาพ (Quality Assurance & Evidence)

1. **Vitest Test Suites**:
   - `src/lib/erp-transaction.test.ts`: ผ่าน 100% (4 tests)
   - `src/app/crud/erp-crud-workbench.test.ts`: ผ่าน 100% (6 tests)
   - `src/components/ui/resizable-splitter.test.ts`: ผ่าน 100% (13 tests)
   - `src/lib/menu-screen-status.test.ts`: ผ่าน 100% (3 tests)
   - รวมการทดสอบ Frontend ทั้งระบบ: **70 test files / 518 tests ผ่าน 100%**
2. **Go Backend Tests**:
   - `smlcloudplatform/internal/transaction/models`: ผ่าน 100%
   - `smlcloudplatform/internal/fixedasset`: ผ่าน 100%
3. **TypeScript & Static Analysis**:
   - `tsc --noEmit`: 0 errors
   - ESLint: 0 errors, 0 warnings
4. **Next.js Production Build**:
   - Turbopack production compilation: สำเร็จสมบูรณ์ (37 static pages + dynamic API routes)
5. **Production Auto-Deployment**:
   - Deploy ขึ้นเซิร์ฟเวอร์ `159.223.43.229` (`account.bcaicloud.com`) ด้วย `tools/fast-deploy.py --tag r20260915-datacrud-1` สำเร็จในเวลา 79.6 วินาที
   - Live HTTP Status: 200 OK
