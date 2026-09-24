package shop

import (
	"context"
	"errors"
	"smlcloudplatform/internal/authentication/models"
	common "smlcloudplatform/internal/models"
	orgpolicy "smlcloudplatform/internal/organization/access"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strings"
	"time"
)

type IShopUserService interface {
	SaveUserFullProfile(holdingCode string, authUsername string, req *models.UserRoleRequest) error
	DeleteUserPermissionShop(holdingCode string, authUsername string, username string) error

	// Holding admin management by email (holdingCode comes from the request, role of the
	// caller is resolved per-holding so it works from the holding-selection screen).
	ListHoldingMembersByAdmin(holdingCode string, authUsername string, pageable micromodels.Pageable) ([]models.ShopUserProfile, common.PaginationData, error)
	EnsureHoldingManager(holdingCode string, authUsername string) error

	InfoShopByUser(holdingCode string, username string) (models.ShopUserProfile, error)
	ListShopByUser(authUsername string, authUserUID string, pageable micromodels.Pageable) ([]models.ShopUserInfo, common.PaginationData, error)
	ListUserInShop(holdingCode string, authUsername string, pageable micromodels.Pageable) ([]models.ShopUserProfile, common.PaginationData, error)
}

type ShopUserService struct {
	repo IShopUserRepository
}

// errCreatorAccessExpiry: an expiry date was set on the Holding's creator.
var errCreatorAccessExpiry = errors.New("creator access expiry not allowed")

func NewShopUserService(shopUserRepo IShopUserRepository) ShopUserService {
	return ShopUserService{
		repo: shopUserRepo,
	}
}

func sameUsername(left string, right string) bool {
	return strings.EqualFold(utils.NormalizeUsername(left), utils.NormalizeUsername(right))
}

func applyAccessStatus(req *models.UserRoleRequest, existing models.ShopUser, authUsername string, now time.Time, isCreator bool) error {
	if isCreator {
		if req.IsAccessDisabled {
			return errors.New("creator_access_cannot_be_disabled")
		}
		// The Holding must always keep a working owner: no expiry date for its creator.
		if strings.TrimSpace(req.AccessExpiryDate) != "" {
			return errCreatorAccessExpiry
		}
		req.IsAccessDisabled = false
		req.AccessDisabledAt = time.Time{}
		req.AccessDisabledBy = ""
		req.AccessEnabledAt = existing.AccessEnabledAt
		req.AccessEnabledBy = existing.AccessEnabledBy
		return nil
	}

	if req.IsAccessDisabled {
		if existing.IsAccessDisabled && !existing.AccessDisabledAt.IsZero() {
			req.AccessDisabledAt = existing.AccessDisabledAt
			req.AccessDisabledBy = existing.AccessDisabledBy
		} else {
			req.AccessDisabledAt = now
			req.AccessDisabledBy = authUsername
		}
		req.AccessEnabledAt = time.Time{}
		req.AccessEnabledBy = ""
		return nil
	}

	if existing.IsAccessDisabled {
		req.AccessEnabledAt = now
		req.AccessEnabledBy = authUsername
	} else {
		req.AccessEnabledAt = existing.AccessEnabledAt
		req.AccessEnabledBy = existing.AccessEnabledBy
	}
	req.AccessDisabledAt = time.Time{}
	req.AccessDisabledBy = ""
	return nil
}

func (svc ShopUserService) InfoShopByUser(holdingCode string, username string) (models.ShopUserProfile, error) {

	shopUserProfile := models.ShopUserProfile{}

	// ดึงข้อมูล ShopUser เพื่อให้ได้ fields ใหม่ด้วย (position, department, LINE, approval)
	// The settings screen addresses members by useruid (Google-only users have no usercode);
	// resolveShopUser falls back to the uid because the PostgreSQL repo returns ErrShopUserNotFound.
	shopUser, err := svc.resolveShopUser(holdingCode, username)
	if err != nil {
		return models.ShopUserProfile{}, err
	}
	if shopUser.Username != "" {
		username = shopUser.Username
	}

	if strings.TrimSpace(username) != "" {
		userProfiles, profileErr := svc.repo.FindUserProfileByUsernames(context.Background(), []string{username})
		if profileErr != nil {
			return models.ShopUserProfile{}, profileErr
		}

		// Profile name
		if len(userProfiles) > 0 {
			shopUserProfile.UID = userProfiles[0].UID
			shopUserProfile.Email = userProfiles[0].Email
			shopUserProfile.UserProfileName = userProfiles[0].Name
		}
	}

	// Basic info
	shopUserProfile.HoldingCode = shopUser.HoldingCode
	shopUserProfile.Username = shopUser.Username
	shopUserProfile.UserUID = shopUser.UserUID
	shopUserProfile.Role = shopUser.Role
	shopUserProfile.IsAccessDisabled = shopUser.IsAccessDisabled
	shopUserProfile.AccessDisabledAt = shopUser.AccessDisabledAt
	shopUserProfile.AccessDisabledBy = shopUser.AccessDisabledBy
	shopUserProfile.AccessEnabledAt = shopUser.AccessEnabledAt
	shopUserProfile.AccessEnabledBy = shopUser.AccessEnabledBy

	createdBy, err := svc.repo.FindShopCreatedBy(context.Background(), holdingCode)
	if err != nil {
		return models.ShopUserProfile{}, err
	}
	shopUserProfile.IsCreator = sameUsername(shopUser.Username, createdBy) || sameUsername(shopUser.UserUID, createdBy)
	if shopUserProfile.IsCreator {
		shopUserProfile.IsAccessDisabled = false
	}

	// === ข้อมูลพนักงาน (ต่อ membership) ===
	shopUserProfile.Position = shopUser.Position
	shopUserProfile.Department = shopUser.Department
	shopUserProfile.Avatar = shopUser.Avatar
	shopUserProfile.AvatarThumb = shopUser.AvatarThumb
	shopUserProfile.AccessExpiryDate = models.AccessExpiryDay(shopUser.AccessExpiryDate)

	// === ข้อมูล LINE OA ===
	shopUserProfile.LineUserID = shopUser.LineUserID
	shopUserProfile.LineDisplayName = shopUser.LineDisplayName
	shopUserProfile.LinePictureURL = shopUser.LinePictureURL

	// === ข้อมูลการอนุมัติ ===
	shopUserProfile.POApproval = shopUser.POApproval
	shopUserProfile.QuotationApproval = shopUser.QuotationApproval
	shopUserProfile.AccessScopes = shopUser.AccessScopes
	shopUserProfile.PermissionSets = shopUser.PermissionSets

	return shopUserProfile, err
}

// resolveShopUser finds a shop user by username first, then by useruid, so
// usercode-less users (Google-only) can be addressed by their stable uid.
// Callers may pass a lowercased uid (NormalizeUsername) — both cases are tried.
// An empty id resolves to nothing on purpose: only users WITH a username may
// be matched through the username query.
func (svc ShopUserService) resolveShopUser(holdingCode string, id string) (models.ShopUser, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return models.ShopUser{}, errors.New("user not found")
	}
	shopUser, err := svc.repo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, id)
	if err == nil && shopUser.Username != "" {
		return shopUser, nil
	}
	if resolved, uidErr := svc.repo.FindByHoldingCodeAndUserUID(context.Background(), holdingCode, id); uidErr == nil {
		return resolved, nil
	}
	if lowered := strings.ToLower(id); lowered != id {
		resolved, uidErr := svc.repo.FindByHoldingCodeAndUserUID(context.Background(), holdingCode, lowered)
		if uidErr == nil {
			return resolved, nil
		}
	}
	return models.ShopUser{}, errors.New("user not found")
}

func (svc ShopUserService) ListShopByUser(authUsername string, authUserUID string, pageable micromodels.Pageable) ([]models.ShopUserInfo, common.PaginationData, error) {

	var docList []models.ShopUserInfo
	var pagination common.PaginationData
	var err error
	authUserUID = strings.TrimSpace(authUserUID)
	if authUserUID != "" {
		docList, pagination, err = svc.repo.FindByUserUIDPage(context.Background(), authUserUID, pageable)
	} else {
		docList, pagination, err = svc.repo.FindByUsernamePage(context.Background(), authUsername, pageable)
	}

	if err != nil {
		return docList, pagination, err
	}

	for idx := range docList {
		docList[idx].IsCreator = authUserUID != "" && strings.TrimSpace(docList[idx].CreatedBy) == authUserUID
	}

	return docList, pagination, err
}

func (svc ShopUserService) ListUserInShop(holdingCode string, authUsername string, pageable micromodels.Pageable) ([]models.ShopUserProfile, common.PaginationData, error) {
	shopUserProfiles := []models.ShopUserProfile{}

	// Authorize by the caller's role IN THIS holding (shopusers), not the JWT-selected role,
	// so an owner/admin of this holding can always list its users (same pattern as holding-member).
	if _, err := svc.requireHoldingManager(holdingCode, authUsername); err != nil {
		return shopUserProfiles, common.PaginationData{}, err
	}

	profileMatchedUsernames := []string{}
	if strings.TrimSpace(pageable.Query) != "" {
		var profileErr error
		profileMatchedUsernames, profileErr = svc.repo.FindUsernamesByProfileQuery(context.Background(), pageable.Query)
		if profileErr != nil {
			return shopUserProfiles, common.PaginationData{}, profileErr
		}
	}

	shopusers, pagination, err := svc.repo.FindByUserInShopPageWithProfileMatches(context.Background(), holdingCode, pageable, profileMatchedUsernames)

	if err != nil {
		return shopUserProfiles, pagination, err
	}

	usernames := []string{}

	for _, doc := range shopusers {
		usernames = append(usernames, doc.Username)
	}

	userProfiles, err := svc.repo.FindUserProfileByUsernames(context.Background(), usernames)

	if err != nil {
		return shopUserProfiles, pagination, err
	}

	createdBy, err := svc.repo.FindShopCreatedBy(context.Background(), holdingCode)
	if err != nil {
		return shopUserProfiles, pagination, err
	}

	dictUserProfiles := map[string]models.UserProfile{}
	for _, doc := range userProfiles {
		dictUserProfiles[doc.Username] = doc
	}

	for _, doc := range shopusers {
		shopUserProfile := models.ShopUserProfile{}

		shopUserProfile.HoldingCode = doc.HoldingCode
		shopUserProfile.Username = doc.Username
		shopUserProfile.UserUID = doc.UserUID
		shopUserProfile.Role = doc.Role
		shopUserProfile.Position = doc.Position
		shopUserProfile.Department = doc.Department
		shopUserProfile.Avatar = doc.Avatar
		shopUserProfile.AvatarThumb = doc.AvatarThumb
		shopUserProfile.AccessExpiryDate = models.AccessExpiryDay(doc.AccessExpiryDate)
		shopUserProfile.POApproval = doc.POApproval
		shopUserProfile.QuotationApproval = doc.QuotationApproval
		shopUserProfile.AccessScopes = doc.AccessScopes
		shopUserProfile.PermissionSets = doc.PermissionSets
		shopUserProfile.IsAccessDisabled = doc.IsAccessDisabled
		shopUserProfile.AccessDisabledAt = doc.AccessDisabledAt
		shopUserProfile.AccessDisabledBy = doc.AccessDisabledBy
		shopUserProfile.AccessEnabledAt = doc.AccessEnabledAt
		shopUserProfile.AccessEnabledBy = doc.AccessEnabledBy
		shopUserProfile.IsCreator = sameUsername(doc.Username, createdBy)
		if shopUserProfile.IsCreator {
			shopUserProfile.IsAccessDisabled = false
		}

		// LINE data
		shopUserProfile.LineUserID = doc.LineUserID
		shopUserProfile.LineDisplayName = doc.LineDisplayName
		shopUserProfile.LinePictureURL = doc.LinePictureURL

		if tempUserProfile, ok := dictUserProfiles[doc.Username]; ok {
			shopUserProfile.UID = tempUserProfile.UID
			shopUserProfile.Email = tempUserProfile.Email
			shopUserProfile.UserProfileName = tempUserProfile.Name
		}

		shopUserProfiles = append(shopUserProfiles, shopUserProfile)
	}

	return shopUserProfiles, pagination, err
}

// SaveUserFullProfile - บันทึกข้อมูลผู้ใช้แบบครบถ้วน (รวม position, department, LINE, approval)
//
// A request without editusername and without useruid ADDS a member: naming someone who is
// already a member is rejected (errMemberAlreadyExists) instead of silently overwriting
// their role and permissions. With either id it EDITS that member and never creates one.
func (svc ShopUserService) SaveUserFullProfile(holdingCode string, authUsername string, req *models.UserRoleRequest) error {
	rawEditID := strings.TrimSpace(req.EditUsername)
	adding := rawEditID == "" && strings.TrimSpace(req.UserUID) == ""
	username := utils.NormalizeUsername(req.Username)
	editusername := utils.NormalizeUsername(req.EditUsername)
	req.EditUsername = editusername

	authUser, err := svc.requireHoldingManager(holdingCode, authUsername)
	if err != nil {
		return err
	}

	createdBy, err := svc.repo.FindShopCreatedBy(context.Background(), holdingCode)
	if err != nil {
		return err
	}

	// Resolve with the RAW edit id first: a useruid keeps its original casing and
	// NormalizeUsername would lowercase it into a lookup miss.
	existingTarget := svc.findSaveTarget(holdingCode, rawEditID, username, req.UserUID)
	targetFound := existingTarget.UserUID != "" || strings.TrimSpace(existingTarget.Username) != ""
	if !adding && !targetFound {
		return errSaveTargetNotFound
	}
	if existingTarget.UserUID != "" {
		req.UserUID = existingTarget.UserUID
	}

	// Self edit: allowed only when the role is unchanged (scope/permission
	// self-service, e.g. an owner granting their own company access). Changing
	// your own role stays forbidden to prevent lock-out — a Holding must always
	// keep a working OWNER (docs/organization.md).
	targetIsCreator := sameUsername(existingTarget.Username, createdBy) ||
		(existingTarget.UserUID != "" && existingTarget.UserUID == createdBy)
	isSelf := sameUsername(authUsername, username) || sameUsername(authUsername, editusername) ||
		(existingTarget.UserUID != "" && existingTarget.UserUID == authUser.UserUID)
	if isSelf && req.Role != existingTarget.Role {
		return errors.New("can not edit self permission")
	}

	// Admins have full access but must not be able to modify an owner or promote anyone to owner.
	if authUser.Role == models.ROLE_ADMIN && (existingTarget.Role == models.ROLE_OWNER || req.Role == models.ROLE_OWNER) {
		return errors.New("permission denied")
	}
	if err = checkScopeGrant(authUser, req); err != nil {
		return err
	}
	if err = checkTargetWithinGrantor(authUser, existingTarget); err != nil {
		return err
	}

	if err = applyAccessStatus(req, existingTarget, authUsername, time.Now().UTC(), targetIsCreator); err != nil {
		return err
	}
	// Checked after the permission checks so a denied caller learns nothing about members.
	if adding && targetFound {
		return errMemberAlreadyExists
	}

	// Normalize username in request
	req.Username = username

	// บันทึกข้อมูลแบบ full profile
	err = svc.repo.SaveFullProfile(context.Background(), holdingCode, req)
	if err != nil {
		return err
	}

	return nil
}

// findSaveTarget resolves the existing member a save would write. The repository upserts by
// useruid, else by username, so the edit id, the username and the client useruid are all
// tried: a bogus editusername must not skip the owner/creator/scope checks of the member
// whose row the upsert would overwrite. A zero ShopUser means a new member.
func (svc ShopUserService) findSaveTarget(holdingCode string, ids ...string) models.ShopUser {
	tried := map[string]bool{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" || tried[id] {
			continue
		}
		tried[id] = true
		if member, err := svc.resolveShopUser(holdingCode, id); err == nil {
			return member
		}
	}
	return models.ShopUser{}
}

// checkTargetWithinGrantor stops a manager from changing or removing a member they could not
// have granted: the member's CURRENT role and scope must fit inside the manager's own scope.
// checkScopeGrant alone only vets the requested scope, so a company-restricted ADMIN could
// otherwise narrow, demote or strip a holding-wide ADMIN, or move another company's staff
// into their own company. A new member (zero ShopUser) has nothing to protect.
func checkTargetWithinGrantor(grantor models.ShopUser, target models.ShopUser) error {
	if target.UserUID == "" && strings.TrimSpace(target.Username) == "" {
		return nil
	}
	return checkScopeGrant(grantor, &models.UserRoleRequest{Role: target.Role, AccessScopes: target.AccessScopes})
}

// checkScopeGrant stops a manager from granting more than they can reach themselves
// (also when editing their own membership). OWNER may grant anything; an ADMIN grants
// the holding-wide scope only when holding-wide themselves, otherwise only companies
// and branches inside their own scope. Scopes must already be hydrated.
func checkScopeGrant(grantor models.ShopUser, req *models.UserRoleRequest) error {
	if grantor.Role == models.ROLE_OWNER || models.HasHoldingScope(grantor.AccessScopes) {
		return nil
	}
	requested := req.AccessScopes
	// OWNER/ADMIN without explicit scopes is holding-wide (organization/access.FindActiveMembership).
	managerTarget := req.Role == models.ROLE_OWNER || req.Role == models.ROLE_ADMIN
	if models.HasHoldingScope(requested) || (len(requested) == 0 && managerTarget) {
		return errAccessScopeExceedsGrantor
	}
	for _, scope := range requested {
		companyUID := strings.TrimSpace(scope.CompanyUID)
		allowed := false
		switch accessScopeType(scope) {
		case "company":
			allowed = models.ScopesAllowCompanySelection(grantor.AccessScopes, companyUID) &&
				(!scope.AllBranches || orgpolicy.AllowsAllBranches(grantor.AccessScopes, companyUID))
		case "branch":
			allowed = models.ScopesAllowBranchSelection(grantor.AccessScopes, companyUID, scope.BranchUID)
		}
		if !allowed {
			return errAccessScopeExceedsGrantor
		}
	}
	return nil
}

func (svc ShopUserService) DeleteUserPermissionShop(holdingCode string, authUsername string, username string) error {

	authUser, err := svc.requireHoldingManager(holdingCode, authUsername)
	if err != nil {
		return err
	}

	findUser, err := svc.resolveShopUser(holdingCode, username)

	if err != nil {
		return err
	}

	createdBy, err := svc.repo.FindShopCreatedBy(context.Background(), holdingCode)
	if err != nil {
		return err
	}

	if sameUsername(findUser.Username, createdBy) || (findUser.UserUID != "" && findUser.UserUID == createdBy) {
		return errors.New("creator_cannot_delete")
	}

	if authUser.Role == models.ROLE_ADMIN && findUser.Role == models.ROLE_OWNER {
		return errors.New("permission denied")
	}

	if sameUsername(findUser.Username, authUsername) || (findUser.UserUID != "" && findUser.UserUID == authUser.UserUID) {
		return errors.New("can't delete your permission")
	}
	if err = checkTargetWithinGrantor(authUser, findUser); err != nil {
		return err
	}

	// Delete the member that was checked above, not whatever else the raw id might match.
	deleteID := username
	if findUser.UserUID != "" {
		deleteID = findUser.UserUID
	}
	err = svc.repo.Delete(context.Background(), holdingCode, deleteID, authUser.UserUID)

	if err != nil {
		return err
	}
	return nil
}

// requireHoldingManager resolves the caller's role IN THE GIVEN holding (not the JWT-selected
// one) and requires OWNER or ADMIN. This lets admins be managed from the holding-selection
// screen for any holding the caller owns/administers.
func (svc ShopUserService) requireHoldingManager(holdingCode string, authUsername string) (models.ShopUser, error) {
	authUser, err := svc.repo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, utils.NormalizeUsername(authUsername))
	if err != nil {
		return models.ShopUser{}, err
	}
	if authUser.Role != models.ROLE_OWNER && authUser.Role != models.ROLE_ADMIN {
		return models.ShopUser{}, errors.New("permission denied")
	}
	if authUser.IsAccessDisabled || models.AccessExpired(authUser.AccessExpiryDate, time.Now()) {
		return models.ShopUser{}, errors.New("permission denied")
	}
	return authUser, nil
}

// EnsureHoldingManager exposes the per-holding OWNER/ADMIN check to handlers (e.g. bulk import)
// that need to gate an operation once before running a batch.
func (svc ShopUserService) EnsureHoldingManager(holdingCode string, authUsername string) error {
	_, err := svc.requireHoldingManager(holdingCode, authUsername)
	return err
}

// ListHoldingMembersByAdmin lists members of a holding for an owner/admin of that holding.
func (svc ShopUserService) ListHoldingMembersByAdmin(holdingCode string, authUsername string, pageable micromodels.Pageable) ([]models.ShopUserProfile, common.PaginationData, error) {
	// ListUserInShop now enforces requireHoldingManager itself (per-holding role).
	return svc.ListUserInShop(holdingCode, authUsername, pageable)
}
