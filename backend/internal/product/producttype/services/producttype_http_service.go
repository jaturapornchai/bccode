package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/logger"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/product/producttype/models"
	"smlcloudplatform/internal/product/producttype/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IProductTypeHttpService interface {
	CreateProductType(holdingCode string, authUsername string, doc models.ProductType) (string, error)
	UpdateProductType(holdingCode string, guid string, authUsername string, doc models.ProductType) error
	DeleteProductType(holdingCode string, guid string, authUsername string) error
	DeleteProductTypeByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoProductType(holdingCode string, guid string) (models.ProductTypeInfo, error)
	InfoProductTypeByCode(holdingCode string, code string) (models.ProductTypeInfo, error)
	SearchProductType(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductTypeInfo, mongopagination.PaginationData, error)
	SearchProductTypeStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.ProductTypeInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.ProductType) (common.BulkImport, error)

	GetModuleName() string
}

type ProductTypeHttpService struct {
	repo               repositories.IProductTypeRepository
	repoProductBarcode productbarcode_repositories.IProductBarcodeRepository
	repoMessageQueue   repositories.IProductTypeMessageQueueRepository
	syncCacheRepo      mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.ProductTypeActivity, models.ProductTypeDeleteActivity]
	contextTimeout time.Duration
}

func NewProductTypeHttpService(
	repo repositories.IProductTypeRepository,
	repoProductBarcode productbarcode_repositories.IProductBarcodeRepository,
	repoMessageQueue repositories.IProductTypeMessageQueueRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	contextTimeout time.Duration,
) *ProductTypeHttpService {

	insSvc := &ProductTypeHttpService{
		repo:               repo,
		repoProductBarcode: repoProductBarcode,
		repoMessageQueue:   repoMessageQueue,
		syncCacheRepo:      syncCacheRepo,
		contextTimeout:     contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.ProductTypeActivity, models.ProductTypeDeleteActivity](repo)

	return insSvc
}

func (svc ProductTypeHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ProductTypeHttpService) CreateProductType(holdingCode string, authUsername string, doc models.ProductType) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", errors.New("code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.ProductTypeDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.ProductType = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	go func() {
		err := svc.repoMessageQueue.Create(docData)

		if err != nil {
			logger.GetLogger().Error(err)
		}

		svc.saveMasterSync(holdingCode)
	}()

	return newGuidFixed, nil
}

func (svc ProductTypeHttpService) UpdateProductType(holdingCode string, guid string, authUsername string, doc models.ProductType) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	docData := findDoc

	docData.ProductType = doc

	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, docData)

	if err != nil {
		return err
	}

	go func() {
		err := svc.repoMessageQueue.Update(docData)

		if err != nil {
			logger.GetLogger().Error(err)
		}

		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc ProductTypeHttpService) DeleteProductType(holdingCode string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	existsInProduct, _ := svc.existsOrderTypeRefInProduct(holdingCode, []string{guid})

	if existsInProduct {
		return fmt.Errorf("\"%s\" is referenced in product barcode", findDoc.Code)
	}

	err = svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)
	if err != nil {
		return err
	}

	go func() {
		err := svc.repoMessageQueue.Delete(findDoc)

		if err != nil {
			logger.GetLogger().Error(err)
		}

		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc ProductTypeHttpService) DeleteProductTypeByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	existsInProduct, _ := svc.existsOrderTypeRefInProduct(holdingCode, GUIDs)

	if existsInProduct {
		return fmt.Errorf("referenced in product")
	}

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	go func() {
		findDocs, err := svc.repo.FindByGuids(ctx, holdingCode, GUIDs)

		if err != nil {
			logger.GetLogger().Error(err)
		} else {
			err = svc.repoMessageQueue.DeleteInBatch(findDocs)

			if err != nil {
				logger.GetLogger().Error(err)
			}
		}

		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc ProductTypeHttpService) InfoProductType(holdingCode string, guid string) (models.ProductTypeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.ProductTypeInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.ProductTypeInfo{}, errors.New("document not found")
	}

	return findDoc.ProductTypeInfo, nil
}

func (svc ProductTypeHttpService) InfoProductTypeByCode(holdingCode string, code string) (models.ProductTypeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)

	if err != nil {
		return models.ProductTypeInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.ProductTypeInfo{}, errors.New("document not found")
	}

	return findDoc.ProductTypeInfo, nil
}

func (svc ProductTypeHttpService) SearchProductType(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductTypeInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.ProductTypeInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc ProductTypeHttpService) SearchProductTypeStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.ProductTypeInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	selectFields := map[string]interface{}{}

	/*
		if langCode != "" {
			selectFields["names"] = bson.M{"$elemMatch": bson.M{"code": langCode}}
		} else {
			selectFields["names"] = 1
		}
	*/

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.ProductTypeInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc ProductTypeHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.ProductType) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.ProductType](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.ProductType, models.ProductTypeDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.ProductType) models.ProductTypeDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.ProductTypeDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.ProductType = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.ProductType, models.ProductTypeDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.ProductTypeDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.ProductTypeDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.ProductType, doc models.ProductTypeDoc) error {

			doc.ProductType = data
			doc.UpdatedBy = authUsername
			doc.UpdatedAt = time.Now()

			err = svc.repo.Update(ctx, holdingCode, doc.GuidFixed, doc)
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

	svc.saveMasterSync(holdingCode)

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc ProductTypeHttpService) getDocIDKey(doc models.ProductType) string {
	return doc.Code
}

func (svc ProductTypeHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc ProductTypeHttpService) GetModuleName() string {
	return "producttype"
}

func (svc ProductTypeHttpService) existsOrderTypeRefInProduct(holdingCode string, GUIDs []string) (bool, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docCount, err := svc.repoProductBarcode.CountByProductTypes(ctx, holdingCode, GUIDs)
	if err != nil {
		return true, err
	}

	if docCount > 0 {
		return true, fmt.Errorf("referenced in product barcode")
	}

	return false, nil
}
