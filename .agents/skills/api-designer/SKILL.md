---
name: api-designer
description: Use when designing or reviewing REST APIs for BC Account.
---

## 1. REST Standards & Versioning
- **Structure**: Prefer Holding-first routes such as `/api/v1/holdings/:holding_code/resources`. Legacy `/shops/...` routes may remain only as explicit compatibility paths during migration; do not introduce new `/shop` API contracts.
- **Scope Parameters**: Holding-owned access APIs use `holding_code` as the tenant key and optional `business_code` / `branch_code` scope filters. Validate company and branch membership server-side before reading or writing scoped records.
- **Versioning**: Baseline is `v1`. Keep `v1` backward-compatible. Add side-by-side `v2` for breaking changes.
- **Payload Headers**: Send `X-BC-Required-Backend-Version`, `X-BC-Client-Platform`, and `X-BC-Client-Version` from clients.

## 2. API Format
- **Success**: `{"success": true, "data": ..., "meta": {"page": 1, "total": 100}}`
- **Error**: `{"success": false, "error": {"code": "...", "message": "...", "details": {}}}`
- **Security Check**: Verify JWT auth, tenant membership permission check, input validation, and rate limits on every route.
