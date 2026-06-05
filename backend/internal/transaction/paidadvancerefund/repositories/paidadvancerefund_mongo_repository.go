package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/paidadvancerefund/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IPaidAdvanceRefundRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.PaidAdvanceRefundDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PaidAdvanceRefundDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.PaidAdvanceRefundDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.PaidAdvanceRefundInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.PaidAdvanceRefundDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.PaidAdvanceRefundDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.PaidAdvanceRefundItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.PaidAdvanceRefundDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.PaidAdvanceRefundInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.PaidAdvanceRefundInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PaidAdvanceRefundDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PaidAdvanceRefundActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PaidAdvanceRefundDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PaidAdvanceRefundActivity, error)

	FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.PaidAdvanceRefundDoc, error)
}

type PaidAdvanceRefundRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PaidAdvanceRefundDoc]
	repositories.SearchRepository[models.PaidAdvanceRefundInfo]
	repositories.GuidRepository[models.PaidAdvanceRefundItemGuid]
	repositories.ActivityRepository[models.PaidAdvanceRefundActivity, models.PaidAdvanceRefundDeleteActivity]
}

func NewPaidAdvanceRefundRepository(pst microservice.IPersisterMongo) *PaidAdvanceRefundRepository {

	insRepo := &PaidAdvanceRefundRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.PaidAdvanceRefundDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PaidAdvanceRefundInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PaidAdvanceRefundItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.PaidAdvanceRefundActivity, models.PaidAdvanceRefundDeleteActivity](pst)

	return insRepo
}
func (repo PaidAdvanceRefundRepository) FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.PaidAdvanceRefundDoc, error) {
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

	doc := models.PaidAdvanceRefundDoc{}
	err := repo.pst.FindOne(ctx, models.PaidAdvanceRefundDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
