---
name: auto-packing
description: >
  Pick & Pack workflow, BOM (Bill of Materials), auto unit conversion/packing suggestion,
  packing list, and warehouse shelf management. Use when the user mentions packing goods,
  preparing shipments, managing BOM combos, or warehouse shelf operations.
---

# Auto Packing Module

## When to Use
- User wants to prepare a packing list from a sale invoice or sale order
- User needs to configure BOM for a combo product (e.g., Gift Set = Item A + Item B + Box)
- User asks why the system suggests a specific unit during packing (auto packing algorithm)
- User wants to track pick & pack status (PENDING → PROCESSING → COMPLETED)
- User needs to manage warehouse shelf structure (location/shelf hierarchy in bclms)

## Anti-Patterns
- Do NOT create a Pick & Pack document without a linked Sale Invoice — orphan documents break reporting
- Do NOT configure BOM with circular references — recursive parent-child must be a DAG (no loops)
- Do NOT confirm Pick & Pack as COMPLETED before all items are physically picked — status is final
- Do NOT modify `productbarcode.barcoderefunitstand/divide` manually — these drive auto packing ratios
- Do NOT assume auto packing cache is always fresh — call `BuildAutoPackingCache()` after BOM changes

## Update Rule
After any packing/BOM/warehouse code change, update this skill:
1. Update feature status in `references/feature-status.md`
2. Update architecture in `references/system-architecture.md`

## System Overview

```
Sale Order -> Sale Invoice -> Pick & Pack Document
                                    |
BOM Config (Product) ---------> Auto Packing (Unit Suggest)
                                    |
Warehouse Location/Shelf -----> Packing List + Label Print
```

## Key Files

### Backend (Go)
| Area | Path |
|------|------|
| Auto Packing Cache | `backend/internal/goapi/process/process-stock/auto_packing_cache.go` |
| Pick & Pack HTTP | `backend/internal/transaction/pickandpack/pickandpack_http.go` |
| Pick & Pack Model | `backend/internal/transaction/pickandpack/models/pickandpack.go` |
| Pick & Pack Service | `backend/internal/transaction/pickandpack/services/pickandpack_http_service.go` |
| BOM Model | `backend/internal/product/bom/models/bom.go` |
| BOM HTTP | `backend/internal/product/bom/bom_http.go` |
| BOM Consumer | `backend/internal/product/bom/bom_consumer.go` |
| Stock Balance | `backend/internal/goapi/process/process-stock/process-stock-balance-by-item-warehouse-location.go` |

### Frontend — bcaiaccount (ERP)
| Area | Path |
|------|------|
| BOM Screen | `frontend/bcaiaccount/lib/screens/config/product_barcode_bom_screen.dart` |
| BOM Widget (Graph) | `frontend/bcaiaccount/lib/screens/config/product_bom_widget.dart` |
| BOM Model | `frontend/bcaiaccount/lib/model/product_bom_model.dart` |

### Frontend — bclms (Warehouse)
| Area | Path |
|------|------|
| Routes (Pick & Pack) | `frontend/bclms/lib/app/routes/app_routes.dart` |
| Warehouse Model | `frontend/bclms/lib/app/modules/warehouse/warehouse_model.dart` |
| Warehouse List | `frontend/bclms/lib/app/modules/warehouse/warehouse_list_page.dart` |

## Architecture

### Auto Packing Algorithm
Unit ratio-based suggestion — largest unit first:
```sql
SELECT unitname, barcoderefunitstand / NULLIF(barcoderefunitdivide, 1) AS unit_ratio
FROM productbarcode WHERE itemcode = $1
ORDER BY unit_ratio DESC
```
- `BuildAutoPackingCache()` — bulk fetch for multiple items
- `FetchPackingForItem()` — cached with on-demand fallback

### BOM Hierarchy
Recursive: parent -> child -> grandchild. `ProductBarcodeBOMView` contains `BOM: []ProductBarcodeBOMView`.

### Pick & Pack Status Flow
```
0 (PENDING) -> 1 (PROCESSING) -> 2 (COMPLETED)
                               -> 3 (CANCELLED)
```

### Database Collections
| DB | Collection | Purpose |
|----|-----------|---------|
| MongoDB | `transactionPickandpack` | Pick & Pack documents |
| MongoDB | `productBarcodeBOMs` | BOM definitions |
| PostgreSQL | `productbarcodeboms` | BOM sync (via Kafka) |
| PostgreSQL | `productbarcode` | Unit conversion data |

## API Endpoints

### Pick & Pack CRUD
```
POST/GET/PUT/DELETE /transaction/pickandpack
GET  /transaction/pickandpack/code/:code
POST /transaction/pickandpack/bulk
```

### Pick & Pack Workflow
```
POST /transaction/pickandpack/approve/:id
PUT  /transaction/pickandpack/updateprint/:docno        (0->1)
PUT  /transaction/pickandpack/confirmpickandpack/:docno (1->2)
PUT  /transaction/pickandpack/cancel/:docno             (->3)
```

### Pick & Pack Dashboard
```
GET /transaction/pickandpack/available-saleinvoice
GET /transaction/pickandpack/by-warehouse
GET /transaction/pickandpack/dashboard
GET /transaction/pickandpack/history
```

### BOM
```
POST/GET/DELETE /product/bom
GET /product/bom/:id
```

## App Responsibility Split

| Feature | bcaiaccount (ERP) | bclms (Warehouse) |
|---------|-------------------|-------------------|
| BOM management | Full CRUD + graph | - |
| Pick & Pack | Stock pickup only | Dedicated module |
| Warehouse structure | Config only | Full CRUD |
| Location/Shelf | - | Full hierarchy |
| Auto packing calc | Backend ready | Same logic |

## References
- [Feature Status](references/feature-status.md) — full feature list + status
- [System Architecture](references/system-architecture.md) — models, data flow, Kafka topics

## Related Skills
- `/product` — BOM definition lives in product barcode; configure here first
- `/stock-inventory` — stock balance must exist before pick & pack can proceed
- `/master-data` — warehouse and unit master data required for packing setup
- `/import-export` — shelf data can be bulk-imported via Excel (`shelf_excel_import`)
