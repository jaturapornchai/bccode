package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/order/setting/models"
	"smlcloudplatform/internal/order/setting/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ISettingHttpService interface {
	CreateSetting(shopID string, authUsername string, doc models.OrderSetting) (string, error)
	UpdateSetting(shopID string, guid string, authUsername string, doc models.OrderSetting) error
	DeleteSetting(shopID string, guid string, authUsername string) error
	DeleteSettingByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoSetting(shopID string, guid string) (models.SettingInfo, error)
	InfoSettingByCode(shopID string, code string) (models.SettingInfo, error)
	SearchSetting(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SettingInfo, mongopagination.PaginationData, error)
	SearchSettingStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.SettingInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.OrderSetting) (common.BulkImport, error)

	GetModuleName() string
}

type SettingHttpService struct {
	repo repositories.ISettingRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.SettingActivity, models.SettingDeleteActivity]
	contextTimeout time.Duration
}

func NewSettingHttpService(repo repositories.ISettingRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *SettingHttpService {

	contextTimeout := time.Duration(15) * time.Second
	insSvc := &SettingHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.SettingActivity, models.SettingDeleteActivity](repo)

	return insSvc
}

func (svc SettingHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc SettingHttpService) CreateSetting(shopID string, authUsername string, doc models.OrderSetting) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDocCode, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if len(findDocCode.GuidFixed) > 0 {
		return "", errors.New("code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.SettingDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.OrderSetting = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(shopID)

	return newGuidFixed, nil
}

func (svc SettingHttpService) UpdateSetting(shopID string, guid string, authUsername string, doc models.OrderSetting) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	findDoc.OrderSetting = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc SettingHttpService) DeleteSetting(shopID string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)
	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc SettingHttpService) DeleteSettingByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	return nil
}

func (svc SettingHttpService) InfoSetting(shopID string, guid string) (models.SettingInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.SettingInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.SettingInfo{}, errors.New("document not found")
	}

	return findDoc.SettingInfo, nil
}

func (svc SettingHttpService) InfoSettingByCode(shopID string, code string) (models.SettingInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)

	if err != nil {
		return models.SettingInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.SettingInfo{}, errors.New("document not found")
	}

	return findDoc.SettingInfo, nil
}

func (svc SettingHttpService) SearchSetting(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SettingInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"devicenumber",
		"location.names.name",
		"warehouse.names.name",
		"branch.names.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.SettingInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc SettingHttpService) SearchSettingStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.SettingInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"devicenumber",
		"location.names.name",
		"warehouse.names.name",
		"branch.names.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.SettingInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc SettingHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.OrderSetting) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.OrderSetting](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.Code)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, shopID, "code", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.Code)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.OrderSetting, models.SettingDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.OrderSetting) models.SettingDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.SettingDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.OrderSetting = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.OrderSetting, models.SettingDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.SettingDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", guid)
		},
		func(doc models.SettingDoc) bool {
			return doc.Code != ""
		},
		func(shopID string, authUsername string, data models.OrderSetting, doc models.SettingDoc) error {

			doc.OrderSetting = data
			doc.UpdatedBy = authUsername
			doc.UpdatedAt = time.Now()

			err = svc.repo.Update(ctx, shopID, doc.GuidFixed, doc)
			if err != nil {
				return nil
			}
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
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.Code)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {

		updateDataKey = append(updateDataKey, doc.Code)
	}

	updateFailDataKey := []string{}
	for _, doc := range updateFailDataList {
		updateFailDataKey = append(updateFailDataKey, svc.getDocIDKey(doc))
	}

	svc.saveMasterSync(shopID)

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc SettingHttpService) getDocIDKey(doc models.OrderSetting) string {
	return doc.Code
}

func (svc SettingHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc SettingHttpService) GetModuleName() string {
	return "pos-setting"
}
