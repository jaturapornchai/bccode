package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/pay/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IPayRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.PayDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PayDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.PayDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.PayInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.PayDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.PayItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.PayDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.PayInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.PayInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PayDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PayActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PayDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PayActivity, error)

	FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.PayDoc, error)
}

type PayRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PayDoc]
	repositories.SearchRepository[models.PayInfo]
	repositories.GuidRepository[models.PayItemGuid]
	repositories.ActivityRepository[models.PayActivity, models.PayDeleteActivity]
}

func NewPayRepository(pst microservice.IPersisterMongo) *PayRepository {

	insRepo := &PayRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.PayDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PayInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PayItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.PayActivity, models.PayDeleteActivity](pst)

	return insRepo
}

func (repo PayRepository) FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.PayDoc, error) {
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

	doc := models.PayDoc{}
	err := repo.pst.FindOne(ctx, models.PayDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
