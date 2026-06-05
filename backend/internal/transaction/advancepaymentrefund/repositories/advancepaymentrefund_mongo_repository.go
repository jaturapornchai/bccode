package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/advancepaymentrefund/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IAdvancePaymentRefundRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.AdvancePaymentRefundDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.AdvancePaymentRefundDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.AdvancePaymentRefundDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.AdvancePaymentRefundInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.AdvancePaymentRefundDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.AdvancePaymentRefundDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.AdvancePaymentRefundItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.AdvancePaymentRefundDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.AdvancePaymentRefundInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.AdvancePaymentRefundInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.AdvancePaymentRefundDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.AdvancePaymentRefundActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.AdvancePaymentRefundDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.AdvancePaymentRefundActivity, error)

	FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.AdvancePaymentRefundDoc, error)
}

type AdvancePaymentRefundRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.AdvancePaymentRefundDoc]
	repositories.SearchRepository[models.AdvancePaymentRefundInfo]
	repositories.GuidRepository[models.AdvancePaymentRefundItemGuid]
	repositories.ActivityRepository[models.AdvancePaymentRefundActivity, models.AdvancePaymentRefundDeleteActivity]
}

func NewAdvancePaymentRefundRepository(pst microservice.IPersisterMongo) *AdvancePaymentRefundRepository {

	insRepo := &AdvancePaymentRefundRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.AdvancePaymentRefundDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.AdvancePaymentRefundInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.AdvancePaymentRefundItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.AdvancePaymentRefundActivity, models.AdvancePaymentRefundDeleteActivity](pst)

	return insRepo
}
func (repo AdvancePaymentRefundRepository) FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.AdvancePaymentRefundDoc, error) {
	filters := bson.M{
		"holdingcode": holdingCode,
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

	doc := models.AdvancePaymentRefundDoc{}
	err := repo.pst.FindOne(ctx, models.AdvancePaymentRefundDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
