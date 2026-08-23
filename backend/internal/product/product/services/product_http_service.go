package services

import (
	"context"
	"errors"
	"fmt"
	creditorRepo "smlcloudplatform/internal/debtaccount/creditor/repositories"
	"smlcloudplatform/internal/product/product/models"
	"smlcloudplatform/internal/product/product/repositories"
	barcodeModel "smlcloudplatform/internal/product/productbarcode/models"
	productBarcodeRepo "smlcloudplatform/internal/product/productbarcode/repositories"
	unitRepo "smlcloudplatform/internal/product/unit/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/apperr"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IProductHttpService interface {
	GetModuleName() string
	GetProduct(holdingCode string, businessCode string, code string) (*models.ProductDoc, error)
	ProductList(holdingCode string, businessCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductInfo, mongopagination.PaginationData, error)
	Create(doc *models.ProductDoc) error
	Update(holdingCode string, businessCode string, code string, authUsername string, doc *models.ProductDoc) (models.ProductDoc, error)
	Delete(holdingCode string, businessCode string, guid string, authUsername string) error
	Resync(holdingCode string, businessCode string) (int, error)
}

type ProductHttpService struct {
	repo                 repositories.IProductRepository
	repoUnit             unitRepo.IUnitRepository
	repomgCreditror      creditorRepo.CreditorRepository
	repomgProductBarcode productBarcodeRepo.ProductBarcodeRepository
	mqRepo               repositories.IProductMessageQueueRepository
	barcodeMQRepo        productBarcodeRepo.IProductBarcodeMessageQueueRepository
	contextTimeout       time.Duration
}

// ✅ **สร้าง Service**
func NewProductHttpService(repo repositories.IProductRepository, repoUnit unitRepo.IUnitRepository, repomgCreditror creditorRepo.CreditorRepository, repomgProductBarcode productBarcodeRepo.ProductBarcodeRepository, mqRepo repositories.IProductMessageQueueRepository, barcodeMQRepo productBarcodeRepo.IProductBarcodeMessageQueueRepository) *ProductHttpService {
	return &ProductHttpService{
		repo:                 repo,
		repoUnit:             repoUnit,
		repomgCreditror:      repomgCreditror,
		repomgProductBarcode: repomgProductBarcode,
		mqRepo:               mqRepo,
		barcodeMQRepo:        barcodeMQRepo,
		contextTimeout:       15 * time.Second,
	}
}

// ✅ **ตั้งค่า Timeout**
func (svc ProductHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ProductHttpService) GetModuleName() string {
	return "product"
}

// ✅ **GetProduct (ดึงข้อมูล Product)**
func (svc ProductHttpService) GetProduct(holdingCode string, businessCode string, code string) (*models.ProductDoc, error) {
	ctx, cancel := svc.getContextTimeout()
	defer cancel()

	// ✅ ดึงข้อมูล Product จาก PostgreSQL
	product, err := svc.repo.FindByGuidInCompany(ctx, holdingCode, businessCode, code)
	if err != nil {
		return nil, err
	}
	if product.ID == primitive.NilObjectID {
		product, err = svc.repo.FindByCodeInCompany(ctx, holdingCode, businessCode, utils.NormalizeBusinessCode(code))
		if err != nil {
			return nil, err
		}
		if product.ID == primitive.NilObjectID {
			return nil, errors.New("document not found")
		}
	}

	// ✅ ดึงข้อมูล Manufacturer ถ้ามีค่า `ManufacturerGUID`
	if product.ManufacturerGUID != "" {
		findDoc, err := svc.repomgCreditror.FindByGuid(ctx, holdingCode, product.ManufacturerGUID)
		if err == nil { // ไม่คืนค่า error ถ้าไม่เจอข้อมูล
			product.ManufacturerCode = findDoc.Code
			product.ManufacturerNames = findDoc.Names
		}
	}

	// ✅ ดึงข้อมูล Barcode จาก MongoDB
	barcodes, err := svc.repomgProductBarcode.FindByItemCodeInCompany(ctx, holdingCode, businessCode, product.Code)
	if err != nil {
		return nil, err
	}
	if barcodes == nil {
		barcodes = []barcodeModel.ProductBarcodeDoc{}
	}

	tempBarcodes := []models.Barcodes{}
	for _, barcode := range barcodes {
		tempPrices := []models.ProductPrice{}
		if barcode.Prices != nil { // ตรวจสอบก่อน loop
			for _, price := range *barcode.Prices {
				tempPrices = append(tempPrices, models.ProductPrice{
					KeyNumber: price.KeyNumber,
					Price:     price.Price,
				})
			}
		}
		divideValue, standValue, foundUnit := product.UnitRatio(barcode.ItemUnitCode)
		if !foundUnit {
			divideValue, standValue = 1, 1
		}

		// ✅ ตรวจสอบ `ItemType`
		if barcode.ItemType == 0 {
			tempImages := []models.ProductImage{}
			if barcode.Images != nil {
				for _, image := range *barcode.Images {
					tempImages = append(tempImages, models.ProductImage{XOrder: image.XOrder, URI: image.URI})
				}
			}
			tempVideos := []models.ProductVideo{}
			if barcode.Videos != nil {
				for _, video := range *barcode.Videos {
					tempVideos = append(tempVideos, models.ProductVideo{XOrder: video.XOrder, URI: video.URI, PosterURI: video.PosterURI})
				}
			}
			tempBarcodes = append(tempBarcodes, models.Barcodes{
				Barcode:       barcode.Barcode,
				ItemUnitCode:  barcode.ItemUnitCode,
				ItemUnitNames: barcode.ItemUnitNames,
				ImageURI:      barcode.ImageURI,
				Images:        &tempImages,
				Videos:        &tempVideos,
				Description:   barcode.Description,
				Prices:        &tempPrices,
				GuidFixed:     barcode.GuidFixed,
				Condition:     false,
				DivideValue:   divideValue,
				StandValue:    standValue,
				Qty:           1,
			})
		}
	}

	product.Barcodes = tempBarcodes // ✅ กำหนดค่า Barcodes ที่เป็น `[]` ถ้าไม่มีข้อมูล

	return &product, nil
}

func (svc ProductHttpService) ProductList(holdingCode string, businessCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductInfo, mongopagination.PaginationData, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"names.name",
		"code",
		"groupcode",
		"groupnames.name",
	}

	docList, pagination, err := svc.repo.FindPageFilterInCompany(ctx, holdingCode, businessCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.ProductInfo{}, pagination, err
	}

	return docList, pagination, nil
}

// ✅ ฟังก์ชันช่วยคำนวณค่า Min/Max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (svc ProductHttpService) ensureProductCodeAvailable(
	ctx context.Context,
	holdingCode string,
	businessCode string,
	code string,
	currentGuid string,
) error {
	product, err := svc.repo.FindByCodeInCompany(ctx, holdingCode, businessCode, code)
	if err != nil {
		return err
	}
	if product.ID != primitive.NilObjectID && product.GuidFixed != currentGuid {
		return errors.New("รหัสสินค้านี้มีอยู่แล้ว")
	}
	return nil
}

func enforceProductBaseUnit(doc *models.ProductDoc) error {
	doc.UnitCode = utils.NormalizeBusinessCode(doc.UnitCode)
	if doc.UnitCode == "" {
		return errors.New("กรุณาเลือกหน่วยนับมาตรฐาน")
	}
	doc.Condition = false
	doc.DivideValue = 1
	doc.StandValue = 1
	seen := map[string]struct{}{doc.UnitCode: {}}
	for index := range doc.UnitConversions {
		unit := &doc.UnitConversions[index]
		unit.UnitCode = utils.NormalizeBusinessCode(unit.UnitCode)
		if unit.UnitCode == "" {
			return fmt.Errorf("กรุณาเลือกหน่วยนับเพิ่มเติมรายการที่ %d", index+1)
		}
		if _, duplicate := seen[unit.UnitCode]; duplicate {
			return fmt.Errorf("รหัสหน่วยนับ %s ซ้ำกับหน่วยอื่นในสินค้า", unit.UnitCode)
		}
		if unit.DivideValue <= 0 || unit.StandValue <= 0 {
			return fmt.Errorf("อัตราส่วนของหน่วยนับ %s ต้องเป็นจำนวนเต็มมากกว่า 0", unit.UnitCode)
		}
		seen[unit.UnitCode] = struct{}{}
	}
	if doc.UnitConversions == nil {
		doc.UnitConversions = []models.ProductUnitConversion{}
	}
	return nil
}

func syncLinkedBarcodeUnitSnapshots(doc models.ProductDoc, barcodes []barcodeModel.ProductBarcodeDoc, actor string, updatedAt time.Time) error {
	for i := range barcodes {
		unit, ok := doc.UnitDefinition(barcodes[i].ItemUnitCode)
		if !ok {
			return fmt.Errorf("หน่วยนับ %s ยังถูกใช้โดยบาร์โค้ด %s", barcodes[i].ItemUnitCode, barcodes[i].Barcode)
		}
		barcodes[i].ItemUnitCode = unit.UnitCode
		barcodes[i].ItemUnitNames = unit.UnitNames
		barcodes[i].ItemUnitGuid = ""
		if unit.UnitCode == doc.UnitCode {
			barcodes[i].ItemUnitGuid = doc.UnitGuid
		}
		barcodes[i].Condition = false
		barcodes[i].DivideValue = unit.DivideValue
		barcodes[i].StandValue = unit.StandValue
		barcodes[i].UpdatedBy = actor
		barcodes[i].UpdatedAt = updatedAt
	}
	return nil
}

func productValidationError(err error) error {
	return apperr.ErrValidation.WithMessage(err.Error()).WithThaiMessage(err.Error()).WithWrap(err)
}

func (svc ProductHttpService) linkedBarcodeUnitSnapshots(ctx context.Context, doc models.ProductDoc, actor string, updatedAt time.Time) ([]barcodeModel.ProductBarcodeDoc, error) {
	barcodes, err := svc.repomgProductBarcode.FindByItemCodeInCompany(ctx, doc.HoldingCode, doc.BusinessCode, doc.Code)
	if err != nil {
		return nil, err
	}
	if err := syncLinkedBarcodeUnitSnapshots(doc, barcodes, actor, updatedAt); err != nil {
		return nil, productValidationError(err)
	}
	return barcodes, nil
}

// ✅ **Create (สร้าง Product ใหม่)**
func (svc ProductHttpService) Create(doc *models.ProductDoc) error {
	ctx, cancel := svc.getContextTimeout()
	defer cancel()

	doc.Code = utils.NormalizeBusinessCode(doc.Code)
	if doc.HoldingCode == "" || doc.BusinessCode == "" || doc.Code == "" {
		return errors.New("HoldingCode, BusinessCode and Code are required")
	}
	if err := svc.repo.EnsureIndexes(ctx); err != nil {
		return err
	}
	if err := svc.ensureProductCodeAvailable(ctx, doc.HoldingCode, doc.BusinessCode, doc.Code, ""); err != nil {
		return err
	}
	if err := enforceProductBaseUnit(doc); err != nil {
		return productValidationError(err)
	}

	if err := barcodeModel.ValidateProductClassification(doc.ItemType, doc.MaterialType); err != nil {
		return err
	}

	// ✅ สร้าง `GuidFixed` ถ้ายังไม่มีค่า
	if doc.GuidFixed == "" {
		doc.GuidFixed = utils.NewGUID() // 🔥 สร้าง GUID ใหม่
	}

	// ✅ กำหนดค่าเริ่มต้นให้ `itemtype` หากไม่ได้ส่งมา
	if doc.ItemType == 0 {
		doc.ItemType = 0
	}

	// ✅ ตั้งค่าเวลาก่อนสร้าง
	doc.CreatedAt = time.Now().UTC()
	doc.UpdatedAt = doc.CreatedAt

	// ✅ เรียก `Create()`
	_, err := svc.repo.Create(ctx, *doc)

	if err != nil {
		return err
	}

	if err := svc.mqRepo.Create(*doc); err != nil {
		return err
	}

	return nil
}

// ✅ **Update (อัปเดต Product)**
func (svc ProductHttpService) Update(holdingCode string, businessCode string, code string, authUsername string, doc *models.ProductDoc) (models.ProductDoc, error) {
	ctx, cancel := svc.getContextTimeout()
	defer cancel()

	if holdingCode == "" || businessCode == "" || code == "" {
		return models.ProductDoc{}, errors.New("HoldingCode, BusinessCode and Code are required")
	}

	findDoc, err := svc.repo.FindByGuidInCompany(ctx, holdingCode, businessCode, code)

	if err != nil {
		return models.ProductDoc{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		findDoc, err = svc.repo.FindByCodeInCompany(ctx, holdingCode, businessCode, utils.NormalizeBusinessCode(code))
		if err != nil {
			return models.ProductDoc{}, err
		}

		if findDoc.ID == primitive.NilObjectID {
			return models.ProductDoc{}, errors.New("document not found")
		}
	}
	doc.Code = utils.NormalizeBusinessCode(doc.Code)
	if doc.Code == "" {
		return models.ProductDoc{}, errors.New("Code is required")
	}
	if doc.Code != utils.NormalizeBusinessCode(findDoc.Code) {
		return models.ProductDoc{}, errors.New("ไม่สามารถเปลี่ยนรหัสสินค้าในหน้าจอแก้ไขได้")
	}
	if err := svc.repo.EnsureIndexes(ctx); err != nil {
		return models.ProductDoc{}, err
	}
	if err := svc.ensureProductCodeAvailable(ctx, holdingCode, businessCode, doc.Code, findDoc.GuidFixed); err != nil {
		return models.ProductDoc{}, err
	}
	docData := findDoc
	docData.ProductData = doc.ProductData
	docData.Code = doc.Code
	// ProductData embeds HoldingCodeentity + DocIdentity, so the assignment above replaces
	// holdingcode/guidfixed with whatever the request body carried (often empty) — which detaches
	// the product from its tenant and breaks every holdingcode-scoped read. Identity always comes
	// from the stored doc, never the request.
	docData.HoldingCode = findDoc.HoldingCode
	docData.BusinessCode = findDoc.BusinessCode
	docData.GuidFixed = findDoc.GuidFixed
	if err := enforceProductBaseUnit(&docData); err != nil {
		return models.ProductDoc{}, productValidationError(err)
	}
	now := time.Now().UTC()
	if err := barcodeModel.ValidateProductClassification(docData.ItemType, docData.MaterialType); err != nil {
		return models.ProductDoc{}, err
	}

	docData.UpdatedBy = authUsername
	docData.UpdatedAt = now

	// Keep the Product-owned unit definitions and every linked Barcode snapshot atomic.
	var linkedBarcodes []barcodeModel.ProductBarcodeDoc
	errx := svc.repomgProductBarcode.Transaction(ctx, func(txCtx context.Context) error {
		var err error
		linkedBarcodes, err = svc.linkedBarcodeUnitSnapshots(txCtx, docData, authUsername, now)
		if err != nil {
			return err
		}
		if err := svc.repo.UpdateInCompany(txCtx, holdingCode, businessCode, findDoc.GuidFixed, docData); err != nil {
			return err
		}
		return svc.repomgProductBarcode.UpdateUnitSnapshotsInCompany(txCtx, holdingCode, businessCode, linkedBarcodes)
	})
	if errx != nil {
		return models.ProductDoc{}, errx
	}

	if err := svc.mqRepo.Update(docData); err != nil {
		return models.ProductDoc{}, err
	}
	if len(linkedBarcodes) > 0 {
		if err := svc.barcodeMQRepo.UpdateInBatch(linkedBarcodes); err != nil {
			return models.ProductDoc{}, err
		}
	}

	return docData, nil
}

// ✅ **Delete (ลบ Product)**
func (svc ProductHttpService) Delete(holdingCode string, businessCode string, guid string, user string) error {
	ctx, cancel := svc.getContextTimeout()
	defer cancel()

	if holdingCode == "" || businessCode == "" || guid == "" {
		return errors.New("HoldingCode, BusinessCode and Code are required")
	}

	deleteGuid := guid
	findDoc, err := svc.repo.FindByGuidInCompany(ctx, holdingCode, businessCode, guid)
	if err != nil {
		return err
	}
	if findDoc.ID == primitive.NilObjectID {
		findDoc, err = svc.repo.FindByCodeInCompany(ctx, holdingCode, businessCode, utils.NormalizeBusinessCode(guid))
		if err != nil {
			return err
		}
		if findDoc.ID == primitive.NilObjectID {
			return errors.New("document not found")
		}
		deleteGuid = findDoc.GuidFixed
	}

	err = svc.repo.DeleteByGuidfixedInCompany(ctx, holdingCode, businessCode, deleteGuid, user)
	if err != nil {
		return err
	}

	if err := svc.mqRepo.Delete(findDoc); err != nil {
		return err
	}

	return nil
}

// ✅ Resync — republish ทุก product ของ tenant เข้า Kafka เพื่อ rebuild PG projection
func (svc ProductHttpService) Resync(holdingCode string, businessCode string) (int, error) {
	ctx, cancel := svc.getContextTimeout()
	defer cancel()

	docs, err := svc.repo.FindFilterInCompany(ctx, holdingCode, businessCode, map[string]interface{}{})
	if err != nil {
		return 0, err
	}

	published := 0
	for _, doc := range docs {
		if err := svc.mqRepo.Create(doc); err != nil {
			return published, err
		}
		published++
	}

	return published, nil
}
