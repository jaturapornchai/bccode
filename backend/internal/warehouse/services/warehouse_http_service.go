package services

import (
	"context"
	"errors"
	"fmt"
	"log"
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

	"github.com/samber/lo"
	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IWarehouseHttpService interface {
	CreateWarehouse(shopID string, authUsername string, doc models.Warehouse) (string, error)
	UpdateWarehouse(shopID string, guid string, authUsername string, doc models.Warehouse) error
	DeleteWarehouse(shopID string, guid string, authUsername string) error
	DeleteWarehouseByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoWarehouse(shopID string, guid string) (models.WarehouseInfo, error)
	InfoWarehouseByCode(shopID string, code string) (models.WarehouseInfo, error)
	SearchWarehouse(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.WarehouseInfo, mongopagination.PaginationData, error)
	SearchWarehouseStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.WarehouseInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.Warehouse) (common.BulkImport, error)

	SearchLocation(shopID string, pageable micromodels.Pageable) ([]models.LocationInfo, mongopagination.PaginationData, error)
	SearchShelf(shopID string, pageable micromodels.Pageable) ([]models.ShelfInfo, mongopagination.PaginationData, error)

	InfoLocation(shopID, warehouseCode, locationCode string) (models.LocationInfo, error)
	CreateLocation(shopID, authUsername, warehouseCode string, doc models.LocationRequest) error
	UpdateLocation(shopID, authUsername, warehouseCode, locationCode string, doc models.LocationRequest) error
	DeleteLocationByCodes(shopID, authUsername, warehouseCode string, locationCodes []string) error

	InfoShelf(shopID, warehouseCode, locationCode, shelfCode string) (models.ShelfInfo, error)
	CreateShelf(shopID, authUsername, warehouseCode, locationCode string, doc models.ShelfRequest) error
	UpdateShelf(shopID, authUsername, warehouseCode, locationCode, shelfCode string, doc models.ShelfRequest) error
	DeleteShelfByCodes(shopID, authUsername, warehouseCode, locationCode string, shelfCodes []string) error

	// ProductBarcode in Shelf management
	AddProductBarcodeToShelf(shopID, authUsername, warehouseCode, locationCode, shelfCode string, productBarcode models.ShelfProductBarcode) error
	RemoveProductBarcodeFromShelf(shopID, authUsername, warehouseCode, locationCode, shelfCode, productGuidFixed string) error
	GetShelfProductBarcodes(shopID, warehouseCode, locationCode, shelfCode string) ([]models.ShelfProductBarcode, error)

	// Bulk ProductBarcode operations
	BulkAddProductBarcodesToShelf(shopID, authUsername, warehouseCode, locationCode, shelfCode string, products []models.ShelfProductBarcode) models.BulkProductOperationResponse
	BulkRemoveProductBarcodesFromShelf(shopID, authUsername, warehouseCode, locationCode, shelfCode string, guidFixedList []string) models.BulkProductOperationResponse

	GetModuleName() string
}

type WarehouseHttpService struct {
	repo   repositories.IWarehouseRepository
	repoMq repositories.IWarehouseMessageQueueRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.WarehouseActivity, models.WarehouseDeleteActivity]
	contextTimeout time.Duration
}

func NewWarehouseHttpService(repo repositories.IWarehouseRepository, repoMq repositories.IWarehouseMessageQueueRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *WarehouseHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &WarehouseHttpService{
		repo:           repo,
		repoMq:         repoMq,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.WarehouseActivity, models.WarehouseDeleteActivity](repo)

	return insSvc
}

func (svc WarehouseHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc WarehouseHttpService) CreateWarehouse(shopID string, authUsername string, doc models.Warehouse) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if findDoc.Code != "" {
		return "", errors.New("code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.WarehouseDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.Warehouse = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	go func() {
		svc.saveMasterSync(shopID)
		err = svc.repoMq.Create(docData)

		if err != nil {
			logger.GetLogger().Errorf("Error create warehouse message queue : %v", err)
		}
	}()

	return newGuidFixed, nil
}

func (svc WarehouseHttpService) UpdateWarehouse(shopID string, guid string, authUsername string, doc models.Warehouse) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	dataDoc := findDoc

	dataDoc.Warehouse = doc

	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, dataDoc)

	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(shopID)
		svc.repoMq.Update(dataDoc)
	}()

	return nil
}

func (svc WarehouseHttpService) CreateLocation(shopID, authUsername, warehouseCode string, doc models.LocationRequest) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", warehouseCode)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	locations := findDoc.Location

	for _, location := range *locations {
		if location.Code == doc.Code {
			return errors.New("location code is exists")
		}
	}

	dataDoc := findDoc

	*dataDoc.Location = append(*dataDoc.Location, models.Location{
		Code:  doc.Code,
		Names: doc.Names,
		Shelf: doc.Shelf,
	})

	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, findDoc.GuidFixed, dataDoc)

	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(shopID)
		svc.repoMq.Update(dataDoc)
	}()

	return nil
}

func (svc WarehouseHttpService) UpdateLocation(shopID, authUsername, warehouseCode, locationCode string, doc models.LocationRequest) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	updateDoc := models.WarehouseDoc{}
	removeDoc := models.WarehouseDoc{}

	findDoc, err := svc.repo.FindWarehouseByLocation(ctx, shopID, warehouseCode, locationCode)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	if warehouseCode != doc.WarehouseCode {

		findDocWarehouse, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", doc.WarehouseCode)

		if err != nil {
			return err
		}

		if len(findDocWarehouse.GuidFixed) < 1 {
			return errors.New("document not found")
		}

		updateDoc = findDocWarehouse

		// clear doc
		removeDoc = findDoc
		tempLocation := []models.Location{}

		for _, location := range *removeDoc.Location {
			if location.Code != locationCode {
				tempLocation = append(tempLocation, location)
			}
		}

		removeDoc.Location = &tempLocation
	} else {
		updateDoc = findDoc
	}

	if warehouseCode == doc.WarehouseCode {
		if locationCode != doc.Code {
			locations := updateDoc.Warehouse.Location

			for _, location := range *locations {
				if location.Code == doc.Code {
					return errors.New("location code is exists")
				}
			}

			isLocationExists := false
			for i, location := range *updateDoc.Location {
				if location.Code == doc.Code {
					isLocationExists = true
					location.Code = doc.Code
					location.Names = doc.Names
					location.Shelf = doc.Shelf
					(*updateDoc.Location)[i] = location
				}
			}

			if !isLocationExists {
				*updateDoc.Location = append(*updateDoc.Location, models.Location{
					Code:  doc.Code,
					Names: doc.Names,
					Shelf: doc.Shelf,
				})
			}

		} else {
			for i, location := range *updateDoc.Location {
				if location.Code == doc.Code {
					location.Names = doc.Names
					location.Shelf = doc.Shelf
					(*updateDoc.Location)[i] = location
				}
			}
		}
	} else {
		for _, location := range *updateDoc.Location {
			if location.Code == doc.Code {
				return errors.New("location code is exists")
			}
		}

		*updateDoc.Location = append(*updateDoc.Location, models.Location{
			Code:  doc.Code,
			Names: doc.Names,
			Shelf: doc.Shelf,
		})
	}

	err = svc.repo.Transaction(ctx, func(ctx context.Context) error {
		updateDoc.UpdatedBy = authUsername
		updateDoc.UpdatedAt = time.Now()

		err = svc.repo.Update(ctx, shopID, updateDoc.GuidFixed, updateDoc)

		if err != nil {
			return err
		}

		if len(removeDoc.GuidFixed) > 0 {
			removeDoc.UpdatedBy = authUsername
			removeDoc.UpdatedAt = time.Now()

			err = svc.repo.Update(ctx, shopID, removeDoc.GuidFixed, removeDoc)

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(shopID)
		svc.repoMq.Update(updateDoc)
		svc.repoMq.Update(removeDoc)
	}()

	return nil
}

func (svc WarehouseHttpService) DeleteLocationByCodes(shopID, authUsername, warehouseCode string, locationCodes []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	removeDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", warehouseCode)

	if err != nil {
		return err
	}

	if len(removeDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	// remove data
	codeIndex := map[string]struct{}{}
	for _, code := range locationCodes {
		codeIndex[code] = struct{}{}
	}

	locationTemp := []models.Location{}
	for _, location := range *removeDoc.Location {
		if _, ok := codeIndex[location.Code]; !ok {
			locationTemp = append(locationTemp, location)
		}
	}
	removeDoc.Location = &locationTemp

	removeDoc.UpdatedBy = authUsername
	removeDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, removeDoc.GuidFixed, removeDoc)

	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(shopID)
		svc.repoMq.Update(removeDoc)
	}()

	return nil
}

func (svc WarehouseHttpService) CreateShelf(shopID, authUsername, warehouseCode, locationCode string, doc models.ShelfRequest) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindWarehouseByShelf(ctx, shopID, warehouseCode, locationCode, doc.Code)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	dataDoc := findDoc
	locations := findDoc.Location

	for indexLocation, location := range *locations {

		if location.Code == locationCode {
			shelves := location.Shelf

			for _, shelf := range shelves {
				if shelf.Code == doc.Code {
					return errors.New("shelf Code is exists")
				}
			}

			tempLocation := (*dataDoc.Location)[indexLocation]
			tempShelf := tempLocation.Shelf

			tempShelf = append(tempShelf, models.Shelf{
				Code: doc.Code,
				Name: doc.Name,
			})

			(*dataDoc.Location)[indexLocation].Shelf = tempShelf

			break
		}
	}

	return nil
}

func (svc WarehouseHttpService) UpdateShelf(shopID, authUsername, warehouseCode, locationCode, shelfCode string, doc models.ShelfRequest) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	updateDoc := models.WarehouseDoc{}
	removeDoc := models.WarehouseDoc{}

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", warehouseCode)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	if warehouseCode != doc.WarehouseCode {

		findDocWarehouse, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", doc.WarehouseCode)

		if err != nil {
			return err
		}

		if len(findDocWarehouse.GuidFixed) < 1 {
			return errors.New("document not found")
		}

		updateDoc = findDocWarehouse
		removeDoc = findDoc

	} else {
		updateDoc = findDoc
		removeDoc = findDoc
	}

	if len(updateDoc.GuidFixed) < 1 || len(removeDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	// remove previous data
	isFoundShelf := false
	for locationIndex, location := range *removeDoc.Location {
		if location.Code == locationCode {
			locationTemp := (*removeDoc.Location)[locationIndex]

			if locationTemp.Shelf == nil {
				break
			}

			shelfTemp := []models.Shelf{}
			for _, shelf := range locationTemp.Shelf {
				if shelf.Code != shelfCode {
					shelfTemp = append(shelfTemp, shelf)
				}
			}
			(*removeDoc.Location)[locationIndex].Shelf = shelfTemp

			isFoundShelf = true
			break
		}
	}

	if !isFoundShelf {
		return errors.New("document not found")
	}

	// update new data
	for locationIndex, location := range *updateDoc.Location {
		if location.Code == doc.LocationCode {
			locationTemp := (*updateDoc.Location)[locationIndex]

			if locationTemp.Shelf == nil {
				locationTemp.Shelf = []models.Shelf{}
			}

			shelfTemp := lo.Filter[models.Shelf](locationTemp.Shelf, func(shelf models.Shelf, i int) bool {
				return shelf.Code != doc.Code
			})

			shelfTemp = append(shelfTemp, models.Shelf{
				Code: doc.Code,
				Name: doc.Name,
			})

			shelfTemp = append(shelfTemp, models.Shelf{
				Code: doc.Code,
				Name: doc.Name,
			})

			(*updateDoc.Location)[locationIndex].Shelf = shelfTemp
			break
		}
	}

	err = svc.repo.Transaction(ctx, func(ctx context.Context) error {

		removeDoc.UpdatedBy = authUsername
		removeDoc.UpdatedAt = time.Now()

		err = svc.repo.Update(ctx, shopID, removeDoc.GuidFixed, removeDoc)

		if err != nil {
			return err
		}

		updateDoc.UpdatedBy = authUsername
		updateDoc.UpdatedAt = time.Now()

		err = svc.repo.Update(ctx, shopID, updateDoc.GuidFixed, updateDoc)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (svc WarehouseHttpService) DeleteShelfByCodes(shopID, authUsername, warehouseCode, locationCode string, shelfCodes []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	removeDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", warehouseCode)

	if err != nil {
		return err
	}

	if len(removeDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	// remove data
	for locationIndex, location := range *removeDoc.Location {
		if location.Code == locationCode {
			locationTemp := (*removeDoc.Location)[locationIndex]

			if locationTemp.Shelf == nil {
				break
			}

			codeIndex := map[string]struct{}{}
			for _, code := range shelfCodes {
				codeIndex[code] = struct{}{}
			}

			shelfTemp := []models.Shelf{}
			for _, shelf := range locationTemp.Shelf {
				if _, ok := codeIndex[shelf.Code]; !ok {
					shelfTemp = append(shelfTemp, shelf)
				}
			}

			(*removeDoc.Location)[locationIndex].Shelf = shelfTemp
			break
		}
	}

	removeDoc.UpdatedBy = authUsername
	removeDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, removeDoc.GuidFixed, removeDoc)

	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(shopID)
		svc.repoMq.Update(removeDoc)
	}()

	return nil
}

func (svc WarehouseHttpService) DeleteWarehouse(shopID string, guid string, authUsername string) error {

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

	go func() {
		svc.saveMasterSync(shopID)
		svc.repoMq.Delete(findDoc)
	}()

	return nil
}

func (svc WarehouseHttpService) DeleteWarehouseByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guid_fixed": bson.M{"$in": GUIDs},
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

func (svc WarehouseHttpService) InfoWarehouse(shopID string, guid string) (models.WarehouseInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.WarehouseInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.WarehouseInfo{}, errors.New("document not found")
	}

	warehouseInfo := findDoc.WarehouseInfo

	// Debug: แสดงข้อมูลที่ได้จาก MongoDB
	log.Printf("DEBUG: WarehouseInfo from MongoDB: %+v", warehouseInfo)
	if warehouseInfo.Location != nil {
		log.Printf("DEBUG: Found %d locations", len(*warehouseInfo.Location))
	}

	// กรองข้อมูล location ที่ไม่สมบูรณ์ออก
	if warehouseInfo.Location != nil {
		filteredLocations := []models.Location{}
		for i, location := range *warehouseInfo.Location {
			// Debug: แสดงข้อมูล shelf ที่ได้จาก MongoDB
			log.Printf("DEBUG: Location[%d] Code=%s, Shelf count=%d", i, location.Code, len(location.Shelf))
			for j, shelf := range location.Shelf {
				log.Printf("DEBUG: Shelf[%d] Code=%s Name=%s", j, shelf.Code, shelf.Name)
			}

			// เก็บเฉพาะ location ที่มี code ไม่ว่างและมี names
			if location.Code != "" && location.Names != nil && len(*location.Names) > 0 {
				filteredLocations = append(filteredLocations, location)
			}
		}
		warehouseInfo.Location = &filteredLocations
	}

	return warehouseInfo, nil
}

func (svc WarehouseHttpService) InfoWarehouseByCode(shopID string, code string) (models.WarehouseInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)

	if err != nil {
		return models.WarehouseInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.WarehouseInfo{}, errors.New("document not found")
	}

	warehouseInfo := findDoc.WarehouseInfo

	// กรองข้อมูล location ที่ไม่สมบูรณ์ออก
	if warehouseInfo.Location != nil {
		filteredLocations := []models.Location{}
		for _, location := range *warehouseInfo.Location {
			// เก็บเฉพาะ location ที่มี code ไม่ว่างและมี names
			if location.Code != "" && location.Names != nil && len(*location.Names) > 0 {
				filteredLocations = append(filteredLocations, location)
			}
		}
		warehouseInfo.Location = &filteredLocations
	}

	return warehouseInfo, nil
}

func (svc WarehouseHttpService) InfoLocation(shopID, warehouseCode, locationCode string) (models.LocationInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", warehouseCode)

	if err != nil {
		return models.LocationInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.LocationInfo{}, errors.New("document not found")
	}

	locationInfo := models.LocationInfo{}

	locationInfo.GuidFixed = findDoc.GuidFixed
	locationInfo.WarehouseCode = findDoc.Code
	locationInfo.WarehouseNames = findDoc.Names

	// ตรวจสอบว่า findDoc.Location เป็น nil หรือไม่
	if findDoc.Location == nil {
		return models.LocationInfo{}, errors.New("no locations found in warehouse")
	}

	// ค้นหา location ที่ต้องการ
	locationFound := false
	for _, location := range *findDoc.Location {
		if location.Code == locationCode {
			locationInfo.LocationCode = location.Code
			locationInfo.LocationNames = location.Names
			locationInfo.Shelf = location.Shelf
			locationFound = true
			break
		}
	}

	// ถ้าไม่เจอ location ที่ต้องการ
	if !locationFound {
		return models.LocationInfo{}, fmt.Errorf("location '%s' not found in warehouse '%s'", locationCode, warehouseCode)
	}

	return locationInfo, nil
}

func (svc WarehouseHttpService) InfoShelf(shopID, warehouseCode, locationCode, shelfCode string) (models.ShelfInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", warehouseCode)

	if err != nil {
		return models.ShelfInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.ShelfInfo{}, errors.New("document not found")
	}

	shelfInfo := models.ShelfInfo{}

	shelfInfo.GuidFixed = findDoc.GuidFixed
	shelfInfo.WarehouseCode = findDoc.Code
	shelfInfo.WarehouseNames = findDoc.Names

	for _, location := range *findDoc.Location {
		if location.Code == locationCode {
			shelfInfo.LocationCode = location.Code
			shelfInfo.LocationNames = location.Names
			for _, shelf := range location.Shelf {
				if shelf.Code == shelfCode {
					shelfInfo.ShelfCode = shelf.Code
					shelfInfo.ShelfName = shelf.Name
					break
				}
			}
			break
		}
	}

	return shelfInfo, nil
}

func (svc WarehouseHttpService) SearchWarehouse(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.WarehouseInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.WarehouseInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc WarehouseHttpService) SearchWarehouseStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.WarehouseInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.WarehouseInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc WarehouseHttpService) SearchLocation(shopID string, pageable micromodels.Pageable) ([]models.LocationInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList, pagination, err := svc.repo.FindLocationPage(ctx, shopID, pageable)

	if err != nil {
		return []models.LocationInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc WarehouseHttpService) SearchShelf(shopID string, pageable micromodels.Pageable) ([]models.ShelfInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList, pagination, err := svc.repo.FindShelfPage(ctx, shopID, pageable)

	if err != nil {
		return []models.ShelfInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc WarehouseHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.Warehouse) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Warehouse](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Warehouse, models.WarehouseDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.Warehouse) models.WarehouseDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.WarehouseDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.Warehouse = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Warehouse, models.WarehouseDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.WarehouseDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", guid)
		},
		func(doc models.WarehouseDoc) bool {
			return doc.Code != ""
		},
		func(shopID string, authUsername string, data models.Warehouse, doc models.WarehouseDoc) error {

			doc.Warehouse = data
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

func (svc WarehouseHttpService) getDocIDKey(doc models.Warehouse) string {
	return doc.Code
}

func (svc WarehouseHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc WarehouseHttpService) GetModuleName() string {
	return "warehouse"
}

// ProductBarcode in Shelf management methods
func (svc WarehouseHttpService) AddProductBarcodeToShelf(shopID, authUsername, warehouseCode, locationCode, shelfCode string, productBarcode models.ShelfProductBarcode) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", warehouseCode)
	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("warehouse not found")
	}

	dataDoc := findDoc

	// Find and update the specific shelf
	for locationIndex, location := range *dataDoc.Location {
		if location.Code == locationCode {
			for shelfIndex, shelf := range location.Shelf {
				if shelf.Code == shelfCode {
					// Add product to shelf
					(*dataDoc.Location)[locationIndex].Shelf[shelfIndex].AddProductBarcode(
						productBarcode.GuidFixed,
						productBarcode.Barcode,
						productBarcode.Names,
					)

					dataDoc.UpdatedBy = authUsername
					dataDoc.UpdatedAt = time.Now()

					return svc.repo.Update(ctx, shopID, findDoc.GuidFixed, dataDoc)
				}
			}
			return errors.New("shelf not found")
		}
	}
	return errors.New("location not found")
}

func (svc WarehouseHttpService) RemoveProductBarcodeFromShelf(shopID, authUsername, warehouseCode, locationCode, shelfCode, productGuidFixed string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", warehouseCode)
	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("warehouse not found")
	}

	dataDoc := findDoc

	// Find and update the specific shelf
	for locationIndex, location := range *dataDoc.Location {
		if location.Code == locationCode {
			for shelfIndex, shelf := range location.Shelf {
				if shelf.Code == shelfCode {
					// Remove product from shelf
					if !(*dataDoc.Location)[locationIndex].Shelf[shelfIndex].RemoveProductBarcode(productGuidFixed) {
						return errors.New("product not found on shelf")
					}

					dataDoc.UpdatedBy = authUsername
					dataDoc.UpdatedAt = time.Now()

					return svc.repo.Update(ctx, shopID, findDoc.GuidFixed, dataDoc)
				}
			}
			return errors.New("shelf not found")
		}
	}
	return errors.New("location not found")
}

func (svc WarehouseHttpService) GetShelfProductBarcodes(shopID, warehouseCode, locationCode, shelfCode string) ([]models.ShelfProductBarcode, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", warehouseCode)
	if err != nil {
		return []models.ShelfProductBarcode{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return []models.ShelfProductBarcode{}, errors.New("warehouse not found")
	}

	// Find the specific shelf and return its products
	for _, location := range *findDoc.Location {
		if location.Code == locationCode {
			for _, shelf := range location.Shelf {
				if shelf.Code == shelfCode {
					return shelf.ProductItems, nil
				}
			}
			return []models.ShelfProductBarcode{}, errors.New("shelf not found")
		}
	}
	return []models.ShelfProductBarcode{}, errors.New("location not found")
}

// Bulk ProductBarcode operations implementations
func (svc WarehouseHttpService) BulkAddProductBarcodesToShelf(shopID, authUsername, warehouseCode, locationCode, shelfCode string, products []models.ShelfProductBarcode) models.BulkProductOperationResponse {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	response := models.BulkProductOperationResponse{
		TotalCount: len(products),
		Results:    []models.BulkOperationResult{},
		Errors:     []models.BulkOperationError{},
	}

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", warehouseCode)
	if err != nil {
		// Return error response for all products
		for i, product := range products {
			response.Errors = append(response.Errors, models.BulkOperationError{
				Index:        i,
				ProductGuid:  product.GuidFixed,
				ErrorMessage: "Warehouse not found: " + err.Error(),
			})
		}
		response.FailedCount = len(products)
		response.Success = false
		return response
	}

	if findDoc.ID == primitive.NilObjectID {
		// Return error response for all products
		for i, product := range products {
			response.Errors = append(response.Errors, models.BulkOperationError{
				Index:        i,
				ProductGuid:  product.GuidFixed,
				ErrorMessage: "Warehouse not found",
			})
		}
		response.FailedCount = len(products)
		response.Success = false
		return response
	}

	dataDoc := findDoc

	// Find and update the specific shelf
	for locationIndex, location := range *dataDoc.Location {
		if location.Code == locationCode {
			for shelfIndex, shelf := range location.Shelf {
				if shelf.Code == shelfCode {
					// Use bulk add method
					bulkResponse := (*dataDoc.Location)[locationIndex].Shelf[shelfIndex].AddMultipleProductBarcodes(products)

					dataDoc.UpdatedBy = authUsername
					dataDoc.UpdatedAt = time.Now()

					if updateErr := svc.repo.Update(ctx, shopID, findDoc.GuidFixed, dataDoc); updateErr != nil {
						// If update fails, return error for all products
						bulkResponse.Success = false
						for i := range bulkResponse.Results {
							bulkResponse.Errors = append(bulkResponse.Errors, models.BulkOperationError{
								Index:        i,
								ProductGuid:  products[i].GuidFixed,
								ErrorMessage: "Database update failed: " + updateErr.Error(),
							})
						}
						bulkResponse.FailedCount = len(products)
						bulkResponse.SuccessCount = 0
						bulkResponse.Results = []models.BulkOperationResult{}
					}

					return bulkResponse
				}
			}
			// Shelf not found
			for i, product := range products {
				response.Errors = append(response.Errors, models.BulkOperationError{
					Index:        i,
					ProductGuid:  product.GuidFixed,
					ErrorMessage: "Shelf not found",
				})
			}
			response.FailedCount = len(products)
			response.Success = false
			return response
		}
	}

	// Location not found
	for i, product := range products {
		response.Errors = append(response.Errors, models.BulkOperationError{
			Index:        i,
			ProductGuid:  product.GuidFixed,
			ErrorMessage: "Location not found",
		})
	}
	response.FailedCount = len(products)
	response.Success = false
	return response
}

func (svc WarehouseHttpService) BulkRemoveProductBarcodesFromShelf(shopID, authUsername, warehouseCode, locationCode, shelfCode string, guidFixedList []string) models.BulkProductOperationResponse {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	response := models.BulkProductOperationResponse{
		TotalCount: len(guidFixedList),
		Results:    []models.BulkOperationResult{},
		Errors:     []models.BulkOperationError{},
	}

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", warehouseCode)
	if err != nil {
		// Return error response for all products
		for i, guid := range guidFixedList {
			response.Errors = append(response.Errors, models.BulkOperationError{
				Index:        i,
				ProductGuid:  guid,
				ErrorMessage: "Warehouse not found: " + err.Error(),
			})
		}
		response.FailedCount = len(guidFixedList)
		response.Success = false
		return response
	}

	if findDoc.ID == primitive.NilObjectID {
		// Return error response for all products
		for i, guid := range guidFixedList {
			response.Errors = append(response.Errors, models.BulkOperationError{
				Index:        i,
				ProductGuid:  guid,
				ErrorMessage: "Warehouse not found",
			})
		}
		response.FailedCount = len(guidFixedList)
		response.Success = false
		return response
	}

	dataDoc := findDoc

	// Find and update the specific shelf
	for locationIndex, location := range *dataDoc.Location {
		if location.Code == locationCode {
			for shelfIndex, shelf := range location.Shelf {
				if shelf.Code == shelfCode {
					// Use bulk remove method
					bulkResponse := (*dataDoc.Location)[locationIndex].Shelf[shelfIndex].RemoveMultipleProductBarcodes(guidFixedList)

					dataDoc.UpdatedBy = authUsername
					dataDoc.UpdatedAt = time.Now()

					if updateErr := svc.repo.Update(ctx, shopID, findDoc.GuidFixed, dataDoc); updateErr != nil {
						// If update fails, return error for all products
						bulkResponse.Success = false
						bulkResponse.Errors = []models.BulkOperationError{}
						for i, guid := range guidFixedList {
							bulkResponse.Errors = append(bulkResponse.Errors, models.BulkOperationError{
								Index:        i,
								ProductGuid:  guid,
								ErrorMessage: "Database update failed: " + updateErr.Error(),
							})
						}
						bulkResponse.FailedCount = len(guidFixedList)
						bulkResponse.SuccessCount = 0
						bulkResponse.Results = []models.BulkOperationResult{}
					}

					return bulkResponse
				}
			}
			// Shelf not found
			for i, guid := range guidFixedList {
				response.Errors = append(response.Errors, models.BulkOperationError{
					Index:        i,
					ProductGuid:  guid,
					ErrorMessage: "Shelf not found",
				})
			}
			response.FailedCount = len(guidFixedList)
			response.Success = false
			return response
		}
	}

	// Location not found
	for i, guid := range guidFixedList {
		response.Errors = append(response.Errors, models.BulkOperationError{
			Index:        i,
			ProductGuid:  guid,
			ErrorMessage: "Location not found",
		})
	}
	response.FailedCount = len(guidFixedList)
	response.Success = false
	return response
}
