# Cross-System Workflow — Frontend <-> Backend

## Architecture Overview

```
+---------------------+   REST API (Dio)   +------------------------+
|   Flutter App        | <---------------> |   Backend (Go)          |
|   (bcaiaccount)      |   port 8888       |   MainAPI + GoAPI       |
|                      |                   |                         |
|  AI works here       |   MCP             |  AI reads/edits here    |
|  Read/edit code      | <---------------> |                         |
|                      |  General: /mcp/   |                         |
|                      |  Dev: /mcp/dev/   |                         |
+---------------------+                   +------------------------+
          |
          | If API missing
          v
+---------------------+
| prompts/api_requests/ |  Send to backend team
| {feature}.md          |  ------------------>  backend creates API
+---------------------+
```

## GoAPI URL Mapping

GoAPI is integrated into MainAPI (port 8888) — uses `/goapi` prefix.

| Old | New |
|-----|-----|
| `goapi:9091/api/health` | `localhost:8888/goapi/api/health` |
| `goapi:9091/get` | `localhost:8888/goapi/get` |
| `goapi:9091/genpdf` | `localhost:8888/goapi/genpdf` |
| `goapi:9091/mcp/sse` | `localhost:8888/goapi/mcp/sse` (General) |
| `goapi:9091/mcp/dev/sse` | `localhost:8888/goapi/mcp/dev/sse` (Dev) |

**MainAPI routes (no /goapi prefix):** `/login`, `/register`, `/shop`, `/product/*`, `/healthz`

## MCP-First Rule (for Flutter dev)

**Before writing any code**, use MCP to check if the API already exists.

### MCP has 2 endpoint sets:

| Endpoint | URL | Visible Tools |
|----------|-----|---------------|
| **General** | `/goapi/mcp/sse` | Business data only |
| **Dev** | `/goapi/mcp/dev/sse` | All tools (General + database + API catalog) |

- Frontend dev -> use General (`/goapi/mcp/sse`)
- Backend dev -> use Dev (`/goapi/mcp/dev/sse`)
- Auth: `x-api-key` header

### Data Lookup Priority
1. **Use MCP tools first** — faster, returns real runtime data
2. If MCP insufficient -> read backend source code
3. Never create Dart models by guessing field names

### When Creating a New Model
```
1. MCP (Dev): get_database_schema -> view columns + types
2. MCP (Dev): get_table_sample -> view sample data
3. Create Dart model matching actual schema
```

## When Required API Does Not Exist

**Never create APIs yourself** — write a spec for the backend team.

### API Specification Prompt Format

Save to `bcaiaccount/prompts/api_requests/{feature}.md`:

```markdown
# API Request: {feature name}

## What is needed
{Describe the API needed and why}

## Endpoint
- Method: POST
- Path: /goapi/api/{resource}/{action}
- Auth: Bearer token (JWT)

## Request Body
{json schema}

## Expected Response
{json schema}

## Use Case
{Which screen, what action}

## Related Tables
{From MCP get_database_schema}
```

## Security Rules (mandatory)

- **Never hardcode AI API keys** in Flutter — must go through backend
- Flutter must not call `api.groq.com`, `openrouter.ai`, `generativelanguage.googleapis.com` directly
- Flutter must not call MCP endpoints directly — MCP is for AI dev tools only

## Post-Task Summary (mandatory)

After completing a feature or bug fix, always report:
- **No backend changes needed** — frontend complete + backend has API/fields ready
- **Backend changes needed** — API missing or backend needs new fields/logic -> spec at `prompts/api_requests/{feature}.md`
