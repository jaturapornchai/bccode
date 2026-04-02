---
name: api-search
description: Search API endpoints from backend via MCP — use to find available endpoints before calling /api-spec
user-invocable: true
---

# API Search — Find API Endpoints

## Usage
`/api-search <keyword>` — Search APIs matching keyword in path or description
`/api-search` — List all APIs

## Steps

### 1. Search API Endpoints
Call MCP tool `list_api_endpoints` via MCP dev endpoint:
```
Tool: list_api_endpoints
Parameters: (none — returns all)
```

### 2. Filter by Keyword
- Search in path, method, description, tags
- Case-insensitive matching
- No keyword → show all grouped by category

### 3. Display Results
Show as table:
| Method | Path | Description |
|--------|------|-------------|
| GET | /goapi/api/health | Health check |
| POST | /goapi/api/v1/chatbot/chat-agent | AI chatbot |

### 4. Suggest Next Steps
- View details → `/api-spec <method> <path>`
- View examples → use MCP tool `get_api_example`
- Generate model → `/model-gen <model_name>`

## Examples
```
/api-search chatbot     → chatbot APIs
/api-search product     → product APIs
/api-search mcp         → MCP endpoints
/api-search stock       → stock APIs
/api-search sale        → sales APIs
```

## MCP Tools Used
| Tool | Endpoint | Purpose |
|------|----------|---------|
| `list_api_endpoints` | Dev only | List all APIs |
| `get_api_spec` | Dev only | View API details |
| `get_api_example` | Dev only | View request/response examples |

## Notes
- Requires MCP server running (`GET http://localhost:9090/goapi/mcp/dev/health`)
- `list_api_endpoints` is on Dev endpoint only (not General)
- URL base: `http://localhost:9090/goapi`
