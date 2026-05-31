---
name: bc-account-expert
description: Use when working with the BC Account business system. Knows domain knowledge, business rules, and system architecture.
---

## 1. Domain & Business Scope
- **Domain**: Thai SME operations (Sales, Purchase, Inventory costing/average, Restaurant menu/service/table flows, Manufacturing/BOM consumption, General Ledger accounts/journal entries, AR/AP aging).
- **Core Workflow Documents**: Quote/Order/Invoice/Receipt, Purchase Order/Bill/Payment, Stock receipt/transfer/count, POS, debitor/creditor aging, GL posting, and audits.
- **Rules Inspection**: Read active repo source, legacy Flutter screen code (in `D:\bcdev`), and DEV database state before altering accounting, inventory, tax, sales, POS, AR/AP, or GL models.

## 2. Multi-Tenant Structure
- **Boundary**: `tenant_id` = one company/business/legal entity/workspace. One user can access many tenants through memberships/roles.
- **Physical Keys**: Core physical storage uses `shopid` (Postgres, Mongo collections, Kafka, ClickHouse). GoAPI/MCP DTOs use `shop_id`. Maintain correct logical-to-physical mapping.
- **Branch Scope**: `branch_id` / branch code is scoped under `tenant_id`. Departments, working days, and holidays are branch-scoped (branches can have different calendars, timezones, and calendars).

## 2.1 Data Store Roles
- **MongoDB**: authoritative operational source for all CRUD, documents, master data, and user-entered business data.
- **Default Source**: When storage is not explicitly specified by Jead, use MongoDB for operational create/read/update/delete/list/detail flows. Do not read business master data from PostgreSQL or ClickHouse just because a projection exists.
- **Cloudflare R2/S3**: only binary/image/file object storage. MongoDB keeps metadata and private paths.
- **PostgreSQL**: relational processing/projection store for postings, balances, VAT/tax, AR/AP, GL, and strict relational calculations.
- **ClickHouse**: BI/analytics/reporting store fed from processed facts. Never treat ClickHouse as transactional source of truth.
- **Product Classification**: `item_type` is 0=Stock, 1=Service, 2=Set, 3=Not Stock. `materialtype` is 0=General, 1=Material, 2=Semi-Finished, 3=Set, 4=Agricultural. Product Set records must use `item_type=2` together with `materialtype=3` in MongoDB and relational projections; API writes and projection consumers must reject mismatched Set classification instead of correcting it silently.

## 3. Legal & Settings Rules
- **Head Office**: Branch code `00000` with Thai name `สำนักงานใหญ่`. Pad/normalize branch code inputs to 5 digits (e.g. `1` -> `00001`). Do not delete `00000`.
- **Branch Details**: Branch records own legal/tax settings: tax ID, registration number, VAT status/rate, company names, base currency, timezone, decimal configurations, roundings, and business flags.
- **Thailand Address**: Selected provinces, districts, subdistricts must use codes. Zip codes must recompute dynamically. Use address API served by backend dataset.

## 4. Tax & Legal Knowledge (Research-First + Cache)
- **Trigger = uncertainty.** Whenever you are unsure about any domain/business/accounting knowledge — and always before building/changing a tax/VAT/WHT/e-Tax/GL/statutory feature — research authoritative sources FIRST, then cache findings. Never guess, never skip. Full rule: core-rules "Tax & Accounting Correctness (Research-First)".
- **Sources**: กรมสรรพากร `rd.go.th`; Revenue Code + Royal Decrees + ministerial regs; competitor Thai ERP (FlowAccount, PEAK, Express, BusinessPlus, SML, Xero TH).
- **Verified knowledge cache — read this BEFORE re-searching**: `tax-legal-cache.md` (same folder). Append dated, source-linked entries after every new research pass; never guess a rate/field/form.
