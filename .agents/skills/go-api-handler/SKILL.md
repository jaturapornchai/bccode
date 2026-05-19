---
name: go-api-handler
description: >
  Create Go API handlers following BC Account patterns.
  Use when: creating a new endpoint, handler, API, or route.
  Trigger: "สร้าง API", "เพิ่ม endpoint", "create handler", "new route"
---

When creating a new Go API handler for BC Account:

1. Use the Gin framework
2. Every handler must accept *gin.Context
3. Use the standard response format:
   - Success: {"success": true, "data": ...}
   - Error: {"success": false, "error": {"code": "...", "message": "..."}}
4. Every handler must derive and authorize `tenant_id` from JWT/session/workspace membership. Legacy storage may map `tenant_id` to `shop_id` / `shopid`.
5. Use service layer pattern: Handler -> Service -> Repository
6. Handle errors with custom error types
7. Validate input with struct tags
8. Always write tests alongside the handler
