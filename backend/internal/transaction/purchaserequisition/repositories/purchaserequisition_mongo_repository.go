package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/purchaserequisition/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IPurchaseRequisitionRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.PurchaseRequisitionDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PurchaseRequisitionDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.PurchaseRequisitionDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.PurchaseRequisitionInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.PurchaseRequisitionDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.PurchaseRequisitionDoc, error)
	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.PurchaseRequisitionItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.PurchaseRequisitionDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.PurchaseRequisitionInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.PurchaseRequisitionInfo, int, error)
	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchaseRequisitionDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchaseRequisitionActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchaseRequisitionDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchaseRequisitionActivity, error)
	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.PurchaseRequisitionDoc, error)
}

type PurchaseRequisitionRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PurchaseRequisitionDoc]
	repositories.SearchRepository[models.PurchaseRequisitionInfo]
	repositories.GuidRepository[models.PurchaseRequisitionItemGuid]
	repositories.ActivityRepository[models.PurchaseRequisitionActivity, models.PurchaseRequisitionDeleteActivity]
}

func NewPurchaseRequisitionRepository(pst microservice.IPersisterMongo) *PurchaseRequisitionRepository {
	insRepo := &PurchaseRequisitionRepository{
		pst: pst,
	}
	insRepo.CrudRepository = repositories.NewCrudRepository[models.PurchaseRequisitionDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PurchaseRequisitionInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PurchaseRequisitionItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.PurchaseRequisitionActivity, models.PurchaseRequisitionDeleteActivity](pst)
	return insRepo
}

func (repo PurchaseRequisitionRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.PurchaseRequisitionDoc, error) {
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
	doc := models.PurchaseRequisitionDoc{}
	err := repo.pst.FindOne(ctx, models.PurchaseRequisitionDoc{}, filters, &doc, &optSort)
	if err != nil {
		return doc, err
	}
	return doc, nil
}
