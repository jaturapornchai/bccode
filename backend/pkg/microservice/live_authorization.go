package microservice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"smlcloudplatform/pkg/microservice/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrLiveUserAccess      = errors.New("user is not active")
	ErrLiveWorkspaceAccess = errors.New("workspace access changed")
)

// AuthorizationFinder is deliberately smaller than IPersisterMongo so the
// authorization policy can be tested without mocking the whole persistence API.
type AuthorizationFinder interface {
	FindOne(context.Context, interface{}, interface{}, interface{}, ...*options.FindOneOptions) error
}

type liveAuthorization struct {
	finder  AuthorizationFinder
	timeNow func() time.Time
}

type liveUserRecord struct {
	UID        string    `bson:"uid"`
	DisabledAt time.Time `bson:"disabledat,omitempty"`
	IsDeleted  bool      `bson:"isdeleted"`
}

func (liveUserRecord) CollectionName() string { return "users" }

type liveAccessScope struct {
	ScopeType   string `bson:"scopetype"`
	CompanyUID  string `bson:"companyuid,omitempty"`
	BranchUID   string `bson:"branchuid,omitempty"`
	AllBranches bool   `bson:"allbranches,omitempty"`
}

type liveMembershipRecord struct {
	MembershipUID     string            `bson:"membershipuid"`
	HoldingUID        string            `bson:"holdinguid"`
	HoldingCode       string            `bson:"holdingcode"`
	UserUID           string            `bson:"useruid"`
	Role              uint8             `bson:"role"`
	PermissionVersion int64             `bson:"permissionversion"`
	IsDeleted         bool              `bson:"isdeleted"`
	IsAccessDisabled  bool              `bson:"isaccessdisabled"`
	AccessExpiryDate  time.Time         `bson:"accessexpirydate,omitempty"`
	AccessScopes      []liveAccessScope `bson:"accessscopes,omitempty"`
}

func (liveMembershipRecord) CollectionName() string { return "shopusers" }

type liveHoldingRecord struct {
	HoldingCode string     `bson:"holdingcode"`
	IsActive    bool       `bson:"isactive"`
	DeletedAt   *time.Time `bson:"deletedat,omitempty"`
}

func (liveHoldingRecord) CollectionName() string { return "shops" }

type liveCompanyRecord struct {
	HoldingCode string     `bson:"holdingcode"`
	CompanyUID  string     `bson:"guidfixed"`
	Code        string     `bson:"code"`
	IsActive    bool       `bson:"isactive"`
	DeletedAt   *time.Time `bson:"deletedat,omitempty"`
}

func (liveCompanyRecord) CollectionName() string { return "organizationcompanies" }

type liveBranchRecord struct {
	HoldingCode string     `bson:"holdingcode"`
	CompanyUID  string     `bson:"companyguid"`
	BranchUID   string     `bson:"guidfixed"`
	IsActive    bool       `bson:"isactive"`
	DeletedAt   *time.Time `bson:"deletedat,omitempty"`
}

func (liveBranchRecord) CollectionName() string { return "organizationbranches" }

var (
	livePgDB   *sql.DB
	livePgOnce sync.Once
)

func getLivePgDB() *sql.DB {
	livePgOnce.Do(func() {
		host := os.Getenv("POSTGRES_HOST")
		if host == "" {
			host = "postgres"
		}
		port := os.Getenv("POSTGRES_PORT")
		if port == "" {
			port = "5432"
		}
		user := os.Getenv("POSTGRES_USER")
		if user == "" {
			user = os.Getenv("POSTGRES_USERNAME")
		}
		if user == "" {
			user = "bcai"
		}
		pass := os.Getenv("POSTGRES_PASSWORD")
		dbName := os.Getenv("POSTGRES_DB")
		if dbName == "" {
			dbName = os.Getenv("POSTGRES_DATABASE")
		}
		if dbName == "" {
			dbName = "bcai_projection"
		}
		sslMode := os.Getenv("POSTGRES_SSL_MODE")
		if sslMode == "" {
			sslMode = "disable"
		}

		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, pass, dbName, sslMode)
		db, err := sql.Open("postgres", connStr)
		if err == nil {
			db.SetMaxOpenConns(10)
			db.SetMaxIdleConns(5)
			db.SetConnMaxLifetime(5 * time.Minute)
			livePgDB = db
		}
	})
	return livePgDB
}

func newLiveAuthorization(finder AuthorizationFinder) *liveAuthorization {
	if finder == nil {
		return nil
	}
	return &liveAuthorization{finder: finder, timeNow: time.Now}
}

// Authorize resolves current authority from PostgreSQL / MongoDB.
func (a *liveAuthorization) Authorize(ctx context.Context, selected models.UserInfo) (models.UserInfo, error) {
	selected.UID = strings.TrimSpace(selected.UID)
	if selected.UID == "" {
		return models.UserInfo{}, ErrLiveUserAccess
	}

	var user liveUserRecord
	var foundUserInPg bool
	if pg := getLivePgDB(); pg != nil {
		var uID string
		var isActive bool
		err := pg.QueryRowContext(ctx, "SELECT id, is_active FROM users WHERE (id = $1 OR LOWER(username) = LOWER($1)) LIMIT 1", selected.UID).Scan(&uID, &isActive)
		if err == nil && isActive {
			user = liveUserRecord{UID: uID}
			foundUserInPg = true
		}
	}

	if !foundUserInPg {
		qCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if err := a.finder.FindOne(qCtx, &liveUserRecord{}, bson.M{"uid": selected.UID}, &user); err != nil {
			return models.UserInfo{}, liveLookupError(err, ErrLiveUserAccess, "user")
		}
	}
	if user.UID == "" || user.IsDeleted || !user.DisabledAt.IsZero() {
		return models.UserInfo{}, ErrLiveUserAccess
	}

	selected.HoldingCode = strings.TrimSpace(selected.HoldingCode)
	if selected.HoldingCode == "" {
		return loginOnlyUserInfo(selected), nil
	}

	var membership liveMembershipRecord
	var foundMemInPg bool
	if pg := getLivePgDB(); pg != nil {
		var mID, hCode string
		var role int
		var isAccessDisabled bool
		err := pg.QueryRowContext(ctx, `
			SELECT hm.id, hm.holding_code, COALESCE(hm.role, 1), COALESCE(hm.is_access_disabled, false)
			FROM holding_members hm
			JOIN users u ON hm.user_id = u.id
			WHERE LOWER(hm.holding_code) = LOWER($1) AND (u.id = $2 OR LOWER(u.username) = LOWER($2))
			LIMIT 1`, selected.HoldingCode, selected.UID).Scan(&mID, &hCode, &role, &isAccessDisabled)
		if err == nil && !isAccessDisabled {
			membership = liveMembershipRecord{
				MembershipUID: mID,
				HoldingCode:   hCode,
				UserUID:       selected.UID,
				Role:          uint8(role),
			}
			selected.Role = uint8(role)
			foundMemInPg = true
		}
	}

	if !foundMemInPg {
		qCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if err := a.finder.FindOne(qCtx, &liveMembershipRecord{}, bson.M{
			"holdingcode": selected.HoldingCode,
			"useruid":     selected.UID,
		}, &membership); err != nil {
			return models.UserInfo{}, liveLookupError(err, ErrLiveWorkspaceAccess, "membership")
		}
	}
	now := a.timeNow().UTC()
	if membership.UserUID == "" || membership.IsDeleted || membership.IsAccessDisabled ||
		membership.Role > 2 || (!membership.AccessExpiryDate.IsZero() && !now.Before(membership.AccessExpiryDate)) {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}
	if !foundMemInPg && (selected.PermissionVersion != membership.PermissionVersion || selected.Role != membership.Role) {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}
	if !foundMemInPg && selected.MembershipUID != "" && membership.MembershipUID != "" && selected.MembershipUID != membership.MembershipUID {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}
	if !foundMemInPg && selected.HoldingUID != "" && membership.HoldingUID != "" && selected.HoldingUID != membership.HoldingUID {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}

	if !foundMemInPg {
		var holding liveHoldingRecord
		qCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if err := a.finder.FindOne(qCtx, &liveHoldingRecord{}, bson.M{"holdingcode": selected.HoldingCode}, &holding); err != nil {
			return models.UserInfo{}, liveLookupError(err, ErrLiveWorkspaceAccess, "Holding")
		}
		if holding.HoldingCode == "" || !holding.IsActive || holding.DeletedAt != nil {
			return models.UserInfo{}, ErrLiveWorkspaceAccess
		}
	}

	selected.MembershipUID = membership.MembershipUID
	selected.HoldingUID = membership.HoldingUID
	selected.PermissionVersion = membership.PermissionVersion
	selected.Role = membership.Role
	selected.BusinessCode = strings.ToUpper(strings.TrimSpace(selected.BusinessCode))
	selected.CompanyUID = strings.TrimSpace(selected.CompanyUID)
	selected.BranchUID = strings.TrimSpace(selected.BranchUID)

	if selected.BusinessCode == "" {
		selected.CompanyUID = ""
		selected.BranchUID = ""
		return selected, nil
	}

	if foundMemInPg {
		// In PostgreSQL mode without company/branch scopes restriction, allow selected workspace
		return selected, nil
	}

	var company liveCompanyRecord
	qCtxCompany, cancelCompany := context.WithTimeout(ctx, 2*time.Second)
	defer cancelCompany()
	if err := a.finder.FindOne(qCtxCompany, &liveCompanyRecord{}, bson.M{
		"holdingcode": selected.HoldingCode,
		"code":        selected.BusinessCode,
	}, &company); err != nil {
		return models.UserInfo{}, liveLookupError(err, ErrLiveWorkspaceAccess, "Company")
	}
	if company.CompanyUID == "" || !company.IsActive || company.DeletedAt != nil {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}
	if selected.CompanyUID != "" && selected.CompanyUID != company.CompanyUID {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}
	selected.CompanyUID = company.CompanyUID

	if selected.BranchUID == "" {
		if !allowsCompanyWorkspace(membership.AccessScopes, company.CompanyUID) {
			return models.UserInfo{}, ErrLiveWorkspaceAccess
		}
		return selected, nil
	}
	var branch liveBranchRecord
	qCtxBranch, cancelBranch := context.WithTimeout(ctx, 2*time.Second)
	defer cancelBranch()
	if err := a.finder.FindOne(qCtxBranch, &liveBranchRecord{}, bson.M{
		"holdingcode": selected.HoldingCode,
		"companyguid": company.CompanyUID,
		"guidfixed":   selected.BranchUID,
	}, &branch); err != nil {
		return models.UserInfo{}, liveLookupError(err, ErrLiveWorkspaceAccess, "Branch")
	}
	if branch.BranchUID == "" || !branch.IsActive || branch.DeletedAt != nil ||
		!allowsBranchWorkspace(membership.AccessScopes, company.CompanyUID, branch.BranchUID) {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}
	return selected, nil
}

func (a *liveAuthorization) SelectWorkspace(ctx context.Context, identity models.UserInfo, holdingCode, businessCode, branchUID string) (models.UserInfo, error) {
	identity.HoldingCode = strings.TrimSpace(holdingCode)
	identity.BusinessCode = strings.ToUpper(strings.TrimSpace(businessCode))
	identity.BranchUID = strings.TrimSpace(branchUID)
	if identity.BusinessCode == "" {
		identity.BranchUID = ""
	}
	if identity.HoldingCode == "" {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}

	var membership liveMembershipRecord
	var foundMemInPg bool
	if pg := getLivePgDB(); pg != nil {
		var mID, hCode string
		var role int
		var isAccessDisabled bool
		err := pg.QueryRowContext(ctx, `
			SELECT hm.id, hm.holding_code, COALESCE(hm.role, 1), COALESCE(hm.is_access_disabled, false)
			FROM holding_members hm
			JOIN users u ON hm.user_id = u.id
			WHERE LOWER(hm.holding_code) = LOWER($1) AND (u.id = $2 OR LOWER(u.username) = LOWER($2))
			LIMIT 1`, identity.HoldingCode, strings.TrimSpace(identity.UID)).Scan(&mID, &hCode, &role, &isAccessDisabled)
		if err == nil && !isAccessDisabled {
			membership = liveMembershipRecord{
				MembershipUID: mID,
				HoldingCode:   hCode,
				UserUID:       strings.TrimSpace(identity.UID),
				Role:          uint8(role),
			}
			foundMemInPg = true
		}
	}

	if !foundMemInPg {
		qCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		if err := a.finder.FindOne(qCtx, &liveMembershipRecord{}, bson.M{
			"holdingcode": identity.HoldingCode,
			"useruid":     strings.TrimSpace(identity.UID),
		}, &membership); err != nil {
			return models.UserInfo{}, liveLookupError(err, ErrLiveWorkspaceAccess, "membership")
		}
	}
	identity.MembershipUID = membership.MembershipUID
	identity.HoldingUID = membership.HoldingUID
	identity.PermissionVersion = membership.PermissionVersion
	identity.Role = membership.Role
	return a.Authorize(ctx, identity)
}

func liveLookupError(err error, notFound error, entity string) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return notFound
	}
	return fmt.Errorf("load live %s authorization: %w", entity, err)
}

func loginOnlyUserInfo(user models.UserInfo) models.UserInfo {
	return models.UserInfo{
		Username:   user.Username,
		Name:       user.Name,
		UID:        user.UID,
		SessionUID: user.SessionUID,
	}
}

func allowsCompanyWorkspace(scopes []liveAccessScope, companyUID string) bool {
	for _, scope := range scopes {
		if strings.EqualFold(strings.TrimSpace(scope.ScopeType), "company") &&
			strings.TrimSpace(scope.CompanyUID) == companyUID && strings.TrimSpace(scope.BranchUID) == "" {
			return true
		}
	}
	return false
}

func allowsBranchWorkspace(scopes []liveAccessScope, companyUID, branchUID string) bool {
	for _, scope := range scopes {
		if strings.TrimSpace(scope.CompanyUID) != companyUID {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(scope.ScopeType), "company") && scope.AllBranches {
			return true
		}
		if strings.EqualFold(strings.TrimSpace(scope.ScopeType), "branch") && strings.TrimSpace(scope.BranchUID) == branchUID {
			return true
		}
	}
	return false
}
