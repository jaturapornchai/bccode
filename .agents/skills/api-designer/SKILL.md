---
name: api-designer
description: Use when designing or reviewing REST APIs for BC Account.
---

## 1. REST Standards & Versioning
- **Structure**: `/api/v1/tenants/:tenantId/resources` (alias to legacy `/shops/:shopId/...`). Map logical `tenantId` to physical `shopid` / `shop_id` in code.
- **Versioning**: Baseline is `v1`. Keep `v1` backward-compatible. Add side-by-side `v2` for breaking changes.
- **Payload Headers**: Send `X-BC-Required-Backend-Version`, `X-BC-Client-Platform`, and `X-BC-Client-Version` from clients.

## 2. API Format
- **Success**: `{"success": true, "data": ..., "meta": {"page": 1, "total": 100}}`
- **Error**: `{"success": false, "error": {"code": "...", "message": "...", "details": {}}}`
- **Security Check**: Verify JWT auth, tenant membership permission check, input validation, and rate limits on every route.
