---
date: 2026-09-12
status: implemented
tags: [bc-account, general-ledger, gl-masters, crud, table, ux, ui-scale-polish]
---

# สถาปัตยกรรมตารางข้อมูลหลักระบบบัญชีมาตรฐาน CRUD เทียบเท่าหน้าจอตั้งค่าพื้นฐาน (GL Masters CRUD Table Parity)

## บริบทและโจทย์ความต้องการ (Context & Requirement)

ลุงจืดได้ส่ง HTML DOM ของตารางรายการในหน้าจอข้อมูลหลักระบบบัญชี (`GLMasters`) เช่น งบประมาณ (`/gl/budget`):
```html
<div class="max-h-[62vh] overflow-auto rounded-xl border border-border" aria-busy="false"><table class="w-full text-left text-[0.95rem] leading-normal [&_td]:py-1 [&_td_button]:min-h-8"><thead class="sticky top-0 bg-muted"><tr><th class="p-2">รหัส</th><th class="p-2">ชื่อ / รายละเอียด</th><th class="p-2">สถานะ</th></tr></thead>...
```
พร้อมคำสั่ง:
> **"พวกนี้ให้เหมือน หน้าจอ แรกๆ เช่น หน่วยนับ ใช้ CRUD"**

เดิมทีหน้าจอ `GLMasters` แสดงผลตารางแบบ Minimalist มีเพียง 3 คอลัมน์ (รหัส, ชื่อ/รายละเอียด, สถานะ) โดยไม่มีปุ่มลบหรือแก้ไขในแต่ละแถว ผู้ใช้ต้องคลิกเปิดเข้าไปดูในแบบฟอร์มด้านขวาเท่านั้นถึงจะเห็นปุ่มลบ และจำนวนเงินในงบประมาณถูกจับไปต่อท้ายสตริงชื่อรายการ (เช่น `... · 100,000.00`) ทำให้อ่านยากและไม่สอดคล้องกับพฤติกรรมในหน้าจอตั้งค่าพื้นฐานของระบบ เช่น หน่วยนับสินค้า (`/productunit` ใน `SettingDataList`) ที่มีคอลัมน์ "จัดการ" พร้อมปุ่มแก้ไข (Pencil) และลบ (Trash2) ในทุกบรรทัด

## การเปลี่ยนแปลงเชิงสถาปัตยกรรมและการปรับปรุง (Implementation Details)

ครอบคลุม 7 หน้าจอที่ใช้งานคอมโพเนนต์ร่วม `GLMasters` (`frontend/src/app/gl/gl-masters.tsx`):
1. `/gl/chartofaccounts` (ผังบัญชี - แท็บ accounts และ fiscal-years)
2. `/gl/budget` (งบประมาณ - budgets)
3. `/gl/periodlock` (ล็อกงวดบัญชี - periods)
4. `/gl/account-groups` (กลุ่มบัญชี - account-groups)
5. `/gl/account-mapping` (เชื่อมโยงผังบัญชี - mappings)
6. `/gl/product-account-groups` (กลุ่มบัญชีสินค้า - product-account-groups)
7. `/report/cashflowforecast` (ประมาณการกระแสเงินสด - forecast)

### 1. เพิ่มคอลัมน์ "จัดการ" (Actions Column) ประจำทุกแถว
- เพิ่มส่วนหัว `<th>จัดการ</th>` จัดชิดขวา
- แต่ละแถวมีปุ่มจัดการ 2 ตัวตามมาตรฐานระบบ:
  - **ปุ่มแก้ไข (Pencil)**: `size-7 rounded-md bg-background text-primary border-primary/30 hover:bg-primary/15 hover:border-primary/50` พร้อม `title="แก้ไข (Edit)"`
  - **ปุ่มลบ (Trash2)**: `size-7 rounded-md bg-background text-red-600 border-red-200/70 hover:bg-red-50 hover:border-red-300` พร้อม `title="ลบ (Delete)"`
- เมื่อคลิกปุ่มลบในแถว จะเรียก `deleteItem(item)` ซึ่งเรียกใช้ `useConfirmDialog` (Danger tone) ยืนยันการลบทันทีโดยไม่ต้องเปิดฟอร์มด้านข้าง
- หากลบผังบัญชี (`accounts`) จะมีข้อความเตือนเด็ดขาดห้ามลบผังบัญชีที่มีการลงรายการแล้ว
- ใช้ `event.stopPropagation()` เพื่อไม่ให้ทริกเกอร์ row click เมื่อกดปุ่มจัดการ

### 2. แยกคอลัมน์ "จำนวนเงิน" (Dedicated Amount Column)
- สำหรับทรัพยากร `budgets` และ `forecast` แยกจำนวนเงินออกจากชื่อรายการเป็นคอลัมน์เฉพาะ
- แสดงผลด้วยฟอนต์ตัวเลขจัดชิดขวา: `text-right font-mono font-semibold tabular-nums whitespace-nowrap`
- จัดรูปแบบทศนิยมและคอมม่าคั่นหลักพันด้วย `formatAmount(item.amount)`

### 3. ป้ายสถานะแบบมีสีสื่อความหมาย (Color-Coded Status Badges)
- ปรับเปลี่ยนจากข้อความธรรมดาเป็น Badge ตามสไตล์ระบบ:
  - `ใช้งาน`: ป้ายสีเขียวมรกต (`bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-500/20`)
  - `ล็อกแล้ว`: ป้ายสีเหลืองอำพัน (`bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-500/20`)
  - `ปิดปีแล้ว` / `ปิดใช้งาน`: ป้ายสีกลางอ่อน (`bg-muted text-muted-foreground border border-border`)

### 4. ปรับปรุงการเลือกแถวและแถบข้อมูลสรุป
- ทั้งแถวสามารถคลิกเพื่อเปิดดูรายละเอียดได้ (`cursor-pointer`)
- แถวที่กำลังเลือกแสดง Ring ไฮไลต์ชัดเจน: `bg-primary/10 ring-1 ring-inset ring-primary/40 font-medium`
- ใต้แถบค้นหาแสดงแถบสรุปจำนวนรายการทั้งหมด: `{total.toLocaleString("th-TH")} รายการ`

### 5. ปรับปรุงแผงฟอร์มแก้ไข (Editor Workbench)
- Header มีไอคอน Pencil หรือ Plus ชัดเจน พร้อมชื่อรหัสรายการ
- มีป้ายเตือนเมื่อมีข้อมูลที่แก้ไขแต่ยังไม่บันทึก (`● มีการเปลี่ยนแปลงที่ยังไม่บันทึก`)
- มีปุ่มปิด (`X`) เพื่อปิดแผงแก้ไข
- Footer มีปุ่ม "บันทึกข้อมูล", "ยกเลิก", "ล็อกงวด/ปลดล็อกงวด" (เฉพาะงวดบัญชี) และปุ่ม "ลบรายการนี้"
- เมื่อยังไม่มีการเลือกแถว แสดง Empty State ที่มีการ์ดและปุ่ม CTA `+ เพิ่มรายการใหม่`

## การตรวจสอบและหลักฐาน (Verification & Evidence)

1. **Unit Tests**:
   - `frontend/src/app/gl/gl-masters.test.ts` (4 ผ่านจาก 4 เทสต์ ครอบคลุมการแสดงผลคอลัมน์จัดการ, ปุ่มแก้ไข/ลบ, คอลัมน์จำนวนเงิน, ป้ายระดับผังบัญชี และป้ายสถานะ)
   - รวมการทดสอบทั้งระบบ 60 Test files / 446 ผ่านทั้งหมด
2. **Build Verification**:
   - `npm run build` สำเร็จ 100% ผ่านการตรวจ TypeScript โดยไม่มีข้อผิดพลาด
3. **Zero Regression**:
   - ไม่กระทบต่อ API หรือการทำงานของโมดูล GL เดิม
