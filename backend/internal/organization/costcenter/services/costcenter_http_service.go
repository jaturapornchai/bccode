package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/organization/costcenter/models"
	"smlcloudplatform/internal/organization/costcenter/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ICostCenterHttpService interface {
	CreateCostCenter(holdingCode string, authUsername string, doc models.CostCenter) (string, error)
	UpdateCostCenter(holdingCode string, guid string, authUsername string, doc models.CostCenter) error
	DeleteCostCenter(holdingCode string, guid string, authUsername string) error
	DeleteCostCenterByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoCostCenter(holdingCode string, guid string) (models.CostCenterInfo, error)
	InfoCostCenterByCode(holdingCode, branchCode, costCenterCode string) (models.CostCenterInfo, error)
	SearchCostCenter(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CostCenterInfo, mongopagination.PaginationData, error)
	SearchCostCenterStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.CostCenterInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.CostCenter) (common.BulkImport, error)

	GetModuleName() string
}

type CostCenterHttpService struct {
	repo             repositories.ICostCenterRepository
	repoMessageQueue repositories.ICostCenterMessageQueueRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.CostCenterActivity, models.CostCenterDeleteActivity]
	contextTimeout time.Duration
}

func NewCostCenterHttpService(repo repositories.ICostCenterRepository, repoMessageQueue repositories.ICostCenterMessageQueueRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *CostCenterHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &CostCenterHttpService{
		repo:             repo,
		repoMessageQueue: repoMessageQueue,
		syncCacheRepo:    syncCacheRepo,
		contextTimeout:   contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.CostCenterActivity, models.CostCenterDeleteActivity](repo)

	return insSvc
}

func (svc CostCenterHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc CostCenterHttpService) CreateCostCenter(holdingCode string, authUsername string, doc models.CostCenter) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if findDoc.Code != "" {
		return "", errors.New("Code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.CostCenterDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.CostCenter = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	go func() {
		svc.repoMessageQueue.Create(docData)
		svc.saveMasterSync(holdingCode)
	}()

	return newGuidFixed, nil
}

func (svc CostCenterHttpService) UpdateCostCenter(holdingCode string, guid string, authUsername string, doc models.CostCenter) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	findDoc.CostCenter = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return err
	}

	go func() {
		svc.repoMessageQueue.Update(findDoc)
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc CostCenterHttpService) DeleteCostCenter(holdingCode string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)
	if err != nil {
		return err
	}

	go func() {
		svc.repoMessageQueue.Delete(findDoc)
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc CostCenterHttpService) DeleteCostCenterByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guid_fixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	return nil
}

func (svc CostCenterHttpService) InfoCostCenter(holdingCode string, guid string) (models.CostCenterInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.CostCenterInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.CostCenterInfo{}, errors.New("document not found")
	}

	return findDoc.CostCenterInfo, nil
}

func (svc CostCenterHttpService) InfoCostCenterByCode(holdingCode, branchCode, costCenterCode string) (models.CostCenterInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindOneByCode(ctx, holdingCode, branchCode, costCenterCode)

	if err != nil {
		return models.CostCenterInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.CostCenterInfo{}, errors.New("document not found")
	}

	return findDoc.CostCenterInfo, nil
}

func (svc CostCenterHttpService) SearchCostCenter(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CostCenterInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.CostCenterInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc CostCenterHttpService) SearchCostCenterStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.CostCenterInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.CostCenterInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc CostCenterHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.CostCenter) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.CostCenter](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.CostCenter, models.CostCenterDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.CostCenter) models.CostCenterDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.CostCenterDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.CostCenter = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.CostCenter, models.CostCenterDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.CostCenterDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.CostCenterDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.CostCenter, doc models.CostCenterDoc) error {

			doc.CostCenter = data
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

func (svc CostCenterHttpService) getDocIDKey(doc models.CostCenter) string {
	return doc.Code
}

func (svc CostCenterHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc CostCenterHttpService) GetModuleName() string {
	return "costcenter"
}
