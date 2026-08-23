package config

import (
	"os"
	"strings"
)

const (
	DataEnvironmentDev = "dev"
	DataEnvironmentUAT = "uat"
	DataEnvironmentPRO = "pro"
)

func ConfiguredDataEnvironment() (string, bool) {
	for _, key := range []string{"BC_ENV", "APP_ENV", "RUN_ENV", "ENVIRONMENT", "MODE"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return NormalizeDataEnvironment(value), true
		}
	}
	return "", false
}

func CurrentDataEnvironment() string {
	if value, configured := ConfiguredDataEnvironment(); configured {
		return value
	}
	return DataEnvironmentDev
}

func NormalizeDataEnvironment(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "uat", "staging", "stage":
		return DataEnvironmentUAT
	case "pro", "prod", "production":
		return DataEnvironmentPRO
	default:
		return DataEnvironmentDev
	}
}

func MongoURIForCurrentEnvironment() string {
	switch CurrentDataEnvironment() {
	case DataEnvironmentUAT:
		return firstNonEmptyEnv("MONGODB_UAT_URI")
	case DataEnvironmentPRO:
		return firstNonEmptyEnv("MONGODB_PRO_URI", "MONGODB_PRODUCTION_URI")
	default:
		return firstNonEmptyEnv("MONGODB_DEV_URI", "MONGODB_URI")
	}
}

func MongoDatabaseForCurrentEnvironment(defaultDB string) string {
	switch CurrentDataEnvironment() {
	case DataEnvironmentUAT:
		return firstNonEmptyEnv("MONGODB_UAT_DB", "MONGODB_UAT_DATABASE")
	case DataEnvironmentPRO:
		return firstNonEmptyEnv("MONGODB_PRO_DB", "MONGODB_PRO_DATABASE", "MONGODB_PRODUCTION_DB")
	default:
		if dbName := firstNonEmptyEnv("MONGODB_DEV_DB", "MONGODB_DEV_DATABASE", "MONGO_DB_NAME", "MONGODB_DB"); dbName != "" {
			return dbName
		}
		return defaultDB
	}
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}
