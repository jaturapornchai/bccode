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
	productmodels "smlcloudplatform/internal/product/product/models"
	productmaster "smlcloudplatform/internal/product/product/repositories"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/product/productbarcode/repositories"
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
	CreateProductBarcode(holdingCode string, authUsername string, doc models.ProductBarcodeRequest) (string, error)
	CreateProductBarcodeInCompany(holdingCode string, businessCode string, authUsername string, doc models.ProductBarcodeRequest) (string, error)
	UpdateProductBarcode(holdingCode string, guid string, authUsername string, doc models.ProductBarcodeRequest) error
	UpdateProductBarcodeInCompany(holdingCode string, businessCode string, guid string, authUsername string, doc models.ProductBarcodeRequest) error
	DeleteProductBarcode(holdingCode string, guid string, authUsername string) error
	DeleteProductBarcodeInCompany(holdingCode string, businessCode string, guid string, authUsername string) error
	DeleteProductBarcodeByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoProductBarcode(holdingCode string, guid string) (models.ProductBarcodeInfo, error)
	InfoProductBarcodeInCompany(holdingCode string, businessCode string, guid string) (models.ProductBarcodeInfo, error)
	InfoProductBarcodeByBarcode(holdingCode string, itemCode string, barcode string) (models.ProductBarcodeInfo, error)
	InfoProductBarcodeByBarcodeInCompany(holdingCode string, businessCode string, itemCode string, barcode string) (models.ProductBarcodeInfo, error)
	InfoWTFArray(holdingCode string, codes []string) ([]interface{}, error)
	InfoWTFArrayMaster(codes []string) ([]interface{}, error)
	GetProductBarcodeByBarcodeRef(holdingCode string, barcodeRef string) ([]models.ProductBarcodeInfo, error)
	GetProductBarcodeByBarcodeRefMultiShops(holdingCodes []string, barcodeRef string) ([]models.ProductBarcodeInfo, error)
	SearchProductBarcode(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)
	SearchProductBarcodeInCompany(holdingCode string, businessCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)
	SearchProductBarcode2(holdingCode string, pageable micromodels.Pageable) ([]models.ProductBarcodeSearch, common.Pagination, error)

	SearchProductBarcodeStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeInfo, int, error)
	SearchProductBarcodeStepMultiShops(langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.ProductBarcode) (common.BulkImport, error)

	XSortsSave(holdingCode string, authUsername string, xsorts []common.XSortModifyReqesut) error
	GetProductBarcodeByBarcodes(holdingCode string, barcodes []string) ([]models.ProductBarcodeInfo, error)

	// Price History methods
	GetPriceHistory(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error)
	GetPriceHistoryByBarcode(holdingCode string, itemCode string, barcode string, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error)

	UpdateProductBarcodeBranch(holdingCode string, authUsername string, branch models.ProductBarcodeBranch, productBarcodeGUIDFixedes []string) error
	UpdateProductBarcodeBusinessType(holdingCode string, authUsername string, businessType models.ProductBarcodeBusinessType, productBarcodeGUIDFixedes []string) error

	GetModuleName() string
	GetProductBarcodeByUnits(holdingCode string, unitCodes []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)
	GetProductBarcodeByGroups(holdingCode string, groupCodes []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error)
	Export(holdingCode string, languageCode string, languageHeader map[string]string) ([][]string, error)

	InfoBomView(holdingCode string, itemCode string, barcode string) (models.ProductBarcodeBOMView, error)
	Import(holdingCode string, authUsername string, dataList []models.ProductBarcode) (common.BulkImport, error)
	ImportRefBarcodeUpdate(holdingCode, authUsername string, requests []models.RefBarcodeImportRequest) (*models.RefBarcodeImportResponse, error)
}

type ProductBarcodeHttpService struct {
	repo            repositories.IProductBarcodeRepository
	repoMaster      productmaster.IProductRepository
	productMQRepo   productmaster.IProductMessageQueueRepository
	repoUnit        unitmaster.IUnitRepository
	unitSvc         unitservices.IUnitHttpService
	repomgCreditror creditorRepo.CreditorRepository
	chRepo          repositories.IProductBarcodeClickhouseRepository
	syncCacheRepo   mastersync.IMasterSyncCacheRepository
	mqRepo          repositories.IProductBarcodeMessageQueueRepository
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
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	priceHistorySvc IProductPriceHistoryService,
	warehouseRepo warehouseRepo.IWarehouseRepository,
	productMQRepos ...productmaster.IProductMessageQueueRepository,
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
		priceHistorySvc: priceHistorySvc,
		warehouseRepo:   warehouseRepo,
		contextTimeout:  contextTimeout,
	}
	insSvc.ActivityService = services.NewActivityService[models.ProductBarcodeActivity, models.ProductBarcodeDeleteActivity](repo)
	if len(productMQRepos) > 0 {
		insSvc.productMQRepo = productMQRepos[0]
	}
	return insSvc
}

func (svc ProductBarcodeHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func normalizeCompanyBarcodeRequest(docReq models.ProductBarcodeRequest) (models.ProductBarcodeRequest, error) {
	docReq = docReq.CoreOnly()
	docReq.ItemCode = utils.NormalizeBusinessCode(docReq.ItemCode)
	docReq.Barcode = utils.NormalizeBusinessCode(docReq.Barcode)
	docReq.ItemUnitCode = utils.NormalizeBusinessCode(docReq.ItemUnitCode)
	if docReq.ItemCode == "" || docReq.Barcode == "" || docReq.ItemUnitCode == "" {
		return models.ProductBarcodeRequest{}, errors.New("ItemCode, Barcode and ItemUnitCode are required")
	}
	if err := validateEAN13Barcode(docReq.Barcode); err != nil {
		return models.ProductBarcodeRequest{}, err
	}
	return docReq, nil
}

func validateEAN13Barcode(barcode string) error {
	if len(barcode) != 13 {
		return nil
	}
	for index := range barcode {
		if barcode[index] < '0' || barcode[index] > '9' {
			return nil
		}
	}

	sum := 0
	for index := 0; index < 12; index++ {
		digit := int(barcode[index] - '0')
		if index%2 == 1 {
			digit *= 3
		}
		sum += digit
	}
	expected := byte('0' + (10-(sum%10))%10)
	if barcode[12] != expected {
		return errors.New("invalid EAN-13 check digit")
	}
	return nil
}

func validateImmutableBarcodeIdentity(docReq models.ProductBarcodeRequest, stored models.ProductBarcodeDoc) error {
	if docReq.ItemCode != stored.ItemCode || docReq.Barcode != stored.Barcode {
		return errors.New("ItemCode and Barcode cannot be changed")
	}
	return nil
}

func minimumProductForBarcode(holdingCode, businessCode, authUsername string, barcode models.ProductBarcodeDoc, now time.Time) productmodels.ProductDoc {
	names := barcode.Names
	if names == nil || len(*names) == 0 {
		names = &[]common.NameX{*common.NewNameXWithCodeName("th", barcode.ItemCode)}
	}

	doc := productmodels.ProductDoc{}
	doc.HoldingCode = holdingCode
	doc.BusinessCode = businessCode
	doc.GuidFixed = utils.NewGUID()
	doc.Code = barcode.ItemCode
	doc.Names = names
	doc.UnitGuid = barcode.ItemUnitGuid
	doc.UnitCode = barcode.ItemUnitCode
	doc.UnitNames = barcode.ItemUnitNames
	doc.DivideValue = 1
	doc.StandValue = 1
	doc.UnitConversions = []productmodels.ProductUnitConversion{}
	doc.CreatedBy = authUsername
	doc.CreatedAt = now
	doc.UpdatedAt = now
	return doc
}

func applyProductUnitToBarcode(barcode *models.ProductBarcodeDoc, product productmodels.ProductDoc) error {
	unit, ok := product.UnitDefinition(barcode.ItemUnitCode)
	if !ok {
		return fmt.Errorf("หน่วยนับ %s ยังไม่ได้กำหนดในสินค้า %s", barcode.ItemUnitCode, product.Code)
	}
	barcode.ItemUnitCode = unit.UnitCode
	barcode.ItemUnitNames = unit.UnitNames
	barcode.ItemUnitGuid = ""
	if unit.UnitCode == product.UnitCode {
		barcode.ItemUnitGuid = product.UnitGuid
	}
	barcode.Condition = false
	barcode.DivideValue = unit.DivideValue
	barcode.StandValue = unit.StandValue
	return nil
}

func (svc ProductBarcodeHttpService) CreateProductBarcode(holdingCode string, authUsername string, docReq models.ProductBarcodeRequest) (string, error) {
	return svc.createProductBarcode(holdingCode, "", authUsername, docReq)
}

func (svc ProductBarcodeHttpService) CreateProductBarcodeInCompany(holdingCode string, businessCode string, authUsername string, docReq models.ProductBarcodeRequest) (string, error) {
	businessCode = utils.NormalizeBusinessCode(businessCode)
	if businessCode == "" {
		return "", errors.New("BusinessCode is required")
	}
	return svc.createProductBarcode(holdingCode, businessCode, authUsername, docReq)
}

func (svc ProductBarcodeHttpService) createProductBarcode(holdingCode string, businessCode string, authUsername string, docReq models.ProductBarcodeRequest) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	var err error
	if businessCode != "" {
		docReq, err = normalizeCompanyBarcodeRequest(docReq)
		if err != nil {
			return "", err
		}
		if svc.productMQRepo == nil {
			return "", errors.New("Product message queue repository is required")
		}
	} else {
		docReq.ItemCode = utils.NormalizeBusinessCode(docReq.ItemCode)
		docReq.Barcode = utils.NormalizeBusinessCode(docReq.Barcode)
		if docReq.Barcode == "" {
			return "", errors.New("Barcode is required")
		}
	}
	if err := svc.repo.EnsureIndexes(ctx); err != nil {
		return "", err
	}
	if businessCode != "" {
		if err := svc.repoMaster.EnsureIndexes(ctx); err != nil {
			return "", err
		}
	}

	findDoc := models.ProductBarcodeDoc{}
	if businessCode != "" {
		findDoc, err = svc.repo.FindByBarcodeInCompany(ctx, holdingCode, businessCode, docReq.Barcode)
	} else {
		findDoc, err = svc.repo.FindOne(ctx, holdingCode, bson.M{
			"itemcode": docReq.ItemCode,
			"barcode":  docReq.Barcode,
		})
	}

	if err != nil {
		return "", err
	}

	if findDoc.Barcode != "" {
		return "", errors.New("บาร์โค้ดนี้มีอยู่แล้วในบริษัท")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.ProductBarcodeDoc{}
	docData.HoldingCode = holdingCode
	docData.BusinessCode = businessCode
	docData.GuidFixed = newGuidFixed
	docData.ProductBarcode = docReq.ToProductBarcode()
	if err := models.ValidateProductClassification(docData.ItemType, docData.MaterialType); err != nil {
		return "", err
	}
	docData.IgnoreBranches = &docReq.IgnoreBranches
	docData.BusinessTypes = &docReq.BusinessTypes

	// Unit validation and creation logic
	if docReq.ItemUnitCode != "" {
		// Check if unit exists in master units
		unitDoc, err := svc.repoUnit.FindByDocIndentityGuid(ctx, holdingCode, "unitcode", docReq.ItemUnitCode)
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
			unitGuid, err := svc.unitSvc.CreateUnit(holdingCode, authUsername, newUnit)
			if err != nil {
				return "", fmt.Errorf("error creating unit: %v", err)
			}

			// Use the created unit data
			docData.ItemUnitGuid = unitGuid
			docData.ItemUnitNames = docReq.ItemUnitNames
		}
	}

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now().UTC()

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

	docData.RefBarcodes, err = svc.prepareRefBarcode(ctx, holdingCode, docReq.RefBarcodes)

	if err != nil {
		return "", err
	}

	docData.BOMs, err = svc.prepareBOMs(ctx, holdingCode, docReq.BOMs)
	if err != nil {
		return "", err
	}

	if docData.BOMs != nil && len(*docData.BOMs) > 0 {
		var activeBOM *[]models.BOMProductBarcode
		now := time.Now().UTC()
		for _, ver := range *docData.BOMs {
			if (ver.StartDate.Before(now) || ver.StartDate.Equal(now)) && (ver.EndDate == nil || ver.EndDate.After(now)) {
				activeBOM = ver.BOM
				break
			}
		}
		if activeBOM == nil {
			activeBOM = (*docData.BOMs)[len(*docData.BOMs)-1].BOM
		}
		docData.BOM = activeBOM
	} else {
		docData.BOM, err = svc.prepareBOM(ctx, holdingCode, docReq.BOM)
		if err != nil {
			return "", err
		}
	}

	var createdProduct *productmodels.ProductDoc
	if businessCode != "" {
		err = svc.repo.Transaction(ctx, func(transactionContext context.Context) error {
			duplicate, err := svc.repo.FindByBarcodeInCompany(transactionContext, holdingCode, businessCode, docData.Barcode)
			if err != nil {
				return err
			}
			if duplicate.ID != primitive.NilObjectID {
				return errors.New("บาร์โค้ดนี้มีอยู่แล้วในบริษัท")
			}

			product, err := svc.repoMaster.FindByCodeInCompany(transactionContext, holdingCode, businessCode, docData.ItemCode)
			if err != nil {
				return err
			}
			if product.ID == primitive.NilObjectID {
				minimumProduct := minimumProductForBarcode(holdingCode, businessCode, authUsername, docData, docData.CreatedAt)
				if err := applyProductUnitToBarcode(&docData, minimumProduct); err != nil {
					return err
				}
				if _, err := svc.repoMaster.Create(transactionContext, minimumProduct); err != nil {
					return err
				}
				createdProduct = &minimumProduct
			} else if err := applyProductUnitToBarcode(&docData, product); err != nil {
				return err
			}

			_, err = svc.repo.Create(transactionContext, docData)
			return err
		})
	} else {
		_, err = svc.repo.Create(ctx, docData)
	}
	if err != nil {
		return "", err
	}

	if createdProduct != nil {
		if err := svc.productMQRepo.Create(*createdProduct); err != nil {
			return "", err
		}
	}
	if err = svc.mqRepo.Create(docData); err != nil {
		return "", err
	}

	// Legacy price history has no BusinessCode. Skip company-owned writes until
	// that collection is partitioned, otherwise duplicate item/barcode keys mix.
	if shouldRecordLegacyPriceHistory(businessCode) && docData.Prices != nil && len(*docData.Prices) > 0 {
		productName := ""
		if docData.Names != nil && len(*docData.Names) > 0 {
			if (*docData.Names)[0].Name != nil {
				productName = *(*docData.Names)[0].Name
			}
		}

		err = svc.priceHistorySvc.RecordPriceChange(
			ctx,
			holdingCode,
			newGuidFixed,
			docData.ItemCode,
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

	svc.saveMasterSync(holdingCode)

	return newGuidFixed, nil
}

func (svc ProductBarcodeHttpService) UpdateProductBarcode(holdingCode string, guid string, authUsername string, docReq models.ProductBarcodeRequest) error {
	return svc.updateProductBarcode(holdingCode, "", guid, authUsername, docReq)
}

func (svc ProductBarcodeHttpService) UpdateProductBarcodeInCompany(holdingCode string, businessCode string, guid string, authUsername string, docReq models.ProductBarcodeRequest) error {
	businessCode = utils.NormalizeBusinessCode(businessCode)
	if businessCode == "" {
		return errors.New("BusinessCode is required")
	}
	return svc.updateProductBarcode(holdingCode, businessCode, guid, authUsername, docReq)
}

func (svc ProductBarcodeHttpService) updateProductBarcode(holdingCode string, businessCode string, guid string, authUsername string, docReq models.ProductBarcodeRequest) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc := models.ProductBarcodeDoc{}
	var err error
	if businessCode != "" {
		findDoc, err = svc.repo.FindByGuidInCompany(ctx, holdingCode, businessCode, guid)
	} else {
		findDoc, err = svc.repo.FindByGuid(ctx, holdingCode, guid)
	}

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}
	if businessCode != "" {
		docReq, err = normalizeCompanyBarcodeRequest(docReq)
		if err != nil {
			return err
		}
		if err := validateImmutableBarcodeIdentity(docReq, findDoc); err != nil {
			return err
		}
		if svc.productMQRepo == nil {
			return errors.New("Product message queue repository is required")
		}
	} else {
		docReq.ItemCode = utils.NormalizeBusinessCode(docReq.ItemCode)
		docReq.Barcode = findDoc.Barcode
	}
	if err := svc.repo.EnsureIndexes(ctx); err != nil {
		return err
	}
	if businessCode != "" {
		if err := svc.repoMaster.EnsureIndexes(ctx); err != nil {
			return err
		}
	}

	duplicate := models.ProductBarcodeDoc{}
	if businessCode != "" {
		duplicate, err = svc.repo.FindByBarcodeInCompany(ctx, holdingCode, businessCode, docReq.Barcode)
	} else {
		duplicate, err = svc.repo.FindOne(ctx, holdingCode, bson.M{
			"itemcode": docReq.ItemCode,
			"barcode":  findDoc.Barcode,
		})
	}
	if err != nil {
		return err
	}
	if duplicate.ID != primitive.NilObjectID && duplicate.GuidFixed != findDoc.GuidFixed {
		return errors.New("บาร์โค้ดนี้มีอยู่แล้วในบริษัท")
	}

	docData := findDoc

	// เก็บราคาเก่าไว้เพื่อเปรียบเทียบ
	oldPrices := []models.ProductPrice{}
	if findDoc.Prices != nil {
		oldPrices = *findDoc.Prices
	}

	docData.ProductBarcode = docReq.ToProductBarcode()
	if err := models.ValidateProductClassification(docData.ItemType, docData.MaterialType); err != nil {
		return err
	}

	docData.Barcode = findDoc.Barcode
	docData.IgnoreBranches = &docReq.IgnoreBranches
	docData.BusinessTypes = &docReq.BusinessTypes

	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now().UTC()

	docData.RefBarcodes, err = svc.prepareRefBarcode(ctx, holdingCode, docReq.RefBarcodes)

	if err != nil {
		return err
	}

	docData.BOMs, err = svc.prepareBOMs(ctx, holdingCode, docReq.BOMs)
	if err != nil {
		return err
	}

	if docData.BOMs != nil && len(*docData.BOMs) > 0 {
		var activeBOM *[]models.BOMProductBarcode
		now := time.Now().UTC()
		for _, ver := range *docData.BOMs {
			if (ver.StartDate.Before(now) || ver.StartDate.Equal(now)) && (ver.EndDate == nil || ver.EndDate.After(now)) {
				activeBOM = ver.BOM
				break
			}
		}
		if activeBOM == nil {
			activeBOM = (*docData.BOMs)[len(*docData.BOMs)-1].BOM
		}
		docData.BOM = activeBOM
	} else {
		docData.BOM, err = svc.prepareBOM(ctx, holdingCode, docReq.BOM)
		if err != nil {
			return err
		}
	}

	var createdProduct *productmodels.ProductDoc
	if businessCode != "" {
		err = svc.repo.Transaction(ctx, func(transactionContext context.Context) error {
			product, err := svc.repoMaster.FindByCodeInCompany(transactionContext, holdingCode, businessCode, docData.ItemCode)
			if err != nil {
				return err
			}
			if product.ID == primitive.NilObjectID {
				minimumProduct := minimumProductForBarcode(holdingCode, businessCode, authUsername, docData, docData.UpdatedAt)
				if err := applyProductUnitToBarcode(&docData, minimumProduct); err != nil {
					return err
				}
				if _, err := svc.repoMaster.Create(transactionContext, minimumProduct); err != nil {
					return err
				}
				createdProduct = &minimumProduct
			} else if err := applyProductUnitToBarcode(&docData, product); err != nil {
				return err
			}

			if err := svc.updateMetaInRefBarcode(transactionContext, holdingCode, businessCode, findDoc.ItemCode, docData); err != nil {
				return err
			}
			if err := svc.updateMetaInBOMBarcode(transactionContext, holdingCode, businessCode, findDoc.ItemCode, docData); err != nil {
				return err
			}
			return svc.repo.UpdateInCompany(transactionContext, holdingCode, businessCode, guid, docData)
		})
	} else {
		if err = svc.updateMetaInRefBarcode(ctx, holdingCode, "", findDoc.ItemCode, docData); err != nil {
			return err
		}
		if err = svc.updateMetaInBOMBarcode(ctx, holdingCode, "", findDoc.ItemCode, docData); err != nil {
			return err
		}
		err = svc.repo.Update(ctx, holdingCode, guid, docData)
	}
	if err != nil {
		return err
	}
	if createdProduct != nil {
		if err := svc.productMQRepo.Create(*createdProduct); err != nil {
			return err
		}
	}

	if businessCode == "" && findDoc.ItemCode != docData.ItemCode {
		if err := svc.mqRepo.Delete(findDoc); err != nil {
			return err
		}
	}

	err = svc.mqRepo.Update(docData)
	if err != nil {
		return err
	}

	if shouldRecordLegacyPriceHistory(businessCode) {
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
			holdingCode,
			guid,
			docData.ItemCode,
			docData.Barcode,
			productName,
			oldPrices,
			newPrices,
			"update",
			authUsername,
			"แก้ไขข้อมูลสินค้า",
		)
		if err != nil {
			// Price history is best-effort and must not fail the master update.
		}
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func shouldRecordLegacyPriceHistory(businessCode string) bool {
	return strings.TrimSpace(businessCode) == ""
}

func (svc ProductBarcodeHttpService) UpdateProductBarcodeBranch(holdingCode string, authUsername string, branch models.ProductBarcodeBranch, productBarcodeGUIDFixedes []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repo.UpdateBranch(ctx, holdingCode, branch, productBarcodeGUIDFixedes)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc ProductBarcodeHttpService) UpdateProductBarcodeBusinessType(holdingCode string, authUsername string, businessType models.ProductBarcodeBusinessType, productBarcodeGUIDFixedes []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repo.UpdateBusinessType(ctx, holdingCode, businessType, productBarcodeGUIDFixedes)

	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc ProductBarcodeHttpService) prepareRefBarcode(ctx context.Context, holdingCode string, barcodes []models.BarcodeRequest) (*[]models.RefProductBarcode, error) {
	tempBarcodes := []models.RefProductBarcode{}
	for _, item := range barcodes {
		childDoc, err := svc.findBarcodeByBusinessKey(ctx, holdingCode, item.ItemCode, item.Barcode)
		if err != nil {
			return &[]models.RefProductBarcode{}, err
		}
		tempRef := childDoc.ToRefBarcode()
		tempRef.Condition = item.Condition
		tempRef.StandValue = item.StandValue
		tempRef.DivideValue = item.DivideValue
		tempRef.Qty = item.Qty

		tempBarcodes = append(tempBarcodes, tempRef)
	}

	return &tempBarcodes, nil
}

func (svc ProductBarcodeHttpService) prepareBOM(ctx context.Context, holdingCode string, barcodes []models.BOMRequest) (*[]models.BOMProductBarcode, error) {
	tempBarcodes := []models.BOMProductBarcode{}
	for _, item := range barcodes {
		childDoc, err := svc.findBarcodeByBusinessKey(ctx, holdingCode, item.ItemCode, item.Barcode)
		if err != nil {
			return &[]models.BOMProductBarcode{}, err
		}
		temp := childDoc.ToBOM()
		temp.Condition = item.Condition
		temp.StandValue = item.StandValue
		temp.DivideValue = item.DivideValue
		temp.Qty = item.Qty

		tempBarcodes = append(tempBarcodes, temp)
	}

	return &tempBarcodes, nil
}

func (svc ProductBarcodeHttpService) findBarcodeByBusinessKey(
	ctx context.Context,
	holdingCode string,
	itemCode string,
	barcode string,
) (models.ProductBarcodeDoc, error) {
	itemCode = utils.NormalizeBusinessCode(itemCode)
	barcode = utils.NormalizeBusinessCode(barcode)
	if barcode == "" {
		return models.ProductBarcodeDoc{}, errors.New("Barcode is required")
	}

	doc, err := svc.repo.FindByBusinessKey(ctx, holdingCode, itemCode, barcode)
	if err != nil {
		return models.ProductBarcodeDoc{}, err
	}
	if doc.ID == primitive.NilObjectID {
		return models.ProductBarcodeDoc{}, fmt.Errorf("product barcode %s/%s not found", itemCode, barcode)
	}
	return doc, nil
}

func (svc ProductBarcodeHttpService) prepareBOMs(ctx context.Context, holdingCode string, reqBOMs []models.BOMVersionRequest) (*[]models.ProductBarcodeBOMVersion, error) {
	if reqBOMs == nil {
		return nil, nil
	}

	var versions []models.ProductBarcodeBOMVersion
	for _, reqVer := range reqBOMs {
		bomItems, err := svc.prepareBOM(ctx, holdingCode, reqVer.BOM)
		if err != nil {
			return nil, err
		}

		guid := reqVer.GuidFixed
		if guid == "" {
			guid = utils.NewGUID()
		}

		versions = append(versions, models.ProductBarcodeBOMVersion{
			GuidFixed: guid,
			StartDate: reqVer.StartDate,
			EndDate:   reqVer.EndDate,
			BOM:       bomItems,
		})
	}

	return &versions, nil
}

func (svc ProductBarcodeHttpService) updateMetaInRefBarcode(
	ctx context.Context,
	holdingCode string,
	businessCode string,
	lookupItemCode string,
	docData models.ProductBarcodeDoc,
) error {

	var findDocs []models.ProductBarcodeDoc
	var err error
	if businessCode != "" {
		findDocs, err = svc.repo.FindByRefKeyInCompany(ctx, holdingCode, businessCode, lookupItemCode, docData.Barcode)
	} else {
		findDocs, err = svc.repo.FindByRefKey(ctx, holdingCode, lookupItemCode, docData.Barcode)
	}
	if err != nil {
		return err
	}

	for _, findDoc := range findDocs {
		tempRefBarcodes := []models.RefProductBarcode{}
		for _, refBarcode := range *findDoc.RefBarcodes {
			if refBarcode.ItemCode == lookupItemCode && refBarcode.Barcode == docData.Barcode {
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
				refBarcode.ItemCode = docData.ItemCode
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

		if businessCode != "" {
			err = svc.repo.UpdateInCompany(ctx, holdingCode, businessCode, findDoc.GuidFixed, findDoc)
		} else {
			err = svc.repo.Update(ctx, holdingCode, findDoc.GuidFixed, findDoc)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func updateBOMBarcodeMetadata(
	items *[]models.BOMProductBarcode,
	lookupItemCode string,
	docData models.ProductBarcodeDoc,
) bool {
	if items == nil {
		return false
	}
	changed := false
	for index := range *items {
		item := &(*items)[index]
		if item.ItemCode != lookupItemCode || item.Barcode != docData.Barcode {
			continue
		}
		item.Names = docData.Names
		item.ItemCode = docData.ItemCode
		item.ItemUnitCode = docData.ItemUnitCode
		item.ItemUnitNames = docData.ItemUnitNames
		changed = true
	}
	return changed
}

func (svc ProductBarcodeHttpService) updateMetaInBOMBarcode(
	ctx context.Context,
	holdingCode string,
	businessCode string,
	lookupItemCode string,
	docData models.ProductBarcodeDoc,
) error {

	var findDocs []models.ProductBarcodeDoc
	var err error
	if businessCode != "" {
		findDocs, err = svc.repo.FindByBOMKeyInCompany(ctx, holdingCode, businessCode, lookupItemCode, docData.Barcode)
	} else {
		findDocs, err = svc.repo.FindByBOMKey(ctx, holdingCode, lookupItemCode, docData.Barcode)
	}
	if err != nil {
		return err
	}

	for _, findDoc := range findDocs {
		changed := updateBOMBarcodeMetadata(findDoc.BOM, lookupItemCode, docData)
		if findDoc.BOMs != nil {
			for index := range *findDoc.BOMs {
				version := &(*findDoc.BOMs)[index]
				if updateBOMBarcodeMetadata(version.BOM, lookupItemCode, docData) {
					changed = true
				}
			}
		}
		if !changed {
			continue
		}
		if businessCode != "" {
			if err := svc.repo.UpdateInCompany(ctx, holdingCode, businessCode, findDoc.GuidFixed, findDoc); err != nil {
				return err
			}
		} else if err := svc.repo.Update(ctx, holdingCode, findDoc.GuidFixed, findDoc); err != nil {
			return err
		}
	}

	return nil
}

func (svc ProductBarcodeHttpService) DeleteProductBarcode(holdingCode string, guid string, authUsername string) error {
	return svc.deleteProductBarcode(holdingCode, "", guid, authUsername)
}

func (svc ProductBarcodeHttpService) DeleteProductBarcodeInCompany(holdingCode string, businessCode string, guid string, authUsername string) error {
	businessCode = utils.NormalizeBusinessCode(businessCode)
	if businessCode == "" {
		return errors.New("BusinessCode is required")
	}
	return svc.deleteProductBarcode(holdingCode, businessCode, guid, authUsername)
}

func (svc ProductBarcodeHttpService) deleteProductBarcode(holdingCode string, businessCode string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc := models.ProductBarcodeDoc{}
	var err error
	if businessCode != "" {
		findDoc, err = svc.repo.FindByGuidInCompany(ctx, holdingCode, businessCode, guid)
	} else {
		findDoc, err = svc.repo.FindByGuid(ctx, holdingCode, guid)
	}

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	countRef := 0
	if businessCode != "" {
		countRef, err = svc.repo.CountByRefKeyInCompany(ctx, holdingCode, businessCode, findDoc.ItemCode, findDoc.Barcode)
	} else {
		countRef, err = svc.repo.CountByRefKey(ctx, holdingCode, findDoc.ItemCode, findDoc.Barcode)
	}

	if err != nil {
		return err
	}

	if countRef > 0 {
		return errors.New("document has other ref barcode referenced")
	}

	countBOM := 0
	if businessCode != "" {
		countBOM, err = svc.repo.CountByBOMKeyInCompany(ctx, holdingCode, businessCode, findDoc.ItemCode, findDoc.Barcode)
	} else {
		countBOM, err = svc.repo.CountByBOMKey(ctx, holdingCode, findDoc.ItemCode, findDoc.Barcode)
	}

	if err != nil {
		return err
	}

	if countBOM > 0 {
		return errors.New("document has other bom referenced")
	}

	if businessCode != "" {
		err = svc.repo.DeleteByGuidfixedInCompany(ctx, holdingCode, businessCode, guid, authUsername)
	} else {
		err = svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)
	}
	if err != nil {
		return err
	}

	err = svc.mqRepo.Delete(findDoc)
	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc ProductBarcodeHttpService) InfoProductBarcode(holdingCode string, guid string) (models.ProductBarcodeInfo, error) {
	return svc.infoProductBarcode(holdingCode, "", guid)
}

func (svc ProductBarcodeHttpService) InfoProductBarcodeInCompany(holdingCode string, businessCode string, guid string) (models.ProductBarcodeInfo, error) {
	businessCode = utils.NormalizeBusinessCode(businessCode)
	if businessCode == "" {
		return models.ProductBarcodeInfo{}, errors.New("BusinessCode is required")
	}
	return svc.infoProductBarcode(holdingCode, businessCode, guid)
}

func (svc ProductBarcodeHttpService) infoProductBarcode(holdingCode string, businessCode string, guid string) (models.ProductBarcodeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc := models.ProductBarcodeDoc{}
	var err error
	if businessCode != "" {
		findDoc, err = svc.repo.FindByGuidInCompany(ctx, holdingCode, businessCode, guid)
	} else {
		findDoc, err = svc.repo.FindByGuid(ctx, holdingCode, guid)
	}
	if err != nil {
		return models.ProductBarcodeInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.ProductBarcodeInfo{}, errors.New("document not found")
	}

	// Product relations resolve by the portable Product code, never by guidfixed.
	if strings.TrimSpace(findDoc.ItemCode) != "" {
		findMasterDoc := productmodels.ProductDoc{}
		if businessCode != "" {
			findMasterDoc, err = svc.repoMaster.FindByCodeInCompany(ctx, holdingCode, businessCode, utils.NormalizeBusinessCode(findDoc.ItemCode))
		} else {
			findMasterDoc, err = svc.repoMaster.FindOneByCode(ctx, holdingCode, utils.NormalizeBusinessCode(findDoc.ItemCode))
		}
		if err == nil && findMasterDoc.ID != primitive.NilObjectID {
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

			if findMasterDoc.ManufacturerGUID != "" {
				findManu, err := svc.repomgCreditror.FindByGuid(ctx, holdingCode, findMasterDoc.ManufacturerGUID)
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

			// Map new classification fields from Product to ProductBarcode
			findDoc.GroupsuboneGuid = findMasterDoc.GroupsuboneGuid
			findDoc.GroupsuboneCode = findMasterDoc.GroupsuboneCode
			findDoc.GroupsuboneNames = findMasterDoc.GroupsuboneNames

			findDoc.GroupsubtwoGuid = findMasterDoc.GroupsubtwoGuid
			findDoc.GroupsubtwoCode = findMasterDoc.GroupsubtwoCode
			findDoc.GroupsubtwoNames = findMasterDoc.GroupsubtwoNames

			findDoc.BrandGuid = findMasterDoc.BrandGuid
			findDoc.BrandCode = findMasterDoc.BrandCode
			findDoc.BrandNames = findMasterDoc.BrandNames

			findDoc.DesignGuid = findMasterDoc.DesignGuid
			findDoc.DesignCode = findMasterDoc.DesignCode
			findDoc.DesignNames = findMasterDoc.DesignNames

			findDoc.ModelGuid = findMasterDoc.ModelGuid
			findDoc.ModelCode = findMasterDoc.ModelCode
			findDoc.ModelNames = findMasterDoc.ModelNames

			findDoc.PatternGuid = findMasterDoc.PatternGuid
			findDoc.PatternCode = findMasterDoc.PatternCode
			findDoc.PatternNames = findMasterDoc.PatternNames

			findDoc.GradeGuid = findMasterDoc.GradeGuid
			findDoc.GradeCode = findMasterDoc.GradeCode
			findDoc.GradeNames = findMasterDoc.GradeNames

			findDoc.CategoryGuid = findMasterDoc.CategoryGuid
			findDoc.CategoryCode = findMasterDoc.CategoryCode
			findDoc.CategoryNames = findMasterDoc.CategoryNames

			findDoc.ClassGuid = findMasterDoc.ClassGuid
			findDoc.ClassCode = findMasterDoc.ClassCode
			findDoc.ClassNames = findMasterDoc.ClassNames

			findDoc.MaterialType = findMasterDoc.MaterialType
			findDoc.TaxType = findMasterDoc.TaxType
			findDoc.VatType = findMasterDoc.VatType
			findDoc.VatCal = int(findMasterDoc.VatType)

			if findMasterDoc.Manufacturers != nil {
				tempManuf := []models.ProductBarcodeManufacturer{}
				for _, m := range *findMasterDoc.Manufacturers {
					tempManuf = append(tempManuf, models.ProductBarcodeManufacturer{
						DocIdentity: common.DocIdentity{GuidFixed: m.GuidFixed},
						Code:        m.Code,
						Names:       m.Names,
					})
				}
				findDoc.Manufacturers = &tempManuf
			} else {
				findDoc.Manufacturers = &[]models.ProductBarcodeManufacturer{}
			}

			if findMasterDoc.Suppliers != nil {
				tempSuppl := []models.ProductBarcodeSupplier{}
				for _, s := range *findMasterDoc.Suppliers {
					tempSuppl = append(tempSuppl, models.ProductBarcodeSupplier{
						DocIdentity: common.DocIdentity{GuidFixed: s.GuidFixed},
						Code:        s.Code,
						Names:       s.Names,
					})
				}
				findDoc.Suppliers = &tempSuppl
			} else {
				findDoc.Suppliers = &[]models.ProductBarcodeSupplier{}
			}

			// Product controls shared color settings; Barcode keeps its own images.
			findDoc.UseImageOrColor = findMasterDoc.UseImageOrColor
			findDoc.ColorSelect = findMasterDoc.ColorSelect
			findDoc.ColorSelectHex = findMasterDoc.ColorSelectHex

			// Map POS / Restaurant settings
			findDoc.IsSumPoint = findMasterDoc.IsSumPoint
			findDoc.IsALaCarte = findMasterDoc.IsALaCarte
			findDoc.IsSplitUnitPrint = findMasterDoc.IsSplitUnitPrint
			findDoc.IsOnlyStaff = findMasterDoc.IsOnlyStaff
			findDoc.FoodType = findMasterDoc.FoodType
			findDoc.IsStockForRestaurant = findMasterDoc.IsStockForRestaurant

			findDoc.Restaurant = models.ProductRestaurant{
				IsForRestaurant:       findMasterDoc.Restaurant.IsForRestaurant,
				IsForTakeAway:         findMasterDoc.Restaurant.IsForTakeAway,
				IsForDelivery:         findMasterDoc.Restaurant.IsForDelivery,
				IsForCustomer:         findMasterDoc.Restaurant.IsForCustomer,
				IsForCustomerPreOrder: findMasterDoc.Restaurant.IsForCustomerPreOrder,
			}

			if findMasterDoc.OrderTypes != nil {
				tempOT := []models.ProductOrderType{}
				for _, ot := range *findMasterDoc.OrderTypes {
					row := models.ProductOrderType{
						Code:  ot.Code,
						Names: ot.Names,
						Price: ot.Price,
					}
					row.GuidFixed = ot.GuidFixed
					tempOT = append(tempOT, row)
				}
				findDoc.OrderTypes = &tempOT
			}

			if findMasterDoc.Options != nil {
				tempOpts := []models.ProductOption{}
				for _, opt := range *findMasterDoc.Options {
					tempChoices := []models.ProductChoice{}
					if opt.Choices != nil {
						for _, ch := range *opt.Choices {
							tempChoices = append(tempChoices, models.ProductChoice{
								GUID:            ch.GUID,
								Names:           ch.Names,
								ImageURI:        ch.ImageURI,
								RefBarcode:      ch.RefBarcode,
								RefBarcodeNames: ch.RefBarcodeNames,
								RefProductCode:  ch.RefProductCode,
								RefUnitCode:     ch.RefUnitCode,
								IsStock:         ch.IsStock,
								IsDefault:       ch.IsDefault,
								Qty:             ch.Qty,
								Price:           ch.Price,
							})
						}
					}
					tempOpts = append(tempOpts, models.ProductOption{
						GUID:       opt.GUID,
						Names:      opt.Names,
						ChoiceType: int8(opt.ChoiceType),
						MinSelect:  uint16(opt.MinSelect),
						MaxSelect:  uint16(opt.MaxSelect),
						Choices:    &tempChoices,
					})
				}
				findDoc.Options = &tempOpts
			}

			// Map shared alerts; Barcode keeps its own description.
			findDoc.IsAlert = findMasterDoc.IsAlert
			findDoc.AlertDescription = findMasterDoc.AlertDescription

			// Map TimeForSales
			if findMasterDoc.TimeForSales != nil {
				tempTfs := []models.ProductTimeForSale{}
				for _, tfs := range *findMasterDoc.TimeForSales {
					tempTfs = append(tempTfs, models.ProductTimeForSale{
						DaysOfWeek: tfs.DaysOfWeek,
						FromDate:   tfs.FromDate,
						ToDate:     tfs.ToDate,
						FromTime:   tfs.FromTime,
						ToTime:     tfs.ToTime,
					})
				}
				findDoc.TimeForSales = &tempTfs
			}

			// Map BusinessTypes and IgnoreBranches
			if findMasterDoc.BusinessTypes != nil {
				tempBt := []models.ProductBarcodeBusinessType{}
				for _, bt := range *findMasterDoc.BusinessTypes {
					row := models.ProductBarcodeBusinessType{
						Code:     bt.Code,
						Names:    bt.Names,
						IsIgnore: bt.IsIgnore,
					}
					row.GuidFixed = bt.GuidFixed
					tempBt = append(tempBt, row)
				}
				findDoc.BusinessTypes = &tempBt
			}

			if findMasterDoc.IgnoreBranches != nil {
				tempIb := []models.ProductBarcodeBranch{}
				for _, ib := range *findMasterDoc.IgnoreBranches {
					row := models.ProductBarcodeBranch{
						Code:     ib.Code,
						Names:    ib.Names,
						IsIgnore: ib.IsIgnore,
					}
					row.GuidFixed = ib.GuidFixed
					tempIb = append(tempIb, row)
				}
				findDoc.IgnoreBranches = &tempIb
			}

			// BOM remains a Barcode reference; unit conversions belong to Product.
			if findMasterDoc.BOM != nil {
				tempBOM := []models.BOMProductBarcode{}
				for _, b := range *findMasterDoc.BOM {
					tempBOM = append(tempBOM, models.BOMProductBarcode{
						BarcodeGuidFixed: b.BarcodeGuidFixed,
						ItemCode:         b.ItemCode,
						Level:            b.Level,
						Names:            b.Names,
						ItemUnitCode:     b.ItemUnitCode,
						ItemUnitNames:    b.ItemUnitNames,
						Barcode:          b.Barcode,
						Condition:        b.Condition,
						DivideValue:      b.DivideValue,
						StandValue:       b.StandValue,
						Qty:              b.Qty,
					})
				}
				findDoc.BOM = &tempBOM
			} else {
				findDoc.BOM = &[]models.BOMProductBarcode{}
			}
		}
	}

	if strings.TrimSpace(findDoc.ItemUnitGuid) != "" {
		unit, err := svc.repoUnit.FindByGuid(ctx, holdingCode, findDoc.ItemUnitGuid)
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

func (svc ProductBarcodeHttpService) InfoProductBarcodeByBarcode(holdingCode string, itemCode string, barcode string) (models.ProductBarcodeInfo, error) {
	return svc.infoProductBarcodeByBarcode(holdingCode, "", itemCode, barcode)
}

func (svc ProductBarcodeHttpService) InfoProductBarcodeByBarcodeInCompany(holdingCode string, businessCode string, itemCode string, barcode string) (models.ProductBarcodeInfo, error) {
	businessCode = utils.NormalizeBusinessCode(businessCode)
	if businessCode == "" {
		return models.ProductBarcodeInfo{}, errors.New("BusinessCode is required")
	}
	return svc.infoProductBarcodeByBarcode(holdingCode, businessCode, itemCode, barcode)
}

func (svc ProductBarcodeHttpService) infoProductBarcodeByBarcode(holdingCode string, businessCode string, itemCode string, barcode string) (models.ProductBarcodeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	itemCode = utils.NormalizeBusinessCode(itemCode)
	barcode = utils.NormalizeBusinessCode(barcode)
	var findDoc models.ProductBarcodeDoc
	var err error
	if businessCode != "" && itemCode == "" {
		findDoc, err = svc.repo.FindByBarcodeInCompany(ctx, holdingCode, businessCode, barcode)
	} else if businessCode != "" {
		findDoc, err = svc.repo.FindByBusinessKeyInCompany(ctx, holdingCode, businessCode, itemCode, barcode)
	} else if itemCode == "" {
		findDoc, err = svc.repo.FindByBarcode(ctx, holdingCode, barcode)
	} else {
		findDoc, err = svc.repo.FindByBusinessKey(ctx, holdingCode, itemCode, barcode)
	}

	if err != nil {
		return models.ProductBarcodeInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.ProductBarcodeInfo{}, errors.New("document not found")
	}

	// Product relations resolve by the portable Product code, never by guidfixed.
	if strings.TrimSpace(findDoc.ItemCode) != "" {
		findMasterDoc := productmodels.ProductDoc{}
		if businessCode != "" {
			findMasterDoc, err = svc.repoMaster.FindByCodeInCompany(ctx, holdingCode, businessCode, utils.NormalizeBusinessCode(findDoc.ItemCode))
		} else {
			findMasterDoc, err = svc.repoMaster.FindOneByCode(ctx, holdingCode, utils.NormalizeBusinessCode(findDoc.ItemCode))
		}
		if err == nil && findMasterDoc.ID != primitive.NilObjectID {
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

			if findMasterDoc.ManufacturerGUID != "" {
				findManu, err := svc.repomgCreditror.FindByGuid(ctx, holdingCode, findMasterDoc.ManufacturerGUID)
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

			// Map new classification fields from Product to ProductBarcode
			findDoc.GroupsuboneGuid = findMasterDoc.GroupsuboneGuid
			findDoc.GroupsuboneCode = findMasterDoc.GroupsuboneCode
			findDoc.GroupsuboneNames = findMasterDoc.GroupsuboneNames

			findDoc.GroupsubtwoGuid = findMasterDoc.GroupsubtwoGuid
			findDoc.GroupsubtwoCode = findMasterDoc.GroupsubtwoCode
			findDoc.GroupsubtwoNames = findMasterDoc.GroupsubtwoNames

			findDoc.BrandGuid = findMasterDoc.BrandGuid
			findDoc.BrandCode = findMasterDoc.BrandCode
			findDoc.BrandNames = findMasterDoc.BrandNames

			findDoc.DesignGuid = findMasterDoc.DesignGuid
			findDoc.DesignCode = findMasterDoc.DesignCode
			findDoc.DesignNames = findMasterDoc.DesignNames

			findDoc.ModelGuid = findMasterDoc.ModelGuid
			findDoc.ModelCode = findMasterDoc.ModelCode
			findDoc.ModelNames = findMasterDoc.ModelNames

			findDoc.PatternGuid = findMasterDoc.PatternGuid
			findDoc.PatternCode = findMasterDoc.PatternCode
			findDoc.PatternNames = findMasterDoc.PatternNames

			findDoc.GradeGuid = findMasterDoc.GradeGuid
			findDoc.GradeCode = findMasterDoc.GradeCode
			findDoc.GradeNames = findMasterDoc.GradeNames

			findDoc.CategoryGuid = findMasterDoc.CategoryGuid
			findDoc.CategoryCode = findMasterDoc.CategoryCode
			findDoc.CategoryNames = findMasterDoc.CategoryNames

			findDoc.ClassGuid = findMasterDoc.ClassGuid
			findDoc.ClassCode = findMasterDoc.ClassCode
			findDoc.ClassNames = findMasterDoc.ClassNames

			findDoc.MaterialType = findMasterDoc.MaterialType
			findDoc.TaxType = findMasterDoc.TaxType
			findDoc.VatType = findMasterDoc.VatType
			findDoc.VatCal = int(findMasterDoc.VatType)

			if findMasterDoc.Manufacturers != nil {
				tempManuf := []models.ProductBarcodeManufacturer{}
				for _, m := range *findMasterDoc.Manufacturers {
					tempManuf = append(tempManuf, models.ProductBarcodeManufacturer{
						DocIdentity: common.DocIdentity{GuidFixed: m.GuidFixed},
						Code:        m.Code,
						Names:       m.Names,
					})
				}
				findDoc.Manufacturers = &tempManuf
			} else {
				findDoc.Manufacturers = &[]models.ProductBarcodeManufacturer{}
			}

			if findMasterDoc.Suppliers != nil {
				tempSuppl := []models.ProductBarcodeSupplier{}
				for _, s := range *findMasterDoc.Suppliers {
					tempSuppl = append(tempSuppl, models.ProductBarcodeSupplier{
						DocIdentity: common.DocIdentity{GuidFixed: s.GuidFixed},
						Code:        s.Code,
						Names:       s.Names,
					})
				}
				findDoc.Suppliers = &tempSuppl
			} else {
				findDoc.Suppliers = &[]models.ProductBarcodeSupplier{}
			}

			// Product controls shared color settings; Barcode keeps its own images.
			findDoc.UseImageOrColor = findMasterDoc.UseImageOrColor
			findDoc.ColorSelect = findMasterDoc.ColorSelect
			findDoc.ColorSelectHex = findMasterDoc.ColorSelectHex

			// Map POS / Restaurant settings
			findDoc.IsSumPoint = findMasterDoc.IsSumPoint
			findDoc.IsALaCarte = findMasterDoc.IsALaCarte
			findDoc.IsSplitUnitPrint = findMasterDoc.IsSplitUnitPrint
			findDoc.IsOnlyStaff = findMasterDoc.IsOnlyStaff
			findDoc.FoodType = findMasterDoc.FoodType
			findDoc.IsStockForRestaurant = findMasterDoc.IsStockForRestaurant

			findDoc.Restaurant = models.ProductRestaurant{
				IsForRestaurant:       findMasterDoc.Restaurant.IsForRestaurant,
				IsForTakeAway:         findMasterDoc.Restaurant.IsForTakeAway,
				IsForDelivery:         findMasterDoc.Restaurant.IsForDelivery,
				IsForCustomer:         findMasterDoc.Restaurant.IsForCustomer,
				IsForCustomerPreOrder: findMasterDoc.Restaurant.IsForCustomerPreOrder,
			}

			if findMasterDoc.OrderTypes != nil {
				tempOT := []models.ProductOrderType{}
				for _, ot := range *findMasterDoc.OrderTypes {
					row := models.ProductOrderType{
						Code:  ot.Code,
						Names: ot.Names,
						Price: ot.Price,
					}
					row.GuidFixed = ot.GuidFixed
					tempOT = append(tempOT, row)
				}
				findDoc.OrderTypes = &tempOT
			}

			if findMasterDoc.Options != nil {
				tempOpts := []models.ProductOption{}
				for _, opt := range *findMasterDoc.Options {
					tempChoices := []models.ProductChoice{}
					if opt.Choices != nil {
						for _, ch := range *opt.Choices {
							tempChoices = append(tempChoices, models.ProductChoice{
								GUID:            ch.GUID,
								Names:           ch.Names,
								ImageURI:        ch.ImageURI,
								RefBarcode:      ch.RefBarcode,
								RefBarcodeNames: ch.RefBarcodeNames,
								RefProductCode:  ch.RefProductCode,
								RefUnitCode:     ch.RefUnitCode,
								IsStock:         ch.IsStock,
								IsDefault:       ch.IsDefault,
								Qty:             ch.Qty,
								Price:           ch.Price,
							})
						}
					}
					tempOpts = append(tempOpts, models.ProductOption{
						GUID:       opt.GUID,
						Names:      opt.Names,
						ChoiceType: int8(opt.ChoiceType),
						MinSelect:  uint16(opt.MinSelect),
						MaxSelect:  uint16(opt.MaxSelect),
						Choices:    &tempChoices,
					})
				}
				findDoc.Options = &tempOpts
			}

			// Map shared alerts; Barcode keeps its own description.
			findDoc.IsAlert = findMasterDoc.IsAlert
			findDoc.AlertDescription = findMasterDoc.AlertDescription

			// Map TimeForSales
			if findMasterDoc.TimeForSales != nil {
				tempTfs := []models.ProductTimeForSale{}
				for _, tfs := range *findMasterDoc.TimeForSales {
					tempTfs = append(tempTfs, models.ProductTimeForSale{
						DaysOfWeek: tfs.DaysOfWeek,
						FromDate:   tfs.FromDate,
						ToDate:     tfs.ToDate,
						FromTime:   tfs.FromTime,
						ToTime:     tfs.ToTime,
					})
				}
				findDoc.TimeForSales = &tempTfs
			}

			// Map BusinessTypes and IgnoreBranches
			if findMasterDoc.BusinessTypes != nil {
				tempBt := []models.ProductBarcodeBusinessType{}
				for _, bt := range *findMasterDoc.BusinessTypes {
					row := models.ProductBarcodeBusinessType{
						Code:     bt.Code,
						Names:    bt.Names,
						IsIgnore: bt.IsIgnore,
					}
					row.GuidFixed = bt.GuidFixed
					tempBt = append(tempBt, row)
				}
				findDoc.BusinessTypes = &tempBt
			}
			if findMasterDoc.IgnoreBranches != nil {
				tempIb := []models.ProductBarcodeBranch{}
				for _, ib := range *findMasterDoc.IgnoreBranches {
					row := models.ProductBarcodeBranch{
						Code:     ib.Code,
						Names:    ib.Names,
						IsIgnore: ib.IsIgnore,
					}
					row.GuidFixed = ib.GuidFixed
					tempIb = append(tempIb, row)
				}
				findDoc.IgnoreBranches = &tempIb
			}

			// BOM remains a Barcode reference; unit conversions belong to Product.
			if findMasterDoc.BOM != nil {
				tempBOM := []models.BOMProductBarcode{}
				for _, b := range *findMasterDoc.BOM {
					tempBOM = append(tempBOM, models.BOMProductBarcode{
						BarcodeGuidFixed: b.BarcodeGuidFixed,
						ItemCode:         b.ItemCode,
						Level:            b.Level,
						Names:            b.Names,
						ItemUnitCode:     b.ItemUnitCode,
						ItemUnitNames:    b.ItemUnitNames,
						Barcode:          b.Barcode,
						Condition:        b.Condition,
						DivideValue:      b.DivideValue,
						StandValue:       b.StandValue,
						Qty:              b.Qty,
					})
				}
				findDoc.BOM = &tempBOM
			} else {
				findDoc.BOM = &[]models.BOMProductBarcode{}
			}
		}
	}

	if strings.TrimSpace(findDoc.ItemUnitGuid) != "" {
		unit, err := svc.repoUnit.FindByGuid(ctx, holdingCode, findDoc.ItemUnitGuid)
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

func (svc ProductBarcodeHttpService) InfoWTFArray(holdingCode string, codes []string) ([]interface{}, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList := []interface{}{}

	for _, code := range codes {
		findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "barcode", code)
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

func (svc ProductBarcodeHttpService) GetProductBarcodeByBarcodeRef(holdingCode string, barcodeRef string) ([]models.ProductBarcodeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDocs, err := svc.repo.FindByRefBarcode(ctx, holdingCode, barcodeRef)
	if err != nil {
		return nil, err
	}

	resultDocs := []models.ProductBarcodeInfo{}
	for _, findDoc := range findDocs {
		resultDocs = append(resultDocs, findDoc.ProductBarcodeInfo)
	}

	return resultDocs, nil
}

func (svc ProductBarcodeHttpService) GetProductBarcodeByBarcodeRefMultiShops(holdingCodes []string, barcodeRef string) ([]models.ProductBarcodeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	resultDocs := []models.ProductBarcodeInfo{}

	// Search across all specified shops
	for _, holdingCode := range holdingCodes {
		findDocs, err := svc.repo.FindByRefBarcode(ctx, holdingCode, barcodeRef)
		if err != nil {
			continue // Continue to next shop if error occurs
		}

		for _, findDoc := range findDocs {
			resultDocs = append(resultDocs, findDoc.ProductBarcodeInfo)
		}
	}

	return resultDocs, nil
}

func (svc ProductBarcodeHttpService) GetProductBarcodeByBarcodes(holdingCode string, barcodes []string) ([]models.ProductBarcodeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	results, err := svc.repo.FindByBarcodes(ctx, holdingCode, barcodes)
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

func (svc ProductBarcodeHttpService) SearchProductBarcode(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error) {
	return svc.searchProductBarcode(holdingCode, "", filters, pageable)
}

func (svc ProductBarcodeHttpService) SearchProductBarcodeInCompany(holdingCode string, businessCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error) {
	businessCode = utils.NormalizeBusinessCode(businessCode)
	if businessCode == "" {
		return []models.ProductBarcodeInfo{}, mongopagination.PaginationData{}, errors.New("BusinessCode is required")
	}
	return svc.searchProductBarcode(holdingCode, businessCode, filters, pageable)
}

func (svc ProductBarcodeHttpService) searchProductBarcode(holdingCode string, businessCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error) {

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
			barcodes, err := svc.getProductBarcodesFromShelves(ctx, holdingCode, shelfFilters)
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

	// Check if holdingcode filter exists (from shopsid parameter)
	var docList []models.ProductBarcodeInfo
	var pagination mongopagination.PaginationData
	var err error

	if businessCode != "" {
		docList, pagination, err = svc.repo.FindPageFilterInCompany(ctx, holdingCode, businessCode, filters, searchInFields, pageable)
	} else if _, hasHoldingCodeFilter := filters["holdingcode"]; hasHoldingCodeFilter {
		// Use FindPageFilterNoHoldingCode when holdingcode filter exists
		docList, pagination, err = svc.repo.FindPageFilterNoHoldingCode(ctx, filters, searchInFields, pageable)
	} else {
		// Use regular FindPageFilter when no holdingcode filter
		docList, pagination, err = svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)
	}

	if err != nil {
		return []models.ProductBarcodeInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc ProductBarcodeHttpService) SearchProductBarcode2(holdingCode string, pageable micromodels.Pageable) ([]models.ProductBarcodeSearch, common.Pagination, error) {

	//fixed holdingCode
	holdingCode = "2Eh6e3pfWvXTp0yV3CyFEhKPjdI"
	docList, pagination, err := svc.chRepo.Search(holdingCode, pageable)

	if err != nil {
		return []models.ProductBarcodeSearch{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc ProductBarcodeHttpService) SearchProductBarcodeStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ProductBarcodeInfo, int, error) {
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

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)
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

	docList, total, err := svc.repo.FindStepNoHoldingCode(ctx, filters, searchInFields, selectFields, pageableStep)
	if err != nil {
		return []models.ProductBarcodeInfo{}, 0, err
	}

	return docList, total, nil
}

func normalizeProductBarcodeImportData(dataList []models.ProductBarcode) []models.ProductBarcode {
	normalized := make([]models.ProductBarcode, len(dataList))
	for index, doc := range dataList {
		doc.ItemCode = utils.NormalizeBusinessCode(doc.ItemCode)
		doc.Barcode = utils.NormalizeBusinessCode(doc.Barcode)
		normalized[index] = doc
	}
	return normalized
}

func (svc ProductBarcodeHttpService) findProductBarcodesForImport(
	ctx context.Context,
	holdingCode string,
	payload []models.ProductBarcode,
) (map[string]models.ProductBarcodeDoc, error) {
	keys := make([]bson.M, 0, len(payload))
	for _, doc := range payload {
		keys = append(keys, bson.M{"itemcode": doc.ItemCode, "barcode": doc.Barcode})
	}
	if len(keys) == 0 {
		return map[string]models.ProductBarcodeDoc{}, nil
	}
	docs, err := svc.repo.Find(ctx, holdingCode, bson.M{"$or": keys})
	if err != nil {
		return nil, err
	}
	result := make(map[string]models.ProductBarcodeDoc, len(docs))
	for _, doc := range docs {
		result[svc.getDocIDKey(doc.ProductBarcode)] = doc
	}
	return result, nil
}

func (svc ProductBarcodeHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.ProductBarcode) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()
	if err := svc.repo.EnsureIndexes(ctx); err != nil {
		return common.BulkImport{}, err
	}

	dataList = normalizeProductBarcodeImportData(dataList)
	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.ProductBarcode](dataList, svc.getDocIDKey)
	existingByKey, err := svc.findProductBarcodesForImport(ctx, holdingCode, payloadList)
	if err != nil {
		return common.BulkImport{}, err
	}
	foundItemGuidList := make([]string, 0, len(existingByKey))
	for key := range existingByKey {
		foundItemGuidList = append(foundItemGuidList, key)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.ProductBarcode, models.ProductBarcodeDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.ProductBarcode) models.ProductBarcodeDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.ProductBarcodeDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.ProductBarcode = doc

			currentTime := time.Now().UTC()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.ProductBarcode, models.ProductBarcodeDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.ProductBarcodeDoc, error) {
			return existingByKey[guid], nil
		},
		func(doc models.ProductBarcodeDoc) bool {
			return doc.Barcode != ""
		},
		func(holdingCode string, authUsername string, dataReq models.ProductBarcode, doc models.ProductBarcodeDoc) error {

			docReq := models.ProductBarcodeRequest{}
			docReq.ProductBarcodeBase = dataReq.ProductBarcodeBase

			tempBarcodes := []models.BarcodeRequest{}

			if dataReq.RefBarcodes != nil {
				for _, docBarcode := range *dataReq.RefBarcodes {
					tempBarcodes = append(tempBarcodes, models.BarcodeRequest{
						ItemCode:    docBarcode.ItemCode,
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

			return svc.UpdateProductBarcode(holdingCode, doc.GuidFixed, authUsername, docReq)
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
							ItemCode:    docBarcode.ItemCode,
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
				_, err = svc.CreateProductBarcode(holdingCode, authUsername, docReq)

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

	svc.saveMasterSync(holdingCode)

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc ProductBarcodeHttpService) getDocIDKey(doc models.ProductBarcode) string {
	return doc.ItemCode + "\x00" + doc.Barcode
}

func (svc ProductBarcodeHttpService) XSortsSave(holdingCode string, authUsername string, xsorts []common.XSortModifyReqesut) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	for _, xsort := range xsorts {
		if len(xsort.GUIDFixed) < 1 {
			continue
		}
		findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, xsort.GUIDFixed)

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
		findDoc.UpdatedAt = time.Now().UTC()

		err = svc.repo.Update(ctx, holdingCode, findDoc.GuidFixed, findDoc)

		if err != nil {
			return err
		}

	}

	svc.saveMasterSync(holdingCode)

	return nil

}

func (svc ProductBarcodeHttpService) DeleteProductBarcodeByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	countRefBarcode, err := svc.repo.CountByRefGuids(ctx, holdingCode, GUIDs)

	if err != nil {
		return err
	}

	if countRefBarcode > 0 {
		return errors.New("document has other barcode referenced")
	}

	countBOM, err := svc.repo.CountByBOMGuids(ctx, holdingCode, GUIDs)

	if err != nil {
		return err
	}

	if countBOM > 0 {
		return errors.New("document has other bom referenced")
	}

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err = svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	svc.saveMasterSync(holdingCode)

	return nil
}

func (svc ProductBarcodeHttpService) GetProductBarcodeByUnits(holdingCode string, unitCodes []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	if len(unitCodes) < 1 {
		return []models.ProductBarcodeInfo{}, mongopagination.PaginationData{}, nil
	}

	results, pagination, err := svc.repo.FindPageByUnits(ctx, holdingCode, unitCodes, pageable)

	if err != nil {
		return []models.ProductBarcodeInfo{}, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (svc ProductBarcodeHttpService) GetProductBarcodeByGroups(holdingCode string, groupCodes []string, pageable micromodels.Pageable) ([]models.ProductBarcodeInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	if len(groupCodes) < 1 {
		return []models.ProductBarcodeInfo{}, mongopagination.PaginationData{}, nil
	}

	results, pagination, err := svc.repo.FindPageByGroups(ctx, holdingCode, groupCodes, pageable)

	if err != nil {
		return []models.ProductBarcodeInfo{}, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (svc ProductBarcodeHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc ProductBarcodeHttpService) GetModuleName() string {
	return "productbarcode"
}

func (svc ProductBarcodeHttpService) Export(holdingCode string, languageCode string, languageHeader map[string]string) ([][]string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	filters := bson.M{}

	docs, err := svc.repo.Find(ctx, holdingCode, filters)

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

func (svc ProductBarcodeHttpService) Import(holdingCode string, authUsername string, dataList []models.ProductBarcode) (common.BulkImport, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()
	if err := svc.repo.EnsureIndexes(ctx); err != nil {
		return common.BulkImport{}, err
	}

	dataList = normalizeProductBarcodeImportData(dataList)
	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.ProductBarcode](dataList, svc.getDocIDKey)
	existingByKey, err := svc.findProductBarcodesForImport(ctx, holdingCode, payloadList)
	if err != nil {
		return common.BulkImport{}, err
	}
	foundItemGuidList := make([]string, 0, len(existingByKey))
	for key := range existingByKey {
		foundItemGuidList = append(foundItemGuidList, key)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.ProductBarcode, models.ProductBarcodeDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.ProductBarcode) models.ProductBarcodeDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.ProductBarcodeDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.ProductBarcode = doc

			currentTime := time.Now().UTC()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	// do update
	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.ProductBarcode, models.ProductBarcodeDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.ProductBarcodeDoc, error) {
			return existingByKey[guid], nil
		},
		func(doc models.ProductBarcodeDoc) bool {
			return doc.Barcode != ""
		},
		func(holdingCode string, authUsername string, dataReq models.ProductBarcode, doc models.ProductBarcodeDoc) error {

			docReq := models.ProductBarcodeRequest{}
			docReq.ProductBarcodeBase = dataReq.ProductBarcodeBase

			tempBarcodes := []models.BarcodeRequest{}

			if dataReq.RefBarcodes != nil {
				for _, docBarcode := range *dataReq.RefBarcodes {
					tempBarcodes = append(tempBarcodes, models.BarcodeRequest{
						ItemCode:    docBarcode.ItemCode,
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

			// svc.UpdateProductBarcode(holdingCode, doc.GuidFixed, authUsername, docReq)

			// findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, doc.GuidFixed)
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
			if err := models.ValidateProductClassification(docData.ItemType, docData.MaterialType); err != nil {
				return err
			}
			docData.Barcode = doc.Barcode
			docData.IgnoreBranches = &docReq.IgnoreBranches
			docData.BusinessTypes = &docReq.BusinessTypes

			docData.UpdatedBy = authUsername
			docData.UpdatedAt = time.Now().UTC()

			err = svc.repo.Update(ctx, holdingCode, doc.GuidFixed, docData)
			if err != nil {
				return err
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
							ItemCode:    docBarcode.ItemCode,
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
				//_, err = svc.CreateProductBarcode(holdingCode, authUsername, docReq)

				newGuidFixed := utils.NewGUID()
				docData := models.ProductBarcodeDoc{}
				docData.HoldingCode = holdingCode
				docData.GuidFixed = newGuidFixed
				docData.ProductBarcode = docReq.ToProductBarcode()
				if err := models.ValidateProductClassification(docData.ItemType, docData.MaterialType); err != nil {
					return err
				}
				docData.IgnoreBranches = &docReq.IgnoreBranches
				docData.BusinessTypes = &docReq.BusinessTypes
				docData.CreatedBy = authUsername
				docData.CreatedAt = time.Now().UTC()

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
func (svc ProductBarcodeHttpService) GetPriceHistory(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error) {
	return svc.priceHistorySvc.GetPriceHistory(holdingCode, filters, pageable)
}

func (svc ProductBarcodeHttpService) GetPriceHistoryByBarcode(holdingCode string, itemCode string, barcode string, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error) {
	return svc.priceHistorySvc.GetPriceHistoryByBarcode(holdingCode, itemCode, barcode, pageable)
}

// getProductBarcodesFromShelves queried the OLD embedded warehouse.location[].shelf[].productitems[]
// structure, which the warehouse master-data redesign removed (see s/warehouse.md —
// warehouse/location/bin are now separate collections, and bin.fixeditems[] is a product-reference
// hint, not a queryable-here index yet). The current frontend never sends the warehousecode/
// locationcode/shelfcode query params that trigger this path, but they are still live and
// Swagger-documented on the public GET /productbarcode endpoint — an external caller (mobile app,
// integration, API consumer) using this filter must get an explicit "no longer supported" error,
// not a silent empty-but-200 result that looks like "zero products found" (Database No Silent
// Fallback Rule). A real replacement needs a new IWarehouseBinRepository dependency threaded
// through this service's 4 constructor call sites — a productbarcode-module change, out of scope
// for the warehouse backend redesign that removed the old structure this relied on.
func (svc ProductBarcodeHttpService) getProductBarcodesFromShelves(ctx context.Context, holdingCode string, shelfFilters map[string]interface{}) ([]string, error) {
	return nil, errors.New("การกรองสินค้าตามคลัง/ที่เก็บสินค้า/ชั้นวางเดิมไม่รองรับแล้วหลังปรับโครงสร้างคลังสินค้าใหม่ (warehousecode/locationcode/shelfcode filter no longer supported after the warehouse master-data redesign)")
}

func (s ProductBarcodeHttpService) ImportRefBarcodeUpdate(holdingCode, authUsername string, requests []models.RefBarcodeImportRequest) (*models.RefBarcodeImportResponse, error) {
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
		err := s.processBatchRefBarcodeUpdate(holdingCode, authUsername, batch, i, response)
		if err != nil {
			fmt.Printf("Error processing batch %d-%d: %v\n", i, end, err)
		}
	}

	if response.Failed > 0 {
		response.Success = false
	}

	return response, nil
}

func (s ProductBarcodeHttpService) processBatchRefBarcodeUpdate(holdingCode, authUsername string, batch []models.RefBarcodeImportRequest, startIndex int, response *models.RefBarcodeImportResponse) error {
	// Extract all barcodes for bulk lookup
	var allBarcodes []string
	barcodeMap := make(map[string]models.RefBarcodeImportRequest)

	for _, req := range batch {
		allBarcodes = append(allBarcodes, req.Barcode, req.BarcodeRef)
		barcodeMap[req.Barcode] = req
	}

	// Bulk find products by barcodes
	products, err := s.repo.FindByBarcodesMap(holdingCode, allBarcodes)
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
				"updatedby":        authUsername,
				"updatedat":        time.Now().UTC(),
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
		fmt.Printf("Updated refbarcodes for %s -> %s (standvalue: %d, dividevalue: %d, condition: %t)\n",
			req.Barcode, req.BarcodeRef, req.StandValue, req.DivideValue, condition)
	}

	return nil
}
