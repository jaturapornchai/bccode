# MCP Tools — Creation & Registration

## What is MCP?

MCP (Model Context Protocol) is a tools bridge for AI dev assistants (Claude Code, Cursor, VSCode), allowing AI to read/write data directly via tools — without reading backend code.

**Flutter app does not call MCP** — Flutter uses regular REST APIs.

## Endpoints — General vs Dev

### General (for Frontend dev)

```
GET  /goapi/mcp/health   <- Health check
GET  /goapi/mcp/tools    <- Tool list (business data only)
GET  /goapi/mcp/sse      <- SSE endpoint (Claude Code)
POST /goapi/mcp/invoke   <- Direct tool invocation
```

### Dev (for Backend dev)

```
GET  /goapi/mcp/dev/health   <- Health check
GET  /goapi/mcp/dev/tools    <- All tools (General + Dev)
GET  /goapi/mcp/dev/sse      <- SSE endpoint
POST /goapi/mcp/dev/invoke   <- Direct tool invocation
```

### Dev-only Tools (16)

| Tool | Category |
|------|----------|
| `get_database_schema` | Database |
| `execute_query` | Database |
| `get_table_sample` | Database |
| `query_mongodb` | MongoDB |
| `list_mongodb_collections` | MongoDB |
| `aggregate_mongodb` | MongoDB |
| `query_clickhouse` | ClickHouse |
| `list_clickhouse_tables` | ClickHouse |
| `execute_pg_command` | Database (dangerous) |
| `execute_ch_command` | Database (dangerous) |
| `list_api_endpoints` | API Catalog |
| `get_api_spec` | API Catalog |
| `get_api_example` | API Catalog |
| `list_enums` | API Catalog |
| `get_model_schema` | API Catalog |
| `rebuild_products` | Sync |

## Creating a New MCP Tool (3 steps)

### Step 1: Create file in `internal/goapi/mcp/tools/`

```go
package tools

import (
    "context"
    "fmt"
    "time"

    "smlcloudplatform/internal/goapi/logger"
    myGlobal "smlcloudplatform/internal/goapi/myglobal"

    "go.mongodb.org/mongo-driver/bson"
)

type FeatureResponse struct {
    Success     bool      `json:"success"`
    Message     string    `json:"message"`
    Data        []Item    `json:"data"`
    Count       int       `json:"count"`
    GeneratedAt time.Time `json:"generated_at"`
}

func GetFeatureList(ctx context.Context, shopID, keyword string, limit int) (*FeatureResponse, error) {
    if shopID == "" {
        return nil, fmt.Errorf("shop_id is required")
    }
    // ... query logic ...
    logger.Info("[MCP GetFeatureList] shopID=%s, count=%d", shopID, len(results))
    return &FeatureResponse{...}, nil
}
```

### Step 2: Register in `server.go` (3 locations)

**Location 1 — AvailableTools (tool description):**
```go
var AvailableTools = []map[string]interface{}{
    {
        "name":        "get_feature_list",
        "description": "Get all features for a shop",
        "parameters": map[string]interface{}{
            "shop_id": "string (required) - Shop ID",
            "keyword": "string (optional) - Search keyword",
            "limit":   "number (optional) - Max results (default: 50)",
        },
    },
}
```

**Location 2 — InvokeTool switch (HTTP handler):**
```go
case "get_feature_list":
    result, err = s.invokeGetFeatureList(ctx, req.Params)
```

**Location 3 — invoke function (parameter parsing):**
```go
func (s *MCPServer) invokeGetFeatureList(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    shopID := getStringParam(params, "shop_id")
    keyword := getStringParam(params, "keyword")
    limit := getIntParam(params, "limit")
    return tools.GetFeatureList(ctx, shopID, keyword, limit)
}
```

### Available Parameter Helpers
```go
getStringParam(params, "key")      // string
getIntParam(params, "key")         // int
getFloatParam(params, "key")       // float64
getBoolParam(params, "key")        // bool
getBoolPtrParam(params, "key")     // *bool (nil if not provided)
```

## Pattern: Formatted Names (important)

Every tool returning a names[] array should include a formatted string for AI readability.

```go
func getThaiName(names []BarcodeNameEntry) string {
    for _, n := range names {
        if n.Code == "th" && n.Name != "" {
            return n.Name
        }
    }
    for _, n := range names {
        if n.Name != "" {
            return n.Name
        }
    }
    return ""
}
```

**Rule:** Every MCP tool returning names[] must include a formatted field alongside it.

## Guidelines

- Tool names use `snake_case` e.g. `get_low_stock_products`
- Description should state what is returned, not just what is done
- Every tool must accept `shop_id` as required parameter
- CRUD tools must publish Kafka event after writing to MongoDB (to sync to PG + CH)
- MongoDB = source of truth -> tools write to MongoDB -> Kafka syncs to PG/CH

## Config Example (Claude Code `~/.claude.json`)

```json
{
  "mcpServers": {
    "bc-erp": {
      "type": "sse",
      "url": "http://localhost:8888/goapi/mcp/sse",
      "headers": { "x-api-key": "YOUR_API_KEY" }
    }
  }
}
```
