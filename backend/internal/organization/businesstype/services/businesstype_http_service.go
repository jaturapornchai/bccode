package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/organization/businesstype/models"
	"smlcloudplatform/internal/organization/businesstype/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IBusinessTypeHttpService interface {
	CreateBusinessType(shopID string, authUsername string, doc models.BusinessType) (string, error)
	UpdateBusinessType(shopID string, guid string, authUsername string, doc models.BusinessType) error
	DeleteBusinessType(shopID string, guid string, authUsername string) error
	DeleteBusinessTypeByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoBusinessType(shopID string, guid string) (models.BusinessTypeInfo, error)
	InfoBusinessTypeByCode(shopID string, code string) (models.BusinessTypeInfo, error)
	SearchBusinessType(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.BusinessTypeInfo, mongopagination.PaginationData, error)
	SearchBusinessTypeStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.BusinessTypeInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.BusinessType) (common.BulkImport, error)

	InfoBusinessTypeDefault(shopID string) (models.BusinessTypeInfo, error)

	GetModuleName() string
}

type BusinessTypeHttpService struct {
	repo repositories.IBusinessTypeRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.BusinessTypeActivity, models.BusinessTypeDeleteActivity]
	contextTimeout time.Duration
}

func NewBusinessTypeHttpService(repo repositories.IBusinessTypeRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *BusinessTypeHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &BusinessTypeHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.BusinessTypeActivity, models.BusinessTypeDeleteActivity](repo)

	return insSvc
}

func (svc BusinessTypeHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc BusinessTypeHttpService) CreateBusinessType(shopID string, authUsername string, doc models.BusinessType) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Check code is exists
	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if findDoc.Code != "" {
		return "", errors.New("code is exists")
	}

	// Create new GuidFixed
	newGuidFixed := utils.NewGUID()

	err = svc.repo.Transaction(ctx, func(tctx context.Context) error {

		if doc.IsDefault {
			err = svc.repo.ClearDefault(tctx, shopID)
			if err != nil {
				return err
			}
		}

		// Create new document
		docData := models.BusinessTypeDoc{}
		docData.ShopID = shopID
		docData.GuidFixed = newGuidFixed
		docData.BusinessType = doc

		docData.CreatedBy = authUsername
		docData.CreatedAt = time.Now()

		// Create document to database
		_, err = svc.repo.Create(ctx, docData)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// Save master sync
	svc.saveMasterSync(shopID)

	return newGuidFixed, nil
}
func (svc BusinessTypeHttpService) UpdateBusinessType(shopID string, guid string, authUsername string, doc models.BusinessType) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Find the business type by the given guid.
	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)
	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	err = svc.repo.Transaction(ctx, func(tctx context.Context) error {

		if doc.IsDefault {
			err = svc.repo.ClearDefault(tctx, shopID)
			if err != nil {
				return err
			}
		}

		// Update the document.
		findDoc.BusinessType = doc
		findDoc.UpdatedBy = authUsername
		findDoc.UpdatedAt = time.Now()

		// Save the updated document.
		err = svc.repo.Update(tctx, shopID, guid, findDoc)
		if err != nil {
			return err
		}

		return nil

	})

	if err != nil {
		return err
	}

	// Set the master sync.
	svc.saveMasterSync(shopID)

	return nil
}

func (svc BusinessTypeHttpService) DeleteBusinessType(shopID string, guid string, authUsername string) error {

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

func (svc BusinessTypeHttpService) DeleteBusinessTypeByGUIDs(shopID string, authUsername string, GUIDs []string) error {

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

func (svc BusinessTypeHttpService) InfoBusinessTypeDefault(shopID string) (models.BusinessTypeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.GetDefault(ctx, shopID)

	if err != nil {
		return models.BusinessTypeInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.BusinessTypeInfo{}, errors.New("document not found")
	}

	return findDoc.BusinessTypeInfo, nil
}

func (svc BusinessTypeHttpService) InfoBusinessType(shopID string, guid string) (models.BusinessTypeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.BusinessTypeInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.BusinessTypeInfo{}, errors.New("document not found")
	}

	return findDoc.BusinessTypeInfo, nil
}

func (svc BusinessTypeHttpService) InfoBusinessTypeByCode(shopID string, code string) (models.BusinessTypeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)

	if err != nil {
		return models.BusinessTypeInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.BusinessTypeInfo{}, errors.New("document not found")
	}

	return findDoc.BusinessTypeInfo, nil
}

func (svc BusinessTypeHttpService) SearchBusinessType(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.BusinessTypeInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.BusinessTypeInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc BusinessTypeHttpService) SearchBusinessTypeStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.BusinessTypeInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.BusinessTypeInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc BusinessTypeHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.BusinessType) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.BusinessType](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.BusinessType, models.BusinessTypeDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.BusinessType) models.BusinessTypeDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.BusinessTypeDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.BusinessType = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.BusinessType, models.BusinessTypeDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.BusinessTypeDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", guid)
		},
		func(doc models.BusinessTypeDoc) bool {
			return doc.Code != ""
		},
		func(shopID string, authUsername string, data models.BusinessType, doc models.BusinessTypeDoc) error {

			doc.BusinessType = data
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

func (svc BusinessTypeHttpService) getDocIDKey(doc models.BusinessType) string {
	return doc.Code
}

func (svc BusinessTypeHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc BusinessTypeHttpService) GetModuleName() string {
	return "businessType"
}
