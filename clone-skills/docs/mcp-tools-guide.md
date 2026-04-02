# MCP Tools Guide — Complete Tool List

## Tool Categories (41 tools)

### Sales (5 tools)
| Tool | Description | Type |
|------|-------------|------|
| `get_daily_sales` | Daily sales totals | readonly |
| `get_sales_by_date_range` | Sales by date range | readonly |
| `get_top_selling_products` | Top-selling products | readonly |
| `get_sales_by_seller` | Sales by employee | readonly |
| `get_monthly_summary` | Monthly summary | readonly |

### Dashboard (2 tools)
| Tool | Description | Type |
|------|-------------|------|
| `get_dashboard_kpis` | KPI Dashboard | readonly |
| `get_business_health` | Business health metrics | readonly |

### Financial (4 tools)
| Tool | Description | Type |
|------|-------------|------|
| `get_profit_analysis` | Profit analysis | readonly |
| `get_accounts_receivable` | Accounts receivable | readonly |
| `get_accounts_payable` | Accounts payable | readonly |
| `get_cash_flow` | Cash flow | readonly |

### Inventory (4 tools)
| Tool | Description | Type |
|------|-------------|------|
| `get_inventory_value` | Inventory valuation | readonly |
| `get_low_stock_alerts` | Low stock alerts | readonly |
| `get_dead_stock` | Dead stock items | readonly |
| `get_inventory_turnover` | Inventory turnover rate | readonly |

### Customers (3 tools)
| Tool | Description | Type |
|------|-------------|------|
| `get_top_customers` | Top customers | readonly |
| `get_customer_growth` | Customer growth | readonly |
| `get_customer_segments` | Customer segments | readonly |

### Comparison (2 tools)
| Tool | Description | Type |
|------|-------------|------|
| `get_yoy_comparison` | Year-over-year comparison | readonly |
| `get_mom_comparison` | Month-over-month comparison | readonly |

### Products (1 tool)
| Tool | Description | Type |
|------|-------------|------|
| `search_products` | Product search | readonly |

### Unit of Measure (7 tools)
| Tool | Description | Type |
|------|-------------|------|
| `list_units` | List units of measure | readonly |
| `search_units` | Search units | readonly |
| `create_unit` | Create unit | **write** |
| `create_units` | Bulk create units | **write** |
| `update_unit` | Update unit | **write** |
| `delete_unit` | Delete unit | **write** |
| `delete_units` | Bulk delete units | **write** |

### API Development (2 tools)
| Tool | Description | Type |
|------|-------------|------|
| `get_api_catalog` | List API endpoints | readonly |
| `get_api_spec` | API endpoint details | readonly |

### Database (2 tools)
| Tool | Description | Type |
|------|-------------|------|
| `query_clickhouse` | Direct ClickHouse query | readonly |
| `query_mongodb` | Direct MongoDB query | readonly |

### Schema (2 tools)
| Tool | Description | Type |
|------|-------------|------|
| `get_model_schema` | Data model structure | readonly |
| `get_enum_catalog` | Enum catalog | readonly |

## Permission Presets
| Preset | `allowed_tools` | Count |
|--------|-----------------|-------|
| **Readonly** | `["readonly"]` | 35 tools |
| **Developer** | `["*"]` | 41 tools (all) |
| **Custom** | Select specific | As selected |

## Agent Chatbot Tools (22 tools)
Agent (`POST /api/v1/chatbot/chat-agent`) uses readonly business tools only:
```
search_products, get_daily_sales, get_sales_by_date_range,
get_top_selling_products, get_sales_by_seller, get_monthly_summary,
get_dashboard_kpis, get_business_health, get_profit_analysis,
get_accounts_receivable, get_accounts_payable, get_cash_flow,
get_inventory_value, get_low_stock_alerts, get_dead_stock,
get_inventory_turnover, get_top_customers, get_customer_growth,
get_customer_segments, get_yoy_comparison, get_mom_comparison,
list_units
```
