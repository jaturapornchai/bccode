package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/masterexpense/models"
	"smlcloudplatform/internal/masterexpense/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IMasterExpenseHttpService interface {
	CreateMasterExpense(holdingCode string, authUsername string, doc models.MasterExpense) (string, error)
	UpdateMasterExpense(holdingCode string, guid string, authUsername string, doc models.MasterExpense) error
	DeleteMasterExpense(holdingCode string, guid string, authUsername string) error
	DeleteMasterExpenseByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoMasterExpense(holdingCode string, guid string) (models.MasterExpenseInfo, error)
	InfoMasterExpenseByCode(holdingCode string, code string) (models.MasterExpenseInfo, error)
	SearchMasterExpense(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MasterExpenseInfo, mongopagination.PaginationData, error)
	SearchMasterExpenseStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.MasterExpenseInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.MasterExpense) (common.BulkImport, error)

	GetModuleName() string
}

type MasterExpenseHttpService struct {
	repo          repositories.IMasterExpenseRepository
	cacheRepo     repositories.IMasterExpenseCacheRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.MasterExpenseActivity, models.MasterExpenseDeleteActivity]
	contextTimeout time.Duration
}

func NewMasterExpenseHttpService(
	repo repositories.IMasterExpenseRepository,
	cacheRepo repositories.IMasterExpenseCacheRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	contextTimeout time.Duration,
) *MasterExpenseHttpService {

	insSvc := &MasterExpenseHttpService{
		repo:           repo,
		cacheRepo:      cacheRepo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.MasterExpenseActivity, models.MasterExpenseDeleteActivity](repo)

	return insSvc
}

func (svc MasterExpenseHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc MasterExpenseHttpService) CreateMasterExpense(holdingCode string, authUsername string, doc models.MasterExpense) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.existsCode(ctx, holdingCode, doc.Code)

	if err != nil {
		return "", err
	}

	// time.Sleep(30 * time.Second)

	newGuidFixed := utils.NewGUID()

	docData := models.MasterExpenseDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.MasterExpense = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
		svc.cacheRepo.ClearCreatedCode(holdingCode, doc.Code)
	}()

	return newGuidFixed, nil
}

func (svc MasterExpenseHttpService) existsCode(ctx context.Context, holdingCode string, code string) error {
	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) > 0 {
		return errors.New("code is exists")
	}

	createCodeSuccess, err := svc.cacheRepo.CreateCode(holdingCode, code, 15*time.Second)

	if err != nil {
		return errors.New("code is exists")
	}

	if !createCodeSuccess {
		return errors.New("code is exists")
	}
	return nil
}

func (svc MasterExpenseHttpService) UpdateMasterExpense(holdingCode string, guid string, authUsername string, doc models.MasterExpense) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	dataDoc := findDoc
	dataDoc.MasterExpense = doc

	dataDoc.Code = findDoc.Code
	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, dataDoc)

	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc MasterExpenseHttpService) DeleteMasterExpense(holdingCode string, guid string, authUsername string) error {

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
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc MasterExpenseHttpService) DeleteMasterExpenseByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc MasterExpenseHttpService) InfoMasterExpense(holdingCode string, guid string) (models.MasterExpenseInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.MasterExpenseInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.MasterExpenseInfo{}, errors.New("document not found")
	}

	return findDoc.MasterExpenseInfo, nil
}

func (svc MasterExpenseHttpService) InfoMasterExpenseByCode(holdingCode string, code string) (models.MasterExpenseInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)

	if err != nil {
		return models.MasterExpenseInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.MasterExpenseInfo{}, errors.New("document not found")
	}

	return findDoc.MasterExpenseInfo, nil
}

func (svc MasterExpenseHttpService) SearchMasterExpense(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MasterExpenseInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.MasterExpenseInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc MasterExpenseHttpService) SearchMasterExpenseStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.MasterExpenseInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	selectFields := map[string]interface{}{}

	/*
		if langCode != "" {
			selectFields["names"] = bson.M{"$elemMatch": bson.M{"code": langCode}}
		} else {
			selectFields["names"] = 1
		}
	*/

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.MasterExpenseInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc MasterExpenseHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.MasterExpense) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.MasterExpense](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.MasterExpense, models.MasterExpenseDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.MasterExpense) models.MasterExpenseDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.MasterExpenseDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.MasterExpense = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.MasterExpense, models.MasterExpenseDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.MasterExpenseDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.MasterExpenseDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.MasterExpense, doc models.MasterExpenseDoc) error {

			doc.MasterExpense = data
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

func (svc MasterExpenseHttpService) getDocIDKey(doc models.MasterExpense) string {
	return doc.Code
}

func (svc MasterExpenseHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc MasterExpenseHttpService) GetModuleName() string {
	return "masterExpense"
}
