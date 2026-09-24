package rolepermission

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/goapi/language"
	orgpolicy "smlcloudplatform/internal/organization/access"
	"smlcloudplatform/pkg/microservice"
)

// postgresMemberRole is the caller's role and permission sets in the selected Holding. A
// membership past its access expiry date (usable through the END of that date in the Holding's
// timezone — authmodels.AccessEndsAt) is orgpolicy.ErrAccessExpired, like every other place
// that resolves a membership.
func postgresMemberRole(ctx microservice.IContext, db *sql.DB, holding string) (string, []byte, error) {
	u := ctx.UserInfo()
	if holding == "" || u.UID == "" || holding != u.HoldingCode {
		return "", nil, errors.New("permission denied")
	}
	requestCtx, cancel := context.WithTimeout(ctx.Request().Context(), rolePermissionTimeout)
	defer cancel()
	var role, timezone string
	var sets []byte
	var expiryDate sql.NullTime
	err := db.QueryRowContext(requestCtx, `SELECT m.role, m.permission_sets, m.access_expiry_date, COALESCE(h.profile->'settings'->>'timezone', '')
		FROM holding_members m JOIN users u ON u.id=m.user_id AND u.is_active=true JOIN holdings h ON h.code=m.holding_code AND h.is_active=true
		WHERE m.holding_code=$1 AND m.user_id::text=$2 AND m.is_active=true`, holding, u.UID).Scan(&role, &sets, &expiryDate, &timezone)
	if err != nil {
		return "", nil, err
	}
	if expiryDate.Valid && authmodels.AccessExpired(authmodels.AccessEndsAt(expiryDate.Time, timezone), time.Now()) {
		return "", nil, orgpolicy.ErrAccessExpired
	}
	return strings.ToUpper(role), sets, nil
}

// respondMemberRoleError answers a refused membership: the expiry message (languages.tsv
// user_access_expired) when access expired, otherwise the given denial.
func respondMemberRoleError(ctx microservice.IContext, err error, denied string) {
	if errors.Is(err, orgpolicy.ErrAccessExpired) {
		denied = language.Text("user_access_expired", requestLanguage(ctx))
	}
	ctx.ResponseError(403, denied)
}

func authorizePostgresRoleManager(ctx microservice.IContext, db *sql.DB, holding string, roleCodes ...string) bool {
	role, _, err := postgresMemberRole(ctx, db, holding)
	if err != nil || (role != "OWNER" && role != "ADMIN") {
		respondMemberRoleError(ctx, err, "ปฏิเสธสิทธิ์: ต้องเป็นผู้ดูแลระบบของ Holding ที่เลือก")
		return false
	}
	for _, code := range roleCodes {
		if role != "OWNER" && !strings.EqualFold(strings.TrimSpace(code), "USER") {
			ctx.ResponseError(403, "ปฏิเสธสิทธิ์: เฉพาะเจ้าของ Holding จัดการชุดสิทธิ์นี้ได้")
			return false
		}
	}
	return true
}

// postgresTargetRole resolves the one stored row id names (resolvePermissionSet) and returns its
// code; otherwise it answers 404/409 itself.
func postgresTargetRole(ctx microservice.IContext, db *sql.DB, holding, id string) (string, bool) {
	role, err := resolvePermissionSet(ctx.Request().Context(), db, strings.TrimSpace(holding), strings.TrimSpace(id))
	if err != nil {
		_ = respondPermissionSetError(ctx, err)
		return "", false
	}
	return role, true
}
