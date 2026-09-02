package microservice

import "testing"

func clearPersisterStorageEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"S3_ACCOUNT_ID",
		"S3_ENDPOINT",
		"S3_REGION",
		"S3_ACCESS_KEY_ID",
		"S3_SECRET_ACCESS_KEY",
		"S3_BUCKET_NAME",
		"R2_ACCOUNT_ID",
		"R2_ENDPOINT",
		"R2_ACCESS_KEY_ID",
		"R2_SECRET_ACCESS_KEY",
		"R2_BUCKET_NAME",
	} {
		t.Setenv(key, "")
	}
}

func setPersisterMinIOEnv(t *testing.T) {
	t.Helper()
	t.Setenv("S3_ENDPOINT", "http://minio:9000")
	t.Setenv("S3_REGION", "us-east-1")
	t.Setenv("S3_ACCESS_KEY_ID", "minio-app-access")
	t.Setenv("S3_SECRET_ACCESS_KEY", "minio-app-secret")
	t.Setenv("S3_BUCKET_NAME", "bcai-account")
}

func TestPersisterMinIOEndpointDoesNotRequireAccountID(t *testing.T) {
	clearPersisterStorageEnv(t)
	setPersisterMinIOEnv(t)

	persister := NewPersisterR2()
	if err := persister.init(); err != nil {
		t.Fatalf("init returned error: %v", err)
	}
	if persister.client == nil || persister.endpoint != "http://minio:9000" || persister.bucketName != "bcai-account" {
		t.Fatalf("unexpected MinIO config: endpoint=%q bucket=%q clientNil=%v", persister.endpoint, persister.bucketName, persister.client == nil)
	}
}

func TestPersisterRetriesAfterMissingConfig(t *testing.T) {
	clearPersisterStorageEnv(t)
	persister := NewPersisterR2()
	if err := persister.init(); err == nil {
		t.Fatal("expected missing storage config error")
	}

	setPersisterMinIOEnv(t)
	if err := persister.init(); err != nil {
		t.Fatalf("retry init returned error: %v", err)
	}
}

func TestPersisterRecognizesOnlyItsPrivateObjectURLs(t *testing.T) {
	clearPersisterStorageEnv(t)
	setPersisterMinIOEnv(t)
	persister := NewPersisterR2()
	if err := persister.init(); err != nil {
		t.Fatalf("init returned error: %v", err)
	}

	key, ok := persister.objectKeyFromURI("http://minio:9000/bcai-account/bc001/images/logo.png")
	if !ok || key != "bc001/images/logo.png" {
		t.Fatalf("expected private object key, key=%q ok=%v", key, ok)
	}
	for _, candidate := range []string{
		"http://other:9000/bcai-account/bc001/images/logo.png",
		"http://minio:9000/other-bucket/bc001/images/logo.png",
		"http://minio:9000/bcai-account/",
	} {
		if _, ok := persister.objectKeyFromURI(candidate); ok {
			t.Fatalf("unexpected match for %q", candidate)
		}
	}
}
