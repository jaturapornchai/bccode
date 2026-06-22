---
name: api-designer
description: Use when designing or reviewing REST APIs for BC Account.
---

## 1. REST Standards & Versioning
- **Structure**: Prefer Holding-first routes such as `/api/v1/holdings/:holdingcode/resources`. Legacy `/shops/...` routes may remain only as explicit compatibility paths during migration; do not introduce new `/shop` API contracts.
- **Scope Parameters**: Holding-owned access APIs use `holdingcode` as the tenant key and optional `businesscode` / `branchcode` scope filters. Validate company and branch membership server-side before reading or writing scoped records.
- **Database Naming**: New or changed MongoDB/PostgreSQL/ClickHouse contracts, database-facing function names, MCP tool/action names, API route/query/body keys tied to database fields, table/collection/index names, fields, and variables/constants that represent database identifiers must be lowercase (underscore allowed; snake_case OK — only uppercase/camelCase is a violation).
- **Accounting Decimal API**: For accounting/finance/stock/sales/purchase/tax/payment/cost/report fields, API requests and responses must carry money, decimal quantity, unit price, exchange rate, debit, credit, and totals as decimal strings, not JSON numbers. Validate decimal strings before persistence and keep project lowercase field names such as `netamount`, `unitprice`, `exchangerate`, and `amountsatang`.
- **Versioning**: Baseline is `v1`. Keep `v1` backward-compatible. Add side-by-side `v2` for breaking changes.
- **Payload Headers**: Send `X-BC-Required-Backend-Version`, `X-BC-Client-Platform`, and `X-BC-Client-Version` from clients.

## 2. API Format
- **Success**: `{"success": true, "data": ..., "meta": {"page": 1, "total": 100}}`
- **Error**: `{"success": false, "error": {"code": "...", "message": "...", "details": {}}}`
- **Security Check**: Verify JWT auth, tenant membership permission check, input validation, and rate limits on every route.
