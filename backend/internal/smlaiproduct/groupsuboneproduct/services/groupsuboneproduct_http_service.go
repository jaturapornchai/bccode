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
	"smlcloudplatform/internal/smlaiproduct/groupsuboneproduct/models"
	"smlcloudplatform/internal/smlaiproduct/groupsuboneproduct/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IGroupsuboneProductHttpService interface {
	CreateGroupsuboneProduct(holdingCode string, authUsername string, doc models.GroupsuboneProduct) (string, error)
	UpdateGroupsuboneProduct(holdingCode string, guid string, authUsername string, doc models.GroupsuboneProduct) error
	DeleteGroupsuboneProduct(holdingCode string, guid string, authUsername string) error
	DeleteGroupsuboneProductByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoGroupsuboneProduct(holdingCode string, guid string) (models.GroupsuboneProductInfo, error)
	InfoGroupsuboneProductByCode(holdingCode string, code string) (models.GroupsuboneProductInfo, error)
	SearchGroupsuboneProduct(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.GroupsuboneProductInfo, mongopagination.PaginationData, error)
	SearchGroupsuboneProductStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.GroupsuboneProductInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.GroupsuboneProduct) (common.BulkImport, error)

	GetModuleName() string
}

type GroupsuboneProductHttpService struct {
	repo               repositories.IGroupsuboneProductRepository
	repoGroup          groupRepositories.IGroupProductRepository
	repoProductBarcode productbarcode_repositories.IProductBarcodeRepository
	syncCacheRepo      mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.GroupsuboneProductActivity, models.GroupsuboneProductDeleteActivity]
	contextTimeout time.Duration
}

func NewGroupsuboneProductHttpService(
	repo repositories.IGroupsuboneProductRepository,
	repoGroup groupRepositories.IGroupProductRepository,
	repoProductBarcode productbarcode_repositories.IProductBarcodeRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	contextTimeout time.Duration,
) *GroupsuboneProductHttpService {

	insSvc := &GroupsuboneProductHttpService{
		repo:               repo,
		repoGroup:          repoGroup,
		repoProductBarcode: repoProductBarcode,

		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.GroupsuboneProductActivity, models.GroupsuboneProductDeleteActivity](repo)

	return insSvc
}

func (svc GroupsuboneProductHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc GroupsuboneProductHttpService) CreateGroupsuboneProduct(holdingCode string, authUsername string, doc models.GroupsuboneProduct) (string, error) {

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

	docData := models.GroupsuboneProductDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.GroupsuboneProduct = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc GroupsuboneProductHttpService) UpdateGroupsuboneProduct(holdingCode string, guid string, authUsername string, doc models.GroupsuboneProduct) error {

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

	docData.GroupsuboneProduct = doc

	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, docData)

	if err != nil {
		return err
	}

	return nil
}

func (svc GroupsuboneProductHttpService) DeleteGroupsuboneProduct(holdingCode string, guid string, authUsername string) error {

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

func (svc GroupsuboneProductHttpService) DeleteGroupsuboneProductByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

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

func (svc GroupsuboneProductHttpService) InfoGroupsuboneProduct(holdingCode string, guid string) (models.GroupsuboneProductInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.GroupsuboneProductInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.GroupsuboneProductInfo{}, errors.New("document not found")
	}

	if findDoc.GroupMainGuid != "" {
		findGroupName, err := svc.repoGroup.FindByGuid(ctx, holdingCode, findDoc.GroupMainGuid)
		if err != nil {
			return models.GroupsuboneProductInfo{}, err
		}

		findDoc.GroupMainNames = findGroupName.Names
	}

	return findDoc.GroupsuboneProductInfo, nil
}

func (svc GroupsuboneProductHttpService) InfoGroupsuboneProductByCode(holdingCode string, code string) (models.GroupsuboneProductInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)

	if err != nil {
		return models.GroupsuboneProductInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.GroupsuboneProductInfo{}, errors.New("document not found")
	}

	if findDoc.GroupMainGuid != "" {
		findGroupName, err := svc.repoGroup.FindByGuid(ctx, holdingCode, findDoc.GroupMainGuid)
		if err != nil {
			return models.GroupsuboneProductInfo{}, err
		}

		findDoc.GroupMainNames = findGroupName.Names
	}

	return findDoc.GroupsuboneProductInfo, nil
}

func (svc GroupsuboneProductHttpService) SearchGroupsuboneProduct(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.GroupsuboneProductInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.GroupsuboneProductInfo{}, pagination, err
	}

	for idx, doc := range docList {
		if doc.GroupMainGuid != "" {
			findGroupName, err := svc.repoGroup.FindByGuid(ctx, holdingCode, doc.GroupMainGuid)
			if err != nil {
				return []models.GroupsuboneProductInfo{}, pagination, err
			}

			docList[idx].GroupMainNames = findGroupName.Names
		}
	}

	return docList, pagination, nil
}

func (svc GroupsuboneProductHttpService) SearchGroupsuboneProductStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.GroupsuboneProductInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.GroupsuboneProductInfo{}, 0, err
	}

	for idx, doc := range docList {
		if doc.GroupMainGuid != "" {
			findGroupName, err := svc.repoGroup.FindByGuid(ctx, holdingCode, doc.GroupMainGuid)
			if err != nil {
				return []models.GroupsuboneProductInfo{}, total, err
			}

			docList[idx].GroupMainNames = findGroupName.Names
		}
	}

	return docList, total, nil
}

func (svc GroupsuboneProductHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.GroupsuboneProduct) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.GroupsuboneProduct](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.GroupsuboneProduct, models.GroupsuboneProductDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.GroupsuboneProduct) models.GroupsuboneProductDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.GroupsuboneProductDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.GroupsuboneProduct = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.GroupsuboneProduct, models.GroupsuboneProductDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.GroupsuboneProductDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.GroupsuboneProductDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.GroupsuboneProduct, doc models.GroupsuboneProductDoc) error {

			doc.GroupsuboneProduct = data
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

func (svc GroupsuboneProductHttpService) getDocIDKey(doc models.GroupsuboneProduct) string {
	return doc.Code
}

func (svc GroupsuboneProductHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc GroupsuboneProductHttpService) GetModuleName() string {
	return "producttype"
}

func (svc GroupsuboneProductHttpService) existsOrderTypeRefInProduct(holdingCode string, GUIDs []string) (bool, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docCount, err := svc.repoProductBarcode.CountByGroupsuboneProducts(ctx, holdingCode, GUIDs)
	if err != nil {
		return true, err
	}

	if docCount > 0 {
		return true, fmt.Errorf("referenced in product barcode")
	}

	return false, nil
}
