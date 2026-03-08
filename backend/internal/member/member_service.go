package member

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	"smlcloudplatform/internal/member/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/shop"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/samber/lo"
	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IMemberService interface {
	AuthWithLine(lineAuth models.LineAuthRequest) (string, error)
	UpdateProfileWithLine(shopID string, lineUID string, doc models.Member) error
	LineProfileInfo(shopID string, lineUID string) (models.MemberInfo, error)

	Create(shopID string, authUsername string, doc models.Member) (string, error)
	Update(shopID string, username string, guid string, doc models.Member) error
	Info(shopID string, guid string) (models.MemberInfo, error)
	SearchMemberInfo(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MemberInfo, mongopagination.PaginationData, error)
	SearchMemberStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.MemberInfo, int, error)
}

type MemberService struct {
	shopService   shop.IShopService
	repo          IMemberRepository
	memberPgRepo  IMemberPGRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository
	authService   *microservice.AuthService

	services.ActivityService[models.MemberActivity, models.MemberDeleteActivity]
	contextTimeout time.Duration
}

func NewMemberService(repo IMemberRepository, memberPgRepo IMemberPGRepository, shopService shop.IShopService, authService *microservice.AuthService, syncCacheRepo mastersync.IMasterSyncCacheRepository) MemberService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := MemberService{
		shopService:    shopService,
		repo:           repo,
		memberPgRepo:   memberPgRepo,
		authService:    authService,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}
	return insSvc
}

func (svc MemberService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc MemberService) UpdateProfileWithLine(shopID string, lineUID string, doc models.Member) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByLineUID(ctx, lineUID)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("guid invalid")
	}

	if findDoc.MemberType != models.MemberTypeLine {
		return errors.New("member type invalid")
	}

	dataDoc := findDoc

	dataDoc.Member = doc

	if dataDoc.Shops == nil {
		dataDoc.Shops = &[]string{}
	}

	*dataDoc.Shops = append(*dataDoc.Shops, shopID)
	*dataDoc.Shops = lo.Uniq[string](*dataDoc.Shops)

	if doc.Addresses == nil {
		dataDoc.Addresses = &[]models.MemberAddress{}

	}

	dataDoc.MemberType = findDoc.MemberType
	dataDoc.LineUID = findDoc.LineUID
	dataDoc.UpdatedBy = lineUID
	dataDoc.UpdatedAt = time.Now()

	dataDoc.LastUpdatedAt = time.Now()

	err = svc.repo.Update(ctx, findDoc.GuidFixed, dataDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc MemberService) registerWithLine(shopID string, lineProfile models.LineProfile) (string, error) {

	idx, err := svc.create(shopID, lineProfile.UserID, models.MemberTypeLine, models.Member{
		LineUID:    lineProfile.UserID,
		Name:       lineProfile.DisplayName,
		PictureUrl: lineProfile.PictureUrl,
		Provider:   &[]string{"line"},
	})

	if err != nil {
		return "", err
	}

	shopInfo, err := svc.shopService.InfoShop(shopID)

	if err != nil {
		return "", err
	}

	if shopInfo.GuidFixed == "" {
		return "", errors.New("shop invalid")
	}

	return idx, nil
}

func (svc MemberService) AuthWithLine(lineAuth models.LineAuthRequest) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	_, lineProfile, err := svc.LineValidator(lineAuth.LineAccessToken)

	if err != nil {
		return "", err
	}

	findDoc, err := svc.repo.FindByLineUID(ctx, lineProfile.UserID)

	if err != nil {
		return "", err
	}

	lineUID := ""
	memberName := ""

	if findDoc.GuidFixed == "" {
		_, err = svc.registerWithLine(lineAuth.ShopID, lineProfile)

		if err != nil {
			return "", err
		}

		lineUID = lineProfile.UserID
		memberName = lineProfile.DisplayName
	} else {
		lineUID = findDoc.Member.LineUID
		memberName = findDoc.Member.Name
	}

	userInfo := micromodels.UserInfo{
		Username: lineUID,
		Name:     memberName,
		ShopID:   lineAuth.ShopID,
		Role:     0,
	}

	tokenID, err := svc.authService.GenerateTokenWithRedisExpire(microservice.AUTHTYPE_BEARER, userInfo, time.Duration(24*30)*time.Hour)

	if err != nil {
		return "", err
	}

	return tokenID, nil

}

func (svc MemberService) LineValidator(lineToken string) (models.LineVerify, models.LineProfile, error) {
	lineVerify, err := svc.LineVerify(lineToken)

	if err != nil {
		return models.LineVerify{}, models.LineProfile{}, err
	}

	if lineVerify.ExpiresIn <= 0 {
		return models.LineVerify{}, models.LineProfile{}, errors.New("line token invalid")
	}

	if lineVerify.ClientID == "" {
		return models.LineVerify{}, models.LineProfile{}, errors.New("line token invalid")
	}

	lineProfile, err := svc.LineProfile(lineToken)

	if err != nil {
		return models.LineVerify{}, models.LineProfile{}, err
	}

	if lineProfile.UserID == "" {
		return models.LineVerify{}, models.LineProfile{}, errors.New("line token invalid")
	}

	return lineVerify, lineProfile, nil
}

func (svc MemberService) LineProfile(lineToken string) (models.LineProfile, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.line.me/v2/profile", nil)

	if err != nil {
		return models.LineProfile{}, err
	}

	req.Header.Add("Authorization", "Bearer "+lineToken)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return models.LineProfile{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return models.LineProfile{}, err
	}

	lineProfile := models.LineProfile{}

	err = json.Unmarshal(body, &lineProfile)
	if err != nil {
		return models.LineProfile{}, err
	}

	return lineProfile, nil
}

func (svc MemberService) LineVerify(lineToken string) (models.LineVerify, error) {

	req, err := http.NewRequest(http.MethodGet, "https://api.line.me/oauth2/v2.1/verify?access_token="+lineToken, nil)

	if err != nil {
		return models.LineVerify{}, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return models.LineVerify{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return models.LineVerify{}, err
	}

	lineVerify := models.LineVerify{}
	err = json.Unmarshal(body, &lineVerify)

	if err != nil {
		return models.LineVerify{}, err
	}

	return lineVerify, nil
}

func (svc MemberService) LineProfileInfo(shopID string, lineUID string) (models.MemberInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	doc, err := svc.repo.FindByLineUID(ctx, lineUID)

	if err != nil {
		return models.MemberInfo{}, err
	}

	return doc.MemberInfo, nil
}

func (svc MemberService) create(shopID string, authUsername string, memberType models.MemberType, doc models.Member) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	dataDoc := models.MemberDoc{}

	newGuid := utils.NewGUID()
	dataDoc.GuidFixed = newGuid
	dataDoc.Member = doc

	if dataDoc.Addresses == nil {
		dataDoc.Addresses = &[]models.MemberAddress{}
	}

	if dataDoc.Provider == nil {
		dataDoc.Provider = &[]string{}
	}

	dataDoc.MemberType = memberType
	dataDoc.CreatedBy = authUsername
	dataDoc.CreatedAt = time.Now()
	dataDoc.LastUpdatedAt = time.Now()
	dataDoc.Shops = &[]string{shopID}

	_, err := svc.repo.Create(ctx, dataDoc)

	if err != nil {
		return "", err
	}

	return newGuid, nil
}

func (svc MemberService) Create(shopID string, authUsername string, doc models.Member) (string, error) {

	doc.LineUID = ""
	idx, err := svc.create(shopID, authUsername, models.MemberTypeCustomer, doc)

	if err != nil {
		return "", err
	}

	return idx, nil
}

func (svc MemberService) Update(shopID string, username string, guid string, doc models.Member) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("guid invalid")
	}

	if findDoc.MemberType != models.MemberTypeCustomer {
		return errors.New("member type invalid")
	}

	dataDoc := findDoc

	dataDoc.Member = doc

	dataDoc.MemberType = findDoc.MemberType
	dataDoc.LineUID = findDoc.LineUID
	dataDoc.UpdatedBy = username
	dataDoc.UpdatedAt = time.Now()
	dataDoc.LastUpdatedAt = time.Now()

	*dataDoc.Shops = append(*dataDoc.Shops, shopID)
	*dataDoc.Shops = lo.Uniq[string](*dataDoc.Shops)

	err = svc.repo.Update(ctx, guid, dataDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc MemberService) Info(shopID string, guid string) (models.MemberInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	doc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.MemberInfo{}, err
	}

	return doc.MemberInfo, nil
}

func (svc MemberService) SearchMemberInfo(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MemberInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"name",
		"surname",
		"telephone",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return []models.MemberInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc MemberService) SearchMemberStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.MemberInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"name",
		"surname",
		"telephone",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.MemberInfo{}, 0, err
	}

	return docList, total, nil
}
