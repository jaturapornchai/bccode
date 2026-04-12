---
name: accounting
description: >
  Double-Entry accounting system covering Chart of Accounts, Journal entries, General Ledger,
  and financial reports. Use when the user mentions accounting, GL, journal, chart of accounts,
  debit/credit entries, GL period, account period closing, or financial reporting.
user-invocable: true
---

# Accounting System (Accounting & General Ledger)

## Overview
Double-Entry accounting system with Journal, Chart of Accounts, and financial reports.

## When to Use
- User says "post journal entry"
- User asks about Chart of Accounts setup or account group configuration
- User mentions "close accounting period" / "account period"
- User wants to generate financial reports (P&L, Balance Sheet)
- User asks about Debit/Credit imbalance or troubleshooting GL discrepancies

## System Structure

### Chart of Accounts
| Section | File | Purpose |
|---------|------|---------|
| Backend | `backend/internal/vfgl/chartofaccount/` | CRUD Chart of Accounts |
| BLoC | `bloc/chart_account/` | State management |
| Screen | `screens/config/` | Chart of Accounts configuration |

### Journal
| Section | File | Purpose |
|---------|------|---------|
| Backend | `backend/internal/vfgl/journal/` | Record accounting entries |
| WebSocket | `backend/cmd/ws/` | Realtime update notifications |
| Consumer | `backend/cmd/journalconsume/` | Process journal events |
| BLoC | `bloc/gl_process/` | GL state management |
| Screen | `screens/gl/` | GL screens |

### Account Group
| Section | File | Purpose |
|---------|------|---------|
| Backend | `backend/internal/vfgl/accountgroup/` | Group accounts |
| Backend | `backend/internal/vfgl/accountperiodmaster/` | Manage accounting periods |

### Journal Report
| Section | File | Purpose |
|---------|------|---------|
| Backend | `backend/internal/vfgl/journalreport/` | Generate financial reports |
| Screen | `screens/report/` | Display reports |

## Double-Entry Principle
```
Every entry must satisfy: Debit = Credit always

Example: Purchase goods for 10,000 THB
  Debit:  Inventory          10,000
  Credit: Accounts Payable   10,000
```

## Account Period
- Manage opening/closing of accounting periods
- Entries cannot be recorded in a closed period
- Supports calendar year / fiscal year

## System Admin
Backend has `systemadmin/` that manages:
- `chartofaccountadmin` — Admin-level Chart of Accounts management
- `journaladmin` — Admin-level journal management
- `creditoradmin` / `debtoradmin` — Admin-level creditor/debtor management

## Frontend (Flutter)

### BLoCs
| BLoC | Purpose |
|------|---------|
| `chart_account_bloc` | Chart of Accounts |
| `gl_process_bloc` | General Ledger |
| `cost_center_bloc` | Cost Center |
| `bank_bloc` | Bank |
| `book_bank_bloc` | Book Bank |

### Repositories
`chart_account_repository`, `cost_center_repository`, `bank_repository`, `book_bank_repository`, `gl_process_repository`

### Models
`accountchart_model`, `accountgroup_model`, `accountbook_model`, `chart_account_model`, `bank_model`, `book_bank_model`, `journal_model`

## Important Notes
- Journal supports **WebSocket** for realtime updates
- There are 2 consumers: `journalconsume` (record) + `journalconsumerdelete` (delete)
- GL has a separate service: `backend/cmd/glservice/`
- Chart of Accounts must comply with Thai Accounting Standards

## Anti-Patterns
- ❌ Never record a journal entry in a closed period — always check account period status first
- ❌ Never hardcode account codes — always fetch from Chart of Accounts via API
- ❌ Never create a journal where Debit ≠ Credit — Double-Entry must always balance
- ❌ Remember that GL service runs in a separate process — must go through Kafka consumer before PG update
- ❌ Never delete a Chart of Accounts entry that has journal entries linked — check references first

## Related Skills
- `/creditor-debtor` — for Accounts Payable / Accounts Receivable
- `/payment` — for recording payments that require journal entries
- `/sales-transaction` — purchase/sale documents that auto-generate journal entries
- `/enum-list` — for account type codes and period status enums
- `/master-data` — for cost center and master data used in GL
