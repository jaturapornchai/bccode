---
date: 2026-09-12
status: proposed
tags: [bc-account, ui, developer-experience, dev-dom-inspector, production]
---

# เปิดใช้งานวิดเจ็ต Copy DOM (DevDomInspector) บน Production (account.bcaicloud.com)

## บริบทและความเป็นมา
เดิมคอมโพเนนต์ `DevDomInspector` (`frontend/src/components/dev-dom-inspector.tsx`) มีเงื่อนไขตรวจสอบ:
```tsx
if (!mounted || process.env.NODE_ENV !== "development") {
  return null;
}
```
ทำให้วิดเจ็ตแสดงเฉพาะบนเครื่องคอมพิวเตอร์สำหรับการพัฒนา (Localhost) เท่านั้น และถูก Next.js build ถอดออก (dead code elimination) บนโหมด Production

ลุงจืดสั่งการเมื่อ 2026-09-12:
> "https://account.bcaicloud.com/
> ให้มี Copy DOM ด้วย เหมือนที่ run บน localhost"

## การตัดสินใจและการแก้ไข
1. **ปลดเงื่อนไข `NODE_ENV !== "development"`:**
   - แก้ไขให้คงไว้เฉพาะการตรวจสอบ `!mounted` สำหรับการป้องกัน hydration mismatch ใน Next.js Server Components
   - ทำให้ปุ่มลอย `Copy DOM [Alt+คลิก]` ทำงานบน Production (`account.bcaicloud.com`) ได้อย่างสมบูรณ์แบบ
2. **ความปลอดภัยและความเป็นส่วนตัว:**
   - วิดเจ็ตคัดลอกเฉพาะ DOM (`target.outerHTML`) ซึ่งเป็นข้อมูลฝั่ง Client ที่เบราว์เซอร์ดาวน์โหลดมาแสดงผลอยู่แล้ว (สามารถเปิดดูได้ผ่าน Developer Tools F12 อยู่แล้ว) ไม่มีการส่งข้อมูลใดๆ ออกนอกเครื่องของผู้ใช้
   - มีการครอบ Capture-phase click listener ด้วย `e.preventDefault()`, `e.stopPropagation()`, `e.stopImmediatePropagation()` ป้องกันการกดโดนลิงก์หรือปุ่มทำงานจริงระหว่างเลือกคัดลอก DOM
3. **การทดสอบและการรักษาคุณภาพ:**
   - เพิ่ม Unit Tests ใน `frontend/src/components/dev-dom-inspector.test.ts`
   - ผ่านการทดสอบ Vitest Suite (56 test files, 414 tests), TypeScript Typecheck, และ Next.js Production Build ครบถ้วน
   - เตรียม Docker Image: `bcai-account-frontend:r20260912-search-dom-1` พร้อม deploy
