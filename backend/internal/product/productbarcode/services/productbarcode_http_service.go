package services

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	creditorRepo "smlcloudplatform/internal/debtaccount/creditor/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productmaster "smlcloudplatform/internal/product/product/repositories"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/product/productbarcode/repositories"
	productcategory_models "smlcloudplatform/internal/product/productcategory/models"
	productcategory_services "smlcloudplatform/internal/product/productcategory/services"
	unit_models "smlcloudplatform/internal/product/unit/models"
	unitmaster "smlcloudplatform/internal/product/unit/repositories"
	unitservices "smlcloudplatform/internal/product/unit/services"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	warehouseRepo "smlcloudplatform/internal/warehouse/repositories"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IProductBarcodeHttpService interface {
	CreateProductBarcode(shopID string, authUsername string, doc models.ProductBarcodeRequest) (string, error)
	UpdateProductBarcode(shopID string, guid string, authUsername string, doc models.ProductBarcodeRequest) error
	DeleteProductBarcode(shopID string, guid string, authUsername string) error
	DeleteProductBarcodeByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoProductBarcode(shopID string, guid string) (models.ProductBarcodeInfo, error)
	InfoProductBarcodeByBarcode(shopID string, barcode string) (models.ProductBarcodeInfo, error)
	InfoWTFArray(shopID string, codes []string) ([]interface{}, error)
	InfoWTFArrayMaster(codes []string) ([]interface{}, error)
	GetProductBarcodeByBarcodeRef(shopID string, barcodeRef string) ([]models.ProductBarcodeInfo, error)
	GetProductBarcodeByBarcodeRefMultiShops(shopIDs []string, barcodeRef string) ([]models.ProductBarcodeInfo, error)
	SearchProductBarcode(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)
	SearchProductBarcode2(shopID string, pageable micromodels.Pageable) ([]models.ProductBarcodeSearch, common.Pagination, error)

	SearchProductBarcodeStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeInfo, int, error)
	SearchProductBarcodeStepMultiShops(langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.ProductBarcode) (common.BulkImport, error)

	XSortsSave(shopID string, authUsername string, xsorts []common.XSortModifyReqesut) error
	GetProductBarcodeByBarcodes(shopID string, barcodes []string) ([]models.ProductBarcodeInfo, error)

	// Price History methods
	GetPriceHistory(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error)
	GetPriceHistoryByBarcode(shopID string, barcode string, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error)

	UpdateProductBarcodeBranch(shopID string, authUsername string, branch models.ProductBarcodeBranch, productBarcodeGUIDFixedes []string) error
	UpdateProductBarcodeBusinessType(shopID string, authUsername string, businessType models.ProductBarcodeBusinessType, productBarcodeGUIDFixedes []string) error

	GetModuleName() string
	GetProductBarcodeByUnits(shopID string, unitCodes []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)
	GetProductBarcodeByGroups(shopID string, groupCodes []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)
	Export(shopID string, languageCode string, languageHeader map[string]string) ([][]string, error)

	InfoBomView(shopID string, barcode string) (models.ProductBarcodeBOMView, error)
	Import(shopID string, authUsername string, dataList []models.ProductBarcode) (common.BulkImport, error)
	ImportRefBarcodeUpdate(shopID, authUsername string, requests []models.RefBarcodeImportRequest) (*models.RefBarcodeImportResponse, error)
}

type ProductBarcodeHttpService struct {
	repo            repositories.IProductBarcodeRepository
	repoMaster      productmaster.IProductRepository
	repoUnit        unitmaster.IUnitRepository
	unitSvc         unitservices.IUnitHttpService
	repomgCreditror creditorRepo.CreditorRepository
	chRepo          repositories.IProductBarcodeClickhouseRepository
	syncCacheRepo   mastersync.IMasterSyncCacheRepository
	mqRepo          repositories.IProductBarcodeMessageQueueRepository
	categorySvc     productcategory_services.IProductCategoryHttpService
	priceHistorySvc IProductPriceHistoryService
	warehouseRepo   warehouseRepo.IWarehouseRepository
	services.ActivityService[models.ProductBarcodeActivity, models.ProductBarcodeDeleteActivity]
	contextTimeout time.Duration
}

func NewProductBarcodeHttpService(
	repo repositories.IProductBarcodeRepository,
	repoMaster productmaster.IProductRepository,
	repoUnit unitmaster.IUnitRepository,
	unitSvc unitservices.IUnitHttpService,
	repomgCreditror creditorRepo.CreditorRepository,
	mqRepo repositories.IProductBarcodeMessageQueueRepository,
	chRepo repositories.IProductBarcodeClickhouseRepository,
	categorySvc productcategory_services.IProductCategoryHttpService,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	priceHistorySvc IProductPriceHistoryService,
	warehouseRepo warehouseRepo.IWarehouseRepository,
) *ProductBarcodeHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &ProductBarcodeHttpService{
		repo:            repo,
		repoMaster:      repoMaster,
		repoUnit:        repoUnit,
		unitSvc:         unitSvc,
		repomgCreditror: repomgCreditror,
		chRepo:          chRepo,
		syncCacheRepo:   syncCacheRepo,
		mqRepo:          mqRepo,
		categorySvc:     categorySvc,
		priceHistorySvc: priceHistorySvc,
		warehouseRepo:   warehouseRepo,
		contextTimeout:  contextTimeout,
	}
	insSvc.ActivityService = services.NewActivityService[models.ProductBarcodeActivity, models.ProductBarcodeDeleteActivity](repo)
	return insSvc
}

func (svc ProductBarcodeHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ProductBarcodeHttpService) CreateProductBarcode(shopID string, authUsername string, docReq models.ProductBarcodeRequest) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "barcode", docReq.Barcode)

	if err != nil {
		return "", err
	}

	if findDoc.Barcode != "" {
		return "", errors.New("barcode is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.ProductBarcodeDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.ProductBarcode = docReq.ToProductBarcode()
	docData.IgnoreBranches = &docReq.IgnoreBranches
	docData.BusinessTypes = &docReq.BusinessTypes

	// Unit validation and creation logic
	if docReq.ItemUnitCode != "" {
		// Check if unit exists in master units
		unitDoc, err := svc.repoUnit.FindByDocIndentityGuid(ctx, shopID, "unitcode", docReq.ItemUnitCode)
		if err != nil {
			return "", fmt.Errorf("error checking unit: %v", err)
		}

		if unitDoc.UnitCode != "" {
			// Unit exists, use its data
			docData.ItemUnitGuid = unitDoc.GuidFixed
			docData.ItemUnitNames = unitDoc.Names
		} else {
			// Unit doesn't exist, create new unit
			newUnit := unit_models.Unit{
				UnitCode: docReq.ItemUnitCode,
				Names:    docReq.ItemUnitNames,
				UnitName: common.UnitName{
					UnitName1: docReq.ItemUnitCode, // Use unit code as default name
				},
			}

			// Create the unit
			unitGuid, err := svc.unitSvc.CreateUnit(shopID, authUsername, newUnit)
			if err != nil {
				return "", fmt.Errorf("error creating unit: %v", err)
			}

			// Use the created unit data
			docData.ItemUnitGuid = unitGuid
			docData.ItemUnitNames = docReq.ItemUnitNames
		}
	}

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	if docReq.Options != nil {
		options := *docReq.Options
		for idxOpt := range options {
			option := &options[idxOpt]
			if len(option.GUID) < 1 {
				option.GUID = utils.NewGUID()
			}

			choices := *option.Choices
			for idxChoice := range choices {
				choice := &choices[idxChoice]
				if len(choice.GUID) < 1 {
					choice.GUID = utils.NewGUID()
				}
			}
		}
	}

	docData.RefBarcodes, err = svc.prepareRefBarcode(ctx, shopID, docReq.RefBarcodes)

	if err != nil {
		return "", err
	}

	docData.BOM, err = svc.prepareBOM(ctx, shopID, docReq.BOM)

	if err != nil {
		return "", err
	}

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	err = svc.mqRepo.Create(docData)
	if err != nil {
		return "", err
	}

	// บันทึกประวัติราคาสำหรับการสร้างใหม่
	if docData.Prices != nil && len(*docData.Prices) > 0 {
		productName := ""
		if docData.Names != nil && len(*docData.Names) > 0 {
			if (*docData.Names)[0].Name != nil {
				productName = *(*docData.Names)[0].Name
			}
		}

		err = svc.priceHistorySvc.RecordPriceChange(
			ctx,
			shopID,
			newGuidFixed,
			docData.Barcode,
			productName,
			[]models.ProductPrice{}, // ไม่มีราคาเก่าสำหรับการสร้างใหม่
			*docData.Prices,
			"create",
			authUsername,
			"สร้างสินค้าใหม่",
		)
		if err != nil {
			// Log error แต่ไม่ return เพื่อไม่ให้การสร้างสินค้าล้มเหลว
			// TODO: Add proper logging
		}
	}

	svc.saveMasterSync(shopID)

	return newGuidFixed, nil
}

func (svc ProductBarcodeHttpService) UpdateProductBarcode(shopID string, guid string, authUsername string, docReq models.ProductBarcodeRequest) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	docData := findDoc

	// เก็บราคาเก่าไว้เพื่อเปรียบเทียบ
	oldPrices := []models.ProductPrice{}
	if findDoc.Prices != nil {
		oldPrices = *findDoc.Prices
	}

	docData.ProductBarcode = docReq.ToProductBarcode()

	docData.Barcode = findDoc.Barcode
	docData.IgnoreBranches = &docReq.IgnoreBranches
	docData.BusinessTypes = &docReq.BusinessTypes

	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	docData.RefBarcodes, err = svc.prepareRefBarcode(ctx, shopID, docReq.RefBarcodes)

	if err != nil {
		return err
	}

	docData.BOM, err = svc.prepareBOM(ctx, shopID, docReq.BOM)

	if err != nil {
		return err
	}

	err = svc.updateMetaInRefBarcode(ctx, shopID, docData)

	if err != nil {
		return err
	}

	err = svc.updateMetaInBOMBarcode(ctx, shopID, docData)

	if err != nil {
		return err
	}

	// print docData
	fmt.Printf("Updated docData: %+v\n", docData)

	err = svc.repo.Update(ctx, shopID, guid, docData)

	if err != nil {
		return err
	}

	err = svc.mqRepo.Update(docData)
	if err != nil {
		return err
	}

	// บันทึกประวัติการเปลี่ยนแปลงราคา
	newPrices := []models.ProductPrice{}
	if docData.Prices != nil {
		newPrices = *docData.Prices
	}

	productName := ""
	if docData.Names != nil && len(*docData.Names) > 0 {
		if (*docData.Names)[0].Name != nil {
			productName = *(*docData.Names)[0].Name
		}
	}

	err = svc.priceHistorySvc.RecordPriceChange(
		ctx,
		shopID,
		guid,
		docData.Barcode,
		productName,
		oldPrices,
		newPrices,
		"update",
		authUsername,
		"แก้ไขข้อมูลสินค้า",
	)
	if err != nil {
		// Log error แต่ไม่ return เพื่อไม่ให้การอัปเดตล้มเหลว
		// TODO: Add proper logging
	}

	categoryBarcode := productcategory_models.CodeXSort{
		Barcode:          docData.Barcode,
		Code:             docData.ItemCode,
		Names:            docData.Names,
		UnitCode:         docData.ItemUnitCode,
		UnitNames:        docData.ItemUnitNames,
		ManufacturerGUID: docData.ManufacturerGUID,
	}

	svc.categorySvc.UpdateBarcode(shopID, categoryBarcode)

	svc.saveMasterSync(shopID)

	return nil
}

func (svc ProductBarcodeHttpService) UpdateProductBarcodeBranch(shopID string, authUsername string, branch models.ProductBarcodeBranch, productBarcodeGUIDFixedes []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repo.UpdateBranch(ctx, shopID, branch, productBarcodeGUIDFixedes)

	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc ProductBarcodeHttpService) UpdateProductBarcodeBusinessType(shopID string, authUsername string, businessType models.ProductBarcodeBusinessType, productBarcodeGUIDFixedes []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repo.UpdateBusinessType(ctx, shopID, businessType, productBarcodeGUIDFixedes)

	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc ProductBarcodeHttpService) prepareRefBarcode(ctx context.Context, shopID string, barcodes []models.BarcodeRequest) (*[]models.RefProductBarcode, error) {
	tempChildrenBarcodes := []string{}
	tempRefBarcodes := map[string]models.BarcodeRequest{}

	for _, item := range barcodes {
		tempChildrenBarcodes = append(tempChildrenBarcodes, item.Barcode)
		tempRefBarcodes[item.Barcode] = item
	}

	findChildrenDocs, err := svc.repo.FindByDocIndentityGuids(ctx, shopID, "barcode", tempChildrenBarcodes)

	if err != nil {
		return &[]models.RefProductBarcode{}, err
	}

	tempBarcodes := []models.RefProductBarcode{}
	for _, childDoc := range findChildrenDocs {
		tempRef := childDoc.ToRefBarcode()

		tempRef.Condition = tempRefBarcodes[tempRef.Barcode].Condition
		tempRef.StandValue = tempRefBarcodes[tempRef.Barcode].StandValue
		tempRef.DivideValue = tempRefBarcodes[tempRef.Barcode].DivideValue
		tempRef.Qty = tempRefBarcodes[tempRef.Barcode].Qty

		tempBarcodes = append(tempBarcodes, tempRef)
	}

	return &tempBarcodes, nil
}

func (svc ProductBarcodeHttpService) prepareBOM(ctx context.Context, shopID string, barcodes []models.BOMRequest) (*[]models.BOMProductBarcode, error) {
	tempChildrenBarcodes := []string{}
	tempBOM := map[string]models.BOMRequest{}

	for _, item := range barcodes {
		tempChildrenBarcodes = append(tempChildrenBarcodes, item.Barcode)
		tempBOM[item.Barcode] = item
	}

	findChildrenDocs, err := svc.repo.FindByDocIndentityGuids(ctx, shopID, "barcode", tempChildrenBarcodes)

	if err != nil {
		return &[]models.BOMProductBarcode{}, err
	}

	tempBarcodes := []models.BOMProductBarcode{}
	for _, childDoc := range findChildrenDocs {
		temp := childDoc.ToBOM()

		temp.Condition = tempBOM[temp.Barcode].Condition
		temp.StandValue = tempBOM[temp.Barcode].StandValue
		temp.DivideValue = tempBOM[temp.Barcode].DivideValue
		temp.Qty = tempBOM[temp.Barcode].Qty

		tempBarcodes = append(tempBarcodes, temp)
	}

	return &tempBarcodes, nil
}

func (svc ProductBarcodeHttpService) updateMetaInRefBarcode(ctx context.Context, shopID string, docData models.ProductBarcodeDoc) error {

	findDocs, err := svc.repo.FindByRefBarcode(ctx, shopID, docData.Barcode)
	if err != nil {
		return err
	}

	for _, findDoc := range findDocs {
		tempRefBarcodes := []models.RefProductBarcode{}
		for _, refBarcode := range *findDoc.RefBarcodes {
			if refBarcode.Barcode == docData.Barcode {
				// เก็บค่าเดิมของ DivideValue และ StandValue ก่อนการอัพเดท
				originalDivideValue := refBarcode.DivideValue
				originalStandValue := refBarcode.StandValue
				originalQty := refBarcode.Qty
				originalCondition := refBarcode.Condition

				// Debug log
				// fmt.Printf("Debug updateMetaInRefBarcode: barcode=%s, originalDivideValue=%f, originalStandValue=%f\n",
				// 	refBarcode.Barcode, originalDivideValue, originalStandValue)

				// อัพเดทเฉพาะข้อมูลที่ต้องการ (metadata เท่านั้น)
				refBarcode.Names = docData.Names
				refBarcode.ItemUnitCode = docData.ItemUnitCode
				refBarcode.ItemUnitNames = docData.ItemUnitNames

				// คืนค่าเดิมให้กับ calculation values ที่ต้องการเก็บไว้
				refBarcode.DivideValue = originalDivideValue
				refBarcode.StandValue = originalStandValue
				refBarcode.Qty = originalQty
				refBarcode.Condition = originalCondition

				// Debug log after restore
				// fmt.Printf("Debug after restore: barcode=%s, divideValue=%f, standValue=%f\n",
				// 	refBarcode.Barcode, refBarcode.DivideValue, refBarcode.StandValue)
			}

			tempRefBarcodes = append(tempRefBarcodes, refBarcode)
		}

		findDoc.RefBarcodes = &tempRefBarcodes
		// fmt.Printf("findDoc.RefBarcodes :%v\n",
		// 	findDoc.RefBarcodes)

		err = svc.repo.Update(ctx, shopID, findDoc.GuidFixed, findDoc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (svc ProductBarcodeHttpService) updateMetaInBOMBarcode(ctx context.Context, shopID string, docData models.ProductBarcodeDoc) error {

	findDocs, err := svc.repo.FindByBOMBarcode(ctx, shopID, docData.Barcode)
	if err != nil {
		return err
	}

	for _, findDoc := range findDocs {
		tempBOMBarcodes := []models.BOMProductBarcode{}
		for _, refBarcode := range *findDoc.BOM {
			if refBarcode.Barcode == docData.Barcode {
				// เก็บค่าเดิมของ DivideValue และ StandValue ก่อนการอัพเดท
				originalDivideValue := refBarcode.DivideValue
				originalStandValue := refBarcode.StandValue
				originalQty := refBarcode.Qty
				originalCondition := refBarcode.Condition

				// Debug log
				fmt.Printf("Debug updateMetaInBOMBarcode: barcode=%s, originalDivideValue=%f, originalStandValue=%f\n",
					refBarcode.Barcode, originalDivideValue, originalStandValue)

				// อัพเดทเฉพาะข้อมูลที่ต้องการ (metadata เท่านั้น)
				refBarcode.Names = docData.Names
				refBarcode.ItemUnitCode = docData.ItemUnitCode
				refBarcode.ItemUnitNames = docData.ItemUnitNames

				// คืนค่าเดิมให้กับ calculation values ที่ต้องการเก็บไว้
				refBarcode.DivideValue = originalDivideValue
				refBarcode.StandValue = originalStandValue
				refBarcode.Qty = originalQty
				refBarcode.Condition = originalCondition

				// Debug log after restore
				fmt.Printf("Debug BOM after restore: barcode=%s, divideValue=%f, standValue=%f\n",
					refBarcode.Barcode, refBarcode.DivideValue, refBarcode.StandValue)
			}

			tempBOMBarcodes = append(tempBOMBarcodes, refBarcode)
		}

		findDoc.BOM = &tempBOMBarcodes

		err = svc.repo.Update(ctx, shopID, findDoc.GuidFixed, findDoc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (svc ProductBarcodeHttpService) DeleteProductBarcode(shopID string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	countRef, err := svc.repo.CountByRefBarcode(ctx, shopID, findDoc.Barcode)

	if err != nil {
		return err
	}

	if countRef > 0 {
		return errors.New("document has other ref barcode referenced")
	}

	countBOM, err := svc.repo.CountByBOM(ctx, shopID, findDoc.Barcode)

	if err != nil {
		return err
	}

	if countBOM > 0 {
		return errors.New("document has other bom referenced")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)
	if err != nil {
		return err
	}

	err = svc.mqRepo.Delete(findDoc)
	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc ProductBarcodeHttpService) InfoProductBarcode(shopID string, guid string) (models.ProductBarcodeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)
	if err != nil {
		return models.ProductBarcodeInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.ProductBarcodeInfo{}, errors.New("document not found")
	}

	// ✅ ตรวจสอบว่า ItemGuid ไม่ใช่ค่าว่างก่อนดึงข้อมูลจาก Master
	if strings.TrimSpace(findDoc.ItemGuid) != "" {
		findMasterDoc, err := svc.repoMaster.FindByGuid(ctx, shopID, findDoc.ItemGuid)
		if err != nil {
			fmt.Printf("Error fetching master data: %v\n", err)
		} else {
			// ✅ ตรวจสอบค่า `findMasterDoc.Names` ก่อนใช้งาน
			tempName := []common.NameX{}
			if findMasterDoc.Names != nil {
				for _, name := range *findMasterDoc.Names {
					tempName = append(tempName, common.NameX{
						Name: name.Name,
						Code: name.Code,
					})
				}
			}

			// ✅ ตรวจสอบค่า `findMasterDoc.GroupName` ก่อนใช้งาน
			tempGroupNames := []common.NameX{}
			if findMasterDoc.GroupNames != nil {
				for _, name := range *findMasterDoc.GroupNames {
					tempGroupNames = append(tempGroupNames, common.NameX{
						Name: name.Name,
						Code: name.Code,
					})
				}
			}

			// ✅ กำหนดค่าให้ findDoc โดยเช็คค่าว่างก่อนใช้งาน
			findDoc.ItemCode = findMasterDoc.Code
			findDoc.Names = &tempName

			if findMasterDoc.ManufacturerGUID != "" {
				findManu, err := svc.repomgCreditror.FindByGuid(ctx, shopID, findMasterDoc.ManufacturerGUID)
				if err == nil { // ไม่คืนค่า error ถ้าไม่เจอข้อมูล
					findDoc.ManufacturerGUID = findManu.GuidFixed
					findDoc.ManufacturerCode = findManu.Code
					findDoc.ManufacturerNames = findManu.Names
				}
			} else {
				findDoc.ManufacturerCode = ""
				findDoc.ManufacturerNames = &[]common.NameX{}
				findDoc.ManufacturerGUID = ""
			}

			if findMasterDoc.GroupCode != "" {
				findDoc.GroupCode = findMasterDoc.GroupCode
			} else {
				findDoc.GroupCode = ""
			}

			findDoc.GroupNames = &tempGroupNames
			findDoc.ItemType = findMasterDoc.ItemType
		}
	}

	if strings.TrimSpace(findDoc.ItemUnitGuid) != "" {
		unit, err := svc.repoUnit.FindByGuid(ctx, shopID, findDoc.ItemUnitGuid)
		if err != nil {
			findDoc.ItemUnitNames = &[]common.NameX{}
		}

		namex := []common.NameX{}
		for _, name := range *unit.Names {
			namex = append(namex, common.NameX{
				Name: name.Name,
				Code: name.Code,
			})
		}
		findDoc.ItemUnitCode = unit.UnitCode
		findDoc.ItemUnitNames = &namex // ✅ กำหนดค่าเฉพาะเมื่อ unit มีข้อมูล
	}

	return findDoc.ProductBarcodeInfo, nil
}

func (svc ProductBarcodeHttpService) InfoProductBarcodeByBarcode(shopID string, barcode string) (models.ProductBarcodeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "barcode", barcode)

	if err != nil {
		return models.ProductBarcodeInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.ProductBarcodeInfo{}, errors.New("document not found")
	}

	// ✅ ตรวจสอบว่า ItemGuid ไม่ใช่ค่าว่างก่อนดึงข้อมูลจาก Master
	if strings.TrimSpace(findDoc.ItemGuid) != "" {
		findMasterDoc, err := svc.repoMaster.FindByGuid(ctx, shopID, findDoc.ItemGuid)
		if err != nil {
			fmt.Printf("Error fetching master data: %v\n", err)
		} else {
			// ✅ ตรวจสอบค่า `findMasterDoc.Names` ก่อนใช้งาน
			tempName := []common.NameX{}
			if findMasterDoc.Names != nil {
				for _, name := range *findMasterDoc.Names {
					tempName = append(tempName, common.NameX{
						Name: name.Name,
						Code: name.Code,
					})
				}
			}

			// ✅ ตรวจสอบค่า `findMasterDoc.GroupName` ก่อนใช้งาน
			tempGroupNames := []common.NameX{}
			if findMasterDoc.GroupNames != nil {
				for _, name := range *findMasterDoc.GroupNames {
					tempGroupNames = append(tempGroupNames, common.NameX{
						Name: name.Name,
						Code: name.Code,
					})
				}
			}

			// ✅ กำหนดค่าให้ findDoc โดยเช็คค่าว่างก่อนใช้งาน
			findDoc.ItemCode = findMasterDoc.Code
			findDoc.Names = &tempName

			if findMasterDoc.ManufacturerGUID != "" {
				findManu, err := svc.repomgCreditror.FindByGuid(ctx, shopID, findMasterDoc.ManufacturerGUID)
				if err == nil { // ไม่คืนค่า error ถ้าไม่เจอข้อมูล
					findDoc.ManufacturerGUID = findManu.GuidFixed
					findDoc.ManufacturerCode = findManu.Code
					findDoc.ManufacturerNames = findManu.Names
				}
			} else {
				findDoc.ManufacturerCode = ""
				findDoc.ManufacturerNames = &[]common.NameX{}
				findDoc.ManufacturerGUID = ""
			}

			if findMasterDoc.GroupCode != "" {
				findDoc.GroupCode = findMasterDoc.GroupCode
			} else {
				findDoc.GroupCode = ""
			}

			findDoc.GroupNames = &tempGroupNames
			findDoc.ItemType = findMasterDoc.ItemType
		}
	}

	if strings.TrimSpace(findDoc.ItemUnitGuid) != "" {
		unit, err := svc.repoUnit.FindByGuid(ctx, shopID, findDoc.ItemUnitGuid)
		if err != nil {
			findDoc.ItemUnitNames = &[]common.NameX{}
		}

		namex := []common.NameX{}
		for _, name := range *unit.Names {
			namex = append(namex, common.NameX{
				Name: name.Name,
				Code: name.Code,
			})
		}
		findDoc.ItemUnitCode = unit.UnitCode
		findDoc.ItemUnitNames = &namex // ✅ กำหนดค่าเฉพาะเมื่อ unit มีข้อมูล
	}

	return findDoc.ProductBarcodeInfo, nil
}

func (svc ProductBarcodeHttpService) InfoWTFArray(shopID string, codes []string) ([]interface{}, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList := []interface{}{}

	for _, code := range codes {
		findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "barcode", code)
		if err != nil || findDoc.ID == primitive.NilObjectID {
			// add item empty
			docList = append(docList, nil)
		} else {
			docList = append(docList, findDoc.ProductBarcodeInfo)
		}
	}

	return docList, nil
}

func (svc ProductBarcodeHttpService) InfoWTFArrayMaster(codes []string) ([]interface{}, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList := []interface{}{}

	findDocList, err := svc.repo.FindMasterInCodes(ctx, codes)

	if err != nil {
		return []interface{}{}, err
	}

	for _, code := range codes {
		findDoc, ok := lo.Find(findDocList, func(item models.ProductBarcodeInfo) bool {
			return item.Barcode == code
		})
		if !ok {
			// add item empty
			docList = append(docList, nil)
		} else {
			docList = append(docList, findDoc)
		}
	}

	return docList, nil
}

func (svc ProductBarcodeHttpService) GetProductBarcodeByBarcodeRef(shopID string, barcodeRef string) ([]models.ProductBarcodeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDocs, err := svc.repo.FindByRefBarcode(ctx, shopID, barcodeRef)
	if err != nil {
		return nil, err
	}

	resultDocs := []models.ProductBarcodeInfo{}
	for _, findDoc := range findDocs {
		resultDocs = append(resultDocs, findDoc.ProductBarcodeInfo)
	}

	return resultDocs, nil
}

func (svc ProductBarcodeHttpService) GetProductBarcodeByBarcodeRefMultiShops(shopIDs []string, barcodeRef string) ([]models.ProductBarcodeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	resultDocs := []models.ProductBarcodeInfo{}

	// Search across all specified shops
	for _, shopID := range shopIDs {
		findDocs, err := svc.repo.FindByRefBarcode(ctx, shopID, barcodeRef)
		if err != nil {
			continue // Continue to next shop if error occurs
		}

		for _, findDoc := range findDocs {
			resultDocs = append(resultDocs, findDoc.ProductBarcodeInfo)
		}
	}

	return resultDocs, nil
}

func (svc ProductBarcodeHttpService) GetProductBarcodeByBarcodes(shopID string, barcodes []string) ([]models.ProductBarcodeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	results, err := svc.repo.FindByBarcodes(ctx, shopID, barcodes)
	if err != nil {
		return nil, err
	}

	tempResults := map[string]models.ProductBarcodeInfo{}
	for _, result := range results {
		tempResults[result.Barcode] = result
	}

	resultDocs := []models.ProductBarcodeInfo{}
	for _, barcode := range barcodes {

		temp, ok := tempResults[barcode]
		if ok {
			resultDocs = append(resultDocs, temp)
		}
	}

	return resultDocs, nil

}

func (svc ProductBarcodeHttpService) SearchProductBarcode(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"barcode",
		"names.name",
		"itemcode",
		"groupcode",
		"groupnames.name",
		"itemunitnames.name",
		"brandcode",
		"brandnames.name",
		"designcode",
		"designnames.name",
		"modelcode",
		"modelnames.name",
		"patterncode",
		"patternnames.name",
		"gradecode",
		"gradenames.name",
		"categorycode",
		"categorynames.name",
		"classcode",
		"classnames.name",
		"groupsubonecode",
		"groupsubonenames.name",
	}

	isalacarte, ok := filters["isalacarte"]
	if ok {
		if !isalacarte.(bool) {
			delete(filters, "isalacarte")
			filters["$or"] = []bson.M{
				{"isalacarte": false},
				{"isalacarte": bson.M{"$exists": false}},
			}
		}

	}

	// Handle shelf search filters
	if shelfSearch, hasShelfSearch := filters["_shelf_search"]; hasShelfSearch {
		// Convert shelf search to map
		shelfFilters, ok := shelfSearch.(map[string]interface{})
		if ok {
			// Find products in the specified shelf by calling warehouse service
			barcodes, err := svc.getProductBarcodesFromShelves(ctx, shopID, shelfFilters)
			if err != nil {
				return []models.ProductBarcodeInfo{}, mongopagination.PaginationData{}, err
			}

			if len(barcodes) == 0 {
				// No products found in the specified shelves, return empty result
				return []models.ProductBarcodeInfo{}, mongopagination.PaginationData{}, nil
			}

			// Add barcode filter to search only products found in shelves
			filters["barcode"] = bson.M{"$in": barcodes}
		}

		// Remove the special shelf search filter as it's processed
		delete(filters, "_shelf_search")
	}

	// Check if shopid filter exists (from shopsid parameter)
	var docList []models.ProductBarcodeInfo
	var pagination mongopagination.PaginationData
	var err error

	if _, hasShopidFilter := filters["shopid"]; hasShopidFilter {
		// Use FindPageFilterNoShopid when shopid filter exists
		docList, pagination, err = svc.repo.FindPageFilterNoShopid(ctx, filters, searchInFields, pageable)
	} else {
		// Use regular FindPageFilter when no shopid filter
		docList, pagination, err = svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)
	}

	if err != nil {
		return []models.ProductBarcodeInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc ProductBarcodeHttpService) SearchProductBarcode2(shopID string, pageable micromodels.Pageable) ([]models.ProductBarcodeSearch, common.Pagination, error) {

	//fixed shopID
	shopID = "2Eh6e3pfWvXTp0yV3CyFEhKPjdI"
	docList, pagination, err := svc.chRepo.Search(shopID, pageable)

	if err != nil {
		return []models.ProductBarcodeSearch{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc ProductBarcodeHttpService) SearchProductBarcodeStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeInfo, int, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"barcode",
		"names.name",
		"itemcode",
		"groupcode",
		"groupnames.name",
		"itemunitnames.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)
	if err != nil {
		return []models.ProductBarcodeInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc ProductBarcodeHttpService) SearchProductBarcodeStepMultiShops(langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeInfo, int, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"barcode",
		"names.name",
		"itemcode",
		"groupcode",
		"groupnames.name",
		"itemunitnames.name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStepNoShopid(ctx, filters, searchInFields, selectFields, pageableStep)
	if err != nil {
		return []models.ProductBarcodeInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc ProductBarcodeHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.ProductBarcode) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.ProductBarcode](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.Barcode)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, shopID, "barcode", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.Barcode)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.ProductBarcode, models.ProductBarcodeDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.ProductBarcode) models.ProductBarcodeDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.ProductBarcodeDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.ProductBarcode = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.ProductBarcode, models.ProductBarcodeDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.ProductBarcodeDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "barcode", guid)
		},
		func(doc models.ProductBarcodeDoc) bool {
			return doc.Barcode != ""
		},
		func(shopID string, authUsername string, dataReq models.ProductBarcode, doc models.ProductBarcodeDoc) error {

			docReq := models.ProductBarcodeRequest{}
			docReq.ProductBarcodeBase = dataReq.ProductBarcodeBase

			tempBarcodes := []models.BarcodeRequest{}

			if dataReq.RefBarcodes != nil {
				for _, docBarcode := range *dataReq.RefBarcodes {
					tempBarcodes = append(tempBarcodes, models.BarcodeRequest{
						Barcode:     docBarcode.Barcode,
						Condition:   docBarcode.Condition,
						StandValue:  docBarcode.StandValue,
						DivideValue: docBarcode.DivideValue,
					})
				}
			}

			docReq.RefBarcodes = tempBarcodes

			if dataReq.BusinessTypes != nil {
				docReq.BusinessTypes = *dataReq.BusinessTypes
			}

			if dataReq.IgnoreBranches != nil {
				docReq.IgnoreBranches = *dataReq.IgnoreBranches
			}

			svc.UpdateProductBarcode(shopID, doc.GuidFixed, authUsername, docReq)

			// err = svc.repo.Update(ctx, shopID, doc.GuidFixed, doc)
			// if err != nil {
			// 	return nil
			// }
			return nil
		},
	)

	if len(createDataList) > 0 {
		svc.repo.Transaction(ctx, func(ctx context.Context) error {
			for _, doc := range createDataList {
				docReq := models.ProductBarcodeRequest{}
				docReq.ProductBarcodeBase = doc.ProductBarcodeBase

				tempBarcodes := []models.BarcodeRequest{}

				if doc.RefBarcodes != nil {
					for _, docBarcode := range *doc.RefBarcodes {
						tempBarcodes = append(tempBarcodes, models.BarcodeRequest{
							Barcode:     docBarcode.Barcode,
							Condition:   docBarcode.Condition,
							StandValue:  docBarcode.StandValue,
							DivideValue: docBarcode.DivideValue,
						})
					}
				}

				if doc.IgnoreBranches != nil {
					docReq.IgnoreBranches = *doc.IgnoreBranches
				}

				if doc.BusinessTypes != nil {
					docReq.BusinessTypes = *doc.BusinessTypes
				}

				docReq.RefBarcodes = tempBarcodes
				_, err = svc.CreateProductBarcode(shopID, authUsername, docReq)

				if err != nil {
					return err
				}

			}

			return nil
		})

		if err != nil {
			return common.BulkImport{}, err
		}

	}

	createDataKey := []string{}

	for _, doc := range createDataList {
		createDataKey = append(createDataKey, doc.Barcode)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.Barcode)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {

		updateDataKey = append(updateDataKey, doc.Barcode)
	}

	updateFailDataKey := []string{}
	for _, doc := range updateFailDataList {
		updateFailDataKey = append(updateFailDataKey, svc.getDocIDKey(doc))
	}

	if len(createDataList) > 0 {
		err = svc.mqRepo.CreateInBatch(createDataList)
		if err != nil {
			return common.BulkImport{}, err
		}
	}

	if len(updateSuccessDataList) > 0 {
		err = svc.mqRepo.UpdateInBatch(updateSuccessDataList)
		if err != nil {
			return common.BulkImport{}, err
		}
	}

	svc.saveMasterSync(shopID)

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc ProductBarcodeHttpService) getDocIDKey(doc models.ProductBarcode) string {
	return doc.Barcode
}

func (svc ProductBarcodeHttpService) XSortsSave(shopID string, authUsername string, xsorts []common.XSortModifyReqesut) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	for _, xsort := range xsorts {
		if len(xsort.GUIDFixed) < 1 {
			continue
		}
		findDoc, err := svc.repo.FindByGuid(ctx, shopID, xsort.GUIDFixed)

		if err != nil {
			return err
		}

		if len(findDoc.GuidFixed) < 1 {
			continue
		}

		if findDoc.XSorts == nil {
			findDoc.XSorts = &[]common.XSort{}
		}

		dictXSorts := map[string]common.XSort{}

		for _, tempXSort := range *findDoc.XSorts {
			dictXSorts[tempXSort.Code] = tempXSort
		}

		dictXSorts[xsort.Code] = common.XSort{
			Code:   xsort.Code,
			XOrder: xsort.XOrder,
		}

		tempXSorts := []common.XSort{}

		for _, tempXSort := range dictXSorts {
			tempXSorts = append(tempXSorts, tempXSort)
		}

		findDoc.XSorts = &tempXSorts
		findDoc.UpdatedBy = authUsername
		findDoc.UpdatedAt = time.Now()

		err = svc.repo.Update(ctx, shopID, findDoc.GuidFixed, findDoc)

		if err != nil {
			return err
		}

	}

	svc.saveMasterSync(shopID)

	return nil

}

func (svc ProductBarcodeHttpService) DeleteProductBarcodeByGUIDs(shopID string, authUsername string, GUIDs []string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	countRefBarcode, err := svc.repo.CountByRefGuids(ctx, shopID, GUIDs)

	if err != nil {
		return err
	}

	if countRefBarcode > 0 {
		return errors.New("document has other barcode referenced")
	}

	countBOM, err := svc.repo.CountByBOMGuids(ctx, shopID, GUIDs)

	if err != nil {
		return err
	}

	if countBOM > 0 {
		return errors.New("document has other bom referenced")
	}

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err = svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc ProductBarcodeHttpService) GetProductBarcodeByUnits(shopID string, unitCodes []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	if len(unitCodes) < 1 {
		return []models.ProductBarcodeInfo{}, mongopagination.PaginationData{}, nil
	}

	results, pagination, err := svc.repo.FindPageByUnits(ctx, shopID, unitCodes, pageable)

	if err != nil {
		return []models.ProductBarcodeInfo{}, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (svc ProductBarcodeHttpService) GetProductBarcodeByGroups(shopID string, groupCodes []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	if len(groupCodes) < 1 {
		return []models.ProductBarcodeInfo{}, mongopagination.PaginationData{}, nil
	}

	results, pagination, err := svc.repo.FindPageByGroups(ctx, shopID, groupCodes, pageable)

	if err != nil {
		return []models.ProductBarcodeInfo{}, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (svc ProductBarcodeHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc ProductBarcodeHttpService) GetModuleName() string {
	return "productbarcode"
}

func (svc ProductBarcodeHttpService) Export(shopID string, languageCode string, languageHeader map[string]string) ([][]string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	filters := bson.M{}

	docs, err := svc.repo.Find(ctx, shopID, filters)

	if err != nil {
		return [][]string{}, err
	}

	keyCols := []string{
		"barcode",        //บาร์โค้ด",
		"productname",    //"ชื่อสินค้า",
		"unitcode",       //"หน่วยนับ",
		"unitname",       //"ชื่อหน่วยนับ",
		"price",          //ราคาขาย",
		"price member",   //ราคาขาย",
		"price delivery", //ราคาขาย",
		"itemtype",       //ประเภทสินค้า",
		"groupcode",      //กลุ่มสินค้า",
	}

	headerRow := []string{}
	for _, keyCol := range keyCols {
		tempVal := keyCol
		if val, ok := languageHeader[keyCol]; ok && val != "" {
			tempVal = val
		}
		headerRow = append(headerRow, tempVal)
	}

	results := [][]string{}

	results = append(results, headerRow)

	temp := prepareDataToCSV(languageCode, docs)
	results = append(results, temp...)

	return results, nil
}

func dataToCSV(data []models.ProductBarcodeDoc, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write your data to the CSV file
	for _, value := range data {
		langCode := "th"

		productName := getName(value.Names, langCode)
		unitName := getName(value.ItemUnitNames, langCode)
		groupName := getName(value.GroupNames, langCode)

		itemType := strconv.Itoa(int(value.ItemType))
		writer.Write([]string{value.Barcode, productName, value.ItemUnitCode, unitName, itemType, value.GroupCode, groupName}) // Adjust fields as per your struct
	}
	return nil
}
func prepareDataToCSV(languageCode string, data []models.ProductBarcodeDoc) [][]string {

	results := [][]string{}

	for _, value := range data {
		langCode := languageCode

		productName := getName(value.Names, langCode)
		unitName := getName(value.ItemUnitNames, langCode)
		groupName := getName(value.GroupNames, langCode)

		price := "0"
		priceMember := "0"
		priceDelivery := "0"

		if value.Prices != nil && len(*value.Prices) > 0 {
			for _, priceItem := range *value.Prices {
				if priceItem.KeyNumber == 1 {
					price = fmt.Sprintf("%.2f", priceItem.Price)
					continue
				}

				if priceItem.KeyNumber == 2 {
					priceMember = fmt.Sprintf("%.2f", priceItem.Price)
					continue
				}

				if priceItem.KeyNumber == 3 {
					priceDelivery = fmt.Sprintf("%.2f", priceItem.Price)
					continue
				}
			}
		}

		itemType := strconv.Itoa(int(value.ItemType))
		results = append(results, []string{value.Barcode, productName, value.ItemUnitCode, unitName, price, priceMember, priceDelivery, itemType, value.GroupCode, groupName})
	}

	return results
}

func getName(names *[]common.NameX, langCode string) string {
	if names == nil {
		return ""
	}

	for _, name := range *names {
		if *name.Code == langCode {
			return *name.Name
		}
	}

	return ""
}

func (svc ProductBarcodeHttpService) Import(shopID string, authUsername string, dataList []models.ProductBarcode) (common.BulkImport, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.ProductBarcode](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.Barcode)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, shopID, "barcode", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.Barcode)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.ProductBarcode, models.ProductBarcodeDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.ProductBarcode) models.ProductBarcodeDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.ProductBarcodeDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.ProductBarcode = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	productBarcodeList, err := svc.repo.FindByDocIndentityGuids(ctx, shopID, "barcode", foundItemGuidList)

	// do update
	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.ProductBarcode, models.ProductBarcodeDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.ProductBarcodeDoc, error) {
			//return svc.repo.FindByDocIndentityGuid(ctx, shopID, "barcode", guid)
			for _, doc := range productBarcodeList {
				if doc.Barcode == guid {
					return doc, nil
				}
			}
			return models.ProductBarcodeDoc{}, nil
		},
		func(doc models.ProductBarcodeDoc) bool {
			return doc.Barcode != ""
		},
		func(shopID string, authUsername string, dataReq models.ProductBarcode, doc models.ProductBarcodeDoc) error {

			docReq := models.ProductBarcodeRequest{}
			docReq.ProductBarcodeBase = dataReq.ProductBarcodeBase

			tempBarcodes := []models.BarcodeRequest{}

			if dataReq.RefBarcodes != nil {
				for _, docBarcode := range *dataReq.RefBarcodes {
					tempBarcodes = append(tempBarcodes, models.BarcodeRequest{
						Barcode:     docBarcode.Barcode,
						Condition:   docBarcode.Condition,
						StandValue:  docBarcode.StandValue,
						DivideValue: docBarcode.DivideValue,
					})
				}
			}

			docReq.RefBarcodes = tempBarcodes

			if dataReq.BusinessTypes != nil {
				docReq.BusinessTypes = *dataReq.BusinessTypes
			}

			if dataReq.IgnoreBranches != nil {
				docReq.IgnoreBranches = *dataReq.IgnoreBranches
			}

			// svc.UpdateProductBarcode(shopID, doc.GuidFixed, authUsername, docReq)

			// findDoc, err := svc.repo.FindByGuid(ctx, shopID, doc.GuidFixed)
			// if findDoc.ID == primitive.NilObjectID {
			// 	return errors.New("document not found")
			// }

			// findDoc := models.ProductBarcodeDoc{}
			// for _, productdBarcodeDoc := range productBarcodeList {
			// 	if productdBarcodeDoc.GuidFixed == doc.GuidFixed {
			// 		findDoc = productdBarcodeDoc
			// 	}
			// }

			// if err != nil {
			// 	return err
			// }
			docData := doc
			docData.ProductBarcode = docReq.ToProductBarcode()
			docData.Barcode = doc.Barcode
			docData.IgnoreBranches = &docReq.IgnoreBranches
			docData.BusinessTypes = &docReq.BusinessTypes

			docData.UpdatedBy = authUsername
			docData.UpdatedAt = time.Now()

			err = svc.repo.Update(ctx, shopID, doc.GuidFixed, doc)
			if err != nil {
				return nil
			}
			return nil
		},
	)

	// do insert
	if len(createDataList) > 0 {
		svc.repo.Transaction(ctx, func(ctx context.Context) error {
			for _, doc := range createDataList {
				docReq := models.ProductBarcodeRequest{}
				docReq.ProductBarcodeBase = doc.ProductBarcodeBase

				tempBarcodes := []models.BarcodeRequest{}

				if doc.RefBarcodes != nil {
					for _, docBarcode := range *doc.RefBarcodes {
						tempBarcodes = append(tempBarcodes, models.BarcodeRequest{
							Barcode:     docBarcode.Barcode,
							Condition:   docBarcode.Condition,
							StandValue:  docBarcode.StandValue,
							DivideValue: docBarcode.DivideValue,
						})
					}
				}

				if doc.IgnoreBranches != nil {
					docReq.IgnoreBranches = *doc.IgnoreBranches
				}

				if doc.BusinessTypes != nil {
					docReq.BusinessTypes = *doc.BusinessTypes
				}

				docReq.RefBarcodes = tempBarcodes
				//_, err = svc.CreateProductBarcode(shopID, authUsername, docReq)

				newGuidFixed := utils.NewGUID()
				docData := models.ProductBarcodeDoc{}
				docData.ShopID = shopID
				docData.GuidFixed = newGuidFixed
				docData.ProductBarcode = docReq.ToProductBarcode()
				docData.IgnoreBranches = &docReq.IgnoreBranches
				docData.BusinessTypes = &docReq.BusinessTypes
				docData.CreatedBy = authUsername
				docData.CreatedAt = time.Now()

				_, err = svc.repo.Create(ctx, docData)
				if err != nil {
					return err
				}

			}

			return nil
		})
	}

	// ให้ส่งไปสร้าง ส่วนอื่น ๆ โดยใช้ kafka
	// if len(createDataList) > 0 {
	// 	err = svc.mqRepo.CreateInBatch(createDataList)
	// 	if err != nil {
	// 		return common.BulkImport{}, err
	// 	}
	// }

	// if len(updateSuccessDataList) > 0 {
	// 	err = svc.mqRepo.UpdateInBatch(updateSuccessDataList)
	// 	if err != nil {
	// 		return common.BulkImport{}, err
	// 	}
	// }

	createDataKey := []string{}

	for _, doc := range createDataList {
		createDataKey = append(createDataKey, doc.Barcode)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.Barcode)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {

		updateDataKey = append(updateDataKey, doc.Barcode)
	}

	updateFailDataKey := []string{}
	for _, doc := range updateFailDataList {
		updateFailDataKey = append(updateFailDataKey, svc.getDocIDKey(doc))
	}

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

// Price History methods
func (svc ProductBarcodeHttpService) GetPriceHistory(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error) {
	return svc.priceHistorySvc.GetPriceHistory(shopID, filters, pageable)
}

func (svc ProductBarcodeHttpService) GetPriceHistoryByBarcode(shopID string, barcode string, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error) {
	return svc.priceHistorySvc.GetPriceHistoryByBarcode(shopID, barcode, pageable)
}

// getProductBarcodesFromShelves queries warehouse system to find product barcodes in specified shelves
func (svc ProductBarcodeHttpService) getProductBarcodesFromShelves(ctx context.Context, shopID string, shelfFilters map[string]interface{}) ([]string, error) {
	// Build filters to query warehouse collection
	filters := make(map[string]interface{})

	// Add warehouse code filter if provided
	if warehouseCode, ok := shelfFilters["warehouse.code"].(string); ok && warehouseCode != "" {
		filters["code"] = warehouseCode
	}

	// Build location and shelf filters based on nested structure
	locationMatch := bson.M{}
	shelfMatch := bson.M{}

	if locationCode, ok := shelfFilters["location.code"].(string); ok && locationCode != "" {
		locationMatch["locations.code"] = locationCode
	}

	if shelfCode, ok := shelfFilters["shelf.code"].(string); ok && shelfCode != "" {
		shelfMatch["locations.shelves.code"] = shelfCode
	}

	// Merge location and shelf filters
	for k, v := range locationMatch {
		filters[k] = v
	}
	for k, v := range shelfMatch {
		filters[k] = v
	}

	// Search for matching warehouses
	searchInFields := []string{"code", "names.name", "location.code", "location.names.name", "location.shelf.code"}
	pageable := micromodels.Pageable{
		Page:  0,
		Limit: 1000, // Limit to reasonable number
	}

	warehouseList, _, err := svc.warehouseRepo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)
	if err != nil {
		return []string{}, err
	}

	// Extract all barcodes from matching shelves
	barcodeSet := make(map[string]bool)
	for _, warehouse := range warehouseList {
		if warehouse.Location == nil {
			continue
		}
		for _, location := range *warehouse.Location {
			// Check if location matches filter (if specified)
			if locationCode, ok := shelfFilters["location.code"].(string); ok && locationCode != "" {
				if location.Code != locationCode {
					continue
				}
			}

			for _, shelf := range location.Shelf {
				// Check if shelf matches filter (if specified)
				if shelfCode, ok := shelfFilters["shelf.code"].(string); ok && shelfCode != "" {
					if shelf.Code != shelfCode {
						continue
					}
				}

				// Extract barcodes from shelf products
				for _, productItem := range shelf.ProductItems {
					if productItem.Barcode != "" {
						barcodeSet[productItem.Barcode] = true
					}
				}
			}
		}
	}

	// Convert set to slice
	barcodes := make([]string, 0, len(barcodeSet))
	for barcode := range barcodeSet {
		barcodes = append(barcodes, barcode)
	}

	return barcodes, nil
}

func (s ProductBarcodeHttpService) ImportRefBarcodeUpdate(shopID, authUsername string, requests []models.RefBarcodeImportRequest) (*models.RefBarcodeImportResponse, error) {
	response := &models.RefBarcodeImportResponse{
		Success:    true,
		TotalItems: len(requests),
		Updated:    0,
		Failed:     0,
		Errors:     []models.RefBarcodeImportError{},
	}

	// Process in batches to avoid memory issues
	batchSize := 100
	for i := 0; i < len(requests); i += batchSize {
		end := i + batchSize
		if end > len(requests) {
			end = len(requests)
		}

		batch := requests[i:end]
		err := s.processBatchRefBarcodeUpdate(shopID, authUsername, batch, i, response)
		if err != nil {
			fmt.Printf("Error processing batch %d-%d: %v\n", i, end, err)
		}
	}

	if response.Failed > 0 {
		response.Success = false
	}

	return response, nil
}

func (s ProductBarcodeHttpService) processBatchRefBarcodeUpdate(shopID, authUsername string, batch []models.RefBarcodeImportRequest, startIndex int, response *models.RefBarcodeImportResponse) error {
	// Extract all barcodes for bulk lookup
	var allBarcodes []string
	barcodeMap := make(map[string]models.RefBarcodeImportRequest)

	for _, req := range batch {
		allBarcodes = append(allBarcodes, req.Barcode, req.BarcodeRef)
		barcodeMap[req.Barcode] = req
	}

	// Bulk find products by barcodes
	products, err := s.repo.FindByBarcodesMap(shopID, allBarcodes)
	if err != nil {
		return err
	}

	// Process each request
	for idx, req := range batch {
		rowNum := startIndex + idx + 1

		// Find target product (the one to be updated)
		targetProduct, exists := products[req.Barcode]
		if !exists {
			response.Failed++
			response.Errors = append(response.Errors, models.RefBarcodeImportError{
				Row:     rowNum,
				Barcode: req.Barcode,
				Error:   "Target barcode not found",
			})
			continue
		}

		// Find reference product
		refProduct, exists := products[req.BarcodeRef]
		if !exists {
			response.Failed++
			response.Errors = append(response.Errors, models.RefBarcodeImportError{
				Row:     rowNum,
				Barcode: req.Barcode,
				Error:   "Reference barcode not found",
			})
			continue
		}

		// Determine condition based on StandValue and DivideValue
		// condition = false when: StandValue = 1 and DivideValue = 1, OR StandValue >= DivideValue
		// condition = true when: StandValue < DivideValue
		condition := false
		if (req.StandValue == 1 && req.DivideValue == 1) || (req.StandValue >= req.DivideValue) {
			condition = false
		} else if req.StandValue < req.DivideValue {
			condition = true
		}

		// Create RefProductBarcode structure
		refBarcode := models.RefProductBarcode{
			GuidFixed:     refProduct.GuidFixed,
			Names:         refProduct.Names,
			ItemUnitCode:  refProduct.ItemUnitCode,
			ItemUnitNames: refProduct.ItemUnitNames,
			Barcode:       req.BarcodeRef,
			Condition:     condition,
			DivideValue:   req.DivideValue,
			StandValue:    req.StandValue,
			Qty:           1,
		}

		// Update target product
		updateData := bson.M{
			"$set": bson.M{
				"refbarcodes":      []models.RefProductBarcode{refBarcode},
				"ismainbarcode":    false,
				"updatedby":        authUsername,
				"updatedat":        time.Now(),
				"isusesubbarcodes": true,
			},
		}

		err := s.repo.UpdateByID(targetProduct.GuidFixed, updateData)
		if err != nil {
			response.Failed++
			response.Errors = append(response.Errors, models.RefBarcodeImportError{
				Row:     rowNum,
				Barcode: req.Barcode,
				Error:   err.Error(),
			})
			continue
		}

		response.Updated++

		// Log the update
		fmt.Printf("Updated refbarcodes for %s -> %s (standvalue: %f, dividevalue: %f, condition: %t, ismainbarcode: false)\n",
			req.Barcode, req.BarcodeRef, req.StandValue, req.DivideValue, condition)
	}

	return nil
}
