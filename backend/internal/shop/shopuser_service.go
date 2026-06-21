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

	InfoShopByUser(holdingCode string, username string) (models.ShopUserProfile, error)
	ListShopByUser(authUsername string, authUserUID string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error)
	ListUserInShop(holdingCode string, pageable micromodels.Pageable) ([]models.ShopUserProfile, mongopagination.PaginationData, error)

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
		return models.ShopUserProfile{}, errors.New("user not found")
	}

	userProfiles, err := svc.repo.FindUserProfileByUsernames(context.Background(), []string{username})
	if err != nil {
		return models.ShopUserProfile{}, err
	}

	// Basic info
	shopUserProfile.HoldingCode = shopUser.HoldingCode
	shopUserProfile.Username = username
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
	shopUserProfile.IsCreator = sameUsername(username, createdBy)
	if shopUserProfile.IsCreator {
		shopUserProfile.IsAccessDisabled = false
	}

	// Profile name
	if len(userProfiles) > 0 {
		shopUserProfile.UID = userProfiles[0].UID
		shopUserProfile.Email = userProfiles[0].Email
		shopUserProfile.UserProfileName = userProfiles[0].Name
		shopUserProfile.Avatar = userProfiles[0].Avatar
		shopUserProfile.AvatarThumb = userProfiles[0].AvatarThumb
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

	return shopUserProfile, err
}

func (svc ShopUserService) ListShopByUser(authUsername string, authUserUID string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error) {

	var docList []models.ShopUserInfo
	var pagination mongopagination.PaginationData
	var err error
	if strings.TrimSpace(authUserUID) != "" {
		docList, pagination, err = svc.repo.FindByUserUIDPage(context.Background(), authUserUID, pageable)
	} else {
		docList, pagination, err = svc.repo.FindByUsernamePage(context.Background(), authUsername, pageable)
	}

	if err != nil {
		return docList, pagination, err
	}

	for idx := range docList {
		docList[idx].IsCreator = sameUsername(authUsername, docList[idx].CreatedBy)
		if docList[idx].IsCreator {
			docList[idx].IsAccessDisabled = false
		}
	}

	return docList, pagination, err
}

func (svc ShopUserService) ListUserInShop(holdingCode string, pageable micromodels.Pageable) ([]models.ShopUserProfile, mongopagination.PaginationData, error) {
	shopUserProfiles := []models.ShopUserProfile{}

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

	if authUsername == username || authUsername == editusername {
		return errors.New("can not edit self permission")
	}

	authUser, err := svc.repo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, authUsername)

	if err != nil {
		return err
	}

	if authUser.Role != models.ROLE_OWNER {
		return errors.New("permission denied")
	}

	editusername = utils.NormalizeUsername(editusername)

	if editusername != "" {

		findEditUser, err := svc.repo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, editusername)

		if err != nil {
			return err
		}

		if findEditUser.Username != "" {
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
	username := utils.NormalizeUsername(req.Username)
	editusername := utils.NormalizeUsername(req.EditUsername)
	req.EditUsername = editusername

	if sameUsername(authUsername, username) || sameUsername(authUsername, editusername) {
		return errors.New("can not edit self permission")
	}

	authUser, err := svc.repo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, authUsername)
	if err != nil {
		return err
	}

	if authUser.Role != models.ROLE_OWNER {
		return errors.New("permission denied")
	}

	createdBy, err := svc.repo.FindShopCreatedBy(context.Background(), holdingCode)
	if err != nil {
		return err
	}

	lookupUsername := username
	if editusername != "" {
		lookupUsername = editusername
	}
	existingTarget, existingTargetErr := svc.repo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, lookupUsername)
	if existingTargetErr != nil {
		existingTarget = models.ShopUser{}
	}
	if existingTarget.UserUID != "" {
		req.UserUID = existingTarget.UserUID
	}

	if err = applyAccessStatus(req, existingTarget, authUsername, time.Now().UTC(), sameUsername(username, createdBy)); err != nil {
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

	authUser, err := svc.repo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, authUsername)

	if err != nil {
		return err
	}

	if authUser.Role != models.ROLE_OWNER && authUser.Role != models.ROLE_ADMIN {
		return errors.New("permission denied")
	}

	findUser, err := svc.repo.FindByHoldingCodeAndUsername(context.Background(), holdingCode, username)

	if err != nil {
		return err
	}

	// ตรวจสอบว่าพบผู้ใช้ที่ต้องการลบหรือไม่
	if findUser.Username == "" {
		return errors.New("user not found")
	}

	createdBy, err := svc.repo.FindShopCreatedBy(context.Background(), holdingCode)
	if err != nil {
		return err
	}

	if sameUsername(findUser.Username, createdBy) {
		return errors.New("creator_cannot_delete")
	}

	if authUser.Role == models.ROLE_ADMIN && findUser.Role == models.ROLE_OWNER {
		return errors.New("permission denied")
	}

	if findUser.Username == authUsername {
		return errors.New("can't delete your permission")
	}

	err = svc.repo.Delete(context.Background(), holdingCode, username)

	if err != nil {
		return err
	}
	return nil
}

// CleanupEmptyUsers - ลบ users ที่ username ว่างออกจาก shop
func (svc ShopUserService) CleanupEmptyUsers(holdingCode string) (int64, error) {
	return svc.repo.DeleteEmptyUsernames(context.Background(), holdingCode)
}

// SyncLineData - sync LINE data จาก LIFF callback (ใช้สำหรับ callback จาก lineoa-liff หลัง linking สำเร็จ)
func (svc ShopUserService) SyncLineData(holdingCode string, username string, lineUserID string, lineDisplayName string, linePictureURL string) error {
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
			// ลบ LINE data ของ user เก่า
			oldReq := &models.UserRoleRequest{
				Username:        existingLineUser.Username,
				Role:            existingLineUser.Role,
				Position:        existingLineUser.Position,
				Department:      existingLineUser.Department,
				LineUserID:      "", // clear LINE data
				LineDisplayName: "",
				LinePictureURL:  "",
				AccessScopes:    existingLineUser.AccessScopes,
			}
			copyAccessStatusToRequest(oldReq, existingLineUser)
			svc.repo.SaveFullProfile(context.Background(), holdingCode, oldReq)
		}
	}

	// สร้าง request สำหรับ update LINE data
	req := &models.UserRoleRequest{
		Username:          username,
		Role:              existingUser.Role, // keep existing role
		Position:          existingUser.Position,
		Department:        existingUser.Department,
		LineUserID:        lineUserID,
		LineDisplayName:   lineDisplayName,
		LinePictureURL:    linePictureURL,
		POApproval:        existingUser.POApproval,
		QuotationApproval: existingUser.QuotationApproval,
		AccessScopes:      existingUser.AccessScopes,
	}
	copyAccessStatusToRequest(req, existingUser)

	return svc.repo.SaveFullProfile(context.Background(), holdingCode, req)
}

// SaveMyLineData - ให้ผู้ใช้อัปเดต LINE data ของตัวเอง (ใช้จาก Flutter หลัง LIFF linking)
func (svc ShopUserService) SaveMyLineData(holdingCode string, username string, lineUserID string, lineDisplayName string, linePictureURL string) error {
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
			// ลบ LINE data ของ user เก่า
			oldReq := &models.UserRoleRequest{
				Username:        existingLineUser.Username,
				Role:            existingLineUser.Role,
				Position:        existingLineUser.Position,
				Department:      existingLineUser.Department,
				LineUserID:      "", // clear LINE data
				LineDisplayName: "",
				LinePictureURL:  "",
				AccessScopes:    existingLineUser.AccessScopes,
			}
			copyAccessStatusToRequest(oldReq, existingLineUser)
			svc.repo.SaveFullProfile(context.Background(), holdingCode, oldReq)
		}
	}

	// สร้าง request สำหรับ update LINE data (คงค่าอื่นๆ ไว้เหมือนเดิม)
	req := &models.UserRoleRequest{
		Username:          username,
		Role:              existingUser.Role,
		Position:          existingUser.Position,
		Department:        existingUser.Department,
		LineUserID:        lineUserID,
		LineDisplayName:   lineDisplayName,
		LinePictureURL:    linePictureURL,
		POApproval:        existingUser.POApproval,
		QuotationApproval: existingUser.QuotationApproval,
		AccessScopes:      existingUser.AccessScopes,
	}
	copyAccessStatusToRequest(req, existingUser)

	return svc.repo.SaveFullProfile(context.Background(), holdingCode, req)
}
