package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/receivedeposit/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IReceiveDepositRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.ReceiveDepositDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ReceiveDepositDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.ReceiveDepositDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.ReceiveDepositInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ReceiveDepositDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.ReceiveDepositDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.ReceiveDepositItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.ReceiveDepositDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ReceiveDepositInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ReceiveDepositInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ReceiveDepositDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ReceiveDepositActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ReceiveDepositDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ReceiveDepositActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.ReceiveDepositDoc, error)
}

type ReceiveDepositRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ReceiveDepositDoc]
	repositories.SearchRepository[models.ReceiveDepositInfo]
	repositories.GuidRepository[models.ReceiveDepositItemGuid]
	repositories.ActivityRepository[models.ReceiveDepositActivity, models.ReceiveDepositDeleteActivity]
}

func NewReceiveDepositRepository(pst microservice.IPersisterMongo) *ReceiveDepositRepository {

	insRepo := &ReceiveDepositRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ReceiveDepositDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ReceiveDepositInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ReceiveDepositItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.ReceiveDepositActivity, models.ReceiveDepositDeleteActivity](pst)

	return insRepo
}
func (repo ReceiveDepositRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.ReceiveDepositDoc, error) {
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

	doc := models.ReceiveDepositDoc{}
	err := repo.pst.FindOne(ctx, models.ReceiveDepositDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
