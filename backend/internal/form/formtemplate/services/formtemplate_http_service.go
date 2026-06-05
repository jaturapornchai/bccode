package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/form/formtemplate/models"
	"smlcloudplatform/internal/form/formtemplate/repositories"
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

type IFormTemplateHttpService interface {
	SaveFormTemplate(holdingCode string, authUsername string, doc models.FormTemplate) (string, error)
	CreateFormTemplate(holdingCode string, authUsername string, doc models.FormTemplate) (string, error)
	UpdateFormTemplate(holdingCode string, guid string, authUsername string, doc models.FormTemplate) error
	DeleteFormTemplate(holdingCode string, guid string, authUsername string) error
	DeleteFormTemplateByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoFormTemplate(holdingCode string, guid string) (models.FormTemplateInfo, error)
	SearchFormTemplate(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.FormTemplateInfo, mongopagination.PaginationData, error)
	SearchFormTemplateStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.FormTemplateInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.FormTemplate) (common.BulkImport, error)
	GetModuleName() string
}

type FormTemplateHttpService struct {
	repo          repositories.IFormTemplateRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.FormTemplateActivity, models.FormTemplateDeleteActivity]
	contextTimeout time.Duration
}

func NewFormTemplateHttpService(
	repo repositories.IFormTemplateRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
) *FormTemplateHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &FormTemplateHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.FormTemplateActivity, models.FormTemplateDeleteActivity](repo)

	return insSvc
}

func (svc FormTemplateHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc FormTemplateHttpService) SaveFormTemplate(holdingCode string, authUsername string, doc models.FormTemplate) (string, error) {

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

func (svc FormTemplateHttpService) CreateFormTemplate(holdingCode string, authUsername string, doc models.FormTemplate) (string, error) {

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

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)

	return docData.GuidFixed, nil
}

func (svc FormTemplateHttpService) create(holdingCode string, authUsername string, doc models.FormTemplate) (models.FormTemplateDoc, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := utils.NewGUID()

	docData := models.FormTemplateDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.FormTemplate = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err := svc.repo.Create(ctx, docData)

	if err != nil {
		return models.FormTemplateDoc{}, err
	}

	return docData, nil
}

func (svc FormTemplateHttpService) UpdateFormTemplate(holdingCode string, guid string, authUsername string, doc models.FormTemplate) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	_, err = svc.update(holdingCode, authUsername, guid, findDoc, doc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc FormTemplateHttpService) update(holdingCode string, authUsername string, guid string, findDoc models.FormTemplateDoc, docUpdate models.FormTemplate) (models.FormTemplateDoc, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docData := findDoc
	docData.FormTemplate = docUpdate
	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	err := svc.repo.Update(ctx, holdingCode, guid, docData)

	if err != nil {
		return models.FormTemplateDoc{}, err
	}

	return docData, nil
}

func (svc FormTemplateHttpService) DeleteFormTemplate(holdingCode, guid, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return nil
	}

	err = svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)
	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc FormTemplateHttpService) DeleteFormTemplateByGUIDs(holdingCode, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc FormTemplateHttpService) InfoFormTemplate(holdingCode string, guid string) (models.FormTemplateInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.FormTemplateInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.FormTemplateInfo{}, errors.New("document not found")
	}

	return findDoc.FormTemplateInfo, nil
}

func (svc FormTemplateHttpService) SearchFormTemplate(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.FormTemplateInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
		"doctype",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.FormTemplateInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc FormTemplateHttpService) SearchFormTemplateStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.FormTemplateInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
		"doctype",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.FormTemplateInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc FormTemplateHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.FormTemplate) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.FormTemplate](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.FormTemplate, models.FormTemplateDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.FormTemplate) models.FormTemplateDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.FormTemplateDoc{}
			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.FormTemplate = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.FormTemplate, models.FormTemplateDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.FormTemplateDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.FormTemplateDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.FormTemplate, doc models.FormTemplateDoc) error {
			doc.FormTemplate = data
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

func (svc FormTemplateHttpService) getDocIDKey(doc models.FormTemplate) string {
	return doc.Code
}

func (svc FormTemplateHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())
		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc FormTemplateHttpService) GetModuleName() string {
	return "formTemplate"
}
