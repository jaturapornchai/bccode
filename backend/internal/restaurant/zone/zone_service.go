package zone

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/restaurant/zone/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IZoneService interface {
	CreateZone(shopID string, authUsername string, doc models.Zone) (string, error)
	UpdateZone(shopID string, guid string, authUsername string, doc models.Zone) error
	DeleteZone(shopID string, guid string, authUsername string) error
	InfoZone(shopID string, guid string) (models.ZoneInfo, error)
	InfoWTFArray(shopID string, codes []string) ([]interface{}, error)
	SearchZone(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ZoneInfo, mongopagination.PaginationData, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.Zone) (common.BulkImport, error)
	DeleteByGUIDs(shopID string, authUsername string, GUIDs []string) error

	GetModuleName() string
}

type ZoneService struct {
	repo          IZoneRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository

	services.ActivityService[models.ZoneActivity, models.ZoneDeleteActivity]
	contextTimeout time.Duration
}

func NewZoneService(repo IZoneRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) ZoneService {
	contextTimeout := time.Duration(15) * time.Second

	insSvc := ZoneService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.ZoneActivity, models.ZoneDeleteActivity](repo)
	return insSvc
}

func (svc ZoneService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ZoneService) CreateZone(shopID string, authUsername string, doc models.Zone) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if len(findDoc.Code) > 0 {
		return "", errors.New("code already exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.ZoneDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.Zone = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	docData.LastUpdatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(shopID)

	return newGuidFixed, nil
}

func (svc ZoneService) UpdateZone(shopID string, guid string, authUsername string, doc models.Zone) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.Zone = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	findDoc.LastUpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc ZoneService) DeleteZone(shopID string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)

	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc ZoneService) DeleteByGUIDs(shopID string, authUsername string, GUIDs []string) error {

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

func (svc ZoneService) InfoZone(shopID string, guid string) (models.ZoneInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.ZoneInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.ZoneInfo{}, errors.New("document not found")
	}

	return findDoc.ZoneInfo, nil

}

func (svc ZoneService) InfoWTFArray(shopID string, codes []string) ([]interface{}, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList := []interface{}{}

	for _, code := range codes {
		findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)
		if err != nil || findDoc.ID == primitive.NilObjectID {
			// add item empty
			docList = append(docList, nil)
		} else {
			docList = append(docList, findDoc.ZoneInfo)
		}
	}

	return docList, nil
}

func (svc ZoneService) SearchZone(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ZoneInfo, mongopagination.PaginationData, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.ZoneInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc ZoneService) SaveInBatch(shopID string, authUsername string, dataList []models.Zone) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadCategoryList, payloadDuplicateCategoryList := importdata.FilterDuplicate[models.Zone](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadCategoryList {
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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Zone, models.ZoneDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadCategoryList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.Zone) models.ZoneDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.ZoneDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.Zone = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			dataDoc.LastUpdatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Zone, models.ZoneDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.ZoneDoc, error) {
			return svc.repo.FindByGuid(ctx, shopID, guid)
		},
		func(doc models.ZoneDoc) bool {
			return false
		},
		func(shopID string, authUsername string, data models.Zone, doc models.ZoneDoc) error {

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

	svc.saveMasterSync(shopID)

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc ZoneService) getDocIDKey(doc models.Zone) string {
	return doc.Code
}

func (svc ZoneService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc ZoneService) GetModuleName() string {
	return "restaurant-zone"
}
