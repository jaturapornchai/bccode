package unified

import (
	"context"
	"net/http"
	"time"

	"smlcloudplatform/internal/goapi/aiprovider"
	unifiedcache "smlcloudplatform/internal/goapi/cache/unified"
	"smlcloudplatform/internal/goapi/logger"

	"github.com/labstack/echo/v4"
)

// UnifiedQueryRequest แทนคำขอ unified query
type UnifiedQueryRequest struct {
	HoldingCode     string                 `json:"holdingcode"`
	Question        string                 `json:"question"`
	QueryType       string                 `json:"querytype"` // "chat", "search", "stock", "document", "unified"
	IncludeRealTime bool                   `json:"includerealtime"`
	Filters         map[string]interface{} `json:"filters"`
	Context         map[string]interface{} `json:"context"`
}

// UnifiedQueryResponse แทนคำตอบ unified query
type UnifiedQueryResponse struct {
	Success        bool                   `json:"success"`
	ResponseType   string                 `json:"responsetype"`
	Data           map[string]interface{} `json:"data"`
	AIResponse     string                 `json:"airesponse"`
	SearchResults  []SearchResult         `json:"searchresults"`
	StockData      *StockData             `json:"stockdata"`
	RealTimeData   interface{}            `json:"realtimedata"`
	CacheStatus    string                 `json:"cachestatus"` // "hit", "miss", "partial"
	ProcessingTime time.Duration          `json:"processingtime"`
	TokenUsage     *TokenUsage            `json:"tokenusage"`
	Timestamp      time.Time              `json:"timestamp"`
}

// SearchResult แทนผลการค้นหา
type SearchResult struct {
	ItemCode       string  `json:"itemcode"`
	Name           string  `json:"name"`
	Barcode        string  `json:"barcode"`
	Unit           string  `json:"unit"`
	ImageURL       string  `json:"imageurl"`
	RelevanceScore float64 `json:"relevancescore"`
	CurrentStock   float64 `json:"currentstock"`
	Warehouse      string  `json:"warehouse"`
	Location       string  `json:"location"`
}

// StockData แทนข้อมูลสต็อก
type StockData struct {
	CurrentBalance float64   `json:"currentbalance"`
	Warehouse      string    `json:"warehouse"`
	Location       string    `json:"location"`
	LastUpdated    time.Time `json:"lastupdated"`
	MovementType   string    `json:"movementtype"` // "increase", "decrease", "transfer"
	ChangeQuantity float64   `json:"changequantity"`
}

// TokenUsage แทนการใช้งาน token
type TokenUsage struct {
	PromptTokens     int     `json:"prompttokens"`
	CompletionTokens int     `json:"completiontokens"`
	TotalTokens      int     `json:"totaltokens"`
	CostUSD          float64 `json:"costusd"`
	CostTHB          float64 `json:"costthb"`
	Model            string  `json:"model"`
}

// UnifiedAPIServer แทน unified API server
type UnifiedAPIServer struct {
	cacheManager *unifiedcache.UnifiedCacheManager
	websocketHub *WebSocketHub
	aiProvider   aiprovider.AIProvider
}

// WebSocketHub แทน WebSocket hub สำหรับ real-time updates
type WebSocketHub struct {
	// WebSocket implementation จะเพิ่มในการใช้งานจริง
	// ในที่นี้เป็น placeholder
	connectedClients map[string]bool
	mu               chan bool
}

// NewUnifiedAPIServer สร้าง UnifiedAPIServer ใหม่
func NewUnifiedAPIServer() *UnifiedAPIServer {
	// สร้าง cache manager
	cacheConfig := &unifiedcache.CacheConfig{
		L1CacheSize:     10000,
		EnableRedis:     false, // จะเปิดในการใช้งานจริง
		EnableDatabase:  false, // จะเปิดในการใช้งานจริง
		AutoCleanup:     true,
		CleanupInterval: 10 * time.Minute,
	}
	cacheManager := unifiedcache.NewUnifiedCacheManager(cacheConfig)

	// สร้าง WebSocket hub
	wsHub := &WebSocketHub{
		connectedClients: make(map[string]bool),
		mu:               make(chan bool, 1),
	}

	// สร้าง AI provider
	ai := aiprovider.GetProvider()

	server := &UnifiedAPIServer{
		cacheManager: cacheManager,
		websocketHub: wsHub,
		aiProvider:   ai,
	}

	logger.Success("✅ UnifiedAPIServer เริ่มต้นเรียบร้อย")
	return server
}

// ProcessUnifiedQuery ประมวลผล unified query
func (s *UnifiedAPIServer) ProcessUnifiedQuery(c echo.Context) error {
	ctx := context.Background()
	startTime := time.Now()

	var req UnifiedQueryRequest
	if err := c.Bind(&req); err != nil {
		logger.Warn("ล้มเหลวในการแปลง request body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success":   false,
			"error":     "Invalid request format",
			"message":   "กรุณาตรวจสอบรูปแบบข้อมูลที่ส่งมา",
			"timestamp": time.Now(),
		})
	}

	// Validate request
	if req.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success":   false,
			"error":     "holdingcode is required",
			"message":   "จำเป็นต้องระบุ holdingcode",
			"timestamp": time.Now(),
		})
	}

	if req.Question == "" && req.QueryType != "search" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success":   false,
			"error":     "question is required",
			"message":   "จำเป็นต้องระบุคำถาม",
			"timestamp": time.Now(),
		})
	}

	// Set default query type
	if req.QueryType == "" {
		req.QueryType = "unified"
	}

	logger.Info("[UnifiedQuery] shop: %s, query_type: %s, include_real_time: %v, question: %s",
		req.HoldingCode, req.QueryType, req.IncludeRealTime, req.Question)

	// ประมวลผลตาม query type
	var response *UnifiedQueryResponse
	var err error

	// ประมวลผล unified query เสมอ (รวมทุกแหล่งข้อมูล)
	response, err = s.ProcessUnifiedMultiSourceQuery(ctx, &req)

	if err != nil {
		logger.Error("ล้มเหลวในการประมวลผล unified query: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success":   false,
			"error":     "processing_failed",
			"message":   "เกิดข้อผิดพลาดในการประมวลผล",
			"details":   err.Error(),
			"timestamp": time.Now(),
		})
	}

	// เพิ่ม processing time
	response.ProcessingTime = time.Since(startTime)
	response.Timestamp = time.Now()

	logger.Info("[UnifiedQuery] เสร็จสิ้นใน %v (cache: %s)", response.ProcessingTime, response.CacheStatus)

	return c.JSON(http.StatusOK, response)
}

// ProcessUnifiedMultiSourceQuery ประมวลผล query หลายแหล่งข้อมูล
func (s *UnifiedAPIServer) ProcessUnifiedMultiSourceQuery(ctx context.Context, req *UnifiedQueryRequest) (*UnifiedQueryResponse, error) {
	// ตรวจสอบ cache ก่อน
	cacheKey := generateQueryCacheKey(req)
	if cached, found := s.cacheManager.Get(ctx, "unified", req.HoldingCode, cacheKey); found {
		logger.Debug("Cache HIT สำหรับ unified query")
		if response, ok := cached.(*UnifiedQueryResponse); ok {
			response.CacheStatus = "hit"
			return response, nil
		}
	}

	// เริ่มต้น response
	response := &UnifiedQueryResponse{
		Success:       true,
		ResponseType:  "unified",
		Data:          make(map[string]interface{}),
		SearchResults: []SearchResult{},
		CacheStatus:   "miss",
		TokenUsage:    &TokenUsage{},
	}

	// 1. ประมวลผล AI Chat (ใช้ Gemini โดยตรง)
	if req.Question != "" {
		// TODO: เรียกใช้ AI Chat handler ที่มีอยู่
		// ในที่นี้เป็น placeholder
		response.AIResponse = "กำลังประมวลผลคำถามด้วย AI..."
		response.TokenUsage = &TokenUsage{
			Model: s.aiProvider.Name(),
		}
	}

	// 2. ประมวลผล Product Search
	searchQuery := extractSearchQuery(req.Question)
	if searchQuery != "" {
		// TODO: เรียกใช้ Product Search handler
		// ในที่นี้เป็น placeholder
		response.SearchResults = []SearchResult{
			{
				ItemCode:       "A001",
				Name:           "สินค้าทดสอบ",
				Barcode:        "123456789",
				Unit:           "ชิ้น",
				ImageURL:       "",
				RelevanceScore: 0.95,
				CurrentStock:   150,
				Warehouse:      "main",
				Location:       "A1",
			},
		}
	}

	// 3. ประมวลผล Real-time Stock (ใช้ existing real-time system)
	if req.IncludeRealTime && len(response.SearchResults) > 0 {
		// TODO: เรียกใช้ Real-time Stock handler ที่มีอยู่
		// ในที่นี้เป็น placeholder
		response.StockData = &StockData{
			CurrentBalance: 150,
			Warehouse:      "main",
			Location:       "A1",
			LastUpdated:    time.Now(),
			MovementType:   "none",
			ChangeQuantity: 0,
		}
		// อัปเดต search results ด้วยข้อมูลสต็อก
		response.SearchResults = augmentSearchResultsWithStock(response.SearchResults, response.StockData)
	}

	// 4. รวมข้อมูลทั้งหมดใน data field
	response.Data["ai_response"] = response.AIResponse
	response.Data["search_results"] = response.SearchResults
	if response.StockData != nil {
		response.Data["stock_data"] = response.StockData
	}
	response.Data["holdingcode"] = req.HoldingCode
	response.Data["query_type"] = "unified"
	response.Data["include_real_time"] = req.IncludeRealTime

	// 5. Cache result
	s.cacheManager.Set(ctx, "unified", req.HoldingCode, cacheKey, response, 0)

	return response, nil
}

// generateQueryCacheKey สร้าง cache key สำหรับ query
func generateQueryCacheKey(req *UnifiedQueryRequest) string {
	key := req.HoldingCode + ":" + req.QueryType + ":" + req.Question
	if req.IncludeRealTime {
		key += ":realtime"
	}
	return key
}

// extractSearchQuery แยก keyword สำหรับ search
func extractSearchQuery(question string) string {
	// Simple keyword extraction
	// ในการใช้งานจริง ควรใช้ NLP ที่ซับซ้อนกว่า
	if question == "" {
		return ""
	}

	// ตรวจสอบว่ามี keyword ที่บ่งบอกการค้นหา
	searchKeywords := []string{"ค้นหา", "หา", "สินค้า", "product", "item", "code", "barcode"}

	for _, keyword := range searchKeywords {
		if containsIgnoreCase(question, keyword) {
			return question
		}
	}

	return ""
}

// containsIgnoreCase ตรวจสอบว่ามี substring โดยไม่สนใจ case
func containsIgnoreCase(text, substring string) bool {
	return len(text) >= len(substring) &&
		findSubstringIgnoreCase(text, substring)
}

func findSubstringIgnoreCase(text, substring string) bool {
	for i := 0; i <= len(text)-len(substring); i++ {
		match := true
		for j := 0; j < len(substring); j++ {
			if toLower(text[i+j]) != toLower(substring[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}

// augmentSearchResultsWithStock เพิ่มข้อมูลสต็อกในผลการค้นหา
func augmentSearchResultsWithStock(results []SearchResult, stockData *StockData) []SearchResult {
	// ในการใช้งานจริง จะเพิ่มข้อมูลสต็อกในแต่ละ result
	// ที่นี่เป็นการเพิ่มแบบง่าย ๆ
	for i := range results {
		if stockData != nil {
			results[i].CurrentStock = stockData.CurrentBalance
			results[i].Warehouse = stockData.Warehouse
			results[i].Location = stockData.Location
		}
	}
	return results
}

// HealthCheck ตรวจสอบสถานะของ unified server
func (s *UnifiedAPIServer) HealthCheck(c echo.Context) error {
	healthData := map[string]interface{}{
		"status":         "healthy",
		"version":        "1.0.0",
		"timestamp":      time.Now().Unix(),
		"unified_server": "running",
		"components": map[string]interface{}{
			"ai_chat":        "ok",
			"product_search": "ok",
			"realtime":       "ok",
			"cache":          "ok",
			"websocket":      "ok",
		},
	}

	// ตรวจสอบ AI provider
	if s.aiProvider == nil {
		healthData["components"].(map[string]interface{})["aiprovider"] = "error"
		healthData["status"] = "degraded"
	}

	// ตรวจสอบ cache manager
	if s.cacheManager == nil {
		healthData["components"].(map[string]interface{})["cache"] = "error"
		healthData["status"] = "degraded"
	}

	// เพิ่ม cache stats
	if s.cacheManager != nil {
		stats := s.cacheManager.GetStats()
		healthData["cache_stats"] = stats
	}

	statusCode := http.StatusOK
	if healthData["status"] == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	return c.JSON(statusCode, healthData)
}

// GetCacheStats ดึงสถิติ cache
func (s *UnifiedAPIServer) GetCacheStats(c echo.Context) error {
	if s.cacheManager == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"success": false,
			"error":   "cache_manager_not_initialized",
			"message": "Cache manager ไม่ได้เริ่มต้น",
		})
	}

	stats := s.cacheManager.GetStats()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}
