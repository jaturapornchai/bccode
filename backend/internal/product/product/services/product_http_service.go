package services

import (
	"context"
	"errors"
	creditorRepo "smlcloudplatform/internal/debtaccount/creditor/repositories"
	"smlcloudplatform/internal/product/product/models"
	"smlcloudplatform/internal/product/product/repositories"
	barcodeModel "smlcloudplatform/internal/product/productbarcode/models"
	productBarcodeRepo "smlcloudplatform/internal/product/productbarcode/repositories"
	unitRepo "smlcloudplatform/internal/product/unit/repositories"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IProductHttpService interface {
	GetModuleName() string
	GetProduct(holdingCode string, code string) (*models.ProductDoc, error)
	ProductList(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductInfo, mongopagination.PaginationData, error)
	Create(doc *models.ProductDoc) error
	Update(holdingCode string, code string, authUsername string, doc *models.ProductDoc) (models.ProductDoc, error)
	Delete(holdingCode string, guid string, authUsername string) error
}

type ProductHttpService struct {
	repo                 repositories.IProductRepository
	repoUnit             unitRepo.IUnitRepository
	repomgCreditror      creditorRepo.CreditorRepository
	repomgProductBarcode productBarcodeRepo.ProductBarcodeRepository
	contextTimeout       time.Duration
}

// ✅ **สร้าง Service**
func NewProductHttpService(repo repositories.IProductRepository, repoUnit unitRepo.IUnitRepository, repomgCreditror creditorRepo.CreditorRepository, repomgProductBarcode productBarcodeRepo.ProductBarcodeRepository) *ProductHttpService {
	return &ProductHttpService{
		repo:                 repo,
		repoUnit:             repoUnit,
		repomgCreditror:      repomgCreditror,
		repomgProductBarcode: repomgProductBarcode,
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
func (svc ProductHttpService) GetProduct(holdingCode string, code string) (*models.ProductDoc, error) {
	ctx, cancel := svc.getContextTimeout()
	defer cancel()

	// ✅ ดึงข้อมูล Product จาก PostgreSQL
	product, err := svc.repo.FindByGuid(ctx, holdingCode, code)
	if err != nil {
		return nil, err
	}
	if product.ID == primitive.NilObjectID {
		product, err = svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)
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
	barcodes, err := svc.repomgProductBarcode.FindByItemCode(ctx, holdingCode, product.Code)
	if err != nil || barcodes == nil {
		// ถ้าไม่เจอข้อมูล หรือเกิดข้อผิดพลาด ให้ตั้งค่า barcodes = []
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
		var condition bool
		var divideValue, standValue, qty float64
		if barcode.RefBarcodes != nil && len(*barcode.RefBarcodes) > 0 { // 🔥 ต้อง Dereference `*RefBarcodes`
			refBarcode := (*barcode.RefBarcodes)[0] // ดึงค่าตัวแรกออกมาใช้งาน
			condition = refBarcode.Condition
			divideValue = refBarcode.DivideValue
			standValue = refBarcode.StandValue
			qty = refBarcode.Qty
		} else {
			condition = false
			divideValue = 1
			standValue = 1
			qty = 1
		}

		// ✅ ตรวจสอบ `ItemType`
		if barcode.ItemType == 0 {
			tempBarcodes = append(tempBarcodes, models.Barcodes{
				Barcode:       barcode.Barcode,
				ItemUnitCode:  barcode.ItemUnitCode,
				ItemUnitNames: barcode.ItemUnitNames,
				Prices:        &tempPrices,
				GuidFixed:     barcode.GuidFixed,
				Condition:     condition,
				DivideValue:   divideValue,
				StandValue:    standValue,
				Qty:           qty,
				IsMainBarcode: barcode.IsMainBarcode,
			})
		}
	}

	product.Barcodes = tempBarcodes // ✅ กำหนดค่า Barcodes ที่เป็น `[]` ถ้าไม่มีข้อมูล

	return &product, nil
}

func (svc ProductHttpService) ProductList(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductInfo, mongopagination.PaginationData, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"names.name",
		"code",
		"groupcode",
		"groupnames.name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

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

// ✅ **Create (สร้าง Product ใหม่)**
func (svc ProductHttpService) Create(doc *models.ProductDoc) error {
	ctx, cancel := svc.getContextTimeout()
	defer cancel()

	if doc.HoldingCode == "" || doc.Code == "" {
		return errors.New("HoldingCode and Code are required")
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
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()

	// ✅ เรียก `Create()`
	_, err := svc.repo.Create(ctx, *doc)

	if err != nil {
		return err
	}

	return nil
}

// ✅ **Update (อัปเดต Product)**
func (svc ProductHttpService) Update(holdingCode string, code string, authUsername string, doc *models.ProductDoc) (models.ProductDoc, error) {
	ctx, cancel := svc.getContextTimeout()
	defer cancel()

	if holdingCode == "" || code == "" {
		return models.ProductDoc{}, errors.New("HoldingCode and Code are required")
	}

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, code)

	if err != nil {
		return models.ProductDoc{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		findDoc, err = svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)
		if err != nil {
			return models.ProductDoc{}, err
		}

		if findDoc.ID == primitive.NilObjectID {
			return models.ProductDoc{}, errors.New("document not found")
		}
	}
	docData := findDoc
	docData.ProductData = doc.ProductData
	docData.Code = doc.Code
	// ProductData embeds HoldingCodeentity + DocIdentity, so the assignment above replaces
	// holdingcode/guidfixed with whatever the request body carried (often empty) — which detaches
	// the product from its tenant and breaks every holdingcode-scoped read. Identity always comes
	// from the stored doc, never the request.
	docData.HoldingCode = findDoc.HoldingCode
	docData.GuidFixed = findDoc.GuidFixed
	if err := barcodeModel.ValidateProductClassification(docData.ItemType, docData.MaterialType); err != nil {
		return models.ProductDoc{}, err
	}

	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	// ✅ เรียก Repository เพื่ออัปเดตข้อมูล
	errx := svc.repo.Update(ctx, holdingCode, code, docData)
	if errx != nil {
		return models.ProductDoc{}, errx
	}

	return docData, nil
}

// ✅ **Delete (ลบ Product)**
func (svc ProductHttpService) Delete(holdingCode string, guid string, user string) error {
	ctx, cancel := svc.getContextTimeout()
	defer cancel()

	if holdingCode == "" || guid == "" {
		return errors.New("HoldingCode and Code are required")
	}

	deleteGuid := guid
	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)
	if err != nil {
		return err
	}
	if findDoc.ID == primitive.NilObjectID {
		findDoc, err = svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", guid)
		if err != nil {
			return err
		}
		if findDoc.ID == primitive.NilObjectID {
			return errors.New("document not found")
		}
		deleteGuid = findDoc.GuidFixed
	}

	err = svc.repo.DeleteByGuidfixed(ctx, holdingCode, deleteGuid, user)
	if err != nil {
		return err
	}

	return nil
}
