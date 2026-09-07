---
tags: [bug, bc-account, go, postgres, frontend, warehouse]
date: 2026-07-01
---

# จอ "คลัง" (warehouse-tree-view) — 3 bug จริง พบระหว่าง UAT

## Bug 1 — สร้างคลังแรกไม่ได้เลย (empty-state silent fail)
`warehouse-tree-view.tsx`: `formType` default = `"editwarehouse"` (ควรเป็น `"createwarehouse"`). tenant ที่ยังไม่มี warehouse เลย → form โชว์เหมือน "เพิ่ม" แต่จริงๆ อยู่ mode "แก้ไข" ของ record ที่ไม่มีอยู่ → `handleSaveForm` early-return เงียบๆ ไม่มี error ใดๆ ทั้งสิ้น.
**Fix:** เปลี่ยน initial state → `"createwarehouse"` (useEffect เดิมยัง override เป็น editwarehouse ถูกต้องเมื่อมี record จริง).

## Bug 2 — wrong backend base (goapi แทน mainapi)
`handleSaveForm`/`handleDeleteWarehouse` fetch `${auth.backendUrl}/warehouse` ตรงๆ — `auth.backendUrl` = goapi base แต่ `/warehouse` อยู่ mainapi (`cmd/app/main.go`). Bug class เดียวกับ company/organization ที่เคยแก้ก่อนหน้า.
**Fix:** wrap ด้วย `deriveMainApiUrl(auth?.backendUrl ?? "")` (import มีอยู่แล้วในไฟล์).

## Bug 3 — table/column name typo (snake_case ผิด)
`internal/warehouse/warehouse_http.go` hardcode raw SQL table/column เป็น snake_case แต่ GORM struct จริงไม่มี underscore:
- `"company_warehouses"` → จริง `"companywarehouses"` (`CompanyWarehousePg.TableName()`)
- `"warehouse_guid"` → จริง `"warehouseguid"` (`gorm:"column:warehouseguid"`)
- `"warehouse_zones"` → จริง `"warehousezones"` (`ZonePg.TableName()`)
- `"zone_guid"` → จริง `"zoneguid"`

ทำให้ UpdateWarehouse/DeleteWarehouse/DeleteZone ฯลฯ 500 `relation "company_warehouses" does not exist`.
**Fix:** sed global replace ทั้ง 4 pattern, verify ด้วย gofmt + real compile (deploy-mainapi-fast.ps1).

## เสริม — defense-in-depth
`ensureTenantSchema` (per-tenant AutoMigrate, เดิมมีแค่ Search/Create) ขยายไปครบ 13 handler เพราะเจอว่า framework's `PersisterTenant` (`pkg/microservice/microservice.go:342`) **cache persister ทันทีแม้ AutoMigrate ทั้งระบบ fail** (ไม่ retry) — ถ้า model อื่นพัง table warehouse จะไม่ถูกสร้างตลอดจน restart. `ensureTenantSchema` ต่อ tenant กู้สถานการณ์นี้ได้.

## Verify
Stagehand full CRUD chain (`scratch/stagehand-uat/uat-warehouse-final.js`): Warehouse create/edit/delete ✅, Location create/edit/delete ✅ (real 200 responses). Shelf create ✅; shelf edit/delete ยังไม่ confirm (test script scoping ปัญหา ไม่ใช่ app bug — title "แก้ไข" ซ้ำ location/shelf).

## Status
Deploy local mainapi เท่านั้น (2 รอบ, healthz ผ่านทั้งคู่) — ยังไม่ deploy .202.

related: [[2026-06-30-barcode-ghost-list-missing-deletedat]]
