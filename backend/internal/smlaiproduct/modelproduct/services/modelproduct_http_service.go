package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/smlaiproduct/modelproduct/models"
	"smlcloudplatform/internal/smlaiproduct/modelproduct/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IModelProductHttpService interface {
	CreateModelProduct(holdingCode string, authUsername string, doc models.ModelProduct) (string, error)
	UpdateModelProduct(holdingCode string, guid string, authUsername string, doc models.ModelProduct) error
	DeleteModelProduct(holdingCode string, guid string, authUsername string) error
	DeleteModelProductByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoModelProduct(holdingCode string, guid string) (models.ModelProductInfo, error)
	InfoModelProductByCode(holdingCode string, code string) (models.ModelProductInfo, error)
	SearchModelProduct(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ModelProductInfo, mongopagination.PaginationData, error)
	SearchModelProductStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.ModelProductInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.ModelProduct) (common.BulkImport, error)

	GetModuleName() string
}

type ModelProductHttpService struct {
	repo               repositories.IModelProductRepository
	repoProductBarcode productbarcode_repositories.IProductBarcodeRepository
	syncCacheRepo      mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.ModelProductActivity, models.ModelProductDeleteActivity]
	contextTimeout time.Duration
}

func NewModelProductHttpService(
	repo repositories.IModelProductRepository,
	repoProductBarcode productbarcode_repositories.IProductBarcodeRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	contextTimeout time.Duration,
) *ModelProductHttpService {

	insSvc := &ModelProductHttpService{
		repo:               repo,
		repoProductBarcode: repoProductBarcode,

		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.ModelProductActivity, models.ModelProductDeleteActivity](repo)

	return insSvc
}

func (svc ModelProductHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ModelProductHttpService) CreateModelProduct(holdingCode string, authUsername string, doc models.ModelProduct) (string, error) {

	// Business Code Uppercase + No-Space rules: the backend normalizes codes
	// itself (never trust the client) so "test br7854" persists as "TESTBR7854"
	// and duplicate checks compare the same normalized form.
	doc.Code = utils.NormalizeBusinessCode(doc.Code)
	if doc.Code == "" {
		return "", errors.New("code is required")
	}

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

	docData := models.ModelProductDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.ModelProduct = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc ModelProductHttpService) UpdateModelProduct(holdingCode string, guid string, authUsername string, doc models.ModelProduct) error {

	// Same business-code normalization as Create (uppercase, no whitespace).
	doc.Code = utils.NormalizeBusinessCode(doc.Code)
	if doc.Code == "" {
		return errors.New("code is required")
	}

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

	docData.ModelProduct = doc

	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, docData)

	if err != nil {
		return err
	}

	return nil
}

func (svc ModelProductHttpService) DeleteModelProduct(holdingCode string, guid string, authUsername string) error {

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

	return nil
}

func (svc ModelProductHttpService) DeleteModelProductByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

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

	return nil
}

func (svc ModelProductHttpService) InfoModelProduct(holdingCode string, guid string) (models.ModelProductInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.ModelProductInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.ModelProductInfo{}, errors.New("document not found")
	}

	return findDoc.ModelProductInfo, nil
}

func (svc ModelProductHttpService) InfoModelProductByCode(holdingCode string, code string) (models.ModelProductInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)

	if err != nil {
		return models.ModelProductInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.ModelProductInfo{}, errors.New("document not found")
	}

	return findDoc.ModelProductInfo, nil
}

func (svc ModelProductHttpService) SearchModelProduct(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ModelProductInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.ModelProductInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc ModelProductHttpService) SearchModelProductStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.ModelProductInfo, int, error) {

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
		return []models.ModelProductInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc ModelProductHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.ModelProduct) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.ModelProduct](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.ModelProduct, models.ModelProductDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.ModelProduct) models.ModelProductDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.ModelProductDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.ModelProduct = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.ModelProduct, models.ModelProductDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.ModelProductDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.ModelProductDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.ModelProduct, doc models.ModelProductDoc) error {

			doc.ModelProduct = data
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

func (svc ModelProductHttpService) getDocIDKey(doc models.ModelProduct) string {
	return doc.Code
}

func (svc ModelProductHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc ModelProductHttpService) GetModuleName() string {
	return "producttype"
}

func (svc ModelProductHttpService) existsOrderTypeRefInProduct(holdingCode string, GUIDs []string) (bool, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docCount, err := svc.repoProductBarcode.CountByModelProducts(ctx, holdingCode, GUIDs)
	if err != nil {
		return true, err
	}

	if docCount > 0 {
		return true, fmt.Errorf("referenced in product barcode")
	}

	return false, nil
}
