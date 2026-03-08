# MCP Tools — การสร้างและ Register

## MCP คืออะไร

MCP (Model Context Protocol) เป็น tools bridge สำหรับ AI dev assistants (Claude Code, Cursor, VSCode)
ให้ AI อ่าน/เขียน data ผ่าน tools โดยตรง — ไม่ต้องอ่าน backend code

**Flutter app ไม่ได้เรียก MCP** — Flutter ใช้ REST API ปกติ

## Endpoints — General vs Dev

### General (สำหรับ Frontend dev)

```
GET  /goapi/mcp/health   ← ตรวจสอบสถานะ
GET  /goapi/mcp/tools    ← รายการ tools (เฉพาะ business data)
GET  /goapi/mcp/sse      ← SSE endpoint (Claude Code)
POST /goapi/mcp/invoke   ← เรียกใช้ tool โดยตรง
```

### Dev (สำหรับ Backend dev)

```
GET  /goapi/mcp/dev/health   ← ตรวจสอบสถานะ
GET  /goapi/mcp/dev/tools    ← รายการ tools ทั้งหมด (General + Dev)
GET  /goapi/mcp/dev/sse      ← SSE endpoint
POST /goapi/mcp/dev/invoke   ← เรียกใช้ tool โดยตรง
```

### Dev-only Tools (16 ตัว)

| Tool | หมวด |
|------|-------|
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

## สร้าง MCP Tool ใหม่ (3 ขั้นตอน)

### ขั้นตอน 1: สร้างไฟล์ใน `internal/goapi/mcp/tools/`

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

### ขั้นตอน 2: Register ใน `server.go` (3 จุด)

**จุดที่ 1 — AvailableTools (tool description):**
```go
var AvailableTools = []map[string]interface{}{
    {
        "name":        "get_feature_list",
        "description": "ดึงรายการ feature ทั้งหมดของร้านค้า",
        "parameters": map[string]interface{}{
            "shop_id": "string (required) - Shop ID",
            "keyword": "string (optional) - Search keyword",
            "limit":   "number (optional) - Max results (default: 50)",
        },
    },
}
```

**จุดที่ 2 — InvokeTool switch (HTTP handler):**
```go
case "get_feature_list":
    result, err = s.invokeGetFeatureList(ctx, req.Params)
```

**จุดที่ 3 — invoke function (parameter parsing):**
```go
func (s *MCPServer) invokeGetFeatureList(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    shopID := getStringParam(params, "shop_id")
    keyword := getStringParam(params, "keyword")
    limit := getIntParam(params, "limit")
    return tools.GetFeatureList(ctx, shopID, keyword, limit)
}
```

### Parameter Helpers ที่มีอยู่แล้ว
```go
getStringParam(params, "key")      // string
getIntParam(params, "key")         // int
getFloatParam(params, "key")       // float64
getBoolParam(params, "key")        // bool
getBoolPtrParam(params, "key")     // *bool (nil if not provided)
```

## Pattern: Formatted Names (สำคัญ)

ทุก tool ที่คืน names[] array → ควรเพิ่ม formatted string ด้วย เพื่อให้ AI อ่านง่าย

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

**กฏ:** ทุก MCP tool ที่คืน names[] → ต้องมี formatted field คู่กันเสมอ

## Guidelines

- Tool name ใช้ `snake_case` เช่น `get_low_stock_products`
- Description ควรบอกว่า return อะไร ไม่ใช่แค่ว่าทำอะไร
- ทุก tool ต้องรับ `shop_id` เป็น required parameter
- CRUD tools ต้อง publish Kafka event หลังเขียน MongoDB (ให้ sync ไป PG + CH)
- MongoDB = source of truth → tools เขียนลง MongoDB → Kafka sync ไป PG/CH

## Config ตัวอย่าง (Claude Code `~/.claude.json`)

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
