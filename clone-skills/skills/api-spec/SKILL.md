---
name: api-spec
description: View API specification (request/response/parameters) — use to inspect endpoint details or generate models
user-invocable: true
---

# API Spec — View API Details

## Usage
`/api-spec <method> <path>` — View spec for a specific endpoint
`/api-spec <keyword>` — Search then display spec

## Steps

### 1. Get API Specification
Call MCP tool `get_api_spec`:
```
Tool: get_api_spec
Parameters:
  - method: "GET" | "POST" | "PUT" | "DELETE"
  - path: "/goapi/api/v1/..."
```

### 2. Get Request/Response Examples
Call MCP tool `get_api_example`:
```
Tool: get_api_example
Parameters:
  - method: "POST"
  - path: "/goapi/api/v1/chatbot/chat-agent"
```

### 3. Display Results
```
## Endpoint: POST /goapi/api/v1/chatbot/chat-agent

### Request Headers
- Authorization: Bearer <JWT token>
- Content-Type: application/json

### Request Body
{
  "message": "string — user message",
  "session_id": "string — session ID (optional)"
}

### Response (200 OK)
{
  "success": true,
  "data": {
    "response": "string — AI response",
    "tools_used": ["string"] — tools called by AI
  }
}

### curl Example
curl -X POST http://localhost:9090/goapi/api/v1/chatbot/chat-agent \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"message": "today sales"}'
```

### 4. Suggest Next Steps
- Generate Dart model → `/model-gen <model_name>`
- View enum values → `/enum-list <keyword>`
- Find other APIs → `/api-search <keyword>`

## Examples
```
/api-spec POST /goapi/api/v1/chatbot/chat-agent
/api-spec GET /goapi/api/health
/api-spec chatbot    → search first then display spec
```

## MCP Tools Used
| Tool | Purpose |
|------|---------|
| `get_api_spec` | View specification (params, body, response) |
| `get_api_example` | View real request/response examples |
| `list_api_endpoints` | Search endpoints (if path unknown) |

## Notes
- If path is unknown → use `/api-search` first
- These tools are on MCP Dev endpoint only
- URL base: `http://localhost:9090/goapi`
