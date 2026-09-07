---
tags: [bug, bc-account, frontend, product-category, uat]
date: 2026-07-01
---

# จอ "สินค้าในหมวด" (productcategorylist) — 2 bug จริง + 1 open finding พบระหว่าง UAT sweep เมนู 16 รายการ

## Bug 1 — `codelist` เป็น raw JSON textarea (ผิดกฎ "ห้ามมี {} []")
`system-setting-screens.ts`: field `codelist` ("รายการสินค้า") ใช้ `jsonField(...)` ไม่มี placeholder → fallback ไปเป็น raw JSON textarea ให้ user พิมพ์เอง `["CODE1","CODE2"]` — ผิดกฎที่ลุงจืดตั้งไว้ใน session นี้เอง ("หน้าจอ ห้ามมี {} [] ต้องออกแบบ ux/ui ให้ user ใช้ได้เลย") และผิด core-rules.md "No raw JSON user input rule" ด้วย.
**Fix:** เปลี่ยนเป็น `stringListField(...)` (chip/tag editor ที่มีอยู่แล้ว ใช้ตัวเดียวกับ "aliases" ของ productcolor/productsize) — reuse ของเดิม ไม่สร้าง component ใหม่.

## Bug 2 — empty-state text อ้างปุ่มที่ไม่มีจริง (dead UI reference)
`product-category-tree-view.tsx`: ข้อความ empty-state ("ไม่พบข้อมูลหมวดสินค้า...กดปุ่ม 'เพิ่มหมวดหลัก' ด้านบน") เป็น hardcode ไม่เช็ค `readOnly` prop — แต่ `productcategorylist` render component นี้ด้วย `readOnly={true}` (ปุ่ม "เพิ่มหมวดหลัก" ไม่ render บนจอนี้เลย เพราะ CRUD หมวดสินค้าทำที่จอ "จัดหมวดสินค้า" เท่านั้น) → user เจอกลุ่มว่างจะเห็นคำแนะนำให้กดปุ่มที่ไม่มีอยู่จริง.
**Fix:** แยกข้อความตาม `readOnly` — จอ CRUD จริง (productcategorygroupselectscreen) ยังคงข้อความเดิม, จอ read-only (productcategorylist) เปลี่ยนเป็นชี้ไปจอที่ถูกต้อง.

## Open finding — cross-screen data staleness (พบแต่ยังไม่ root-cause, ไม่ได้แก้)
สร้างหมวดสินค้าที่ `productcategorygroupselectscreen` กลุ่มเดียวกัน แล้วไปเปิด `productcategorylist` (full page reload ใหม่ทั้งหมด ไม่ใช่แค่ client nav) — **ไม่เห็นข้อมูลที่เพิ่งสร้าง แม้ poll นานถึง 30 วินาที**. ตรวจแล้ว: ข้อมูลอยู่ใน MongoDB ถูกต้อง (`appdb.productcategories`, ชื่อตรง, ไม่ deleted), proxy route (`route.ts`) ใช้ `cache: "no-store"` (ไม่ใช่ HTTP cache ชัดๆ), ทั้ง 2 slug สร้าง request ด้วย `group-number` filter เหมือนกันทุกประการ. **ยังไม่รู้ root cause จริง** — น่าจะเป็น React state/effect timing ใน `loadRecords`/`groupNumber` ของ `system-settings-screen.tsx`, หรือ backend scoping ที่ต่างกันโดยไม่เห็นจาก frontend code. ข้อมูลไม่หาย แค่จอ read ไม่ fresh.

## Verify
`frontend/e2e/product-category-list-crud.spec.ts` (2 test: empty-state text ถูกต้องบนจอ read-only + จอ CRUD เดิมยังทำงานปกติ) ผ่าน 2 รอบติด. tsc exit 0. Mongo cleanup ครบ (0 live test record เหลือใน productcategories/dimension/productcolors/brandproductmaster).

## Status
2 bug fix แล้ว + deploy local dev (HMR, next dev). Open finding ต้องตามต่อ session หน้า — ยังไม่กระทบข้อมูลจริง (แค่ freshness ของจออ่าน).

related: [[2026-06-30-dimension-items-json-object-vs-array]], [[2026-07-01-warehouse-screen-3-bugs]]
