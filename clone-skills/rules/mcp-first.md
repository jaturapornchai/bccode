# MCP-First Development

## Principle
MCP (Model Context Protocol) bridges frontend AI to backend without reading source code directly.

## Rules

### Cross-Project Access
**AI can read + edit both projects:**
- **Backend** `D:\bcdev\backend` — read + write
- **Frontend** `D:\bcdev\bcaiaccount` — read + write
- Cross-check freely, e.g., verify backend response then fix frontend model to match

### Test Project Rules (bcaitest01)
When working in `bcaitest01`:
- **Backend** `D:\bcdev\backend` — **read only, never edit**
- **Frontend** `D:\bcdev\bcaiaccount` — **read only, never edit**
- Always cross-check backend + frontend when testing or debugging
- On finding bugs via MCP tools → check source code for root cause → **report to Jead** (never fix directly)

### MCP Remains Essential For:
1. **API discovery**: `get_api_catalog`, `get_api_spec`
2. **Schema/enum lookup**: `get_model_schema`, `get_enum_catalog`
3. **New API requests**: Write spec in `prompts/api_requests/{feature}.md`

### Backend AI (Claude Code in backend)
1. Read + edit backend code directly
2. On receiving frontend prompt → build API per spec + add MCP tool if needed
3. MCP tools must cover every API that frontend requires

### MCP Endpoints — General vs Dev

| | General (Frontend dev) | Dev (Backend dev) |
|---|---|---|
| **SSE** | `GET /goapi/mcp/sse` | `GET /goapi/mcp/dev/sse` |
| **Tools** | `GET /goapi/mcp/tools` | `GET /goapi/mcp/dev/tools` |
| **Health** | `GET /goapi/mcp/health` | `GET /goapi/mcp/dev/health` |

- **General** — business tools only (sales, inventory, products, barcodes, units, etc.)
- **Dev** — all tools including database access, API catalog, model schema, enums
- **Dev-only tools (16):** `get_database_schema`, `execute_query`, `get_table_sample`, `query_mongodb`, `list_mongodb_collections`, `aggregate_mongodb`, `query_clickhouse`, `list_clickhouse_tables`, `execute_pg_command`, `execute_ch_command`, `list_api_endpoints`, `get_api_spec`, `get_api_example`, `list_enums`, `get_model_schema`, `rebuild_products`
- API Keys: `POST/GET/PUT/DELETE /goapi/api/mcp/keys` (same key for both endpoints)

### MCP Config Locations
- **Claude Code**: `~/.claude.json` → `mcpServers` → bc-erp (SSE)
  - Frontend dev: `/goapi/mcp/sse`
  - Backend dev: `/goapi/mcp/dev/sse`
- **Claude Desktop**: `%APPDATA%/Claude/claude_desktop_config.json` → `mcp-remote` global
- **Frontend project**: `.mcp.json` in project root (General endpoint)

## prompts/api_requests/ Pattern
When frontend needs a new API:
1. Frontend AI writes spec in `prompts/api_requests/{feature}.md`
2. Backend AI reads spec and builds API
3. Backend AI adds MCP tool if needed
4. Frontend AI verifies new API via MCP then implements
