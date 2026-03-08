package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stockreturnproduct/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IStockReturnProductRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.StockReturnProductDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.StockReturnProductDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.StockReturnProductDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.StockReturnProductInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.StockReturnProductDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.StockReturnProductDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.StockReturnProductItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.StockReturnProductDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.StockReturnProductInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.StockReturnProductInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockReturnProductDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockReturnProductActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockReturnProductDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockReturnProductActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.StockReturnProductDoc, error)
}

type StockReturnProductRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.StockReturnProductDoc]
	repositories.SearchRepository[models.StockReturnProductInfo]
	repositories.GuidRepository[models.StockReturnProductItemGuid]
	repositories.ActivityRepository[models.StockReturnProductActivity, models.StockReturnProductDeleteActivity]
}

func NewStockReturnProductRepository(pst microservice.IPersisterMongo) *StockReturnProductRepository {

	insRepo := &StockReturnProductRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.StockReturnProductDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.StockReturnProductInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.StockReturnProductItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.StockReturnProductActivity, models.StockReturnProductDeleteActivity](pst)

	return insRepo
}

func (repo StockReturnProductRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.StockReturnProductDoc, error) {
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

	doc := models.StockReturnProductDoc{}
	err := repo.pst.FindOne(ctx, models.StockReturnProductDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
