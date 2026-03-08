package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/smlaiproduct/categoryproduct/models"
	"smlcloudplatform/internal/smlaiproduct/categoryproduct/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ICategoryProductHttpService interface {
	CreateCategoryProduct(shopID string, authUsername string, doc models.CategoryProduct) (string, error)
	UpdateCategoryProduct(shopID string, guid string, authUsername string, doc models.CategoryProduct) error
	DeleteCategoryProduct(shopID string, guid string, authUsername string) error
	DeleteCategoryProductByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoCategoryProduct(shopID string, guid string) (models.CategoryProductInfo, error)
	InfoCategoryProductByCode(shopID string, code string) (models.CategoryProductInfo, error)
	SearchCategoryProduct(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CategoryProductInfo, mongopagination.PaginationData, error)
	SearchCategoryProductStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.CategoryProductInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.CategoryProduct) (common.BulkImport, error)

	GetModuleName() string
}

type CategoryProductHttpService struct {
	repo               repositories.ICategoryProductRepository
	repoProductBarcode productbarcode_repositories.IProductBarcodeRepository
	syncCacheRepo      mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.CategoryProductActivity, models.CategoryProductDeleteActivity]
	contextTimeout time.Duration
}

func NewCategoryProductHttpService(
	repo repositories.ICategoryProductRepository,
	repoProductBarcode productbarcode_repositories.IProductBarcodeRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	contextTimeout time.Duration,
) *CategoryProductHttpService {

	insSvc := &CategoryProductHttpService{
		repo:               repo,
		repoProductBarcode: repoProductBarcode,

		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.CategoryProductActivity, models.CategoryProductDeleteActivity](repo)

	return insSvc
}

func (svc CategoryProductHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc CategoryProductHttpService) CreateCategoryProduct(shopID string, authUsername string, doc models.CategoryProduct) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", errors.New("code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.CategoryProductDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.CategoryProduct = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc CategoryProductHttpService) UpdateCategoryProduct(shopID string, guid string, authUsername string, doc models.CategoryProduct) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	docData := findDoc

	docData.CategoryProduct = doc

	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, docData)

	if err != nil {
		return err
	}

	return nil
}

func (svc CategoryProductHttpService) DeleteCategoryProduct(shopID string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	existsInProduct, _ := svc.existsOrderTypeRefInProduct(shopID, []string{guid})

	if existsInProduct {
		return fmt.Errorf("\"%s\" is referenced in product barcode", findDoc.Code)
	}

	err = svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)
	if err != nil {
		return err
	}

	return nil
}

func (svc CategoryProductHttpService) DeleteCategoryProductByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	existsInProduct, _ := svc.existsOrderTypeRefInProduct(shopID, GUIDs)

	if existsInProduct {
		return fmt.Errorf("referenced in product")
	}

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	return nil
}

func (svc CategoryProductHttpService) InfoCategoryProduct(shopID string, guid string) (models.CategoryProductInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.CategoryProductInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.CategoryProductInfo{}, errors.New("document not found")
	}

	return findDoc.CategoryProductInfo, nil
}

func (svc CategoryProductHttpService) InfoCategoryProductByCode(shopID string, code string) (models.CategoryProductInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)

	if err != nil {
		return models.CategoryProductInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.CategoryProductInfo{}, errors.New("document not found")
	}

	return findDoc.CategoryProductInfo, nil
}

func (svc CategoryProductHttpService) SearchCategoryProduct(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CategoryProductInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.CategoryProductInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc CategoryProductHttpService) SearchCategoryProductStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.CategoryProductInfo, int, error) {

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

	docList, total, err := svc.repo.FindStep(ctx, shopID, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.CategoryProductInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc CategoryProductHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.CategoryProduct) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.CategoryProduct](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.CategoryProduct, models.CategoryProductDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.CategoryProduct) models.CategoryProductDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.CategoryProductDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.CategoryProduct = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.CategoryProduct, models.CategoryProductDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.CategoryProductDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", guid)
		},
		func(doc models.CategoryProductDoc) bool {
			return doc.Code != ""
		},
		func(shopID string, authUsername string, data models.CategoryProduct, doc models.CategoryProductDoc) error {

			doc.CategoryProduct = data
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

func (svc CategoryProductHttpService) getDocIDKey(doc models.CategoryProduct) string {
	return doc.Code
}

func (svc CategoryProductHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc CategoryProductHttpService) GetModuleName() string {
	return "productType"
}

func (svc CategoryProductHttpService) existsOrderTypeRefInProduct(shopID string, GUIDs []string) (bool, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docCount, err := svc.repoProductBarcode.CountByCategoryProducts(ctx, shopID, GUIDs)
	if err != nil {
		return true, err
	}

	if docCount > 0 {
		return true, fmt.Errorf("referenced in product barcode")
	}

	return false, nil
}
