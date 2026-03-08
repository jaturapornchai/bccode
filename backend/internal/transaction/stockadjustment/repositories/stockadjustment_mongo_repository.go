package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stockadjustment/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IStockAdjustmentRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.StockAdjustmentDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.StockAdjustmentDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.StockAdjustmentDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.StockAdjustmentInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.StockAdjustmentDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.StockAdjustmentDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.StockAdjustmentItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.StockAdjustmentDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.StockAdjustmentInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.StockAdjustmentInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockAdjustmentDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockAdjustmentActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockAdjustmentDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockAdjustmentActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.StockAdjustmentDoc, error)
}

type StockAdjustmentRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.StockAdjustmentDoc]
	repositories.SearchRepository[models.StockAdjustmentInfo]
	repositories.GuidRepository[models.StockAdjustmentItemGuid]
	repositories.ActivityRepository[models.StockAdjustmentActivity, models.StockAdjustmentDeleteActivity]
}

func NewStockAdjustmentRepository(pst microservice.IPersisterMongo) *StockAdjustmentRepository {

	insRepo := &StockAdjustmentRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.StockAdjustmentDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.StockAdjustmentInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.StockAdjustmentItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.StockAdjustmentActivity, models.StockAdjustmentDeleteActivity](pst)

	return insRepo
}
func (repo StockAdjustmentRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.StockAdjustmentDoc, error) {
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

	doc := models.StockAdjustmentDoc{}
	err := repo.pst.FindOne(ctx, models.StockAdjustmentDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
