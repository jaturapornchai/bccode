---
name: api-designer
description: Use when designing or reviewing REST APIs for BC Account
---
## API Design Standards

### URL Structure
```
GET    /api/v1/tenants/:tenantId/products         # list
GET    /api/v1/tenants/:tenantId/products/:id     # detail
POST   /api/v1/tenants/:tenantId/products         # create
PUT    /api/v1/tenants/:tenantId/products/:id     # update
DELETE /api/v1/tenants/:tenantId/products/:id     # delete
```

Existing `/shops/:shopId/...` APIs are legacy-compatible aliases. New APIs should use `tenantId` in the route/request contract, but for existing data the `tenantId` value is the same value as core `shopid`. Map to the module's real physical field inside persistence code: usually `shopid`, sometimes `shop_id` for newer GoAPI/MCP DTOs. `tenantId` identifies the selected company/business, not the owner user; user access must come from membership/role checks.

### Version Compatibility
- Backend API contracts start at `v1`. Future breaking changes must add `v2`, `v3`, and so on; do not silently change `v1` behavior.
- Supported API versions must run side-by-side. If a `v2` route/contract is added, keep the matching `v1` route/contract callable until every dependent client is migrated.
- Version selection must be explicit per request through URL versioning and/or request metadata, never a global backend mode that disables old clients.
- Keep `v1` compatible for web, iOS, and Android clients that have not updated yet.
- Clients should send `X-BC-Required-Backend-Version`, `X-BC-Client-Platform`, and `X-BC-Client-Version` on API calls or during bootstrap/version checks.
- Version-aware backend responses should include active and supported backend versions when practical.
- Any API design or review must state whether it is additive-compatible with `v1` or requires a new version.

### Response Format
```json
{
  "success": true,
  "data": { },
  "meta": { "page": 1, "total": 100 },
  "error": null
}
```

### Error Format
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "message": "Product not found",
    "details": {}
  }
}
```

### Required Middleware for every route
- JWT Authentication
- Tenant permission check from authenticated workspace membership (`tenant_id`; legacy `shop_id/shopid` only as alias)
- Request logging
- Rate limiting (sensitive endpoints)
- Input validation
