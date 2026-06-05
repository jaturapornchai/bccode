package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/goapi/setupconfig"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/labstack/echo/v4"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	_ "github.com/lib/pq"
)

// ==========================================
// Setup Config Handler
// จัดการ config ระบบ (API URLs, Database connections)
// เก็บใน MongoDB collection: systemconfig
// ==========================================

const (
	setupConfigCollection   = "systemconfig"
	setupPasswordCollection = "systemsetuppassword"
	defaultSetupPassword    = "12345"
)

// SetupConfigEntry โครงสร้างข้อมูล config
type SetupConfigEntry struct {
	Category    string    `json:"category" bson:"category"`       // "serviceurls", "mongodb", "postgresql", "clickhouse", "redis", "kafka", "integrations"
	Key         string    `json:"key" bson:"key"`                 // เช่น "mainapiurl", "MONGODB_URI"
	Value       string    `json:"value" bson:"value"`             // ค่า config
	IsSecret    bool      `json:"issecret" bson:"issecret"`       // true = แสดงเป็น *** ใน response
	Description string    `json:"description" bson:"description"` // คำอธิบาย
	UpdatedAt   time.Time `json:"updatedat" bson:"updatedat"`
	UpdatedBy   string    `json:"updatedby" bson:"updatedby"`
}

// SetupPasswordDoc โครงสร้างเก็บ password
type SetupPasswordDoc struct {
	PasswordHash string    `json:"passwordhash" bson:"passwordhash"`
	UpdatedAt    time.Time `json:"updatedat" bson:"updatedat"`
}

// ==========================================
// Helper Functions
// ==========================================

func getSetupDB() *mongo.Database {
	return myglobal.GetMongoDatabase()
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func verifySetupPassword(password string) bool {
	db := getSetupDB()
	if db == nil {
		logger.Error("[SetupConfig] ไม่สามารถเชื่อมต่อ MongoDB ได้")
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var doc SetupPasswordDoc
	err := db.Collection(setupPasswordCollection).FindOne(ctx, bson.M{}).Decode(&doc)
	if err != nil {
		// ถ้ายังไม่มี password document → ใช้ default password
		if err == mongo.ErrNoDocuments {
			return password == defaultSetupPassword
		}
		logger.Error("[SetupConfig] ตรวจสอบ password ล้มเหลว: %v", err)
		return false
	}

	return doc.PasswordHash == hashPassword(password)
}

// ==========================================
// Verify Password Handler
// POST /api/setup/verify-password
// ==========================================

func SetupVerifyPasswordHandler(c echo.Context) error {
	var req struct {
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "ข้อมูลไม่ถูกต้อง",
		})
	}

	if req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "กรุณาระบุรหัสผ่าน",
		})
	}

	if !verifySetupPassword(req.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "รหัสผ่านไม่ถูกต้อง",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "ยืนยันรหัสผ่านสำเร็จ",
	})
}

// ==========================================
// Change Password Handler
// POST /api/setup/change-password
// ==========================================

func SetupChangePasswordHandler(c echo.Context) error {
	var req struct {
		CurrentPassword string `json:"currentpassword"`
		NewPassword     string `json:"newpassword"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "ข้อมูลไม่ถูกต้อง",
		})
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "กรุณาระบุรหัสผ่านเดิมและรหัสผ่านใหม่",
		})
	}

	if len(req.NewPassword) < 4 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "รหัสผ่านใหม่ต้องมีอย่างน้อย 4 ตัวอักษร",
		})
	}

	if !verifySetupPassword(req.CurrentPassword) {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "รหัสผ่านเดิมไม่ถูกต้อง",
		})
	}

	db := getSetupDB()
	if db == nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "ไม่สามารถเชื่อมต่อ MongoDB ได้",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newDoc := SetupPasswordDoc{
		PasswordHash: hashPassword(req.NewPassword),
		UpdatedAt:    time.Now(),
	}

	opts := options.Replace().SetUpsert(true)
	_, err := db.Collection(setupPasswordCollection).ReplaceOne(ctx, bson.M{}, newDoc, opts)
	if err != nil {
		logger.Error("[SetupConfig] เปลี่ยนรหัสผ่านล้มเหลว: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "เปลี่ยนรหัสผ่านล้มเหลว: " + err.Error(),
		})
	}

	logger.Info("[SetupConfig] เปลี่ยนรหัสผ่าน setup สำเร็จ")
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "เปลี่ยนรหัสผ่านสำเร็จ",
	})
}

// ==========================================
// Get Config Handler
// POST /api/setup/config/get
// ==========================================

func SetupGetConfigHandler(c echo.Context) error {
	var req struct {
		Password string `json:"password"`
		Category string `json:"category"` // optional: filter by category
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "ข้อมูลไม่ถูกต้อง",
		})
	}

	if req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "กรุณาระบุรหัสผ่าน",
		})
	}

	if !verifySetupPassword(req.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "รหัสผ่านไม่ถูกต้อง",
		})
	}

	// อ่าน config จาก bootstrap.json (ไม่ใช้ MongoDB)
	configs, err := setupconfig.ReadBootstrapAsConfigEntries()
	if err != nil {
		logger.Error("[SetupConfig] อ่าน bootstrap.json ล้มเหลว: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "อ่าน bootstrap.json ล้มเหลว: " + err.Error(),
		})
	}

	// กรอง category ถ้าระบุ
	if req.Category != "" {
		filtered := make([]setupconfig.BootstrapConfigEntry, 0)
		for _, cfg := range configs {
			if cfg.Category == req.Category {
				filtered = append(filtered, cfg)
			}
		}
		configs = filtered
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    configs,
	})
}

// ==========================================
// Get Config Raw Handler (ดึงค่า secret จริง)
// POST /api/setup/config/get-raw
// ==========================================

func SetupGetConfigRawHandler(c echo.Context) error {
	var req struct {
		Password string `json:"password"`
		Category string `json:"category"`
		Key      string `json:"key"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "ข้อมูลไม่ถูกต้อง",
		})
	}

	if req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "กรุณาระบุรหัสผ่าน",
		})
	}

	if !verifySetupPassword(req.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "รหัสผ่านไม่ถูกต้อง",
		})
	}

	// อ่าน config จาก bootstrap.json (ไม่ใช้ MongoDB)
	allConfigs, err := setupconfig.ReadBootstrapAsConfigEntries()
	if err != nil {
		logger.Error("[SetupConfig] อ่าน bootstrap.json ล้มเหลว: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "อ่าน bootstrap.json ล้มเหลว: " + err.Error(),
		})
	}

	// กรอง category + key ถ้าระบุ
	configs := make([]setupconfig.BootstrapConfigEntry, 0)
	for _, cfg := range allConfigs {
		if req.Category != "" && cfg.Category != req.Category {
			continue
		}
		if req.Key != "" && cfg.Key != req.Key {
			continue
		}
		configs = append(configs, cfg)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    configs,
	})
}

// ==========================================
// Save Config Handler
// POST /api/setup/config/save
// ==========================================

func SetupSaveConfigHandler(c echo.Context) error {
	var req struct {
		Password string             `json:"password"`
		Configs  []SetupConfigEntry `json:"configs"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "ข้อมูลไม่ถูกต้อง",
		})
	}

	if req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "กรุณาระบุรหัสผ่าน",
		})
	}

	if !verifySetupPassword(req.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "รหัสผ่านไม่ถูกต้อง",
		})
	}

	if len(req.Configs) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "กรุณาระบุ config อย่างน้อย 1 รายการ",
		})
	}

	// เขียน config ลง bootstrap.json เท่านั้น (ไม่เขียน MongoDB)
	bootstrapEntries := make([]setupconfig.ConfigUpdateEntry, 0, len(req.Configs))
	for _, cfg := range req.Configs {
		if cfg.Value == "***" || cfg.Category == "" || cfg.Key == "" {
			continue
		}
		bootstrapEntries = append(bootstrapEntries, setupconfig.ConfigUpdateEntry{
			Category: cfg.Category,
			Key:      cfg.Key,
			Value:    cfg.Value,
		})
	}

	if err := setupconfig.UpdateBootstrapJSON(bootstrapEntries); err != nil {
		logger.Error("[SetupConfig] เขียน bootstrap.json ล้มเหลว: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "เขียน bootstrap.json ล้มเหลว: " + err.Error(),
		})
	}

	savedCount := len(bootstrapEntries)
	logger.Info("[SetupConfig] บันทึก config ลง bootstrap.json สำเร็จ %d รายการ", savedCount)

	// Auto-reload: อ่าน config ใหม่จาก bootstrap.json + reconnect databases ทันที
	go func() {
		setupconfig.ReloadAndReconnect()
	}()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("บันทึก config ลง bootstrap.json สำเร็จ %d รายการ (กำลัง reload connections...)", savedCount),
	})
}

// ==========================================
// Client Config Handler (PUBLIC — ไม่ต้อง password)
// GET /api/setup/client-config
// Flutter apps เรียกตอน startup เพื่อรับ API URLs
// ==========================================

func SetupClientConfigHandler(c echo.Context) error {
	db := getSetupDB()
	if db == nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "ไม่สามารถเชื่อมต่อ MongoDB ได้",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ดึงเฉพาะ category = serviceurls (ไม่เป็น secret)
	filter := bson.M{
		"category": "serviceurls",
		"issecret": bson.M{"$ne": true},
	}

	cursor, err := db.Collection(setupConfigCollection).Find(ctx, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "ดึง config ล้มเหลว",
		})
	}
	defer cursor.Close(ctx)

	var configs []SetupConfigEntry
	if err := cursor.All(ctx, &configs); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "อ่าน config ล้มเหลว",
		})
	}

	// สร้าง key-value map สำหรับ client
	data := make(map[string]string)
	for _, cfg := range configs {
		data[cfg.Key] = cfg.Value
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// ==========================================
// Test Connection Handler
// POST /api/setup/test-connection
// ==========================================

func SetupTestConnectionHandler(c echo.Context) error {
	var req struct {
		Password  string `json:"password"`
		Type      string `json:"type"`      // "mongodb", "postgresql", "clickhouse", "redis", "kafka", "http"
		URI       string `json:"uri"`       // สำหรับ mongodb
		Host      string `json:"host"`      // สำหรับ pg, ch, redis, kafka
		Port      string `json:"port"`      // สำหรับ pg, ch, redis, kafka
		User      string `json:"user"`      // สำหรับ pg, ch
		Password2 string `json:"password2"` // สำหรับ pg, ch (ใช้ password2 เพราะ password ถูกใช้สำหรับ setup password)
		Database  string `json:"database"`  // สำหรับ pg, ch
		URL       string `json:"url"`       // สำหรับ http test
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "ข้อมูลไม่ถูกต้อง",
		})
	}

	if req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "กรุณาระบุรหัสผ่าน",
		})
	}

	if !verifySetupPassword(req.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "รหัสผ่านไม่ถูกต้อง",
		})
	}

	start := time.Now()

	switch req.Type {
	case "mongodb":
		return testMongoDBConnection(c, req.URI, req.Database, start)
	case "postgresql":
		return testPostgreSQLConnection(c, req.Host, req.Port, req.User, req.Password2, req.Database, start)
	case "clickhouse":
		return testClickHouseConnection(c, req.Host, req.Port, req.User, req.Password2, req.Database, start)
	case "redis":
		return testRedisConnection(c, req.Host, req.Port, start)
	case "kafka":
		return testKafkaConnection(c, req.Host, req.Port, start)
	case "http":
		return testHTTPConnection(c, req.URL, start)
	default:
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("ประเภทการเชื่อมต่อไม่ถูกต้อง: %s", req.Type),
		})
	}
}

// ==========================================
// Connection Test Functions
// ==========================================

func testMongoDBConnection(c echo.Context, uri string, database string, start time.Time) error {
	if uri == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "กรุณาระบุ MongoDB URI",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   false,
			"message":   "เชื่อมต่อ MongoDB ล้มเหลว: " + err.Error(),
			"latencyms": time.Since(start).Milliseconds(),
		})
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   false,
			"message":   "Ping MongoDB ล้มเหลว: " + err.Error(),
			"latencyms": time.Since(start).Milliseconds(),
		})
	}

	// ดึงรายชื่อ database ที่มีอยู่
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   false,
			"message":   "ดึงรายชื่อ database ล้มเหลว: " + err.Error(),
			"latencyms": time.Since(start).Milliseconds(),
		})
	}

	// ตรวจสอบว่า database ที่ระบุมีอยู่หรือไม่
	if database != "" {
		dbExists := false
		for _, dbName := range databases {
			if dbName == database {
				dbExists = true
				break
			}
		}

		if !dbExists {
			// database ไม่มี → สร้างอัตโนมัติโดย create collection systemconfig
			logger.Info("[MongoDB] database '%s' ไม่มี — กำลังสร้างอัตโนมัติ...", database)
			db := client.Database(database)
			createErr := db.CreateCollection(ctx, "systemconfig")
			if createErr != nil {
				// ถ้า collection มีอยู่แล้ว ถือว่า database มีอยู่แล้ว → ไม่ใช่ error
				errMsg := createErr.Error()
				if !strings.Contains(errMsg, "already exists") && !strings.Contains(errMsg, "NamespaceExists") {
					return c.JSON(http.StatusOK, map[string]interface{}{
						"success":   false,
						"message":   fmt.Sprintf("สร้าง database '%s' ล้มเหลว: %s", database, errMsg),
						"latencyms": time.Since(start).Milliseconds(),
					})
				}
			}
			logger.Info("[MongoDB] สร้าง database '%s' สำเร็จ", database)
			return c.JSON(http.StatusOK, map[string]interface{}{
				"success":   true,
				"message":   fmt.Sprintf("เชื่อมต่อ MongoDB สำเร็จ (สร้าง database '%s' ใหม่)", database),
				"latencyms": time.Since(start).Milliseconds(),
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":   true,
		"message":   fmt.Sprintf("เชื่อมต่อ MongoDB สำเร็จ (พบ %d databases)", len(databases)),
		"latencyms": time.Since(start).Milliseconds(),
	})
}

func testPostgreSQLConnection(c echo.Context, host, port, user, password, database string, start time.Time) error {
	if host == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "กรุณาระบุ PostgreSQL host",
		})
	}
	if port == "" {
		port = "5432"
	}
	if database == "" {
		database = "postgres"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=10",
		host, port, user, password, database)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   false,
			"message":   "เปิดการเชื่อมต่อ PostgreSQL ล้มเหลว: " + err.Error(),
			"latencyms": time.Since(start).Milliseconds(),
		})
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   false,
			"message":   "Ping PostgreSQL ล้มเหลว: " + err.Error(),
			"latencyms": time.Since(start).Milliseconds(),
		})
	}

	// ดึง version
	var version string
	_ = db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	versionShort := ""
	if version != "" {
		parts := strings.SplitN(version, ",", 2)
		versionShort = " (" + parts[0] + ")"
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":   true,
		"message":   "เชื่อมต่อ PostgreSQL สำเร็จ" + versionShort,
		"latencyms": time.Since(start).Milliseconds(),
	})
}

func testClickHouseConnection(c echo.Context, host, port, user, password, database string, start time.Time) error {
	if host == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "กรุณาระบุ ClickHouse host",
		})
	}
	if port == "" {
		port = "9000"
	}
	if database == "" {
		database = "default"
	}

	// แยก host:port ถ้า host มี : อยู่แล้ว
	addr := host
	if !strings.Contains(host, ":") {
		addr = host + ":" + port
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: database,
			Username: user,
			Password: password,
		},
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   false,
			"message":   "เปิดการเชื่อมต่อ ClickHouse ล้มเหลว: " + err.Error(),
			"latencyms": time.Since(start).Milliseconds(),
		})
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		// ตรวจ error code 81 = UNKNOWN_DATABASE (database ยังไม่ถูกสร้าง)
		var chException *clickhouse.Exception
		if errors.As(err, &chException) && chException.Code == 81 {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"success":   false,
				"message":   fmt.Sprintf("Ping ClickHouse ล้มเหลว: %s", err.Error()),
				"errorcode": "databasenotexist",
				"database":  database,
				"latencyms": time.Since(start).Milliseconds(),
			})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   false,
			"message":   "Ping ClickHouse ล้มเหลว: " + err.Error(),
			"latencyms": time.Since(start).Milliseconds(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":   true,
		"message":   fmt.Sprintf("เชื่อมต่อ ClickHouse สำเร็จ (database: %s)", database),
		"latencyms": time.Since(start).Milliseconds(),
	})
}

// ==========================================
// POST /api/setup/create-clickhouse-database
// สร้าง ClickHouse database ถ้ายังไม่มี
// ==========================================

func SetupCreateClickHouseDatabaseHandler(c echo.Context) error {
	var req struct {
		Password  string `json:"password"`
		Host      string `json:"host"`
		Port      string `json:"port"`
		User      string `json:"user"`
		Password2 string `json:"password2"`
		Database  string `json:"database"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "ข้อมูลไม่ถูกต้อง",
		})
	}
	if !verifySetupPassword(req.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "รหัสผ่านไม่ถูกต้อง",
		})
	}
	if req.Host == "" || req.Database == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "กรุณาระบุ host และ database",
		})
	}

	port := req.Port
	if port == "" {
		port = "9000"
	}
	addr := req.Host
	if !strings.Contains(req.Host, ":") {
		addr = req.Host + ":" + port
	}

	// เชื่อมต่อโดยใช้ "default" database เพื่อสร้าง database ใหม่
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: req.User,
			Password: req.Password2,
		},
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": false,
			"message": "เชื่อมต่อ ClickHouse ล้มเหลว: " + err.Error(),
		})
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", req.Database)
	if err := conn.Exec(ctx, query); err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": false,
			"message": "สร้าง database ล้มเหลว: " + err.Error(),
		})
	}

	logger.Info(fmt.Sprintf("[SetupConfig] สร้าง ClickHouse database '%s' สำเร็จ", req.Database))
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("สร้าง database '%s' สำเร็จ", req.Database),
	})
}

func testRedisConnection(c echo.Context, host, port string, start time.Time) error {
	if host == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "กรุณาระบุ Redis host",
		})
	}
	if port == "" {
		port = "6379"
	}

	addr := host + ":" + port
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   false,
			"message":   "เชื่อมต่อ Redis ล้มเหลว: " + err.Error(),
			"latencyms": time.Since(start).Milliseconds(),
		})
	}
	defer conn.Close()

	// ส่ง PING command
	_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   false,
			"message":   "ส่งคำสั่ง PING ไปยัง Redis ล้มเหลว: " + err.Error(),
			"latencyms": time.Since(start).Milliseconds(),
		})
	}

	// อ่าน response
	buf := make([]byte, 64)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   false,
			"message":   "อ่าน response จาก Redis ล้มเหลว: " + err.Error(),
			"latencyms": time.Since(start).Milliseconds(),
		})
	}

	response := string(buf[:n])
	if strings.Contains(response, "PONG") {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   true,
			"message":   "เชื่อมต่อ Redis สำเร็จ (PONG)",
			"latencyms": time.Since(start).Milliseconds(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":   false,
		"message":   fmt.Sprintf("Redis response ไม่ถูกต้อง: %s", response),
		"latencyms": time.Since(start).Milliseconds(),
	})
}

func testKafkaConnection(c echo.Context, host, port string, start time.Time) error {
	if host == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "กรุณาระบุ Kafka broker host",
		})
	}
	if port == "" {
		port = "9092"
	}

	addr := host
	if !strings.Contains(host, ":") {
		addr = host + ":" + port
	}

	conn, err := kafka.Dial("tcp", addr)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   false,
			"message":   "เชื่อมต่อ Kafka ล้มเหลว: " + err.Error(),
			"latencyms": time.Since(start).Milliseconds(),
		})
	}
	defer conn.Close()

	// ดึง broker list
	brokers, err := conn.Brokers()
	brokerInfo := ""
	if err == nil {
		brokerInfo = fmt.Sprintf(" (พบ %d brokers)", len(brokers))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":   true,
		"message":   "เชื่อมต่อ Kafka สำเร็จ" + brokerInfo,
		"latencyms": time.Since(start).Milliseconds(),
	})
}

func testHTTPConnection(c echo.Context, url string, start time.Time) error {
	if url == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "กรุณาระบุ URL",
		})
	}

	// เพิ่ม https:// ถ้าไม่มี protocol
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success":   false,
			"message":   "เชื่อมต่อ URL ล้มเหลว: " + err.Error(),
			"latencyms": time.Since(start).Milliseconds(),
		})
	}
	defer resp.Body.Close()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":    true,
		"message":    fmt.Sprintf("เชื่อมต่อ URL สำเร็จ (HTTP %d)", resp.StatusCode),
		"latencyms":  time.Since(start).Milliseconds(),
		"statuscode": resp.StatusCode,
	})
}

// ==========================================
// Seed Default Config (เรียกจาก environment variables)
// POST /api/setup/config/seed
// ==========================================

func SetupSeedConfigHandler(c echo.Context) error {
	var req struct {
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "ข้อมูลไม่ถูกต้อง",
		})
	}

	if !verifySetupPassword(req.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "รหัสผ่านไม่ถูกต้อง",
		})
	}

	// ดึงค่าจาก environment variables ที่ใช้อยู่
	envConfigs := []SetupConfigEntry{
		// Service URLs
		{Category: "serviceurls", Key: "mainapiurl", Value: os.Getenv("MAINAPI_URL"), Description: "Main API URL (mainapi)", IsSecret: false},
		{Category: "serviceurls", Key: "goapiurl", Value: os.Getenv("GOAPI_URL"), Description: "Go API URL (goapi)", IsSecret: false},

		// MongoDB DEV ใช้สำหรับ run บน local เท่านั้น; UAT/PRO จะตั้งค่าตอน deploy บน internet ภายหลัง
		{Category: "mongodbdev", Key: "uri", Value: envFirst("MONGODB_DEV_URI", "MONGODB_URI"), Description: "MongoDB DEV Local Connection URI", IsSecret: true},
		{Category: "mongodbdev", Key: "database", Value: envFirst("MONGODB_DEV_DB", "MONGODB_DEV_DATABASE", "MONGO_DB_NAME", "MONGODB_DB", "MONGODB_DATABASE_NAME"), Description: "MongoDB DEV Local Database Name", IsSecret: false},
		{Category: "mongodb", Key: "uri", Value: os.Getenv("MONGODB_URI"), Description: "Legacy MongoDB URI (explicit compatibility only)", IsSecret: true},
		{Category: "mongodb", Key: "database", Value: envFirst("MONGODB_DB", "MONGO_DB_NAME", "MONGODB_DATABASE_NAME"), Description: "Legacy MongoDB Database Name (explicit compatibility only)", IsSecret: false},

		// PostgreSQL
		{Category: "postgresql", Key: "host", Value: os.Getenv("POSTGRES_HOST"), Description: "PostgreSQL Host", IsSecret: false},
		{Category: "postgresql", Key: "port", Value: os.Getenv("POSTGRES_PORT"), Description: "PostgreSQL Port", IsSecret: false},
		{Category: "postgresql", Key: "user", Value: os.Getenv("POSTGRES_USER"), Description: "PostgreSQL User", IsSecret: false},
		{Category: "postgresql", Key: "password", Value: os.Getenv("POSTGRES_PASSWORD"), Description: "PostgreSQL Password", IsSecret: true},
		{Category: "postgresql", Key: "sslmode", Value: os.Getenv("POSTGRES_SSL_MODE"), Description: "PostgreSQL SSL Mode", IsSecret: false},

		// ClickHouse
		{Category: "clickhouse", Key: "host", Value: os.Getenv("CLICKHOUSE_HOST"), Description: "ClickHouse Host", IsSecret: false},
		{Category: "clickhouse", Key: "port", Value: os.Getenv("CLICKHOUSE_PORT"), Description: "ClickHouse Port", IsSecret: false},
		{Category: "clickhouse", Key: "user", Value: os.Getenv("CLICKHOUSE_USER"), Description: "ClickHouse User", IsSecret: false},
		{Category: "clickhouse", Key: "password", Value: os.Getenv("CLICKHOUSE_PASSWORD"), Description: "ClickHouse Password", IsSecret: true},
		{Category: "clickhouse", Key: "databasename", Value: os.Getenv("CH_DATABASE_NAME"), Description: "ClickHouse Database Name", IsSecret: false},

		// Redis
		{Category: "redis", Key: "host", Value: os.Getenv("REDIS_HOST"), Description: "Redis Host", IsSecret: false},
		{Category: "redis", Key: "port", Value: os.Getenv("REDIS_PORT"), Description: "Redis Port", IsSecret: false},

		// Kafka
		{Category: "kafka", Key: "serverurl", Value: os.Getenv("KAFKA_SERVER_URL"), Description: "Kafka Broker URL", IsSecret: false},

		// Integrations
		{Category: "integrations", Key: "geminiapikey", Value: os.Getenv("GEMINI_API_KEY"), Description: "Google Gemini API Key", IsSecret: true},
		{Category: "integrations", Key: "r2accountid", Value: os.Getenv("R2_ACCOUNT_ID"), Description: "Cloudflare R2 Account ID", IsSecret: false},
		{Category: "integrations", Key: "r2accesskeyid", Value: os.Getenv("R2_ACCESS_KEY_ID"), Description: "Cloudflare R2 Access Key", IsSecret: true},
		{Category: "integrations", Key: "r2secretaccesskey", Value: os.Getenv("R2_SECRET_ACCESS_KEY"), Description: "Cloudflare R2 Secret Key", IsSecret: true},
		{Category: "integrations", Key: "r2bucketname", Value: os.Getenv("R2_BUCKET_NAME"), Description: "Cloudflare R2 Bucket Name", IsSecret: false},
	}

	db := getSetupDB()
	if db == nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "ไม่สามารถเชื่อมต่อ MongoDB ได้",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := db.Collection(setupConfigCollection)

	// สร้าง unique index
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "category", Value: 1}, {Key: "key", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, _ = collection.Indexes().CreateOne(ctx, indexModel)

	seededCount := 0
	for _, cfg := range envConfigs {
		if cfg.Value == "" {
			continue // ข้ามถ้า env var ไม่ได้ตั้งค่า
		}

		cfg.UpdatedAt = time.Now()
		cfg.UpdatedBy = "seed"

		// ใช้ upsert — ถ้ามีค่าอยู่แล้วจะไม่เขียนทับ
		filter := bson.M{
			"category": cfg.Category,
			"key":      cfg.Key,
		}

		// ตรวจสอบว่ามีอยู่แล้วหรือไม่
		count, _ := collection.CountDocuments(ctx, filter)
		if count > 0 {
			continue // ข้ามถ้ามีอยู่แล้ว
		}

		_, err := collection.InsertOne(ctx, cfg)
		if err != nil {
			logger.Warn("[SetupConfig] Seed config [%s/%s] ล้มเหลว: %v", cfg.Category, cfg.Key, err)
			continue
		}
		seededCount++
	}

	logger.Info("[SetupConfig] Seed config สำเร็จ %d รายการ", seededCount)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Seed config จาก environment variables สำเร็จ %d รายการ", seededCount),
	})
}

func envFirst(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}
