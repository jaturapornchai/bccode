---
name: bc-account-expert
description: Use when working with the BC Account business system. Knows domain knowledge, business rules, and system architecture.
---

## 1. Domain & Business Scope
- **Domain**: Thai SME operations (Sales, Purchase, Inventory costing/average, Restaurant menu/service/table flows, Manufacturing/BOM consumption, General Ledger accounts/journal entries, AR/AP aging).
- **Core Workflow Documents**: Quote/Order/Invoice/Receipt, Purchase Order/Bill/Payment, Stock receipt/transfer/count, POS, debitor/creditor aging, GL posting, and audits.
- **Rules Inspection**: Read active repo source, legacy Flutter screen code (in `D:\bcdev`), and DEV database state before altering accounting, inventory, tax, sales, POS, AR/AP, or GL models.

## 2. Holding, Company, Branch Structure
- **Boundary**: `holding_code` is the root tenant/workspace boundary. A Holding can contain many companies, and each company can contain many branches.
- **Local Model Authority**: For new/changed ERP model contracts, use the active `D:\bccode` source, runtime evidence, and local project rules. Do not depend on external model-document folders.
- **Company Scope**: `business_code` is the user-facing company code and must be normalized to uppercase before validation, duplicate checks, search, save, and sync.
- **Branch Scope**: `branch_code` is scoped under `holding_code + business_code`, normalized to 5 digits, and owns branch legal/tax/calendar settings. Departments, working days, and holidays are branch-scoped.
- **Access Scope**: Users, screen permissions, permission groups, user permission assignments, and approval rights are Holding-owned records under `holding_code`, with explicit `access_scopes[]`, `scope_rules[]`, or `approval_rules[]` for `holding`, `company`, or `branch` applicability.

## 2.1 Data Store Roles
- **MongoDB**: the only authoritative operational source for all CRUD, documents, master data, settings, transactions, and user-entered business data.
- **Default Source**: When storage is not explicitly specified by Jead, use MongoDB for operational create/read/update/delete/list/detail flows. Do not read business master data, settings, permissions, approvals, companies, branches, users, products, or transactions from PostgreSQL or ClickHouse just because a projection exists.
- **Operational CRUD Pipeline**: Operational writes flow `MongoDB -> Kafka -> PostgreSQL -> ClickHouse`. Write MongoDB first, emit Kafka or a durable MongoDB outbox event, let projection consumers rebuild PostgreSQL, then feed ClickHouse for BI/reporting. Do not let user-facing CRUD write PostgreSQL or ClickHouse directly.
- **Frontend CRUD Mutations**: Frontend create/edit/delete flows call MongoDB-backed operational APIs only. Projection or BI endpoints are never mutation sources for normal screens; Kafka and downstream projection sync are backend responsibilities.
- **Cloudflare R2/S3**: only binary/image/file object storage. MongoDB keeps metadata and private paths.
- **PostgreSQL**: rebuildable relational processing/projection store for postings, balances, VAT/tax, AR/AP, GL, strict relational calculations, and integration-ready relational outputs.
- **ClickHouse**: rebuildable BI/analytics/reporting store fed from processed facts. Never treat ClickHouse as transactional source of truth.
- **Projection Conflict Rule**: If PostgreSQL or ClickHouse differs from MongoDB, MongoDB wins. Fix sync/rebuild code or data pipelines instead of treating projections as operational truth.
- **Product Classification**: `item_type` is 0=Stock, 1=Service, 2=Set, 3=Not Stock. `materialtype` is 0=General, 1=Material, 2=Semi-Finished, 3=Set, 4=Agricultural. Product Set records must use `item_type=2` together with `materialtype=3` in MongoDB and relational projections; API writes and projection consumers must reject mismatched Set classification instead of correcting it silently.
- **Product Unit Access**: Product unit master data uses `units.business_codes` for company-level availability. Do not model or edit branch-level access on product units. Products, barcodes, and business documents reference units by `unit_code`; `guid_fixed` remains the immutable CRUD/sync identity. Read legacy `company_guids` or `unitcode` only as transition aliases.
- **Product Stock, Cost, Marketplace Stock, and Dimension Price**: Accounting stock belongs to product-level stock records and must support product total balance, warehouse-level balance, and storage-location-level balance. Costing must calculate from accounting stock only: normally one cost per product; when the explicit warehouse-cost option `cost_by_warehouse` is enabled, calculate cost per warehouse first and aggregate back to product cost. Do not calculate inventory cost by marketplace dimensions such as color or size. Marketplace stock is a separate availability projection that may expose product-level available balance and product-dimension available balance such as red, XL, black 128GB, SIM package, network, lot, or serial dimensions. Store per-marketplace SKU dimension availability under `marketplace_sku_mappings[].marketplace_dimension_stocks[]`; use it for channel availability/sync only, not stock deduction, cost, or GL. Selling prices may be defined at product, barcode/SKU, price-level, marketplace, and detailed dimension level; create/use a separate dimension price table/model when the normal product/barcode price array cannot represent marketplace variants clearly. Barcode/SKU rows are sellable identifiers for scanning, units, prices, marketplace mapping, images, and lookup only; they must not own authoritative accounting `qty`, `balance_qty`, average cost, or stock ledger balance. Existing barcode stock/balance fields are legacy/projection/display compatibility fields only and must not be used as the source for stock deduction or costing.
- **Product Variant Copy Workflow**: Repeated option patterns such as color > size, size > color, storage > color, flavor > pack, or package > network should be duplicated through `เพิ่ม (Copy)` from an existing `product_variant_matrix` row when possible. Copy opens create mode and carries option tiers, SKU rows, media, specs, import mappings, and marketplace-related setup from the selected row so the user only adjusts what differs.

## 3. Legal & Settings Rules
- **Head Office**: Branch code `00000` with Thai name `สำนักงานใหญ่`. Pad/normalize branch code inputs to 5 digits (e.g. `1` -> `00001`). Do not delete `00000`.
- **Branch Details**: Branch records own legal/tax settings: tax ID, registration number, VAT status/rate, company names, base currency, timezone, decimal configurations, roundings, and business flags.
- **Thailand Address**: Selected provinces, districts, subdistricts must use codes. Zip codes must recompute dynamically. Use address API served by backend dataset.

## 4. Tax & Legal Knowledge (Local-First + Cache)
- **No news/web search by default.** Use active repo source, local docs, tests, DEV/runtime evidence, and `tax-legal-cache.md` first.
- **Never guess.** If the cache and local evidence are not enough for tax/VAT/WHT/e-Tax/GL/statutory behavior, mark the point unverified and ask Jead before using external sources or changing business logic.
- **External verification is opt-in.** Only when Jead explicitly asks, use authoritative sources such as `rd.go.th`, Thai tax law, official specs, or known Thai accounting/ERP references, then cite the source/date.
