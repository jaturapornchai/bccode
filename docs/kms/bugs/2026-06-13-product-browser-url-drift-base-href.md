# Product Browser URL drift จาก base href + replaceState

#bc-account #frontend #astro

## Symptom
เปิด `/product` แล้วคลิกเอกสารไปเรื่อยๆ URL กลายเป็น `/product/#detail=...` (มี trailing slash) — พอ refresh เจอ Astro 404 "trailingSlash set to never"

## Root cause
Shell ฉีด `<base href="/product/">` เพื่อให้ relative link ใช้ได้ แต่ `history.replaceState(null, "", "#detail=...")` resolve fragment ผ่าน base → path เลื่อนจาก `/product` เป็น `/product/` เงียบๆ แล้วชนกับ `trailingSlash: "never"` ใน astro.config.mjs

## Fix
1. [[product browser]] shell ([D:\bccode-model\product\index.html]): ใส่ path เต็มใน replaceState — `${location.pathname}${location.search}#detail=...`
2. astro.config.mjs: `trailingSlash: "ignore"` รับทั้งสองแบบ (URL เก่าที่ user bookmark ไว้มี slash)

## Regression test
`tests/ui-modernization.test.mjs` pin โค้ดใหม่ของ shell; `tests/product-browser.test.mjs` กัน shell 3 สำเนา drift (product/index.html = product.astro = public/product/index.html)

> ⚠️ ตรวจ 2026-09-09: ไม่มีไฟล์ `tests/*.test.mjs` ใน repo นี้เลย — `git ls-files | grep test.mjs` ว่าง และ `git log --all --diff-filter=A --name-only -- '*.test.mjs'` ก็ไม่คืนอะไร (ไม่ใช่แค่ถูกลบ); `tests/` มีแต่ Playwright `.spec.ts` → regression guard ชุดนี้ **ไม่มีผลบังคับ** ถ้าจะกันซ้ำต้องเขียน test ใหม่
