package goapi

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/dataimport"
	"smlcloudplatform/internal/goapi/handlers"
	"smlcloudplatform/internal/goapi/handlers/aichat"
	"smlcloudplatform/internal/goapi/handlers/approval"
	"smlcloudplatform/internal/goapi/handlers/datahistory"
	"smlcloudplatform/internal/goapi/handlers/kafka"
	"smlcloudplatform/internal/goapi/handlers/lineoa"
	"smlcloudplatform/internal/goapi/handlers/unified"
	"smlcloudplatform/internal/goapi/inventory"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mcp"
	"smlcloudplatform/internal/goapi/myclickhouse"
	"smlcloudplatform/internal/goapi/mydb"
	"smlcloudplatform/internal/goapi/myglobal"
	myPg "smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/mypostgres"
	"smlcloudplatform/internal/goapi/workers"

	serviceConfig "smlcloudplatform/internal/goapi/config"
	build "smlcloudplatform/internal/goapi/process/build"
	"smlcloudplatform/internal/goapi/setupconfig"

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

	// 3. MongoDB Atlas (separate connection)
	if err := handlers.InitMongoAtlas(); err != nil {
		logger.Warn("GoAPI: Failed to initialize MongoDB Atlas: %v", err)
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
		Skipper: func(c echo.Context) bool {
			path := c.Path()
			return strings.HasSuffix(path, "/reportget") ||
				strings.HasSuffix(path, "/reportpost") ||
				strings.HasSuffix(path, "/mongogetdata") ||
				strings.HasSuffix(path, "/getdoc") ||
				strings.HasSuffix(path, "/copymongouattodev") ||
				strings.HasSuffix(path, "/previewcopymongo") ||
				strings.HasSuffix(path, "/listsourceshops") ||
				strings.Contains(path, "/rebuild/progress/") ||
				strings.HasSuffix(path, "/mcp/sse") ||
				strings.HasSuffix(path, "/mcp/message")
		},
		ErrorMessage: "Request timeout",
		Timeout:      30 * time.Second,
	}))
}

// RegisterRoutes registers all goapi routes on the Echo group
// prefix คือ URL prefix ของ group เช่น "/goapi" หรือ "" (standalone)
func (s *GoAPIServer) RegisterRoutes(g *echo.Group, prefix string) {
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
		shopId := c.QueryParam("shopid")
		if shopId != "" {
			db, err := myPg.Connect(shopId)
			if err != nil {
				logger.Info("สร้าง Database สำหรับ shopId %s: %v", shopId, err)
				go build.DatabaseChecker(shopId, true)
			} else {
				defer db.Close()
			}
		}
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
	g.GET("/api/health/queue/:shopid", handlers.QueueShopStatusHandler)
	g.GET("/api/health/database", handlers.DatabaseHealthHandler)
	g.GET("/api/health/system", handlers.SystemHealthHandler)

	g.GET("/reportget", handlers.ReportGetHandler)

	// Database operation routes
	g.POST("/get", handlers.PgSelectHandler)
	g.POST("/exec", handlers.PgExecHandler)
	g.POST("/getdoc", handlers.PgGetDocHandler)
	g.POST("/mongogetdata", handlers.MongoGetDataHandler)
	g.POST("/reportpost", handlers.ReportPostHandler)
	g.GET("/rebuild/progress/:jobId", handlers.RebuildProgressSSEHandler)

	// Result table endpoints
	g.POST("/resultfromquery", handlers.ResultFromQueryHandler)
	g.POST("/resultget", handlers.ResultGetHandler)
	g.POST("/resulttopdf", handlers.ResultToPDFHandler)

	// Generate PDF
	g.POST("/genpdf", handlers.GenPDFHandler)
	g.GET("/genpdf/history", handlers.PdfHistoryGetHandler)
	g.POST("/genpdf/history", handlers.PdfHistoryListHandler)
	g.GET("/genpdf/reprint/:id", handlers.PdfReprintHandler)

	// Inventory Costing
	inventoryGroup := g.Group("/api")
	inventory.RegisterRoutes(inventoryGroup)

	// Stock
	g.POST("/processstockcalccost", handlers.ProcessStockCalcCostHandler)
	g.POST("/api/stockcost/query", handlers.ProcessStockCostHandler)
	g.POST("/api/stockcost/summary", handlers.ProcessStockCostSummaryHandler)
	g.POST("/api/stockcost/check", handlers.ProcessStockCostCheckHandler)

	// Transaction Calculator
	g.POST("/api/transaction/calculate", handlers.TransactionCalculatorHandler)
	g.POST("/api/transaction/quick-calc", handlers.QuickCalculatorHandler)
	g.POST("/api/transaction/validate-payment", handlers.ValidatePaymentHandler)
	g.POST("/api/transaction/purchase-history", handlers.PurchaseHistoryHandler)

	// Sales Report
	g.POST("/api/report/sales/by-document", handlers.SalesReportByDocumentHandler)
	g.POST("/api/report/sales/summary", handlers.SalesReportSummaryHandler)

	// Product Cache
	g.POST("/api/product/search", handlers.ProductSearchHandler)
	g.POST("/api/product/barcode", handlers.ProductBarcodeSearchHandler)
	g.POST("/api/product/barcode/list", handlers.BarcodeListHandler)
	g.GET("/api/search/aliases", handlers.SearchAliasListHandler)
	g.POST("/api/search/aliases", handlers.SearchAliasCreateHandler)
	g.DELETE("/api/search/aliases/:id", handlers.SearchAliasDeleteHandler)
	g.POST("/api/process/product-balance", handlers.ProductBalanceUpdateHandler)
	g.GET("/api/product/cache/stats", handlers.ProductCacheStatsHandler)
	g.POST("/api/product/cache/clear", handlers.ProductCacheClearHandler)
	g.POST("/api/product/search/unified", handlers.UnifiedProductSearchHandler)

	// Stock Report Lookup
	g.POST("/api/stock-report/barcodes", handlers.StockReportBarcodesHandler)
	g.POST("/api/stock-report/warehouses", handlers.StockReportWarehousesHandler)

	// Line OA
	g.POST("/api/lineoa/configs", lineoa.GetConfigsHandler)
	g.POST("/api/lineoa/config", lineoa.GetConfigHandler)
	g.POST("/api/lineoa/config/save", lineoa.SaveConfigHandler)
	g.POST("/api/lineoa/test", lineoa.TestHandler)
	g.POST("/api/lineoa/employees", lineoa.GetEmployeesHandler)
	g.POST("/api/lineoa/employee/add", lineoa.AddEmployeeHandler)
	g.POST("/api/lineoa/employee/remove", lineoa.RemoveEmployeeHandler)
	g.POST("/api/lineoa/employee/link", lineoa.GenerateLinkHandler)
	g.POST("/api/user/lineoa/link", lineoa.UserLinkHandler)
	g.POST("/api/user/lineoa/callback", lineoa.CallbackHandler)
	g.POST("/api/user/lineoa/profile", lineoa.GetUserProfileHandler)
	g.POST("/api/lineoa/webhook", lineoa.WebhookHandler)
	g.POST("/webhook/lineoa", lineoa.WebhookHandler)

	// Approval System
	g.POST("/api/approval/po-settings", approval.GetPOApprovalSettingsHandler)
	g.POST("/api/approval/po-setting", approval.GetPOApprovalSettingHandler)
	g.POST("/api/approval/po-setting/save", approval.SavePOApprovalSettingHandler)
	g.POST("/api/approval/po-setting/delete", approval.DeletePOApprovalSettingHandler)
	g.POST("/api/approval/po-status/get", approval.GetPOApprovalStatusHandler)
	g.POST("/api/approval/po-status/batch", approval.GetBatchPOApprovalStatusHandler)
	g.POST("/api/approval/po-status/submit", approval.SubmitPOApprovalHandler)
	g.POST("/api/approval/po-status/approve", approval.ApprovePOHandler)
	g.POST("/api/approval/po-status/reject", approval.RejectPOHandler)
	g.POST("/api/approval/po-status/withdraw", approval.WithdrawPOHandler)
	g.POST("/api/approval/po-status/pending", approval.GetPendingApprovalsHandler)
	g.POST("/api/approval/po-status/rejected", approval.GetRejectedPOListHandler)
	g.POST("/api/approval/notification/check", approval.CheckNotificationSentHandler)
	g.POST("/api/approval/notification/send", approval.SendApprovalNotificationHandler)
	g.POST("/api/approval/notification/process", approval.ProcessPendingNotificationsHandler)
	g.POST("/api/approval/notification/logs", approval.GetNotificationLogsHandler)
	g.POST("/api/approval/notification/mark-opened", approval.MarkNotificationOpenedHandler)
	g.POST("/api/approval/timeline", approval.GetApprovalTimelineHandler)
	g.POST("/api/approval/notification/send-real", approval.SendRealApprovalNotificationHandler)
	g.POST("/api/approval/notification/resend", approval.ResendApprovalNotificationHandler)
	g.GET("/api/approval/smtp-status", approval.GetSMTPStatusHandler)
	g.POST("/api/approval/test-email", approval.SendTestEmailHandler)
	g.GET("/api/approval/lineoa-config-status", approval.GetLineOAConfigStatusHandler)
	g.GET("/api/approval/lineoa-configs", approval.ListAllLineOAConfigsHandler)
	g.POST("/api/approval/test-line-push", approval.TestLinePushHandler)
	g.GET("/api/approval/action", approval.ApproveViaTokenHandler)
	g.POST("/api/approval/action", approval.ApproveViaTokenHandler)
	g.GET("/api/approval/token-info", approval.GetApprovalTokenInfoHandler)
	g.POST("/api/approval/po-details", approval.GetPODetailsForLIFFHandler)
	g.POST("/api/approval/liff-approve", approval.LiffApproveHandler)

	// PR Approval System (ใบขอซื้อ) — ใช้ approval engine เดียวกับ PO
	// Frontend ส่ง purchase_type_code = "PR" เพื่อแยกจาก PO
	g.POST("/api/approval/pr-settings", approval.GetPOApprovalSettingsHandler)
	g.POST("/api/approval/pr-setting", approval.GetPOApprovalSettingHandler)
	g.POST("/api/approval/pr-setting/save", approval.SavePOApprovalSettingHandler)
	g.POST("/api/approval/pr-setting/delete", approval.DeletePOApprovalSettingHandler)
	g.POST("/api/approval/pr-status/get", approval.GetPOApprovalStatusHandler)
	g.POST("/api/approval/pr-status/batch", approval.GetBatchPOApprovalStatusHandler)
	g.POST("/api/approval/pr-status/submit", approval.SubmitPOApprovalHandler)
	g.POST("/api/approval/pr-status/approve", approval.ApprovePOHandler)
	g.POST("/api/approval/pr-status/reject", approval.RejectPOHandler)
	g.POST("/api/approval/pr-status/withdraw", approval.WithdrawPOHandler)
	g.POST("/api/approval/pr-status/pending", approval.GetPendingApprovalsHandler)
	g.POST("/api/approval/pr-status/rejected", approval.GetRejectedPOListHandler)

	// RFQ Approval System (สืบราคา) — ใช้ approval engine เดียวกับ PO
	// Frontend ส่ง purchase_type_code = "RFQ" เพื่อแยกจาก PO
	g.POST("/api/approval/rfq-settings", approval.GetPOApprovalSettingsHandler)
	g.POST("/api/approval/rfq-setting", approval.GetPOApprovalSettingHandler)
	g.POST("/api/approval/rfq-setting/save", approval.SavePOApprovalSettingHandler)
	g.POST("/api/approval/rfq-setting/delete", approval.DeletePOApprovalSettingHandler)
	g.POST("/api/approval/rfq-status/get", approval.GetPOApprovalStatusHandler)
	g.POST("/api/approval/rfq-status/batch", approval.GetBatchPOApprovalStatusHandler)
	g.POST("/api/approval/rfq-status/submit", approval.SubmitPOApprovalHandler)
	g.POST("/api/approval/rfq-status/approve", approval.ApprovePOHandler)
	g.POST("/api/approval/rfq-status/reject", approval.RejectPOHandler)
	g.POST("/api/approval/rfq-status/withdraw", approval.WithdrawPOHandler)
	g.POST("/api/approval/rfq-status/pending", approval.GetPendingApprovalsHandler)
	g.POST("/api/approval/rfq-status/rejected", approval.GetRejectedPOListHandler)

	// Data History
	g.GET("/api/datahistory", datahistory.GetHistoryHandler)
	g.GET("/api/datahistory/po", datahistory.GetPOHistoryHandler)

	// Purchase Order Manual Close
	g.POST("/api/purchase-order/manual-close", handlers.ManualClosePOHandler)

	// Migration
	g.GET("/api/migrate/currency", handlers.MigrateCurrencyColumnsHandler)
	g.GET("/api/migrate/currency-backfill", handlers.BackfillCurrencyDataHandler)
	g.GET("/api/migrate/clickhouse-softdelete", handlers.MigrateClickHouseSoftDeleteHandler)

	// MongoDB copy
	g.POST("/copymongouattodev", handlers.CopyMongoUatToDevHandler)
	g.POST("/previewcopymongo", handlers.PreviewCopyMongoHandler)
	g.GET("/listsourceshops", handlers.ListSourceShopsHandler)

	// MongoDB Atlas
	g.POST("/atlas/get", handlers.MongoAtlasGetHandler)
	g.POST("/atlas/update", handlers.MongoAtlasUpdateHandler)
	g.POST("/atlas/delete", handlers.MongoAtlasDeleteHandler)

	// ClickHouse
	g.POST("/clickhouse/query", handlers.ClickHouseQueryHandler)
	g.POST("/clickhouse/querys", handlers.ClickHouseMultiQueryHandler)
	g.POST("/clickhouse/select", handlers.ClickHouseSelectHandler)

	// Test endpoints
	g.POST("/test/sale-order", kafka.TestSaleOrderHandler)
	g.POST("/test/purchase", kafka.TestPurchaseHandler)
	g.POST("/test/purchase-order", kafka.TestPurchaseOrderHandler)
	g.POST("/test/purchase-partial", kafka.TestPurchasePartialHandler)

	// S3 File Proxy
	g.GET("/s3/file/*", handlers.S3FileProxyHandler)

	// Upload endpoints
	g.POST("/upload", handlers.FileUploadHandler)
	g.POST("/upload/init", handlers.InitChunkedUploadHandler)
	g.POST("/upload/chunk", handlers.UploadChunkHandler)
	g.POST("/upload/merge", handlers.MergeChunksHandler)
	g.GET("/upload/status/:uploadID", handlers.GetUploadStatusHandler)
	g.DELETE("/upload/cancel/:uploadID", handlers.CancelUploadHandler)

	// Language API
	g.GET("/api/language/:lang", handlers.GetLanguageHandler)

	// Image endpoints
	g.POST("/image/upload", handlers.ImageUploadHandler)
	g.POST("/image/list", handlers.ImageListHandler)
	g.POST("/image/get", handlers.ImageGetHandler)
	g.POST("/image/info", handlers.ImageInfoHandler)
	g.POST("/image/delete", handlers.ImageDeleteHandler)
	g.POST("/image/promptpayverify", handlers.ImageVerifyHandler)

	// Excel Product Import
	g.POST("/xlsx/product/start", dataimport.StartProductPrepareHandler)

	// Chatbot API (Gemini)
	chatbotV1 := g.Group("/api/v1/chatbot")
	chatbotV1.POST("/chat-gemini", aichat.ChatGemini)
	chatbotV1.POST("/chat-agent", aichat.ChatAgent)
	chatbotV1.POST("/analyze-document", aichat.AnalyzeDocument)

	// Unified API
	unifiedV1 := g.Group("/api/v1/unified")
	unifiedServer := unified.NewUnifiedAPIServer()
	unifiedV1.POST("/query", unifiedServer.ProcessUnifiedQuery)
	unifiedV1.GET("/health", unifiedServer.HealthCheck)
	unifiedV1.GET("/cache/stats", unifiedServer.GetCacheStats)

	// MCP Server Routes (prefix-aware)
	mcpServer := mcp.NewMCPServer()
	mcpServer.RegisterRoutesOnGroup(g)
	mcpServer.RegisterSSERoutesOnGroup(g, prefix)

	// MCP API Key Management
	mcpAPIKeyHandler := handlers.NewMCPAPIKeyHandler()
	g.POST("/api/mcp/keys", mcpAPIKeyHandler.CreateAPIKeyHandler)
	g.GET("/api/mcp/keys", mcpAPIKeyHandler.ListAPIKeysHandler)
	g.GET("/api/mcp/keys/:id", mcpAPIKeyHandler.GetAPIKeyHandler)
	g.PUT("/api/mcp/keys/:id", mcpAPIKeyHandler.UpdateAPIKeyHandler)
	g.DELETE("/api/mcp/keys/:id", mcpAPIKeyHandler.DeleteAPIKeyHandler)
	g.GET("/api/mcp/keys/:id/export", mcpAPIKeyHandler.ExportAPIKeyHandler)
	g.POST("/api/mcp/keys/create-with-export", mcpAPIKeyHandler.CreateAPIKeyWithExportHandler)
	g.GET("/api/mcp/audit-logs", mcpAPIKeyHandler.GetAuditLogsHandler)
	g.GET("/api/mcp/available-tools", mcpAPIKeyHandler.GetAvailableToolsHandler)

	// Setup Config Routes
	g.POST("/api/setup/verify-password", handlers.SetupVerifyPasswordHandler)
	g.POST("/api/setup/change-password", handlers.SetupChangePasswordHandler)
	g.POST("/api/setup/config/get", handlers.SetupGetConfigHandler)
	g.POST("/api/setup/config/get-raw", handlers.SetupGetConfigRawHandler)
	g.POST("/api/setup/config/save", handlers.SetupSaveConfigHandler)
	g.POST("/api/setup/config/seed", handlers.SetupSeedConfigHandler)
	g.POST("/api/setup/test-connection", handlers.SetupTestConnectionHandler)
	g.POST("/api/setup/create-clickhouse-database", handlers.SetupCreateClickHouseDatabaseHandler)
	g.GET("/api/setup/client-config", handlers.SetupClientConfigHandler)

	// Deploy Webhook Routes
	g.POST("/api/deploy/backend", handlers.DeployBackendHandler)
	g.POST("/api/deploy/frontend", handlers.DeployFrontendHandler)
	g.GET("/api/deploy/status", handlers.DeployStatusHandler)

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

	// Heavy operations - low limit
	heavyTier := RateLimitTier{
		rateLimit: 10,
		limiterConfig: middleware.RateLimiterConfig{
			Store: middleware.NewRateLimiterMemoryStoreWithConfig(
				middleware.RateLimiterMemoryStoreConfig{Rate: 10, Burst: 2, ExpiresIn: 60 * time.Second}),
			IdentifierExtractor: func(ctx echo.Context) (string, error) { return ctx.RealIP(), nil },
			DenyHandler:         createRateLimitDenyHandler(10),
		},
	}
	tiers["/reportget"] = heavyTier
	tiers["/reportpost"] = heavyTier

	tiers["/mongogetdata"] = RateLimitTier{
		rateLimit: 20,
		limiterConfig: middleware.RateLimiterConfig{
			Store: middleware.NewRateLimiterMemoryStoreWithConfig(
				middleware.RateLimiterMemoryStoreConfig{Rate: 20, Burst: 5, ExpiresIn: 60 * time.Second}),
			IdentifierExtractor: func(ctx echo.Context) (string, error) { return ctx.RealIP(), nil },
			DenyHandler:         createRateLimitDenyHandler(20),
		},
	}

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
	tiers["/get"] = dbQueryTier
	tiers["/pg/select"] = dbQueryTier
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

func splitAndTrim(s string, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func formatInt(n int) string {
	if n < 10 {
		return string(rune(n + '0'))
	}
	return fmt.Sprintf("%d", n)
}
