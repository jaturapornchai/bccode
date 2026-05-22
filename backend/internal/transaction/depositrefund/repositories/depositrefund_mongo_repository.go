package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/depositrefund/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IDepositRefundRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.DepositRefundDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.DepositRefundDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.DepositRefundDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.DepositRefundInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.DepositRefundDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.DepositRefundDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.DepositRefundItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.DepositRefundDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.DepositRefundInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.DepositRefundInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DepositRefundDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DepositRefundActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DepositRefundDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DepositRefundActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.DepositRefundDoc, error)
}

type DepositRefundRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.DepositRefundDoc]
	repositories.SearchRepository[models.DepositRefundInfo]
	repositories.GuidRepository[models.DepositRefundItemGuid]
	repositories.ActivityRepository[models.DepositRefundActivity, models.DepositRefundDeleteActivity]
}

func NewDepositRefundRepository(pst microservice.IPersisterMongo) *DepositRefundRepository {

	insRepo := &DepositRefundRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.DepositRefundDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.DepositRefundInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.DepositRefundItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.DepositRefundActivity, models.DepositRefundDeleteActivity](pst)

	return insRepo
}
func (repo DepositRefundRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.DepositRefundDoc, error) {
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

	doc := models.DepositRefundDoc{}
	err := repo.pst.FindOne(ctx, models.DepositRefundDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
