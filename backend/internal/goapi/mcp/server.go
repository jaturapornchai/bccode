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
	ErrorCode string      `json:"errorcode,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Tool      string      `json:"tool"`
}

// AvailableTools lists all available MCP tools
var AvailableTools = []map[string]interface{}{
	// Product Search
	{
		"name":        "searchproducts",
		"description": "Search products from the PostgreSQL projection/read model built from MongoDB operational product data, with Thai full-text search and stock balance.",
		"parameters": map[string]interface{}{
			"holdingcode":    "string (required) - Holding Code",
			"keyword":        "string (required) - Search keyword (Thai or English)",
			"whcode":         "string (optional) - Warehouse code to filter stock balance",
			"locationcode":   "string (optional) - Location code to filter stock balance",
			"limit":          "number (optional) - Max products to return (default: 50, max: 200)",
			"includebalance": "boolean (optional) - Include stock balance (default: true)",
		},
	},
	// Sales Tools
	{
		"name":        "getdailysales",
		"description": "Get daily sales summary for a specific date",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"date":        "string (required) - Date in YYYY-MM-DD format",
			"branchcode":  "string (optional) - Filter by branch/warehouse code",
		},
	},
	{
		"name":        "getsalesbydaterange",
		"description": "Get sales data grouped by day/week/month for a date range",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"fromdate":    "string (required) - Start date in YYYY-MM-DD format",
			"todate":      "string (required) - End date in YYYY-MM-DD format",
			"branchcode":  "string (optional) - Filter by branch/warehouse code",
			"groupby":     "string (optional) - Group by: day, week, month (default: day)",
		},
	},
	{
		"name":        "gettopsellingproducts",
		"description": "Get top selling products for a date range",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"fromdate":    "string (required) - Start date in YYYY-MM-DD format",
			"todate":      "string (required) - End date in YYYY-MM-DD format",
			"limit":       "number (optional) - Number of products to return (default: 10, max: 100)",
			"branchcode":  "string (optional) - Filter by branch/warehouse code",
		},
	},
	{
		"name":        "getsalesbyseller",
		"description": "Get sales data grouped by seller/salesperson",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"fromdate":    "string (required) - Start date in YYYY-MM-DD format",
			"todate":      "string (required) - End date in YYYY-MM-DD format",
		},
	},
	{
		"name":        "getmonthlysummary",
		"description": "Get monthly sales summary with comparison to previous month",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"year":        "number (required) - Year (e.g., 2025)",
			"month":       "number (required) - Month (1-12)",
		},
	},
	// Dashboard Tools
	{
		"name":        "getdashboardkpis",
		"description": "Get comprehensive KPI dashboard from processed relational projections. Includes sales, orders, profit, customers, inventory metrics with trends.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"period":      "string (optional) - Period: today, thisweek, thismonth, thisyear (default: thismonth)",
		},
	},
	{
		"name":        "getbusinesshealth",
		"description": "Get overall business health score and key indicators. Returns health score (0-100), alerts, and recommendations.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
		},
	},
	// Financial Tools
	{
		"name":        "getprofitanalysis",
		"description": "Get detailed profit analysis with revenue breakdown by category, gross margin, and profit trends.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"fromdate":    "string (required) - Start date in YYYY-MM-DD format",
			"todate":      "string (required) - End date in YYYY-MM-DD format",
		},
	},
	{
		"name":        "getaccountsreceivable",
		"description": "Get accounts receivable summary with aging buckets (current, 30, 60, 90+ days) and top debtors.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"limit":       "number (optional) - Number of top debtors to return (default: 10)",
		},
	},
	{
		"name":        "getaccountspayable",
		"description": "Get accounts payable summary with aging buckets and top creditors.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"limit":       "number (optional) - Number of top creditors to return (default: 10)",
		},
	},
	{
		"name":        "getcashflow",
		"description": "Get cash flow analysis showing inflows, outflows, and net position over time.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"fromdate":    "string (required) - Start date in YYYY-MM-DD format",
			"todate":      "string (required) - End date in YYYY-MM-DD format",
		},
	},
	// Inventory Tools
	{
		"name":        "getinventoryvalue",
		"description": "Get inventory valuation from PostgreSQL relational projections, with breakdown by category and warehouse.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"whcode":      "string (optional) - Filter by warehouse code",
		},
	},
	{
		"name":        "getlowstockalerts",
		"description": "Get products that are below minimum stock level or out of stock.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"threshold":   "number (optional) - Stock threshold to consider low (default: 10)",
			"limit":       "number (optional) - Number of alerts to return (default: 50)",
		},
	},
	{
		"name":        "getdeadstock",
		"description": "Get products with no movement for specified days (slow-moving/dead stock).",
		"parameters": map[string]interface{}{
			"holdingcode":    "string (required) - Holding Code",
			"daysnomovement": "number (optional) - Days without movement (default: 90)",
			"limit":          "number (optional) - Number of products to return (default: 50)",
		},
	},
	{
		"name":        "getinventoryturnover",
		"description": "Get inventory turnover ratio and days of inventory for products.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"fromdate":    "string (required) - Start date in YYYY-MM-DD format",
			"todate":      "string (required) - End date in YYYY-MM-DD format",
			"limit":       "number (optional) - Number of products to return (default: 50)",
		},
	},
	// Customer Tools
	{
		"name":        "gettopcustomers",
		"description": "Get top customers by revenue with purchase history and trends.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"fromdate":    "string (required) - Start date in YYYY-MM-DD format",
			"todate":      "string (required) - End date in YYYY-MM-DD format",
			"limit":       "number (optional) - Number of customers to return (default: 10)",
		},
	},
	{
		"name":        "getcustomergrowth",
		"description": "Get customer acquisition and retention metrics over time.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"fromdate":    "string (required) - Start date in YYYY-MM-DD format",
			"todate":      "string (required) - End date in YYYY-MM-DD format",
		},
	},
	{
		"name":        "getcustomersegments",
		"description": "Get customer segmentation analysis (RFM: Recency, Frequency, Monetary).",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
		},
	},
	// Comparison Tools
	{
		"name":        "getyoycomparison",
		"description": "Get year-over-year comparison of revenue, orders, and profit with monthly breakdown.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"year":        "number (optional) - Year to compare (default: current year)",
			"month":       "number (optional) - Specific month to compare (1-12, optional)",
		},
	},
	{
		"name":        "getmomcomparison",
		"description": "Get month-over-month comparison with weekly breakdown and daily trends.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"year":        "number (optional) - Year (default: current year)",
			"month":       "number (optional) - Month to compare (1-12, default: current month)",
		},
	},
	// Database Tools
	{
		"name":        "getdatabaseschema",
		"description": "Get PostgreSQL database structure (tables, columns, types, relationships). PostgreSQL is used for relational processing — joins, aggregations, reports. For raw data, use MongoDB. For OLAP analytics, use ClickHouse.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"tablename":   "string (optional) - Filter by table name (partial match)",
		},
	},
	{
		"name":        "executequery",
		"description": "Execute a readonly SQL query on PostgreSQL (SELECT only). PostgreSQL handles relational processing — joins, aggregations, reports.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"query":       "string (required) - SQL SELECT query",
			"limit":       "number (optional) - Max rows to return (default: 100, max: 1000)",
		},
	},
	{
		"name":        "gettablesample",
		"description": "Get sample data from a PostgreSQL table. Quick way to see what relational data looks like.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"tablename":   "string (required) - Table name",
			"limit":       "number (optional) - Number of rows (default: 10, max: 100)",
		},
	},
	// MongoDB Tools — Main data store (source of truth)
	{
		"name":        "querymongodb",
		"description": "Query MongoDB collection (readonly). MongoDB is the main data store (source of truth) — all documents, transactions, and master data live here.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (optional) - Holding Code for data isolation",
			"database":    "string (optional) - Database name (default: from config)",
			"collection":  "string (required) - Collection name",
			"filter":      "string (optional) - JSON filter e.g. {\"transflag\":6} (default: {})",
			"limit":       "number (optional) - Max documents (default: 20, max: 100)",
		},
	},
	{
		"name":        "listmongodbcollections",
		"description": "List all collections in MongoDB (main data store). Use to discover available raw data.",
		"parameters": map[string]interface{}{
			"database": "string (optional) - Database name (default: from config)",
		},
	},
	{
		"name":        "aggregatemongodb",
		"description": "Run aggregation pipeline on MongoDB (main data store, readonly). Blocks $out and $merge stages. Use for complex queries on raw source data.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (optional) - Holding Code for data isolation",
			"database":    "string (optional) - Database name (default: from config)",
			"collection":  "string (required) - Collection name",
			"pipeline":    "string (required) - JSON array of pipeline stages e.g. [{\"$match\":{\"transflag\":6}},{\"$group\":{\"_id\":\"$currency\",\"count\":{\"$sum\":1}}}]",
			"limit":       "number (optional) - Max results (default: 100, max: 100)",
		},
	},
	// ClickHouse Tools — Dimensional/OLAP processing
	{
		"name":        "queryclickhouse",
		"description": "Execute readonly SELECT/SHOW query on ClickHouse (OLAP/dimensional analytics). Use for time-series analysis, BI dashboards, and large-scale aggregations.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (optional) - Holding Code for data isolation",
			"database":    "string (optional) - Database name (default: from env CH_DATABASE_NAME)",
			"query":       "string (required) - SQL SELECT or SHOW query",
			"limit":       "number (optional) - Max rows (default: 100, max: 1000)",
		},
	},
	{
		"name":        "listclickhousetables",
		"description": "List all tables in ClickHouse (OLAP/dimensional) with engine, row count and size info. Use to discover analytics data.",
		"parameters": map[string]interface{}{
			"database": "string (optional) - Database name (default: from env CH_DATABASE_NAME)",
		},
	},
	// Dev Database Tools — ไม่จำกัด readonly (ใช้ด้วยความระมัดระวัง)
	{
		"name":        "executepgcommand",
		"description": "⚡ DEV TOOL: Execute ANY SQL on PostgreSQL (SELECT, DELETE, INSERT, UPDATE, ALTER, TRUNCATE, DROP). No readonly restriction. Use with caution.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code (= PostgreSQL database name)",
			"query":       "string (required) - Any SQL command",
			"limit":       "number (optional) - Max rows for SELECT (default: 100, max: 10000)",
		},
	},
	{
		"name":        "executechcommand",
		"description": "⚡ DEV TOOL: Execute ANY SQL on ClickHouse (SELECT, ALTER TABLE DELETE, INSERT, DROP, TRUNCATE). No readonly restriction. Use with caution.",
		"parameters": map[string]interface{}{
			"database": "string (optional) - Database name (default: from env CH_DATABASE_NAME)",
			"query":    "string (required) - Any SQL command",
			"limit":    "number (optional) - Max rows for SELECT (default: 100, max: 10000)",
		},
	},
	// API Catalog Tool
	{
		"name":        "listapiendpoints",
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
		"name":        "getapispec",
		"description": "Get detailed request/response specification for a specific API endpoint. Returns full spec with parameters, request body schema, response schema, curl example, and usage notes.",
		"parameters": map[string]interface{}{
			"path":   "string (required) - API path or partial match (e.g., '/product/search', '/login')",
			"method": "string (optional) - HTTP method filter (GET, POST, PUT, DELETE)",
		},
	},
	// API Example Tool
	{
		"name":        "getapiexample",
		"description": "Get real request/response examples for a specific API endpoint. Returns curl command, request body, response body, Dart code snippet, and TypeScript code snippet.",
		"parameters": map[string]interface{}{
			"path":   "string (required) - API path or partial match (e.g., '/product/search', '/transaction/calculate')",
			"method": "string (optional) - HTTP method filter",
		},
	},
	// Enum Catalog Tool
	{
		"name":        "listenums",
		"description": "List all enum values and constants used in the backend (transflag, payment types, approval status, VAT types, etc.). Essential for frontend to use correct values.",
		"parameters": map[string]interface{}{
			"category": "string (optional) - Filter by category (transaction, payment, approval, datahistory, kafka, etc.)",
			"keyword":  "string (optional) - Search in enum name, label, or description",
		},
	},
	// Web Search Tool (ค้นหาข้อมูลจาก internet)
	{
		"name":        "websearch",
		"description": "Search the internet for information using DuckDuckGo. Use when you need external knowledge not available in the shop's data (e.g. tax rates, regulations, product info, market data).",
		"parameters": map[string]interface{}{
			"query": "string (required) - Search query (Thai or English)",
			"limit": "number (optional) - Max results (default: 5, max: 10)",
		},
	},
	// Unit of Measure Tools (หน่วยนับ)
	{
		"name":        "listunits",
		"description": "List/search units of measure (หน่วยนับ). Returns unit codes and names.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"keyword":     "string (optional) - Search by unit code or name",
			"limit":       "number (optional) - Max results (default: 50, max: 200)",
		},
	},
	{
		"name":        "createunit",
		"description": "Create a new unit of measure (หน่วยนับ). Uses names[] for multi-language display names.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"unitcode":    "string (required) - Unit code e.g. EA, BOX, KG",
			"names":       "string (required) - JSON array [{\"code\":\"th\",\"name\":\"ชิ้น\"},{\"code\":\"en\",\"name\":\"Each\"}]",
		},
	},
	{
		"name":        "createunits",
		"description": "Create multiple units of measure at once (bulk). Skips duplicates.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"units":       "string (required) - JSON array e.g. [{\"unitcode\":\"EA\",\"names\":[{\"code\":\"th\",\"name\":\"ชิ้น\"}]},{\"unitcode\":\"BOX\",\"names\":[{\"code\":\"th\",\"name\":\"กล่อง\"}]}]",
		},
	},
	{
		"name":        "updateunit",
		"description": "Update an existing unit of measure by unit code.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"unitcode":    "string (required) - Unit code to update",
			"names":       "string (optional) - JSON array of language names [{\"code\":\"th\",\"name\":\"ชิ้น\"}]",
		},
	},
	{
		"name":        "deleteunit",
		"description": "Delete a unit of measure by unit code.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"unitcode":    "string (required) - Unit code to delete",
		},
	},
	{
		"name":        "deleteunits",
		"description": "Delete multiple units of measure at once (bulk). Reports which were deleted and which were not found.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"unitcodes":   "string (required) - JSON array of unit codes e.g. [\"EA\",\"BOX\",\"KG\"]. Max 100 items.",
		},
	},
	{
		"name":        "getunitschema",
		"description": "Get the data structure/schema of unit of measure (หน่วยนับ) documents.",
		"parameters":  map[string]interface{}{},
	},
	// Product Barcode Tools (สินค้า/บาร์โค้ด)
	{
		"name":        "listbarcodes",
		"description": "List/search product barcodes (สินค้า/บาร์โค้ด). Search by barcode, item code, or product name.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"keyword":     "string (optional) - Search by barcode, item code, or product name",
			"limit":       "number (optional) - Max results (default: 50, max: 200)",
		},
	},
	{
		"name":        "createbarcode",
		"description": "Create a new product barcode (สินค้า/บาร์โค้ด). Uses names[] for multi-language product names.",
		"parameters": map[string]interface{}{
			"holdingcode":   "string (required) - Holding Code",
			"barcode":       "string (required) - Barcode e.g. 8859100001234",
			"itemcode":      "string (required) - Item/product code e.g. SKU001",
			"names":         "string (required) - JSON array [{\"code\":\"th\",\"name\":\"สินค้า A\"},{\"code\":\"en\",\"name\":\"Product A\"}]",
			"itemunitcode":  "string (optional) - Unit code e.g. EA, BOX",
			"itemunitnames": "string (optional) - JSON array of unit names [{\"code\":\"th\",\"name\":\"ชิ้น\"}]",
			"prices":        "string (optional) - JSON array of prices [{\"keynumber\":1,\"price\":100.00}]",
			"standvalue":    "number (optional) - Unit conversion numerator (default: 1). e.g. BOX=24 means 1 BOX = 24 base units",
			"dividevalue":   "number (optional) - Unit conversion denominator (default: 1)",
			"ismainbarcode": "boolean (optional) - Is main barcode? Auto-detected if not provided: true when standvalue=1 & dividevalue=1",
			"groupcode":     "string (optional) - Product group code (กลุ่มสินค้า)",
			"groupnames":    "string (optional) - JSON array of group names [{\"code\":\"th\",\"name\":\"อาหาร\"}]",
			"categorycode":  "string (optional) - Product category guidfixed (หมวดสินค้า)",
			"categorynames": "string (optional) - JSON array of category names [{\"code\":\"th\",\"name\":\"เนื้อสัตว์\"}]",
		},
	},
	{
		"name":        "createbarcodes",
		"description": "Create multiple product barcodes at once (bulk). Skips duplicates.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"barcodes":    "string (required) - JSON array e.g. [{\"barcode\":\"123\",\"itemcode\":\"SKU1\",\"names\":[{\"code\":\"th\",\"name\":\"สินค้า\"}]}]",
		},
	},
	{
		"name":        "updatebarcode",
		"description": "Update an existing product barcode by guidfixed.",
		"parameters": map[string]interface{}{
			"holdingcode":   "string (required) - Holding Code",
			"guidfixed":     "string (required) - GuidFixed of the barcode to update",
			"names":         "string (optional) - JSON array of language names [{\"code\":\"th\",\"name\":\"ชื่อใหม่\"}]",
			"itemunitcode":  "string (optional) - New unit code",
			"itemunitnames": "string (optional) - JSON array of unit names",
			"prices":        "string (optional) - JSON array of prices [{\"keynumber\":1,\"price\":150.00}]",
			"groupcode":     "string (optional) - Product group code (กลุ่มสินค้า)",
			"groupnames":    "string (optional) - JSON array of group names [{\"code\":\"th\",\"name\":\"อาหาร\"}]",
			"categorycode":  "string (optional) - Product category guidfixed (หมวดสินค้า)",
			"categorynames": "string (optional) - JSON array of category names [{\"code\":\"th\",\"name\":\"เนื้อสัตว์\"}]",
		},
	},
	{
		"name":        "deletebarcode",
		"description": "Delete a product barcode by guidfixed.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"guidfixed":   "string (required) - GuidFixed of the barcode to delete",
		},
	},
	{
		"name":        "deletebarcodes",
		"description": "Delete multiple product barcodes at once (bulk). Reports which were deleted and which were not found.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"guidfixeds":  "string (required) - JSON array of guidfixed values e.g. [\"guid1\",\"guid2\"]. Max 100 items.",
		},
	},
	{
		"name":        "getbarcodeschema",
		"description": "Get the data structure/schema of product barcode (สินค้า/บาร์โค้ด) documents.",
		"parameters":  map[string]interface{}{},
	},
	// Reference Barcodes + Multi-Unit
	{
		"name":        "getrefbarcodes",
		"description": "Get reference barcodes and unit chain for a product. Shows all units (e.g., ชิ้น→ลัง) and how they reference each other. Returns product_name and unitname as formatted strings (auto from names[]).",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"itemcode":    "string (optional) - Item code to get all barcodes for (if not specified, use barcode to find)",
			"barcode":     "string (optional) - Barcode to find item code from (one of itemcode/barcode required)",
		},
	},
	{
		"name":        "setrefbarcode",
		"description": "Set reference barcode for a barcode (unit conversion). E.g., BOX → EA means 1 BOX = 24 EA. Checks for circular references.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"barcode":     "string (required) - Source barcode (e.g., BOX barcode)",
			"refbarcode":  "string (required) - Target reference barcode (e.g., EA barcode). Must be same itemcode.",
			"qty":         "number (optional) - Quantity for condition-based reference",
			"standvalue":  "number (optional) - Override stand value (default: use existing)",
			"dividevalue": "number (optional) - Override divide value (default: use existing)",
			"condition":   "boolean (optional) - Is conditional reference (default: false)",
		},
	},
	{
		"name":        "createmultiunitbarcode",
		"description": "Create a product with multiple units at once (e.g., ชิ้น + ลัง + แพ็ค). Automatically sets reference barcodes. The unit with standvalue=1 becomes the base unit.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"itemcode":    "string (required) - Item code for all units",
			"names":       "string (optional) - Default JSON names array (used when unit doesn't have its own names)",
			"units":       "string (required) - JSON array of units e.g. [{\"barcode\":\"EA-001\",\"itemunitcode\":\"EA\",\"standvalue\":1,\"dividevalue\":1,...},{\"barcode\":\"BOX-001\",\"itemunitcode\":\"BOX\",\"standvalue\":24,\"dividevalue\":1,...}]",
		},
	},
	// Rebuild Products (Full Sync — กรณี Kafka sync ผิดพลาด)
	{
		"name":        "rebuildproducts",
		"description": "Full rebuild: sync ALL product barcodes from MongoDB → PostgreSQL + ClickHouse. Use when Kafka sync fails or data is out of sync. Same as frontend 'สร้างสินค้าใหม่' button.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code to rebuild products for",
		},
	},
	// Rebuild Embeddings (สร้าง vector embeddings สำหรับ semantic search)
	{
		"name":        "rebuildembeddings",
		"description": "Build vector embeddings for semantic search on PostgreSQL projection tables with Ollama -> pgvector. Supports product, debtor, creditor, customer. Use entitytype=all to rebuild all.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"entitytype":  "string (optional) - product|debtor|creditor|customer|all (default: product)",
			"forceall":    "boolean (optional) - true=สร้างใหม่ทั้งหมด, false=เฉพาะที่ยังไม่มี (default: false)",
		},
	},
	// Product Group Tools (กลุ่มสินค้า)
	{
		"name":        "listproductgroups",
		"description": "List/search product groups (กลุ่มสินค้า). Returns group codes and names with multi-language support.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"keyword":     "string (optional) - Search by group code or name (e.g., 'อาหาร', 'FOOD')",
			"limit":       "number (optional) - Max results to return (default: 50, max: 200)",
		},
	},
	{
		"name":        "createproductgroup",
		"description": "Create a new product group (กลุ่มสินค้า). Uses names[] for multi-language display names.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"code":        "string (required) - Group code (e.g., 'FOOD', 'DRINK', 'TOOL')",
			"names":       "string (required) - JSON array [{\"code\":\"th\",\"name\":\"อาหาร\"},{\"code\":\"en\",\"name\":\"Food\"}]",
		},
	},
	{
		"name":        "createproductgroups",
		"description": "Create multiple product groups at once (bulk). Skips duplicates automatically.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"groups":      "string (required) - JSON array e.g. [{\"code\":\"FOOD\",\"names\":[{\"code\":\"th\",\"name\":\"อาหาร\"}]},{\"code\":\"DRINK\",\"names\":[{\"code\":\"th\",\"name\":\"เครื่องดื่ม\"}]}]. Max 100 items.",
		},
	},
	{
		"name":        "updateproductgroup",
		"description": "Update an existing product group by code.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"code":        "string (required) - Group code to update",
			"names":       "string (optional) - JSON array of multi-language names [{\"code\":\"th\",\"name\":\"อาหาร\"}]",
		},
	},
	{
		"name":        "deleteproductgroup",
		"description": "Delete a product group by code.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"code":        "string (required) - Group code to delete",
		},
	},
	{
		"name":        "deleteproductgroups",
		"description": "Delete multiple product groups at once (bulk). Reports which were deleted and which were not found.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"codes":       "string (required) - JSON array of group codes e.g. [\"FOOD\",\"DRINK\",\"TOOL\"]. Max 100 items.",
		},
	},
	{
		"name":        "getproductgroupschema",
		"description": "Get the data structure/schema of product group (กลุ่มสินค้า) documents with examples.",
		"parameters":  map[string]interface{}{},
	},
	// Product Category Tools (หมวดสินค้า)
	{
		"name":        "listproductcategories",
		"description": "List/search product categories (หมวดสินค้า). Returns category names, hierarchy, and group numbers.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"keyword":     "string (optional) - Search by category name (e.g., 'เนื้อสัตว์', 'ผัก')",
			"limit":       "number (optional) - Max results to return (default: 50, max: 200)",
		},
	},
	{
		"name":        "createproductcategory",
		"description": "Create a new product category (หมวดสินค้า). Uses names[] for multi-language display names. Supports hierarchy via parentguid.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"names":       "string (required) - JSON array [{\"code\":\"th\",\"name\":\"เนื้อสัตว์\"},{\"code\":\"en\",\"name\":\"Meat\"}]",
			"parentguid":  "string (optional) - Parent category guidfixed (for hierarchy)",
			"groupnumber": "number (optional) - Group/sort number",
		},
	},
	{
		"name":        "createproductcategories",
		"description": "Create multiple product categories at once (bulk). Auto-generates guidfixed for each.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"categories":  "string (required) - JSON array e.g. [{\"names\":[{\"code\":\"th\",\"name\":\"เนื้อสัตว์\"}],\"groupnumber\":1},{\"names\":[{\"code\":\"th\",\"name\":\"ผัก\"}],\"groupnumber\":2}]. Max 100 items.",
		},
	},
	{
		"name":        "updateproductcategory",
		"description": "Update an existing product category by guidfixed.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"guidfixed":   "string (required) - GuidFixed of the category to update",
			"names":       "string (optional) - JSON array of multi-language names [{\"code\":\"th\",\"name\":\"เนื้อสัตว์\"}]",
			"parentguid":  "string (optional) - New parent category guidfixed",
			"groupnumber": "number (optional) - New group/sort number",
		},
	},
	{
		"name":        "deleteproductcategory",
		"description": "Delete a product category by guidfixed.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"guidfixed":   "string (required) - GuidFixed of the category to delete",
		},
	},
	{
		"name":        "deleteproductcategories",
		"description": "Delete multiple product categories at once (bulk). Reports which were deleted and which were not found.",
		"parameters": map[string]interface{}{
			"holdingcode": "string (required) - Holding Code",
			"guidfixeds":  "string (required) - JSON array of guidfixed values e.g. [\"guid1\",\"guid2\"]. Max 100 items.",
		},
	},
	{
		"name":        "getproductcategoryschema",
		"description": "Get the data structure/schema of product category (หมวดสินค้า) documents with examples.",
		"parameters":  map[string]interface{}{},
	},
	// Creditor Tools (เจ้าหนี้)
	{
		"name":        "listcreditors",
		"description": "List/search creditors (เจ้าหนี้). Returns creditor codes, names, tax ID, and contact info.",
		"parameters": map[string]interface{}{
			"keyword": "string (optional) — Search by creditor code, name, or tax ID",
			"limit":   "number (optional) — Max results (default: 50, max: 200)",
		},
	},
	{
		"name":        "createcreditor",
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
		"name":        "createcreditors",
		"description": "Create multiple creditors at once (bulk). Skips duplicates automatically. Max 100 items.",
		"parameters": map[string]interface{}{
			"creditors": "string (required) — JSON array of creditors. Each needs code (required) and names (required).",
		},
	},
	{
		"name":        "updatecreditor",
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
		"name":        "deletecreditor",
		"description": "Delete a creditor by guidfixed (soft delete).",
		"parameters": map[string]interface{}{
			"guidfixed": "string (required) — GuidFixed of the creditor to delete",
		},
	},
	{
		"name":        "deletecreditors",
		"description": "Delete multiple creditors at once (bulk soft delete). Max 100 items.",
		"parameters": map[string]interface{}{
			"guidfixeds": "string (required) — JSON array of guidfixed values to delete",
		},
	},
	{
		"name":        "getcreditorschema",
		"description": "Get the data structure/schema of creditor (เจ้าหนี้) documents with examples.",
		"parameters":  map[string]interface{}{},
	},
	// Debtor Tools (ลูกหนี้)
	{
		"name":        "listdebtors",
		"description": "List/search debtors (ลูกหนี้). Returns debtor codes, names, tax ID, and contact info.",
		"parameters": map[string]interface{}{
			"keyword": "string (optional) — Search by debtor code, name, or tax ID",
			"limit":   "number (optional) — Max results (default: 50, max: 200)",
		},
	},
	{
		"name":        "createdebtor",
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
		"name":        "createdebtors",
		"description": "Create multiple debtors at once (bulk). Skips duplicates automatically. Max 100 items.",
		"parameters": map[string]interface{}{
			"debtors": "string (required) — JSON array of debtors. Each needs code (required) and names (required).",
		},
	},
	{
		"name":        "updatedebtor",
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
		"name":        "deletedebtor",
		"description": "Delete a debtor by guidfixed (soft delete).",
		"parameters": map[string]interface{}{
			"guidfixed": "string (required) — GuidFixed of the debtor to delete",
		},
	},
	{
		"name":        "deletedebtors",
		"description": "Delete multiple debtors at once (bulk soft delete). Max 100 items.",
		"parameters": map[string]interface{}{
			"guidfixeds": "string (required) — JSON array of guidfixed values to delete",
		},
	},
	{
		"name":        "getdebtorschema",
		"description": "Get the data structure/schema of debtor (ลูกหนี้) documents with examples.",
		"parameters":  map[string]interface{}{},
	},
	// Purchase Order Tools (ใบสั่งซื้อ)
	{
		"name":        "listpurchaseorders",
		"description": "List/search purchase orders (ใบสั่งซื้อ). Returns docno, creditor, amounts, status.",
		"parameters": map[string]interface{}{
			"keyword": "string (optional) — Search by docno, creditor code/name, or description",
			"limit":   "number (optional) — Max results (default: 50, max: 200)",
		},
	},
	{
		"name":        "createpurchaseorder",
		"description": "Create a new purchase order (ใบสั่งซื้อ). Requires docno.",
		"parameters": map[string]interface{}{
			"docno":       "string (required) — Document number e.g. 'PO-2026-0001'",
			"custcode":    "string (optional) — Creditor code",
			"custnames":   "string (optional) — JSON array of creditor names e.g. [{\"code\":\"th\",\"name\":\"บริษัท ABC\"}]",
			"details":     "string (optional) — JSON array of line items [{\"barcode\":\"123\",\"itemcode\":\"SKU1\",\"qty\":10,\"price\":100,\"sumamount\":1000}]",
			"description": "string (optional) — Description/remark",
			"transflag":   "number (optional) — Transaction flag (default: 0)",
			"vattype":     "number (optional) — VAT type (0=none, 1=inclusive, 2=exclusive)",
			"vatrate":     "number (optional) — VAT rate %",
			"totalamount": "number (optional) — Total amount",
		},
	},
	{
		"name":        "updatepurchaseorder",
		"description": "Update an existing purchase order by guidfixed.",
		"parameters": map[string]interface{}{
			"guidfixed":   "string (required) — GuidFixed of the purchase order to update",
			"custnames":   "string (optional) — JSON array of creditor names",
			"details":     "string (optional) — JSON array of line items",
			"description": "string (optional) — New description",
			"totalamount": "number (optional) — New total amount",
			"status":      "number (optional) — New status (0=draft, 1=pending, 2=approved, 3=rejected)",
		},
	},
	{
		"name":        "deletepurchaseorder",
		"description": "Delete a purchase order by guidfixed (soft delete).",
		"parameters": map[string]interface{}{
			"guidfixed": "string (required) — GuidFixed of the purchase order to delete",
		},
	},
	{
		"name":        "getpurchaseorderschema",
		"description": "Get the data structure/schema of purchase order (ใบสั่งซื้อ) documents with examples.",
		"parameters":  map[string]interface{}{},
	},
	// Purchase Requisition Tools (ใบขอซื้อ PR)
	{
		"name":        "listpurchaserequisitions",
		"description": "List/search purchase requisitions (ใบขอซื้อ PR). Returns docno, requester, department, urgency, status.",
		"parameters": map[string]interface{}{
			"keyword": "string (optional) — Search by docno, requester code/name, department, purpose",
			"limit":   "number (optional) — Max results (default: 50, max: 200)",
		},
	},
	{
		"name":        "createpurchaserequisition",
		"description": "Create a new purchase requisition (ใบขอซื้อ PR). AI ใช้สร้าง PR อัตโนมัติได้. Requires docno, requestercode, departmentcode, purpose.",
		"parameters": map[string]interface{}{
			"docno":                 "string (required) — Document number e.g. 'PR20260314-00001'",
			"requestercode":         "string (required) — Employee code of requester",
			"requestername":         "string (optional) — Name of requester",
			"departmentcode":        "string (required) — Department code e.g. 'IT', 'ACC'",
			"departmentnames":       "string (optional) — JSON array [{\"code\":\"th\",\"name\":\"ฝ่ายไอที\"}]",
			"purpose":               "string (required) — Purpose/reason for purchase",
			"budgetcode":            "string (optional) — Budget code",
			"budgetamount":          "number (optional) — Budget amount",
			"urgency":               "number (optional) — 1=Normal, 2=Urgent, 3=Critical (default: 1)",
			"requesteddeliverydate": "string (optional) — Requested delivery date (YYYY-MM-DD)",
			"details":               "string (optional) — JSON array of items [{\"barcode\":\"123\",\"itemcode\":\"SKU1\",\"qty\":10,\"price\":100,\"sumamount\":1000}]",
			"description":           "string (optional) — Description/remark",
			"totalamount":           "number (optional) — Total amount",
		},
	},
	{
		"name":        "updatepurchaserequisition",
		"description": "Update an existing purchase requisition by guidfixed.",
		"parameters": map[string]interface{}{
			"guidfixed":             "string (required) — GuidFixed of the PR to update",
			"requestercode":         "string (optional) — New requester code",
			"requestername":         "string (optional) — New requester name",
			"departmentcode":        "string (optional) — New department code",
			"departmentnames":       "string (optional) — JSON array of department names",
			"purpose":               "string (optional) — New purpose",
			"budgetcode":            "string (optional) — New budget code",
			"budgetamount":          "number (optional) — New budget amount",
			"urgency":               "number (optional) — 1=Normal, 2=Urgent, 3=Critical",
			"requesteddeliverydate": "string (optional) — New delivery date",
			"details":               "string (optional) — JSON array of items",
			"description":           "string (optional) — New description",
			"totalamount":           "number (optional) — New total amount",
			"status":                "number (optional) — New status (0=draft, 1=pending, 2=approved, 3=rejected)",
			"conversionstatus":      "string (optional) — none/converted_to_rfq/converted_to_po",
		},
	},
	{
		"name":        "deletepurchaserequisition",
		"description": "Delete a purchase requisition by guidfixed (soft delete).",
		"parameters": map[string]interface{}{
			"guidfixed": "string (required) — GuidFixed of the PR to delete",
		},
	},
	{
		"name":        "getpurchaserequisitionschema",
		"description": "Get the data structure/schema of purchase requisition (ใบขอซื้อ PR) documents with examples.",
		"parameters":  map[string]interface{}{},
	},
	// Model Schema Tool
	{
		"name":        "getmodelschema",
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
	case "searchproducts":
		return s.invokeSearchProducts(ctx, params)
	case "getdailysales":
		return s.invokeGetDailySales(ctx, params)
	case "getsalesbydaterange":
		return s.invokeGetSalesByDateRange(ctx, params)
	case "gettopsellingproducts":
		return s.invokeGetTopSellingProducts(ctx, params)
	case "getsalesbyseller":
		return s.invokeGetSalesBySeller(ctx, params)
	case "getmonthlysummary":
		return s.invokeGetMonthlySummary(ctx, params)
	case "getdashboardkpis":
		return s.invokeGetDashboardKPIs(ctx, params)
	case "getbusinesshealth":
		return s.invokeGetBusinessHealth(ctx, params)
	case "getprofitanalysis":
		return s.invokeGetProfitAnalysis(ctx, params)
	case "getaccountsreceivable":
		return s.invokeGetAccountsReceivable(ctx, params)
	case "getaccountspayable":
		return s.invokeGetAccountsPayable(ctx, params)
	case "getcashflow":
		return s.invokeGetCashFlow(ctx, params)
	case "getinventoryvalue":
		return s.invokeGetInventoryValue(ctx, params)
	case "getlowstockalerts":
		return s.invokeGetLowStockAlerts(ctx, params)
	case "getdeadstock":
		return s.invokeGetDeadStock(ctx, params)
	case "getinventoryturnover":
		return s.invokeGetInventoryTurnover(ctx, params)
	case "gettopcustomers":
		return s.invokeGetTopCustomers(ctx, params)
	case "getcustomergrowth":
		return s.invokeGetCustomerGrowth(ctx, params)
	case "getcustomersegments":
		return s.invokeGetCustomerSegments(ctx, params)
	case "getyoycomparison":
		return s.invokeGetYoYComparison(ctx, params)
	case "getmomcomparison":
		return s.invokeGetMoMComparison(ctx, params)
	case "listunits":
		return s.invokeListUnits(ctx, params)
	case "websearch":
		return s.invokeWebSearch(ctx, params)
	case "querymongodb":
		return s.invokeQueryMongoDB(ctx, params)
	case "listmongodbcollections":
		return s.invokeListMongoDBCollections(ctx, params)
	case "aggregatemongodb":
		return s.invokeAggregateMongoDB(ctx, params)
	case "queryclickhouse":
		return s.invokeQueryClickHouse(ctx, params)
	case "listclickhousetables":
		return s.invokeListClickHouseTables(ctx, params)
	case "querypostgresql":
		return s.invokeExecutePgCommand(ctx, params)
	case "executejs":
		return s.invokeExecuteJS(ctx, params)
	case "executepython":
		return s.invokeExecutePython(ctx, params)
	// Entity semantic search tools
	case "searchdebtors":
		return s.invokeSearchDebtors(ctx, params)
	case "searchcreditors":
		return s.invokeSearchCreditors(ctx, params)
	case "searchcustomers":
		return s.invokeSearchCustomers(ctx, params)
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

	// SECURITY (2026-06-21): the tenant is ALWAYS the API key's tenant. Never honor a
	// caller-supplied holdingcode — it allowed cross-tenant access (and raw-SQL via the
	// dev tools). Force it unconditionally, matching the SSE handler (sse_handler.go).
	if req.Params == nil {
		req.Params = make(map[string]interface{})
	}
	req.Params["holdingcode"] = apiKey.HoldingCode

	// Execute tool
	ctx := c.Request().Context()
	startTime := time.Now()

	var result interface{}
	var err error

	switch req.Tool {
	// Product Search
	case "searchproducts":
		result, err = s.invokeSearchProducts(ctx, req.Params)
	// Sales Tools
	case "getdailysales":
		result, err = s.invokeGetDailySales(ctx, req.Params)
	case "getsalesbydaterange":
		result, err = s.invokeGetSalesByDateRange(ctx, req.Params)
	case "gettopsellingproducts":
		result, err = s.invokeGetTopSellingProducts(ctx, req.Params)
	case "getsalesbyseller":
		result, err = s.invokeGetSalesBySeller(ctx, req.Params)
	case "getmonthlysummary":
		result, err = s.invokeGetMonthlySummary(ctx, req.Params)
	// Dashboard Tools
	case "getdashboardkpis":
		result, err = s.invokeGetDashboardKPIs(ctx, req.Params)
	case "getbusinesshealth":
		result, err = s.invokeGetBusinessHealth(ctx, req.Params)
	// Financial Tools
	case "getprofitanalysis":
		result, err = s.invokeGetProfitAnalysis(ctx, req.Params)
	case "getaccountsreceivable":
		result, err = s.invokeGetAccountsReceivable(ctx, req.Params)
	case "getaccountspayable":
		result, err = s.invokeGetAccountsPayable(ctx, req.Params)
	case "getcashflow":
		result, err = s.invokeGetCashFlow(ctx, req.Params)
	// Inventory Tools
	case "getinventoryvalue":
		result, err = s.invokeGetInventoryValue(ctx, req.Params)
	case "getlowstockalerts":
		result, err = s.invokeGetLowStockAlerts(ctx, req.Params)
	case "getdeadstock":
		result, err = s.invokeGetDeadStock(ctx, req.Params)
	case "getinventoryturnover":
		result, err = s.invokeGetInventoryTurnover(ctx, req.Params)
	// Customer Tools
	case "gettopcustomers":
		result, err = s.invokeGetTopCustomers(ctx, req.Params)
	case "getcustomergrowth":
		result, err = s.invokeGetCustomerGrowth(ctx, req.Params)
	case "getcustomersegments":
		result, err = s.invokeGetCustomerSegments(ctx, req.Params)
	// Comparison Tools
	case "getyoycomparison":
		result, err = s.invokeGetYoYComparison(ctx, req.Params)
	case "getmomcomparison":
		result, err = s.invokeGetMoMComparison(ctx, req.Params)
	// Database Tools
	case "getdatabaseschema":
		result, err = s.invokeGetDatabaseSchema(ctx, req.Params)
	case "executequery":
		result, err = s.invokeExecuteQuery(ctx, req.Params)
	case "gettablesample":
		result, err = s.invokeGetTableSample(ctx, req.Params)
	// MongoDB Tools
	case "querymongodb":
		result, err = s.invokeQueryMongoDB(ctx, req.Params)
	case "listmongodbcollections":
		result, err = s.invokeListMongoDBCollections(ctx, req.Params)
	case "aggregatemongodb":
		result, err = s.invokeAggregateMongoDB(ctx, req.Params)
	// ClickHouse Tools
	case "queryclickhouse":
		result, err = s.invokeQueryClickHouse(ctx, req.Params)
	case "listclickhousetables":
		result, err = s.invokeListClickHouseTables(ctx, req.Params)
	// Dev Database Tools
	case "executepgcommand":
		result, err = s.invokeExecutePgCommand(ctx, req.Params)
	case "executechcommand":
		result, err = s.invokeExecuteChCommand(ctx, req.Params)
	// Web Search
	case "websearch":
		result, err = s.invokeWebSearch(ctx, req.Params)
	// Unit of Measure Tools (หน่วยนับ)
	case "listunits":
		result, err = s.invokeListUnits(ctx, req.Params)
	case "createunit":
		result, err = s.invokeCreateUnit(ctx, req.Params)
	case "createunits":
		result, err = s.invokeCreateUnits(ctx, req.Params)
	case "updateunit":
		result, err = s.invokeUpdateUnit(ctx, req.Params)
	case "deleteunit":
		result, err = s.invokeDeleteUnit(ctx, req.Params)
	case "deleteunits":
		result, err = s.invokeDeleteUnits(ctx, req.Params)
	case "getunitschema":
		result, err = s.invokeGetUnitSchema(ctx, req.Params)
	// Product Barcode Tools (สินค้า/บาร์โค้ด)
	case "listbarcodes":
		result, err = s.invokeListBarcodes(ctx, req.Params)
	case "createbarcode":
		result, err = s.invokeCreateBarcode(ctx, req.Params)
	case "createbarcodes":
		result, err = s.invokeCreateBarcodes(ctx, req.Params)
	case "updatebarcode":
		result, err = s.invokeUpdateBarcode(ctx, req.Params)
	case "deletebarcode":
		result, err = s.invokeDeleteBarcode(ctx, req.Params)
	case "deletebarcodes":
		result, err = s.invokeDeleteBarcodes(ctx, req.Params)
	case "getbarcodeschema":
		result, err = s.invokeGetBarcodeSchema(ctx, req.Params)
	// Reference Barcodes + Multi-Unit
	case "getrefbarcodes":
		result, err = s.invokeGetRefBarcodes(ctx, req.Params)
	case "setrefbarcode":
		result, err = s.invokeSetRefBarcode(ctx, req.Params)
	case "createmultiunitbarcode":
		result, err = s.invokeCreateMultiUnitBarcode(ctx, req.Params)
	case "rebuildproducts":
		result, err = s.invokeRebuildProducts(ctx, req.Params)
	case "rebuildembeddings":
		result, err = s.invokeRebuildEmbeddings(ctx, req.Params)
	// Product Group Tools (กลุ่มสินค้า)
	case "listproductgroups":
		result, err = s.invokeListProductGroups(ctx, req.Params)
	case "createproductgroup":
		result, err = s.invokeCreateProductGroup(ctx, req.Params)
	case "createproductgroups":
		result, err = s.invokeCreateProductGroups(ctx, req.Params)
	case "updateproductgroup":
		result, err = s.invokeUpdateProductGroup(ctx, req.Params)
	case "deleteproductgroup":
		result, err = s.invokeDeleteProductGroup(ctx, req.Params)
	case "deleteproductgroups":
		result, err = s.invokeDeleteProductGroups(ctx, req.Params)
	case "getproductgroupschema":
		result, err = s.invokeGetProductGroupSchema(ctx, req.Params)
	// Product Category Tools (หมวดสินค้า)
	case "listproductcategories":
		result, err = s.invokeListProductCategories(ctx, req.Params)
	case "createproductcategory":
		result, err = s.invokeCreateProductCategory(ctx, req.Params)
	case "createproductcategories":
		result, err = s.invokeCreateProductCategories(ctx, req.Params)
	case "updateproductcategory":
		result, err = s.invokeUpdateProductCategory(ctx, req.Params)
	case "deleteproductcategory":
		result, err = s.invokeDeleteProductCategory(ctx, req.Params)
	case "deleteproductcategories":
		result, err = s.invokeDeleteProductCategories(ctx, req.Params)
	case "getproductcategoryschema":
		result, err = s.invokeGetProductCategorySchema(ctx, req.Params)
	// Creditor Tools (เจ้าหนี้)
	case "listcreditors":
		result, err = s.invokeListCreditors(ctx, req.Params)
	case "createcreditor":
		result, err = s.invokeCreateCreditor(ctx, req.Params)
	case "createcreditors":
		result, err = s.invokeCreateCreditors(ctx, req.Params)
	case "updatecreditor":
		result, err = s.invokeUpdateCreditor(ctx, req.Params)
	case "deletecreditor":
		result, err = s.invokeDeleteCreditor(ctx, req.Params)
	case "deletecreditors":
		result, err = s.invokeDeleteCreditors(ctx, req.Params)
	case "getcreditorschema":
		result, err = s.invokeGetCreditorSchema(ctx, req.Params)
	// Debtor Tools (ลูกหนี้)
	case "listdebtors":
		result, err = s.invokeListDebtors(ctx, req.Params)
	case "createdebtor":
		result, err = s.invokeCreateDebtor(ctx, req.Params)
	case "createdebtors":
		result, err = s.invokeCreateDebtors(ctx, req.Params)
	case "updatedebtor":
		result, err = s.invokeUpdateDebtor(ctx, req.Params)
	case "deletedebtor":
		result, err = s.invokeDeleteDebtor(ctx, req.Params)
	case "deletedebtors":
		result, err = s.invokeDeleteDebtors(ctx, req.Params)
	case "getdebtorschema":
		result, err = s.invokeGetDebtorSchema(ctx, req.Params)
	// Purchase Order Tools (ใบสั่งซื้อ)
	case "listpurchaseorders":
		result, err = s.invokeListPurchaseOrders(ctx, req.Params)
	case "createpurchaseorder":
		result, err = s.invokeCreatePurchaseOrder(ctx, req.Params)
	case "updatepurchaseorder":
		result, err = s.invokeUpdatePurchaseOrder(ctx, req.Params)
	case "deletepurchaseorder":
		result, err = s.invokeDeletePurchaseOrder(ctx, req.Params)
	case "getpurchaseorderschema":
		result, err = s.invokeGetPurchaseOrderSchema(ctx, req.Params)
	// Purchase Requisition Tools (ใบขอซื้อ PR)
	case "listpurchaserequisitions":
		result, err = s.invokeListPurchaseRequisitions(ctx, req.Params)
	case "createpurchaserequisition":
		result, err = s.invokeCreatePurchaseRequisition(ctx, req.Params)
	case "updatepurchaserequisition":
		result, err = s.invokeUpdatePurchaseRequisition(ctx, req.Params)
	case "deletepurchaserequisition":
		result, err = s.invokeDeletePurchaseRequisition(ctx, req.Params)
	case "getpurchaserequisitionschema":
		result, err = s.invokeGetPurchaseRequisitionSchema(ctx, req.Params)
	// API Catalog & Frontend Dev Tools
	case "listapiendpoints":
		result, err = s.invokeListAPIEndpoints(ctx, req.Params)
	case "getapispec":
		result, err = s.invokeGetAPISpec(ctx, req.Params)
	case "getapiexample":
		result, err = s.invokeGetAPIExample(ctx, req.Params)
	case "listenums":
		result, err = s.invokeListEnums(ctx, req.Params)
	case "getmodelschema":
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

	// Get API key data and set holdingcode
	apiKey, _ := auth.GetAPIKeyFromContext(c)
	if params.HoldingCode == "" {
		params.HoldingCode = apiKey.HoldingCode
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
	if params.HoldingCode == "" {
		params.HoldingCode = apiKey.HoldingCode
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
	if params.HoldingCode == "" {
		params.HoldingCode = apiKey.HoldingCode
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
	if params.HoldingCode == "" {
		params.HoldingCode = apiKey.HoldingCode
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
	if params.HoldingCode == "" {
		params.HoldingCode = apiKey.HoldingCode
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
	req.HoldingCode = getStringParam(params, "holdingcode")
	req.Date = getStringParam(params, "date")
	req.BranchCode = getStringParam(params, "branchcode")
	return s.sales.GetDailySales(ctx, req)
}

// invokeGetSalesByDateRange invokes the sales by date range tool
func (s *MCPServer) invokeGetSalesByDateRange(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var req tools.SalesByDateRangeRequest
	req.HoldingCode = getStringParam(params, "holdingcode")
	req.FromDate = getStringParam(params, "fromdate")
	req.ToDate = getStringParam(params, "todate")
	req.BranchCode = getStringParam(params, "branchcode")
	req.GroupBy = getStringParam(params, "groupby")
	return s.sales.GetSalesByDateRange(ctx, req)
}

// invokeGetTopSellingProducts invokes the top selling products tool
func (s *MCPServer) invokeGetTopSellingProducts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var req tools.TopSellingProductsRequest
	req.HoldingCode = getStringParam(params, "holdingcode")
	req.FromDate = getStringParam(params, "fromdate")
	req.ToDate = getStringParam(params, "todate")
	req.Limit = getIntParam(params, "limit")
	req.BranchCode = getStringParam(params, "branchcode")
	return s.sales.GetTopSellingProducts(ctx, req)
}

// invokeGetSalesBySeller invokes the sales by seller tool
func (s *MCPServer) invokeGetSalesBySeller(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var req tools.SalesBySellerRequest
	req.HoldingCode = getStringParam(params, "holdingcode")
	req.FromDate = getStringParam(params, "fromdate")
	req.ToDate = getStringParam(params, "todate")
	return s.sales.GetSalesBySeller(ctx, req)
}

// invokeGetMonthlySummary invokes the monthly summary tool
func (s *MCPServer) invokeGetMonthlySummary(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var req tools.MonthlySummaryRequest
	req.HoldingCode = getStringParam(params, "holdingcode")
	req.Year = getIntParam(params, "year")
	req.Month = getIntParam(params, "month")
	return s.sales.GetMonthlySummary(ctx, req)
}

// invokeSearchProducts invokes the product search tool with Thai full-text search
func (s *MCPServer) invokeSearchProducts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	keyword := getStringParam(params, "keyword")
	whcode := getStringParam(params, "whcode")
	locationcode := getStringParam(params, "locationcode")
	limit := getIntParam(params, "limit")
	includeBalance := getBoolParam(params, "includebalance")

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
	if _, exists := params["includebalance"]; !exists {
		includeBalance = true
	}

	// Call the product search tool
	return tools.SearchProducts(ctx, holdingCode, keyword, whcode, locationcode, limit, includeBalance)
}

// ==================== Dashboard Tool Invocations ====================

// invokeGetDashboardKPIs invokes the dashboard KPIs tool
func (s *MCPServer) invokeGetDashboardKPIs(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	period := getStringParam(params, "period")
	if period == "" {
		period = "thismonth"
	}
	return tools.GetDashboardKPIs(ctx, holdingCode, period)
}

// invokeGetBusinessHealth invokes the business health tool
func (s *MCPServer) invokeGetBusinessHealth(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	return tools.GetBusinessHealth(ctx, holdingCode)
}

// ==================== Financial Tool Invocations ====================

// invokeGetProfitAnalysis invokes the profit analysis tool
func (s *MCPServer) invokeGetProfitAnalysis(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	fromDate := getStringParam(params, "fromdate")
	toDate := getStringParam(params, "todate")
	return tools.GetProfitAnalysis(ctx, holdingCode, fromDate, toDate)
}

// invokeGetAccountsReceivable invokes the accounts receivable tool
func (s *MCPServer) invokeGetAccountsReceivable(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	return tools.GetAccountsReceivable(ctx, holdingCode)
}

// invokeGetAccountsPayable invokes the accounts payable tool
func (s *MCPServer) invokeGetAccountsPayable(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	return tools.GetAccountsPayable(ctx, holdingCode)
}

// invokeGetCashFlow invokes the cash flow tool
func (s *MCPServer) invokeGetCashFlow(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	fromDate := getStringParam(params, "fromdate")
	toDate := getStringParam(params, "todate")
	return tools.GetCashFlow(ctx, holdingCode, fromDate, toDate)
}

// ==================== Inventory Tool Invocations ====================

// invokeGetInventoryValue invokes the inventory value tool
func (s *MCPServer) invokeGetInventoryValue(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	whcode := getStringParam(params, "whcode")
	return tools.GetInventoryValue(ctx, holdingCode, whcode)
}

// invokeGetLowStockAlerts invokes the low stock alerts tool
func (s *MCPServer) invokeGetLowStockAlerts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	threshold := getIntParam(params, "threshold")
	limit := getIntParam(params, "limit")
	if threshold <= 0 {
		threshold = 10
	}
	if limit <= 0 {
		limit = 50
	}
	return tools.GetLowStockAlerts(ctx, holdingCode, threshold, limit)
}

// invokeGetDeadStock invokes the dead stock tool
func (s *MCPServer) invokeGetDeadStock(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	daysNoMovement := getIntParam(params, "daysnomovement")
	limit := getIntParam(params, "limit")
	if daysNoMovement <= 0 {
		daysNoMovement = 90
	}
	if limit <= 0 {
		limit = 50
	}
	return tools.GetDeadStock(ctx, holdingCode, daysNoMovement, limit)
}

// invokeGetInventoryTurnover invokes the inventory turnover tool
func (s *MCPServer) invokeGetInventoryTurnover(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	fromDate := getStringParam(params, "fromdate")
	toDate := getStringParam(params, "todate")
	return tools.GetInventoryTurnover(ctx, holdingCode, fromDate, toDate)
}

// ==================== Customer Tool Invocations ====================

func (s *MCPServer) invokeSearchCustomers(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.SearchCustomers(ctx, holdingCode, keyword, limit)
}

// invokeGetTopCustomers invokes the top customers tool
func (s *MCPServer) invokeGetTopCustomers(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	fromDate := getStringParam(params, "fromdate")
	toDate := getStringParam(params, "todate")
	limit := getIntParam(params, "limit")
	sortBy := getStringParam(params, "sortby")
	if limit <= 0 {
		limit = 10
	}
	if sortBy == "" {
		sortBy = "amount"
	}
	return tools.GetTopCustomers(ctx, holdingCode, fromDate, toDate, limit, sortBy)
}

// invokeGetCustomerGrowth invokes the customer growth tool
func (s *MCPServer) invokeGetCustomerGrowth(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	fromDate := getStringParam(params, "fromdate")
	toDate := getStringParam(params, "todate")
	return tools.GetCustomerGrowth(ctx, holdingCode, fromDate, toDate)
}

// invokeGetCustomerSegments invokes the customer segments tool
func (s *MCPServer) invokeGetCustomerSegments(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	fromDate := getStringParam(params, "fromdate")
	toDate := getStringParam(params, "todate")
	return tools.GetCustomerSegments(ctx, holdingCode, fromDate, toDate)
}

// ==================== Comparison Tool Invocations ====================

// invokeGetYoYComparison invokes the year-over-year comparison tool
func (s *MCPServer) invokeGetYoYComparison(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	year := getIntParam(params, "year")
	month := getIntParam(params, "month")
	return tools.GetYoYComparison(ctx, holdingCode, year, month)
}

// invokeGetMoMComparison invokes the month-over-month comparison tool
func (s *MCPServer) invokeGetMoMComparison(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	year := getIntParam(params, "year")
	month := getIntParam(params, "month")
	return tools.GetMoMComparison(ctx, holdingCode, year, month)
}

// ==================== Database Tool Invocations ====================

// invokeGetDatabaseSchema invokes the database schema tool
func (s *MCPServer) invokeGetDatabaseSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	tableName := getStringParam(params, "tablename")
	return tools.GetDatabaseSchema(ctx, holdingCode, tableName)
}

// invokeExecuteQuery invokes the execute query tool (readonly)
func (s *MCPServer) invokeExecuteQuery(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	query := getStringParam(params, "query")
	limit := getIntParam(params, "limit")
	if limit <= 0 {
		limit = 100
	}
	return tools.ExecuteReadonlyQuery(ctx, holdingCode, query, limit)
}

// invokeGetTableSample invokes the table sample tool
func (s *MCPServer) invokeGetTableSample(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	tableName := getStringParam(params, "tablename")
	limit := getIntParam(params, "limit")
	if limit <= 0 {
		limit = 10
	}
	return tools.GetTableSample(ctx, holdingCode, tableName, limit)
}

// ==================== MongoDB Tool Invocations ====================

// invokeQueryMongoDB invokes the MongoDB query tool (readonly)
func (s *MCPServer) invokeQueryMongoDB(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	database := getStringParam(params, "database")
	collection := getStringParam(params, "collection")
	filter := getStringParam(params, "filter")
	limit := getIntParam(params, "limit")
	return tools.QueryMongoDB(ctx, holdingCode, database, collection, filter, limit)
}

// invokeListMongoDBCollections invokes the MongoDB list collections tool
func (s *MCPServer) invokeListMongoDBCollections(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	database := getStringParam(params, "database")
	return tools.ListMongoDBCollections(ctx, database)
}

// invokeAggregateMongoDB invokes the MongoDB aggregation tool (readonly)
func (s *MCPServer) invokeAggregateMongoDB(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	database := getStringParam(params, "database")
	collection := getStringParam(params, "collection")
	pipeline := getStringParam(params, "pipeline")
	limit := getIntParam(params, "limit")
	return tools.AggregateMongoDB(ctx, holdingCode, database, collection, pipeline, limit)
}

// ==================== ClickHouse Tool Invocations ====================

// invokeQueryClickHouse invokes the ClickHouse query tool (readonly)
func (s *MCPServer) invokeQueryClickHouse(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	database := getStringParam(params, "database")
	query := getStringParam(params, "query")
	limit := getIntParam(params, "limit")
	return tools.QueryClickHouse(ctx, holdingCode, database, query, limit)
}

// invokeListClickHouseTables invokes the ClickHouse list tables tool
func (s *MCPServer) invokeListClickHouseTables(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	database := getStringParam(params, "database")
	return tools.ListClickHouseTables(ctx, database)
}

// invokeExecutePgCommand invokes the dev PostgreSQL command tool
func (s *MCPServer) invokeExecutePgCommand(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	query := getStringParam(params, "query")
	limit := getIntParam(params, "limit")
	return tools.ExecutePgCommand(ctx, holdingCode, query, limit)
}

// invokeExecuteJS รัน JavaScript ใน Goja sandbox (readonly)
// AI เขียน JS เอง → รัน → ดูผล → แก้ → รันใหม่ จนได้คำตอบ
func (s *MCPServer) invokeExecuteJS(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	code := getStringParam(params, "code")
	return tools.ExecuteJS(ctx, holdingCode, code)
}

// invokeExecutePython รัน Python 3 script ใน subprocess sandbox (readonly)
// LLM เขียน Python เก่งที่สุด → ใช้เป็น primary tool สำหรับงาน data/query
func (s *MCPServer) invokeExecutePython(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	code := getStringParam(params, "code")
	return tools.ExecutePython(ctx, holdingCode, code)
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
		HoldingCode:     apiKey.HoldingCode,
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
	holdingCode := getStringParam(params, "holdingcode")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.ListUnits(ctx, holdingCode, keyword, limit)
}

// invokeCreateUnit invokes the create unit tool
func (s *MCPServer) invokeCreateUnit(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	unitCode := getStringParam(params, "unitcode")
	names := getStringParam(params, "names")
	return tools.CreateUnit(ctx, holdingCode, unitCode, names)
}

// invokeCreateUnits invokes the bulk create units tool
func (s *MCPServer) invokeCreateUnits(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	units := getStringParam(params, "units")
	return tools.CreateUnits(ctx, holdingCode, units)
}

// invokeUpdateUnit invokes the update unit tool
func (s *MCPServer) invokeUpdateUnit(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	unitCode := getStringParam(params, "unitcode")
	names := getStringParam(params, "names")
	return tools.UpdateUnit(ctx, holdingCode, unitCode, names)
}

// invokeDeleteUnit invokes the delete unit tool
func (s *MCPServer) invokeDeleteUnit(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	unitCode := getStringParam(params, "unitcode")
	return tools.DeleteUnit(ctx, holdingCode, unitCode)
}

// invokeDeleteUnits invokes the bulk delete units tool
func (s *MCPServer) invokeDeleteUnits(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	unitcodes := getStringParam(params, "unitcodes")
	return tools.DeleteUnits(ctx, holdingCode, unitcodes)
}

// invokeGetUnitSchema invokes the get unit schema tool
func (s *MCPServer) invokeGetUnitSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return tools.GetUnitSchema(), nil
}

// ==================== Web Search Tool Invocation ====================

// invokeWebSearch invokes the web search tool
func (s *MCPServer) invokeWebSearch(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query := getStringParam(params, "query")
	limit := getIntParam(params, "limit")
	return tools.WebSearch(ctx, query, limit)
}

// ==================== Product Barcode Tool Invocations ====================

// invokeListBarcodes invokes the list/search barcodes tool
func (s *MCPServer) invokeListBarcodes(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.ListBarcodes(ctx, holdingCode, keyword, limit)
}

// invokeCreateBarcode invokes the create barcode tool
func (s *MCPServer) invokeCreateBarcode(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
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
	return tools.CreateBarcode(ctx, holdingCode, barcode, itemCode, names, itemUnitCode, itemUnitNames, prices, standValue, divideValue, isMainBarcode, groupCode, groupNames, categoryCode, categoryNames)
}

// invokeCreateBarcodes invokes the bulk create barcodes tool
func (s *MCPServer) invokeCreateBarcodes(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	barcodes := getStringParam(params, "barcodes")
	return tools.CreateBarcodes(ctx, holdingCode, barcodes)
}

// invokeUpdateBarcode invokes the update barcode tool
func (s *MCPServer) invokeUpdateBarcode(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	guidFixed := getStringParam(params, "guidfixed")
	names := getStringParam(params, "names")
	itemUnitCode := getStringParam(params, "itemunitcode")
	itemUnitNames := getStringParam(params, "itemunitnames")
	prices := getStringParam(params, "prices")
	groupCode := getStringParam(params, "groupcode")
	groupNames := getStringParam(params, "groupnames")
	categoryCode := getStringParam(params, "categorycode")
	categoryNames := getStringParam(params, "categorynames")
	return tools.UpdateBarcode(ctx, holdingCode, guidFixed, names, itemUnitCode, itemUnitNames, prices, groupCode, groupNames, categoryCode, categoryNames)
}

// invokeDeleteBarcode invokes the delete barcode tool
func (s *MCPServer) invokeDeleteBarcode(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	guidFixed := getStringParam(params, "guidfixed")
	return tools.DeleteBarcode(ctx, holdingCode, guidFixed)
}

// invokeDeleteBarcodes invokes the bulk delete barcodes tool
func (s *MCPServer) invokeDeleteBarcodes(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	guidfixeds := getStringParam(params, "guidfixeds")
	return tools.DeleteBarcodes(ctx, holdingCode, guidfixeds)
}

// invokeGetBarcodeSchema invokes the get barcode schema tool
func (s *MCPServer) invokeGetBarcodeSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return tools.GetBarcodeSchema(), nil
}

// invokeGetRefBarcodes invokes the get reference barcodes tool
func (s *MCPServer) invokeGetRefBarcodes(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	itemCode := getStringParam(params, "itemcode")
	barcode := getStringParam(params, "barcode")
	return tools.GetRefBarcodes(ctx, holdingCode, itemCode, barcode)
}

// invokeSetRefBarcode invokes the set reference barcode tool
func (s *MCPServer) invokeSetRefBarcode(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	barcode := getStringParam(params, "barcode")
	refBarcode := getStringParam(params, "refbarcode")
	qty := getFloatParam(params, "qty")
	standValue := getFloatParam(params, "standvalue")
	divideValue := getFloatParam(params, "dividevalue")
	condition := getBoolParam(params, "condition")
	return tools.SetRefBarcode(ctx, holdingCode, barcode, refBarcode, qty, standValue, divideValue, condition)
}

// invokeCreateMultiUnitBarcode invokes the create multi-unit barcode tool
func (s *MCPServer) invokeCreateMultiUnitBarcode(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	itemCode := getStringParam(params, "itemcode")
	names := getStringParam(params, "names")
	units := getStringParam(params, "units")
	return tools.CreateMultiUnitBarcode(ctx, holdingCode, itemCode, names, units)
}

func (s *MCPServer) invokeRebuildProducts(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	return tools.RebuildProducts(ctx, holdingCode)
}

func (s *MCPServer) invokeRebuildEmbeddings(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	forceAll := getBoolParam(params, "forceall")
	entityType := getStringParam(params, "entitytype")
	if entityType == "all" {
		return tools.RebuildAllEmbeddings(ctx, holdingCode, forceAll)
	}
	return tools.RebuildEmbeddings(ctx, holdingCode, forceAll, entityType)
}

// ==================== Product Group Tool Invocations ====================

func (s *MCPServer) invokeListProductGroups(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.ListProductGroups(ctx, holdingCode, keyword, limit)
}

func (s *MCPServer) invokeCreateProductGroup(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	code := getStringParam(params, "code")
	names := getStringParam(params, "names")
	return tools.CreateProductGroup(ctx, holdingCode, code, names)
}

func (s *MCPServer) invokeCreateProductGroups(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	groups := getStringParam(params, "groups")
	return tools.CreateProductGroups(ctx, holdingCode, groups)
}

func (s *MCPServer) invokeUpdateProductGroup(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	code := getStringParam(params, "code")
	names := getStringParam(params, "names")
	return tools.UpdateProductGroup(ctx, holdingCode, code, names)
}

func (s *MCPServer) invokeDeleteProductGroup(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	code := getStringParam(params, "code")
	return tools.DeleteProductGroup(ctx, holdingCode, code)
}

func (s *MCPServer) invokeDeleteProductGroups(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	codes := getStringParam(params, "codes")
	return tools.DeleteProductGroups(ctx, holdingCode, codes)
}

func (s *MCPServer) invokeGetProductGroupSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return tools.GetProductGroupSchema(), nil
}

// ==================== Product Category Tool Invocations ====================

func (s *MCPServer) invokeListProductCategories(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.ListProductCategories(ctx, holdingCode, keyword, limit)
}

func (s *MCPServer) invokeCreateProductCategory(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	names := getStringParam(params, "names")
	parentGUID := getStringParam(params, "parentguid")
	groupNumber := getIntParam(params, "groupnumber")
	return tools.CreateProductCategory(ctx, holdingCode, names, parentGUID, groupNumber)
}

func (s *MCPServer) invokeCreateProductCategories(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	categories := getStringParam(params, "categories")
	return tools.CreateProductCategories(ctx, holdingCode, categories)
}

func (s *MCPServer) invokeUpdateProductCategory(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	guidFixed := getStringParam(params, "guidfixed")
	names := getStringParam(params, "names")
	parentGUID := getStringParam(params, "parentguid")
	groupNumber := getIntParam(params, "groupnumber")
	return tools.UpdateProductCategory(ctx, holdingCode, guidFixed, names, parentGUID, groupNumber)
}

func (s *MCPServer) invokeDeleteProductCategory(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	guidFixed := getStringParam(params, "guidfixed")
	return tools.DeleteProductCategory(ctx, holdingCode, guidFixed)
}

func (s *MCPServer) invokeDeleteProductCategories(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	guidfixeds := getStringParam(params, "guidfixeds")
	return tools.DeleteProductCategories(ctx, holdingCode, guidfixeds)
}

func (s *MCPServer) invokeGetProductCategorySchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return tools.GetProductCategorySchema(), nil
}

// ==================== Creditor Invokers ====================

func (s *MCPServer) invokeListCreditors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.ListCreditors(ctx, holdingCode, keyword, limit)
}

func (s *MCPServer) invokeSearchCreditors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.SearchCreditors(ctx, holdingCode, keyword, limit)
}

func (s *MCPServer) invokeCreateCreditor(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	code := getStringParam(params, "code")
	names := getStringParam(params, "names")
	taxid := getStringParam(params, "taxid")
	email := getStringParam(params, "email")
	personaltype := int8(getIntParam(params, "personaltype"))
	creditday := getIntParam(params, "creditday")
	address := getStringParam(params, "addressforbilling")
	return tools.CreateCreditor(ctx, holdingCode, code, names, taxid, email, personaltype, creditday, address)
}

func (s *MCPServer) invokeCreateCreditors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	creditors := getStringParam(params, "creditors")
	return tools.CreateCreditors(ctx, holdingCode, creditors)
}

func (s *MCPServer) invokeUpdateCreditor(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	guidfixed := getStringParam(params, "guidfixed")
	names := getStringParam(params, "names")
	taxid := getStringParam(params, "taxid")
	email := getStringParam(params, "email")
	creditday := getIntParam(params, "creditday")
	address := getStringParam(params, "addressforbilling")
	return tools.UpdateCreditor(ctx, holdingCode, guidfixed, names, taxid, email, creditday, address)
}

func (s *MCPServer) invokeDeleteCreditor(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	guidfixed := getStringParam(params, "guidfixed")
	return tools.DeleteCreditor(ctx, holdingCode, guidfixed)
}

func (s *MCPServer) invokeDeleteCreditors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	guidfixeds := getStringParam(params, "guidfixeds")
	return tools.DeleteCreditors(ctx, holdingCode, guidfixeds)
}

func (s *MCPServer) invokeGetCreditorSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return tools.GetCreditorSchema(), nil
}

// ==================== Debtor Invokers ====================

func (s *MCPServer) invokeListDebtors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.ListDebtors(ctx, holdingCode, keyword, limit)
}

func (s *MCPServer) invokeSearchDebtors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.SearchDebtors(ctx, holdingCode, keyword, limit)
}

func (s *MCPServer) invokeCreateDebtor(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	code := getStringParam(params, "code")
	names := getStringParam(params, "names")
	taxid := getStringParam(params, "taxid")
	email := getStringParam(params, "email")
	personaltype := int8(getIntParam(params, "personaltype"))
	creditday := getIntParam(params, "creditday")
	address := getStringParam(params, "addressforbilling")
	return tools.CreateDebtor(ctx, holdingCode, code, names, taxid, email, personaltype, creditday, address)
}

func (s *MCPServer) invokeCreateDebtors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	debtors := getStringParam(params, "debtors")
	return tools.CreateDebtors(ctx, holdingCode, debtors)
}

func (s *MCPServer) invokeUpdateDebtor(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	guidfixed := getStringParam(params, "guidfixed")
	names := getStringParam(params, "names")
	taxid := getStringParam(params, "taxid")
	email := getStringParam(params, "email")
	creditday := getIntParam(params, "creditday")
	address := getStringParam(params, "addressforbilling")
	return tools.UpdateDebtor(ctx, holdingCode, guidfixed, names, taxid, email, creditday, address)
}

func (s *MCPServer) invokeDeleteDebtor(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	guidfixed := getStringParam(params, "guidfixed")
	return tools.DeleteDebtor(ctx, holdingCode, guidfixed)
}

func (s *MCPServer) invokeDeleteDebtors(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	guidfixeds := getStringParam(params, "guidfixeds")
	return tools.DeleteDebtors(ctx, holdingCode, guidfixeds)
}

func (s *MCPServer) invokeGetDebtorSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return tools.GetDebtorSchema(), nil
}

// ==================== Purchase Order Invokers ====================

func (s *MCPServer) invokeListPurchaseOrders(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.ListPurchaseOrders(ctx, holdingCode, keyword, limit)
}

func (s *MCPServer) invokeCreatePurchaseOrder(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	docno := getStringParam(params, "docno")
	custcode := getStringParam(params, "custcode")
	custnames := getStringParam(params, "custnames")
	details := getStringParam(params, "details")
	description := getStringParam(params, "description")
	transflag := getIntParam(params, "transflag")
	vattype := int8(getIntParam(params, "vattype"))
	vatrate := getFloatParam(params, "vatrate")
	totalamount := getFloatParam(params, "totalamount")
	return tools.CreatePurchaseOrder(ctx, holdingCode, docno, custcode, custnames, details, description, transflag, vattype, vatrate, totalamount)
}

func (s *MCPServer) invokeUpdatePurchaseOrder(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	guidfixed := getStringParam(params, "guidfixed")
	custnames := getStringParam(params, "custnames")
	details := getStringParam(params, "details")
	description := getStringParam(params, "description")
	totalamount := getFloatParam(params, "totalamount")
	status := int8(getIntParam(params, "status"))
	return tools.UpdatePurchaseOrder(ctx, holdingCode, guidfixed, custnames, details, description, totalamount, status)
}

func (s *MCPServer) invokeDeletePurchaseOrder(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	guidfixed := getStringParam(params, "guidfixed")
	return tools.DeletePurchaseOrder(ctx, holdingCode, guidfixed)
}

func (s *MCPServer) invokeGetPurchaseOrderSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return tools.GetPurchaseOrderSchema(), nil
}

// ==================== Purchase Requisition Invokers ====================

func (s *MCPServer) invokeListPurchaseRequisitions(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	keyword := getStringParam(params, "keyword")
	limit := getIntParam(params, "limit")
	return tools.ListPurchaseRequisitions(ctx, holdingCode, keyword, limit)
}

func (s *MCPServer) invokeCreatePurchaseRequisition(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	docno := getStringParam(params, "docno")
	requesterCode := getStringParam(params, "requestercode")
	requesterName := getStringParam(params, "requestername")
	departmentCode := getStringParam(params, "departmentcode")
	departmentNames := getStringParam(params, "departmentnames")
	purpose := getStringParam(params, "purpose")
	budgetCode := getStringParam(params, "budgetcode")
	budgetAmount := getFloatParam(params, "budgetamount")
	urgency := int8(getIntParam(params, "urgency"))
	requestedDeliveryDate := getStringParam(params, "requesteddeliverydate")
	details := getStringParam(params, "details")
	description := getStringParam(params, "description")
	totalamount := getFloatParam(params, "totalamount")
	return tools.CreatePurchaseRequisition(ctx, holdingCode, docno, requesterCode, requesterName, departmentCode, departmentNames, purpose, budgetCode, budgetAmount, urgency, requestedDeliveryDate, details, description, totalamount)
}

func (s *MCPServer) invokeUpdatePurchaseRequisition(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	guidfixed := getStringParam(params, "guidfixed")
	requesterCode := getStringParam(params, "requestercode")
	requesterName := getStringParam(params, "requestername")
	departmentCode := getStringParam(params, "departmentcode")
	departmentNames := getStringParam(params, "departmentnames")
	purpose := getStringParam(params, "purpose")
	budgetCode := getStringParam(params, "budgetcode")
	budgetAmount := getFloatParam(params, "budgetamount")
	urgency := int8(getIntParam(params, "urgency"))
	requestedDeliveryDate := getStringParam(params, "requesteddeliverydate")
	details := getStringParam(params, "details")
	description := getStringParam(params, "description")
	totalamount := getFloatParam(params, "totalamount")
	status := int8(getIntParam(params, "status"))
	conversionStatus := getStringParam(params, "conversionstatus")
	return tools.UpdatePurchaseRequisition(ctx, holdingCode, guidfixed, requesterCode, requesterName, departmentCode, departmentNames, purpose, budgetCode, budgetAmount, urgency, requestedDeliveryDate, details, description, totalamount, status, conversionStatus)
}

func (s *MCPServer) invokeDeletePurchaseRequisition(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	holdingCode := getStringParam(params, "holdingcode")
	guidfixed := getStringParam(params, "guidfixed")
	return tools.DeletePurchaseRequisition(ctx, holdingCode, guidfixed)
}

func (s *MCPServer) invokeGetPurchaseRequisitionSchema(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return tools.GetPurchaseRequisitionSchema(), nil
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

	// SECURITY (2026-06-21): tool catalogs require an API key (they advertise the tool
	// surface incl. the raw-SQL dev tools). Only the liveness probe /mcp/health stays public.
	mcpGroup.GET("/tools", s.ListTools)
	g.GET("/mcp/health", s.HealthCheck)

	// Dev endpoints (เห็นทุก tools) — auth required
	mcpGroup.GET("/dev/tools", s.ListToolsDev)
	mcpGroup.GET("/dev/health", s.HealthCheck)
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
