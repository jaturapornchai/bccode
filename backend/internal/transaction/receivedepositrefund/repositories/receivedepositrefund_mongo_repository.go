package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/receivedepositrefund/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IReceiveDepositRefundRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.ReceiveDepositRefundDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ReceiveDepositRefundDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.ReceiveDepositRefundDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.ReceiveDepositRefundInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.ReceiveDepositRefundDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.ReceiveDepositRefundDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.ReceiveDepositRefundItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.ReceiveDepositRefundDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ReceiveDepositRefundInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ReceiveDepositRefundInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ReceiveDepositRefundDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ReceiveDepositRefundActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ReceiveDepositRefundDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ReceiveDepositRefundActivity, error)

	FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.ReceiveDepositRefundDoc, error)
}

type ReceiveDepositRefundRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ReceiveDepositRefundDoc]
	repositories.SearchRepository[models.ReceiveDepositRefundInfo]
	repositories.GuidRepository[models.ReceiveDepositRefundItemGuid]
	repositories.ActivityRepository[models.ReceiveDepositRefundActivity, models.ReceiveDepositRefundDeleteActivity]
}

func NewReceiveDepositRefundRepository(pst microservice.IPersisterMongo) *ReceiveDepositRefundRepository {

	insRepo := &ReceiveDepositRefundRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ReceiveDepositRefundDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ReceiveDepositRefundInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ReceiveDepositRefundItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ReceiveDepositRefundActivity, models.ReceiveDepositRefundDeleteActivity](pst)

	return insRepo
}
func (repo ReceiveDepositRefundRepository) FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.ReceiveDepositRefundDoc, error) {
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

	doc := models.ReceiveDepositRefundDoc{}
	err := repo.pst.FindOne(ctx, models.ReceiveDepositRefundDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
