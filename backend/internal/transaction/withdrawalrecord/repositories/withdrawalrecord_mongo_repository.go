package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/withdrawalrecord/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IWithdrawalRecordRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.WithdrawalRecordDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.WithdrawalRecordDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.WithdrawalRecordDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.WithdrawalRecordInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.WithdrawalRecordDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.WithdrawalRecordDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.WithdrawalRecordItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.WithdrawalRecordDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.WithdrawalRecordInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.WithdrawalRecordInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.WithdrawalRecordDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.WithdrawalRecordActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.WithdrawalRecordDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.WithdrawalRecordActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.WithdrawalRecordDoc, error)
}

type WithdrawalRecordRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.WithdrawalRecordDoc]
	repositories.SearchRepository[models.WithdrawalRecordInfo]
	repositories.GuidRepository[models.WithdrawalRecordItemGuid]
	repositories.ActivityRepository[models.WithdrawalRecordActivity, models.WithdrawalRecordDeleteActivity]
}

func NewWithdrawalRecordRepository(pst microservice.IPersisterMongo) *WithdrawalRecordRepository {

	insRepo := &WithdrawalRecordRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.WithdrawalRecordDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.WithdrawalRecordInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.WithdrawalRecordItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.WithdrawalRecordActivity, models.WithdrawalRecordDeleteActivity](pst)

	return insRepo
}
func (repo WithdrawalRecordRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.WithdrawalRecordDoc, error) {
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

	doc := models.WithdrawalRecordDoc{}
	err := repo.pst.FindOne(ctx, models.WithdrawalRecordDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
