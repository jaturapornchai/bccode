package main

import (
	"strings"
	"testing"
)

func TestLoadConfigRequiresAllEnvironmentVariables(t *testing.T) {
	t.Setenv("R2_ACCOUNT_ID", "")
	t.Setenv("R2_ACCESS_KEY_ID", "")
	t.Setenv("R2_SECRET_ACCESS_KEY", "")
	t.Setenv("R2_BUCKET_NAME", "")
	t.Setenv("BC_R2_SMOKE_HOLDING_CODE", "")

	_, err := loadConfig()
	if err == nil {
		t.Fatal("expected missing environment error")
	}
	for _, want := range []string{"R2_ACCOUNT_ID", "R2_ACCESS_KEY_ID", "R2_SECRET_ACCESS_KEY", "R2_BUCKET_NAME", "BC_R2_SMOKE_HOLDING_CODE"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected error to mention %s, got %v", want, err)
		}
	}
}

func TestLoadConfigRejectsNestedHoldingCode(t *testing.T) {
	t.Setenv("R2_ACCOUNT_ID", "account")
	t.Setenv("R2_ACCESS_KEY_ID", "access")
	t.Setenv("R2_SECRET_ACCESS_KEY", "secret")
	t.Setenv("R2_BUCKET_NAME", "bucket")
	t.Setenv("BC_R2_SMOKE_HOLDING_CODE", "SHOP001/other")

	_, err := loadConfig()
	if err == nil || !strings.Contains(err.Error(), "single holdingcode segment") {
		t.Fatalf("expected single holdingcode segment error, got %v", err)
	}
}

func TestLoadConfigOK(t *testing.T) {
	t.Setenv("R2_ACCOUNT_ID", "account")
	t.Setenv("R2_ACCESS_KEY_ID", "access")
	t.Setenv("R2_SECRET_ACCESS_KEY", "secret")
	t.Setenv("R2_BUCKET_NAME", "bucket")
	t.Setenv("BC_R2_SMOKE_HOLDING_CODE", "SHOP001")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if cfg.AccountID != "account" || cfg.AccessKeyID != "access" || cfg.SecretAccessKey != "secret" || cfg.BucketName != "bucket" || cfg.HoldingCode != "SHOP001" {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}
