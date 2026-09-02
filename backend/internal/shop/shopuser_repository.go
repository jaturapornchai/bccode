package shop

import (
	"context"
	"errors"
	"net/mail"
	"smlcloudplatform/internal/authentication/models"
	companymodels "smlcloudplatform/internal/organization/company/models"
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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IShopUserRepository interface {
	Create(ctx context.Context, shopUser *models.ShopUser) error
	Update(ctx context.Context, id primitive.ObjectID, holdingCode string, username string, role models.UserRole) error
	Save(ctx context.Context, holdingCode string, username string, role models.UserRole) error
	SaveStable(ctx context.Context, holdingCode string, holdingUID string, userUID string, username string, role models.UserRole, createdAt time.Time) error
	SaveFullProfile(ctx context.Context, holdingCode string, req *models.UserRoleRequest) error
	UpdateLineFields(ctx context.Context, holdingCode string, userUID string, lineUserID string, lineDisplayName string, linePictureURL string) error
	UpdateLastAccess(ctx context.Context, holdingCode string, userUID string, lastAccessedAt time.Time) error
	SaveFavorite(ctx context.Context, holdingCode string, userUID string, isFavorite bool) error
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
	ResolveCompanyUID(ctx context.Context, holdingCode string, businessCode string) (string, error)
}

func (svc ShopUserRepository) SaveStable(ctx context.Context, holdingCode string, holdingUID string, userUID string, username string, role models.UserRole, createdAt time.Time) error {
	holdingUID = strings.TrimSpace(holdingUID)
	userUID = strings.TrimSpace(userUID)
	if holdingUID == "" || userUID == "" {
		return errors.New("stable membership identity is required")
	}
	membership := &models.ShopUser{
		Version: 0,
		ShopUserBase: models.ShopUserBase{
			MembershipUID: utils.NewGUID(),
			HoldingUID:    holdingUID,
			HoldingCode:   holdingCode,
			UserUID:       userUID,
			Username:      strings.TrimSpace(username),
			Role:          role,
		},
		PermissionVersion: 0,
		IsDeleted:         false,
		CreatedAt:         createdAt.UTC(),
		CreatedBy:         userUID,
		IsAccessDisabled:  false,
		AccessScopes:      []models.AccessScope{},
	}
	_, err := svc.pst.Create(ctx, &models.ShopUser{}, membership)
	return err
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
		"useruid":          userUID,
		"username":         req.Username,
		"role":             req.Role,
		"isaccessdisabled": req.IsAccessDisabled,
		"accessdisabledat": req.AccessDisabledAt,
		"accessdisabledby": req.AccessDisabledBy,
		"accessenabledat":  req.AccessEnabledAt,
		"accessenabledby":  req.AccessEnabledBy,
		"accessexpirydate": req.AccessExpiryDate,
		"position":         req.Position,
		"department":       req.Department,
		"lineuserid":       req.LineUserID,
		"linedisplayname":  req.LineDisplayName,
		"linepictureurl":   req.LinePictureURL,
		"accessscopes":     req.AccessScopes,
		"permissionsets":   normalizePermissionSets(req.PermissionSets),
	}

	// เพิ่มข้อมูลการอนุมัติแยกตามประเภทเอกสาร
	if req.POApproval != nil {
		updateData["poapproval"] = req.POApproval
	}
	if req.QuotationApproval != nil {
		updateData["quotationapproval"] = req.QuotationApproval
	}

	// Upsert keyed by the stable identity. username (usercode) when present —
	// one membership per user per holding, and useruid is NOT the match key
	// because email-created memberships carry an empty useruid until first
	// login. Usercode-less members (Google-only, docs login.md) are keyed by
	// the stable useruid instead, and no global login profile is synced.
	var filter bson.M
	if strings.TrimSpace(req.Username) != "" {
		filter = bson.M{"holdingcode": holdingCode, "username": req.Username}
		if strings.TrimSpace(req.EditUsername) != "" {
			filter = bson.M{"holdingcode": holdingCode, "username": req.EditUsername}
		}
		if err := svc.saveUserLoginProfile(ctx, holdingCode, req); err != nil {
			return err
		}
		if userUID == "" {
			userUID = svc.lookupUserUID(ctx, req.Username)
			updateData["useruid"] = userUID
		}
	} else {
		filter = bson.M{"holdingcode": holdingCode, "useruid": userUID}
	}

	// Zero expiry means "no expiry": drop the field instead of writing the zero
	// time — listing queries match accessexpirydate $exists:false OR > now, so a
	// stored zero value would hide the membership from every listing.
	if req.AccessExpiryDate.IsZero() {
		delete(updateData, "accessexpirydate")
	}
	update := bson.M{"$set": updateData}
	if req.AccessExpiryDate.IsZero() {
		update["$unset"] = bson.M{"accessexpirydate": ""}
	}

	optUpdate := options.Update().SetUpsert(true)
	return svc.pst.Update(ctx, &models.ShopUser{}, filter, update, optUpdate)
}

func (svc ShopUserRepository) UpdateLineFields(ctx context.Context, holdingCode string, userUID string, lineUserID string, lineDisplayName string, linePictureURL string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	userUID = strings.TrimSpace(userUID)
	if holdingCode == "" || userUID == "" {
		return errors.New("membership identity is required")
	}

	return svc.pst.Update(ctx, &models.ShopUser{}, bson.M{
		"holdingcode": holdingCode,
		"useruid":     userUID,
		"isdeleted":   bson.M{"$ne": true},
	}, bson.M{"$set": bson.M{
		"lineuserid":      lineUserID,
		"linedisplayname": lineDisplayName,
		"linepictureurl":  linePictureURL,
	}})
}

func (svc ShopUserRepository) saveUserLoginProfile(ctx context.Context, holdingCode string, req *models.UserRoleRequest) error {
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return errors.New("user identity is required")
	}

	// Editing: verify the usercode is not being changed by holding administration.
	// editusername may be the original usercode or, for legacy callers, the useruid.
	editUsername := strings.TrimSpace(req.EditUsername)
	if editUsername != "" {
		var existingMembership models.ShopUser
		err := svc.pst.FindOne(ctx, &models.ShopUser{}, bson.M{
			"holdingcode": holdingCode,
			"isdeleted":   bson.M{"$ne": true},
			"$or": bson.A{
				bson.M{"username": editUsername},
				bson.M{"useruid": editUsername},
			},
		}, &existingMembership)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return err
		}
		if existingMembership.Username != "" && !strings.EqualFold(existingMembership.Username, username) {
			return errors.New("global usercode cannot be changed by holding administration")
		}
	}

	existing := &models.UserDoc{}
	if err := svc.pst.FindOne(ctx, &models.UserDoc{}, bson.M{"username": username}, existing); err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	if existing.Username == "" || existing.UID == "" {
		return errors.New("user must sign in with Google and accept an invitation before membership is created")
	}
	email := strings.TrimSpace(req.Email)
	if email != "" {
		if !strings.EqualFold(strings.TrimSpace(existing.Email), email) {
			return errors.New("membership email must match the user's verified identity")
		}
	}
	return nil
}

func (svc ShopUserRepository) UpdateLastAccess(ctx context.Context, holdingCode string, userUID string, lastAccessedAt time.Time) error {
	if strings.TrimSpace(userUID) == "" {
		return errors.New("user identity is required")
	}
	err := svc.pst.Update(ctx, &models.ShopUser{}, bson.M{"holdingcode": holdingCode, "useruid": userUID}, bson.M{"$set": bson.M{"lastaccessedat": lastAccessedAt}})

	if err != nil {
		return err
	}

	return nil
}

func (svc ShopUserRepository) SaveFavorite(ctx context.Context, holdingCode string, userUID string, isFavorite bool) error {
	if strings.TrimSpace(userUID) == "" {
		return errors.New("user identity is required")
	}
	err := svc.pst.Update(ctx, &models.ShopUser{}, bson.M{"holdingcode": holdingCode, "useruid": userUID}, bson.M{"$set": bson.M{"isfavorite": isFavorite}})

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

func (svc ShopUserRepository) ResolveCompanyUID(ctx context.Context, holdingCode string, businessCode string) (string, error) {
	company := &companymodels.CompanyDoc{}
	err := svc.pst.FindOne(ctx, company, bson.M{
		"holdingcode": holdingCode,
		"code":        companymodels.NormalizeCompanyCode(businessCode),
		"deletedat":   bson.M{"$exists": false},
	}, company)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(company.GuidFixed) == "" || !company.IsActive {
		return "", errors.New("company is inactive or invalid")
	}
	return strings.TrimSpace(company.GuidFixed), nil
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

	// Fall back to a username lookup when there is no userUID, the userUID query errored,
	// OR it matched nothing. A membership created by email (before that person ever logged in)
	// keeps an empty/old useruid, while login later assigns a fresh uid — so the useruid query
	// misses and we must match by the stable email username instead.
	if userUID == "" || err != nil || shopUser.Username == "" {
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

	err := svc.pst.FindOne(ctx, &models.ShopUser{}, bson.M{"holdingcode": holdingCode, "lineuserid": lineUserID}, shopUser)
	if err != nil {
		return models.ShopUser{}, err
	}

	return *shopUser, nil
}

// FindByLineUserID - ค้นหาผู้ใช้จาก LINE User ID (ไม่ filter holdingCode)
// สำหรับ LINE Login — หา user ที่เชื่อมต่อ LINE ไว้ ไม่ว่าจะอยู่ shop ไหน
func (svc ShopUserRepository) FindByLineUserID(ctx context.Context, lineUserID string) (models.ShopUser, error) {
	shopUser := &models.ShopUser{}

	err := svc.pst.FindOne(ctx, &models.ShopUser{}, bson.M{"lineuserid": lineUserID}, shopUser)
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
	shopusers := &[]models.ShopUser{}

	err := svc.pst.Find(ctx, &models.ShopUser{}, bson.M{"holdingcode": holdingCode}, shopusers)

	if err != nil {
		return nil, err
	}

	return shopusers, nil
}

func (svc ShopUserRepository) FindByUsername(ctx context.Context, username string) (*[]models.ShopUser, error) {
	shopusers := &[]models.ShopUser{}

	err := svc.pst.Find(ctx, &models.ShopUser{}, bson.M{"username": username}, shopusers)

	if err != nil {
		return nil, err
	}

	return shopusers, nil
}

func (repo ShopUserRepository) FindByUsernamePage(ctx context.Context, username string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error) {
	return repo.findByUserPage(ctx, bson.M{"username": username}, pageable)
}

func (repo ShopUserRepository) FindByUserUIDPage(ctx context.Context, userUID string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error) {
	return repo.findByUserPage(ctx, bson.M{"useruid": userUID}, pageable)
}

func (repo ShopUserRepository) findByUserPage(ctx context.Context, userMatch bson.M, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error) {
	docList := []models.ShopUserInfo{}
	now := time.Now().UTC()
	matchFilter := bson.M{
		"isdeleted":        false,
		"isaccessdisabled": false,
		"$or": bson.A{
			bson.M{"accessexpirydate": bson.M{"$exists": false}},
			bson.M{"accessexpirydate": bson.M{"$gt": now}},
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
			"localField":   "holdinguid",
			"foreignField": "holdinguid",
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
				}},
				{"$project": bson.M{"code": 1}},
			},
			"as": "currencyInfo",
		}},
		bson.M{
			"$match": bson.M{
				"shopInfo": bson.M{"$elemMatch": bson.M{
					"isactive":  true,
					"isdeleted": false,
				}},
			},
		},
		bson.M{
			"$project": bson.M{
				"_id":              1,
				"holdinguid":       1,
				"role":             1,
				"isfavorite":       1,
				"lastaccessedat":   1,
				"isaccessdisabled": 1,
				"accessdisabledat": 1,
				"accessdisabledby": 1,
				"accessenabledat":  1,
				"accessenabledby":  1,
				"main_holdingcode": bson.M{"$first": "$shopInfo.mainholdingcode"},
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
		"linedisplayname",
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
		"lineuserid",
		"linedisplayname",
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

func (svc ShopUserRepository) shopUserIdentityFilter(_ context.Context, holdingCode string, username string) bson.M {
	// Memberships created before first login may not have useruid yet; username is
	// the stable holding-membership key used by SaveFullProfile.
	return bson.M{"holdingcode": holdingCode, "username": username}
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

// normalizePermissionSets upper-cases, trims and dedups permission set codes so
// they always match role_permission.rolecode; never nil so the field is stored.
func normalizePermissionSets(codes []string) []string {
	out := make([]string, 0, len(codes))
	seen := map[string]struct{}{}
	for _, code := range codes {
		code = strings.ToUpper(strings.TrimSpace(code))
		if code == "" {
			continue
		}
		if _, dup := seen[code]; dup {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}
	return out
}
