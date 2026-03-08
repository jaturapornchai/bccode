package config

import (
	"smlcloudplatform/internal/goapi/logger"
	"os"
)

type IServiceConfig interface {
	MongodbURI() string
	MongodbDatabaseName() string

	// MongoDB Dev Server Configuration
	MongoServerIP() string
	MongoServerPort() string
	MongoUsername() string
	MongoPassword() string
	MongoAuthDB() string
	MongoDBName() string

	PostgresHost() string
	PostgresPort() string
	PostgresUser() string
	PostgresPassword() string
	PostgresDatabase() string
	PostgresSSLMode() string

	KafkaURI() string
	KafKaSecurityProtocol() string
	KafKaSSLKeyFile() string
	KafKaSSLCAFile() string
	KafKaSSLCertFile() string
	KafkaConsumerGroupVersion() string
}

type ServiceConfig struct{}

func LoadEnv() {
	// ไม่โหลด .env อีกต่อไป — config ทั้งหมดมาจาก bootstrap.json + MongoDB system_config
	logger.Info("[Config] config มาจาก bootstrap.json + MongoDB system_config (ไม่ใช้ .env)")
}

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

// ValidateRequiredEnvVars checks that critical environment variables are set.
// Database connection vars (PostgreSQL, ClickHouse, Redis, Kafka) ไม่ต้อง validate ที่นี่
// เพราะจะถูก override จาก Setup Config (MongoDB system_config) ภายหลัง
func ValidateRequiredEnvVars() {
	// Database connection vars มาจาก Setup Config ไม่ต้อง validate ใน .env
	// Bootstrap: MONGODB_URI มาจาก bootstrap.json
	// ที่เหลือ: มาจาก MongoDB system_config collection
}

func (ServiceConfig) MongodbURI() string {
	return getEnv("MONGODB_URI", "") // mongodb://localhost:27017
}

func (ServiceConfig) MongodbDatabaseName() string {
	// ลำดับความสำคัญ: MONGO_DB_NAME (dev) > MONGODB_DB (cloud/production)
	dbName := getEnv("MONGO_DB_NAME", "")
	if dbName != "" {
		return dbName
	}
	return getEnv("MONGODB_DB", "bcaiclouddb")
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
	return getEnv("POSTGRES_DATABASE", "postgres") // database_name
}

func (ServiceConfig) PostgresSSLMode() string {
	return getEnv("POSTGRES_SSL_MODE", "disable") // disable, require, verify-ca, verify-full
}

func (ServiceConfig) KafkaURI() string {
	return getEnv("KAFKA_SERVER_URL", "") // localhost:9094
}

func (ServiceConfig) KafKaSecurityProtocol() string {
	return getEnv("KAFKA_SECURITY_PROTOCOL", "plaintext") // SASL_SSL
}

func (ServiceConfig) KafKaSSLKeyFile() string {
	return getEnv("KAFKA_SSL_KEY_FILE", "") // /path/to/key
}

func (ServiceConfig) KafKaSSLCAFile() string {
	return getEnv("KAFKA_SSL_CA_FILE", "") // /path/to/ca
}

func (ServiceConfig) KafKaSSLCertFile() string {
	return getEnv("KAFKA_SSL_CERT_FILE", "") // /path/to/cert
}

// KafkaConsumerGroupVersion returns the consumer group version suffix
// Change this to replay all messages from beginning with a new consumer group
// Production: v1 (stable), Development: v2 (replay on first start), etc.
func (ServiceConfig) KafkaConsumerGroupVersion() string {
	return getEnv("KAFKA_CONSUMER_GROUP_VERSION", "v1")
}

// MongoDB Dev Server Configuration Methods
func (ServiceConfig) MongoServerIP() string {
	return getEnv("MONGO_SERVER_IP", "")
}

func (ServiceConfig) MongoServerPort() string {
	return getEnv("MONGO_SERVER_PORT", "27017")
}

func (ServiceConfig) MongoUsername() string {
	return getEnv("MONGO_USERNAME", "")
}

func (ServiceConfig) MongoPassword() string {
	return getEnv("MONGO_PASSWORD", "")
}

func (ServiceConfig) MongoAuthDB() string {
	return getEnv("MONGO_AUTH_DB", "admin")
}

func (ServiceConfig) MongoDBName() string {
	// ลำดับความสำคัญ: MONGO_DB_NAME (dev) > MONGODB_DB (cloud/production)
	dbName := getEnv("MONGO_DB_NAME", "")
	if dbName != "" {
		return dbName
	}
	return getEnv("MONGODB_DB", "bcaiclouddb")
}

