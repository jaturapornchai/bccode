---
name: mcp-tool
description: Create new MCP tools for BC Account.
---

- **Language**: Use Go. Follow the MCP Spec (2025-11-25).
- **Annotations**: Include `readOnlyHint`, `destructiveHint`, and `idempotentHint`.
- **Validation & Multi-Tenant**: Validate input with JSON schema. Every tool reading data must filter by authorized `tenant_id` (physically mapped to shop ID).
- **Attestation**: Write unit tests, write document tags, output error messages in English/Thai.
