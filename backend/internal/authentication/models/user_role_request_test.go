package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUserRoleRequestAllowsEmptyAccessExpiryDate(t *testing.T) {
	var request UserRoleRequest
	if err := json.Unmarshal([]byte(`{"username":"uat-user","accessexpirydate":""}`), &request); err != nil {
		t.Fatalf("empty access expiry must be accepted: %v", err)
	}
	if !request.AccessExpiryDate.IsZero() {
		t.Fatalf("empty access expiry must become zero time: %v", request.AccessExpiryDate)
	}
}

func TestUserRoleRequestParsesAccessExpiryDate(t *testing.T) {
	var request UserRoleRequest
	if err := json.Unmarshal([]byte(`{"accessexpirydate":"2026-08-31T00:00:00Z"}`), &request); err != nil {
		t.Fatalf("valid access expiry rejected: %v", err)
	}
	want := time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC)
	if !request.AccessExpiryDate.Equal(want) {
		t.Fatalf("access expiry = %v, want %v", request.AccessExpiryDate, want)
	}
}
