package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/advancepayment/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IAdvancePaymentRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.AdvancePaymentDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.AdvancePaymentDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.AdvancePaymentDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.AdvancePaymentInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.AdvancePaymentDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.AdvancePaymentDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.AdvancePaymentItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.AdvancePaymentDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.AdvancePaymentInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.AdvancePaymentInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.AdvancePaymentDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.AdvancePaymentActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.AdvancePaymentDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.AdvancePaymentActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.AdvancePaymentDoc, error)
}

type AdvancePaymentRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.AdvancePaymentDoc]
	repositories.SearchRepository[models.AdvancePaymentInfo]
	repositories.GuidRepository[models.AdvancePaymentItemGuid]
	repositories.ActivityRepository[models.AdvancePaymentActivity, models.AdvancePaymentDeleteActivity]
}

func NewAdvancePaymentRepository(pst microservice.IPersisterMongo) *AdvancePaymentRepository {

	insRepo := &AdvancePaymentRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.AdvancePaymentDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.AdvancePaymentInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.AdvancePaymentItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.AdvancePaymentActivity, models.AdvancePaymentDeleteActivity](pst)

	return insRepo
}
func (repo AdvancePaymentRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.AdvancePaymentDoc, error) {
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

	doc := models.AdvancePaymentDoc{}
	err := repo.pst.FindOne(ctx, models.AdvancePaymentDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
