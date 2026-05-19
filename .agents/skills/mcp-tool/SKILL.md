---
name: mcp-tool
description: >
  Create new MCP tools for BC Account.
  Use when: creating an MCP tool, new tool, or adding an MCP server.
  Trigger: "สร้าง MCP", "new tool", "create MCP"
---

When creating an MCP tool:

1. Use Go implementation
2. Follow MCP Protocol spec 2025-11-25
3. Include tool annotations: readOnlyHint, destructiveHint, idempotentHint
4. Validate with JSON Schema
5. Error messages in both Thai and English
6. Write comprehensive unit tests
7. Documentation in README.md
8. Every tool that reads customer data must filter by authorized `tenant_id`. Legacy storage may map `tenant_id` to `shop_id` / `shopid`.
