package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/saleorder/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ISaleOrderRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.SaleOrderDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.SaleOrderDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.SaleOrderDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.SaleOrderInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.SaleOrderDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.SaleOrderDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.SaleOrderItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.SaleOrderDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.SaleOrderInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.SaleOrderInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SaleOrderDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SaleOrderActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SaleOrderDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SaleOrderActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.SaleOrderDoc, error)
}

type SaleOrderRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.SaleOrderDoc]
	repositories.SearchRepository[models.SaleOrderInfo]
	repositories.GuidRepository[models.SaleOrderItemGuid]
	repositories.ActivityRepository[models.SaleOrderActivity, models.SaleOrderDeleteActivity]
}

func NewSaleOrderRepository(pst microservice.IPersisterMongo) *SaleOrderRepository {

	insRepo := &SaleOrderRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.SaleOrderDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.SaleOrderInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.SaleOrderItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.SaleOrderActivity, models.SaleOrderDeleteActivity](pst)

	return insRepo
}
func (repo SaleOrderRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.SaleOrderDoc, error) {
	filters := bson.M{
		"shopid": shopID,
		"deleted_at": bson.M{
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

	doc := models.SaleOrderDoc{}
	err := repo.pst.FindOne(ctx, models.SaleOrderDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
