---
description: Generate CRUD single pass. Uses thinking_level=low.
---

# Quick CRUD — thinking_level: LOW

Usage: `/qcrud <Resource> <field1:type> <field2:type> ...`

Example: `/qcrud Product name:string price:int stock:int category_id:uuid`

## Auto-generated
- Migration (PK, FK, common-filter indexes)
- Model + DTO
- Handler (list paginated, get, create, update, delete)
- Validator (derived from type)
- Test (table-driven: 1 happy + 2 error per endpoint)
- Route registered

## Smart skip
- Resource exists → ask "merge or replace?"
- Don't duplicate timestamp if in base
- `--no-db` flag → skip migration

## Output flow
1. Diff for review (NOT applied)
2. Wait "ok"
3. Apply + run test + paste output
4. If MCP db-mcp available → run migration auto

## Token budget: ~3-5k output (low thinking)
