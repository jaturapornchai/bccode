---
name: api-spec
description: >
  View API specification (request/response/parameters) — use to inspect endpoint details or
  generate models. Use when the user wants to know request body, response format, or parameters
  of an API. Uses MCP get_api_spec tool.
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

## When to Use
- Need to know the request body structure before writing a Dart API call
- Need to view response format to create a model or parse JSON
- Need a curl example to manually test an endpoint
- Need to know if an endpoint requires an Authorization header
- Generating a Dart model and need the complete field list

## Anti-Patterns
- Do NOT guess request body fields from convention or intuition — always use `get_api_spec`
- Do NOT read Go handler source directly to guess structs — MCP provides more accurate information
- Do NOT hardcode response fields without checking spec — backend may add/change fields
- Do NOT skip `get_api_example` — examples show edge cases that spec may not cover
- Do NOT call `/api-spec` without knowing the path — use `/api-search` first

## Related Skills
- `/api-search` — Find endpoint path before viewing spec
- `/model-gen` — Generate Dart class from the response schema
- `/enum-list` — View enum values for fields that are enum types
