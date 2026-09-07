---
tags: [bug, bc-account, go, mongodb, warehouse, uat]
date: 2026-07-02
---

# zone/shelf sub-routes ของคลัง (Mongo-backed) — 4 bug จริงซ่อนกันเป็นชั้น พบระหว่าง UAT post-migration

Code path นี้ (`/warehouse/:guid/zone[...]`) frontend ไม่เคยเรียก เลยไม่เคยมีใครเจอ — บั๊กตัวหลังโผล่ได้ต่อเมื่อแก้ตัวหน้าก่อน (ชั้นที่ 4 เป็นของใหม่จากการ migrate เมื่อวาน).

## Bug 1 — POST zone panic (nil-deref)
Warehouse ที่สร้างโดยไม่มี `location` key เก็บ `location: null` ลง Mongo → `CreateLocation` ทำ `range *locations` → panic (ยืนยันจาก stack trace `docker logs mainapi`: warehouse_http_service.go:185).
**Fix (choke point เดียว):** `normalizeWarehouse()` — Location nil → `[]` ในทุก write path ของ service (CreateWarehouse / UpdateWarehouse / SaveInBatch ทั้ง create+update closure). DB disposable → wipe docs null เก่าแทน nil-guard 10 จุด.

## Bug 2 — POST shelf ตอบ "document not found" เสมอ
`CreateShelf` ใช้ `FindWarehouseByShelf` ที่ filter `location.shelf.code == <โค้ดที่กำลังสร้าง>` — shelf ใหม่ไม่มีทาง match. (ไม่เคยโผล่เพราะบั๊กพี่น้อง "CreateShelf ไม่เคย persist" บังไว้จนแก้เมื่อวาน)
**Fix:** เปลี่ยนเป็น `FindWarehouseByLocation` (dup-check ใน loop มีอยู่แล้ว).

## Bug 3 — Zone rename งอกซ้ำ + ทิ้ง entry เก่า
`UpdateLocation` same-warehouse rename branch loop หา `doc.Code` (โค้ดใหม่ที่เพิ่ง dup-check ว่าไม่มี!) แทน `locationCode` เดิม → append entry ใหม่เสมอ + entry เก่าค้างเป็น orphan.
**Fix:** rewrite branch — หาด้วยโค้ดเดิม, update in-place, `guidfixed` เดิมรอด.

## Bug 4 — Cross-warehouse move ตอบ 200 แต่ไม่ย้าย (regression จากการ migrate เอง)
HTTP handler (เขียนใหม่ตอน migrate) ทับ `req.WarehouseCode = warehouseInfo.Code` เสมอ → ทำลาย move target → เข้า same-warehouse branch ตลอด (ผลข้างเคียง: ชื่อ zone เปลี่ยนแต่ไม่ย้าย — 200 หลอกๆ).
**Fix:** default เฉพาะเมื่อว่าง (`if req.WarehouseCode == ""`) ทั้ง UpdateZone และ UpdateShelf (LocationCode ด้วย).

## เสริม (in-scope)
- zone/shelf ผ่าน sub-routes ได้ `guidfixed` จริงแล้ว (เดิม `""`) + **preserve ข้าม rename/move** (`movedShelf`/`movedLocation` carry min/max/productitems/dimensions ที่ request ไม่มี — เดิม rebuild จาก Code+Name = ล้างเงียบ)

## Regression test
`frontend/e2e/product-warehouse-crud.spec.ts` เพิ่ม test "warehouse zone/shelf sub-route API — create/rename/move/delete" (Playwright request context + token จาก UI login) cover ทั้ง 4 บั๊ก — ผ่าน 2 รอบติด, Mongo live leftovers = 0.

## บทเรียน
- Code path ที่ "มีอยู่แต่ไม่มีใครเรียก" = บั๊กสะสมเป็นชั้น — บั๊กแรกทำให้ path ตายทั้งเส้น บั๊กที่เหลือเลยไม่เคยถูกเจอ
- ตอน migrate/rewrite handler: field ที่ handler "เติมให้" จาก path param ต้อง default-เมื่อว่างเท่านั้น ห้ามทับของที่ client ส่งมา (ทำลาย semantics เช่น move)

related: [[2026-07-01-warehouse-mongodb-migration]], [[2026-07-01-warehouse-screen-3-bugs]]
