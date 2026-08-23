package organization

import (
	"errors"
	"testing"
	"time"

	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/pkg/apperr"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

func TestActiveGoogleIdentityRejectsRevokedOrIncompleteRecords(t *testing.T) {
	now := time.Now().UTC()
	valid := authmodels.GoogleIdentity{
		ID: primitive.NewObjectID(), Issuer: "https://accounts.google.com", Subject: "subject", IsActive: true,
	}
	if !isActiveGoogleIdentity(valid) {
		t.Fatal("complete active identity must be accepted")
	}

	for name, identity := range map[string]authmodels.GoogleIdentity{
		"revoked":  func() authmodels.GoogleIdentity { value := valid; value.RevokedAt = &now; return value }(),
		"inactive": func() authmodels.GoogleIdentity { value := valid; value.IsActive = false; return value }(),
		"no issuer": func() authmodels.GoogleIdentity {
			value := valid
			value.Issuer = ""
			return value
		}(),
		"no subject": func() authmodels.GoogleIdentity {
			value := valid
			value.Subject = ""
			return value
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			if isActiveGoogleIdentity(identity) {
				t.Fatal("invalid Google identity must be rejected")
			}
		})
	}
}

func TestCreatorFindErrorDistinguishesMissingIdentityFromDatabaseFailure(t *testing.T) {
	missing := creatorFindError(mongo.ErrNoDocuments, apperr.ErrForbidden)
	if missing == nil || missing.Code != apperr.ErrForbidden.Code {
		t.Fatalf("missing identity code = %v, want FORBIDDEN", missing)
	}

	databaseFailure := creatorFindError(errors.New("database unavailable"), apperr.ErrForbidden)
	if databaseFailure == nil || databaseFailure.Code != apperr.ErrInternal.Code {
		t.Fatalf("database failure code = %v, want INTERNAL", databaseFailure)
	}
}

func TestHoldingAdminMembershipFilterRejectsDeletedMembership(t *testing.T) {
	filter := holdingAdminMembershipFilter("holding-1", "user-1")
	if filter["holdingcode"] != "holding-1" || filter["useruid"] != "user-1" || filter["isdeleted"] != false {
		t.Fatalf("unsafe membership filter: %#v", filter)
	}
}
