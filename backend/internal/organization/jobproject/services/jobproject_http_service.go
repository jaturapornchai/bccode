package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/organization/jobproject/models"
	"smlcloudplatform/internal/organization/jobproject/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IJobProjectHttpService interface {
	CreateJobProject(holdingCode string, authUsername string, doc models.JobProject) (string, error)
	UpdateJobProject(holdingCode string, guid string, authUsername string, doc models.JobProject) error
	DeleteJobProject(holdingCode string, guid string, authUsername string) error
	DeleteJobProjectByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoJobProject(holdingCode string, guid string) (models.JobProjectInfo, error)
	InfoJobProjectByCode(holdingCode, branchCode, jobProjectCode string) (models.JobProjectInfo, error)
	SearchJobProject(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.JobProjectInfo, mongopagination.PaginationData, error)
	SearchJobProjectStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.JobProjectInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.JobProject) (common.BulkImport, error)

	GetModuleName() string
}

type JobProjectHttpService struct {
	repo             repositories.IJobProjectRepository
	repoMessageQueue repositories.IJobProjectMessageQueueRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.JobProjectActivity, models.JobProjectDeleteActivity]
	contextTimeout time.Duration
}

func NewJobProjectHttpService(repo repositories.IJobProjectRepository, repoMessageQueue repositories.IJobProjectMessageQueueRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *JobProjectHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &JobProjectHttpService{
		repo:             repo,
		repoMessageQueue: repoMessageQueue,
		syncCacheRepo:    syncCacheRepo,
		contextTimeout:   contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.JobProjectActivity, models.JobProjectDeleteActivity](repo)

	return insSvc
}

func (svc JobProjectHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc JobProjectHttpService) CreateJobProject(holdingCode string, authUsername string, doc models.JobProject) (string, error) {

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

	docData := models.JobProjectDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.JobProject = doc

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

func (svc JobProjectHttpService) UpdateJobProject(holdingCode string, guid string, authUsername string, doc models.JobProject) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	findDoc.JobProject = doc

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

func (svc JobProjectHttpService) DeleteJobProject(holdingCode string, guid string, authUsername string) error {

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

func (svc JobProjectHttpService) DeleteJobProjectByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	return nil
}

func (svc JobProjectHttpService) InfoJobProject(holdingCode string, guid string) (models.JobProjectInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.JobProjectInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.JobProjectInfo{}, errors.New("document not found")
	}

	return findDoc.JobProjectInfo, nil
}

func (svc JobProjectHttpService) InfoJobProjectByCode(holdingCode, branchCode, jobProjectCode string) (models.JobProjectInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindOneByCode(ctx, holdingCode, branchCode, jobProjectCode)

	if err != nil {
		return models.JobProjectInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.JobProjectInfo{}, errors.New("document not found")
	}

	return findDoc.JobProjectInfo, nil
}

func (svc JobProjectHttpService) SearchJobProject(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.JobProjectInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.JobProjectInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc JobProjectHttpService) SearchJobProjectStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.JobProjectInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.JobProjectInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc JobProjectHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.JobProject) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.JobProject](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.JobProject, models.JobProjectDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.JobProject) models.JobProjectDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.JobProjectDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.JobProject = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.JobProject, models.JobProjectDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.JobProjectDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.JobProjectDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.JobProject, doc models.JobProjectDoc) error {

			doc.JobProject = data
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

func (svc JobProjectHttpService) getDocIDKey(doc models.JobProject) string {
	return doc.Code
}

func (svc JobProjectHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc JobProjectHttpService) GetModuleName() string {
	return "jobproject"
}
