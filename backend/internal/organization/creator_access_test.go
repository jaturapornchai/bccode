package organization

import (
	"testing"
	"time"

	authmodels "smlcloudplatform/internal/authentication/models"
)

func TestValidateCreatorPolicy(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		role         authmodels.UserRole
		disabledAt   time.Time
		requireAdmin bool
		wantCode     string
	}{
		{name: "holding bootstrap with email", email: "owner@example.com"},
		{name: "missing email", wantCode: "EMAIL_REQUIRED"},
		{name: "invalid email", email: "not-an-email", wantCode: "EMAIL_REQUIRED"},
		{name: "disabled account", email: "disabled@example.com", disabledAt: time.Now(), wantCode: "FORBIDDEN"},
		{name: "ordinary user", email: "user@example.com", role: authmodels.ROLE_USER, requireAdmin: true, wantCode: "FORBIDDEN"},
		{name: "admin", email: "admin@example.com", role: authmodels.ROLE_ADMIN, requireAdmin: true},
		{name: "owner", email: "owner@example.com", role: authmodels.ROLE_OWNER, requireAdmin: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreatorPolicy(tt.email, tt.disabledAt, tt.role, tt.requireAdmin)
			if tt.wantCode == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantCode != "" && (err == nil || err.Code != tt.wantCode) {
				t.Fatalf("error code = %v, want %s", err, tt.wantCode)
			}
		})
	}
}
