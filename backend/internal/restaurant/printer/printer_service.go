package printer

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/restaurant/printer/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IPrinterService interface {
	CreatePrinter(holdingCode string, authUsername string, doc models.Printer) (string, error)
	UpdatePrinter(holdingCode string, guid string, authUsername string, doc models.Printer) error
	DeletePrinter(holdingCode string, guid string, authUsername string) error
	InfoPrinter(holdingCode string, guid string) (models.PrinterInfo, error)
	SearchPrinter(holdingCode string, pageable micromodels.Pageable) ([]models.PrinterInfo, mongopagination.PaginationData, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.Printer) (common.BulkImport, error)
	SearchPrinterStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.PrinterInfo, int, error)

	GetModuleName() string
}

type PrinterService struct {
	repo          PrinterRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository

	services.ActivityService[models.PrinterActivity, models.PrinterDeleteActivity]
	contextTimeout time.Duration
}

func NewPrinterService(repo PrinterRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) PrinterService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := PrinterService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.PrinterActivity, models.PrinterDeleteActivity](repo)
	return insSvc
}

func (svc PrinterService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc PrinterService) CreatePrinter(holdingCode string, authUsername string, doc models.Printer) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if len(findDoc.Code) > 0 {
		return "", errors.New("code already exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.PrinterDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.Printer = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)

	return newGuidFixed, nil
}

func (svc PrinterService) UpdatePrinter(holdingCode string, guid string, authUsername string, doc models.Printer) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.Printer = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc PrinterService) DeletePrinter(holdingCode string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc PrinterService) InfoPrinter(holdingCode string, guid string) (models.PrinterInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.PrinterInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.PrinterInfo{}, errors.New("document not found")
	}

	return findDoc.PrinterInfo, nil

}

func (svc PrinterService) SearchPrinter(holdingCode string, pageable micromodels.Pageable) ([]models.PrinterInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	for i := range [5]bool{} {
		searchInFields = append(searchInFields, fmt.Sprintf("name%d", (i+1)))
	}

	docList, pagination, err := svc.repo.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return []models.PrinterInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc PrinterService) SearchPrinterStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.PrinterInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.PrinterInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc PrinterService) SaveInBatch(holdingCode string, authUsername string, dataList []models.Printer) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadCategoryList, payloadDuplicateCategoryList := importdata.FilterDuplicate[models.Printer](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadCategoryList {
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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Printer, models.PrinterDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadCategoryList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.Printer) models.PrinterDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.PrinterDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.Printer = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			dataDoc.LastUpdatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Printer, models.PrinterDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.PrinterDoc, error) {
			return svc.repo.FindByGuid(ctx, holdingCode, guid)
		},
		func(doc models.PrinterDoc) bool {
			return false
		},
		func(holdingCode string, authUsername string, data models.Printer, doc models.PrinterDoc) error {

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
	for _, doc := range payloadDuplicateCategoryList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.Code)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {
		updateDataKey = append(updateDataKey, doc.Code)
	}

	updateFailDataKey := []string{}
	for _, doc := range updateFailDataList {
		updateFailDataKey = append(updateFailDataKey, doc.Code)
	}

	svc.saveMasterSync(holdingCode)

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc PrinterService) getDocIDKey(doc models.Printer) string {
	return doc.Code
}

func (svc PrinterService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc PrinterService) GetModuleName() string {
	return "restaurant-printer"
}
