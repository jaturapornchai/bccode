package goapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/dataimport"
	"smlcloudplatform/internal/goapi/handlers"
	"smlcloudplatform/internal/goapi/handlers/aichat"
	"smlcloudplatform/internal/goapi/handlers/approval"
	"smlcloudplatform/internal/goapi/handlers/datahistory"
	"smlcloudplatform/internal/goapi/handlers/knowledgebase"
	"smlcloudplatform/internal/goapi/handlers/lineoa"
	"smlcloudplatform/internal/goapi/handlers/unified"
	"smlcloudplatform/internal/goapi/inventory"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myclickhouse"
	"smlcloudplatform/internal/goapi/mydb"
	"smlcloudplatform/internal/goapi/myglobal"
	myPg "smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/mypostgres"
	"smlcloudplatform/internal/goapi/workers"

	appConfig "smlcloudplatform/internal/config"
	serviceConfig "smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/setupconfig"
	"smlcloudplatform/pkg/microservice"
	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const VersionNumber = "1.1.1121"

// GoAPIServer encapsulates all goapi services and routes
type GoAPIServer struct {
	workerManager *workers.WorkerManager
}

// New creates a new GoAPIServer instance
func New() *GoAPIServer {
	return &GoAPIServer{}
}

// Init initializes all goapi services (DB, workers, background tasks)
func (s *GoAPIServer) Init() error {
	logger.Info("GoAPI: เริ่มต้น initialization...")

	// 1. Load goapi bootstrap config (set goapi-specific env vars)
	setupconfig.LoadBootstrapConfig()

	config := serviceConfig.NewServiceConfig()
	_ = config.MongodbDatabaseName()

	// 2. MongoDB connection
	_ = myglobal.SafeMongoConnectFast()

	// 3. MongoDB (separate connection)
	if err := handlers.InitMongoAtlas(); err != nil {
		logger.Warn("GoAPI: Failed to initialize MongoDB: %v", err)
	}

	// 4. LINE OA + Approval init
	atlasClient, atlasDB := handlers.GetAtlasConnection()
	if atlasClient != nil && atlasDB != nil {
		lineoa.Init(atlasClient, atlasDB)
		logger.Success("GoAPI: ✅ Line OA handlers initialized")
		approval.Init(atlasClient, atlasDB)
		logger.Success("GoAPI: ✅ Approval handlers initialized")
	}

	// 5. S3/R2 Client
	if err := handlers.InitR2Client(); err != nil {
		logger.Warn("GoAPI: Failed to initialize R2 client: %v", err)
	}

	// 6. Database Manager Pool (PostgreSQL + ClickHouse)
	logger.Info("GoAPI: กำลังเริ่มต้น Database Manager Pool...")
	dbConfig := mydb.DatabaseManagerConfig{
		PostgreSQLHost:     os.Getenv("POSTGRES_HOST"),
		PostgreSQLPort:     os.Getenv("POSTGRES_PORT"),
		PostgreSQLUser:     os.Getenv("POSTGRES_USER"),
		PostgreSQLPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgreSQLSSLMode:  os.Getenv("POSTGRES_SSL_MODE"),
		ClickHouseHost:     os.Getenv("CLICKHOUSE_HOST"),
		ClickHousePort:     os.Getenv("CLICKHOUSE_PORT"),
		ClickHouseUser:     os.Getenv("CLICKHOUSE_USER"),
		ClickHousePassword: os.Getenv("CLICKHOUSE_PASSWORD"),
	}
	mydb.InitManagerPool(dbConfig)
	logger.Success("GoAPI: ✅ Database Manager Pool เริ่มต้นเรียบร้อย")

	// 7. Global Database Connection สำหรับ Queue System
	logger.Info("GoAPI: กำลังตั้งค่า Global Database Connection...")
	pool := mydb.GetGlobalManagerPool()
	globalDBManager, err := pool.GetManager("postgres")
	if err != nil {
		logger.Error("GoAPI: ❌ Failed to get global database manager: %v", err)
	} else {
		globalDB, err := globalDBManager.GetPostgreSQLConnection()
		if err != nil {
			logger.Error("GoAPI: ❌ Failed to get global PostgreSQL connection: %v", err)
		} else {
			myglobal.SetGlobalDatabaseConnection(globalDB)
			myglobal.SetGlobalDatabaseProvider(func() (*sql.DB, error) {
				manager, mgrErr := pool.GetManager("postgres")
				if mgrErr != nil {
					return nil, mgrErr
				}
				return manager.GetPostgreSQLConnection()
			})
			logger.Success("GoAPI: ✅ Global Database Connection ตั้งค่าเรียบร้อย")

			if err := globalDB.Ping(); err != nil {
				logger.Error("GoAPI: ❌ Global Database ping failed: %v", err)
			}

			// Queue Schema
			if err := mypostgres.InitQueueSchema(globalDB); err != nil {
				logger.Error("GoAPI: ❌ Failed to initialize queue schema: %v", err)
			}
		}
	}

	// 8. Worker System
	logger.Info("GoAPI: กำลังเริ่มต้น Worker System...")
	s.workerManager = workers.NewWorkerManager(0)
	if s.workerManager != nil {
		if err := s.workerManager.Start(); err != nil {
			logger.Error("GoAPI: Failed to start worker manager: %v", err)
		} else {
			logger.Success("GoAPI: ✅ Workers เริ่มต้นเรียบร้อย")
		}
	}

	// 9. Background goroutines
	go handlers.StartConnectionHealthChecker()
	go handlers.StartBackgroundTask()
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			myPg.CleanupPreparedStatements()
		}
	}()

	// 10. Kafka consumers
	enableKafka := os.Getenv("ENABLE_KAFKA")
	if enableKafka == "true" {
		logger.Info("GoAPI: 🚀 เริ่มต้น Kafka consumers...")
		handlers.StartConsumers()
		logger.Info("GoAPI: ✅ Kafka consumers initialized")
	}

	logger.Success("GoAPI: ✅ Initialization เสร็จสมบูรณ์")
	return nil
}

// RegisterMiddleware applies goapi-specific middleware to the group
func (s *GoAPIServer) RegisterMiddleware(g *echo.Group) {
	// 1. Panic Recovery
	g.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize:         1 << 10,
		LogLevel:          0,
		DisablePrintStack: true,
		DisableStackAll:   true,
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			logger.Error("❌ PANIC RECOVERED: %v", err)
			logger.Error("Request: %s %s", c.Request().Method, c.Request().URL.Path)
			logger.Error("Stack trace: %s", string(stack))
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"status":  "error",
				"code":    500,
				"message": "Internal server error (panic recovered)",
				"error":   err.Error(),
			})
		},
	}))

	// 2. Security Headers
	g.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self'; object-src 'none';",
	}))

	// 3. Tiered Rate Limiting
	g.Use(createTieredRateLimiter())

	// 4. Request Timeout
	g.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		ErrorMessage: "Request timeout",
		Timeout:      30 * time.Second,
	}))
}

func createGoAPIAuthMiddleware(authService *microservice.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenText, err := getBearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
			if err != nil {
				logger.Warn("GoAPI auth failed: %v", err)
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"success": false,
					"message": "Unauthorized",
				})
			}

			userInfo, ok := authenticateGoAPIRedisToken(c.Request().Context(), authService, tokenText)
			if !ok {
				logger.Warn("GoAPI auth failed: inactive session")
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"success": false,
					"message": "Unauthorized",
				})
			}
			if userInfo.HoldingCode == "" {
				logger.Warn("GoAPI auth failed: shop not selected")
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"success": false,
					"message": "Shop not selected",
				})
			}
			requestedHoldingCode, err := goAPIRequestHoldingCode(c)
			if err != nil {
				logger.Warn("GoAPI auth failed: invalid tenant payload")
				return c.JSON(http.StatusBadRequest, map[string]interface{}{
					"success": false,
					"message": "Invalid request payload",
				})
			}
			if requestedHoldingCode != "" && requestedHoldingCode != userInfo.HoldingCode {
				logger.Warn("GoAPI tenant blocked: route=%s requested_shop=%s token_shop=%s", c.Path(), requestedHoldingCode, userInfo.HoldingCode)
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"success": false,
					"message": "Forbidden",
				})
			}

			if isDevelopmentMode() {
				logger.Info("[DEV][GoAPI auth] method=%s route=%s user=%s holdingcode=%s requested_holdingcode=%s",
					c.Request().Method, c.Path(), userInfo.Username, userInfo.HoldingCode, requestedHoldingCode)
			}

			c.Set("UserInfo", userInfo)
			return next(c)
		}
	}
}

func authenticateGoAPIRedisToken(ctx context.Context, authService *microservice.AuthService, tokenText string) (msmodels.UserInfo, bool) {
	if authService == nil {
		return msmodels.UserInfo{}, false
	}
	userInfo, err := authService.AuthenticateAccessToken(ctx, tokenText)
	if err != nil {
		return msmodels.UserInfo{}, false
	}
	return userInfo, true
}

func getBearerToken(authorization string) (string, error) {
	parts := strings.SplitN(strings.TrimSpace(authorization), " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" || strings.TrimSpace(parts[1]) == "" {
		return "", fmt.Errorf("missing authorization bearer")
	}
	return strings.TrimSpace(parts[1]), nil
}

func goAPIRequestHoldingCode(c echo.Context) (string, error) {
	for _, key := range []string{"holdingcode", "tenantid", "database", "dbname"} {
		if value := strings.TrimSpace(c.QueryParam(key)); value != "" {
			return value, nil
		}
	}

	if c.Request().Body == nil {
		return "", nil
	}
	contentType := strings.ToLower(c.Request().Header.Get(echo.HeaderContentType))
	if !strings.Contains(contentType, "application/json") {
		return "", nil
	}

	bodyBytes, err := io.ReadAll(c.Request().Body)
	c.Request().Body = io.NopCloser(bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	if len(bytes.TrimSpace(bodyBytes)) == 0 {
		return "", nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		return "", err
	}
	return holdingCodeFromPayload(payload)
}

func holdingCodeFromPayload(payload map[string]interface{}) (string, error) {
	for _, key := range []string{"holdingcode", "tenantid", "database", "dbname"} {
		if value := payloadString(payload[key]); value != "" {
			return value, nil
		}
	}

	if nested, ok := payload["body"].(map[string]interface{}); ok {
		return holdingCodeFromPayload(nested)
	}
	if nestedRaw, ok := payload["body"].(string); ok && strings.TrimSpace(nestedRaw) != "" {
		var nested map[string]interface{}
		if err := json.Unmarshal([]byte(nestedRaw), &nested); err != nil {
			return "", err
		}
		return holdingCodeFromPayload(nested)
	}
	return "", nil
}

func payloadString(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	default:
		return ""
	}
}

func isDevelopmentMode() bool {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("MODE")))
	return mode == "" || mode == "dev" || mode == "development" || mode == "local"
}

// RegisterRoutes registers all goapi routes on the Echo group
// prefix คือ URL prefix ของ group เช่น "/goapi" หรือ "" (standalone)
func (s *GoAPIServer) RegisterRoutes(g *echo.Group, prefix string, authorizationFinders ...microservice.AuthorizationFinder) {
	authCacher := microservice.NewCacher((&appConfig.Config{}).CacherConfig())
	authService := microservice.NewAuthService(authCacher, 15*time.Minute, 8*time.Hour, authorizationFinders...)
	authGroup := g.Group("", createGoAPIAuthMiddleware(authService))

	// Health & Status
	g.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message":   "Hello, World!",
			"version":   VersionNumber,
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})
	g.GET("/version", func(c echo.Context) error {
		logger.Info("เรียกใช้ Version")
		return c.String(http.StatusOK, "API Version: "+VersionNumber)
	})
	g.GET("/api/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":    "healthy",
			"version":   VersionNumber,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Health check endpoints
	g.GET("/api/health/kafka", handlers.KafkaHealthHandler)
	g.GET("/api/health/background", handlers.BackgroundTaskStatusHandler)
	g.GET("/api/health/queue", handlers.QueueStatusHandler)
	g.GET("/api/health/queue/:holdingcode", handlers.QueueShopStatusHandler)
	g.GET("/api/health/database", handlers.DatabaseHealthHandler)
	g.GET("/api/health/system", handlers.SystemHealthHandler)

	// Inventory Costing
	inventoryGroup := authGroup.Group("/api")
	inventory.RegisterRoutes(inventoryGroup)

	// Stock
	authGroup.POST("/processstockcalccost", handlers.ProcessStockCalcCostHandler)
	authGroup.POST("/api/stockcost/query", handlers.ProcessStockCostHandler)
	authGroup.POST("/api/stockcost/summary", handlers.ProcessStockCostSummaryHandler)
	authGroup.POST("/api/stockcost/check", handlers.ProcessStockCostCheckHandler)

	// Transaction Calculator
	authGroup.POST("/api/transaction/calculate", handlers.TransactionCalculatorHandler)
	authGroup.POST("/api/transaction/quick-calc", handlers.QuickCalculatorHandler)
	authGroup.POST("/api/transaction/validate-payment", handlers.ValidatePaymentHandler)
	authGroup.POST("/api/transaction/purchase-history", handlers.PurchaseHistoryHandler)

	// Sales Report
	authGroup.POST("/api/report/sales/by-document", handlers.SalesReportByDocumentHandler)
	authGroup.POST("/api/report/sales/summary", handlers.SalesReportSummaryHandler)

	// Product Cache
	authGroup.POST("/api/product/search", handlers.ProductSearchHandler)
	authGroup.POST("/api/product/barcode", handlers.ProductBarcodeSearchHandler)
	authGroup.POST("/api/product/barcode/list", handlers.BarcodeListHandler)
	authGroup.GET("/api/search/aliases", handlers.SearchAliasListHandler)
	authGroup.POST("/api/search/aliases", handlers.SearchAliasCreateHandler)
	authGroup.DELETE("/api/search/aliases/:id", handlers.SearchAliasDeleteHandler)
	authGroup.POST("/api/process/product-balance", handlers.ProductBalanceUpdateHandler)
	authGroup.GET("/api/product/cache/stats", handlers.ProductCacheStatsHandler)
	authGroup.POST("/api/product/cache/clear", handlers.ProductCacheClearHandler)
	authGroup.POST("/api/product/search/unified", handlers.UnifiedProductSearchHandler)

	// ชั้นลงขาย (Product Listing API v2) — backend/architecture/product-listing-api-v2.md
	authGroup.POST("/product/v2/item/get", handlers.ProductV2ItemGetHandler)
	authGroup.POST("/product/v2/item/update-listing", handlers.ProductV2ItemUpdateListingHandler)
	authGroup.POST("/product/v2/item/readiness", handlers.ProductV2ItemReadinessHandler)
	authGroup.POST("/product/v2/tier/init", handlers.ProductV2TierInitHandler)
	authGroup.POST("/product/v2/tier/update", handlers.ProductV2TierUpdateHandler)

	// Stock Report Lookup
	authGroup.POST("/api/stock-report/barcodes", handlers.StockReportBarcodesHandler)
	authGroup.POST("/api/stock-report/warehouses", handlers.StockReportWarehousesHandler)

	// Line OA
	authGroup.POST("/api/lineoa/configs", lineoa.GetConfigsHandler)
	authGroup.POST("/api/lineoa/config", lineoa.GetConfigHandler)
	authGroup.POST("/api/lineoa/config/save", lineoa.SaveConfigHandler)
	authGroup.POST("/api/lineoa/employees", lineoa.GetEmployeesHandler)
	authGroup.POST("/api/lineoa/employee/add", lineoa.AddEmployeeHandler)
	authGroup.POST("/api/lineoa/employee/remove", lineoa.RemoveEmployeeHandler)
	authGroup.POST("/api/lineoa/employee/link", lineoa.GenerateLinkHandler)
	authGroup.POST("/api/user/lineoa/link", lineoa.UserLinkHandler)
	authGroup.POST("/api/user/lineoa/callback", lineoa.CallbackHandler)
	authGroup.POST("/api/user/lineoa/profile", lineoa.GetUserProfileHandler)
	g.POST("/api/lineoa/webhook", lineoa.WebhookHandler)
	g.POST("/webhook/lineoa", lineoa.WebhookHandler)

	// Approval System
	authGroup.POST("/api/approval/po-settings", approval.GetPOApprovalSettingsHandler)
	authGroup.POST("/api/approval/po-setting", approval.GetPOApprovalSettingHandler)
	authGroup.POST("/api/approval/po-setting/save", approval.SavePOApprovalSettingHandler)
	authGroup.POST("/api/approval/po-setting/delete", approval.DeletePOApprovalSettingHandler)
	authGroup.POST("/api/approval/po-status/get", approval.GetPOApprovalStatusHandler)
	authGroup.POST("/api/approval/po-status/batch", approval.GetBatchPOApprovalStatusHandler)
	authGroup.POST("/api/approval/po-status/submit", approval.SubmitPOApprovalHandler)
	authGroup.POST("/api/approval/po-status/approve", approval.ApprovePOHandler)
	authGroup.POST("/api/approval/po-status/reject", approval.RejectPOHandler)
	authGroup.POST("/api/approval/po-status/withdraw", approval.WithdrawPOHandler)
	authGroup.POST("/api/approval/po-status/pending", approval.GetPendingApprovalsHandler)
	authGroup.POST("/api/approval/po-status/rejected", approval.GetRejectedPOListHandler)
	authGroup.POST("/api/approval/notification/check", approval.CheckNotificationSentHandler)
	authGroup.POST("/api/approval/notification/send", approval.SendApprovalNotificationHandler)
	authGroup.POST("/api/approval/notification/process", approval.ProcessPendingNotificationsHandler)
	authGroup.POST("/api/approval/notification/logs", approval.GetNotificationLogsHandler)
	authGroup.POST("/api/approval/notification/mark-opened", approval.MarkNotificationOpenedHandler)
	authGroup.POST("/api/approval/timeline", approval.GetApprovalTimelineHandler)
	authGroup.POST("/api/approval/notification/send-real", approval.SendRealApprovalNotificationHandler)
	authGroup.POST("/api/approval/notification/resend", approval.ResendApprovalNotificationHandler)
	authGroup.GET("/api/approval/smtp-status", approval.GetSMTPStatusHandler)
	authGroup.GET("/api/approval/lineoa-config-status", approval.GetLineOAConfigStatusHandler)
	authGroup.GET("/api/approval/lineoa-configs", approval.ListAllLineOAConfigsHandler)
	g.GET("/api/approval/action", approval.ApproveViaTokenHandler)
	g.POST("/api/approval/action", approval.ApproveViaTokenHandler)
	g.GET("/api/approval/token-info", approval.GetApprovalTokenInfoHandler)
	authGroup.POST("/api/approval/po-details", approval.GetPODetailsForLIFFHandler)
	authGroup.POST("/api/approval/liff-approve", approval.LiffApproveHandler)

	// PR Approval System (ใบขอซื้อ) — ใช้ approval engine เดียวกับ PO
	// Frontend ส่ง purchasetypecode = "PR" เพื่อแยกจาก PO
	authGroup.POST("/api/approval/pr-settings", approval.GetPOApprovalSettingsHandler)
	authGroup.POST("/api/approval/pr-setting", approval.GetPOApprovalSettingHandler)
	authGroup.POST("/api/approval/pr-setting/save", approval.SavePOApprovalSettingHandler)
	authGroup.POST("/api/approval/pr-setting/delete", approval.DeletePOApprovalSettingHandler)
	authGroup.POST("/api/approval/pr-status/get", approval.GetPOApprovalStatusHandler)
	authGroup.POST("/api/approval/pr-status/batch", approval.GetBatchPOApprovalStatusHandler)
	authGroup.POST("/api/approval/pr-status/submit", approval.SubmitPOApprovalHandler)
	authGroup.POST("/api/approval/pr-status/approve", approval.ApprovePOHandler)
	authGroup.POST("/api/approval/pr-status/reject", approval.RejectPOHandler)
	authGroup.POST("/api/approval/pr-status/withdraw", approval.WithdrawPOHandler)
	authGroup.POST("/api/approval/pr-status/pending", approval.GetPendingApprovalsHandler)
	authGroup.POST("/api/approval/pr-status/rejected", approval.GetRejectedPOListHandler)

	// RFQ Approval System (สืบราคา) — ใช้ approval engine เดียวกับ PO
	// Frontend ส่ง purchasetypecode = "RFQ" เพื่อแยกจาก PO
	authGroup.POST("/api/approval/rfq-settings", approval.GetPOApprovalSettingsHandler)
	authGroup.POST("/api/approval/rfq-setting", approval.GetPOApprovalSettingHandler)
	authGroup.POST("/api/approval/rfq-setting/save", approval.SavePOApprovalSettingHandler)
	authGroup.POST("/api/approval/rfq-setting/delete", approval.DeletePOApprovalSettingHandler)
	authGroup.POST("/api/approval/rfq-status/get", approval.GetPOApprovalStatusHandler)
	authGroup.POST("/api/approval/rfq-status/batch", approval.GetBatchPOApprovalStatusHandler)
	authGroup.POST("/api/approval/rfq-status/submit", approval.SubmitPOApprovalHandler)
	authGroup.POST("/api/approval/rfq-status/approve", approval.ApprovePOHandler)
	authGroup.POST("/api/approval/rfq-status/reject", approval.RejectPOHandler)
	authGroup.POST("/api/approval/rfq-status/withdraw", approval.WithdrawPOHandler)
	authGroup.POST("/api/approval/rfq-status/pending", approval.GetPendingApprovalsHandler)
	authGroup.POST("/api/approval/rfq-status/rejected", approval.GetRejectedPOListHandler)

	// Data History
	authGroup.GET("/api/datahistory", datahistory.GetHistoryHandler)
	authGroup.GET("/api/datahistory/po", datahistory.GetPOHistoryHandler)

	// Purchase Order Manual Close
	authGroup.POST("/api/purchase-order/manual-close", handlers.ManualClosePOHandler)

	// S3 File Proxy
	authGroup.GET("/s3/file/*", handlers.S3FileProxyHandler)

	// Upload endpoints
	authGroup.POST("/upload", handlers.FileUploadHandler)
	authGroup.POST("/upload/init", handlers.InitChunkedUploadHandler)
	authGroup.POST("/upload/chunk", handlers.UploadChunkHandler)
	authGroup.POST("/upload/merge", handlers.MergeChunksHandler)
	authGroup.GET("/upload/status/:uploadID", handlers.GetUploadStatusHandler)
	authGroup.DELETE("/upload/cancel/:uploadID", handlers.CancelUploadHandler)

	// Language API
	g.GET("/api/language/:lang", handlers.GetLanguageHandler)
	g.GET("/api/address/thailand", handlers.GetThailandAddressHandler)

	// Image endpoints
	authGroup.POST("/image/upload", handlers.ImageUploadHandler)
	authGroup.POST("/video/upload", handlers.VideoUploadHandler)
	authGroup.POST("/image/list", handlers.ImageListHandler)
	authGroup.POST("/image/get", handlers.ImageGetHandler)
	authGroup.POST("/image/info", handlers.ImageInfoHandler)
	authGroup.POST("/image/delete", handlers.ImageDeleteHandler)
	authGroup.POST("/image/promptpayverify", handlers.ImageVerifyHandler)

	// Attachment endpoints (private, shop-scoped)
	authGroup.POST("/api/attachment/upload", handlers.AttachmentUploadHandler)
	authGroup.POST("/api/attachment/list", handlers.AttachmentListHandler)
	authGroup.POST("/api/attachment/delete", handlers.AttachmentDeleteHandler)
	authGroup.GET("/api/attachment/download/:id", handlers.AttachmentDownloadHandler)

	// Excel Product Import
	authGroup.POST("/xlsx/product/start", dataimport.StartProductPrepareHandler)

	// Chatbot API (Gemini)
	chatbotV1 := authGroup.Group("/api/v1/chatbot")
	chatbotV1.POST("/chat-gemini", aichat.ChatGemini)
	chatbotV1.POST("/analyze-document", aichat.AnalyzeDocument)

	// Knowledge Base (RAGFlow-backed)
	knowledgebase.RegisterRoutes(authGroup.Group("/api/v1/kb"))

	// OpenAI-compatible gateway (สำหรับ OpenClaw / client ที่พูด OpenAI protocol)
	openaiGW := authGroup.Group("/api/aichat/v1")
	openaiGW.GET("/models", aichat.OpenAIGatewayListModels)

	// Result store (frontend ดึง full tool result ที่ truncate ใน chat ออกไป)
	authGroup.GET("/api/aichat/result/:id", aichat.GetAIChatResult)

	// AI Provider Config (per-shop settings stored in MongoDB)
	aiProviderV1 := authGroup.Group("/api/v1/ai-provider")
	aiProviderV1.POST("/list", aichat.ListAIProviders)
	aiProviderV1.POST("/save", aichat.SaveAIProvider)
	aiProviderV1.POST("/delete", aichat.DeleteAIProvider)
	aiProviderV1.POST("/models", aichat.ListAIModels)
	aiProviderV1.POST("/status", aichat.AIProviderStatus)
	aiProviderV1.POST("/question-history", aichat.ListQuestionHistory)
	aiProviderV1.POST("/complaint", aichat.SubmitComplaint)

	// Unified API
	unifiedV1 := authGroup.Group("/api/v1/unified")
	unifiedServer := unified.NewUnifiedAPIServer()
	unifiedV1.POST("/query", unifiedServer.ProcessUnifiedQuery)
	unifiedV1.GET("/health", unifiedServer.HealthCheck)
	unifiedV1.GET("/cache/stats", unifiedServer.GetCacheStats)

	logger.Success("GoAPI: ✅ Routes registered successfully")
}

// Shutdown cleans up all goapi resources
func (s *GoAPIServer) Shutdown() {
	logger.Info("GoAPI: กำลังปิดระบบ...")
	if s.workerManager != nil {
		s.workerManager.Stop()
	}
	myglobal.DisconnectMongo()
	handlers.DisconnectMongoAtlas()
	mydb.CloseAllManagers()
	myclickhouse.CloseClickHouseConnection()
	myPg.CleanupPreparedStatements()
	myPg.CloseAllPools()
	logger.Success("GoAPI: ✅ ปิดระบบเรียบร้อย")
}

// ========================================
// Helper functions (ย้ายมาจาก cmd/goapi/main.go)
// ========================================

func createTieredRateLimiter() echo.MiddlewareFunc {
	type RateLimitTier struct {
		limiterConfig middleware.RateLimiterConfig
		rateLimit     int
	}

	tiers := make(map[string]RateLimitTier)

	// Health checks - very high limit
	healthTier := RateLimitTier{
		rateLimit: 1000,
		limiterConfig: middleware.RateLimiterConfig{
			Store: middleware.NewRateLimiterMemoryStoreWithConfig(
				middleware.RateLimiterMemoryStoreConfig{Rate: 1000, Burst: 100, ExpiresIn: 60 * time.Second}),
			IdentifierExtractor: func(ctx echo.Context) (string, error) { return ctx.RealIP(), nil },
			DenyHandler:         createRateLimitDenyHandler(1000),
		},
	}
	tiers["/api/health"] = healthTier
	tiers["/health"] = healthTier

	// Standard APIs
	standardTier := RateLimitTier{
		rateLimit: 200,
		limiterConfig: middleware.RateLimiterConfig{
			Store: middleware.NewRateLimiterMemoryStoreWithConfig(
				middleware.RateLimiterMemoryStoreConfig{Rate: 200, Burst: 20, ExpiresIn: 60 * time.Second}),
			IdentifierExtractor: func(ctx echo.Context) (string, error) { return ctx.RealIP(), nil },
			DenyHandler:         createRateLimitDenyHandler(200),
		},
	}
	tiers["/tokenize"] = standardTier
	tiers["/transliterate"] = standardTier

	tiers["/upload"] = RateLimitTier{
		rateLimit: 50,
		limiterConfig: middleware.RateLimiterConfig{
			Store: middleware.NewRateLimiterMemoryStoreWithConfig(
				middleware.RateLimiterMemoryStoreConfig{Rate: 50, Burst: 10, ExpiresIn: 60 * time.Second}),
			IdentifierExtractor: func(ctx echo.Context) (string, error) { return ctx.RealIP(), nil },
			DenyHandler:         createRateLimitDenyHandler(50),
		},
	}

	// Database query endpoints - high limit
	dbQueryTier := RateLimitTier{
		rateLimit: 500,
		limiterConfig: middleware.RateLimiterConfig{
			Store: middleware.NewRateLimiterMemoryStoreWithConfig(
				middleware.RateLimiterMemoryStoreConfig{Rate: 500, Burst: 50, ExpiresIn: 60 * time.Second}),
			IdentifierExtractor: func(ctx echo.Context) (string, error) { return ctx.RealIP(), nil },
			DenyHandler:         createRateLimitDenyHandler(500),
		},
	}
	tiers["/api/product/search/unified"] = dbQueryTier

	// Default rate limiter
	defaultConfig := middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{Rate: 100, Burst: 10, ExpiresIn: 60 * time.Second}),
		IdentifierExtractor: func(ctx echo.Context) (string, error) { return ctx.RealIP(), nil },
		DenyHandler:         createRateLimitDenyHandler(100),
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Path()
			// Strip group prefix for matching (เช่น /goapi/get → /get)
			for key, tier := range tiers {
				if strings.HasSuffix(path, key) {
					c.Response().Header().Set("X-RateLimit-Limit", formatInt(tier.rateLimit))
					return middleware.RateLimiterWithConfig(tier.limiterConfig)(next)(c)
				}
			}
			c.Response().Header().Set("X-RateLimit-Limit", "100")
			return middleware.RateLimiterWithConfig(defaultConfig)(next)(c)
		}
	}
}

func createRateLimitDenyHandler(limit int) func(echo.Context, string, error) error {
	return func(c echo.Context, identifier string, err error) error {
		c.Response().Header().Set("X-RateLimit-Remaining", "0")
		c.Response().Header().Set("Retry-After", "60")
		return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
			"error":       "Rate limit exceeded",
			"code":        "RATE_LIMIT_EXCEEDED",
			"limit":       limit,
			"retry_after": 60,
		})
	}
}

func formatInt(n int) string {
	if n < 10 {
		return string(rune(n + '0'))
	}
	return fmt.Sprintf("%d", n)
}
