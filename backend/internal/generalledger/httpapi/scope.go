package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	authmodels "smlcloudplatform/internal/authentication/models"
	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/mcptoken"
	"smlcloudplatform/pkg/microservice"
)

func resolveScope(ctx context.Context, request microservice.IContext, connect func(string) (*sql.DB, error)) (requestScope, error) {
	u := request.UserInfo()
	result := requestScope{
		Scope:       gl.Scope{Holding: u.HoldingCode, Company: u.BusinessCode, Actor: u.UID},
		Permissions: map[string]bool{},
	}
	if u.HoldingCode == "" || u.BusinessCode == "" {
		return result, fmt.Errorf("กรุณาเลือกบริษัทก่อนใช้งานบัญชี")
	}
	if !validHoldingRegex.MatchString(u.HoldingCode) || u.UID == "" {
		return result, errScopeDenied
	}
	// Membership and organization metadata live centrally; journal data stays in the Holding database.
	db, err := connect(mcptoken.ControlDatabase)
	if err != nil || db == nil {
		return result, errScopeDenied
	}
	var companyCode string
	if err := db.QueryRowContext(ctx, `SELECT code FROM companies WHERE holding_code=$1 AND code=$2 AND is_active=true`, u.HoldingCode, u.BusinessCode).Scan(&companyCode); err != nil {
		return result, errScopeDenied
	}
	if u.BranchUID != "" {
		var branchCode string
		err := db.QueryRowContext(ctx, `SELECT code FROM branches WHERE holding_code = $1 AND code = $2 AND company_code = $3 AND is_active = true`, u.HoldingCode, u.BranchUID, u.BusinessCode).Scan(&branchCode)
		if err != nil || branchCode == "" {
			return result, errScopeDenied
		}
		result.Scope.Branch = branchCode
	}
	var role, timezone string
	var permSetsJSON, accessScopesJSON []byte
	var expiryDate sql.NullTime
	err = db.QueryRowContext(ctx, `SELECT m.role, m.permission_sets, m.access_scopes, m.access_expiry_date, COALESCE(h.profile->'settings'->>'timezone', '') FROM holding_members m JOIN users u ON u.id=m.user_id JOIN holdings h ON h.code=m.holding_code WHERE m.holding_code=$1 AND m.user_id::text=$2 AND m.is_active=true AND u.is_active=true AND h.is_active=true`, u.HoldingCode, u.UID).Scan(&role, &permSetsJSON, &accessScopesJSON, &expiryDate, &timezone)
	if err != nil || strings.TrimSpace(role) == "" {
		return result, errScopeDenied
	}
	// วันหมดอายุการเข้าใช้งาน: ใช้ได้ถึงสิ้นวันตามเขตเวลาของกลุ่มกิจการ — กฎเดียวกับ login/organization
	// (authmodels.AccessEndsAt); session หรือ API/MCP token ที่ออกก่อนหมดอายุต้องถูกปิดที่นี่ด้วย
	if expiryDate.Valid && authmodels.AccessExpired(authmodels.AccessEndsAt(expiryDate.Time, timezone), time.Now()) {
		return result, errAccessExpired
	}
	manager := strings.EqualFold(role, "OWNER") || strings.EqualFold(role, "ADMIN")
	// Every GL request revalidates membership scopes, API/MCP token requests included
	// (u.UID is then the token issuer): a token never reaches further than its issuer now.
	if !sessionScopeAllowed(accessScopesJSON, manager, companyCode, u.BranchUID) {
		return result, errScopeDenied
	}
	_, tokenRequest := request.(*mcpGLContext)
	companyWide := sessionScopeAllowed(accessScopesJSON, manager, companyCode, "")
	if tokenRequest {
		// Legacy tokens may retain a narrower branch grant than their owner.
		companyWide = u.BranchUID == ""
	}
	if manager {
		result.CompanyWide = companyWide
		result.Permissions["*"] = true
		return result, nil
	}
	var permSets []string
	if len(permSetsJSON) > 0 && json.Unmarshal(permSetsJSON, &permSets) != nil {
		return result, errScopeDenied
	}
	// Publish permissions only after every assigned role has been verified.
	permissions := map[string]bool{}
	for _, code := range append([]string{role}, permSets...) {
		var permsJSON []byte
		if err := db.QueryRowContext(ctx, `SELECT permissions FROM role_permissions WHERE holding_code = $1 AND UPPER(role_code) = UPPER($2) AND is_active=true`, u.HoldingCode, code).Scan(&permsJSON); err != nil {
			return result, errScopeDenied
		}
		var perms []string
		if err := json.Unmarshal(permsJSON, &perms); err != nil {
			return result, errScopeDenied
		}
		for _, permission := range perms {
			permissions[permission] = true
		}
	}
	result.CompanyWide = companyWide
	result.Permissions = permissions
	return result, nil
}

// sessionScopeAllowed checks the stored access scopes for the company/branch that
// resolveScope already verified as active. A holding rule covers them through the
// shared predicates; it grants access only, never GL permissions.
func sessionScopeAllowed(raw []byte, manager bool, company, branch string) bool {
	var scopes []authmodels.AccessScope
	// The central schema historically represents an empty manager grant as {}.
	if string(raw) == "{}" {
		return manager
	}
	if len(raw) > 0 && json.Unmarshal(raw, &scopes) != nil {
		return false
	}
	if manager && len(scopes) == 0 {
		return true
	}
	if branch != "" {
		return authmodels.ScopesAllowBranchSelection(scopes, company, branch)
	}
	return authmodels.ScopesAllowCompanySelection(scopes, company)
}

// Company operations must distinguish the selected branch from the current grant.
func (s requestScope) companyScope() (requestScope, error) {
	if !s.CompanyWide {
		return s, errScopeDenied
	}
	s.Scope.Branch = ""
	return s, nil
}

// ResolveSessionScope validates a browser session against the central PostgreSQL membership
// metadata (company, branch, access scopes) — shared by modules that post through the ledger.
func ResolveSessionScope(ctx context.Context, request microservice.IContext) (gl.Scope, error) {
	scope, err := resolveScope(ctx, request, mypg.PgSqlFastConnect)
	return scope.Scope, err
}
