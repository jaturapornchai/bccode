# ERP Conventions — BC AI Cloud

## Platform
BC AI Cloud = Cloud-based ERP/Accounting system for retail businesses.

## Database Architecture
| Database | Role | Purpose |
|----------|------|---------|
| **MongoDB** | Main data store | Source of truth — documents, transactions, masters |
| **PostgreSQL** | Relational processing | Joins, aggregations, reports |
| **ClickHouse** | OLAP/Analytics | Time-series, BI dashboards, analytics |

## Key Concepts

### shop_id
- Every request requires `shop_id` (store identifier)
- Format: string (e.g., `"2jgDFkVsFdah2JnSMC89rM2eBMy"`)
- Every query must filter by shop_id

### Document Types
Sale Invoice, Sale Return, Purchase, Purchase Return, Stock Transfer, Stock Adjustment, Stock Balance, Quotation, Credit Note, Debit Note

### Master Data
- **Products** — barcode, name, price, unit
- **Units** — unit code, names (multi-language)
- **Customers** — customer code, name, contact
- **Suppliers** — supplier code, name, contact

## Port Standards
| Service | Port | Note |
|---------|------|------|
| MainAPI (gateway) | **8888** | Single entry point |
| MongoDB | 27017 | Native (not in Docker) |
| PostgreSQL | 5432 | Native (not in Docker) |
| Redis | 6379 | Docker internal |
| ClickHouse HTTP | 8123 | Docker internal |
| Kafka | 9092 | Docker internal |

## Domain
- VPS dev: `api.bcaicloud.com` → Hetzner
- Always use domain, never raw IP

## Multi-language
- System supports multi-language (TH, EN, etc.)
- Use `assets/language/languages.json` for i18n
- Multi-language names stored as `map[string]string` (`{"th": "...", "en": "..."}`)

## Backend-Frontend Coordination (Critical)

| Project | Path | Tech |
|---------|------|------|
| **Backend** | `D:\bcdev\backend` | Go (Echo v4) |
| **Frontend** | `D:\bcdev\bcaiaccount` | Flutter (BLoC + Dio) |

### Coordination Rules:
1. **Data structures must match** — backend response fields must align with frontend model
2. **Backend change → notify frontend** — if response format changes, create prompt in `prompts/api_requests/`
3. **Frontend change → verify backend** — use MCP tools (`get_api_spec`, `get_model_schema`) before editing
4. **JSON field names must match** — backend uses `json:"field_name"` (snake_case), frontend model must map accordingly
5. **Null safety** — backend fields with `omitempty` may be absent → frontend must handle nullable

### Warnings:
- Modifying backend response → **verify frontend still works**
- Adding new field → safe (old frontend still works)
- Removing/renaming field → **dangerous** (frontend breaks) → notify + fix frontend
- Changing field type (e.g., `int` → `string`) → **dangerous** → fix both sides
