package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/debtaccount/creditorgroup/models"
	"smlcloudplatform/internal/debtaccount/creditorgroup/repositories"
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

type ICreditorGroupHttpService interface {
	CreateCreditorGroup(holdingCode string, authUsername string, doc models.CreditorGroup) (string, error)
	UpdateCreditorGroup(holdingCode string, guid string, authUsername string, doc models.CreditorGroup) error
	DeleteCreditorGroup(holdingCode string, guid string, authUsername string) error
	DeleteCreditorGroupByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoCreditorGroup(holdingCode string, guid string) (models.CreditorGroupInfo, error)
	SearchCreditorGroup(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CreditorGroupInfo, mongopagination.PaginationData, error)
	SearchCreditorGroupStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.CreditorGroupInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.CreditorGroup) (common.BulkImport, error)

	GetModuleName() string
}

type CreditorGroupHttpService struct {
	repo          repositories.ICreditorGroupRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.CreditorGroupActivity, models.CreditorGroupDeleteActivity]
	contextTimeout time.Duration
}

func NewCreditorGroupHttpService(repo repositories.ICreditorGroupRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *CreditorGroupHttpService {
	contextTimeout := time.Duration(15) * time.Second

	insSvc := &CreditorGroupHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.CreditorGroupActivity, models.CreditorGroupDeleteActivity](repo)

	return insSvc
}

func (svc CreditorGroupHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc CreditorGroupHttpService) CreateCreditorGroup(holdingCode string, authUsername string, doc models.CreditorGroup) (string, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "group_code", doc.GroupCode)

	if err != nil {
		return "", err
	}

	if findDoc.GroupCode != "" {
		return "", errors.New("GroupCode is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.CreditorGroupDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.CreditorGroup = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)

	return newGuidFixed, nil
}

func (svc CreditorGroupHttpService) UpdateCreditorGroup(holdingCode string, guid string, authUsername string, doc models.CreditorGroup) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.CreditorGroup = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc CreditorGroupHttpService) DeleteCreditorGroup(holdingCode string, guid string, authUsername string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()
	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)
	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc CreditorGroupHttpService) DeleteCreditorGroupByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {
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

func (svc CreditorGroupHttpService) InfoCreditorGroup(holdingCode string, guid string) (models.CreditorGroupInfo, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.CreditorGroupInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.CreditorGroupInfo{}, errors.New("document not found")
	}

	return findDoc.CreditorGroupInfo, nil

}

func (svc CreditorGroupHttpService) SearchCreditorGroup(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CreditorGroupInfo, mongopagination.PaginationData, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"group_code",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.CreditorGroupInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc CreditorGroupHttpService) SearchCreditorGroupStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.CreditorGroupInfo, int, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"group_code",
		"names.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.CreditorGroupInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc CreditorGroupHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.CreditorGroup) (common.BulkImport, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.CreditorGroup](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.GroupCode)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, holdingCode, "group_code", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.GroupCode)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.CreditorGroup, models.CreditorGroupDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.CreditorGroup) models.CreditorGroupDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.CreditorGroupDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.CreditorGroup = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.CreditorGroup, models.CreditorGroupDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.CreditorGroupDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "group_code", guid)
		},
		func(doc models.CreditorGroupDoc) bool {
			return doc.GroupCode != ""
		},
		func(holdingCode string, authUsername string, data models.CreditorGroup, doc models.CreditorGroupDoc) error {

			doc.CreditorGroup = data
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

	svc.saveMasterSync(holdingCode)

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc CreditorGroupHttpService) getDocIDKey(doc models.CreditorGroup) string {
	return doc.GroupCode
}

func (svc CreditorGroupHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc CreditorGroupHttpService) GetModuleName() string {
	return "creditorGroup"
}
