package table

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/restaurant/table/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ITableService interface {
	CreateTable(holdingCode string, authUsername string, doc models.Table) (string, error)
	UpdateTable(holdingCode string, guid string, authUsername string, doc models.Table) error
	DeleteTable(holdingCode string, guid string, authUsername string) error
	DeleteTableByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	DeleteByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoTable(holdingCode string, guid string) (models.TableInfo, error)
	SearchTable(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.TableInfo, mongopagination.PaginationData, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.Table) (common.BulkImport, error)
	SaveXOrder(holdingCode string, authUsername string, xOrders []models.XOrderRequest) error

	GetModuleName() string
}

type TableService struct {
	repo          ITableRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository

	services.ActivityService[models.TableActivity, models.TableDeleteActivity]
	contextTimeout time.Duration
}

func NewTableService(repo ITableRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *TableService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := TableService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.TableActivity, models.TableDeleteActivity](repo)
	return &insSvc
}

func (svc TableService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc TableService) CreateTable(holdingCode string, authUsername string, doc models.Table) (string, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByTwoColumns(ctx, holdingCode, "number", doc.Number, "groupnumber", doc.GroupNumber)

	if err != nil {
		return "", err
	}

	if len(findDoc.Number) > 0 {
		return "", errors.New("code already exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.TableDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.Table = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)

	return newGuidFixed, nil
}

func (svc TableService) UpdateTable(holdingCode string, guid string, authUsername string, doc models.Table) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.Table = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc TableService) DeleteTable(holdingCode string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc TableService) DeleteTableByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

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

func (svc TableService) DeleteByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

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

func (svc TableService) InfoTable(holdingCode string, guid string) (models.TableInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.TableInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.TableInfo{}, errors.New("document not found")
	}

	return findDoc.TableInfo, nil

}

func (svc TableService) SearchTable(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.TableInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	for i := range [5]bool{} {
		searchInFields = append(searchInFields, fmt.Sprintf("name%d", (i+1)))
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.TableInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc TableService) SaveInBatch(holdingCode string, authUsername string, dataList []models.Table) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadCategoryList, payloadDuplicateCategoryList := importdata.FilterDuplicate[models.Table](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadCategoryList {
		itemCodeGuidList = append(itemCodeGuidList, doc.Number)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, holdingCode, "code", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.Code)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Table, models.TableDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadCategoryList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.Table) models.TableDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.TableDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.Table = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			dataDoc.LastUpdatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Table, models.TableDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.TableDoc, error) {
			return svc.repo.FindByGuid(ctx, holdingCode, guid)
		},
		func(doc models.TableDoc) bool {
			return false
		},
		func(holdingCode string, authUsername string, data models.Table, doc models.TableDoc) error {

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
		createDataKey = append(createDataKey, doc.Number)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateCategoryList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.Number)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {
		updateDataKey = append(updateDataKey, doc.Number)
	}

	updateFailDataKey := []string{}
	for _, doc := range updateFailDataList {
		updateFailDataKey = append(updateFailDataKey, doc.Number)
	}

	svc.saveMasterSync(holdingCode)

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc TableService) getDocIDKey(doc models.Table) string {
	return doc.Number
}

func (svc TableService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc TableService) GetModuleName() string {
	return "restauranttable"
}

func (svc TableService) SaveXOrder(holdingCode string, authUsername string, xOrders []models.XOrderRequest) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	for _, xOrder := range xOrders {
		svc.repo.SaveXOrder(ctx, holdingCode, xOrder.GuidFixed, xOrder.XOrder)
	}

	return nil
}
