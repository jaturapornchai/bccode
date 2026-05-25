---
name: bc-account-expert
description: Use when working with the BC Account business system. Knows domain knowledge, business rules, and system architecture.
---
## BC Account Domain Knowledge

### Business Context
- Accounting/POS for Thai SMEs — 2,500+ shops
- Primary customers: construction materials, wholesale
- Multi-tenant boundary: `tenant_id` = one company/business/workspace. One owner/user can access many tenants through membership/roles, and each tenant can have many branches.
- Existing `shop_id` / `shopid` values are legacy aliases for `tenant_id` during migration.
- Company records are intentionally minimal in the UI: company name and company address only. Operational/legal/document settings belong to branch records.
- Every company must have at least one branch. For Thailand tax/VAT branch numbering, the default head office branch is code `00000` with Thai name `สำนักงานใหญ่` (head office). Do not use `00001` for head office; `00001` means the first branch office. Persist Thai branch codes as normalized five-digit strings; backend branch create/update/import must enforce normalization, and normal CRUD must not delete the `00000` head-office branch.
- Branch records own document company names, branch names, branch contact/address parts, branch phone, logo/image, country, tax ID, company registration number, VAT status/rate/type, base currency, timezone/date/year/decimal settings, business type, payment rounding, point config, machine type, coupon use type, departments, and business-property flags.

### Core Modules
- GL: General Ledger (Chart of Accounts, Journal)
- AR: Accounts Receivable (Invoice, Receipt)
- AP: Accounts Payable (PO, Bill, Payment)
- INV: Inventory (Product, Stock, Moving Average/FIFO)
- POS: Point of Sale
- PR/RFQ: Purchase Requisition / Request for Quotation

### Pricing & Costing
- Moving Average Cost (default)
- FIFO / FEFO for perishables
- Tiered pricing (wholesale/retail/special)

### Integration Points
- LINE OA: via OpenClaw/HiClaw
- MCP Tools: 66+ tools for AI agent
- ESL: Electronic Shelf Label

### Architecture
- Backend: Go/Gin -> REST API
- Frontend target: Next.js in `D:\bccode\frontend` from `https://github.com/jaturapornchai/bccode`
- Frontend migration reference/template: Flutter in `D:\bcdev\frontend` from `https://github.com/jaturapornchai/bcdev` (especially `bcaiaccount`)
- DB: PostgreSQL tenant-scoped data + ClickHouse analytics + MongoDB documents
- AI: Codex Agent SDK + MCP
