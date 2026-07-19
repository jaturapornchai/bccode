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
	"strings"
	"time"

	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IProductGroupHttpService interface {
	SaveProductGroup(holdingCode string, authUsername string, doc models.ProductGroup) (string, error)
	CreateProductGroup(holdingCode string, authUsername string, doc models.ProductGroup) (string, error)
	UpdateProductGroup(holdingCode string, guid string, authUsername string, doc models.ProductGroup) error
	DeleteProductGroup(holdingCode string, guid string, authUsername string) error
	DeleteProductGroupByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoProductGroup(holdingCode string, guid string) (models.ProductGroupInfo, error)
	InfoWTFArray(holdingCode string, unitCodes []string) ([]interface{}, error)
	SearchProductGroup(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductGroupInfo, mongopagination.PaginationData, error)
	SearchProductGroupStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.ProductGroupInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.ProductGroup) (common.BulkImport, error)
	XSortsSave(holdingCode string, authUsername string, xsorts []common.XSortModifyReqesut) error

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

// A group must keep at least one non-empty name. The struct validator cannot
// catch this: `required` on NameX.Name (*string) only rejects nil pointers, so
// empty strings slip through and used to persist nameless groups.
func validateProductGroupHasName(doc models.ProductGroup) error {
	if doc.Names != nil {
		for _, name := range *doc.Names {
			if name.Name != nil && strings.TrimSpace(*name.Name) != "" {
				return nil
			}
		}
	}
	return errors.New("at least one product group name is required")
}

func (svc ProductGroupHttpService) SaveProductGroup(holdingCode string, authUsername string, doc models.ProductGroup) (string, error) {

	if err := validateProductGroupHasName(doc); err != nil {
		return "", err
	}

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if len(findDoc.Code) > 0 {

		docData, err := svc.update(holdingCode, authUsername, findDoc.GuidFixed, findDoc, doc)

		if err != nil {
			return "", err
		}

		return docData.GuidFixed, nil

	} else {
		docData, err := svc.create(holdingCode, authUsername, doc)

		if err != nil {
			return "", err
		}

		return docData.GuidFixed, nil
	}
}

func (svc ProductGroupHttpService) CreateProductGroup(holdingCode string, authUsername string, doc models.ProductGroup) (string, error) {

	if err := validateProductGroupHasName(doc); err != nil {
		return "", err
	}

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if findDoc.Code != "" {
		return "", errors.New("code is exists")
	}

	docData, err := svc.create(holdingCode, authUsername, doc)

	go func() {
		svc.repoMessageQueue.Create(docData)

		svc.saveMasterSync(holdingCode)
	}()

	return docData.GuidFixed, err
}

func (svc ProductGroupHttpService) create(holdingCode string, authUsername string, doc models.ProductGroup) (models.ProductGroupDoc, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := utils.NewGUID()

	docData := models.ProductGroupDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.ProductGroup = doc

	docData.EmptyOnNil()

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err := svc.repo.Create(ctx, docData)

	if err != nil {
		return models.ProductGroupDoc{}, err
	}

	return docData, nil
}

func (svc ProductGroupHttpService) UpdateProductGroup(holdingCode string, guid string, authUsername string, doc models.ProductGroup) error {

	if err := validateProductGroupHasName(doc); err != nil {
		return err
	}

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.ProductGroup = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	docData, err := svc.update(holdingCode, authUsername, guid, findDoc, doc)

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

func (svc ProductGroupHttpService) update(holdingCode string, authUsername string, guid string, findDoc models.ProductGroupDoc, docUpdate models.ProductGroup) (models.ProductGroupDoc, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docData := findDoc

	docData.ProductGroup = docUpdate

	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	err := svc.repo.Update(ctx, holdingCode, guid, docData)

	if err != nil {
		return models.ProductGroupDoc{}, err
	}

	return docData, nil
}

func (svc ProductGroupHttpService) DeleteProductGroup(holdingCode, guid, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return nil
	}

	existsInProduct, err := svc.existsGroupRefInProduct(
		holdingCode,
		[]string{findDoc.GuidFixed},
		[]string{findDoc.Code},
	)
	if err != nil {
		return err
	}

	if existsInProduct {
		return errors.New("ไม่สามารถลบกลุ่มสินค้าที่ถูกใช้งานในบาร์โค้ดสินค้า")
	}

	// Block deletion of a group that still has sub-groups so children are not
	// silently orphaned (their parentguid would dangle and re-root in the tree).
	childDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "parentguid", guid)
	if err != nil {
		return err
	}
	if childDoc.GuidFixed != "" {
		return errors.New("ไม่สามารถลบกลุ่มสินค้าที่มีกลุ่มย่อยได้ กรุณาย้ายหรือลบกลุ่มย่อยก่อน")
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

func (svc ProductGroupHttpService) DeleteProductGroupByGUIDs(holdingCode, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDocs, err := svc.repo.FindByGuids(ctx, holdingCode, GUIDs)

	if err != nil {
		return err
	}

	if len(findDocs) == 0 {
		return nil
	}

	groupCodes := []string{}
	groupGUIDs := []string{}
	for _, v := range findDocs {
		groupGUIDs = append(groupGUIDs, v.GuidFixed)
		groupCodes = append(groupCodes, v.Code)
	}

	existsInProduct, err := svc.existsGroupRefInProduct(holdingCode, groupGUIDs, groupCodes)
	if err != nil {
		return err
	}

	if existsInProduct {
		return errors.New("ไม่สามารถลบกลุ่มสินค้าที่ถูกใช้งานในบาร์โค้ดสินค้า")
	}

	// Block batch deletion when any selected group still has sub-groups, so
	// children are not silently orphaned. Caller must handle sub-groups first.
	for _, v := range findDocs {
		childDoc, errChild := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "parentguid", v.GuidFixed)
		if errChild != nil {
			return errChild
		}
		if childDoc.GuidFixed != "" {
			return errors.New("ไม่สามารถลบกลุ่มสินค้าที่มีกลุ่มย่อยได้ กรุณาย้ายหรือลบกลุ่มย่อยก่อน")
		}
	}

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err = svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	go func() {
		err := svc.repoMessageQueue.DeleteInBatch(findDocs)

		if err != nil {
			logger.GetLogger().Error(err)
		}

		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc ProductGroupHttpService) InfoProductGroup(holdingCode string, guid string) (models.ProductGroupInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.ProductGroupInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.ProductGroupInfo{}, errors.New("document not found")
	}

	return findDoc.ProductGroupInfo, nil

}

func (svc ProductGroupHttpService) InfoWTFArray(holdingCode string, codes []string) ([]interface{}, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList := []interface{}{}

	for _, code := range codes {
		findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)
		if err != nil || findDoc.ID == primitive.NilObjectID {
			// add item empty
			docList = append(docList, nil)
		} else {
			docList = append(docList, findDoc.ProductGroupInfo)
		}
	}

	return docList, nil
}

func (svc ProductGroupHttpService) SearchProductGroup(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductGroupInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.ProductGroupInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc ProductGroupHttpService) SearchProductGroupStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.ProductGroupInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.ProductGroupInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc ProductGroupHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.ProductGroup) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.ProductGroup](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.ProductGroup, models.ProductGroupDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.ProductGroup) models.ProductGroupDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.ProductGroupDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.ProductGroup = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.ProductGroup, models.ProductGroupDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.ProductGroupDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.ProductGroupDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.ProductGroup, doc models.ProductGroupDoc) error {

			doc.ProductGroup = data
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

// XSortsSave applies batch sibling-reorder requests: for each request it merges the
// new xorder into the target document's xsorts list (keyed by code) and persists it.
func (svc ProductGroupHttpService) XSortsSave(holdingCode string, authUsername string, xsorts []common.XSortModifyReqesut) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	for _, xsort := range xsorts {
		if len(xsort.GUIDFixed) < 1 {
			continue
		}
		findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, xsort.GUIDFixed)

		if err != nil {
			return err
		}

		if len(findDoc.GuidFixed) < 1 {
			continue
		}

		if findDoc.XSorts == nil {
			findDoc.XSorts = &[]common.XSort{}
		}

		dictXSorts := map[string]common.XSort{}

		for _, tempXSort := range *findDoc.XSorts {
			dictXSorts[tempXSort.Code] = tempXSort
		}

		dictXSorts[xsort.Code] = common.XSort{
			Code:   xsort.Code,
			XOrder: xsort.XOrder,
		}

		tempXSorts := []common.XSort{}

		for _, tempXSort := range dictXSorts {
			tempXSorts = append(tempXSorts, tempXSort)
		}

		findDoc.XSorts = &tempXSorts

		err = svc.repo.UpdateXSorts(ctx, holdingCode, findDoc.GuidFixed, tempXSorts, authUsername, time.Now())

		if err != nil {
			return err
		}
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc ProductGroupHttpService) getDocIDKey(doc models.ProductGroup) string {
	return doc.Code
}

func (svc ProductGroupHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc ProductGroupHttpService) GetModuleName() string {
	return "productgroup"
}

func (svc ProductGroupHttpService) existsGroupRefInProduct(holdingCode string, groupGUIDs, groupCodes []string) (bool, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	nonEmptyGUIDs := make([]string, 0, len(groupGUIDs))
	for _, guid := range groupGUIDs {
		if strings.TrimSpace(guid) != "" {
			nonEmptyGUIDs = append(nonEmptyGUIDs, guid)
		}
	}
	if len(nonEmptyGUIDs) > 0 {
		docCount, err := svc.repoProductBarcode.CountByGroupGUIDs(ctx, holdingCode, nonEmptyGUIDs)
		if err != nil {
			return false, err
		}
		if docCount > 0 {
			return true, nil
		}
	}

	nonEmptyCodes := make([]string, 0, len(groupCodes))
	for _, code := range groupCodes {
		if strings.TrimSpace(code) != "" {
			nonEmptyCodes = append(nonEmptyCodes, code)
		}
	}
	if len(nonEmptyCodes) > 0 {
		docCount, err := svc.repoProductBarcode.CountByGroupCodes(ctx, holdingCode, nonEmptyCodes)
		if err != nil {
			return false, err
		}
		if docCount > 0 {
			return true, nil
		}
	}

	return false, nil
}
