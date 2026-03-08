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
	CreateMasterIncome(shopID string, authUsername string, doc models.MasterIncome) (string, error)
	UpdateMasterIncome(shopID string, guid string, authUsername string, doc models.MasterIncome) error
	DeleteMasterIncome(shopID string, guid string, authUsername string) error
	DeleteMasterIncomeByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoMasterIncome(shopID string, guid string) (models.MasterIncomeInfo, error)
	InfoMasterIncomeByCode(shopID string, code string) (models.MasterIncomeInfo, error)
	SearchMasterIncome(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MasterIncomeInfo, mongopagination.PaginationData, error)
	SearchMasterIncomeStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.MasterIncomeInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.MasterIncome) (common.BulkImport, error)

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

func (svc MasterIncomeHttpService) CreateMasterIncome(shopID string, authUsername string, doc models.MasterIncome) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.existsCode(ctx, shopID, doc.Code)
	if err != nil {
		return "", err
	}

	newGuidFixed := utils.NewGUID()

	docData := models.MasterIncomeDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.MasterIncome = doc

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

func (svc MasterIncomeHttpService) existsCode(ctx context.Context, shopID string, code string) error {
	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) > 0 {
		return errors.New("code is exists")
	}

	createCodeSuccess, err := svc.cacheRepo.CreateCode(shopID, code, 60*time.Second)

	if err != nil {
		return errors.New("code is exists")
	}

	if !createCodeSuccess {
		return errors.New("code is exists")
	}
	return nil
}

func (svc MasterIncomeHttpService) UpdateMasterIncome(shopID string, guid string, authUsername string, doc models.MasterIncome) error {

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
	dataDoc.MasterIncome = doc

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

func (svc MasterIncomeHttpService) DeleteMasterIncome(shopID string, guid string, authUsername string) error {

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

func (svc MasterIncomeHttpService) DeleteMasterIncomeByGUIDs(shopID string, authUsername string, GUIDs []string) error {

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

func (svc MasterIncomeHttpService) InfoMasterIncome(shopID string, guid string) (models.MasterIncomeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.MasterIncomeInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.MasterIncomeInfo{}, errors.New("document not found")
	}

	return findDoc.MasterIncomeInfo, nil
}

func (svc MasterIncomeHttpService) InfoMasterIncomeByCode(shopID string, code string) (models.MasterIncomeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)

	if err != nil {
		return models.MasterIncomeInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.MasterIncomeInfo{}, errors.New("document not found")
	}

	return findDoc.MasterIncomeInfo, nil
}

func (svc MasterIncomeHttpService) SearchMasterIncome(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MasterIncomeInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.MasterIncomeInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc MasterIncomeHttpService) SearchMasterIncomeStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.MasterIncomeInfo, int, error) {

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
		return []models.MasterIncomeInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc MasterIncomeHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.MasterIncome) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.MasterIncome](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.MasterIncome, models.MasterIncomeDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.MasterIncome) models.MasterIncomeDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.MasterIncomeDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.MasterIncome = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.MasterIncome, models.MasterIncomeDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.MasterIncomeDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", guid)
		},
		func(doc models.MasterIncomeDoc) bool {
			return doc.Code != ""
		},
		func(shopID string, authUsername string, data models.MasterIncome, doc models.MasterIncomeDoc) error {

			doc.MasterIncome = data
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

func (svc MasterIncomeHttpService) getDocIDKey(doc models.MasterIncome) string {
	return doc.Code
}

func (svc MasterIncomeHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc MasterIncomeHttpService) GetModuleName() string {
	return "masterIncome"
}
