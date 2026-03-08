package config

import (
	"log"
	"os"
	"strings"
)

type IConfig interface {
	ConfigMode() string
	ApplicationName() string
	IsDebugMode() bool
	PathPrefix() string
	PersisterConfig() IPersisterConfig
	MongoPersisterConfig() IPersisterMongoConfig
	ClickHouseConfig() IPersisterClickHouseConfig
	ElkPersisterConfig() IPersisterElkConfig
	OpenSearchPersisterConfig() IPersisterOpenSearchConfig
	CacherConfig() ICacherConfig
	MQConfig() IMQConfig
	TopicName() string
	HttpCORS() []string

	// SignKeyPath() string
	// VerifyKeyPath() string
	JwtSecretKey() string
	HttpConfig() IHttpConfig
	LoggerConfig() ILoggerConfig
	ProductGroupServiceConfig() IProductGroupServiceConfig
	LineClientId() string
}

func GetEnv(key string, fallback string) string {
	return getEnv(key, fallback)
}

func getEnv(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}

type Config struct {
	Mode string
}

func NewConfig() IConfig {
	config := &Config{}
	config.LoadConfig()
	return config
}

func (cfg *Config) LoadConfig() {
	cfg.Mode = cfg.ConfigMode()
	// ไม่โหลด .env อีกต่อไป — config ทั้งหมดมาจาก bootstrap.json + MongoDB system_config
	log.Println("[Config] config มาจาก bootstrap.json + MongoDB system_config (ไม่ใช้ .env)")
}

func (c *Config) ConfigMode() string {
	env := os.Getenv("MODE")
	if env == "" {
		os.Setenv("MODE", "development")
		env = "development"
	}
	return env
}

func (*Config) ApplicationName() string {
	return getEnv("SERVICE_NAME", "microservice")
}

func (c *Config) IsDebugMode() bool {
	if c.ConfigMode() == "development" {
		return true
	}
	return false
}

func (cfg *Config) PathPrefix() string {
	return getEnv("PATH_PREFIX", "")
}

func (*Config) PersisterConfig() IPersisterConfig {
	return NewPersisterConfig()
}

func (cfg *Config) MongoPersisterConfig() IPersisterMongoConfig {
	return NewMongoPersisterConfig()
}

func (cfg *Config) ClickHouseConfig() IPersisterClickHouseConfig {
	return NewPersisterClickHouseConfig()
}

func (*Config) TopicName() string {
	return os.Getenv("TOPIC_NAME")
}

func (*Config) HttpCORS() []string {
	rawCORS := getEnv("HTTP_CORS", "*")

	return strings.Split(rawCORS, " ")
}

// func (*Config) SignKeyPath() string {
// 	return getEnv("PUBLIC_KEY_PATH", "./../../private.key")
// }

// func (*Config) VerifyKeyPath() string {
// 	return getEnv("PRIVATE_KEY_PATH", "./../../public.key")
// }

func (*Config) JwtSecretKey() string {
	return getEnv("JWT_SECRET_KEY", "54cfcbf5437a029d48a9f67552eeb04b48a65703")
}

func (cfg *Config) ElkPersisterConfig() IPersisterElkConfig {
	return NewPersisterElkConfig()
}

func (cfg *Config) OpenSearchPersisterConfig() IPersisterOpenSearchConfig {
	return NewPersisterOpenSearchConfig()
}

///

func (cfg *Config) CacherConfig() ICacherConfig {
	return NewCacherConfig()
}

func (*Config) HttpConfig() IHttpConfig {
	return NewHttpConfig()
}

func (*Config) LoggerConfig() ILoggerConfig {
	return NewLoggerConfig()
}

func (*Config) ProductGroupServiceConfig() IProductGroupServiceConfig {
	return NewProductGroupServiceConfig()
}

func (*Config) LineClientId() string {
	return getEnv("LINE_CLIENT_ID", "1657004770")
}
