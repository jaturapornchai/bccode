package shop

import (
	"context"
	"net/mail"
	"smlcloudplatform/internal/authentication/models"
	shopmodels "smlcloudplatform/internal/shop/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IShopUserRepository interface {
	Create(ctx context.Context, shopUser *models.ShopUser) error
	Update(ctx context.Context, id primitive.ObjectID, shopID string, username string, role models.UserRole) error
	Save(ctx context.Context, shopID string, username string, role models.UserRole) error
	SaveFullProfile(ctx context.Context, shopID string, req *models.UserRoleRequest) error
	UpdateLastAccess(ctx context.Context, shopID string, username string, lastAccessedAt time.Time) error
	SaveFavorite(ctx context.Context, shopID string, username string, isFavorite bool) error
	Delete(ctx context.Context, shopID string, username string) error
	DeleteEmptyUsernames(ctx context.Context, shopID string) (int64, error)
	FindByShopIDAndUsernameInfo(ctx context.Context, shopID string, username string) (models.ShopUserInfo, error)
	FindByShopIDAndUsername(ctx context.Context, shopID string, username string) (models.ShopUser, error)
	FindByShopIDAndLineUserID(ctx context.Context, shopID string, lineUserID string) (models.ShopUser, error)
	FindByLineUserID(ctx context.Context, lineUserID string) (models.ShopUser, error)
	FindShopCreatedBy(ctx context.Context, shopID string) (string, error)
	FindRole(ctx context.Context, shopID string, username string) (models.UserRole, error)
	FindByShopID(ctx context.Context, shopID string) (*[]models.ShopUser, error)
	FindByUsername(ctx context.Context, username string) (*[]models.ShopUser, error)
	FindByUsernamePage(ctx context.Context, username string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error)
	FindByUserInShopPage(ctx context.Context, shopID string, pageable micromodels.Pageable) ([]models.ShopUser, mongopagination.PaginationData, error)
	FindUserProfileByUsernames(ctx context.Context, usernames []string) ([]models.UserProfile, error)
}

type ShopUserRepository struct {
	pst microservice.IPersisterMongo
}

func NewShopUserRepository(pst microservice.IPersisterMongo) ShopUserRepository {
	return ShopUserRepository{
		pst: pst,
	}
}

func (svc ShopUserRepository) Create(ctx context.Context, shopUser *models.ShopUser) error {

	_, err := svc.pst.Create(ctx, &models.ShopUser{}, shopUser)

	if err != nil {
		return err
	}

	return nil
}

func (svc ShopUserRepository) Update(ctx context.Context, id primitive.ObjectID, shopID string, username string, role models.UserRole) error {

	err := svc.pst.Update(ctx, &models.ShopUser{}, bson.M{"_id": id, "shopid": shopID}, bson.M{"$set": bson.M{"username": username, "role": role}})

	if err != nil {
		return err
	}

	return nil
}

func (svc ShopUserRepository) Save(ctx context.Context, shopID string, username string, role models.UserRole) error {

	optUpdate := options.Update().SetUpsert(true)
	err := svc.pst.Update(ctx, &models.ShopUser{}, bson.M{"shopid": shopID, "username": username}, bson.M{"$set": bson.M{"role": role}}, optUpdate)

	if err != nil {
		return err
	}

	return nil
}

// SaveFullProfile - บันทึกข้อมูลผู้ใช้แบบครบถ้วน (รวม position, department, LINE, approval)
func (svc ShopUserRepository) SaveFullProfile(ctx context.Context, shopID string, req *models.UserRoleRequest) error {
	updateData := bson.M{
		"role":              req.Role,
		"isaccessdisabled":  req.IsAccessDisabled,
		"accessdisabledat":  req.AccessDisabledAt,
		"accessdisabledby":  req.AccessDisabledBy,
		"accessenabledat":   req.AccessEnabledAt,
		"accessenabledby":   req.AccessEnabledBy,
		"position":          req.Position,
		"department":        req.Department,
		"line_user_id":      req.LineUserID,
		"line_display_name": req.LineDisplayName,
		"line_picture_url":  req.LinePictureURL,
	}

	// เพิ่มข้อมูลการอนุมัติแยกตามประเภทเอกสาร
	if req.POApproval != nil {
		updateData["po_approval"] = req.POApproval
	}
	if req.QuotationApproval != nil {
		updateData["quotation_approval"] = req.QuotationApproval
	}

	optUpdate := options.Update().SetUpsert(true)
	err := svc.pst.Update(ctx, &models.ShopUser{}, bson.M{"shopid": shopID, "username": req.Username}, bson.M{"$set": updateData}, optUpdate)

	if err != nil {
		return err
	}

	return nil
}

func (svc ShopUserRepository) UpdateLastAccess(ctx context.Context, shopID string, username string, lastAccessedAt time.Time) error {

	optUpdate := options.Update().SetUpsert(true)
	err := svc.pst.Update(ctx, &models.ShopUser{}, bson.M{"shopid": shopID, "username": username}, bson.M{"$set": bson.M{"lastaccessedat": lastAccessedAt}}, optUpdate)

	if err != nil {
		return err
	}

	return nil
}

func (svc ShopUserRepository) SaveFavorite(ctx context.Context, shopID string, username string, isFavorite bool) error {

	optUpdate := options.Update().SetUpsert(true)
	err := svc.pst.Update(ctx, &models.ShopUser{}, bson.M{"shopid": shopID, "username": username}, bson.M{"$set": bson.M{"isfavorite": isFavorite}}, optUpdate)

	if err != nil {
		return err
	}

	return nil
}

func (svc ShopUserRepository) Delete(ctx context.Context, shopID string, username string) error {

	err := svc.pst.Delete(ctx, &models.ShopUser{}, bson.M{"shopid": shopID, "username": username})

	if err != nil {
		return err
	}

	return nil
}

// DeleteEmptyUsernames - ลบ users ที่ username ว่างหรือเป็น empty string
func (svc ShopUserRepository) DeleteEmptyUsernames(ctx context.Context, shopID string) (int64, error) {
	// ใช้ Exec เพื่อเข้าถึง collection โดยตรง
	collection, err := svc.pst.Exec(ctx, &models.ShopUser{})
	if err != nil {
		return 0, err
	}

	// ลบ users ที่ username ว่าง, เป็น null, หรือเป็น empty string
	filter := bson.M{
		"shopid": shopID,
		"$or": []bson.M{
			{"username": ""},
			{"username": bson.M{"$exists": false}},
			{"username": nil},
		},
	}

	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

func (svc ShopUserRepository) FindByShopIDAndUsernameInfo(ctx context.Context, shopID string, username string) (models.ShopUserInfo, error) {

	shopUser := &models.ShopUserInfo{}

	err := svc.pst.FindOne(ctx, &models.ShopUserInfo{}, bson.M{"shopid": shopID, "username": username}, shopUser)
	if err != nil {
		return models.ShopUserInfo{}, err
	}

	return *shopUser, nil
}

func (svc ShopUserRepository) FindByShopIDAndUsername(ctx context.Context, shopID string, username string) (models.ShopUser, error) {

	shopUser := &models.ShopUser{}

	err := svc.pst.FindOne(ctx, &models.ShopUser{}, bson.M{"shopid": shopID, "username": username}, shopUser)
	if err != nil {
		return models.ShopUser{}, err
	}

	return *shopUser, nil
}

func (svc ShopUserRepository) FindShopCreatedBy(ctx context.Context, shopID string) (string, error) {
	shopDoc := &shopmodels.ShopDoc{}

	err := svc.pst.FindOne(ctx, &shopmodels.ShopDoc{}, bson.M{"guidfixed": shopID}, shopDoc)
	if err != nil {
		return "", err
	}

	return shopDoc.CreatedBy, nil
}

// FindByShopIDAndLineUserID - ค้นหาผู้ใช้ตาม LINE User ID (สำหรับตรวจสอบ duplicate)
func (svc ShopUserRepository) FindByShopIDAndLineUserID(ctx context.Context, shopID string, lineUserID string) (models.ShopUser, error) {
	shopUser := &models.ShopUser{}

	err := svc.pst.FindOne(ctx, &models.ShopUser{}, bson.M{"shopid": shopID, "line_user_id": lineUserID}, shopUser)
	if err != nil {
		return models.ShopUser{}, err
	}

	return *shopUser, nil
}

// FindByLineUserID - ค้นหาผู้ใช้จาก LINE User ID (ไม่ filter shopID)
// สำหรับ LINE Login — หา user ที่เชื่อมต่อ LINE ไว้ ไม่ว่าจะอยู่ shop ไหน
func (svc ShopUserRepository) FindByLineUserID(ctx context.Context, lineUserID string) (models.ShopUser, error) {
	shopUser := &models.ShopUser{}

	err := svc.pst.FindOne(ctx, &models.ShopUser{}, bson.M{"line_user_id": lineUserID}, shopUser)
	if err != nil {
		return models.ShopUser{}, err
	}

	return *shopUser, nil
}

func (svc ShopUserRepository) FindRole(ctx context.Context, shopID string, username string) (models.UserRole, error) {

	shopUser := &models.ShopUser{}

	err := svc.pst.FindOne(ctx, &models.ShopUser{}, bson.M{"shopid": shopID, "username": username}, shopUser)

	if err != nil {
		return models.ROLE_USER, err
	}

	return shopUser.Role, nil
}

func (svc ShopUserRepository) FindByShopID(ctx context.Context, shopID string) (*[]models.ShopUser, error) {
	shopUsers := &[]models.ShopUser{}

	err := svc.pst.Find(ctx, &models.ShopUser{}, bson.M{"shopid": shopID}, shopUsers)

	if err != nil {
		return nil, err
	}

	return shopUsers, nil
}

func (svc ShopUserRepository) FindByUsername(ctx context.Context, username string) (*[]models.ShopUser, error) {
	shopUsers := &[]models.ShopUser{}

	err := svc.pst.Find(ctx, &models.ShopUser{}, bson.M{"username": username}, shopUsers)

	if err != nil {
		return nil, err
	}

	return shopUsers, nil
}

func (repo ShopUserRepository) FindByUsernamePage(ctx context.Context, username string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error) {

	docList := []models.ShopUserInfo{}

	searchFilterList := []interface{}{}

	searchInFields := []string{
		"shopid",
		"names",
	}

	for _, colName := range searchInFields {
		searchFilterList = append(searchFilterList, bson.M{colName: bson.M{"$regex": primitive.Regex{
			Pattern: ".*" + pageable.Query + ".*",
			Options: "",
		}}})
	}

	aggPaginatedData, err := repo.pst.AggregatePage(ctx, &models.ShopUser{}, pageable,
		bson.M{"$match": bson.M{
			"username": username,
			"deletedat": bson.M{
				"$exists": false,
			},
		}},
		bson.M{"$lookup": bson.M{
			"from":         "shops",
			"localField":   "shopid",
			"foreignField": "guidfixed",
			"as":           "shopInfo",
		}},
		bson.M{
			"$match": bson.M{"shopInfo.deletedAt": bson.M{"$exists": false}},
		},
		bson.M{
			"$project": bson.M{
				"_id":              1,
				"role":             1,
				"shopid":           1,
				"isfavorite":       1,
				"lastaccessedat":   1,
				"isaccessdisabled": 1,
				"accessdisabledat": 1,
				"accessdisabledby": 1,
				"accessenabledat":  1,
				"accessenabledby":  1,
				"names":            bson.M{"$first": "$shopInfo.names"},
				"branchcode":       bson.M{"$first": "$shopInfo.branchcode"},
				"createdby":        bson.M{"$first": "$shopInfo.createdby"},
			},
		},
		bson.M{
			"$match": bson.M{
				"$or": searchFilterList,
			},
		},
		bson.M{
			"$sort": bson.M{
				"lastaccessedat": -1,
			},
		},
	)

	if err != nil {
		return []models.ShopUserInfo{}, mongopagination.PaginationData{}, err
	}

	for _, raw := range aggPaginatedData.Data {
		var doc *models.ShopUserInfo

		if marshallErr := bson.Unmarshal(raw, &doc); marshallErr == nil {
			docList = append(docList, *doc)
		}

	}

	return docList, aggPaginatedData.Pagination, nil
}

func (repo ShopUserRepository) FindByUserInShopPage(ctx context.Context, shopID string, pageable micromodels.Pageable) ([]models.ShopUser, mongopagination.PaginationData, error) {

	docList := []models.ShopUser{}

	searchInFields := []string{
		"username",
	}

	searchFilterList := []interface{}{}

	for _, colName := range searchInFields {
		searchFilterList = append(searchFilterList, bson.M{colName: bson.M{"$regex": primitive.Regex{
			Pattern: ".*" + pageable.Query + ".*",
			Options: "",
		}}})
	}

	filtter := bson.M{
		"shopid": shopID,
		"$or":    searchFilterList,
	}

	paginattion, err := repo.pst.FindPage(ctx, &models.ShopUser{}, filtter, pageable, &docList)

	if err != nil {
		return []models.ShopUser{}, mongopagination.PaginationData{}, err
	}

	return docList, paginattion, nil
}

func (repo ShopUserRepository) FindUserProfileByUsernames(ctx context.Context, usernames []string) ([]models.UserProfile, error) {

	docList := []models.UserProfile{}

	filters := bson.M{
		"username": bson.M{
			"$in": usernames,
		},
	}

	err := repo.pst.Find(ctx, &models.UserProfile{}, filters, &docList)

	if err != nil {
		return []models.UserProfile{}, err
	}

	for idx := range docList {
		username := strings.TrimSpace(docList[idx].Username)
		if username == "" {
			continue
		}

		updateSet := bson.M{}
		if strings.TrimSpace(docList[idx].UID) == "" {
			uid := utils.NewGUID()
			updateSet["uid"] = uid
			docList[idx].UID = uid
		}
		// Legacy Google/email accounts used username as the registered email before
		// the users.email field existed. Do not overwrite a stored email.
		if strings.TrimSpace(docList[idx].Email) == "" && isEmailUsername(username) {
			email := strings.ToLower(username)
			updateSet["email"] = email
			docList[idx].Email = email
		}
		if len(updateSet) == 0 {
			continue
		}

		err = repo.pst.Update(
			ctx,
			&models.UserDoc{},
			bson.M{"username": docList[idx].Username},
			bson.M{"$set": updateSet},
		)
		if err != nil {
			return []models.UserProfile{}, err
		}
	}

	return docList, nil
}

func isEmailUsername(username string) bool {
	address, err := mail.ParseAddress(username)
	if err != nil {
		return false
	}
	return strings.EqualFold(address.Address, strings.TrimSpace(username))
}
