package shop

import (
	"context"
	"errors"
	"net/mail"
	"smlcloudplatform/internal/authentication/models"
	shopmodels "smlcloudplatform/internal/shop/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/search"
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
	Update(ctx context.Context, id primitive.ObjectID, holdingCode string, username string, role models.UserRole) error
	Save(ctx context.Context, holdingCode string, username string, role models.UserRole) error
	SaveFullProfile(ctx context.Context, holdingCode string, req *models.UserRoleRequest) error
	UpdateLastAccess(ctx context.Context, holdingCode string, username string, lastAccessedAt time.Time) error
	SaveFavorite(ctx context.Context, holdingCode string, username string, isFavorite bool) error
	Delete(ctx context.Context, holdingCode string, username string) error
	DeleteEmptyUsernames(ctx context.Context, holdingCode string) (int64, error)
	FindByHoldingCodeAndUserUIDInfo(ctx context.Context, holdingCode string, userUID string) (models.ShopUserInfo, error)
	FindByHoldingCodeAndUserUID(ctx context.Context, holdingCode string, userUID string) (models.ShopUser, error)
	FindByHoldingCodeAndUsernameInfo(ctx context.Context, holdingCode string, username string) (models.ShopUserInfo, error)
	FindByHoldingCodeAndUsername(ctx context.Context, holdingCode string, username string) (models.ShopUser, error)
	FindByHoldingCodeAndLineUserID(ctx context.Context, holdingCode string, lineUserID string) (models.ShopUser, error)
	FindByLineUserID(ctx context.Context, lineUserID string) (models.ShopUser, error)
	FindShopCreatedBy(ctx context.Context, holdingCode string) (string, error)
	FindRole(ctx context.Context, holdingCode string, username string) (models.UserRole, error)
	FindByHoldingCode(ctx context.Context, holdingCode string) (*[]models.ShopUser, error)
	FindByUsername(ctx context.Context, username string) (*[]models.ShopUser, error)
	FindByUsernamePage(ctx context.Context, username string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error)
	FindByUserUIDPage(ctx context.Context, userUID string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error)
	FindByUserInShopPage(ctx context.Context, holdingCode string, pageable micromodels.Pageable) ([]models.ShopUser, mongopagination.PaginationData, error)
	FindByUserInShopPageWithProfileMatches(ctx context.Context, holdingCode string, pageable micromodels.Pageable, profileUsernames []string) ([]models.ShopUser, mongopagination.PaginationData, error)
	FindUsernamesByProfileQuery(ctx context.Context, query string) ([]string, error)
	FindUserProfileByUsernames(ctx context.Context, usernames []string) ([]models.UserProfile, error)
	ResolveHoldingCodeByHoldingCode(ctx context.Context, holdingCode string) (string, error)
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
	if shopUser.UserUID == "" {
		shopUser.UserUID = svc.lookupUserUID(ctx, shopUser.Username)
	}

	_, err := svc.pst.Create(ctx, &models.ShopUser{}, shopUser)

	if err != nil {
		return err
	}

	return nil
}

func (svc ShopUserRepository) Update(ctx context.Context, id primitive.ObjectID, holdingCode string, username string, role models.UserRole) error {
	existing := &models.ShopUser{}
	_ = svc.pst.FindOne(ctx, &models.ShopUser{}, bson.M{"_id": id, "holdingcode": holdingCode}, existing)
	userUID := strings.TrimSpace(existing.UserUID)
	if userUID == "" {
		userUID = svc.lookupUserUID(ctx, username)
	}

	err := svc.pst.Update(ctx, &models.ShopUser{}, bson.M{"_id": id, "holdingcode": holdingCode}, bson.M{"$set": bson.M{"username": username, "role": role, "useruid": userUID}})

	if err != nil {
		return err
	}

	return nil
}

func (svc ShopUserRepository) Save(ctx context.Context, holdingCode string, username string, role models.UserRole) error {
	userUID := svc.lookupUserUID(ctx, username)
	filter := bson.M{"holdingcode": holdingCode, "username": username}
	if userUID != "" {
		filter = bson.M{"holdingcode": holdingCode, "useruid": userUID}
	}

	optUpdate := options.Update().SetUpsert(true)
	err := svc.pst.Update(ctx, &models.ShopUser{}, filter, bson.M{"$set": bson.M{"username": username, "role": role, "useruid": userUID}}, optUpdate)

	if err != nil {
		return err
	}

	return nil
}

func (svc ShopUserRepository) SaveFullProfile(ctx context.Context, holdingCode string, req *models.UserRoleRequest) error {
	userUID := strings.TrimSpace(req.UserUID)
	if userUID == "" {
		userUID = svc.lookupUserUID(ctx, req.Username)
	}
	updateData := bson.M{
		"useruid":           userUID,
		"username":          req.Username,
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
		"accessscopes":      req.AccessScopes,
	}

	// เพิ่มข้อมูลการอนุมัติแยกตามประเภทเอกสาร
	if req.POApproval != nil {
		updateData["po_approval"] = req.POApproval
	}
	if req.QuotationApproval != nil {
		updateData["quotation_approval"] = req.QuotationApproval
	}

	filter := bson.M{"holdingcode": holdingCode, "username": req.Username}
	if userUID != "" {
		filter = bson.M{"holdingcode": holdingCode, "useruid": userUID}
	} else if strings.TrimSpace(req.EditUsername) != "" {
		filter = bson.M{"holdingcode": holdingCode, "username": req.EditUsername}
	}

	optUpdate := options.Update().SetUpsert(true)
	err := svc.pst.Update(ctx, &models.ShopUser{}, filter, bson.M{"$set": updateData}, optUpdate)

	if err != nil {
		return err
	}

	if err := svc.saveUserLoginProfile(ctx, req); err != nil {
		return err
	}

	return nil
}

func (svc ShopUserRepository) saveUserLoginProfile(ctx context.Context, req *models.UserRoleRequest) error {
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return nil
	}

	updateData := bson.M{}
	if name := strings.TrimSpace(req.UserProfileName); name != "" {
		updateData["name"] = name
	}
	if !isEmailUsername(username) {
		updateData["email"] = strings.TrimSpace(req.Email)
	}
	if len(updateData) == 0 {
		return nil
	}
	updateData["updatedat"] = time.Now().UTC()

	optUpdate := options.Update().SetUpsert(false)
	return svc.pst.Update(ctx, &models.UserDoc{}, bson.M{"username": username}, bson.M{"$set": updateData}, optUpdate)
}

func (svc ShopUserRepository) UpdateLastAccess(ctx context.Context, holdingCode string, username string, lastAccessedAt time.Time) error {

	optUpdate := options.Update().SetUpsert(true)
	err := svc.pst.Update(ctx, &models.ShopUser{}, svc.shopUserIdentityFilter(ctx, holdingCode, username), bson.M{"$set": bson.M{"lastaccessedat": lastAccessedAt}}, optUpdate)

	if err != nil {
		return err
	}

	return nil
}

func (svc ShopUserRepository) SaveFavorite(ctx context.Context, holdingCode string, username string, isFavorite bool) error {

	optUpdate := options.Update().SetUpsert(true)
	err := svc.pst.Update(ctx, &models.ShopUser{}, svc.shopUserIdentityFilter(ctx, holdingCode, username), bson.M{"$set": bson.M{"isfavorite": isFavorite}}, optUpdate)

	if err != nil {
		return err
	}

	return nil
}

func (svc ShopUserRepository) Delete(ctx context.Context, holdingCode string, username string) error {

	err := svc.pst.Delete(ctx, &models.ShopUser{}, svc.shopUserIdentityFilter(ctx, holdingCode, username))

	if err != nil {
		return err
	}

	return nil
}

// DeleteEmptyUsernames - ลบ users ที่ username ว่างหรือเป็น empty string
func (svc ShopUserRepository) DeleteEmptyUsernames(ctx context.Context, holdingCode string) (int64, error) {
	// ใช้ Exec เพื่อเข้าถึง collection โดยตรง
	collection, err := svc.pst.Exec(ctx, &models.ShopUser{})
	if err != nil {
		return 0, err
	}

	// ลบ users ที่ username ว่าง, เป็น null, หรือเป็น empty string
	filter := bson.M{
		"holdingcode": holdingCode,
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

func (svc ShopUserRepository) FindByHoldingCodeAndUserUIDInfo(ctx context.Context, holdingCode string, userUID string) (models.ShopUserInfo, error) {
	shopUser := &models.ShopUserInfo{}
	err := svc.pst.FindOne(ctx, &models.ShopUserInfo{}, bson.M{"holdingcode": holdingCode, "useruid": userUID}, shopUser)
	if err != nil {
		return models.ShopUserInfo{}, err
	}
	return *shopUser, nil
}

func (svc ShopUserRepository) ResolveHoldingCodeByHoldingCode(ctx context.Context, holdingCode string) (string, error) {
	shopDoc := &shopmodels.ShopDoc{}
	err := svc.pst.FindOne(ctx, &shopmodels.ShopDoc{}, bson.M{"holdingcode": holdingCode, "deletedat": bson.M{"$exists": false}}, shopDoc)
	if err != nil || strings.TrimSpace(shopDoc.GuidFixed) == "" {
		legacyDoc := &shopmodels.ShopDoc{}
		legacyErr := svc.pst.FindOne(ctx, &shopmodels.ShopDoc{}, bson.M{"guidfixed": holdingCode, "deletedat": bson.M{"$exists": false}}, legacyDoc)
		if legacyErr != nil {
			if err != nil {
				return "", err
			}
			return "", legacyErr
		}
		shopDoc = legacyDoc
	}
	if strings.TrimSpace(shopDoc.GuidFixed) == "" {
		return "", errors.New("holding not found")
	}
	return shopDoc.GuidFixed, nil
}

func (svc ShopUserRepository) FindByHoldingCodeAndUserUID(ctx context.Context, holdingCode string, userUID string) (models.ShopUser, error) {
	shopUser := &models.ShopUser{}
	err := svc.pst.FindOne(ctx, &models.ShopUser{}, bson.M{"holdingcode": holdingCode, "useruid": userUID}, shopUser)
	if err != nil {
		return models.ShopUser{}, err
	}
	return *shopUser, nil
}

func (svc ShopUserRepository) FindByHoldingCodeAndUsernameInfo(ctx context.Context, holdingCode string, username string) (models.ShopUserInfo, error) {

	shopUser := &models.ShopUserInfo{}
	userUID := svc.lookupUserUID(ctx, username)

	var err error
	if userUID != "" {
		err = svc.pst.FindOne(ctx, &models.ShopUserInfo{}, bson.M{"holdingcode": holdingCode, "useruid": userUID}, shopUser)
	}

	if err != nil || userUID == "" {
		err = svc.pst.FindOne(ctx, &models.ShopUserInfo{}, bson.M{"holdingcode": holdingCode, "username": username}, shopUser)
	}

	if err != nil {
		return models.ShopUserInfo{}, err
	}

	return *shopUser, nil
}

func (svc ShopUserRepository) FindByHoldingCodeAndUsername(ctx context.Context, holdingCode string, username string) (models.ShopUser, error) {

	shopUser := &models.ShopUser{}
	userUID := svc.lookupUserUID(ctx, username)

	var err error
	if userUID != "" {
		err = svc.pst.FindOne(ctx, &models.ShopUser{}, bson.M{"holdingcode": holdingCode, "useruid": userUID}, shopUser)
	}

	if err != nil || userUID == "" {
		err = svc.pst.FindOne(ctx, &models.ShopUser{}, bson.M{"holdingcode": holdingCode, "username": username}, shopUser)
	}

	if err != nil {
		return models.ShopUser{}, err
	}

	return *shopUser, nil
}

func (svc ShopUserRepository) FindShopCreatedBy(ctx context.Context, holdingCode string) (string, error) {
	shopDoc := &shopmodels.ShopDoc{}

	err := svc.pst.FindOne(ctx, &shopmodels.ShopDoc{}, bson.M{"guidfixed": holdingCode}, shopDoc)
	if err != nil {
		return "", err
	}

	return shopDoc.CreatedBy, nil
}

// FindByHoldingCodeAndLineUserID - ค้นหาผู้ใช้ตาม LINE User ID (สำหรับตรวจสอบ duplicate)
func (svc ShopUserRepository) FindByHoldingCodeAndLineUserID(ctx context.Context, holdingCode string, lineUserID string) (models.ShopUser, error) {
	shopUser := &models.ShopUser{}

	err := svc.pst.FindOne(ctx, &models.ShopUser{}, bson.M{"holdingcode": holdingCode, "line_user_id": lineUserID}, shopUser)
	if err != nil {
		return models.ShopUser{}, err
	}

	return *shopUser, nil
}

// FindByLineUserID - ค้นหาผู้ใช้จาก LINE User ID (ไม่ filter holdingCode)
// สำหรับ LINE Login — หา user ที่เชื่อมต่อ LINE ไว้ ไม่ว่าจะอยู่ shop ไหน
func (svc ShopUserRepository) FindByLineUserID(ctx context.Context, lineUserID string) (models.ShopUser, error) {
	shopUser := &models.ShopUser{}

	err := svc.pst.FindOne(ctx, &models.ShopUser{}, bson.M{"line_user_id": lineUserID}, shopUser)
	if err != nil {
		return models.ShopUser{}, err
	}

	return *shopUser, nil
}

func (svc ShopUserRepository) FindRole(ctx context.Context, holdingCode string, username string) (models.UserRole, error) {

	shopUser, err := svc.FindByHoldingCodeAndUsername(ctx, holdingCode, username)

	if err != nil {
		return models.ROLE_USER, err
	}

	return shopUser.Role, nil
}

func (svc ShopUserRepository) FindByHoldingCode(ctx context.Context, holdingCode string) (*[]models.ShopUser, error) {
	shopUsers := &[]models.ShopUser{}

	err := svc.pst.Find(ctx, &models.ShopUser{}, bson.M{"holdingcode": holdingCode}, shopUsers)

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
	return repo.findByUserPage(ctx, bson.M{"username": username}, pageable)
}

func (repo ShopUserRepository) FindByUserUIDPage(ctx context.Context, userUID string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error) {
	return repo.findByUserPage(ctx, bson.M{"useruid": userUID}, pageable)
}

func (repo ShopUserRepository) findByUserPage(ctx context.Context, userMatch bson.M, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error) {
	docList := []models.ShopUserInfo{}
	matchFilter := bson.M{
		"deletedat": bson.M{
			"$exists": false,
		},
	}
	for key, value := range userMatch {
		matchFilter[key] = value
	}

	searchFilterList := []interface{}{}

	searchInFields := []string{
		"holdingcode",
		"names",
	}

	for _, colName := range searchInFields {
		searchFilterList = append(searchFilterList, bson.M{colName: bson.M{"$regex": primitive.Regex{
			Pattern: ".*" + pageable.Query + ".*",
			Options: "",
		}}})
	}

	aggPaginatedData, err := repo.pst.AggregatePage(ctx, &models.ShopUser{}, pageable,
		bson.M{"$match": matchFilter},
		bson.M{"$lookup": bson.M{
			"from":         "shops",
			"localField":   "holdingcode",
			"foreignField": "guidfixed",
			"as":           "shopInfo",
		}},
		bson.M{"$lookup": bson.M{
			"from": "currency",
			"let":  bson.M{"holdingCode": "$holdingcode"},
			"pipeline": []bson.M{
				{"$match": bson.M{
					"$expr": bson.M{"$and": []interface{}{
						bson.M{"$eq": []interface{}{"$holdingcode", "$$holdingCode"}},
						bson.M{"$ne": []interface{}{"$isdisabled", true}},
					}},
					"deletedat": bson.M{"$exists": false},
					"deletedat": bson.M{"$exists": false},
				}},
				{"$project": bson.M{"code": 1}},
			},
			"as": "currencyInfo",
		}},
		bson.M{
			"$match": bson.M{
				"shopInfo.0":         bson.M{"$exists": true},
				"shopInfo.deletedat": bson.M{"$exists": false},
				"shopInfo.deletedAt": bson.M{"$exists": false},
				"shopInfo.deletedat": bson.M{"$exists": false},
			},
		},
		bson.M{
			"$project": bson.M{
				"_id":              1,
				"role":             1,
				"isfavorite":       1,
				"lastaccessedat":   1,
				"isaccessdisabled": 1,
				"accessdisabledat": 1,
				"accessdisabledby": 1,
				"accessenabledat":  1,
				"accessenabledby":  1,
				"main_holdingcode": bson.M{"$first": "$shopInfo.main_holdingcode"},
				"holdingcode":      bson.M{"$first": "$shopInfo.holdingcode"},
				"names":            bson.M{"$first": "$shopInfo.names"},
				"branchcode":       bson.M{"$first": "$shopInfo.branchcode"},
				"createdby":        bson.M{"$first": "$shopInfo.createdby"},
				"language":         bson.M{"$first": "$shopInfo.settings.language"},
				"languageconfigs":  bson.M{"$ifNull": []interface{}{bson.M{"$first": "$shopInfo.settings.languageconfigs"}, []interface{}{}}},
				"basecurrency":     bson.M{"$first": "$shopInfo.settings.basecurrency"},
				"currencies": bson.M{"$map": bson.M{
					"input": "$currencyInfo",
					"as":    "currency",
					"in":    "$$currency.code",
				}},
				"timezone":            bson.M{"$first": "$shopInfo.settings.timezone"},
				"timezonelabel":       bson.M{"$first": "$shopInfo.settings.timezonelabel"},
				"timezoneoffset":      bson.M{"$first": "$shopInfo.settings.timezoneoffset"},
				"dateformat":          bson.M{"$first": "$shopInfo.settings.dateformat"},
				"usebuddhistcalendar": bson.M{"$first": "$shopInfo.settings.usebuddhistcalendar"},
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

func (repo ShopUserRepository) FindByUserInShopPage(ctx context.Context, holdingCode string, pageable micromodels.Pageable) ([]models.ShopUser, mongopagination.PaginationData, error) {
	return repo.FindByUserInShopPageWithProfileMatches(ctx, holdingCode, pageable, nil)
}

func (repo ShopUserRepository) FindByUserInShopPageWithProfileMatches(ctx context.Context, holdingCode string, pageable micromodels.Pageable, profileUsernames []string) ([]models.ShopUser, mongopagination.PaginationData, error) {
	docList := []models.ShopUser{}

	searchInFields := []string{
		"username",
		"position",
		"department",
		"line_display_name",
	}

	searchFilterList := search.CreateTextFilter(searchInFields, pageable.Query)
	if len(profileUsernames) > 0 {
		searchFilterList = append(searchFilterList, bson.M{"username": bson.M{"$in": profileUsernames}})
	}

	filtter := bson.M{
		"holdingcode": holdingCode,
	}
	if len(searchFilterList) > 0 {
		filtter["$or"] = searchFilterList
	}

	paginattion, err := repo.pst.FindPage(ctx, &models.ShopUser{}, filtter, pageable, &docList)

	if err != nil {
		return []models.ShopUser{}, mongopagination.PaginationData{}, err
	}

	return docList, paginattion, nil
}

func (repo ShopUserRepository) FindUsernamesByProfileQuery(ctx context.Context, query string) ([]string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []string{}, nil
	}

	docList := []models.UserProfile{}
	filters := bson.M{}
	searchFilterList := search.CreateTextFilter([]string{
		"username",
		"uid",
		"name",
		"email",
		"line_user_id",
		"line_display_name",
	}, query)
	if len(searchFilterList) > 0 {
		filters["$or"] = searchFilterList
	}

	if err := repo.pst.Find(ctx, &models.UserProfile{}, filters, &docList); err != nil {
		return []string{}, err
	}

	usernames := make([]string, 0, len(docList))
	seen := map[string]struct{}{}
	for _, doc := range docList {
		username := strings.TrimSpace(doc.Username)
		if username == "" {
			continue
		}
		if _, ok := seen[username]; ok {
			continue
		}
		seen[username] = struct{}{}
		usernames = append(usernames, username)
	}
	return usernames, nil
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

func (svc ShopUserRepository) shopUserIdentityFilter(ctx context.Context, holdingCode string, username string) bson.M {
	filter := bson.M{"holdingcode": holdingCode, "username": username}
	if userUID := svc.lookupUserUID(ctx, username); userUID != "" {
		filter = bson.M{"holdingcode": holdingCode, "useruid": userUID}
	}
	return filter
}

func (svc ShopUserRepository) lookupUserUID(ctx context.Context, username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return ""
	}
	userDoc := &models.UserDoc{}
	err := svc.pst.FindOne(ctx, &models.UserDoc{}, bson.M{"username": username}, userDoc)
	if err != nil {
		return ""
	}
	return userDoc.UID
}
