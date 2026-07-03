package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/logger"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	"smlcloudplatform/internal/warehouse/models"
	"smlcloudplatform/internal/warehouse/repositories"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IWarehouseHttpService interface {
	CreateWarehouse(holdingCode string, authUsername string, doc models.Warehouse) (string, error)
	UpdateWarehouse(holdingCode string, guid string, authUsername string, doc models.Warehouse) error
	DeleteWarehouse(holdingCode string, guid string, authUsername string) error
	DeleteWarehouseByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoWarehouse(holdingCode string, guid string) (models.WarehouseInfo, error)
	InfoWarehouseByCode(holdingCode string, code string) (models.WarehouseInfo, error)
	SearchWarehouse(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.WarehouseInfo, mongopagination.PaginationData, error)
	SearchWarehouseStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.WarehouseInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.Warehouse) (common.BulkImport, error)

	GetModuleName() string
}

type WarehouseHttpService struct {
	repo   repositories.IWarehouseRepository
	repoMq repositories.IWarehouseMessageQueueRepository

	// repoLocation is used only to re-validate the company-scope subset invariant (every location's
	// CompanyGuids must stay a subset of its warehouse's) when a warehouse's own CompanyGuids is
	// narrowed on update — see UpdateWarehouse.
	repoLocation repositories.IWarehouseLocationRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.WarehouseActivity, models.WarehouseDeleteActivity]
	contextTimeout time.Duration
}

func NewWarehouseHttpService(repo repositories.IWarehouseRepository, repoMq repositories.IWarehouseMessageQueueRepository, repoLocation repositories.IWarehouseLocationRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *WarehouseHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &WarehouseHttpService{
		repo:           repo,
		repoMq:         repoMq,
		repoLocation:   repoLocation,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.WarehouseActivity, models.WarehouseDeleteActivity](repo)

	return insSvc
}

func (svc WarehouseHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc WarehouseHttpService) CreateWarehouse(holdingCode string, authUsername string, doc models.Warehouse) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if findDoc.Code != "" {
		return "", errors.New("code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.WarehouseDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.Warehouse = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
		err = svc.repoMq.Create(docData)

		if err != nil {
			logger.GetLogger().Errorf("Error create warehouse message queue : %v", err)
		}
	}()

	return newGuidFixed, nil
}

func (svc WarehouseHttpService) UpdateWarehouse(holdingCode string, guid string, authUsername string, doc models.Warehouse) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	// If this update narrows CompanyGuids (or restricts a previously-unrestricted warehouse), every
	// existing location under it must still be a subset of the new set — otherwise a location would
	// be left silently referencing a company its own parent warehouse no longer allows, breaking the
	// invariant validateLocationCompanyScope enforces at location create/update time.
	if svc.repoLocation != nil {
		locations, err := svc.repoLocation.FindByWarehouseGuids(ctx, holdingCode, []string{guid})
		if err != nil {
			return err
		}
		for _, location := range locations {
			if err := validateLocationCompanyScope(location.CompanyGuids, doc.CompanyGuids); err != nil {
				return fmt.Errorf("ไม่สามารถแก้ไขสิทธิ์บริษัทของคลังได้ เนื่องจากที่เก็บสินค้า %s จะไม่อยู่ในสิทธิ์ที่อนุญาตอีกต่อไป: %w", location.Code, err)
			}
		}
	}

	dataDoc := findDoc
	dataDoc.Warehouse = doc

	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, dataDoc)

	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
		svc.repoMq.Update(dataDoc)
	}()

	return nil
}

func (svc WarehouseHttpService) DeleteWarehouse(holdingCode string, guid string, authUsername string) error {

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

	go func() {
		svc.saveMasterSync(holdingCode)
		svc.repoMq.Delete(findDoc)
	}()

	return nil
}

func (svc WarehouseHttpService) DeleteWarehouseByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

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

func (svc WarehouseHttpService) InfoWarehouse(holdingCode string, guid string) (models.WarehouseInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.WarehouseInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.WarehouseInfo{}, errors.New("document not found")
	}

	return findDoc.WarehouseInfo, nil
}

func (svc WarehouseHttpService) InfoWarehouseByCode(holdingCode string, code string) (models.WarehouseInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)

	if err != nil {
		return models.WarehouseInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.WarehouseInfo{}, errors.New("document not found")
	}

	return findDoc.WarehouseInfo, nil
}

func (svc WarehouseHttpService) SearchWarehouse(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.WarehouseInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.WarehouseInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc WarehouseHttpService) SearchWarehouseStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.WarehouseInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.WarehouseInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc WarehouseHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.Warehouse) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Warehouse](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Warehouse, models.WarehouseDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.Warehouse) models.WarehouseDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.WarehouseDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.Warehouse = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Warehouse, models.WarehouseDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.WarehouseDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.WarehouseDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.Warehouse, doc models.WarehouseDoc) error {

			doc.Warehouse = data
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

func (svc WarehouseHttpService) getDocIDKey(doc models.Warehouse) string {
	return doc.Code
}

func (svc WarehouseHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc WarehouseHttpService) GetModuleName() string {
	return "warehouse"
}
