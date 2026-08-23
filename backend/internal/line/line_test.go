//go:build integration

package line

import (
	"os"
	"testing"
)

func TestValidateTokenIntegration(t *testing.T) {
	channelID := os.Getenv("LINE_TEST_CHANNEL_ID")
	token := os.Getenv("LINE_TEST_ID_TOKEN")
	if channelID == "" || token == "" {
		t.Skip("LINE_TEST_CHANNEL_ID and LINE_TEST_ID_TOKEN are required for the LINE integration test")
	}

	adapter := NewLineAdapter(channelID)

	user, err := adapter.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate LINE token: %v", err)
	}
	if user == nil {
		t.Fatal("validated LINE user is nil")
	}
	if user.UserId == "" {
		t.Fatal("validated LINE user id is empty")
	}
}
