---
date: 2026-09-14
status: accepted
tags: [bc-account, ui, general-ledger, textbox, account-select, appearance-none, ui-polish]
---

# แก้ไขปัญหาไอคอนในช่องเลือกผังบัญชีซ้อนทับกัน (Fix AccountSelect Addon Icons Overlap)

## Context

จากการทดสอบระบบจริง ลุงจืดได้ส่งภาพถ่ายหน้าจอของช่องเลือกบัญชีแม่ (`AccountSelect`) ที่มีข้อความ `BM69-1000 · ข้อมูลตัวอย่าง - สินทรัพย์` และพบปัญหา:
> *"textbox icon ทับกัน แก้ให้ด้วย ให้สวยๆ"*

### สาเหตุของปัญหา (Root Cause)
1. **ลูกศร Dropdown ดั้งเดิมของบราวเซอร์แสดงผลซ้อนอยู่ใต้ปุ่มค้นหา**:
   - แท็ก `<select>` ไม่ได้กำหนดคลาส `appearance-none [&::-ms-expand]:hidden`
   - บราวเซอร์บน Windows (เช่น Chrome / Edge) จะวาดติ่งลูกศรดั้งเดิม (`▼`) ที่ขอบขวาของ `<select>` โดยอัตโนมัติ
   - เมื่อมีปุ่มค้นหาแบบ Custom วางอยู่ที่ `absolute right-1.5` ปุ่มค้นหาจึงไปทับอยู่บนติ่งลูกศรดังกล่าว ทำให้เห็นหัวลูกศรสีดำโผล่แลบออกมาจากใต้ปุ่มค้นหา
2. **ปุ่ม Clean (`X`) และปุ่มค้นหา (`🔍`) เบียดชิดกันเกินไป**:
   - ระยะห่างเดิมใช้ `gap-1` (4px) และตำแหน่ง `right-1.5` (6px) อยู่ชิดขอบโค้งมน `rounded-xl` เกินไป
3. **Padding ฝั่งขวาของกล่องข้อความไม่เพียงพอ**:
   - เดิมใช้ `!pr-14` (56px) ขณะที่ความกว้างรวมของทั้ง 2 ปุ่มคือ 66px ทำให้ข้อความยาววิ่งเข้าไปซ้อนทับอยู่ใต้ปุ่ม Clean

## Decision

1. **ปรับปรุง `AccountSelect` ใน [`frontend/src/app/gl/gl-common.tsx`](file:///d:/bccode/frontend/src/app/gl/gl-common.tsx)**:
   - เพิ่ม `appearance-none [&::-ms-expand]:hidden` ให้กับแท็ก `<select>` เพื่อลบลูกศรดั้งเดิมของบราวเซอร์ออก 100%
   - ขยาย Padding ข้อความฝั่งขวาเป็น `!pr-20` (80px) เมื่อมี 2 ปุ่ม และ `!pr-12` (48px) เมื่อมีปุ่มเดียว
   - ปรับการจัดวางปุ่มด้านขวาเป็น:
     ```tsx
     <div className="absolute right-2 top-1/2 -translate-y-1/2 z-10 flex items-center gap-1.5 pointer-events-auto">
       {allowEmpty && value && !disabled && (
         <button
           type="button"
           tabIndex={-1}
           onClick={(e) => {
             e.preventDefault();
             e.stopPropagation();
             onChange("");
           }}
           className="inline-flex !size-6 !min-h-0 !max-h-none !min-w-0 !p-0 items-center justify-center rounded-full text-muted-foreground/70 hover:bg-muted hover:text-foreground transition-colors cursor-pointer shrink-0"
           title="ล้างค่าที่เลือก"
           aria-label="ล้างค่าที่เลือก"
         >
           <X className="size-3.5" />
         </button>
       )}
       <button
         type="button"
         tabIndex={-1}
         disabled={disabled}
         onClick={(e) => {
           e.preventDefault();
           e.stopPropagation();
           setDialogOpen(true);
         }}
         className="inline-flex !size-7 !min-h-0 !max-h-none !min-w-0 !p-0 items-center justify-center rounded-[8px] border border-primary/20 bg-primary/10 hover:bg-primary/20 text-primary transition-all active:scale-95 disabled:pointer-events-none disabled:opacity-50 cursor-pointer shadow-2xs shrink-0"
         title="เปิดระบบค้นหาผังบัญชีแบบเต็มจอ (F2)"
         aria-label="ค้นหาผังบัญชี"
       >
         <Search className="size-3.5" />
       </button>
     </div>
     ```

2. **แก้ไขปัญหาไอคอนไม่กึ่งกลาง (Centering & Anti-Selector Clashing)**:
   - ตรวจพบว่า `button[class*="rounded-lg"]` ใน `frontend/src/app/globals.css:6507` บังคับสไตล์ระบบ `height: 36px !important; padding-inline: 0.7em !important;`
   - เมื่อปุ่ม `size-7` ใช้คลาส `rounded-lg` สไตล์นี้ทำให้ปุ่มยืดกลายเป็นวงรีสูง 36px และ `padding-inline: 0.7em` ผลักไอคอน SVG หลุดแนวแกนไปทางขวา 7.53px (Not Centered)
   - ปรับเปลี่ยนคลาสปุ่มมาใช้ `rounded-[8px]` ร่วมกับ `inline-flex !size-7 !min-h-0 !max-h-none !min-w-0 !p-0 items-center justify-center shrink-0` เพื่อหลบ global selector และรับประกันการจัดกึ่งกลางเรขาคณิตสมบูรณ์แบบ (Pixel-perfect center)
   - ปรับปรุงปุ่ม Clear ใน `SearchInput` เป็น `rounded-full inline-flex !size-7 !min-h-0 !max-h-none !min-w-0 !p-0 items-center justify-center shrink-0` เช่นเดียวกัน

3. **ผลการตรวจสอบ (Verification)**:
   - `npx tsc --noEmit` $\to$ 0 errors
   - `npm test` $\to$ 61 test suites, 449 unit tests ผ่าน 100%
   - `npm run build` $\to$ ผ่านสมบูรณ์ทั้ง 37 routes
   - Auto-deploy สู่ Production ([account.bcaicloud.com](https://account.bcaicloud.com/)) ด้วย Fast Streamed Zero-Disk Deployment (Release: `r20260914-center-icon-1`)
   - อัปเกรด Skill ส่วนที่ 8.22 ใน `.agents/skills/ui-scale-polish/SKILL.md`

## Consequences

- ✅ ไม่มีติ่งลูกศรดั้งเดิมของบราวเซอร์แลบออกมาซ้อนทับปุ่มค้นหาอีกต่อไป
- ✅ ปุ่ม Clean (`X`) และปุ่มค้นหา (`🔍`) มีระยะห่างที่สวยงาม ดูเป็นระเบียบ สะอาดตา พรีเมี่ยม
- ✅ ไอคอนค้นหาจัดวางกึ่งกลางเป๊ะทั้งแนวตั้งและแนวนอน (Mathematical Center 100%) ในกรอบสี่เหลี่ยมจัตุรัสโค้งมน ไม่บิดเบี้ยว
- ✅ ข้อความรหัสและชื่อบัญชียาวๆ จะถูกตัดทอน (`truncate`) ก่อนถึงปุ่ม ไม่ซ้อนทับหรือลอดใต้ปุ่ม
