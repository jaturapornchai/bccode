package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/saleinvoicebomprice/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ISaleInvoiceBomPriceRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.SaleInvoiceBomPriceDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.SaleInvoiceBomPriceDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.SaleInvoiceBomPriceDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.SaleInvoiceBomPriceInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.SaleInvoiceBomPriceDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.SaleInvoiceBomPriceDoc, error)
	FindByDocNo(ctx context.Context, shopID string, docNo string) ([]models.SaleInvoiceBomPriceInfo, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.SaleInvoiceBomPriceItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.SaleInvoiceBomPriceDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.SaleInvoiceBomPriceInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.SaleInvoiceBomPriceInfo, int, error)
}

type SaleInvoiceBomPriceRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.SaleInvoiceBomPriceDoc]
	repositories.SearchRepository[models.SaleInvoiceBomPriceInfo]
	repositories.GuidRepository[models.SaleInvoiceBomPriceItemGuid]
}

func NewSaleInvoiceBomPriceRepository(pst microservice.IPersisterMongo) *SaleInvoiceBomPriceRepository {

	insRepo := &SaleInvoiceBomPriceRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.SaleInvoiceBomPriceDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.SaleInvoiceBomPriceInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.SaleInvoiceBomPriceItemGuid](pst)

	return insRepo
}

func (repo SaleInvoiceBomPriceRepository) FindByDocNo(ctx context.Context, shopID string, docNo string) ([]models.SaleInvoiceBomPriceInfo, error) {
	var docs []models.SaleInvoiceBomPriceInfo
	filters := map[string]interface{}{
		"shopid":    shopID,
		"deletedat": bson.M{"$exists": false},
		"docno":     docNo,
	}
	err := repo.pst.Find(ctx, models.SaleInvoiceBomPriceInfo{}, filters, &docs)
	return docs, err
}
