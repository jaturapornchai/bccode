# Decision Record: Full-Height Auto-Expanded Workbench Standard (2026-09-14)

## Context
ลุงจืดสั่งการเรื่องความสูงของหน้าจอ Workbench ระบบบัญชีแยกประเภท (`/gl/*`):
> *"ความสูง ต้อง expanded ส่วนที่เหลือด้วย auto ด้วย"*

จากการตรวจสอบสภาพจริงบน Production:
- เมื่อเปิดหน้าจอผังบัญชี (`/gl/chartofaccounts`) หรือสมุดรายวัน (`/gl/journal/*`) บนจอขนาด 900px viewport:
  - คอนเทนเนอร์รายการ (`listPaneBox`) มีความสูงเพียง 270px
  - มีพื้นที่สีพื้นว่างเปล่าด้านล่าง (`spaceBelowList`) มากถึง 523.9px (กว่า 58% ของหน้าจอไม่ได้ถูกใช้งาน!)
  - ตัวตารางข้อมูลใช้คลาสล็อกความสูงสูงสุดคงที่ `max-h-[62vh]` หรือ `max-h-[65vh]`
  - ใน `main-menu-screen.tsx` แท็บคอนเทนเนอร์มีเงื่อนไข `activeTabNeedsFixedViewport` ซึ่งรองรับเฉพาะ `/product`, `/productbarcode`, `/datamodelgraph` แต่ตกหล่น `isGeneralLedgerRoute(activeWorkTab.route)` ส่งผลให้คอนเทนเนอร์ใช้ `lg:overflow-y-auto` แทนที่จะเป็น `lg:h-full lg:overflow-hidden`

## Decision
1. **Viewport Lock ใน `main-menu-screen.tsx`**:
   - เพิ่ม `isGeneralLedgerRoute(activeWorkTab.route)` เข้าใน `activeTabNeedsFixedViewport`
   - เพิ่ม `isGeneralLedgerRoute(tab.route)` ในคลาสของแท็บ `<section>` เพื่อส่งมอบ `lg:h-full lg:min-h-0 lg:overflow-hidden`
2. **ขยายความสูงของ `GeneralLedgerScreen`**:
   - ปรับ `<main className="gl-workbench ...">` ให้ใช้ `flex min-w-0 flex-1 flex-col gap-2` ร่วมกับ `h-full min-h-0 overflow-hidden` เมื่อ `embedded = true`
   - ห่อหุ้ม `{content}` ด้วย `<div className="flex-1 min-h-0 flex flex-col">{content}</div>`
3. **ขยาย `SplitWorkbench` ใน `gl-common.tsx`**:
   - คอนเทนเนอร์หลักใช้ `flex min-w-0 flex-1 flex-col gap-2 xl:flex-row xl:h-full xl:min-h-0`
   - ทั้ง 2 พาเนล (`data-gl-pane="list"` และ `data-gl-pane="editor"`) ใช้ `flex flex-col min-h-0` (ไม่ใส่ `h-full` เพราะจะไปปิดการทำงานของ `align-items: stretch` ตามธรรมชาติของ Flex row) เพื่อยืดเต็ม 100% ของความสูง row เสมอ
4. **ขยายตารางข้อมูลอัตโนมัติ (Auto-expanded Table)**:
   - เปลี่ยนจาก `max-h-[62vh]` / `max-h-[65vh]` เป็น `flex-1 min-h-[300px] overflow-auto rounded-xl border border-border`
   - ตรึง Search bar และ Stats ไว้บนสุด (`shrink-0`) และตรึง Pager ไว้ล่างสุด (`shrink-0`)
5. **ขยาย Editor Pane และ Scrollable Fieldset**:
   - ตรึง Header และ Footer ของพาเนลแก้ไขไว้บน-ล่าง (`shrink-0`)
   - ปรับ Fieldset / Lines ให้อยู่ในคอนเทนเนอร์เลื่อนแนวตั้งเฉพาะส่วน `flex-1 min-h-0 overflow-y-auto pr-1`
   - ปรับ Empty State ให้จัดกึ่งกลางพื้นที่พาเนลแนวตั้งด้วย `flex flex-col flex-1 h-full min-h-64 items-center justify-center`
6. **อัปเกรด Skill**:
   - บันทึกแบบแผนลงใน `.agents/skills/ui-scale-polish/SKILL.md` หัวข้อ 8.23 ตามกฎ Mandatory Skill Upgrade

## Files Changed
- `frontend/src/app/menu/main-menu-screen.tsx`
- `frontend/src/app/gl/general-ledger-screen.tsx`
- `frontend/src/app/gl/gl-common.tsx`
- `frontend/src/app/gl/gl-masters.tsx`
- `frontend/src/app/gl/gl-journals.tsx`
- `frontend/src/app/gl/gl-statement-designer.tsx`
- `frontend/src/app/gl/gl-reports.tsx`
- `.agents/skills/ui-scale-polish/SKILL.md`
