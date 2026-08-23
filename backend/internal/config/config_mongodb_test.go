package config_test

import (
	"smlcloudplatform/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMongodbConfigDoesNotBuildLegacyServerURI(t *testing.T) {
	clearMongoEnvironment(t)
	t.Setenv("MODE", "development")

	giveProtocal := "mongodb"
	giveServer := "demo-mongo-server"
	givePort := "00000"
	giveUser := "mongo-user"
	givePassword := "mongo-password"
	giveDBName := "mongodb-db"
	giveSSLMode := "true"
	giveCAFile := "/cert/ca.cert"

	t.Setenv("MONGODB_PROTOCAL", giveProtocal)
	t.Setenv("MONGODB_SERVER", giveServer)
	t.Setenv("MONGODB_PORT", givePort)
	t.Setenv("MONGODB_USERNAME", giveUser)
	t.Setenv("MONGODB_PASSWORD", givePassword)
	t.Setenv("MONGODB_DB", giveDBName)
	t.Setenv("MONGODB_SSL", giveSSLMode)
	t.Setenv("MONGODB_TLS_CA_FILE", giveCAFile)

	mongoConfig := &config.MongoPersisterConfig{}

	assert.Equal(t, giveProtocal, mongoConfig.MongodbProtocal())
	assert.Equal(t, "", mongoConfig.MongodbURI())
}

func TestMongodbConfigUsesDevLegacyURIOnlyAsURIFallback(t *testing.T) {
	clearMongoEnvironment(t)
	t.Setenv("MODE", "development")

	giveDBName := "mongodb-db"
	giveURI := "mongodb+srv://legacy-dev.example.mongodb.net"
	t.Setenv("MONGODB_URI", giveURI)
	t.Setenv("MONGODB_DB", giveDBName)

	mongoConfig := &config.MongoPersisterConfig{}

	assert.Equal(t, giveURI, mongoConfig.MongodbURI())
	assert.Equal(t, giveDBName, mongoConfig.DB())
}

func TestMongodbConfigUsesDevAtlasEnvironment(t *testing.T) {
	clearMongoEnvironment(t)
	t.Setenv("MODE", "development")
	t.Setenv("MONGODB_DEV_URI", "mongodb+srv://dev-user:dev-pass@example-dev.mongodb.net")
	t.Setenv("MONGODB_DEV_DB", "bc_dev")
	t.Setenv("MONGODB_URI", "mongodb+srv://generic-user:generic-pass@example.mongodb.net")
	t.Setenv("MONGODB_DB", "generic_db")

	mongoConfig := &config.MongoPersisterConfig{}

	assert.Equal(t, "dev", config.CurrentDataEnvironment())
	assert.Equal(t, "mongodb+srv://dev-user:dev-pass@example-dev.mongodb.net", mongoConfig.MongodbURI())
	assert.Equal(t, "bc_dev", mongoConfig.DB())
}

func TestMongodbConfigUsesUATAtlasEnvironment(t *testing.T) {
	clearMongoEnvironment(t)
	t.Setenv("MODE", "uat")
	t.Setenv("MONGODB_UAT_URI", "mongodb+srv://uat-user:uat-pass@example-uat.mongodb.net")
	t.Setenv("MONGODB_UAT_DB", "bc_uat")
	t.Setenv("MONGODB_URI", "mongodb+srv://generic-user:generic-pass@example.mongodb.net")
	t.Setenv("MONGODB_DB", "generic_db")

	mongoConfig := &config.MongoPersisterConfig{}

	assert.Equal(t, "uat", config.CurrentDataEnvironment())
	assert.Equal(t, "mongodb+srv://uat-user:uat-pass@example-uat.mongodb.net", mongoConfig.MongodbURI())
	assert.Equal(t, "bc_uat", mongoConfig.DB())
}

func TestMongodbConfigUsesPROAtlasEnvironment(t *testing.T) {
	clearMongoEnvironment(t)
	t.Setenv("MODE", "production")
	t.Setenv("MONGODB_PRODUCTION_URI", "mongodb+srv://pro-user:pro-pass@example-pro.mongodb.net")
	t.Setenv("MONGODB_PRODUCTION_DB", "bc_pro")
	t.Setenv("MONGODB_URI", "mongodb+srv://generic-user:generic-pass@example.mongodb.net")
	t.Setenv("MONGODB_DB", "generic_db")

	mongoConfig := &config.MongoPersisterConfig{}

	assert.Equal(t, "pro", config.CurrentDataEnvironment())
	assert.Equal(t, "mongodb+srv://pro-user:pro-pass@example-pro.mongodb.net", mongoConfig.MongodbURI())
	assert.Equal(t, "bc_pro", mongoConfig.DB())
}

func TestMongodbConfigDoesNotFallbackToGenericInUAT(t *testing.T) {
	clearMongoEnvironment(t)
	t.Setenv("MODE", "uat")
	t.Setenv("MONGODB_URI", "mongodb+srv://generic-user:generic-pass@example.mongodb.net")
	t.Setenv("MONGODB_DB", "generic_db")

	mongoConfig := &config.MongoPersisterConfig{}

	assert.Equal(t, "", mongoConfig.MongodbURI())
	assert.Equal(t, "", mongoConfig.DB())
}

func TestConfiguredDataEnvironmentRequiresExplicitValue(t *testing.T) {
	clearMongoEnvironment(t)

	dataEnvironment, configured := config.ConfiguredDataEnvironment()

	assert.False(t, configured)
	assert.Empty(t, dataEnvironment)
	assert.Equal(t, config.DataEnvironmentDev, config.CurrentDataEnvironment())
}

func TestConfiguredDataEnvironmentNormalizesExplicitValue(t *testing.T) {
	clearMongoEnvironment(t)
	t.Setenv("BC_ENV", " production ")

	dataEnvironment, configured := config.ConfiguredDataEnvironment()

	assert.True(t, configured)
	assert.Equal(t, config.DataEnvironmentPRO, dataEnvironment)
}

func clearMongoEnvironment(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"BC_ENV",
		"APP_ENV",
		"RUN_ENV",
		"ENVIRONMENT",
		"MODE",
		"MONGODB_URI",
		"MONGODB_DB",
		"MONGO_DB_NAME",
		"MONGODB_DEV_URI",
		"MONGODB_DEV_DB",
		"MONGODB_DEV_DATABASE",
		"MONGODB_UAT_URI",
		"MONGODB_UAT_DB",
		"MONGODB_UAT_DATABASE",
		"MONGODB_PRO_URI",
		"MONGODB_PRO_DB",
		"MONGODB_PRO_DATABASE",
		"MONGODB_PRODUCTION_URI",
		"MONGODB_PRODUCTION_DB",
		"MONGODB_PROTOCAL",
		"MONGODB_SERVER",
		"MONGODB_PORT",
		"MONGODB_USERNAME",
		"MONGODB_PASSWORD",
		"MONGODB_SSL",
		"MONGODB_TLS_CA_FILE",
	} {
		t.Setenv(key, "")
	}
}
