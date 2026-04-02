---
name: dashboard-report
description: Dashboard, reports, PDF, CSV, BI Analytics — Dashboard & Reports
user-invocable: true
---

# Dashboard & Reports

## Overview
Display business summary data, charts, sales/stock/financial reports.

## System Structure

### Dashboard
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/dashboard/` | Main dashboard + DB data |
| BLoC | `bloc/bi_report/` | BI report state management |
| Procurement | `screens/procurement_dashboard/` | Procurement dashboard |

### Reports
| Section | File | Purpose |
|---------|------|---------|
| Screen | `screens/report/` | Display reports |
| PDF | `screens/report/pdf_report_*` | Generate PDF |
| Stock Report | `screens/report/stock_report*` | Stock reports |
| BLoC | `bloc/report/` | State management |
| Export CSV | `bloc/export_csv/` | CSV export |

### Backend — 2 Report Engines
| Engine | File | Database | Best for |
|--------|------|----------|----------|
| reportqueryc | `backend/internal/report/reportqueryc/` | ClickHouse | Large data, analytics |
| reportquerym | `backend/internal/report/reportquerym/` | MongoDB | Current data |

### GoAPI Analytics
| Handler | Purpose |
|---------|---------|
| `handlers/reports/` | General reports |
| `handlers/sales_report/` | Sales reports |
| `handlers/stock_report_lookup/` | Stock lookup |
| `handlers/clickhouse/` | Direct ClickHouse queries |

## Report Types
| Report | Description |
|--------|-------------|
| Daily Sales | Daily sales summary |
| Sales by Product | Best/worst selling products |
| Sales by Employee | Individual performance |
| Stock Balance | Warehouse inventory levels |
| Low Stock Alert | Reorder notifications |
| Profit & Loss | Revenue - Cost - Expenses |
| Outstanding Payables | Amounts owed to suppliers |
| Outstanding Receivables | Amounts owed by customers |
| Procurement Report | PR → RFQ → PO tracking |

## Report Download
| Platform | Method | File |
|----------|--------|------|
| Web | Download via browser | `file_download_web.dart` |
| Mobile | Save to device | `file_download_mobile.dart` |
| Desktop | Save to disk | `file_download_desktop.dart` |

## Frontend (Flutter)

### BLoCs
| BLoC | Purpose |
|------|---------|
| `report_bloc` | General reports |
| `bi_report_bloc` | BI Analytics |
| `export_csv_bloc` | CSV export |

### Repositories
`report_repository`, `bi_report_repository`

### Services
`sales_report_api_service`, `result_table_api_service`, `pdf_service`

### BI Report Models (23 files)
`sale_daily_report_model`, `sale_report_summary`, `sale_return_model`, `gross_profit_by_document_model`, `gross_profit_by_product_model`, `payment_daily_model`, `purchase_partial_model`, `stock_balance_model`, `stock_movment_model`, `vat_buy_model`, `vat_sale_model` + summary models

## Important Notes
- Reports requiring **large data** → use **ClickHouse** (reportqueryc)
- Reports requiring **latest data** → use **MongoDB** (reportquerym)
- ClickHouse may have slight delay (data sync via Kafka)
- **Never use mock data** in reports — if API is not ready, show an error
- PDF generation uses fixed colors (`PdfColors.*`) — do not change based on theme
- CSV export uses a separate BLoC (`bloc/export_csv/`)
