# Auto Packing Feature Status

Last updated: 2026-03-14

## Legend
- **done** = Complete and usable
- **backend-done** = Backend ready, frontend pending
- **in-progress** = Currently being worked on
- **pending** = Not started but planned
- **planned** = On roadmap
- **future** = Enterprise feature for later

---

## 1. Auto Packing (Unit Conversion)

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 1 | Unit ratio calculation | **done** | `auto_packing_cache.go` — ratio DESC ordering |
| 2 | Multi-unit product detection | **done** | IsAutoPacking flag in stock balance |
| 3 | Packing suggestion (largest unit first) | **done** | `FetchPackingForItem()` |
| 4 | Bulk cache optimization | **done** | `BuildAutoPackingCache()` — batch query |
| 5 | Incremental stock calculation | **done** | Checksum-based skip for unchanged items |
| 6 | Auto packing UI in stock pickup | **pending** | bcaiaccount stock pickup screen |
| 7 | Auto packing UI in pick & pack | **pending** | bclms pick & pack screen |

## 2. Pick & Pack Workflow

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 8 | Pick & Pack model (MongoDB) | **done** | `transactionPickandpack` collection |
| 9 | Pick & Pack CRUD API | **done** | Full REST endpoints |
| 10 | Status workflow (pending>processing>completed) | **done** | 4 statuses: 0,1,2,3 |
| 11 | Print tracking | **done** | `updateprint/:docno` |
| 12 | Confirm tracking | **done** | `confirmpickandpack/:docno` |
| 13 | Cancel support | **done** | `cancel/:docno` + `cancel-by-saleinvoice/:id` |
| 14 | Close job | **done** | `closejob/:id` |
| 15 | Available sale invoices | **done** | `available-saleinvoice` endpoint |
| 16 | Group by warehouse | **done** | `by-warehouse` endpoint |
| 17 | Packing progress per SI | **done** | `saleinvoice-status` endpoint |
| 18 | Dashboard (overall) | **done** | `dashboard` endpoint |
| 19 | Dashboard per warehouse | **done** | `dashboard/warehouse` endpoint |
| 20 | History with date range | **done** | `history` endpoint |
| 21 | Bulk import | **done** | `POST /bulk` endpoint |
| 22 | Approve pick & pack | **done** | `approve/:id` endpoint |
| 23 | Kafka MQ sync | **done** | Create/Update/Delete topics |
| 24 | Pick & Pack list screen (bclms) | **pending** | Route exists, placeholder |
| 25 | Pick & Pack edit screen (bclms) | **pending** | Not started |
| 26 | Pick & Pack dashboard UI (bclms) | **pending** | Backend ready |
| 27 | Barcode scanning for picking | **pending** | bclms device integration |

## 3. BOM (Bill of Materials)

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 28 | BOM model (MongoDB + PG) | **done** | Recursive hierarchy, checksum |
| 29 | BOM CRUD API | **done** | Upsert with checksum comparison |
| 30 | BOM Kafka consumer (PG sync) | **done** | `bom_consumer.go` |
| 31 | BOM list screen (bcaiaccount) | **done** | `product_barcode_bom_screen.dart` |
| 32 | BOM tree graph view | **done** | `product_bom_widget.dart` + GraphView |
| 33 | BOM unit conversion | **done** | Multi-unit support |
| 34 | Conditional BOM | **done** | `condition` flag |
| 35 | Sale Invoice BOM pricing | **done** | `transactionSaleinvoiceBOMPrices` |
| 36 | Combo/Gift set via BOM | **done** | BOM structure for bundling |
| 37 | BOM image support | **done** | `imageuri` field |

## 4. Warehouse Management (bclms)

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 38 | Warehouse CRUD | **done** | Full API + UI in bclms |
| 39 | Location hierarchy | **done** | Warehouse > Location |
| 40 | Shelf management | **done** | Location > Shelf > ShelfProduct |
| 41 | Warehouse device management | **backend-done** | Admin/Warehouse/Display types |
| 42 | Device PIN authentication | **backend-done** | `activepin` field |
| 43 | Employee-device assignment | **backend-done** | `employees[]` on device model |

## 5. Packing List & Shipping Label

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 44 | Packing list generation | **pending** | PDF/print from pick & pack doc |
| 45 | Shipping label template | **planned** | Needs label design + print API |
| 46 | Tracking number field | **planned** | Not yet in pickandpack model |
| 47 | Courier integration (Kerry/Flash/etc) | **future** | Third-party API |
| 48 | QR code on packing label | **planned** | Barcode/QR lib |

## 6. Bin Packing (3D)

| # | Feature | Status | Notes |
|---|---------|--------|-------|
| 49 | Box/carton master data | **planned** | Needs dimension fields |
| 50 | 3D bin packing algorithm | **future** | Optimize box usage |
| 51 | Carton suggestion per order | **future** | Auto-select smallest box |
| 52 | Weight-based packing | **future** | Weight limit per box |

---

## Summary

| Category | Done | Backend-Done | Pending | Planned | Future |
|----------|------|-------------|---------|---------|--------|
| Auto Packing | 5 | 0 | 2 | 0 | 0 |
| Pick & Pack | 16 | 0 | 4 | 0 | 0 |
| BOM | 10 | 0 | 0 | 0 | 0 |
| Warehouse | 3 | 3 | 0 | 0 | 0 |
| Packing List | 0 | 0 | 1 | 2 | 1 |
| Bin Packing | 0 | 0 | 0 | 1 | 3 |
| **Total** | **34** | **3** | **7** | **3** | **4** |
