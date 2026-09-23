package setupconfig

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

// bootstrapConfig โครงสร้าง JSON สำหรับ bootstrap ทุก config section
type bootstrapConfig struct {
	PostgreSQL   map[string]string `json:"postgresql"`
	Service      map[string]string `json:"service"`
	Integrations map[string]string `json:"integrations"`
	Storage      map[string]string `json:"storage"`
}

// configMapping กำหนดว่า Setup Config key ไหน map กับ env var อะไรบ้าง
// mainapi ใช้ env var names ต่างจาก goapi บางตัว
var configMapping = map[string]map[string][]string{
	"postgresql": {
		"host":        {"POSTGRES_HOST"},
		"port":        {"POSTGRES_PORT"},
		"user":        {"POSTGRES_USERNAME"}, // mainapi ใช้ POSTGRES_USERNAME (ไม่ใช่ POSTGRES_USER)
		"password":    {"POSTGRES_PASSWORD"},
		"sslmode":     {"POSTGRES_SSL_MODE"},
		"dbname":      {"POSTGRES_DB_NAME"},
		"timezone":    {"POSTGRES_TIMEZONE"},
		"loggerlevel": {"POSTGRES_LOGGER_LEVEL"},
	},
	"service": {
		"loglevel":          {"LOG_LEVEL"},
		"jwtsecretkey":      {"JWT_SECRET_KEY"},
		"devapimode":        {"DEV_API_MODE"},
		"serviceport":       {"SERVICE_PORT"},
		"hostapi":           {"HOST_API"},
		"mode":              {"MODE"},
		"httpcors":          {"HTTP_CORS"},
		"firebaseprojectid": {"FIREBASE_PROJECT_ID"},
	},
	"integrations": {
		"geminiapikey": {"GEMINI_API_KEY"},
		"geminimodel":  {"GEMINI_MODEL"},
	},
	"storage": {
		"datapath":           {"STORAGE_DATA_PATH"},
		"datauri":            {"STORAGE_DATA_URI"},
		"azureaccountname":   {"AZURE_STORAGE_ACCOUNT_NAME"},
		"azureaccountkey":    {"AZURE_STORAGE_ACCOUNT_KEY"},
		"azurecontainername": {"AZURE_STORAGE_CONTAINER_NAME"},
		"azuretenantid":      {"AZURE_TENANT_ID"},
		"s3endpoint":         {"S3_ENDPOINT"},
		"s3publicendpoint":   {"S3_PUBLIC_ENDPOINT"},
		"s3accesskeyid":      {"S3_ACCESS_KEY_ID"},
		"s3secretaccesskey":  {"S3_SECRET_ACCESS_KEY"},
		"s3bucketname":       {"S3_BUCKET_NAME"},
	},
}

// secretKeys รายชื่อ key ที่ต้อง mask ใน log
var secretKeys = map[string]bool{
	"password":          true,
	"secretaccesskey":   true,
	"accountkey":        true,
	"azureaccountkey":   true,
	"jwtsecretkey":      true,
	"geminiapikey":      true,
	"s3secretaccesskey": true,
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

	mergeMap(&base.PostgreSQL, custom.PostgreSQL)
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
	overrideCount += applyBootstrapSection("postgresql", cfg.PostgreSQL)
	overrideCount += applyBootstrapSection("service", cfg.Service)
	overrideCount += applyBootstrapSection("integrations", cfg.Integrations)
	overrideCount += applyBootstrapSection("storage", cfg.Storage)

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

// isSecretKey ตรวจสอบว่า key เป็น secret ที่ต้อง mask ใน log หรือไม่
func isSecretKey(key string) bool {
	if secretKeys[key] {
		return true
	}
	lowerKey := strings.ToLower(key)
	return strings.Contains(lowerKey, "password") ||
		strings.Contains(lowerKey, "secret") ||
		strings.Contains(lowerKey, "apikey")
}

// maskURI ซ่อน password ใน URI สำหรับ log
func maskURI(uri string) string {
	if len(uri) > 30 {
		return uri[:30] + "..."
	}
	return uri
}
