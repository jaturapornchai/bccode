package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/debtaccount/customergroup/models"
	"smlcloudplatform/internal/debtaccount/customergroup/repositories"
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

type ICustomerGroupHttpService interface {
	CreateCustomerGroup(shopID string, authUsername string, doc models.CustomerGroup) (string, error)
	UpdateCustomerGroup(shopID string, guid string, authUsername string, doc models.CustomerGroup) error
	DeleteCustomerGroup(shopID string, guid string, authUsername string) error
	DeleteCustomerGroupByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoCustomerGroup(shopID string, guid string) (models.CustomerGroupInfo, error)
	SearchCustomerGroup(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CustomerGroupInfo, mongopagination.PaginationData, error)
	SearchCustomerGroupStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.CustomerGroupInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.CustomerGroup) (common.BulkImport, error)

	GetModuleName() string
}

type CustomerGroupHttpService struct {
	repo repositories.ICustomerGroupRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.CustomerGroupActivity, models.CustomerGroupDeleteActivity]
	contextTimeout time.Duration
}

func NewCustomerGroupHttpService(repo repositories.ICustomerGroupRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *CustomerGroupHttpService {
	contextTimeout := time.Duration(15) * time.Second

	insSvc := &CustomerGroupHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.CustomerGroupActivity, models.CustomerGroupDeleteActivity](repo)

	return insSvc
}

func (svc CustomerGroupHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc CustomerGroupHttpService) CreateCustomerGroup(shopID string, authUsername string, doc models.CustomerGroup) (string, error) {

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

	docData := models.CustomerGroupDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.CustomerGroup = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(shopID)

	return newGuidFixed, nil
}

func (svc CustomerGroupHttpService) UpdateCustomerGroup(shopID string, guid string, authUsername string, doc models.CustomerGroup) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.CustomerGroup = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc CustomerGroupHttpService) DeleteCustomerGroup(shopID string, guid string, authUsername string) error {
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

func (svc CustomerGroupHttpService) DeleteCustomerGroupByGUIDs(shopID string, authUsername string, GUIDs []string) error {
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

func (svc CustomerGroupHttpService) InfoCustomerGroup(shopID string, guid string) (models.CustomerGroupInfo, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.CustomerGroupInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.CustomerGroupInfo{}, errors.New("document not found")
	}

	return findDoc.CustomerGroupInfo, nil

}

func (svc CustomerGroupHttpService) SearchCustomerGroup(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CustomerGroupInfo, mongopagination.PaginationData, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"groupcode",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.CustomerGroupInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc CustomerGroupHttpService) SearchCustomerGroupStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.CustomerGroupInfo, int, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"groupcode",
		"names.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.CustomerGroupInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc CustomerGroupHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.CustomerGroup) (common.BulkImport, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.CustomerGroup](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.CustomerGroup, models.CustomerGroupDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.CustomerGroup) models.CustomerGroupDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.CustomerGroupDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.CustomerGroup = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.CustomerGroup, models.CustomerGroupDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.CustomerGroupDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "groupcode", guid)
		},
		func(doc models.CustomerGroupDoc) bool {
			return doc.GroupCode != ""
		},
		func(shopID string, authUsername string, data models.CustomerGroup, doc models.CustomerGroupDoc) error {

			doc.CustomerGroup = data
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

func (svc CustomerGroupHttpService) getDocIDKey(doc models.CustomerGroup) string {
	return doc.GroupCode
}

func (svc CustomerGroupHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc CustomerGroupHttpService) GetModuleName() string {
	return "customerGroup"
}
