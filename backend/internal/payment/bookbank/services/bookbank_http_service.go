package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/payment/bookbank/models"
	"smlcloudplatform/internal/payment/bookbank/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IBookBankHttpService interface {
	CreateBookBank(shopID string, authUsername string, doc models.BookBank) (string, error)
	UpdateBookBank(shopID string, guid string, authUsername string, doc models.BookBank) error
	DeleteBookBank(shopID string, guid string, authUsername string) error
	DeleteBookBankByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoBookBank(shopID string, guid string) (models.BookBankInfo, error)
	SearchBookBank(shopID string, pageable micromodels.Pageable) ([]models.BookBankInfo, mongopagination.PaginationData, error)
	SearchBookBankStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.BookBankInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.BookBank) (common.BulkImport, error)

	GetModuleName() string
}

type BookBankHttpService struct {
	repo repositories.IBookBankRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.BookBankActivity, models.BookBankDeleteActivity]
	contextTimeout time.Duration
}

func NewBookBankHttpService(repo repositories.IBookBankRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *BookBankHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &BookBankHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.BookBankActivity, models.BookBankDeleteActivity](repo)

	return insSvc
}

func (svc BookBankHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc BookBankHttpService) CreateBookBank(shopID string, authUsername string, doc models.BookBank) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "passbook", doc.PassBook)

	if err != nil {
		return "", err
	}

	if findDoc.PassBook != "" {
		return "", errors.New("PassBook is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.BookBankDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.BookBank = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(shopID)

	return newGuidFixed, nil
}

func (svc BookBankHttpService) UpdateBookBank(shopID string, guid string, authUsername string, doc models.BookBank) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.BookBank = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc BookBankHttpService) DeleteBookBank(shopID string, guid string, authUsername string) error {

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

func (svc BookBankHttpService) DeleteBookBankByGUIDs(shopID string, authUsername string, GUIDs []string) error {

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

func (svc BookBankHttpService) InfoBookBank(shopID string, guid string) (models.BookBankInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.BookBankInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.BookBankInfo{}, errors.New("document not found")
	}

	return findDoc.BookBankInfo, nil

}

func (svc BookBankHttpService) SearchBookBank(shopID string, pageable micromodels.Pageable) ([]models.BookBankInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"bookcode",
		"bankcode",
		"passbook",
		"names.name",
		"banknames.name",
	}

	docList, pagination, err := svc.repo.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return []models.BookBankInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc BookBankHttpService) SearchBookBankStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.BookBankInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"bookcode",
		"bankcode",
		"passbook",
		"names.name",
		"banknames.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.BookBankInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc BookBankHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.BookBank) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.BookBank](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.PassBook)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, shopID, "passbook", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.PassBook)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.BookBank, models.BookBankDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.BookBank) models.BookBankDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.BookBankDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.BookBank = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.BookBank, models.BookBankDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.BookBankDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "passbook", guid)
		},
		func(doc models.BookBankDoc) bool {
			return doc.PassBook != ""
		},
		func(shopID string, authUsername string, data models.BookBank, doc models.BookBankDoc) error {

			doc.BookBank = data
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
		createDataKey = append(createDataKey, doc.PassBook)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.PassBook)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {

		updateDataKey = append(updateDataKey, doc.PassBook)
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

func (svc BookBankHttpService) getDocIDKey(doc models.BookBank) string {
	return doc.PassBook
}

func (svc BookBankHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc BookBankHttpService) GetModuleName() string {
	return "bookbank"
}
