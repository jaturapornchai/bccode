package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/debtaccount/creditor/models"
	"smlcloudplatform/internal/debtaccount/creditor/repositories"
	groupModels "smlcloudplatform/internal/debtaccount/creditorgroup/models"
	groupRepositories "smlcloudplatform/internal/debtaccount/creditorgroup/repositories"
	"smlcloudplatform/internal/logger"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/samber/lo"
	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ICreditorHttpService interface {
	CreateCreditor(shopID string, authUsername string, doc models.CreditorRequest) (string, error)
	UpdateCreditor(shopID string, guid string, authUsername string, doc models.CreditorRequest) error
	DeleteCreditor(shopID string, guid string, authUsername string) error
	DeleteCreditorByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoCreditor(shopID string, guid string) (models.CreditorInfo, error)
	InfoCreditorByCode(shopID string, code string) (models.CreditorInfo, error)
	SearchCreditor(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CreditorInfo, mongopagination.PaginationData, error)
	SearchCreditorStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CreditorInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.CreditorRequest) (common.BulkImport, error)

	GetModuleName() string
}

type CreditorHttpService struct {
	repo          repositories.ICreditorRepository
	repoMq        repositories.ICreditorMessageQueueRepository
	repoGroup     groupRepositories.ICreditorGroupRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.CreditorActivity, models.CreditorDeleteActivity]

	contextTimeout time.Duration
}

func NewCreditorHttpService(
	repo repositories.ICreditorRepository,
	repoMq repositories.ICreditorMessageQueueRepository,
	repoGroup groupRepositories.ICreditorGroupRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository) *CreditorHttpService {
	contextTimeout := time.Duration(15) * time.Second
	insSvc := &CreditorHttpService{
		repo:           repo,
		repoMq:         repoMq,
		repoGroup:      repoGroup,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.CreditorActivity, models.CreditorDeleteActivity](repo)

	return insSvc
}

func (svc CreditorHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc CreditorHttpService) CreateCreditor(shopID string, authUsername string, doc models.CreditorRequest) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if findDoc.Code != "" {
		return "", errors.New("code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.CreditorDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.Creditor = doc.Creditor
	docData.GroupGUIDs = &doc.Groups

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	go func() {
		svc.saveMasterSync(shopID)
		err = svc.repoMq.Create(docData)
		if err != nil {
			logger.GetLogger().Errorf("Create creditor message queue error :: %s", err.Error())
		}
	}()

	return newGuidFixed, nil
}

func (svc CreditorHttpService) UpdateCreditor(shopID string, guid string, authUsername string, doc models.CreditorRequest) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	dataDoc := findDoc

	dataDoc.Creditor = doc.Creditor
	dataDoc.GroupGUIDs = &doc.Groups

	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, dataDoc)

	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(shopID)
		err = svc.repoMq.Update(dataDoc)
		if err != nil {
			logger.GetLogger().Errorf("Update creditor message queue error :: %s", err.Error())
		}
	}()

	return nil
}

func (svc CreditorHttpService) DeleteCreditor(shopID string, guid string, authUsername string) error {
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

	go func() {
		svc.saveMasterSync(shopID)
		err = svc.repoMq.Delete(findDoc)
		if err != nil {
			logger.GetLogger().Errorf("Delete creditor message queue error :: %s", err.Error())
		}
	}()

	return nil
}

func (svc CreditorHttpService) DeleteCreditorByGUIDs(shopID string, authUsername string, GUIDs []string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()
	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	findDocs, err := svc.repo.FindByGuids(ctx, shopID, GUIDs)

	if err != nil {
		return err
	}

	err = svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(shopID)
		err = svc.repoMq.DeleteInBatch(findDocs)
		if err != nil {
			logger.GetLogger().Errorf("Delete creditor message queue error :: %s", err.Error())
		}
	}()

	return nil
}

func (svc CreditorHttpService) InfoCreditor(shopID string, guid string) (models.CreditorInfo, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Find the document by guid
	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)
	if err != nil {
		return models.CreditorInfo{}, err
	}

	// Check if the document is not found
	if findDoc.ID == primitive.NilObjectID {
		return models.CreditorInfo{}, errors.New("document not found")
	}

	// Check if GroupGUIDs is nil
	if findDoc.GroupGUIDs == nil {
		return findDoc.CreditorInfo, nil
	}

	// Find groups by GroupGUIDs
	findGroups, err := svc.repoGroup.FindByGuids(ctx, shopID, *findDoc.GroupGUIDs)
	if err != nil {
		return models.CreditorInfo{}, err
	}

	// Map group documents to group info
	groupInfo := lo.Map[groupModels.CreditorGroupDoc, groupModels.CreditorGroupInfo](
		findGroups,
		func(docGroup groupModels.CreditorGroupDoc, idx int) groupModels.CreditorGroupInfo {
			return docGroup.CreditorGroupInfo
		})

	// Assign group info to the creditor info
	findDoc.CreditorInfo.Groups = &groupInfo

	return findDoc.CreditorInfo, nil
}

func (svc CreditorHttpService) InfoCreditorByCode(shopID string, code string) (models.CreditorInfo, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)

	if err != nil {
		return models.CreditorInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.CreditorInfo{}, errors.New("document not found")
	}

	if findDoc.GroupGUIDs != nil {
		findGroups, err := svc.repoGroup.FindByGuids(ctx, shopID, *findDoc.GroupGUIDs)

		if err != nil {
			return models.CreditorInfo{}, err
		}

		groupInfo := lo.Map[groupModels.CreditorGroupDoc, groupModels.CreditorGroupInfo](
			findGroups,
			func(docGroup groupModels.CreditorGroupDoc, idx int) groupModels.CreditorGroupInfo {
				return docGroup.CreditorGroupInfo
			})

		findDoc.CreditorInfo.Groups = &groupInfo

	}

	return findDoc.CreditorInfo, nil

}

func (svc CreditorHttpService) SearchCreditor(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CreditorInfo, mongopagination.PaginationData, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
		"groups",
		"fundcode",
		"addressforbilling.address.0",
		"addressforbilling.phoneprimary",
		"addressforbilling.phonesecondary",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.CreditorInfo{}, pagination, err
	}

	for idx, doc := range docList {
		if doc.GroupGUIDs != nil {
			findCustGroups, err := svc.repoGroup.FindByGuids(ctx, shopID, *doc.GroupGUIDs)
			if err != nil {
				return []models.CreditorInfo{}, pagination, err
			}

			custGroupInfo := lo.Map[groupModels.CreditorGroupDoc, groupModels.CreditorGroupInfo](
				findCustGroups,
				func(docGroup groupModels.CreditorGroupDoc, idx int) groupModels.CreditorGroupInfo {
					return docGroup.CreditorGroupInfo
				})

			docList[idx].Groups = &custGroupInfo
		}
	}

	return docList, pagination, nil
}

func (svc CreditorHttpService) SearchCreditorStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CreditorInfo, int, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
		"groups",
		"fundcode",
		"addressforbilling.address.0",
		"addressforbilling.phoneprimary",
		"addressforbilling.phonesecondary",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.CreditorInfo{}, 0, err
	}

	for idx, doc := range docList {
		if doc.GroupGUIDs != nil {
			findCustGroups, err := svc.repoGroup.FindByGuids(ctx, shopID, *doc.GroupGUIDs)
			if err != nil {
				return []models.CreditorInfo{}, 0, err
			}

			custGroupInfo := lo.Map[groupModels.CreditorGroupDoc, groupModels.CreditorGroupInfo](
				findCustGroups,
				func(docGroup groupModels.CreditorGroupDoc, idx int) groupModels.CreditorGroupInfo {
					return docGroup.CreditorGroupInfo
				})

			docList[idx].Groups = &custGroupInfo
		}
	}

	return docList, total, nil
}

func (svc CreditorHttpService) SaveInBatch(shopID string, authUsername string, dataListParam []models.CreditorRequest) (common.BulkImport, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	dataList := []models.Creditor{}
	for _, doc := range dataListParam {
		doc.GroupGUIDs = &doc.Groups
		dataList = append(dataList, doc.Creditor)
	}

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Creditor](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Creditor, models.CreditorDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.Creditor) models.CreditorDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.CreditorDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.Creditor = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Creditor, models.CreditorDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.CreditorDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", guid)
		},
		func(doc models.CreditorDoc) bool {
			return doc.Code != ""
		},
		func(shopID string, authUsername string, data models.Creditor, doc models.CreditorDoc) error {

			doc.Creditor = data
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

	go func() {
		svc.saveMasterSync(shopID)
		err = svc.repoMq.CreateInBatch(createDataList)
		if err != nil {
			logger.GetLogger().Errorf("Create creditor message queue error :: %s", err.Error())
		}
		svc.repoMq.UpdateInBatch(updateSuccessDataList)

		if err != nil {
			logger.GetLogger().Errorf("Update creditor message queue error :: %s", err.Error())
		}
	}()

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc CreditorHttpService) getDocIDKey(doc models.Creditor) string {
	return doc.Code
}

func (svc CreditorHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc CreditorHttpService) GetModuleName() string {
	return "creditor"
}
