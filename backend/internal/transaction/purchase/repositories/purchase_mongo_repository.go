package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/purchase/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IPurchaseRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.PurchaseDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PurchaseDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.PurchaseDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.PurchaseInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.PurchaseDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.PurchaseDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.PurchaseItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.PurchaseDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.PurchaseInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.PurchaseInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchaseDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchaseActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchaseDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchaseActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.PurchaseDoc, error)
}

type PurchaseRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PurchaseDoc]
	repositories.SearchRepository[models.PurchaseInfo]
	repositories.GuidRepository[models.PurchaseItemGuid]
	repositories.ActivityRepository[models.PurchaseActivity, models.PurchaseDeleteActivity]
}

func NewPurchaseRepository(pst microservice.IPersisterMongo) *PurchaseRepository {

	insRepo := &PurchaseRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.PurchaseDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PurchaseInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PurchaseItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.PurchaseActivity, models.PurchaseDeleteActivity](pst)

	return insRepo
}
func (repo PurchaseRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.PurchaseDoc, error) {
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

	doc := models.PurchaseDoc{}
	err := repo.pst.FindOne(ctx, models.PurchaseDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
