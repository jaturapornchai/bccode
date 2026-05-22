package datatransfer

import "testing"

func TestSourceDatabaseConfigPrefersExplicitSourceURI(t *testing.T) {
	clearDataTransferEnv(t)
	t.Setenv("MONGODB_SOURCE_URI", "mongodb://source-private")
	t.Setenv("MONGODB_SOURCE_DB", "source_db")
	t.Setenv("MONGODB_UAT_URI", "mongodb://uat-private")
	t.Setenv("MONGODB_UAT_DB", "uat_db")

	cfg := SourceDatabaseConfig{}

	if got := cfg.MongodbURI(); got != "mongodb://source-private" {
		t.Fatalf("MongodbURI() = %q", got)
	}
	if got := cfg.DB(); got != "source_db" {
		t.Fatalf("DB() = %q", got)
	}
}

func TestSourceDatabaseConfigFallsBackToUAT(t *testing.T) {
	clearDataTransferEnv(t)
	t.Setenv("MONGODB_UAT_URI", "mongodb://uat-private")
	t.Setenv("MONGODB_UAT_DB", "uat_db")

	cfg := SourceDatabaseConfig{}

	if got := cfg.MongodbURI(); got != "mongodb://uat-private" {
		t.Fatalf("MongodbURI() = %q", got)
	}
	if got := cfg.DB(); got != "uat_db" {
		t.Fatalf("DB() = %q", got)
	}
}

func TestDestinationDatabaseConfigPrefersExplicitDestinationURI(t *testing.T) {
	clearDataTransferEnv(t)
	t.Setenv("MONGODB_DESTINATION_URI", "mongodb://destination-private")
	t.Setenv("MONGODB_DESTINATION_DB", "destination_db")
	t.Setenv("MONGODB_DEV_URI", "mongodb://dev-private")
	t.Setenv("MONGODB_DEV_DB", "dev_db")

	cfg := DestinationDatabaseConfig{}

	if got := cfg.MongodbURI(); got != "mongodb://destination-private" {
		t.Fatalf("MongodbURI() = %q", got)
	}
	if got := cfg.DB(); got != "destination_db" {
		t.Fatalf("DB() = %q", got)
	}
}

func TestDestinationDatabaseConfigFallsBackToCurrentEnvironment(t *testing.T) {
	clearDataTransferEnv(t)
	t.Setenv("MODE", "dev")
	t.Setenv("MONGODB_DEV_URI", "mongodb://dev-private")
	t.Setenv("MONGODB_DEV_DB", "dev_db")

	cfg := DestinationDatabaseConfig{}

	if got := cfg.MongodbURI(); got != "mongodb://dev-private" {
		t.Fatalf("MongodbURI() = %q", got)
	}
	if got := cfg.DB(); got != "dev_db" {
		t.Fatalf("DB() = %q", got)
	}
}

func clearDataTransferEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"BC_ENV",
		"APP_ENV",
		"RUN_ENV",
		"ENVIRONMENT",
		"MODE",
		"MONGODB_SOURCE_URI",
		"MONGODB_SOURCE_CONNECTION",
		"MONGODB_SOURCE_DB",
		"MONGODB_SOURCE_DATABASE",
		"MONGODB_DESTINATION_URI",
		"MONGODB_DESTINATION_CONNECTION",
		"MONGODB_DESTINATION_DB",
		"MONGODB_DESTINATION_DATABASE",
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
		"MONGODB_URI",
		"MONGODB_DB",
		"MONGO_DB_NAME",
	} {
		t.Setenv(key, "")
	}
}
