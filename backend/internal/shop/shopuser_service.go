package shop

import (
	"context"
	"errors"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
)

type IShopUserService interface {
	SaveUserPermissionShop(shopID string, authUsername string, editusername string, username string, role models.UserRole) error
	SaveUserFullProfile(shopID string, authUsername string, req *models.UserRoleRequest) error
	DeleteUserPermissionShop(shopID string, authUsername string, username string) error
	CleanupEmptyUsers(shopID string) (int64, error)

	InfoShopByUser(shopID string, username string) (models.ShopUserProfile, error)
	ListShopByUser(authUsername string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error)
	ListUserInShop(shopID string, pageable micromodels.Pageable) ([]models.ShopUserProfile, mongopagination.PaginationData, error)

	// SyncLineData - sync LINE data จาก LIFF (ใช้สำหรับ callback จาก lineoa-liff)
	SyncLineData(shopID string, username string, lineUserID string, lineDisplayName string, linePictureURL string) error

	// SaveMyLineData - ให้ผู้ใช้อัปเดต LINE data ของตัวเอง
	SaveMyLineData(shopID string, username string, lineUserID string, lineDisplayName string, linePictureURL string) error
}

type ShopUserService struct {
	repo IShopUserRepository
}

func NewShopUserService(shopUserRepo IShopUserRepository) ShopUserService {
	return ShopUserService{
		repo: shopUserRepo,
	}
}

func (svc ShopUserService) InfoShopByUser(shopID string, username string) (models.ShopUserProfile, error) {

	shopUserProfile := models.ShopUserProfile{}

	// ดึงข้อมูล ShopUser เพื่อให้ได้ fields ใหม่ด้วย (position, department, LINE, approval)
	shopUser, err := svc.repo.FindByShopIDAndUsername(context.Background(), shopID, username)
	if err != nil {
		return models.ShopUserProfile{}, err
	}

	userProfiles, err := svc.repo.FindUserProfileByUsernames(context.Background(), []string{username})
	if err != nil {
		return models.ShopUserProfile{}, err
	}

	// Basic info
	shopUserProfile.ShopID = shopUser.ShopID
	shopUserProfile.Username = username
	shopUserProfile.Role = shopUser.Role

	// Profile name
	if len(userProfiles) > 0 {
		shopUserProfile.UserProfileName = userProfiles[0].Name
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

	return shopUserProfile, err
}

func (svc ShopUserService) ListShopByUser(authUsername string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error) {

	docList, pagination, err := svc.repo.FindByUsernamePage(context.Background(), authUsername, pageable)

	if err != nil {
		return docList, pagination, err
	}

	return docList, pagination, err
}

func (svc ShopUserService) ListUserInShop(shopID string, pageable micromodels.Pageable) ([]models.ShopUserProfile, mongopagination.PaginationData, error) {
	shopUserProfiles := []models.ShopUserProfile{}

	shopUsers, pagination, err := svc.repo.FindByUserInShopPage(context.Background(), shopID, pageable)

	if err != nil {
		return shopUserProfiles, pagination, err
	}

	usernames := []string{}

	for _, doc := range shopUsers {
		usernames = append(usernames, doc.Username)
	}

	userProfiles, err := svc.repo.FindUserProfileByUsernames(context.Background(), usernames)

	if err != nil {
		return shopUserProfiles, pagination, err
	}

	for _, doc := range shopUsers {
		shopUserProfile := models.ShopUserProfile{}

		shopUserProfile.ShopID = doc.ShopID
		shopUserProfile.Username = doc.Username
		shopUserProfile.Role = doc.Role

		// LINE data
		shopUserProfile.LineUserID = doc.LineUserID
		shopUserProfile.LineDisplayName = doc.LineDisplayName
		shopUserProfile.LinePictureURL = doc.LinePictureURL

		shopUserProfiles = append(shopUserProfiles, shopUserProfile)
	}

	dictUserProfiles := map[string]models.UserProfile{}
	for _, doc := range userProfiles {
		dictUserProfiles[doc.Username] = doc
	}

	for idx, doc := range userProfiles {
		tempUserProfile := dictUserProfiles[doc.Username]

		shopUserProfiles[idx].UserProfileName = tempUserProfile.Name
	}

	return shopUserProfiles, pagination, err
}

func (svc ShopUserService) SaveUserPermissionShop(shopID string, authUsername string, editusername string, username string, role models.UserRole) error {

	username = utils.NormalizeUsername(username)

	if authUsername == username || authUsername == editusername {
		return errors.New("can not edit self permission")
	}

	authUser, err := svc.repo.FindByShopIDAndUsername(context.Background(), shopID, authUsername)

	if err != nil {
		return err
	}

	if authUser.Role != models.ROLE_OWNER {
		return errors.New("permission denied")
	}

	editusername = utils.NormalizeUsername(editusername)

	if editusername != "" {

		findEditUser, err := svc.repo.FindByShopIDAndUsername(context.Background(), shopID, editusername)

		if err != nil {
			return err
		}

		if findEditUser.Username != "" {
			tempID := findEditUser.ID

			err = svc.repo.Update(context.Background(), tempID, shopID, username, role)
			if err != nil {
				return err
			}
		} else {
			err = svc.repo.Save(context.Background(), shopID, username, role)

			if err != nil {
				return err
			}
		}

	} else {

		err = svc.repo.Save(context.Background(), shopID, username, role)

		if err != nil {
			return err
		}
	}
	return nil
}

func (svc ShopUserService) create(ctx context.Context, shopID string, username string, role models.UserRole) error {

	tempShopUser := models.ShopUser{}
	tempShopUser.ShopID = shopID
	tempShopUser.Username = username
	tempShopUser.Role = role

	err := svc.repo.Create(ctx, &tempShopUser)
	if err != nil {
		return err
	}
	return nil
}

// SaveUserFullProfile - บันทึกข้อมูลผู้ใช้แบบครบถ้วน (รวม position, department, LINE, approval)
func (svc ShopUserService) SaveUserFullProfile(shopID string, authUsername string, req *models.UserRoleRequest) error {
	username := utils.NormalizeUsername(req.Username)
	editusername := utils.NormalizeUsername(req.EditUsername)

	if authUsername == username || authUsername == editusername {
		return errors.New("can not edit self permission")
	}

	authUser, err := svc.repo.FindByShopIDAndUsername(context.Background(), shopID, authUsername)
	if err != nil {
		return err
	}

	if authUser.Role != models.ROLE_OWNER {
		return errors.New("permission denied")
	}

	// ตรวจสอบว่า LINE ID ไม่ซ้ำกับผู้ใช้คนอื่น - ถ้าซ้ำให้ auto-unlink คนเก่า
	if req.LineUserID != "" {
		existingUser, _ := svc.repo.FindByShopIDAndLineUserID(context.Background(), shopID, req.LineUserID)
		// ถ้าพบผู้ใช้ที่ใช้ LINE ID นี้แล้ว และไม่ใช่ผู้ใช้คนเดียวกัน ให้ลบ LINE data ของคนเก่า (auto-unlink)
		if existingUser.Username != "" && existingUser.Username != username {
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
			}
			svc.repo.SaveFullProfile(context.Background(), shopID, oldReq)
		}
	}

	// Normalize username in request
	req.Username = username

	// บันทึกข้อมูลแบบ full profile
	err = svc.repo.SaveFullProfile(context.Background(), shopID, req)
	if err != nil {
		return err
	}

	return nil
}

func (svc ShopUserService) DeleteUserPermissionShop(shopID string, authUsername string, username string) error {

	authUser, err := svc.repo.FindByShopIDAndUsername(context.Background(), shopID, authUsername)

	if err != nil {
		return err
	}

	if authUser.Role != models.ROLE_OWNER && authUser.Role != models.ROLE_ADMIN {
		return errors.New("permission denied")
	}

	findUser, err := svc.repo.FindByShopIDAndUsername(context.Background(), shopID, username)

	if err != nil {
		return err
	}

	// ตรวจสอบว่าพบผู้ใช้ที่ต้องการลบหรือไม่
	if findUser.Username == "" {
		return errors.New("user not found")
	}

	if authUser.Role == models.ROLE_ADMIN && findUser.Role == models.ROLE_OWNER {
		return errors.New("permission denied")
	}

	if findUser.Username == authUsername {
		return errors.New("can't delete your permission")
	}

	err = svc.repo.Delete(context.Background(), shopID, username)

	if err != nil {
		return err
	}
	return nil
}

// CleanupEmptyUsers - ลบ users ที่ username ว่างออกจาก shop
func (svc ShopUserService) CleanupEmptyUsers(shopID string) (int64, error) {
	return svc.repo.DeleteEmptyUsernames(context.Background(), shopID)
}

// SyncLineData - sync LINE data จาก LIFF callback (ใช้สำหรับ callback จาก lineoa-liff หลัง linking สำเร็จ)
func (svc ShopUserService) SyncLineData(shopID string, username string, lineUserID string, lineDisplayName string, linePictureURL string) error {
	username = utils.NormalizeUsername(username)

	if username == "" {
		return errors.New("username is required")
	}

	// ตรวจสอบว่ามี user นี้ใน shop หรือไม่
	existingUser, err := svc.repo.FindByShopIDAndUsername(context.Background(), shopID, username)
	if err != nil {
		return errors.New("user not found in shop")
	}

	if existingUser.Username == "" {
		return errors.New("user not found in shop")
	}

	// ถ้า LINE ID ถูกใช้โดยผู้ใช้คนอื่นแล้ว ให้ลบ LINE data ของคนนั้นก่อน (auto-unlink)
	if lineUserID != "" {
		existingLineUser, _ := svc.repo.FindByShopIDAndLineUserID(context.Background(), shopID, lineUserID)
		if existingLineUser.Username != "" && existingLineUser.Username != username {
			// ลบ LINE data ของ user เก่า
			oldReq := &models.UserRoleRequest{
				Username:        existingLineUser.Username,
				Role:            existingLineUser.Role,
				Position:        existingLineUser.Position,
				Department:      existingLineUser.Department,
				LineUserID:      "", // clear LINE data
				LineDisplayName: "",
				LinePictureURL:  "",
			}
			svc.repo.SaveFullProfile(context.Background(), shopID, oldReq)
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
	}

	return svc.repo.SaveFullProfile(context.Background(), shopID, req)
}

// SaveMyLineData - ให้ผู้ใช้อัปเดต LINE data ของตัวเอง (ใช้จาก Flutter หลัง LIFF linking)
func (svc ShopUserService) SaveMyLineData(shopID string, username string, lineUserID string, lineDisplayName string, linePictureURL string) error {
	username = utils.NormalizeUsername(username)

	if username == "" {
		return errors.New("username is required")
	}

	// ตรวจสอบว่ามี user นี้ใน shop หรือไม่
	existingUser, err := svc.repo.FindByShopIDAndUsername(context.Background(), shopID, username)
	if err != nil {
		return errors.New("user not found in shop")
	}

	if existingUser.Username == "" {
		return errors.New("user not found in shop")
	}

	// ถ้า LINE ID ถูกใช้โดยผู้ใช้คนอื่นแล้ว ให้ลบ LINE data ของคนนั้นก่อน (auto-unlink)
	if lineUserID != "" {
		existingLineUser, _ := svc.repo.FindByShopIDAndLineUserID(context.Background(), shopID, lineUserID)
		if existingLineUser.Username != "" && existingLineUser.Username != username {
			// ลบ LINE data ของ user เก่า
			oldReq := &models.UserRoleRequest{
				Username:        existingLineUser.Username,
				Role:            existingLineUser.Role,
				Position:        existingLineUser.Position,
				Department:      existingLineUser.Department,
				LineUserID:      "", // clear LINE data
				LineDisplayName: "",
				LinePictureURL:  "",
			}
			svc.repo.SaveFullProfile(context.Background(), shopID, oldReq)
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
	}

	return svc.repo.SaveFullProfile(context.Background(), shopID, req)
}
