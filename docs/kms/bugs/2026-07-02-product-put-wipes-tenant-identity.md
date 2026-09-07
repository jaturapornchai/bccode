---
tags: [bug, bc-account, go, mongodb, product, multi-tenant, uat]
date: 2026-07-02
---

# PUT /product/:guid ล้าง holdingcode + guidfixed — สินค้าหลุดจาก tenant ทันทีที่ update

พบระหว่าง UAT จอสูตรผลิต (uat-crud-mongo): หลัง PUT product 2 ตัว จู่ๆ list หาไม่เจอทั้งคู่ (`/product?materialtype=1,2,4` → 0) ทั้งที่ doc ยังอยู่ใน Mongo.

## Root cause
`internal/product/product/services/product_http_service.go` `Update()`:
```go
docData := findDoc
docData.ProductData = doc.ProductData   // ← ProductData embeds HoldingCodeentity + DocIdentity
```
assign embedded struct ทั้งก้อนจาก request → `holdingcode` และ `guidfixed` กลายเป็นค่าที่ client ส่งมา (มักว่าง เพราะ client ไม่ echo identity กลับ) → ทุก read ที่ scope ด้วย holdingcode มองไม่เห็น doc นั้นอีกเลย = **data corruption ระดับ tenant-isolation** (ไม่ใช่แค่ display bug).

## Fix
Restore identity จาก doc เดิมหลัง assign เสมอ — identity มาจาก stored doc + auth context เท่านั้น ห้ามมาจาก request body:
```go
docData.HoldingCode = findDoc.HoldingCode
docData.GuidFixed = findDoc.GuidFixed
```
ซ่อมข้อมูล: ลบ 2 docs ที่ tenant-detached (หาไม่เจอผ่าน API แล้วเพราะ holdingcode ว่าง) → สร้างใหม่ผ่าน API. Regression-check: PUT ซ้ำ → identity อยู่ครบ + list เจอ.

## บทเรียน (pattern class เดียวกับ warehouse migration เมื่อวาน)
**"Embedded-struct whole-assign ใน update path" = จุดเสี่ยงล้าง identity/field ที่ client ไม่ได้ส่ง.** เจอ 2 วันติด 2 module (warehouse: ล้าง location เมื่อ payload ไม่มี key; product: ล้าง holdingcode/guidfixed). เวลา review update handler ตัวไหนก็ตาม เช็คเสมอว่า assign ก้อนใหญ่จาก request แล้ว field ไหนต้อง survive จาก stored doc.

## หมายเหตุ harness (ไม่ใช่บั๊กแอป)
ข้อมูล seed ไทยเป็น `????` — curl.exe บน Windows แปลง argv ไทย. Seed ด้วย `--data-binary @file.json` (UTF-8) เท่านั้น ห้ามใส่ไทยใน `-d` inline.

related: [[2026-07-01-warehouse-mongodb-migration]], [[2026-07-02-warehouse-subroutes-4-bugs]]
