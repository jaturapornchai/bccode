package access

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	authmodels "smlcloudplatform/internal/authentication/models"
	micromodels "smlcloudplatform/pkg/microservice/models"
)

var (
	ErrActiveMembershipRequired = errors.New("active membership required")
	ErrHoldingManagerRequired   = errors.New("Holding OWNER or ADMIN required")
	ErrHoldingOwnerRequired     = errors.New("Holding OWNER required")
	ErrActiveHoldingRequired    = errors.New("active Holding required")
)

// Querier is satisfied by *sql.DB and *sql.Tx on the central control database.
type Querier interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// Holding is the active Holding row from the central holdings table.
type Holding struct {
	Code string
	Name string
}

func RequireActiveHolding(ctx context.Context, db Querier, holdingCode string) error {
	_, err := FindActiveHolding(ctx, db, holdingCode)
	return err
}

func FindActiveHolding(ctx context.Context, db Querier, holdingCode string) (Holding, error) {
	holdingCode = strings.TrimSpace(holdingCode)
	if holdingCode == "" {
		return Holding{}, ErrActiveHoldingRequired
	}
	var holding Holding
	err := db.QueryRowContext(ctx, `SELECT code, name FROM holdings WHERE code = $1 AND is_active = true`, holdingCode).Scan(&holding.Code, &holding.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return Holding{}, ErrActiveHoldingRequired
	}
	if err != nil {
		return Holding{}, fmt.Errorf("load Holding: %w", err)
	}
	return holding, nil
}

// FindActiveMembership resolves the caller's current membership of the selected Holding.
// Inactive users, memberships and Holdings fail closed. OWNER/ADMIN memberships without
// explicit scopes cover every active company of the Holding.
func FindActiveMembership(ctx context.Context, db Querier, userInfo micromodels.UserInfo, _ time.Time) (authmodels.ShopUser, error) {
	holdingCode := strings.TrimSpace(userInfo.HoldingCode)
	userUID := strings.TrimSpace(userInfo.UID)
	if holdingCode == "" || userUID == "" {
		return authmodels.ShopUser{}, ErrActiveMembershipRequired
	}

	var (
		membership     authmodels.ShopUser
		roleText       string
		permissionSets []byte
		accessScopes   []byte
	)
	err := db.QueryRowContext(ctx, `
		SELECT m.id::text, u.id::text, u.username, m.role, m.permission_sets, m.access_scopes
		FROM holding_members m
		JOIN users u ON u.id = m.user_id AND u.is_active = true
		JOIN holdings h ON h.code = m.holding_code AND h.is_active = true
		WHERE m.holding_code = $1 AND m.user_id::text = $2 AND m.is_active = true`,
		holdingCode, userUID).Scan(&membership.MembershipUID, &membership.UserUID, &membership.Username, &roleText, &permissionSets, &accessScopes)
	if errors.Is(err, sql.ErrNoRows) {
		return authmodels.ShopUser{}, ErrActiveMembershipRequired
	}
	if err != nil {
		return authmodels.ShopUser{}, fmt.Errorf("load membership: %w", err)
	}
	role := authmodels.RoleFromText(roleText)
	membership.ID = membership.MembershipUID
	membership.HoldingCode = holdingCode
	membership.HoldingUID = holdingCode
	membership.Role = role
	if len(permissionSets) > 0 && json.Unmarshal(permissionSets, &membership.PermissionSets) != nil {
		return authmodels.ShopUser{}, fmt.Errorf("load membership: invalid permission sets")
	}
	scopes, err := authmodels.ParseAccessScopes(accessScopes)
	if err != nil {
		return authmodels.ShopUser{}, fmt.Errorf("load membership: %w", err)
	}
	if len(scopes) == 0 && (role == authmodels.ROLE_OWNER || role == authmodels.ROLE_ADMIN) {
		if scopes, err = allCompanyScopes(ctx, db, holdingCode); err != nil {
			return authmodels.ShopUser{}, err
		}
	}
	membership.AccessScopes = scopes
	return membership, nil
}

func allCompanyScopes(ctx context.Context, db Querier, holdingCode string) ([]authmodels.AccessScope, error) {
	rows, err := db.QueryContext(ctx, `SELECT code FROM companies WHERE holding_code = $1 AND is_active = true ORDER BY code`, holdingCode)
	if err != nil {
		return nil, fmt.Errorf("load Holding companies: %w", err)
	}
	defer rows.Close()
	scopes := []authmodels.AccessScope{}
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("load Holding companies: %w", err)
		}
		scopes = append(scopes, authmodels.AccessScope{ScopeType: "company", CompanyUID: code, BusinessCode: code, AllBranches: true})
	}
	return scopes, rows.Err()
}

func FindActiveHoldingManager(ctx context.Context, db Querier, userInfo micromodels.UserInfo, now time.Time) (authmodels.ShopUser, error) {
	membership, err := FindActiveMembership(ctx, db, userInfo, now)
	if err != nil {
		return authmodels.ShopUser{}, err
	}
	if membership.Role != authmodels.ROLE_OWNER && membership.Role != authmodels.ROLE_ADMIN {
		return authmodels.ShopUser{}, ErrHoldingManagerRequired
	}
	return membership, nil
}

func FindActiveHoldingOwner(ctx context.Context, db Querier, userInfo micromodels.UserInfo, now time.Time) (authmodels.ShopUser, error) {
	membership, err := FindActiveMembership(ctx, db, userInfo, now)
	if err != nil {
		return authmodels.ShopUser{}, err
	}
	if membership.Role != authmodels.ROLE_OWNER {
		return authmodels.ShopUser{}, ErrHoldingOwnerRequired
	}
	return membership, nil
}
