package config

import "os"

// IPersisterConfig is interface for persister
type IPersisterConfig interface {
	Host() string
	Port() string
	DB() string
	Username() string
	Password() string
	SSLMode() string
	TimeZone() string
	LoggerLevel() string
}

type PersisterConfig struct{}

func NewPersisterConfig() *PersisterConfig {
	return &PersisterConfig{}
}

func (cfg *PersisterConfig) Host() string {
	return os.Getenv("POSTGRES_HOST")
}

func (cfg *PersisterConfig) Port() string {
	return os.Getenv("POSTGRES_PORT")
}

func (cfg *PersisterConfig) DB() string {
	dbName := os.Getenv("POSTGRES_DB_NAME")
	if dbName == "" {
		dbName = "postgres" // default database สำหรับ test connection ตอน startup
	}
	return dbName
}

func (cfg *PersisterConfig) Username() string {
	return os.Getenv("POSTGRES_USERNAME")
}

func (cfg *PersisterConfig) Password() string {
	return os.Getenv("POSTGRES_PASSWORD")
}

func (cfg *PersisterConfig) SSLMode() string {
	sslMode := os.Getenv("POSTGRES_SSL_MODE")
	if sslMode == "" {
		sslMode = "disable"
	}
	return sslMode
}

func (cfg *PersisterConfig) TimeZone() string {
	return getEnv("POSTGRES_TIMEZONE", "Asia/Bangkok")
}

func (cfg *PersisterConfig) LoggerLevel() string {
	loggerLevel := getEnv("POSTGRES_LOGGER_LEVEL", "")
	return loggerLevel
}
