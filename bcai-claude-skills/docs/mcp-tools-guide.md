# MCP Tools Guide — รายการ Tools ทั้งหมด

## Tool Categories (41 tools)

### Sales (5 tools)
| Tool | คำอธิบาย | Type |
|------|----------|------|
| `get_daily_sales` | ยอดขายรายวัน | readonly |
| `get_sales_by_date_range` | ยอดขายตามช่วงเวลา | readonly |
| `get_top_selling_products` | สินค้าขายดี | readonly |
| `get_sales_by_seller` | ยอดขายตามพนักงาน | readonly |
| `get_monthly_summary` | สรุปรายเดือน | readonly |

### Dashboard (2 tools)
| Tool | คำอธิบาย | Type |
|------|----------|------|
| `get_dashboard_kpis` | KPI Dashboard | readonly |
| `get_business_health` | สุขภาพธุรกิจ | readonly |

### Financial (4 tools)
| Tool | คำอธิบาย | Type |
|------|----------|------|
| `get_profit_analysis` | วิเคราะห์กำไร | readonly |
| `get_accounts_receivable` | ลูกหนี้การค้า | readonly |
| `get_accounts_payable` | เจ้าหนี้การค้า | readonly |
| `get_cash_flow` | กระแสเงินสด | readonly |

### Inventory (4 tools)
| Tool | คำอธิบาย | Type |
|------|----------|------|
| `get_inventory_value` | มูลค่าสินค้าคงเหลือ | readonly |
| `get_low_stock_alerts` | สินค้าใกล้หมด | readonly |
| `get_dead_stock` | สินค้าค้างสต็อก | readonly |
| `get_inventory_turnover` | อัตราหมุนเวียนสินค้า | readonly |

### Customers (3 tools)
| Tool | คำอธิบาย | Type |
|------|----------|------|
| `get_top_customers` | ลูกค้ารายใหญ่ | readonly |
| `get_customer_growth` | การเติบโตลูกค้า | readonly |
| `get_customer_segments` | กลุ่มลูกค้า | readonly |

### Comparison (2 tools)
| Tool | คำอธิบาย | Type |
|------|----------|------|
| `get_yoy_comparison` | เปรียบเทียบปีต่อปี | readonly |
| `get_mom_comparison` | เปรียบเทียบเดือนต่อเดือน | readonly |

### Products (1 tool)
| Tool | คำอธิบาย | Type |
|------|----------|------|
| `search_products` | ค้นหาสินค้า | readonly |

### Unit of Measure (7 tools)
| Tool | คำอธิบาย | Type |
|------|----------|------|
| `list_units` | รายการหน่วยนับ | readonly |
| `search_units` | ค้นหาหน่วยนับ | readonly |
| `create_unit` | สร้างหน่วยนับ | **write** |
| `create_units` | สร้างหน่วยนับหลายรายการ | **write** |
| `update_unit` | แก้ไขหน่วยนับ | **write** |
| `delete_unit` | ลบหน่วยนับ | **write** |
| `delete_units` | ลบหน่วยนับหลายรายการ | **write** |

### API Development (2 tools)
| Tool | คำอธิบาย | Type |
|------|----------|------|
| `get_api_catalog` | รายการ API endpoints | readonly |
| `get_api_spec` | รายละเอียด API endpoint | readonly |

### Database (2 tools)
| Tool | คำอธิบาย | Type |
|------|----------|------|
| `query_clickhouse` | Query ClickHouse โดยตรง | readonly |
| `query_mongodb` | Query MongoDB โดยตรง | readonly |

### Schema (2 tools)
| Tool | คำอธิบาย | Type |
|------|----------|------|
| `get_model_schema` | ดูโครงสร้าง data model | readonly |
| `get_enum_catalog` | ดูรายการ enums | readonly |

## Permission Presets
| Preset | `allowed_tools` | จำนวน |
|--------|-----------------|-------|
| **Readonly** | `["readonly"]` | 35 tools |
| **Developer** | `["*"]` | 41 tools (all) |
| **Custom** | เลือกเอง | ตามที่เลือก |

## Agent Chatbot Tools (22 tools)
Agent (`POST /api/v1/chatbot/chat-agent`) ใช้เฉพาะ readonly business tools:
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
