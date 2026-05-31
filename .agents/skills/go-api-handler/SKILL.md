---
name: go-api-handler
description: Create Go API handlers following BC Account patterns.
---

- **Route Framework**: Use Gin. Route handlers must accept `*gin.Context`.
- **Response Format**:
  - Success: `{"success": true, "data": ...}`
  - Error: `{"success": false, "error": {"code": "...", "message": "..."}}`
- **Tenancy Scope**: Authorize `tenant_id` from JWT/session/workspace and map to storage physical key.
- **Validation**: Validate input structs with tags. Handle exceptions with custom types. Write unit tests.
