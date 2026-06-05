package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/masterincome/models"
	"smlcloudplatform/internal/masterincome/repositories"
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

type IMasterIncomeHttpService interface {
	CreateMasterIncome(holdingCode string, authUsername string, doc models.MasterIncome) (string, error)
	UpdateMasterIncome(holdingCode string, guid string, authUsername string, doc models.MasterIncome) error
	DeleteMasterIncome(holdingCode string, guid string, authUsername string) error
	DeleteMasterIncomeByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoMasterIncome(holdingCode string, guid string) (models.MasterIncomeInfo, error)
	InfoMasterIncomeByCode(holdingCode string, code string) (models.MasterIncomeInfo, error)
	SearchMasterIncome(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MasterIncomeInfo, mongopagination.PaginationData, error)
	SearchMasterIncomeStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.MasterIncomeInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.MasterIncome) (common.BulkImport, error)

	GetModuleName() string
}

type MasterIncomeHttpService struct {
	repo          repositories.IMasterIncomeRepository
	cacheRepo     repositories.IMasterIncomeCacheRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.MasterIncomeActivity, models.MasterIncomeDeleteActivity]
	contextTimeout time.Duration
}

func NewMasterIncomeHttpService(
	repo repositories.IMasterIncomeRepository,
	cacheRepo repositories.IMasterIncomeCacheRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	contextTimeout time.Duration,
) *MasterIncomeHttpService {

	insSvc := &MasterIncomeHttpService{
		repo:           repo,
		cacheRepo:      cacheRepo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.MasterIncomeActivity, models.MasterIncomeDeleteActivity](repo)

	return insSvc
}

func (svc MasterIncomeHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc MasterIncomeHttpService) CreateMasterIncome(holdingCode string, authUsername string, doc models.MasterIncome) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.existsCode(ctx, holdingCode, doc.Code)
	if err != nil {
		return "", err
	}

	newGuidFixed := utils.NewGUID()

	docData := models.MasterIncomeDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.MasterIncome = doc

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

func (svc MasterIncomeHttpService) existsCode(ctx context.Context, holdingCode string, code string) error {
	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) > 0 {
		return errors.New("code is exists")
	}

	createCodeSuccess, err := svc.cacheRepo.CreateCode(holdingCode, code, 60*time.Second)

	if err != nil {
		return errors.New("code is exists")
	}

	if !createCodeSuccess {
		return errors.New("code is exists")
	}
	return nil
}

func (svc MasterIncomeHttpService) UpdateMasterIncome(holdingCode string, guid string, authUsername string, doc models.MasterIncome) error {

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
	dataDoc.MasterIncome = doc

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

func (svc MasterIncomeHttpService) DeleteMasterIncome(holdingCode string, guid string, authUsername string) error {

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

func (svc MasterIncomeHttpService) DeleteMasterIncomeByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

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

func (svc MasterIncomeHttpService) InfoMasterIncome(holdingCode string, guid string) (models.MasterIncomeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.MasterIncomeInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.MasterIncomeInfo{}, errors.New("document not found")
	}

	return findDoc.MasterIncomeInfo, nil
}

func (svc MasterIncomeHttpService) InfoMasterIncomeByCode(holdingCode string, code string) (models.MasterIncomeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)

	if err != nil {
		return models.MasterIncomeInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.MasterIncomeInfo{}, errors.New("document not found")
	}

	return findDoc.MasterIncomeInfo, nil
}

func (svc MasterIncomeHttpService) SearchMasterIncome(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MasterIncomeInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.MasterIncomeInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc MasterIncomeHttpService) SearchMasterIncomeStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.MasterIncomeInfo, int, error) {

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
		return []models.MasterIncomeInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc MasterIncomeHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.MasterIncome) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.MasterIncome](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.MasterIncome, models.MasterIncomeDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.MasterIncome) models.MasterIncomeDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.MasterIncomeDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.MasterIncome = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.MasterIncome, models.MasterIncomeDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.MasterIncomeDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.MasterIncomeDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.MasterIncome, doc models.MasterIncomeDoc) error {

			doc.MasterIncome = data
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

func (svc MasterIncomeHttpService) getDocIDKey(doc models.MasterIncome) string {
	return doc.Code
}

func (svc MasterIncomeHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc MasterIncomeHttpService) GetModuleName() string {
	return "masterIncome"
}
