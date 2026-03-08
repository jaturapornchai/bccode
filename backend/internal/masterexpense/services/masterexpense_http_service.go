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
	CreateMasterExpense(shopID string, authUsername string, doc models.MasterExpense) (string, error)
	UpdateMasterExpense(shopID string, guid string, authUsername string, doc models.MasterExpense) error
	DeleteMasterExpense(shopID string, guid string, authUsername string) error
	DeleteMasterExpenseByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoMasterExpense(shopID string, guid string) (models.MasterExpenseInfo, error)
	InfoMasterExpenseByCode(shopID string, code string) (models.MasterExpenseInfo, error)
	SearchMasterExpense(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MasterExpenseInfo, mongopagination.PaginationData, error)
	SearchMasterExpenseStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.MasterExpenseInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.MasterExpense) (common.BulkImport, error)

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

func (svc MasterExpenseHttpService) CreateMasterExpense(shopID string, authUsername string, doc models.MasterExpense) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.existsCode(ctx, shopID, doc.Code)

	if err != nil {
		return "", err
	}

	// time.Sleep(30 * time.Second)

	newGuidFixed := utils.NewGUID()

	docData := models.MasterExpenseDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.MasterExpense = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	go func() {
		svc.saveMasterSync(shopID)
		svc.cacheRepo.ClearCreatedCode(shopID, doc.Code)
	}()

	return newGuidFixed, nil
}

func (svc MasterExpenseHttpService) existsCode(ctx context.Context, shopID string, code string) error {
	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) > 0 {
		return errors.New("code is exists")
	}

	createCodeSuccess, err := svc.cacheRepo.CreateCode(shopID, code, 15*time.Second)

	if err != nil {
		return errors.New("code is exists")
	}

	if !createCodeSuccess {
		return errors.New("code is exists")
	}
	return nil
}

func (svc MasterExpenseHttpService) UpdateMasterExpense(shopID string, guid string, authUsername string, doc models.MasterExpense) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

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

	err = svc.repo.Update(ctx, shopID, guid, dataDoc)

	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc MasterExpenseHttpService) DeleteMasterExpense(shopID string, guid string, authUsername string) error {

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

	go func() {
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc MasterExpenseHttpService) DeleteMasterExpenseByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc MasterExpenseHttpService) InfoMasterExpense(shopID string, guid string) (models.MasterExpenseInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.MasterExpenseInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.MasterExpenseInfo{}, errors.New("document not found")
	}

	return findDoc.MasterExpenseInfo, nil
}

func (svc MasterExpenseHttpService) InfoMasterExpenseByCode(shopID string, code string) (models.MasterExpenseInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)

	if err != nil {
		return models.MasterExpenseInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.MasterExpenseInfo{}, errors.New("document not found")
	}

	return findDoc.MasterExpenseInfo, nil
}

func (svc MasterExpenseHttpService) SearchMasterExpense(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MasterExpenseInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.MasterExpenseInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc MasterExpenseHttpService) SearchMasterExpenseStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.MasterExpenseInfo, int, error) {

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

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.MasterExpenseInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc MasterExpenseHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.MasterExpense) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.MasterExpense](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.MasterExpense, models.MasterExpenseDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.MasterExpense) models.MasterExpenseDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.MasterExpenseDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.MasterExpense = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.MasterExpense, models.MasterExpenseDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.MasterExpenseDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", guid)
		},
		func(doc models.MasterExpenseDoc) bool {
			return doc.Code != ""
		},
		func(shopID string, authUsername string, data models.MasterExpense, doc models.MasterExpenseDoc) error {

			doc.MasterExpense = data
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

func (svc MasterExpenseHttpService) getDocIDKey(doc models.MasterExpense) string {
	return doc.Code
}

func (svc MasterExpenseHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc MasterExpenseHttpService) GetModuleName() string {
	return "masterExpense"
}
