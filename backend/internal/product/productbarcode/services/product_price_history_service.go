package services

import (
	"context"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/product/productbarcode/repositories"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IProductPriceHistoryService interface {
	RecordPriceChange(ctx context.Context, shopID string, productBarcodeGUID string, barcode string, productName string, oldPrices, newPrices []models.ProductPrice, action string, username string, remark string) error
	GetPriceHistory(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error)
	GetPriceHistoryByBarcode(shopID string, barcode string, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error)
	GetPriceHistoryByProductGUID(shopID string, productBarcodeGUID string, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error)
}

type ProductPriceHistoryService struct {
	repo         repositories.IProductPriceHistoryRepository
	generateGUID func() string
	timeNow      func() time.Time
}

func NewProductPriceHistoryService(
	repo repositories.IProductPriceHistoryRepository,
	generateGUID func() string,
	timeNow func() time.Time,
) *ProductPriceHistoryService {
	return &ProductPriceHistoryService{
		repo:         repo,
		generateGUID: generateGUID,
		timeNow:      timeNow,
	}
}

func (svc *ProductPriceHistoryService) RecordPriceChange(
	ctx context.Context,
	shopID string,
	productBarcodeGUID string,
	barcode string,
	productName string,
	oldPrices, newPrices []models.ProductPrice,
	action string,
	username string,
	remark string,
) error {
	now := svc.timeNow()

	// สร้าง map สำหรับ oldPrices เพื่อเข้าถึงได้ง่าย
	oldPriceMap := make(map[int]float64)
	for _, price := range oldPrices {
		oldPriceMap[price.KeyNumber] = price.Price
	}

	var histories []models.ProductPriceHistory

	// เปรียบเทียบราคาแต่ละประเภท
	for _, newPrice := range newPrices {
		oldPrice := oldPriceMap[newPrice.KeyNumber]

		// บันทึกเฉพาะเมื่อราคาเปลี่ยน หรือเป็นการสร้างใหม่
		if action == "create" || oldPrice != newPrice.Price {
			priceType := getPriceTypeName(newPrice.KeyNumber)

			history := models.ProductPriceHistory{}
			history.ShopIdentity.ShopID = shopID
			history.DocIdentity.GuidFixed = svc.generateGUID()
			history.ProductBarcodeGUID = productBarcodeGUID
			history.Barcode = barcode
			history.ProductName = productName
			history.PriceType = priceType
			history.KeyNumber = newPrice.KeyNumber
			history.OldPrice = oldPrice
			history.NewPrice = newPrice.Price
			history.PriceDifference = newPrice.Price - oldPrice
			history.Action = action
			history.CreatedBy = username
			history.CreatedAt = now
			history.Remark = remark

			histories = append(histories, history)
		}
	}

	// บันทึกประวัติทั้งหมด
	if len(histories) > 0 {
		return svc.repo.CreateInBatch(ctx, histories)
	}

	return nil
}

func (svc *ProductPriceHistoryService) GetPriceHistory(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	searchInFields := []string{"barcode", "productname", "createdby"}
	return svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)
}

func (svc *ProductPriceHistoryService) GetPriceHistoryByBarcode(shopID string, barcode string, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return svc.repo.FindByBarcode(ctx, shopID, barcode, pageable)
}

func (svc *ProductPriceHistoryService) GetPriceHistoryByProductGUID(shopID string, productBarcodeGUID string, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return svc.repo.FindByProductBarcode(ctx, shopID, productBarcodeGUID, pageable)
}

// Helper function เพื่อแปลง KeyNumber เป็นชื่อประเภทราคา
func getPriceTypeName(keyNumber int) string {
	switch {
	case keyNumber == 1:
		return "normal"
	case keyNumber == 2:
		return "member"
	case keyNumber >= 3 && keyNumber <= 5:
		return "delivery"
	case keyNumber >= 6 && keyNumber <= 17:
		return "price level"
	default:
		return "other"
	}
}
