package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stockbalance/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IStockBalanceRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.StockBalanceDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.StockBalanceDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.StockBalanceDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.StockBalanceInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.StockBalanceDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.StockBalanceDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.StockBalanceItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.StockBalanceDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.StockBalanceInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.StockBalanceInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockBalanceDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockBalanceActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockBalanceDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockBalanceActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.StockBalanceDoc, error)
	Transaction(ctx context.Context, execFunc func(context.Context) error) error
}

type StockBalanceRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.StockBalanceDoc]
	repositories.SearchRepository[models.StockBalanceInfo]
	repositories.GuidRepository[models.StockBalanceItemGuid]
	repositories.ActivityRepository[models.StockBalanceActivity, models.StockBalanceDeleteActivity]
}

func NewStockBalanceRepository(pst microservice.IPersisterMongo) *StockBalanceRepository {

	insRepo := &StockBalanceRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.StockBalanceDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.StockBalanceInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.StockBalanceItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.StockBalanceActivity, models.StockBalanceDeleteActivity](pst)

	return insRepo
}

func (repo StockBalanceRepository) Transaction(ctx context.Context, execFunc func(context.Context) error) error {
	return repo.pst.Transaction(ctx, execFunc)
}

func (repo StockBalanceRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.StockBalanceDoc, error) {
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

	doc := models.StockBalanceDoc{}
	err := repo.pst.FindOne(ctx, models.StockBalanceDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
