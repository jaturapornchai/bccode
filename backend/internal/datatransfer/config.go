package datatransfer

import (
	"os"
	"strings"

	coreconfig "smlcloudplatform/internal/config"
)

type SourceDatabaseConfig struct{}

func (SourceDatabaseConfig) MongodbURI() string {
	return firstNonEmptyEnv("MONGODB_SOURCE_URI", "MONGODB_SOURCE_CONNECTION", "MONGODB_UAT_URI")
}

func (SourceDatabaseConfig) DB() string {
	return firstNonEmptyEnv("MONGODB_SOURCE_DB", "MONGODB_SOURCE_DATABASE", "MONGODB_UAT_DB", "MONGODB_UAT_DATABASE")
}

func (SourceDatabaseConfig) Debug() bool {
	return false
}

type DestinationDatabaseConfig struct{}

func (DestinationDatabaseConfig) MongodbURI() string {
	if uri := firstNonEmptyEnv("MONGODB_DESTINATION_URI", "MONGODB_DESTINATION_CONNECTION"); uri != "" {
		return uri
	}
	return coreconfig.MongoURIForCurrentEnvironment()
}

func (DestinationDatabaseConfig) DB() string {
	if dbName := firstNonEmptyEnv("MONGODB_DESTINATION_DB", "MONGODB_DESTINATION_DATABASE"); dbName != "" {
		return dbName
	}
	return coreconfig.MongoDatabaseForCurrentEnvironment("")
}

func (DestinationDatabaseConfig) Debug() bool {
	return false
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}
