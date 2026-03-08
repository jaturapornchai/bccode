package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/purchasereturn/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IPurchaseReturnRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.PurchaseReturnDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PurchaseReturnDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.PurchaseReturnDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.PurchaseReturnInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.PurchaseReturnDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.PurchaseReturnDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.PurchaseReturnItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.PurchaseReturnDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.PurchaseReturnInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.PurchaseReturnInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchaseReturnDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchaseReturnActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchaseReturnDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchaseReturnActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.PurchaseReturnDoc, error)
}

type PurchaseReturnRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PurchaseReturnDoc]
	repositories.SearchRepository[models.PurchaseReturnInfo]
	repositories.GuidRepository[models.PurchaseReturnItemGuid]
	repositories.ActivityRepository[models.PurchaseReturnActivity, models.PurchaseReturnDeleteActivity]
	contextTimeout time.Duration
}

func NewPurchaseReturnRepository(pst microservice.IPersisterMongo) *PurchaseReturnRepository {

	contextTimeout := time.Duration(15) * time.Second

	insRepo := &PurchaseReturnRepository{
		pst:            pst,
		contextTimeout: contextTimeout,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.PurchaseReturnDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PurchaseReturnInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PurchaseReturnItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.PurchaseReturnActivity, models.PurchaseReturnDeleteActivity](pst)

	return insRepo
}

func (repo PurchaseReturnRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.PurchaseReturnDoc, error) {
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

	doc := models.PurchaseReturnDoc{}
	err := repo.pst.FindOne(ctx, models.PurchaseReturnDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
