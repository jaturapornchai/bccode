package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/services"
	groupRepositories "smlcloudplatform/internal/smlaiproduct/groupproduct/repositories"
	groupsuboneRepositories "smlcloudplatform/internal/smlaiproduct/groupsuboneproduct/repositories"
	"smlcloudplatform/internal/smlaiproduct/groupsubtwoproduct/models"
	"smlcloudplatform/internal/smlaiproduct/groupsubtwoproduct/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IGroupsubtwoProductHttpService interface {
	CreateGroupsubtwoProduct(holdingCode string, authUsername string, doc models.GroupsubtwoProduct) (string, error)
	UpdateGroupsubtwoProduct(holdingCode string, guid string, authUsername string, doc models.GroupsubtwoProduct) error
	DeleteGroupsubtwoProduct(holdingCode string, guid string, authUsername string) error
	DeleteGroupsubtwoProductByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoGroupsubtwoProduct(holdingCode string, guid string) (models.GroupsubtwoProductInfo, error)
	InfoGroupsubtwoProductByCode(holdingCode string, code string) (models.GroupsubtwoProductInfo, error)
	SearchGroupsubtwoProduct(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.GroupsubtwoProductInfo, mongopagination.PaginationData, error)
	SearchGroupsubtwoProductStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.GroupsubtwoProductInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.GroupsubtwoProduct) (common.BulkImport, error)

	GetModuleName() string
}

type GroupsubtwoProductHttpService struct {
	repo               repositories.IGroupsubtwoProductRepository
	repoGroup          groupRepositories.IGroupProductRepository
	repoGroupSubOne    groupsuboneRepositories.IGroupsuboneProductRepository
	repoProductBarcode productbarcode_repositories.IProductBarcodeRepository
	syncCacheRepo      mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.GroupsubtwoProductActivity, models.GroupsubtwoProductDeleteActivity]
	contextTimeout time.Duration
}

func NewGroupsubtwoProductHttpService(
	repo repositories.IGroupsubtwoProductRepository,
	repoGroup groupRepositories.IGroupProductRepository,
	repoGroupSubOne groupsuboneRepositories.IGroupsuboneProductRepository,
	repoProductBarcode productbarcode_repositories.IProductBarcodeRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	contextTimeout time.Duration,
) *GroupsubtwoProductHttpService {

	insSvc := &GroupsubtwoProductHttpService{
		repo:               repo,
		repoGroup:          repoGroup,
		repoGroupSubOne:    repoGroupSubOne,
		repoProductBarcode: repoProductBarcode,

		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.GroupsubtwoProductActivity, models.GroupsubtwoProductDeleteActivity](repo)

	return insSvc
}

func (svc GroupsubtwoProductHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc GroupsubtwoProductHttpService) CreateGroupsubtwoProduct(holdingCode string, authUsername string, doc models.GroupsubtwoProduct) (string, error) {

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

	docData := models.GroupsubtwoProductDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.GroupsubtwoProduct = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc GroupsubtwoProductHttpService) UpdateGroupsubtwoProduct(holdingCode string, guid string, authUsername string, doc models.GroupsubtwoProduct) error {

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

	docData.GroupsubtwoProduct = doc

	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, docData)

	if err != nil {
		return err
	}

	return nil
}

func (svc GroupsubtwoProductHttpService) DeleteGroupsubtwoProduct(holdingCode string, guid string, authUsername string) error {

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

func (svc GroupsubtwoProductHttpService) DeleteGroupsubtwoProductByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

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

func (svc GroupsubtwoProductHttpService) InfoGroupsubtwoProduct(holdingCode string, guid string) (models.GroupsubtwoProductInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.GroupsubtwoProductInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.GroupsubtwoProductInfo{}, errors.New("document not found")
	}

	if findDoc.GroupMainGuid != "" {
		findGroupName, err := svc.repoGroup.FindByGuid(ctx, holdingCode, findDoc.GroupMainGuid)
		if err != nil {
			return models.GroupsubtwoProductInfo{}, err
		}

		findDoc.GroupMainNames = findGroupName.Names
	}
	if findDoc.GroupSubGuid != "" {
		findGroupName, err := svc.repoGroupSubOne.FindByGuid(ctx, holdingCode, findDoc.GroupSubGuid)
		if err != nil {
			return models.GroupsubtwoProductInfo{}, err
		}
		findDoc.GroupSubNames = findGroupName.Names
	}

	return findDoc.GroupsubtwoProductInfo, nil
}

func (svc GroupsubtwoProductHttpService) InfoGroupsubtwoProductByCode(holdingCode string, code string) (models.GroupsubtwoProductInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)

	if err != nil {
		return models.GroupsubtwoProductInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.GroupsubtwoProductInfo{}, errors.New("document not found")
	}

	if findDoc.GroupMainGuid != "" {
		findGroupName, err := svc.repoGroup.FindByGuid(ctx, holdingCode, findDoc.GroupMainGuid)
		if err != nil {
			return models.GroupsubtwoProductInfo{}, err
		}

		findDoc.GroupMainNames = findGroupName.Names
	}
	if findDoc.GroupSubGuid != "" {
		findGroupName, err := svc.repoGroupSubOne.FindByGuid(ctx, holdingCode, findDoc.GroupSubGuid)
		if err != nil {
			return models.GroupsubtwoProductInfo{}, err
		}
		findDoc.GroupSubNames = findGroupName.Names
	}

	return findDoc.GroupsubtwoProductInfo, nil
}

func (svc GroupsubtwoProductHttpService) SearchGroupsubtwoProduct(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.GroupsubtwoProductInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.GroupsubtwoProductInfo{}, pagination, err
	}

	for idx, doc := range docList {
		if doc.GroupMainGuid != "" {
			findGroupName, err := svc.repoGroup.FindByGuid(ctx, holdingCode, doc.GroupMainGuid)
			if err != nil {
				return []models.GroupsubtwoProductInfo{}, pagination, err
			}

			docList[idx].GroupMainNames = findGroupName.Names
		}
		if doc.GroupSubGuid != "" {
			findGroupName, err := svc.repoGroupSubOne.FindByGuid(ctx, holdingCode, doc.GroupSubGuid)
			if err != nil {
				return []models.GroupsubtwoProductInfo{}, pagination, err
			}

			docList[idx].GroupSubNames = findGroupName.Names
		}

	}

	return docList, pagination, nil
}

func (svc GroupsubtwoProductHttpService) SearchGroupsubtwoProductStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.GroupsubtwoProductInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.GroupsubtwoProductInfo{}, 0, err
	}

	for idx, doc := range docList {
		if doc.GroupMainGuid != "" {
			findGroupName, err := svc.repoGroup.FindByGuid(ctx, holdingCode, doc.GroupMainGuid)
			if err != nil {
				return []models.GroupsubtwoProductInfo{}, total, err
			}

			docList[idx].GroupMainNames = findGroupName.Names
		}
		if doc.GroupSubGuid != "" {
			findGroupName, err := svc.repoGroupSubOne.FindByGuid(ctx, holdingCode, doc.GroupSubGuid)
			if err != nil {
				return []models.GroupsubtwoProductInfo{}, total, err
			}

			docList[idx].GroupSubNames = findGroupName.Names
		}
	}

	return docList, total, nil
}

func (svc GroupsubtwoProductHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.GroupsubtwoProduct) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.GroupsubtwoProduct](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.GroupsubtwoProduct, models.GroupsubtwoProductDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.GroupsubtwoProduct) models.GroupsubtwoProductDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.GroupsubtwoProductDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.GroupsubtwoProduct = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.GroupsubtwoProduct, models.GroupsubtwoProductDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.GroupsubtwoProductDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.GroupsubtwoProductDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.GroupsubtwoProduct, doc models.GroupsubtwoProductDoc) error {

			doc.GroupsubtwoProduct = data
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

func (svc GroupsubtwoProductHttpService) getDocIDKey(doc models.GroupsubtwoProduct) string {
	return doc.Code
}

func (svc GroupsubtwoProductHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc GroupsubtwoProductHttpService) GetModuleName() string {
	return "producttype"
}

func (svc GroupsubtwoProductHttpService) existsOrderTypeRefInProduct(holdingCode string, GUIDs []string) (bool, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docCount, err := svc.repoProductBarcode.CountByGroupsubtwoProducts(ctx, holdingCode, GUIDs)
	if err != nil {
		return true, err
	}

	if docCount > 0 {
		return true, fmt.Errorf("referenced in product barcode")
	}

	return false, nil
}
