---
name: bc-account-expert
description: Use when working with the BC Account ERP system. Knows domain knowledge, business rules, and system architecture.
---
## BC Account Domain Knowledge

### Business Context
- ERP/Accounting/POS for Thai SMEs — 2,500+ shops
- Primary customers: construction materials, wholesale
- Multi-tenant boundary: `tenant_id` = one company/business/workspace. One owner/user can access many tenants through membership/roles, and each tenant can have many branches.
- Existing `shop_id` / `shopid` values are legacy aliases for `tenant_id` during migration.

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
