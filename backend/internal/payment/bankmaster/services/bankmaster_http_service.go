package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/payment/bankmaster/models"
	"smlcloudplatform/internal/payment/bankmaster/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IBankMasterHttpService interface {
	CreateBankMaster(holdingCode string, authUsername string, doc models.BankMaster) (string, error)
	UpdateBankMaster(holdingCode string, guid string, authUsername string, doc models.BankMaster) error
	DeleteBankMaster(holdingCode string, guid string, authUsername string) error
	DeleteBankMasterByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoBankMaster(holdingCode string, guid string) (models.BankMasterInfo, error)
	SearchBankMaster(holdingCode string, pageable micromodels.Pageable) ([]models.BankMasterInfo, mongopagination.PaginationData, error)
	SearchBankMasterStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.BankMasterInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.BankMaster) (common.BulkImport, error)

	GetModuleName() string
}

type BankMasterHttpService struct {
	repo repositories.IBankMasterRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.BankMasterActivity, models.BankMasterDeleteActivity]
	contextTimeout time.Duration
}

func NewBankMasterHttpService(repo repositories.IBankMasterRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *BankMasterHttpService {
	contextTimeout := time.Duration(15) * time.Second

	insSvc := &BankMasterHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.BankMasterActivity, models.BankMasterDeleteActivity](repo)

	return insSvc
}

func (svc BankMasterHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc BankMasterHttpService) CreateBankMaster(holdingCode string, authUsername string, doc models.BankMaster) (string, error) {

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

	docData := models.BankMasterDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.BankMaster = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)

	return newGuidFixed, nil
}

func (svc BankMasterHttpService) UpdateBankMaster(holdingCode string, guid string, authUsername string, doc models.BankMaster) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.BankMaster = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc BankMasterHttpService) DeleteBankMaster(holdingCode string, guid string, authUsername string) error {

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

func (svc BankMasterHttpService) DeleteBankMasterByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

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

func (svc BankMasterHttpService) InfoBankMaster(holdingCode string, guid string) (models.BankMasterInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.BankMasterInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.BankMasterInfo{}, errors.New("document not found")
	}

	return findDoc.BankMasterInfo, nil

}

func (svc BankMasterHttpService) SearchBankMaster(holdingCode string, pageable micromodels.Pageable) ([]models.BankMasterInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return []models.BankMasterInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc BankMasterHttpService) SearchBankMasterStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.BankMasterInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	selectFields := map[string]interface{}{
		"guid_fixed": 1,
		"code":       1,
		"logo":       1,
	}

	if langCode != "" {
		selectFields["names"] = bson.M{"$elemMatch": bson.M{"code": langCode}}
	} else {
		selectFields["names"] = 1
	}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.BankMasterInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc BankMasterHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.BankMaster) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.BankMaster](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.BankMaster, models.BankMasterDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.BankMaster) models.BankMasterDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.BankMasterDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.BankMaster = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.BankMaster, models.BankMasterDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.BankMasterDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.BankMasterDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.BankMaster, doc models.BankMasterDoc) error {

			doc.BankMaster = data
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

func (svc BankMasterHttpService) getDocIDKey(doc models.BankMaster) string {
	return doc.Code
}

func (svc BankMasterHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc BankMasterHttpService) GetModuleName() string {
	return "bankmaster"
}
