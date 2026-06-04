package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/quotation/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IQuotationRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.QuotationDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.QuotationDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.QuotationDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.QuotationInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.QuotationDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.QuotationDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.QuotationItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.QuotationDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.QuotationInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.QuotationInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.QuotationDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.QuotationActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.QuotationDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.QuotationActivity, error)

	FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.QuotationDoc, error)
}

type QuotationRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.QuotationDoc]
	repositories.SearchRepository[models.QuotationInfo]
	repositories.GuidRepository[models.QuotationItemGuid]
	repositories.ActivityRepository[models.QuotationActivity, models.QuotationDeleteActivity]
}

func NewQuotationRepository(pst microservice.IPersisterMongo) *QuotationRepository {

	insRepo := &QuotationRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.QuotationDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.QuotationInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.QuotationItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.QuotationActivity, models.QuotationDeleteActivity](pst)

	return insRepo
}
func (repo QuotationRepository) FindLastDocNo(ctx context.Context, holdingCode string, prefixDocNo string) (models.QuotationDoc, error) {
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

	doc := models.QuotationDoc{}
	err := repo.pst.FindOne(ctx, models.QuotationDoc{}, filters, &doc, &optSort)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
