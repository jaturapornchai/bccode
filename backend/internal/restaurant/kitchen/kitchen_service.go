package kitchen

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/restaurant/kitchen/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IKitchenService interface {
	CreateKitchen(shopID string, authUsername string, doc models.Kitchen) (string, error)
	UpdateKitchen(shopID string, guid string, authUsername string, doc models.Kitchen) error
	DeleteKitchen(shopID string, guid string, authUsername string) error
	InfoKitchen(shopID string, guid string) (models.KitchenInfo, error)
	SearchKitchen(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.KitchenInfo, mongopagination.PaginationData, error)
	SearchKitchenStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.KitchenInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.Kitchen) (common.BulkImport, error)
	GetProductBarcodeKitchen(shopID string) ([]models.KitchenProductBarcode, error)

	// LastActivity(shopID string, action string, lastUpdatedDate time.Time, pageable micromodels.Pageable) (common.LastActivity, mongopagination.PaginationData, error)

	GetModuleName() string
}

type KitchenService struct {
	repo          IKitchenRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.KitchenActivity, models.KitchenDeleteActivity]
	contextTimeout time.Duration
}

func NewKitchenService(repo KitchenRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) KitchenService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := KitchenService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.KitchenActivity, models.KitchenDeleteActivity](repo)
	return insSvc
}

func (svc KitchenService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc KitchenService) CreateKitchen(shopID string, authUsername string, doc models.Kitchen) (string, error) {

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

	docData := models.KitchenDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.Kitchen = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err

	}

	svc.saveMasterSync(shopID)

	return newGuidFixed, nil
}

func (svc KitchenService) UpdateKitchen(shopID string, guid string, authUsername string, doc models.Kitchen) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.Kitchen = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc KitchenService) DeleteKitchen(shopID string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)

	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc KitchenService) InfoKitchen(shopID string, guid string) (models.KitchenInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.KitchenInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.KitchenInfo{}, errors.New("document not found")
	}

	return findDoc.KitchenInfo, nil

}

func (svc KitchenService) SearchKitchen(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.KitchenInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	for i := range [5]bool{} {
		searchInFields = append(searchInFields, fmt.Sprintf("name%d", (i+1)))
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.KitchenInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc KitchenService) SearchKitchenStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.KitchenInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"names.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.KitchenInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc KitchenService) SaveInBatch(shopID string, authUsername string, dataList []models.Kitchen) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadCategoryList, payloadDuplicateCategoryList := importdata.FilterDuplicate[models.Kitchen](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Kitchen, models.KitchenDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadCategoryList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.Kitchen) models.KitchenDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.KitchenDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.Kitchen = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			dataDoc.LastUpdatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Kitchen, models.KitchenDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.KitchenDoc, error) {
			return svc.repo.FindByGuid(ctx, shopID, guid)
		},
		func(doc models.KitchenDoc) bool {
			return false
		},
		func(shopID string, authUsername string, data models.Kitchen, doc models.KitchenDoc) error {

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

func (svc KitchenService) getDocIDKey(doc models.Kitchen) string {
	return doc.Code
}

func (svc KitchenService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc KitchenService) GetModuleName() string {
	return "restaurant-kitchen"
}

func (svc KitchenService) GetProductBarcodeKitchen(shopID string) ([]models.KitchenProductBarcode, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	isNotFinished := true

	currentPage := 1
	limit := 20

	tempDocs := map[string][]models.KitchenInfo{}

	for isNotFinished {
		findDocs, pagination, err := svc.repo.FindPage(ctx, shopID, []string{}, micromodels.Pageable{
			Page:  currentPage,
			Limit: limit,
		})

		if err != nil {
			return []models.KitchenProductBarcode{}, err
		}

		for _, doc := range findDocs {
			for _, barcode := range *doc.Products {
				if barcode != "" {
					if _, ok := tempDocs[barcode]; !ok {
						tempDocs[barcode] = []models.KitchenInfo{}
					}
					tempDocs[barcode] = append(tempDocs[barcode], doc)
				}
			}
		}

		if int64(currentPage) >= pagination.TotalPage {
			isNotFinished = false
		} else {
			currentPage++
		}

	}

	docs := []models.KitchenProductBarcode{}
	for barcode, doc := range tempDocs {

		kitchens := []models.KitchenBarcode{}
		for _, item := range doc {
			kitchens = append(kitchens, models.KitchenBarcode{
				Code:  item.Code,
				Names: item.Names,
			})
		}

		docs = append(docs, models.KitchenProductBarcode{
			Barcode:  barcode,
			Kitchens: kitchens,
		})
	}

	return docs, nil
}
