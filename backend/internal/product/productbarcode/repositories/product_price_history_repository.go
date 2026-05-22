package repositories

import (
	"context"
	"smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IProductPriceHistoryRepository interface {
	Create(ctx context.Context, doc models.ProductPriceHistory) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ProductPriceHistory) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error)
	FindByProductBarcode(ctx context.Context, shopID string, productBarcodeGUID string, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error)
	FindByBarcode(ctx context.Context, shopID string, barcode string, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error)
	CountByDateRange(ctx context.Context, shopID string, fromDate, toDate time.Time) (int, error)
}

type ProductPriceHistoryRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ProductPriceHistory]
	repositories.SearchRepository[models.ProductPriceHistoryInfo]
}

func NewProductPriceHistoryRepository(pst microservice.IPersisterMongo) *ProductPriceHistoryRepository {
	repo := &ProductPriceHistoryRepository{
		pst: pst,
	}

	repo.CrudRepository = repositories.NewCrudRepository[models.ProductPriceHistory](pst)
	repo.SearchRepository = repositories.NewSearchRepository[models.ProductPriceHistoryInfo](pst)

	return repo
}

func (repo ProductPriceHistoryRepository) FindByProductBarcode(ctx context.Context, shopID string, productBarcodeGUID string, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error) {
	filters := map[string]interface{}{
		"productbarcodeguid": productBarcodeGUID,
	}

	searchInFields := []string{"barcode", "productname", "createdby"}

	return repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)
}

func (repo ProductPriceHistoryRepository) FindByBarcode(ctx context.Context, shopID string, barcode string, pageable micromodels.Pageable) ([]models.ProductPriceHistoryInfo, mongopagination.PaginationData, error) {
	filters := map[string]interface{}{
		"barcode": barcode,
	}

	searchInFields := []string{"barcode", "productname", "createdby"}

	return repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)
}

func (repo ProductPriceHistoryRepository) CountByDateRange(ctx context.Context, shopID string, fromDate, toDate time.Time) (int, error) {
	filters := map[string]interface{}{
		"created_at": map[string]interface{}{
			"$gte": fromDate,
			"$lte": toDate,
		},
	}

	result, err := repo.pst.Count(ctx, models.ProductPriceHistory{}, filters)
	return int(result), err
}
