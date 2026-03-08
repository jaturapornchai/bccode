package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mcp/auth"
	"smlcloudplatform/internal/goapi/mcp/mongodb"
	"smlcloudplatform/internal/goapi/mcp/redis"
	"smlcloudplatform/internal/goapi/mcp/tools"

	"github.com/labstack/echo/v4"
)

// MCPServer represents the MCP server
type MCPServer struct {
	keysRepo  *mongodb.KeysRepository
	cache     *redis.Cache
	auth      *auth.AuthMiddleware
	sales     *tools.SalesTool
	ssePrefix string // URL prefix for SSE message endpoint (e.g., "/goapi")
}

// ToolRequest represents a generic tool request
type ToolRequest struct {
	Tool   string                 `json:"tool"`
	Params map[string]interface{} `json:"params"`
}

// ToolResponse represents a generic tool response
type ToolResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	ErrorCode string      `json:"error_code,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Tool      string      `json:"tool"`
}

// AvailableTools lists all available MCP tools
var AvailableTools = []map[string]interface{}{
	// Product Search
	{
		"name":        "search_products",
		"description": "Search products with Thai full-text search and stock balance (same as cart system)",
		"parameters": map[string]interface{}{
			"shop_id":         "string (required) - Shop ID",
			"keyword":         "string (required) - Search keyword (Thai or English)",
			"whcode":          "string (optional) - Warehouse code to filter stock balance",
			"locationcode":    "string (optional) - Location code to filter stock balance",
			"limit":           "number (optional) - Max products to return (default: 50, max: 200)",
			"include_balance": "boolean (optional) - Include stock balance (default: true)",
		},
	},
	// Sales Tools
	{
		"name":        "get_daily_sales",
		"description": "Get daily sales summary for a specific date",
		"parameters": map[string]interface{}{
			"shop_id":     "string (required) - Shop ID",
			"date":        "string (required) - Date in YYYY-MM-DD format",
			"branch_code": "string (optional) - Filter by branch/warehouse code",
		},
	},
	{
		"name":        "get_sales_by_date_range",
		"description": "Get sales data grouped by day/week/month for a date range",
		"parameters": map[string]interface{}{
			"shop_id":     "string (required) - Shop ID",
			"from_date":   "string (required) - Start date in YYYY-MM-DD format",
			"to_date":     "string (required) - End date in YYYY-MM-DD format",
			"branch_code": "string (optional) - Filter by branch/warehouse code",
			"group_by":    "string (optional) - Group by: day, week, month (default: day)",
		},
	},
	{
		"name":        "get_top_selling_products",
		"description": "Get top selling products for a date range",
		"parameters": map[string]interface{}{
			"shop_id":     "string (required) - Shop ID",
			"from_date":   "string (required) - Start date in YYYY-MM-DD format",
			"to_date":     "string (required) - End date in YYYY-MM-DD format",
			"limit":       "number (optional) - Number of products to return (default: 10, max: 100)",
			"branch_code": "string (optional) - Filter by branch/warehouse code",
		},
	},
	{
		"name":        "get_sales_by_seller",
		"description": "Get sales data grouped by seller/salesperson",
		"parameters": map[string]interface{}{
			"shop_id":   "string (required) - Shop ID",
			"from_date": "string (required) - Start date in YYYY-MM-DD format",
			"to_date":   "string (required) - End date in YYYY-MM-DD format",
		},
	},
	{
		"name":        "get_monthly_summary",
		"description": "Get monthly sales summary with comparison to previous month",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"year":    "number (required) - Year (e.g., 2025)",
			"month":   "number (required) - Month (1-12)",
		},
	},
	// Dashboard Tools
	{
		"name":        "get_dashboard_kpis",
		"description": "Get comprehensive KPI dashboard for CEO/executives. Includes sales, orders, profit, customers, inventory metrics with trends.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"period":  "string (optional) - Period: today, this_week, this_month, this_year (default: this_month)",
		},
	},
	{
		"name":        "get_business_health",
		"description": "Get overall business health score and key indicators. Returns health score (0-100), alerts, and recommendations.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
		},
	},
	// Financial Tools
	{
		"name":        "get_profit_analysis",
		"description": "Get detailed profit analysis with revenue breakdown by category, gross margin, and profit trends.",
		"parameters": map[string]interface{}{
			"shop_id":   "string (required) - Shop ID",
			"from_date": "string (required) - Start date in YYYY-MM-DD format",
			"to_date":   "string (required) - End date in YYYY-MM-DD format",
		},
	},
	{
		"name":        "get_accounts_receivable",
		"description": "Get accounts receivable summary with aging buckets (current, 30, 60, 90+ days) and top debtors.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"limit":   "number (optional) - Number of top debtors to return (default: 10)",
		},
	},
	{
		"name":        "get_accounts_payable",
		"description": "Get accounts payable summary with aging buckets and top creditors.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"limit":   "number (optional) - Number of top creditors to return (default: 10)",
		},
	},
	{
		"name":        "get_cash_flow",
		"description": "Get cash flow analysis showing inflows, outflows, and net position over time.",
		"parameters": map[string]interface{}{
			"shop_id":   "string (required) - Shop ID",
			"from_date": "string (required) - Start date in YYYY-MM-DD format",
			"to_date":   "string (required) - End date in YYYY-MM-DD format",
		},
	},
	// Inventory Tools
	{
		"name":        "get_inventory_value",
		"description": "Get total inventory valuation with breakdown by category and warehouse.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"whcode":  "string (optional) - Filter by warehouse code",
		},
	},
	{
		"name":        "get_low_stock_alerts",
		"description": "Get products that are below minimum stock level or out of stock.",
		"parameters": map[string]interface{}{
			"shop_id":   "string (required) - Shop ID",
			"threshold": "number (optional) - Stock threshold to consider low (default: 10)",
			"limit":     "number (optional) - Number of alerts to return (default: 50)",
		},
	},
	{
		"name":        "get_dead_stock",
		"description": "Get products with no movement for specified days (slow-moving/dead stock).",
		"parameters": map[string]interface{}{
			"shop_id":          "string (required) - Shop ID",
			"days_no_movement": "number (optional) - Days without movement (default: 90)",
			"limit":            "number (optional) - Number of products to return (default: 50)",
		},
	},
	{
		"name":        "get_inventory_turnover",
		"description": "Get inventory turnover ratio and days of inventory for products.",
		"parameters": map[string]interface{}{
			"shop_id":   "string (required) - Shop ID",
			"from_date": "string (required) - Start date in YYYY-MM-DD format",
			"to_date":   "string (required) - End date in YYYY-MM-DD format",
			"limit":     "number (optional) - Number of products to return (default: 50)",
		},
	},
	// Customer Tools
	{
		"name":        "get_top_customers",
		"description": "Get top customers by revenue with purchase history and trends.",
		"parameters": map[string]interface{}{
			"shop_id":   "string (required) - Shop ID",
			"from_date": "string (required) - Start date in YYYY-MM-DD format",
			"to_date":   "string (required) - End date in YYYY-MM-DD format",
			"limit":     "number (optional) - Number of customers to return (default: 10)",
		},
	},
	{
		"name":        "get_customer_growth",
		"description": "Get customer acquisition and retention metrics over time.",
		"parameters": map[string]interface{}{
			"shop_id":   "string (required) - Shop ID",
			"from_date": "string (required) - Start date in YYYY-MM-DD format",
			"to_date":   "string (required) - End date in YYYY-MM-DD format",
		},
	},
	{
		"name":        "get_customer_segments",
		"description": "Get customer segmentation analysis (RFM: Recency, Frequency, Monetary).",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
		},
	},
	// Comparison Tools
	{
		"name":        "get_yoy_comparison",
		"description": "Get year-over-year comparison of revenue, orders, and profit with monthly breakdown.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"year":    "number (optional) - Year to compare (default: current year)",
			"month":   "number (optional) - Specific month to compare (1-12, optional)",
		},
	},
	{
		"name":        "get_mom_comparison",
		"description": "Get month-over-month comparison with weekly breakdown and daily trends.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"year":    "number (optional) - Year (default: current year)",
			"month":   "number (optional) - Month to compare (1-12, default: current month)",
		},
	},
	// Database Tools
	{
		"name":        "get_database_schema",
		"description": "Get PostgreSQL database structure (tables, columns, types, relationships). PostgreSQL is used for relational processing — joins, aggregations, reports. For raw data, use MongoDB. For OLAP analytics, use ClickHouse.",
		"parameters": map[string]interface{}{
			"shop_id":    "string (required) - Shop ID",
			"table_name": "string (optional) - Filter by table name (partial match)",
		},
	},
	{
		"name":        "execute_query",
		"description": "Execute a readonly SQL query on PostgreSQL (SELECT only). PostgreSQL handles relational processing — joins, aggregations, reports.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"query":   "string (required) - SQL SELECT query",
			"limit":   "number (optional) - Max rows to return (default: 100, max: 1000)",
		},
	},
	{
		"name":        "get_table_sample",
		"description": "Get sample data from a PostgreSQL table. Quick way to see what relational data looks like.",
		"parameters": map[string]interface{}{
			"shop_id":    "string (required) - Shop ID",
			"table_name": "string (required) - Table name",
			"limit":      "number (optional) - Number of rows (default: 10, max: 100)",
		},
	},
	// MongoDB Tools — Main data store (source of truth)
	{
		"name":        "query_mongodb",
		"description": "Query MongoDB collection (readonly). MongoDB is the main data store (source of truth) — all documents, transactions, and master data live here.",
		"parameters": map[string]interface{}{
			"shop_id":    "string (optional) - Shop ID for data isolation",
			"database":   "string (optional) - Database name (default: from config)",
			"collection": "string (required) - Collection name",
			"filter":     "string (optional) - JSON filter e.g. {\"transflag\":6} (default: {})",
			"limit":      "number (optional) - Max documents (default: 20, max: 100)",
		},
	},
	{
		"name":        "list_mongodb_collections",
		"description": "List all collections in MongoDB (main data store). Use to discover available raw data.",
		"parameters": map[string]interface{}{
			"database": "string (optional) - Database name (default: from config)",
		},
	},
	{
		"name":        "aggregate_mongodb",
		"description": "Run aggregation pipeline on MongoDB (main data store, readonly). Blocks $out and $merge stages. Use for complex queries on raw source data.",
		"parameters": map[string]interface{}{
			"shop_id":    "string (optional) - Shop ID for data isolation",
			"database":   "string (optional) - Database name (default: from config)",
			"collection": "string (required) - Collection name",
			"pipeline":   "string (required) - JSON array of pipeline stages e.g. [{\"$match\":{\"transflag\":6}},{\"$group\":{\"_id\":\"$currency\",\"count\":{\"$sum\":1}}}]",
			"limit":      "number (optional) - Max results (default: 100, max: 100)",
		},
	},
	// ClickHouse Tools — Dimensional/OLAP processing
	{
		"name":        "query_clickhouse",
		"description": "Execute readonly SELECT/SHOW query on ClickHouse (OLAP/dimensional analytics). Use for time-series analysis, BI dashboards, and large-scale aggregations.",
		"parameters": map[string]interface{}{
			"shop_id":  "string (optional) - Shop ID for data isolation",
			"database": "string (optional) - Database name (default: from env CH_DATABASE_NAME)",
			"query":    "string (required) - SQL SELECT or SHOW query",
			"limit":    "number (optional) - Max rows (default: 100, max: 1000)",
		},
	},
	{
		"name":        "list_clickhouse_tables",
		"description": "List all tables in ClickHouse (OLAP/dimensional) with engine, row count and size info. Use to discover analytics data.",
		"parameters": map[string]interface{}{
			"database": "string (optional) - Database name (default: from env CH_DATABASE_NAME)",
		},
	},
	// Dev Database Tools — ไม่จำกัด readonly (ใช้ด้วยความระมัดระวัง)
	{
		"name":        "execute_pg_command",
		"description": "⚡ DEV TOOL: Execute ANY SQL on PostgreSQL (SELECT, DELETE, INSERT, UPDATE, ALTER, TRUNCATE, DROP). No readonly restriction. Use with caution.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID (= PostgreSQL database name)",
			"query":   "string (required) - Any SQL command",
			"limit":   "number (optional) - Max rows for SELECT (default: 100, max: 10000)",
		},
	},
	{
		"name":        "execute_ch_command",
		"description": "⚡ DEV TOOL: Execute ANY SQL on ClickHouse (SELECT, ALTER TABLE DELETE, INSERT, DROP, TRUNCATE). No readonly restriction. Use with caution.",
		"parameters": map[string]interface{}{
			"database": "string (optional) - Database name (default: from env CH_DATABASE_NAME)",
			"query":    "string (required) - Any SQL command",
			"limit":    "number (optional) - Max rows for SELECT (default: 100, max: 10000)",
		},
	},
	// API Catalog Tool
	{
		"name":        "list_api_endpoints",
		"description": "List all available API endpoints with full details (method, path, parameters, request/response schema). Use to discover APIs for frontend development. Returns GoAPI routes and MainAPI routes (from Swagger).",
		"parameters": map[string]interface{}{
			"category": "string (optional) - Filter by category (e.g., product, sales, transaction, approval, restaurant, lineoa)",
			"keyword":  "string (optional) - Search in path or description",
			"method":   "string (optional) - Filter by HTTP method (GET, POST, PUT, DELETE)",
			"source":   "string (optional) - Filter by source: goapi, mainapi (default: all)",
			"limit":    "number (optional) - Max endpoints to return (default: 50, max: 500)",
		},
	},
	// API Spec Tool
	{
		"name":        "get_api_spec",
		"description": "Get detailed request/response specification for a specific API endpoint. Returns full spec with parameters, request body schema, response schema, curl example, and usage notes.",
		"parameters": map[string]interface{}{
			"path":   "string (required) - API path or partial match (e.g., '/product/search', '/login')",
			"method": "string (optional) - HTTP method filter (GET, POST, PUT, DELETE)",
		},
	},
	// API Example Tool
	{
		"name":        "get_api_example",
		"description": "Get real request/response examples for a specific API endpoint. Returns curl command, request body, response body, Dart code snippet, and TypeScript code snippet.",
		"parameters": map[string]interface{}{
			"path":   "string (required) - API path or partial match (e.g., '/product/search', '/transaction/calculate')",
			"method": "string (optional) - HTTP method filter",
		},
	},
	// Enum Catalog Tool
	{
		"name":        "list_enums",
		"description": "List all enum values and constants used in the backend (transflag, payment types, approval status, VAT types, etc.). Essential for frontend to use correct values.",
		"parameters": map[string]interface{}{
			"category": "string (optional) - Filter by category (transaction, payment, approval, datahistory, kafka, etc.)",
			"keyword":  "string (optional) - Search in enum name, label, or description",
		},
	},
	// Unit of Measure Tools (หน่วยนับ)
	{
		"name":        "list_units",
		"description": "List/search units of measure (หน่วยนับ). Returns unit codes and names.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"keyword": "string (optional) - Search by unit code or name",
			"limit":   "number (optional) - Max results (default: 50, max: 200)",
		},
	},
	{
		"name":        "create_unit",
		"description": "Create a new unit of measure (หน่วยนับ). Uses names[] for multi-language display names.",
		"parameters": map[string]interface{}{
			"shop_id":  "string (required) - Shop ID",
			"unitcode": "string (required) - Unit code e.g. EA, BOX, KG",
			"names":    "string (required) - JSON array [{\"code\":\"th\",\"name\":\"ชิ้น\"},{\"code\":\"en\",\"name\":\"Each\"}]",
		},
	},
	{
		"name":        "create_units",
		"description": "Create multiple units of measure at once (bulk). Skips duplicates.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"units":   "string (required) - JSON array e.g. [{\"unitcode\":\"EA\",\"names\":[{\"code\":\"th\",\"name\":\"ชิ้น\"}]},{\"unitcode\":\"BOX\",\"names\":[{\"code\":\"th\",\"name\":\"กล่อง\"}]}]",
		},
	},
	{
		"name":        "update_unit",
		"description": "Update an existing unit of measure by unit code.",
		"parameters": map[string]interface{}{
			"shop_id":  "string (required) - Shop ID",
			"unitcode": "string (required) - Unit code to update",
			"names":    "string (optional) - JSON array of language names [{\"code\":\"th\",\"name\":\"ชิ้น\"}]",
		},
	},
	{
		"name":        "delete_unit",
		"description": "Delete a unit of measure by unit code.",
		"parameters": map[string]interface{}{
			"shop_id":  "string (required) - Shop ID",
			"unitcode": "string (required) - Unit code to delete",
		},
	},
	{
		"name":        "delete_units",
		"description": "Delete multiple units of measure at once (bulk). Reports which were deleted and which were not found.",
		"parameters": map[string]interface{}{
			"shop_id":   "string (required) - Shop ID",
			"unitcodes": "string (required) - JSON array of unit codes e.g. [\"EA\",\"BOX\",\"KG\"]. Max 100 items.",
		},
	},
	{
		"name":        "get_unit_schema",
		"description": "Get the data structure/schema of unit of measure (หน่วยนับ) documents.",
		"parameters":  map[string]interface{}{},
	},
	// Product Barcode Tools (สินค้า/บาร์โค้ด)
	{
		"name":        "list_barcodes",
		"description": "List/search product barcodes (สินค้า/บาร์โค้ด). Search by barcode, item code, or product name.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"keyword": "string (optional) - Search by barcode, item code, or product name",
			"limit":   "number (optional) - Max results (default: 50, max: 200)",
		},
	},
	{
		"name":        "create_barcode",
		"description": "Create a new product barcode (สินค้า/บาร์โค้ด). Uses names[] for multi-language product names.",
		"parameters": map[string]interface{}{
			"shop_id":        "string (required) - Shop ID",
			"barcode":        "string (required) - Barcode e.g. 8859100001234",
			"itemcode":       "string (required) - Item/product code e.g. SKU001",
			"names":          "string (required) - JSON array [{\"code\":\"th\",\"name\":\"สินค้า A\"},{\"code\":\"en\",\"name\":\"Product A\"}]",
			"itemunitcode":   "string (optional) - Unit code e.g. EA, BOX",
			"itemunitnames":  "string (optional) - JSON array of unit names [{\"code\":\"th\",\"name\":\"ชิ้น\"}]",
			"prices":         "string (optional) - JSON array of prices [{\"keynumber\":1,\"price\":100.00}]",
			"standvalue":     "number (optional) - Unit conversion numerator (default: 1). e.g. BOX=24 means 1 BOX = 24 base units",
			"dividevalue":    "number (optional) - Unit conversion denominator (default: 1)",
			"ismainbarcode":  "boolean (optional) - Is main barcode? Auto-detected if not provided: true when standvalue=1 & dividevalue=1",
			"groupcode":      "string (optional) - Product group code (กลุ่มสินค้า)",
			"groupnames":     "string (optional) - JSON array of group names [{\"code\":\"th\",\"name\":\"อาหาร\"}]",
			"categorycode":   "string (optional) - Product category guidfixed (หมวดสินค้า)",
			"categorynames":  "string (optional) - JSON array of category names [{\"code\":\"th\",\"name\":\"เนื้อสัตว์\"}]",
		},
	},
	{
		"name":        "create_barcodes",
		"description": "Create multiple product barcodes at once (bulk). Skips duplicates.",
		"parameters": map[string]interface{}{
			"shop_id":  "string (required) - Shop ID",
			"barcodes": "string (required) - JSON array e.g. [{\"barcode\":\"123\",\"itemcode\":\"SKU1\",\"names\":[{\"code\":\"th\",\"name\":\"สินค้า\"}]}]",
		},
	},
	{
		"name":        "update_barcode",
		"description": "Update an existing product barcode by guidfixed.",
		"parameters": map[string]interface{}{
			"shop_id":   "string (required) - Shop ID",
			"guidfixed": "string (required) - GuidFixed of the barcode to update",
			"names":     "string (optional) - JSON array of language names [{\"code\":\"th\",\"name\":\"ชื่อใหม่\"}]",
			"itemunitcode":   "string (optional) - New unit code",
			"itemunitnames":  "string (optional) - JSON array of unit names",
			"prices":         "string (optional) - JSON array of prices [{\"keynumber\":1,\"price\":150.00}]",
			"groupcode":      "string (optional) - Product group code (กลุ่มสินค้า)",
			"groupnames":     "string (optional) - JSON array of group names [{\"code\":\"th\",\"name\":\"อาหาร\"}]",
			"categorycode":   "string (optional) - Product category guidfixed (หมวดสินค้า)",
			"categorynames":  "string (optional) - JSON array of category names [{\"code\":\"th\",\"name\":\"เนื้อสัตว์\"}]",
		},
	},
	{
		"name":        "delete_barcode",
		"description": "Delete a product barcode by guidfixed.",
		"parameters": map[string]interface{}{
			"shop_id":   "string (required) - Shop ID",
			"guidfixed": "string (required) - GuidFixed of the barcode to delete",
		},
	},
	{
		"name":        "delete_barcodes",
		"description": "Delete multiple product barcodes at once (bulk). Reports which were deleted and which were not found.",
		"parameters": map[string]interface{}{
			"shop_id":    "string (required) - Shop ID",
			"guidfixeds": "string (required) - JSON array of guidfixed values e.g. [\"guid1\",\"guid2\"]. Max 100 items.",
		},
	},
	{
		"name":        "get_barcode_schema",
		"description": "Get the data structure/schema of product barcode (สินค้า/บาร์โค้ด) documents.",
		"parameters":  map[string]interface{}{},
	},
	// Reference Barcodes + Multi-Unit
	{
		"name":        "get_ref_barcodes",
		"description": "Get reference barcodes and unit chain for a product. Shows all units (e.g., ชิ้น→ลัง) and how they reference each other. Returns product_name and unit_name as formatted strings (auto from names[]).",
		"parameters": map[string]interface{}{
			"shop_id":  "string (required) - Shop ID",
			"itemcode": "string (optional) - Item code to get all barcodes for (if not specified, use barcode to find)",
			"barcode":  "string (optional) - Barcode to find item code from (one of itemcode/barcode required)",
		},
	},
	{
		"name":        "set_ref_barcode",
		"description": "Set reference barcode for a barcode (unit conversion). E.g., BOX → EA means 1 BOX = 24 EA. Checks for circular references.",
		"parameters": map[string]interface{}{
			"shop_id":      "string (required) - Shop ID",
			"barcode":      "string (required) - Source barcode (e.g., BOX barcode)",
			"ref_barcode":  "string (required) - Target reference barcode (e.g., EA barcode). Must be same itemcode.",
			"qty":          "number (optional) - Quantity for condition-based reference",
			"standvalue":   "number (optional) - Override stand value (default: use existing)",
			"dividevalue":  "number (optional) - Override divide value (default: use existing)",
			"condition":    "boolean (optional) - Is conditional reference (default: false)",
		},
	},
	{
		"name":        "create_multi_unit_barcode",
		"description": "Create a product with multiple units at once (e.g., ชิ้น + ลัง + แพ็ค). Automatically sets reference barcodes. The unit with standvalue=1 becomes the base unit.",
		"parameters": map[string]interface{}{
			"shop_id":  "string (required) - Shop ID",
			"itemcode": "string (required) - Item code for all units",
			"names":    "string (optional) - Default JSON names array (used when unit doesn't have its own names)",
			"units":    "string (required) - JSON array of units e.g. [{\"barcode\":\"EA-001\",\"itemunitcode\":\"EA\",\"standvalue\":1,\"dividevalue\":1,...},{\"barcode\":\"BOX-001\",\"itemunitcode\":\"BOX\",\"standvalue\":24,\"dividevalue\":1,...}]",
		},
	},
	// Rebuild Products (Full Sync — กรณี Kafka sync ผิดพลาด)
	{
		"name":        "rebuild_products",
		"description": "Full rebuild: sync ALL product barcodes from MongoDB → PostgreSQL + ClickHouse. Use when Kafka sync fails or data is out of sync. Same as frontend 'สร้างสินค้าใหม่' button.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID to rebuild products for",
		},
	},
	// Product Group Tools (กลุ่มสินค้า)
	{
		"name":        "list_product_groups",
		"description": "List/search product groups (กลุ่มสินค้า). Returns group codes and names with multi-language support.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"keyword": "string (optional) - Search by group code or name (e.g., 'อาหาร', 'FOOD')",
			"limit":   "number (optional) - Max results to return (default: 50, max: 200)",
		},
	},
	{
		"name":        "create_product_group",
		"description": "Create a new product group (กลุ่มสินค้า). Uses names[] for multi-language display names.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"code":    "string (required) - Group code (e.g., 'FOOD', 'DRINK', 'TOOL')",
			"names":   "string (required) - JSON array [{\"code\":\"th\",\"name\":\"อาหาร\"},{\"code\":\"en\",\"name\":\"Food\"}]",
		},
	},
	{
		"name":        "create_product_groups",
		"description": "Create multiple product groups at once (bulk). Skips duplicates automatically.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"groups":  "string (required) - JSON array e.g. [{\"code\":\"FOOD\",\"names\":[{\"code\":\"th\",\"name\":\"อาหาร\"}]},{\"code\":\"DRINK\",\"names\":[{\"code\":\"th\",\"name\":\"เครื่องดื่ม\"}]}]. Max 100 items.",
		},
	},
	{
		"name":        "update_product_group",
		"description": "Update an existing product group by code.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"code":    "string (required) - Group code to update",
			"names":   "string (optional) - JSON array of multi-language names [{\"code\":\"th\",\"name\":\"อาหาร\"}]",
		},
	},
	{
		"name":        "delete_product_group",
		"description": "Delete a product group by code.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"code":    "string (required) - Group code to delete",
		},
	},
	{
		"name":        "delete_product_groups",
		"description": "Delete multiple product groups at once (bulk). Reports which were deleted and which were not found.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"codes":   "string (required) - JSON array of group codes e.g. [\"FOOD\",\"DRINK\",\"TOOL\"]. Max 100 items.",
		},
	},
	{
		"name":        "get_product_group_schema",
		"description": "Get the data structure/schema of product group (กลุ่มสินค้า) documents with examples.",
		"parameters":  map[string]interface{}{},
	},
	// Product Category Tools (หมวดสินค้า)
	{
		"name":        "list_product_categories",
		"description": "List/search product categories (หมวดสินค้า). Returns category names, hierarchy, and group numbers.",
		"parameters": map[string]interface{}{
			"shop_id": "string (required) - Shop ID",
			"keyword": "string (optional) - Search by category name (e.g., 'เนื้อสัตว์', 'ผัก')",
			"limit":   "number (optional) - Max results to return (default: 50, max: 200)",
		},
	},
	{
		"name":        "create_product_category",
		"description": "Create a new product category (หมวดสินค้า). Uses names[] for multi-language display names. Supports hierarchy via parentguid.",
		"parameters": map[string]interface{}{
			"shop_id":      "string (required) - Shop ID",
			"names":        "string (required) - JSON array [{\"code\":\"th\",\"name\":\"เนื้อสัตว์\"},{\"code\":\"en\",\"name\":\"Meat\"}]",
			"parentguid":   "string (optional) - Parent category guidfixed (for hierarchy)",
			"groupnumber":  "number (optional) - Group/sort number",
		},
	},
	{
		"name":        "create_product_categories",
		"description": "Create multiple product categories at once (bulk). Auto-generates guidfixed for each.",
		"parameters": map[string]interface{}{
			"shop_id":    "string (required) - Shop ID",
			"categories": "string (required) - JSON array e.g. [{\"names\":[{\"code\":\"th\",\"name\":\"เนื้อสัตว์\"}],\"groupnumber\":1},{\"names\":[{\"code\":\"th\",\"name\":\"ผัก\"}],\"groupnumber\":2}]. Max 100 items.",
		},
	},
	{
		"name":        "update_product_category",
		"description": "Update an existing product category by guidfixed.",
		"parameters": map[string]interface{}{
			"shop_id":     "string (required) - Shop ID",
			"guidfixed":   "string (required) - GuidFixed of the category to update",
			"names":       "string (optional) - JSON array of multi-language names [{\"code\":\"th\",\"name\":\"เนื้อสัตว์\"}]",
			"parentguid":  "string (optional) - New parent category guidfixed",
			"groupnumber": "number (optional) - New group/sort number",
		},
	},
	{
		"name":        "delete_product_category",
		"description": "Delete a product category by guidfixed.",
		"parameters": map[string]interface{}{
			"shop_id":   "string (required) - Shop ID",
			"guidfixed": "string (required) - GuidFixed of the category to delete",
		},
	},
	{
		"name":        "delete_product_categories",
		"description": "Delete multiple product categories at once (bulk). Reports which were deleted and which were not found.",
		"parameters": map[string]interface{}{
			"shop_id":    "string (required) - Shop ID",
			"guidfixeds": "string (required) - JSON array of guidfixed values e.g. [\"guid1\",\"guid2\"]. Max 100 items.",
		},
	},
	{
		"name":        "get_product_category_schema",
		"description": "Get the data structure/schema of product category (หมวดสินค้า) documents with examples.",
		"parameters":  map[string]interface{}{},
	},
	// Creditor Tools (เจ้าหนี้)
	{
		"name":        "list_creditors",
		"description": "List/search creditors (เจ้าหนี้). Returns creditor codes, names, tax ID, and contact info.",
		"parameters": map[string]interface{}{
			"keyword": "string (optional) — Search by creditor code, name, or tax ID",
			"limit":   "number (optional) — Max results (default: 50, max: 200)",
		},
	},
	{
		"name":        "create_creditor",
		"description": "Create a new creditor (เจ้าหนี้). Requires code and names.",
		"parameters": map[string]interface{}{
			"code":              "string (required) — Creditor code e.g. 'CR-001'",
			"names":             "string (required) — JSON array of names e.g. [{\"code\":\"th\",\"name\":\"บริษัท ABC\"}]",
			"personaltype":      "number (optional) — 1=บุคคลธรรมดา, 2=นิติบุคคล (default: 0)",
			"taxid":             "string (optional) — Tax ID",
			"email":             "string (optional) — Email",
			"creditday":         "number (optional) — Credit days",
			"addressforbilling": "string (optional) — JSON object {address, countrycode, provincecode, districtcode, subdistrictcode, zipcode, phoneprimary}",
		},
	},
	{
		"name":        "create_creditors",
		"description": "Create multiple creditors at once (bulk). Skips duplicates automatically. Max 100 items.",
		"parameters": map[string]interface{}{
			"creditors": "string (required) — JSON array of creditors. Each needs code (required) and names (required).",
		},
	},
	{
		"name":        "update_creditor",
		"description": "Update an existing creditor by guidfixed.",
		"parameters": map[string]interface{}{
			"guidfixed":         "string (required) — GuidFixed of the creditor to update",
			"names":             "string (optional) — JSON array of names",
			"taxid":             "string (optional) — New tax ID",
			"email":             "string (optional) — New email",
			"creditday":         "number (optional) — New credit days",
			"addressforbilling": "string (optional) — JSON object for billing address",
		},
	},
	{
		"name":        "delete_creditor",
		"description": "Delete a creditor by guidfixed (soft delete).",
		"parameters": map[string]interface{}{
			"guidfixed": "string (required) — GuidFixed of the creditor to delete",
		},
	},
	{
		"name":        "delete_creditors",
		"description": "Delete multiple creditors at once (bulk soft delete). Max 100 items.",
		"parameters": map[string]interface{}{
			"guidfixeds": "string (required) — JSON array of guidfixed values to delete",
		},
	},
	{
		"name":        "get_creditor_schema",
		"description": "Get the data structure/schema of creditor (เจ้าหนี้) documents with examples.",
		"parameters":  map[string]interface{}{},
	},
	// Debtor Tools (ลูกหนี้)
	{
		"name":        "list_debtors",
		"description": "List/search debtors (ลูกหนี้). Returns debtor codes, names, tax ID, and contact info.",
		"parameters": map[string]interface{}{
			"keyword": "string (optional) — Search by debtor code, name, or tax ID",
			"limit":   "number (optional) — Max results (default: 50, max: 200)",
		},
	},
	{
		"name":        "create_debtor",
		"description": "Create a new debtor (ลูกหนี้). Requires code and names.",
		"parameters": map[string]interface{}{
			"code":              "string (required) — Debtor code e.g. 'DB-001'",
			"names":             "string (required) — JSON array of names e.g. [{\"code\":\"th\",\"name\":\"ร้าน XYZ\"}]",
			"personaltype":      "number (optional) — 1=บุคคลธรรมดา, 2=นิติบุคคล (default: 0)",
			"taxid":             "string (optional) — Tax ID",
			"email":             "string (optional) — Email",
			"creditday":         "number (optional) — Credit days",
			"addressforbilling": "string (optional) — JSON object {address, countrycode, provincecode, districtcode, subdistrictcode, zipcode, phoneprimary}",
		},
	},
	{
		"name":        "create_debtors",
		"description": "Create multiple debtors at once (bulk). Skips duplicates automatically. Max 100 items.",
		"parameters": map[string]interface{}{
			"debtors": "string (required) — JSON array of debtors. Each needs code (required) and names (required).",
		},
	},
	{
		"name":        "update_debtor",
		"description": "Update an existing debtor by guidfixed.",
		"parameters": map[string]interface{}{
			"guidfixed":         "string (required) — GuidFixed of the debtor to update",
			"names":             "string (optional) — JSON array of names",
			"taxid":             "string (optional) — New tax ID",
			"email":             "string (optional) — New email",
			"creditday":         "number (optional) — New credit days",
			"addressforbilling": "string (optional) — JSON object for billing address",
		},
	},
	{
		"name":        "delete_debtor",
		"description": "Delete a debtor by guidfixed (soft delete).",
		"parameters": map[string]interface{}{
			"guidfixed": "string (required) — GuidFixed of the debtor to delete",
		},
	},
	{
		"name":        "delete_debtors",
		"description": "Delete multiple debtors at once (bulk soft delete). Max 100 items.",
		"parameters": map[string]interface{}{
			"guidfixeds": "string (required) — JSON array of guidfixed values to delete",
		},
	},
	{
		"name":        "get_debtor_schema",
		"description": "Get the data structure/schema of debtor (ลูกหนี้) documents with examples.",
		"parameters":  map[string]interface{}{},
	},
	// Model Schema Tool
	{
		"name":        "get_model_schema",
		"description": "Get Go struct definitions as JSON schema with auto-generated Dart class and TypeScript interface. Use to create frontend models that match backend exactly.",
		"parameters": map[string]interface{}{
			"model":    "string (optional) - Model name or partial match (e.g., 'MongoDocModel', 'Customer', 'Barcode')",
			"category": "string (optional) - Filter by category (transaction, product, master, stock, file, pdf, query)",
			"keyword":  "string (optional) - Search in model name or description",
		},
	},
}

// NewMCPServer creates a new MCP server
// defaultServer — singleton สำหรับ agent loop เรียก ExecuteToolDirect
var defaultServer *MCPServer

func NewMCPServer() *MCPServer {
	keysRepo := mongodb.NewKeysRepository()
	cache := redis.NewCache()
	authMiddleware := auth.NewAuthMiddleware(keysRepo, cache)
	salesTool := tools.NewSalesTool()

	s := &MCPServer{
		keysRepo: keysRepo,
		cache:    cache,
		auth:     authMiddleware,
		sales:    salesTool,
	}
	defaultServer = s
	return s
}

// GetDefaultServer returns the singleton MCP server instance
func GetDefaultServer() *MCPServer {
	return defaultServer
}

// ExecuteToolDirect executes a tool by name with params (for agent loop)
// ไม่ check permission — agent กำหนด tool whitelist เอง
func (s *MCPServer) ExecuteToolDirect(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error) {
	if params == nil {
		params = make(map[string]interface{})
	}

	switch toolName {
	case "search_products":
		return s.invokeSearchProducts(ctx, params)
	case "get_daily_sales":
		return s.invokeGetDailySales(ctx, params)
	case "get_sales_by_date_range":
		return s.invokeGetSalesByDateRange(ctx, params)
	case "get_top_selling_products":
		return s.invokeGetTopSellingProducts(ctx, params)
	case "get_sales_by_seller":
		return s.invokeGetSalesBySeller(ctx, params)
	case "get_monthly_summary":
		return s.invokeGetMonthlySummary(ctx, params)
	case "get_dashboard_kpis":
		return s.invokeGetDashboardKPIs(ctx, params)
	case "get_business_health":
		return s.invokeGetBusinessHealth(ctx, params)
	case "get_profit_analysis":
		return s.invokeGetProfitAnalysis(ctx, params)
	case "get_accounts_receivable":
		return s.invokeGetAccountsReceivable(ctx, params)
	case "get_accounts_payable":
		return s.invokeGetAccountsPayable(ctx, params)
	case "get_cash_flow":
		return s.invokeGetCashFlow(ctx, params)
	case "get_inventory_value":
		return s.invokeGetInventoryValue(ctx, params)
	case "get_low_stock_alerts":
		return s.invokeGetLowStockAlerts(ctx, params)
	case "get_dead_stock":
		return s.invokeGetDeadStock(ctx, params)
	case "get_inventory_turnover":
		return s.invokeGetInventoryTurnover(ctx, params)
	case "get_top_customers":
		return s.invokeGetTopCustomers(ctx, params)
	case "get_customer_growth":
		return s.invokeGetCustomerGrowth(ctx, params)
	case "get_customer_segments":
		return s.invokeGetCustomerSegments(ctx, params)
	case "get_yoy_comparison":
		return s.invokeGetYoYComparison(ctx, params)
	case "get_mom_comparison":
		return s.invokeGetMoMComparison(ctx, params)
	case "list_units":
		return s.invokeListUnits(ctx, params)
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// RegisterRoutes registers MCP routes with Echo
func (s *MCPServer) RegisterRoutes(e *echo.Echo) {
	// MCP API Group with authentication
	mcpGroup := e.Group("/mcp")
	mcpGroup.Use(s.auth.APIKeyAuth)

	// Tool execution endpoint
	mcpGroup.POST("/invoke", s.InvokeTool)

	// Specific tool endpoints
	mcpGroup.POST("/tools/daily-sales", s.GetDailySales)
	mcpGroup.POST("/tools/sales-by-date-range", s.GetSalesByDateRange)
	mcpGroup.POST("/tools/top-selling-products", s.GetTopSellingProducts)
	mcpGroup.POST("/tools/sales-by-seller", s.GetSalesBySeller)
	mcpGroup.POST("/tools/monthly-summary", s.GetMonthlySummary)

	// Public endpoints (no auth required)
	e.GET("/mcp/tools", s.ListTools)
	e.GET("/mcp/health", s.HealthCheck)
}

// ListTools returns available tools (general only — ซ่อน dev tools)
func (s *MCPServer) ListTools(c echo.Context) error {
	var general []map[string]interface{}
	for _, t := range AvailableTools {
		name, _ := t["name"].(string)
		if devToolNames[name] {
			continue
		}
		general = append(general, t)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"tools":   general,
		"count":   len(general),
		"mode":    "general",
		"version": "1.0.0",
	})
}

// ListToolsDev returns all available tools (general + dev)
func (s *MCPServer) ListToolsDev(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"tools":   AvailableTools,
		"count":   len(AvailableTools),
		"mode":    "dev",
		"version": "1.0.0",
	})
}

// HealthCheck returns server health status
func (s *MCPServer) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"service":   "mcp-server",
	})
}

// InvokeTool invokes a tool by name
func (s *MCPServer) InvokeTool(c echo.Context) error {
	var req ToolRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ToolResponse{
			Success:   false,
			Error:     "Invalid request body",
			ErrorCode: "INVALID_REQUEST",
			Timestamp: time.Now(),
			Tool:      req.Tool,
		})
	}

	if req.Tool == "" {
		return c.JSON(http.StatusBadRequest, ToolResponse{
			Success:   false,
			Error:     "Tool name is required",
			ErrorCode: "MISSING_TOOL",
			Timestamp: time.Now(),
		})
	}

	// Check if API key has permission to use this tool
	apiKey, ok := auth.GetAPIKeyFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, ToolResponse{
			Success:   false,
			Error:     "Unauthorized",
			ErrorCode: "UNAUTHORIZED",
			Timestamp: time.Now(),
			Tool:      req.Tool,
		})
	}

	if !apiKey.IsToolAllowed(req.Tool) {
		return c.JSON(http.StatusForbidden, ToolResponse{
			Success:   false,
			Error:     fmt.Sprintf("Tool '%s' is not allowed for this API key", req.Tool),
			ErrorCode: "TOOL_NOT_ALLOWED",
			Timestamp: time.Now(),
			Tool:      req.Tool,
		})
	}

	// Add shop_id from API key if not provided or empty
	if req.Params == nil {
		req.Params = make(map[string]interface{})
	}
	if shopID, exists := req.Params["shop_id"]; !exists || shopID == nil || shopID == "" {
		req.Params["shop_id"] = apiKey.ShopID
	}

	// Execute tool
	ctx := c.Request().Context()
	startTime := time.Now()

	var result interface{}
	var err error

	switch req.Tool {
	// Product Search
	case "search_products":
		result, err = s.invokeSearchProducts(ctx, req.Params)
	// Sales Tools
	case "get_daily_sales":
		result, err = s.invokeGetDailySales(ctx, req.Params)
	case "get_sales_by_date_range":
		result, err = s.invokeGetSalesByDateRange(ctx, req.Params)
	case "get_top_selling_products":
		result, err = s.invokeGetTopSellingProducts(ctx, req.Params)
	case "get_sales_by_seller":
		result, err = s.invokeGetSalesBySeller(ctx, req.Params)
	case "get_monthly_summary":
		result, err = s.invokeGetMonthlySummary(ctx, req.Params)
	// Dashboard Tools
	case "get_dashboard_kpis":
		result, err = s.invokeGetDashboardKPIs(ctx, req.Params)
	case "get_business_health":
		result, err = s.invokeGetBusinessHealth(ctx, req.Params)
	// Financial Tools
	case "get_profit_analysis":
		result, err = s.invokeGetProfitAnalysis(ctx, req.Params)
	case "get_accounts_receivable":
		result, err = s.invokeGetAccountsReceivable(ctx, req.Params)
	case "get_accounts_payable":
		result, err = s.invokeGetAccountsPayable(ctx, req.Params)
	case "get_cash_flow":
		result, err = s.invokeGetCashFlow(ctx, req.Params)
	// Inventory Tools
	case "get_inventory_value":
		result, err = s.invokeGetInventoryValue(ctx, req.Params)
	case "get_low_stock_alerts":
		result, err = s.invokeGetLowStockAlerts(ctx, req.Params)
	case "get_dead_stock":
		result, err = s.invokeGetDeadStock(ctx, req.Params)
	case "get_inventory_turnover":
		result, err = s.invokeGetInventoryTurnover(ctx, req.Params)
	// Customer Tools
	case "get_top_customers":
		result, err = s.invokeGetTopCustomers(ctx, req.Params)
	case "get_customer_growth":
		result, err = s.invokeGetCustomerGrowth(ctx, req.Params)
	case "get_customer_segments":
		result, err = s.invokeGetCustomerSegments(ctx, req.Params)
	// Comparison Tools
	case "get_yoy_comparison":
		result, err = s.invokeGetYoYComparison(ctx, req.Params)
	case "get_mom_comparison":
		result, err = s.invokeGetMoMComparison(ctx, req.Params)
	// Database Tools
	case "get_database_schema":
		result, err = s.invokeGetDatabaseSchema(ctx, req.Params)
	case "execute_query":
		result, err = s.invokeExecuteQuery(ctx, req.Params)
	case "get_table_sample":
		result, err = s.invokeGetTableSample(ctx, req.Params)
	// MongoDB Tools
	case "query_mongodb":
		result, err = s.invokeQueryMongoDB(ctx, req.Params)
	case "list_mongodb_collections":
		result, err = s.invokeListMongoDBCollections(ctx, req.Params)
	case "aggregate_mongodb":
		result, err = s.invokeAggregateMongoDB(ctx, req.Params)
	// ClickHouse Tools
	case "query_clickhouse":
		result, err = s.invokeQueryClickHouse(ctx, req.Params)
	case "list_clickhouse_tables":
		result, err = s.invokeListClickHouseTables(ctx, req.Params)
	// Dev Database Tools
	case "execute_pg_command":
		result, err = s.invokeExecutePgCommand(ctx, req.Params)
	case "execute_ch_command":
		result, err = s.invokeExecuteChCommand(ctx, req.Params)
	// Unit of Measure Tools (หน่วยนับ)
	case "list_units":
		result, err = s.invokeListUnits(ctx, req.Params)
	case "create_unit":
		result, err = s.invokeCreateUnit(ctx, req.Params)
	case "create_units":
		result, err = s.invokeCreateUnits(ctx, req.Params)
	case "update_unit":
		result, err = s.invokeUpdateUnit(ctx, req.Params)
	case "delete_unit":
		result, err = s.invokeDeleteUnit(ctx, req.Params)
	case "delete_units":
		result, err = s.invokeDeleteUnits(ctx, req.Params)
	case "get_unit_schema":
		result, err = s.invokeGetUnitSchema(ctx, req.Params)
	// Product Barcode Tools (สินค้า/บาร์โค้ด)
	case "list_barcodes":
		result, err = s.invokeListBarcodes(ctx, req.Params)
	case "create_barcode":
		result, err = s.invokeCreateBarcode(ctx, req.Params)
	case "create_barcodes":
		result, err = s.invokeCreateBarcodes(ctx, req.Params)
	case "update_barcode":
		result, err = s.invokeUpdateBarcode(ctx, req.Params)
	case "delete_barcode":
		result, err = s.invokeDeleteBarcode(ctx, req.Params)
	case "delete_barcodes":
		result, err = s.invokeDeleteBarcodes(ctx, req.Params)
	case "get_barcode_schema":
		result, err = s.invokeGetBarcodeSchema(ctx, req.Params)
	// Reference Barcodes + Multi-Unit
	case "get_ref_barcodes":
		result, err = s.invokeGetRefBarcodes(ctx, req.Params)
	case "set_ref_barcode":
		result, err = s.invokeSetRefBarcode(ctx, req.Params)
	case "create_multi_unit_barcode":
		result, err = s.invokeCreateMultiUnitBarcode(ctx, req.Params)
	case "rebuild_products":
		result, err = s.invokeRebuildProducts(ctx, req.Params)
	// Product Group Tools (กลุ่มสินค้า)
	case "list_product_groups":
		result, err = s.invokeListProductGroups(ctx, req.Params)
	case "create_product_group":
		result, err = s.invokeCreateProductGroup(ctx, req.Params)
	case "create_product_groups":
		result, err = s.invokeCreateProductGroups(ctx, req.Params)
	case "update_product_group":
		result, err = s.invokeUpdateProductGroup(ctx, req.Params)
	case "delete_product_group":
		result, err = s.invokeDeleteProductGroup(ctx, req.Params)
	case "delete_product_groups":
		result, err = s.invokeDeleteProductGroups(ctx, req.Params)
	case "get_product_group_schema":
		result, err = s.invokeGetProductGroupSchema(ctx, req.Params)
	// Product Category Tools (หมวดสินค้า)
	case "list_product_categories":
		result, err = s.invokeListProductCategories(ctx, req.Params)
	case "create_product_category":
		result, err = s.invokeCreateProductCategory(ctx, req.Params)
	case "create_product_categories":
		result, err = s.invokeCreateProductCategories(ctx, req.Params)
	case "update_product_category":
		result, err = s.invokeUpdateProductCategory(ctx, req.Params)
	case "delete_product_category":
		result, err = s.invokeDeleteProductCategory(ctx, req.Params)
	case "delete_product_categories":
		result, err = s.invokeDeleteProductCategories(ctx, req.Params)
	case "get_product_category_schema":
		result, err = s.invokeGetProductCategorySchema(ctx, req.Params)
	// Creditor Tools (เจ้าหนี้)
	case "list_creditors":
		result, err = s.invokeListCreditors(ctx, req.Params)
	case "create_creditor":
		result, err = s.invokeCreateCreditor(ctx, req.Params)
	case "create_creditors":
		result, err = s.invokeCreateCreditors(ctx, req.Params)
	case "update_creditor":
		result, err = s.invokeUpdateCreditor(ctx, req.Params)
	case "delete_creditor":
		result, err = s.invokeDeleteCreditor(ctx, req.Params)
	case "delete_creditors":
		result, err = s.invokeDeleteCreditors(ctx, req.Params)
	case "get_creditor_schema":
		result, err = s.invokeGetCreditorSchema(ctx, req.Params)
	// Debtor Tools (ลูกหนี้)
	case "list_debtors":
		result, err = s.invokeListDebtors(ctx, req.Params)
	case "create_debtor":
		result, err = s.invokeCreateDebtor(ctx, req.Params)
	case "create_debtors":
		result, err = s.invokeCreateDebtors(ctx, req.Params)
	case "update_debtor":
		result, err = s.invokeUpdateDebtor(ctx, req.Params)
	case "delete_debtor":
		result, err = s.invokeDeleteDebtor(ctx, req.Params)
	case "delete_debtors":
		result, err = s.invokeDeleteDebtors(ctx, req.Params)
	case "get_debtor_schema":
		result, err = s.invokeGetDebtorSchema(ctx, req.Params)
	// API Catalog & Frontend Dev Tools
	case "list_api_endpoints":
		result, err = s.invokeListAPIEndpoints(ctx, req.Params)
	case "get_api_spec":
		result, err = s.invokeGetAPISpec(ctx, req.Params)
	case "get_api_example":
		result, err = s.invokeGetAPIExample(ctx, req.Params)
	case "list_enums":
		result, err = s.invokeListEnums(ctx, req.Params)
	case "get_model_schema":
		result, err = s.invokeGetModelSchema(ctx, req.Params)
	default:
		return c.JSON(http.StatusBadRequest, ToolResponse{
			Success:   false,
			Error:     fmt.Sprintf("Unknown tool: %s", req.Tool),
			ErrorCode: "UNKNOWN_TOOL",
			Timestamp: time.Now(),
			Tool:      req.Tool,
		})
	}

	// Log execution
	executionTime := time.Since(startTime).Milliseconds()
	go s.logAudit(apiKey, req.Tool, req.Params, err, executionTime)

	if err != nil {
		logger.Error("Tool execution failed: %s - %v", req.Tool, err)
		return c.JSON(http.StatusInternalServerError, ToolResponse{
			Success:   false,
			Error:     err.Error(),
			ErrorCode: "EXECUTION_ERROR",
			Timestamp: time.Now(),
			Tool:      req.Tool,
		})
	}

	return c.JSON(http.StatusOK, ToolResponse{
		Success:   true,
		Data:      result,
		Timestamp: time.Now(),
		Tool:      req.Tool,
	})
}

// GetDailySales handles the daily sales tool endpoint
func (s *MCPServer) GetDailySales(c echo.Context) error {
	var params tools.DailySalesRequest
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Get API key data and set shop_id
	apiKey, _ := auth.GetAPIKeyFromContext(c)
	if params.ShopID == "" {
		params.ShopID = apiKey.ShopID
	}

	result, err := s.sales.GetDailySales(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

// GetSalesByDateRange handles the sales by date range tool endpoint
func (s *MCPServer) GetSalesByDateRange(c echo.Context) error {
	var params tools.SalesByDateRangeRequest
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	apiKey, _ := auth.GetAPIKeyFromContext(c)
	if params.ShopID == "" {
		params.ShopID = apiKey.ShopID
	}

	result, err := s.sales.GetSalesByDateRange(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

// GetTopSellingProducts handles the top selling products tool endpoint
func (s *MCPServer) GetTopSellingProducts(c echo.Context) error {
	var params tools.TopSellingProductsRequest
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	apiKey, _ := auth.GetAPIKeyFromContext(c)
	if params.ShopID == "" {
		params.ShopID = apiKey.ShopID
	}

	result, err := s.sales.GetTopSellingProducts(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

// GetSalesBySeller handles the sales by seller tool endpoint
func (s *MCPServer) GetSalesBySeller(c echo.Context) error {
	var params tools.SalesBySellerRequest
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	apiKey, _ := auth.GetAPIKeyFromContext(c)
	if params.ShopID == "" {
		params.ShopID = apiKey.ShopID
	}

	result, err := s.sales.GetSalesBySeller(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

// GetMonthlySummary handles the monthly summary tool endpoint
func (s *MCPServer) GetMonthlySummary(c echo.Context) error {
	var params tools.MonthlySummaryRequest
	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	apiKey, _ := auth.GetAPIKeyFromContext(c)
	if params.ShopID == "" {
		params.ShopID = apiKey.ShopID
	}

	result, err := s.sales.GetMonthlySummary(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

// invokeGetDailySales invokes the daily sales tool
func (s *MCPServer) invokeGetDailySales(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var req tools.DailySalesRequest
	req.ShopID = getStringParam(params, "shop_id")
	req.Date = getStringParam(params, "date")
	req.BranchCode = getStringParam(params, "branch_code")
	return s.sales.GetDailySales(ctx, req)
}

// invokeGetSalesByDateRange invokes the sales by date range tool
func (s *MCPServer) invokeGetSalesByDateRange(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var req tools.SalesByDateRangeRequest
	req.ShopID = getStringParam(params, "shop_id")
	req.FromDate = getStringParam(params, "from_date")
	req.ToDate = getStringParam(params, "to_date")
	req.BranchCode = getStringParam(params, "branch_code")
	req.GroupBy = getStringParam(params, "group_by")
	return s.sales.GetSalesByDateRange(ctx, req)
}

// invokeGetTopSellingProducts invokes the top selling products tool
func (s *MCPServer) invokeGetTopSellingProducts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var req tools.TopSellingProductsRequest
	req.ShopID = getStringParam(params, "shop_id")
	req.FromDate = getStringParam(params, "from_date")
	req.ToDate = getStringParam(params, "to_date")
	req.Limit = getIntParam(params, "limit")
	req.BranchCode = getStringParam(params, "branch_code")
	return s.sales.GetTopSellingProducts(ctx, req)
}

// invokeGetSalesBySeller invokes the sales by seller tool
func (s *MCPServer) invokeGetSalesBySeller(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var req tools.SalesBySellerRequest
	req.ShopID = getStringParam(params, "shop_id")
	req.FromDate = getStringParam(params, "from_date")
	req.ToDate = getStringParam(params, "to_date")
	return s.sales.GetSalesBySeller(ctx, req)
}

// invokeGetMonthlySummary invokes the monthly summary tool
func (s *MCPServer) invokeGetMonthlySummary(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var req tools.MonthlySummaryRequest
	req.ShopID = getStringParam(params, "shop_id")
	req.Year = getIntParam(params, "year")
	req.Month = getIntParam(params, "month")
	return s.sales.GetMonthlySummary(ctx, req)
}

// invokeSearchProducts invokes the product search tool with Thai full-text search
func (s *MCPServer) invokeSearchProducts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	keyword := getStringParam(params, "keyword")
	whcode := getStringParam(params, "whcode")
	locationcode := getStringParam(params, "locationcode")
	limit := getIntParam(params, "limit")
	includeBalance := getBoolParam(params, "include_balance")

	if keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}

	// Default values
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	// Default to true if not specified
	if _, exists := params["include_balance"]; !exists {
		includeBalance = true
	}

	// Call the product search tool
	return tools.SearchProducts(ctx, shopID, keyword, whcode, locationcode, limit, includeBalance)
}

// ==================== Dashboard Tool Invocations ====================

// invokeGetDashboardKPIs invokes the dashboard KPIs tool
func (s *MCPServer) invokeGetDashboardKPIs(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	period := getStringParam(params, "period")
	if period == "" {
		period = "this_month"
	}
	return tools.GetDashboardKPIs(ctx, shopID, period)
}

// invokeGetBusinessHealth invokes the business health tool
func (s *MCPServer) invokeGetBusinessHealth(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	return tools.GetBusinessHealth(ctx, shopID)
}

// ==================== Financial Tool Invocations ====================

// invokeGetProfitAnalysis invokes the profit analysis tool
func (s *MCPServer) invokeGetProfitAnalysis(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	fromDate := getStringParam(params, "from_date")
	toDate := getStringParam(params, "to_date")
	return tools.GetProfitAnalysis(ctx, shopID, fromDate, toDate)
}

// invokeGetAccountsReceivable invokes the accounts receivable tool
func (s *MCPServer) invokeGetAccountsReceivable(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	return tools.GetAccountsReceivable(ctx, shopID)
}

// invokeGetAccountsPayable invokes the accounts payable tool
func (s *MCPServer) invokeGetAccountsPayable(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	return tools.GetAccountsPayable(ctx, shopID)
}

// invokeGetCashFlow invokes the cash flow tool
func (s *MCPServer) invokeGetCashFlow(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	fromDate := getStringParam(params, "from_date")
	toDate := getStringParam(params, "to_date")
	return tools.GetCashFlow(ctx, shopID, fromDate, toDate)
}

// ==================== Inventory Tool Invocations ====================

// invokeGetInventoryValue invokes the inventory value tool
func (s *MCPServer) invokeGetInventoryValue(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	whcode := getStringParam(params, "whcode")
	return tools.GetInventoryValue(ctx, shopID, whcode)
}

// invokeGetLowStockAlerts invokes the low stock alerts tool
func (s *MCPServer) invokeGetLowStockAlerts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	threshold := getIntParam(params, "threshold")
	limit := getIntParam(params, "limit")
	if threshold <= 0 {
		threshold = 10
	}
	if limit <= 0 {
		limit = 50
	}
	return tools.GetLowStockAlerts(ctx, shopID, threshold, limit)
}

// invokeGetDeadStock invokes the dead stock tool
func (s *MCPServer) invokeGetDeadStock(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	daysNoMovement := getIntParam(params, "days_no_movement")
	limit := getIntParam(params, "limit")
	if daysNoMovement <= 0 {
		daysNoMovement = 90
	}
	if limit <= 0 {
		limit = 50
	}
	return tools.GetDeadStock(ctx, shopID, daysNoMovement, limit)
}

// invokeGetInventoryTurnover invokes the inventory turnover tool
func (s *MCPServer) invokeGetInventoryTurnover(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	fromDate := getStringParam(params, "from_date")
	toDate := getStringParam(params, "to_date")
	return tools.GetInventoryTurnover(ctx, shopID, fromDate, toDate)
}

// ==================== Customer Tool Invocations ====================

// invokeGetTopCustomers invokes the top customers tool
func (s *MCPServer) invokeGetTopCustomers(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	fromDate := getStringParam(params, "from_date")
	toDate := getStringParam(params, "to_date")
	limit := getIntParam(params, "limit")
	sortBy := getStringParam(params, "sort_by")
	if limit <= 0 {
		limit = 10
	}
	if sortBy == "" {
		sortBy = "amount"
	}
	return tools.GetTopCustomers(ctx, shopID, fromDate, toDate, limit, sortBy)
}

// invokeGetCustomerGrowth invokes the customer growth tool
func (s *MCPServer) invokeGetCustomerGrowth(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	fromDate := getStringParam(params, "from_date")
	toDate := getStringParam(params, "to_date")
	return tools.GetCustomerGrowth(ctx, shopID, fromDate, toDate)
}

// invokeGetCustomerSegments invokes the customer segments tool
func (s *MCPServer) invokeGetCustomerSegments(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	fromDate := getStringParam(params, "from_date")
	toDate := getStringParam(params, "to_date")
	return tools.GetCustomerSegments(ctx, shopID, fromDate, toDate)
}

// ==================== Comparison Tool Invocations ====================

// invokeGetYoYComparison invokes the year-over-year comparison tool
func (s *MCPServer) invokeGetYoYComparison(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	year := getIntParam(params, "year")
	month := getIntParam(params, "month")
	return tools.GetYoYComparison(ctx, shopID, year, month)
}

// invokeGetMoMComparison invokes the month-over-month comparison tool
func (s *MCPServer) invokeGetMoMComparison(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	year := getIntParam(params, "year")
	month := getIntParam(params, "month")
	return tools.GetMoMComparison(ctx, shopID, year, month)
}

// ==================== Database Tool Invocations ====================

// invokeGetDatabaseSchema invokes the database schema tool
func (s *MCPServer) invokeGetDatabaseSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	tableName := getStringParam(params, "table_name")
	return tools.GetDatabaseSchema(ctx, shopID, tableName)
}

// invokeExecuteQuery invokes the execute query tool (readonly)
func (s *MCPServer) invokeExecuteQuery(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	query := getStringParam(params, "query")
	limit := getIntParam(params, "limit")
	if limit <= 0 {
		limit = 100
	}
	return tools.ExecuteReadonlyQuery(ctx, shopID, query, limit)
}

// invokeGetTableSample invokes the table sample tool
func (s *MCPServer) invokeGetTableSample(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	tableName := getStringParam(params, "table_name")
	limit := getIntParam(params, "limit")
	if limit <= 0 {
		limit = 10
	}
	return tools.GetTableSample(ctx, shopID, tableName, limit)
}

// ==================== MongoDB Tool Invocations ====================

// invokeQueryMongoDB invokes the MongoDB query tool (readonly)
func (s *MCPServer) invokeQueryMongoDB(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	database := getStringParam(params, "database")
	collection := getStringParam(params, "collection")
	filter := getStringParam(params, "filter")
	limit := getIntParam(params, "limit")
	return tools.QueryMongoDB(ctx, shopID, database, collection, filter, limit)
}

// invokeListMongoDBCollections invokes the MongoDB list collections tool
func (s *MCPServer) invokeListMongoDBCollections(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	database := getStringParam(params, "database")
	return tools.ListMongoDBCollections(ctx, database)
}

// invokeAggregateMongoDB invokes the MongoDB aggregation tool (readonly)
func (s *MCPServer) invokeAggregateMongoDB(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	database := getStringParam(params, "database")
	collection := getStringParam(params, "collection")
	pipeline := getStringParam(params, "pipeline")
	limit := getIntParam(params, "limit")
	return tools.AggregateMongoDB(ctx, shopID, database, collection, pipeline, limit)
}

// ==================== ClickHouse Tool Invocations ====================

// invokeQueryClickHouse invokes the ClickHouse query tool (readonly)
func (s *MCPServer) invokeQueryClickHouse(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	database := getStringParam(params, "database")
	query := getStringParam(params, "query")
	limit := getIntParam(params, "limit")
	return tools.QueryClickHouse(ctx, shopID, database, query, limit)
}

// invokeListClickHouseTables invokes the ClickHouse list tables tool
func (s *MCPServer) invokeListClickHouseTables(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	database := getStringParam(params, "database")
	return tools.ListClickHouseTables(ctx, database)
}

// invokeExecutePgCommand invokes the dev PostgreSQL command tool
func (s *MCPServer) invokeExecutePgCommand(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	query := getStringParam(params, "query")
	limit := getIntParam(params, "limit")
	return tools.ExecutePgCommand(ctx, shopID, query, limit)
}

// invokeExecuteChCommand invokes the dev ClickHouse command tool
func (s *MCPServer) invokeExecuteChCommand(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	database := getStringParam(params, "database")
	query := getStringParam(params, "query")
	limit := getIntParam(params, "limit")
	return tools.ExecuteChCommand(ctx, database, query, limit)
}

// logAudit logs tool execution to database
func (s *MCPServer) logAudit(apiKey *mongodb.APIKey, toolName string, params map[string]interface{}, err error, executionTimeMs int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status := "success"
	errorMsg := ""
	if err != nil {
		status = "error"
		errorMsg = err.Error()
	}

	paramsJSON, _ := json.Marshal(params)
	var paramsBson map[string]interface{}
	json.Unmarshal(paramsJSON, &paramsBson)

	log := &mongodb.AuditLog{
		APIKeyID:        apiKey.ID,
		ShopID:          apiKey.ShopID,
		ToolName:        toolName,
		RequestParams:   paramsBson,
		ResponseStatus:  status,
		ErrorMessage:    errorMsg,
		ExecutionTimeMs: executionTimeMs,
	}

	if err := s.keysRepo.CreateAuditLog(ctx, log); err != nil {
		logger.Error("Failed to create audit log: %v", err)
	}
}

// Helper functions for parameter parsing
func getStringParam(params map[string]interface{}, key string) string {
	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getIntParam(params map[string]interface{}, key string) int {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case int64:
			return int(v)
		}
	}
	return 0
}

func getBoolParam(params map[string]interface{}, key string) bool {
	if val, ok := params[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func getFloatParam(params map[string]interface{}, key string) float64 {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

// getBoolPtrParam returns *bool (nil if not provided, pointer to value if provided)
func getBoolPtrParam(params map[string]interface{}, key string) *bool {
	if val, ok := params[key]; ok {
		if b, ok := val.(bool); ok {
			return &b
		}
	}
	return nil
}

// invokeListAPIEndpoints invokes the API catalog tool
func (s *MCPServer) invokeListAPIEndpoints(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	req := tools.APICatalogRequest{
		Category: getStringParam(params, "category"),
		Keyword:  getStringParam(params, "keyword"),
		Method:   getStringParam(params, "method"),
		Source:   getStringParam(params, "source"),
		Limit:    getIntParam(params, "limit"),
	}
	return tools.GetAPICatalog(req)
}

// invokeGetAPISpec invokes the API spec tool
func (s *MCPServer) invokeGetAPISpec(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	req := tools.APISpecRequest{
		Path:   getStringParam(params, "path"),
		Method: getStringParam(params, "method"),
	}
	return tools.GetAPISpec(req)
}

// invokeGetAPIExample invokes the API example tool
func (s *MCPServer) invokeGetAPIExample(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	req := tools.APIExampleRequest{
		Path:   getStringParam(params, "path"),
		Method: getStringParam(params, "method"),
	}
	return tools.GetAPIExample(req)
}

// invokeListEnums invokes the enum catalog tool
func (s *MCPServer) invokeListEnums(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	req := tools.EnumCatalogRequest{
		Category: getStringParam(params, "category"),
		Keyword:  getStringParam(params, "keyword"),
	}
	return tools.GetEnumCatalog(req)
}

// invokeGetModelSchema invokes the model schema tool
func (s *MCPServer) invokeGetModelSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	req := tools.ModelSchemaRequest{
		Model:    getStringParam(params, "model"),
		Category: getStringParam(params, "category"),
		Keyword:  getStringParam(params, "keyword"),
	}
	return tools.GetModelSchema(req)
}

// ==================== Unit of Measure Tool Invocations ====================

// invokeListUnits invokes the list/search units tool
func (s *MCPServer) invokeListUnits(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.ListUnits(ctx, shopID, keyword, limit)
}

// invokeCreateUnit invokes the create unit tool
func (s *MCPServer) invokeCreateUnit(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	unitCode := getStringParam(params, "unitcode")
	names := getStringParam(params, "names")
	return tools.CreateUnit(ctx, shopID, unitCode, names)
}

// invokeCreateUnits invokes the bulk create units tool
func (s *MCPServer) invokeCreateUnits(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	units := getStringParam(params, "units")
	return tools.CreateUnits(ctx, shopID, units)
}

// invokeUpdateUnit invokes the update unit tool
func (s *MCPServer) invokeUpdateUnit(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	unitCode := getStringParam(params, "unitcode")
	names := getStringParam(params, "names")
	return tools.UpdateUnit(ctx, shopID, unitCode, names)
}

// invokeDeleteUnit invokes the delete unit tool
func (s *MCPServer) invokeDeleteUnit(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	unitCode := getStringParam(params, "unitcode")
	return tools.DeleteUnit(ctx, shopID, unitCode)
}

// invokeDeleteUnits invokes the bulk delete units tool
func (s *MCPServer) invokeDeleteUnits(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	unitcodes := getStringParam(params, "unitcodes")
	return tools.DeleteUnits(ctx, shopID, unitcodes)
}

// invokeGetUnitSchema invokes the get unit schema tool
func (s *MCPServer) invokeGetUnitSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return tools.GetUnitSchema(), nil
}

// ==================== Product Barcode Tool Invocations ====================

// invokeListBarcodes invokes the list/search barcodes tool
func (s *MCPServer) invokeListBarcodes(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.ListBarcodes(ctx, shopID, keyword, limit)
}

// invokeCreateBarcode invokes the create barcode tool
func (s *MCPServer) invokeCreateBarcode(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	barcode := getStringParam(params, "barcode")
	itemCode := getStringParam(params, "itemcode")
	names := getStringParam(params, "names")
	itemUnitCode := getStringParam(params, "itemunitcode")
	itemUnitNames := getStringParam(params, "itemunitnames")
	prices := getStringParam(params, "prices")
	standValue := getFloatParam(params, "standvalue")
	divideValue := getFloatParam(params, "dividevalue")
	isMainBarcode := getBoolPtrParam(params, "ismainbarcode")
	groupCode := getStringParam(params, "groupcode")
	groupNames := getStringParam(params, "groupnames")
	categoryCode := getStringParam(params, "categorycode")
	categoryNames := getStringParam(params, "categorynames")
	return tools.CreateBarcode(ctx, shopID, barcode, itemCode, names, itemUnitCode, itemUnitNames, prices, standValue, divideValue, isMainBarcode, groupCode, groupNames, categoryCode, categoryNames)
}

// invokeCreateBarcodes invokes the bulk create barcodes tool
func (s *MCPServer) invokeCreateBarcodes(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	barcodes := getStringParam(params, "barcodes")
	return tools.CreateBarcodes(ctx, shopID, barcodes)
}

// invokeUpdateBarcode invokes the update barcode tool
func (s *MCPServer) invokeUpdateBarcode(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	guidFixed := getStringParam(params, "guidfixed")
	names := getStringParam(params, "names")
	itemUnitCode := getStringParam(params, "itemunitcode")
	itemUnitNames := getStringParam(params, "itemunitnames")
	prices := getStringParam(params, "prices")
	groupCode := getStringParam(params, "groupcode")
	groupNames := getStringParam(params, "groupnames")
	categoryCode := getStringParam(params, "categorycode")
	categoryNames := getStringParam(params, "categorynames")
	return tools.UpdateBarcode(ctx, shopID, guidFixed, names, itemUnitCode, itemUnitNames, prices, groupCode, groupNames, categoryCode, categoryNames)
}

// invokeDeleteBarcode invokes the delete barcode tool
func (s *MCPServer) invokeDeleteBarcode(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	guidFixed := getStringParam(params, "guidfixed")
	return tools.DeleteBarcode(ctx, shopID, guidFixed)
}

// invokeDeleteBarcodes invokes the bulk delete barcodes tool
func (s *MCPServer) invokeDeleteBarcodes(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	guidfixeds := getStringParam(params, "guidfixeds")
	return tools.DeleteBarcodes(ctx, shopID, guidfixeds)
}

// invokeGetBarcodeSchema invokes the get barcode schema tool
func (s *MCPServer) invokeGetBarcodeSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return tools.GetBarcodeSchema(), nil
}

// invokeGetRefBarcodes invokes the get reference barcodes tool
func (s *MCPServer) invokeGetRefBarcodes(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	itemCode := getStringParam(params, "itemcode")
	barcode := getStringParam(params, "barcode")
	return tools.GetRefBarcodes(ctx, shopID, itemCode, barcode)
}

// invokeSetRefBarcode invokes the set reference barcode tool
func (s *MCPServer) invokeSetRefBarcode(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	barcode := getStringParam(params, "barcode")
	refBarcode := getStringParam(params, "ref_barcode")
	qty := getFloatParam(params, "qty")
	standValue := getFloatParam(params, "standvalue")
	divideValue := getFloatParam(params, "dividevalue")
	condition := getBoolParam(params, "condition")
	return tools.SetRefBarcode(ctx, shopID, barcode, refBarcode, qty, standValue, divideValue, condition)
}

// invokeCreateMultiUnitBarcode invokes the create multi-unit barcode tool
func (s *MCPServer) invokeCreateMultiUnitBarcode(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	itemCode := getStringParam(params, "itemcode")
	names := getStringParam(params, "names")
	units := getStringParam(params, "units")
	return tools.CreateMultiUnitBarcode(ctx, shopID, itemCode, names, units)
}

func (s *MCPServer) invokeRebuildProducts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	return tools.RebuildProducts(ctx, shopID)
}

// ==================== Product Group Tool Invocations ====================

func (s *MCPServer) invokeListProductGroups(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.ListProductGroups(ctx, shopID, keyword, limit)
}

func (s *MCPServer) invokeCreateProductGroup(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	code := getStringParam(params, "code")
	names := getStringParam(params, "names")
	return tools.CreateProductGroup(ctx, shopID, code, names)
}

func (s *MCPServer) invokeCreateProductGroups(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	groups := getStringParam(params, "groups")
	return tools.CreateProductGroups(ctx, shopID, groups)
}

func (s *MCPServer) invokeUpdateProductGroup(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	code := getStringParam(params, "code")
	names := getStringParam(params, "names")
	return tools.UpdateProductGroup(ctx, shopID, code, names)
}

func (s *MCPServer) invokeDeleteProductGroup(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	code := getStringParam(params, "code")
	return tools.DeleteProductGroup(ctx, shopID, code)
}

func (s *MCPServer) invokeDeleteProductGroups(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	codes := getStringParam(params, "codes")
	return tools.DeleteProductGroups(ctx, shopID, codes)
}

func (s *MCPServer) invokeGetProductGroupSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return tools.GetProductGroupSchema(), nil
}

// ==================== Product Category Tool Invocations ====================

func (s *MCPServer) invokeListProductCategories(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.ListProductCategories(ctx, shopID, keyword, limit)
}

func (s *MCPServer) invokeCreateProductCategory(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	names := getStringParam(params, "names")
	parentGUID := getStringParam(params, "parentguid")
	groupNumber := getIntParam(params, "groupnumber")
	return tools.CreateProductCategory(ctx, shopID, names, parentGUID, groupNumber)
}

func (s *MCPServer) invokeCreateProductCategories(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	categories := getStringParam(params, "categories")
	return tools.CreateProductCategories(ctx, shopID, categories)
}

func (s *MCPServer) invokeUpdateProductCategory(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	guidFixed := getStringParam(params, "guidfixed")
	names := getStringParam(params, "names")
	parentGUID := getStringParam(params, "parentguid")
	groupNumber := getIntParam(params, "groupnumber")
	return tools.UpdateProductCategory(ctx, shopID, guidFixed, names, parentGUID, groupNumber)
}

func (s *MCPServer) invokeDeleteProductCategory(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	guidFixed := getStringParam(params, "guidfixed")
	return tools.DeleteProductCategory(ctx, shopID, guidFixed)
}

func (s *MCPServer) invokeDeleteProductCategories(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	guidfixeds := getStringParam(params, "guidfixeds")
	return tools.DeleteProductCategories(ctx, shopID, guidfixeds)
}

func (s *MCPServer) invokeGetProductCategorySchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return tools.GetProductCategorySchema(), nil
}

// ==================== Creditor Invokers ====================

func (s *MCPServer) invokeListCreditors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.ListCreditors(ctx, shopID, keyword, limit)
}

func (s *MCPServer) invokeCreateCreditor(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	code := getStringParam(params, "code")
	names := getStringParam(params, "names")
	taxid := getStringParam(params, "taxid")
	email := getStringParam(params, "email")
	personaltype := int8(getIntParam(params, "personaltype"))
	creditday := getIntParam(params, "creditday")
	address := getStringParam(params, "addressforbilling")
	return tools.CreateCreditor(ctx, shopID, code, names, taxid, email, personaltype, creditday, address)
}

func (s *MCPServer) invokeCreateCreditors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	creditors := getStringParam(params, "creditors")
	return tools.CreateCreditors(ctx, shopID, creditors)
}

func (s *MCPServer) invokeUpdateCreditor(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	guidfixed := getStringParam(params, "guidfixed")
	names := getStringParam(params, "names")
	taxid := getStringParam(params, "taxid")
	email := getStringParam(params, "email")
	creditday := getIntParam(params, "creditday")
	address := getStringParam(params, "addressforbilling")
	return tools.UpdateCreditor(ctx, shopID, guidfixed, names, taxid, email, creditday, address)
}

func (s *MCPServer) invokeDeleteCreditor(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	guidfixed := getStringParam(params, "guidfixed")
	return tools.DeleteCreditor(ctx, shopID, guidfixed)
}

func (s *MCPServer) invokeDeleteCreditors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	guidfixeds := getStringParam(params, "guidfixeds")
	return tools.DeleteCreditors(ctx, shopID, guidfixeds)
}

func (s *MCPServer) invokeGetCreditorSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return tools.GetCreditorSchema(), nil
}

// ==================== Debtor Invokers ====================

func (s *MCPServer) invokeListDebtors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.ListDebtors(ctx, shopID, keyword, limit)
}

func (s *MCPServer) invokeCreateDebtor(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	code := getStringParam(params, "code")
	names := getStringParam(params, "names")
	taxid := getStringParam(params, "taxid")
	email := getStringParam(params, "email")
	personaltype := int8(getIntParam(params, "personaltype"))
	creditday := getIntParam(params, "creditday")
	address := getStringParam(params, "addressforbilling")
	return tools.CreateDebtor(ctx, shopID, code, names, taxid, email, personaltype, creditday, address)
}

func (s *MCPServer) invokeCreateDebtors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	debtors := getStringParam(params, "debtors")
	return tools.CreateDebtors(ctx, shopID, debtors)
}

func (s *MCPServer) invokeUpdateDebtor(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	guidfixed := getStringParam(params, "guidfixed")
	names := getStringParam(params, "names")
	taxid := getStringParam(params, "taxid")
	email := getStringParam(params, "email")
	creditday := getIntParam(params, "creditday")
	address := getStringParam(params, "addressforbilling")
	return tools.UpdateDebtor(ctx, shopID, guidfixed, names, taxid, email, creditday, address)
}

func (s *MCPServer) invokeDeleteDebtor(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	guidfixed := getStringParam(params, "guidfixed")
	return tools.DeleteDebtor(ctx, shopID, guidfixed)
}

func (s *MCPServer) invokeDeleteDebtors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	shopID := getStringParam(params, "shop_id")
	guidfixeds := getStringParam(params, "guidfixeds")
	return tools.DeleteDebtors(ctx, shopID, guidfixeds)
}

func (s *MCPServer) invokeGetDebtorSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return tools.GetDebtorSchema(), nil
}

// RegisterRoutesOnGroup registers MCP routes on an Echo Group (for embedded mode)
func (s *MCPServer) RegisterRoutesOnGroup(g *echo.Group) {
	mcpGroup := g.Group("/mcp")
	mcpGroup.Use(s.auth.APIKeyAuth)

	mcpGroup.POST("/invoke", s.InvokeTool)
	mcpGroup.POST("/tools/daily-sales", s.GetDailySales)
	mcpGroup.POST("/tools/sales-by-date-range", s.GetSalesByDateRange)
	mcpGroup.POST("/tools/top-selling-products", s.GetTopSellingProducts)
	mcpGroup.POST("/tools/sales-by-seller", s.GetSalesBySeller)
	mcpGroup.POST("/tools/monthly-summary", s.GetMonthlySummary)

	// General endpoints (ซ่อน dev tools)
	g.GET("/mcp/tools", s.ListTools)
	g.GET("/mcp/health", s.HealthCheck)

	// Dev endpoints (เห็นทุก tools)
	g.GET("/mcp/dev/tools", s.ListToolsDev)
	g.GET("/mcp/dev/health", s.HealthCheck)
}

// RegisterSSERoutesOnGroup registers SSE routes on an Echo Group with prefix awareness
func (s *MCPServer) RegisterSSERoutesOnGroup(g *echo.Group, prefix string) {
	s.ssePrefix = prefix
	// General endpoint — ซ่อน dev tools
	g.GET("/mcp/sse", s.HandleSSE)
	g.POST("/mcp/message", s.HandleMessage)
	// Dev endpoint — เห็นทุก tools (database, schema, raw SQL, rebuild, etc.)
	g.GET("/mcp/dev/sse", s.HandleSSEDev)
	g.POST("/mcp/dev/message", s.HandleMessage)
}
