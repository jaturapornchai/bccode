package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/restaurant/device/models"
	"smlcloudplatform/internal/restaurant/device/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IDeviceHttpService interface {
	CreateDevice(holdingCode string, authUsername string, doc models.Device) (string, error)
	UpdateDevice(holdingCode string, guid string, authUsername string, doc models.Device) error
	DeleteDevice(holdingCode string, guid string, authUsername string) error
	DeleteDeviceByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoDevice(holdingCode string, guid string) (models.DeviceInfo, error)
	SearchDevice(holdingCode string, pageable micromodels.Pageable) ([]models.DeviceInfo, mongopagination.PaginationData, error)
	SearchDeviceStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.DeviceInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.Device) (common.BulkImport, error)

	GetModuleName() string
}

type DeviceHttpService struct {
	repo repositories.IDeviceRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.DeviceActivity, models.DeviceDeleteActivity]
	contextTimeout time.Duration
}

func NewDeviceHttpService(repo repositories.IDeviceRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *DeviceHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &DeviceHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.DeviceActivity, models.DeviceDeleteActivity](repo)

	return insSvc
}

func (svc DeviceHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc DeviceHttpService) CreateDevice(holdingCode string, authUsername string, doc models.Device) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if findDoc.Code != "" {
		return "", errors.New("code is already exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.DeviceDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.Device = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(holdingCode)

	return newGuidFixed, nil
}

func (svc DeviceHttpService) UpdateDevice(holdingCode string, guid string, authUsername string, doc models.Device) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.Device = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc DeviceHttpService) DeleteDevice(holdingCode string, guid string, authUsername string) error {

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

func (svc DeviceHttpService) DeleteDeviceByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

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

func (svc DeviceHttpService) InfoDevice(holdingCode string, guid string) (models.DeviceInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.DeviceInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.DeviceInfo{}, errors.New("document not found")
	}

	return findDoc.DeviceInfo, nil

}

func (svc DeviceHttpService) SearchDevice(holdingCode string, pageable micromodels.Pageable) ([]models.DeviceInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return []models.DeviceInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc DeviceHttpService) SearchDeviceStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.DeviceInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.DeviceInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc DeviceHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.Device) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Device](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Device, models.DeviceDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.Device) models.DeviceDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.DeviceDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.Device = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Device, models.DeviceDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.DeviceDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		},
		func(doc models.DeviceDoc) bool {
			return doc.Code != ""
		},
		func(holdingCode string, authUsername string, data models.Device, doc models.DeviceDoc) error {

			doc.Device = data
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

func (svc DeviceHttpService) getDocIDKey(doc models.Device) string {
	return doc.Code
}

func (svc DeviceHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc DeviceHttpService) GetModuleName() string {
	return "restaurant-device"
}
