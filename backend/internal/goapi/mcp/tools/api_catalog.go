package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/logger"
)

// ==================== API Catalog Structs ====================

type APIEndpoint struct {
	Method       string          `json:"method"`
	Path         string          `json:"path"`
	Description  string          `json:"description"`
	Category     string          `json:"category"`
	Source       string          `json:"source"` // "goapi" or "mainapi"
	AuthRequired bool            `json:"auth_required"`
	Parameters   []APIParam      `json:"parameters,omitempty"`
	RequestBody  *APIRequestBody `json:"request_body,omitempty"`
	Response     *APIResponse    `json:"response,omitempty"`
}

type APIParam struct {
	Name        string `json:"name"`
	In          string `json:"in"` // query, path, header, body
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type APIRequestBody struct {
	ContentType string                 `json:"content_type"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
	Fields      []SchemaField          `json:"fields,omitempty"`
	ModelName   string                 `json:"model_name,omitempty"`
	Example     map[string]interface{} `json:"example,omitempty"`
}

type APIResponse struct {
	ContentType string                 `json:"content_type"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
	Fields      []SchemaField          `json:"fields,omitempty"`
	ModelName   string                 `json:"model_name,omitempty"`
	Example     map[string]interface{} `json:"example,omitempty"`
}

type SchemaField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required,omitempty"`
	Description string `json:"description,omitempty"`
	Ref         string `json:"ref,omitempty"`
}

type APICatalogRequest struct {
	Category string `json:"category"`
	Keyword  string `json:"keyword"`
	Method   string `json:"method"`
	Source   string `json:"source"`
	Limit    int    `json:"limit"`
}

type APICatalogResponse struct {
	Endpoints     []APIEndpoint `json:"endpoints"`
	TotalCount    int           `json:"total_count"`
	FilteredCount int           `json:"filtered_count"`
	Categories    []string      `json:"categories"`
	GeneratedAt   time.Time     `json:"generated_at"`
}

// ==================== Swagger Parsing Structs ====================

type swaggerDoc struct {
	Paths       map[string]map[string]json.RawMessage `json:"paths"`
	Definitions map[string]json.RawMessage            `json:"definitions,omitempty"`
}

type swaggerDefinition struct {
	Type       string                        `json:"type"`
	Properties map[string]swaggerPropertyDef `json:"properties"`
	Required   []string                      `json:"required"`
}

type swaggerPropertyDef struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Format      string `json:"format"`
	Ref         string `json:"$ref"`
}

type swaggerOperation struct {
	Tags        []string                 `json:"tags"`
	Description string                   `json:"description"`
	Summary     string                   `json:"summary"`
	Parameters  []swaggerParam           `json:"parameters"`
	Responses   map[string]interface{}   `json:"responses"`
	Security    []map[string]interface{} `json:"security"`
	Consumes    []string                 `json:"consumes"`
}

type swaggerParam struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
	Schema      interface{} `json:"schema,omitempty"`
}

// ==================== Swagger Cache ====================

var (
	swaggerOnce        sync.Once
	swaggerEndpoints   []APIEndpoint
	swaggerErr         error
	swaggerDefinitions map[string]json.RawMessage // cached definitions for $ref resolve
)

// ==================== Main Function ====================

func GetAPICatalog(req APICatalogRequest) (*APICatalogResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 2000 {
		req.Limit = 2000
	}

	// รวม endpoints จากทั้ง 2 sources
	var allEndpoints []APIEndpoint

	// GoAPI endpoints (hardcoded catalog)
	if req.Source == "" || req.Source == "goapi" {
		allEndpoints = append(allEndpoints, getGoAPIEndpoints()...)
	}

	// MainAPI endpoints (จาก swagger.json)
	if req.Source == "" || req.Source == "mainapi" {
		mainEndpoints, err := getMainAPIEndpoints()
		if err != nil {
			logger.Warn("[API Catalog] ไม่สามารถโหลด MainAPI endpoints: %v", err)
			// ไม่ return error — ยังส่ง GoAPI endpoints ได้
		} else {
			allEndpoints = append(allEndpoints, mainEndpoints...)
		}
	}

	totalCount := len(allEndpoints)

	// Apply filters
	filtered := filterEndpoints(allEndpoints, req)

	// Collect categories
	categorySet := make(map[string]bool)
	for _, ep := range allEndpoints {
		categorySet[ep.Category] = true
	}
	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}

	// Apply limit
	if len(filtered) > req.Limit {
		filtered = filtered[:req.Limit]
	}

	return &APICatalogResponse{
		Endpoints:     filtered,
		TotalCount:    totalCount,
		FilteredCount: len(filtered),
		Categories:    categories,
		GeneratedAt:   time.Now(),
	}, nil
}

// ==================== Filter ====================

func filterEndpoints(endpoints []APIEndpoint, req APICatalogRequest) []APIEndpoint {
	if req.Category == "" && req.Keyword == "" && req.Method == "" {
		return endpoints
	}

	var result []APIEndpoint
	keyword := strings.ToLower(req.Keyword)
	method := strings.ToUpper(req.Method)
	category := strings.ToLower(req.Category)

	for _, ep := range endpoints {
		// Filter by method
		if method != "" && ep.Method != method {
			continue
		}
		// Filter by category
		if category != "" && strings.ToLower(ep.Category) != category {
			continue
		}
		// Filter by keyword (search in path + description)
		if keyword != "" {
			pathMatch := strings.Contains(strings.ToLower(ep.Path), keyword)
			descMatch := strings.Contains(strings.ToLower(ep.Description), keyword)
			catMatch := strings.Contains(strings.ToLower(ep.Category), keyword)
			if !pathMatch && !descMatch && !catMatch {
				continue
			}
		}
		result = append(result, ep)
	}
	return result
}

// ==================== MainAPI: Swagger Parser ====================

func getMainAPIEndpoints() ([]APIEndpoint, error) {
	swaggerOnce.Do(func() {
		swaggerEndpoints, swaggerErr = parseSwaggerFile()
	})
	return swaggerEndpoints, swaggerErr
}

func parseSwaggerFile() ([]APIEndpoint, error) {
	// หา swagger.json
	paths := []string{
		"docs/swagger.json",
		"/app/docs/swagger.json",
	}

	var data []byte
	var err error
	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil {
			logger.Info("[API Catalog] โหลด swagger.json จาก: %s", p)
			break
		}
	}
	if data == nil {
		return nil, fmt.Errorf("ไม่พบ docs/swagger.json")
	}

	var doc swaggerDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse swagger.json failed: %w", err)
	}

	// เก็บ definitions ไว้ใช้ resolve $ref
	swaggerDefinitions = doc.Definitions

	var endpoints []APIEndpoint
	for path, methods := range doc.Paths {
		for method, rawOp := range methods {
			method = strings.ToUpper(method)
			if method == "OPTIONS" || method == "HEAD" {
				continue
			}

			var op swaggerOperation
			if err := json.Unmarshal(rawOp, &op); err != nil {
				continue
			}

			// Category จาก tags
			category := "other"
			if len(op.Tags) > 0 {
				category = op.Tags[0]
			}

			// Description (+ override สำหรับ endpoints สำคัญ)
			desc := op.Summary
			if desc == "" {
				desc = op.Description
			}
			if override, ok := getDescriptionOverride(method, path); ok {
				desc = override
			}

			// Auth required
			authRequired := len(op.Security) > 0

			// Parameters
			var params []APIParam
			var reqBody *APIRequestBody
			for _, sp := range op.Parameters {
				if sp.In == "body" {
					reqBody = &APIRequestBody{
						ContentType: "application/json",
					}
					if sp.Schema != nil {
						if schemaMap, ok := sp.Schema.(map[string]interface{}); ok {
							reqBody.Schema = schemaMap
							// Resolve $ref → fields
							if ref, ok := schemaMap["$ref"].(string); ok {
								fields, modelName := resolveSwaggerRef(ref)
								reqBody.Fields = fields
								reqBody.ModelName = modelName
							}
						}
					}
					continue
				}
				paramType := sp.Type
				if paramType == "" {
					paramType = "string"
				}
				params = append(params, APIParam{
					Name:        sp.Name,
					In:          sp.In,
					Type:        paramType,
					Required:    sp.Required,
					Description: sp.Description,
				})
			}

			// Response schema (200)
			var response *APIResponse
			if resp200, ok := op.Responses["200"]; ok {
				if respMap, ok := resp200.(map[string]interface{}); ok {
					response = &APIResponse{
						ContentType: "application/json",
					}
					if schema, ok := respMap["schema"]; ok {
						if schemaMap, ok := schema.(map[string]interface{}); ok {
							response.Schema = schemaMap
							// Resolve $ref → fields
							if ref, ok := schemaMap["$ref"].(string); ok {
								fields, modelName := resolveSwaggerRef(ref)
								response.Fields = fields
								response.ModelName = modelName
							}
						}
					}
				}
			}

			endpoints = append(endpoints, APIEndpoint{
				Method:       method,
				Path:         path,
				Description:  desc,
				Category:     category,
				Source:       "mainapi",
				AuthRequired: authRequired,
				Parameters:   params,
				RequestBody:  reqBody,
				Response:     response,
			})
		}
	}

	logger.Success("[API Catalog] โหลด MainAPI endpoints: %d routes", len(endpoints))
	return endpoints, nil
}

// ==================== $ref Resolver ====================

// resolveSwaggerRef — resolve $ref เป็น field list จาก swagger definitions
func resolveSwaggerRef(ref string) ([]SchemaField, string) {
	// Extract definition name from "#/definitions/models.AuthResponse"
	prefix := "#/definitions/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, ""
	}
	defName := ref[len(prefix):]

	if swaggerDefinitions == nil {
		return nil, defName
	}

	rawDef, ok := swaggerDefinitions[defName]
	if !ok {
		return nil, defName
	}

	var def swaggerDefinition
	if err := json.Unmarshal(rawDef, &def); err != nil {
		return nil, defName
	}

	// สร้าง required set
	requiredSet := make(map[string]bool)
	for _, r := range def.Required {
		requiredSet[r] = true
	}

	// Convert properties → SchemaField
	var fields []SchemaField
	for propName, prop := range def.Properties {
		f := SchemaField{
			Name:        propName,
			Type:        prop.Type,
			Required:    requiredSet[propName],
			Description: prop.Description,
		}
		if prop.Format != "" {
			f.Type = prop.Type + "(" + prop.Format + ")"
		}
		// ถ้า property เป็น $ref อีก → แสดง ref name
		if prop.Ref != "" {
			refPrefix := "#/definitions/"
			if strings.HasPrefix(prop.Ref, refPrefix) {
				f.Ref = prop.Ref[len(refPrefix):]
				f.Type = "object"
			}
		}
		if f.Type == "" && f.Ref != "" {
			f.Type = "object"
		}
		fields = append(fields, f)
	}

	// Extract short model name (e.g., "models.AuthResponse" → "AuthResponse")
	shortName := defName
	if idx := strings.LastIndex(defName, "."); idx >= 0 {
		shortName = defName[idx+1:]
	}

	return fields, shortName
}

// ==================== Description Overrides ====================

// getDescriptionOverride — override description สำหรับ MainAPI endpoints ที่ Swagger annotation ไม่ชัด
func getDescriptionOverride(method, path string) (string, bool) {
	overrides := map[string]string{
		// ===== Authentication =====
		"POST /login":             "Login with username/password, returns JWT access token",
		"POST /login/email":       "Login with email/password",
		"POST /login/phonenumber": "Login with phone number/password",
		"POST /login/line":        "Login with LINE access token",
		"POST /login/google":      "Login with Google account",
		"POST /login/pos":         "POS machine login (shopid + username)",
		"POST /register":          "Register new user account",
		"POST /refresh":           "Refresh expired JWT access token",
		"POST /logout":            "Logout and invalidate token",

		// ===== Shop =====
		"GET /list-shop":    "List all shops for current user",
		"GET /shop/{id}":    "Get shop profile by ID",
		"PUT /shop/{id}":    "Update shop information",
		"DELETE /shop/{id}": "Delete shop",
		"POST /select-shop": "Select active shop for session",

		// ===== Branch =====
		"GET /shop/branch":              "List all branches with search and pagination",
		"GET /shop/branch/list":         "Search branches with limit/offset pagination",
		"GET /shop/branch/{id}":         "Get branch detail by ID",
		"POST /shop/branch":             "Create new branch",
		"DELETE /shop/branch":           "Delete branch",
		"GET /organization/branch/list": "List branches with search, offset/limit pagination, and language filter",
		"GET /organization/branch/{id}": "Get branch detail by guidfixed",

		// ===== Product Barcode =====
		"GET /product/barcode/by-code":       "Get product barcodes by code array (JSON encoded)",
		"POST /product/barcode/bulk":         "Bulk create product barcodes",
		"GET /product/barcode/pk/{barcode}":  "Get product barcode by primary key (barcode)",
		"GET /product/barcode/bom/{barcode}": "Get BOM (Bill of Materials) for product barcode",
		"GET /product/barcode/export":        "Export product barcodes",

		// ===== Warehouse =====
		"GET /warehouse":    "List warehouses with search and pagination",
		"POST /warehouse":   "Create new warehouse",
		"DELETE /warehouse": "Delete warehouse",

		// ===== Customer =====
		"GET /debtaccount/customer":      "List customers with search and pagination",
		"POST /debtaccount/customer":     "Create new customer",
		"DELETE /debtaccount/customer":   "Delete customer",
		"GET /debtaccount/customer/{id}": "Get customer detail by ID",
		"PUT /debtaccount/customer/{id}": "Update customer",

		// ===== Creditor =====
		"GET /debtaccount/creditor":       "List creditors (suppliers) with search and pagination",
		"POST /debtaccount/creditor":      "Create new creditor",
		"GET /debtaccount/creditor/{id}":  "Get creditor detail by ID",
		"PUT /debtaccount/creditor/{id}":  "Update creditor",
		"POST /debtaccount/creditor/bulk": "Bulk create creditors",

		// ===== Currency =====
		"GET /currency":    "List currencies",
		"POST /currency":   "Create new currency",
		"DELETE /currency": "Delete currency",

		// ===== Member =====
		"GET /member":      "List members with search and pagination",
		"POST /member":     "Create new member",
		"GET /member/{id}": "Get member detail by ID",
		"PUT /member/{id}": "Update member",

		// ===== Employee =====
		"GET /shop/employee":         "List employees with search and pagination",
		"POST /shop/employee":        "Create new employee",
		"GET /shop/employee/{id}":    "Get employee detail by ID",
		"PUT /shop/employee/{id}":    "Update employee",
		"DELETE /shop/employee/{id}": "Delete employee",

		// ===== User/Permission =====
		"GET /shop/permission/{username}":    "Get shop user permission and profile by username",
		"PUT /shop/permission":               "Save shop user permission and profile (position, department, LINE, approval)",
		"DELETE /shop/permission/{username}": "Delete shop user permission",

		// ===== Transaction: Sale Invoice =====
		"GET /transaction/sale-invoice":      "List sale invoices with search and pagination",
		"POST /transaction/sale-invoice":     "Create sale invoice",
		"DELETE /transaction/sale-invoice":   "Delete sale invoice",
		"GET /transaction/sale-invoice/{id}": "Get sale invoice detail by ID",

		// ===== Transaction: Sale Invoice Return =====
		"GET /transaction/sale-invoice-return":    "List sale returns",
		"POST /transaction/sale-invoice-return":   "Create sale return",
		"DELETE /transaction/sale-invoice-return": "Delete sale return",

		// ===== Transaction: Purchase Order =====
		"GET /transaction/purchase-order":             "List purchase orders with search and pagination",
		"POST /transaction/purchase-order":            "Create purchase order",
		"DELETE /transaction/purchase-order":          "Delete purchase order",
		"GET /transaction/purchase-order/list":        "Search purchase orders with limit/offset pagination",
		"GET /transaction/purchase-order/code/{code}": "Get purchase order by document code",

		// ===== Transaction: Purchase =====
		"GET /transaction/purchase":    "List purchases with search and pagination",
		"POST /transaction/purchase":   "Create purchase",
		"DELETE /transaction/purchase": "Delete purchase",

		// ===== Stock =====
		"GET /transaction/stock-transfer":    "List stock transfers",
		"POST /transaction/stock-transfer":   "Create stock transfer",
		"GET /transaction/stock-adjustment":  "List stock adjustments",
		"POST /transaction/stock-adjustment": "Create stock adjustment",
		"GET /transaction/stock-balance":     "List stock balance (opening balance)",
		"POST /transaction/stock-balance":    "Create stock balance record",

		// ===== Image/File =====
		"POST /upload/productimage": "Upload product image",
		"POST /upload/shoplogo":     "Upload shop logo image",

		// ===== Settings =====
		"GET /setting": "Get shop settings/configuration",
		"PUT /setting": "Update shop settings/configuration",
	}

	key := method + " " + path
	if desc, ok := overrides[key]; ok {
		return desc, true
	}
	return "", false
}

// ==================== GoAPI: Hardcoded Catalog ====================

func getGoAPIEndpoints() []APIEndpoint {
	return []APIEndpoint{
		// ===== Health & Status =====
		{Method: "GET", Path: "/goapi/", Description: "GoAPI root - returns version and status", Category: "health", Source: "goapi", AuthRequired: false,
			Response: &APIResponse{ContentType: "application/json", Example: map[string]interface{}{"message": "Hello, World!", "version": "1.1.1121", "status": "healthy"}}},
		{Method: "GET", Path: "/goapi/version", Description: "Get API version", Category: "health", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/health", Description: "Health check - returns status and version", Category: "health", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/health/kafka", Description: "Kafka connection health check", Category: "health", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/health/background", Description: "Background task status", Category: "health", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/health/queue", Description: "Queue system status", Category: "health", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/health/queue/:shopid", Description: "Queue status for specific shop", Category: "health", Source: "goapi", AuthRequired: false,
			Parameters: []APIParam{{Name: "shopid", In: "path", Type: "string", Required: true, Description: "Shop ID"}}},
		{Method: "GET", Path: "/goapi/api/health/database", Description: "Database connection health check (PostgreSQL + ClickHouse)", Category: "health", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/health/system", Description: "System health - memory, goroutines, uptime", Category: "health", Source: "goapi", AuthRequired: false},

		// ===== Database Operations =====
		{Method: "GET", Path: "/goapi/reportget", Description: "Get report data via GET method", Category: "database", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/get", Description: "Execute PostgreSQL SELECT query", Category: "database", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type":     "object",
					"required": []string{"shopid", "sql"},
					"properties": map[string]interface{}{
						"shopid": map[string]interface{}{"type": "string", "description": "Shop ID (database name)"},
						"sql":    map[string]interface{}{"type": "string", "description": "SQL SELECT query to execute"},
					},
				},
				Example: map[string]interface{}{"shopid": "SHOP001", "sql": "SELECT * FROM products LIMIT 10"}}},
		{Method: "POST", Path: "/goapi/exec", Description: "Execute PostgreSQL command (INSERT/UPDATE/DELETE)", Category: "database", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type":     "object",
					"required": []string{"shopid", "sql"},
					"properties": map[string]interface{}{
						"shopid": map[string]interface{}{"type": "string", "description": "Shop ID (database name)"},
						"sql":    map[string]interface{}{"type": "string", "description": "SQL command (INSERT/UPDATE/DELETE)"},
					},
				},
				Example: map[string]interface{}{"shopid": "SHOP001", "sql": "UPDATE products SET name='test' WHERE id=1"}}},
		{Method: "POST", Path: "/goapi/getdoc", Description: "Get document data from PostgreSQL", Category: "database", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type":     "object",
					"required": []string{"shopid", "docno"},
					"properties": map[string]interface{}{
						"shopid":    map[string]interface{}{"type": "string", "description": "Shop ID"},
						"docno":     map[string]interface{}{"type": "string", "description": "Document number"},
						"transflag": map[string]interface{}{"type": "integer", "description": "Transaction type flag (optional filter)"},
					},
				},
				Example: map[string]interface{}{"shopid": "SHOP001", "docno": "INV-001"}}},
		{Method: "POST", Path: "/goapi/mongogetdata", Description: "Query MongoDB collection", Category: "database", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type":     "object",
					"required": []string{"database", "collection"},
					"properties": map[string]interface{}{
						"database":   map[string]interface{}{"type": "string", "description": "MongoDB database name"},
						"collection": map[string]interface{}{"type": "string", "description": "MongoDB collection name"},
						"filter":     map[string]interface{}{"type": "object", "description": "MongoDB query filter"},
						"sort":       map[string]interface{}{"type": "object", "description": "Sort order (e.g., {\"_id\": -1})"},
						"limit":      map[string]interface{}{"type": "integer", "description": "Max results (default: 100)"},
					},
				},
				Example: map[string]interface{}{"database": "dbname", "collection": "collname", "filter": map[string]interface{}{"shopid": "SHOP001"}}}},
		{Method: "POST", Path: "/goapi/reportpost", Description: "Generate report via POST", Category: "database", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/rebuild/progress/:jobId", Description: "SSE endpoint for rebuild progress tracking", Category: "database", Source: "goapi", AuthRequired: false,
			Parameters: []APIParam{{Name: "jobId", In: "path", Type: "string", Required: true, Description: "Job ID for progress tracking"}}},

		// ===== Result Table =====
		{Method: "POST", Path: "/goapi/resultfromquery", Description: "Generate result table from SQL query", Category: "database", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"shopid": "SHOP001", "sql": "SELECT * FROM products"}}},
		{Method: "POST", Path: "/goapi/resultget", Description: "Get pre-generated result table", Category: "database", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/resulttopdf", Description: "Convert result table to PDF", Category: "database", Source: "goapi", AuthRequired: false},

		// ===== PDF Generation =====
		{Method: "POST", Path: "/goapi/genpdf", Description: "Generate PDF document from template", Category: "pdf", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type":     "object",
					"required": []string{"shopid", "docno"},
					"properties": map[string]interface{}{
						"shopid":      map[string]interface{}{"type": "string", "description": "Shop ID"},
						"docno":       map[string]interface{}{"type": "string", "description": "Document number"},
						"template":    map[string]interface{}{"type": "string", "description": "PDF template name (e.g., invoice, receipt, po)"},
						"orientation": map[string]interface{}{"type": "string", "description": "L=landscape, P=portrait (see pdf_orientation enum)"},
						"page_size":   map[string]interface{}{"type": "string", "description": "A4, A5, Letter (see pdf_page_size enum)"},
					},
				},
				Example: map[string]interface{}{"shopid": "SHOP001", "docno": "INV-001", "template": "invoice"}},
			Response: &APIResponse{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"success":   map[string]interface{}{"type": "boolean"},
						"url":       map[string]interface{}{"type": "string", "description": "PDF file URL path"},
						"file_size": map[string]interface{}{"type": "integer", "description": "File size in bytes"},
					},
				}}},
		{Method: "GET", Path: "/goapi/genpdf/history", Description: "Get PDF generation history", Category: "pdf", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/genpdf/history", Description: "List PDF generation history with filters", Category: "pdf", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/genpdf/reprint/:id", Description: "Reprint previously generated PDF", Category: "pdf", Source: "goapi", AuthRequired: false,
			Parameters: []APIParam{{Name: "id", In: "path", Type: "string", Required: true, Description: "PDF history ID"}}},

		// ===== Stock =====
		{Method: "POST", Path: "/goapi/processstockcalccost", Description: "Process stock calculation and costing", Category: "stock", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"shopid": "SHOP001"}}},
		{Method: "POST", Path: "/goapi/api/stockcost/query", Description: "Query stock cost data", Category: "stock", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"shopid": "SHOP001", "barcode": "1234567890"}}},
		{Method: "POST", Path: "/goapi/api/stockcost/summary", Description: "Get stock cost summary", Category: "stock", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"shopid": "SHOP001"}}},
		{Method: "POST", Path: "/goapi/api/stockcost/check", Description: "Check stock cost status", Category: "stock", Source: "goapi", AuthRequired: false},

		// ===== Transaction Calculator =====
		{Method: "POST", Path: "/goapi/api/transaction/calculate", Description: "Calculate transaction totals (tax, discount, net amount)", Category: "transaction", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type":     "object",
					"required": []string{"shopid", "items"},
					"properties": map[string]interface{}{
						"shopid":          map[string]interface{}{"type": "string", "description": "Shop ID"},
						"items":           map[string]interface{}{"type": "array", "description": "Line items", "items": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"barcode": map[string]interface{}{"type": "string"}, "qty": map[string]interface{}{"type": "number"}, "price": map[string]interface{}{"type": "number"}, "discount": map[string]interface{}{"type": "string", "description": "Discount text e.g. '10%'"}}}},
						"tax_type":        map[string]interface{}{"type": "integer", "description": "0=excluded, 1=included, 2=non-taxable (see vat_type enum)"},
						"discount_amount": map[string]interface{}{"type": "number", "description": "Document-level discount amount"},
					},
				},
				Example: map[string]interface{}{
					"shopid": "SHOP001", "items": []map[string]interface{}{{"barcode": "001", "qty": 2, "price": 100}},
					"tax_type": 1, "discount_amount": 50,
				}},
			Response: &APIResponse{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"total_amount":    map[string]interface{}{"type": "number", "description": "Total before discount"},
						"discount_amount": map[string]interface{}{"type": "number", "description": "Total discount"},
						"net_amount":      map[string]interface{}{"type": "number", "description": "Net amount after discount"},
						"tax_amount":      map[string]interface{}{"type": "number", "description": "VAT amount"},
						"items":           map[string]interface{}{"type": "array", "description": "Calculated line items with amounts"},
					},
				},
				Example: map[string]interface{}{
					"total_amount": 200, "discount_amount": 50, "net_amount": 150, "tax_amount": 9.81,
				}}},
		{Method: "POST", Path: "/goapi/api/transaction/quick-calc", Description: "Quick calculation without full transaction context", Category: "transaction", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/transaction/validate-payment", Description: "Validate payment amounts and methods", Category: "transaction", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/transaction/purchase-history", Description: "Get purchase history for products", Category: "transaction", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"shopid": "SHOP001", "barcode": "001"}}},

		// ===== Sales Report =====
		{Method: "POST", Path: "/goapi/api/report/sales/by-document", Description: "Sales report grouped by document", Category: "sales-report", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type":     "object",
					"required": []string{"shopid", "from_date", "to_date"},
					"properties": map[string]interface{}{
						"shopid":    map[string]interface{}{"type": "string", "description": "Shop ID"},
						"from_date": map[string]interface{}{"type": "string", "format": "date", "description": "Start date (YYYY-MM-DD)"},
						"to_date":   map[string]interface{}{"type": "string", "format": "date", "description": "End date (YYYY-MM-DD)"},
						"whcode":    map[string]interface{}{"type": "string", "description": "Filter by warehouse code"},
						"transflag": map[string]interface{}{"type": "integer", "description": "Filter by transaction type (see transflag enum)"},
					},
				},
				Example: map[string]interface{}{
					"shopid": "SHOP001", "from_date": "2025-01-01", "to_date": "2025-12-31",
				}}},
		{Method: "POST", Path: "/goapi/api/report/sales/summary", Description: "Sales summary report with totals", Category: "sales-report", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type":     "object",
					"required": []string{"shopid", "from_date", "to_date"},
					"properties": map[string]interface{}{
						"shopid":    map[string]interface{}{"type": "string", "description": "Shop ID"},
						"from_date": map[string]interface{}{"type": "string", "format": "date", "description": "Start date (YYYY-MM-DD)"},
						"to_date":   map[string]interface{}{"type": "string", "format": "date", "description": "End date (YYYY-MM-DD)"},
					},
				},
				Example: map[string]interface{}{
					"shopid": "SHOP001", "from_date": "2025-01-01", "to_date": "2025-12-31",
				}},
			Response: &APIResponse{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"total_amount":   map[string]interface{}{"type": "number", "description": "Total sales amount"},
						"total_cost":     map[string]interface{}{"type": "number", "description": "Total cost"},
						"total_profit":   map[string]interface{}{"type": "number", "description": "Total profit"},
						"document_count": map[string]interface{}{"type": "integer", "description": "Number of documents"},
					},
				}}},

		// ===== Product Search =====
		{Method: "POST", Path: "/goapi/api/product/search", Description: "Search products with Thai full-text search, returns stock balance", Category: "product", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type":     "object",
					"required": []string{"shopid"},
					"properties": map[string]interface{}{
						"shopid":  map[string]interface{}{"type": "string", "description": "Shop ID"},
						"keyword": map[string]interface{}{"type": "string", "description": "Search keyword (Thai full-text supported)"},
						"whcode":  map[string]interface{}{"type": "string", "description": "Warehouse code filter"},
						"limit":   map[string]interface{}{"type": "integer", "description": "Max results (default: 50, max: 200)"},
						"offset":  map[string]interface{}{"type": "integer", "description": "Offset for pagination"},
					},
				},
				Example: map[string]interface{}{
					"shopid": "SHOP001", "keyword": "น้ำตาล", "whcode": "WH01", "limit": 50,
				}},
			Response: &APIResponse{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"products":    map[string]interface{}{"type": "array", "description": "Matched products with barcode, name, price, balance, unit info"},
						"total_count": map[string]interface{}{"type": "integer", "description": "Total matching products"},
					},
				},
				Example: map[string]interface{}{
					"products": []map[string]interface{}{{"barcode": "001", "name": "น้ำตาล", "price": 35.00, "balance": "1 กล่อง x 2 โหล"}},
				}}},
		{Method: "POST", Path: "/goapi/api/product/barcode", Description: "Search product by exact barcode", Category: "product", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type":     "object",
					"required": []string{"shopid", "barcode"},
					"properties": map[string]interface{}{
						"shopid":  map[string]interface{}{"type": "string", "description": "Shop ID"},
						"barcode": map[string]interface{}{"type": "string", "description": "Exact barcode to lookup"},
					},
				},
				Example: map[string]interface{}{"shopid": "SHOP001", "barcode": "8850999220017"}}},
		{Method: "GET", Path: "/goapi/api/product/cache/stats", Description: "Get product cache statistics", Category: "product", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/product/cache/clear", Description: "Clear product cache for a shop", Category: "product", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/product/search/unified", Description: "Unified product search across multiple sources", Category: "product", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"shopid": "SHOP001", "keyword": "สินค้า", "limit": 20}}},

		// ===== Stock Report =====
		{Method: "POST", Path: "/goapi/api/stock-report/barcodes", Description: "Get stock report by barcodes", Category: "stock-report", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"shopid": "SHOP001", "barcodes": []string{"001", "002"}}}},
		{Method: "POST", Path: "/goapi/api/stock-report/warehouses", Description: "Get stock report by warehouses", Category: "stock-report", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"shopid": "SHOP001", "whcodes": []string{"WH01"}}}},

		// ===== LINE OA =====
		{Method: "POST", Path: "/goapi/api/lineoa/configs", Description: "Get all LINE OA configurations", Category: "lineoa", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"shopid": "SHOP001"}}},
		{Method: "POST", Path: "/goapi/api/lineoa/config", Description: "Get specific LINE OA config", Category: "lineoa", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/lineoa/config/save", Description: "Save LINE OA configuration", Category: "lineoa", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/lineoa/test", Description: "Test LINE OA configuration (send test message)", Category: "lineoa", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/lineoa/employees", Description: "Get linked LINE employees", Category: "lineoa", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/lineoa/employee/add", Description: "Add employee to LINE OA", Category: "lineoa", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/lineoa/employee/remove", Description: "Remove employee from LINE OA", Category: "lineoa", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/lineoa/employee/link", Description: "Generate LINE link token for employee", Category: "lineoa", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/user/lineoa/link", Description: "User-facing LINE OA link", Category: "lineoa", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/user/lineoa/callback", Description: "LINE OA OAuth callback", Category: "lineoa", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/user/lineoa/profile", Description: "Get user LINE OA profile", Category: "lineoa", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/lineoa/webhook", Description: "LINE OA webhook receiver", Category: "lineoa", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/webhook/lineoa", Description: "LINE OA webhook (alternate path)", Category: "lineoa", Source: "goapi", AuthRequired: false},

		// ===== Approval System =====
		{Method: "POST", Path: "/goapi/api/approval/po-settings", Description: "Get all PO approval settings", Category: "approval", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"shopid": "SHOP001"}}},
		{Method: "POST", Path: "/goapi/api/approval/po-setting", Description: "Get specific PO approval setting", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/po-setting/save", Description: "Save PO approval setting (create/update)", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/po-setting/delete", Description: "Delete PO approval setting", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/po-status/get", Description: "Get PO approval status", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/po-status/batch", Description: "Get batch PO approval statuses", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/po-status/submit", Description: "Submit PO for approval", Category: "approval", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type":     "object",
					"required": []string{"shopid", "docno", "submitted_by"},
					"properties": map[string]interface{}{
						"shopid":       map[string]interface{}{"type": "string", "description": "Shop ID"},
						"docno":        map[string]interface{}{"type": "string", "description": "PO document number"},
						"submitted_by": map[string]interface{}{"type": "string", "description": "Employee code who submits"},
					},
				}}},
		{Method: "POST", Path: "/goapi/api/approval/po-status/approve", Description: "Approve PO", Category: "approval", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type":     "object",
					"required": []string{"shopid", "docno", "approved_by"},
					"properties": map[string]interface{}{
						"shopid":      map[string]interface{}{"type": "string", "description": "Shop ID"},
						"docno":       map[string]interface{}{"type": "string", "description": "PO document number"},
						"approved_by": map[string]interface{}{"type": "string", "description": "Approver employee code"},
						"comment":     map[string]interface{}{"type": "string", "description": "Approval comment"},
					},
				}}},
		{Method: "POST", Path: "/goapi/api/approval/po-status/reject", Description: "Reject PO", Category: "approval", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type":     "object",
					"required": []string{"shopid", "docno", "rejected_by"},
					"properties": map[string]interface{}{
						"shopid":      map[string]interface{}{"type": "string", "description": "Shop ID"},
						"docno":       map[string]interface{}{"type": "string", "description": "PO document number"},
						"rejected_by": map[string]interface{}{"type": "string", "description": "Rejecter employee code"},
						"reason":      map[string]interface{}{"type": "string", "description": "Rejection reason"},
					},
				}}},
		{Method: "POST", Path: "/goapi/api/approval/po-status/withdraw", Description: "Withdraw PO from approval", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/po-status/pending", Description: "Get pending PO approvals", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/po-status/rejected", Description: "Get rejected PO list", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/notification/check", Description: "Check if notification was sent", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/notification/send", Description: "Send approval notification", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/notification/process", Description: "Process pending notifications", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/notification/logs", Description: "Get notification logs", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/notification/mark-opened", Description: "Mark notification as opened", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/timeline", Description: "Get approval timeline", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/notification/send-real", Description: "Send real approval notification (email/LINE)", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/notification/resend", Description: "Resend approval notification", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/approval/smtp-status", Description: "Get SMTP configuration status", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/test-email", Description: "Send test email", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/approval/lineoa-config-status", Description: "Get LINE OA config status for approval", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/approval/lineoa-configs", Description: "List all LINE OA configs for approval", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/test-line-push", Description: "Test LINE push message", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/approval/action", Description: "Approve/reject via token link (GET)", Category: "approval", Source: "goapi", AuthRequired: false,
			Parameters: []APIParam{{Name: "token", In: "query", Type: "string", Required: true, Description: "Approval token"}}},
		{Method: "POST", Path: "/goapi/api/approval/action", Description: "Approve/reject via token link (POST)", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/approval/token-info", Description: "Get approval token information", Category: "approval", Source: "goapi", AuthRequired: false,
			Parameters: []APIParam{{Name: "token", In: "query", Type: "string", Required: true, Description: "Approval token"}}},
		{Method: "POST", Path: "/goapi/api/approval/po-details", Description: "Get PO details for LIFF app", Category: "approval", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/approval/liff-approve", Description: "Approve PO via LIFF app", Category: "approval", Source: "goapi", AuthRequired: false},

		// ===== Data History =====
		{Method: "GET", Path: "/goapi/api/datahistory", Description: "Get data change history", Category: "datahistory", Source: "goapi", AuthRequired: false,
			Parameters: []APIParam{
				{Name: "shopid", In: "query", Type: "string", Required: true, Description: "Shop ID"},
				{Name: "collection", In: "query", Type: "string", Required: true, Description: "Collection/table name"},
			}},
		{Method: "GET", Path: "/goapi/api/datahistory/po", Description: "Get PO-specific data history", Category: "datahistory", Source: "goapi", AuthRequired: false},

		// ===== Purchase Order =====
		{Method: "POST", Path: "/goapi/api/purchase-order/manual-close", Description: "Manually close a purchase order", Category: "purchase-order", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"shopid": "SHOP001", "docno": "PO-001"}}},

		// ===== Migration =====
		{Method: "GET", Path: "/goapi/api/migrate/currency", Description: "Migrate currency columns", Category: "migration", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/migrate/currency-backfill", Description: "Backfill currency data", Category: "migration", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/migrate/clickhouse-softdelete", Description: "Migrate ClickHouse soft delete columns", Category: "migration", Source: "goapi", AuthRequired: false},

		// ===== MongoDB Operations =====
		{Method: "POST", Path: "/goapi/copymongouattodev", Description: "DEV-only copy MongoDB data from UAT/PRO to DEV", Category: "mongodb", Source: "goapi", AuthRequired: true},
		{Method: "POST", Path: "/goapi/previewcopymongo", Description: "Preview DEV-only MongoDB copy from UAT/PRO to DEV", Category: "mongodb", Source: "goapi", AuthRequired: true},
		{Method: "GET", Path: "/goapi/listsourceshops", Description: "List UAT/PRO source shops for DEV-only MongoDB copy", Category: "mongodb", Source: "goapi", AuthRequired: true},
		{Method: "POST", Path: "/goapi/atlas/get", Description: "Get data from MongoDB", Category: "mongodb", Source: "goapi", AuthRequired: true,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"database": "dbname", "collection": "collname", "filter": map[string]interface{}{}}}},
		{Method: "POST", Path: "/goapi/atlas/update", Description: "Update data in MongoDB", Category: "mongodb", Source: "goapi", AuthRequired: true},
		{Method: "POST", Path: "/goapi/atlas/delete", Description: "Delete data from MongoDB", Category: "mongodb", Source: "goapi", AuthRequired: true},

		// ===== ClickHouse =====
		{Method: "POST", Path: "/goapi/clickhouse/query", Description: "Execute ClickHouse query", Category: "clickhouse", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"shopid": "SHOP001", "sql": "SELECT count() FROM sales"}}},
		{Method: "POST", Path: "/goapi/clickhouse/querys", Description: "Execute multiple ClickHouse queries", Category: "clickhouse", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/clickhouse/select", Description: "Execute ClickHouse SELECT query", Category: "clickhouse", Source: "goapi", AuthRequired: false},

		// ===== Test (Kafka) =====
		{Method: "POST", Path: "/goapi/test/sale-order", Description: "Test: publish sale order Kafka message", Category: "test", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/test/purchase", Description: "Test: publish purchase Kafka message", Category: "test", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/test/purchase-order", Description: "Test: publish purchase order Kafka message", Category: "test", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/test/purchase-partial", Description: "Test: publish purchase partial Kafka message", Category: "test", Source: "goapi", AuthRequired: false},

		// ===== File / S3 =====
		{Method: "GET", Path: "/goapi/s3/file/*", Description: "Stream private object from S3/R2 storage — requires auth, object key must be under caller's shopid", Category: "file", Source: "goapi", AuthRequired: true},

		// ===== Upload (private, shop-scoped) =====
		{Method: "POST", Path: "/goapi/upload", Description: "Upload file (single file) — shopid resolved from auth context, object key is shopid/...", Category: "upload", Source: "goapi", AuthRequired: true,
			RequestBody: &APIRequestBody{ContentType: "multipart/form-data"}},
		{Method: "POST", Path: "/goapi/upload/init", Description: "Initialize chunked upload session", Category: "upload", Source: "goapi", AuthRequired: true,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"file_name": "file.xlsx", "total_chunks": 5}}},
		{Method: "POST", Path: "/goapi/upload/chunk", Description: "Upload a file chunk", Category: "upload", Source: "goapi", AuthRequired: true,
			RequestBody: &APIRequestBody{ContentType: "multipart/form-data"}},
		{Method: "POST", Path: "/goapi/upload/merge", Description: "Merge uploaded chunks into final file — shopid resolved from auth context", Category: "upload", Source: "goapi", AuthRequired: true},
		{Method: "GET", Path: "/goapi/upload/status/:uploadID", Description: "Get chunked upload status", Category: "upload", Source: "goapi", AuthRequired: true,
			Parameters: []APIParam{{Name: "uploadID", In: "path", Type: "string", Required: true, Description: "Upload session ID"}}},
		{Method: "DELETE", Path: "/goapi/upload/cancel/:uploadID", Description: "Cancel chunked upload", Category: "upload", Source: "goapi", AuthRequired: true,
			Parameters: []APIParam{{Name: "uploadID", In: "path", Type: "string", Required: true, Description: "Upload session ID"}}},

		// ===== Language =====
		{Method: "GET", Path: "/goapi/api/language/:lang", Description: "Get language translations", Category: "language", Source: "goapi", AuthRequired: false,
			Parameters: []APIParam{{Name: "lang", In: "path", Type: "string", Required: true, Description: "Language code (e.g., th, en)"}}},

		// ===== Image (private, shop-scoped) =====
		{Method: "POST", Path: "/goapi/image/upload", Description: "Upload image — object key is {shopid}/..., shopid resolved from auth context; mismatched form shopid is rejected with 403", Category: "image", Source: "goapi", AuthRequired: true,
			RequestBody: &APIRequestBody{ContentType: "multipart/form-data"}},
		{Method: "POST", Path: "/goapi/image/list", Description: "List images for the caller's shop; response URLs are backend proxy URLs only", Category: "image", Source: "goapi", AuthRequired: true,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"shopid": "SHOP001", "prefix": "products/"}}},
		{Method: "POST", Path: "/goapi/image/get", Description: "Get image by filename for the caller's shop", Category: "image", Source: "goapi", AuthRequired: true},
		{Method: "POST", Path: "/goapi/image/info", Description: "Get image metadata for the caller's shop", Category: "image", Source: "goapi", AuthRequired: true},
		{Method: "POST", Path: "/goapi/image/delete", Description: "Delete image for the caller's shop", Category: "image", Source: "goapi", AuthRequired: true},
		{Method: "POST", Path: "/goapi/image/promptpayverify", Description: "Verify PromptPay QR from image", Category: "image", Source: "goapi", AuthRequired: true},

		// ===== Attachment (private, shop-scoped) =====
		{Method: "POST", Path: "/goapi/api/attachment/upload", Description: "Upload document attachment — object key is {shopid}/attachments/..., shopid resolved from auth context", Category: "attachment", Source: "goapi", AuthRequired: true,
			RequestBody: &APIRequestBody{ContentType: "multipart/form-data"}},
		{Method: "POST", Path: "/goapi/api/attachment/list", Description: "List document attachments for the caller's shop; URLs are backend proxy URLs only", Category: "attachment", Source: "goapi", AuthRequired: true,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"shopid": "SHOP001", "screen_type": "purchaseorder", "docno": "PO-0001"}}},
		{Method: "POST", Path: "/goapi/api/attachment/delete", Description: "Delete document attachment for the caller's shop", Category: "attachment", Source: "goapi", AuthRequired: true},
		{Method: "GET", Path: "/goapi/api/attachment/download/:id", Description: "Stream attachment through backend after shopid check (no presigned URL redirect)", Category: "attachment", Source: "goapi", AuthRequired: true,
			Parameters: []APIParam{{Name: "id", In: "path", Type: "string", Required: true, Description: "Attachment Mongo ObjectID"}}},

		// ===== Excel Import =====
		{Method: "POST", Path: "/goapi/xlsx/product/start", Description: "Start Excel product import preparation", Category: "import", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "multipart/form-data"}},

		// ===== Chatbot (Gemini AI) =====
		{Method: "POST", Path: "/goapi/api/v1/chatbot/chat-gemini", Description: "Chat with Gemini AI", Category: "chatbot", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type":     "object",
					"required": []string{"shopid", "message"},
					"properties": map[string]interface{}{
						"shopid":          map[string]interface{}{"type": "string", "description": "Shop ID"},
						"message":         map[string]interface{}{"type": "string", "description": "User message (Thai/English)"},
						"conversation_id": map[string]interface{}{"type": "string", "description": "Conversation ID for context continuity"},
					},
				},
				Example: map[string]interface{}{"shopid": "SHOP001", "message": "ยอดขายวันนี้เท่าไร?"}}},
		{Method: "POST", Path: "/goapi/api/v1/chatbot/analyze-document", Description: "Analyze document with Gemini AI", Category: "chatbot", Source: "goapi", AuthRequired: false},

		// ===== Unified API =====
		{Method: "POST", Path: "/goapi/api/v1/unified/query", Description: "Unified query - natural language to SQL/MongoDB query", Category: "unified", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type":     "object",
					"required": []string{"shopid", "query"},
					"properties": map[string]interface{}{
						"shopid": map[string]interface{}{"type": "string", "description": "Shop ID"},
						"query":  map[string]interface{}{"type": "string", "description": "Natural language query (Thai/English)"},
					},
				},
				Example: map[string]interface{}{"shopid": "SHOP001", "query": "ยอดขายเดือนนี้"}}},
		{Method: "GET", Path: "/goapi/api/v1/unified/health", Description: "Unified API health check", Category: "unified", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/v1/unified/cache/stats", Description: "Unified API cache statistics", Category: "unified", Source: "goapi", AuthRequired: false},

		// ===== MCP =====
		{Method: "POST", Path: "/goapi/mcp/invoke", Description: "Invoke MCP tool by name (requires API key)", Category: "mcp", Source: "goapi", AuthRequired: true,
			RequestBody: &APIRequestBody{ContentType: "application/json",
				Schema: map[string]interface{}{
					"type":     "object",
					"required": []string{"tool"},
					"properties": map[string]interface{}{
						"tool":   map[string]interface{}{"type": "string", "description": "MCP tool name (e.g., get_daily_sales, search_products)"},
						"params": map[string]interface{}{"type": "object", "description": "Tool-specific parameters"},
					},
				},
				Example: map[string]interface{}{"tool": "get_daily_sales", "params": map[string]interface{}{"date": "2025-01-01"}}},
			Parameters: []APIParam{{Name: "X-API-Key", In: "header", Type: "string", Required: true, Description: "MCP API Key"}}},
		{Method: "GET", Path: "/goapi/mcp/tools", Description: "List all available MCP tools (no auth)", Category: "mcp", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/mcp/health", Description: "MCP server health check", Category: "mcp", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/mcp/sse", Description: "MCP Server-Sent Events connection (JSON-RPC 2.0)", Category: "mcp", Source: "goapi", AuthRequired: true},
		{Method: "POST", Path: "/goapi/mcp/message", Description: "MCP SSE message handler (JSON-RPC 2.0)", Category: "mcp", Source: "goapi", AuthRequired: true},

		// ===== MCP API Key Management =====
		{Method: "POST", Path: "/goapi/api/mcp/keys", Description: "Create new MCP API key", Category: "mcp-keys", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{
				"shop_id": "SHOP001", "name": "Frontend Dev", "allowed_tools": []string{"get_daily_sales", "search_products"},
			}}},
		{Method: "GET", Path: "/goapi/api/mcp/keys", Description: "List all MCP API keys", Category: "mcp-keys", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/mcp/keys/:id", Description: "Get specific MCP API key", Category: "mcp-keys", Source: "goapi", AuthRequired: false,
			Parameters: []APIParam{{Name: "id", In: "path", Type: "string", Required: true, Description: "API Key ID"}}},
		{Method: "PUT", Path: "/goapi/api/mcp/keys/:id", Description: "Update MCP API key", Category: "mcp-keys", Source: "goapi", AuthRequired: false},
		{Method: "DELETE", Path: "/goapi/api/mcp/keys/:id", Description: "Delete MCP API key", Category: "mcp-keys", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/mcp/keys/:id/export", Description: "Export MCP API key with visible token", Category: "mcp-keys", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/mcp/keys/create-with-export", Description: "Create and immediately export MCP API key", Category: "mcp-keys", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/mcp/audit-logs", Description: "Get MCP tool usage audit logs", Category: "mcp-keys", Source: "goapi", AuthRequired: false},

		// ===== Setup Config =====
		{Method: "POST", Path: "/goapi/api/setup/verify-password", Description: "Verify setup password", Category: "setup", Source: "goapi", AuthRequired: false,
			RequestBody: &APIRequestBody{ContentType: "application/json", Example: map[string]interface{}{"password": "***"}}},
		{Method: "POST", Path: "/goapi/api/setup/change-password", Description: "Change setup password", Category: "setup", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/setup/config/get", Description: "Get bootstrap config (masked secrets)", Category: "setup", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/setup/config/get-raw", Description: "Get raw bootstrap config", Category: "setup", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/setup/config/save", Description: "Save bootstrap config", Category: "setup", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/setup/config/seed", Description: "Seed initial config", Category: "setup", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/setup/test-connection", Description: "Test database connection", Category: "setup", Source: "goapi", AuthRequired: false},
		{Method: "POST", Path: "/goapi/api/setup/create-clickhouse-database", Description: "Create ClickHouse database for shop", Category: "setup", Source: "goapi", AuthRequired: false},
		{Method: "GET", Path: "/goapi/api/setup/client-config", Description: "Get client-side config (public settings)", Category: "setup", Source: "goapi", AuthRequired: false},

		// ===== Deploy =====
		{Method: "POST", Path: "/goapi/api/deploy/backend", Description: "Trigger backend deployment webhook", Category: "deploy", Source: "goapi", AuthRequired: false,
			Parameters: []APIParam{{Name: "token", In: "query", Type: "string", Required: true, Description: "Deploy token"}}},
		{Method: "POST", Path: "/goapi/api/deploy/frontend", Description: "Trigger frontend deployment webhook", Category: "deploy", Source: "goapi", AuthRequired: false,
			Parameters: []APIParam{{Name: "token", In: "query", Type: "string", Required: true, Description: "Deploy token"}}},
		{Method: "GET", Path: "/goapi/api/deploy/status", Description: "Get deployment status", Category: "deploy", Source: "goapi", AuthRequired: false},
	}
}
