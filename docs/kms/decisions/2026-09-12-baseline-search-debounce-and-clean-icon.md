---
date: 2026-09-12
status: implemented
tags: [bc-account, ui, baseline-search, debounce, clean-icon, general-ledger]
---

# ระบบค้นหาหลัก Baseline Toolbar — ค้นหาอัตโนมัติ (Auto 2s) และปุ่ม Clean ล้างคำค้น

## บริบทและความต้องการ
ลุงจืดได้ใช้ฟังก์ชัน Copy DOM คัดลอกแถบค้นหาของหน้าระบบบัญชีแยกประเภท:
```html
<form class="flex flex-wrap gap-2">
  <input class="..." aria-label="ค้นหารหัสหรือชื่อ" placeholder="ค้นหารหัสหรือชื่อ" value="">
  <button type="submit">ค้นหา</button>
  <button type="button">โหลดใหม่</button>
  <button type="button">เพิ่มรายการ</button>
  <button type="button">ขยายบรรทัด</button>
</form>
```
และสั่งการ:
> "baseline
> ตอนค้นหา ให้เป็น auto เลย รอ 2 วินาที ถ้าไม่กดเริ่มค้นหา และเพิ่ม icon clean ด้วย"

## การตัดสินใจและการออกแบบ (Architecture & UI Pattern)
1. **สร้าง Hook `useDebouncedSearch` (`frontend/src/app/gl/gl-common.tsx`)**:
   - จัดการ State คำค้นหา (`query`)
   - ตั้งค่าตัวหน่วงเวลาอัตโนมัติ `debounceMs = 2000` (2 วินาที)
   - หากผู้ใช้หยุดพิมพ์ครบ 2 วินาที จะกระตุ้นการค้นหาอัตโนมัติ (`onSearch(query)`)
   - หากผู้ใช้กดปุ่ม "ค้นหา" หรือกด Enter ก่อนครบ 2 วินาที ฟังก์ชัน `searchNow()` จะยกเลิกตัวหน่วงเวลาทันทีและค้นหาทันที
   - หากผู้ใช้กดปุ่ม Clean ฟังก์ชัน `clear()` จะยกเลิกตัวหน่วงเวลาทันที เคลียร์ค่าเป็นว่าง และสั่งค้นหาด้วยค่าว่างทันที
2. **สร้างคอมโพเนนต์ `SearchInput` (`frontend/src/app/gl/gl-common.tsx`)**:
   - ครอบ Input ด้วย Wrapper `relative flex items-center`
   - เมื่อมีข้อความ (`value.length > 0`) จะแสดงปุ่มไอคอน Clean (`X`) ที่ขอบขวา (`absolute right-2`)
   - กำหนด `pr-9` ให้กับ Input เพื่อไม่ให้ข้อความยาวทับกับปุ่ม Clean
   - รองรับปุ่ม `Escape` เพื่อเรียกคำสั่ง Clean ล้างคำค้นหา
   - กำหนด `tabIndex={-1}` ให้ปุ่ม Clean เพื่อรักษาคีย์บอร์ด Tab navigation ให้ราบรื่น
3. **การนำไปใช้งานใน Baseline Forms**:
   - `frontend/src/app/gl/gl-masters.tsx` (ผังบัญชี, ปีบัญชี, งวดบัญชี, งบประมาณ)
   - `frontend/src/app/gl/gl-journals.tsx` (สมุดรายวัน)
   - `frontend/src/app/gl/gl-statement-designer.tsx` (แม่แบบงบการเงิน)
4. **การทดสอบและการรักษาคุณภาพ**:
   - Unit Tests: `frontend/src/app/gl/search-input.test.ts` (5/5 tests PASS)
   - Full Vitest Suite: 57 test files / 419 tests (100% PASS)
   - TypeScript: 0 errors
   - Next.js Build: ผ่านสมบูรณ์ 37/37 routes
