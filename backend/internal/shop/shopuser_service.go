package shop

import (
	"context"
	"errors"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IShopUserService interface {
	SaveUserPermissionShop(holdingCode string, authUsername string, editusername string, username string, role models.UserRole) error
	SaveUserFullProfile(holdingCode string, authUsername string, req *models.UserRoleRequest) error
	DeleteUserPermissionShop(holdingCode string, authUsername string, username string) error
	CleanupEmptyUsers(holdingCode string) (int64, error)

	// Holding admin management by email (holdingCode comes from the request, role of the
	// caller is resolved per-holding so it works from the holding-selection screen).
	AddHoldingAdminByEmail(holdingCode string, authUsername string, targetEmail string) error
	RemoveHoldingMember(holdingCode string, authUsername string, targetEmail string) error
	ListHoldingMembersByAdmin(holdingCode string, authUsername string, pageable micromodels.Pageable) ([]models.ShopUserProfile, mongopagination.PaginationData, error)
	EnsureHoldingManager(holdingCode string, authUsername string) error

	InfoShopByUser(holdingCode string, username string) (models.ShopUserProfile, error)
	ListShopByUser(authUsername string, authUserUID string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error)
	ListUserInShop(holdingCode string, authUsername string, pageable micromodels.Pageable) ([]models.ShopUserProfile, mongopagination.PaginationData, error)

	// SyncLineData - sync LINE data จาก LIFF (ใช้สำหรับ callback จาก lineoa-liff)
	SyncLineData(holdingCode string, username string, lineUserID string, lineDisplayName string, linePictureURL string) error

	// SaveMyLineData - ให้ผู้ใช้อัปเดต LINE data ของตัวเอง
	SaveMyLineData(holdingCode string, username string, lineUserID string, lineDisplayName string, linePictureURL string) error
}

type ShopUserService struct {
	repo IShopUserRepository
}

func NewShopUserService(shopUserRepo IShopUserRepository) ShopUserService {
	return ShopUserService{
		repo: shopUserRepo,
	}
}

func sameUsername(left string, right string) bool {
	return strings.EqualFold(utils.NormalizeUsername(left), utils.NormalizeUsername(right))
}

func copyAccessStatusToRequest(req *models.UserRoleRequest, user models.ShopUser) {
	req.IsAccessDisabled = user.IsAccessDisabled
	req.AccessDisabledAt = user.AccessDisabledAt
	req.AccessDisabledBy = user.AccessDisabledBy
	req.AccessEnabledAt = user.AccessEnabledAt
	req.AccessEnabledBy = user.AccessEnabledBy
	req.AccessExpiryDate = user.AccessExpiryDate
}

func applyAccessStatus(req *models.UserRoleRequest, existing models.ShopUser, authUsername string, now time.Time, isCreator bool) error {
	if isCreator {
		if req.IsAccessDisabled {
			return errors.New("creator_access_cannot_be_disabled")
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
	shopUser, err := svc.repo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, username)
	if err != nil {
		return models.ShopUserProfile{}, err
	}
	if shopUser.Username == "" {
		// Users without a usercode (Google-only) are addressed by their stable useruid
		// because their shopuser row keeps an empty username by design.
		resolved, uidErr := svc.repo.FindByHoldingCodeAndUserUID(context.Background(), holdingCode, username)
		if uidErr != nil {
			return models.ShopUserProfile{}, errors.New("user not found")
		}
		shopUser = resolved
		err = nil
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
			shopUserProfile.Avatar = userProfiles[0].Avatar
			shopUserProfile.AvatarThumb = userProfiles[0].AvatarThumb
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

	// === ข้อมูลพนักงาน ===
	shopUserProfile.Position = shopUser.Position
	shopUserProfile.Department = shopUser.Department

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

func (svc ShopUserService) ListShopByUser(authUsername string, authUserUID string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error) {

	var docList []models.ShopUserInfo
	var pagination mongopagination.PaginationData
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

func (svc ShopUserService) ListUserInShop(holdingCode string, authUsername string, pageable micromodels.Pageable) ([]models.ShopUserProfile, mongopagination.PaginationData, error) {
	shopUserProfiles := []models.ShopUserProfile{}

	// Authorize by the caller's role IN THIS holding (shopusers), not the JWT-selected role,
	// so an owner/admin of this holding can always list its users (same pattern as holding-member).
	if _, err := svc.requireHoldingManager(holdingCode, authUsername); err != nil {
		return shopUserProfiles, mongopagination.PaginationData{}, err
	}

	profileMatchedUsernames := []string{}
	if strings.TrimSpace(pageable.Query) != "" {
		var profileErr error
		profileMatchedUsernames, profileErr = svc.repo.FindUsernamesByProfileQuery(context.Background(), pageable.Query)
		if profileErr != nil {
			return shopUserProfiles, mongopagination.PaginationData{}, profileErr
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
			shopUserProfile.Avatar = tempUserProfile.Avatar
			shopUserProfile.AvatarThumb = tempUserProfile.AvatarThumb
		}

		shopUserProfiles = append(shopUserProfiles, shopUserProfile)
	}

	return shopUserProfiles, pagination, err
}

func (svc ShopUserService) SaveUserPermissionShop(holdingCode string, authUsername string, editusername string, username string, role models.UserRole) error {

	username = utils.NormalizeUsername(username)

	authUser, err := svc.repo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, authUsername)

	if err != nil {
		return err
	}

	if authUser.Role != models.ROLE_OWNER && authUser.Role != models.ROLE_ADMIN {
		return errors.New("permission denied")
	}

	// Resolve the target with the RAW id first: a useruid keeps its original
	// casing and NormalizeUsername would lowercase it into a lookup miss.
	findEditUser, err := svc.resolveShopUser(holdingCode, editusername)

	if err != nil {
		return err
	}

	editusername = utils.NormalizeUsername(editusername)

	// Self edit: allowed only when the role is unchanged (see SaveUserFullProfile).
	isSelf := sameUsername(authUsername, username) || sameUsername(authUsername, editusername) ||
		(findEditUser.UserUID != "" && findEditUser.UserUID == authUser.UserUID)
	if isSelf && role != findEditUser.Role {
		return errors.New("can not edit self permission")
	}

	if findEditUser.Username != "" || findEditUser.UserUID != "" {
		tempID := findEditUser.ID

		err = svc.repo.Update(context.Background(), tempID, holdingCode, username, role)
		if err != nil {
			return err
		}
	} else {
		err = svc.repo.Save(context.Background(), holdingCode, username, role)

		if err != nil {
			return err
		}
	}
	return nil
}

func (svc ShopUserService) create(ctx context.Context, holdingCode string, username string, role models.UserRole) error {

	tempShopUser := models.ShopUser{}
	tempShopUser.HoldingCode = holdingCode
	tempShopUser.Username = username
	tempShopUser.Role = role

	err := svc.repo.Create(ctx, &tempShopUser)
	if err != nil {
		return err
	}
	return nil
}

// SaveUserFullProfile - บันทึกข้อมูลผู้ใช้แบบครบถ้วน (รวม position, department, LINE, approval)
func (svc ShopUserService) SaveUserFullProfile(holdingCode string, authUsername string, req *models.UserRoleRequest) error {
	rawEditID := strings.TrimSpace(req.EditUsername)
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

	lookupUsername := rawEditID
	if lookupUsername == "" {
		lookupUsername = username
	}
	existingTarget := models.ShopUser{}
	if lookupUsername != "" {
		// Resolve with the RAW id: a useruid keeps its original casing and
		// NormalizeUsername would lowercase it into a lookup miss.
		existingTarget, err = svc.resolveShopUser(holdingCode, lookupUsername)
		if err != nil {
			existingTarget = models.ShopUser{}
		}
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

	if err = applyAccessStatus(req, existingTarget, authUsername, time.Now().UTC(), targetIsCreator); err != nil {
		return err
	}

	// ตรวจสอบว่า LINE ID ไม่ซ้ำกับผู้ใช้คนอื่น - ถ้าซ้ำให้ auto-unlink คนเก่า
	if req.LineUserID != "" {
		existingUser, existingUserErr := svc.repo.FindByHoldingCodeAndLineUserID(context.Background(), holdingCode, req.LineUserID)
		// ถ้าพบผู้ใช้ที่ใช้ LINE ID นี้แล้ว และไม่ใช่ผู้ใช้คนเดียวกัน ให้ลบ LINE data ของคนเก่า (auto-unlink)
		if existingUserErr == nil && existingUser.Username != "" && !sameUsername(existingUser.Username, username) {
			// ลบ LINE data ของ user เก่า
			oldReq := &models.UserRoleRequest{
				Username:          existingUser.Username,
				Role:              existingUser.Role,
				Position:          existingUser.Position,
				Department:        existingUser.Department,
				LineUserID:        "", // clear LINE data
				LineDisplayName:   "",
				LinePictureURL:    "",
				POApproval:        existingUser.POApproval,
				QuotationApproval: existingUser.QuotationApproval,
				AccessScopes:      existingUser.AccessScopes,
				PermissionSets:    existingUser.PermissionSets,
			}
			copyAccessStatusToRequest(oldReq, existingUser)
			svc.repo.SaveFullProfile(context.Background(), holdingCode, oldReq)
		}
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

	err = svc.repo.Delete(context.Background(), holdingCode, username)

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
	if authUser.IsAccessDisabled || (!authUser.AccessExpiryDate.IsZero() && time.Now().After(authUser.AccessExpiryDate)) {
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

// AddHoldingAdminByEmail grants ADMIN access to a holding by email. The email is stored as the
// shopuser username (normalized lowercase) so it binds when that person later logs in by the
// same email (Google/password) — no prior login required. Idempotent (safe to call repeatedly).
func (svc ShopUserService) AddHoldingAdminByEmail(holdingCode string, authUsername string, targetEmail string) error {
	if _, err := svc.requireHoldingManager(holdingCode, authUsername); err != nil {
		return err
	}

	target := utils.NormalizeUsername(targetEmail)
	if target == "" {
		return errors.New("email is required")
	}

	existing, existingErr := svc.repo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, target)
	if existingErr == nil && existing.Role == models.ROLE_OWNER {
		// Never downgrade or re-grant an owner through the admin tool.
		return errors.New("cannot change owner")
	}

	req := &models.UserRoleRequest{
		Username: target,
		Email:    target,
		Role:     models.ROLE_ADMIN,
	}
	if existingErr == nil && existing.UserUID != "" {
		req.UserUID = existing.UserUID
	}

	return svc.repo.SaveFullProfile(context.Background(), holdingCode, req)
}

// RemoveHoldingMember removes a member from a holding. Owners (and the shop creator) are
// protected and cannot be removed; callers cannot remove themselves.
func (svc ShopUserService) RemoveHoldingMember(holdingCode string, authUsername string, targetEmail string) error {
	if _, err := svc.requireHoldingManager(holdingCode, authUsername); err != nil {
		return err
	}

	target := utils.NormalizeUsername(targetEmail)
	if target == "" {
		return errors.New("email is required")
	}
	if sameUsername(target, authUsername) {
		return errors.New("can't remove yourself")
	}

	findUser, err := svc.repo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, target)
	if err != nil {
		return err
	}
	if findUser.Username == "" {
		return errors.New("user not found")
	}
	if findUser.Role == models.ROLE_OWNER {
		return errors.New("cannot remove owner")
	}

	createdBy, createdByErr := svc.repo.FindShopCreatedBy(context.Background(), holdingCode)
	if createdByErr == nil && sameUsername(findUser.Username, createdBy) {
		return errors.New("cannot remove creator")
	}

	return svc.repo.Delete(context.Background(), holdingCode, target)
}

// ListHoldingMembersByAdmin lists members of a holding for an owner/admin of that holding.
func (svc ShopUserService) ListHoldingMembersByAdmin(holdingCode string, authUsername string, pageable micromodels.Pageable) ([]models.ShopUserProfile, mongopagination.PaginationData, error) {
	// ListUserInShop now enforces requireHoldingManager itself (per-holding role).
	return svc.ListUserInShop(holdingCode, authUsername, pageable)
}

// CleanupEmptyUsers - ลบ users ที่ username ว่างออกจาก shop
func (svc ShopUserService) CleanupEmptyUsers(holdingCode string) (int64, error) {
	return svc.repo.DeleteEmptyUsernames(context.Background(), holdingCode)
}

// SyncLineData - sync LINE data จาก LIFF callback (ใช้สำหรับ callback จาก lineoa-liff หลัง linking สำเร็จ)
func (svc ShopUserService) SyncLineData(holdingCode string, username string, lineUserID string, lineDisplayName string, linePictureURL string) error {
	return svc.updateLineFields(holdingCode, username, lineUserID, lineDisplayName, linePictureURL)
}

// SaveMyLineData - ให้ผู้ใช้อัปเดต LINE data ของตัวเอง (ใช้จาก Flutter หลัง LIFF linking)
func (svc ShopUserService) SaveMyLineData(holdingCode string, username string, lineUserID string, lineDisplayName string, linePictureURL string) error {
	return svc.updateLineFields(holdingCode, username, lineUserID, lineDisplayName, linePictureURL)
}

func (svc ShopUserService) updateLineFields(holdingCode string, username string, lineUserID string, lineDisplayName string, linePictureURL string) error {
	username = utils.NormalizeUsername(username)

	if username == "" {
		return errors.New("username is required")
	}

	// ตรวจสอบว่ามี user นี้ใน shop หรือไม่
	existingUser, err := svc.repo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, username)
	if err != nil {
		return errors.New("user not found in shop")
	}

	if existingUser.Username == "" {
		return errors.New("user not found in shop")
	}

	// ถ้า LINE ID ถูกใช้โดยผู้ใช้คนอื่นแล้ว ให้ลบ LINE data ของคนนั้นก่อน (auto-unlink)
	if lineUserID != "" {
		existingLineUser, existingLineUserErr := svc.repo.FindByHoldingCodeAndLineUserID(context.Background(), holdingCode, lineUserID)
		if existingLineUserErr == nil && existingLineUser.Username != "" && !sameUsername(existingLineUser.Username, username) {
			if err := svc.repo.UpdateLineFields(context.Background(), holdingCode, existingLineUser.UserUID, "", "", ""); err != nil {
				return err
			}
		}
	}

	return svc.repo.UpdateLineFields(context.Background(), holdingCode, existingUser.UserUID, lineUserID, lineDisplayName, linePictureURL)
}
