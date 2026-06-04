package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/depositrecord/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IDepositRecordRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.DepositRecordDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.DepositRecordDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.DepositRecordDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.DepositRecordInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.DepositRecordDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.DepositRecordDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.DepositRecordItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.DepositRecordDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.DepositRecordInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.DepositRecordInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DepositRecordDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DepositRecordActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DepositRecordDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DepositRecordActivity, error)

	FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.DepositRecordDoc, error)
}

type DepositRecordRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.DepositRecordDoc]
	repositories.SearchRepository[models.DepositRecordInfo]
	repositories.GuidRepository[models.DepositRecordItemGuid]
	repositories.ActivityRepository[models.DepositRecordActivity, models.DepositRecordDeleteActivity]
}

func NewDepositRecordRepository(pst microservice.IPersisterMongo) *DepositRecordRepository {

	insRepo := &DepositRecordRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.DepositRecordDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.DepositRecordInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.DepositRecordItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.DepositRecordActivity, models.DepositRecordDeleteActivity](pst)

	return insRepo
}
func (repo DepositRecordRepository) FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.DepositRecordDoc, error) {
	filters := bson.M{
		"holding_code": holdingCode,
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

	doc := models.DepositRecordDoc{}
	err := repo.pst.FindOne(ctx, models.DepositRecordDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
