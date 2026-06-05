package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mcp/mongodb"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// MCP Protocol Types
const (
	MCPProtocolVersion = "2024-11-05"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP Tool Definition
type MCPToolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputschema"`
}

// SSE Session represents an active SSE connection
type SSESession struct {
	ID        string
	APIKey    *mongodb.APIKey
	IsDevMode bool // true = dev endpoint → เห็นทุก tools, false = general → ซ่อน dev tools
	eventsCh  chan string
	Done      chan struct{}
	mu        sync.Mutex
}

// devToolNames — tools ที่แสดงเฉพาะใน dev mode
// general endpoint จะไม่เห็น tools เหล่านี้
var devToolNames = map[string]bool{
	"getdatabaseschema":      true,
	"executequery":           true,
	"gettablesample":         true,
	"querymongodb":           true,
	"listmongodbcollections": true,
	"aggregatemongodb":       true,
	"queryclickhouse":        true,
	"listclickhousetables":   true,
	"executepgcommand":       true,
	"executechcommand":       true,
	"listapiendpoints":       true,
	"getapispec":             true,
	"getapiexample":          true,
	"listenums":              true,
	"getmodelschema":         true,
	"rebuildproducts":        true,
}

// Active SSE sessions
var (
	sseSessions   = make(map[string]*SSESession)
	sseSessionsMu sync.RWMutex
)

// RegisterSSERoutes registers SSE routes for MCP (standalone mode)
func (s *MCPServer) RegisterSSERoutes(e *echo.Echo) {
	s.ssePrefix = "" // standalone mode — no prefix
	e.GET("/mcp/sse", s.HandleSSE)
	e.POST("/mcp/message", s.HandleMessage)
}

// HandleSSEDev handles SSE connections for dev mode (all tools visible)
func (s *MCPServer) HandleSSEDev(c echo.Context) error {
	return s.handleSSEInternal(c, true)
}

// HandleSSE handles SSE connections from MCP clients (general tools only)
func (s *MCPServer) HandleSSE(c echo.Context) error {
	return s.handleSSEInternal(c, false)
}

// handleSSEInternal — shared SSE handler logic
func (s *MCPServer) handleSSEInternal(c echo.Context, devMode bool) error {
	// Get API key from query param or header
	apiKeyStr := c.QueryParam("apikey")
	if apiKeyStr == "" {
		apiKeyStr = c.Request().Header.Get("X-API-Key")
	}
	if apiKeyStr == "" {
		authHeader := c.Request().Header.Get("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			apiKeyStr = authHeader[7:]
		}
	}
	if apiKeyStr == "" {
		apiKeyStr = c.Request().Header.Get("MCP-API-Key")
	}

	if apiKeyStr == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "API key required",
		})
	}

	// Validate API key
	apiKey, err := s.keysRepo.GetAPIKeyByKey(c.Request().Context(), apiKeyStr)
	if err != nil || apiKey == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid API key",
		})
	}

	// Set SSE headers
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("X-Accel-Buffering", "no")

	// Create session
	sessionID := uuid.New().String()
	session := &SSESession{
		ID:        sessionID,
		APIKey:    apiKey,
		IsDevMode: devMode,
		eventsCh:  make(chan string, 100),
		Done:      make(chan struct{}),
	}

	// Store session
	sseSessionsMu.Lock()
	sseSessions[sessionID] = session
	sseSessionsMu.Unlock()

	defer func() {
		sseSessionsMu.Lock()
		delete(sseSessions, sessionID)
		sseSessionsMu.Unlock()
		close(session.Done)
	}()

	logger.Info("MCP SSE session started: %s (shop: %s)", sessionID, apiKey.HoldingCode)

	// Get the underlying writer - try to get Flusher
	w := c.Response().Writer
	flusher, hasFlusher := w.(http.Flusher)

	// Helper function to flush if supported
	flushWriter := func() {
		if hasFlusher {
			flusher.Flush()
		}
	}

	// Send endpoint event with session ID for message posting
	// ใช้ ssePrefix เพื่อรองรับ embedded mode (เช่น /goapi/mcp/message)
	messagePath := "/mcp/message"
	if devMode {
		messagePath = "/mcp/dev/message"
	}
	messageEndpoint := fmt.Sprintf("%s%s?sessionid=%s", s.ssePrefix, messagePath, sessionID)

	// Write directly to the underlying writer
	fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", messageEndpoint)
	flushWriter()

	// Keep connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request().Context().Done():
			logger.Info("MCP SSE session closed: %s", sessionID)
			return nil
		case <-session.Done:
			return nil
		case event := <-session.eventsCh:
			fmt.Fprint(w, event)
			flushWriter()
		case <-ticker.C:
			fmt.Fprintf(w, "event: ping\ndata: %s\n\n", time.Now().Format(time.RFC3339))
			flushWriter()
		}
	}
}

// HandleMessage handles JSON-RPC messages from MCP clients
func (s *MCPServer) HandleMessage(c echo.Context) error {
	sessionID := c.QueryParam("sessionid")
	if sessionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "sessionid required",
		})
	}

	// Get session
	sseSessionsMu.RLock()
	session, exists := sseSessions[sessionID]
	sseSessionsMu.RUnlock()

	if !exists {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Session not found",
		})
	}

	// Parse JSON-RPC request
	var req JSONRPCRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32700,
				Message: "Parse error",
			},
		})
	}

	// Handle the request
	response := s.handleJSONRPCRequest(c.Request().Context(), session, &req)

	// Send response via SSE channel
	responseJSON, _ := json.Marshal(response)
	event := fmt.Sprintf("event: message\ndata: %s\n\n", string(responseJSON))

	select {
	case session.eventsCh <- event:
	default:
		logger.Warn("SSE event channel full for session %s", sessionID)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "accepted",
	})
}

// handleJSONRPCRequest processes JSON-RPC requests
func (s *MCPServer) handleJSONRPCRequest(ctx context.Context, session *SSESession, req *JSONRPCRequest) *JSONRPCResponse {
	response := &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		response.Result = s.handleInitialize(req.Params)
	case "initialized":
		response.Result = map[string]interface{}{}
	case "tools/list":
		response.Result = s.handleToolsListFiltered(session)
	case "tools/call":
		result, err := s.handleToolCall(ctx, session, req.Params)
		if err != nil {
			response.Error = &JSONRPCError{
				Code:    -32000,
				Message: err.Error(),
			}
		} else {
			response.Result = result
		}
	case "ping":
		response.Result = map[string]interface{}{}
	default:
		response.Error = &JSONRPCError{
			Code:    -32601,
			Message: fmt.Sprintf("Method not found: %s", req.Method),
		}
	}

	return response
}

// handleInitialize handles the initialize request
func (s *MCPServer) handleInitialize(params json.RawMessage) map[string]interface{} {
	return map[string]interface{}{
		"protocolVersion": MCPProtocolVersion,
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "BC Cloud MCP Server",
			"version": "1.0.0",
		},
	}
}

// handleToolsList returns the list of available tools
func (s *MCPServer) handleToolsList() map[string]interface{} {
	tools := []MCPToolDef{
		{
			Name:        "searchproducts",
			Description: "Search products from the PostgreSQL projection/read model built from MongoDB operational product data. Returns product info, units, prices, and stock balance by warehouse/location in formatted word (e.g., '1 กล่อง x 2 โหล x 3 ชิ้น')",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":        "string",
						"description": "Search keyword (Thai or English). Will be tokenized automatically.",
					},
					"whcode": map[string]interface{}{
						"type":        "string",
						"description": "Optional warehouse code to filter stock balance",
					},
					"locationcode": map[string]interface{}{
						"type":        "string",
						"description": "Optional location code to filter stock balance",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Maximum number of products to return (default: 50, max: 200)",
					},
					"includebalance": map[string]interface{}{
						"type":        "boolean",
						"description": "Include stock balance information (default: true)",
					},
				},
				"required": []string{"keyword"},
			},
		},
		{
			Name:        "getdailysales",
			Description: "Get daily sales summary for a specific date",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"date": map[string]interface{}{
						"type":        "string",
						"description": "Date in YYYY-MM-DD format",
					},
					"branchcode": map[string]interface{}{
						"type":        "string",
						"description": "Optional branch/warehouse code filter",
					},
				},
				"required": []string{"date"},
			},
		},
		{
			Name:        "getsalesbydaterange",
			Description: "Get sales data grouped by day/week/month for a date range",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"fromdate": map[string]interface{}{
						"type":        "string",
						"description": "Start date in YYYY-MM-DD format",
					},
					"todate": map[string]interface{}{
						"type":        "string",
						"description": "End date in YYYY-MM-DD format",
					},
					"groupby": map[string]interface{}{
						"type":        "string",
						"description": "Group by: day, week, month (default: day)",
						"enum":        []string{"day", "week", "month"},
					},
					"branchcode": map[string]interface{}{
						"type":        "string",
						"description": "Optional branch/warehouse code filter",
					},
				},
				"required": []string{"fromdate", "todate"},
			},
		},
		{
			Name:        "gettopsellingproducts",
			Description: "Get top selling products for a date range",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"fromdate": map[string]interface{}{
						"type":        "string",
						"description": "Start date in YYYY-MM-DD format",
					},
					"todate": map[string]interface{}{
						"type":        "string",
						"description": "End date in YYYY-MM-DD format",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of products to return (default: 10, max: 100)",
					},
					"branchcode": map[string]interface{}{
						"type":        "string",
						"description": "Optional branch/warehouse code filter",
					},
				},
				"required": []string{"fromdate", "todate"},
			},
		},
		{
			Name:        "getsalesbyseller",
			Description: "Get sales data grouped by seller/salesperson",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"fromdate": map[string]interface{}{
						"type":        "string",
						"description": "Start date in YYYY-MM-DD format",
					},
					"todate": map[string]interface{}{
						"type":        "string",
						"description": "End date in YYYY-MM-DD format",
					},
				},
				"required": []string{"fromdate", "todate"},
			},
		},
		{
			Name:        "getmonthlysummary",
			Description: "Get monthly sales summary with comparison to previous month",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"year": map[string]interface{}{
						"type":        "number",
						"description": "Year (e.g., 2025). Defaults to current year if not specified.",
					},
					"month": map[string]interface{}{
						"type":        "number",
						"description": "Month (1-12). Defaults to current month if not specified.",
					},
				},
			},
		},
		// Dashboard Tools
		{
			Name:        "getdashboardkpis",
			Description: "Get comprehensive KPI dashboard from processed relational projections. Includes sales, orders, profit, customers, inventory metrics.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"period": map[string]interface{}{
						"type":        "string",
						"description": "Period: today, thisweek, thismonth, thisyear (default: thismonth)",
						"enum":        []string{"today", "thisweek", "thismonth", "thisyear"},
					},
				},
			},
		},
		{
			Name:        "getbusinesshealth",
			Description: "Get overall business health score (0-100) with key indicators and recommendations.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		// Financial Tools
		{
			Name:        "getprofitanalysis",
			Description: "Get detailed profit analysis with revenue breakdown, costs, and margins.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"fromdate": map[string]interface{}{
						"type":        "string",
						"description": "Start date in YYYY-MM-DD format",
					},
					"todate": map[string]interface{}{
						"type":        "string",
						"description": "End date in YYYY-MM-DD format",
					},
				},
			},
		},
		{
			Name:        "getaccountsreceivable",
			Description: "Get accounts receivable summary with aging buckets and top debtors.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "getaccountspayable",
			Description: "Get accounts payable summary with aging buckets and top creditors.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "getcashflow",
			Description: "Get cash flow analysis showing inflows, outflows, and net position.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"fromdate": map[string]interface{}{
						"type":        "string",
						"description": "Start date in YYYY-MM-DD format",
					},
					"todate": map[string]interface{}{
						"type":        "string",
						"description": "End date in YYYY-MM-DD format",
					},
				},
			},
		},
		// Inventory Tools
		{
			Name:        "getinventoryvalue",
			Description: "Get inventory valuation from PostgreSQL relational projections, with breakdown by category and warehouse.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"whcode": map[string]interface{}{
						"type":        "string",
						"description": "Optional warehouse code to filter",
					},
				},
			},
		},
		{
			Name:        "getlowstockalerts",
			Description: "Get products that are below minimum stock level or out of stock.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"threshold": map[string]interface{}{
						"type":        "number",
						"description": "Stock threshold to consider low (default: 10)",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of alerts to return (default: 50)",
					},
				},
			},
		},
		{
			Name:        "getdeadstock",
			Description: "Get products with no movement for specified days (slow-moving/dead stock).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"daysnomovement": map[string]interface{}{
						"type":        "number",
						"description": "Days without movement (default: 90)",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of products to return (default: 50)",
					},
				},
			},
		},
		{
			Name:        "getinventoryturnover",
			Description: "Get inventory turnover ratio and days of inventory.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"fromdate": map[string]interface{}{
						"type":        "string",
						"description": "Start date in YYYY-MM-DD format",
					},
					"todate": map[string]interface{}{
						"type":        "string",
						"description": "End date in YYYY-MM-DD format",
					},
				},
			},
		},
		// Customer Tools
		{
			Name:        "gettopcustomers",
			Description: "Get top customers by revenue with purchase history.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"fromdate": map[string]interface{}{
						"type":        "string",
						"description": "Start date in YYYY-MM-DD format",
					},
					"todate": map[string]interface{}{
						"type":        "string",
						"description": "End date in YYYY-MM-DD format",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of customers to return (default: 10)",
					},
					"sortby": map[string]interface{}{
						"type":        "string",
						"description": "Sort by: amount, orders, profit (default: amount)",
						"enum":        []string{"amount", "orders", "profit"},
					},
				},
			},
		},
		{
			Name:        "getcustomergrowth",
			Description: "Get customer acquisition and retention metrics.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"fromdate": map[string]interface{}{
						"type":        "string",
						"description": "Start date in YYYY-MM-DD format",
					},
					"todate": map[string]interface{}{
						"type":        "string",
						"description": "End date in YYYY-MM-DD format",
					},
				},
			},
		},
		{
			Name:        "getcustomersegments",
			Description: "Get customer segmentation analysis (RFM: Recency, Frequency, Monetary).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"fromdate": map[string]interface{}{
						"type":        "string",
						"description": "Start date in YYYY-MM-DD format",
					},
					"todate": map[string]interface{}{
						"type":        "string",
						"description": "End date in YYYY-MM-DD format",
					},
				},
			},
		},
		// Comparison Tools
		{
			Name:        "getyoycomparison",
			Description: "Get year-over-year comparison of revenue, orders, and profit.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"year": map[string]interface{}{
						"type":        "number",
						"description": "Year to compare (default: current year)",
					},
					"month": map[string]interface{}{
						"type":        "number",
						"description": "Specific month to compare (1-12, optional)",
					},
				},
			},
		},
		{
			Name:        "getmomcomparison",
			Description: "Get month-over-month comparison with weekly breakdown.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"year": map[string]interface{}{
						"type":        "number",
						"description": "Year (default: current year)",
					},
					"month": map[string]interface{}{
						"type":        "number",
						"description": "Month to compare (1-12, default: current month)",
					},
				},
			},
		},
		// Database Tools
		{
			Name:        "getdatabaseschema",
			Description: "Get database structure including tables, columns, data types. Use to explore available data.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"tablename": map[string]interface{}{
						"type":        "string",
						"description": "Filter by table name (partial match, optional)",
					},
				},
			},
		},
		{
			Name:        "executequery",
			Description: "Execute a readonly SQL query (SELECT only). For custom data retrieval.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "SQL SELECT query to execute",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Max rows to return (default: 100, max: 1000)",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "gettablesample",
			Description: "Get sample data from a specific table.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"tablename": map[string]interface{}{
						"type":        "string",
						"description": "Table name to sample",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Number of rows (default: 10, max: 100)",
					},
				},
				"required": []string{"tablename"},
			},
		},
		// MongoDB Tools
		{
			Name:        "querymongodb",
			Description: "Query MongoDB collection with JSON filter (readonly). Returns documents matching the filter.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database name (default: from config)",
					},
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name to query",
					},
					"filter": map[string]interface{}{
						"type":        "string",
						"description": "JSON filter e.g. {\"transflag\":6} (default: {})",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Max documents to return (default: 20, max: 100)",
					},
				},
				"required": []string{"collection"},
			},
		},
		{
			Name:        "listmongodbcollections",
			Description: "List all collections in a MongoDB database.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database name (default: from config)",
					},
				},
			},
		},
		{
			Name:        "aggregatemongodb",
			Description: "Run aggregation pipeline on MongoDB collection (readonly). Blocks $out and $merge stages.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database name (default: from config)",
					},
					"collection": map[string]interface{}{
						"type":        "string",
						"description": "Collection name",
					},
					"pipeline": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of pipeline stages",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Max results (default: 100, max: 100)",
					},
				},
				"required": []string{"collection", "pipeline"},
			},
		},
		// ClickHouse Tools
		{
			Name:        "queryclickhouse",
			Description: "Execute readonly SELECT/SHOW query on ClickHouse (OLAP analytics database).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database name (default: from env CH_DATABASE_NAME)",
					},
					"query": map[string]interface{}{
						"type":        "string",
						"description": "SQL SELECT or SHOW query",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Max rows (default: 100, max: 1000)",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "listclickhousetables",
			Description: "List all tables in ClickHouse database with engine, row count and size info.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database name (default: from env CH_DATABASE_NAME)",
					},
				},
			},
		},
		// Dev Database Tools
		{
			Name:        "executepgcommand",
			Description: "⚡ DEV TOOL: Execute ANY SQL on PostgreSQL (SELECT, DELETE, INSERT, UPDATE, ALTER, TRUNCATE, DROP). No readonly restriction.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"holdingcode": map[string]interface{}{
						"type":        "string",
						"description": "Holding Code (= PostgreSQL database name)",
					},
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Any SQL command",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Max rows for SELECT (default: 100, max: 10000)",
					},
				},
				"required": []string{"holdingcode", "query"},
			},
		},
		{
			Name:        "executechcommand",
			Description: "⚡ DEV TOOL: Execute ANY SQL on ClickHouse (SELECT, ALTER TABLE DELETE, INSERT, DROP, TRUNCATE). No readonly restriction.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database name (default: from env CH_DATABASE_NAME)",
					},
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Any SQL command",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Max rows for SELECT (default: 100, max: 10000)",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "listapiendpoints",
			Description: "List all available API endpoints with full details (method, path, parameters, request/response schema). Use to discover APIs for frontend development.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"category": map[string]interface{}{
						"type":        "string",
						"description": "Filter by category (e.g., product, sales, transaction, approval)",
					},
					"keyword": map[string]interface{}{
						"type":        "string",
						"description": "Search in path or description",
					},
					"method": map[string]interface{}{
						"type":        "string",
						"description": "Filter by HTTP method (GET, POST, PUT, DELETE)",
					},
					"source": map[string]interface{}{
						"type":        "string",
						"description": "Filter by source: goapi or mainapi (default: all)",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Max endpoints to return (default: 50, max: 500)",
					},
				},
			},
		},
		{
			Name:        "getapispec",
			Description: "Get detailed request/response specification for a specific API endpoint with curl example and usage notes.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "API path or partial match (e.g., '/product/search')",
					},
					"method": map[string]interface{}{
						"type":        "string",
						"description": "HTTP method filter (GET, POST, PUT, DELETE)",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "getapiexample",
			Description: "Get real request/response examples with curl command, Dart code snippet, and TypeScript code snippet.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "API path or partial match",
					},
					"method": map[string]interface{}{
						"type":        "string",
						"description": "HTTP method filter",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "listenums",
			Description: "List all enum values and constants used in the backend (transflag, payment types, approval status, VAT types, etc.).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"category": map[string]interface{}{
						"type":        "string",
						"description": "Filter by category (transaction, payment, approval, datahistory, kafka)",
					},
					"keyword": map[string]interface{}{
						"type":        "string",
						"description": "Search in enum name, label, or description",
					},
				},
			},
		},
		{
			Name:        "getmodelschema",
			Description: "Get Go struct definitions with auto-generated Dart class and TypeScript interface. Use to create frontend models matching backend.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"model": map[string]interface{}{
						"type":        "string",
						"description": "Model name or partial match (e.g., 'MongoDocModel', 'Customer')",
					},
					"category": map[string]interface{}{
						"type":        "string",
						"description": "Filter by category (transaction, product, master, stock, file)",
					},
					"keyword": map[string]interface{}{
						"type":        "string",
						"description": "Search in model name or description",
					},
				},
			},
		},
		// Unit of Measure Tools (หน่วยนับ)
		{
			Name:        "listunits",
			Description: "List/search units of measure (หน่วยนับ). Returns unit codes and names with multi-language support.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":        "string",
						"description": "Search by unit code or name (e.g., 'ชิ้น', 'EA', 'กล่อง')",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Max results to return (default: 50, max: 200)",
					},
				},
			},
		},
		{
			Name:        "createunit",
			Description: "Create a new unit of measure (หน่วยนับ). Uses names[] for multi-language display names.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"unitcode": map[string]interface{}{
						"type":        "string",
						"description": "Unit code (e.g., 'EA', 'BOX', 'KG', 'PACK')",
					},
					"names": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of multi-language names e.g. [{\"code\":\"th\",\"name\":\"ชิ้น\"},{\"code\":\"en\",\"name\":\"Each\"}]",
					},
				},
				"required": []string{"unitcode", "names"},
			},
		},
		{
			Name:        "createunits",
			Description: "Create multiple units of measure at once (bulk). Skips duplicates automatically.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"units": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of units. Each item needs unitcode (required) and optionally names array. e.g. [{\"unitcode\":\"EA\",\"names\":[{\"code\":\"th\",\"name\":\"ชิ้น\"}]},{\"unitcode\":\"BOX\",\"names\":[{\"code\":\"th\",\"name\":\"กล่อง\"}]}]. Max 100 items.",
					},
				},
				"required": []string{"units"},
			},
		},
		{
			Name:        "updateunit",
			Description: "Update an existing unit of measure by unit code.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"unitcode": map[string]interface{}{
						"type":        "string",
						"description": "Unit code to update",
					},
					"names": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of multi-language names [{\"code\":\"th\",\"name\":\"ชิ้น\"}]",
					},
				},
				"required": []string{"unitcode"},
			},
		},
		{
			Name:        "deleteunit",
			Description: "Delete a unit of measure by unit code.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"unitcode": map[string]interface{}{
						"type":        "string",
						"description": "Unit code to delete",
					},
				},
				"required": []string{"unitcode"},
			},
		},
		{
			Name:        "deleteunits",
			Description: "Delete multiple units of measure at once (bulk). Reports which were deleted and which were not found.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"unitcodes": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of unit codes to delete e.g. [\"EA\",\"BOX\",\"KG\"]. Max 100 items.",
					},
				},
				"required": []string{"unitcodes"},
			},
		},
		{
			Name:        "getunitschema",
			Description: "Get the data structure/schema of unit of measure (หน่วยนับ) documents with examples.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		// Product Barcode Tools (สินค้า/บาร์โค้ด)
		{
			Name:        "listbarcodes",
			Description: "List/search product barcodes (สินค้า/บาร์โค้ด). Search by barcode, item code, or product name. Returns barcode, names, unit, prices, and classification.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":        "string",
						"description": "Search by barcode, item code, or product name (e.g., '8859100001234', 'SKU001', 'น้ำดื่ม')",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Max results to return (default: 50, max: 200)",
					},
				},
			},
		},
		{
			Name:        "createbarcode",
			Description: "Create a new product barcode (สินค้า/บาร์โค้ด). Uses names[] for multi-language product names.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"barcode": map[string]interface{}{
						"type":        "string",
						"description": "Barcode value (e.g., '8859100001234')",
					},
					"itemcode": map[string]interface{}{
						"type":        "string",
						"description": "Item/product code (e.g., 'SKU001')",
					},
					"names": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of multi-language names e.g. [{\"code\":\"th\",\"name\":\"สินค้า A\"},{\"code\":\"en\",\"name\":\"Product A\"}]",
					},
					"itemunitcode": map[string]interface{}{
						"type":        "string",
						"description": "Unit code (e.g., 'EA', 'BOX')",
					},
					"itemunitnames": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of unit names [{\"code\":\"th\",\"name\":\"ชิ้น\"}]",
					},
					"prices": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of prices [{\"keynumber\":1,\"price\":100.00}]",
					},
					"standvalue": map[string]interface{}{
						"type":        "number",
						"description": "Unit conversion numerator (default: 1). e.g. BOX=24 means 1 BOX = 24 base units",
					},
					"dividevalue": map[string]interface{}{
						"type":        "number",
						"description": "Unit conversion denominator (default: 1)",
					},
					"ismainbarcode": map[string]interface{}{
						"type":        "boolean",
						"description": "Is main barcode? Auto-detected if not provided: true when standvalue=1 & dividevalue=1, false otherwise",
					},
					"groupcode": map[string]interface{}{
						"type":        "string",
						"description": "Product group code (e.g., 'GRP-ELEC'). Auto-fills groupnames from productGroups collection.",
					},
					"categorycode": map[string]interface{}{
						"type":        "string",
						"description": "Product category code (e.g., 'CAT-WIRE'). Auto-fills categorynames from productCategories collection.",
					},
				},
				"required": []string{"barcode", "itemcode", "names"},
			},
		},
		{
			Name:        "createbarcodes",
			Description: "Create multiple product barcodes at once (bulk). Skips duplicates automatically.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"barcodes": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of barcodes. Each needs barcode (required), itemcode (required), names (required). e.g. [{\"barcode\":\"123\",\"itemcode\":\"SKU1\",\"names\":[{\"code\":\"th\",\"name\":\"สินค้า\"}]}]. Max 100 items.",
					},
				},
				"required": []string{"barcodes"},
			},
		},
		{
			Name:        "updatebarcode",
			Description: "Update an existing product barcode by guidfixed.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"guidfixed": map[string]interface{}{
						"type":        "string",
						"description": "GuidFixed of the barcode to update",
					},
					"names": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of multi-language names [{\"code\":\"th\",\"name\":\"ชื่อใหม่\"}]",
					},
					"itemunitcode": map[string]interface{}{
						"type":        "string",
						"description": "New unit code",
					},
					"itemunitnames": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of unit names",
					},
					"prices": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of prices [{\"keynumber\":1,\"price\":150.00}]",
					},
					"groupcode": map[string]interface{}{
						"type":        "string",
						"description": "Product group code. Auto-fills groupnames from productGroups collection.",
					},
					"categorycode": map[string]interface{}{
						"type":        "string",
						"description": "Product category code. Auto-fills categorynames from productCategories collection.",
					},
				},
				"required": []string{"guidfixed"},
			},
		},
		{
			Name:        "deletebarcode",
			Description: "Delete a product barcode by guidfixed.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"guidfixed": map[string]interface{}{
						"type":        "string",
						"description": "GuidFixed of the barcode to delete",
					},
				},
				"required": []string{"guidfixed"},
			},
		},
		{
			Name:        "deletebarcodes",
			Description: "Delete multiple product barcodes at once (bulk). Reports which were deleted and which were not found.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"guidfixeds": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of guidfixed values to delete e.g. [\"guid1\",\"guid2\"]. Max 100 items.",
					},
				},
				"required": []string{"guidfixeds"},
			},
		},
		{
			Name:        "getbarcodeschema",
			Description: "Get the data structure/schema of product barcode (สินค้า/บาร์โค้ด) documents with examples.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		// Reference Barcodes + Multi-Unit
		{
			Name:        "getrefbarcodes",
			Description: "Get reference barcodes and unit chain for a product. Shows all units (e.g., ชิ้น→ลัง) and how they reference each other.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"itemcode": map[string]interface{}{
						"type":        "string",
						"description": "Item code to get all barcodes for",
					},
					"barcode": map[string]interface{}{
						"type":        "string",
						"description": "Barcode to find item code from (use when itemcode is unknown)",
					},
				},
			},
		},
		{
			Name:        "setrefbarcode",
			Description: "Set reference barcode for a barcode (unit conversion). E.g., BOX → EA means 1 BOX = 24 EA. Checks for circular references.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"barcode": map[string]interface{}{
						"type":        "string",
						"description": "Source barcode (e.g., BOX barcode)",
					},
					"refbarcode": map[string]interface{}{
						"type":        "string",
						"description": "Target reference barcode (e.g., EA barcode). Must be same itemcode.",
					},
					"qty": map[string]interface{}{
						"type":        "number",
						"description": "Quantity for condition-based reference",
					},
					"standvalue": map[string]interface{}{
						"type":        "number",
						"description": "Override stand value (default: use existing)",
					},
					"dividevalue": map[string]interface{}{
						"type":        "number",
						"description": "Override divide value (default: use existing)",
					},
					"condition": map[string]interface{}{
						"type":        "boolean",
						"description": "Is conditional reference (default: false)",
					},
				},
				"required": []string{"barcode", "refbarcode"},
			},
		},
		{
			Name:        "createmultiunitbarcode",
			Description: "Create a product with multiple units at once (e.g., ชิ้น + ลัง + แพ็ค). Automatically sets reference barcodes. The unit with standvalue=1 becomes the base unit.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"itemcode": map[string]interface{}{
						"type":        "string",
						"description": "Item code for all units",
					},
					"names": map[string]interface{}{
						"type":        "string",
						"description": "Default JSON names array (used when unit doesn't have its own names) e.g. [{\"code\":\"th\",\"name\":\"น้ำดื่ม\"}]",
					},
					"units": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of units. Each needs barcode, itemunitcode, standvalue. e.g. [{\"barcode\":\"EA-001\",\"itemunitcode\":\"EA\",\"standvalue\":1,\"dividevalue\":1},{\"barcode\":\"BOX-001\",\"itemunitcode\":\"BOX\",\"standvalue\":24,\"dividevalue\":1}]",
					},
					"groupcode": map[string]interface{}{
						"type":        "string",
						"description": "Product group code (shared for all units). Auto-fills groupnames.",
					},
					"categorycode": map[string]interface{}{
						"type":        "string",
						"description": "Product category code (shared for all units). Auto-fills categorynames.",
					},
				},
				"required": []string{"itemcode", "units"},
			},
		},
		// Product Group Tools (กลุ่มสินค้า)
		{
			Name:        "listproductgroups",
			Description: "List/search product groups (กลุ่มสินค้า). Returns group codes and names with multi-language support.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":        "string",
						"description": "Search by group code or name (e.g., 'อาหาร', 'FOOD')",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Max results to return (default: 50, max: 200)",
					},
				},
			},
		},
		{
			Name:        "createproductgroup",
			Description: "Create a new product group (กลุ่มสินค้า). Uses names[] for multi-language display names.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":        "string",
						"description": "Group code (e.g., 'FOOD', 'DRINK', 'TOOL')",
					},
					"names": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of multi-language names e.g. [{\"code\":\"th\",\"name\":\"อาหาร\"},{\"code\":\"en\",\"name\":\"Food\"}]",
					},
				},
				"required": []string{"code", "names"},
			},
		},
		{
			Name:        "createproductgroups",
			Description: "Create multiple product groups at once (bulk). Skips duplicates automatically.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"groups": map[string]interface{}{
						"type":        "string",
						"description": "JSON array e.g. [{\"code\":\"FOOD\",\"names\":[{\"code\":\"th\",\"name\":\"อาหาร\"}]}]. Max 100 items.",
					},
				},
				"required": []string{"groups"},
			},
		},
		{
			Name:        "updateproductgroup",
			Description: "Update an existing product group by code.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":        "string",
						"description": "Group code to update",
					},
					"names": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of multi-language names [{\"code\":\"th\",\"name\":\"อาหาร\"}]",
					},
				},
				"required": []string{"code"},
			},
		},
		{
			Name:        "deleteproductgroup",
			Description: "Delete a product group by code.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":        "string",
						"description": "Group code to delete",
					},
				},
				"required": []string{"code"},
			},
		},
		{
			Name:        "deleteproductgroups",
			Description: "Delete multiple product groups at once (bulk). Reports which were deleted and which were not found.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"codes": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of group codes e.g. [\"FOOD\",\"DRINK\"]. Max 100 items.",
					},
				},
				"required": []string{"codes"},
			},
		},
		{
			Name:        "getproductgroupschema",
			Description: "Get the data structure/schema of product group (กลุ่มสินค้า) documents with examples.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		// Product Category Tools (หมวดสินค้า)
		{
			Name:        "listproductcategories",
			Description: "List/search product categories (หมวดสินค้า). Returns category names, hierarchy, and group numbers.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":        "string",
						"description": "Search by category name (e.g., 'เนื้อสัตว์', 'ผัก')",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Max results to return (default: 50, max: 200)",
					},
				},
			},
		},
		{
			Name:        "createproductcategory",
			Description: "Create a new product category (หมวดสินค้า). Uses names[] for multi-language display names. Supports hierarchy via parentguid.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"names": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of multi-language names e.g. [{\"code\":\"th\",\"name\":\"เนื้อสัตว์\"},{\"code\":\"en\",\"name\":\"Meat\"}]",
					},
					"parentguid": map[string]interface{}{
						"type":        "string",
						"description": "Parent category guidfixed (for hierarchy)",
					},
					"groupnumber": map[string]interface{}{
						"type":        "number",
						"description": "Group/sort number",
					},
				},
				"required": []string{"names"},
			},
		},
		{
			Name:        "createproductcategories",
			Description: "Create multiple product categories at once (bulk). Auto-generates guidfixed for each.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"categories": map[string]interface{}{
						"type":        "string",
						"description": "JSON array e.g. [{\"names\":[{\"code\":\"th\",\"name\":\"เนื้อสัตว์\"}],\"groupnumber\":1}]. Max 100 items.",
					},
				},
				"required": []string{"categories"},
			},
		},
		{
			Name:        "updateproductcategory",
			Description: "Update an existing product category by guidfixed.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"guidfixed": map[string]interface{}{
						"type":        "string",
						"description": "GuidFixed of the category to update",
					},
					"names": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of multi-language names [{\"code\":\"th\",\"name\":\"เนื้อสัตว์\"}]",
					},
					"parentguid": map[string]interface{}{
						"type":        "string",
						"description": "New parent category guidfixed",
					},
					"groupnumber": map[string]interface{}{
						"type":        "number",
						"description": "New group/sort number",
					},
				},
				"required": []string{"guidfixed"},
			},
		},
		{
			Name:        "deleteproductcategory",
			Description: "Delete a product category by guidfixed.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"guidfixed": map[string]interface{}{
						"type":        "string",
						"description": "GuidFixed of the category to delete",
					},
				},
				"required": []string{"guidfixed"},
			},
		},
		{
			Name:        "deleteproductcategories",
			Description: "Delete multiple product categories at once (bulk). Reports which were deleted and which were not found.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"guidfixeds": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of guidfixed values e.g. [\"guid1\",\"guid2\"]. Max 100 items.",
					},
				},
				"required": []string{"guidfixeds"},
			},
		},
		{
			Name:        "getproductcategoryschema",
			Description: "Get the data structure/schema of product category (หมวดสินค้า) documents with examples.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		// Creditor Tools (เจ้าหนี้)
		{
			Name:        "listcreditors",
			Description: "List/search creditors (เจ้าหนี้). Returns creditor codes, names, tax ID, and contact info.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":        "string",
						"description": "Search by creditor code, name, or tax ID (e.g., 'บริษัท ABC', 'CR-001', '0105XXX')",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Max results to return (default: 50, max: 200)",
					},
				},
			},
		},
		{
			Name:        "createcreditor",
			Description: "Create a new creditor (เจ้าหนี้). Uses names[] for multi-language display names.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":        "string",
						"description": "Creditor code (e.g., 'CR-001')",
					},
					"names": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of multi-language names e.g. [{\"code\":\"th\",\"name\":\"บริษัท ABC\"},{\"code\":\"en\",\"name\":\"ABC Co.\"}]",
					},
					"taxid": map[string]interface{}{
						"type":        "string",
						"description": "Tax ID (optional)",
					},
					"email": map[string]interface{}{
						"type":        "string",
						"description": "Email (optional)",
					},
					"personaltype": map[string]interface{}{
						"type":        "number",
						"description": "Personal type: 0=company, 1=individual (default: 0)",
					},
					"creditday": map[string]interface{}{
						"type":        "number",
						"description": "Credit days (default: 0)",
					},
					"addressforbilling": map[string]interface{}{
						"type":        "string",
						"description": "JSON object for billing address (optional)",
					},
				},
				"required": []string{"code", "names"},
			},
		},
		{
			Name:        "createcreditors",
			Description: "Create multiple creditors at once (bulk). Skips duplicates automatically.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"creditors": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of creditors. Each needs code (required) and names (required). Max 100 items.",
					},
				},
				"required": []string{"creditors"},
			},
		},
		{
			Name:        "updatecreditor",
			Description: "Update an existing creditor by guidfixed.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"guidfixed": map[string]interface{}{
						"type":        "string",
						"description": "GuidFixed of the creditor to update",
					},
					"names": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of multi-language names",
					},
					"taxid": map[string]interface{}{
						"type":        "string",
						"description": "New tax ID",
					},
					"email": map[string]interface{}{
						"type":        "string",
						"description": "New email",
					},
					"creditday": map[string]interface{}{
						"type":        "number",
						"description": "New credit days",
					},
					"addressforbilling": map[string]interface{}{
						"type":        "string",
						"description": "JSON object for billing address",
					},
				},
				"required": []string{"guidfixed"},
			},
		},
		{
			Name:        "deletecreditor",
			Description: "Delete a creditor by guidfixed (soft delete).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"guidfixed": map[string]interface{}{
						"type":        "string",
						"description": "GuidFixed of the creditor to delete",
					},
				},
				"required": []string{"guidfixed"},
			},
		},
		{
			Name:        "deletecreditors",
			Description: "Delete multiple creditors at once (bulk soft delete). Reports which were deleted and which were not found.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"guidfixeds": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of guidfixed values to delete e.g. [\"guid1\",\"guid2\"]. Max 100 items.",
					},
				},
				"required": []string{"guidfixeds"},
			},
		},
		{
			Name:        "getcreditorschema",
			Description: "Get the data structure/schema of creditor (เจ้าหนี้) documents with examples.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		// Debtor Tools (ลูกหนี้)
		{
			Name:        "listdebtors",
			Description: "List/search debtors (ลูกหนี้). Returns debtor codes, names, tax ID, and contact info.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"keyword": map[string]interface{}{
						"type":        "string",
						"description": "Search by debtor code, name, or tax ID (e.g., 'ร้าน ABC', 'DB-001', '0105XXX')",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Max results to return (default: 50, max: 200)",
					},
				},
			},
		},
		{
			Name:        "createdebtor",
			Description: "Create a new debtor (ลูกหนี้). Uses names[] for multi-language display names.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":        "string",
						"description": "Debtor code (e.g., 'DB-001')",
					},
					"names": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of multi-language names e.g. [{\"code\":\"th\",\"name\":\"ร้าน ABC\"},{\"code\":\"en\",\"name\":\"ABC Shop\"}]",
					},
					"taxid": map[string]interface{}{
						"type":        "string",
						"description": "Tax ID (optional)",
					},
					"email": map[string]interface{}{
						"type":        "string",
						"description": "Email (optional)",
					},
					"personaltype": map[string]interface{}{
						"type":        "number",
						"description": "Personal type: 0=company, 1=individual (default: 0)",
					},
					"creditday": map[string]interface{}{
						"type":        "number",
						"description": "Credit days (default: 0)",
					},
					"addressforbilling": map[string]interface{}{
						"type":        "string",
						"description": "JSON object for billing address (optional)",
					},
				},
				"required": []string{"code", "names"},
			},
		},
		{
			Name:        "createdebtors",
			Description: "Create multiple debtors at once (bulk). Skips duplicates automatically.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"debtors": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of debtors. Each needs code (required) and names (required). Max 100 items.",
					},
				},
				"required": []string{"debtors"},
			},
		},
		{
			Name:        "updatedebtor",
			Description: "Update an existing debtor by guidfixed.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"guidfixed": map[string]interface{}{
						"type":        "string",
						"description": "GuidFixed of the debtor to update",
					},
					"names": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of multi-language names",
					},
					"taxid": map[string]interface{}{
						"type":        "string",
						"description": "New tax ID",
					},
					"email": map[string]interface{}{
						"type":        "string",
						"description": "New email",
					},
					"creditday": map[string]interface{}{
						"type":        "number",
						"description": "New credit days",
					},
					"addressforbilling": map[string]interface{}{
						"type":        "string",
						"description": "JSON object for billing address",
					},
				},
				"required": []string{"guidfixed"},
			},
		},
		{
			Name:        "deletedebtor",
			Description: "Delete a debtor by guidfixed (soft delete).",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"guidfixed": map[string]interface{}{
						"type":        "string",
						"description": "GuidFixed of the debtor to delete",
					},
				},
				"required": []string{"guidfixed"},
			},
		},
		{
			Name:        "deletedebtors",
			Description: "Delete multiple debtors at once (bulk soft delete). Reports which were deleted and which were not found.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"guidfixeds": map[string]interface{}{
						"type":        "string",
						"description": "JSON array of guidfixed values to delete e.g. [\"guid1\",\"guid2\"]. Max 100 items.",
					},
				},
				"required": []string{"guidfixeds"},
			},
		},
		{
			Name:        "getdebtorschema",
			Description: "Get the data structure/schema of debtor (ลูกหนี้) documents with examples.",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "rebuildproducts",
			Description: "Full rebuild: sync ALL product barcodes from MongoDB → PostgreSQL + ClickHouse. Use when Kafka sync fails or data is out of sync. Same as frontend 'สร้างสินค้าใหม่' button.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"holdingcode": map[string]interface{}{
						"type":        "string",
						"description": "Holding Code to rebuild products for",
					},
				},
				"required": []string{"holdingcode"},
			},
		},
	}

	return map[string]interface{}{
		"tools": tools,
	}
}

// handleToolsListFiltered returns only tools the session's API key can access
// ถ้า IsDevMode=false → ซ่อน dev tools ออก
func (s *MCPServer) handleToolsListFiltered(session *SSESession) map[string]interface{} {
	allTools := s.handleToolsList()
	toolDefs := allTools["tools"].([]MCPToolDef)

	var filtered []MCPToolDef
	for _, t := range toolDefs {
		// กรอง dev tools ออกถ้าไม่ใช่ dev mode
		if !session.IsDevMode && devToolNames[t.Name] {
			continue
		}
		if session.APIKey.IsToolAllowed(t.Name) {
			filtered = append(filtered, t)
		}
	}

	mode := "general"
	if session.IsDevMode {
		mode = "dev"
	}
	logger.Info("[MCP] tools/list mode=%s, visible=%d/%d", mode, len(filtered), len(toolDefs))

	return map[string]interface{}{
		"tools": filtered,
	}
}

// handleToolCall handles tool invocation
func (s *MCPServer) handleToolCall(ctx context.Context, session *SSESession, params json.RawMessage) (interface{}, error) {
	var callParams struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(params, &callParams); err != nil {
		return nil, fmt.Errorf("invalid tool call params: %w", err)
	}

	// Check tool permission
	if !session.APIKey.IsToolAllowed(callParams.Name) {
		return nil, fmt.Errorf("tool '%s' is not allowed for this API key", callParams.Name)
	}

	// Add holdingcode from API key
	if callParams.Arguments == nil {
		callParams.Arguments = make(map[string]interface{})
	}
	callParams.Arguments["holdingcode"] = session.APIKey.HoldingCode

	// Execute tool
	startTime := time.Now()
	var result interface{}
	var err error

	switch callParams.Name {
	// Product Search
	case "searchproducts":
		result, err = s.invokeSearchProducts(ctx, callParams.Arguments)
	// Sales Tools
	case "getdailysales":
		result, err = s.invokeGetDailySales(ctx, callParams.Arguments)
	case "getsalesbydaterange":
		result, err = s.invokeGetSalesByDateRange(ctx, callParams.Arguments)
	case "gettopsellingproducts":
		result, err = s.invokeGetTopSellingProducts(ctx, callParams.Arguments)
	case "getsalesbyseller":
		result, err = s.invokeGetSalesBySeller(ctx, callParams.Arguments)
	case "getmonthlysummary":
		result, err = s.invokeGetMonthlySummary(ctx, callParams.Arguments)
	// Dashboard Tools
	case "getdashboardkpis":
		result, err = s.invokeGetDashboardKPIs(ctx, callParams.Arguments)
	case "getbusinesshealth":
		result, err = s.invokeGetBusinessHealth(ctx, callParams.Arguments)
	// Financial Tools
	case "getprofitanalysis":
		result, err = s.invokeGetProfitAnalysis(ctx, callParams.Arguments)
	case "getaccountsreceivable":
		result, err = s.invokeGetAccountsReceivable(ctx, callParams.Arguments)
	case "getaccountspayable":
		result, err = s.invokeGetAccountsPayable(ctx, callParams.Arguments)
	case "getcashflow":
		result, err = s.invokeGetCashFlow(ctx, callParams.Arguments)
	// Inventory Tools
	case "getinventoryvalue":
		result, err = s.invokeGetInventoryValue(ctx, callParams.Arguments)
	case "getlowstockalerts":
		result, err = s.invokeGetLowStockAlerts(ctx, callParams.Arguments)
	case "getdeadstock":
		result, err = s.invokeGetDeadStock(ctx, callParams.Arguments)
	case "getinventoryturnover":
		result, err = s.invokeGetInventoryTurnover(ctx, callParams.Arguments)
	// Customer Tools
	case "gettopcustomers":
		result, err = s.invokeGetTopCustomers(ctx, callParams.Arguments)
	case "getcustomergrowth":
		result, err = s.invokeGetCustomerGrowth(ctx, callParams.Arguments)
	case "getcustomersegments":
		result, err = s.invokeGetCustomerSegments(ctx, callParams.Arguments)
	// Comparison Tools
	case "getyoycomparison":
		result, err = s.invokeGetYoYComparison(ctx, callParams.Arguments)
	case "getmomcomparison":
		result, err = s.invokeGetMoMComparison(ctx, callParams.Arguments)
	// Database Tools
	case "getdatabaseschema":
		result, err = s.invokeGetDatabaseSchema(ctx, callParams.Arguments)
	case "executequery":
		result, err = s.invokeExecuteQuery(ctx, callParams.Arguments)
	case "gettablesample":
		result, err = s.invokeGetTableSample(ctx, callParams.Arguments)
	// MongoDB Tools
	case "querymongodb":
		result, err = s.invokeQueryMongoDB(ctx, callParams.Arguments)
	case "listmongodbcollections":
		result, err = s.invokeListMongoDBCollections(ctx, callParams.Arguments)
	case "aggregatemongodb":
		result, err = s.invokeAggregateMongoDB(ctx, callParams.Arguments)
	// ClickHouse Tools
	case "queryclickhouse":
		result, err = s.invokeQueryClickHouse(ctx, callParams.Arguments)
	case "listclickhousetables":
		result, err = s.invokeListClickHouseTables(ctx, callParams.Arguments)
	// Dev Database Tools
	case "executepgcommand":
		result, err = s.invokeExecutePgCommand(ctx, callParams.Arguments)
	case "executechcommand":
		result, err = s.invokeExecuteChCommand(ctx, callParams.Arguments)
	// API Catalog & Frontend Dev Tools
	case "listapiendpoints":
		result, err = s.invokeListAPIEndpoints(ctx, callParams.Arguments)
	case "getapispec":
		result, err = s.invokeGetAPISpec(ctx, callParams.Arguments)
	case "getapiexample":
		result, err = s.invokeGetAPIExample(ctx, callParams.Arguments)
	case "listenums":
		result, err = s.invokeListEnums(ctx, callParams.Arguments)
	case "getmodelschema":
		result, err = s.invokeGetModelSchema(ctx, callParams.Arguments)
	// Unit of Measure Tools (หน่วยนับ)
	case "listunits":
		result, err = s.invokeListUnits(ctx, callParams.Arguments)
	case "createunit":
		result, err = s.invokeCreateUnit(ctx, callParams.Arguments)
	case "createunits":
		result, err = s.invokeCreateUnits(ctx, callParams.Arguments)
	case "updateunit":
		result, err = s.invokeUpdateUnit(ctx, callParams.Arguments)
	case "deleteunit":
		result, err = s.invokeDeleteUnit(ctx, callParams.Arguments)
	case "deleteunits":
		result, err = s.invokeDeleteUnits(ctx, callParams.Arguments)
	case "getunitschema":
		result, err = s.invokeGetUnitSchema(ctx, callParams.Arguments)
	// Product Barcode Tools (สินค้า/บาร์โค้ด)
	case "listbarcodes":
		result, err = s.invokeListBarcodes(ctx, callParams.Arguments)
	case "createbarcode":
		result, err = s.invokeCreateBarcode(ctx, callParams.Arguments)
	case "createbarcodes":
		result, err = s.invokeCreateBarcodes(ctx, callParams.Arguments)
	case "updatebarcode":
		result, err = s.invokeUpdateBarcode(ctx, callParams.Arguments)
	case "deletebarcode":
		result, err = s.invokeDeleteBarcode(ctx, callParams.Arguments)
	case "deletebarcodes":
		result, err = s.invokeDeleteBarcodes(ctx, callParams.Arguments)
	case "getbarcodeschema":
		result, err = s.invokeGetBarcodeSchema(ctx, callParams.Arguments)
	// Reference Barcodes + Multi-Unit
	case "getrefbarcodes":
		result, err = s.invokeGetRefBarcodes(ctx, callParams.Arguments)
	case "setrefbarcode":
		result, err = s.invokeSetRefBarcode(ctx, callParams.Arguments)
	case "createmultiunitbarcode":
		result, err = s.invokeCreateMultiUnitBarcode(ctx, callParams.Arguments)
	case "rebuildproducts":
		result, err = s.invokeRebuildProducts(ctx, callParams.Arguments)
	// Product Group Tools
	case "listproductgroups":
		result, err = s.invokeListProductGroups(ctx, callParams.Arguments)
	case "createproductgroup":
		result, err = s.invokeCreateProductGroup(ctx, callParams.Arguments)
	case "createproductgroups":
		result, err = s.invokeCreateProductGroups(ctx, callParams.Arguments)
	case "updateproductgroup":
		result, err = s.invokeUpdateProductGroup(ctx, callParams.Arguments)
	case "deleteproductgroup":
		result, err = s.invokeDeleteProductGroup(ctx, callParams.Arguments)
	case "deleteproductgroups":
		result, err = s.invokeDeleteProductGroups(ctx, callParams.Arguments)
	case "getproductgroupschema":
		result, err = s.invokeGetProductGroupSchema(ctx, callParams.Arguments)
	// Product Category Tools
	case "listproductcategories":
		result, err = s.invokeListProductCategories(ctx, callParams.Arguments)
	case "createproductcategory":
		result, err = s.invokeCreateProductCategory(ctx, callParams.Arguments)
	case "createproductcategories":
		result, err = s.invokeCreateProductCategories(ctx, callParams.Arguments)
	case "updateproductcategory":
		result, err = s.invokeUpdateProductCategory(ctx, callParams.Arguments)
	case "deleteproductcategory":
		result, err = s.invokeDeleteProductCategory(ctx, callParams.Arguments)
	case "deleteproductcategories":
		result, err = s.invokeDeleteProductCategories(ctx, callParams.Arguments)
	case "getproductcategoryschema":
		result, err = s.invokeGetProductCategorySchema(ctx, callParams.Arguments)
	// Creditor Tools (เจ้าหนี้)
	case "listcreditors":
		result, err = s.invokeListCreditors(ctx, callParams.Arguments)
	case "createcreditor":
		result, err = s.invokeCreateCreditor(ctx, callParams.Arguments)
	case "createcreditors":
		result, err = s.invokeCreateCreditors(ctx, callParams.Arguments)
	case "updatecreditor":
		result, err = s.invokeUpdateCreditor(ctx, callParams.Arguments)
	case "deletecreditor":
		result, err = s.invokeDeleteCreditor(ctx, callParams.Arguments)
	case "deletecreditors":
		result, err = s.invokeDeleteCreditors(ctx, callParams.Arguments)
	case "getcreditorschema":
		result, err = s.invokeGetCreditorSchema(ctx, callParams.Arguments)
	// Debtor Tools (ลูกหนี้)
	case "listdebtors":
		result, err = s.invokeListDebtors(ctx, callParams.Arguments)
	case "createdebtor":
		result, err = s.invokeCreateDebtor(ctx, callParams.Arguments)
	case "createdebtors":
		result, err = s.invokeCreateDebtors(ctx, callParams.Arguments)
	case "updatedebtor":
		result, err = s.invokeUpdateDebtor(ctx, callParams.Arguments)
	case "deletedebtor":
		result, err = s.invokeDeleteDebtor(ctx, callParams.Arguments)
	case "deletedebtors":
		result, err = s.invokeDeleteDebtors(ctx, callParams.Arguments)
	case "getdebtorschema":
		result, err = s.invokeGetDebtorSchema(ctx, callParams.Arguments)
	default:
		return nil, fmt.Errorf("unknown tool: %s", callParams.Name)
	}

	// Log execution
	executionTime := time.Since(startTime).Milliseconds()
	go s.logAudit(session.APIKey, callParams.Name, callParams.Arguments, err, executionTime)

	if err != nil {
		return map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Error: %s", err.Error()),
				},
			},
			"isError": true,
		}, nil
	}

	// Format result as MCP content
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": string(resultJSON),
			},
		},
	}, nil
}
