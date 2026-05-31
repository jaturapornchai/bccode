package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"smlcloudplatform/internal/logger"
	common_models "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/product/bom/models"
	"smlcloudplatform/internal/product/bom/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/checksum"
	"strings"
	"time"

	micromodels "smlcloudplatform/pkg/microservice/models"

	product_models "smlcloudplatform/internal/product/productbarcode/models"
	product_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	product_services "smlcloudplatform/internal/product/productbarcode/services"
	saleinvoicebom_services "smlcloudplatform/internal/transaction/saleinvoicebomprice/services"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IBOMHttpService interface {
	UpsertBOM(shopID string, authUsername string, dcoNo string, barcode string) (string, error)
	SaveRecipeBOM(shopID string, authUsername string, guid string, req models.ProductBarcodeBOMSaveRequest) (string, error)
	DeleteBOM(shopID string, guid string, authUsername string) error
	InfoBOM(shopID string, guid string) (models.ProductBarcodeBOMViewInfo, error)
	SearchBOM(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeBOMViewInfo, mongopagination.PaginationData, error)
	SearchBOMStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeBOMViewInfo, int, error)
}

type BOMHttpService struct {
	repo              repositories.IBomRepository
	repoMq            repositories.IBomMessageQueueRepository
	saleInvoiceBomSvc saleinvoicebom_services.ISaleInvoiceBomPriceService
	productRepo       product_repositories.IProductBarcodeRepository
	services.ActivityService[models.ProductBarcodeBOMViewActivity, models.ProductBarcodeBOMViewDeleteActivity]
	contextTimeout time.Duration
}

func NewBOMHttpService(repo repositories.IBomRepository, repoMq repositories.IBomMessageQueueRepository, productRepo product_repositories.IProductBarcodeRepository, saleInvoiceBomSvc saleinvoicebom_services.ISaleInvoiceBomPriceService) *BOMHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &BOMHttpService{
		repo:              repo,
		repoMq:            repoMq,
		productRepo:       productRepo,
		saleInvoiceBomSvc: saleInvoiceBomSvc,
		contextTimeout:    contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.ProductBarcodeBOMViewActivity, models.ProductBarcodeBOMViewDeleteActivity](repo)

	return insSvc
}

func (svc BOMHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc BOMHttpService) UpsertBOM(shopID string, authUsername string, docNo string, barcode string) (string, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	doc, err := svc.productRepo.FindByBarcode(ctx, shopID, barcode)
	if err != nil {
		return "", err
	}

	if len(doc.ProductBarcode.Barcode) == 0 {
		return "", fmt.Errorf("barcode not found")
	}
	if doc.ItemType == product_models.ItemTypeSet && doc.MaterialType != product_models.MaterialTypeSet {
		return "", fmt.Errorf("product set barcode %s has materialtype %d; expected 3 (Set)", barcode, doc.MaterialType)
	}

	// 1. Delete all cached BOM entries for this barcode to start with a clean state
	err = svc.repo.Delete(ctx, shopID, authUsername, map[string]interface{}{"barcode": barcode})
	if err != nil {
		return "", err
	}

	lastGUID := ""
	now := time.Now()

	// 2. Process versioned BOMs if available
	if doc.BOMs != nil && len(*doc.BOMs) > 0 {
		for _, ver := range *doc.BOMs {
			bomViewDict := map[string]*product_models.ProductBarcodeBOMView{}
			bomView := product_models.ProductBarcodeBOMView{}
			bomView.FromProductBarcode(doc.ProductBarcodeData)
			bomView.Level = 1

			if _, ok := bomViewDict[doc.Barcode]; !ok {
				bomViewDict[doc.Barcode] = &bomView
			}

			productBarcodeDict := map[string]product_models.ProductBarcodeDoc{}
			if ver.BOM != nil && len(*ver.BOM) > 0 {
				err = product_services.BuildBOMView(ctx, svc.productRepo.FindByBarcode, bomView.Level, &productBarcodeDict, &bomViewDict, shopID, ver.BOM, &bomView.BOM)
				if err != nil {
					return "", err
				}
			}

			bomBytes, err := json.Marshal(bomView)
			if err != nil {
				return "", err
			}

			docBomView := models.ProductBarcodeBOMView{}
			err = json.Unmarshal(bomBytes, &docBomView)
			if err != nil {
				return "", err
			}

			checkSumStr, err := checksum.Sum(docBomView)
			if err != nil {
				return "", err
			}

			// Determine if this version covers current time
			isCurrentUse := (ver.StartDate.Before(now) || ver.StartDate.Equal(now)) && (ver.EndDate == nil || ver.EndDate.After(now))

			guidFixed := ver.GuidFixed
			if guidFixed == "" {
				guidFixed = utils.NewGUID()
			}

			lastGUID, err = svc.createWithParams(ctx, shopID, authUsername, checkSumStr, guidFixed, ver.StartDate, ver.EndDate, isCurrentUse, docBomView)
			if err != nil {
				return lastGUID, err
			}

			// Sync sale invoice BOM price for intermediate components
			bomBarcodes := []string{}
			for tempBarcode := range productBarcodeDict {
				bomBarcodes = append(bomBarcodes, tempBarcode)
			}
			_, err = svc.saleInvoiceBomSvc.CreateSaleInvoiceBomPrice(shopID, authUsername, docNo, guidFixed, bomBarcodes)
			if err != nil {
				return lastGUID, err
			}
		}
		return lastGUID, nil
	}

	// 3. Fallback to single BOM structure if no versioned BOMs exist
	bomViewDict := map[string]*product_models.ProductBarcodeBOMView{}
	bomView := product_models.ProductBarcodeBOMView{}
	bomView.FromProductBarcode(doc.ProductBarcodeData)
	bomView.Level = 1

	if _, ok := bomViewDict[doc.Barcode]; !ok {
		bomViewDict[doc.Barcode] = &bomView
	}

	productBarcodeDict := map[string]product_models.ProductBarcodeDoc{}
	if doc.BOM != nil && len(*doc.BOM) > 0 {
		err = product_services.BuildBOMView(ctx, svc.productRepo.FindByBarcode, bomView.Level, &productBarcodeDict, &bomViewDict, shopID, doc.BOM, &bomView.BOM)
		if err != nil {
			return "", err
		}
	}

	bomBytes, err := json.Marshal(bomView)
	if err != nil {
		return "", err
	}

	docBomView := models.ProductBarcodeBOMView{}
	err = json.Unmarshal(bomBytes, &docBomView)
	if err != nil {
		return "", err
	}

	checkSumStr, err := checksum.Sum(docBomView)
	if err != nil {
		return "", err
	}

	newGUID, err := svc.createWithParams(ctx, shopID, authUsername, checkSumStr, "", time.Now(), nil, true, docBomView)
	if err != nil {
		return newGUID, err
	}

	bomBarcodes := []string{}
	for tempBarcode := range productBarcodeDict {
		bomBarcodes = append(bomBarcodes, tempBarcode)
	}
	_, err = svc.saleInvoiceBomSvc.CreateSaleInvoiceBomPrice(shopID, authUsername, docNo, newGUID, bomBarcodes)
	if err != nil {
		return newGUID, err
	}

	return newGUID, nil
}

func (svc BOMHttpService) SaveRecipeBOM(shopID string, authUsername string, guid string, req models.ProductBarcodeBOMSaveRequest) (string, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	recipeCode := strings.TrimSpace(req.Barcode)
	if recipeCode == "" {
		return "", fmt.Errorf("recipe code is required")
	}
	if req.Names == nil || len(*req.Names) == 0 {
		return "", fmt.Errorf("recipe names are required")
	}

	itemUnitCode := strings.TrimSpace(req.ItemUnitCode)
	if itemUnitCode == "" {
		itemUnitCode = "RECIPE"
	}
	itemUnitNames := req.ItemUnitNames
	if itemUnitNames == nil {
		itemUnitNames = &[]common_models.NameX{}
	}

	preparedVersions, activeBOM, startDate, endDate, err := svc.prepareRecipeBOMVersions(ctx, shopID, recipeCode, req)
	if err != nil {
		return "", err
	}

	root := models.ProductBarcodeBOMView{}
	root.BarcodeGuidFixed = guid
	root.Barcode = recipeCode
	root.RefType = "recipe"
	root.Names = req.Names
	root.ItemUnitCode = itemUnitCode
	root.ItemUnitNames = itemUnitNames
	root.Condition = true
	root.DivideValue = 1
	root.StandValue = 1
	root.Qty = 1
	root.Price = req.Price
	root.BOM = &activeBOM
	root.EmptyOnNil()

	now := time.Now()
	targetGuid := strings.TrimSpace(guid)
	if targetGuid == "" || strings.HasPrefix(targetGuid, "virtual-") {
		targetGuid = strings.TrimSpace(req.GuidFixed)
	}
	if targetGuid == "" || strings.HasPrefix(targetGuid, "virtual-") {
		targetGuid = utils.NewGUID()
	}
	root.BarcodeGuidFixed = targetGuid

	existingByCode, err := svc.repo.FindUseBOMByBarcode(ctx, shopID, recipeCode)
	if err == nil && existingByCode.GuidFixed != "" && existingByCode.GuidFixed != targetGuid {
		return "", fmt.Errorf("recipe code %s already exists", recipeCode)
	}
	if err != nil {
		errText := strings.ToLower(err.Error())
		if !strings.Contains(errText, "no documents") && !strings.Contains(errText, "not found") {
			return "", err
		}
	}

	checkSumStr, err := checksum.Sum(root)
	if err != nil {
		return "", err
	}

	docData := models.ProductBarcodeBOMViewDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = targetGuid
	docData.ProductBarcodeBOMView = root
	docData.CheckSum = checkSumStr
	docData.IsCurrentUse = true
	docData.UseInDate = now
	docData.StartDate = startDate
	docData.EndDate = endDate
	docData.BOMs = &preparedVersions
	docData.EmptyOnNil()

	isCreate := guid == "" || strings.HasPrefix(guid, "virtual-")
	if isCreate {
		docData.CreatedBy = authUsername
		docData.CreatedAt = now
		_, err = svc.repo.Create(ctx, docData)
		if err != nil {
			return "", err
		}
		go func() {
			if err := svc.repoMq.Create(docData); err != nil {
				logger.GetLogger().Error(err)
			}
		}()
		return targetGuid, nil
	}

	existing, err := svc.repo.FindByGuid(ctx, shopID, targetGuid)
	if err != nil {
		return "", err
	}
	if existing.GuidFixed == "" {
		return "", fmt.Errorf("recipe not found")
	}
	docData.CreatedBy = existing.CreatedBy
	docData.CreatedAt = existing.CreatedAt
	docData.UpdatedBy = authUsername
	docData.UpdatedAt = now

	err = svc.repo.Update(ctx, shopID, targetGuid, docData)
	if err != nil {
		return "", err
	}
	go func() {
		if err := svc.repoMq.Update(docData); err != nil {
			logger.GetLogger().Error(err)
		}
	}()

	return targetGuid, nil
}

func (svc BOMHttpService) prepareRecipeBOMVersions(
	ctx context.Context,
	shopID string,
	recipeCode string,
	req models.ProductBarcodeBOMSaveRequest,
) ([]models.ProductBarcodeBOMVersion, []models.ProductBarcodeBOMView, time.Time, *time.Time, error) {
	now := time.Now()
	if len(req.BOMs) == 0 {
		items, err := svc.prepareRecipeBOMItems(ctx, shopID, recipeCode, req.BOM)
		if err != nil {
			return nil, nil, now, nil, err
		}
		return []models.ProductBarcodeBOMVersion{}, items, now, nil, nil
	}

	versions := []models.ProductBarcodeBOMVersion{}
	activeIndex := -1
	for idx, ver := range req.BOMs {
		startDate := ver.StartDate
		if startDate.IsZero() {
			startDate = now
		}

		rawItems := []models.ProductBarcodeBOMView{}
		if ver.BOM != nil {
			rawItems = *ver.BOM
		}
		items, err := svc.prepareRecipeBOMItems(ctx, shopID, recipeCode, rawItems)
		if err != nil {
			return nil, nil, now, nil, err
		}

		guidFixed := strings.TrimSpace(ver.GuidFixed)
		if guidFixed == "" {
			guidFixed = utils.NewGUID()
		}
		itemsPtr := &items
		versions = append(versions, models.ProductBarcodeBOMVersion{
			GuidFixed: guidFixed,
			StartDate: startDate,
			EndDate:   ver.EndDate,
			BOM:       itemsPtr,
		})

		if (startDate.Before(now) || startDate.Equal(now)) && (ver.EndDate == nil || ver.EndDate.After(now)) {
			activeIndex = idx
		}
	}

	if len(versions) == 0 {
		return []models.ProductBarcodeBOMVersion{}, []models.ProductBarcodeBOMView{}, now, nil, nil
	}
	if activeIndex < 0 {
		activeIndex = len(versions) - 1
	}

	activeItems := []models.ProductBarcodeBOMView{}
	if versions[activeIndex].BOM != nil {
		activeItems = *versions[activeIndex].BOM
	}
	return versions, activeItems, versions[activeIndex].StartDate, versions[activeIndex].EndDate, nil
}

func (svc BOMHttpService) prepareRecipeBOMItems(
	ctx context.Context,
	shopID string,
	recipeCode string,
	reqItems []models.ProductBarcodeBOMView,
) ([]models.ProductBarcodeBOMView, error) {
	items := []models.ProductBarcodeBOMView{}
	for _, reqItem := range reqItems {
		barcode := strings.TrimSpace(reqItem.Barcode)
		if barcode == "" {
			return nil, fmt.Errorf("component barcode or recipe code is required")
		}

		refType := strings.ToLower(strings.TrimSpace(reqItem.RefType))
		if refType == "" {
			refType = "product"
		}
		if refType == "recipe" && barcode == recipeCode {
			return nil, fmt.Errorf("recipe %s cannot reference itself", recipeCode)
		}

		normalized := normalizeRecipeBOMRequestItem(reqItem)
		var item models.ProductBarcodeBOMView

		switch refType {
		case "recipe":
			recipeDoc, err := svc.repo.FindUseBOMByBarcode(ctx, shopID, barcode)
			if err != nil {
				return nil, err
			}
			if recipeDoc.GuidFixed == "" {
				return nil, fmt.Errorf("recipe %s not found", barcode)
			}
			item = recipeDoc.ProductBarcodeBOMView
			item.BarcodeGuidFixed = recipeDoc.GuidFixed
			item.RefType = "recipe"
			item.Barcode = barcode
			if item.BOM == nil {
				empty := []models.ProductBarcodeBOMView{}
				item.BOM = &empty
			}
		case "product":
			productDoc, err := svc.productRepo.FindByBarcode(ctx, shopID, barcode)
			if err != nil {
				return nil, err
			}
			if productDoc.Barcode == "" {
				return nil, fmt.Errorf("product barcode %s not found", barcode)
			}
			if !isAllowedRecipeComponentMaterialType(productDoc.MaterialType) {
				return nil, fmt.Errorf("recipe component barcode %s has materialtype %d; allowed values are 1=Material, 2=Semi-Finished, 4=Agricultural", barcode, productDoc.MaterialType)
			}
			item = models.ProductBarcodeBOMView{}
			item.BarcodeGuidFixed = productDoc.GuidFixed
			item.RefType = "product"
			item.Barcode = productDoc.Barcode
			item.Names = productDoc.Names
			item.ItemUnitCode = productDoc.ItemUnitCode
			item.ItemUnitNames = productDoc.ItemUnitNames
			item.MaterialType = productDoc.MaterialType
			empty := []models.ProductBarcodeBOMView{}
			item.BOM = &empty
		default:
			return nil, fmt.Errorf("invalid component ref_type %s", refType)
		}

		item.Level = normalized.Level
		item.Condition = normalized.Condition
		item.DivideValue = normalized.DivideValue
		item.StandValue = normalized.StandValue
		item.Qty = normalized.Qty
		item.YieldPercent = normalized.YieldPercent
		item.AverageCost = normalized.AverageCost
		item.Price = normalized.Price
		item.EmptyOnNil()
		items = append(items, item)
	}
	return items, nil
}

func isAllowedRecipeComponentMaterialType(materialType int8) bool {
	switch materialType {
	case product_models.MaterialTypeMaterial, product_models.MaterialTypeSemiFinished, product_models.MaterialTypeAgricultural:
		return true
	default:
		return false
	}
}

func normalizeRecipeBOMRequestItem(item models.ProductBarcodeBOMView) models.ProductBarcodeBOMView {
	if item.Level == 0 {
		item.Level = 2
	}
	if item.DivideValue == 0 {
		item.DivideValue = 1
	}
	if item.StandValue == 0 {
		item.StandValue = 1
	}
	if item.Qty == 0 {
		item.Qty = 1
	}
	if item.YieldPercent == 0 {
		item.YieldPercent = 100
	}
	item.Condition = true
	return item
}

func (svc BOMHttpService) clearUseBOMByBarcode(ctx context.Context, shopID string, barcode string) error {
	return svc.repo.ClearUseBOMByBarcode(ctx, shopID, barcode)
}

func (svc BOMHttpService) createWithParams(
	ctx context.Context,
	shopID string,
	authUsername string,
	checkSum string,
	guidFixed string,
	startDate time.Time,
	endDate *time.Time,
	isCurrentUse bool,
	doc models.ProductBarcodeBOMView,
) (string, error) {
	currentDate := time.Now()
	if guidFixed == "" {
		guidFixed = utils.NewGUID()
	}

	docData := models.ProductBarcodeBOMViewDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = guidFixed
	docData.ProductBarcodeBOMView = doc
	docData.CheckSum = checkSum
	docData.IsCurrentUse = isCurrentUse
	docData.UseInDate = currentDate
	docData.StartDate = startDate
	docData.EndDate = endDate

	docData.EmptyOnNil()

	docData.CreatedBy = authUsername
	docData.CreatedAt = currentDate

	_, err := svc.repo.Create(ctx, docData)
	if err != nil {
		return "", err
	}

	go func() {
		err := svc.repoMq.Create(docData)
		if err != nil {
			logger.GetLogger().Error(err)
		}
	}()

	return guidFixed, nil
}

func (svc BOMHttpService) create(ctx context.Context, shopID string, authUsername string, checkSum string, doc models.ProductBarcodeBOMView) (string, error) {

	currentDate := time.Now()
	newGuidFixed := utils.NewGUID()

	docData := models.ProductBarcodeBOMViewDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.ProductBarcodeBOMView = doc
	docData.CheckSum = checkSum
	docData.IsCurrentUse = true
	docData.UseInDate = currentDate

	docData.EmptyOnNil()

	docData.CreatedBy = authUsername
	docData.CreatedAt = currentDate

	_, err := svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	go func() {
		err := svc.repoMq.Create(docData)

		if err != nil {
			logger.GetLogger().Error(err)
		}
	}()

	return newGuidFixed, nil
}

func (svc BOMHttpService) DeleteBOM(shopID string, guid string, authUsername string) error {

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
		err := svc.repoMq.Delete(findDoc)

		if err != nil {
			logger.GetLogger().Error(err)
		}
	}()

	return nil
}

func (svc BOMHttpService) InfoBOM(shopID string, guid string) (models.ProductBarcodeBOMViewInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.ProductBarcodeBOMViewInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.ProductBarcodeBOMViewInfo{}, errors.New("document not found")
	}

	return findDoc.ProductBarcodeBOMViewInfo, nil

}

func (svc BOMHttpService) SearchBOM(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeBOMViewInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"barcode",
		"names.name",
	}

	if len(pageable.Sorts) == 0 {
		pageable.Sorts = []micromodels.KeyInt{
			{Key: "barcode", Value: 1},
			{Key: "level", Value: 1},
		}
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.ProductBarcodeBOMViewInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc BOMHttpService) SearchBOMStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeBOMViewInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"barcode",
		"names.name",
	}

	if len(pageableStep.Sorts) == 0 {
		pageableStep.Sorts = []micromodels.KeyInt{
			{Key: "code", Value: 1},
		}
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.ProductBarcodeBOMViewInfo{}, 0, err
	}

	return docList, total, nil
}
