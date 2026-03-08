package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/stocktransfer/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IStockTransferRepository interface {
	FindDocOne(ctx context.Context, shopID, docno string, transFlag int) (models.StockTransferDoc, error)
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.StockTransferDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.StockTransferDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.StockTransferDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.StockTransferInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.StockTransferDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.StockTransferDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.StockTransferItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.StockTransferDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.StockTransferInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.StockTransferInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockTransferDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockTransferActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockTransferDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.StockTransferActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.StockTransferDoc, error)
}

type StockTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.StockTransferDoc]
	repositories.SearchRepository[models.StockTransferInfo]
	repositories.GuidRepository[models.StockTransferItemGuid]
	repositories.ActivityRepository[models.StockTransferActivity, models.StockTransferDeleteActivity]
}

func NewStockTransferRepository(pst microservice.IPersisterMongo) *StockTransferRepository {

	insRepo := &StockTransferRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.StockTransferDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.StockTransferInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.StockTransferItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.StockTransferActivity, models.StockTransferDeleteActivity](pst)

	return insRepo
}

func (repo StockTransferRepository) FindDocOne(ctx context.Context, shopID, docno string, transFlag int) (models.StockTransferDoc, error) {
	doc := models.StockTransferDoc{}
	err := repo.pst.FindOne(ctx, models.StockTransferDoc{}, bson.M{"shopid": shopID, "docno": docno, "transflag": transFlag}, &doc)

	if err != nil {
		return doc, err
	}

	return doc, nil
}

func (repo StockTransferRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.StockTransferDoc, error) {
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

	doc := models.StockTransferDoc{}
	err := repo.pst.FindOne(ctx, models.StockTransferDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
