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
	HoldingCode        string   `json:"holdingcode"`
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	AllowedTools       []string `json:"allowedtools"`
	RateLimitPerMinute int      `json:"ratelimitperminute"`
	ExpiresAt          *string  `json:"expiresat,omitempty"` // Format: YYYY-MM-DD
	CreatedBy          string   `json:"createdby"`
}

// CreateAPIKeyResponse represents the response for creating an API key
type CreateAPIKeyResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	HoldingCode string    `json:"holdingcode"`
	CreatedAt   time.Time `json:"createdat"`
	ExpiresAt   *string   `json:"expiresat,omitempty"`
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
	if req.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode is required",
			"code":  "MISSING_HOLDING_CODE",
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

	// Parse expiresat if provided
	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", *req.ExpiresAt)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid expiresat format, use YYYY-MM-DD",
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
		HoldingCode:        req.HoldingCode,
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

	// Format expiresat for response
	var expiresAtStr *string
	if createdKey.ExpiresAt != nil {
		str := createdKey.ExpiresAt.Format("2006-01-02")
		expiresAtStr = &str
	}

	return c.JSON(http.StatusCreated, CreateAPIKeyResponse{
		ID:          createdKey.ID.Hex(),
		Name:        createdKey.Name,
		HoldingCode: createdKey.HoldingCode,
		CreatedAt:   createdKey.CreatedAt,
		ExpiresAt:   expiresAtStr,
	})
}

// ListAPIKeysHandler lists all API keys for a shop
type ListAPIKeysRequest struct {
	HoldingCode string `query:"holdingcode"`
}

func (h *MCPAPIKeyHandler) ListAPIKeysHandler(c echo.Context) error {
	holdingCode := c.QueryParam("holdingcode")
	if holdingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode is required",
			"code":  "MISSING_HOLDING_CODE",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	apiKeys, err := h.keysRepo.GetAPIKeysByShop(ctx, holdingCode)
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
			"holdingcode":           key.HoldingCode,
			"isactive":              key.IsActive,
			"allowedtools":          key.AllowedTools,
			"ratelimitperminute":    key.RateLimitPerMinute,
			"createdat":             key.CreatedAt,
			"createdby":             key.CreatedBy,
		}

		if key.ExpiresAt != nil {
			keyData["expiresat"] = key.ExpiresAt.Format("2006-01-02")
		}
		if key.LastUsedAt != nil {
			keyData["lastusedat"] = key.LastUsedAt.Format("2006-01-02 15:04:05")
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
	KeyID string `json:"keyid"`
}

// DeleteAPIKeyHandler soft deletes an API key
func (h *MCPAPIKeyHandler) DeleteAPIKeyHandler(c echo.Context) error {
	keyID := c.Param("id")
	if keyID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "keyid is required",
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
	Name               *string  `json:"name,omitempty"`
	Description        *string  `json:"description,omitempty"`
	IsActive           *bool    `json:"isactive,omitempty"`
	AllowedTools       []string `json:"allowedtools,omitempty"`
	RateLimitPerMinute *int     `json:"ratelimitperminute,omitempty"`
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
		updates["isactive"] = *req.IsActive
	}
	if req.AllowedTools != nil {
		updates["allowedtools"] = req.AllowedTools
	}
	if req.RateLimitPerMinute != nil {
		updates["ratelimitperminute"] = *req.RateLimitPerMinute
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
			"error": "keyid is required",
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
		"holdingcode":           apiKey.HoldingCode,
		"isactive":              apiKey.IsActive,
		"allowedtools":          apiKey.AllowedTools,
		"ratelimitperminute":    apiKey.RateLimitPerMinute,
		"createdat":             apiKey.CreatedAt,
		"createdby":             apiKey.CreatedBy,
	}

	if apiKey.ExpiresAt != nil {
		response["expiresat"] = apiKey.ExpiresAt.Format("2006-01-02")
	}
	if apiKey.LastUsedAt != nil {
		response["lastusedat"] = apiKey.LastUsedAt.Format("2006-01-02 15:04:05")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    response,
	})
}

// GetAuditLogsHandler gets audit logs for a shop
type GetAuditLogsRequest struct {
	HoldingCode string `query:"holdingcode"`
	Limit       int64  `query:"limit"`
	Skip        int64  `query:"skip"`
}

func (h *MCPAPIKeyHandler) GetAuditLogsHandler(c echo.Context) error {
	holdingCode := c.QueryParam("holdingcode")
	if holdingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode is required",
			"code":  "MISSING_HOLDING_CODE",
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

	logs, err := h.keysRepo.GetAuditLogsByShop(ctx, holdingCode, limit, skip)
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
			"apikeyid":         log.APIKeyID.Hex(),
			"holdingcode":       log.HoldingCode,
			"toolname":         log.ToolName,
			"requestparams":    log.RequestParams,
			"responsestatus":   log.ResponseStatus,
			"executiontimems":  log.ExecutionTimeMs,
			"createdat":         log.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if log.ErrorMessage != "" {
			logData["errormessage"] = log.ErrorMessage
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
		apiKey.HoldingCode,
		apiKey.Name,
		serverURL,
		apiKey.ExpiresAt,
		apiKey.CreatedAt,
	)

	// บันทึก token file (background)
	go func() {
		if saveErr := tools.SaveTokenFile(apiKey.HoldingCode, keyID, exportData); saveErr != nil {
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

	if req.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode is required",
			"code":  "MISSING_HOLDING_CODE",
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

	// Parse expiresat
	var expiresAt *time.Time
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", *req.ExpiresAt)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid expiresat format, use YYYY-MM-DD",
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
		HoldingCode:        req.HoldingCode,
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
		createdKey.HoldingCode,
		createdKey.Name,
		serverURL,
		createdKey.ExpiresAt,
		createdKey.CreatedAt,
	)

	// บันทึก token file (background)
	go func() {
		if saveErr := tools.SaveTokenFile(createdKey.HoldingCode, createdKey.ID.Hex(), exportData); saveErr != nil {
			logger.Error("Failed to save token file: %v", saveErr)
		}
	}()

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"id":          createdKey.ID.Hex(),
			"name":        createdKey.Name,
			"holdingcode": createdKey.HoldingCode,
			"createdat":   createdKey.CreatedAt,
			"expiresat": func() *string {
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
	Name     string `json:"name"`
	Category string `json:"category"`
	IsWrite  bool   `json:"iswrite"`
}

// GetAvailableToolsHandler returns all MCP tools grouped by category + write flag
// Frontend ใช้ endpoint นี้แทนการ hardcode รายการ tools
func (h *MCPAPIKeyHandler) GetAvailableToolsHandler(c echo.Context) error {
	// Tool catalog — source of truth สำหรับ frontend
	catalog := []ToolInfo{
		// Products
		{"searchproducts", "Products", false},
		// Sales
		{"getdailysales", "Sales", false},
		{"getsalesbydaterange", "Sales", false},
		{"gettopsellingproducts", "Sales", false},
		{"getsalesbyseller", "Sales", false},
		{"getmonthlysummary", "Sales", false},
		// Dashboard
		{"getdashboardkpis", "Dashboard", false},
		{"getbusinesshealth", "Dashboard", false},
		// Financial
		{"getprofitanalysis", "Financial", false},
		{"getaccountsreceivable", "Financial", false},
		{"getaccountspayable", "Financial", false},
		{"getcashflow", "Financial", false},
		// Inventory
		{"getinventoryvalue", "Inventory", false},
		{"getlowstockalerts", "Inventory", false},
		{"getdeadstock", "Inventory", false},
		{"getinventoryturnover", "Inventory", false},
		// Customers
		{"gettopcustomers", "Customers", false},
		{"getcustomergrowth", "Customers", false},
		{"getcustomersegments", "Customers", false},
		// Comparison
		{"getyoycomparison", "Comparison", false},
		{"getmomcomparison", "Comparison", false},
		// Database (PostgreSQL)
		{"getdatabaseschema", "Database", false},
		{"executequery", "Database", false},
		{"gettablesample", "Database", false},
		// Database (MongoDB)
		{"querymongodb", "Database", false},
		{"listmongodbcollections", "Database", false},
		{"aggregatemongodb", "Database", false},
		// Database (ClickHouse)
		{"queryclickhouse", "Database", false},
		{"listclickhousetables", "Database", false},
		// API Development
		{"listapiendpoints", "API Development", false},
		{"getapispec", "API Development", false},
		{"getapiexample", "API Development", false},
		// Schema & Enums
		{"listenums", "Schema", false},
		{"getmodelschema", "Schema", false},
		// Unit of Measure (readonly)
		{"listunits", "Unit of Measure", false},
		{"getunitschema", "Unit of Measure", false},
		// Unit of Measure (write)
		{"createunit", "Unit of Measure", true},
		{"createunits", "Unit of Measure", true},
		{"updateunit", "Unit of Measure", true},
		{"deleteunit", "Unit of Measure", true},
		{"deleteunits", "Unit of Measure", true},
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
		Tools    []ToolInfo `json:"tools"`
	}
	var groups []CategoryGroup
	for _, cat := range categoryOrder {
		groups = append(groups, CategoryGroup{Category: cat, Tools: categories[cat]})
	}

	// Permission presets
	presets := []map[string]interface{}{
		{
			"value":        "readonly",
			"label":        "Readonly",
			"description":  "ดูข้อมูลอย่างเดียว (เหมาะกับ Claude Desktop, frontend dev)",
			"allowedtools": []string{"readonly"},
		},
		{
			"value":        "developer",
			"label":        "Developer",
			"description":  "ทุก tool รวม create/update/delete (เหมาะกับ backend dev)",
			"allowedtools": []string{"*"},
		},
		{
			"value":        "custom",
			"label":        "Custom",
			"description":  "เลือก tools เองทีละตัว",
			"allowedtools": nil,
		},
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":       true,
		"categories":    groups,
		"presets":       presets,
		"totaltools":    len(catalog),
		"writetools":    len(tools.WriteTools),
		"readonlytools": len(catalog) - len(tools.WriteTools),
	})
}
