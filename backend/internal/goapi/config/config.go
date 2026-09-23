package config

import (
	"os"
)

// IServiceConfig - ค่าเชื่อมต่อ PostgreSQL จาก env (ฐานข้อมูลเดียวของระบบ)
type IServiceConfig interface {
	PostgresHost() string
	PostgresPort() string
	PostgresUser() string
	PostgresPassword() string
	PostgresDatabase() string
	PostgresSSLMode() string
}

type ServiceConfig struct{}

func getEnv(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}

func NewServiceConfig() IServiceConfig {
	return &ServiceConfig{}
}

func (ServiceConfig) PostgresHost() string {
	return getEnv("POSTGRES_HOST", "localhost") // localhost
}

func (ServiceConfig) PostgresPort() string {
	return getEnv("POSTGRES_PORT", "5432") // 5432
}

func (ServiceConfig) PostgresUser() string {
	return getEnv("POSTGRES_USER", "postgres") // postgres
}

func (ServiceConfig) PostgresPassword() string {
	return getEnv("POSTGRES_PASSWORD", "") // password
}

func (ServiceConfig) PostgresDatabase() string {
	// Use POSTGRES_DB if available, otherwise fall back to POSTGRES_DATABASE
	dbName := getEnv("POSTGRES_DB", "")
	if dbName != "" {
		return dbName
	}
	return getEnv("POSTGRES_DATABASE", "postgres") // databasename
}

func (ServiceConfig) PostgresSSLMode() string {
	return getEnv("POSTGRES_SSL_MODE", "disable") // disable, require, verify-ca, verify-full
}
