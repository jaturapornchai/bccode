---
date: 2026-07-01
status: accepted
tags: [bc-account, architecture, mongodb, postgres, warehouse]
---

# คลัง (warehouse/zone/shelf) ย้าย backend จาก PostgreSQL-direct กลับไป MongoDB-first

## Context
UAT sweep พบว่าจอ "คลัง" (`/productwarehousescreen`) รัน CRUD 100% บน PostgreSQL ตรงๆ (GORM, 3 relational table, per-tenant database) — ผิดกฎ "Iron MongoDB Source Rule" ที่มีอยู่แล้วในโปรเจกต์ (MongoDB ต้องเป็น operational source เดียว) แต่ไม่มีที่ไหน document ว่าเป็น exception ที่ตั้งใจ. ลุงจืด confirm: **MongoDB = เก็บข้อมูลดิบ, PostgreSQL = คำนวณแบบ relation เท่านั้น** แล้วสั่งย้าย.

พบระหว่างสำรวจโค้ด: มี MongoDB implementation ของคลังอยู่แล้ว (`warehouse_http_service.go` + `warehouse_mongo_repository.go` + Kafka projection ไป Postgres) — ถูกสถาปัตยกรรมตั้งแต่แรก แต่ route จริงไม่เคยถูก wire ไปที่ implementation นี้ กลับไป wire Postgres-ตรงแทน (น่าจะเพราะ Mongo version มีบั๊กจริง 2 จุดที่ไม่เคยถูกแก้ — `CreateShelf` ไม่ persist, `UpdateShelf` append ซ้ำ — ทำให้ดูเหมือนใช้ไม่ได้).

## Decision
Rewire `/warehouse` routes กลับไปใช้ MongoDB implementation เดิมที่มีอยู่ (ไม่ได้ออกแบบใหม่ทั้งหมด) — แก้บั๊ก 2 จุด, เพิ่ม field ที่ขาด (`CompanyGuids`, `GuidFixed` บน Location/Shelf), เขียน HTTP handler ใหม่ที่คง constructor/route signature เดิมทุกจุด (main.go ไม่ต้องแก้เลย) และคง API contract เดิม 100% (frontend ไม่ต้องแก้เลย).

จุดสำคัญที่สุด: frontend มี 2 endpoint ที่ยิง `PUT /warehouse/:id` ด้วย payload คนละแบบ — แก้ชื่อคลังอย่างเดียว (ไม่มี `location` key) vs แก้โซน/ชั้นวาง (มี `location` array เต็ม). Handler ใหม่ต้องแยกให้ออกว่า "ไม่มี key" (เก็บของเดิมไว้) กับ "มี key แม้จะว่าง" (แทนที่) — ใช้ pointer type (`*[]Location`, `*[]string`) แก้ปัญหานี้.

## Alternatives
- **ออกแบบ Mongo schema ใหม่ทั้งหมด**: ไม่เลือก เพราะของเดิมที่มีอยู่แล้วถูกสถาปัตยกรรมและตรง field name กับที่ frontend ใช้อยู่แล้ว (`location`/`shelf` matching `MongoDB aggregate embed rule`) — reuse ถูกกว่าและเสี่ยงน้อยกว่า
- **เก็บ Postgres ไว้เป็น primary, sync ไป Mongo ทีหลัง**: ไม่เลือก เพราะขัดกฎ Iron MongoDB Source Rule ตรงๆ และ ลุงจืด สั่งชัดว่าต้องอยู่ Mongo

## Consequences
- ✅ ตรงกฎสถาปัตยกรรมแล้ว, ไม่ต้องมี per-tenant Postgres connection/AutoMigrate dance อีกต่อไป (ตัวเหตุ 2 ใน 3 บั๊กที่เจอก่อนหน้านี้ในวันเดียวกัน)
- ✅ verify แล้วจริง: Playwright test เดิม (ไม่แก้ไฟล์) ผ่าน 2 รอบ, ทดสอบ partial-update property ตรงผ่าน curl จริง (แก้ชื่ออย่างเดียวไม่ล้าง location)
- ⚠️ ตาราง Postgres เดิม (`warehouse`/`warehousezones`/`warehouseshelves`/`companywarehouses`) กลายเป็นขยะ — ไม่ลบ ไม่ migrate (pre-launch disposable rule)
- ⚠️ Kafka projection ไป Postgres (`warehouse_consumer_service.go` → `WarehousePG` flat JSONB) ยังไม่ได้ตรวจ/ทดสอบรอบนี้ — เป็นส่วน "pgsql คำนวณ relation" ของกฎใหม่ ที่ยังไม่มีใครใช้จริง รอ use case
- ⚠️ zone/shelf sub-routes (`/warehouse/:guid/zone/...`) rewire แล้วแต่ frontend ไม่ใช้เลย ยังไม่ได้ UAT ตรงๆ

related: [[2026-07-01-warehouse-screen-3-bugs]], [[2026-07-01-productcategorylist-codelist-and-emptystate]]
