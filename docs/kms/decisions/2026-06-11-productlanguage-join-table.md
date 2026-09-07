---
date: 2026-06-11
status: accepted
tags: [bc-account, postgres, read-model, i18n]
---

# pgsql เลือกชื่อตามภาษาด้วยตาราง productlanguage join แทน function picklangname

## Context
Product read model (PostgreSQL, per-holding tenant DB) เก็บชื่อหลายภาษาเป็น `names`/`unitnamesmain` jsonb (`[]NameX`) และเลือกชื่อตามภาษาใน select ด้วย function `picklangname(names, langcode, fallbacklangcode, defaultname)` (SQL IMMUTABLE, planner inline) ซึ่งต้องดูแลใน migration ของทุก tenant DB และอ่าน query ยาก

## Decision
ลุงจืดสั่ง (2026-06-11): สร้างตาราง `productlanguage` (1 แถวต่อ `productcode` + `langcode`, เก็บ `name` + `unitnamemain`) แล้วเลือกชื่อใน select ด้วย `LEFT JOIN` 2 ครั้ง (langcode + fallbacklangcode) + `COALESCE`; **ลบ `picklangname` ออกจาก pgsql ทั้งหมด** และตั้งกฎห้ามสร้าง function เลือกภาษาใน database
- FK `productcode` → `product.code` (immutable business code) `ON DELETE CASCADE`
- UNIQUE `(productcode, langcode)` รองรับ join ตรงๆ ไม่ต้องมี index เพิ่ม
- CHECK กันแถวที่ `name` และ `unitnamemain` ว่างทั้งคู่
- projection upsert `ON CONFLICT (productcode, langcode)` + delete ภาษาที่หายจาก source

## Alternatives
- คง `picklangname` — inline ได้เร็วจริง แต่ logic ซ่อนใน function + ต้อง migrate ทุก tenant DB
- `LEFT JOIN LATERAL` (เทียบเท่า function 100% รวม fallback "ภาษาแรกที่มีค่า") — ไม่เลือกเพราะซับซ้อนกว่า และ test guard เดิมห้าม `LEFT JOIN LATERAL` ใน doc นี้
- ลบ `names` jsonb ออกจาก product ด้วย — ไม่ทำ (scope); jsonb ยังใช้กับ namessearch (GIN trgm cross-language search) + raw fallback ให้ UI

## Consequences
- fallback chain เปลี่ยนจาก `lang → fallbacklang → ภาษาแรกที่มีค่า → code` เป็น `lang → fallbacklang → code` (tier-3 หาย; UI ใช้ raw `names` ชดเชยได้)
- ชื่อถูกเก็บ 2 ที่ใน pgsql (product.names jsonb + productlanguage) — projection ต้อง sync 2 ตาราง
- ClickHouse ยังมี `picklangname` (lambda UDF) ของ layer ตัวเอง — ยังไม่แตะ ถ้าจะ align ต้องตัดสินใจแยก
- verify แล้วบน postgres:16-alpine (DDL, upsert, fallback ทุก tier, CHECK, FK cascade ผ่านหมด); regression อยู่ใน `tests/ui-modernization.test.mjs` (test "PostgreSQL productlanguage read model documents per-language join table")
- ไฟล์: `pgsql/productlanguage-schema.html` (ใหม่), `pgsql/product-schema.html`, nav `index.html`/`src/pages/index.astro`, copy ใน `public/pgsql/` + `public/scripts/pgsql/`

โยง [[2026-06-10-product-readmodel-parity-taxtype-rename]]
