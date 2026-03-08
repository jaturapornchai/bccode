package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/paid/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IPaidRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.PaidDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PaidDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.PaidDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.PaidInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.PaidDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.PaidItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.PaidDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.PaidInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.PaidInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PaidDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PaidActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PaidDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PaidActivity, error)

	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.PaidDoc, error)
}

type PaidRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PaidDoc]
	repositories.SearchRepository[models.PaidInfo]
	repositories.GuidRepository[models.PaidItemGuid]
	repositories.ActivityRepository[models.PaidActivity, models.PaidDeleteActivity]
}

func NewPaidRepository(pst microservice.IPersisterMongo) *PaidRepository {

	insRepo := &PaidRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.PaidDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PaidInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PaidItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.PaidActivity, models.PaidDeleteActivity](pst)

	return insRepo
}

func (repo PaidRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.PaidDoc, error) {
	filters := bson.M{
		"shopid": shopID,
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

	doc := models.PaidDoc{}
	err := repo.pst.FindOne(ctx, models.PaidDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
