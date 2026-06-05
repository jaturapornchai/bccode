---
name: mcp-tool
description: Create new MCP tools for BC Account.
---

- **Language**: Use Go. Follow the MCP Spec (2025-11-25).
- **Annotations**: Include `readOnlyHint`, `destructiveHint`, and `idempotentHint`.
- **Validation & Multi-Tenant**: Validate input with JSON schema. Every tool reading data must filter by authorized `tenantid` or `holdingcode`.
- **Naming**: New MCP tool/action names, parameter keys, database-facing function names, and database-related variables/constants must be lowercase with no underscore, for example `searchproducts`, `getdailysales`, `querymongodb`, and `allowedtools`.
- **Attestation**: Write unit tests, write document tags, output error messages in English/Thai.
