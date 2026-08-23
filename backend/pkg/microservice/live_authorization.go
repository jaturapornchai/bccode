package microservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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

func newLiveAuthorization(finder AuthorizationFinder) *liveAuthorization {
	if finder == nil {
		return nil
	}
	return &liveAuthorization{finder: finder, timeNow: time.Now}
}

// Authorize resolves current authority from MongoDB. Redis session fields are
// only a versioned workspace selection; they are never accepted as authority.
func (a *liveAuthorization) Authorize(ctx context.Context, selected models.UserInfo) (models.UserInfo, error) {
	selected.UID = strings.TrimSpace(selected.UID)
	if selected.UID == "" {
		return models.UserInfo{}, ErrLiveUserAccess
	}

	var user liveUserRecord
	if err := a.finder.FindOne(ctx, &liveUserRecord{}, bson.M{"uid": selected.UID}, &user); err != nil {
		return models.UserInfo{}, liveLookupError(err, ErrLiveUserAccess, "user")
	}
	if user.UID == "" || user.IsDeleted || !user.DisabledAt.IsZero() {
		return models.UserInfo{}, ErrLiveUserAccess
	}

	selected.HoldingCode = strings.TrimSpace(selected.HoldingCode)
	if selected.HoldingCode == "" {
		return loginOnlyUserInfo(selected), nil
	}

	var membership liveMembershipRecord
	if err := a.finder.FindOne(ctx, &liveMembershipRecord{}, bson.M{
		"holdingcode": selected.HoldingCode,
		"useruid":     selected.UID,
	}, &membership); err != nil {
		return models.UserInfo{}, liveLookupError(err, ErrLiveWorkspaceAccess, "membership")
	}
	now := a.timeNow().UTC()
	if membership.UserUID == "" || membership.IsDeleted || membership.IsAccessDisabled ||
		membership.Role > 2 || (!membership.AccessExpiryDate.IsZero() && !now.Before(membership.AccessExpiryDate)) {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}
	if selected.PermissionVersion != membership.PermissionVersion || selected.Role != membership.Role {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}
	if selected.MembershipUID != "" && membership.MembershipUID != "" && selected.MembershipUID != membership.MembershipUID {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}
	if selected.HoldingUID != "" && membership.HoldingUID != "" && selected.HoldingUID != membership.HoldingUID {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
	}

	var holding liveHoldingRecord
	if err := a.finder.FindOne(ctx, &liveHoldingRecord{}, bson.M{"holdingcode": selected.HoldingCode}, &holding); err != nil {
		return models.UserInfo{}, liveLookupError(err, ErrLiveWorkspaceAccess, "Holding")
	}
	if holding.HoldingCode == "" || !holding.IsActive || holding.DeletedAt != nil {
		return models.UserInfo{}, ErrLiveWorkspaceAccess
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

	var company liveCompanyRecord
	if err := a.finder.FindOne(ctx, &liveCompanyRecord{}, bson.M{
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
	if err := a.finder.FindOne(ctx, &liveBranchRecord{}, bson.M{
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

	// Resolve membership first so Authorize can version-fence the selection.
	var membership liveMembershipRecord
	if err := a.finder.FindOne(ctx, &liveMembershipRecord{}, bson.M{
		"holdingcode": identity.HoldingCode,
		"useruid":     strings.TrimSpace(identity.UID),
	}, &membership); err != nil {
		return models.UserInfo{}, liveLookupError(err, ErrLiveWorkspaceAccess, "membership")
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
