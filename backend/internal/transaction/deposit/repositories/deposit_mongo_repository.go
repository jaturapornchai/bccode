package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/deposit/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IDepositRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.DepositDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.DepositDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.DepositDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.DepositInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.DepositDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.DepositDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.DepositItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.DepositDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.DepositInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.DepositInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DepositDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DepositActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DepositDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DepositActivity, error)

	FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.DepositDoc, error)
}

type DepositRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.DepositDoc]
	repositories.SearchRepository[models.DepositInfo]
	repositories.GuidRepository[models.DepositItemGuid]
	repositories.ActivityRepository[models.DepositActivity, models.DepositDeleteActivity]
}

func NewDepositRepository(pst microservice.IPersisterMongo) *DepositRepository {

	insRepo := &DepositRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.DepositDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.DepositInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.DepositItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.DepositActivity, models.DepositDeleteActivity](pst)

	return insRepo
}
func (repo DepositRepository) FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.DepositDoc, error) {
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

	doc := models.DepositDoc{}
	err := repo.pst.FindOne(ctx, models.DepositDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
