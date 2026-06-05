package settings

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/restaurant/settings/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IRestaurantSettingsService interface {
	CreateRestaurantSettings(holdingCode string, authUsername string, doc models.RestaurantSettings) (string, error)
	UpdateRestaurantSettings(holdingCode string, guid string, authUsername string, doc models.RestaurantSettings) error
	DeleteByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	DeleteRestaurantSettings(holdingCode string, guid string, authUsername string) error
	InfoRestaurantSettings(holdingCode string, guid string) (models.RestaurantSettingsInfo, error)
	SearchRestaurantSettings(holdingCode string, pageable micromodels.Pageable) ([]models.RestaurantSettingsInfo, mongopagination.PaginationData, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.RestaurantSettings) (common.BulkImport, error)
	ListRestaurantSettingsByCode(holdingCode string, code string, pagable micromodels.Pageable) ([]models.RestaurantSettingsInfo, mongopagination.PaginationData, error)

	GetModuleName() string
}

type RestaurantSettingsService struct {
	repo          IRestaurantSettingsRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository

	services.ActivityService[models.RestaurantSettingsActivity, models.RestaurantSettingsDeleteActivity]
	contextTimeout time.Duration
}

func NewRestaurantSettingsService(repo IRestaurantSettingsRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) RestaurantSettingsService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := RestaurantSettingsService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.RestaurantSettingsActivity, models.RestaurantSettingsDeleteActivity](repo)
	return insSvc
}

func (svc RestaurantSettingsService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc RestaurantSettingsService) CreateRestaurantSettings(holdingCode string, authUsername string, doc models.RestaurantSettings) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if findDoc.ID != primitive.NilObjectID {
		return "", errors.New("document already exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.RestaurantSettingsDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.RestaurantSettings = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	docData.LastUpdatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)

	return newGuidFixed, nil
}

func (svc RestaurantSettingsService) UpdateRestaurantSettings(holdingCode string, guid string, authUsername string, doc models.RestaurantSettings) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.RestaurantSettings = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	findDoc.LastUpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc RestaurantSettingsService) DeleteRestaurantSettings(holdingCode string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc RestaurantSettingsService) DeleteByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	return nil
}

func (svc RestaurantSettingsService) InfoRestaurantSettings(holdingCode string, guid string) (models.RestaurantSettingsInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.RestaurantSettingsInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.RestaurantSettingsInfo{}, errors.New("document not found")
	}

	return findDoc.RestaurantSettingsInfo, nil
}

func (svc RestaurantSettingsService) ListRestaurantSettingsByCode(holdingCode string, code string, pagable micromodels.Pageable) ([]models.RestaurantSettingsInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, map[string]interface{}{"code": code}, []string{"body"}, pagable)

	if err != nil {
		return []models.RestaurantSettingsInfo{}, mongopagination.PaginationData{}, err
	}

	return docList, pagination, nil

}

func (svc RestaurantSettingsService) SearchRestaurantSettings(holdingCode string, pageable micromodels.Pageable) ([]models.RestaurantSettingsInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		// "body",
	}

	docList, pagination, err := svc.repo.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return []models.RestaurantSettingsInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc RestaurantSettingsService) SaveInBatch(holdingCode string, authUsername string, dataList []models.RestaurantSettings) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadCategoryList, payloadDuplicateCategoryList := importdata.FilterDuplicate[models.RestaurantSettings](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadCategoryList {
		itemCodeGuidList = append(itemCodeGuidList, doc.Code)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, holdingCode, "code", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.Code)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.RestaurantSettings, models.RestaurantSettingsDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadCategoryList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.RestaurantSettings) models.RestaurantSettingsDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.RestaurantSettingsDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.RestaurantSettings = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			dataDoc.LastUpdatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.RestaurantSettings, models.RestaurantSettingsDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.RestaurantSettingsDoc, error) {
			return svc.repo.FindByGuid(ctx, holdingCode, guid)
		},
		func(doc models.RestaurantSettingsDoc) bool {
			return false
		},
		func(holdingCode string, authUsername string, data models.RestaurantSettings, doc models.RestaurantSettingsDoc) error {

			return nil
		},
	)

	if len(createDataList) > 0 {
		err = svc.repo.CreateInBatch(ctx, createDataList)

		if err != nil {
			return common.BulkImport{}, err
		}
	}
	createDataKey := []string{}

	for _, doc := range createDataList {
		createDataKey = append(createDataKey, doc.Code)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateCategoryList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.Code)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {
		updateDataKey = append(updateDataKey, doc.Code)
	}

	updateFailDataKey := []string{}
	for _, doc := range updateFailDataList {
		updateFailDataKey = append(updateFailDataKey, doc.Code)
	}

	svc.saveMasterSync(holdingCode)

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc RestaurantSettingsService) getDocIDKey(doc models.RestaurantSettings) string {
	return doc.Code
}

func (svc RestaurantSettingsService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc RestaurantSettingsService) GetModuleName() string {
	return "restaurantsettings"
}
