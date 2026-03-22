# Inventory Costing System — Backend Reference

## Overview
ระบบคำนวณต้นทุนสินค้าคงคลัง รองรับ 7 วิธีคิดต้นทุน

## Costing Methods (Phase 1 MVP)
| Method | Constant | Engine File |
|--------|----------|-------------|
| Moving Average | `moving_average` | `costing/moving_average.go` |
| FIFO | `fifo` | `costing/fifo.go` |
| LIFO | `lifo` | `costing/lifo.go` |
| FEFO | `fefo` | `costing/fefo.go` |
| Standard Cost | `standard` | `costing/standard.go` |

### Phase 2 (ยังไม่ implement)
- Periodic Average (`periodic_average`)
- Lot-Based (`lot`)

## Package Structure
```
internal/goapi/inventory/
├── handler.go         ← HTTP handlers (per-request DB via mypg.PgSqlFastConnect)
├── service.go         ← Service orchestration (transaction wrapping)
├── database.go        ← CREATE TABLE (8 tables + 9 indexes)
├── models/
│   └── models.go      ← All structs + constants
└── costing/
    ├── engine.go       ← CostingEngine interface + shared helpers
    ├── moving_average.go
    ├── fifo.go
    ├── lifo.go
    ├── fefo.go
    └── standard.go
```

## API Endpoints (registered at `/goapi/api/`)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/products/:itemcode/costing-config?shopid=` | ดู costing config |
| PUT | `/products/:itemcode/costing-config?shopid=` | แก้ costing config |
| GET | `/products/:itemcode/cost-layers?shopid=&whcode=` | ดู cost layers |
| POST | `/inventory/receipt` | รับสินค้าเข้า |
| POST | `/inventory/issue` | ตัดสินค้าออก |
| POST | `/inventory/transfer` | โอนย้ายคลัง |
| POST | `/inventory/adjustment` | ปรับปรุง stock |
| POST | `/inventory/sales-return` | รับคืนจากลูกค้า |
| POST | `/inventory/purchase-return` | ส่งคืน supplier |
| GET | `/reports/inventory-valuation?shopid=` | รายงานมูลค่าสินค้า |
| GET | `/reports/stock-card/:itemcode?shopid=&from=&to=` | Stock Card |
| POST | `/inventory/create-tables?shopid=` | สร้างตาราง |

## Database Tables (PostgreSQL)
1. `product_costing_config` — costing method per product
2. `inventory_cost_layers` — FIFO/LIFO/FEFO layers
3. `inventory_stock_balances` — current qty/avg/total per product×warehouse
4. `inventory_cost_transactions` — full audit trail
5. `inventory_variances` — standard cost variances
6. `inventory_landed_costs` — landed cost header
7. `inventory_landed_cost_allocations` — landed cost allocation
8. `inventory_accounting_periods` — period management

## Key Patterns
- **Per-request DB**: handler ใช้ `mypg.PgSqlFastConnect(shopID)` ตาม shopID
- **Transaction wrapping**: `service.withTransaction()` ใช้ `sql.LevelReadCommitted`
- **Layer locking**: `SELECT ... FOR UPDATE` ป้องกัน concurrent consumption
- **Engine factory**: `costing.NewCostingEngine(method)` สร้าง engine ตาม method
- **Default method**: ถ้าไม่ตั้งค่า → ใช้ `moving_average`

## Frontend (Flutter)
| File | Description |
|------|-------------|
| `model/inventory_costing_model.dart` | Dart models (json_serializable) |
| `repositories/inventory_costing_repository.dart` | Dio API calls |
| `bloc/inventory_costing/` | BLoC (event/state/bloc) |
| `screens/config/inventory_costing_screen.dart` | Config screen |
