---
name: mcp-check
description: Check MCP server status, available tools, API key — use when MCP has issues or tool calls fail
user-invocable: true
---

# MCP Check — Verify MCP Status

## Usage
`/mcp-check` — Check both General and Dev MCP servers
`/mcp-check tools` — List available tools

## Steps

### 1. Check Health
```bash
# General endpoint
GET http://localhost:9090/goapi/mcp/health

# Dev endpoint
GET http://localhost:9090/goapi/mcp/dev/health
```

### 2. Check Available Tools
```bash
# General tools (business data — safe for AI agents)
GET http://localhost:9090/goapi/mcp/tools

# Dev tools (all — includes database access)
GET http://localhost:9090/goapi/mcp/dev/tools
```

### 3. Display Results
```
## MCP Server Status

| Endpoint | Status | URL |
|----------|--------|-----|
| General | ✅ Online | http://localhost:9090/goapi/mcp/health |
| Dev | ✅ Online | http://localhost:9090/goapi/mcp/dev/health |

## Available Tools

### General Tools (25 — safe for AI)
| Category | Count | Examples |
|----------|-------|---------|
| Sales | 5 | get_daily_sales, get_top_selling_products |
| Dashboard | 3 | get_dashboard_kpis, get_business_health |
| Inventory | 5 | get_inventory_value, get_low_stock_alerts |
| Finance | 5 | get_profit_analysis, get_cash_flow |
| Customer | 3 | get_top_customers, get_customer_segments |
| Master Data | 4 | list_units, search_products |

### Dev-Only Tools (16 — database access)
| Category | Count | Examples |
|----------|-------|---------|
| PostgreSQL | 4 | get_database_schema, execute_query |
| MongoDB | 3 | query_mongodb, aggregate_mongodb |
| ClickHouse | 3 | query_clickhouse, list_clickhouse_tables |
| API/Schema | 6 | list_api_endpoints, get_model_schema |
```

### 4. Check API Key (if applicable)
- API Key format: `bc_live_{32chars}`
- Auth: Header `X-API-Key` or Query `?api_key=`

## Troubleshooting

### MCP Not Responding
1. Verify backend is running: `GET http://localhost:9090/goapi/api/health`
2. Check port — use **9090** (not 8888 or 9091)
3. GoAPI is merged into MainAPI — always use `/goapi` prefix

### MCP Tool Not Working
1. Verify correct endpoint (General vs Dev)
2. Dev tools only work on `/goapi/mcp/dev/` endpoint
3. Check API key permissions

## Endpoint Summary
| Endpoint | Port | Path | Purpose |
|----------|------|------|---------|
| Health (General) | 9090 | `/goapi/mcp/health` | Status check |
| Health (Dev) | 9090 | `/goapi/mcp/dev/health` | Dev status check |
| Tools (General) | 9090 | `/goapi/mcp/tools` | Business tools |
| Tools (Dev) | 9090 | `/goapi/mcp/dev/tools` | All tools |
| SSE (General) | 9090 | `/goapi/mcp/sse` | Claude Desktop/AI |
| SSE (Dev) | 9090 | `/goapi/mcp/dev/sse` | Dev AI clients |
