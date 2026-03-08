package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/debtaccount/debtorgroup/models"
	"smlcloudplatform/internal/debtaccount/debtorgroup/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IDebtorGroupHttpService interface {
	CreateDebtorGroup(shopID string, authUsername string, doc models.DebtorGroup) (string, error)
	UpdateDebtorGroup(shopID string, guid string, authUsername string, doc models.DebtorGroup) error
	DeleteDebtorGroup(shopID string, guid string, authUsername string) error
	DeleteDebtorGroupByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoDebtorGroup(shopID string, guid string) (models.DebtorGroupInfo, error)
	SearchDebtorGroup(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DebtorGroupInfo, mongopagination.PaginationData, error)
	SearchDebtorGroupStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.DebtorGroupInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.DebtorGroup) (common.BulkImport, error)

	GetModuleName() string
}

type DebtorGroupHttpService struct {
	repo repositories.IDebtorGroupRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.DebtorGroupActivity, models.DebtorGroupDeleteActivity]
	contextTimeout time.Duration
}

func NewDebtorGroupHttpService(repo repositories.IDebtorGroupRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *DebtorGroupHttpService {
	contextTimeout := time.Duration(15) * time.Second

	insSvc := &DebtorGroupHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.DebtorGroupActivity, models.DebtorGroupDeleteActivity](repo)

	return insSvc
}

func (svc DebtorGroupHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc DebtorGroupHttpService) CreateDebtorGroup(shopID string, authUsername string, doc models.DebtorGroup) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "groupcode", doc.GroupCode)

	if err != nil {
		return "", err
	}

	if findDoc.GroupCode != "" {
		return "", errors.New("GroupCode is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.DebtorGroupDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.DebtorGroup = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(shopID)

	return newGuidFixed, nil
}

func (svc DebtorGroupHttpService) UpdateDebtorGroup(shopID string, guid string, authUsername string, doc models.DebtorGroup) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.DebtorGroup = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc DebtorGroupHttpService) DeleteDebtorGroup(shopID string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)
	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc DebtorGroupHttpService) DeleteDebtorGroupByGUIDs(shopID string, authUsername string, GUIDs []string) error {

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

func (svc DebtorGroupHttpService) InfoDebtorGroup(shopID string, guid string) (models.DebtorGroupInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.DebtorGroupInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.DebtorGroupInfo{}, errors.New("document not found")
	}

	return findDoc.DebtorGroupInfo, nil

}

func (svc DebtorGroupHttpService) SearchDebtorGroup(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DebtorGroupInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"groupcode",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.DebtorGroupInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc DebtorGroupHttpService) SearchDebtorGroupStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.DebtorGroupInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"groupcode",
		"names.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.DebtorGroupInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc DebtorGroupHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.DebtorGroup) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.DebtorGroup](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.GroupCode)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, shopID, "groupcode", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.GroupCode)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.DebtorGroup, models.DebtorGroupDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.DebtorGroup) models.DebtorGroupDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.DebtorGroupDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.DebtorGroup = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.DebtorGroup, models.DebtorGroupDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.DebtorGroupDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "groupcode", guid)
		},
		func(doc models.DebtorGroupDoc) bool {
			return doc.GroupCode != ""
		},
		func(shopID string, authUsername string, data models.DebtorGroup, doc models.DebtorGroupDoc) error {

			doc.DebtorGroup = data
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
		createDataKey = append(createDataKey, doc.GroupCode)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.GroupCode)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {

		updateDataKey = append(updateDataKey, doc.GroupCode)
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

func (svc DebtorGroupHttpService) getDocIDKey(doc models.DebtorGroup) string {
	return doc.GroupCode
}

func (svc DebtorGroupHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc DebtorGroupHttpService) GetModuleName() string {
	return "debtorGroup"
}
