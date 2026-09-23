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

	"smlcloudplatform/internal/goapi/handlers"
	"smlcloudplatform/internal/goapi/inventory"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mydb"
	"smlcloudplatform/internal/goapi/myglobal"
	myPg "smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/goapi/mypostgres"
	processstock "smlcloudplatform/internal/goapi/process/process-stock"
	"smlcloudplatform/internal/goapi/process/stockengine"

	"smlcloudplatform/internal/goapi/setupconfig"
	"smlcloudplatform/pkg/microservice"
	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const VersionNumber = "1.1.1121"

// GoAPIServer encapsulates all goapi services and routes
type GoAPIServer struct{}

// New creates a new GoAPIServer instance
func New() *GoAPIServer {
	return &GoAPIServer{}
}

// Init initializes all goapi services (DB, workers, background tasks)
func (s *GoAPIServer) Init() error {
	logger.Info("GoAPI: เริ่มต้น initialization...")

	// 1. Load goapi bootstrap config (set goapi-specific env vars)
	setupconfig.LoadBootstrapConfig()

	// 5. S3/R2 Client
	if err := handlers.InitR2Client(); err != nil {
		logger.Warn("GoAPI: Failed to initialize R2 client: %v", err)
	}

	// 6. Database Manager Pool (PostgreSQL)
	logger.Info("GoAPI: กำลังเริ่มต้น Database Manager Pool...")
	dbConfig := mydb.DatabaseManagerConfig{
		PostgreSQLHost:     os.Getenv("POSTGRES_HOST"),
		PostgreSQLPort:     os.Getenv("POSTGRES_PORT"),
		PostgreSQLUser:     os.Getenv("POSTGRES_USER"),
		PostgreSQLPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgreSQLSSLMode:  os.Getenv("POSTGRES_SSL_MODE"),
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

	// 9. Background goroutines
	go handlers.StartConnectionHealthChecker()
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			myPg.CleanupPreparedStatements()
		}
	}()

	// 10. Stock engine workers (คิดต้นทุนจากคิวแทนการคิดคาสด ๆ ในตัว consumer)
	//
	// ยอดคงเหลือในตารางสินค้าต้องอัปเดตหลังคำนวณเสร็จเท่านั้น ไม่ใช่ตอนรับเอกสาร
	// เพราะตอนรับเอกสารสมุดสต็อกยังไม่ถูกเขียน จอสินค้าจะเห็นยอดของรอบก่อน
	stockengine.AfterRecalculate = func(db *sql.DB, holdingCode, businessCode, itemCode string) {
		if err := processstock.ProcessProductBalanceUpdateByItems(db, holdingCode, businessCode, []string{itemCode}); err != nil {
			logger.Error("update product balance for %s/%s: %v", businessCode, itemCode, err)
		}
	}
	if stockengine.WorkerEnabledFromEnv() {
		stockengine.StartWorkers(context.Background(), myglobal.TransFlagsToProcess)
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

			userInfo, ok := authenticateGoAPIToken(c.Request().Context(), authService, tokenText)
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

func authenticateGoAPIToken(ctx context.Context, authService *microservice.AuthService, tokenText string) (msmodels.UserInfo, bool) {
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
func (s *GoAPIServer) RegisterRoutes(g *echo.Group, prefix string, cacher microservice.ICacher, authorizationFinders ...microservice.AuthorizationFinder) {
	authService := microservice.NewAuthService(cacher, 15*time.Minute, 8*time.Hour, authorizationFinders...)
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

	// Transaction calculator
	authGroup.POST("/api/transaction/calculate", handlers.TransactionCalculatorHandler)
	authGroup.POST("/api/transaction/quick-calc", handlers.QuickCalculatorHandler)
	authGroup.POST("/api/transaction/validate-payment", handlers.ValidatePaymentHandler)
	authGroup.POST("/api/transaction/purchase-history", handlers.PurchaseHistoryHandler)

	// Reports (PostgreSQL)
	authGroup.POST("/api/report/sales/by-document", handlers.SalesReportByDocumentHandler)
	authGroup.POST("/api/report/sales/summary", handlers.SalesReportSummaryHandler)
	authGroup.POST("/api/report/tax/vat-register", handlers.TaxVatRegisterHandler)
	authGroup.POST("/api/report/tax/pp30-summary", handlers.PP30SummaryHandler)
	authGroup.POST("/api/report/tax/wht", handlers.TaxWithholdingHandler)
	authGroup.POST("/api/report/tax/wht/certificate", handlers.WhtCertificateHandler)
	authGroup.POST("/api/report/debt/query", handlers.DebtReportHandler)

	// Product search
	authGroup.POST("/api/product/search", handlers.ProductSearchHandler)
	authGroup.POST("/api/product/barcode", handlers.ProductBarcodeSearchHandler)
	authGroup.GET("/api/search/aliases", handlers.SearchAliasListHandler)
	authGroup.POST("/api/search/aliases", handlers.SearchAliasCreateHandler)
	authGroup.DELETE("/api/search/aliases/:id", handlers.SearchAliasDeleteHandler)
	authGroup.POST("/api/process/product-balance", handlers.ProductBalanceUpdateHandler)
	authGroup.POST("/api/process/queue-status", handlers.StockQueueStatusHandler)
	authGroup.GET("/api/product/cache/stats", handlers.ProductCacheStatsHandler)
	authGroup.POST("/api/product/cache/clear", handlers.ProductCacheClearHandler)
	authGroup.POST("/api/product/search/unified", handlers.UnifiedProductSearchHandler)

	// Stock report lookups
	authGroup.POST("/api/stock-report/barcodes", handlers.StockReportBarcodesHandler)
	authGroup.POST("/api/stock-report/warehouses", handlers.StockReportWarehousesHandler)

	// Files (S3 / MinIO)
	authGroup.GET("/s3/file/*", handlers.S3FileProxyHandler)
	authGroup.POST("/upload", handlers.FileUploadHandler)
	authGroup.POST("/upload/init", handlers.InitChunkedUploadHandler)
	authGroup.POST("/upload/chunk", handlers.UploadChunkHandler)
	authGroup.POST("/upload/merge", handlers.MergeChunksHandler)
	authGroup.GET("/upload/status/:uploadID", handlers.GetUploadStatusHandler)
	authGroup.DELETE("/upload/cancel/:uploadID", handlers.CancelUploadHandler)
	authGroup.POST("/image/upload", handlers.ImageUploadHandler)
	authGroup.POST("/video/upload", handlers.VideoUploadHandler)

	// Language & address (public)
	g.GET("/api/language/:lang", handlers.GetLanguageHandler)
	g.GET("/api/address/thailand", handlers.GetThailandAddressHandler)

	logger.Success("GoAPI: ✅ Routes registered successfully")
}

// Shutdown cleans up all goapi resources
func (s *GoAPIServer) Shutdown() {
	logger.Info("GoAPI: กำลังปิดระบบ...")
	mydb.CloseAllManagers()
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
