---
name: crud-fast
description: Generate CRUD endpoint with model, handler, validator, test. Trigger on "add CRUD", "create endpoint", "new resource", "scaffold api". Uses thinking_level=low.
---

# CRUD Generation (Gemini 3.5 Flash — low thinking)

## Set thinking_level: low (boilerplate doesn't need deep reasoning)

## Token-saving flow
1. Ask ONLY missing fields. Skip if obvious.
2. ONE pass — no preview, no "should I?".
3. Output diff format.
4. **Schema-first output** — return structured spec, not prose

## Output structure (JSON-like)
```json
{
  "resource": "<name>",
  "fields": [
    {"name": "<field_name>", "type": "<type>", "required": true, "validation": "<rules>"}
  ],
  "files": [
    "migrations/<timestamp>_<resource>.sql",
    "internal/models/<resource>.go",
    "internal/handlers/<resource>.go",
    "internal/validators/<resource>.go",
    "internal/handlers/<resource>_test.go"
  ]
}
```

## Generation order
1. Migration (PK, FK indexes, filtered cols)
2. Model
3. Handler — list (paginated), get, create, update, delete
4. Validator (at handler boundary)
5. Test — table-driven, 1 happy + 2 error per endpoint

## Patterns
- Pagination: `limit` (def 20, max 100), `offset`
- Soft delete: `deleted_at`
- Audit: `created_at`, `updated_at` auto
- Errors: 400/404/409/500
- Response: `{data, meta:{total,page,limit}}`

## REJECT
- ❌ All rows without pagination
- ❌ N+1 in list
- ❌ Logic in handler
- ❌ SELECT *

## Use MCP if available
- db-mcp → run migration directly after gen
- github-mcp → open PR with diff
