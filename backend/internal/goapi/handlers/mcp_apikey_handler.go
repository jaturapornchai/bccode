package handlers

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mcp/auth"
	"smlcloudplatform/internal/goapi/mcp/mongodb"
	"smlcloudplatform/internal/goapi/mcp/tools"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
)

// MCPAPIKeyHandler handles MCP API key management HTTP requests
type MCPAPIKeyHandler struct {
	keysRepo *mongodb.KeysRepository
}

// NewMCPAPIKeyHandler creates a new MCP API key handler
func NewMCPAPIKeyHandler() *MCPAPIKeyHandler {
	return &MCPAPIKeyHandler{
		keysRepo: mongodb.NewKeysRepository(),
	}
}

// CreateAPIKeyRequest represents a request to create an API key
type CreateAPIKeyRequest struct {
	ShopID string   `json:"shop_id"`
	Name string   `json:"name"`
	Description string   `json:"description"`
	AllowedTools []string `json:"allowed_tools"`
	RateLimitPerMinute int      `json:"rate_limit_per_minute"`
	ExpiresAt *string  `json:"expires_at,omitempty"` // Format: YYYY-MM-DD
	CreatedBy string   `json:"created_by"`
}

// CreateAPIKeyResponse represents the response for creating an API key
type CreateAPIKeyResponse struct {
	ID string    `json:"id"`
	Name string    `json:"name"`
	ShopID string    `json:"shop_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt *string   `json:"expires_at,omitempty"`
}

// CreateAPIKeyHandler creates a new MCP API key
func (h *MCPAPIKeyHandler) CreateAPIKeyHandler(c echo.Context) error {
	var req CreateAPIKeyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	// Validate required fields
	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "shop_id is required",
			"code":  "MISSING_SHOP_ID",
		})
	}
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "name is required",
			"code":  "MISSING_NAME",
		})
	}

	// Generate API key
	apiKeyStr := auth.GenerateAPIKey()

	// Parse expires_at if provided
	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", *req.ExpiresAt)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid expires_at format, use YYYY-MM-DD",
				"code":  "INVALID_DATE_FORMAT",
			})
		}
		expiresAt = &t
	}

	// Set default rate limit
	if req.RateLimitPerMinute <= 0 {
		req.RateLimitPerMinute = 600
	}

	// Set default allowed tools if not provided
	if len(req.AllowedTools) == 0 {
		req.AllowedTools = []string{"readonly"}
	}

	// Create API key document
	apiKey := &mongodb.APIKey{
		APIKey:             apiKeyStr,
		ShopID:             req.ShopID,
		Name:               req.Name,
		Description:        req.Description,
		IsActive:           true,
		AllowedTools:       req.AllowedTools,
		RateLimitPerMinute: req.RateLimitPerMinute,
		ExpiresAt:          expiresAt,
		CreatedBy:          req.CreatedBy,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	createdKey, err := h.keysRepo.CreateAPIKey(ctx, apiKey)
	if err != nil {
		logger.Error("Failed to create API key: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create API key",
			"code":  "CREATE_FAILED",
		})
	}

	// Format expires_at for response
	var expiresAtStr *string
	if createdKey.ExpiresAt != nil {
		str := createdKey.ExpiresAt.Format("2006-01-02")
		expiresAtStr = &str
	}

	return c.JSON(http.StatusCreated, CreateAPIKeyResponse{
		ID:        createdKey.ID.Hex(),
		Name:      createdKey.Name,
		ShopID:    createdKey.ShopID,
		CreatedAt: createdKey.CreatedAt,
		ExpiresAt: expiresAtStr,
	})
}

// ListAPIKeysHandler lists all API keys for a shop
type ListAPIKeysRequest struct {
	ShopID string `query:"shop_id"`
}

func (h *MCPAPIKeyHandler) ListAPIKeysHandler(c echo.Context) error {
	shopID := c.QueryParam("shop_id")
	if shopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "shop_id is required",
			"code":  "MISSING_SHOP_ID",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	apiKeys, err := h.keysRepo.GetAPIKeysByShop(ctx, shopID)
	if err != nil {
		logger.Error("Failed to get API keys: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get API keys",
			"code":  "FETCH_FAILED",
		})
	}

	// Filter out inactive keys in the response (or include them with a flag)
	var response []map[string]interface{}
	for _, key := range apiKeys {
		keyData := map[string]interface{}{
			"id":                    key.ID.Hex(),
			"name":                  key.Name,
			"description":           key.Description,
			"shop_id":               key.ShopID,
			"is_active":             key.IsActive,
			"allowed_tools":         key.AllowedTools,
			"rate_limit_per_minute": key.RateLimitPerMinute,
			"created_at":            key.CreatedAt,
			"created_by":            key.CreatedBy,
		}


		if key.ExpiresAt != nil {
			keyData["expires_at"] = key.ExpiresAt.Format("2006-01-02")
		}
		if key.LastUsedAt != nil {
			keyData["last_used_at"] = key.LastUsedAt.Format("2006-01-02 15:04:05")
		}

		response = append(response, keyData)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    response,
		"count":   len(response),
	})
}

// DeleteAPIKeyRequest represents a request to delete an API key
type DeleteAPIKeyRequest struct {
	KeyID string `json:"key_id"`
}

// DeleteAPIKeyHandler soft deletes an API key
func (h *MCPAPIKeyHandler) DeleteAPIKeyHandler(c echo.Context) error {
	keyID := c.Param("id")
	if keyID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "key_id is required",
			"code":  "MISSING_KEY_ID",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if key exists
	existingKey, err := h.keysRepo.GetAPIKeyByID(ctx, keyID)
	if err != nil {
		logger.Error("Failed to get API key: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get API key",
			"code":  "FETCH_FAILED",
		})
	}

	if existingKey == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "API key not found",
			"code":  "KEY_NOT_FOUND",
		})
	}

	// Delete the key (soft delete)
	if err := h.keysRepo.DeleteAPIKey(ctx, keyID); err != nil {
		logger.Error("Failed to delete API key: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete API key",
			"code":  "DELETE_FAILED",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "API key deleted successfully",
	})
}

// UpdateAPIKeyRequest represents a request to update an API key
type UpdateAPIKeyRequest struct {
	Name *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
	AllowedTools []string `json:"allowed_tools,omitempty"`
	RateLimitPerMinute *int     `json:"rate_limit_per_minute,omitempty"`
}

// UpdateAPIKeyHandler updates an API key
func (h *MCPAPIKeyHandler) UpdateAPIKeyHandler(c echo.Context) error {
	keyID := c.Param("id")
	if keyID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "key_id is required",
			"code":  "MISSING_KEY_ID",
		})
	}

	var req UpdateAPIKeyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	// Build update document
	updates := bson.M{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.AllowedTools != nil {
		updates["allowed_tools"] = req.AllowedTools
	}
	if req.RateLimitPerMinute != nil {
		updates["rate_limit_per_minute"] = *req.RateLimitPerMinute
	}

	if len(updates) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No fields to update",
			"code":  "NO_UPDATES",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if key exists
	existingKey, err := h.keysRepo.GetAPIKeyByID(ctx, keyID)
	if err != nil {
		logger.Error("Failed to get API key: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get API key",
			"code":  "FETCH_FAILED",
		})
	}

	if existingKey == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "API key not found",
			"code":  "KEY_NOT_FOUND",
		})
	}

	// Update the key
	if err := h.keysRepo.UpdateAPIKey(ctx, keyID, updates); err != nil {
		logger.Error("Failed to update API key: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update API key",
			"code":  "UPDATE_FAILED",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "API key updated successfully",
	})
}

// GetAPIKeyHandler gets a single API key by ID
func (h *MCPAPIKeyHandler) GetAPIKeyHandler(c echo.Context) error {
	keyID := c.Param("id")
	if keyID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "key_id is required",
			"code":  "MISSING_KEY_ID",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	apiKey, err := h.keysRepo.GetAPIKeyByID(ctx, keyID)
	if err != nil {
		logger.Error("Failed to get API key: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get API key",
			"code":  "FETCH_FAILED",
		})
	}

	if apiKey == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "API key not found",
			"code":  "KEY_NOT_FOUND",
		})
	}

	response := map[string]interface{}{
		"id":                    apiKey.ID.Hex(),
		"name":                  apiKey.Name,
		"description":           apiKey.Description,
		"shop_id":               apiKey.ShopID,
		"is_active":             apiKey.IsActive,
		"allowed_tools":         apiKey.AllowedTools,
		"rate_limit_per_minute": apiKey.RateLimitPerMinute,
		"created_at":            apiKey.CreatedAt,
		"created_by":            apiKey.CreatedBy,
	}

	if apiKey.ExpiresAt != nil {
		response["expires_at"] = apiKey.ExpiresAt.Format("2006-01-02")
	}
	if apiKey.LastUsedAt != nil {
		response["last_used_at"] = apiKey.LastUsedAt.Format("2006-01-02 15:04:05")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    response,
	})
}

// GetAuditLogsHandler gets audit logs for a shop
type GetAuditLogsRequest struct {
	ShopID string `query:"shop_id"`
	Limit  int64  `query:"limit"`
	Skip   int64  `query:"skip"`
}

func (h *MCPAPIKeyHandler) GetAuditLogsHandler(c echo.Context) error {
	shopID := c.QueryParam("shop_id")
	if shopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "shop_id is required",
			"code":  "MISSING_SHOP_ID",
		})
	}

	limit := int64(50)
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 64); err == nil {
			limit = parsed
		}
	}

	skip := int64(0)
	if s := c.QueryParam("skip"); s != "" {
		if parsed, err := strconv.ParseInt(s, 10, 64); err == nil {
			skip = parsed
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logs, err := h.keysRepo.GetAuditLogsByShop(ctx, shopID, limit, skip)
	if err != nil {
		logger.Error("Failed to get audit logs: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get audit logs",
			"code":  "FETCH_FAILED",
		})
	}

	var response []map[string]interface{}
	for _, log := range logs {
		logData := map[string]interface{}{
			"id":                log.ID.Hex(),
			"api_key_id":        log.APIKeyID.Hex(),
			"shop_id":           log.ShopID,
			"tool_name":         log.ToolName,
			"request_params":    log.RequestParams,
			"response_status":   log.ResponseStatus,
			"execution_time_ms": log.ExecutionTimeMs,
			"created_at":        log.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if log.ErrorMessage != "" {
			logData["error_message"] = log.ErrorMessage
		}

		response = append(response, logData)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    response,
		"count":   len(response),
		"limit":   limit,
		"skip":    skip,
	})
}

// ExportAPIKeyHandler export token เป็น JSON config สำหรับ Claude Desktop/Code
func (h *MCPAPIKeyHandler) ExportAPIKeyHandler(c echo.Context) error {
	keyID := c.Param("id")
	if keyID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "key_id is required",
			"code":  "MISSING_KEY_ID",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	apiKey, err := h.keysRepo.GetAPIKeyByID(ctx, keyID)
	if err != nil {
		logger.Error("Failed to get API key for export: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get API key",
			"code":  "FETCH_FAILED",
		})
	}

	if apiKey == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "API key not found",
			"code":  "KEY_NOT_FOUND",
		})
	}

	// ดึง server URL จาก environment หรือใช้ค่า default
	serverURL := os.Getenv("MCP_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8888"
	}

	// สร้าง export data
	exportData := tools.GenerateTokenExport(
		apiKey.APIKey,
		apiKey.ShopID,
		apiKey.Name,
		serverURL,
		apiKey.ExpiresAt,
		apiKey.CreatedAt,
	)

	// บันทึก token file (background)
	go func() {
		if saveErr := tools.SaveTokenFile(apiKey.ShopID, keyID, exportData); saveErr != nil {
			logger.Error("Failed to save token file: %v", saveErr)
		}
	}()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    exportData,
	})
}

// CreateAPIKeyWithExportHandler สร้าง API key แล้ว export config กลับมาด้วย (สำหรับ Flutter)
func (h *MCPAPIKeyHandler) CreateAPIKeyWithExportHandler(c echo.Context) error {
	var req CreateAPIKeyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "shop_id is required",
			"code":  "MISSING_SHOP_ID",
		})
	}
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "name is required",
			"code":  "MISSING_NAME",
		})
	}

	// Generate API key
	apiKeyStr := auth.GenerateAPIKey()

	// Parse expires_at
	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", *req.ExpiresAt)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid expires_at format, use YYYY-MM-DD",
				"code":  "INVALID_DATE_FORMAT",
			})
		}
		// ตั้งเวลาเป็นสิ้นวัน
		t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		expiresAt = &t
	}

	if req.RateLimitPerMinute <= 0 {
		req.RateLimitPerMinute = 600
	}

	// Default = readonly (ถ้าต้องการ write ต้องระบุ ["*"] มาเอง)
	if len(req.AllowedTools) == 0 {
		req.AllowedTools = []string{"readonly"}
	}

	apiKey := &mongodb.APIKey{
		APIKey:             apiKeyStr,
		ShopID:             req.ShopID,
		Name:               req.Name,
		Description:        req.Description,
		IsActive:           true,
		AllowedTools:       req.AllowedTools,
		RateLimitPerMinute: req.RateLimitPerMinute,
		ExpiresAt:          expiresAt,
		CreatedBy:          req.CreatedBy,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	createdKey, err := h.keysRepo.CreateAPIKey(ctx, apiKey)
	if err != nil {
		logger.Error("Failed to create API key: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create API key",
			"code":  "CREATE_FAILED",
		})
	}

	// สร้าง export data
	serverURL := os.Getenv("MCP_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8888"
	}

	exportData := tools.GenerateTokenExport(
		createdKey.APIKey,
		createdKey.ShopID,
		createdKey.Name,
		serverURL,
		createdKey.ExpiresAt,
		createdKey.CreatedAt,
	)

	// บันทึก token file (background)
	go func() {
		if saveErr := tools.SaveTokenFile(createdKey.ShopID, createdKey.ID.Hex(), exportData); saveErr != nil {
			logger.Error("Failed to save token file: %v", saveErr)
		}
	}()

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"id":         createdKey.ID.Hex(),
			"name":       createdKey.Name,
			"shop_id":    createdKey.ShopID,
			"created_at": createdKey.CreatedAt,
			"expires_at": func() *string {
				if createdKey.ExpiresAt != nil {
					s := createdKey.ExpiresAt.Format("2006-01-02")
					return &s
				}
				return nil
			}(),
			"export": exportData,
		},
	})
}

// ToolInfo describes a single MCP tool for the frontend catalog
type ToolInfo struct {
	Name string `json:"name"`
	Category string `json:"category"`
	IsWrite bool   `json:"is_write"`
}

// GetAvailableToolsHandler returns all MCP tools grouped by category + write flag
// Frontend ใช้ endpoint นี้แทนการ hardcode รายการ tools
func (h *MCPAPIKeyHandler) GetAvailableToolsHandler(c echo.Context) error {
	// Tool catalog — source of truth สำหรับ frontend
	catalog := []ToolInfo{
		// Products
		{"search_products", "Products", false},
		// Sales
		{"get_daily_sales", "Sales", false},
		{"get_sales_by_date_range", "Sales", false},
		{"get_top_selling_products", "Sales", false},
		{"get_sales_by_seller", "Sales", false},
		{"get_monthly_summary", "Sales", false},
		// Dashboard
		{"get_dashboard_kpis", "Dashboard", false},
		{"get_business_health", "Dashboard", false},
		// Financial
		{"get_profit_analysis", "Financial", false},
		{"get_accounts_receivable", "Financial", false},
		{"get_accounts_payable", "Financial", false},
		{"get_cash_flow", "Financial", false},
		// Inventory
		{"get_inventory_value", "Inventory", false},
		{"get_low_stock_alerts", "Inventory", false},
		{"get_dead_stock", "Inventory", false},
		{"get_inventory_turnover", "Inventory", false},
		// Customers
		{"get_top_customers", "Customers", false},
		{"get_customer_growth", "Customers", false},
		{"get_customer_segments", "Customers", false},
		// Comparison
		{"get_yoy_comparison", "Comparison", false},
		{"get_mom_comparison", "Comparison", false},
		// Database (PostgreSQL)
		{"get_database_schema", "Database", false},
		{"execute_query", "Database", false},
		{"get_table_sample", "Database", false},
		// Database (MongoDB)
		{"query_mongodb", "Database", false},
		{"list_mongodb_collections", "Database", false},
		{"aggregate_mongodb", "Database", false},
		// Database (ClickHouse)
		{"query_clickhouse", "Database", false},
		{"list_clickhouse_tables", "Database", false},
		// API Development
		{"list_api_endpoints", "API Development", false},
		{"get_api_spec", "API Development", false},
		{"get_api_example", "API Development", false},
		// Schema & Enums
		{"list_enums", "Schema", false},
		{"get_model_schema", "Schema", false},
		// Unit of Measure (readonly)
		{"list_units", "Unit of Measure", false},
		{"get_unit_schema", "Unit of Measure", false},
		// Unit of Measure (write)
		{"create_unit", "Unit of Measure", true},
		{"create_units", "Unit of Measure", true},
		{"update_unit", "Unit of Measure", true},
		{"delete_unit", "Unit of Measure", true},
		{"delete_units", "Unit of Measure", true},
	}

	// Verify write flags match permissions.go
	for i := range catalog {
		catalog[i].IsWrite = tools.IsWriteTool(catalog[i].Name)
	}

	// Group by category
	categories := make(map[string][]ToolInfo)
	var categoryOrder []string
	seen := make(map[string]bool)
	for _, t := range catalog {
		if !seen[t.Category] {
			categoryOrder = append(categoryOrder, t.Category)
			seen[t.Category] = true
		}
		categories[t.Category] = append(categories[t.Category], t)
	}

	// Build ordered response
	type CategoryGroup struct {
		Category string     `json:"category"`
		Tools []ToolInfo `json:"tools"`
	}
	var groups []CategoryGroup
	for _, cat := range categoryOrder {
		groups = append(groups, CategoryGroup{Category: cat, Tools: categories[cat]})
	}

	// Permission presets
	presets := []map[string]interface{}{
		{
			"value":       "readonly",
			"label":       "Readonly",
			"description": "ดูข้อมูลอย่างเดียว (เหมาะกับ Claude Desktop, frontend dev)",
			"allowed_tools": []string{"readonly"},
		},
		{
			"value":       "developer",
			"label":       "Developer",
			"description": "ทุก tool รวม create/update/delete (เหมาะกับ backend dev)",
			"allowed_tools": []string{"*"},
		},
		{
			"value":       "custom",
			"label":       "Custom",
			"description": "เลือก tools เองทีละตัว",
			"allowed_tools": nil,
		},
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":       true,
		"categories":    groups,
		"presets":       presets,
		"total_tools":   len(catalog),
		"write_tools":   len(tools.WriteTools),
		"readonly_tools": len(catalog) - len(tools.WriteTools),
	})
}
