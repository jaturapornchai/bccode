---
name: creditor-debtor
description: >
  Creditor (supplier/vendor), debtor (customers with outstanding balances), and general
  customer master data management. Use when the user mentions creditor, debtor, supplier,
  vendor, customer master, supplier setup, outstanding balance, credit terms, or customer
  grouping.
user-invocable: true
---

# Creditor & Debtor System

## Overview
Manage creditor (supplier/vendor) and debtor (customers with outstanding payments) data for the shop.

## When to Use
- User says "create new creditor" or "add new supplier"
- User asks about customer credit terms or outstanding balance
- User wants to set up debtor groups or customer categories
- User asks whether a creditor/supplier can be safely deleted
- User needs to search or filter creditor/debtor data for purchase or sale documents

## System Structure

### Creditor (= Supplier we owe money to)
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/config/creditor_screen.dart` | View/edit creditor data |
| BLoC | `bloc/creditor/` | State management |
| Repository | `repositories/creditor_repository.dart` | API integration |
| Backend Model | `backend/internal/debtaccount/creditor/` | Go model + handler |
| Backend Group | `backend/internal/debtaccount/creditorgroup/` | Creditor groups |

### Debtor (= Customer who owes us money)
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/config/debtor_screen.dart` | View/edit debtor data |
| BLoC | `bloc/debtor/` | State management |
| Repository | `repositories/debtor_repository.dart` | API integration |
| Backend Model | `backend/internal/debtaccount/debtor/` | Go model + handler |
| Backend Group | `backend/internal/debtaccount/debtorgroup/` | Debtor groups |

### Customer (= General customer, not necessarily a debtor)
| Section | File | Purpose |
|---------|------|---------|
| BLoC | `bloc/customer/` | State management |
| Backend | `backend/internal/debtaccount/customer/` | Go model + handler |
| Backend Group | `backend/internal/debtaccount/customergroup/` | Customer groups |

## Key Differences
| | Creditor | Debtor | Customer |
|---|---|---|---|
| Who | Supplier/Vendor | Customer with outstanding payment | General customer |
| Debt direction | We owe them | They owe us | Not applicable |
| Used with | Purchasing (PO/PR) | Credit sales (Invoice) | General sales |

## Data Flow
```
MongoDB (save) → Kafka → PostgreSQL (upsert) + ClickHouse (upsert)
```
- Both creditor and debtor must sync across all 3 DBs
- Backend has `creditor_migration.go` / `debtor_migration.go` for creating PG tables

## API Endpoints
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/creditor/list` | List creditors |
| POST | `/api/v1/creditor` | Create new creditor |
| PUT | `/api/v1/creditor/:id` | Edit creditor |
| DELETE | `/api/v1/creditor/:id` | Delete creditor |
| GET | `/api/v1/debtor/list` | List debtors |
| POST | `/api/v1/debtor` | Create new debtor |

## Frontend (Flutter)

### BLoCs
| BLoC | Purpose |
|------|---------|
| `creditor_bloc` | Creditor CRUD |
| `creditor_group_bloc` | Creditor groups |
| `creditor_filter_cubit` | Creditor filtering |
| `debtor_bloc` | Debtor CRUD |
| `debtor_group_bloc` | Debtor groups |
| `debtor_filter_cubit` | Debtor filtering |
| `customer_bloc` | Customer CRUD |
| `customer_group_bloc` | Customer groups |

### Repositories
`creditor_repository`, `creditor_group_repository`, `debtor_repository`, `debtor_group_repository`, `customer_repository`, `customer_group_repository`

### Models
`creditor_model`, `creditor_group_model`, `debtor_model`, `debtor_group_model`, `debtor_creditor_model`, `customer_model`, `customer_group_model`, `customer_address_model`

## Important Notes
- Creditors/debtors must always have a `shop_id`
- Deleting a creditor referenced by a PO is dangerous — must check first
- Groups are used for categorization, e.g., "Domestic Suppliers", "Foreign Suppliers"
- Backend has `creditorprocess` / `debtorprocess` for calculating outstanding balances

## Anti-Patterns
- ❌ Never delete a creditor/debtor without checking for linked PO/Invoice — data may be corrupted
- ❌ Always include `shop_id` — every creditor/debtor must be linked to a shop
- ❌ Never confuse Creditor (we owe them) with Debtor (they owe us) — opposite directions
- ❌ Never save creditor/debtor to MongoDB only — must sync via Kafka to PG + ClickHouse
- ❌ Never hardcode creditor/debtor group codes — fetch from group master data

## Related Skills
- `/sales-transaction` — uses debtor/customer in sale documents
- `/payment` — creditor for making payments, debtor for receiving payments
- `/accounting` — Accounts Payable (creditor) / Accounts Receivable (debtor) in GL
- `/master-data` — for group and category master data
- `/enum-list` — for creditor/debtor status and type codes
