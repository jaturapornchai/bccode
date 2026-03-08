package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/purchaseorder/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IPurchaseOrderRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.PurchaseOrderDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PurchaseOrderDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.PurchaseOrderDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.PurchaseOrderInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.PurchaseOrderDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.PurchaseOrderDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.PurchaseOrderItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.PurchaseOrderDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.PurchaseOrderInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.PurchaseOrderInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchaseOrderDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchaseOrderActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchaseOrderDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchaseOrderActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.PurchaseOrderDoc, error)
}

type PurchaseOrderRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PurchaseOrderDoc]
	repositories.SearchRepository[models.PurchaseOrderInfo]
	repositories.GuidRepository[models.PurchaseOrderItemGuid]
	repositories.ActivityRepository[models.PurchaseOrderActivity, models.PurchaseOrderDeleteActivity]
}

func NewPurchaseOrderRepository(pst microservice.IPersisterMongo) *PurchaseOrderRepository {

	insRepo := &PurchaseOrderRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.PurchaseOrderDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PurchaseOrderInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PurchaseOrderItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.PurchaseOrderActivity, models.PurchaseOrderDeleteActivity](pst)

	return insRepo
}
func (repo PurchaseOrderRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.PurchaseOrderDoc, error) {
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

	doc := models.PurchaseOrderDoc{}
	err := repo.pst.FindOne(ctx, models.PurchaseOrderDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
