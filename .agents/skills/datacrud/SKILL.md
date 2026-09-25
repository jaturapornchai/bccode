---
name: datacrud
description: มาตรฐานการสร้างและปรับปรุงหน้าจอ CRUD และ Master-Detail ของระบบ BC Ai Account (คอลัมน์รายการซ้าย + รายละเอียด/ฟอร์มขวา ปรับขนาดได้)
---

# กฎบัตรและมาตรฐานหน้าจอ CRUD (Master-Detail Workbench)

## 1. วัตถุประสงค์และขอบเขต
- บังคับใช้กับทุกหน้าจอในระบบที่มีการแสดงรายการข้อมูล (Data List / Tree) และส่วนแสดงรายละเอียดหรือฟอร์มเพิ่ม/แก้ไขข้อมูล (Detail / Form Pane)
- ยึดหลัก **UX สำหรับคนไทยอายุ 40+** และ **ความพรีเมี่ยมเป็นหนึ่งเดียวทั้งระบบ** ตามข้อกำหนดใน `AGENTS.md`
- สอดคล้องกับมาตรฐานความหนาแน่นสูง (High Information Density) และดีไซน์ซิสเต็มใน `.agents/skills/ui-scale-polish/SKILL.md`

---

## 2. Shared CRUD Workbench Contract
- **โครงสร้างหลัก**: รายการซ้าย (List/Tree) + รายละเอียด/ฟอร์มขวา (Detail/Edit Pane) — **ห้ามเปิด Modal Popup ซ้อนเพื่อแก้ไขข้อมูลปกติ** (ยกเว้น Confirm Dialog หรือ Picker เฉพาะกิจ)
- **แหล่งข้อมูล (Data Provider) — PostgreSQL อย่างเดียว**:
  - จอเรียก BFF `/api/...` (Next route) → Go `mainapi` → PostgreSQL ผ่าน `sql.Tx` ทั้งอ่านและเขียน · business logic และการคำนวณอยู่ที่ backend ส่วนจอทำแค่แสดงผลและตรวจฟอร์มเบื้องต้น
  - ตัวอย่าง client: `fetchErpTransactions` / `saveErpTransaction` / `deleteErpTransaction` ใน `frontend/src/lib/erp-transaction.ts`
  - ถ้า endpoint ยังไม่มี (HTTP 404) ให้ขึ้น "จอนี้ยังไม่เปิดใช้งาน" (`module_not_available`, `erp-transaction.ts` ~บรรทัด 874) แทน error สีแดง เพราะผู้ใช้แก้เองไม่ได้
  - MongoDB / Kafka / Redis / ClickHouse ถอดออกแล้วเมื่อ 2026-09-23 ([ADR](../../../docs/kms/decisions/2026-09-23-remove-mongo-kafka-redis-clickhouse.md)) ห้ามออกแบบจอที่พึ่งพาของเหล่านี้
- **ความสดใหม่ของข้อมูล (Cache & Selection Freshness)**:
  - หลัง Create, Update หรือ Delete สำเร็จ ต้องรีเฟรช State ทันที ไม่ทิ้งรายการเก่าที่ค้าง หรือ id ที่ถูกลบไปแล้ว
  - การคลิกแถวรายการ (Row Click) ทำหน้าที่เพียง "เลือกดูรายละเอียด (Select & View Read-only Detail)" เท่านั้น **ห้ามเปลี่ยนเข้าโหมดแก้ไข (Edit Mode) อัตโนมัติจากการคลิกแถว** การแก้ไขต้องกดปุ่มไอคอนดินสอ (Pencil) หรือปุ่มแก้ไขอย่างชัดเจน

---

## 3. Layout & Resizing Contract (กฎการปรับความกว้าง — เลื่อนได้ทุกจอ)
1. **ต้องปรับความกว้างได้ทุกจอ (Mandatory Universal Resizability)**:
   - ทุกหน้าจอที่มี 2 คอลัมน์ (Master-Detail, Tree-Form, List-Detail) **ต้องติดตั้ง `<ResizableSplitter />` จาก `@/components/ui/resizable-splitter` เสมอ**
   - **ห้ามล็อกความกว้างตายตัว**: เช่น `w-80`, `w-[320px]`, `clamp(300px, 26vw, 380px)` หรือ fixed Tailwind grid โดยไม่มีตัวเลื่อนเด็ดขาด
2. **การโต้ตอบครบวงจร (Interaction & Controls)**:
   - **Mouse & Touch Dragging**: ลากเลื่อนปรับความกว้างได้อย่างนุ่มนวล มี Hitbox กว้างพอ (`w-3`) พร้อมคอร์เซอร์ `col-resize` และระงับ text selection ขณะลาก
   - **Double-click Reset**: ดับเบิ้ลคลิกที่ตัวจับเพื่อคืนค่าความกว้างเริ่มต้น (Default Width) ทันที
   - **Keyboard Accessibility**: รองรับปุ่ม `ArrowLeft` / `ArrowRight` (ปรับทีละ 1-2% หรือ 16px), `Home` (ไปจุดต่ำสุด min), `End` (ไปจุดสูงสุด max)
   - **Persistence**: บันทึกค่าความกว้างลงใน `localStorage` เสมอ (เช่น `bc_org_tree_sidebar_width`, `bc_product_set_sidebar_width`) เพื่อให้ผู้ใช้ไม่ต้องลากปรับใหม่ทุกครั้งที่เข้าจอ
   - **ARIA Attributes**: ต้องมี `role="separator"`, `aria-orientation="vertical"`, `aria-valuenow`, `aria-valuemin`, `aria-valuemax`, `tabIndex={0}`
   - **แบบแผนมาตรฐาน**: ใช้ hook `useSplitPercent({ storageKey, defaultLeft, min, max, containerRef })` คู่กับ `<ResizableSplitter>` (export ทั้งคู่จาก `@/components/ui/resizable-splitter`) — hook จัดการลาก คีย์บอร์ด รีเซ็ต และ `localStorage` ให้แล้ว ไม่ต้องเขียนเอง · ตัวอย่าง `frontend/src/app/crud/erp-crud-workbench.tsx` (hook ~บรรทัด 95, splitter ~658)
3. **การเลื่อนหน้าจอ (Scrolling & Height Management)**:
   - ฝั่งซ้ายและฝั่งขวาต้องมีอิสระในการ Scroll ของตนเอง (`overflow-y-auto`)
   - คำนวณความสูงให้พอดีกับ Viewport (`h-[calc(100dvh-...)]`) ป้องกันไม่ให้เกิด Scrollbar นอกของเบราว์เซอร์บน Desktop

---

## 4. มาตรฐานตารางและรายการ (List / Table Behavior)
- **CSS Classes กลาง**:
  - `.bc-list-toolbar`: แถบสรุปบนสุด (หัวข้อ + จำนวนรายการ)
  - `.bc-list-header`: แถวหัวตาราง ตรึงด้านบน (Sticky) ตัวอักษร `0.7rem` font-extrabold
  - `.bc-list-row`: แถวข้อมูล ตัวอักษร `0.75rem` padding กระชับ `3px 8px`
- **ความสม่ำเสมอของขนาดตัวอักษร**: คอลัมน์ในแถวต้องใช้ขนาดฟอนต์ `0.75rem` (12px) เท่ากัน ห้ามใส่ `text-sm` หรือ `text-base` เฉพาะคอลัมน์ใดคอลัมน์หนึ่งจนเสียจังหวะบรรทัด
- **Action Buttons**:
  - ปุ่มแก้ไข: ดินสอสีฟ้า/primary (`Pencil`)
  - ปุ่มลบ: ถังขยะสีแดง (`Trash2`)
  - ต้องมี `e.stopPropagation()` เพื่อไม่ให้การกดปุ่มแอ็กชันไปทริกเกอร์การเลือกแถว

---

## 5. มาตรฐานฟอร์มและฟิลด์ (Right Pane & Form Behavior)
- **ฟิลด์หลายภาษา (Multilingual Fields)**:
  - ใช้คอมโพเนนต์กลาง `<NamesEditor>` จาก `@/components/product-barcode/names-editor` เสมอ เพื่อแสดงภาษาตาม `settings.languageconfigs` พร้อมธงชาติ
- **การป้องกันข้อมูลสูญหาย (Dirty Form Guard)**:
  - หากผู้ใช้กำลังกรอกหรือแก้ไขข้อมูลอยู่ แล้วพยายามเปลี่ยนแถวหรือออกจากหน้าจอ ต้องมี Confirm Dialog ภาษาไทยแจ้งเตือนเสมอ (ห้ามใช้ `window.confirm`)
  - ใช้ `useConfirmDialog` จาก `@/components/ui/confirm-dialog` ส่งป้ายผ่าน `backendText(...)` · ตัวอย่าง dirty guard ตอนเปลี่ยนแถวใน `erp-crud-workbench.tsx` ~บรรทัด 127
- **การวางตำแหน่งปุ่มบันทึก (Actions Placement)**:
  - ปุ่มบันทึก (Save) และยกเลิก (Cancel) ต้องอยู่ในตำแหน่งที่เข้าถึงได้ทันที เช่น ตรึงที่ Header ด้านบน หรือ Pinned Footer ด้านล่าง ไม่บังคับให้ผู้ใช้ต้องเลื่อนลงไปล่างสุดของฟอร์มยาวๆ เพื่อกดบันทึก

---

## 6. รายชื่อหน้าจอที่ต้องมี `ResizableSplitter` (Current Roster)
1. [`warehouse-tree-view.tsx`](../../../frontend/src/app/system-settings/warehouse-tree-view.tsx) — ผังคลังสินค้า (ต้นแบบ)
2. [`system-settings-screen.tsx`](../../../frontend/src/app/system-settings/system-settings-screen.tsx) — SettingMasterDetail (หน้าจอตั้งค่าระบบทั่วไป)
3. [`system-settings-screen.tsx`](../../../frontend/src/app/system-settings/system-settings-screen.tsx) — BOM Editor (สูตรการผลิต)
4. [`system-settings-screen.tsx`](../../../frontend/src/app/system-settings/system-settings-screen.tsx) — ProductGroupTreeView & ProductSubgroupTreeView (กลุ่มสินค้าและกลุ่มย่อย)
5. [`system-settings-screen.tsx`](../../../frontend/src/app/system-settings/system-settings-screen.tsx) — ProductCategoryTreeView (หมวดหมู่สินค้า)
6. [`company-branch-tree-view.tsx`](../../../frontend/src/app/system-settings/company-branch-tree-view.tsx) — โครงสร้างองค์กร (บริษัทและสาขา)
7. [`product-screen.tsx`](../../../frontend/src/app/menu/product-screen.tsx) — ข้อมูลสินค้า
8. [`product-barcode-screen.tsx`](../../../frontend/src/app/menu/product-barcode-screen.tsx) — บาร์โค้ดสินค้า
9. [`product-set-screen.tsx`](../../../frontend/src/app/menu/product-set-screen.tsx) — สินค้าชุด (Product Bundles)
10. [`product-barcode-shelf-screen.tsx`](../../../frontend/src/app/menu/product-barcode-shelf-screen.tsx) — พิมพ์ป้ายสินค้า (Label Printing)
11. [`manage-shortcuts-screen.tsx`](../../../frontend/src/app/menu/manage-shortcuts-screen.tsx) — จัดการทางลัด (Shortcuts Catalog & My Shortcuts)
12. [`erp-crud-workbench.tsx`](../../../frontend/src/app/crud/erp-crud-workbench.tsx) — workbench กลางของเอกสาร ERP (เปิดจาก `main-menu-screen.tsx`) — ตัวอย่างอ้างอิงแบบแผน `useSplitPercent`
13. [`gl-common.tsx`](../../../frontend/src/app/gl/gl-common.tsx) — `SplitWorkbench` ของจอบัญชีแยกประเภท (GL) จำความกว้างใน `bc_gl_split_percent`
14. [`token-screen.tsx`](../../../frontend/src/app/mcp-tokens/token-screen.tsx) — จัดการ MCP token
