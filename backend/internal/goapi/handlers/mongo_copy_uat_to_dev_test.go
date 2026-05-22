package handlers

import (
	"strings"
	"testing"

	"smlcloudplatform/internal/goapi/models"
)

func TestValidateMongoCopyEnvironmentAllowsUATToDevOnlyInDevMode(t *testing.T) {
	t.Setenv("MODE", "development")

	sourceEnv, targetEnv, err := validateMongoCopyEnvironment(models.PayLoadCopyMongoStruct{
		SourceEnvironment: "uat",
		TargetEnvironment: "dev",
	})
	if err != nil {
		t.Fatalf("validateMongoCopyEnvironment returned error: %v", err)
	}
	if sourceEnv != mongoCopyEnvUAT || targetEnv != mongoCopyEnvDev {
		t.Fatalf("source=%s target=%s", sourceEnv, targetEnv)
	}
}

func TestValidateMongoCopyEnvironmentAllowsPROToDevOnlyInDevMode(t *testing.T) {
	t.Setenv("MODE", "dev")

	sourceEnv, targetEnv, err := validateMongoCopyEnvironment(models.PayLoadCopyMongoStruct{
		SourceEnvironment: "production",
		TargetEnvironment: "development",
	})
	if err != nil {
		t.Fatalf("validateMongoCopyEnvironment returned error: %v", err)
	}
	if sourceEnv != mongoCopyEnvPRO || targetEnv != mongoCopyEnvDev {
		t.Fatalf("source=%s target=%s", sourceEnv, targetEnv)
	}
}

func TestValidateMongoCopyEnvironmentBlocksOutsideDevMode(t *testing.T) {
	t.Setenv("MODE", "production")

	_, _, err := validateMongoCopyEnvironment(models.PayLoadCopyMongoStruct{
		SourceEnvironment: "uat",
		TargetEnvironment: "dev",
	})
	if err == nil || !strings.Contains(err.Error(), "DEV mode") {
		t.Fatalf("err = %v, want DEV mode error", err)
	}
}

func TestValidateMongoCopyEnvironmentBlocksWritesOutsideDev(t *testing.T) {
	t.Setenv("MODE", "development")

	_, _, err := validateMongoCopyEnvironment(models.PayLoadCopyMongoStruct{
		SourceEnvironment: "uat",
		TargetEnvironment: "pro",
	})
	if err == nil || !strings.Contains(err.Error(), "target must be DEV") {
		t.Fatalf("err = %v, want target DEV error", err)
	}
}

func TestBuildDevMongoURIUsesEnvSpecificValuesBeforeLegacy(t *testing.T) {
	clearMongoCopyEnv(t)
	t.Setenv("MONGODB_DEV_URI", "mongodb+srv://dev.example.mongodb.net")
	t.Setenv("MONGODB_DEV_DB", "bc_dev")
	t.Setenv("MONGODB_URI", "mongodb+srv://legacy.example.mongodb.net")
	t.Setenv("MONGODB_DB", "legacy_db")

	uri, dbName, err := buildDevMongoURI()
	if err != nil {
		t.Fatalf("buildDevMongoURI returned error: %v", err)
	}
	if uri != "mongodb+srv://dev.example.mongodb.net" || dbName != "bc_dev" {
		t.Fatalf("uri=%q db=%q", uri, dbName)
	}
}

func TestBuildProductionMongoURIUsesPROAliasBeforeProductionAlias(t *testing.T) {
	clearMongoCopyEnv(t)
	t.Setenv("MONGODB_PRO_URI", "mongodb+srv://pro.example.mongodb.net")
	t.Setenv("MONGODB_PRO_DB", "bc_pro")
	t.Setenv("MONGODB_PRODUCTION_URI", "mongodb+srv://production.example.mongodb.net")
	t.Setenv("MONGODB_PRODUCTION_DB", "bc_production")

	uri, dbName, err := buildProductionMongoURI()
	if err != nil {
		t.Fatalf("buildProductionMongoURI returned error: %v", err)
	}
	if uri != "mongodb+srv://pro.example.mongodb.net" || dbName != "bc_pro" {
		t.Fatalf("uri=%q db=%q", uri, dbName)
	}
}

func TestBuildProductionMongoURIDoesNotUseLegacyHostParts(t *testing.T) {
	clearMongoCopyEnv(t)
	t.Setenv("MONGO_SERVER_IP", "legacy-host")
	t.Setenv("MONGO_SERVER_PORT", "27017")
	t.Setenv("MONGO_USERNAME", "legacy-user")
	t.Setenv("MONGO_PASSWORD", "legacy-pass")
	t.Setenv("MONGO_DB_NAME", "legacy-db")

	_, _, err := buildProductionMongoURI()
	if err == nil || !strings.Contains(err.Error(), "MONGODB_PRO_URI") {
		t.Fatalf("err = %v, want MONGODB_PRO_URI required error", err)
	}
}

func clearMongoCopyEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"MONGODB_DEV_URI",
		"MONGODB_DEV_DB",
		"MONGODB_DEV_DATABASE",
		"MONGODB_URI",
		"MONGODB_DB",
		"MONGO_DB_NAME",
		"MONGODB_PRO_URI",
		"MONGODB_PRO_DB",
		"MONGODB_PRO_DATABASE",
		"MONGODB_PRODUCTION_URI",
		"MONGODB_PRODUCTION_DB",
		"MONGO_SERVER_IP",
		"MONGO_SERVER_PORT",
		"MONGO_USERNAME",
		"MONGO_PASSWORD",
		"MONGO_AUTH_DB",
	} {
		t.Setenv(key, "")
	}
}
