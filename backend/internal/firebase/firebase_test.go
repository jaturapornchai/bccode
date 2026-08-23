//go:build integration

package firebase_test

import (
	"os"
	"smlcloudplatform/internal/firebase"
	"testing"
)

func TestVerifyTokenIntegration(t *testing.T) {
	token := os.Getenv("FIREBASE_TEST_ID_TOKEN")
	if token == "" {
		t.Skip("FIREBASE_TEST_ID_TOKEN is required for the Firebase integration test")
	}

	adapter := firebase.NewFirebaseAdapter()
	user, err := adapter.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate Firebase token: %v", err)
	}
	if user == nil {
		t.Fatal("validated Firebase user is nil")
	}
}
