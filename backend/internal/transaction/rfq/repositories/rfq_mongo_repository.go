package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/rfq/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IRFQRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.RFQDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.RFQDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.RFQDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.RFQInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.RFQDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.RFQDoc, error)
	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.RFQItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.RFQDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.RFQInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.RFQInfo, int, error)
	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.RFQDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.RFQActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.RFQDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.RFQActivity, error)
	FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.RFQDoc, error)
}

type RFQRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.RFQDoc]
	repositories.SearchRepository[models.RFQInfo]
	repositories.GuidRepository[models.RFQItemGuid]
	repositories.ActivityRepository[models.RFQActivity, models.RFQDeleteActivity]
}

func NewRFQRepository(pst microservice.IPersisterMongo) *RFQRepository {
	insRepo := &RFQRepository{
		pst: pst,
	}
	insRepo.CrudRepository = repositories.NewCrudRepository[models.RFQDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.RFQInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.RFQItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.RFQActivity, models.RFQDeleteActivity](pst)
	return insRepo
}

func (repo RFQRepository) FindLastDocNo(ctx context.Context, shopID string, prefixDocNo string) (models.RFQDoc, error) {
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
	doc := models.RFQDoc{}
	err := repo.pst.FindOne(ctx, models.RFQDoc{}, filters, &doc, &optSort)
	if err != nil {
		return doc, err
	}
	return doc, nil
}
