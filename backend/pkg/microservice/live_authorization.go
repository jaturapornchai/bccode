package microservice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	authmodels "smlcloudplatform/internal/authentication/models"
	orgpolicy "smlcloudplatform/internal/organization/access"
	"smlcloudplatform/pkg/microservice/models"
)

var (
	ErrLiveUserAccess      = errors.New("user is not active")
	ErrLiveWorkspaceAccess = errors.New("workspace access changed")
)

// AuthorizationFinder is the central PostgreSQL handle (*sql.DB) that live
// authorization reads users, memberships, companies and branches from.
type AuthorizationFinder interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

const liveAuthorizationTimeout = 2 * time.Second

type liveAuthorization struct {
	db      AuthorizationFinder
	timeNow func() time.Time
}

func newLiveAuthorization(db AuthorizationFinder) *liveAuthorization {
	if db == nil {
		return nil
	}
	return &liveAuthorization{db: db, timeNow: time.Now}
}

// Authorize re-resolves the caller's current authority on every request, so a
// disabled user, membership, Holding, company or branch loses access immediately.
func (a *liveAuthorization) Authorize(ctx context.Context, selected models.UserInfo) (models.UserInfo, error) {
	selected.UID = strings.TrimSpace(selected.UID)
	if _, err := uuid.Parse(selected.UID); err != nil {
		return models.UserInfo{}, ErrLiveUserAccess
	}
	ctx, cancel := context.WithTimeout(ctx, liveAuthorizationTimeout)
	defer cancel()

	var active bool
	err := a.db.QueryRowContext(ctx, `SELECT is_active FROM users WHERE id = $1`, selected.UID).Scan(&active)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !active) {
		return models.UserInfo{}, ErrLiveUserAccess
	}
	if err != nil {
		return models.UserInfo{}, fmt.Errorf("load live user authorization: %w", err)
	}

	selected.HoldingCode = strings.TrimSpace(selected.HoldingCode)
	if selected.HoldingCode == "" {
		return loginOnlyUserInfo(selected), nil
	}
	membership, err := orgpolicy.FindActiveMembership(ctx, a.db, selected, a.timeNow().UTC())
	if errors.Is(err, orgpolicy.ErrActiveMembershipRequired) {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}
	if err != nil {
		return models.UserInfo{}, fmt.Errorf("load live membership authorization: %w", err)
	}
	if membership.Role > authmodels.ROLE_OWNER {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}

	selected.MembershipUID = membership.MembershipUID
	selected.HoldingUID = selected.HoldingCode
	selected.Role = membership.Role
	selected.BusinessCode = strings.ToUpper(strings.TrimSpace(selected.BusinessCode))
	selected.CompanyUID = strings.TrimSpace(selected.CompanyUID)
	selected.BranchUID = strings.TrimSpace(selected.BranchUID)
	if selected.BusinessCode == "" {
		selected.CompanyUID = ""
		selected.BranchUID = ""
		return selected, nil
	}

	var companyCode string
	err = a.db.QueryRowContext(ctx, `SELECT code FROM companies WHERE holding_code = $1 AND UPPER(code) = $2 AND is_active = true`,
		selected.HoldingCode, selected.BusinessCode).Scan(&companyCode)
	if errors.Is(err, sql.ErrNoRows) {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}
	if err != nil {
		return models.UserInfo{}, fmt.Errorf("load live company authorization: %w", err)
	}
	if selected.CompanyUID != "" && !strings.EqualFold(selected.CompanyUID, companyCode) {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}
	selected.CompanyUID = companyCode

	if selected.BranchUID == "" {
		if !allowsCompanyWorkspace(membership.AccessScopes, companyCode) {
			return models.UserInfo{}, ErrLiveWorkspaceAccess
		}
		return selected, nil
	}
	var branchActive bool
	err = a.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM branches
		WHERE holding_code = $1 AND company_code = $2 AND code = $3 AND is_active = true)`,
		selected.HoldingCode, companyCode, selected.BranchUID).Scan(&branchActive)
	if err != nil {
		return models.UserInfo{}, fmt.Errorf("load live branch authorization: %w", err)
	}
	if !branchActive || !orgpolicy.AllowsBranch(membership.AccessScopes, companyCode, selected.BranchUID) {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}
	return selected, nil
}

// SelectWorkspace switches the session to another Holding/company/branch; the
// stable IDs are resolved from the database, never trusted from the old token.
func (a *liveAuthorization) SelectWorkspace(ctx context.Context, identity models.UserInfo, holdingCode, businessCode, branchUID string) (models.UserInfo, error) {
	identity.HoldingCode = strings.TrimSpace(holdingCode)
	identity.BusinessCode = strings.ToUpper(strings.TrimSpace(businessCode))
	identity.BranchUID = strings.TrimSpace(branchUID)
	identity.CompanyUID = ""
	if identity.BusinessCode == "" {
		identity.BranchUID = ""
	}
	if identity.HoldingCode == "" {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}
	return a.Authorize(ctx, identity)
}

func loginOnlyUserInfo(user models.UserInfo) models.UserInfo {
	return models.UserInfo{
		Username:   user.Username,
		Name:       user.Name,
		UID:        user.UID,
		SessionUID: user.SessionUID,
	}
}

// allowsCompanyWorkspace requires a company-level scope; a branch-only scope
// does not open the whole company.
func allowsCompanyWorkspace(scopes []authmodels.AccessScope, companyUID string) bool {
	for _, scope := range scopes {
		if strings.EqualFold(strings.TrimSpace(scope.ScopeType), "company") &&
			strings.EqualFold(strings.TrimSpace(scope.CompanyUID), companyUID) && strings.TrimSpace(scope.BranchUID) == "" {
			return true
		}
	}
	return false
}
