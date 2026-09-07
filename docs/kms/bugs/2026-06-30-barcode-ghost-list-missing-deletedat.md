---
tags: [bug, bc-account, go, goapi, mongodb, cqrs]
date: 2026-06-30
---

# barcode ที่ลบแล้วยังโชว์ใน list (ghost) → ดูเหมือน barcode ซ้ำ

## Symptom
หน้าจอที่ list barcode (รวมจอ marketplace) โชว์ barcode ที่ **ลบไปแล้ว** ค้างอยู่ → เห็น barcode เดียวกันหลายแถว = ดูเหมือนซ้ำ. ทั้งที่ uniqueness ต่อ holding ทำงาน (create barcode เดิมซ้ำ → 400 "barcode is exists").

## Root cause
Frontend list (`/api/product-barcode/list`) เรียก **goapi** `BarcodeListHandler` (`backend/internal/goapi/handlers/barcode_list.go`), อ่าน MongoDB `productbarcodes` ด้วย filter `{holdingcode}` **เท่านั้น — ไม่มี `deletedat: {$exists:false}`**.
- DELETE = **soft delete** (set `deletedat`, ไม่ลบ doc).
- query ที่ไม่ filter deletedat → คืน doc ที่ลบแล้ว = **ghost**.
- uniqueness check บล็อกแค่ doc ที่ LIVE → ลบแล้ว create ใหม่ได้ → ghost(เก่า)+ตัวใหม่ อยู่ใน list พร้อมกัน = "ซ้ำ".
- mainapi `SearchRepository.FindStep` filter deletedat ถูกอยู่แล้ว; เฉพาะ goapi handler นี้ลืม (เป็น goapi ตัวเดียวที่อ่าน `productbarcodes` ตรงๆ).

## Fix
เพิ่ม `"deletedat": bson.M{"$exists": false}` เข้า filter (filter var เดียว ใช้ทั้ง find+count) ที่ `barcode_list.go`. Build+deploy local mainapi (`deploy-mainapi-fast.ps1`).

## Regression test (verified)
`scratch/stagehand-uat/mkt-ghost-check.js`: create barcode unique → list=1 → delete → **list=0** (ก่อน fix=1) → recreate barcode เดิม=201. Ghost หาย.

## Rule
ตั้งกฏ: ทุก read ของ collection ที่ soft-delete (`productbarcodes` ฯลฯ) ผ่าน raw bson ต้อง filter `deletedat:{$exists:false}` (find+count). ดู `.agents/rules/bc-account-core-rules.md` "Barcode / Soft-Delete Read Integrity".

related: [[2026-06-30-marketplace-channelmappings-field]]
