package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/smlaiproduct/groupproduct/models"
	"smlcloudplatform/internal/smlaiproduct/groupproduct/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IGroupProductHttpService interface {
	CreateGroupProduct(holdingCode string, authUsername string, doc models.GroupProduct) (string, error)
	UpdateGroupProduct(holdingCode string, guid string, authUsername string, doc models.GroupProduct) error
	DeleteGroupProduct(holdingCode string, guid string, authUsername string) error
	DeleteGroupProductByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoGroupProduct(holdingCode string, guid string) (models.GroupProductInfo, error)
	InfoGroupProductByCode(holdingCode string, code string) (models.GroupProductInfo, error)
	SearchGroupProduct(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.GroupProductInfo, mongopagination.PaginationData, error)
	SearchGroupProductStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.GroupProductInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.GroupProduct) (common.BulkImport, error)

	GetModuleName() string
}

type GroupProductHttpService struct {
	repo               repositories.IGroupProductRepository
	repoProductBarcode productbarcode_repositories.IProductBarcodeRepository
	syncCacheRepo      mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.GroupProductActivity, models.GroupProductDeleteActivity]
	contextTimeout time.Duration
}

func NewGroupProductHttpService(
	repo repositories.IGroupProductRepository,
	repoProductBarcode productbarcode_repositories.IProductBarcodeRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	contextTimeout time.Duration,
) *GroupProductHttpService {

	insSvc := &GroupProductHttpService{
		repo:               repo,
		repoProductBarcode: repoProductBarcode,

		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.GroupProductActivity, models.GroupProductDeleteActivity](repo)

	return insSvc
}

func (svc GroupProductHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc GroupProductHttpService) CreateGroupProduct(holdingCode string, authUsername string, doc models.GroupProduct) (string, error) {

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

	docData := models.GroupProductDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.GroupProduct = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc GroupProductHttpService) UpdateGroupProduct(holdingCode string, guid string, authUsername string, doc models.GroupProduct) error {

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

	docData.GroupProduct = doc

	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, docData)

	if err != nil {
		return err
	}

	return nil
}

func (svc GroupProductHttpService) DeleteGroupProduct(holdingCode string, guid string, authUsername string) error {

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

func (svc GroupProductHttpService) DeleteGroupProductByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

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

func (svc GroupProductHttpService) InfoGroupProduct(holdingCode string, guid string) (models.GroupProductInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.GroupProductInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.GroupProductInfo{}, errors.New("document not found")
	}

	return findDoc.GroupProductInfo, nil
}

func (svc GroupProductHttpService) InfoGroupProductByCode(holdingCode string, code string) (models.GroupProductInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)

	if err != nil {
		return models.GroupProductInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.GroupProductInfo{}, errors.New("document not found")
	}

	return findDoc.GroupProductInfo, nil
}

func (svc GroupProductHttpService) SearchGroupProduct(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.GroupProductInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.GroupProductInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc GroupProductHttpService) SearchGroupProductStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.GroupProductInfo, int, error) {

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
		return []models.GroupProductInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc GroupProductHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.GroupProduct) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.GroupProduct](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.GroupProduct, models.GroupProductDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.GroupProduct) models.GroupProductDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.GroupProductDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.GroupProduct = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.GroupProduct, models.GroupProductDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.GroupProductDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.GroupProductDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.GroupProduct, doc models.GroupProductDoc) error {

			doc.GroupProduct = data
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

func (svc GroupProductHttpService) getDocIDKey(doc models.GroupProduct) string {
	return doc.Code
}

func (svc GroupProductHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc GroupProductHttpService) GetModuleName() string {
	return "producttype"
}

func (svc GroupProductHttpService) existsOrderTypeRefInProduct(holdingCode string, GUIDs []string) (bool, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docCount, err := svc.repoProductBarcode.CountByGroupProducts(ctx, holdingCode, GUIDs)
	if err != nil {
		return true, err
	}

	if docCount > 0 {
		return true, fmt.Errorf("referenced in product barcode")
	}

	return false, nil
}
