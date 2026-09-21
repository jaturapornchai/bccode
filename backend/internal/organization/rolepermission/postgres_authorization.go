package rolepermission

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"smlcloudplatform/pkg/microservice"
)

func postgresMemberRole(ctx microservice.IContext, db *sql.DB, holding string) (string, []byte, error) {
	u := ctx.UserInfo()
	if holding == "" || u.UID == "" || holding != u.HoldingCode {
		return "", nil, errors.New("permission denied")
	}
	requestCtx, cancel := context.WithTimeout(ctx.Request().Context(), rolePermissionTimeout)
	defer cancel()
	var role string
	var sets []byte
	err := db.QueryRowContext(requestCtx, `SELECT m.role,m.permission_sets FROM holding_members m JOIN users u ON u.id=m.user_id AND u.is_active=true JOIN holdings h ON h.code=m.holding_code AND h.is_active=true WHERE m.holding_code=$1 AND m.user_id::text=$2 AND m.is_active=true`, holding, u.UID).Scan(&role, &sets)
	return strings.ToUpper(role), sets, err
}

func authorizePostgresRoleManager(ctx microservice.IContext, db *sql.DB, holding string, roleCodes ...string) bool {
	role, _, err := postgresMemberRole(ctx, db, holding)
	if err != nil || (role != "OWNER" && role != "ADMIN") {
		ctx.ResponseError(403, "ปฏิเสธสิทธิ์: ต้องเป็นผู้ดูแลระบบของ Holding ที่เลือก")
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

func postgresTargetRole(ctx microservice.IContext, db *sql.DB, holding, id string) (string, bool) {
	var role string
	err := db.QueryRowContext(ctx.Request().Context(), `SELECT role_code FROM role_permissions WHERE holding_code=$1 AND (id=$2 OR UPPER(role_code)=UPPER($2))`, holding, id).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.ResponseError(404, "ไม่พบชุดสิทธิ์")
		} else {
			ctx.ResponseError(503, "ไม่สามารถตรวจสอบชุดสิทธิ์ได้")
		}
		return "", false
	}
	return role, true
}
