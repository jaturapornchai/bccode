---
name: stock-inventory
description: Stock & Inventory — receive, issue, transfer, 7 costing methods
user-invocable: true
---

# Stock & Inventory System

## Overview
Manage inventory balances, warehouses, stock receive/issue/transfer, and cost calculation.

## System Structure

### Warehouse
| Part | File | Purpose |
|------|------|---------|
| Backend | `backend/internal/warehouse/` | Warehouse CRUD |
| BLoC | `bloc/warehouse/` + `bloc/warehouse_location/` | State management |

### Stock
| Type | Backend Path | Purpose |
|------|-------------|---------|
| Balance | `transaction/stockbalance/` | View current stock balance |
| Detail | `transaction/stockbalancedetail/` | View details per lot |
| Transfer | `transaction/stocktransfer/` | Move goods between warehouses |
| Receive | `transaction/stockreceiveproduct/` | Receive goods into warehouse |
| Return | `transaction/stockreturnproduct/` | Return goods from warehouse |
| Pickup | `transaction/stockpickupproduct/` | Pick goods out of warehouse |
| Adjustment | `transaction/stockadjustment/` | Adjust stock balance (increase/decrease) |
| Import | `backend/internal/stockbalanceimport/` | Import stock data |

### Inventory Costing
| Part | File | Purpose |
|------|------|---------|
| Backend | `backend/internal/goapi/inventory/` | Calculate cost using 7 methods |
| BLoC | `bloc/inventory_costing/` | State management |
| Config | `screens/config/inventory_costing_screen.dart` | Configure costing method |
| Reference | `references/go/inventory-costing.md` | Reference documentation |

### Stock Process
| Part | File | Purpose |
|------|------|---------|
| Backend | `backend/internal/stockprocess/` | Stock processing |
| Product Balance | `goapi/handlers/product_balance/` | Calculate product balance |
| Product Cache | `goapi/handlers/product_cache/` | Product data cache |

## Stock Movement Flow
```
Purchase
  |
  v
Stock Receive -> Stock increases
  |
  v
Sale Invoice -> Stock decreases
  |
  v
Return -> Stock increases back

Inter-warehouse transfer:
Warehouse A -> Stock Transfer -> Warehouse B

Stock adjustment (physical count != system):
Stock Adjustment -> Correct the balance
```

## 7 Costing Methods
| # | Name | Description |
|---|------|-------------|
| 1 | FIFO | First in, first out |
| 2 | LIFO | Last in, first out |
| 3 | Weighted Average | Weighted average cost |
| 4 | Moving Average | Moving average cost |
| 5 | Specific ID | Identify specific lot |
| 6 | Standard Cost | Standard cost |
| 7 | Retail Method | Retail price method |

## Frontend (Flutter)

### BLoCs
| BLoC | Purpose |
|------|---------|
| `inventory_bloc` | Inventory |
| `inventory_costing_bloc` | Cost calculation |
| `warehouse_bloc` | Warehouse |
| `warehouse_location_bloc` | Warehouse location |
| `stock_balance_bloc` | Stock balance |
| `shelf_product_bloc` | Shelf products |
| `shelf_product_selector_bloc` | Shelf product selector |

### Repositories
`inventory_repository`, `inventory_costing_repository`, `warehouse_repository`, `warehouse_location_repository`, `stock_balance_import_repository`, `shelf_product_repository`

### Services
`stock_cost_api_service`, `stock_report_lookup_service`

### Models
`inventory_model`, `inventory_costing_model`, `warehouse_model`, `warehouse_location_model`, `stock_balance_import_model`, `shelf_model`, `shelf_product_model`

## Important Notes
- Stock must sync across 3 DBs: MongoDB -> Kafka -> PG + ClickHouse
- `stockbalanceimport` is used to import Opening Balance
- Stock adjustments must always include a reason
- Inter-warehouse transfers must balance: sender (-) + receiver (+) = total unchanged
- Pick & Pack is for preparing goods before shipment -> see skill `/auto-packing`
- Product search uses GoAPI (`product_search/`) + Thai tokenizer
