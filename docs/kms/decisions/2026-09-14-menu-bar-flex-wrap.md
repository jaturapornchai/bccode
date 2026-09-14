---
date: 2026-09-14
status: accepted
tags: [bc-account, frontend, menu, responsive, flex-wrap, ux-polish]
---

# ปรับแถบเมนูนำทางด้านบนให้ตัดขึ้นบรรทัดใหม่ (Flex Wrap) แทนการมีแถบเลื่อนแนวนอน (No Horizontal Scroll)

## Context

แถบเมนูนำทางด้านบน (Top Navigation Bar) ของระบบบัญชีเดิมใช้ CSS:
```html
<div class="flex min-w-0 gap-1 overflow-x-auto">
```
ซึ่งเมื่อหน้าจอมีขนาดเล็กกว่าความกว้างรวมของปุ่ม หรือเมื่อผู้ใช้ย่อหน้าต่างบราวเซอร์ ตัวบราวเซอร์จะแสดงแถบเลื่อนแนวนอน (Horizontal Scrollbar) ทำให้ผู้ใช้ต้องเลื่อนซ้าย-ขวาเพื่อกดปุ่มเมนูที่อยู่ด้านหลัง (เช่น บัญชีแยกประเภท หรือ ภาษีมูลค่าเพิ่ม)

ลุงจืดสั่งการ:
> *"ไม่ต้องการให้มีตัวเลื่อน ให้ wrap ลงมาเลย ถ้าล้น"*

## Decision

1. **ปรับปรุงแถบเมนูใน [`frontend/src/app/menu/main-menu-screen.tsx`](file:///d:/bccode/frontend/src/app/menu/main-menu-screen.tsx)**:
   - เปลี่ยนจาก:
     ```tsx
     <div className="flex min-w-0 gap-1 overflow-x-auto">
     ```
   - เป็น:
     ```tsx
     <div className="flex flex-wrap items-center gap-1">
     ```
   - ปลด `overflow-x-auto` ออกโดยสิ้นเชิง ไม่เกิดแถบเลื่อนแนวนอนอีกต่อไป
   - เมื่อความกว้างหน้าจอไม่พอ ปุ่มเมนูจะตัดขึ้นบรรทัดใหม่อย่างเป็นธรรมชาติ (`flex-wrap`)

2. **ปรับปรุงการคำนวณตำแหน่ง Popover เมนูย่อย (`openSectionLeft`)**:
   - เพิ่มการ Clamp ตำแหน่งแนวนอน `clampedLeft` ไม่ให้พาเนลเมนูย่อยหลุดออกนอกขอบขวาของจอ:
     ```tsx
     const containerWidth = menuRootRef.current?.clientWidth ?? window.innerWidth;
     const maxLeft = Math.max(0, containerWidth - 300);
     const clampedLeft = Math.min(Math.max(0, target.offsetLeft), maxLeft);
     setOpenSectionLeft(clampedLeft);
     ```
   - ทั้งใน `positionOpenSection` (hover/focus) และใน `onClick` ของปุ่มเมนู
   - พาเนลลอยจะแสดงผลอยู่ใต้แถบเมนูรวมทุกแถวเสมอ (`top-[calc(100%+6px)]` ของ `relative container`) ไม่ทับซ้อนปุ่มเมนู

3. **ผลการตรวจสอบ (Verification)**:
   - `npx tsc --noEmit` $\to$ 0 errors
   - `npm test` $\to$ 61 suites, 449 unit tests ผ่าน 100%
   - `npm run build` $\to$ ผ่านสมบูรณ์ 37/37 routes
   - Auto-deploy ขึ้น Production ([account.bcaicloud.com](https://account.bcaicloud.com/)) ด้วย Fast Streamed Zero-Disk Deployment ทันที (Release: `r20260914-menu-wrap-1`)

## Consequences

- ✅ ไม่มีแถบเลื่อนแนวนอนกวนสายตาหรือต้องเลื่อนซ้าย-ขวาอีกต่อไป
- ✅ ผู้ใช้บนหน้าจอขนาดต่างๆ หรือเปิดหน้าต่างแบบแบ่งครึ่งจอ (Split screen / iPad) มองเห็นปุ่มเมนูครบทุกโมดูลทันที
- ✅ ป๊อปอัปเมนูย่อยเปิดแสดงผลในตำแหน่งที่ถูกต้อง ไม่ล้นตกขอบจอ
