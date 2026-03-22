---
name: auto-packing
description: >
  BC Account Auto Packing Module — Pick & Pack workflow, BOM (Bill of Materials),
  unit conversion/auto packing algorithm, packing list, shipping label, warehouse
  management feature tracker and development guide. Use when: (1) adding/modifying
  pick & pack screens or backend, (2) working with BOM/product assembly/combo/gift set,
  (3) auto packing unit conversion logic, (4) packing list or shipping label generation,
  (5) warehouse shelf/location management, (6) checking packing feature status.
  Triggers on: "packing", "pick and pack", "pick & pack", "BOM", "bill of materials",
  "auto pack", "แพ็ค", "หยิบสินค้า", "จัดส่ง", "กล่อง", "carton", "shipping label",
  "packing list", "combo", "gift set", "ชุดสินค้า", "สินค้าสำเร็จรูป", "unit conversion",
  "หน่วยนับ", "คลังสินค้า", "warehouse", "shelf", "ชั้นวาง".
---

# Auto Packing Module — Development Guide

## Mandatory Update Rule

**ทุกครั้งที่แก้ code packing/BOM/warehouse → ต้อง update skill นี้ทันที:**
1. แก้ feature ไหน → update status ใน `references/feature-status.md`
2. เพิ่ม feature ใหม่ → เพิ่มแถวใน feature status
3. แก้ backend model → update `references/system-architecture.md`
4. แก้ frontend UI → update status จาก pending → done
5. **ห้ามจบ session โดยไม่ update** — ถ้ามีการแก้ packing code

## System Overview

```
┌─────────────┐     ┌──────────────┐     ┌────────────────┐
│  Sale Order  │────▶│ Sale Invoice  │────▶│  Pick & Pack   │
│  (SO)        │     │  (SI)        │     │  Document      │
└─────────────┘     └──────────────┘     └────────┬───────┘
                                                   │
                    ┌──────────────┐     ┌─────────▼───────┐
                    │  BOM Config  │────▶│  Auto Packing   │
                    │  (Product)   │     │  (Unit Suggest)  │
                    └──────────────┘     └────────┬───────┘
                                                   │
                    ┌──────────────┐     ┌─────────▼───────┐
                    │  Warehouse   │────▶│  Packing List   │
                    │  Location    │     │  + Label Print   │
                    │  Shelf       │     └─────────────────┘
                    └──────────────┘
```

## Key Files

### Backend (Go)

| Area | Path |
|------|------|
| **Auto Packing Cache** | `backend/internal/goapi/process/process-stock/auto_packing_cache.go` |
| **Pick & Pack HTTP** | `backend/internal/transaction/pickandpack/pickandpack_http.go` |
| **Pick & Pack Model** | `backend/internal/transaction/pickandpack/models/pickandpack.go` |
| **Pick & Pack Service** | `backend/internal/transaction/pickandpack/services/pickandpack_http_service.go` |
| **Pick & Pack Mongo Repo** | `backend/internal/transaction/pickandpack/repositories/pickandpack_mongo_repository.go` |
| **Pick & Pack MQ Repo** | `backend/internal/transaction/pickandpack/repositories/pickandpack_messagequeue_repository.go` |
| **BOM Model** | `backend/internal/product/bom/models/bom.go` |
| **BOM HTTP** | `backend/internal/product/bom/bom_http.go` |
| **BOM Service** | `backend/internal/product/bom/services/bom_http_service.go` |
| **BOM Consumer** | `backend/internal/product/bom/bom_consumer.go` |
| **BOM PG Model** | `backend/internal/product/bom/models/bom_postgres.go` |
| **PP Device Model** | `backend/internal/pickandpack/models/device.go` |
| **Stock Balance** | `backend/internal/goapi/process/process-stock/process-stock-balance-by-item-warehouse-location.go` |
| **TransFlag Constants** | `backend/internal/goapi/handlers/kafka/constants.go` |

### Frontend — bcaiaccount (ERP App)

| Area | Path |
|------|------|
| **BOM Screen** | `frontend/bcaiaccount/lib/screens/config/product_barcode_bom_screen.dart` |
| **BOM Widget (Graph)** | `frontend/bcaiaccount/lib/screens/config/product_bom_widget.dart` |
| **BOM Model** | `frontend/bcaiaccount/lib/model/product_bom_model.dart` |
| **Barcode Repository** | `frontend/bcaiaccount/lib/repositories/product_barcode_repository.dart` |
| **Barcode BLoC** | `frontend/bcaiaccount/lib/bloc/product_barcode/product_barcode_bloc.dart` |

### Frontend — bclms (Warehouse/LMS App)

| Area | Path |
|------|------|
| **Routes (Pick & Pack)** | `frontend/bclms/lib/app/routes/app_routes.dart` |
| **Warehouse Model** | `frontend/bclms/lib/app/modules/warehouse/warehouse_model.dart` |
| **Warehouse Service** | `frontend/bclms/lib/app/modules/warehouse/warehouse_service.dart` |
| **Warehouse List** | `frontend/bclms/lib/app/modules/warehouse/warehouse_list_page.dart` |

## Architecture Patterns

### Auto Packing Algorithm

Unit ratio-based packing suggestion — เรียง unit ใหญ่สุดก่อน:

```sql
SELECT DISTINCT unitname, barcoderefunitstand, barcoderefunitdivide,
       barcoderefunitstand / NULLIF(barcoderefunitdivide, 1) AS unit_ratio
FROM productbarcode
WHERE itemcode = $1 AND barcoderefunitstand > 0 AND barcoderefunitdivide > 0
ORDER BY unit_ratio DESC
```

- `BuildAutoPackingCache()` — bulk fetch packing info for multiple items
- `FetchPackingForItem()` — get cached with fallback on-demand query
- Example: Product A → EA(1), BOX(24), PALLET(48) → suggest PALLET first

### BOM Hierarchy

Recursive nested structure — parent → child → grandchild:

```
ProductBarcodeBOMView {
  BarcodeGuidFixed, Level, Barcode, Qty,
  StandValue, DivideValue,  // unit conversion
  Condition,                // conditional BOM
  BOM: []ProductBarcodeBOMView  // recursive children
}
```

### Pick & Pack Status Flow

```
0 (PENDING) → 1 (PROCESSING) → 2 (COMPLETED)
                              → 3 (CANCELLED)
```

- `UpdatePrint(docno)` → status 0→1
- `ConfirmPickandpack(docno)` → status 1→2
- `CancelPickandpack(docno)` → status→3

### Database Collections

| Database | Collection/Table | Purpose |
|----------|-----------------|---------|
| MongoDB | `transactionPickandpack` | Pick & Pack documents |
| MongoDB | `productBarcodeBOMs` | BOM definitions |
| MongoDB | `transactionSaleinvoiceBOMPrices` | BOM pricing per SI |
| MongoDB | `pickandpackDevices` | Warehouse device config |
| PostgreSQL | `productbarcodeboms` | BOM sync (via Kafka) |
| PostgreSQL | `productbarcode` | Unit conversion data |
| PostgreSQL | `stock_calculation_state` | Incremental calc tracking |

## API Endpoints

### Pick & Pack CRUD

```
POST   /transaction/pickandpack           → Create
GET    /transaction/pickandpack           → Search paginated
GET    /transaction/pickandpack/list      → Search offset/limit
GET    /transaction/pickandpack/:id       → Get by GUID
GET    /transaction/pickandpack/code/:code → Get by docno
PUT    /transaction/pickandpack/:id       → Update
DELETE /transaction/pickandpack/:id       → Delete
DELETE /transaction/pickandpack           → Bulk delete
POST   /transaction/pickandpack/bulk      → Bulk import
```

### Pick & Pack Workflow

```
POST /transaction/pickandpack/approve/:id              → Approve
PUT  /transaction/pickandpack/updateprint/:docno        → Mark printed (0→1)
PUT  /transaction/pickandpack/confirmpickandpack/:docno → Confirm done (1→2)
PUT  /transaction/pickandpack/cancel/:docno             → Cancel (→3)
PUT  /transaction/pickandpack/status/:id/:status        → Set status
PUT  /transaction/pickandpack/closejob/:id              → Close job
DELETE /transaction/pickandpack/cancel-by-saleinvoice/:id → Cancel by SI
```

### Pick & Pack Dashboard

```
GET /transaction/pickandpack/available-saleinvoice   → SI not yet packed
GET /transaction/pickandpack/by-warehouse            → Group by warehouse
GET /transaction/pickandpack/saleinvoice-status       → Packing progress per SI
GET /transaction/pickandpack/dashboard                → Overall stats
GET /transaction/pickandpack/dashboard/warehouse      → Per-warehouse stats
GET /transaction/pickandpack/history                  → Historical data
```

### BOM

```
POST   /product/bom      → Create/upsert
GET    /product/bom      → Search paginated
GET    /product/bom/list → Search offset/limit
GET    /product/bom/:id  → Get detail
DELETE /product/bom/:id  → Delete
```

## App Responsibility Split

| Feature | bcaiaccount (ERP) | bclms (Warehouse) |
|---------|-------------------|-------------------|
| BOM management | ✅ Full CRUD + graph view | ❌ |
| Stock transactions | ✅ Pickup/Receive/Return | ❌ |
| Warehouse structure | ❌ Config only | ✅ Full CRUD |
| Pick & Pack | ❌ Stock pickup only | ✅ Dedicated module |
| Location/Shelf | ❌ | ✅ Full hierarchy |
| Barcode scanning | ✅ Product search | ✅ Picking validation |
| Auto packing calc | ✅ Backend ready | ✅ Same logic |
| Device management | ❌ | ✅ Admin/Warehouse/Display |
| Dashboard | ✅ General | ✅ Warehouse pick/pack |

## Feature Status

ดู `references/feature-status.md` สำหรับ feature list ครบทุกข้อ + status

## References

- **Feature Status**: `references/feature-status.md` — feature list + status ครบทุกข้อ
- **System Architecture**: `references/system-architecture.md` — models, data flow, Kafka topics
