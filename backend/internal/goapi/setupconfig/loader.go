package setupconfig

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myclickhouse"
	"smlcloudplatform/internal/goapi/mydb"
	"sort"
	"strings"
	"time"
)

// bootstrapConfig โครงสร้าง JSON สำหรับ bootstrap ทุก config section
// Redis และ Kafka ไม่อยู่ใน bootstrap.json — ใช้ค่า default (redis:6379, kafka:9092)
type bootstrapConfig struct {
	MongoDB      map[string]string `json:"mongodb"`
	MongoDBDev   map[string]string `json:"mongodb_dev"`
	MongoDBUAT   map[string]string `json:"mongodb_uat"`
	MongoDBPRO   map[string]string `json:"mongodb_pro"`
	PostgreSQL   map[string]string `json:"postgresql"`
	ClickHouse   map[string]string `json:"clickhouse"`
	Service      map[string]string `json:"service"`
	Integrations map[string]string `json:"integrations"`
	Storage      map[string]string `json:"storage"`
	MongoDBProd  map[string]string `json:"mongodb_production"`
	Kafka        map[string]string `json:"kafka"`
}

// ConfigUpdateEntry โครงสร้างสำหรับอัพเดท config กลับไปที่ bootstrap.json
type ConfigUpdateEntry struct {
	Category string
	Key      string
	Value    string
}

// BootstrapConfigEntry โครงสร้างสำหรับ API response (อ่าน config จาก bootstrap.json)
type BootstrapConfigEntry struct {
	Category    string `json:"category"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	IsSecret    bool   `json:"is_secret"`
	Description string `json:"description"`
}

// configMapping กำหนดว่า Setup Config key ไหน map กับ env var อะไรบ้าง
// บาง key อาจ map กับหลาย env vars (เช่น ClickHouse มี 2 ชุด)
// ครอบคลุมทั้ง goapi และ mainapi env vars
var configMapping = map[string]map[string][]string{
	"mongodb": {
		"uri":      {"MONGODB_URI", "MONGODB_DEV_URI", "MONGODB_UAT_URI", "MONGODB_PRO_URI", "MONGODB_PRODUCTION_URI"},
		"database": {"MONGODB_DB", "MONGO_DB_NAME", "MONGODB_DEV_DB", "MONGODB_UAT_DB", "MONGODB_PRO_DB", "MONGODB_PRODUCTION_DB"},
		"host":     {"MONGODB_HOST", "MONGODB_DEV_HOST", "MONGODB_UAT_HOST", "MONGODB_PRO_HOST"},
		"port":     {"MONGODB_PORT", "MONGODB_DEV_PORT", "MONGODB_UAT_PORT", "MONGODB_PRO_PORT"},
		"username": {"MONGODB_USERNAME", "MONGODB_USER", "MONGODB_DEV_USER", "MONGODB_UAT_USER", "MONGODB_PRO_USER"},
		"password": {"MONGODB_PASSWORD", "MONGODB_DEV_PASSWORD", "MONGODB_UAT_PASSWORD", "MONGODB_PRO_PASSWORD"},
	},
	"postgresql": {
		"host":         {"POSTGRES_HOST"},
		"port":         {"POSTGRES_PORT"},
		"user":         {"POSTGRES_USER", "POSTGRES_USERNAME"}, // goapi ใช้ POSTGRES_USER, mainapi ใช้ POSTGRES_USERNAME
		"password":     {"POSTGRES_PASSWORD"},
		"ssl_mode":     {"POSTGRES_SSL_MODE"},
		"db_name":      {"POSTGRES_DB_NAME"},
		"timezone":     {"POSTGRES_TIMEZONE"},
		"logger_level": {"POSTGRES_LOGGER_LEVEL"},
	},
	"clickhouse": {
		"host":          {"CLICKHOUSE_HOST", "CH_SERVER_ADDRESS"},
		"port":          {"CLICKHOUSE_PORT"},
		"user":          {"CLICKHOUSE_USER", "CH_USERNAME"},
		"password":      {"CLICKHOUSE_PASSWORD", "CH_PASSWORD"},
		"database_name": {"CH_DATABASE_NAME"},
	},
	"service": {
		"enable_kafka":                 {"ENABLE_KAFKA"},
		"kafka_consumer_group_version": {"KAFKA_CONSUMER_GROUP_VERSION"},
		"enable_clone_clickhouse":      {"ENABLE_CLONE_CLICKHOUSE"},
		"log_level":                    {"LOG_LEVEL"},
		"jwt_secret_key":               {"JWT_SECRET_KEY"},
		"dev_api_mode":                 {"DEV_API_MODE"},
		"service_port":                 {"SERVICE_PORT"},
		"host_api":                     {"HOST_API"},
		"mode":                         {"MODE"},
		"http_cors":                    {"HTTP_CORS"},
		"cors_allowed_origins":         {"CORS_ALLOWED_ORIGINS"},
		"firebase_project_id":          {"FIREBASE_PROJECT_ID"},
	},
	"integrations": {
		"ai_provider":          {"AI_PROVIDER"}, // "gemini" | "openrouter" | "groq" | "deepseek"
		"gemini_api_key":       {"GEMINI_API_KEY"},
		"gemini_model":         {"GEMINI_MODEL"},
		"openrouter_api_key":   {"OPENROUTER_API_KEY"},
		"openrouter_model":     {"OPENROUTER_MODEL"},
		"groq_api_key":         {"GROQ_API_KEY"},
		"groq_model":           {"GROQ_MODEL"},
		"deepseek_api_key":     {"DEEPSEEK_API_KEY"},
		"deepseek_model":       {"DEEPSEEK_MODEL"},
		"r2_account_id":        {"R2_ACCOUNT_ID"},
		"r2_access_key_id":     {"R2_ACCESS_KEY_ID"},
		"r2_secret_access_key": {"R2_SECRET_ACCESS_KEY"},
		"r2_bucket_name":       {"R2_BUCKET_NAME"},
		"s3_endpoint":          {"S3_ENDPOINT"},
		"s3_public_endpoint":   {"S3_PUBLIC_ENDPOINT"},
		"s3_access_key_id":     {"S3_ACCESS_KEY_ID"},
		"s3_secret_access_key": {"S3_SECRET_ACCESS_KEY"},
		"s3_bucket_name":       {"S3_BUCKET_NAME"},
		"thunder_api_key":      {"THUNDER_API_KEY"},
		"brevo_api_key":        {"BREVO_API_KEY"},
		"brevo_from_email":     {"BREVO_FROM_EMAIL"},
		"brevo_from_name":      {"BREVO_FROM_NAME"},
		"bcweaviate_url":       {"BCWEAVIATE_URL"},
	},
	"kafka": {
		"server_url": {"KAFKA_SERVER_URL"},
	},
	"storage": {
		"data_path":            {"STORAGE_DATA_PATH"},
		"data_uri":             {"STORAGE_DATA_URI"},
		"azure_account_name":   {"AZURE_STORAGE_ACCOUNT_NAME"},
		"azure_account_key":    {"AZURE_STORAGE_ACCOUNT_KEY"},
		"azure_container_name": {"AZURE_STORAGE_CONTAINER_NAME"},
		"azure_tenant_id":      {"AZURE_TENANT_ID"},
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
	"thunder_api_key":      true,
	"brevo_api_key":        true,
	"r2_secret_access_key": true,
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
	mergeMap(&base.PostgreSQL, custom.PostgreSQL)
	mergeMap(&base.ClickHouse, custom.ClickHouse)
	mergeMap(&base.Service, custom.Service)
	mergeMap(&base.Integrations, custom.Integrations)
	mergeMap(&base.Storage, custom.Storage)
	mergeMap(&base.MongoDBProd, custom.MongoDBProd)
	mergeMap(&base.Kafka, custom.Kafka)

	return nil
}

// LoadBootstrapConfig อ่าน bootstrap.json แล้ว set env vars ทุก section
// เรียกก่อน MongoDB connect เพื่อให้ได้ MongoDB URI จาก file
// และก่อน init database connections ทั้งหมด
func LoadBootstrapConfig() {
	logger.Info("[Bootstrap] กำลังค้นหา bootstrap.json...")

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
		logger.Error("[Bootstrap] ไม่พบ bootstrap.json — ไม่สามารถโหลด config ได้")
		logger.Error("[Bootstrap] ค้นหาแล้วที่: %s", strings.Join(bootstrapPaths, ", "))
		return
	}

	var cfg bootstrapConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		logger.Error("[Bootstrap] อ่าน bootstrap.json ล้มเหลว: %v", err)
		return
	}

	logger.Info("[Bootstrap] อ่าน bootstrap.json จาก %s", bootstrapPath)

	// โหลด custom_config.json ถ้ามี เพื่อ override ค่าจาก bootstrap.json
	customPath := getCustomConfigPath(bootstrapPath)
	if customPath != "" {
		if customData, err := os.ReadFile(customPath); err == nil {
			logger.Info("[Bootstrap] พบ custom_config.json จาก %s - กำลัง merge...", customPath)
			if err := mergeCustomConfig(&cfg, customData); err != nil {
				logger.Error("[Bootstrap] อ่าน/รวม custom_config.json ล้มเหลว: %v", err)
			} else {
				logger.Success("[Bootstrap] รวม custom_config.json สำเร็จ")
			}
		}
	}

	overrideCount := 0

	// โหลดทุก section ที่มีใน bootstrap.json ตาม configMapping
	overrideCount += applyBootstrapSection("mongodb", cfg.MongoDB)
	overrideCount += applyBootstrapSection("mongodb_dev", cfg.MongoDBDev)
	overrideCount += applyBootstrapSection("mongodb_uat", cfg.MongoDBUAT)
	overrideCount += applyBootstrapSection("mongodb_pro", cfg.MongoDBPRO)
	overrideCount += applyBootstrapSection("postgresql", cfg.PostgreSQL)
	overrideCount += applyBootstrapSection("clickhouse", cfg.ClickHouse)
	overrideCount += applyBootstrapSection("service", cfg.Service)
	overrideCount += applyBootstrapSection("integrations", cfg.Integrations)
	overrideCount += applyBootstrapSection("storage", cfg.Storage)
	overrideCount += applyBootstrapSection("mongodb_production", cfg.MongoDBProd)
	overrideCount += applyBootstrapSection("kafka", cfg.Kafka)

	// ตั้งค่า default สำหรับ Redis และ Kafka (ถ้ายังไม่ได้ตั้งค่า)
	setDefaultEnvVars()

	// compose CH_SERVER_ADDRESS = host:port (ClickHouse client ต้องการ host:port)
	composeClickHouseAddress()

	if overrideCount > 0 {
		logger.Success("[Bootstrap] ✅ โหลด config สำเร็จ (%d env vars)", overrideCount)
	} else {
		logger.Warn("[Bootstrap] bootstrap.json ว่างหรือไม่มี config ที่ตรงกับ configMapping")
	}
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
				logger.Info("[Bootstrap] override %s = ***", envVar)
			} else if strings.Contains(strings.ToLower(envVar), "uri") {
				logger.Info("[Bootstrap] override %s = %s", envVar, maskURI(value))
			} else {
				logger.Info("[Bootstrap] override %s = %s", envVar, value)
			}
			count++
		}
	}

	return count
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
	os.Setenv("CLICKHOUSE_HOST", composed)
	logger.Info("[Bootstrap] compose CH_SERVER_ADDRESS = %s", composed)
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
			logger.Info("[Bootstrap] default %s = %s", envVar, defaultVal)
		}
	}
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

// ReloadAndReconnect อ่าน bootstrap.json ใหม่แล้ว reconnect database connections
// เรียกหลังจาก save config จาก Frontend เพื่อให้ backend ใช้ config ใหม่ทันที
func ReloadAndReconnect() error {
	logger.Info("[ReloadConfig] กำลัง reload config จาก bootstrap.json และ reconnect databases...")

	// 1. อ่าน bootstrap.json ใหม่ → override env vars
	LoadBootstrapConfig()

	// 2. ปิด PostgreSQL connection pools ทั้งหมด (จะสร้างใหม่อัตโนมัติเมื่อมี request)
	if err := mydb.CloseAllManagers(); err != nil {
		logger.Error("[ReloadConfig] ปิด PostgreSQL pools ล้มเหลว: %v", err)
	} else {
		logger.Info("[ReloadConfig] ปิด PostgreSQL pools แล้ว — จะ reconnect อัตโนมัติ")
	}

	// 3. ปิด ClickHouse connection (จะ reconnect อัตโนมัติเมื่อมี request)
	if err := myclickhouse.CloseClickHouseConnection(); err != nil {
		logger.Error("[ReloadConfig] ปิด ClickHouse connection ล้มเหลว: %v", err)
	} else {
		logger.Info("[ReloadConfig] ปิด ClickHouse connection แล้ว — จะ reconnect อัตโนมัติ")
	}

	// 4. แจ้ง mainapi ให้ reload ด้วย
	notifyMainAPIReload()

	logger.Success("[ReloadConfig] ✅ reload config + reconnect สำเร็จ — connections จะสร้างใหม่เมื่อมี request")
	return nil
}

// notifyMainAPIReload แจ้ง mainapi ให้ reload config ใหม่จาก bootstrap.json
func notifyMainAPIReload() {
	mainAPIURL := os.Getenv("MAINAPI_INTERNAL_URL")
	if mainAPIURL == "" {
		mainAPIURL = "http://mainapi:8080" // Docker network default
	}

	url := mainAPIURL + "/reload-config"
	logger.Info("[ReloadConfig] กำลังแจ้ง mainapi ให้ reload config: %s", url)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		logger.Warn("[ReloadConfig] สร้าง request ไป mainapi ล้มเหลว: %v", err)
		return
	}

	// ส่ง shared secret สำหรับ authentication
	secret := os.Getenv("RELOAD_CONFIG_SECRET")
	if secret != "" {
		req.Header.Set("X-Reload-Secret", secret)
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Warn("[ReloadConfig] แจ้ง mainapi reload ล้มเหลว: %v (mainapi อาจยังไม่พร้อม)", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		logger.Success("[ReloadConfig] ✅ แจ้ง mainapi reload สำเร็จ")
	} else {
		logger.Warn("[ReloadConfig] mainapi ตอบกลับ status %d", resp.StatusCode)
	}
}

// maskURI ซ่อน password ใน URI สำหรับ log
func maskURI(uri string) string {
	if len(uri) > 30 {
		return uri[:30] + "..."
	}
	return uri
}

// ReadBootstrapAsConfigEntries อ่าน bootstrap.json แล้วคืนเป็น []BootstrapConfigEntry
// สำหรับ API response ที่ Flutter frontend ใช้แสดงหน้า Setup Config
func ReadBootstrapAsConfigEntries() ([]BootstrapConfigEntry, error) {
	// หา bootstrap.json ที่ใช้งานอยู่
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

	if data == nil {
		return nil, fmt.Errorf("ไม่พบ bootstrap.json — ค้นหาแล้วที่: %s", strings.Join(bootstrapPaths, ", "))
	}

	// อ่านเป็น generic map เพื่อรองรับทุก section
	var raw map[string]map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("อ่าน bootstrap.json ล้มเหลว: %v", err)
	}

	// อ่าน custom_config.json ถ้ามี เพื่อทำการ merge
	customPath := getCustomConfigPath(bootstrapPath)
	if customPath != "" {
		if customData, err := os.ReadFile(customPath); err == nil {
			var customRaw map[string]map[string]interface{}
			if err := json.Unmarshal(customData, &customRaw); err == nil {
				// merge customRaw ลงใน raw
				for cat, values := range customRaw {
					if raw[cat] == nil {
						raw[cat] = make(map[string]interface{})
					}
					for k, v := range values {
						raw[cat][k] = v
					}
				}
			}
		}
	}

	var entries []BootstrapConfigEntry
	for category, values := range raw {
		for key, val := range values {
			strVal := fmt.Sprintf("%v", val)
			entries = append(entries, BootstrapConfigEntry{
				Category: category,
				Key:      key,
				Value:    strVal,
				IsSecret: isSecretKey(key),
			})
		}
	}

	// เรียงตาม category → key
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Category != entries[j].Category {
			return entries[i].Category < entries[j].Category
		}
		return entries[i].Key < entries[j].Key
	})

	return entries, nil
}

// UpdateBootstrapJSON เขียนค่า config กลับไปที่ custom_config.json
// เรียกหลังจาก save config จาก Frontend เพื่อให้มีค่าล่าสุด
// เพื่อหลีกเลี่ยงสิทธิ์ในการแก้ไข bootstrap.json โดยตรง
func UpdateBootstrapJSON(configs []ConfigUpdateEntry) error {
	logger.Info("[Bootstrap] กำลังเขียน config กลับไปที่ custom_config.json...")

	// หา bootstrap.json ที่ใช้งานอยู่เพื่อหาโฟลเดอร์สำหรับ custom_config.json
	var bootstrapPath string
	var err error

	for _, path := range bootstrapPaths {
		if _, err := os.Stat(path); err == nil {
			bootstrapPath = path
			break
		}
	}

	if bootstrapPath == "" {
		return fmt.Errorf("ไม่พบ bootstrap.json — ค้นหาแล้วที่: %s", strings.Join(bootstrapPaths, ", "))
	}

	customPath := getCustomConfigPath(bootstrapPath)
	if customPath == "" {
		return fmt.Errorf("ไม่สามารถสร้าง custom_config.json path ได้")
	}

	// อ่านค่าปัจจุบันจาก custom_config.json ถ้ามีไฟล์อยู่แล้ว
	// หากไม่มี ให้ใช้ค่าจาก bootstrap.json เป็นฐานเริ่มต้น
	var current map[string]map[string]interface{}
	customData, err := os.ReadFile(customPath)
	if err == nil {
		// custom_config.json มีอยู่แล้ว -> อ่านขึ้นมาอัพเดทต่อ
		if err := json.Unmarshal(customData, &current); err != nil {
			return fmt.Errorf("อ่าน custom_config.json ล้มเหลว: %v", err)
		}
	} else {
		// ยังไม่มี custom_config.json -> อ่านจาก bootstrap.json เป็นค่าตั้งต้น
		baseData, err := os.ReadFile(bootstrapPath)
		if err != nil {
			return fmt.Errorf("อ่าน bootstrap.json ล้มเหลว: %v", err)
		}
		if err := json.Unmarshal(baseData, &current); err != nil {
			return fmt.Errorf("ถอดรหัส bootstrap.json ล้มเหลว: %v", err)
		}
	}

	// อัพเดทค่าจาก configs ที่ส่งมา
	updatedCount := 0
	for _, cfg := range configs {
		if cfg.Category == "" || cfg.Key == "" || cfg.Value == "" || cfg.Value == "***" {
			continue
		}

		// สร้าง section ถ้ายังไม่มี
		if current[cfg.Category] == nil {
			current[cfg.Category] = make(map[string]interface{})
		}

		oldValue, exists := current[cfg.Category][cfg.Key]
		current[cfg.Category][cfg.Key] = cfg.Value

		if !exists || oldValue != cfg.Value {
			if isSecretKey(cfg.Key) {
				logger.Info("[Bootstrap] เขียน %s/%s = ***", cfg.Category, cfg.Key)
			} else {
				logger.Info("[Bootstrap] เขียน %s/%s = %s", cfg.Category, cfg.Key, cfg.Value)
			}
			updatedCount++
		}
	}

	if updatedCount == 0 {
		logger.Info("[Bootstrap] ไม่มีค่าที่เปลี่ยนแปลง — ไม่เขียน custom_config.json")
		return nil
	}

	// เขียนกลับด้วย JSON format สวยๆ
	newData, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return fmt.Errorf("สร้าง JSON ล้มเหลว: %v", err)
	}

	// เขียนไฟล์ (append newline ท้ายไฟล์)
	newData = append(newData, '\n')
	if err := os.WriteFile(customPath, newData, 0644); err != nil {
		return fmt.Errorf("เขียน custom_config.json ล้มเหลว: %v", err)
	}

	logger.Success("[Bootstrap] ✅ เขียน custom_config.json สำเร็จ (%d ค่า) → %s", updatedCount, customPath)
	return nil
}

// GetConfigMapping คืน configMapping สำหรับใช้ใน handler อื่นๆ
func GetConfigMapping() map[string]map[string][]string {
	return configMapping
}

// LogCurrentConfig แสดง env vars ปัจจุบันที่เกี่ยวข้องกับ config สำหรับ debug
func LogCurrentConfig() {
	logger.Info("[Config] ===== สรุป config ปัจจุบัน =====")
	for category, keys := range configMapping {
		for key, envVars := range keys {
			for _, envVar := range envVars {
				value := os.Getenv(envVar)
				if value == "" {
					continue
				}
				if isSecretKey(key) {
					logger.Info("[Config] %s/%s → %s = ***", category, key, envVar)
				} else {
					logger.Info("[Config] %s/%s → %s = %s", category, key, envVar, value)
				}
			}
		}
	}
	logger.Info("[Config] ===== จบสรุป config =====")
}

// GetBootstrapPaths คืน bootstrap file paths สำหรับ test
func GetBootstrapPaths() []string {
	return bootstrapPaths
}

// SetBootstrapPaths ตั้ง bootstrap file paths สำหรับ test
func SetBootstrapPaths(paths []string) {
	bootstrapPaths = paths
}

// init log ว่าโหลด package setupconfig แล้ว
func init() {
	fmt.Println("[SetupConfig] โหลด setupconfig package สำเร็จ")
}
