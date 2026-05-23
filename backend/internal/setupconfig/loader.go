package setupconfig

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

// bootstrapConfig โครงสร้าง JSON สำหรับ bootstrap ทุก config section
// Redis และ Kafka ไม่อยู่ใน bootstrap.json — ใช้ค่า default (redis:6379, kafka:9092)
type bootstrapConfig struct {
	MongoDB           map[string]string `json:"mongodb"`
	MongoDBDev        map[string]string `json:"mongodb_dev"`
	MongoDBUAT        map[string]string `json:"mongodb_uat"`
	MongoDBPRO        map[string]string `json:"mongodb_pro"`
	MongoDBProduction map[string]string `json:"mongodb_production"`
	PostgreSQL        map[string]string `json:"postgresql"`
	ClickHouse        map[string]string `json:"clickhouse"`
	Service           map[string]string `json:"service"`
	Integrations      map[string]string `json:"integrations"`
	Storage           map[string]string `json:"storage"`
}

// configMapping กำหนดว่า Setup Config key ไหน map กับ env var อะไรบ้าง
// mainapi ใช้ env var names ต่างจาก goapi บางตัว
var configMapping = map[string]map[string][]string{
	"mongodb": {
		"uri":      {"MONGODB_URI", "MONGODB_DEV_URI", "MONGODB_UAT_URI", "MONGODB_PRO_URI", "MONGODB_PRODUCTION_URI"},
		"database": {"MONGODB_DB", "MONGODB_DEV_DB", "MONGODB_UAT_DB", "MONGODB_PRO_DB", "MONGODB_PRODUCTION_DB"},
		"host":     {"MONGODB_HOST", "MONGODB_DEV_HOST", "MONGODB_UAT_HOST", "MONGODB_PRO_HOST"},
		"port":     {"MONGODB_PORT", "MONGODB_DEV_PORT", "MONGODB_UAT_PORT", "MONGODB_PRO_PORT"},
		"username": {"MONGODB_USERNAME", "MONGODB_USER", "MONGODB_DEV_USER", "MONGODB_UAT_USER", "MONGODB_PRO_USER"},
		"password": {"MONGODB_PASSWORD", "MONGODB_DEV_PASSWORD", "MONGODB_UAT_PASSWORD", "MONGODB_PRO_PASSWORD"},
	},
	"postgresql": {
		"host":         {"POSTGRES_HOST"},
		"port":         {"POSTGRES_PORT"},
		"user":         {"POSTGRES_USERNAME"}, // mainapi ใช้ POSTGRES_USERNAME (ไม่ใช่ POSTGRES_USER)
		"password":     {"POSTGRES_PASSWORD"},
		"ssl_mode":     {"POSTGRES_SSL_MODE"},
		"db_name":      {"POSTGRES_DB_NAME"},
		"timezone":     {"POSTGRES_TIMEZONE"},
		"logger_level": {"POSTGRES_LOGGER_LEVEL"},
	},
	"clickhouse": {
		"host":          {"CH_SERVER_ADDRESS"},
		"port":          {"CLICKHOUSE_PORT"},
		"user":          {"CH_USERNAME"},
		"password":      {"CH_PASSWORD"},
		"database_name": {"CH_DATABASE_NAME"},
	},
	"service": {
		"log_level":           {"LOG_LEVEL"},
		"jwt_secret_key":      {"JWT_SECRET_KEY"},
		"dev_api_mode":        {"DEV_API_MODE"},
		"service_port":        {"SERVICE_PORT"},
		"host_api":            {"HOST_API"},
		"mode":                {"MODE"},
		"http_cors":           {"HTTP_CORS"},
		"firebase_project_id": {"FIREBASE_PROJECT_ID"},
	},
	"integrations": {
		"gemini_api_key": {"GEMINI_API_KEY"},
		"gemini_model":   {"GEMINI_MODEL"},
	},
	"storage": {
		"data_path":            {"STORAGE_DATA_PATH"},
		"data_uri":             {"STORAGE_DATA_URI"},
		"azure_account_name":   {"AZURE_STORAGE_ACCOUNT_NAME"},
		"azure_account_key":    {"AZURE_STORAGE_ACCOUNT_KEY"},
		"azure_container_name": {"AZURE_STORAGE_CONTAINER_NAME"},
		"azure_tenant_id":      {"AZURE_TENANT_ID"},
		"s3_endpoint":          {"S3_ENDPOINT"},
		"s3_public_endpoint":   {"S3_PUBLIC_ENDPOINT"},
		"s3_access_key_id":     {"S3_ACCESS_KEY_ID"},
		"s3_secret_access_key": {"S3_SECRET_ACCESS_KEY"},
		"s3_bucket_name":       {"S3_BUCKET_NAME"},
	},
}

// secretKeys รายชื่อ key ที่ต้อง mask ใน log
var secretKeys = map[string]bool{
	"password":             true,
	"secret_access_key":    true,
	"account_key":          true,
	"azure_account_key":    true,
	"jwt_secret_key":       true,
	"gemini_api_key":       true,
	"s3_secret_access_key": true,
}

// bootstrapPaths ลำดับความสำคัญในการหา bootstrap.json
var bootstrapPaths = []string{
	"/app/bootstrap/bootstrap.json", // Docker volume mount (primary)
	"/app/bootstrap.json",           // Docker WORKDIR
	"bootstrap.json",                // Current directory (local dev)
	"config/bootstrap.json",         // Monorepo root (local dev)
}

func getCustomConfigPath(bootstrapPath string) string {
	if bootstrapPath == "" {
		return ""
	}
	return strings.Replace(bootstrapPath, "bootstrap.json", "custom_config.json", 1)
}

func mergeCustomConfig(base *bootstrapConfig, customData []byte) error {
	var custom bootstrapConfig
	if err := json.Unmarshal(customData, &custom); err != nil {
		return err
	}

	mergeMap := func(baseMap *map[string]string, customMap map[string]string) {
		if customMap == nil {
			return
		}
		if *baseMap == nil {
			*baseMap = make(map[string]string)
		}
		for k, v := range customMap {
			(*baseMap)[k] = v
		}
	}

	mergeMap(&base.MongoDB, custom.MongoDB)
	mergeMap(&base.MongoDBDev, custom.MongoDBDev)
	mergeMap(&base.MongoDBUAT, custom.MongoDBUAT)
	mergeMap(&base.MongoDBPRO, custom.MongoDBPRO)
	mergeMap(&base.MongoDBProduction, custom.MongoDBProduction)
	mergeMap(&base.PostgreSQL, custom.PostgreSQL)
	mergeMap(&base.ClickHouse, custom.ClickHouse)
	mergeMap(&base.Service, custom.Service)
	mergeMap(&base.Integrations, custom.Integrations)
	mergeMap(&base.Storage, custom.Storage)

	return nil
}

// LoadBootstrapConfig อ่าน bootstrap.json แล้ว set env vars ทุก section
// เรียกก่อน config.NewConfig() เพื่อให้ env vars พร้อมใช้งาน
func LoadBootstrapConfig() {
	log.Println("[Bootstrap] กำลังค้นหา bootstrap.json...")

	var bootstrapPath string
	var data []byte
	var err error

	for _, path := range bootstrapPaths {
		data, err = os.ReadFile(path)
		if err == nil {
			bootstrapPath = path
			break
		}
	}

	if bootstrapPath == "" {
		log.Printf("[Bootstrap] ไม่พบ bootstrap.json — ไม่สามารถโหลด config ได้")
		log.Printf("[Bootstrap] ค้นหาแล้วที่: %s", strings.Join(bootstrapPaths, ", "))
		return
	}

	var cfg bootstrapConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("[Bootstrap] อ่าน bootstrap.json ล้มเหลว: %v", err)
		return
	}

	log.Printf("[Bootstrap] อ่าน bootstrap.json จาก %s", bootstrapPath)

	// โหลด custom_config.json ถ้ามี เพื่อ override ค่าจาก bootstrap.json
	customPath := getCustomConfigPath(bootstrapPath)
	if customPath != "" {
		if customData, err := os.ReadFile(customPath); err == nil {
			log.Printf("[Bootstrap] พบ custom_config.json จาก %s - กำลัง merge...", customPath)
			if err := mergeCustomConfig(&cfg, customData); err != nil {
				log.Printf("[Bootstrap] อ่าน/รวม custom_config.json ล้มเหลว: %v", err)
			} else {
				log.Printf("[Bootstrap] รวม custom_config.json สำเร็จ")
			}
		}
	}

	overrideCount := 0

	// โหลดทุก section ที่มีใน bootstrap.json ตาม configMapping
	overrideCount += applyBootstrapSection("mongodb", cfg.MongoDB)
	overrideCount += applyBootstrapSection("mongodb_dev", cfg.MongoDBDev)
	overrideCount += applyBootstrapSection("mongodb_uat", cfg.MongoDBUAT)
	overrideCount += applyBootstrapSection("mongodb_pro", cfg.MongoDBPRO)
	overrideCount += applyBootstrapSection("mongodb_production", cfg.MongoDBProduction)
	overrideCount += applyBootstrapSection("postgresql", cfg.PostgreSQL)
	overrideCount += applyBootstrapSection("clickhouse", cfg.ClickHouse)
	overrideCount += applyBootstrapSection("service", cfg.Service)
	overrideCount += applyBootstrapSection("integrations", cfg.Integrations)
	overrideCount += applyBootstrapSection("storage", cfg.Storage)

	// ตั้งค่า default สำหรับ Redis และ Kafka (ไม่ต้องตั้งใน bootstrap.json)
	setDefaultEnvVars()

	// compose CH_SERVER_ADDRESS = host:port (ClickHouse client ต้องการ host:port)
	composeClickHouseAddress()

	if overrideCount > 0 {
		log.Printf("[Bootstrap] ✅ โหลด config สำเร็จ (%d env vars)", overrideCount)
	} else {
		log.Printf("[Bootstrap] ⚠️ bootstrap.json ว่างหรือไม่มี config ที่ตรงกับ configMapping")
	}
}

// ReloadConfig อ่าน bootstrap.json ใหม่ — เรียกเมื่อได้รับ reload-config จาก goapi
func ReloadConfig() {
	log.Println("[ReloadConfig] กำลัง reload config จาก bootstrap.json...")
	LoadBootstrapConfig()
	log.Println("[ReloadConfig] ✅ reload config สำเร็จ")
}

// applyBootstrapSection อ่าน section จาก bootstrap.json แล้ว set env vars ตาม configMapping
func applyBootstrapSection(category string, values map[string]string) int {
	if values == nil || len(values) == 0 {
		return 0
	}

	categoryMapping, exists := configMapping[category]
	if !exists {
		return 0
	}

	count := 0
	for key, value := range values {
		if value == "" {
			continue
		}

		envVarNames, keyExists := categoryMapping[key]
		if !keyExists {
			continue
		}

		for _, envVar := range envVarNames {
			os.Setenv(envVar, value)
			if isSecretKey(key) {
				log.Printf("[Bootstrap] override %s = ***", envVar)
			} else if strings.Contains(strings.ToLower(envVar), "uri") {
				log.Printf("[Bootstrap] override %s = %s", envVar, maskURI(value))
			} else {
				log.Printf("[Bootstrap] override %s = %s", envVar, value)
			}
			count++
		}
	}

	return count
}

// setDefaultEnvVars ตั้งค่า default สำหรับ Redis และ Kafka
// ค่าเหล่านี้ไม่ต้องอยู่ใน bootstrap.json — ใช้ค่า default ที่ตรงกับ Docker network
func setDefaultEnvVars() {
	defaults := map[string]string{
		"REDIS_HOST":       "redis",
		"REDIS_PORT":       "6379",
		"REDIS_CACHE_URI":  "redis:6379",
		"KAFKA_SERVER_URL": "kafka:9092",
	}
	for envVar, defaultVal := range defaults {
		if os.Getenv(envVar) == "" {
			if envVar == "KAFKA_SERVER_URL" && os.Getenv("ENABLE_KAFKA") == "false" {
				continue
			}
			os.Setenv(envVar, defaultVal)
			log.Printf("[Bootstrap] default %s = %s", envVar, defaultVal)
		}
	}
}

// composeClickHouseAddress compose CH_SERVER_ADDRESS = host:port
// ClickHouse client ต้องการ address ในรูปแบบ host:port (เช่น 103.13.30.32:9000)
// แต่ bootstrap.json แยก host กับ port คนละ field
func composeClickHouseAddress() {
	host := os.Getenv("CH_SERVER_ADDRESS")
	port := os.Getenv("CLICKHOUSE_PORT")
	if host == "" || port == "" {
		return
	}
	// ถ้า host มี port อยู่แล้ว (เช่น host:9000) → ไม่ต้องเพิ่ม
	if strings.Contains(host, ":") {
		return
	}
	composed := host + ":" + port
	os.Setenv("CH_SERVER_ADDRESS", composed)
	log.Printf("[Bootstrap] compose CH_SERVER_ADDRESS = %s", composed)
}

// isSecretKey ตรวจสอบว่า key เป็น secret ที่ต้อง mask ใน log หรือไม่
func isSecretKey(key string) bool {
	if secretKeys[key] {
		return true
	}
	lowerKey := strings.ToLower(key)
	return strings.Contains(lowerKey, "password") ||
		strings.Contains(lowerKey, "secret") ||
		strings.Contains(lowerKey, "api_key")
}

// maskURI ซ่อน password ใน URI สำหรับ log
func maskURI(uri string) string {
	if len(uri) > 30 {
		return uri[:30] + "..."
	}
	return uri
}
