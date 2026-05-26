---
name: bc-account-expert
description: Use when working with the BC Account business system. Knows domain knowledge, business rules, and system architecture.
---
## BC Account Domain Knowledge

### Business Context
- BC Ai Account is a Thai SME business platform, not a generic CRUD/admin system.
- Agents using this skill must act as domain experts in Thai SME accounting and operations: accounting, marketing, sales, purchasing, trading/distribution, restaurant operations, light manufacturing, general ledger, inventory accounting, accounts receivable, accounts payable, tax/VAT-aware workflows, branch/company operations, reporting, and auditability.
- Existing users include Thai SME shops. Important customer patterns include construction materials, wholesale/trading, retail/POS, restaurant, and other SME workflows.
- Multi-tenant boundary: `tenant_id` = one company/business/workspace. One owner/user can access many tenants through membership/roles, and each tenant can have many branches.
- Existing `shop_id` / `shopid` values are legacy aliases for `tenant_id` during migration.
- Company records are intentionally minimal in the UI: company name and company address only. Operational/legal/document settings belong to branch records.
- Every company must have at least one branch. For Thailand tax/VAT branch numbering, the default head office branch is code `00000` with Thai name `สำนักงานใหญ่` (head office). Do not use `00001` for head office; `00001` means the first branch office. Persist Thai branch codes as normalized five-digit strings; backend branch create/update/import must enforce normalization, and normal CRUD must not delete the `00000` head-office branch.
- Branch records own document company names, branch names, branch contact/address parts, branch phone, logo/image, country, tax ID, company registration number, VAT status/rate/type, base currency, timezone/date/year/decimal settings, business type, payment rounding, point config, machine type, coupon use type, departments, and business-property flags.

### Business Design Rules
- Model features around real operating workflows and documents, not standalone storage tables.
- Preserve document lifecycle, numbering, tax/VAT treatment, stock impact, accounting impact, permissions, reports, and traceability.
- Before changing accounting, inventory, sales, purchase, AR/AP, restaurant, production, tax, or reporting behavior, inspect active source, legacy Flutter behavior when relevant, data contracts, and real DEV behavior.
- Surface business impact and compatibility risk before changing schemas, APIs, GL postings, stock movement, tax logic, or report contracts.

### Core Modules
- GL: General Ledger (Chart of Accounts, Journal)
- AR: Accounts Receivable (customer, invoice, receipt, aging, credit control)
- AP: Accounts Payable (supplier, purchase order, bill, payment, aging)
- Sales: quotation, sale order, invoice, receipt, POS, pricing, promotion, customer workflow
- Purchasing: purchase request, RFQ, purchase order, receiving, bill, supplier payment
- INV: Inventory, stock card, warehouse/location, stock receiving/issuing/transfer/counting, product costing
- Restaurant: menu/category, table/order flow, kitchen/service flow, bill/receipt, promotion
- Manufacturing: BOM/recipe, material consumption, finished goods receipt, production cost
- GL: General Ledger, chart of accounts, journal, posting, financial statements
- Marketing: customer segmentation, promotion, coupon, campaign, loyalty/points

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
