# Inventory Costing System — Backend Reference

## Overview
Inventory costing system supporting 7 costing methods.

## Costing Methods (Phase 1 MVP)
| Method | Constant | Engine File |
|--------|----------|-------------|
| Moving Average | `moving_average` | `costing/moving_average.go` |
| FIFO | `fifo` | `costing/fifo.go` |
| LIFO | `lifo` | `costing/lifo.go` |
| FEFO | `fefo` | `costing/fefo.go` |
| Standard Cost | `standard` | `costing/standard.go` |

### Phase 2 (not yet implemented)
- Periodic Average (`periodic_average`)
- Lot-Based (`lot`)

## Package Structure
```
internal/goapi/inventory/
├── handler.go         <- HTTP handlers (per-request DB via mypg.PgSqlFastConnect)
├── service.go         <- Service orchestration (transaction wrapping)
├── database.go        <- CREATE TABLE (8 tables + 9 indexes)
├── models/
│   └── models.go      <- All structs + constants
└── costing/
    ├── engine.go       <- CostingEngine interface + shared helpers
    ├── moving_average.go
    ├── fifo.go
    ├── lifo.go
    ├── fefo.go
    └── standard.go
```

## API Endpoints (registered at `/goapi/api/`)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/products/:itemcode/costing-config?shopid=` | Get costing config |
| PUT | `/products/:itemcode/costing-config?shopid=` | Update costing config |
| GET | `/products/:itemcode/cost-layers?shopid=&whcode=` | View cost layers |
| POST | `/inventory/receipt` | Goods receipt |
| POST | `/inventory/issue` | Goods issue |
| POST | `/inventory/transfer` | Warehouse transfer |
| POST | `/inventory/adjustment` | Stock adjustment |
| POST | `/inventory/sales-return` | Customer return |
| POST | `/inventory/purchase-return` | Supplier return |
| GET | `/reports/inventory-valuation?shopid=` | Inventory valuation report |
| GET | `/reports/stock-card/:itemcode?shopid=&from=&to=` | Stock Card |
| POST | `/inventory/create-tables?shopid=` | Create tables |

## Database Tables (PostgreSQL)
1. `product_costing_config` — costing method per product
2. `inventory_cost_layers` — FIFO/LIFO/FEFO layers
3. `inventory_stock_balances` — current qty/avg/total per product x warehouse
4. `inventory_cost_transactions` — full audit trail
5. `inventory_variances` — standard cost variances
6. `inventory_landed_costs` — landed cost header
7. `inventory_landed_cost_allocations` — landed cost allocation
8. `inventory_accounting_periods` — period management

## Key Patterns
- **Per-request DB**: handler uses `mypg.PgSqlFastConnect(shopID)` per shopID
- **Transaction wrapping**: `service.withTransaction()` uses `sql.LevelReadCommitted`
- **Layer locking**: `SELECT ... FOR UPDATE` prevents concurrent consumption
- **Engine factory**: `costing.NewCostingEngine(method)` creates engine by method
- **Default method**: if not configured -> uses `moving_average`

## Frontend (Flutter)
| File | Description |
|------|-------------|
| `model/inventory_costing_model.dart` | Dart models (json_serializable) |
| `repositories/inventory_costing_repository.dart` | Dio API calls |
| `bloc/inventory_costing/` | BLoC (event/state/bloc) |
| `screens/config/inventory_costing_screen.dart` | Config screen |
