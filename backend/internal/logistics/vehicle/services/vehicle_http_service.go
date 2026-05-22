package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/logistics/vehicle/models"
	"smlcloudplatform/internal/logistics/vehicle/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IVehicleHttpService interface {
	CreateVehicle(shopID string, authUsername string, doc models.Vehicle) (string, error)
	UpdateVehicle(shopID string, guid string, authUsername string, doc models.Vehicle) error
	DeleteVehicle(shopID string, guid string, authUsername string) error
	DeleteVehicleByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoVehicle(shopID string, guid string) (models.VehicleInfo, error)
	InfoVehicleByCode(shopID string, code string) (models.VehicleInfo, error)
	SearchVehicle(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.VehicleInfo, mongopagination.PaginationData, error)
	SearchVehicleStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.VehicleInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.Vehicle) (common.BulkImport, error)

	GetModuleName() string
}

type VehicleHttpService struct {
	repo          repositories.IVehicleRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.VehicleActivity, models.VehicleDeleteActivity]
	contextTimeout time.Duration
}

func NewVehicleHttpService(repo repositories.IVehicleRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *VehicleHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &VehicleHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.VehicleActivity, models.VehicleDeleteActivity](repo)

	return insSvc
}

func (svc VehicleHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc VehicleHttpService) CreateVehicle(shopID string, authUsername string, doc models.Vehicle) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// ตรวจสอบว่ามี vehicle code นี้อยู่แล้วหรือไม่
	findDoc, err := svc.repo.FindOneFilter(
		ctx,
		shopID,
		map[string]interface{}{
			"code": doc.Code,
		},
	)
	if err != nil {
		return "", err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", errors.New("รหัสยานพาหนะนี้มีอยู่ในระบบแล้ว")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.VehicleDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.Vehicle = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(shopID)

	return newGuidFixed, nil
}

func (svc VehicleHttpService) UpdateVehicle(shopID string, guid string, authUsername string, doc models.Vehicle) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("ไม่พบข้อมูลยานพาหนะ")
	}

	docData := findDoc

	docData.Vehicle = doc

	docData.GuidFixed = findDoc.GuidFixed
	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, docData)

	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc VehicleHttpService) DeleteVehicle(shopID string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("ไม่พบข้อมูลยานพาหนะ")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)
	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc VehicleHttpService) DeleteVehicleByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guid_fixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc VehicleHttpService) InfoVehicle(shopID string, guid string) (models.VehicleInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.VehicleInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.VehicleInfo{}, errors.New("ไม่พบข้อมูลยานพาหนะ")
	}

	return findDoc.VehicleInfo, nil
}

func (svc VehicleHttpService) InfoVehicleByCode(shopID string, code string) (models.VehicleInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)

	if err != nil {
		return models.VehicleInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.VehicleInfo{}, errors.New("ไม่พบข้อมูลยานพาหนะ")
	}

	return findDoc.VehicleInfo, nil
}

func (svc VehicleHttpService) SearchVehicle(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.VehicleInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
		"licenseplate",
		"brand",
		"drivername",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.VehicleInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc VehicleHttpService) SearchVehicleStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.VehicleInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
		"licenseplate",
		"brand",
		"drivername",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.VehicleInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc VehicleHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.Vehicle) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Vehicle](dataList, svc.getDocIDKey)

	itemCodeGuidList := []interface{}{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.Code)
	}

	findItemGuid, err := svc.repo.FindInItemGuids(ctx, shopID, "code", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.Code)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Vehicle, models.VehicleDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.Vehicle) models.VehicleDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.VehicleDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.Vehicle = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Vehicle, models.VehicleDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.VehicleDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", guid)
		},
		func(doc models.VehicleDoc) bool {
			return doc.Code != ""
		},
		func(shopID string, authUsername string, data models.Vehicle, doc models.VehicleDoc) error {

			doc.Vehicle = data
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

	createDataKey := []interface{}{}

	for _, doc := range createDataList {
		createDataKey = append(createDataKey, doc.Code)
	}

	payloadDuplicateDataKey := []interface{}{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.Code)
	}

	updateDataKey := []interface{}{}
	for _, doc := range updateSuccessDataList {

		updateDataKey = append(updateDataKey, doc.Code)
	}

	updateFailDataKey := []interface{}{}
	for _, doc := range updateFailDataList {
		updateFailDataKey = append(updateFailDataKey, svc.getDocIDKey(doc))
	}

	svc.saveMasterSync(shopID)

	tempCreateDataKey := svc.toSliceString(createDataKey)
	tempUpdateDataKey := svc.toSliceString(updateDataKey)
	tempUpdateFailDataKey := svc.toSliceString(updateFailDataKey)
	tempPayloadDuplicateDataKey := svc.toSliceString(payloadDuplicateDataKey)

	return common.BulkImport{
		Created:          tempCreateDataKey,
		Updated:          tempUpdateDataKey,
		UpdateFailed:     tempUpdateFailDataKey,
		PayloadDuplicate: tempPayloadDuplicateDataKey,
	}, nil
}

func (svc VehicleHttpService) toSliceString(data []interface{}) []string {
	tempData := make([]string, len(data))
	for i, v := range data {
		tempData[i] = v.(string)
	}

	return tempData
}

func (svc VehicleHttpService) getDocIDKey(doc models.Vehicle) string {
	return doc.Code
}

func (svc VehicleHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("บันทึก %s cache ผิดพลาด :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc VehicleHttpService) GetModuleName() string {
	return "vehicle"
}
