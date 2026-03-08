package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/product/productgroup/models"
	"smlcloudplatform/internal/product/productgroup/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IProductGroupHttpService interface {
	SaveProductGroup(shopID string, authUsername string, doc models.ProductGroup) (string, error)
	CreateProductGroup(shopID string, authUsername string, doc models.ProductGroup) (string, error)
	UpdateProductGroup(shopID string, guid string, authUsername string, doc models.ProductGroup) error
	DeleteProductGroup(shopID string, guid string, authUsername string) error
	DeleteProductGroupByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoProductGroup(shopID string, guid string) (models.ProductGroupInfo, error)
	InfoWTFArray(shopID string, unitCodes []string) ([]interface{}, error)
	SearchProductGroup(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductGroupInfo, mongopagination.PaginationData, error)
	SearchProductGroupStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.ProductGroupInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.ProductGroup) (common.BulkImport, error)

	GetModuleName() string
}

type ProductGroupHttpService struct {
	repo               repositories.IProductGroupRepository
	repoMessageQueue   repositories.IProductGroupMessageQueueRepository
	repoProductBarcode productbarcode_repositories.IProductBarcodeRepository
	syncCacheRepo      mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.ProductGroupActivity, models.ProductGroupDeleteActivity]
	productGroupServiceConfig config.IProductGroupServiceConfig
	contextTimeout            time.Duration
}

func NewProductGroupHttpService(
	repo repositories.IProductGroupRepository,
	repoMessageQueue repositories.IProductGroupMessageQueueRepository,
	repoProductBarcode productbarcode_repositories.IProductBarcodeRepository,
	productGroupServiceConfig config.IProductGroupServiceConfig,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
) *ProductGroupHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &ProductGroupHttpService{
		repo:                      repo,
		repoMessageQueue:          repoMessageQueue,
		repoProductBarcode:        repoProductBarcode,
		syncCacheRepo:             syncCacheRepo,
		productGroupServiceConfig: productGroupServiceConfig,
		contextTimeout:            contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.ProductGroupActivity, models.ProductGroupDeleteActivity](repo)

	return insSvc
}

func (svc ProductGroupHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ProductGroupHttpService) SaveProductGroup(shopID string, authUsername string, doc models.ProductGroup) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if len(findDoc.Code) > 0 {

		docData, err := svc.update(shopID, authUsername, findDoc.GuidFixed, findDoc, doc)

		if err != nil {
			return "", err
		}

		return docData.GuidFixed, nil

	} else {
		docData, err := svc.create(shopID, authUsername, doc)

		if err != nil {
			return "", err
		}

		return docData.GuidFixed, nil
	}
}

func (svc ProductGroupHttpService) CreateProductGroup(shopID string, authUsername string, doc models.ProductGroup) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if findDoc.Code != "" {
		return "", errors.New("code is exists")
	}

	docData, err := svc.create(shopID, authUsername, doc)

	go func() {
		svc.repoMessageQueue.Create(docData)

		svc.saveMasterSync(shopID)
	}()

	return docData.GuidFixed, err
}

func (svc ProductGroupHttpService) create(shopID string, authUsername string, doc models.ProductGroup) (models.ProductGroupDoc, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := utils.NewGUID()

	docData := models.ProductGroupDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.ProductGroup = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err := svc.repo.Create(ctx, docData)

	if err != nil {
		return models.ProductGroupDoc{}, err
	}

	return docData, nil
}

func (svc ProductGroupHttpService) UpdateProductGroup(shopID string, guid string, authUsername string, doc models.ProductGroup) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.ProductGroup = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	docData, err := svc.update(shopID, authUsername, guid, findDoc, doc)

	if err != nil {
		return err
	}

	go func() {
		err := svc.repoMessageQueue.Update(docData)

		if err != nil {
			logger.GetLogger().Error(err)
		}

		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc ProductGroupHttpService) update(shopID string, authUsername string, guid string, findDoc models.ProductGroupDoc, docUpdate models.ProductGroup) (models.ProductGroupDoc, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docData := findDoc

	docData.ProductGroup = docUpdate

	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	err := svc.repo.Update(ctx, shopID, guid, docData)

	if err != nil {
		return models.ProductGroupDoc{}, err
	}

	return docData, nil
}

func (svc ProductGroupHttpService) DeleteProductGroup(shopID, guid, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return nil
	}

	existsInProduct, _ := svc.existsGroupRefInProduct(shopID, []string{findDoc.Code})

	if existsInProduct {
		return fmt.Errorf("group code \"%s\" is referenced in product barcode", findDoc.Code)
	}

	err = svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)
	if err != nil {
		return err
	}

	go func() {
		err := svc.repoMessageQueue.Delete(findDoc)

		if err != nil {
			logger.GetLogger().Error(err)
		}

		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc ProductGroupHttpService) DeleteProductGroupByGUIDs(shopID, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDocs, err := svc.repo.FindByGuids(ctx, shopID, GUIDs)

	if err != nil {
		return err
	}

	if len(findDocs) == 0 {
		return nil
	}

	groupCodes := []string{}
	for _, v := range findDocs {
		groupCodes = append(groupCodes, v.Code)
	}

	existsInProduct, _ := svc.existsGroupRefInProduct(shopID, groupCodes)

	if existsInProduct {
		return fmt.Errorf("referenced in product")
	}

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err = svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	go func() {
		err := svc.repoMessageQueue.DeleteInBatch(findDocs)

		if err != nil {
			logger.GetLogger().Error(err)
		}

		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc ProductGroupHttpService) InfoProductGroup(shopID string, guid string) (models.ProductGroupInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.ProductGroupInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.ProductGroupInfo{}, errors.New("document not found")
	}

	return findDoc.ProductGroupInfo, nil

}

func (svc ProductGroupHttpService) InfoWTFArray(shopID string, codes []string) ([]interface{}, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList := []interface{}{}

	for _, code := range codes {
		findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)
		if err != nil || findDoc.ID == primitive.NilObjectID {
			// add item empty
			docList = append(docList, nil)
		} else {
			docList = append(docList, findDoc.ProductGroupInfo)
		}
	}

	return docList, nil
}

func (svc ProductGroupHttpService) SearchProductGroup(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductGroupInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.ProductGroupInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc ProductGroupHttpService) SearchProductGroupStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.ProductGroupInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.ProductGroupInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc ProductGroupHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.ProductGroup) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.ProductGroup](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.ProductGroup, models.ProductGroupDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.ProductGroup) models.ProductGroupDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.ProductGroupDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.ProductGroup = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.ProductGroup, models.ProductGroupDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.ProductGroupDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", guid)
		},
		func(doc models.ProductGroupDoc) bool {
			return doc.Code != ""
		},
		func(shopID string, authUsername string, data models.ProductGroup, doc models.ProductGroupDoc) error {

			doc.ProductGroup = data
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

func (svc ProductGroupHttpService) getDocIDKey(doc models.ProductGroup) string {
	return doc.Code
}

func (svc ProductGroupHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc ProductGroupHttpService) GetModuleName() string {
	return "productGroup"
}

func (svc ProductGroupHttpService) existsGroupRefInProduct(shopID string, groupCodes []string) (bool, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docCount, err := svc.repoProductBarcode.CountByGroupCodes(ctx, shopID, groupCodes)

	if err != nil {
		return true, err
	}

	if docCount > 0 {
		return true, fmt.Errorf("referenced in product barcode")
	}

	return false, nil
}
