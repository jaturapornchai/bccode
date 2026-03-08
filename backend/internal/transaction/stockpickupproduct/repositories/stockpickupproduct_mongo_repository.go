package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stockpickupproduct/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IStockPickupProductRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.StockPickupProductDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.StockPickupProductDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.StockPickupProductDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.StockPickupProductInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.StockPickupProductDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.StockPickupProductDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.StockPickupProductItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.StockPickupProductDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.StockPickupProductInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.StockPickupProductInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockPickupProductDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockPickupProductActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockPickupProductDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockPickupProductActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.StockPickupProductDoc, error)
}

type StockPickupProductRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.StockPickupProductDoc]
	repositories.SearchRepository[models.StockPickupProductInfo]
	repositories.GuidRepository[models.StockPickupProductItemGuid]
	repositories.ActivityRepository[models.StockPickupProductActivity, models.StockPickupProductDeleteActivity]
}

func NewStockPickupProductRepository(pst microservice.IPersisterMongo) *StockPickupProductRepository {

	insRepo := &StockPickupProductRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.StockPickupProductDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.StockPickupProductInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.StockPickupProductItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.StockPickupProductActivity, models.StockPickupProductDeleteActivity](pst)

	return insRepo
}
func (repo StockPickupProductRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.StockPickupProductDoc, error) {
	filters := bson.M{
		"shopid": shopID,
		"deletedat": bson.M{
			"$exists": false,
		},
		"docno": bson.M{
			"$regex": "^" + prefixDocNo + ".*$",
		},
	}

	optSort := options.FindOneOptions{}
	optSort.SetSort(bson.M{
		"docno": -1,
	})

	doc := models.StockPickupProductDoc{}
	err := repo.pst.FindOne(ctx, models.StockPickupProductDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
