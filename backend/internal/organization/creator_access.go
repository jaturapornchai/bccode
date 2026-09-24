package organization

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/demo"
	orgpolicy "smlcloudplatform/internal/organization/access"
	"smlcloudplatform/pkg/apperr"
	micromodels "smlcloudplatform/pkg/microservice/models"
)

var (
	errCreatorMissing  = apperr.ErrUnauthorized.WithMessage("authenticated user no longer exists").WithThaiMessage("ไม่พบบัญชีผู้ใช้ที่เข้าสู่ระบบ")
	errCreatorDisabled = apperr.ErrForbidden.WithMessage("account is disabled").WithThaiMessage("บัญชีผู้ใช้ถูกปิดใช้งาน")
	errGoogleRequired  = apperr.ErrForbidden.WithMessage("an active verified Google identity is required").WithThaiMessage("ต้องเชื่อม Google Identity ที่ยืนยันแล้วก่อนสร้างองค์กร")
	errHoldingRequired = apperr.ErrForbidden.WithMessage("a Holding must be selected").WithThaiMessage("กรุณาเลือก Holding ก่อนสร้างข้อมูลโครงสร้างองค์กร")
	errManagerRequired = apperr.ErrForbidden.WithMessage("Holding OWNER or ADMIN permission is required").WithThaiMessage("เฉพาะ OWNER หรือ ADMIN ของ Holding เท่านั้นที่สร้างได้")
)

// RequireEmailedAccount authorizes the root Holding bootstrap. A Holding has no
// scoped role until after it exists; its creator becomes OWNER in ShopService.
func RequireEmailedAccount(ctx context.Context, db orgpolicy.Querier, userInfo micromodels.UserInfo) *apperr.AppError {
	return requireOrganizationCreator(ctx, db, userInfo, false)
}

// RequireHoldingAdmin authorizes Company/Branch creation from the current Holding.
func RequireHoldingAdmin(ctx context.Context, db orgpolicy.Querier, userInfo micromodels.UserInfo) *apperr.AppError {
	return requireOrganizationCreator(ctx, db, userInfo, true)
}

func requireOrganizationCreator(ctx context.Context, db orgpolicy.Querier, userInfo micromodels.UserInfo, requireAdmin bool) *apperr.AppError {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	userUID := strings.TrimSpace(userInfo.UID)
	if userUID == "" {
		return errCreatorMissing
	}
	var (
		username    string
		isActive    bool
		hasIdentity bool
	)
	err := db.QueryRowContext(ctx, `
		SELECT u.username, u.is_active,
		       EXISTS (SELECT 1 FROM user_identities i
		               WHERE i.user_id = u.id AND i.provider = 'google'
		                 AND COALESCE(i.extra->>'active', 'true') <> 'false'
		                 AND COALESCE(i.extra->>'revokedAt', '') = '')
		FROM users u WHERE u.id::text = $1`, userUID).Scan(&username, &isActive, &hasIdentity)
	if err != nil {
		return creatorFindError(err, errCreatorMissing)
	}
	var disabledAt time.Time
	if !isActive {
		disabledAt = time.Now()
	}
	if policyErr := validateCreatorPolicy(disabledAt, authmodels.ROLE_USER, false); policyErr != nil {
		return policyErr
	}
	// The public demo account has no Google identity by design (docs/login.md "บัญชี Demo").
	if !demo.IsDemoUser(username) && !hasIdentity {
		return errGoogleRequired
	}
	if !requireAdmin {
		return nil
	}
	if strings.TrimSpace(userInfo.HoldingCode) == "" {
		return errHoldingRequired
	}
	membership, err := orgpolicy.FindActiveMembership(ctx, db, userInfo, time.Now())
	if err != nil {
		if expired := AccessExpiredError(err, "th"); expired != nil {
			return expired
		}
		if errors.Is(err, orgpolicy.ErrActiveMembershipRequired) {
			return errManagerRequired
		}
		return apperr.ErrInternal.WithWrap(err)
	}
	return validateCreatorPolicy(time.Time{}, membership.Role, true)
}

func creatorFindError(err error, notFound *apperr.AppError) *apperr.AppError {
	if errors.Is(err, sql.ErrNoRows) {
		return notFound
	}
	return apperr.ErrInternal.WithWrap(err)
}

func validateCreatorPolicy(disabledAt time.Time, role authmodels.UserRole, requireAdmin bool) *apperr.AppError {
	if !disabledAt.IsZero() {
		return errCreatorDisabled
	}
	if requireAdmin && role != authmodels.ROLE_ADMIN && role != authmodels.ROLE_OWNER {
		return errManagerRequired
	}
	return nil
}
