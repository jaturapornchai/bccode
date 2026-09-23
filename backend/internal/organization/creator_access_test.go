package organization

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/pkg/apperr"

)

func TestValidateCreatorPolicy(t *testing.T) {
	tests := []struct {
		name         string
		role         authmodels.UserRole
		disabledAt   time.Time
		requireAdmin bool
		wantCode     string
	}{
		{name: "holding bootstrap"},
		{name: "disabled account", disabledAt: time.Now(), wantCode: "FORBIDDEN"},
		{name: "ordinary user", role: authmodels.ROLE_USER, requireAdmin: true, wantCode: "FORBIDDEN"},
		{name: "admin", role: authmodels.ROLE_ADMIN, requireAdmin: true},
		{name: "owner", role: authmodels.ROLE_OWNER, requireAdmin: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreatorPolicy(tt.disabledAt, tt.role, tt.requireAdmin)
			if tt.wantCode == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantCode != "" && (err == nil || err.Code != tt.wantCode) {
				t.Fatalf("error code = %v, want %s", err, tt.wantCode)
			}
		})
	}
}

func TestCreatorFindErrorDistinguishesMissingIdentityFromDatabaseFailure(t *testing.T) {
	missing := creatorFindError(sql.ErrNoRows, apperr.ErrForbidden)
	if missing == nil || missing.Code != apperr.ErrForbidden.Code {
		t.Fatalf("missing identity code = %v, want FORBIDDEN", missing)
	}

	databaseFailure := creatorFindError(errors.New("database unavailable"), apperr.ErrForbidden)
	if databaseFailure == nil || databaseFailure.Code != apperr.ErrInternal.Code {
		t.Fatalf("database failure code = %v, want INTERNAL", databaseFailure)
	}
}
