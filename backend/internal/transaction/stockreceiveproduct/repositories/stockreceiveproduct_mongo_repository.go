package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stockreceiveproduct/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IStockReceiveProductRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.StockReceiveProductDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.StockReceiveProductDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.StockReceiveProductDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.StockReceiveProductInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.StockReceiveProductDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.StockReceiveProductDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.StockReceiveProductItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.StockReceiveProductDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.StockReceiveProductInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.StockReceiveProductInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockReceiveProductDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockReceiveProductActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockReceiveProductDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockReceiveProductActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.StockReceiveProductDoc, error)
}

type StockReceiveProductRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.StockReceiveProductDoc]
	repositories.SearchRepository[models.StockReceiveProductInfo]
	repositories.GuidRepository[models.StockReceiveProductItemGuid]
	repositories.ActivityRepository[models.StockReceiveProductActivity, models.StockReceiveProductDeleteActivity]
}

func NewStockReceiveProductRepository(pst microservice.IPersisterMongo) *StockReceiveProductRepository {

	insRepo := &StockReceiveProductRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.StockReceiveProductDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.StockReceiveProductInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.StockReceiveProductItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.StockReceiveProductActivity, models.StockReceiveProductDeleteActivity](pst)

	return insRepo
}

func (repo StockReceiveProductRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.StockReceiveProductDoc, error) {
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

	doc := models.StockReceiveProductDoc{}
	err := repo.pst.FindOne(ctx, models.StockReceiveProductDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
