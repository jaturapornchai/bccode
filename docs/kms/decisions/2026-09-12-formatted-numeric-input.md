---
date: 2026-09-12
status: implemented
tags: [bc-account, ui, numeric-input, amount-input, formatting, decimal, comma, text-right, tabular-nums]
---

# ระบบจัดรูปแบบตัวเลขและจำนวนเงิน (Formatted Numeric Input Standard — Comma, Decimal, Right-Aligned & Clean Edit Mode)

## บริบทและความต้องการ
ลุงจืดได้ใช้ฟังก์ชัน Copy DOM คัดลอกช่องกรอกตัวเลขจำนวนเงินงบประมาณจากหน้าระบบบัญชีแยกประเภท:
```html
<input class="min-h-[2.6em] w-full rounded-xl border border-input bg-background px-3 py-1.5 text-[0.95rem] leading-normal text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-60" inputmode="decimal" required="" value="70000">
```
และสั่งการ:
> "ตัวเลขทั้งหมด ในระบบ ต้องมี ทศนิยม และ comma และต้องชิดขวา ยกเว้นต้อนเข้าไปแก้ไข ให้เป็น text ธรรมดา ยังไม่ต้องมี comma ลองคิดให้ด้วย"

## การตัดสินใจและการออกแบบ (Architecture & UI Pattern)
1. **การแสดงผลเมื่ออยู่นิ่ง / ไม่ได้โฟกัส (Idle / Blur Mode)**:
   - ตัวเลขทุกจำนวนในระบบต้องจัดรูปแบบให้มี Thousands Comma และทศนิยมครบถ้วนตาม Scale เสมอ (ค่าเริ่มต้น 2 ตำแหน่ง เช่น `70,000.00`, `1,234,567.89`)
   - ต้องจัดข้อความชิดขวาเสมอ (`text-right`) และระบุ `tabular-nums` เพื่อให้ความกว้างของตัวเลขทุกตัวเท่ากัน จัดแนวตรงจุดทศนิยมอย่างเป็นระเบียบ
2. **โหมดแก้ไขเมื่อโฟกัส (Focus / Edit Mode)**:
   - เมื่อผู้ใช้คลิกหรือแท็บเข้ามาในช่องกรอก ให้สลับการแสดงผลเป็นข้อความธรรมดาที่ไม่มี Comma (Plain Text without Commas) เช่น `70000.00` ทันที
   - ทำการเลือกข้อความทั้งหมด (`event.target.select()`) ทันที พร้อมดัก `onMouseUp` (`e.preventDefault()`) เพื่อไม่ให้การปล่อยเมาส์ล้างข้อความที่เลือก ทำให้ผู้ใช้สามารถพิมพ์ตัวเลขใหม่แทนที่ได้ทันที
   - ขณะพิมพ์ตัวเลข (Active Typing) ไม่แทรก Comma ในทุกคีย์สโตรก เพื่อป้องกันปัญหาเคอร์เซอร์กระโดดหรือเลื่อนตำแหน่งผิดพลาด
3. **การปรับข้อมูลเมื่อหลุดโฟกัส (Blur / Commit Mode)**:
   - เมื่อหลุดโฟกัส (`onBlur`) หรือกดปุ่ม Enter (`onKeyDown` Enter $\to$ `blur()`): ให้ตัดช่องว่าง, ปรับทศนิยมให้ครบตาม Scale ด้วยการปัดเศษครึ่งหนึ่งขึ้น (Half-up Rounding) ด้วย BigInt ไม่ให้สูญเสียความแม่นยำ, จัดรูปแบบด้วย Thousands Comma, และส่งค่า (Emit) ตัวเลขที่สะอาดกลับไปยัง State ของระบบ
4. **คอมโพเนนต์และการนำไปใช้งาน**:
   - `AmountInput`, `formatAmountValue`, `cleanAmountValue` ใน `frontend/src/app/gl/gl-common.tsx`
   - `NumericInput` ใน `frontend/src/components/ui/numeric-input.tsx`
   - นำไปใช้ใน:
     - งบประมาณและประมาณการกระแสเงินสด (`frontend/src/app/gl/gl-masters.tsx`)
     - รายการเดบิต / เครดิตในสมุดรายวันทั่วไป (`frontend/src/app/gl/gl-journals.tsx`)
     - ข้อมูลสินค้าและการกำหนดราคา (`frontend/src/app/menu/product-tab-shared.tsx`)
5. **การปรับปรุงเพิ่มเติม (2026-09-14: Focus-Zero-as-Empty & Leading Zero Normalization)**:
   - **ปัญหาที่พบจากผู้ใช้งาน (User Feedback)**: เมื่อเลื่อนเข้าช่องกรอกที่มีค่า `0` (เช่น ค่าเริ่มต้นของเดบิต/เครดิต หรือราคา) เคอร์เซอร์จะตกอยู่หน้าเลข 0 ทำให้พิมพ์ `1` แล้วกลายเป็น `10`
   - **ทางแก้ (Solution)**:
     * หากค่าเดิมเป็น 0 หรือว่าง ให้ตั้งข้อความตอนโฟกัสเป็นค่าว่าง `""` ทันที และอาศัย HTML `placeholder` แสดงตัวเลขจางๆ แทน
     * หากค่าเดิมไม่ใช่ 0 ให้เลือกข้อความทั้งหมดด้วย `requestAnimationFrame(() => target.select())` เพื่อให้พิมพ์แทนที่ได้ทันที
     * ป้องกันตัวเลขกลายพันธุ์ใน `onChange` ด้วยการตัดเลขศูนย์นำหน้าก่อนเลข 1–9 (`/^(-?)0+([1-9])/ -> "$1$2"`)
6. **การทดสอบและการรักษาคุณภาพ**:
   - Unit Tests: `frontend/src/app/gl/amount-input.test.ts` (17/17 tests PASS)
   - Unit Tests: `frontend/src/components/ui/numeric-input.test.ts` (6/6 tests PASS)
   - Playwright Browser Automation: `scratch/test_numeric_focus.cjs` & `scratch/test_numeric_replace.cjs` (PASS 100%)
   - Full Vitest Suite: 59 test files / 449 tests (100% PASS)
   - TypeScript: 0 errors
   - Next.js Build: ผ่านสมบูรณ์ 37/37 routes

