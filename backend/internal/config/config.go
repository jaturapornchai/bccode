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
	HttpCORS() []string

	// SignKeyPath() string
	// VerifyKeyPath() string
	JwtSecretKey() string
	HttpConfig() IHttpConfig
	LoggerConfig() ILoggerConfig
	GoogleClientId() string
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
	// ไม่โหลด .env อีกต่อไป — config ทั้งหมดมาจาก bootstrap.json + custom_config.json
	log.Println("[Config] config มาจาก bootstrap.json + custom_config.json (ไม่ใช้ .env)")
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
	return getEnv("JWT_SECRET_KEY", "")
}

func (*Config) HttpConfig() IHttpConfig {
	return NewHttpConfig()
}

func (*Config) LoggerConfig() ILoggerConfig {
	return NewLoggerConfig()
}

// GoogleClientId is the public Google OAuth client ID used to verify the audience (aud)
// of Google ID tokens at /googlelogin. A client ID is public (it ships in the web bundle),
// not a secret. Override via bootstrap.json -> GOOGLE_CLIENT_ID when needed.
func (*Config) GoogleClientId() string {
	return getEnv("GOOGLE_CLIENT_ID", "212036599086-c7aqvm005jiv2kqi4duju8spd9b3jb94.apps.googleusercontent.com")
}
