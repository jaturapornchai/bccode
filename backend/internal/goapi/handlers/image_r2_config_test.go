package handlers

import (
	"strings"
	"testing"
)

func resetObjectStorageClientForTest() {
	r2InitMu.Lock()
	defer r2InitMu.Unlock()
	r2Client = nil
	r2BucketName = ""
}

func clearObjectStorageEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"S3_ACCOUNT_ID",
		"S3_ENDPOINT",
		"S3_REGION",
		"S3_ACCESS_KEY_ID",
		"S3_SECRET_ACCESS_KEY",
		"S3_BUCKET_NAME",
		"S3_FORCE_PATH_STYLE",
		"R2_ACCOUNT_ID",
		"R2_ENDPOINT",
		"R2_ACCESS_KEY_ID",
		"R2_SECRET_ACCESS_KEY",
		"R2_BUCKET_NAME",
		"R2_FORCE_PATH_STYLE",
	} {
		t.Setenv(key, "")
	}
	resetObjectStorageClientForTest()
	t.Cleanup(resetObjectStorageClientForTest)
}

func setMinIOEnv(t *testing.T, bucket string) {
	t.Helper()
	t.Setenv("S3_ENDPOINT", "http://minio:9000")
	t.Setenv("S3_ACCESS_KEY_ID", "minio-app-access")
	t.Setenv("S3_SECRET_ACCESS_KEY", "minio-app-secret")
	t.Setenv("S3_BUCKET_NAME", bucket)
	t.Setenv("S3_FORCE_PATH_STYLE", "true")
}

func TestMinIOExplicitEndpointDoesNotRequireAccountID(t *testing.T) {
	clearObjectStorageEnv(t)
	setMinIOEnv(t, "bcai-account")

	client, err := GetR2Client()
	if err != nil {
		t.Fatalf("GetR2Client returned error: %v", err)
	}
	if client == nil {
		t.Fatal("expected an S3-compatible client")
	}
	if r2BucketName != "bcai-account" {
		t.Fatalf("expected S3 bucket, got %q", r2BucketName)
	}
}

func TestCloudflareWithoutEndpointRequiresAccountID(t *testing.T) {
	clearObjectStorageEnv(t)
	t.Setenv("R2_ACCESS_KEY_ID", "r2-access")
	t.Setenv("R2_SECRET_ACCESS_KEY", "r2-secret")
	t.Setenv("R2_BUCKET_NAME", "r2-bucket")

	_, err := GetR2Client()
	if err == nil || !strings.Contains(err.Error(), "S3_ENDPOINT or R2_ACCOUNT_ID") {
		t.Fatalf("expected endpoint/account validation error, got %v", err)
	}
}

func TestObjectStoragePrefersS3Config(t *testing.T) {
	clearObjectStorageEnv(t)
	setMinIOEnv(t, "s3-bucket")
	t.Setenv("R2_ACCOUNT_ID", "legacy-account")
	t.Setenv("R2_ACCESS_KEY_ID", "legacy-access")
	t.Setenv("R2_SECRET_ACCESS_KEY", "legacy-secret")
	t.Setenv("R2_BUCKET_NAME", "legacy-bucket")

	if _, err := GetR2Client(); err != nil {
		t.Fatalf("GetR2Client returned error: %v", err)
	}
	if r2BucketName != "s3-bucket" {
		t.Fatalf("expected S3 config to win, got bucket %q", r2BucketName)
	}
}

func TestObjectStorageInitRetriesAfterValidationFailure(t *testing.T) {
	clearObjectStorageEnv(t)

	if _, err := GetR2Client(); err == nil {
		t.Fatal("expected initial configuration error")
	}
	setMinIOEnv(t, "retry-bucket")

	client, err := GetR2Client()
	if err != nil || client == nil {
		t.Fatalf("expected retry to initialize client, client=%v err=%v", client, err)
	}
}

func TestSuccessfulObjectStorageClientRemainsCached(t *testing.T) {
	clearObjectStorageEnv(t)
	setMinIOEnv(t, "original-bucket")

	first, err := GetR2Client()
	if err != nil {
		t.Fatalf("first GetR2Client returned error: %v", err)
	}
	t.Setenv("S3_BUCKET_NAME", "changed-bucket")
	second, err := GetR2Client()
	if err != nil {
		t.Fatalf("second GetR2Client returned error: %v", err)
	}
	if first != second || r2BucketName != "original-bucket" {
		t.Fatal("successful client should stay cached until process restart")
	}
}
